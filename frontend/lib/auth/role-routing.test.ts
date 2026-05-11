import { describe, it, expect } from "vitest";
import { defaultLandingFor, isSystemAdmin, pathRequiresSystemAdmin } from "./role-routing";

describe("defaultLandingFor", () => {
  it("routes system_admin to /admin", () => {
    expect(defaultLandingFor("System Admin")).toBe("/admin");
  });

  it("routes manager to /manager", () => {
    expect(defaultLandingFor("Manager")).toBe("/manager");
  });

  it("routes developer to /developer", () => {
    expect(defaultLandingFor("Developer")).toBe("/developer");
  });

  it("falls back to /developer for null/undefined", () => {
    expect(defaultLandingFor(null)).toBe("/developer");
    expect(defaultLandingFor(undefined)).toBe("/developer");
  });
});

describe("isSystemAdmin", () => {
  it("is true only for System Admin", () => {
    expect(isSystemAdmin("System Admin")).toBe(true);
    expect(isSystemAdmin("Manager")).toBe(false);
    expect(isSystemAdmin("Developer")).toBe(false);
    expect(isSystemAdmin(null)).toBe(false);
    expect(isSystemAdmin(undefined)).toBe(false);
  });
});

describe("pathRequiresSystemAdmin", () => {
  it("recognises /admin and /admin/* as system_admin gated", () => {
    expect(pathRequiresSystemAdmin("/admin")).toBe(true);
    expect(pathRequiresSystemAdmin("/admin/settings")).toBe(true);
    expect(pathRequiresSystemAdmin("/admin/settings/users")).toBe(true);
  });

  it("recognises legacy /organization paths until PR-S2 retirement", () => {
    expect(pathRequiresSystemAdmin("/organization")).toBe(true);
    expect(pathRequiresSystemAdmin("/organization/units")).toBe(true);
  });

  it("does not gate developer/manager/auth paths", () => {
    expect(pathRequiresSystemAdmin("/developer")).toBe(false);
    expect(pathRequiresSystemAdmin("/manager")).toBe(false);
    expect(pathRequiresSystemAdmin("/account")).toBe(false);
    expect(pathRequiresSystemAdmin("/auth/login")).toBe(false);
    expect(pathRequiresSystemAdmin("/")).toBe(false);
  });
});
