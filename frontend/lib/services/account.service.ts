/**
 * Account Service
 *
 * Self-service (PR-L3, DEC-2=B): updateMyPassword drives the Kratos public
 * settings flow directly from the browser; the Kratos session cookie does
 * the authentication. Backend is not involved.
 *
 * Admin operations (PR-S3): issueAccount / forceResetPassword /
 * disableAccount / unlockAccount call backend /api/v1/accounts endpoints
 * which proxy Kratos admin API. Backend ownership keeps the Kratos admin
 * URL off the browser and lands every action in DevHub audit_logs.
 */

import { apiClient, ApiError } from "@/lib/services/api-client";

const KRATOS_PUBLIC_URL = (
  process.env.NEXT_PUBLIC_KRATOS_PUBLIC_URL ?? "http://localhost:4433"
).replace(/\/$/, "");

export type SettingsFlowErrorCode =
  | "REAUTH_REQUIRED"
  | "UNAUTHENTICATED"
  | "VALIDATION"
  | "FLOW_INIT_FAILED"
  | "SUBMIT_FAILED";

// SettingsFlowError carries enough context for the UI to either show a
// validation message inline or send the user back through Kratos for a
// privileged-session re-auth (kratos.yaml privileged_session_max_age=15m).
export class SettingsFlowError extends Error {
  constructor(
    public code: SettingsFlowErrorCode,
    public redirectURL: string | null,
    message?: string,
  ) {
    super(message ?? code);
    this.name = "SettingsFlowError";
  }
}

interface KratosUiNode {
  attributes?: { name?: string; value?: unknown };
  messages?: { text: string }[];
}

interface KratosUi {
  action?: string;
  method?: string;
  nodes?: KratosUiNode[];
  messages?: { text: string }[];
}

interface KratosFlow {
  id: string;
  ui: KratosUi;
}

interface KratosErrorEnvelope {
  redirect_browser_to?: string;
  error?: { id?: string; reason?: string; message?: string };
  ui?: KratosUi;
}

export interface AccountInfo {
  id: number;
  user_id: string;
  login_id: string;
  status: 'active' | 'disabled' | 'locked' | 'password_reset_required';
  last_login_at?: string;
}

class AccountService {
  /**
   * User self-service: change own password via Kratos settings flow.
   *
   * The current password is not sent to the settings POST — Kratos uses the
   * privileged session cookie as proof. If that session is older than 15m
   * (privileged_session_max_age), Kratos returns a redirect that re-runs the
   * login flow with `refresh=true`; we surface that as REAUTH_REQUIRED so
   * the UI can hand the user off to /login. After re-auth they can retry.
   */
  async updateMyPassword(currentPass: string, newPass: string): Promise<void> {
    void currentPass;
    const flow = await this.fetchSettingsFlow();
    const csrfToken = this.extractCsrfToken(flow.ui);
    const action = flow.ui.action ?? `${KRATOS_PUBLIC_URL}/self-service/settings?flow=${encodeURIComponent(flow.id)}`;
    await this.submitPasswordChange(action, csrfToken, newPass);
  }

  private async fetchSettingsFlow(): Promise<KratosFlow> {
    let res: Response;
    try {
      res = await fetch(`${KRATOS_PUBLIC_URL}/self-service/settings/browser`, {
        credentials: "include",
        headers: { Accept: "application/json" },
      });
    } catch (err) {
      throw new SettingsFlowError(
        "FLOW_INIT_FAILED",
        null,
        err instanceof Error ? err.message : "Network error",
      );
    }

    if (res.status === 401 || res.status === 403) {
      const body = (await res.json().catch(() => null)) as KratosErrorEnvelope | null;
      if (body?.redirect_browser_to) {
        throw new SettingsFlowError(
          "REAUTH_REQUIRED",
          body.redirect_browser_to,
          "Re-authentication required (privileged session expired)",
        );
      }
      throw new SettingsFlowError(
        "UNAUTHENTICATED",
        null,
        body?.error?.reason ?? "Not signed in to Kratos",
      );
    }
    if (!res.ok) {
      throw new SettingsFlowError("FLOW_INIT_FAILED", null, `HTTP ${res.status}`);
    }
    return (await res.json()) as KratosFlow;
  }

  private async submitPasswordChange(action: string, csrfToken: string, newPassword: string): Promise<void> {
    let res: Response;
    try {
      res = await fetch(action, {
        method: "POST",
        credentials: "include",
        // redirect: "manual" prevents fetch from silently following Kratos's
        // 303 hand-off to a fresh login flow when the settings flow is
        // expired or the privileged session lapsed mid-submission. Without
        // this, the followed response can be 200 and the caller would
        // incorrectly report success.
        redirect: "manual",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          csrf_token: csrfToken,
          method: "password",
          password: newPassword,
        }),
      });
    } catch (err) {
      throw new SettingsFlowError(
        "SUBMIT_FAILED",
        null,
        err instanceof Error ? err.message : "Network error",
      );
    }

    // redirect: "manual" surfaces 3xx responses as an opaque-redirect Response
    // (status 0, type "opaqueredirect"). The browser cannot read the target,
    // but its mere presence means Kratos punted us elsewhere — treat as
    // re-auth required so the UI can send the user back through /login.
    if (res.type === "opaqueredirect" || (res.status >= 300 && res.status < 400)) {
      throw new SettingsFlowError(
        "REAUTH_REQUIRED",
        null,
        "Settings flow redirected; re-authentication required",
      );
    }

    if (res.ok) {
      // Even with 200 OK Kratos signals success only when state==="success".
      // A 200 carrying state==="show_form" means the flow is still pending
      // (e.g., validation messages were attached without HTTP error code).
      const successBody = (await res.json().catch(() => null)) as KratosFlow & { state?: string } | null;
      if (successBody?.state === "success") return;
      if (successBody?.ui) {
        const messages = this.collectMessages(successBody.ui);
        if (messages) {
          throw new SettingsFlowError("VALIDATION", null, messages);
        }
      }
      throw new SettingsFlowError(
        "SUBMIT_FAILED",
        null,
        "Settings flow returned 200 without a success state",
      );
    }

    const body = (await res.json().catch(() => null)) as KratosErrorEnvelope | null;
    if (res.status === 403 && body?.error?.id === "session_refresh_required") {
      throw new SettingsFlowError(
        "REAUTH_REQUIRED",
        body.redirect_browser_to ?? null,
        "Re-authentication required",
      );
    }
    if (res.status === 400 && body?.ui) {
      const messages = this.collectMessages(body.ui);
      throw new SettingsFlowError(
        "VALIDATION",
        null,
        messages || "Password did not pass validation",
      );
    }
    throw new SettingsFlowError(
      "SUBMIT_FAILED",
      null,
      body?.error?.message ?? `HTTP ${res.status}`,
    );
  }

  private extractCsrfToken(ui: KratosUi): string {
    for (const node of ui.nodes ?? []) {
      if (node.attributes?.name === "csrf_token") {
        return String(node.attributes.value ?? "");
      }
    }
    return "";
  }

  private collectMessages(ui: KratosUi): string {
    const out: string[] = [];
    for (const m of ui.messages ?? []) out.push(m.text);
    for (const n of ui.nodes ?? []) {
      for (const m of n.messages ?? []) out.push(m.text);
    }
    return out.join("; ");
  }

  // ------------------------------------------------------------------
  // Admin operations — still mocked, scheduled for PR-S3.
  // ------------------------------------------------------------------

  /**
   * Issue (or re-create) a Kratos identity + DevHub user pair. login_id is
   * accepted for UI symmetry but is currently the same as user_id on the
   * backend — Kratos identifies by traits.email + metadata_public.user_id.
   */
  async issueAccount(
    userId: string,
    loginId: string,
    forceReset: boolean,
    options?: { email?: string; displayName?: string; role?: string },
  ): Promise<{ tempPassword: string; identityId?: string }> {
    void forceReset;
    void loginId;
    const payload = await apiClient<{ data: { temp_password: string; identity_id?: string } }>(
      "POST",
      "/api/v1/accounts",
      {
        user_id: userId,
        email: options?.email ?? `${userId}@example.com`,
        display_name: options?.displayName ?? userId,
        role: options?.role,
      },
    );
    return { tempPassword: payload.data.temp_password, identityId: payload.data.identity_id };
  }

  async forceResetPassword(userId: string): Promise<{ tempPassword: string }> {
    const payload = await apiClient<{ data: { temp_password: string } }>(
      "PUT",
      `/api/v1/accounts/${encodeURIComponent(userId)}/password`,
      {},
    );
    return { tempPassword: payload.data.temp_password };
  }

  async disableAccount(userId: string, reason: string): Promise<void> {
    void reason; // 1차에서는 audit reason 미수신 (backend 가 actor 만 기록)
    await apiClient<{ status: string }>(
      "PATCH",
      `/api/v1/accounts/${encodeURIComponent(userId)}`,
      { status: "disabled" },
    );
  }

  async unlockAccount(userId: string): Promise<void> {
    await apiClient<{ status: string }>(
      "PATCH",
      `/api/v1/accounts/${encodeURIComponent(userId)}`,
      { status: "active" },
    );
  }

  async deleteAccount(userId: string): Promise<void> {
    await apiClient<{ status: string }>(
      "DELETE",
      `/api/v1/accounts/${encodeURIComponent(userId)}`,
    );
  }
}

export { ApiError };

export const accountService = new AccountService();
