/**
 * Wire types — API/WebSocket payload shapes exactly as backend emits them.
 *
 * Convention:
 *   - keep field names in snake_case to match the JSON wire format
 *   - never reference UI display strings (e.g. "Developer") here
 *   - service modules (lib/services/*.service.ts) are responsible for
 *     converting wire shapes into UI shapes (lib/services/types.ts)
 *   - UI components must NOT import from this file directly; importing UI
 *     types from types.ts keeps the wire/UI boundary in services
 *
 * Source-of-truth (DEC-1=A, work_26_05_11-b sprint): docs/backend_api_contract.md.
 */

// Standard success envelope — docs/backend_api_contract.md §1.
//   - 성공: { status: "ok"|"created"|..., data: T, meta?: ... }
//   - 실패: { status: "rejected"|"unavailable"|..., error: string, code?: string }
//     (apiClient throws ApiError on non-2xx, so call sites that receive an
//     ApiResponse can assume the success branch — `data` is non-nullable
//     here. For endpoints that return no payload (DELETE), parameterise
//     with `unknown` and ignore the result.)
export interface ApiResponse<T> {
  status: string;
  data: T;
  meta?: Record<string, unknown>;
}

// Failure shape returned in the rejected envelope. Surfaced via ApiError;
// services rarely need this directly.
export interface ApiErrorResponse {
  status: string;
  error: string;
  code?: string;
}

// Role wire format — docs/backend_api_contract.md §2.
export type ApiUserRole = "developer" | "manager" | "system_admin";

// Dashboard metric wire shape — GET /api/v1/dashboard/metrics?role=...
export interface ApiMetric {
  label: string;
  value: string;
  trend: string;
  trend_direction: "up" | "down" | "flat";
  numeric_value?: number;
  unit?: string;
}

// WebSocket event envelope — docs/backend_api_contract.md §8 (5 keys).
//   - schema_version: integer-as-string per DEC-3=A (work_26_05_11-b)
//   - event_id: prefixed UUID (Hub.prefixedEventID())
//   - occurred_at: ISO-8601 UTC
export interface WSEvent<T = unknown> {
  schema_version: string;
  type: string;
  event_id: string;
  occurred_at: string;
  data: T;
}

export type WSEventHandler<T = unknown> = (event: WSEvent<T>) => void;
