// 서비스 간 주소 default 를 한 곳에서 관리한다.
// 정책 (CLAUDE.md "native default, docker optional"):
//   - 코드 default 는 모두 native(localhost) 기준이다.
//   - docker / 다른 환경에서 띄울 때는 env 로 override 한다.
//     예: docker compose 의 environment 에 `BACKEND_API_URL=http://backend-core:8080` 주입.
//   - .env.local.example 의 주석에 docker 케이스 사용법을 명시한다.
//
// client-side(`NEXT_PUBLIC_*`) 변수는 브라우저로 inline 되므로 빌드 시점에 고정된다.
// server-only 변수는 런타임에 평가된다.

const stripTrailingSlash = (u: string) => u.replace(/\/$/, "");

// Retrieve and normalize NEXT_PUBLIC_BASE_PATH (e.g. "/devhub")
// sprint -s (PR #187): exported for auth.service.ts logout URI 정합 (basePath 포함 logout
// — sprint -j codex review #9 #4 backend 확장 carve #3 구현).
export const BASE_PATH = process.env.NEXT_PUBLIC_BASE_PATH
  ? `/${process.env.NEXT_PUBLIC_BASE_PATH.replace(/^\//, "").replace(/\/$/, "")}`
  : "";

const isBrowser = typeof window !== "undefined";

// --- client-side API ---
// When reverse proxy is active (BASE_PATH is set), direct fetch to same-origin relative BASE_PATH
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ?? (BASE_PATH ? BASE_PATH : "");

// --- realtime / websocket ---
// Dynamically resolve protocol (ws/wss) and host at runtime in browser for maximum cloud-native portability.
export const WS_BASE_URL =
  process.env.NEXT_PUBLIC_WS_URL ??
  (isBrowser
    ? `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}${BASE_PATH ? BASE_PATH : ""}/api/v1/realtime/ws`
    : "ws://localhost:8080/api/v1/realtime/ws");

// --- IdP (Keycloak OIDC default) ---
export const IDP_PROVIDER = process.env.NEXT_PUBLIC_IDP_PROVIDER ?? "keycloak";

export const OIDC_ISSUER_URL = stripTrailingSlash(
  process.env.NEXT_PUBLIC_OIDC_ISSUER_URL ?? "",
);

export const OIDC_AUTH_URL =
  process.env.NEXT_PUBLIC_OIDC_AUTH_URL ??
  `${OIDC_ISSUER_URL}/protocol/openid-connect/auth`;

// Dynamically construct OIDC redirect callback URL based on current origin to support multiple environments seamlessly.
export const OIDC_REDIRECT_URI =
  process.env.NEXT_PUBLIC_OIDC_REDIRECT_URI ??
  (isBrowser
    ? `${window.location.origin}${BASE_PATH}/auth/callback`
    : "http://localhost:3000/auth/callback");

// --- server-only (next.config / route handlers / tests) ---
// next.config.ts 의 rewrites 가 사용. docker 에서는 compose env 로 override.
export const BACKEND_API_URL_SERVER =
  process.env.BACKEND_API_URL ?? "http://localhost:8080";

// e2e/global-setup + fixtures 는 DEVHUB_KEYCLOAK_ADMIN_URL env 를 직접 참조한다
// (sprint -m design — docs/planning/e2e_keycloak_migration.md, 별도 carve).

// Keycloak Admin Console URL.
// caller (admin users page) 가 빈 string 받으면 <a href=""> 가 되어 same-page reload 위험.
// Stage 3 hotfix (PR #246 review): 환경변수 미설정 시 명시 null 반환 + caller 측 fallback 분기.
export const getKCAdminConsoleUrl = (): string | null => {
  if (process.env.NEXT_PUBLIC_KC_ADMIN_URL) return process.env.NEXT_PUBLIC_KC_ADMIN_URL;
  if (!OIDC_ISSUER_URL) return null;
  try {
    const url = new URL(OIDC_ISSUER_URL);
    const realmMatch = url.pathname.match(/\/realms\/([^/]+)/);
    const realm = realmMatch ? realmMatch[1] : 'master';
    return `${url.origin}/admin/${realm}/console/`;
  } catch {
    // Invalid OIDC_ISSUER_URL — no fallback (caller hides the link).
    return null;
  }
};
