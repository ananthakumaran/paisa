import { fireEvent, render, screen } from "@testing-library/svelte";
import { expect, test } from "vitest";
import FileTree from "./FileTree.svelte";

test("selects a nested ledger file and displays unsaved state", async () => {
  const files = [{
    type: "directory" as const,
    name: "accounts",
    children: [{
      type: "file" as const,
      name: "accounts/main.ledger",
      content: "",
      versions: [],
    }],
  }];
  const { component } = render(FileTree, {
    files,
    path: "",
    selectedFileName: "accounts/main.ledger",
    hasUnsavedChanges: true,
  });
  let selected = "";
  component.$on("select", (event) => selected = event.detail.name);

  expect(screen.getByText("unsaved")).toBeInTheDocument();
  await fireEvent.click(screen.getByTitle("main.ledger"));
  expect(selected).toBe("accounts/main.ledger");
});
