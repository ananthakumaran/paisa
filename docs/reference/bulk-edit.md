# Bulk Edit

Paisa provides bulk transaction editor to search and modify multiple
transactions at once. The interface is made of two parts:

1) Search input box allows you to narrow down the transactions you are
interested in making changes

2) Bulk Edit form allows you to make changes to the narrowed down set
of transactions.

## Search

#### Plain

Paisa provides a powerful search query interface. Let's start with a
few example queries.

```
Expenses:Utilities:Electricity
```

This will search for all transactions that have a posting with account
named Expenses:Utilities:Electricity. By default, the search is case
insensitive and will do a substring match. So, `Expenses:Utilities`
will match Expenses:Utilities:Electricity account as well as. If you
want to search an account name which has special characters like space
in it, you can use double quotes to enclose it like
`"Expenses:Utilities:Hair Cut"`.

You can also search on transaction date. For example, if you want to
show all the transactions made on 1st Jan 2023, just type
`[2023-01-01]`. If you want to see all made on that month, just leave
out the day part `[2023-01]`. You can do the same with year, `[2023]`
will show all the transactions made in 2023.

There is experimental support for natural language date. You can do
queries like `[last month]`, `[last year]`, `[this month]`, `[last
week]`, `[jan 2023]`, etc.

Let's say you want to search by amount. You can do that by typing
`42`, it will show all the transactions that have a posting with that
amount.

if you want to search a exact Account, you can do that using Regular
Expression. Just type `/^Assets:Equity:APPLE$/`, you can also do case
insensitive search by using the modifier `i` like
`/^Assets:Equity:APPLE$/i`.

#### Property

You can also search based on properties like account, commodity,
amount, total, filename, payee and date.

```
account = Expenses:Utilities:Electricity
payee =~ /uber/i
payee = "Advance Tax"
total > 5000
commodity = GOLD
date >= [2023-01-01]
filename = creditcard/2023/jan.ledger
```

The general format is `property operator value`. The property can be
any of the following:

- account (posting account)
- commodity (posting commodity)
- amount (posting amount)
- total (transaction total)
- filename (name of the file the transaction is in)
- payee (transaction payee)
- date (transaction date)

The operator can be any of the following:

- = (equal)
- =~ (regular expression match)
- < (less than)
- <= (less than or equal)
- \> (greater than)
- \>= (greater than or equal)

Not all the combinations of property, operator and value would work,
if in doubt, just try it out, the UI will show you an error if the
query is not valid.

In fact, in the previous format we saw, the property and operator is
not specified and a default set is chosen based on the value type. For
example, `42` will be treated as `amount = 42`, `Expenses:Utilities`
will be treated as `account = Expenses:Utilities`,
`/Expenses:Utilities/i` will be treated as `account =~
/Expenses:Utilities/i`, `[2023-01]` will be treated as `date =
[2023-01]`.

#### Conditional

You can combine multiple property based queries using `AND` and `OR`,
you can negate them using `NOT`

```
account = Expenses:Utilities AND payee =~ /uber/i
commodity = GOLD OR total > 5000
date >= [2023-01-01] AND date < [2023-04-01]
account = Expenses:Utilities AND payee =~ /uber/i AND (total > 5000 OR total < 1000)
total < 5000 AND NOT account = Expenses:Utilities
[last year] AND (payee =~ /swiggy/i OR payee =~ /phonepe/i)
```

If you leave out the conditional operator, it will be treated as
`AND`. Both of the below queries are the same

```
account = Expenses:Utilities AND payee =~ /uber/i
account = Expenses:Utilities payee =~ /uber/i
```


## Bulk Edit Form

Currently bulk edit form supports only account rename feature. More
will be added later. The preview button allows you to see the changes
before you save them. It will show a side by side diff of the changes.
