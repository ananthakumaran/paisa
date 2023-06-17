# Config

```yaml
# Absolute path to your journal file. The main journal file can
# refer other files using `include` as long as all the files are in
# the same directory
# REQUIRED
journal_path: /home/john/finance/personal.ledger

# Absolute path to your database file. The database file will be
# created if it does not exist.
# REQUIRED
db_path: /home/ananthakumaran/work/repos/finance/paisa.db

# The ledger client to use
# OPTIONAL, DEFAULT: ledger, ENUM: ledger, hledger
ledger_cli: hledger

# The default currency to use. NOTE: Paisa doesn't handle multiple
# currencies well.
# OPTIONAL, DEFAULT: INR
default_currency: INR

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
    type: mutualfund
    code: 145552
    harvest: 1095
    tax_category: debt
  - name: NIFTY
    type: mutualfund
    code: 120716
    harvest: 365
    tax_category: equity
  - name: APPLE
    type: stock
    code: AAPL
    harvest: 1095
    tax_category: equity
```
