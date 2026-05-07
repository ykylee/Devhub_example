/**
 * Account Service
 * Provides methods for user account management, password resets, and admin controls.
 * Currently uses Mock implementations (Promises) pending Phase 13 IdP Backend integration.
 */

export interface AccountInfo {
  id: number;
  user_id: string;
  login_id: string;
  status: 'active' | 'disabled' | 'locked' | 'password_reset_required';
  last_login_at?: string;
}

class AccountService {
  /**
   * Mock API delay helper
   */
  private delay(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  /**
   * User self-service: Update own password
   */
  async updateMyPassword(currentPass: string, newPass: string): Promise<void> {
    await this.delay(800);
    // Mock validation
    if (currentPass !== "password") { // Mock current password is 'password'
      throw new Error("Invalid current password");
    }
    console.log("[AccountService] Password updated successfully");
  }

  /**
   * Admin: Issue a new account for an existing user
   */
  async issueAccount(userId: string, loginId: string, forceReset: boolean): Promise<{ tempPassword: string }> {
    await this.delay(600);
    console.log(`[AccountService] Issued account for ${userId} with login ${loginId}`);
    return { tempPassword: `Temp${Math.floor(Math.random() * 10000)}!` };
  }

  /**
   * Admin: Force reset password for a user
   */
  async forceResetPassword(userId: string): Promise<{ tempPassword: string }> {
    await this.delay(600);
    console.log(`[AccountService] Forced password reset for ${userId}`);
    return { tempPassword: `Reset${Math.floor(Math.random() * 10000)}!` };
  }

  /**
   * Admin: Revoke (Disable) an account
   */
  async disableAccount(userId: string, reason: string): Promise<void> {
    await this.delay(600);
    console.log(`[AccountService] Disabled account for ${userId}. Reason: ${reason}`);
  }

  /**
   * Admin: Unlock an account
   */
  async unlockAccount(userId: string): Promise<void> {
    await this.delay(600);
    console.log(`[AccountService] Unlocked account for ${userId}`);
  }
}

export const accountService = new AccountService();
