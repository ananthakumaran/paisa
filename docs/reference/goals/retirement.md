---
description: "How to configure your retirement goal in Paisa"
---

# Retirement

Paisa will help you plan your retirement and track your progress. The
first part is figuring out what should be your retirement corpus. This
will be your target. Instead of specifying the amount explicitly, you
can specify your expected yearly expenses and the safe withdrawal
rate.

```yaml
goals:
  retirement:
    - name: Retirement
      icon: mdi:palm-tree
      swr: 3.3
      yearly_expenses: 1100000
```

If you use paisa to track expenses, instead of specifying the
`yearly_expenses`, you can specify the list of accounts. Paisa will
take the average of the last 3 year expenses

```yaml
goals:
  retirement:
    - name: Retirement
      icon: mdi:palm-tree
      swr: 2
      expenses:
        - Expenses:Entertainment
        - Expenses:Gift
        - Expenses:Insurance
        - Expenses:Misc
        - Expenses:Shopping
        - Expenses:Utilities
```

Now that the target is specified, you need to specify the list of
accounts where you keep your retirement savings.

```yaml
goals:
  retirement:
    - name: Retirement
      icon: mdi:palm-tree
      swr: 2
      expenses:
        - Expenses:Entertainment
        - Expenses:Gift
        - Expenses:Insurance
        - Expenses:Misc
        - Expenses:Shopping
        - Expenses:Utilities
      savings:
        - Assets:Equity:*
        - Assets:Debt:*
```
