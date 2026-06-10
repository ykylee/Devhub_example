"use client";

// Branding client-side context + hook.
//
// server-only `shared/config/branding.ts` 는 process.env 를 직접 읽으므로
// client component (브라우저) 에서 직접 import 하면 `process.env` 가 build 시점에
// inline 되지 않아 (NEXT_PUBLIC_ prefix 없음) undefined 가 된다.
//
// 본 모듈은 `/api/runtime-config` 응답의 branding 필드를 client context 로 주입.
//   - Provider: `<BrandingProvider>` 가 root layout (또는 dashboard layout) 의
//     client component 안에서 fetch + children 에 주입.
//   - Hook: `useBranding()` 으로 `{ appName, appShortName, brandTagline }` 반환.
//
// SSR 단계에서는 branding 값이 비어있다가 hydration 후 채워진다. 대부분의
// user-facing text (page heading, sidebar brand) 는 hydration 후 렌더되며
// 자연스럽게 새 branding 으로 갱신. 단, metadata (HTML <title>) 는 SSR 단계에
// inline 되어야 하므로 server-only `branding.ts` 가 별도 책임진다 (app/layout.tsx).

import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

export interface ClientBranding {
  appName: string;
  appShortName: string;
  brandTagline: string;
}

const BrandingContext = createContext<ClientBranding | null>(null);

interface RuntimeConfigResponse {
  app_name?: string;
  app_short_name?: string;
  brand_tagline?: string;
}

const FALLBACK_BRANDING: ClientBranding = {
  appName: "DevHub Example",
  appShortName: "DevHub",
  brandTagline: "Role-Prioritized Entry Team Hub",
};

interface BrandingProviderProps {
  children: ReactNode;
  // SSR 단계에서 server-side branding 을 주입받아 첫 paint 까지 빈 string 회피.
  initialBranding?: ClientBranding;
}

export function BrandingProvider({ children, initialBranding }: BrandingProviderProps) {
  // SSR 첫 paint 시점에도 server-side default 표시를 위해 initialBranding 사용.
  // client hydration 후엔 /api/runtime-config 의 최신 값으로 갱신.
  const [branding, setBranding] = useState<ClientBranding>(
    initialBranding ?? FALLBACK_BRANDING,
  );

  useEffect(() => {
    // client-side fetch: runtime-config 가 이미 캐시된 최신 branding 제공.
    // 실패 시 SSR initialBranding 유지 (graceful fallback).
    let cancelled = false;
    fetch("/api/runtime-config", { cache: "no-store" })
      .then((res) => (res.ok ? res.json() : null))
      .then((data: RuntimeConfigResponse | null) => {
        if (cancelled || !data) return;
        if (
          typeof data.app_name === "string" &&
          typeof data.app_short_name === "string" &&
          typeof data.brand_tagline === "string"
        ) {
          setBranding({
            appName: data.app_name,
            appShortName: data.app_short_name,
            brandTagline: data.brand_tagline,
          });
        }
      })
      .catch(() => {
        // network/parse 실패 → 기존 SSR 값 유지.
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <BrandingContext.Provider value={branding}>{children}</BrandingContext.Provider>
  );
}

export function useBranding(): ClientBranding {
  const ctx = useContext(BrandingContext);
  if (ctx) return ctx;
  return FALLBACK_BRANDING;
}
