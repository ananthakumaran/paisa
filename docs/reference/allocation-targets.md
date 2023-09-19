# Allocation Targets

Paisa allows you to set a allocation target for a group of
accounts. The allocation page shows how far your current allocation is
from the allocation target. For example, to keep a 40:60 split between
debt and equity, use the following configuration. The account name can
have `*` which matches any characters

```yaml
allocation_targets:
  - name: Debt
    target: 40
    accounts:
      - Assets:Debt:*
  - name: Equity
    target: 60
    accounts:
      - Assets:Equity:*
```
