"use client";

import { useState } from "react";
import { AlertTriangle, Eye, TestTube2 } from "lucide-react";
import type { InboundSourceType } from "@/domain/integration-registry/schema/integration.types";

/**
 * PatternPreview — X-2 multi-provider 운영 UI 의 3번째 widget.
 * system_admin 이 custom_external_ref_pattern 의 regex 컴파일 가능 여부 + sample match 검증.
 * 1차 PR #586 의 auto_route.go 의 InboundSourceRoutingConfig.CustomExternalRefPattern +
 * 3차 PR #588 의 GenericWebhookAdapter.MatchExternalRefPattern 와 1:1 정합.
 *
 * 제공 예시: provider-specific (GITEA-<n> | JIRA-<n> | #<n> | !<n>) + custom regex 검증.
 */
const SAMPLE_EXTERNAL_REFS: Record<InboundSourceType, string[]> = {
  gitea: ["GITEA-123", "GITEA-4567", "PUSH-devhub/core"],
  jira: ["DEV-456", "PROJ-789", "JIRA-1"],
  other: ["CUSTOM-999", "OTHER-123", "#789", "!100"],
  "": [],
};

const PROVIDER_PATTERN: Record<InboundSourceType, string | null> = {
  gitea: "^GITEA-\\d+$",
  jira: "^([A-Z][A-Z0-9_]{1,9})-\\d+$",
  other: null, // generic — uses custom_external_ref_pattern
  "": null,
};

function tryCompile(pattern: string): { ok: true } | { ok: false; error: string } {
  try {
    new RegExp(pattern);
    return { ok: true };
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : "invalid regex" };
  }
}

export function PatternPreview({
  type,
  customPattern,
}: {
  type: InboundSourceType;
  customPattern: string;
}) {
  const [sample, setSample] = useState(SAMPLE_EXTERNAL_REFS[type][0] || "");
  const samples = SAMPLE_EXTERNAL_REFS[type] || [];
  const providerPattern = PROVIDER_PATTERN[type];
  const effectivePattern = type === "other" ? customPattern : providerPattern || "";
  const compileResult = effectivePattern ? tryCompile(effectivePattern) : null;

  return (
    <div className="glass border-border rounded-2xl p-5 space-y-3 shadow-sm">
      <div className="flex items-center gap-2">
        <Eye className="w-4 h-4 text-primary" />
        <h3 className="text-sm font-bold uppercase tracking-wider text-foreground">
          Pattern Preview
        </h3>
      </div>

      {!type && (
        <p className="text-xs text-muted-foreground">
          inbound_source_type 이 Disabled. pattern preview 미표시.
        </p>
      )}

      {type && (
        <>
          <div className="space-y-2">
            <label
              htmlFor="pattern-preview-sample"
              className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground"
            >
              Sample external_ref
            </label>
            <select
              id="pattern-preview-sample"
              value={sample}
              onChange={(e) => setSample(e.target.value)}
              className="w-full px-3 py-2 rounded-md border border-border bg-background text-foreground text-xs"
            >
              {samples.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </div>

          {effectivePattern && (
            <div className="px-3 py-2 rounded-md bg-muted/30 font-mono text-xs text-foreground">
              <span className="text-muted-foreground">pattern:</span> {effectivePattern}
            </div>
          )}

          {compileResult && !compileResult.ok && (
            <div className="flex items-start gap-2 px-3 py-2 rounded-md bg-rose-500/10 text-rose-700 dark:text-rose-300 text-xs">
              <AlertTriangle className="w-3.5 h-3.5 mt-0.5 shrink-0" />
              <span>{compileResult.error}</span>
            </div>
          )}

          {compileResult && compileResult.ok && sample && (
            <div className="flex items-center gap-2 px-3 py-2 rounded-md bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 text-xs">
              <TestTube2 className="w-3.5 h-3.5" />
              <span>
                <code className="font-mono">{sample}</code> →{" "}
                {new RegExp(effectivePattern).test(sample) ? "MATCH" : "NO MATCH"}
              </span>
            </div>
          )}
        </>
      )}
    </div>
  );
}
