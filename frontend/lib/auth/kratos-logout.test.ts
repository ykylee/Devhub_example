import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { performKratosBrowserLogout, killKratosSession } from "./kratos-logout";

const originalLocation = window.location;

describe("performKratosBrowserLogout", () => {
  let assignSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    assignSpy = vi.fn();
    // jsdom's window.location is read-only at the property level; replace
    // the whole object so we can spy on assign.
    Object.defineProperty(window, "location", {
      configurable: true,
      value: { ...originalLocation, assign: assignSpy, origin: "http://localhost:3000" },
    });
  });

  afterEach(() => {
    Object.defineProperty(window, "location", { configurable: true, value: originalLocation });
    vi.restoreAllMocks();
  });

  it("navigates to logout_url when Kratos returns one", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({
      ok: true,
      json: async () => ({ logout_url: "http://localhost:4433/self-service/logout?token=abc" }),
    })));

    await performKratosBrowserLogout("/");

    expect(assignSpy).toHaveBeenCalledWith(
      "http://localhost:4433/self-service/logout?token=abc",
    );
  });

  it("falls back to the supplied returnTo when Kratos has no session (flow init fails)", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false, status: 401, json: async () => ({}) })));

    await performKratosBrowserLogout("/fallback");

    expect(assignSpy).toHaveBeenCalledWith("/fallback");
  });

  it("falls back to returnTo when fetch throws", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => {
      throw new Error("network down");
    }));

    await performKratosBrowserLogout("/fallback");

    expect(assignSpy).toHaveBeenCalledWith("/fallback");
  });
});

describe("killKratosSession", () => {
  it("issues GET /self-service/logout/browser and follows up with GET logout_url (manual redirect)", async () => {
    const fetchSpy = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ logout_url: "http://localhost:4433/self-service/logout?token=t" }),
      })
      .mockResolvedValueOnce({ ok: true });
    vi.stubGlobal("fetch", fetchSpy);

    await killKratosSession();

    expect(fetchSpy).toHaveBeenCalledTimes(2);
    const secondCall = fetchSpy.mock.calls[1];
    expect(secondCall[0]).toBe("http://localhost:4433/self-service/logout?token=t");
    expect(secondCall[1]).toMatchObject({ credentials: "include", redirect: "manual" });
  });

  it("returns quietly when there is no Kratos session", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ({ ok: false })));
    await expect(killKratosSession()).resolves.toBeUndefined();
  });
});
