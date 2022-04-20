# Accounts

Even though **ledger** doesn't have any Account naming convention,
**paisa** makes lot of assumptions and expects you to follow the same
naming convention.

### Asset

All your assets should go under `Asset:`. The level of granularity is
up to you. The recommended convention is to use
`Asset:{instrument_type}:{instrument_name}`. The instrument type may
be `Cash`, `Equity`, `Debt`, etc. The instrument name may be the name of
the fund, stock, etc

### Income

All your income should come from `Income:`.

* `Income:Salary` - salary debit account
* `Income:Interest:{name}` - interest debit account
