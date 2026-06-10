// App branding — 사외/사내 deploy tier 별 다른 명칭.
//
// 정책 (`docs/governance/worker_division.md` §6 + 2026-06-10 결정):
//   - 사외 (GitHub main) deploy: default "DevHub Example" / "DevHub" / tagline
//   - 사내 (사내 SCM) deploy: env var override 로 다른 명칭 주입
//     예: DEVHUB_APP_NAME="내부 포털" / DEVHUB_APP_SHORT_NAME="내부" / DEVHUB_BRAND_TAGLINE="..."
//   - 코드 default = 사외 명칭 (사내 build 시 override)
//
// Tier 분류: **공용** (사외 + 사내 양쪽 build 가 같은 모듈 사용).
//   - 본 모듈은 pure env reader. 사내 한정 정보 (host, secret, IdP) 미포함.
//   - 사외 build 시 default 사외 명칭 노출, 사내 build 시 env var 로 다른 명칭.
//
// 사용:
//   - server-side (metadata export, runtime-config API, server components):
//     `import { branding } from "@/shared/config/branding"` — SSR inline 가능.
//   - client-side (브라우저 component): 본 모듈 직접 import 시 NEXT_PUBLIC_* prefix 가
//     없어 inline 되지 않으므로 **반드시** `useBranding()` hook (lib/branding-context.tsx)
//     으로 `/api/runtime-config` 응답의 branding 필드를 읽어야 한다.
//
// Branding default 변경 시 `frontend/app/api/runtime-config/route.ts` 의 default
// 도 함께 갱신할 것 (양쪽 default 가 drift 되지 않도록).

export const branding = {
  appName:
    process.env.DEVHUB_APP_NAME?.trim() && process.env.DEVHUB_APP_NAME.trim() !== ""
      ? process.env.DEVHUB_APP_NAME.trim()
      : "DevHub Example",
  appShortName:
    process.env.DEVHUB_APP_SHORT_NAME?.trim() &&
    process.env.DEVHUB_APP_SHORT_NAME.trim() !== ""
      ? process.env.DEVHUB_APP_SHORT_NAME.trim()
      : "DevHub",
  brandTagline:
    process.env.DEVHUB_BRAND_TAGLINE?.trim() &&
    process.env.DEVHUB_BRAND_TAGLINE.trim() !== ""
      ? process.env.DEVHUB_BRAND_TAGLINE.trim()
      : "Role-Prioritized Entry Team Hub",
} as const;
