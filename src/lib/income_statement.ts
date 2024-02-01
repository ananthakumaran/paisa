import * as d3 from "d3";
import {
  formatCurrency,
  formatCurrencyCrude,
  tooltip,
  type IncomeStatement,
  rem,
  firstNames
} from "./utils";
import COLORS from "./colors";
import _ from "lodash";
import { iconGlyph, iconify } from "./icon";
import { pathArrows } from "d3-path-arrows";

export function renderIncomeStatement(element: Element) {
  const BARS = 4;
  const BAR_HEIGHT = 100;

  const svg = d3.select(element),
    margin = { top: rem(20), right: rem(20), bottom: rem(10), left: rem(110) },
    width = Math.max(element.parentElement.clientWidth, 600) - margin.left - margin.right,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const height = BAR_HEIGHT * BARS;
  svg
    .attr("height", height + margin.top + margin.bottom)
    .attr("width", width + margin.left + margin.right);

  const sum = (object: Record<string, number>) => Object.values(object).reduce((a, b) => a + b, 0);
  const y = d3.scaleBand().range([height, 0]).paddingInner(0.4).paddingOuter(0.6);
  const x = d3.scaleLinear().range([0, width]);

  const xAxis = g
    .append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")");

  const yAxis = g.append("g").attr("class", "axis y dark is-large");

  const garrows = g.append("g");
  const gbars = g.append("g");
  const glines = g.append("g");
  const gmarks = g.append("g");
  const gamounts = g.append("g");
  const gicons = g.append("g");

  interface Bar {
    label: string;
    value: number;
    start: number;
    end: number;
    color: string;
    breakdown: Record<string, number>;
    multiplier: number;
  }

  const arrows = pathArrows()
    .arrowLength(10)
    .gapLength(100)
    .arrowHeadSize(3)
    .path((d: Bar) => {
      const path = d3.path();

      const startY = y(d.label) + y.bandwidth() + 4;
      const startX = x(d.start);
      const endX = x(d.end);

      path.moveTo(startX, startY);
      path.lineTo(endX, startY);
      return path.toString();
    });

  let firstRender = true;
  return function (statement: IncomeStatement) {
    const incomeStart = statement.startingBalance;
    const income = sum(statement.income) * -1;
    const taxStart = incomeStart + income;
    const tax = sum(statement.tax) * -1;
    const interestStart = taxStart + tax;
    const interest = sum(statement.interest) * -1;
    const pnlStart = interestStart + interest;
    const pnl = sum(statement.pnl);
    const equityStart = pnlStart + pnl;
    const equity = sum(statement.equity) * -1;
    const liabilitiesStart = equityStart + equity;
    const liabilities = sum(statement.liabilities) * -1;
    const expensesStart = liabilitiesStart + liabilities;
    const expenses = sum(statement.expenses) * -1;
    const expensesEnd = expensesStart + expenses;
    const t = svg.transition().duration(firstRender ? 0 : 750);
    firstRender = false;

    const bars: Bar[] = [
      {
        label: "Income",
        start: incomeStart,
        end: incomeStart + income,
        color: COLORS.income,
        value: income,
        breakdown: statement.income,
        multiplier: -1
      },
      {
        label: "Tax",
        start: taxStart,
        end: taxStart + tax,
        color: COLORS.expenses,
        value: tax,
        breakdown: statement.tax,
        multiplier: -1
      },
      {
        label: "Interest",
        start: interestStart,
        end: interestStart + interest,
        color: COLORS.income,
        value: interest,
        breakdown: statement.interest,
        multiplier: -1
      },
      {
        label: "Gain / Loss",
        start: pnlStart,
        end: pnlStart + pnl,
        color: pnl > 0 ? COLORS.gain : COLORS.loss,
        value: pnl,
        breakdown: statement.pnl,
        multiplier: 1
      },
      {
        label: "Equity",
        start: equityStart,
        end: equityStart + equity,
        color: COLORS.equity,
        value: equity,
        breakdown: statement.equity,
        multiplier: -1
      },
      {
        label: "Liabilities",
        start: liabilitiesStart,
        end: liabilitiesStart + liabilities,
        color: COLORS.liabilities,
        value: liabilities,
        breakdown: statement.liabilities,
        multiplier: -1
      },
      {
        label: "Expenses",
        start: expensesStart,
        end: expensesStart + expenses,
        color: COLORS.expenses,
        value: expenses,
        breakdown: statement.expenses,
        multiplier: -1
      }
    ];

    interface Line {
      label: string;
      value: number;
      anchor: string;
      down?: boolean;
      icon?: string;
    }

    const lines: Line[] = [
      { label: "Income", value: incomeStart, anchor: "start", icon: "fa6-solid:caret-down" },
      { label: "Tax", value: taxStart, anchor: "end" },
      { label: "Interest", value: interestStart, anchor: "end" },
      { label: "Gain / Loss", value: pnlStart, anchor: "end" },
      { label: "Equity", value: equityStart, anchor: "end" },
      { label: "Liabilities", value: liabilitiesStart, anchor: "end" },
      { label: "Expenses", value: expensesStart, anchor: "end" },
      {
        label: "Expenses",
        value: expensesEnd,
        down: true,
        anchor: "end",
        icon: "fa6-solid:caret-up"
      }
    ];

    y.domain(bars.map((d) => d.label).reverse());
    x.domain(
      d3.extent([
        incomeStart,
        interestStart,
        taxStart,
        pnlStart,
        equityStart,
        liabilitiesStart,
        expensesStart,
        expensesEnd
      ])
    );

    xAxis.transition(t).call(d3.axisTop(x).tickSize(height).tickFormat(formatCurrencyCrude));
    yAxis.transition(t).call(d3.axisLeft(y).tickSize(-width).tickPadding(10));

    garrows.selectAll("g").remove();
    t.on("end", () => {
      garrows.selectAll("g").data(bars).join("g").attr("class", "g-arrow is-light").call(arrows);
    });

    gbars
      .selectAll("rect")
      .data(bars)
      .join("rect")
      .attr("stroke", (d) => d.color)
      .attr("fill", (d) => d.color)
      .attr("fill-opacity", 0.5)
      .attr("data-tippy-content", (d) => {
        const secondLevelBreakdown = _.chain(d.breakdown)
          .toPairs()
          .groupBy((pair) => firstNames(pair[0], 2))
          .map((pairs, label) => [label, _.sumBy(pairs, (pair) => pair[1])])
          .fromPairs()
          .value();

        return tooltip(
          _.map(secondLevelBreakdown, (value, label) => [
            iconify(label),
            [formatCurrency(value * d.multiplier), "has-text-right has-text-weight-bold"]
          ]),
          { header: d.label, total: formatCurrency(d.value) }
        );
      })
      .transition(t)
      .attr("x", function (d) {
        if (d.value < 0) {
          return x(d.end);
        }
        return x(d.start);
      })
      .attr("y", function (d) {
        return y(d.label) + (y.bandwidth() - Math.min(y.bandwidth(), BAR_HEIGHT)) / 2;
      })
      .attr("width", function (d) {
        if (d.value < 0) {
          return x(d.start) - x(d.end);
        }
        return x(d.end) - x(d.start);
      })
      .attr("height", y.bandwidth());

    glines
      .selectAll("line")
      .data(lines)
      .join("line")
      .attr("class", "svg-grey")
      .attr("stroke-width", 1)
      .attr("stroke-dasharray", "2,2")
      .attr("stroke-opacity", 0.5)
      .transition(t)
      .attr("x1", function (d) {
        return x(d.value);
      })
      .attr("x2", function (d) {
        return x(d.value);
      })
      .attr("y1", function (d) {
        if (d.down) {
          return y(d.label);
        } else {
          return y(d.label) - y.step() * y.paddingInner();
        }
      })
      .attr("y2", function (d) {
        if (d.down) {
          return y(d.label) + y.bandwidth() + y.step() * y.paddingInner();
        } else {
          return y(d.label) + y.bandwidth();
        }
      });

    gmarks
      .selectAll("text")
      .data(lines)
      .join("text")
      .attr("text-anchor", (d) => d.anchor)
      .attr("font-size", "0.7rem")
      .attr("class", "svg-text-grey")
      .attr("dy", (d) => (d.down ? "-0.5rem" : "1rem"))
      .attr("dx", (d) => (d.anchor === "start" ? "0.3rem" : "-0.3rem"))
      .transition(t)
      .attr("x", function (d) {
        return x(d.value);
      })
      .attr("y", function (d) {
        if (d.down) {
          return y(d.label) + y.bandwidth() + y.step() * y.paddingInner();
        } else {
          return y(d.label) - y.step() * y.paddingInner();
        }
      })
      .text((d) => formatCurrency(d.value));

    gamounts
      .selectAll("text")
      .data(bars)
      .join("text")
      .attr("dy", "0.3rem")
      .attr("text-anchor", "middle")
      .attr("class", "svg-text-black-ter has-text-weight-bold")
      .transition(t)
      .attr("x", function (d) {
        return (x(d.start) + x(d.end)) / 2;
      })
      .attr("y", function (d) {
        return y(d.label) + y.bandwidth() / 2;
      })
      .text((d) => formatCurrency(d.value));

    gicons
      .selectAll("text")
      .data([_.first(lines), _.last(lines)])
      .join("text")
      .attr("text-anchor", "middle")
      .attr("font-size", "1.2rem")
      .attr("class", "svg-text-grey")
      .attr("dy", (d) => (d.down ? "0.8rem" : "0.2rem"))
      .transition(t)
      .attr("x", function (d) {
        return x(d.value);
      })
      .attr("y", function (d) {
        if (d.down) {
          return y(d.label) + y.bandwidth() + y.step() * y.paddingInner();
        } else {
          return y(d.label) - y.step() * y.paddingInner();
        }
      })
      .text((d) => iconGlyph(d.icon));
  };
}
