"use client";

import { useState, useMemo, FormEvent } from "react";
import { Plug, X, Eye, EyeOff } from "lucide-react";
import { motion } from "framer-motion";
import { integrationService } from "@/lib/services/integration.service";
import type {
  IntegrationProvider,
  IntegrationProviderType,
  IntegrationAuthMode,
  CreateIntegrationProviderInput,
} from "@/lib/services/integration.types";
import {
  VENDOR_PRESETS,
  getVendorPreset,
  KNOWN_CAPABILITIES,
  SDK_VENDORS,
  composeCredentialsRef,
  parseCredentialsRef,
  type WebhookSignatureStrategy,
  type SdkVendor,
} from "@/lib/services/integration-provider-presets";

interface ProviderModalProps {
  /** edit 모드 시 기존 provider, create 모드 시 undefined */
  initial?: IntegrationProvider;
  onClose: () => void;
  onSaved: (provider: IntegrationProvider) => void;
}

const providerTypeOptions: IntegrationProviderType[] = ["alm", "scm", "ci_cd", "doc", "infra"];
const authModeOptions: IntegrationAuthMode[] = ["token", "basic", "oauth2", "app_password", "agent"];
const signatureStrategyOptions: { value: WebhookSignatureStrategy; label: string }[] = [
  { value: "hmac_sha256", label: "HMAC-SHA256 (generic webhook secret)" },
  { value: "provider_sdk", label: "Provider SDK (vendor-bound)" },
  { value: "shared_token", label: "Shared token (plain compare)" },
];

const inputCls =
  "w-full px-4 py-3 rounded-xl bg-muted/30 border border-border text-foreground dark:text-primary-foreground text-sm focus:outline-none focus:border-accent";
const labelCls = "block text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2";

/** 마스킹 토글이 달린 secret 입력 필드. write-only 자격증명 입력에 공통 사용. */
function SecretField({
  id,
  label,
  value,
  onChange,
  show,
  onToggle,
  placeholder,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (v: string) => void;
  show: boolean;
  onToggle: () => void;
  placeholder: string;
}) {
  return (
    <div>
      <label htmlFor={id} className={labelCls}>
        {label}
      </label>
      <div className="relative">
        <input
          id={id}
          type={show ? "text" : "password"}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          className={`${inputCls} font-mono pr-12`}
          autoComplete="off"
        />
        <button
          type="button"
          onClick={onToggle}
          className="absolute right-3 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground"
          aria-label={show ? "Hide secret" : "Show secret"}
        >
          {show ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
        </button>
      </div>
    </div>
  );
}

export function ProviderModal({ initial, onClose, onSaved }: ProviderModalProps) {
  const isEdit = Boolean(initial);
  const parsedInitial = useMemo(
    () => (initial ? parseCredentialsRef(initial.credentials_ref) : null),
    [initial],
  );

  const [vendorPresetId, setVendorPresetId] = useState("custom");
  const [providerKey, setProviderKey] = useState(initial?.provider_key ?? "");
  const [providerType, setProviderType] = useState<IntegrationProviderType>(initial?.provider_type ?? "scm");
  const [displayName, setDisplayName] = useState(initial?.display_name ?? "");
  const [authMode, setAuthMode] = useState<IntegrationAuthMode>(initial?.auth_mode ?? "token");
  const [baseUrl, setBaseUrl] = useState(initial?.base_url ?? "");
  // api_token 은 write-only — edit 모드에서도 빈 값으로 시작 (blank=keep). 설정 여부는
  // initial.api_token_set 로 표시.
  const [apiToken, setApiToken] = useState("");
  const [showApiToken, setShowApiToken] = useState(false);
  // auth_mode 별 구조화 outbound 자격증명. 비밀 외 필드는 기존 값 prefill,
  // authSecret 은 write-only (blank=keep).
  const [authUsername, setAuthUsername] = useState(initial?.auth_username ?? "");
  const [authClientId, setAuthClientId] = useState(initial?.auth_client_id ?? "");
  const [authTokenUrl, setAuthTokenUrl] = useState(initial?.auth_token_url ?? "");
  const [authSecret, setAuthSecret] = useState("");
  const [showAuthSecret, setShowAuthSecret] = useState(false);
  const [enabled, setEnabled] = useState(initial?.enabled ?? true);

  // 가이드 자격증명 입력 (#1) — strategy + vendor + secret 를 분리 입력받아 조합.
  const [sigStrategy, setSigStrategy] = useState<WebhookSignatureStrategy>(parsedInitial?.strategy ?? "hmac_sha256");
  const [sdkVendor, setSdkVendor] = useState<SdkVendor>(parsedInitial?.sdkVendor ?? "gitea");
  const [secret, setSecret] = useState("");
  const [showSecret, setShowSecret] = useState(false);

  // capabilities 체크박스 (#4) — 표준 어휘 + 기존 provider 의 비표준 항목 union.
  const capabilityOptions = useMemo(() => {
    const extras = (initial?.capabilities ?? []).filter(
      (c) => !(KNOWN_CAPABILITIES as readonly string[]).includes(c),
    );
    return [...KNOWN_CAPABILITIES, ...extras];
  }, [initial]);
  const [capabilities, setCapabilities] = useState<string[]>(initial?.capabilities ?? []);

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 연결 테스트 (#5)
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ ok: boolean; msg: string } | null>(null);

  const handleTestConnection = async () => {
    if (!baseUrl.trim()) return;
    setTesting(true);
    setTestResult(null);
    try {
      const r = await integrationService.testConnection(baseUrl.trim());
      setTestResult(
        r.reachable
          ? { ok: true, msg: `Reachable (HTTP ${r.status_code ?? "?"}, ${r.latency_ms ?? "?"}ms)` }
          : { ok: false, msg: `Unreachable: ${r.error ?? "no response"}` },
      );
    } catch (err) {
      setTestResult({ ok: false, msg: err instanceof Error ? err.message : "test failed" });
    } finally {
      setTesting(false);
    }
  };

  // vendor 템플릿 (#3) — 선택 시 type/auth/strategy/vendor/capabilities 자동 채움.
  const applyVendorPreset = (id: string) => {
    setVendorPresetId(id);
    if (id === "custom") return;
    const preset = getVendorPreset(id);
    setProviderType(preset.providerType);
    setAuthMode(preset.authMode);
    setSigStrategy(preset.signatureStrategy);
    if (preset.sdkVendor) setSdkVendor(preset.sdkVendor);
    setCapabilities(preset.capabilities);
  };

  const toggleCapability = (cap: string, checked: boolean) => {
    setCapabilities((prev) =>
      checked ? Array.from(new Set([...prev, cap])) : prev.filter((c) => c !== cap),
    );
  };

  // auth_mode 에 맞는 outbound 자격증명 payload 만 구성. 비밀 외 필드는 trim 그대로
  // (blank → 클리어), secret(api_token/auth_secret)은 blank=keep (omit).
  const outboundAuthPayload = (): Partial<CreateIntegrationProviderInput> => {
    switch (authMode) {
      case "token":
        return { api_token: apiToken.trim() || undefined };
      case "basic":
      case "app_password":
        return { auth_username: authUsername.trim(), auth_secret: authSecret.trim() || undefined };
      case "oauth2":
        return {
          auth_client_id: authClientId.trim(),
          auth_token_url: authTokenUrl.trim(),
          auth_secret: authSecret.trim() || undefined,
        };
      case "agent":
        return { auth_username: authUsername.trim() };
      default:
        return {};
    }
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      let saved: IntegrationProvider;
      if (isEdit && initial) {
        // secret 입력 시에만 credentials_ref 재조합, blank 면 기존 유지 (omit).
        const nextCredentials = secret.trim()
          ? composeCredentialsRef(sigStrategy, sdkVendor, secret)
          : undefined;
        saved = await integrationService.updateProvider(initial.provider_id, {
          enabled,
          display_name: displayName.trim() || undefined,
          credentials_ref: nextCredentials,
          capabilities,
          base_url: baseUrl.trim(),
          ...outboundAuthPayload(), // auth_mode 별 outbound 자격증명 (secret blank=keep)
        });
      } else {
        if (!providerKey.trim()) {
          setError("provider_key 는 필수입니다.");
          return;
        }
        if (!displayName.trim()) {
          setError("display_name 은 필수입니다.");
          return;
        }
        if (!secret.trim()) {
          setError("secret 은 필수입니다 (자격증명 전략에 사용).");
          return;
        }
        saved = await integrationService.createProvider({
          provider_key: providerKey.trim(),
          provider_type: providerType,
          display_name: displayName.trim(),
          auth_mode: authMode,
          credentials_ref: composeCredentialsRef(sigStrategy, sdkVendor, secret),
          capabilities,
          base_url: baseUrl.trim() || undefined,
          ...outboundAuthPayload(), // auth_mode 별 outbound 자격증명
        });
      }
      onSaved(saved);
      onClose();
    } catch (err) {
      const msg = err instanceof Error ? err.message : "저장에 실패했습니다.";
      setError(msg);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
    >
      <motion.div
        initial={{ opacity: 0, y: 20, scale: 0.98 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        exit={{ opacity: 0, y: 20, scale: 0.98 }}
        className="glass border-border rounded-3xl w-full max-w-2xl max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between p-6 border-b border-border/60">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-accent/10 flex items-center justify-center border border-accent/20">
              <Plug className="w-5 h-5 text-accent" />
            </div>
            <div>
              <h3 className="text-lg font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                {isEdit ? "Edit Provider" : "Register Provider"}
              </h3>
              <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest mt-0.5">
                {isEdit ? `provider_key: ${initial?.provider_key}` : "신규 외부 시스템 연동 등록"}
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-muted/50 transition-colors"
            aria-label="Close"
          >
            <X className="w-4 h-4 text-muted-foreground" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          {!isEdit && (
            <>
              <div>
                <label htmlFor="vendor_preset" className={labelCls}>
                  Vendor Template
                </label>
                <select
                  id="vendor_preset"
                  value={vendorPresetId}
                  onChange={(e) => applyVendorPreset(e.target.value)}
                  className={inputCls}
                >
                  {VENDOR_PRESETS.map((p) => (
                    <option key={p.id} value={p.id}>{p.label}</option>
                  ))}
                </select>
                <p className="text-[10px] text-muted-foreground mt-1.5">
                  vendor 선택 시 type / auth / signature / capabilities 가 자동 설정됩니다. Custom 은 수동 입력.
                </p>
              </div>

              <div>
                <label htmlFor="provider_key" className={labelCls}>
                  Provider Key *
                </label>
                <input
                  id="provider_key"
                  type="text"
                  value={providerKey}
                  onChange={(e) => setProviderKey(e.target.value)}
                  placeholder={getVendorPreset(vendorPresetId).providerKeyHint}
                  className={`${inputCls} font-mono`}
                  required
                />
                <p className="text-[10px] text-muted-foreground mt-1.5">URL-safe 식별자. 발급 후 변경 불가.</p>
              </div>
            </>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="provider_type" className={labelCls}>Type *</label>
              <select
                id="provider_type"
                value={providerType}
                onChange={(e) => setProviderType(e.target.value as IntegrationProviderType)}
                disabled={isEdit}
                className={`${inputCls} disabled:opacity-60`}
              >
                {providerTypeOptions.map((t) => (
                  <option key={t} value={t}>{t}</option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="auth_mode" className={labelCls}>Auth Mode *</label>
              <select
                id="auth_mode"
                value={authMode}
                onChange={(e) => setAuthMode(e.target.value as IntegrationAuthMode)}
                disabled={isEdit}
                className={`${inputCls} disabled:opacity-60`}
              >
                {authModeOptions.map((m) => (
                  <option key={m} value={m}>{m}</option>
                ))}
              </select>
            </div>
          </div>

          <div>
            <label htmlFor="display_name" className={labelCls}>Display Name *</label>
            <input
              id="display_name"
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Gitea Main / Jenkins Production"
              className={inputCls}
            />
          </div>

          {/* endpoint/base URL (#2) + 연결 테스트 (#5) */}
          <div>
            <label htmlFor="base_url" className={labelCls}>Base URL</label>
            <div className="flex gap-2">
              <input
                id="base_url"
                type="url"
                value={baseUrl}
                onChange={(e) => { setBaseUrl(e.target.value); setTestResult(null); }}
                placeholder={getVendorPreset(vendorPresetId).baseUrlHint ?? "https://external-system.example.com"}
                className={`${inputCls} font-mono`}
              />
              <button
                type="button"
                onClick={() => void handleTestConnection()}
                disabled={!baseUrl.trim() || testing}
                className="shrink-0 px-4 py-3 rounded-xl border border-accent/40 bg-accent/10 text-accent font-black uppercase tracking-widest text-[10px] hover:bg-accent/20 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              >
                {testing ? "Testing…" : "Test"}
              </button>
            </div>
            {testResult ? (
              <p className={`text-[10px] mt-1.5 font-bold ${testResult.ok ? "text-emerald-500" : "text-destructive"}`}>
                {testResult.ok ? "✓ " : "✗ "}{testResult.msg}
              </p>
            ) : (
              <p className="text-[10px] text-muted-foreground mt-1.5">
                외부 시스템 endpoint (sync 대상). webhook 전용이면 비워둘 수 있습니다.
              </p>
            )}
          </div>

          {/* Outbound Auth — auth_mode 에 맞춰 자격증명 입력 필드가 바뀐다. repo/issue/PR
              sync·pull 인증용 (inbound webhook secret 과 별개). auth_mode 는 등록 시
              결정되며 edit 에서 변경 불가. */}
          <div className="space-y-3 rounded-2xl border border-border/60 bg-muted/10 p-4">
            <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">
              Outbound Auth · {authMode}
            </p>

            {authMode === "token" && (
              <SecretField
                id="api_token"
                label={`API Token${isEdit && initial?.api_token_set ? " (set — blank=keep)" : ""}`}
                value={apiToken}
                onChange={setApiToken}
                show={showApiToken}
                onToggle={() => setShowApiToken((v) => !v)}
                placeholder={isEdit && initial?.api_token_set ? "•••••• (set, leave blank to keep)" : "Personal Access Token (outbound sync 인증)"}
              />
            )}

            {(authMode === "basic" || authMode === "app_password") && (
              <>
                <div>
                  <label htmlFor="auth_username" className={labelCls}>Username *</label>
                  <input
                    id="auth_username"
                    type="text"
                    value={authUsername}
                    onChange={(e) => setAuthUsername(e.target.value)}
                    placeholder="service-account"
                    className={`${inputCls} font-mono`}
                  />
                </div>
                <SecretField
                  id="auth_secret"
                  label={`${authMode === "app_password" ? "App Password" : "Password"}${isEdit && initial?.auth_secret_set ? " (set — blank=keep)" : ""}`}
                  value={authSecret}
                  onChange={setAuthSecret}
                  show={showAuthSecret}
                  onToggle={() => setShowAuthSecret((v) => !v)}
                  placeholder={isEdit && initial?.auth_secret_set ? "•••••• (set, leave blank to keep)" : authMode === "app_password" ? "application password" : "account password"}
                />
              </>
            )}

            {authMode === "oauth2" && (
              <>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label htmlFor="auth_client_id" className={labelCls}>Client ID *</label>
                    <input
                      id="auth_client_id"
                      type="text"
                      value={authClientId}
                      onChange={(e) => setAuthClientId(e.target.value)}
                      placeholder="oauth2 client_id"
                      className={`${inputCls} font-mono`}
                    />
                  </div>
                  <div>
                    <label htmlFor="auth_token_url" className={labelCls}>Token URL *</label>
                    <input
                      id="auth_token_url"
                      type="url"
                      value={authTokenUrl}
                      onChange={(e) => setAuthTokenUrl(e.target.value)}
                      placeholder="https://idp.example.com/oauth/token"
                      className={`${inputCls} font-mono`}
                    />
                  </div>
                </div>
                <SecretField
                  id="auth_secret"
                  label={`Client Secret${isEdit && initial?.auth_secret_set ? " (set — blank=keep)" : ""}`}
                  value={authSecret}
                  onChange={setAuthSecret}
                  show={showAuthSecret}
                  onToggle={() => setShowAuthSecret((v) => !v)}
                  placeholder={isEdit && initial?.auth_secret_set ? "•••••• (set, leave blank to keep)" : "oauth2 client_secret (client-credentials grant)"}
                />
              </>
            )}

            {authMode === "agent" && (
              <div>
                <label htmlFor="auth_username" className={labelCls}>Agent Identifier</label>
                <input
                  id="auth_username"
                  type="text"
                  value={authUsername}
                  onChange={(e) => setAuthUsername(e.target.value)}
                  placeholder="agent id"
                  className={`${inputCls} font-mono`}
                />
                <p className="text-[10px] text-muted-foreground mt-1.5">
                  agent 모드는 별도 agent 프로세스가 외부 시스템 인증을 수행합니다. 서버 직접 sync 에는 사용되지 않습니다.
                </p>
              </div>
            )}

            <p className="text-[10px] text-muted-foreground">
              외부 시스템 sync/pull 인증용. webhook 전용 provider 면 비워둘 수 있습니다. 아래 webhook secret 과 별개.
            </p>
          </div>

          {/* 가이드 자격증명 (#1) — strategy + secret 분리 입력 → credentials_ref 자동 조합 */}
          <div className="space-y-3 rounded-2xl border border-border/60 bg-muted/10 p-4">
            <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Webhook Credentials</p>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="sig_strategy" className={labelCls}>Signature Strategy *</label>
                <select
                  id="sig_strategy"
                  value={sigStrategy}
                  onChange={(e) => setSigStrategy(e.target.value as WebhookSignatureStrategy)}
                  className={inputCls}
                >
                  {signatureStrategyOptions.map((o) => (
                    <option key={o.value} value={o.value}>{o.label}</option>
                  ))}
                </select>
              </div>
              {sigStrategy === "provider_sdk" && (
                <div>
                  <label htmlFor="sdk_vendor" className={labelCls}>SDK Vendor *</label>
                  <select
                    id="sdk_vendor"
                    value={sdkVendor}
                    onChange={(e) => setSdkVendor(e.target.value as SdkVendor)}
                    className={inputCls}
                  >
                    {SDK_VENDORS.map((v) => (
                      <option key={v} value={v}>{v}</option>
                    ))}
                  </select>
                </div>
              )}
            </div>
            <div>
              <label htmlFor="secret" className={labelCls}>
                Secret {isEdit ? "(blank = keep current)" : "*"}
              </label>
              <div className="relative">
                <input
                  id="secret"
                  type={showSecret ? "text" : "password"}
                  value={secret}
                  onChange={(e) => setSecret(e.target.value)}
                  placeholder={isEdit && parsedInitial?.hasSecret ? "•••••• (set, leave blank to keep)" : "webhook signing secret / token"}
                  className={`${inputCls} font-mono pr-12`}
                  autoComplete="off"
                />
                <button
                  type="button"
                  onClick={() => setShowSecret((v) => !v)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground"
                  aria-label={showSecret ? "Hide secret" : "Show secret"}
                >
                  {showSecret ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>
              <p className="text-[10px] text-muted-foreground mt-1.5">
                저장 시 <code>credentials_ref</code> 로 자동 조합됩니다 (내부 인코딩 직접 입력 불필요).
              </p>
            </div>
          </div>

          {/* capabilities 체크박스 (#4) */}
          <div>
            <label className={labelCls}>Capabilities</label>
            <div className="flex flex-wrap gap-2">
              {capabilityOptions.map((cap) => {
                const checked = capabilities.includes(cap);
                return (
                  <label
                    key={cap}
                    className={`flex items-center gap-2 px-3 py-2 rounded-xl border text-xs font-bold cursor-pointer transition-colors ${
                      checked ? "border-accent bg-accent/10 text-foreground" : "border-border bg-muted/20 text-muted-foreground"
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={(e) => toggleCapability(cap, e.target.checked)}
                      className="w-3.5 h-3.5 accent-accent"
                    />
                    {cap}
                  </label>
                );
              })}
            </div>
          </div>

          {isEdit && (
            <div className="flex items-center gap-3 p-4 rounded-xl bg-muted/20 border border-border/60">
              <input
                id="enabled"
                type="checkbox"
                checked={enabled}
                onChange={(e) => setEnabled(e.target.checked)}
                className="w-4 h-4 accent-accent"
              />
              <label htmlFor="enabled" className="text-xs font-bold text-foreground dark:text-primary-foreground cursor-pointer">
                Enabled (수신/sync 활성)
              </label>
            </div>
          )}

          {error && (
            <div className="p-3 rounded-xl bg-destructive/10 border border-destructive/30 text-destructive text-xs font-bold">
              {error}
            </div>
          )}

          <div className="flex gap-3 pt-4 border-t border-border/60">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-6 py-3 rounded-2xl border border-border text-foreground dark:text-primary-foreground font-black uppercase tracking-widest text-[10px] hover:bg-muted/30 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="flex-1 px-6 py-3 rounded-2xl bg-primary text-primary-foreground font-black uppercase tracking-widest text-[10px] shadow-xl hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {submitting ? "Saving..." : isEdit ? "Save" : "Register"}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
