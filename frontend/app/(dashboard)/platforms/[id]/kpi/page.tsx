"use client";

import { use } from "react";
import { PlatformKPISection } from "@/domain/platform-lifecycle/view/PlatformKPISection";
import { KpiTestDetailPage } from "@/shared/ui-foundation/components/KpiTestDetailPage";

// 이슈 3 (KPI/test 카드 → 별도 페이지 drill-down) — platforms/[id]/kpi page
// (2026-06-17 정공법). 기존 inline 사용 (platforms/[id]/page.tsx 의 PlatformKPISection)
// 의 "자세히 보기" link 가 본 page 로 navigation. 본 page 는 Sprint C 의
// PlatformKPISection component 를 단일 column 으로 rendering + back link.

export default function PlatformKpiPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  return (
    <KpiTestDetailPage entityType="platform" kind="kpi" entityId={id}>
      <PlatformKPISection platformId={id} />
    </KpiTestDetailPage>
  );
}
