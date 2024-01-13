---
description: "How to use notepad calculator called sheets in Paisa"
status: new
---

# Sheets

## Introduction

Sheet is a notepad calculator that computes the answer as you type.
It has full access to your ledger and can be used to calculate the
answers for a wide variety of financial questions.

!!! example "Experimental"

    Sheet is an experimental feature. Give it a try and let me know
    how it goes, specially what is missing and what can be improved.
    The syntax and semantics of sheet may change in future releases.

#### Calculator

Sheet can act as a normal calculator. The example below shows how to
calculate the monthly EMI for a home loan.

<div class="split-codeview">
<div class="sheet" markdown>
```{ .sheet .no-copy }
# Home Loan
price = 40,00,000
down_payment = 20% * price
finance_amount = price - down_payment
interest_rate = 8.6%
term = 30

n = term * 12
r = interest_rate / 12
# EMI
monthly_payment = r / (1 - (1 + r) ^ (-n)) * finance_amount
```
</div>
<div class="sheet-result sheet-result-1" markdown>
```{ .text .no-copy }
# Home Loan
40,00,000
8,00,000
32,00,000
0
30

360
0
# EMI
24832
```
</div>
</div>

#### Function

Sheets comes with a set of [built-in functions](#functions). You can also define
your own functions. The example below shows how to define a function

<div class="split-codeview">
<div class="sheet" markdown>

```{ .sheet .no-copy }
# Years to Double
years_to_double(rate) = 72 / rate

years_to_double(2)
years_to_double(4)
years_to_double(6)
years_to_double(8)
years_to_double(10)
years_to_double(12)
years_to_double(14)
```
</div>
<div class="sheet-result sheet-result-2" markdown>
```{ .text .no-copy }
# Years to Double


36
18
12
9
7
6
5
```
</div>
</div>

#### Query

Query is what makes sheets powerful. It allows you to query your
ledger postings and do calculations on them. The example below shows
how to calculate cost basis of your assets so you can report them to
Income Tax department.

<div class="split-codeview">
<div class="sheet" markdown>

```{ .sheet .no-copy }
# Schedule AL
date_query = {date <= [2023-03-31]}
cost_basis(x) = cost(fifo(x AND date_query))
cost_basis_negative(x) = cost(fifo(negate(x AND date_query)))

# Immovable
immovable = cost_basis({account = Assets:House})

# Movable
metal = 0
art = 0
vehicle = 0
bank = cost_basis({account = /^Assets:Checking/})
share = cost_basis({account =~ /^Assets:Equity:.*/ OR
                    account =~ /^Assets:Debt:.*/})
insurance = 0
loan = 0
cash = 0

# Liability
liability = cost_basis_negative({account =~ /^Liabilities:Homeloan/})

# Total
total = immovable + metal + art + vehicle + bank + share + insurance + loan + cash - liability
```
</div>
<div class="sheet-result sheet-result-3" markdown>
```{ .text .no-copy }
# Schedule AL




# Immovable
25,00,000

# Movable
0
0
0
1,21,402
66,98,880

0
0
0

# Liability
6,21,600

# Total
86,98,682
```
</div>
</div>

## Syntax

#### Number

Sheet allows comma as a separator. `%` is a syntax sugar for dividing
the number by 100. So `8%` is same as `0.08`.

```sheet
100,00
100.00
-100
8%
```

#### Operators

Sheet supports the following operators. `^` is the exponentiation operator.

```sheet
1 + 1
1 - 1
1 * 1
1 / 1
1 ^ 2
```

#### Variable

You can define variables using `=` operator. Variables have to be
defined before they are used.

```sheet
x = 1
y = 2

z = x + y
```

#### Function

Sheet comes with a set of [built-in functions](#functions). You can also define your
own functions. Function definition starts with the function name and
then a list of arguments. The body of the function is a single
expression. To call a function, use the function name followed by
arguments.

```sheet
sum(a, b) = a + b

sum(1, 2)
```

#### Query

Query is a first class citizen in sheet. You can think of a query as a
list of postings that match the conditions. The query syntax is same
as the [search syntax](./bulk-edit.md#search) used in postings and transactions pages. It
has to be enclosed inside a `{}`. Query can be treated as any other
value, can be passed as an argument, can be assigned to a variable. In
fact you can combine two queries using `#!sheet AND` and `#!sheet OR`
operator as well.

```sheet
query = {account = Expenses:Utilities AND payee =~ /uber/i}

upto_this_fy = {date < [2024-04-01]}
assets = { account =~ /^Assets:.*/ }
liabilities = { account =~ /^Liabilities:.*/ }

assets_upto_this_fy = upto_this_fy AND assets
liabilities_upto_this_fy = upto_this_fy AND liabilities
```

<details markdown>
  <summary>What is posting?</summary>
```ledger hl_lines="2 3"
2022/01/01 Salary
    Income:Salary:Acme        -100,000 INR
    Assets:Checking            100,000 INR
```
In the above transaction, there are two postings. A transaction consists
of several postings that sum to zero. If you want to know the cost of
your checking account, you would use `#!sheet cost({account = Assets:Checking})`

</details>

#### Comment

Sheet supports single line comments starting with `;` or `//`.

```sheet
// set x
x = 1
; set y
y = 2

z = x + y // add x and y
```



## Functions

You can find the full list of built-in functions here. This list is
very slim as of now, please start a discussion if you want some
functions to be added here.

#### `#!typescript cost(q: Posting[] | Query): number`

Returns the sum of the cost of all the postings.

#### `#!typescript fifo(q: Posting[] | Query): Posting[]`

Returns the list of postings after performing a FIFO adjustment. For
example, assume you have 3 postings

```ledger
2020/01/01 Buy
    Assets:Bank:Checking   100 USD
2020/02/01 Buy
    Assets:Bank:Checking   100 USD
2020/03/01 Sell
    Assets:Bank:Checking   -50 USD
```

after FIFO adjustment, you will get the following postings

```ledger
2020/01/01 Buy
    Assets:Bank:Checking    50 USD
2020/02/01 Buy
    Assets:Bank:Checking   100 USD
```

All your sell postings will be adjusted against the oldest buy
postings and you will have the remaining buy postings.


#### `#!typescript negate(q: Posting[] | Query): Posting[]`

Negates the amount and quantity of all the postings. This would be
useful when you want to do some calculation on liabilities or income
which are negative in nature.
