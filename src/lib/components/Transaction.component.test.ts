import { render, screen } from "@testing-library/svelte";
import { expect, test } from "vitest";
import dayjs from "dayjs";
import Transaction from "./Transaction.svelte";

test("renders a transaction payee, date, and debit/credit postings", () => {
  render(Transaction, {
    compact: false,
    t: {
      id: 1,
      date: dayjs("2022-02-07"),
      payee: "Grocery Store",
      note: "",
      postings: [
        {
          account: "Expenses:Food",
          amount: 500,
          quantity: 500,
          commodity: "INR",
          payee: "Grocery Store",
          date: dayjs("2022-02-07"),
        },
        {
          account: "Assets:Checking",
          amount: -500,
          quantity: -500,
          commodity: "INR",
          payee: "Grocery Store",
          date: dayjs("2022-02-07"),
        },
      ],
    },
  });
  expect(screen.getByText("Grocery Store")).toBeVisible();
  expect(screen.getByText("Expenses:Food")).toBeVisible();
  expect(screen.getByText("Assets:Checking")).toBeVisible();
});
