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

// --- client-side API ---
// next.config.ts 의 rewrites 로 `/api/*` 를 BACKEND_API_URL 로 프록시하므로
// 클라이언트는 기본적으로 relative path 를 쓴다 (CORS 회피).
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "";

// --- realtime / websocket ---
export const WS_BASE_URL =
  process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080/api/v1/realtime/ws";

// --- IdP (Kratos / Hydra) ---
export const KRATOS_PUBLIC_URL = stripTrailingSlash(
  process.env.NEXT_PUBLIC_KRATOS_PUBLIC_URL ?? "http://localhost:4433",
);

export const OIDC_AUTH_URL =
  process.env.NEXT_PUBLIC_OIDC_AUTH_URL ?? "http://localhost:4444/oauth2/auth";

export const HYDRA_PUBLIC_BASE = OIDC_AUTH_URL.replace(/\/oauth2\/auth\/?$/, "");

export const OIDC_REDIRECT_URI =
  process.env.NEXT_PUBLIC_OIDC_REDIRECT_URI ?? "http://localhost:3000/auth/callback";

// --- server-only (next.config / route handlers / tests) ---
// next.config.ts 의 rewrites 가 사용. docker 에서는 compose env 로 override.
export const BACKEND_API_URL_SERVER =
  process.env.BACKEND_API_URL ?? "http://localhost:8080";

// e2e/global-setup 에서 사용. native 기본값.
export const KRATOS_ADMIN_URL_SERVER = stripTrailingSlash(
  process.env.KRATOS_ADMIN_URL ?? "http://localhost:4434",
);
