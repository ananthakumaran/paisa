---
description: "Track NAV history of mainland China mutual funds via 天天基金 (Eastmoney)."
---

# 天天基金 (Eastmoney) <sub>:flag_cn:</sub>

The `cn-ttjj` provider fetches daily NAV (单位净值) history for mainland
China mutual funds from [天天基金](https://fund.eastmoney.com/) (a.k.a.
Eastmoney). It is the recommended source for funds that have a 6-digit
Eastmoney code such as `000311` (景顺长城沪深300指数增强).

## Configuration

```yaml
commodities:
  - name: HS300 # (1)!
    type: mutualfund # (2)!
    price:
        provider: cn-ttjj # (3)!
        code: "000311" # (4)!
```

1. commodity name as used in your journal
1. commodity type (always `mutualfund` for this provider)
1. price provider name
1. 6-digit Eastmoney fund code; keep the leading zeros — wrap in quotes
   so YAML treats it as a string

## Data source

The provider parses `Data_netWorthTrend` out of
`https://fund.eastmoney.com/pingzhongdata/<code>.js`, which contains the
fund's entire NAV history. Each entry has a millisecond timestamp and a
NAV value; timestamps are interpreted in `Asia/Shanghai` and truncated
to the trading day.

Delisted (清盘) funds expose an empty array; in that case the provider
returns an empty price list without error so other prices in the run
are not lost.

## Notes

- No API key required; please be respectful of the upstream service and
  avoid polling more than necessary.
- For funds traded only on Hong Kong / overseas markets, prefer
  [Yahoo](../commodities.md#yahoo) or
  [Alpha Vantage](../commodities.md#alpha-vantage) instead.
