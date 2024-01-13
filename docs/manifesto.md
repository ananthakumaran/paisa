---
description: "Paisa's Manifesto"
hide:
  - navigation
---

# Manifesto

## 1. Data Ownership :fontawesome-solid-key:

User owns the data. It should be possible to migrate all the data to
another app or service. Paisa will strive to make this as easy as
possible by choosing open standards and formats.

> [Ledger](https://ledger-cli.org/) text format is used to store all the transaction data.

## 2. Privacy :simple-gnuprivacyguard:

User's data is private. Paisa will not collect or send any data to any
server[^1]. Paisa will not use any third party analytics or tracking on
the app[^2].

> Paisa's source code is open source and could be audited by anyone.

## 3. Longevity :octicons-infinity-24:

The app should be available for a long time. It takes a lot of effort
to collect and maintain the transaction data. The app should not just
disappear one day. Paisa will strive to avoid unnecessary dependencies
and build a self contained app.

> Paisa is licensed under AGPL v3, which helps with this issue to some
> extent. I have yet to figure out a solution to make the app
> development and maintenance process sustainable in the long
> term. But rest assured, any decision that would be made related to
> this will not override the first two points.


[^1]: Paisa fetches commodity price information from third party
    servers. Since Paisa will send the commodity identifier to the
    server, the third party server might be able to connect the
    commodity list with the user's IP address. This is an opt-in
    feature, as you have to explicitly configure Paisa to fetch
    price. VPN could be used if you want to avoid this.

[^2]: This doesn't include the website paisa.fyi, which is hosted on
    third party servers and has consent based analytics.

