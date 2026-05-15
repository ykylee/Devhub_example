"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { Key, Plus } from "lucide-react";
import { devRequestTokenService } from "@/lib/services/dev_request_token.service";
import type { DevRequestIntakeToken } from "@/lib/services/dev_request_token.types";
import { IntakeTokenTable } from "@/components/dev-request/IntakeTokenTable";
import { IssueIntakeTokenModal } from "@/components/dev-request/IssueIntakeTokenModal";
import { useToast } from "@/components/ui/Toast";

export default function AdminSettingsDevRequestTokensPage() {
  const [tokens, setTokens] = useState<DevRequestIntakeToken[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showIssue, setShowIssue] = useState(false);
  const [revokingTokenID, setRevokingTokenID] = useState<string | null>(null);
  const { toast } = useToast();

  const refresh = async () => {
    setIsLoading(true);
    try {
      const result = await devRequestTokenService.list();
      setTokens(result.data);
    } catch (error) {
      console.error("[admin/settings/dev-request-tokens] load failed:", error);
      toast("intake token 목록을 불러오지 못했습니다.", "error");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setIsLoading(true);
      try {
        const result = await devRequestTokenService.list();
        if (!cancelled) setTokens(result.data);
      } catch (error) {
        if (!cancelled) console.error("[admin/settings/dev-request-tokens] load failed:", error);
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const handleIssued = (tok: DevRequestIntakeToken) => {
    // 발급 직후 목록 새로고침 — modal 의 reveal phase 가 닫히기 전이라도
    // 상태 일관성을 위해 즉시 반영.
    setTokens((prev) => [tok, ...prev]);
  };

  const handleRevoke = async (tok: DevRequestIntakeToken) => {
    if (!window.confirm(`'${tok.client_label}' 토큰을 revoke 하시겠습니까? 이 작업은 즉시 적용되며 인증이 차단됩니다.`)) {
      return;
    }
    setRevokingTokenID(tok.token_id);
    try {
      const updated = await devRequestTokenService.revoke(tok.token_id);
      setTokens((prev) => prev.map((t) => (t.token_id === updated.token_id ? updated : t)));
      toast(`토큰 '${tok.client_label}' 이 revoke 되었습니다.`, "success");
    } catch (error) {
      console.error("[admin/settings/dev-request-tokens] revoke failed:", error);
      toast("토큰 revoke 에 실패했습니다.", "error");
    } finally {
      setRevokingTokenID(null);
    }
  };

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex items-center gap-3">
          <div className="w-12 h-12 bg-orange-500/10 border border-orange-500/30 rounded-2xl flex items-center justify-center">
            <Key className="w-6 h-6 text-orange-400" />
          </div>
          <div>
            <h2 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
              Intake <span className="text-orange-400">Tokens</span>
            </h2>
            <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest mt-1">
              외부 시스템 → DevHub Dev Request 수신 인증 자격
            </p>
          </div>
        </motion.div>

        <button
          type="button"
          onClick={() => setShowIssue(true)}
          className="inline-flex items-center gap-2 px-6 py-3 rounded-2xl bg-orange-500 text-white font-black uppercase tracking-widest text-[10px] shadow-xl shadow-orange-500/20 hover:scale-[1.02] active:scale-[0.98] transition-all"
        >
          <Plus className="w-4 h-4" />
          Issue Token
        </button>
      </div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-orange-500/20 border-t-orange-500 rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">
            Loading Intake Tokens...
          </p>
        </div>
      ) : (
        <IntakeTokenTable items={tokens} onRevoke={handleRevoke} revokingTokenID={revokingTokenID} />
      )}

      {showIssue && (
        <IssueIntakeTokenModal
          onClose={() => {
            setShowIssue(false);
            void refresh();
          }}
          onIssued={handleIssued}
        />
      )}
    </div>
  );
}
