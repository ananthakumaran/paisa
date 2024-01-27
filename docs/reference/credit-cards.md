---
description: "How to configure credit cards in Paisa"
---

# Credit Cards

Paisa allows you to track your credit card bills and payments. Let's
say you have a liability account called
`Liabilities:CreditCard:Freedom` to track your real world credit card,
you can configure it in Paisa as follows:

```yaml
credit_cards:
    - account: Liabilities:CreditCard:Freedom #(1)!
      credit_limit: 150000 #(2)!
      statement_end_day: 8 #(3)!
      due_day: 20 #(4)!
      network: visa #(5)!
      number: "0007" #(6)!
      expiration_date: "2029-05-01" #(7)!
```

1. Account name
2. Credit limit of the card
3. The day of the month when the statement is generated
4. The day of the month when the payment is due
5. The network of the card
6. The last 4 digits of the card number
7. The expiration date of the card

The above configuration can be done from the `More > Configuration`
page. Expand the `Credit Cards` section and click
:fontawesome-solid-circle-plus: icon to add a new one

Once configured, the credit card will show up on the `Liabilities >
Credit Cards` page. Paisa will automatically calculate and display
various information like the amount due, payment due date, credit
limit utilized, etc.
