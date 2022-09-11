# Tax Harvesting

Mutual Fund gains are subject to capital gains tax. The taxation rate
itself differs based on how long you hold your investment, whether
it's debt or equity, whether you have more than 1 Lakh gains in a year
etc. To further complicate things, there might be exit load
applicable.

Paisa takes a generic approach and let the user configure when a
commodity is safe to harvest. There are two fields related to tax
harvesting.

1. harvest - specifies the number of days after which the commodity is
   eligible for tax harvesting. This should be set based on fund exit
   load, fund category etc

2. tax_category - The value can be either `equity` or `debt`. Based on
   which various rules like grandfathering, cost inflation index
   adjustment, etc. would be applied.


```yaml
commodities:
  - name: NIFTY
    type: mutualfund
    code: 120716
    harvest: 365
    tax_category: equity
  - name: ABSTF
    type: mutualfund
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
