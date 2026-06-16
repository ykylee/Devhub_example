"use client";

import { useState } from "react";
import { ChevronRight, Info, Server, Settings, Zap, AlertCircle } from "lucide-react";
import { cn } from "@/shared/utils";
import { motion } from "framer-motion";

// DomainPicker — Sprint D (kpi-tests-per-domain-scope.md §6.4)
//
// KPI / Tests 글로벌 페이지(/kpis, /tests) 의 상단에 위치하는 도메인 picker.
// scope (Platform/Project/Repository) 선택 → 해당 scope 의 entity list 표시 →
// entity 클릭 시 해당 entity 의 sub-section 이 있는 도메인 페이지로 redirect.
//
// Sprint A (Repository) 만 sub-section 활성화. Sprint B (Project) + Sprint C
// (Platform) 미구현 → 미구현 scope 의 entity 클릭 시 '준비 중' 안내.
//
// 정공법: page 가 entity list fetch (projectService.listPlatforms/getPlatforms
// + repositoryService.listRepositories) 후 props 로 주입. picker 는 순수 UI.

export type DomainScope = "platform" | "project" | "repository";

export interface DomainEntity {
  id: string;
  name: string;
  description?: string;
}

interface DomainPickerProps {
  /** 현재 페이지의 기본 scope. 페이지 mount 시 미리 선택됨. */
  defaultScope?: DomainScope;
  /** Page 가 fetch 한 entity list (scope 별로 분리). */
  platforms?: DomainEntity[];
  projects?: DomainEntity[];
  repositories?: DomainEntity[];
  /** Page 가 fetch 한 로딩/에러 상태 — picker 가 본문 표시. */
  loading?: boolean;
  error?: string | null;
}

const SCOPE_OPTIONS: Array<{ value: DomainScope; label: string; icon: typeof Zap; sublabel: string; ready: boolean }> = [
  { value: "platform", label: "Platform", icon: Zap, sublabel: "sub-project rollup", ready: false },
  { value: "project", label: "Project", icon: Settings, sublabel: "weighted repository rollup", ready: true },
  { value: "repository", label: "Repository", icon: Server, sublabel: "raw metric (weight=1)", ready: true },
];

export function DomainPicker({
  defaultScope = "repository",
  platforms = [],
  projects = [],
  repositories = [],
  loading = false,
  error = null,
}: DomainPickerProps) {
  const [scope, setScope] = useState<DomainScope>(defaultScope);

  const entities: DomainEntity[] =
    scope === "platform" ? platforms : scope === "project" ? projects : repositories;

  const scopeMeta = SCOPE_OPTIONS.find((o) => o.value === scope)!;
  const Icon = scopeMeta.icon;

  return (
    <section
      aria-label="Domain picker"
      data-testid="domain-picker"
      className="glass border border-border rounded-2xl p-5 space-y-4"
    >
      <header className="space-y-1">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <Icon className="w-5 h-5" />
          {scopeMeta.label} Analytics
        </h2>
        <p className="text-sm text-muted-foreground">
          {scopeMeta.sublabel} · Sprint {scopeMeta.ready ? "A" : scope === "project" ? "B (예정)" : "C (예정)"}
        </p>
      </header>

      {/* scope tabs */}
      <div role="tablist" aria-label="Domain scope" className="flex flex-wrap gap-2">
        {SCOPE_OPTIONS.map((opt) => {
          const OptIcon = opt.icon;
          const active = opt.value === scope;
          return (
            <button
              key={opt.value}
              type="button"
              role="tab"
              aria-selected={active}
              data-testid={`domain-picker-scope-${opt.value}`}
              onClick={() => setScope(opt.value)}
              className={cn(
                "inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors",
                active
                  ? "bg-primary/12 text-primary border border-primary/30"
                  : "bg-muted/30 text-muted-foreground hover:bg-muted/60 border border-transparent",
              )}
            >
              <OptIcon className="w-4 h-4" />
              {opt.label}
              {!opt.ready && (
                <span className="text-[10px] uppercase tracking-wide text-muted-foreground/70 ml-1">soon</span>
              )}
            </button>
          );
        })}
      </div>

      {/* entity list */}
      <div data-testid={`domain-picker-entity-list-${scope}`} className="space-y-1.5">
        {loading && (
          <p className="text-sm text-muted-foreground px-2 py-3">Loading {scope} entities…</p>
        )}
        {!loading && error && (
          <p className="text-sm text-red-600 dark:text-red-300 flex items-center gap-1.5 px-2 py-3">
            <AlertCircle className="w-4 h-4" /> {error}
          </p>
        )}
        {!loading && !error && entities.length === 0 && (
          <p className="text-sm text-muted-foreground px-2 py-3">
            No {scope} entities available.
          </p>
        )}
        {!loading && !error && entities.map((entity) => (
          <EntityLink key={entity.id} scope={scope} entity={entity} ready={scopeMeta.ready} />
        ))}
      </div>

      {/* helper hint */}
      <footer className="text-xs text-muted-foreground flex items-start gap-1.5 pt-2 border-t border-border/40">
        <Info className="w-3.5 h-3.5 mt-0.5 shrink-0" aria-hidden />
        <span>
          도메인 상세 페이지의 sub-section 이 1차 진입점입니다. 글로벌 페이지는
          cross-reference picker 역할입니다.
        </span>
      </footer>
    </section>
  );
}

function EntityLink({ scope, entity, ready }: { scope: DomainScope; entity: DomainEntity; ready: boolean }) {
  const href =
    scope === "platform"
      ? `/platforms/${encodeURIComponent(entity.id)}`
      : scope === "project"
        ? `/projects/${encodeURIComponent(entity.id)}`
        : `/repositories/${entity.id}`;

  return (
    <motion.a
      href={href}
      data-testid={`domain-picker-entity-${entity.id}`}
      whileHover={{ x: 2 }}
      className={cn(
        "flex items-center justify-between gap-2 px-3 py-2 rounded-lg border border-border/40 bg-background/30",
        "hover:bg-muted/40 transition-colors",
      )}
    >
      <div className="min-w-0 flex-1">
        <div className="text-sm font-medium truncate">{entity.name}</div>
        {entity.description && (
          <div className="text-xs text-muted-foreground truncate">{entity.description}</div>
        )}
      </div>
      <div className="flex items-center gap-1.5 shrink-0">
        {!ready && (
          <span className="text-[10px] uppercase tracking-wide text-muted-foreground/70 px-1.5 py-0.5 rounded border border-border/40">
            sub-section 예정
          </span>
        )}
        <ChevronRight className="w-4 h-4 text-muted-foreground" />
      </div>
    </motion.a>
  );
}
