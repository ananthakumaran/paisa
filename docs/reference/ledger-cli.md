---
description: "Paisa supports all the top 3 plain text account clients namely ledger, hledger and beancount"
---

# Ledger CLI

Paisa is compatible with [ledger](https://www.ledger-cli.org), [hledger](https://hledger.org/) and
[beancount](https://beancount.github.io/). By default paisa will try to use **ledger**, if you
prefer to use **hledger** or **beancount** instead, change the
`ledger_cli` value in [paisa.yaml](./config.md)

```yaml
# OPTIONAL, DEFAULT: ledger, ENUM: ledger, hledger, beancount
ledger_cli: hledger
```

Paisa ships with ledger binary. If you use hledger or beancount, make
sure that the binaries are installed.
