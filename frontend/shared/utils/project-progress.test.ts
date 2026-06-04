import { describe, expect, it } from "vitest";
import type { ProjectTaskItem } from "@/domain/application-lifecycle/schema/project.types";
import { computeProjectProgress } from "./project-progress";

const task = (status: ProjectTaskItem["status"]): ProjectTaskItem => ({
  id: `t-${status}`,
  title: `Task ${status}`,
  priority: "medium",
  status,
});

describe("computeProjectProgress", () => {
  it("returns null when no tasks", () => {
    expect(computeProjectProgress([])).toBeNull();
  });

  it("returns 0 when no done tasks", () => {
    const tasks = [task("todo"), task("in_progress"), task("review")];
    expect(computeProjectProgress(tasks)).toBe(0);
  });

  it("returns 100 when all done", () => {
    const tasks = [task("done"), task("done"), task("done")];
    expect(computeProjectProgress(tasks)).toBe(100);
  });

  it("rounds percentage correctly (33% of 3 tasks)", () => {
    const tasks = [task("done"), task("todo"), task("todo")];
    expect(computeProjectProgress(tasks)).toBe(33);
  });

  it("rounds percentage correctly (67% of 3 tasks)", () => {
    const tasks = [task("done"), task("done"), task("todo")];
    expect(computeProjectProgress(tasks)).toBe(67);
  });

  it("handles single done task", () => {
    expect(computeProjectProgress([task("done")])).toBe(100);
  });

  it("handles single todo task", () => {
    expect(computeProjectProgress([task("todo")])).toBe(0);
  });
});
