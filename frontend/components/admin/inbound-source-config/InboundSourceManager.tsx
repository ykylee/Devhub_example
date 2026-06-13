"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Save, AlertTriangle, CheckCircle2, Loader2 } from "lucide-react";
import type {
  PlatformInboundSourceView,
  InboundSourceType,
  InboundSourceRoutingConfig,
} from "@/domain/integration-registry/schema/integration.types";
import { InboundSourceTypeSelector } from "./InboundSourceTypeSelector";
import { InboundSourceConfigEditor } from "./InboundSourceConfigEditor";
import { PatternPreview } from "./PatternPreview";

/**
 * InboundSourceManager — X-2 multi-provider 운영 UI 의 4번째 widget (전체 통합 view).
 * system_admin 이 platform 별 inbound_source_type + inbound_source_config 를 관리.
 * 3 widget (InboundSourceTypeSelector + InboundSourceConfigEditor + PatternPreview) +
 * 저장 버튼 + audit log.
 *
 * 1차 PR #586 의 auto_route.go 의 InboundSourceRoutingConfig + 2~3차 PR 의 WebhookAdapter
 * (Gitea/Jira/Generic) 와 1:1 정합.
 */
export function InboundSourceManager({
  platforms,
  onSave,
}: {
  platforms: PlatformInboundSourceView[];
  onSave: (
    platformId: string,
    next: {
      inbound_source_type: InboundSourceType;
      inbound_source_config: InboundSourceRoutingConfig;
    },
  ) => void | Promise<void>;
}) {
  const router = useRouter();
  const [selectedId, setSelectedId] = useState(platforms[0]?.platform_id || "");
  const selected = platforms.find((p) => p.platform_id === selectedId);
  const [draftType, setDraftType] = useState<InboundSourceType>(selected?.inbound_source_type || "");
  const [draftConfig, setDraftConfig] = useState<InboundSourceRoutingConfig>(
    selected?.inbound_source_config || {},
  );
  const [saveState, setSaveState] = useState<"idle" | "saving" | "saved" | "error">("idle");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const handleSelect = (id: string) => {
    setSelectedId(id);
    const p = platforms.find((x) => x.platform_id === id);
    if (p) {
      setDraftType(p.inbound_source_type);
      setDraftConfig(p.inbound_source_config);
    }
  };

  const handleSave = async () => {
    if (!selected) return;
    setSaveState("saving");
    setErrorMsg(null);
    try {
      await onSave(selected.platform_id, {
        inbound_source_type: draftType,
        inbound_source_config: draftConfig,
      });
      setSaveState("saved");
      setTimeout(() => setSaveState("idle"), 2000);
      router.refresh();
    } catch (e) {
      setSaveState("error");
      setErrorMsg(e instanceof Error ? e.message : "save failed");
    }
  };

  return (
    <div className="space-y-4 animate-fade-in">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-bold text-foreground">Inbound Source (Multi-Provider Webhook)</h2>
        <select
          value={selectedId}
          onChange={(e) => handleSelect(e.target.value)}
          className="px-3 py-2 rounded-md border border-border bg-background text-foreground text-sm"
        >
          {platforms.map((p) => (
            <option key={p.platform_id} value={p.platform_id}>
              {p.platform_key} — {p.platform_name}
            </option>
          ))}
        </select>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="space-y-4">
          <InboundSourceTypeSelector
            value={draftType}
            onChange={setDraftType}
            disabled={saveState === "saving"}
          />
          <PatternPreview type={draftType} customPattern={draftConfig.custom_external_ref_pattern || ""} />
        </div>
        <div>
          <InboundSourceConfigEditor
            config={draftConfig}
            onSave={(next) => setDraftConfig(next)}
            isSaving={saveState === "saving"}
          />
        </div>
      </div>

      {errorMsg && (
        <div className="flex items-start gap-2 px-3 py-2 rounded-md bg-rose-500/10 text-rose-700 dark:text-rose-300 text-xs">
          <AlertTriangle className="w-3.5 h-3.5 mt-0.5 shrink-0" />
          <span>{errorMsg}</span>
        </div>
      )}

      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={handleSave}
          disabled={saveState === "saving" || !selected}
          className="bg-primary text-primary-foreground px-5 py-2.5 rounded-md text-xs font-bold hover:bg-primary/90 transition-colors flex items-center gap-2 disabled:opacity-50"
        >
          {saveState === "saving" ? (
            <Loader2 className="w-3.5 h-3.5 animate-spin" />
          ) : saveState === "saved" ? (
            <CheckCircle2 className="w-3.5 h-3.5" />
          ) : (
            <Save className="w-3.5 h-3.5" />
          )}
          {saveState === "saving" ? "Saving..." : saveState === "saved" ? "Saved" : "Save Inbound Source"}
        </button>
        {saveState === "saved" && (
          <span className="text-xs text-emerald-600 dark:text-emerald-400">저장 완료 (2초 후 자동 reset)</span>
        )}
      </div>
    </div>
  );
}
