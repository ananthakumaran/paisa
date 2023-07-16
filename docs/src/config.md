# Config

**paisa** checks for a config file named `paisa.yaml` in the current
directory. It can also be set via flag `--config` or env variable
`PAISA_CONFIG`.

```yaml
# Absolute path to your journal file. The main journal file can
# refer other files using `include` as long as all the files are in
# the same directory
# REQUIRED
journal_path: /home/john/finance/personal.ledger

# Absolute path to your database file. The database file will be
# created if it does not exist.
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
  # By default average of last 3 year expenses will be used to
  # calculate your yearly expenses. This can be overriden by setting
  # this config to positive value
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
    # differs based on type
    code: 145552
    harvest: 1095
    # Optional, ENUM: equity65, equity35, debt, unlisted_equity
    tax_category: debt
  - name: NIFTY
    type: mutualfund
    code: 120716
    harvest: 365
    tax_category: equity65
  - name: APPLE
    type: stock
    code: AAPL
    harvest: 1095
    tax_category: equity65
```
