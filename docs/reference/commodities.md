---
description: "How to configure commodities in Paisa to track mutual funds, stocks, gold, etc."
---

# Commodities

There are no restrictions on the type of **commodities** that can be
used in Paisa. Anything like gold, mutual fund, NPS, etc can be
tracked as a commodity. Few example transactions can be found below.

```ledger
2019/02/18 NPS
    Assets:Equity:NPS:SBI:E/*(1)!*/  /*(2)!*/15.9378 NPS_SBI_E/*(3)!*/ @ /*(4)!*/23.5289 INR/*(5)!*/
    Assets:Checking

2019/02/21 NPS
    Assets:Equity:NPS:SBI:E      1557.2175 NPS_SBI_E @ 23.8406 INR
    Assets:Checking

2020/06/25 Gold
    Assets:Gold                         40 GOLD @ 4650 INR
    Assets:Checking
```

1.  Account name
2.  Number of units purchased
3.  Commodity Name
4.  Purchase Price per Unit
5.  Currency

Paisa comes with inbuilt support for fetching the latest price of some
commodities like mutual fund, NPS, stocks, etc from few providers. For
others, it will try to use the latest purchase price specified in the
journal. For example, when you enter the second NPS transaction on
`#!ledger 2019/02/21`, the valuation of your existing holdings will be
adjusted based on the new purchase price.

To link a commodity with a commodity price provider, Go to `Configuration`
page and expand the `Commodities` section. You can click the
:fontawesome-solid-circle-plus: icon to add a new one. Edit the name
to commodity name. Click the :fontawesome-solid-pen-to-square: icon
near Price section and select the price provider details. Once done,
save the configuration and click the `Update Prices` from the top right hand
side menu. If you had done everything correctly, you would see the
latest price of the commodity under `Assets` :material-chevron-right:
`Balance`. You can also view the full price history on `Ledger`
:material-chevron-right: `Price`

## Prices in another currency

Suppose you track your finances in INR and own Apple shares. Yahoo gives
you Apple's price in USD, such as 200 USD per share. To show its value in
INR, Paisa also needs the USD-to-INR exchange rate.

In Apple's price settings, enter **USD** in **Quote commodity**. Then add a
USD commodity with Yahoo's `USDINR=X` as its price source:

```yaml
default_currency: INR
commodities:
  - name: APPLE
    type: stock
    price:
      provider: com-yahoo
      code: AAPL
      commodity: USD # Apple's price is in USD
  - name: USD
    type: currency
    price:
      provider: com-yahoo
      code: USDINR=X
```

Save the configuration and select **Update Prices**. If one USD is worth
83 INR, Paisa shows the Apple share's value as 16,600 INR.

For an Indian mutual fund whose price is already in INR, leave
**Quote commodity** empty.

## MF API Mutual Fund <sub>:flag_in:</sub>

To automatically track the latest value of your mutual funds holdings,
you need to link the commodity and the fund scheme code.

```yaml
commodities:
  - name: NIFTY # (1)!
    type: mutualfund # (2)!
    price:
        provider: in-mfapi # (3)!
        code: 120716 # (4)!
```

1. commodity name
1. commodity type
1. price provider name
1. mutual fund scheme code

The example configuration above links nifty commodity with the respective
mutual fund scheme code.

### Reporting in another currency

MFAPI quotes NAVs in INR. With USD reporting, set the fund's quote commodity
to INR and configure an INR-to-USD price source:

```yaml
default_currency: USD
commodities:
  - name: NIFTY
    type: mutualfund
    price:
      provider: in-mfapi
      code: 120716
      commodity: INR
  - name: INR
    type: currency
    price:
      provider: com-yahoo
      code: INRUSD=X
```

Use the same setup for INR-denominated FinBodhi NPS and metal prices.

## Yahoo <sub>:globe_with_meridians:</sub>

To automatically track the latest value of your stock holdings,
you need to link the commodity and the stock ticker name.

```yaml
commodities:
  - name: APPLE # (1)!
    type: stock # (2)!
    price:
        provider: com-yahoo # (3)!
        code: AAPL # (4)!
        commodity: USD # (5)!
```

1. commodity name
1. type
1. price provider name
1. stock ticker code
1. currency of the stock price

Stock prices are fetched from Yahoo Finance. The ticker code should match
the code used on Yahoo. For INR reporting, add the USD exchange-rate source
shown in [Prices in another currency](#prices-in-another-currency).

## FinBodhi Stocks <sub>:globe_with_meridians:</sub>

Supports Indian and US stocks and ETFs, and GIFT City funds. Choose
**FinBodhi Stocks** in the price provider picker and search by name,
ticker or ISIN.

```yaml
commodities:
  - name: RELIANCE
    type: stock
    price:
      provider: com-finbodhi-stock
      code: INE002A01018
```

For US stocks and GIFT City funds, set **Quote commodity** to USD and
configure a [USD exchange-rate source](#prices-in-another-currency).

## Alpha Vantage <sub>:globe_with_meridians:</sub>

Supports 100,000+ stocks, ETFs, mutual funds etc. Set the quote commodity
and configure any required currency source as described above.

```yaml
commodities:
  - name: RELIANCE # (1)!
    type: stock # (2)!
    price:
        provider: co-alphavantage # (3)!
        code: 7GAURDT55LU7OQKK:RELIANCE.BSE:INR # (4)!
```

1. commodity name
1. type
1. price provider name
1. code

Code is made of three parts, `apiKey`, `symbol` and `currency`. For
example in `7GAURDT55LU7OQKK:RELIANCE.BSE:INR`, `7GAURDT55LU7OQKK` is
the api key, `RELIANCE.BSE` is the symbol and `INR` is the
currency. Alpha Vantage provides [free api key](https://www.alphavantage.co/support/#api-key) with 25 requests per day
limit.

## FinBodhi NPS <sub>:flag_in:</sub>

To automatically track the latest value of your nps funds holdings,
you need to link the commodity and the fund scheme code.

```yaml
commodities:
  - name: NPS_HDFC_E # (1)!
    type: nps # (2)!
    price:
        provider: com-finbodhi-nps # (3)!
        code: SM008002 # (4)!
```

1. commodity name
1. type
1. price provider name
1. nps fund scheme code

The example configuration above links NPS fund commodity with their
respective NPS fund scheme code.

## FinBodhi SIF <sub>:flag_in:</sub>

Supports NAV history for [Specialized Investment Funds](https://www.amfiindia.com/sif) in India. In the
price provider picker, choose **FinBodhi SIF India** and search by scheme
name or SIF number. Each result shows the full NAV name and number;
select the correct plan and option.

```yaml
commodities:
  - name: ALTIVA_EX_TOP100_LONG_SHORT
    type: sif
    price:
      provider: com-finbodhi-sif
      code: SIF-120
```


## FinBodhi Metals <sub>:flag_in:</sub>

To automatically track the latest price of gold or silver at various
level of purity, you need to link the commodity and the metal
code. The price is for 1 gram of the metal.

```yaml
commodities:
  - name: GOLD # (1)!
    type: metal # (2)!
    price:
        provider: com-finbodhi-metal # (3)!
        code: gold-999 # (4)!
```

1. commodity name
1. type
1. price provider name
1. metal name along with purity


The following metals and purity combinations are supported.

| Metal  | Purity | Code       |
|--------|--------|------------|
| Gold   | 999    | gold-999   |
| Gold   | 995    | gold-995   |
| Gold   | 916    | gold-916   |
| Gold   | 750    | gold-750   |
| Gold   | 585    | gold-585   |
| Silver | 999    | silver-999 |


## RealEstate

Some commodities like real estate are bought once and the price
changes over time. Ledger allows you to set the price as on date.

```ledger
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

```ledger
P 2023/05/01 00:00:00 USD 81.75 INR

2023/05/01 Freelance Income
    ;; conversion rate will be picked up
    ;; from the price directive above
    Income:Freelance      -100 USD
    Assets:Checking

2023/06/01 Freelance Income
    ;; conversion rate is specified inline
    Income:Freelance      -200 USD @ 82.75 INR
    Assets:Checking

2023/07/01 Netflix
    ;; if not available for a date,
    ;; will use previous known conversion rate (82.75)
    Expenses:Entertainment  10 USD
    Assets:Checking
```

## Update

Paisa fetches the latest price of the commodities only when you
**update** the prices. This can be done via UI using the dropdown in
the top right hand side corner or via `paisa update` command. Make
sure to update the prices after you make any changes to your journal
file or you want to fetch the latest value of the commodities.
