---
description: "How to configure recurring transactions in Paisa"
---

# Recurring

Some of the transactions recur on a regular interval and it might be
useful to know the next due date for such transactions. Recurring page
shows the upcoming or recently missed transactions.

Paisa depends on the posting metadata to identify which transactions
are recurring. This metadata can be added in couple of ways. Let's say
you pay rent every month and you want to mark it as recurring, a
typical journal would like below

```ledger
2023/07/01 Rent
    Expenses:Rent             15,000 INR
    Assets:Checking

2023/08/01 Rent
    Expenses:Rent             15,000 INR
    Assets:Checking
```

You can manually tag a posting by adding `; Recurring: Rent`.

```ledger
2023/07/01 Rent
    ; Recurring: Rent
    Expenses:Rent             15,000 INR
    Assets:Checking

2023/08/01 Rent
    ; Recurring: Rent
    Expenses:Rent             15,000 INR
    Assets:Checking
```

The first part of the metadata before the colon is called tag name. It
should be `Recurring`. The second part is the tag value. This value is
used to group transactions.

Tagging each and every posting can be tiresome. Ledger has a feature
called [Automated Transaction](https://ledger-cli.org/doc/ledger3.html#Automated-Transactions) which can make this process simpler.

```ledger
= Expenses:Rent
    ; Recurring: Rent
```

The first line is the predicate and the line below it will get added
to any matching posting. By default, it will match the posting account
name. But you can target other attributes like payee, amount etc. You
can find more examples below, more info about predicate is available on
Ledger [docs](https://ledger-cli.org/doc/ledger3.html#Complex-expressions)

```ledger
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

## Period

Paisa will try to infer the recurring period of the transactions
automatically, but this might not be perfect. Recurring period can
also be explicitly specified via metadata.

```ledger
= expr payee=~/Savings Interest/
    ; Recurring: Savings Interest
    ; Period: L MAR,JUN,SEP,DEC ?
```

Let's say your bank deposits the interest on the last day of the last
month of the quarter, we can specify like the example above. Paisa
editor recognizes **period syntax** and shows the upcoming 3 schedules
right next to period metadata.

```
┌─────────── day of the month 1-31
│  ┌─────────── month 1-12
│  │  ┌─────────── day of the week 0-6 (Sunday to Saturday)
│  │  │
1  *  ?
```

The syntax of the period is similar to [cron](https://en.wikipedia.org/wiki/Cron), with the omission of
seconds and hours.

| Field        | Allowed values      | Special characters |
|--------------|---------------------|--------------------|
| Day of month | `1–31`              | `* , - ? L W`      |
| Month        | `1-12` or `JAN-DEC` | `* , -`            |
| Day of week  | `0-6` or `SUN-SAT`  | `* , - ? L`        |

`*` also known as wildcard represents all valid values. `?` means you
want to omit the field, usually you use it on the day of month or day
of week. `L` means last day of the month or week. `,` can be used to
specify multiple entries. `-` can be used to specify range. `W` means
the closest business day to given day of month

Multiple cron expressions can be specified by joining them using
`|`. Refer the [wikipedia](https://en.wikipedia.org/wiki/Cron) for more information. If you are not
sure, just type it out and the editor will show you whether it is
valid and the next 3 schedules if valid.


### Examples

* Last day of every month `#!ledger ; Period: L * ?`
* 5<sup>th</sup> every month `#!ledger ; Period: 5 * ?`
* Every Sunday `#!ledger ; Period: ? * 0`
* 1<sup>st</sup> of Jan and 7<sup>th</sup> of Feb `#!ledger ; Period 1 JAN ? | 7 FEB ?`
* Closest business day to the 15<sup>th</sup> day of every month. `#!ledger ; Period 15W * ?`
