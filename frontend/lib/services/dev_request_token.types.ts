/**
 * Dev Request Intake Token admin types.
 * Mirrors backend-core/internal/httpapi/dev_request_intake_tokens_admin.go (API-66..68).
 * Sprint claude/work_260515-p (DREQ-Admin-UI frontend). ADR-0014.
 *
 * Note: `hashed_token` 은 절대 응답에 포함되지 않는다. `plain_token` 은 발급
 * 응답 (POST /api/v1/dev-request-tokens) 에서만 1회 노출되며, 이후 list / revoke
 * 응답에는 없다.
 */

export interface DevRequestIntakeToken {
  token_id: string;
  client_label: string;
  source_system: string;
  allowed_ips: string[];
  created_at: string;
  created_by: string;
  last_used_at: string | null;
  revoked_at: string | null;
}

/** Response shape of `POST /api/v1/dev-request-tokens` — same as DevRequestIntakeToken
 *  but with `plain_token` (1회 노출). */
export interface IssuedDevRequestIntakeToken extends DevRequestIntakeToken {
  plain_token: string;
}

export interface IssueDevRequestIntakeTokenInput {
  client_label: string;
  source_system: string;
  allowed_ips: string[];
}
