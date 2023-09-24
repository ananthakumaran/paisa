# Configuration

All the configuration related to paisa is stored in a yaml file named
`paisa.yaml`. The configuration can be edited via the web
interface. The sequence in which paisa looks for the file is described
below

1. `PAISA_CONFIG` environment variable
1. via `--config` flag
1. Current working directory
1. `paisa/paisa.yaml` file inside User Documents folder.

If it can't find the configuration file, it will create a default
configuration file named `paisa/paisa.yaml` inside User Documents folder. The
default configuration is tuned for Indians, users from other countries
would have to change the `default_currency` and `locale`.

```yaml
# Path to your journal file. It can be absolute or relative to the
# configuration file. The main journal file can refer other files using
# `include` as long as all the files are in the same or sub directory
# REQUIRED
journal_path: /home/john/finance/main.ledger

# Path to your database file. It can be absolute or relative to the
# configuration file. The database file will be created if it does not exist.
# REQUIRED
db_path: /home/john/finance/paisa.db

# The ledger client to use
# OPTIONAL, DEFAULT: ledger, ENUM: ledger, hledger
ledger_cli: hledger

# The default currency to use. NOTE: Paisa tries to convert other
# currencies to default currency, so make sure it's possible to
# convert to default currency by specifying the exchange rate.
#
# OPTIONAL, DEFAULT: INR
default_currency: INR

# The locale used to format numbers. The list of locales supported
# depends on your browser. It's known to work well with en-US and en-IN.
#
# OPTIONAL, DEFAULT: en-IN
locale: en-IN

# First month of the financial year. This can be set to 1 to follow
# January to December.
#
# OPTIONAL, DEFAULT: 4
financial_year_starting_month: 4

## Retirement
retirement:
  # Safe Withdrawal Rate
  # OPTIONAL, DEFAULT: 4
  swr: 2
  # List of expense accounts
  # OPTIONAL, DEFAULT: Expenses:*
  expenses:
    - Expenses:Clothing
    - Expenses:Education
    - Expenses:Entertainment
    - Expenses:Food
    - Expenses:Gift
    - Expenses:Insurance
    - Expenses:Misc
    - Expenses:Restaurant
    - Expenses:Shopping
    - Expenses:Utilities
  # List of accounts where you keep retirement savings
  # OPTIONAL, DEFAULT: Assets:*
  savings:
    - Assets:Equity:*
    - Assets:Debt:*
  # By default, average of last 3 year expenses will be used to
  # calculate your yearly expenses. This can be overridden by setting
  # this configuration to positive value
  # OPTIONAL, DEFAULT: 0
  yearly_expenses: 0

## Schedule AL
# OPTIONAL, DEFAULT: []
schedule_al:
  # Code
  # REQUIRED, ENUM: immovable, metal, art, vehicle, bank, share,
  # insurance, loan, cash, liability
  - code: metal
    accounts:
      - Assets:Gold
  - code: bank
    accounts:
      - Assets:Checking
      - Assets:Debt:Cash:FD
  - code: share
    accounts:
      - Assets:Equity:*
  - code: insurance
    accounts:
      - Assets:Debt:Insurance

## Allocation Target
# OPTIONAL, DEFAULT: []
allocation_targets:
  - name: Debt
    target: 30
    accounts:
      - Assets:Debt:*
      - Assets:Checking
  - name: Equity
    target: 60
    accounts:
      - Assets:Equity:*
  - name: Equity Foreign
    target: 20
    accounts:
      - Assets:Equity:NASDAQ
  - name: Equity Index
    target: 20
    accounts:
      - Assets:Equity:NIFTY
  - name: Equity Active
    target: 20
    accounts:
      - Assets:Equity:PPFAS
  - name: Others
    target: 10
    accounts:
      - Assets:Gold
      - Assets:RealEstate

## Commodities
# OPTIONAL, DEFAULT: []
commodities:
  - name: NASDAQ
    # Required, ENUM: mutualfund, stock, nps, unknown
    type: mutualfund
    price:
      # Required, ENUM: in-mfapi, com-yahoo, com-purifiedbytes-nps
      provider: in-mfapi
      # differs based on provider
      code: 145552
    harvest: 1095
    # Optional, ENUM: equity65, equity35, debt, unlisted_equity
    tax_category: debt
  - name: NIFTY
    type: mutualfund
    price:
      provider: in-mfapi
      code: 120716
    harvest: 365
    tax_category: equity65
  - name: APPLE
    type: stock
    price:
      provider: com-yahoo
      code: AAPL
    harvest: 1095
    tax_category: equity65
```
