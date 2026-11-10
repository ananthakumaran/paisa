# Capital Gains

Mutual Funds are considered as capital asset and capital gains tax
will be charged on any gains. Capital Gains page shows the financial
year wise capital gains tax based on the withdrawal transaction. The
sum total, account wise breakdown, and per buy/sell match information
are available.

Only commodities with `tax_category` are considered. Please check [Tax
Harvesting](tax-harvesting.md) page for configuration details.

### Gain

The difference between the purchase and sell price.

### Taxable Gain

The gain after applying Indexation or Grandfathering provision.


Based on the duration between the purchase and sell date, an asset
might be classified as Short Term or Long Term.

### Short Term Tax

This would be a percentage of the Taxable Gain. The percentage varies
based on the asset class.

### Long Term Tax

This would be a percentage of the Taxable Gain. The percentage varies
based on the asset class.

### Taxable at Slab Rate

There are cases where the gain would be added to taxable income and
charged **Slab** rate. Since the tax rate would depend on the person,
the whole taxable amount is shown instead of the tax. You can multiply
this with your slab rate to get the tax amount.

## Stock Split

A stock split is a transaction that removes the old units and adds the
new units under the **same** account, without involving any other
account or currency. It's not considered as a sale. The existing
purchases are carried forward with the same purchase date and total
cost, only the number of units and the unit price would change. The
same applies to reverse split.

=== "ledger"

    ```ledger
    2019/06/03 Split 1:2
        Assets:Equity:ACME           -10 ACME @ 300 INR
        Assets:Equity:ACME            20 ACME @ 150 INR
    ```

=== "hledger"

    ```ledger
    2019/06/03 Split 1:2
        Assets:Equity:ACME           -10 ACME @ 300 INR
        Assets:Equity:ACME            20 ACME @ 150 INR
    ```

=== "beancount"

    ```
    2019-06-03 * "Split 1:2"
      Assets:Equity:ACME             -10 ACME {300 INR}
      Assets:Equity:ACME              20 ACME {150 INR}
    ```

A split applies to all the units held in the account on that day. If
there are multiple lots, you can specify them in a single transaction
or as a separate transaction for each lot. Splits and bonuses are
applied before the purchases and sales made on the same day.

If the units were purchased before 1st Feb 2018, the price used for
grandfathering is also adjusted for the splits that happened later,
unless the price is from Yahoo, which is already adjusted.

## Bonus

Bonus units should come from an income account. The cost of the bonus
units is considered as zero and the date of the bonus as the purchase
date. The cost and purchase date of the existing units are not
affected.

```ledger
2020/01/02 Bonus 1:1
    Assets:Equity:ACME                20 ACME
    Income:Bonus:ACME                -20 ACME
```

You can also specify the price as zero, `#!ledger 20 ACME @ 0 INR` in
case of ledger and hledger or `20 ACME {0 INR}` in case of beancount.

Units that come from `#!ledger Income:Salary`, `#!ledger
Income:Dividend` and `#!ledger Income:Interest` are **not** considered
as bonus, as they are usually taxed when you receive them. The market
value on that day would be the cost. Likewise, if the price is
specified and the income is recorded in currency, the specified price
would be the cost.

```ledger
2020/01/01 RSU Vest
    Assets:Equity:ACME                10 ACME @ 50 INR
    Income:Salary:Acme              -500 INR
```

## Sale

If the units are sold at [lot price](../accounts.md#capital-gains),
the gain posted to the corresponding `#!ledger
Income:CapitalGains:{name}` account is added to arrive at the sell
price.
