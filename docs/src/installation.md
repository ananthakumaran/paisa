# Installation

* Follow the installation instruction in [ledger](https://www.ledger-cli.org/download.html) site and install **ledger**
* Install paisa

## Quick Start

Run the following commands to checkout **paisa** features. Read the
tutorial to understand the details.

```shell
❯ mkdir finance
❯ cd finance
❯ paisa init
INFO Generating config file: /home/john/finance/paisa.yaml
INFO Generating journal file: /home/john/finance/personal.ledger
❯ paisa update
INFO Using config file: /home/john/finance/paisa.yaml
INFO Syncing transactions from journal
INFO Fetching mutual fund history
INFO Fetching commodity NIFTY
INFO Fetching commodity NIFTY_JR
❯ paisa serve
INFO Using config file: /home/john/finance/paisa.yaml
INFO Listening on 7500
```
Go to [http://localhost:7500](http://localhost:7500)
