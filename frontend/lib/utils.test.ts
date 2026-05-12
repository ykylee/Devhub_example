import { describe, it, expect } from "vitest";
import { cn } from "./utils";

describe("cn (tailwind-merge helper)", () => {
  it("should merge basic class names", () => {
    expect(cn("px-2", "py-2")).toBe("px-2 py-2");
  });

  it("should handle conditional classes", () => {
    expect(cn("px-2", true && "py-2", false && "m-2")).toBe("px-2 py-2");
  });

  it("should merge tailwind classes correctly (last one wins)", () => {
    // tailwind-merge should resolve px-2 vs px-4
    expect(cn("px-2", "px-4")).toBe("px-4");
  });

  it("should handle undefined and null", () => {
    expect(cn("px-2", undefined, null, "py-2")).toBe("px-2 py-2");
  });
});
