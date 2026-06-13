"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import type {
  PlatformInboundSourceView,
  InboundSourceRoutingConfig,
} from "@/domain/integration-registry/schema/integration.types";
import { AlertTriangle, FileJson, Save, RotateCcw } from "lucide-react";

/**
 * InboundSourceConfigEditor — X-2 multi-provider 운영 UI 의 2번째 widget.
 * system_admin 이 platform 별 InboundSourceRoutingConfig (JSONB) 을 직접 편집.
 * 'other' provider 의 custom_external_ref_pattern 등 사용자 정의 regex 정공법.
 *
 * 본 widget 은 raw JSONB textarea editor (Monaco-like) — 운영자 audit trail +
 * 즉시 preview + 저장 시 backend PATCH /api/v1/platforms/:id inbound_source_config 호출.
 *
 * 1차 PR #586 의 auto_route.go 의 InboundSourceRoutingConfig struct 와 1:1 정합.
 */
export function InboundSourceConfigEditor({
  config,
  onSave,
  isSaving = false,
}: {
  config: InboundSourceRoutingConfig;
  onSave: (next: InboundSourceRoutingConfig) => void | Promise<void>;
  isSaving?: boolean;
}) {
  const [text, setText] = useState(() => JSON.stringify(config, null, 2));
  const [parseError, setParseError] = useState<string | null>(null);

  const handleTextChange = (next: string) => {
    setText(next);
    try {
      const parsed = JSON.parse(next);
      if (typeof parsed !== "object" || Array.isArray(parsed) || parsed === null) {
        setParseError("must be a JSON object");
        return;
      }
      setParseError(null);
    } catch (e) {
      setParseError(e instanceof Error ? e.message : "invalid JSON");
    }
  };

  const handleSave = () => {
    if (parseError) return;
    const parsed = JSON.parse(text) as InboundSourceRoutingConfig;
    void onSave(parsed);
  };

  const handleReset = () => {
    setText(JSON.stringify(config, null, 2));
    setParseError(null);
  };

  return (
    <div className="glass border-border rounded-2xl p-5 space-y-3 shadow-sm">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <FileJson className="w-4 h-4 text-primary" />
          <h3 className="text-sm font-bold uppercase tracking-wider text-foreground">
            Inbound Source Config (JSONB)
          </h3>
        </div>
        <button
          type="button"
          onClick={handleReset}
          className="text-muted-foreground hover:text-primary transition-colors"
          title="Reset to saved config"
        >
          <RotateCcw className="w-3.5 h-3.5" />
        </button>
      </div>

      <textarea
        value={text}
        onChange={(e) => handleTextChange(e.target.value)}
        rows={8}
        spellCheck={false}
        className="w-full px-3 py-2 rounded-md border border-border bg-background text-foreground font-mono text-xs resize-vertical"
        placeholder='{"custom_external_ref_pattern": "^CUSTOM-\\d+$"}'
      />

      {parseError && (
        <div className="flex items-start gap-2 px-3 py-2 rounded-md bg-rose-500/10 text-rose-700 dark:text-rose-300 text-xs">
          <AlertTriangle className="w-3.5 h-3.5 mt-0.5 shrink-0" />
          <span>{parseError}</span>
        </div>
      )}

      <button
        type="button"
        onClick={handleSave}
        disabled={parseError !== null || isSaving}
        className="bg-primary text-primary-foreground px-4 py-2 rounded-md text-xs font-bold hover:bg-primary/90 transition-colors flex items-center gap-2 disabled:opacity-50"
      >
        <Save className="w-3.5 h-3.5" /> {isSaving ? "Saving..." : "Save Config"}
      </button>
    </div>
  );
}
