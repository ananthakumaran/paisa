# Yahoo Finance (HK, US, China-ADR Stocks + FX)

The `yahoo` price provider fetches daily closing prices from Yahoo Finance for
a wide range of instruments:

- US stocks and ETFs: `AAPL`, `UBER`, `SPY`
- Hong Kong stocks: `0700.HK` (Tencent), `1810.HK` (Xiaomi)
- China ADRs: `BABA`, `BIDU`
- Indices: `^GSPC`, `^HSI`
- FX pairs: `USDCNY=X`, `EURUSD=X` (handled by the FX subsystem; see
  [issue #11](https://github.com/1Feng/paisa/issues/11))

It is implemented in `internal/scraper/yahoo/` and registered under the
provider code `yahoo`. The legacy provider `com-yahoo`
(`internal/scraper/stock/yahoo.go`) is still available for backwards
compatibility.

## Configuration

```yaml
commodities:
  - name: TENCENT # (1)!
    type: stock # (2)!
    price:
      provider: yahoo # (3)!
      code: "0700.HK" # (4)!

  - name: UBER
    type: stock
    price:
      provider: yahoo
      code: UBER
```

1. commodity name
2. commodity type
3. provider code (`yahoo`)
4. Yahoo Finance ticker symbol

## Symbol routing

The provider classifies each input symbol syntactically:

| Pattern               | Kind       | Backend |
|-----------------------|------------|---------|
| `=X` suffix           | FX rate    | FX subsystem (issue #11) |
| `^` prefix            | Index      | Stock fetcher |
| anything else         | Stock/ETF  | Stock fetcher |

If the quote currency reported by Yahoo differs from your `default_currency`,
the value is automatically converted using the corresponding `<from><to>=X`
FX series from Yahoo.

## Resilience

- **Backoff on 429** — retries with exponential backoff (default 4 retries,
  500 ms base). The `Retry-After` header is honoured when present.
- **Graceful 404** — delisted or unknown tickers return `ErrNotFound` and the
  provider logs a warning instead of crashing the sync; previously cached
  prices remain untouched.
- **Network timeout** — the underlying HTTP client uses a 30 s timeout and
  rotates a small pool of common browser User-Agent strings to avoid being
  filtered by Yahoo's edge.

## Example: full HK + US portfolio

```yaml
commodities:
  - name: TENCENT
    type: stock
    price: { provider: yahoo, code: "0700.HK" }
  - name: XIAOMI
    type: stock
    price: { provider: yahoo, code: "1810.HK" }
  - name: UBER
    type: stock
    price: { provider: yahoo, code: UBER }
  - name: APPLE
    type: stock
    price: { provider: yahoo, code: AAPL }
  - name: ALIBABA
    type: stock
    price: { provider: yahoo, code: BABA }
  - name: BAIDU
    type: stock
    price: { provider: yahoo, code: BIDU }
```
