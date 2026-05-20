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

const DEV_FALLBACK_OIDC_AUTH_URL =
  "http://localhost:8180/devhub/auth/keycloak/realms/devhub/protocol/openid-connect/auth";
const DEV_FALLBACK_OIDC_ISSUER_URL =
  "http://localhost:8180/devhub/auth/keycloak/realms/devhub";

export async function GET(request: NextRequest) {
  // Use Next.js parsed origin to avoid trusting spoofable forwarded headers.
  const origin = request.nextUrl.origin;
  const runtimeEnv = process.env;
  const isProduction = runtimeEnv.NODE_ENV === "production";

  const oidcAuthURL =
    runtimeEnv["OIDC_AUTH_URL"] ??
    runtimeEnv["NEXT_PUBLIC_OIDC_AUTH_URL"] ??
    (isProduction ? undefined : DEV_FALLBACK_OIDC_AUTH_URL);
  const oidcRedirectURI =
    runtimeEnv["OIDC_REDIRECT_URI"] ??
    runtimeEnv["NEXT_PUBLIC_OIDC_REDIRECT_URI"] ??
    `${trimTrailingSlash(origin)}/auth/callback`;
  const oidcIssuerURL =
    runtimeEnv["OIDC_ISSUER_URL"] ??
    runtimeEnv["NEXT_PUBLIC_OIDC_ISSUER_URL"] ??
    (isProduction ? undefined : DEV_FALLBACK_OIDC_ISSUER_URL);

  // ADR-0019 §4.2 환경설정 계약 + 단일 포트 컨셉 (ADR-0018) 정합:
  // production 에서 OIDC issuer / auth URL 이 미설정이면 fail-fast.
  // localhost fallback 이 그대로 응답되면 외부 브라우저가 localhost:8180 으로 redirect 시도 →
  // "동작 안 함" 컨셉 위배. docs/reports/2026-05-20-network-docker-single-port-review.md §4.3 정정.
  if (isProduction && (!oidcAuthURL || !oidcIssuerURL)) {
    return NextResponse.json(
      {
        error: "runtime_config_misconfigured",
        message:
          "Production OIDC env vars missing (OIDC_AUTH_URL / OIDC_ISSUER_URL). " +
          "운영 환경에서 외부 Keycloak issuer/auth URL 을 환경변수로 주입해야 한다.",
      },
      { status: 500, headers: { "Cache-Control": "no-store" } },
    );
  }

  const payload: RuntimeConfigResponse = {
    oidc_auth_url: oidcAuthURL ?? DEV_FALLBACK_OIDC_AUTH_URL,
    oidc_redirect_uri: oidcRedirectURI,
    oidc_issuer_url: oidcIssuerURL ?? DEV_FALLBACK_OIDC_ISSUER_URL,
  };

  return NextResponse.json(payload, {
    headers: {
      // 환경변수 전환 직후에도 stale 값이 남지 않도록 no-store.
      "Cache-Control": "no-store",
    },
  });
}
