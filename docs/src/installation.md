# Installation

* Follow the installation instruction in [ledger](https://www.ledger-cli.org/download.html) site and install
  **ledger**. Windows users can install ledger via [chocolatey](https://community.chocolatey.org/packages/ledger).
* Download the latest prebuilt [binary](https://github.com/ananthakumaran/paisa/releases/latest)
* Once downloaded, you can perform the following steps to install
it.
```shell
❯ mv paisa-* paisa
❯ chmod u+x paisa
❯ xattr -dr com.apple.quarantine paisa # applicable only on Mac
❯ mv paisa /usr/local/bin
```

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
INFO Fetching commodities price history
INFO Fetching commodity NIFTY
INFO Fetching commodity NIFTY_JR
❯ paisa serve
INFO Using config file: /home/john/finance/paisa.yaml
INFO Listening on 7500
```
Go to [http://localhost:7500](http://localhost:7500)
