import { describe, expect, it, vi } from "vitest";

describe("endpoints configuration", () => {
  it("resolves BASE_PATH and APP_ORIGIN based on environment variables", async () => {
    vi.stubEnv("NEXT_PUBLIC_BASE_PATH", "/my-base");
    vi.stubEnv("NEXT_PUBLIC_APP_ORIGIN", "https://my-app.com/");
    vi.stubEnv("NEXT_PUBLIC_API_URL", "https://api.my-app.com");
    vi.stubEnv("NEXT_PUBLIC_WS_URL", "wss://ws.my-app.com");
    vi.stubEnv("NEXT_PUBLIC_IDP_PROVIDER", "keycloak");

    const { BASE_PATH, API_BASE_URL, WS_BASE_URL, IDP_PROVIDER } = await import("./endpoints?t=1");

    expect(BASE_PATH).toBe("/my-base");
    expect(API_BASE_URL).toBe("https://api.my-app.com");
    expect(WS_BASE_URL).toBe("wss://ws.my-app.com");
    expect(IDP_PROVIDER).toBe("keycloak");

    vi.unstubAllEnvs();
  });

  it("handles empty environment defaults", async () => {
    vi.stubEnv("NEXT_PUBLIC_BASE_PATH", "");
    vi.stubEnv("NEXT_PUBLIC_APP_ORIGIN", "");
    vi.stubEnv("NEXT_PUBLIC_API_URL", "");
    vi.stubEnv("NEXT_PUBLIC_WS_URL", "");
    vi.stubEnv("NEXT_PUBLIC_IDP_PROVIDER", "");

    const { BASE_PATH, API_BASE_URL, BACKEND_API_URL_SERVER } = await import("./endpoints?t=2");

    expect(BASE_PATH).toBe("");
    expect(API_BASE_URL).toBe("");
    expect(BACKEND_API_URL_SERVER).toBe("http://localhost:8080");

    vi.unstubAllEnvs();
  });

  it("calculates Keycloak Admin Console URL gracefully", async () => {
    // 1. Explicit env
    vi.stubEnv("NEXT_PUBLIC_KC_ADMIN_URL", "https://kc.admin.com");
    const { getKCAdminConsoleUrl } = await import("./endpoints?t=3");
    expect(getKCAdminConsoleUrl()).toBe("https://kc.admin.com");

    // 2. Fallback from issuer with realm
    vi.stubEnv("NEXT_PUBLIC_KC_ADMIN_URL", "");
    vi.stubEnv("NEXT_PUBLIC_OIDC_ISSUER_URL", "https://auth.my-app.com/realms/myrealm");
    const { getKCAdminConsoleUrl: getUrlWithRealm } = await import("./endpoints?t=4");
    expect(getUrlWithRealm()).toBe("https://auth.my-app.com/admin/myrealm/console/");

    // 3. Fallback from issuer without realms (master)
    vi.stubEnv("NEXT_PUBLIC_OIDC_ISSUER_URL", "https://auth.my-app.com/somepath");
    const { getKCAdminConsoleUrl: getUrlMaster } = await import("./endpoints?t=5");
    expect(getUrlMaster()).toBe("https://auth.my-app.com/admin/master/console/");

    // 4. Empty issuer url
    vi.stubEnv("NEXT_PUBLIC_OIDC_ISSUER_URL", "");
    const { getKCAdminConsoleUrl: getUrlEmpty } = await import("./endpoints?t=6");
    expect(getUrlEmpty()).toBeNull();

    // 5. Invalid issuer url (trigger catch block)
    vi.stubEnv("NEXT_PUBLIC_OIDC_ISSUER_URL", "not-a-valid-url");
    const { getKCAdminConsoleUrl: getUrlInvalid } = await import("./endpoints?t=7");
    expect(getUrlInvalid()).toBeNull();

    vi.unstubAllEnvs();
  });
});
