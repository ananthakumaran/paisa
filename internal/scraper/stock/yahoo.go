package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/google/btree"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/utils"
)

var UserAgents = []string{
	// Chrome
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",

	// # Firefox
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.7; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (X11; Linux i686; rv:135.0) Gecko/20100101 Firefox/135.0",

	// # Safari
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_7_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Safari/605.1.15",

	// # Edge
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/131.0.2903.86",
}

type UserAgent struct {
	sync.Once
	name string
}

var agent UserAgent

func selectAgent() {
	agent.name = UserAgents[rand.Intn(len(UserAgents))]
}

// Yahoo rejects requests that do not carry a session cookie along with a
// matching crumb. Fetching the cookie first and then the crumb mirrors what a
// browser does; without it the API replies with 401 "Invalid Cookie" or a plain
// text "Edge: Too Many Requests" body.
var (
	yahooCookieURL = "https://fc.yahoo.com"
	yahooCrumbURL  = "https://query1.finance.yahoo.com/v1/test/getcrumb"
	yahooChartURL  = "https://query2.finance.yahoo.com/v8/finance/chart/%s"
)

type yahooSession struct {
	mutex  sync.Mutex
	client *http.Client
	crumb  string
}

var session yahooSession

// httpClient returns a http.Client backed by a cookie jar, so that the session
// cookie obtained while fetching the crumb is sent along with every subsequent
// request.
func (s *yahooSession) httpClient() (*http.Client, error) {
	if s.client == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
		s.client = &http.Client{Jar: jar, Timeout: 30 * time.Second}
	}
	return s.client, nil
}

func (s *yahooSession) get(url string) (*http.Response, error) {
	client, err := s.httpClient()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	agent.Do(func() { selectAgent() })
	req.Header.Add("User-Agent", agent.name)

	return client.Do(req)
}

// getCrumb returns the cached crumb, fetching a new one when the cache is empty
// or when a refresh is requested. The crumb is tied to the session cookie held
// by the cookie jar, so both are obtained together and reused for every ticker
// of a sync.
func (s *yahooSession) getCrumb(refresh bool) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.crumb != "" && !refresh {
		return s.crumb, nil
	}

	// Seeds the session cookie. Yahoo answers with 404 here, the body is
	// irrelevant, only the Set-Cookie header matters.
	resp, err := s.get(yahooCookieURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch yahoo session cookie: %w", err)
	}
	resp.Body.Close()

	resp, err = s.get(yahooCrumbURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch yahoo crumb: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch yahoo crumb, status %d: %s", resp.StatusCode, summarize(respBytes))
	}

	crumb := strings.TrimSpace(string(respBytes))
	if crumb == "" || strings.ContainsAny(crumb, "{<") {
		return "", fmt.Errorf("received an unexpected yahoo crumb: %s", summarize(respBytes))
	}

	s.crumb = crumb
	return s.crumb, nil
}

// summarize trims a response body so that it can be embedded in an error
// message without flooding the log.
func summarize(body []byte) string {
	text := strings.TrimSpace(string(body))
	if len(text) > 200 {
		return text[:200] + "..."
	}
	return text
}

type Quote struct {
	Close []float64
}

type Indicators struct {
	Quote []Quote
}

type Meta struct {
	Currency string
}

type Result struct {
	Timestamp  []int64
	Indicators Indicators
	Meta       Meta
}

type ResponseError struct {
	Code        string
	Description string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Description)
}

type Chart struct {
	Result []Result
	Error  *ResponseError
}
type Response struct {
	Chart Chart
}

type ExchangePrice struct {
	Timestamp int64
	Close     float64
}

func (p ExchangePrice) Less(o btree.Item) bool {
	return p.Timestamp < (o.(ExchangePrice).Timestamp)
}

func GetHistory(ticker string, commodityName string) ([]*price.Price, error) {
	log.Info("Fetching stock price history from Yahoo")
	response, err := getTicker(ticker)
	if err != nil {
		return nil, err
	}

	result, err := firstResult(response, ticker)
	if err != nil {
		return nil, err
	}

	var prices []*price.Price
	needExchangePrice := false
	var exchangePrice *btree.BTree

	if !utils.IsCurrency(result.Meta.Currency) {
		needExchangePrice = true
		exchangeTicker := fmt.Sprintf("%s%s=X", result.Meta.Currency, config.DefaultCurrency())
		exchangeResponse, err := getTicker(exchangeTicker)
		if err != nil {
			return nil, err
		}

		exchangeResult, err := firstResult(exchangeResponse, exchangeTicker)
		if err != nil {
			return nil, err
		}

		exchangePrice = btree.New(2)
		for i, t := range exchangeResult.Timestamp {
			exchangePrice.ReplaceOrInsert(ExchangePrice{Timestamp: t, Close: exchangeResult.Indicators.Quote[0].Close[i]})
		}
	}

	for i, timestamp := range result.Timestamp {
		date := time.Unix(timestamp, 0)
		value := result.Indicators.Quote[0].Close[i]

		if needExchangePrice {
			exchangePrice := utils.BTreeDescendFirstLessOrEqual(exchangePrice, ExchangePrice{Timestamp: timestamp})
			value = value * exchangePrice.Close
		}

		price := price.Price{Date: date, CommodityType: config.Stock, CommodityID: ticker, CommodityName: commodityName, Value: decimal.NewFromFloat(value)}

		prices = append(prices, &price)
	}
	return prices, nil
}

// firstResult validates that the response actually carries a usable series.
// Yahoo returns an empty result for delisted or misspelled tickers, and
// indexing into it blindly used to panic and take the whole sync down.
func firstResult(response *Response, ticker string) (*Result, error) {
	if len(response.Chart.Result) == 0 {
		if response.Chart.Error != nil {
			return nil, fmt.Errorf("yahoo returned no data for %s, %s", ticker, response.Chart.Error.Error())
		}
		return nil, fmt.Errorf("yahoo returned no data for %s", ticker)
	}

	result := response.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 || len(result.Indicators.Quote[0].Close) == 0 {
		return nil, fmt.Errorf("yahoo returned no price quotes for %s", ticker)
	}

	return &result, nil
}

type statusError struct {
	StatusCode int
	Body       string
}

func (e *statusError) Error() string {
	return fmt.Sprintf("yahoo request failed with status %d: %s", e.StatusCode, e.Body)
}

// retryable reports whether the request was turned down because the session
// went stale, in which case a fresh cookie and crumb are worth a try.
func retryable(err error) bool {
	statusErr, ok := err.(*statusError)
	if !ok {
		return false
	}

	switch statusErr.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests:
		return true
	}
	return false
}

func getTicker(ticker string) (*Response, error) {
	response, err := fetchTicker(ticker, false)
	if err != nil && retryable(err) {
		log.WithError(err).Info("Refreshing Yahoo crumb and retrying")
		return fetchTicker(ticker, true)
	}

	return response, err
}

func fetchTicker(ticker string, refreshCrumb bool) (*Response, error) {
	crumb, err := session.getCrumb(refreshCrumb)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(yahooChartURL, ticker) + fmt.Sprintf("?interval=1d&range=50y&crumb=%s", crumb)
	resp, err := session.get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response

	// A rejected request may carry a plain text body such as
	// "Edge: Too Many Requests", so the status is checked before the body is
	// treated as JSON.
	if resp.StatusCode != http.StatusOK {
		if err := json.Unmarshal(respBytes, &response); err == nil && response.Chart.Error != nil {
			return nil, fmt.Errorf("yahoo request for %s failed, %s", ticker, response.Chart.Error.Error())
		}
		return nil, &statusError{StatusCode: resp.StatusCode, Body: summarize(respBytes)}
	}

	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse yahoo response for %s: %w", ticker, err)
	}

	return &response, nil
}

type YahooPriceProvider struct {
}

func (p *YahooPriceProvider) Code() string {
	return "com-yahoo"
}

func (p *YahooPriceProvider) Label() string {
	return "Yahoo Finance"
}

func (p *YahooPriceProvider) Description() string {
	return "Supports a large set of stocks, ETFs, mutual funds, currencies, bonds, commodities, and cryptocurrencies. The stock price will be automatically converted to your default currency using the yahoo exchange rate."
}

func (p *YahooPriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Ticker", ID: "ticker", Help: "Stock ticker symbol, can be located on Yahoo's website. For example, AAPL is the ticker symbol for Apple Inc. (AAPL)", InputType: "text"},
	}
}

func (p *YahooPriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{}
}

func (p *YahooPriceProvider) ClearCache(db *gorm.DB) {
}

func (p *YahooPriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	return GetHistory(code, commodityName)
}
