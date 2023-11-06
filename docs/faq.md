---
description: "Frequently asked Questions about Paisa: personal finance manager"
hide:
  - navigation
---

# FAQs

[__I already use ledger/hledger/beancount. How do I get started?__](#existing-user-getting-started)

Go through the [installation](./getting-started/installation.md) docs and get the app working. Go to
[configuration](./reference/config.md) page and update `journal_path`, `ledger_cli`,
`default_currency` and `locale`. At this point you should be able to
view your journal data. Read the [accounts](./reference/accounts.md) docs to understand the
account naming conventions followed by paisa.

[__How do I get started?__](#getting-started)

You have installed paisa and gone through the demo and you like it,
but you don't know what to do next. There is no way you would sit and
type all the transactions you have made in the last decade or so.

Paisa is an accounting application and you can focus on things that
matters to you the most. For example, if you are not interested in
tracking expenses, you can just have a single expense transaction per
month. You decide the **granularity** at which you want to record the
transactions. Just recording a few transactions per month like your
salary, monthly expense and investments will go a long way and will
give you pretty good picture of your finances.

[__Mac Desktop app fails with executable file not found in $PATH
error__](#exec-not-found)

Mac Desktop apps are usually started with different PATH than what you
would normally get in a Terminal app. Paisa does a fallback search in
the following folders before throwing the error. Check if your
executable binary is in any of the folder listed below, if not, create
a symbolic link from one of the folders to your executable binary.

```
/bin
/usr/bin
/usr/local/bin
/sbin
/usr/sbin
/opt/homebrew/bin
```
