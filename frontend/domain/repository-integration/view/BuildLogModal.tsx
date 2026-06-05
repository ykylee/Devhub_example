"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { X, Terminal, Copy, Check, Loader2 } from "lucide-react";
import { repositoryService } from "@/domain/repository-integration/service/repository.service";

interface BuildLogModalProps {
  repositoryId: number;
  runId: number;
  runExternalId: string;
  onClose: () => void;
}

export function BuildLogModal({ repositoryId, runId, runExternalId, onClose }: BuildLogModalProps) {
  const [log, setLog] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const fetchLog = async () => {
      try {
        setLoading(true);
        setError(null);
        const logData = await repositoryService.getBuildRunLog(repositoryId, runId);
        setLog(logData);
      } catch (err) {
        console.error(err);
        setError("Failed to retrieve build logs. SCM connection might be degraded.");
      } finally {
        setLoading(false);
      }
    };
    void fetchLog();
  }, [repositoryId, runId]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(log);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy log", err);
    }
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-6">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
      />

      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        role="dialog"
        aria-modal="true"
        className="relative w-full max-w-3xl glass border-border rounded-3xl shadow-2xl overflow-hidden flex flex-col max-h-[85vh]"
      >
        {/* Header */}
        <div className="p-6 border-b border-border flex items-center justify-between bg-muted/10">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-primary/10 rounded-xl flex items-center justify-center text-primary">
              <Terminal className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-lg font-black text-foreground dark:text-primary-foreground tracking-tight">
                Build Run <span className="text-primary">Console Output</span>
              </h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                External ID: {runExternalId} • Run #{runId}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {!loading && !error && (
              <button
                onClick={handleCopy}
                className="p-2 hover:bg-muted/30 rounded-xl text-muted-foreground transition-all flex items-center gap-1.5 text-xs font-bold"
                title="Copy full log"
              >
                {copied ? (
                  <>
                    <Check className="w-4 h-4 text-success" />
                    <span className="text-success text-[10px] font-black uppercase tracking-wider">Copied</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-4 h-4" />
                    <span className="text-[10px] font-black uppercase tracking-wider">Copy</span>
                  </>
                )}
              </button>
            )}
            <button
              onClick={onClose}
              className="p-2 hover:bg-muted/30 rounded-xl text-muted-foreground transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 p-6 overflow-y-auto bg-black/90 dark:bg-black/95 text-slate-300 font-mono text-xs leading-relaxed custom-scrollbar">
          {loading ? (
            <div className="h-64 flex flex-col items-center justify-center gap-3">
              <Loader2 className="w-8 h-8 animate-spin text-primary" />
              <p className="text-xs text-muted-foreground animate-pulse font-sans font-bold">Streaming console buffer...</p>
            </div>
          ) : error ? (
            <div className="p-4 bg-destructive/10 border border-destructive/20 rounded-2xl text-[11px] text-destructive font-sans font-bold flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-destructive animate-ping" />
              {error}
            </div>
          ) : (
            <pre className="whitespace-pre-wrap select-text selection:bg-primary/30 selection:text-white">
              {log.split("\n").map((line, idx) => {
                let colorClass = "text-slate-300";
                if (line.includes("[ERROR]")) {
                  colorClass = "text-destructive font-bold";
                } else if (line.includes("[WARN]")) {
                  colorClass = "text-warning font-bold";
                } else if (line.includes("[INFO]")) {
                  colorClass = "text-info/80";
                } else if (line.includes("✕")) {
                  colorClass = "text-destructive font-bold pl-2";
                }
                return (
                  <div key={idx} className={colorClass}>
                    {line}
                  </div>
                );
              })}
            </pre>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-border bg-muted/5 flex justify-end">
          <button
            onClick={onClose}
            className="px-6 py-2.5 bg-muted/20 border border-border/80 hover:bg-muted/40 text-foreground dark:text-primary-foreground font-black rounded-xl transition-all uppercase tracking-widest text-[10px]"
          >
            Close Console
          </button>
        </div>
      </motion.div>
    </div>
  );
}
