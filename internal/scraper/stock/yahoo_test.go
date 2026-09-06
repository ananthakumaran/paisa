package stock

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

const chartResponse = `{"chart":{"result":[{"meta":{"currency":"INR"},"timestamp":[1700000000,1700086400],"indicators":{"quote":[{"close":[100.5,101.25]}]}}],"error":null}}`

// stubYahoo points the scraper at a test server and restores the real
// endpoints, along with the cached session, once the test is done.
func stubYahoo(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(handler)

	cookieURL, crumbURL, chartURL := yahooCookieURL, yahooCrumbURL, yahooChartURL
	yahooCookieURL = server.URL + "/cookie"
	yahooCrumbURL = server.URL + "/crumb"
	yahooChartURL = server.URL + "/chart/%s"
	session = yahooSession{}

	t.Cleanup(func() {
		server.Close()
		yahooCookieURL, yahooCrumbURL, yahooChartURL = cookieURL, crumbURL, chartURL
		session = yahooSession{}
	})

	return server
}

// yahooHandler mimics Yahoo: the cookie endpoint sets a session cookie, the
// crumb endpoint refuses to answer without it, and the chart endpoint refuses
// to answer without a matching crumb.
func yahooHandler(chart http.HandlerFunc) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "A3", Value: "session-cookie", Path: "/"})
		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/crumb", func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie("A3"); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"finance":{"result":null,"error":{"code":"Unauthorized","description":"Invalid Cookie"}}}`)
			return
		}
		fmt.Fprint(w, "test-crumb")
	})

	mux.HandleFunc("/chart/", chart)

	return mux
}

func TestGetTickerSendsCookieAndCrumb(t *testing.T) {
	stubYahoo(t, yahooHandler(func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie("A3"); err != nil {
			t.Errorf("expected the session cookie to be sent along with the chart request")
		}
		if got := r.URL.Query().Get("crumb"); got != "test-crumb" {
			t.Errorf("expected crumb %q, got %q", "test-crumb", got)
		}
		fmt.Fprint(w, chartResponse)
	}))

	response, err := getTicker("TATAPOWER.NS")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result, err := firstResult(response, "TATAPOWER.NS")
	if err != nil {
		t.Fatalf("expected a usable result, got %v", err)
	}

	if result.Meta.Currency != "INR" {
		t.Errorf("expected currency INR, got %q", result.Meta.Currency)
	}
	if len(result.Timestamp) != 2 {
		t.Errorf("expected 2 timestamps, got %d", len(result.Timestamp))
	}
	if got := result.Indicators.Quote[0].Close[0]; got != 100.5 {
		t.Errorf("expected first close 100.5, got %v", got)
	}
}

// The crumb is tied to the session, fetching it once per process is enough.
func TestCrumbIsCachedAcrossTickers(t *testing.T) {
	var crumbRequests int32

	mux := http.NewServeMux()
	mux.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "A3", Value: "session-cookie", Path: "/"})
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/crumb", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&crumbRequests, 1)
		fmt.Fprint(w, "test-crumb")
	})
	mux.HandleFunc("/chart/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, chartResponse)
	})
	stubYahoo(t, mux)

	for _, ticker := range []string{"TATAPOWER.NS", "PREMIERENE.NS", "AXISCETF.NS"} {
		if _, err := getTicker(ticker); err != nil {
			t.Fatalf("expected no error for %s, got %v", ticker, err)
		}
	}

	if got := atomic.LoadInt32(&crumbRequests); got != 1 {
		t.Errorf("expected the crumb to be fetched once, got %d requests", got)
	}
}

// Regression test for #327: a rate limited request answers with a plain text
// body, which must surface as a readable error rather than a JSON parse
// failure.
func TestGetTickerRateLimited(t *testing.T) {
	stubYahoo(t, yahooHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, "Edge: Too Many Requests")
	}))

	_, err := getTicker("TATAPOWER.NS")
	if err == nil {
		t.Fatal("expected an error for a rate limited request")
	}
	if !strings.Contains(err.Error(), "429") || !strings.Contains(err.Error(), "Edge: Too Many Requests") {
		t.Errorf("expected the status and body in the error, got %v", err)
	}
}

// A stale session is retried once with a fresh cookie and crumb.
func TestGetTickerRefreshesCrumbOnce(t *testing.T) {
	var chartRequests, crumbRequests int32

	mux := http.NewServeMux()
	mux.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "A3", Value: "session-cookie", Path: "/"})
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/crumb", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&crumbRequests, 1)
		fmt.Fprintf(w, "crumb-%d", atomic.LoadInt32(&crumbRequests))
	})
	mux.HandleFunc("/chart/", func(w http.ResponseWriter, r *http.Request) {
		// The first request is answered as if the crumb had expired.
		if atomic.AddInt32(&chartRequests, 1) == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Invalid Crumb")
			return
		}
		if got := r.URL.Query().Get("crumb"); got != "crumb-2" {
			t.Errorf("expected the retry to use a refreshed crumb, got %q", got)
		}
		fmt.Fprint(w, chartResponse)
	})
	stubYahoo(t, mux)

	if _, err := getTicker("TATAPOWER.NS"); err != nil {
		t.Fatalf("expected the retry to succeed, got %v", err)
	}
	if got := atomic.LoadInt32(&crumbRequests); got != 2 {
		t.Errorf("expected 2 crumb requests, got %d", got)
	}
}

func TestGetTickerCrumbEndpointFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	// No cookie is set, so the crumb endpoint turns the request down.
	mux.HandleFunc("/crumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"finance":{"result":null,"error":{"code":"Unauthorized","description":"Invalid Cookie"}}}`)
	})
	mux.HandleFunc("/chart/", func(w http.ResponseWriter, r *http.Request) {
		t.Error("the chart endpoint must not be called without a crumb")
	})
	stubYahoo(t, mux)

	_, err := getTicker("TATAPOWER.NS")
	if err == nil {
		t.Fatal("expected an error when the crumb cannot be fetched")
	}
	if !strings.Contains(err.Error(), "crumb") {
		t.Errorf("expected the error to mention the crumb, got %v", err)
	}
}

// Regression test for #337: an empty result used to panic with
// "index out of range [0] with length 0".
func TestGetHistoryEmptyResult(t *testing.T) {
	stubYahoo(t, yahooHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"chart":{"result":null,"error":{"code":"Not Found","description":"No data found, symbol may be delisted"}}}`)
	}))

	prices, err := GetHistory("DELISTED.NS", "Delisted")
	if err == nil {
		t.Fatal("expected an error for a delisted ticker")
	}
	if prices != nil {
		t.Errorf("expected no prices, got %d", len(prices))
	}
	if !strings.Contains(err.Error(), "No data found, symbol may be delisted") {
		t.Errorf("expected the yahoo description in the error, got %v", err)
	}
}

// A 200 response can still carry an empty result.
func TestGetHistoryEmptyResultWithOKStatus(t *testing.T) {
	stubYahoo(t, yahooHandler(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"chart":{"result":[],"error":null}}`)
	}))

	if _, err := GetHistory("EMPTY.NS", "Empty"); err == nil {
		t.Fatal("expected an error for an empty result")
	}
}

func TestFirstResultWithoutQuotes(t *testing.T) {
	response := &Response{Chart: Chart{Result: []Result{{Meta: Meta{Currency: "INR"}}}}}

	if _, err := firstResult(response, "NOQUOTES.NS"); err == nil {
		t.Fatal("expected an error when the result carries no quotes")
	}
}
