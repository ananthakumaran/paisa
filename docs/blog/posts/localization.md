---
date: 2024-01-03
comments: true
categories:
  - configuration
  - locale
hide:
  - feedback
---

# Localization

The way numbers are formatted varies across various [countries](https://en.wikipedia.org/wiki/Decimal_separator#Examples_of_use). In
this post, I am going to talk about how to configure localization in
Paisa so you can use the number formatting you are familiar with.

The User Interface localization is controlled by [configuration](../../reference/config.md),
and the Journal file localization is controlled by the commodity
directive.

<!-- more -->

## User Interface

### `locale`

The way the numbers are formatted by the User Interface is controlled
by the [configuration](../../reference/config.md) key called `locale`. By default, this value
is set to `en-IN`. It can be edited via the configuration page. The
value is made of two parts: the first part is the language name, and
the second part is the region name. Refer to Wikipedia for the full list
of supported [language](https://en.wikipedia.org/wiki/List_of_ISO_639_language_codes) and [country](https://en.wikipedia.org/wiki/List_of_ISO_3166_country_codes) codes.

### `default_currency`

Paisa supports multiple currencies, but it needs you to set one of the
currencies as the default currency. The User Interface shows all the
numbers in the default currency.

### `display_precision`

Paisa also allows you to control the number of digits shown after the
decimal point. By default, it's set to zero.


## Journal

In most cases, you don't have to worry about the journal file since
you can enter the value in the correct format and the ledger will parse it
fine. But if you want to follow European number formatting, then you
would have to use [commodity directive](https://ledger-cli.org/doc/ledger3.html#index-commodity-1) to declare the number
formatting. In the example below, we instruct the ledger to treat `,`
as decimal separator and `.` as thousand separator.

```ledger
commodity EUR
	format 1.000,00 EUR

2024/01/01 Salary
    Income:Salary:Acme                    -11.000,00 EUR
    Assets:Checking                        11.000,00 EUR
```
