/**
 * Account Service
 *
 * Self-service (PR-L4, DEC-A=A): updateMyPassword calls the backend proxy
 * `POST /api/v1/account/password`, which verifies current_password, drives
 * Kratos api-mode settings flow server-side, and refreshes the cached
 * Kratos session. Browser-mode cookies are no longer required, so the api-
 * mode OIDC token alone authenticates the request.
 *
 * Admin operations (PR-S3): issueAccount / forceResetPassword /
 * disableAccount / unlockAccount call backend /api/v1/accounts endpoints
 * which proxy Kratos admin API. Backend ownership keeps the Kratos admin
 * URL off the browser and lands every action in DevHub audit_logs.
 */

import { apiClient, ApiError } from "@/lib/services/api-client";

export type SettingsFlowErrorCode =
  | "REAUTH_REQUIRED"
  | "UNAUTHENTICATED"
  | "VALIDATION"
  | "FLOW_INIT_FAILED"
  | "SUBMIT_FAILED"
  | "CURRENT_PASSWORD_INVALID";

// SettingsFlowError carries enough context for the UI to either show a
// validation message inline or send the user back through Kratos for a
// privileged-session re-auth.
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

export interface AccountInfo {
  id: number;
  user_id: string;
  login_id: string;
  status: 'active' | 'disabled' | 'locked' | 'password_reset_required';
  last_login_at?: string;
}

function payloadCode(payload: unknown): string {
  if (payload && typeof payload === "object" && !Array.isArray(payload)) {
    const code = (payload as { code?: unknown }).code;
    if (typeof code === "string") return code;
  }
  return "";
}

// mapApiErrorToSettingsFlowError translates the backend's `code` field
// (account_password.go.respondSettingsError) onto the SettingsFlowError
// vocabulary the /account page already consumes. Unknown codes fall through
// to SUBMIT_FAILED so the UI surfaces a generic error rather than swallow.
function mapApiErrorToSettingsFlowError(err: ApiError): SettingsFlowError {
  const code = payloadCode(err.payload);
  switch (code) {
    case "validation":
      return new SettingsFlowError("VALIDATION", null, err.message);
    case "reauth_required":
      return new SettingsFlowError("REAUTH_REQUIRED", null, err.message);
    case "current_password_invalid":
      return new SettingsFlowError("CURRENT_PASSWORD_INVALID", null, err.message);
    case "flow_expired":
      return new SettingsFlowError("FLOW_INIT_FAILED", null, err.message);
    default:
      return new SettingsFlowError("SUBMIT_FAILED", null, err.message);
  }
}

class AccountService {
  /**
   * User self-service: change own password via the backend proxy. The proxy
   * runs a fresh Kratos api-mode login with current_password, opens a
   * settings flow, and submits the new password — all server-side. The
   * browser only carries the OIDC access token via apiClient's
   * Authorization header.
   */
  async updateMyPassword(currentPass: string, newPass: string): Promise<void> {
    try {
      await apiClient<{ status: string; data: { user_id: string } }>(
        "POST",
        "/api/v1/account/password",
        { current_password: currentPass, new_password: newPass },
      );
    } catch (err) {
      if (err instanceof ApiError) {
        throw mapApiErrorToSettingsFlowError(err);
      }
      throw err;
    }
  }

  // ------------------------------------------------------------------
  // Admin operations (PR-S3) — unchanged.
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
