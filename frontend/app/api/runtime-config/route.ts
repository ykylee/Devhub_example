import { NextRequest, NextResponse } from "next/server";

export const dynamic = "force-dynamic";

interface RuntimeConfigResponse {
  oidc_auth_url: string;
  oidc_redirect_uri: string;
  oidc_issuer_url: string;
}

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, "");
}

function normalizeBasePath(raw: string | undefined): string {
  if (!raw || raw.trim() === "") return "";
  return `/${raw.replace(/^\//, "").replace(/\/$/, "")}`;
}

const DEV_FALLBACK_OIDC_ISSUER_URL =
  "http://localhost:8180/devhub/auth/keycloak/realms/devhub";

// 표준 Keycloak OIDC authorize path — issuer URL 이 주어지면 본 suffix 로 derive.
const OIDC_AUTH_PATH_SUFFIX = "/protocol/openid-connect/auth";

export async function GET(request: NextRequest) {
  // Use Next.js parsed origin to avoid trusting spoofable forwarded headers.
  const origin = request.nextUrl.origin;
  const runtimeEnv = process.env;
  const isProduction = runtimeEnv.NODE_ENV === "production";
  const basePath = normalizeBasePath(runtimeEnv["NEXT_PUBLIC_BASE_PATH"]);
  const publicBaseURL = runtimeEnv["DEVHUB_PUBLIC_BASE_URL"]?.trim() ?? "";
  const redirectPath = `${basePath}/auth/callback`;

  const issuerEnv =
    runtimeEnv["OIDC_ISSUER_URL"] ??
    runtimeEnv["NEXT_PUBLIC_OIDC_ISSUER_URL"];
  const authURLEnv =
    runtimeEnv["OIDC_AUTH_URL"] ??
    runtimeEnv["NEXT_PUBLIC_OIDC_AUTH_URL"];

  // ADR-0019 §4.2 환경설정 계약 + 단일 포트 컨셉 (ADR-0018) 정합:
  // production 에서 issuer URL 이 미설정이면 fail-fast. localhost fallback 이 그대로
  // 응답되면 외부 브라우저가 localhost:8180 으로 redirect 시도 → "동작 안 함" 컨셉 위배.
  // (docs/reports/2026-05-20-network-docker-single-port-review.md §4.3 정정 + codex P1).
  //
  // OIDC_AUTH_URL 은 strict 요구 아니다 — issuer URL 에서 표준 path 로 derive 가능
  // (compose 도 OIDC_AUTH_URL=${OIDC_AUTH_URL:-} 으로 optional). 강제하면 healthcheck
  // /api/runtime-config 가 500 → frontend unhealthy → nginx depends_on 체인 차단 회귀.
  if (isProduction && (!issuerEnv || issuerEnv.trim() === "")) {
    return NextResponse.json(
      {
        error: "runtime_config_misconfigured",
        message:
          "Production OIDC_ISSUER_URL 미설정. 외부 Keycloak issuer URL 을 환경변수로 주입해야 한다.",
      },
      { status: 500, headers: { "Cache-Control": "no-store" } },
    );
  }

  const oidcIssuerURL = issuerEnv && issuerEnv.trim() !== ""
    ? issuerEnv
    : DEV_FALLBACK_OIDC_ISSUER_URL;

  // auth URL 우선순위: env 명시값 → issuer-derived → dev fallback (개발 환경 only).
  const oidcAuthURL = authURLEnv && authURLEnv.trim() !== ""
    ? authURLEnv
    : `${trimTrailingSlash(oidcIssuerURL)}${OIDC_AUTH_PATH_SUFFIX}`;

  const originForRedirect =
    publicBaseURL !== "" ? trimTrailingSlash(publicBaseURL) : trimTrailingSlash(origin);
  const oidcRedirectURI =
    runtimeEnv["OIDC_REDIRECT_URI"] ??
    runtimeEnv["NEXT_PUBLIC_OIDC_REDIRECT_URI"] ??
    `${originForRedirect}${redirectPath}`;

  const payload: RuntimeConfigResponse = {
    oidc_auth_url: oidcAuthURL,
    oidc_redirect_uri: oidcRedirectURI,
    oidc_issuer_url: oidcIssuerURL,
  };

  return NextResponse.json(payload, {
    headers: {
      // 환경변수 전환 직후에도 stale 값이 남지 않도록 no-store.
      "Cache-Control": "no-store",
    },
  });
}
