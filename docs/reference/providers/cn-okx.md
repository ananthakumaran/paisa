---
description: "How to track crypto prices in Paisa via the OKX price provider."
---

# OKX Crypto (`cn-okx`)

The `cn-okx` provider fetches daily close prices for crypto pairs from
OKX's public `history-candles` endpoint. No API key is required.

```yaml
commodities:
  - name: BTC
    type: unknown
    price:
        provider: cn-okx
        code: BTC-USDT
```

## Code format

The `code` is an OKX instrument id, in the form `<base>-<quote>`. Examples:

| Symbol     | Description           |
|------------|-----------------------|
| `BTC-USDT` | Bitcoin priced in USDT |
| `ETH-USDT` | Ether priced in USDT   |
| `BTC-USD`  | Bitcoin priced in USD  |
| `SOL-USDT` | Solana priced in USDT  |

Any instrument id that OKX serves via
`https://www.okx.com/api/v5/market/history-candles` will work.

## Currency

Prices are reported in the **quote currency** of the instrument
(typically `USDT` or `USD`). Paisa's FX layer is responsible for
converting these into your `default_currency` (e.g. `CNY`). Make sure
you have an `USDT` → `default_currency` (or `USD` → `default_currency`)
rate available via either inline `P` price directives in your journal
or a configured FX provider.

## Pagination & limits

The provider walks the OKX `after` cursor backwards in time, 100
candles at a time, and stops when an empty page is returned. (Per the
OKX v5 contract, `after=<ts_ms>` returns candles strictly older than
the given timestamp; `before=<ts_ms>` would walk forward — the wrong
direction for backfill.) There is also a hard upper bound of 100 pages
(~27 years of daily data) and a stalled-cursor guard to guarantee
termination if the upstream returns degenerate responses.

## Date / timezone

Crypto markets trade 24/7 and OKX's daily candle boundary is **UTC
midnight**. The stored `Price.Date` is anchored in UTC so the calendar
day does not drift under negative timezones.

## Example transaction

```ledger
2024/01/15 Buy BTC
    Assets:Crypto:BTC     0.05 BTC @ 42000 USDT
    Assets:Bank
```

With the configuration above, Paisa will fetch the historical close
prices for `BTC-USDT` from OKX and value the holding accordingly.
