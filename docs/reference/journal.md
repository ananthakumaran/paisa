---
description: "The journal format used by Paisa to capture transactions, metadata, etc."
---

# Journal

All your transactions are stored in plain text files called
journal. You can organize your transactions into multiple journal
files. The [journal_path](./config.md) configuration refers your main journal
file. The main journal file can refer other journal files using
[include](https://ledger-cli.org/doc/ledger3.html#index-include) directive. The `#!ledger include` directive supports
wildcards `*` as well. Transactions are sourced **only** from the main
journal file and other journal files included from the main journal
file.


```ledger
include investments.ledger
include expenses/*.ledger
```

## Editor

Paisa comes with a journal editor. It allows you to edit all the files
with the same file extension as your main journal and in the same or
sub directories as your main journal.


## Backup

Paisa tries its best to keep your journal safe. It creates a backup of
your journal file every time you save it. The backup file is created
with the same name as your journal file with a `.backup.{timestamp}`
extension. You can revert back to old versions of your journal file
via the editor. You can also delete the backup files from the editor.

!!! warning

    It is recommended to keep your journal files and `paisa.yaml`
    under version control or other backup mechanism. Paisa's backup
    mechanism is not a replacement for a proper backup. You can ignore
    `paisa.db` file from your version control system, the data in db
    file can be recreated from your journal files.

## Syntax

The journal syntax of the features you use normally along with paisa
is documented here. Refer [ledger](https://ledger-cli.org/doc/ledger3.html#Journal-Format) documentation for more details.

##### Transaction

```ledger
2022/01/01 Salary
    Income:Salary:Acme      -100,000 INR
    Assets:Checking          100,000 INR
```

A `transaction` should start with a `date` followed by
`description`. Following that you can have 2 or more `postings`. The
posting line should have at least 2 leading spaces. The `account` name
and the `amount` should be separated by at least 2 spaces.

##### Commodity

```ledger
2022/01/07 Investment
    Assets:Checking         -20,000 INR
    Assets:Equity:NIFTY   168.690 NIFTY @ 118.56 INR
```

`commodity` cost can be specified using the `@` syntax. Here `118.56`
is the per unit cost and `168.690` is the quantity you have bought.

##### Comment

```ledger
2023/07/01 Rent
    ; This is a transaction comment
    Expenses:Rent             15,000 INR
    Assets:Checking   ; This is a posting comment
```

Any text after `;` is treated as a comment. Comment can be at whole
transaction level or individual posting level. Comment is also referred
as **note** in many places, both are same.

##### Tags

```ledger
2023/07/01 Rent
    ; Recurring: Rent
    Expenses:Rent             15,000 INR
    Assets:Checking
```

Transactions can be tagged with extra metadata called tags. Tag has
two parts: tag name and value. In the above example, `Recurring` is
the name and `Rent` is the value. Tag should be inside comment.

##### Include

```ledger
include investments.ledger
include expenses/*.ledger
```

Include directive can be used to include other journal files. It
supports wildcards `*`.
