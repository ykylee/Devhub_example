"use client";

import { Loader2, RefreshCcw } from "lucide-react";

interface PageLoadingProps {
  label?: string;
}

export function PageLoading({ label = "Loading..." }: PageLoadingProps) {
  return (
    <div className="flex items-center justify-center h-[60vh]">
      <div className="flex items-center gap-3 text-muted-foreground">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
        <span className="text-sm font-bold">{label}</span>
      </div>
    </div>
  );
}

interface PageErrorProps {
  message: string;
  onRetry?: () => void;
}

export function PageError({ message, onRetry }: PageErrorProps) {
  return (
    <div className="glass-card p-4 border border-destructive/30 bg-destructive/10 text-destructive">
      <p className="text-sm font-medium">{message}</p>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="mt-3 inline-flex items-center gap-2 rounded-lg border border-destructive/40 px-3 py-1.5 text-xs font-bold hover:bg-destructive/10"
        >
          <RefreshCcw className="w-3.5 h-3.5" />
          Retry
        </button>
      )}
    </div>
  );
}

interface PageEmptyProps {
  message: string;
}

export function PageEmpty({ message }: PageEmptyProps) {
  return (
    <div className="text-center py-20 glass-card">
      <p className="text-muted-foreground font-black uppercase tracking-widest text-xs opacity-60">{message}</p>
    </div>
  );
}
