# Accounts

If you take a typical transaction, money moves between two parties. In
ledger, you call them **Account**. At least two accounts are involved
in any transaction. Accounts are also hierarchical. This helps with
the organization. For example you can treat EPF and PPF as two
accounts namely `Assets:Debt:EPF` and `Assets:Debt:PPF`. The
hierarchy helps you ask questions like what is the balance of
`Assets:Debt`, which will include both EPF and PPF

Even though **ledger** doesn't have any strict Account naming
convention, **paisa** expects you to follow the standard naming
convention.

There are four types of account namely

1. Assets
1. Liabilities
1. Income
1. Expenses

All the accounts you create should be under one of these
accounts. This naming convention is a necessity, because without
which, it's not possible to tell whether you are spending money or
investing money. A transaction from Assets account to Expenses account
implies that you are spending money.

Money typically flows from Income to Assets, from Assets to either
Expenses, Liabilities or other Assets, from Liabilities to Expenses.

``` mermaid
graph LR
  I[Income] --> A[Assets];
  A --> E[Expenses];
  A --> A;
  A --> L[Liabilities];
  L --> E[Expenses];
```

As a general principle, try not to create too many accounts at second
level. The UI works best when you create less than or equal to 12
second level accounts under each type. For example, you can have 12
accounts under `Expenses`. But if you want more, try to add them under
3 level, example `Expenses:Food:Subway`.


### Assets

All your assets should go under `Assets:`. The level of granularity is
up to you. The recommended convention is to use
`Assets:{instrument_type}:{instrument_name}`. The instrument type may
be `Cash`, `Equity`, `Debt`, etc. The instrument name may be the name of
the fund, stock, etc

### Checking

`Assets:Checking` is a special account where you keep your money for
daily use. This will be included in your net worth, but will not be
treated as an investment. So gain page for example, will exclude this
account and won't show the returns.


### Income

All your income should come from `Income:`. The typical way is to
treat each employer as a separate account.

* `Income:Salary:{company}` - salary debit account

### Interest

`Income:Interest` is a special type of account from the perspective of
returns calculation. Let's assume you have bought APPLE stock. You
might be buying them at regular intervals. To calculate your returns,
we can compute the difference between purchase price and current
price.

Now in case of FD, you will get your interest credited to your
account. The returns is the difference between the amount you
deposited and the final balance. It's essential we need to know which
transactions are deposits and which are interest credits.

Any money that comes from the sub account of `Income:Interest` will be
treated as interest. This convention allows paisa to calculate the
returns of any debt instrument without explicitly specifying anything
else.

* `Income:Interest:{name}` - interest debit account

### Tax

Income tax paid to government should be credited to `Expenses:Tax`
account. This is used to calculate your Net Income and your Savings
Rate.

### Expenses

All your expenses should go to `Expenses:{category}` accounts. You can
also have more than 2 levels as well. The expense page will roll it up
to 2 level wherever necessary.
