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

export async function GET(request: NextRequest) {
  // Use Next.js parsed origin to avoid trusting spoofable forwarded headers.
  const origin = request.nextUrl.origin;
  const runtimeEnv = process.env;
  const oidcAuthURL =
    runtimeEnv["OIDC_AUTH_URL"] ??
    runtimeEnv["NEXT_PUBLIC_OIDC_AUTH_URL"] ??
    "http://localhost:8180/devhub/auth/keycloak/realms/devhub/protocol/openid-connect/auth";
  const oidcRedirectURI =
    runtimeEnv["OIDC_REDIRECT_URI"] ??
    runtimeEnv["NEXT_PUBLIC_OIDC_REDIRECT_URI"] ??
    `${trimTrailingSlash(origin)}/auth/callback`;
  const oidcIssuerURL =
    runtimeEnv["OIDC_ISSUER_URL"] ??
    runtimeEnv["NEXT_PUBLIC_OIDC_ISSUER_URL"] ??
    "http://localhost:8180/devhub/auth/keycloak/realms/devhub";

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
