# Commodities

There are no restrictions on the type of **commodities** that can be
used in **paisa**. Anything like gold, mutual fund, NPS, etc can be
tracked as a commodity. Few example transactions can be found below.

```go
2019/02/18 NPS
    Assets:Equity:NPS:SBI:E           15.9378 NPS_SBI_E @ 23.5289 INR
;// account name                      units   commodity   purchase price
;//                                           name        per unit
    Assets:Checking

2019/02/21 NPS
    Assets:Equity:NPS:SBI:E           1557.2175 NPS_SBI_E @ 23.8406 INR
    Assets:Checking

2020/06/25 Gold
    Assets:Gold                       40 GOLD @ 4650 INR
    Assets:Checking
```

**paisa** comes with inbuilt support for fetching the latest price of
some commodities like mutual fund and NPS. For others, it will try to
use the latest purchase price specified in the journal. For example,
when you enter the second NPS transaction on `2019/02/21`, the
valuation of your existing holdings will be adjusted based on the new
purchase price.

## Mutual Fund

To automatically track the latest value of your mutual funds holdings,
you need to link the commodity and the fund scheme code.

```yaml
commodities:
  - name: NIFTY      # commodity name
    type: mutualfund # type
    code: 120716     # mutual fund scheme code
  - name: NIFTY_JR
    type: mutualfund
    code: 120684
```

The example config above links two commodities with their respective
mutual fund scheme code. The scheme code can be found using the
*search* command.

```shell
❯ paisa search mutualfund
INFO Using config file: /home/john/finance/paisa.yaml
INFO Using cached results; pass '-u' to update the cache
✔ ICICI Prudential Asset Management Company Limited
✔ ICICI Prudential Nifty Next 50 Index Fund - Direct Plan -  Growth
INFO Mutual Fund Scheme Code: 120684
```

## NPS

To automatically track the latest value of your nps funds holdings,
you need to link the commodity and the fund scheme code.

```yaml
commodities:
  - name: NPS_HDFC_E      # commodity name
    type: nps             # type
    code: SM008002        # nps fund scheme code
```

The example config above links NPS fund commodity with their
respective NPS fund scheme code. The scheme code can be found using
the *search* command.

```
❯ paisa search nps
INFO Using config file: /home/john/finance/paisa.yaml
INFO Using cached results; pass '-u' to update the cache
✔ HDFC Pension Management Company Limited
✔ HDFC PENSION MANAGEMENT COMPANY LIMITED SCHEME C - TIER I
INFO NPS Fund Scheme Code: SM008002
```

## Stock

To automatically track the latest value of your stock holdings,
you need to link the commodity and the stock ticker name.

```yaml
commodities:
  - name: APPLE           # commodity name
    type: stock           # type
    code: AAPL            # stock ticker code
```

Stock prices are fetched from yahoo finance website. The ticker code
should match the code used in yahoo.

## RealEstate

Some commodities like real estate are bought once and the price
changes over time. Ledger allows you to set the price as on date.

```go
2014/01/01 Home purchase
    Assets:House                                1 APT @ 4000000 INR
    Liabilities:Homeloan

P 2016/01/01 00:00:00 APT 5000000 INR
P 2018/01/01 00:00:00 APT 6500000 INR
P 2020/01/01 00:00:00 APT 6700000 INR
P 2021/01/01 00:00:00 APT 6300000 INR
P 2022/01/01 00:00:00 APT 8000000 INR
```

## Currencies

If you need to deal with multiple currencies, just treat them as you
would treat any commodity. Since paisa is a reporting tool, it will
always try to convert other currencies to the
[default_currency](./config.md). As long as the exchange rate from a currency to
default\_currency is available, paisa would work without issue

```go
P 2023/05/01 00:00:00 USD 81.75 INR

2023/05/01 Freelance Income
    ;; conversion rate will be picked up from the price directive above
    Income:Freelance                            -100 USD
    Assets:Checking

2023/06/01 Freelance Income
    ;; conversion rate is specified inline
    Income:Freelance                            -200 USD @ 82.75 INR
    Assets:Checking

2023/07/01 Netflix
    ;; if not available for a date, will use previous known conversion rate (82.75)
    Expenses:Entertainment                        10 USD
    Assets:Checking
```

## Updates

**paisa** fetches the latest price of the commodities only when
*update* command is used. Make sure to run `paisa update` command
after you make any changes to your journal file or you want to fetch
the latest value of the commodities.
