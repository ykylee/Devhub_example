"use client";

import { use } from "react";
import { ProjectKPISection } from "@/domain/platform-lifecycle/view/ProjectKPISection";
import { KpiTestDetailPage } from "@/shared/ui-foundation/components/KpiTestDetailPage";

// 이슈 3 (KPI/test 카드 → 별도 페이지 drill-down) — projects/[id]/kpi page
// (2026-06-17 정공법).

export default function ProjectKpiPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  return (
    <KpiTestDetailPage entityType="project" kind="kpi" entityId={id}>
      <ProjectKPISection projectId={id} />
    </KpiTestDetailPage>
  );
}
