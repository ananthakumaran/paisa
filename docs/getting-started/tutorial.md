# Tutorial

This tutorial will introduce all the concepts necessary to get
started. **Paisa** builds on top of the **[ledger](https://ledger-cli.org/)**[^1], a command
line tool that follows the principles of [plain text accounting](https://plaintextaccounting.org/).
**ledger** primarily focuses on command line users and doesn't provide
any graphical user interface. Paisa aims to create a low-friction
graphical user interface on top of **ledger**, thereby making it
accessible to a wider range of users.

As an end user, you should be familiar with the [terms and
concepts](https://github.com/ledger/ledger/blob/master/doc/GLOSSARY.md) used by **ledger**, which we will cover below. Paisa
comes with an embedded **ledger** and you are not required to use
**ledger** via command line unless you want to.

!!! tip

    Even though the tutorial focuses on Indian users, Paisa is
    capable of handling any currency. You can change the default
    currency, locale and financial year starting month etc. Check the
    [configuration](../reference/config.md) reference for more details.


## :fontawesome-regular-file-lines: Journal

A journal file captures all your financial transactions. A transaction
may represent a mutual fund purchase, retirement contribution, grocery
purchase and so on. Paisa creates a journal named `main.ledger`, Let's
add our first transaction there. To open the editor, go to `Ledger`
:material-chevron-right: `Editor`

```ledger
2022/01/01/*(1)!*/ Salary/*(2)!*/
    Income:Salary:Acme/*(3)!*/   -100,000 INR/*(6)!*/
    Assets:Checking/*(4)!*/    /*(5)!*/100,000 INR
```

1. Transaction `Date`
2. Transaction description, also called as `Payee`
3. Debit `Account`
4. Credit `Account`
5. `Amount`
6. `Currency`

**ledger** follows the double-entry accounting system. In simple terms, it
tracks the movement of money from debit account to credit
account. Here `#!ledger Income:Salary:Acme` is the debit account and
`#!ledger Assets:Checking` is the credit account. The date at which the
transaction took place and a description of the transaction is written
in the first line followed by the list of credit or debit
entry. Account [naming conventions](../reference/accounts.md) are explained later. The `:` in the account name
represents hierarchy.

```ledger
2022/01/01 Salary
    Income:Salary:Acme      -100,000 INR
    Assets:Checking          100,000 INR

2022/02/01 Salary
    Income:Salary:Acme      -100,000 INR
    Assets:Checking          100,000 INR

2022/03/01 Salary
    Income:Salary:Acme      -100,000 INR
    Assets:Checking          100,000 INR
```

Let's add few more transactions. As you edit your journal file, the
balance of the journal will be shown on the right hand side.

```
         300,000 INR  Assets:Checking
        -300,000 INR  Income:Salary:Acme
--------------------
                   0
```

You would notice zero balance and a checking account with 3 lakhs and
an income account with -3 lakhs. Double-entry accounting will always
results in 0 balance since you have to always enter both the credit
and debit side.


Let's say your company deducts `#!ledger 12,000 INR` and contributes it to EPF,
we could represent it as follows

```ledger
2022/01/01 Salary
    Income:Salary:Acme    -100,000 INR
    Assets:Checking         88,000 INR
    Assets:Debt:EPF         12,000 INR

2022/02/01 Salary
    Income:Salary:Acme    -100,000 INR
    Assets:Checking         88,000 INR
    Assets:Debt:EPF         12,000 INR

2022/03/01 Salary
    Income:Salary:Acme    -100,000 INR
    Assets:Checking         88,000 INR
    Assets:Debt:EPF         12,000 INR
```

You can now see the use of `:` hierarchy in the account name.

```
         300,000 INR  Assets
         264,000 INR    Checking
          36,000 INR    Debt:EPF
        -300,000 INR  Income:Salary:Acme
--------------------
                   0
```

## :material-gold: Commodity

So far we have only dealt with INR. **ledger** can handle commodity as
well. Let's say you are also investing `#!ledger 10,000 INR` in UTI Nifty Index
Fund and `#!ledger 10,000 INR` in ICICI Nifty Next 50 Index Fund every
month.

```ledger
2018/01/01 Investment
    Assets:Checking               -20,000 INR
    Assets:Equity:NIFTY        148.0865 NIFTY @ 67.5281 INR
    Assets:Equity:NIFTY_JR  358.6659 NIFTY_JR @ 27.8811 INR

2018/02/01 Investment
    Assets:Checking               -20,000 INR
    Assets:Equity:NIFTY        140.2870 NIFTY @ 71.2824 INR
    Assets:Equity:NIFTY_JR  363.2242 NIFTY_JR @ 27.5312 INR

2018/03/01 Investment
    Assets:Checking               -20,000 INR
    Assets:Equity:NIFTY        147.5908 NIFTY @ 67.7549 INR
    Assets:Equity:NIFTY_JR  378.4323 NIFTY_JR @ 26.4248 INR
```

Let's consider `#!ledger 148.0865 NIFTY @ 67.5281 INR`. Here `#!ledger
NIFTY` is the name of the commodity and we have bought `#!ledger
148.0865` units at `#!ledger 67.5281 INR` per unit.

Paisa has support for fetching commodity price history from few
[providers](../reference/commodities.md). Go to `Configuration` page and expand the `Commodities`
section. You can click the :fontawesome-solid-circle-plus: icon to
add a new one. Edit the name to `#!ledger NIFTY`. Click the
:fontawesome-solid-pen-to-square: icon near Price section and select
the price provider details. Once done, save the configuration and click the
`Update Prices` from the top right hand side menu. If you had done
everything correctly, you would see the latest price of the commodity
under `Assets` :material-chevron-right: `Balance`

## :fontawesome-solid-hand-holding-dollar: Interest

There are many instruments like EPF, FD, etc that pay interest at
regular intervals. We can treat it as just another transaction. Any
income account that has a prefix `#!ledger Income:Interest:` can be
used as the debit account. It's not mandatory to specify the amount at
bot side. If you leave one side, **ledger** will deduct it.

```ledger
2022/03/31 EPF Interest
    Income:Interest:EPF     -5,000 INR
    Assets:Debt:EPF
```

```
           5,000 INR  Assets:Debt:EPF
          -5,000 INR  Income:Interest:EPF
--------------------
                   0
```

## Configuration

All the configuration related to paisa is stored in a yaml file named
`paisa.yaml`. The configuration can be edited via the web interface. The
sequence in which it look for the file is described below

1. `PAISA_CONFIG` environment variable
1. via --config flag
1. Current working directory
1. `paisa/paisa.yaml` file inside User Documents folder.

If it can't find the configuration file, it will create a default
configuration file named `paisa/paisa.yaml` inside User Documents
folder. The default configuration is tuned for Indians, users from
other countries would have to change the `default_currency` and
`locale`. Check the [configuration](../reference/config.md) reference for details.

## Update

Paisa fetches the latest price of the commodities only when
*update* command is used. Make sure to run `paisa update` command
after you make any changes to your journal file or you want to fetch
the latest value of the commodities. The update can be performed from
the UI as well via the dropdown in the top right hand side corner.

[^1]: [hledger](hled) is also supported, refer to the [configuration](../reference/config.md)
    to see how to modify the `ledger_cli`.
