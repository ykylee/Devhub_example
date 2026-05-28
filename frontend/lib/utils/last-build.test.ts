import { describe, expect, it } from "vitest";
import {
  applicationBuildStatusView,
  repositoryLastBuildView,
} from "./last-build";

describe("applicationBuildStatusView", () => {
  it("healthy → success badge", () => {
    expect(applicationBuildStatusView("healthy")).toEqual({
      label: "Healthy",
      variant: "success",
      tone: "positive",
    });
  });

  it("broken → danger badge", () => {
    expect(applicationBuildStatusView("broken")).toEqual({
      label: "Broken",
      variant: "danger",
      tone: "negative",
    });
  });

  it("unknown/null/undefined → secondary fallback", () => {
    expect(applicationBuildStatusView("unknown").variant).toBe("secondary");
    expect(applicationBuildStatusView(null).variant).toBe("secondary");
    expect(applicationBuildStatusView(undefined).variant).toBe("secondary");
    expect(applicationBuildStatusView("unexpected").variant).toBe("secondary");
  });
});

describe("repositoryLastBuildView", () => {
  it("success → success badge", () => {
    const v = repositoryLastBuildView("success");
    expect(v.variant).toBe("success");
    expect(v.tone).toBe("positive");
  });

  it("failed → danger badge", () => {
    const v = repositoryLastBuildView("failed");
    expect(v.variant).toBe("danger");
    expect(v.tone).toBe("negative");
  });

  it("cancelled → warning tone negative", () => {
    const v = repositoryLastBuildView("cancelled");
    expect(v.variant).toBe("warning");
    expect(v.tone).toBe("negative");
  });

  it("running → warning tone neutral", () => {
    const v = repositoryLastBuildView("running");
    expect(v.variant).toBe("warning");
    expect(v.tone).toBe("neutral");
  });

  it.each(["queued", "skipped", "unknown", null, undefined, "unexpected"])(
    "%s → secondary neutral",
    (input) => {
      const v = repositoryLastBuildView(input);
      expect(v.variant).toBe("secondary");
      expect(v.tone).toBe("neutral");
    },
  );
});
