"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { X, Key, Loader2, Plus, Trash2, Check, AlertTriangle } from "lucide-react";
import { devRequestTokenService } from "@/lib/services/dev_request_token.service";
import type { DevRequestIntakeToken } from "@/lib/services/dev_request_token.types";

interface EditIntakeTokenModalProps {
  token: DevRequestIntakeToken;
  onClose: () => void;
  onUpdated: (tok: DevRequestIntakeToken) => void;
}

export function EditIntakeTokenModal({ token, onClose, onUpdated }: EditIntakeTokenModalProps) {
  // ISO8601 시각을 <input type="datetime-local"> 값에 맞게 포맷팅합니다.
  const formatDatetimeLocal = (isoString: string | null) => {
    if (!isoString) return "";
    const d = new Date(isoString);
    if (isNaN(d.getTime())) return "";
    const pad = (n: number) => n.toString().padStart(2, "0");
    const yyyy = d.getFullYear();
    const mm = pad(d.getMonth() + 1);
    const dd = pad(d.getDate());
    const hh = pad(d.getHours());
    const min = pad(d.getMinutes());
    return `${yyyy}-${mm}-${dd}T${hh}:${min}`;
  };

  const [allowedIPs, setAllowedIPs] = useState<string[]>(
    token.allowed_ips.length > 0 ? token.allowed_ips : [""]
  );
  const [expiresAt, setExpiresAt] = useState(formatDatetimeLocal(token.expires_at));
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape" && !submitting) onClose();
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [submitting, onClose]);

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

      // expires_at 변경 처리: 날짜가 지정되었으면 ISO 문자열로 변환하고, 지정되지 않았으면 null 전송
      const finalExpiry = expiresAt ? new Date(expiresAt).toISOString() : null;

      const result = await devRequestTokenService.update(token.token_id, {
        allowed_ips: cleanedIPs,
        expires_at: finalExpiry,
      });

      setSuccess(true);
      onUpdated(result);
      setTimeout(() => {
        onClose();
      }, 800);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to update token";
      setError(msg);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-6">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={!submitting ? onClose : undefined}
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
      />

      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        className="relative w-full max-w-xl glass border-border rounded-3xl shadow-2xl overflow-hidden"
        role="dialog"
        aria-modal="true"
        aria-labelledby="edit-token-modal-title"
      >
        <div className="p-8 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-accent/20 rounded-xl flex items-center justify-center">
              <Key className="w-5 h-5 text-accent" />
            </div>
            <div>
              <h2 id="edit-token-modal-title" className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                Edit <span className="text-accent">Intake Token</span>
              </h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                토큰 운영 설정 변경 및 기간 연장
              </p>
            </div>
          </div>
          {!submitting && (
            <button
              onClick={onClose}
              disabled={submitting}
              className="p-2 hover:bg-muted/30 rounded-xl text-muted-foreground transition-colors disabled:opacity-50"
            >
              <X className="w-5 h-5" />
            </button>
          )}
        </div>

        <form onSubmit={handleSubmit} className="p-8 space-y-6 max-h-[70vh] overflow-y-auto custom-scrollbar">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1 bg-muted/20 border border-border/40 rounded-2xl p-4">
              <span className="text-[9px] font-black text-muted-foreground uppercase tracking-widest">Client Label</span>
              <p className="text-sm font-semibold text-foreground dark:text-primary-foreground truncate">
                {token.client_label}
              </p>
            </div>

            <div className="space-y-1 bg-muted/20 border border-border/40 rounded-2xl p-4">
              <span className="text-[9px] font-black text-muted-foreground uppercase tracking-widest">Source System</span>
              <p className="text-sm font-semibold text-foreground dark:text-primary-foreground truncate">
                {token.source_system}
              </p>
            </div>
          </div>

          <div className="space-y-2">
            <label htmlFor="expiresAt" className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Expires At (Optional)</label>
            <input
              id="expiresAt"
              type="datetime-local"
              value={expiresAt}
              onChange={(e) => setExpiresAt(e.target.value)}
              className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-accent/50"
            />
            <p className="text-[10px] text-muted-foreground/60 px-1">토큰 만료 일시. 비워두면 무기한 사용 가능한 토큰으로 연장됩니다.</p>
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
                  className="flex-1 bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm font-mono text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-accent/50"
                />
                <button
                  type="button"
                  onClick={() => removeIP(idx)}
                  disabled={allowedIPs.length === 1 && ip === ""}
                  className="glass border-border px-4 rounded-2xl hover:bg-destructive/10 hover:border-destructive/30 hover:text-destructive transition-all text-muted-foreground disabled:opacity-30"
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
              허용 대역 수정 — 최소 1개 이상 필요합니다.
            </p>
          </div>

          {error && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              className="p-4 bg-accent/10 border border-accent/20 rounded-2xl text-[11px] text-accent font-medium flex items-center gap-2"
            >
              <AlertTriangle className="w-4 h-4" />
              {error}
            </motion.div>
          )}

          {success && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="p-4 bg-success/10 border border-success/20 rounded-2xl text-[11px] text-success font-bold flex items-center gap-2"
            >
              <Check className="w-4 h-4" />
              성공적으로 변경 사항이 갱신되었습니다!
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
              disabled={submitting || success}
              className="flex-1 bg-warning text-warning-foreground font-black py-4 px-8 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all shadow-xl disabled:opacity-50 uppercase tracking-widest text-[10px] flex items-center justify-center gap-2"
            >
              {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <>Save Changes</>}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
