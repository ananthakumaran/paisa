package metal

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestPricesPerGram(t *testing.T) {
	for _, tc := range []struct {
		code, close, want string
	}{
		{code: "gold-999", close: "137576", want: "13757.6"},
		{code: "silver-999", close: "176425", want: "176.425"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			original := http.DefaultClient
			http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				require.Equal(t, "/api/metal/"+tc.code+"/price.json", req.URL.Path)
				body := fmt.Sprintf(`{"data":[{"date":"2026-11-03","close":%s}]}`, tc.close)
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
			})}
			t.Cleanup(func() { http.DefaultClient = original })

			prices, err := (&PriceProvider{}).GetPrices(tc.code, "METAL")
			require.NoError(t, err)
			require.Len(t, prices, 1)
			require.Equal(t, tc.want, prices[0].Value.String())
		})
	}
}
