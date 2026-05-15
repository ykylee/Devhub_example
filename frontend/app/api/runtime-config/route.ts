import { NextRequest, NextResponse } from "next/server";

export const dynamic = "force-dynamic";

interface RuntimeConfigResponse {
  oidc_auth_url: string;
  oidc_redirect_uri: string;
}

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, "");
}

function requestOrigin(request: NextRequest): string {
  const forwardedProto = request.headers.get("x-forwarded-proto");
  const forwardedHost = request.headers.get("x-forwarded-host");
  if (forwardedProto && forwardedHost) {
    return `${forwardedProto}://${forwardedHost}`;
  }
  return request.nextUrl.origin;
}

export async function GET(request: NextRequest) {
  const origin = requestOrigin(request);
  const runtimeEnv = process.env;
  const oidcAuthURL =
    runtimeEnv["OIDC_AUTH_URL"] ??
    runtimeEnv["NEXT_PUBLIC_OIDC_AUTH_URL"] ??
    "http://localhost:4444/oauth2/auth";
  const oidcRedirectURI =
    runtimeEnv["OIDC_REDIRECT_URI"] ??
    runtimeEnv["NEXT_PUBLIC_OIDC_REDIRECT_URI"] ??
    `${trimTrailingSlash(origin)}/auth/callback`;

  const payload: RuntimeConfigResponse = {
    oidc_auth_url: oidcAuthURL,
    oidc_redirect_uri: oidcRedirectURI,
  };

  return NextResponse.json(payload, {
    headers: {
      // 환경변수 전환 직후에도 stale 값이 남지 않도록 no-store.
      "Cache-Control": "no-store",
    },
  });
}
