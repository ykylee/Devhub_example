import { describe, it, expect } from "vitest";
import { cn, formatBytes } from "@/shared/utils";

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

describe("formatBytes", () => {
  it("returns '0 B' for 0", () => {
    expect(formatBytes(0)).toBe("0 B");
  });

  it("formats bytes under 1024 with B unit", () => {
    expect(formatBytes(512)).toBe("512 B");
  });

  it("formats kilobytes with KB unit and 1 decimal", () => {
    expect(formatBytes(2048)).toBe("2 KB");
    expect(formatBytes(1536)).toBe("1.5 KB");
  });

  it("formats megabytes with MB unit", () => {
    expect(formatBytes(1024 * 1024)).toBe("1 MB");
    expect(formatBytes(5 * 1024 * 1024)).toBe("5 MB");
  });

  it("formats gigabytes with GB unit", () => {
    expect(formatBytes(2 * 1024 * 1024 * 1024)).toBe("2 GB");
  });

  it("formats terabytes with TB unit", () => {
    expect(formatBytes(1024 ** 4)).toBe("1 TB");
  });
});
