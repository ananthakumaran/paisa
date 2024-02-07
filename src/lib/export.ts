import type { BalancedPosting } from "./utils";
import Papa from "papaparse";

export function download(balancedPostings: BalancedPosting[]) {
  const rows = balancedPostings.map((balancedPosting) => {
    console.log(balancedPosting);
    return {
      Date: balancedPosting.from.date.toISOString(),
      Payee: balancedPosting.from.payee,
      FromAccount: balancedPosting.from.account,
      FromQuantity: balancedPosting.from.quantity,
      FromAmount: balancedPosting.from.amount,
      FromCommodity: balancedPosting.from.commodity,
      ToAccount: balancedPosting.to.account,
      ToQuantity: balancedPosting.to.quantity,
      ToAmount: balancedPosting.to.amount,
      ToCommodity: balancedPosting.to.commodity
    };
  });

  const csv = Papa.unparse(rows);
  const downloadLink = document.createElement("a");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  downloadLink.href = window.URL.createObjectURL(blob);
  downloadLink.download = "paisa-transactions.csv";
  downloadLink.click();
}
