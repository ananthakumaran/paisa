import { describe, expect, test } from "@jest/globals";
import { parseDate } from "./search_query_editor";

function assertRange(text: string, start: string, end: string) {
  const result = parseDate(text, reference);
  expect(result.start.format("YYYY-MM-DDTHH:mm:ss")).toEqual(start);
  expect(result.end.format("YYYY-MM-DDTHH:mm:ss")).toEqual(end);
}

const reference = new Date("2023-09-03");
describe("parseDate", () => {
  test("natural language", () => {
    assertRange("today", "2023-09-03T00:00:00", "2023-09-03T23:59:59");
    assertRange("this month", "2023-09-01T00:00:00", "2023-09-30T23:59:59");
    assertRange("last month", "2023-08-01T00:00:00", "2023-08-31T23:59:59");
    assertRange("this year", "2023-01-01T00:00:00", "2023-12-31T23:59:59");
    assertRange("2023", "2023-01-01T00:00:00", "2023-12-31T23:59:59");
    assertRange("2023-01", "2023-01-01T00:00:00", "2023-01-31T23:59:59");
    assertRange("last week", "2023-08-27T00:00:00", "2023-09-02T23:59:59");
    assertRange("jan 2023", "2023-01-01T00:00:00", "2023-01-31T23:59:59");
  });
});
