# Tax

Paisa provides few features to help with tax filing and tax
optimization.

## Disclaimer

The author of this tool is not an accountant, the calculations are
made based on their interpretation of rules and regulation. This tool
is distributed free hoping it would be useful. The users are advised
to consult a tax accountant before making any decision based on the
information provided here. Check the [LICENSE](https://github.com/ananthakumaran/paisa/blob/master/COPYING) of paisa for more
details.

## Stock Split and Bonus

[Capital Gains](capital-gains.md), [Tax
Harvesting](tax-harvesting.md) and [Schedule AL](schedule-al.md)
take stock splits and bonus issues into account.

A transaction with postings only to a single asset account, all in
commodity units, is considered a stock split. The units of the
existing purchases are adjusted by the split ratio, while their
purchase date and cost stay the same. The amounts booked in the split
transaction are ignored. Either of the below forms can be used.

```ledger
2022/01/01 ABC Split 1:2
    Assets:Equity:ABC                -20 ABC @ 150 INR
    Assets:Equity:ABC                 40 ABC @ 75 INR

2022/01/01 ABC Split 1:2
    Assets:Equity:ABC                 20 ABC @ 0 INR
```

```beancount
2022-01-01 * "ABC Split 1:2"
  Assets:Equity:ABC                  20 ABC {0 INR}
```

Units credited to an asset account against an Income account are
considered a bonus issue. The bonus units are treated as acquired on
the allotment date at zero cost, irrespective of the value booked
against the Income account. Units credited against
`Income:CapitalGains`, `Income:Dividend` or `Income:Interest` are not
considered a bonus, as they are reinvested payouts with a real cost.

```ledger
2022/06/01 ABC Bonus 1:1
    Assets:Equity:ABC                 40 ABC @ 120 INR
    Income:Bonus:ABC

2022/06/01 ABC Bonus 1:1
    Assets:Equity:ABC                 40 ABC
    Income:Bonus:ABC                 -40 ABC
```

```beancount
2022-06-01 * "ABC Bonus 1:1"
  Assets:Equity:ABC                  40 ABC {0 INR}
  Income:Bonus:ABC                    0 INR
```

!!! note

    Ledger CLI drops postings with zero amount, so a bonus booked as
    `40 ABC @ 0 INR` against an Income account would be considered a
    stock split. Use one of the forms above instead.
