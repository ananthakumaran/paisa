---
description: "How to configure your savings goal in Paisa"
---

# Savings

Savings represents a general financial objective that you want to
achieve, like buying a car, or a house, or a vacation.

```yaml
goals:
  savings:
    - name: House
      target: 5000000
      icon: fluent-emoji-high-contrast:house-with-garden
      accounts:
        - Assets:Equity:*
        - Assets:Debt:*
```

Specify the target amount, and the accounts that you keep the money
in.

### Forecast

When you save towards an objective, you will have a target date in
mind. In the below config, you are specifying the target date and
rate. Paisa will calculate the monthly contribution required to
achieve the goal.

```yaml hl_lines="5-6"
goals:
  savings:
    - name: House
      target: 5000000
      target_date: 2030-05-01
      rate: 10
      icon: fluent-emoji-high-contrast:house-with-garden
      accounts:
        - Assets:Equity:*
        - Assets:Debt:*
```

If on the other hand, you know how much you can afford to save every
month, but want to know when the goal will be achieved, you can
specify the payment per period and the rate. Paisa will calculate the
target date.

```yaml hl_lines="5-6"
goals:
  savings:
    - name: House
      target: 5000000
      payment_per_period: 50000
      rate: 10
      icon: fluent-emoji-high-contrast:house-with-garden
      accounts:
        - Assets:Equity:*
        - Assets:Debt:*
```

### Math

This section will give some rough idea of how the calculations are
done. You can skip this section if you are not interested in the math.

<math xmlns="http://www.w3.org/1998/Math/MathML" display="block" class="tml-display" style="display:block math;">
  <semantics>
    <mrow>
      <mi>F</mi>
      <mi>V</mi>
      <mo>+</mo>
      <mi>P</mi>
      <mi>V</mi>
      <mo>×</mo>
      <mo form="prefix" stretchy="false">(</mo>
      <mn>1</mn>
      <mo>+</mo>
      <mi>R</mi>
      <mi>A</mi>
      <mi>T</mi>
      <mi>E</mi>
      <msup>
        <mo form="postfix" stretchy="false">)</mo>
        <mrow>
          <mi>N</mi>
          <mi>P</mi>
          <mi>E</mi>
          <mi>R</mi>
        </mrow>
      </msup>
      <mo>+</mo>
      <mi>P</mi>
      <mi>M</mi>
      <mi>T</mi>
      <mo>×</mo>
      <mfrac>
        <mrow>
          <mo form="prefix" stretchy="false">(</mo>
          <mn>1</mn>
          <mo>+</mo>
          <mi>R</mi>
          <mi>A</mi>
          <mi>T</mi>
          <mi>E</mi>
          <msup>
            <mo form="postfix" stretchy="false">)</mo>
            <mrow>
              <mi>N</mi>
              <mi>P</mi>
              <mi>E</mi>
              <mi>R</mi>
            </mrow>
          </msup>
          <mo>−</mo>
          <mn>1</mn>
        </mrow>
        <mrow>
          <mi>R</mi>
          <mi>A</mi>
          <mi>T</mi>
          <mi>E</mi>
        </mrow>
      </mfrac>
      <mo>=</mo>
      <mn>0</mn>
    </mrow>
    <annotation encoding="application/x-tex">FV + PV \times (1+RATE)^{NPER} + PMT \times \frac{(1+RATE)^{NPER} - 1}{RATE} =0</annotation>
  </semantics>
</math>

Where

| Variable | Description        | Paisa                                            |
|----------|--------------------|--------------------------------------------------|
| FV       | Future Value       | `target`                                         |
| PV       | Present Value      | Current Savings                                  |
| PMT      | Payment per period | `payment_per_period` / Monthly Investment needed |
| RATE     | Rate of return     | `rate`                                           |
| NPER     | Number of periods  | `target_date`                                    |

Out of the 5 variables, if you know any 4, the 5<sup>th</sup> can be calculated by
solving the equation. You can refer the [blog post](https://ciju.in/posts/understanding-financial-functions-excel-sheets) for more details.
