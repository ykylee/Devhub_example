"use client";

import { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { ChevronLeft, ChevronRight, FileText, RefreshCw, Search } from "lucide-react";
import { auditService } from "@/lib/services/audit.service";
import type { AuditLogEntry, AuditLogFilters } from "@/lib/services/audit.types";

const PAGE_SIZE = 50;

// AdminSettingsAuditPage — system_admin only (AuthGuard + layout guard).
// Backed by GET /api/v1/audit-logs (handler.listAuditLogs in audit.go).
// Filter inputs map 1:1 to ListAuditLogsOptions; debounce the change handler
// so users can type freely without flooding the backend.
export default function AdminSettingsAuditPage() {
  const [entries, setEntries] = useState<AuditLogEntry[]>([]);
  const [count, setCount] = useState<number>(0);
  const [offset, setOffset] = useState<number>(0);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const [draftFilters, setDraftFilters] = useState<AuditLogFilters>({});
  const [appliedFilters, setAppliedFilters] = useState<AuditLogFilters>({});
  // reloadTick bumps when the operator clicks Refresh; useEffect re-runs
  // without us having to call setState synchronously from outside (which
  // the react-hooks/set-state-in-effect rule rejects for cascading-render
  // reasons).
  const [reloadTick, setReloadTick] = useState(0);

  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      try {
        const result = await auditService.getLogs({ ...appliedFilters, limit: PAGE_SIZE, offset });
        if (cancelled) return;
        setEntries(result.entries);
        setCount(result.entries.length);
        setLoadError(null);
      } catch (err) {
        if (cancelled) return;
        const message = err instanceof Error ? err.message : "Failed to load audit logs";
        console.error("[admin/settings/audit] load failed:", err);
        setLoadError(message);
        setEntries([]);
        setCount(0);
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    };
    run();
    return () => {
      cancelled = true;
    };
  }, [appliedFilters, offset, reloadTick]);

  const applyFilters = () => {
    setIsLoading(true);
    setOffset(0);
    setAppliedFilters(draftFilters);
  };

  const clearFilters = () => {
    setIsLoading(true);
    setDraftFilters({});
    setOffset(0);
    setAppliedFilters({});
  };

  const refresh = () => {
    setIsLoading(true);
    setReloadTick((t) => t + 1);
  };

  const goPrev = () => {
    setIsLoading(true);
    setOffset(Math.max(0, offset - PAGE_SIZE));
  };

  const goNext = () => {
    setIsLoading(true);
    setOffset(offset + PAGE_SIZE);
  };

  const canPrev = offset > 0;
  const canNext = count === PAGE_SIZE; // backend may have more if the page is full

  // Stable list of action labels seen in the current page — feeds the
  // "common values" hints below each filter so operators can discover
  // valid filter values without consulting docs.
  const sampledActions = useMemo(() => Array.from(new Set(entries.map((e) => e.action))).slice(0, 6), [entries]);
  const sampledTargetTypes = useMemo(() => Array.from(new Set(entries.map((e) => e.target_type))).slice(0, 6), [entries]);

  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="glass border-border/50 rounded-2xl p-5 space-y-4"
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <FileText className="w-4 h-4 text-orange-400" />
            <h2 className="text-xs font-black uppercase tracking-widest text-foreground">Audit Log Filters</h2>
          </div>
          <button
            type="button"
            onClick={refresh}
            className="text-[10px] font-black uppercase tracking-widest text-muted-foreground hover:text-foreground flex items-center gap-1 transition-colors"
            title="Refresh current page"
          >
            <RefreshCw className="w-3 h-3" /> Refresh
          </button>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
          <FilterField
            label="Actor login"
            value={draftFilters.actor_login ?? ""}
            onChange={(v) => setDraftFilters({ ...draftFilters, actor_login: v || undefined })}
            placeholder="e.g. alice"
            hints={[]}
          />
          <FilterField
            label="Action"
            value={draftFilters.action ?? ""}
            onChange={(v) => setDraftFilters({ ...draftFilters, action: v || undefined })}
            placeholder="e.g. auth.login.succeeded"
            hints={sampledActions}
          />
          <FilterField
            label="Target type"
            value={draftFilters.target_type ?? ""}
            onChange={(v) => setDraftFilters({ ...draftFilters, target_type: v || undefined })}
            placeholder="e.g. user, command"
            hints={sampledTargetTypes}
          />
          <FilterField
            label="Target id"
            value={draftFilters.target_id ?? ""}
            onChange={(v) => setDraftFilters({ ...draftFilters, target_id: v || undefined })}
            placeholder="exact id"
            hints={[]}
          />
        </div>

        <div className="flex justify-end gap-2 pt-1">
          <button
            type="button"
            onClick={clearFilters}
            className="px-4 py-2 rounded-xl border border-border text-[10px] font-black uppercase tracking-widest text-muted-foreground hover:text-foreground transition-colors"
          >
            Clear
          </button>
          <button
            type="button"
            onClick={applyFilters}
            className="px-4 py-2 rounded-xl bg-primary text-primary-foreground text-[10px] font-black uppercase tracking-widest hover:opacity-90 transition-opacity flex items-center gap-1"
          >
            <Search className="w-3 h-3" /> Apply
          </button>
        </div>
      </motion.div>

      {loadError && (
        <div className="glass border-rose-500/30 rounded-2xl p-4 text-xs text-rose-400">
          Load failed: {loadError}
        </div>
      )}

      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="glass border-border/50 rounded-2xl overflow-hidden"
      >
        <div className="px-5 py-3 border-b border-border/30 flex items-center justify-between">
          <h2 className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
            {isLoading ? "Loading audit logs..." : `Showing ${count} entries (offset ${offset})`}
          </h2>
          <div className="flex items-center gap-2">
            <button
              type="button"
              disabled={!canPrev || isLoading}
              onClick={goPrev}
              className="p-1.5 rounded-lg border border-border text-muted-foreground hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
              title="Previous page"
            >
              <ChevronLeft className="w-3.5 h-3.5" />
            </button>
            <button
              type="button"
              disabled={!canNext || isLoading}
              onClick={goNext}
              className="p-1.5 rounded-lg border border-border text-muted-foreground hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
              title="Next page"
            >
              <ChevronRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <div className="divide-y divide-border/30">
          {entries.length === 0 && !isLoading && (
            <div className="p-10 text-center text-xs text-muted-foreground">
              No audit log entries match the current filters.
            </div>
          )}
          {entries.map((entry) => {
            const isExpanded = expandedId === entry.audit_id;
            return (
              <button
                key={entry.audit_id}
                type="button"
                onClick={() => setExpandedId(isExpanded ? null : entry.audit_id)}
                className="w-full text-left px-5 py-3 hover:bg-primary/5 transition-colors"
              >
                <div className="flex items-center gap-4 text-xs">
                  <span className="font-mono text-[10px] text-muted-foreground w-44 shrink-0">
                    {formatTimestamp(entry.created_at)}
                  </span>
                  <span className="font-bold text-foreground w-28 shrink-0 truncate" title={entry.actor_login}>
                    {entry.actor_login || "(system)"}
                  </span>
                  <span className="font-mono text-accent w-64 shrink-0 truncate" title={entry.action}>
                    {entry.action}
                  </span>
                  <span className="text-muted-foreground w-32 shrink-0 truncate" title={entry.target_type}>
                    {entry.target_type}
                  </span>
                  <span className="font-mono text-muted-foreground truncate flex-1" title={entry.target_id}>
                    {entry.target_id}
                  </span>
                </div>
                {isExpanded && (
                  <div className="mt-3 ml-44 space-y-2">
                    <Detail label="audit_id" value={entry.audit_id} mono />
                    {entry.command_id && <Detail label="command_id" value={entry.command_id} mono />}
                    {entry.source_type && <Detail label="source_type" value={entry.source_type} />}
                    {entry.source_ip && <Detail label="source_ip" value={entry.source_ip} mono />}
                    {entry.request_id && <Detail label="request_id" value={entry.request_id} mono />}
                    <Detail
                      label="payload"
                      value={JSON.stringify(entry.payload ?? {}, null, 2)}
                      mono
                      multiline
                    />
                  </div>
                )}
              </button>
            );
          })}
        </div>
      </motion.div>
    </div>
  );
}

interface FilterFieldProps {
  label: string;
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  hints?: string[];
}

function FilterField({ label, value, onChange, placeholder, hints = [] }: FilterFieldProps) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-[9px] font-bold uppercase tracking-widest text-muted-foreground">{label}</span>
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="bg-input border border-border rounded-lg px-3 py-1.5 text-xs text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 transition-all"
      />
      {hints.length > 0 && (
        <div className="flex flex-wrap gap-1 mt-0.5">
          {hints.map((h) => (
            <button
              key={h}
              type="button"
              onClick={() => onChange(h)}
              className="text-[9px] font-mono text-muted-foreground hover:text-primary transition-colors px-1.5 py-0.5 rounded border border-border/50"
            >
              {h}
            </button>
          ))}
        </div>
      )}
    </label>
  );
}

interface DetailProps {
  label: string;
  value: string;
  mono?: boolean;
  multiline?: boolean;
}

function Detail({ label, value, mono, multiline }: DetailProps) {
  return (
    <div className="text-[11px] flex gap-2">
      <span className="font-bold uppercase tracking-widest text-muted-foreground w-24 shrink-0">{label}</span>
      <span
        className={[
          mono ? "font-mono" : "",
          multiline ? "whitespace-pre-wrap break-words" : "truncate",
          "text-foreground/90 flex-1",
        ].join(" ")}
      >
        {value}
      </span>
    </div>
  );
}

function formatTimestamp(iso: string): string {
  if (!iso) return "";
  try {
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return iso;
    return d.toISOString().replace("T", " ").replace(/\.\d+Z$/, "Z");
  } catch {
    return iso;
  }
}
