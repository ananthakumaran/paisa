---
description: "Use Eastmoney (东方财富) to fetch historical daily prices for Chinese A-share stocks in Paisa."
---

# Eastmoney A-Share <sub>:flag_cn:</sub>

Provider code: `cn-eastmoney`

Fetches historical daily close prices for Chinese A-shares (Shanghai and
Shenzhen) from [Eastmoney](https://www.eastmoney.com/) (东方财富).

## Configuration

For A-share stocks you can just write the 6-digit stock code — the
scraper will auto-resolve the Shanghai / Shenzhen prefix:

```yaml
commodities:
  - name: PUFA # (1)!
    type: stock # (2)!
    price:
        provider: cn-eastmoney # (3)!
        code: "600000" # (4)!
```

1. commodity name (matches the symbol used in your ledger journal)
1. commodity type — use `stock`
1. price provider name
1. 6-digit A-share code (Shanghai 6xxxxx, Shenzhen 0xxxxx / 3xxxxx)

### Auto-resolved market prefixes

| Code prefix | Market               | Auto secid    |
| ----------- | -------------------- | ------------- |
| `6xxxxx`    | Shanghai (沪)        | `1.<code>`    |
| `0xxxxx`    | Shenzhen main (深)   | `0.<code>`    |
| `3xxxxx`    | Shenzhen ChiNext     | `0.<code>`    |

### Explicit market.code

For any other market, write the secid directly (`market.code`). Paisa
will pass it to Eastmoney unchanged.

| Prefix | Market                     | Example     |
| ------ | -------------------------- | ----------- |
| `1`    | Shanghai                   | `1.600000`  |
| `0`    | Shenzhen                   | `0.000001`  |
| `116`  | Hong Kong                  | `116.00700` |
| `105`  | Nasdaq                     | `105.AAPL`  |
| `106`  | NYSE                       | `106.BABA`  |

```yaml
commodities:
  - name: TENCENT
    type: stock
    price:
        provider: cn-eastmoney
        code: "116.00700"
```

## Notes

- The daily close (`收盘价`) is used as the price for each trading day.
- Non-trading days, halted stocks and delisted tickers degrade
  gracefully — the scraper returns whatever history is available
  instead of failing the entire sync.
- Successful fetches are cached in-memory for 30 minutes per `secid`
  to avoid hammering Eastmoney during repeated `paisa update` calls.
