import type { Issue, Severity } from "./utils";

// Display order: most actionable first. Drives both the top counter and the
// grouped layout on the Doctor page.
export const SEVERITY_ORDER: Severity[] = ["error", "warning", "info"];

export interface SeverityGroup {
  severity: Severity;
  issues: Issue[];
}

// Bulma color class for a severity. Severity is the semantic name; bulma
// classes use "danger" for errors.
export function bulmaClassFor(severity: Severity): string {
  switch (severity) {
    case "error":
      return "is-danger";
    case "warning":
      return "is-warning";
    case "info":
      return "is-info";
  }
}

// groupBySeverity partitions issues by severity in SEVERITY_ORDER. Severities
// with no issues are omitted so the UI does not render empty buckets.
export function groupBySeverity(issues: Issue[]): SeverityGroup[] {
  const buckets: Record<Severity, Issue[]> = {
    error: [],
    warning: [],
    info: []
  };
  for (const issue of issues) {
    // Fallback: treat unknown / missing severity as error so users still see
    // it (loud-failure is safer than silent-drop).
    const sev: Severity =
      issue.severity === "warning" || issue.severity === "info" ? issue.severity : "error";
    buckets[sev].push(issue);
  }
  return SEVERITY_ORDER.filter((s) => buckets[s].length > 0).map((severity) => ({
    severity,
    issues: buckets[severity]
  }));
}

// Human-readable label for the counter, e.g. "errors" / "warnings" / "infos".
export function severityLabel(severity: Severity, count: number): string {
  const base = severity === "error" ? "error" : severity === "warning" ? "warning" : "info";
  return count === 1 ? base : `${base}s`;
}
