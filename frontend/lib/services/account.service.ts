/**
 * Account Service
 *
 * Admin operations (PR-S3): issueAccount / forceResetPassword /
 * disableAccount / unlockAccount / deleteAccount call backend
 * /api/v1/accounts endpoints which proxy the IdP (Keycloak since
 * ADR-0019) Admin API. Backend ownership keeps the admin URL off the
 * browser and lands every action in DevHub audit_logs.
 *
 * Self-service password change is delegated to the Keycloak Account
 * Console (ADR-0019, sprint claude/work_260519-ad) — DevHub no longer
 * proxies that flow.
 */

import { apiClient, ApiError } from "@/lib/services/api-client";

export interface AccountInfo {
  id: number;
  user_id: string;
  login_id: string;
  status: 'active' | 'disabled' | 'locked' | 'password_reset_required';
  last_login_at?: string;
}

class AccountService {
  /**
   * Issue (or re-create) an IdP identity + DevHub user pair. login_id is
   * accepted for UI symmetry but is currently the same as user_id on the
   * backend — Keycloak identifies by username + attributes.devhub_user_id.
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
    void reason;
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
