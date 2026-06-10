"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { motion } from "framer-motion";
import { TestTube, Send, RefreshCcw, Inbox, AlertTriangle, CheckCircle2, ExternalLink, Loader2 } from "lucide-react";
import { DashboardHeader } from "@/shared/ui-foundation/components/DashboardHeader";
import { PageEmpty, PageError, PageLoading } from "@/shared/ui-foundation/components/PageState";
import { useToast } from "@/shared/ui-foundation/components/Toast";
import { useStore } from "@/lib/store";
import { isSystemAdmin } from "@/domain/auth-session/service/role-routing";
import { devRequestService } from "@/domain/dev-request/service/dev_request.service";
import { devRequestTokenService } from "@/domain/dev-request/service/dev_request_token.service";
import { DevRequest } from "@/domain/dev-request/schema/dev_request.types";
import { IssuedDevRequestIntakeToken } from "@/domain/dev-request/schema/dev_request_token.types";
import { toUserErrorMessage } from "@/shared/utils/error-message";

type SubmissionSource = "quick" | "intake";

interface SubmissionRecord {
  source: SubmissionSource;
  dreq: DevRequest;
  token?: IssuedDevRequestIntakeToken;
  submittedAt: number;
}

const initialForm = {
  title: "[Test] Reception flow from external system",
  details: "Auto-submitted from /admin/reception-test for QA verification.",
  requester: "Reception Test Panel",
  assignee_user_id: "tester",
  external_ref: "",
};

// client_label regex: `^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$` (dev_request_intake_tokens.label_format check).
// Date.now() is a numeric string → safe; toISOString() would include 'T'/':'/'.' → rejected.
const intakeTokenLabel = (): string => `reception-test-${Date.now()}`;

export default function ReceptionTestPage() {
  const { toast } = useToast();
  const actor = useStore((s) => s.actor);
  const isAdmin = isSystemAdmin(actor?.role);

  const [form, setForm] = useState(initialForm);
  const [quickBusy, setQuickBusy] = useState(false);
  const [intakeBusy, setIntakeBusy] = useState(false);
  const [tokenBusy, setTokenBusy] = useState(false);
  const [currentToken, setCurrentToken] = useState<IssuedDevRequestIntakeToken | null>(null);
  const [history, setHistory] = useState<SubmissionRecord[]>([]);
  const [recentError, setRecentError] = useState<string | null>(null);

  // Quick test: uses backend `DEVHUB_DEBUG_TOKEN` bypass path.
  const handleQuickTest = useCallback(async () => {
    setQuickBusy(true);
    setRecentError(null);
    try {
      const dreq = await devRequestService.createDebugDreq(form.assignee_user_id);
      setHistory((prev) => [{ source: "quick", dreq, submittedAt: Date.now() } as SubmissionRecord, ...prev].slice(0, 20));
      toast(`Quick test: DREQ ${dreq.id} created (status=${dreq.status})`, "success");
    } catch (err) {
      setRecentError(toUserErrorMessage(err, "Quick test failed"));
      toast(toUserErrorMessage(err, "Quick test failed"), "error");
    } finally {
      setQuickBusy(false);
    }
  }, [form.assignee_user_id, toast]);

  // Issue a fresh intake token via admin API. The token's plain value is returned
  // exactly once and used to POST /api/v1/dev-requests.
  const handleIssueToken = useCallback(async () => {
    setTokenBusy(true);
    setRecentError(null);
    try {
      const issued = await devRequestTokenService.issue({
        allowed_ips: ["127.0.0.1/32"],
        client_label: intakeTokenLabel(),
        source_system: "reception_test_panel",
      });
      setCurrentToken(issued);
      toast(`Intake token issued (token_id=${issued.token_id})`, "success");
    } catch (err) {
      setRecentError(toUserErrorMessage(err, "Token issuance failed"));
      toast(toUserErrorMessage(err, "Token issuance failed"), "error");
    } finally {
      setTokenBusy(false);
    }
  }, [toast]);

  // Full intake test: requires a token in state. Submits via the public intake
  // endpoint (no auth header — token IS the auth).
  const handleIntakeSubmit = useCallback(async () => {
    if (!currentToken) {
      toast("Issue an intake token first", "warning");
      return;
    }
    if (!form.title.trim() || !form.requester.trim()) {
      toast("title + requester are required", "warning");
      return;
    }
    setIntakeBusy(true);
    setRecentError(null);
    try {
      const externalRef = form.external_ref.trim() ||
        `RECEP-TEST-${Date.now()}-${Math.floor(Math.random() * 1000)}`;

      // Public intake endpoint — bypass the standard auth interceptor and pass the
      // intake token in the Authorization header directly. The backend
      // requireIntakeToken middleware validates the token.
      const res = await fetch("/api/v1/dev-requests", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${currentToken.plain_token}`,
        },
        body: JSON.stringify({
          title: form.title.trim(),
          details: form.details.trim(),
          requester: form.requester.trim(),
          assignee_user_id: form.assignee_user_id.trim() || undefined,
          external_ref: externalRef,
        }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(`intake ${res.status}: ${body?.error ?? res.statusText}`);
      }
      const body = await res.json();
      const dreq = body?.data as DevRequest;
      setHistory((prev) => [
        { source: "intake", dreq, token: currentToken, submittedAt: Date.now() } as SubmissionRecord,
        ...prev,
      ].slice(0, 20));
      toast(`Intake submit: DREQ ${dreq.id} created (external_ref=${externalRef})`, "success");
    } catch (err) {
      setRecentError(toUserErrorMessage(err, "Intake submit failed"));
      toast(toUserErrorMessage(err, "Intake submit failed"), "error");
    } finally {
      setIntakeBusy(false);
    }
  }, [currentToken, form, toast]);

  // Refresh the test history by re-listing recent test submissions. Filters
  // for source markers used by both flows ("debug_system" / "Reception Test Panel").
  const refreshHistory = useCallback(async () => {
    setRecentError(null);
    try {
      const all = await devRequestService.list({ limit: 50 });
      const filtered = all.data
        .filter((d) =>
          d.requester?.toLowerCase().includes("reception test") ||
          d.requester?.toLowerCase().includes("debug"),
        )
        .slice(0, 20)
        .map<SubmissionRecord>((d) => ({
          source: d.requester?.toLowerCase().includes("reception test") ? "intake" : "quick",
          dreq: d,
          submittedAt: new Date(d.created_at).getTime(),
        }));
      setHistory(filtered);
    } catch (err) {
      setRecentError(toUserErrorMessage(err, "History refresh failed"));
    }
  }, []);

  useEffect(() => {
    if (!isAdmin) return;
    void refreshHistory();
  }, [isAdmin, refreshHistory]);

  if (!actor) {
    return <PageLoading label="Loading session..." />;
  }
  if (!isAdmin) {
    return (
      <div className="pb-20 space-y-6">
        <DashboardHeader
          titlePrefix="Reception"
          titleGradient="Test"
          subtitle="system_admin only. Ask your admin to grant system_admin role."
        />
        <PageError message="Insufficient permissions: system_admin required." />
      </div>
    );
  }

  return (
    <div className="pb-20 space-y-10">
      <DashboardHeader
        titlePrefix="Reception"
        titleGradient="Test"
        subtitle="QA simulation for the external-system intake flow. Quick debug bypass + full intake-token round trip."
      />

      <div className="rounded-2xl border border-orange-500/30 bg-orange-500/10 px-5 py-3 text-xs font-bold text-orange-700 dark:text-orange-300 flex items-start gap-3">
        <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
        <div className="space-y-1">
          <p>테스트 페이지 — 운영 데이터에 dev request 가 생성됩니다. cleanup 은 Admin Settings → Dev Requests 에서 close / reject 하세요.</p>
          <p className="opacity-80 font-normal">
            external_ref 가 동일하면 idempotent (같은 ref 재제출 시 기존 DREQ 반환). 매번 다른 ref 가 자동 생성됩니다.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Card 1: Quick Test (debug bypass) */}
        <TestCard
          title="Quick Test"
          subtitle="DEVHUB_DEBUG_TOKEN bypass path. 1-click, no intake token required."
          icon={<TestTube className="w-5 h-5" />}
          accent="text-orange-700 dark:text-orange-300"
        >
          <div className="space-y-3 text-sm">
            <p className="text-muted-foreground">
              내부 디버그 토큰 (<code className="px-1.5 py-0.5 bg-muted rounded text-xs">debug-token-bypass-dev</code>) 으로 빠르게 dev request 1건 생성. assignee 는 본인 또는 지정.
            </p>
            <FieldRow label="assignee_user_id">
              <input
                type="text"
                value={form.assignee_user_id}
                onChange={(e) => setForm({ ...form, assignee_user_id: e.target.value })}
                className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm font-mono"
                placeholder="tester / alice / (blank = actor)"
              />
            </FieldRow>
            <ActionButton
              onClick={handleQuickTest}
              disabled={quickBusy}
              icon={quickBusy ? <Loader2 className="w-4 h-4 animate-spin" /> : <TestTube className="w-4 h-4" />}
            >
              {quickBusy ? "Creating..." : "Create Debug DREQ"}
            </ActionButton>
          </div>
        </TestCard>

        {/* Card 2: Full Intake Test */}
        <TestCard
          title="Full Intake Test"
          subtitle="Real intake-token round trip (issue token → POST /dev-requests)."
          icon={<Inbox className="w-5 h-5" />}
          accent="text-sky-700 dark:text-sky-300"
        >
          <div className="space-y-3 text-sm">
            <div className="rounded-lg border border-border/60 bg-muted/30 p-3 space-y-1.5">
              <div className="flex items-center justify-between gap-2">
                <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                  Step 1 — Issue intake token
                </p>
                <button
                  type="button"
                  onClick={handleIssueToken}
                  disabled={tokenBusy}
                  className="text-[10px] font-black uppercase tracking-widest px-2.5 py-1 rounded-md bg-primary/10 text-primary hover:bg-primary/20 disabled:opacity-50 transition-all flex items-center gap-1.5"
                >
                  {tokenBusy ? <Loader2 className="w-3 h-3 animate-spin" /> : <Send className="w-3 h-3" />}
                  {tokenBusy ? "Issuing..." : "Issue Token"}
                </button>
              </div>
              {currentToken ? (
                <div className="space-y-1 text-xs">
                  <p className="font-mono break-all">
                    <span className="text-muted-foreground">token_id:</span> {currentToken.token_id}
                  </p>
                  <p className="font-mono break-all text-orange-700 dark:text-orange-300">
                    <span className="text-muted-foreground">plain:</span> {currentToken.plain_token}
                  </p>
                  <p className="text-[10px] text-muted-foreground">
                    ⚠ plain value is shown only once. Submit immediately or re-issue.
                  </p>
                </div>
              ) : (
                <p className="text-xs text-muted-foreground italic">no token issued yet</p>
              )}
            </div>

            <div className="rounded-lg border border-border/60 bg-muted/30 p-3 space-y-2">
              <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                Step 2 — Submit reception payload
              </p>
              <FieldRow label="title *">
                <input
                  type="text"
                  value={form.title}
                  onChange={(e) => setForm({ ...form, title: e.target.value })}
                  className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm"
                />
              </FieldRow>
              <FieldRow label="requester *">
                <input
                  type="text"
                  value={form.requester}
                  onChange={(e) => setForm({ ...form, requester: e.target.value })}
                  className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm"
                />
              </FieldRow>
              <FieldRow label="details">
                <textarea
                  value={form.details}
                  onChange={(e) => setForm({ ...form, details: e.target.value })}
                  className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm h-16"
                />
              </FieldRow>
              <FieldRow label="external_ref (auto if empty)">
                <input
                  type="text"
                  value={form.external_ref}
                  onChange={(e) => setForm({ ...form, external_ref: e.target.value })}
                  className="w-full bg-background border border-border rounded-md px-3 py-2 text-sm font-mono"
                  placeholder="RECEP-TEST-..."
                />
              </FieldRow>
              <ActionButton
                onClick={handleIntakeSubmit}
                disabled={intakeBusy || !currentToken}
                icon={intakeBusy ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
                variant="primary"
              >
                {intakeBusy ? "Submitting..." : "Submit via Intake Token"}
              </ActionButton>
            </div>
          </div>
        </TestCard>
      </div>

      {/* Recent submissions */}
      <TestCard
        title="Recent Test Submissions"
        subtitle={`최근 20건의 Reception Test / Debug DREQ — 새로고침 버튼으로 backend 에서 재조회`}
        icon={<RefreshCcw className="w-5 h-5" />}
        accent="text-emerald-700 dark:text-emerald-300"
      >
        <div className="flex justify-end mb-3">
          <button
            type="button"
            onClick={refreshHistory}
            className="text-xs font-bold px-3 py-1.5 rounded-md bg-muted/40 hover:bg-muted/70 flex items-center gap-1.5"
          >
            <RefreshCcw className="w-3.5 h-3.5" /> Refresh
          </button>
        </div>
        {recentError && <PageError message={recentError} onRetry={refreshHistory} />}
        {!recentError && history.length === 0 && (
          <PageEmpty message="아직 제출된 테스트 없음. Quick Test 또는 Full Intake Test 로 dev request 1건 생성 후 refresh." />
        )}
        {!recentError && history.length > 0 && (
          <div className="space-y-2">
            {history.map((rec) => (
              <SubmissionRow key={`${rec.source}-${rec.dreq.id}-${rec.submittedAt}`} rec={rec} />
            ))}
          </div>
        )}
      </TestCard>
    </div>
  );
}

function TestCard({
  title,
  subtitle,
  icon,
  accent,
  children,
}: {
  title: string;
  subtitle: string;
  icon: React.ReactNode;
  accent: string;
  children: React.ReactNode;
}) {
  return (
    <motion.section
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="glass-card p-6 space-y-4"
    >
      <div className="flex items-start gap-3">
        <div className={`p-2 rounded-xl bg-primary/10 ${accent}`}>{icon}</div>
        <div className="flex-1 min-w-0">
          <h2 className="text-lg font-black tracking-tighter">{title}</h2>
          <p className="text-xs text-muted-foreground mt-0.5">{subtitle}</p>
        </div>
      </div>
      {children}
    </motion.section>
  );
}

function FieldRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block space-y-1">
      <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
        {label}
      </span>
      {children}
    </label>
  );
}

function ActionButton({
  onClick,
  disabled,
  icon,
  children,
  variant = "secondary",
}: {
  onClick: () => void;
  disabled?: boolean;
  icon: React.ReactNode;
  children: React.ReactNode;
  variant?: "primary" | "secondary";
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={
        variant === "primary"
          ? "w-full inline-flex items-center justify-center gap-2 rounded-md bg-primary text-primary-foreground px-4 py-2 text-sm font-bold hover:bg-primary/90 disabled:opacity-50 transition-all"
          : "w-full inline-flex items-center justify-center gap-2 rounded-md bg-muted/60 hover:bg-muted px-4 py-2 text-sm font-bold disabled:opacity-50 transition-all"
      }
    >
      {icon}
      {children}
    </button>
  );
}

function SubmissionRow({ rec }: { rec: SubmissionRecord }) {
  const dreq = rec.dreq;
  return (
    <div className="flex items-center gap-3 rounded-lg border border-border/40 bg-background/40 p-3 text-sm">
      <div className="flex-shrink-0">
        {rec.source === "intake" ? (
          <Inbox className="w-4 h-4 text-sky-700 dark:text-sky-300" />
        ) : (
          <TestTube className="w-4 h-4 text-orange-700 dark:text-orange-300" />
        )}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-bold truncate">{dreq.title}</span>
          <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground px-1.5 py-0.5 rounded bg-muted/50">
            {dreq.status}
          </span>
        </div>
        <p className="text-xs text-muted-foreground font-mono truncate">
          {dreq.id} · {dreq.requester}
          {dreq.external_ref && ` · ext:${dreq.external_ref}`}
        </p>
      </div>
      <Link
        href={`/dev-requests`}
        className="text-[10px] font-black uppercase tracking-widest text-primary hover:underline flex items-center gap-1"
      >
        <ExternalLink className="w-3 h-3" /> view
      </Link>
    </div>
  );
}
