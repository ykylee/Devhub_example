"use client";

import { ReactNode } from "react";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";

// KpiTestDetailPage — 이슈 3 (KPI/test 카드 → 별도 페이지 drill-down) 의 공통 wrapper
// (2026-06-17 fix). 6 page (platforms/projects/repositories × kpi/test-results) 가 본
// wrapper 사용 — entity type (platform | project | repository) + kind (kpi | tests) +
// entity id 로 header + back link 일관.
//
// Sprint A/B/C 의 KPI/test endpoint 5 개 (repositories/:id/kpi + /test-results,
// projects/:id/kpi + /test-results, platforms/:id/kpi + /test-results) 의
// 정공법 응답을 단일 page 구조로 drill-down. 기존 inline 사용 (platforms/[id]/page.tsx,
// projects/[id]/page.tsx, ManagerView 의 RepositoryKPI/Tests section) 의 section 위치에
// "자세히 보기" link 가 본 wrapper 의 page 로 navigation.

interface KpiTestDetailPageProps {
  /** "platform" | "project" | "repository" — entity type */
  entityType: "platform" | "project" | "repository";
  /** "kpi" | "tests" */
  kind: "kpi" | "tests";
  /** entity id (string) */
  entityId: string;
  /** entity 표시명 (header 에 표시) */
  entityName?: string;
  /** 본문 — 기존 component (PlatformKPISection / PlatformTestsSection / ...) */
  children: ReactNode;
}

export function KpiTestDetailPage({ entityType, kind, entityId, entityName, children }: KpiTestDetailPageProps) {
  // entity 상세 page 로 back (예: /platforms/[id], /projects/[id], /repositories/[id])
  const basePath = entityType === "platform" ? "platforms" : entityType === "project" ? "projects" : "repositories";
  const kindLabel = kind === "kpi" ? "KPI" : "Test Results";
  return (
    <div className="space-y-6 pb-20 px-4 md:px-8">
      <div className="flex items-center gap-3">
        <Link
          href={`/${basePath}/${entityId}`}
          data-testid={`kpi-test-detail-back-${entityType}-${kind}`}
          className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
        >
          <ArrowLeft className="w-3.5 h-3.5" />
          Back to {entityType} detail
        </Link>
      </div>
      <header className="space-y-1">
        <h1 className="text-2xl font-black tracking-tight text-foreground dark:text-primary-foreground">
          {entityName ? `${entityName} — ` : ""}{kindLabel}
        </h1>
        <p className="text-xs text-muted-foreground/70">
          Drill-down 상세 페이지 (2026-06-17 정공법). 본 page 의 content 는 Sprint A/B/C 의
          KPI/test endpoint 의 raw response 를 재사용.
        </p>
      </header>
      {children}
    </div>
  );
}
