import { describe, it, expect } from "vitest";
import {
  VENDOR_PRESETS,
  getVendorPreset,
  composeCredentialsRef,
  parseCredentialsRef,
  SDK_VENDORS,
} from "./integration-provider-presets";

describe("integration provider presets", () => {
  it("includes a custom preset and one preset per supported SDK vendor", () => {
    expect(getVendorPreset("custom").id).toBe("custom");
    for (const vendor of ["gitea", "github", "gitlab", "jira", "jenkins", "bitbucket", "bamboo"]) {
      const preset = getVendorPreset(vendor);
      expect(preset.id).toBe(vendor);
      expect(preset.sdkVendor).toBe(vendor);
      expect((SDK_VENDORS as readonly string[]).includes(vendor)).toBe(true);
    }
  });

  it("maps vendor presets to the correct provider_type", () => {
    expect(getVendorPreset("gitea").providerType).toBe("scm");
    expect(getVendorPreset("github").providerType).toBe("scm");
    expect(getVendorPreset("jira").providerType).toBe("alm");
    expect(getVendorPreset("jenkins").providerType).toBe("ci_cd");
  });

  it("falls back to the custom preset for unknown ids", () => {
    expect(getVendorPreset("nope").id).toBe("custom");
    expect(VENDOR_PRESETS[0].id).toBe("custom");
  });
});

describe("composeCredentialsRef", () => {
  it("composes hmac_sha256 strategy", () => {
    expect(composeCredentialsRef("hmac_sha256", undefined, "  s3cret ")).toBe("hmac_sha256:s3cret");
  });

  it("composes provider_sdk strategy with vendor", () => {
    expect(composeCredentialsRef("provider_sdk", "github", "tok")).toBe("provider_sdk:github:tok");
  });

  it("defaults provider_sdk vendor to gitea when missing", () => {
    expect(composeCredentialsRef("provider_sdk", undefined, "tok")).toBe("provider_sdk:gitea:tok");
  });

  it("returns the raw token for shared_token strategy", () => {
    expect(composeCredentialsRef("shared_token", undefined, "plain-token")).toBe("plain-token");
  });
});

describe("parseCredentialsRef", () => {
  it("round-trips hmac_sha256", () => {
    const parsed = parseCredentialsRef("hmac_sha256:abc");
    expect(parsed.strategy).toBe("hmac_sha256");
    expect(parsed.hasSecret).toBe(true);
  });

  it("round-trips provider_sdk with vendor + secret", () => {
    const parsed = parseCredentialsRef("provider_sdk:gitlab:xyz");
    expect(parsed.strategy).toBe("provider_sdk");
    expect(parsed.sdkVendor).toBe("gitlab");
    expect(parsed.hasSecret).toBe(true);
  });

  it("handles provider_sdk without secret", () => {
    const parsed = parseCredentialsRef("provider_sdk:gitea");
    expect(parsed.strategy).toBe("provider_sdk");
    expect(parsed.sdkVendor).toBe("gitea");
    expect(parsed.hasSecret).toBe(false);
  });

  it("marks unknown sdk vendor as undefined", () => {
    const parsed = parseCredentialsRef("provider_sdk:weirdvendor:s");
    expect(parsed.strategy).toBe("provider_sdk");
    expect(parsed.sdkVendor).toBeUndefined();
    expect(parsed.hasSecret).toBe(true);
  });

  it("treats anything else as a shared token", () => {
    const parsed = parseCredentialsRef("plain-token");
    expect(parsed.strategy).toBe("shared_token");
    expect(parsed.hasSecret).toBe(true);
  });

  it("reports no secret for empty input", () => {
    expect(parseCredentialsRef("").hasSecret).toBe(false);
  });

  it("does not leak the secret value (only presence)", () => {
    const parsed = parseCredentialsRef("hmac_sha256:supersecret");
    expect(JSON.stringify(parsed)).not.toContain("supersecret");
  });
});
