---
description: "Paisa supports all the top 3 plain text account clients namely ledger, hledger and beancount"
---

# Ledger CLI

Paisa is compatible with [Ledger](https://www.ledger-cli.org), [hledger](https://hledger.org/) and
[Beancount](https://beancount.github.io/). By default Paisa uses **hledger**. To choose another
engine, set the existing `ledger_cli` option in [paisa.yaml](./config.md) to `ledger`, `hledger`, or
`beancount`.

```yaml
# OPTIONAL, DEFAULT: hledger, ENUM: ledger, hledger, beancount
ledger_cli: hledger
```

Paisa's Linux packages and Docker image include hledger. Other builds need hledger on `PATH` unless
they bundle it. Paisa continues to support Ledger and Beancount; install their executables and select
them with `ledger_cli` when needed. The Ledger binary remains embedded in the prebuilt Paisa binary.

## Unavailable Features

Some of the features that are available in Paisa are not supported
by **hledger** and **beancount**

### hledger

* [Recurring](./recurring.md) (partial) - hledger supports automated transactions,
  but it doesn't allow to add **only** metadata to transaction. So
  some of the examples in the recurring section will not work with
  hledger. It's possible to add metadata to transactions manually.


### beancount

Both Beancount 2 and 3 use `ledger_cli: beancount`. Beancount 2 works with its
bundled commands. With Beancount 3, install the separate
[`beanquery`](https://github.com/beancount/beanquery) package and make sure
`bean-check` and `bean-query` are on your `PATH`.

* [Budget](./budget.md) - budget is based on periodic transactions that is no
    supported by beancount.

* [Recurring](./recurring.md) (partial) - beancount doesn't support automated
  transactions. It's possible to add metadata to transactions
  manually. Use lowercase for tag name.

* Search filter in Transactions and Postings pages doesn't support note
  [property](./bulk-edit.md#property).
