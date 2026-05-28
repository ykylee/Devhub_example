import { describe, expect, it } from "vitest";
import {
  lifecycleStatusBadgeVariant,
  type LifecycleBadgeVariant,
} from "./lifecycle-status";

describe("lifecycleStatusBadgeVariant", () => {
  const cases: Array<{ status: string; variant: LifecycleBadgeVariant }> = [
    { status: "active", variant: "success" },
    { status: "planning", variant: "primary" },
    { status: "on_hold", variant: "warning" },
    { status: "closed", variant: "secondary" },
    { status: "archived", variant: "glass" },
  ];

  it.each(cases)("$status → $variant", ({ status, variant }) => {
    expect(lifecycleStatusBadgeVariant(status)).toBe(variant);
  });

  it("unknown status → secondary (defensive fallback)", () => {
    expect(lifecycleStatusBadgeVariant("")).toBe("secondary");
    expect(lifecycleStatusBadgeVariant("unknown_future_value")).toBe("secondary");
    expect(lifecycleStatusBadgeVariant("ACTIVE")).toBe("secondary"); // case-sensitive
  });
});
