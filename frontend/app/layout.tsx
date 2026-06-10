import type { Metadata } from "next";
import "./globals.css";
import { ToastContainer } from "@/shared/ui-foundation/components/Toast";
import { LogoutOverlay } from "@/shared/ui-foundation/components/LogoutOverlay";
import { branding } from "@/shared/config/branding";
import { BrandingProviderClient } from "@/lib/branding-provider-client";

// 2026-06-10 결정 — 사외/사내 2-tier deploy 명칭 정책
// (`docs/governance/worker_division.md` §6):
//   - 사외 (GitHub main) deploy: default "DevHub Example" / "DevHub" / tagline
//   - 사내 (사내 SCM) deploy: DEVHUB_APP_NAME / DEVHUB_APP_SHORT_NAME /
//     DEVHUB_BRAND_TAGLINE env var override 로 다른 명칭 주입
//   - 코드 default = 사외 명칭 (사내 build 시 override)
//   - 본 `<title>` 은 SSR metadata 단계에서 server-side `branding` 모듈의 값을
//     직접 inline. 사내 deploy 시 자동으로 다른 명칭 노출. Client-side
//     branding (useBranding hook) 은 page heading / sidebar 등 CSR 단계용.
export const metadata: Metadata = {
  title: `${branding.appShortName} - ${branding.brandTagline}`,
  description: `${branding.appName} — role-prioritized entry team hub for modern engineering teams.`,
};

// Paint 전에 저장된 theme 을 html element 에 반영해 light→dark FOUC 를 회피.
// localStorage 키는 Header dropdown 의 theme toggle 과 공유한다 ("devhub-theme").
const themeBootstrap = `try{var t=localStorage.getItem("devhub-theme");if(t==="dark")document.documentElement.classList.add("theme-dark");}catch(e){}`;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="h-full antialiased">
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeBootstrap }} />
      </head>
      <body className="min-h-full flex flex-col font-sans">
        <BrandingProviderClient
          initialBranding={{
            appName: branding.appName,
            appShortName: branding.appShortName,
            brandTagline: branding.brandTagline,
          }}
        >
          {children}
          <LogoutOverlay />
          <ToastContainer />
        </BrandingProviderClient>
      </body>
    </html>
  );
}
