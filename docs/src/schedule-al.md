# Schedule AL

As per the Indian Income tax law, citizens are obligated to report
their entire Assets and Liabilities if the total income exceeds ₹50
lakh. Paisa helps with the computation of the amount.

| code      | Section        | Details                                                                      |
|-----------|----------------|------------------------------------------------------------------------------|
| immovable | A (1)          | Immovable Assets                                                             |
| metal     | B (1) (i)      | Jewellery, bullion etc.                                                      |
| art       | B (1) (ii)     | Archaeological collections, drawings, painting, sculpture or any work of art |
| vechicle  | B (1) (iii)    | Vehicles, yachts, boats and aircrafts                                        |
| bank      | B (1) (iv) (a) | Financial assets: Bank (including all deposits)                              |
| share     | B (1) (iv) (b) | Financial assets: Shares and securities                                      |
| insurance | B (1) (iv) (c) | Financial assets: Insurance policies                                         |
| loan      | B (1) (iv) (d) | Financial assets: Loans and advances given                                   |
| cash      | B (1) (iv) (e) | Financial assets: Cash in hand                                               |
| liability | C (1)          | Liabilities                                                                  |


All you need to do is to specify the accounts that belong to each
section. Paisa will compute the total as on the last day of the
previous financial year.


```yaml
schedule_al:
    - code: bank
      accounts:
          - Assets:Checking
    - code: share
      accounts:
          - Assets:Equity:*
          - Assets:Debt:*

```
