# CHANGELOG

### 0.6.5 (2024-02-02)

* Add Liabilities > [Credit Card](https://paisa.fyi/reference/credit-cards) page
* Support password protected XLSX file
* Allow user to configure timezone
* Bug fixes

### 0.6.4 (2024-01-22)

* Add checking accounts balance to dashboard
* Improve template management UI
* Improve spinner and page transition
* Bug fixes

### 0.6.3 (2024-01-13)

* Introduce [Sheets](https://paisa.fyi/reference/sheets/): A notepad calculator with access to your ledger
* Remove flat option from cashflow > yearly page
* Dockerimage now installs paisa to /usr/bin
* Improve legends rendering on all pages
* Allow user to cancel pdf password prompt
* Add new warning for missing assets accounts from allocation target
* Support hledger's balance assertion
* Bug fixes

### 0.6.2 (2023-12-23)

* New logo
* Allow goals to be reordered
* Show goals on the dashboard page
* Bug fixes

### 0.6.1 (2023-12-16)

* Add new price provider: [Alpha Vantage](https://paisa.fyi/reference/commodities/#alpha-vantage)
* Make first day of the week configurable
* Support ledger strict mode
* Add user login support, go to `User Accounts` section in configuration page to enable it
* Show notes associated with a transaction/posting
* Bug fixes

### 0.6.0 (2023-12-09)

* Add individual account balance on goals page
* Add [keyboard shortcuts](https://paisa.fyi/reference/editor/) to format/save file on editor page
* Add ability to search posting/transaction by note
* Add option to reverse the order of generated transactions on import page
* Add option to clear price cache
* Bug fixes

### 0.5.9 (2023-11-26)

* Improve postings page
* Add income statement page (Cash Flow > Income Statement)
* Bug fixes

### 0.5.8 (2023-11-18)

* Add ability to specify rate, target date or monthly contribution to
  [savings goal](https://paisa.fyi/reference/goals/savings/)
* Improve price page
* Bug fixes

### 0.5.7 (2023-11-11)

* Add [goals](https://paisa.fyi/reference/goals)
* Remove retirement page (available under goals)
* Bug fixes

#### Breaking Changes :rotating_light:

Retirement page has been moved under goals. If you have used
retirement, you need to setup a new [retirement goal](https://paisa.fyi/reference/goals)

### 0.5.6 (2023-11-04)

* Add support for Income:CapitalGains
* Add option to control display precision
* Add new price provider for gold and silver (IBJA India)
* Add option to disable budget rollover
* Bug fixes

### 0.5.5 (2023-10-07)

* Support account icon customization
* Add beancount ledger client support

### 0.5.4 (2023-10-07)

* Add calendar view to recurring page
* Support [recurring period](https://paisa.fyi/reference/recurring/#period) configuration
* Support European number format
* Bug fixes

### 0.5.3 (2023-09-30)

* Add Docker Image
* Add Linux Application (deb package)
* Move import templates to configuration file
* Bug fixes

#### Breaking Changes :rotating_light:

User's custom import templates used to be stored in Database, which is
a bad idea in hindsight. It's being moved to the configuration
file. With this change, all the data in paisa.db would be transient
and can be deleted and re created from the journal and configuration
files without any data loss.

If you have custom template, take a backup before you upgrade and add
it again via new version. If you have already upgraded, you can still
get the data directly from the db file using the following query
`sqlite3 paisa.db "select * from templates";`

### 0.5.2 (2023-09-22)

* Add Desktop app
* Support password protected PDF on import page
* Bug fixes

#### Breaking Changes :rotating_light:

* The structure of price code configuration has been updated to make
  it easier to add more price provider in the future. In addition to
  the code, the provider name also has to be added. Refer the
  [config](https://paisa.fyi/reference/config/) documentation for more details

```diff
     type: mutualfund
-    code: 122639
+    price:
+      provider: in-mfapi
+      code: 122639
     harvest: 365
```


### 0.5.0 (2023-09-16)

* Add config page
* Embed ledger binary inside paisa
* Bug fixes

### 0.4.9 (2023-09-09)

* Add [search query](https://paisa.fyi/reference/bulk-edit/#search) support in transaction page
* Spends at child accounts level would be included in the budget of
  parent account.
* Fix the windows build, which was broken by the recent changes to
  ledger import
* Bug fixes

### 0.4.8 (2023-09-01)

* Add budget
* Add hierarchial cash flow
* Switch from float64 to decimal
* Bug fixes


### 0.4.7 (2023-08-19)

* Add dark mode
* Add bulk transaction editor
