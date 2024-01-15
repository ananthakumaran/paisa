---
hide:
  - navigation
  - toc
  - feedback
---

<div class="hero" markdown>
# Paisa
<p>Personal Finance Manager</p>
</div>

<div class="home" markdown>
=== ":material-currency-inr: INR"

    ```ledger
    2022/01/01 Salary
        Income:Salary:Acme     -100,000 INR
        Assets:Checking         100,000 INR

    2022/01/03 Rent
        Assets:Checking         -20,000 INR
        Expenses:Rent

    2022/01/07 Investment
        Assets:Checking         -20,000 INR
        Assets:Equity:NIFTY   168.690 NIFTY @ 118.56 INR
    ```

=== ":material-currency-usd: USD"

    ```ledger
    2022/01/01 Salary
        Income:Salary:Acme      $-5,000
        Assets:Checking          $5,000

    2022/01/03 Rent
        Assets:Checking         $-2,000
        Expenses:Rent

    2022/01/07 Investment
        Assets:Checking         $-1,000
        Assets:Equity:AAPL   6.452 AAPL @ $154.97
    ```

=== ":fontawesome-solid-euro-sign: EURO"

    ```ledger
    commodity €
        format €1.000,00

    commodity AAPL
        format 1.000,00 AAPL

    2022/01/01 Salary
        Income:Salary:Acme      €-5.000
        Assets:Checking          €5.000

    2022/01/03 Rent
        Assets:Checking         €-2.000
        Expenses:Rent

    2022/01/07 Investment
        Assets:Checking      €-1.000,02
        Assets:Equity:AAPL   6,453 AAPL @ €154,97
    ```


<p style="text-align: center; margin-bottom: 3rem">
  <a class="md-button md-button--primary" style="margin-right: 50px;" href="/getting-started/installation/">Install</a>
  <a class="md-button md-button--primary" href="https://demo.paisa.fyi">Demo</a>
</p>

<div class="features-container" markdown>
<div class="features" markdown>
- :fontawesome-regular-file-lines: Builds on top of the **[ledger](https://www.ledger-cli.org/)** double entry accounting tool.
- :octicons-shield-lock-16: Your financial **data** never leaves your system.
- :simple-git: The journal and configuration information are stored in **plain text** files
  that can be easily version controlled. You can collaborate with
  others by giving access to the files.
- :octicons-graph-24: Track the latest market **price** of your Mutual Fund, NPS Fund
  and Stock holdings.
- :fontawesome-regular-credit-card: Track and **budget** your **Expenses**.
- :material-microsoft-excel: **Convert** CSV, Excel and PDF files to Ledger journal.
- :fontawesome-solid-calculator: **Calculate** your taxes, emis, etc using **[sheets](./reference/sheets.md)**.
- :octicons-goal-16: Track your **goals**.
- :material-timer-sync: View your **recurring** transactions.
- :material-beach: Plan your **retirement**.
- :material-chart-bar: And many more **visualizations** to help you make any financial
  decisions.
</div>

<div class="thumbnail-container app-frame win dark" data-title="Paisa">
  <div class="thumbnail">
    <iframe src="https://demo1.paisa.fyi" frameborder="0" scrolling="no"></iframe>
  </div>
</div>
</div>
</div>


<div class="feature-card-container" markdown>
<div class="feature-card" markdown>
<div class="feature-card-left feature-card-icon feature-card-icon-logo" markdown>
:fontawesome-regular-file-lines:{ .feature-card-icon }
</div>
<div class="feature-card-right" markdown>
# Plain Text

All your financial data is stored in plain text files. Of course, we
will not ask you to draw a picture on blank canvas. We provide
enough guard rails to make data entry easy and error free. The editor
comes with syntax highlighting, auto completion, error checking and
auto formatting.
</div>
</div>


<div class="feature-card" markdown>
<div class="feature-card-left feature-card-icon feature-card-icon-equity" markdown>
:octicons-shield-lock-16:{ .feature-card-icon }
</div>

<div class="feature-card-right" markdown>
# Privacy

Selling your data is not the indirect goal of the app. The app will
never collect or send any data to any server. All your data is stored
in your system. Checkout the [manifesto](./manifesto.md) for more.
</div>
</div>

<div class="feature-card" markdown>
<div class="feature-card-left feature-card-icon feature-card-icon-asset" markdown>
:octicons-graph-24:{ .feature-card-icon }
</div>
<div class="feature-card-right" markdown>
# Price Tracking

Paisa supports various [price](./reference/commodities.md) data providers, so it can keep track of
the latest price of all your assets. It also allows the user to enter
the price manually, so you can use it to revalue your assets like
house, car, land, etc.
</div>
</div>


<div class="feature-card" markdown>
<div class="feature-card-left feature-card-icon feature-card-icon-expense" markdown>
:fontawesome-regular-credit-card:{ .feature-card-icon }
</div>

<div class="feature-card-right" markdown>
# Budget and Expenses

Paisa allows you to track your expenses at the granularity of your
choice. You can customize the categories and subcategories to suit
your needs, even icon is customizable. If you are tight on money,
don't worry, paisa got you covered. You can set a [budget](./reference/budget.md) for each
category and make sure you don't overspend.

</div>
</div>

<div class="feature-card" markdown>
<div class="feature-card-left feature-card-icon feature-card-icon-income" markdown>
:material-microsoft-excel:{ .feature-card-icon }
</div>
<div class="feature-card-right" markdown>
# Data Import

You don't have to sit and manually enter all your transactions. You
can get the account statements from your bank or credit card provider
and import them into paisa. Paisa [import](./reference/import.md) system is flexible
enough to handle most of the formats out there in the wild. Once you
setup the import template, it will hardly take 5 minutes to import and
categorize all your transactions each month.
</div>

</div>


<div class="feature-card" markdown>
<div class="feature-card-left feature-card-icon feature-card-icon-liability" markdown>
:fontawesome-solid-calculator:{ .feature-card-icon }
</div>

<div class="feature-card-right" markdown>
# Sheets

Ever wanted to calculate your taxes, emis, etc? Paisa comes with a
notepad calculator called [sheet](./reference/sheets.md) that can do all the calculations
with the data from your ledger. The sheets are live and interactive,
and will update automatically when you change the data in your ledger.

</div>
</div>

<div class="feature-card" markdown>
<div class="feature-card-left feature-card-icon feature-card-icon-logo" markdown>
:octicons-goal-16:{ .feature-card-icon }
</div>
<div class="feature-card-right" markdown>
# Goals

Want to know how much you need to save for your next vacation? or how
long it will take to buy a new car? Paisa can help you with planning
and tracking your [goals](./reference/goals/index.md). If you know how much you can afford to
save each month, paisa can tell you when you can achieve your goal. On
the other hand, if you have a deadline, paisa can tell you how much
you need to save each month to achieve your goal.
</div>

</div>

<div class="feature-card" markdown>
<div class="feature-card-left feature-card-icon feature-card-icon-expense" markdown>
:material-timer-sync:{ .feature-card-icon }
</div>

<div class="feature-card-right" markdown>
# Bills

Paisa can help you track your [recurring](./reference/recurring.md) transactions like rent, emi,
credit card bills, etc. The calendar view will show you all the bills
that are due in the current month. You can set the recurring period of
the transaction, which is flexible enough to handle any kind of
schedule like weekly, monthly, quarterly, last day of the month, last
day of the quarter, etc.

</div>
</div>

<p></p>
</div>
