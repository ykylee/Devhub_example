"use client";

// Client-side wrapper for BrandingProvider — root layout (server component) 에서
// import 가능한 client boundary. SSR 단계의 initialBranding 은 server-side
// `branding` 모듈에서 server component 가 props 로 주입.

import type { ReactNode } from "react";
import {
  BrandingProvider,
  type ClientBranding,
} from "@/lib/branding-context";

interface BrandingProviderClientProps {
  children: ReactNode;
  initialBranding: ClientBranding;
}

export function BrandingProviderClient({
  children,
  initialBranding,
}: BrandingProviderClientProps) {
  return (
    <BrandingProvider initialBranding={initialBranding}>
      {children}
    </BrandingProvider>
  );
}
