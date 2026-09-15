import { describe, expect, it } from "vitest";

import { taskListPath } from "@/features/task-workflows/task-workflow-api";

describe("taskListPath", () => {
  it("builds filtered task master list paths", () => {
    expect(
      taskListPath("types", {
        search: "cycle count",
        active: "active",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/master/task-types?page=2&page_size=10&search=cycle+count&active=true",
    );
  });

  it("omits optional filters", () => {
    expect(
      taskListPath("transitions", {
        search: "",
        active: "all",
        page: 1,
        pageSize: 100,
      }),
    ).toBe("/api/v1/master/task-transitions?page=1&page_size=100");
  });
});
