# Tax Harvesting

Gains from asset types like Mutual Fund, Stock are subject to capital
gains tax. The taxation rate itself differs based on how long you hold
your investment, whether it's debt or equity, whether you have more
than 1 Lakh gains in a year etc. To further complicate things, there
might be exit load applicable.

Paisa takes a generic approach and let the user configure when a
commodity is safe to harvest. There are two fields related to tax
harvesting.

1. harvest - specifies the number of days after which the commodity is
   eligible for tax harvesting. This should be set based on fund exit
   load, fund category etc

2. tax_category - Choose the category that matches your investment below.

## Tax categories

| `tax_category` | Use it for |
| --- | --- |
| `equity65` | Stocks traded on an Indian exchange and equity mutual funds, such as Nifty 50 index funds and NIFTYBEES. |
| `equity35` | Hybrid mutual funds that don’t qualify as equity or debt funds, such as a 40% equity / 60% debt fund. |
| `debt` | Debt mutual funds that don’t trade on a stock exchange. |
| `listed_debt` | Debt funds that trade on an Indian stock exchange, such as LIQUIDBEES. |
| `unlisted_equity` | Company shares that don’t trade on an Indian stock exchange. |

```yaml
commodities:
  - name: NIFTY
    type: mutualfund
    price:
      provider: in-mfapi
      code: 120716
    harvest: 365
    tax_category: equity65
  - name: ABSTF
    type: mutualfund
    price:
      provider: in-mfapi
      code: 119533
    harvest: 1095
    tax_category: debt
```

The tax harvest page will show the list of accounts eligible for
harvesting.

## Multiple Folios

Paisa uses FIFO method within an Account. If you have multiple folios,
this might result in incorrect values. The issue can be solved by
using a different Account for each folio.
