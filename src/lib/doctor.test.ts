import { describe, expect, test } from "bun:test";
import { groupBySeverity, SEVERITY_ORDER, bulmaClassFor } from "./doctor";
import type { Issue } from "./utils";

function makeIssue(severity: Issue["severity"], summary = "test"): Issue {
  return {
    severity,
    level: severity === "error" ? "danger" : severity,
    summary,
    description: "",
    details: ""
  };
}

describe("doctor", () => {
  test("Issue type accepts severity", () => {
    const issue = makeIssue("info");
    expect(issue.severity).toBe("info");
  });

  test("groupBySeverity buckets issues by severity in declared order", () => {
    const issues: Issue[] = [
      makeIssue("info", "a"),
      makeIssue("error", "b"),
      makeIssue("warning", "c"),
      makeIssue("info", "d")
    ];
    const groups = groupBySeverity(issues);
    expect(groups.map((g) => g.severity)).toEqual(["error", "warning", "info"]);
    expect(groups.find((g) => g.severity === "info")!.issues.length).toBe(2);
    expect(groups.find((g) => g.severity === "error")!.issues.length).toBe(1);
    expect(groups.find((g) => g.severity === "warning")!.issues.length).toBe(1);
  });

  test("groupBySeverity omits empty severities", () => {
    const groups = groupBySeverity([makeIssue("info")]);
    expect(groups.map((g) => g.severity)).toEqual(["info"]);
  });

  test("SEVERITY_ORDER ranks error > warning > info", () => {
    expect(SEVERITY_ORDER).toEqual(["error", "warning", "info"]);
  });

  test("bulmaClassFor maps severity to bulma class", () => {
    expect(bulmaClassFor("error")).toBe("is-danger");
    expect(bulmaClassFor("warning")).toBe("is-warning");
    expect(bulmaClassFor("info")).toBe("is-info");
  });
});
