# Commodities

There are no restrictions on the type of **commodities** that can be
used in **paisa**. Anything like gold, mutual fund, NPS, etc can be
tracked as a commodity. Few example transactions can be found below.

```go
2019/02/18 NPS
    Asset:Equity:NPS:SBI:E            15.9378 NPS_SBI_E @ 23.5289 INR
;// account name                      units   commodity   purchase price
;//                                           name        per unit
    Checking

2019/02/21 NPS
    Asset:Equity:NPS:SBI:E            1557.2175 NPS_SBI_E @ 23.8406 INR
    Checking

2020/06/25 Gold
    Asset:Gold                        40 GOLD @ 4650 INR
    Checking
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

## Updates

**paisa** fetches the latest price of the commodities only when
*update* command is used. Make sure to run `paisa update` command
after you make any changes to your journal file or you want to fetch
the latest value of the commodities.
