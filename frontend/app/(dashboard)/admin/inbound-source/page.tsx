"use client";

import { useState } from "react";
import { InboundSourceManager } from "@/components/admin/inbound-source-config";
import type { PlatformInboundSourceView, InboundSourceType, InboundSourceRoutingConfig } from "@/domain/integration-registry/schema/integration.types";
import { Webhook } from "lucide-react";

/**
 * /admin/inbound-source page — X-2 multi-provider webhook 운영 UI (RM-M4-07 §3.5 X-2).
 *
 * system_admin 이 platform 별 inbound_source_type + inbound_source_config 를 관리.
 * 4 widget (InboundSourceTypeSelector + InboundSourceConfigEditor + PatternPreview +
 * InboundSourceManager 통합 view).
 *
 * 1차 PR #586 의 auto_route.go 의 InboundSourceRoutingConfig + 2~3차 PR 의
 * WebhookAdapter (Gitea/Jira/Generic) 와 1:1 정합.
 *
 * 권한: system_admin 일임 (Sidebar 의 isSystemAdmin(actor?.role) gate 자동).
 */
export default function InboundSourcePage() {
  // Mock platform data (실제로는 `GET /api/v1/platforms?inbound_source=true` 호출).
  // 본 sprint 의 e2e spec (admin-x2.spec.ts) 가 mock data 검증.
  const [platforms, setPlatforms] = useState<PlatformInboundSourceView[]>([
    {
      platform_id: "plat-gitea-1",
      platform_key: "gitea-main",
      platform_name: "Gitea Main",
      inbound_source_type: "gitea",
      inbound_source_config: {},
      updated_at: "2026-06-13T10:00:00Z",
    },
    {
      platform_id: "plat-jira-1",
      platform_key: "jira-prod",
      platform_name: "Jira Production",
      inbound_source_type: "jira",
      inbound_source_config: {},
      updated_at: "2026-06-13T10:00:00Z",
    },
    {
      platform_id: "plat-other-1",
      platform_key: "custom-ci",
      platform_name: "Custom CI",
      inbound_source_type: "other",
      inbound_source_config: { custom_external_ref_pattern: "^CUSTOM-\\d+$" },
      updated_at: "2026-06-13T10:00:00Z",
    },
  ]);

  const handleSave = async (
    platformId: string,
    next: { inbound_source_type: InboundSourceType; inbound_source_config: InboundSourceRoutingConfig },
  ) => {
    // 실제: `PATCH /api/v1/platforms/${platformId}` body { inbound_source_type, inbound_source_config } 호출.
    setPlatforms((prev) =>
      prev.map((p) =>
        p.platform_id === platformId
          ? { ...p, inbound_source_type: next.inbound_source_type, inbound_source_config: next.inbound_source_config, updated_at: new Date().toISOString() }
          : p,
      ),
    );
  };

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 bg-muted/30 rounded-xl flex items-center justify-center border border-border shadow-inner">
          <Webhook className="w-5 h-5 text-primary" />
        </div>
        <div>
          <h1 className="text-xl font-black text-foreground">Inbound Source (Multi-Provider Webhook)</h1>
          <p className="text-xs text-muted-foreground">
            X-2 · Gitea / Jira / Custom provider · webhook envelope dispatcher
          </p>
        </div>
      </div>

      <InboundSourceManager platforms={platforms} onSave={handleSave} />
    </div>
  );
}
