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

2. tax_category - This defines how the taxes are calculated as the
   government usually tweaks the tax code regularly with various rules
   like grandfathering, cost inflation index adjustment, etc.

    1. `equity65` - This is for 65% or more investment in Indian
       equity.

    2. `equity35` - This if for 35% or more investment in India equity
       but less than 65%

    3. `debt` - This is for debt funds.

    4. `unlisted_equity` - This is for unlisted foreign on Indian equity.


```yaml
commodities:
  - name: NIFTY
    type: mutualfund
    code: 120716
    harvest: 365
    tax_category: equity65
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
