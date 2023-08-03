# Recurring

Some of the transactions recur on a regular interval and it might be
useful to know the next due date for such transactions. Recurring page
shows the upcoming or recently missed transactions.

Paisa depends on the posting metadata to identify which transactions
are recurring. This metadata can be added in couple of ways. Let's say
you pay rent every month and you want to mark it as recurring, a
typical journal would like below

```go
2023/07/01 Rent
    Expenses:Rent                             15,000 INR
    Assets:Checking

2023/08/01 Rent
    Expenses:Rent                             15,000 INR
    Assets:Checking
```

You can manually tag a posting by adding `; Recurring: Rent`.

```go
2023/07/01 Rent
    ; Recurring: Rent
    Expenses:Rent                             15,000 INR
    Assets:Checking

2023/08/01 Rent
    ; Recurring: Rent
    Expenses:Rent                             15,000 INR
    Assets:Checking
```

The first part of the metadata before the colon is called tag name. It
should be `Recurring`. The second part is the tag value. This value is
used to group transactions.

Tagging each and every posting can be tiresome. Ledger has a feature
called [Automated Transaction](https://ledger-cli.org/doc/ledger3.html#Automated-Transactions) which can make this process simpler.

```go
= Expenses:Rent
    ; Recurring: Rent
```

The first line is the predicate and the line below it will get added
to any matching posting. By default, it will match the posting account
name. But you can target other attributes like payee, amount etc. You
can find more examples below, more info about predicate is available on
Ledger [docs](https://ledger-cli.org/doc/ledger3.html#Complex-expressions)

```go
= expr payee=~/^PPF$/
    ; Recurring: PPF

= expr payee=~/Mutual Fund/
    ; Recurring: Mutual Fund

= expr 'account=~/Expenses:Insurance/ and (payee=~/HDFC/)'
    ; Recurring: Life Insurance

= expr 'account=~/Expenses:Insurance/ and !(payee=~/HDFC/)'
    ; Recurring: Bike Insurance

= expr payee=~/Savings Interest/
    ; Recurring: Savings Interest
```
