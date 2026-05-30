import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { tokenStore } from "./token-store";
import { refreshAccessToken, _resetRefreshCoordinator } from "./refresh";

// OIDC_ISSUER_URL 모듈 레벨 상수를 빈 값으로 mock -> resolveTokenEndpoint 가
// 항상 /api/runtime-config 를 fetch 하도록 강제. (NEXT_PUBLIC_* env var 는 Next.js
// 빌드 시 인라인되므로 vi.stubEnv 로 격리 불가.)
vi.mock("@/shared/config/endpoints", () => ({
  OIDC_ISSUER_URL: "",
  API_BASE_URL: "",
}));

// 테스트 헬퍼: resolveTokenEndpoint 가 OIDC_ISSUER_URL 미설정 시 `/api/runtime-config`
// 를 fetch 하므로, URL 패턴별로 적절한 응답을 반환하는 mock 을 구성한다.
// (`tokenEndpointFetch` 가 token endpoint 호출에 대해 무엇을 반환할지 시나리오별 지정.)
function setupFetchMock(tokenEndpointFetch: () => Promise<Response>): ReturnType<typeof vi.fn> {
  const fetchMock = vi.fn().mockImplementation(async (url: string | URL) => {
    const s = String(url);
    if (s.includes("/api/runtime-config")) {
      return new Response(JSON.stringify({ oidc_issuer_url: "http://test-issuer" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }
    return tokenEndpointFetch();
  });
  global.fetch = fetchMock as unknown as typeof fetch;
  return fetchMock;
}

describe("refresh.refreshAccessToken (single-flight + outcome classification)", () => {
  let originalFetch: typeof fetch;

  beforeEach(() => {
    sessionStorage.clear();
    tokenStore.clear();
    _resetRefreshCoordinator();
    originalFetch = global.fetch;
  });

  afterEach(() => {
    global.fetch = originalFetch;
  });

  function seedRefreshToken() {
    tokenStore.save({
      access_token: "old",
      refresh_token: "r1",
      expires_in: 300,
      token_type: "Bearer",
    });
  }

  it("no refresh_token → auth_failed", async () => {
    // 토큰 비어있는 상태.
    const outcome = await refreshAccessToken();
    expect(outcome).toEqual({ kind: "auth_failed", reason: "no_refresh_token" });
  });

  it("HTTP 200 + valid body → ok + tokenStore 갱신", async () => {
    seedRefreshToken();
    setupFetchMock(async () =>
      new Response(JSON.stringify({
        access_token: "new",
        refresh_token: "r2",
        expires_in: 600,
        token_type: "Bearer",
      }), { status: 200, headers: { "Content-Type": "application/json" } }),
    );

    const outcome = await refreshAccessToken();
    expect(outcome).toEqual({ kind: "ok" });
    expect(tokenStore.getAccessToken()).toBe("new");
    expect(tokenStore.getRefreshToken()).toBe("r2");
  });

  it("HTTP 4xx (refresh token 거부) → auth_failed", async () => {
    seedRefreshToken();
    setupFetchMock(async () => new Response(JSON.stringify({ error: "invalid_grant" }), { status: 400 }));

    const outcome = await refreshAccessToken();
    expect(outcome.kind).toBe("auth_failed");
    if (outcome.kind === "auth_failed") {
      expect(outcome.reason).toBe("http_400");
    }
  });

  it("HTTP 5xx (IdP 일시 장애) → transient_failed", async () => {
    seedRefreshToken();
    setupFetchMock(async () => new Response("", { status: 503 }));

    const outcome = await refreshAccessToken();
    expect(outcome.kind).toBe("transient_failed");
    if (outcome.kind === "transient_failed") {
      expect(outcome.reason).toBe("http_503");
    }
  });

  it("network error → transient_failed", async () => {
    seedRefreshToken();
    setupFetchMock(async () => { throw new TypeError("Failed to fetch"); });

    const outcome = await refreshAccessToken();
    expect(outcome.kind).toBe("transient_failed");
    if (outcome.kind === "transient_failed") {
      expect(outcome.reason).toBe("network_error");
    }
  });

  it("동시 호출은 single-flight — token endpoint fetch 는 1회만, 동일 outcome 공유", async () => {
    seedRefreshToken();
    let tokenEndpointCalls = 0;
    setupFetchMock(async () => {
      tokenEndpointCalls += 1;
      return new Promise<Response>((resolve) =>
        setTimeout(() => resolve(new Response(JSON.stringify({
          access_token: "new",
          refresh_token: "r2",
          expires_in: 600,
          token_type: "Bearer",
        }), { status: 200 })), 50),
      );
    });

    const [a, b, c] = await Promise.all([
      refreshAccessToken(),
      refreshAccessToken(),
      refreshAccessToken(),
    ]);
    expect(a).toEqual({ kind: "ok" });
    expect(b).toEqual({ kind: "ok" });
    expect(c).toEqual({ kind: "ok" });
    // token endpoint 는 단 1회만 호출 — Keycloak 의 refresh_token 한 번만 소비.
    expect(tokenEndpointCalls).toBe(1);
  });

  it("실패 outcome 도 동시 호출 시 공유되며 token endpoint 호출 1회", async () => {
    seedRefreshToken();
    let tokenEndpointCalls = 0;
    setupFetchMock(async () => {
      tokenEndpointCalls += 1;
      return new Promise<Response>((resolve) =>
        setTimeout(() => resolve(new Response("", { status: 400 })), 50),
      );
    });

    const [a, b] = await Promise.all([refreshAccessToken(), refreshAccessToken()]);
    expect(a.kind).toBe("auth_failed");
    expect(b.kind).toBe("auth_failed");
    expect(tokenEndpointCalls).toBe(1);
  });

  it("연속 호출(이전 완료 후) 은 새 token endpoint fetch 발생", async () => {
    seedRefreshToken();
    let tokenEndpointCalls = 0;
    setupFetchMock(async () => {
      tokenEndpointCalls += 1;
      return new Response(JSON.stringify({
        access_token: "new",
        refresh_token: "r2",
        expires_in: 600,
        token_type: "Bearer",
      }), { status: 200 });
    });

    await refreshAccessToken();
    await refreshAccessToken();
    expect(tokenEndpointCalls).toBe(2);
  });

  it("runtime-config fetch !ok → no_token_endpoint (transient_failed)", async () => {
    seedRefreshToken();
    global.fetch = vi.fn().mockImplementation(async (url: string | URL) => {
      const s = String(url);
      if (s.includes("/api/runtime-config")) return new Response("", { status: 503 });
      throw new Error("token endpoint should not be called");
    }) as unknown as typeof fetch;

    const outcome = await refreshAccessToken();
    expect(outcome).toEqual({ kind: "transient_failed", reason: "no_token_endpoint" });
  });

  it("runtime-config returns empty oidc_issuer_url → no_token_endpoint", async () => {
    seedRefreshToken();
    global.fetch = vi.fn().mockImplementation(async (url: string | URL) => {
      const s = String(url);
      if (s.includes("/api/runtime-config")) {
        return new Response(JSON.stringify({ oidc_issuer_url: "   " }), { status: 200 });
      }
      throw new Error("token endpoint should not be called");
    }) as unknown as typeof fetch;

    const outcome = await refreshAccessToken();
    expect(outcome).toEqual({ kind: "transient_failed", reason: "no_token_endpoint" });
  });

  it("runtime-config returns no oidc_issuer_url field → no_token_endpoint", async () => {
    seedRefreshToken();
    global.fetch = vi.fn().mockImplementation(async (url: string | URL) => {
      const s = String(url);
      if (s.includes("/api/runtime-config")) {
        return new Response(JSON.stringify({}), { status: 200 });
      }
      throw new Error("token endpoint should not be called");
    }) as unknown as typeof fetch;

    const outcome = await refreshAccessToken();
    expect(outcome).toEqual({ kind: "transient_failed", reason: "no_token_endpoint" });
  });

  it("runtime-config fetch throws → no_token_endpoint (catch branch)", async () => {
    seedRefreshToken();
    global.fetch = vi.fn().mockImplementation(async (url: string | URL) => {
      const s = String(url);
      if (s.includes("/api/runtime-config")) throw new TypeError("network down");
      throw new Error("token endpoint should not be called");
    }) as unknown as typeof fetch;

    const outcome = await refreshAccessToken();
    expect(outcome).toEqual({ kind: "transient_failed", reason: "no_token_endpoint" });
  });

  it("response.json parse error after 200 → transient_failed parse_error", async () => {
    seedRefreshToken();
    setupFetchMock(async () => new Response("not-json {", { status: 200, headers: { "Content-Type": "application/json" } }));

    const outcome = await refreshAccessToken();
    expect(outcome.kind).toBe("transient_failed");
    if (outcome.kind === "transient_failed") {
      expect(outcome.reason).toBe("parse_error");
    }
  });
});
