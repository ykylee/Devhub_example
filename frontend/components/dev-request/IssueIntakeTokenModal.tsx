"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { X, Key, Loader2, Plus, Trash2, Copy, AlertTriangle, Check, Eye, EyeOff } from "lucide-react";
import { devRequestTokenService } from "@/lib/services/dev_request_token.service";
import type {
  DevRequestIntakeToken,
  IssuedDevRequestIntakeToken,
} from "@/lib/services/dev_request_token.types";

interface IssueIntakeTokenModalProps {
  onClose: () => void;
  onIssued: (tok: DevRequestIntakeToken) => void;
}

type ModalPhase = "form" | "reveal";

export function IssueIntakeTokenModal({ onClose, onIssued }: IssueIntakeTokenModalProps) {
  const [phase, setPhase] = useState<ModalPhase>("form");
  const [clientLabel, setClientLabel] = useState("");
  const [sourceSystem, setSourceSystem] = useState("");
  const [allowedIPs, setAllowedIPs] = useState<string[]>([""]);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [issued, setIssued] = useState<IssuedDevRequestIntakeToken | null>(null);
  const [copied, setCopied] = useState(false);
  const [showToken, setShowToken] = useState(false);

  const canCloseForm = phase === "form" && !submitting;

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape" && canCloseForm) onClose();
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [canCloseForm, onClose]);

  const updateIP = (idx: number, value: string) => {
    setAllowedIPs((prev) => prev.map((v, i) => (i === idx ? value : v)));
  };
  const addIP = () => setAllowedIPs((prev) => [...prev, ""]);
  const removeIP = (idx: number) =>
    setAllowedIPs((prev) => prev.filter((_, i) => i !== idx).concat(prev.length === 1 ? [""] : []));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      const cleanedIPs = allowedIPs.map((v) => v.trim()).filter(Boolean);
      if (cleanedIPs.length === 0) {
        setError("allowed_ips 는 최소 1개 이상의 CIDR 또는 IP 가 필요합니다 (deny-by-default).");
        setSubmitting(false);
        return;
      }
      const result = await devRequestTokenService.issue({
        client_label: clientLabel.trim(),
        source_system: sourceSystem.trim(),
        allowed_ips: cleanedIPs,
      });
      setIssued(result);
      setPhase("reveal");
      onIssued(result);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to issue token";
      setError(msg);
    } finally {
      setSubmitting(false);
    }
  };

  const handleCopy = async () => {
    if (!issued) return;
    try {
      await navigator.clipboard.writeText(issued.plain_token);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2000);
    } catch {
      // clipboard 가 안 되는 환경 — 사용자가 수동 copy
    }
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-6">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={canCloseForm ? onClose : undefined}
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
      />

      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        className="relative w-full max-w-xl glass border-border rounded-3xl shadow-2xl overflow-hidden"
        role="dialog"
        aria-modal="true"
        aria-labelledby="intake-token-modal-title"
      >
        <div className="p-8 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-orange-500/20 rounded-xl flex items-center justify-center">
              <Key className="w-5 h-5 text-orange-400" />
            </div>
            <div>
              <h2 id="intake-token-modal-title" className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                Issue <span className="text-orange-400">Intake Token</span>
              </h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                {phase === "form" ? "외부 시스템 인증 자격 발급" : "토큰 1회 노출 — 안전 보관"}
              </p>
            </div>
          </div>
          {phase === "form" && (
            <button
              onClick={onClose}
              disabled={submitting}
              className="p-2 hover:bg-muted/30 rounded-xl text-muted-foreground transition-colors disabled:opacity-50"
            >
              <X className="w-5 h-5" />
            </button>
          )}
        </div>

        {phase === "form" && (
          <form onSubmit={handleSubmit} className="p-8 space-y-6 max-h-[70vh] overflow-y-auto custom-scrollbar">
            <div className="space-y-2">
              <label htmlFor="clientLabel" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Client Label</label>
              <input
                id="clientLabel"
                required
                value={clientLabel}
                onChange={(e) => setClientLabel(e.target.value)}
                placeholder="e.g. ops_portal"
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-orange-400/50"
              />
              <p className="text-[10px] text-muted-foreground/60 px-1">운영 식별자 — log/audit 에 보이는 token 의 이름.</p>
            </div>

            <div className="space-y-2">
              <label htmlFor="sourceSystem" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Source System</label>
              <input
                id="sourceSystem"
                required
                value={sourceSystem}
                onChange={(e) => setSourceSystem(e.target.value)}
                placeholder="e.g. ops, jira, servicedesk"
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-orange-400/50"
              />
              <p className="text-[10px] text-muted-foreground/60 px-1">intake 요청의 dev_request.source_system 자동 채움 값 (body spoof 방지).</p>
            </div>

            <div className="space-y-3">
              <label htmlFor="allowedIPs-0" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Allowed IPs / CIDRs</label>
              {allowedIPs.map((ip, idx) => (
                <div key={idx} className="flex gap-2">
                  <input
                    id={`allowedIPs-${idx}`}
                    value={ip}
                    onChange={(e) => updateIP(idx, e.target.value)}
                    placeholder="e.g. 10.0.0.0/24 or 192.0.2.7"
                    className="flex-1 bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm font-mono text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-orange-400/50"
                  />
                  <button
                    type="button"
                    onClick={() => removeIP(idx)}
                    disabled={allowedIPs.length === 1 && ip === ""}
                    className="glass border-border px-4 rounded-2xl hover:bg-red-500/10 hover:border-red-500/30 hover:text-red-400 transition-all text-muted-foreground disabled:opacity-30"
                    aria-label={`Remove IP ${idx + 1}`}
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              ))}
              <button
                type="button"
                onClick={addIP}
                className="flex items-center gap-2 px-4 py-2 rounded-xl bg-muted/20 border border-dashed border-border text-muted-foreground hover:bg-muted/40 hover:text-foreground dark:hover:text-primary-foreground transition-all text-[10px] font-black uppercase tracking-widest"
              >
                <Plus className="w-3.5 h-3.5" />
                Add IP / CIDR
              </button>
              <p className="text-[10px] text-muted-foreground/60 px-1">
                deny-by-default — 최소 1개 이상. backend 가 빈 배열 거절 (invalid_allowed_ips).
              </p>
            </div>

            {error && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: "auto" }}
                className="p-4 bg-orange-500/10 border border-orange-500/20 rounded-2xl text-[11px] text-orange-400 font-medium"
              >
                {error}
              </motion.div>
            )}

            <div className="flex gap-4 pt-4 border-t border-border/60">
              <button
                type="button"
                onClick={onClose}
                disabled={submitting}
                className="flex-1 glass border-border text-foreground dark:text-primary-foreground font-bold py-4 rounded-2xl hover:bg-muted/30 transition-all uppercase tracking-widest text-[10px]"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={submitting}
                className="flex-1 bg-orange-500 text-white font-black py-4 px-8 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all shadow-xl shadow-orange-500/20 disabled:opacity-50 uppercase tracking-widest text-[10px] flex items-center justify-center gap-2"
              >
                {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <>Issue Token</>}
              </button>
            </div>
          </form>
        )}

        {phase === "reveal" && issued && (
          <div className="p-8 space-y-6">
            <div className="p-4 bg-amber-500/10 border border-amber-500/30 rounded-2xl flex gap-3">
              <AlertTriangle className="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" />
              <div className="space-y-1">
                <p className="text-xs font-black text-amber-400 uppercase tracking-widest">Token shown once</p>
                <p className="text-[11px] text-amber-400/80">
                  이 토큰은 본 화면에서만 1회 노출됩니다. 즉시 안전한 저장소 (vault, 비밀 관리자) 에 옮긴 뒤 이 창을 닫으세요.
                  창을 닫으면 server 의 SHA-256 해시만 남으며, 잃어버린 토큰은 복구할 수 없습니다 — revoke 후 재발급 해야 합니다.
                </p>
              </div>
            </div>

            <div className="space-y-3">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Plain Token</label>
              <div className="flex gap-2">
                <code className="flex-1 bg-muted/40 border border-border rounded-2xl px-4 py-3 text-xs font-mono text-foreground dark:text-primary-foreground break-all">
                  {showToken ? issued.plain_token : "•".repeat(32)}
                </code>
                <button
                  type="button"
                  onClick={() => setShowToken(!showToken)}
                  className="glass border-border px-4 rounded-2xl hover:bg-muted/40 transition-all text-foreground dark:text-primary-foreground flex items-center justify-center text-[10px] font-black uppercase tracking-widest"
                  aria-label={showToken ? "Hide token" : "Show token"}
                >
                  {showToken ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
                <button
                  type="button"
                  onClick={handleCopy}
                  className="glass border-border px-4 rounded-2xl hover:bg-muted/40 transition-all text-foreground dark:text-primary-foreground flex items-center gap-2 text-[10px] font-black uppercase tracking-widest"
                >
                  {copied ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
                  {copied ? "Copied" : "Copy"}
                </button>
              </div>
            </div>

            <div className="space-y-2 text-[11px] text-muted-foreground">
              <p>
                <span className="font-black uppercase tracking-widest text-[9px] text-muted-foreground/60 mr-2">Client</span>
                {issued.client_label}
              </p>
              <p>
                <span className="font-black uppercase tracking-widest text-[9px] text-muted-foreground/60 mr-2">Source</span>
                {issued.source_system}
              </p>
              <p>
                <span className="font-black uppercase tracking-widest text-[9px] text-muted-foreground/60 mr-2">Allowed</span>
                <span className="font-mono">{issued.allowed_ips.join(", ")}</span>
              </p>
            </div>

            <div className="pt-4 border-t border-border/60">
              <button
                type="button"
                onClick={onClose}
                className="w-full bg-primary text-primary-foreground font-black py-4 px-8 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all shadow-xl shadow-primary/20 uppercase tracking-widest text-[10px]"
              >
                저장 완료 — 닫기
              </button>
            </div>
          </div>
        )}
      </motion.div>
    </div>
  );
}
