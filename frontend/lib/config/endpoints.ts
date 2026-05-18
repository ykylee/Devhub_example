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
const BASE_PATH = process.env.NEXT_PUBLIC_BASE_PATH
  ? `/${process.env.NEXT_PUBLIC_BASE_PATH.replace(/^\//, "").replace(/\/$/, "")}`
  : "";

const isBrowser = typeof window !== "undefined";

// --- client-side API ---
// When reverse proxy is active (BASE_PATH is set), direct fetch to same-origin relative '/devhub/api'
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ?? (BASE_PATH ? `${BASE_PATH}/api` : "");

// --- realtime / websocket ---
// Dynamically resolve protocol (ws/wss) and host at runtime in browser for maximum cloud-native portability.
export const WS_BASE_URL =
  process.env.NEXT_PUBLIC_WS_URL ??
  (isBrowser
    ? `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}${BASE_PATH ? BASE_PATH : ""}/api/v1/realtime/ws`
    : "ws://localhost:8080/api/v1/realtime/ws");

// --- IdP (Kratos / Hydra) ---
export const KRATOS_PUBLIC_URL = stripTrailingSlash(
  process.env.NEXT_PUBLIC_KRATOS_PUBLIC_URL ??
    (BASE_PATH ? `${BASE_PATH}/auth/kratos` : "http://localhost:4433"),
);

export const OIDC_AUTH_URL =
  process.env.NEXT_PUBLIC_OIDC_AUTH_URL ??
  (BASE_PATH
    ? `${BASE_PATH}/auth/hydra/oauth2/auth`
    : "http://localhost:4444/oauth2/auth");

export const HYDRA_PUBLIC_BASE = OIDC_AUTH_URL.replace(/\/oauth2\/auth\/?$/, "");

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

// e2e/global-setup 에서 사용. native 기본값.
export const KRATOS_ADMIN_URL_SERVER = stripTrailingSlash(
  process.env.KRATOS_ADMIN_URL ?? "http://localhost:4434",
);
