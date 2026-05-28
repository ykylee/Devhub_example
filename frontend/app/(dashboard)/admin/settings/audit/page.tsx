"use client";

import { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { ChevronLeft, ChevronRight, FileText, RefreshCw, Search, XCircle, Filter } from "lucide-react";
import { auditService } from "@/domain/audit-ops/service/audit.service";
import type { AuditLogEntry, AuditLogFilters } from "@/domain/audit-ops/schema/audit.types";
import { cn } from "@/shared/utils";

const PAGE_SIZE = 50;

export default function AdminSettingsAuditPage() {
  const [entries, setEntries] = useState<AuditLogEntry[]>([]);
  const [count, setCount] = useState<number>(0);
  const [offset, setOffset] = useState<number>(0);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const [draftFilters, setDraftFilters] = useState<AuditLogFilters>({});
  const [appliedFilters, setAppliedFilters] = useState<AuditLogFilters>({});
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
  const canNext = count === PAGE_SIZE;

  const sampledActions = useMemo(() => Array.from(new Set(entries.map((e) => e.action))).slice(0, 6), [entries]);
  const sampledTargetTypes = useMemo(() => Array.from(new Set(entries.map((e) => e.target_type))).slice(0, 6), [entries]);

  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="glass border border-border/50 rounded-2xl p-6 space-y-6 shadow-xl"
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary/10 rounded-lg">
              <Filter className="w-4 h-4 text-primary" />
            </div>
            <div>
              <h2 className="text-xs font-black uppercase tracking-widest text-foreground">Audit Log Intelligence</h2>
              <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-tighter mt-1">Refine system audit events</p>
            </div>
          </div>
          <button
            type="button"
            onClick={refresh}
            className="px-3 py-1.5 rounded-lg glass border border-border/60 text-[10px] font-black uppercase tracking-widest text-muted-foreground hover:text-foreground flex items-center gap-2 transition-all"
          >
            <RefreshCw className={cn("w-3 h-3", isLoading && "animate-spin")} /> Refresh
          </button>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <FilterField
            label="Actor login"
            value={draftFilters.actor_login ?? ""}
            onChange={(v) => setDraftFilters({ ...draftFilters, actor_login: v || undefined })}
            placeholder="e.g. yklee"
            hints={[]}
          />
          <FilterField
            label="Action"
            value={draftFilters.action ?? ""}
            onChange={(v) => setDraftFilters({ ...draftFilters, action: v || undefined })}
            placeholder="e.g. rbac.policy.updated"
            hints={sampledActions}
          />
          <FilterField
            label="Target type"
            value={draftFilters.target_type ?? ""}
            onChange={(v) => setDraftFilters({ ...draftFilters, target_type: v || undefined })}
            placeholder="e.g. role, app"
            hints={sampledTargetTypes}
          />
          <FilterField
            label="Target id"
            value={draftFilters.target_id ?? ""}
            onChange={(v) => setDraftFilters({ ...draftFilters, target_id: v || undefined })}
            placeholder="Search by ID..."
            hints={[]}
          />
        </div>

        <div className="flex justify-end gap-3 pt-2 border-t border-border/30">
          <button
            type="button"
            onClick={clearFilters}
            className="px-6 py-2.5 rounded-xl glass border border-border text-[10px] font-black uppercase tracking-widest text-muted-foreground hover:text-foreground transition-all flex items-center gap-2"
          >
            <XCircle className="w-3.5 h-3.5" /> Reset
          </button>
          <button
            type="button"
            onClick={applyFilters}
            className="px-8 py-2.5 rounded-xl bg-primary text-primary-foreground text-[10px] font-black uppercase tracking-widest hover:opacity-90 transition-all flex items-center gap-2 shadow-lg shadow-primary/20 border border-primary/20"
          >
            <Search className="w-3.5 h-3.5" /> Execute Filter
          </button>
        </div>
      </motion.div>

      {loadError && (
        <div className="glass border-destructive/30 rounded-2xl p-4 text-xs text-destructive flex items-center gap-3">
          <div className="w-2 h-2 bg-destructive rounded-full animate-pulse" />
          Load failed: {loadError}
        </div>
      )}

      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="glass-card border border-border/50 overflow-hidden shadow-2xl"
      >
        <div className="px-6 py-4 border-b border-border/30 flex items-center justify-between bg-muted/5">
          <div className="flex items-center gap-3">
            <FileText className="w-4 h-4 text-primary" />
            <h2 className="text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground">
              {isLoading ? "Analyzing Logs..." : `Audit Stream (${count} entries)`}
            </h2>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              disabled={!canPrev || isLoading}
              onClick={goPrev}
              className="p-2 rounded-xl glass border border-border/60 text-muted-foreground hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed transition-all"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <button
              type="button"
              disabled={!canNext || isLoading}
              onClick={goNext}
              className="p-2 rounded-xl glass border border-border/60 text-muted-foreground hover:text-foreground disabled:opacity-30 disabled:cursor-not-allowed transition-all"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>

        <div className="divide-y divide-border/30">
          {entries.length === 0 && !isLoading && (
            <div className="p-20 text-center space-y-4">
              <div className="w-16 h-16 bg-muted/20 rounded-full flex items-center justify-center mx-auto border border-border/40">
                <Search className="w-8 h-8 text-muted-foreground/40" />
              </div>
              <p className="text-xs text-muted-foreground font-medium">No audit log entries match the current filters.</p>
            </div>
          )}
          {entries.map((entry) => {
            const isExpanded = expandedId === entry.audit_id;
            return (
              <button
                key={entry.audit_id}
                type="button"
                onClick={() => setExpandedId(isExpanded ? null : entry.audit_id)}
                className={cn(
                  "w-full text-left px-6 py-4 hover:bg-primary/5 transition-all group relative overflow-hidden",
                  isExpanded && "bg-primary/5"
                )}
              >
                {isExpanded && <div className="absolute left-0 top-0 w-1 h-full bg-primary" />}
                <div className="flex items-center gap-6 text-xs">
                  <span className="font-mono text-[10px] text-muted-foreground w-48 shrink-0 flex items-center gap-2">
                    <div className="w-1.5 h-1.5 bg-muted-foreground/30 rounded-full" />
                    {formatTimestamp(entry.created_at)}
                  </span>
                  <span className="font-black text-foreground w-32 shrink-0 truncate group-hover:text-primary transition-colors">
                    {entry.actor_login || "(system)"}
                  </span>
                  <span className="font-mono text-accent w-64 shrink-0 truncate font-bold">
                    {entry.action}
                  </span>
                  <span className="text-muted-foreground w-32 shrink-0 truncate font-medium uppercase tracking-tighter text-[10px]">
                    {entry.target_type}
                  </span>
                  <span className="font-mono text-muted-foreground/60 truncate flex-1 text-[10px]" title={entry.target_id}>
                    {entry.target_id}
                  </span>
                </div>
                {isExpanded && (
                  <motion.div 
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: "auto" }}
                    className="mt-6 ml-48 space-y-4 p-4 glass rounded-xl border border-border/40"
                  >
                    <Detail label="Audit ID" value={entry.audit_id} mono />
                    {entry.command_id && <Detail label="Command ID" value={entry.command_id} mono />}
                    {entry.source_type && <Detail label="Source" value={entry.source_type} />}
                    {entry.source_ip && <Detail label="IP Address" value={entry.source_ip} mono />}
                    <Detail
                      label="Payload"
                      value={JSON.stringify(entry.payload ?? {}, null, 2)}
                      mono
                      multiline
                    />
                  </motion.div>
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
    <label className="flex flex-col gap-2">
      <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground px-1">{label}</span>
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="glass border border-border/60 rounded-xl px-4 py-2.5 text-xs text-foreground focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/40 transition-all font-medium placeholder:text-muted-foreground/40"
      />
      {hints.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mt-1 px-1">
          {hints.map((h) => (
            <button
              key={h}
              type="button"
              onClick={() => onChange(h)}
              className="text-[9px] font-black uppercase tracking-tighter text-muted-foreground hover:text-primary transition-all px-2 py-0.5 rounded-lg border border-border/40 hover:border-primary/30"
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
    <div className="text-[11px] flex gap-4">
      <span className="font-black uppercase tracking-widest text-muted-foreground/60 w-24 shrink-0">{label}</span>
      <span
        className={cn(
          "flex-1",
          mono ? "font-mono text-foreground/80" : "font-medium text-foreground",
          multiline ? "whitespace-pre-wrap break-words bg-black/20 p-3 rounded-lg border border-border/20" : "truncate"
        )}
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
