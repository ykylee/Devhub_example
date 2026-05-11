/**
 * Account Service
 *
 * Self-service: updateMyPassword drives the Kratos public settings flow
 * (DEC-2=B, PR-L3). Browser cookie carries the Kratos session that Kratos
 * uses to identify the user; backend is not involved.
 *
 * Admin operations (issueAccount / forceResetPassword / disableAccount /
 * unlockAccount) remain mocked here — they will move to backend-proxied
 * Kratos admin API calls in PR-S3.
 */

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
  /** Mock API delay helper (still used by admin mocks). */
  private delay(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

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

    if (res.ok) return;

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

  async issueAccount(userId: string, loginId: string, forceReset: boolean): Promise<{ tempPassword: string }> {
    void forceReset;
    await this.delay(600);
    console.log(`[AccountService] (mock) Issued account for ${userId} with login ${loginId}`);
    return { tempPassword: `Temp${Math.floor(Math.random() * 10000)}!` };
  }

  async forceResetPassword(userId: string): Promise<{ tempPassword: string }> {
    await this.delay(600);
    console.log(`[AccountService] (mock) Forced password reset for ${userId}`);
    return { tempPassword: `Reset${Math.floor(Math.random() * 10000)}!` };
  }

  async disableAccount(userId: string, reason: string): Promise<void> {
    await this.delay(600);
    console.log(`[AccountService] (mock) Disabled account for ${userId}. Reason: ${reason}`);
  }

  async unlockAccount(userId: string): Promise<void> {
    await this.delay(600);
    console.log(`[AccountService] (mock) Unlocked account for ${userId}`);
  }
}

export const accountService = new AccountService();
