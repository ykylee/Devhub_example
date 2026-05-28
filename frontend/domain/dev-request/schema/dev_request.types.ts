/**
 * Dev Request (DREQ) Domain Types
 * Based on backend-core/internal/domain/dev_request.go and ARCH-DREQ-05 / API-59..65.
 * Sprint claude/work_260515-j (DREQ-Frontend).
 */

export type DevRequestStatus =
  | "received"
  | "pending"
  | "in_review"
  | "registered"
  | "rejected"
  | "closed";

export type DevRequestTargetType = "application" | "project";

export interface DevRequest {
  id: string;
  title: string;
  details: string;
  requester: string;
  assignee_user_id: string;
  source_system: string;
  external_ref: string;
  status: DevRequestStatus;
  registered_target_type?: DevRequestTargetType | "";
  registered_target_id?: string;
  rejected_reason?: string;
  received_at: string;
  created_at: string;
  updated_at: string;
}

export interface DevRequestRegisterPayload {
  target_type: DevRequestTargetType;
  target_id: string;
}
