---
description: "How to secure Paisa by adding user account"
---

# User Authentication

You can setup username and password to make sure only you can access
the application. Go to `Configuration` page and expand the `User
Accounts` section. You can click the :fontawesome-solid-circle-plus:
icon to add a new username and password. Once you save the
configuration, you will get logged out if you had not logged in
earlier via another account.

It is important to understand that authentication only protects the
application user interface. If someone can access your computer and
they can access the folder where Paisa stores the ledger and database
files, they will be able to view your data.

## Implementation details

Paisa uses [cryptographic hash function](https://en.wikipedia.org/wiki/Cryptographic_hash_function) sha-256 to convert your
password to digest before storing it in the configuration file. This
has few implications. No one can look at the configuration file and
get the password, this includes you as well. If you forget the
password, you can remove the user accounts from [configuration](./config.md)
file to get back access.

!!! warning

    If you run paisa on a server and access it over public internet,
    make sure to use it over **https** and use a **strong password**. If you
    run it over plain http, [man in the
    middle](https://en.wikipedia.org/wiki/Man-in-the-middle_attack)
    attack can be performed to obtain the data including your password.
