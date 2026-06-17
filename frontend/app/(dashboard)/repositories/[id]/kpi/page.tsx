"use client";

import { use } from "react";
import { RepositoryKPISection } from "@/domain/repository-integration/view/RepositoryKPISection";
import { KpiTestDetailPage } from "@/shared/ui-foundation/components/KpiTestDetailPage";

// 이슈 3 (KPI/test 카드 → 별도 페이지 drill-down) — repositories/[id]/kpi page
// (2026-06-17 정공법). 기존 ManagerView 의 RepositoryKPISection 사용처의
// "자세히 보기" link 가 본 page 로 navigation.

export default function RepositoryKpiPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  return (
    <KpiTestDetailPage entityType="repository" kind="kpi" entityId={id}>
      <RepositoryKPISection repoId={Number(id)} />
    </KpiTestDetailPage>
  );
}
