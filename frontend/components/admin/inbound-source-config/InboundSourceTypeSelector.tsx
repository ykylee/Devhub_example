"use client";

import type { InboundSourceType } from "@/domain/integration-registry/schema/integration.types";
import { AlertCircle, Webhook } from "lucide-react";

const SUPPORTED_TYPES: { value: InboundSourceType; label: string; hint: string }[] = [
  {
    value: "",
    label: "Disabled (manual routing)",
    hint: "inbound_source 자동 routing 비활성화. voc 단계 유지 (현행 동작 보존).",
  },
  {
    value: "gitea",
    label: "Gitea / Forgejo / Gogs",
    hint: "GITEA-<number> external_ref + X-Gitea-* header + provider_type='scm' (default adapter).",
  },
  {
    value: "jira",
    label: "Jira / Atlassian",
    hint: "<PROJECT_KEY>-<number> external_ref + X-Atlassian-* header + provider_type='alm'.",
  },
  {
    value: "other",
    label: "Other (custom)",
    hint: "Custom external_ref pattern (InboundSourceRoutingConfig.CustomExternalRefPattern) + provider_type='other' (generic adapter).",
  },
];

/**
 * InboundSourceTypeSelector — X-2 multi-provider 운영 UI 의 1번째 widget.
 * system_admin 이 platform 별 inbound_source_type 을 선택 (Gitea/Jira/Other/Disabled).
 * raw HTML <select> + 선택된 옵션의 hint 표시.
 */
export function InboundSourceTypeSelector({
  value,
  onChange,
  disabled = false,
}: {
  value: InboundSourceType;
  onChange: (v: InboundSourceType) => void;
  disabled?: boolean;
}) {
  return (
    <div className="glass border-border rounded-2xl p-5 space-y-3 shadow-sm">
      <div className="flex items-center gap-2">
        <Webhook className="w-4 h-4 text-primary" />
        <h3 className="text-sm font-bold uppercase tracking-wider text-foreground">
          Inbound Source Type
        </h3>
      </div>

      <div className="space-y-2">
        <label
          htmlFor="inbound-source-type-select"
          className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground"
        >
          Provider Type
        </label>
        <select
          id="inbound-source-type-select"
          value={value || "_disabled_"}
          onChange={(e) =>
            onChange(e.target.value === "_disabled_" ? "" : (e.target.value as InboundSourceType))
          }
          disabled={disabled}
          className="w-full px-3 py-2 rounded-md border border-border bg-background text-foreground text-sm"
        >
          {SUPPORTED_TYPES.map((opt) => (
            <option key={opt.value || "_disabled_"} value={opt.value || "_disabled_"}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      {value && (
        <div className="flex items-start gap-2 px-3 py-2 rounded-md bg-muted/30 text-xs text-muted-foreground">
          <AlertCircle className="w-3.5 h-3.5 mt-0.5 shrink-0" />
          <span>{SUPPORTED_TYPES.find((o) => o.value === value)?.hint}</span>
        </div>
      )}
    </div>
  );
}
