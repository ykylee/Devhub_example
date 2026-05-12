"use client";

import { apiClient } from "./api-client";
import type { ApiResponse } from "./wire";
import type { AuditLogEntry, AuditLogFilters, AuditLogListMeta } from "./audit.types";

export interface AuditLogListResult {
  entries: AuditLogEntry[];
  meta: AuditLogListMeta;
}

class AuditService {
  private static instance: AuditService;
  private constructor() {}
  public static getInstance(): AuditService {
    if (!AuditService.instance) AuditService.instance = new AuditService();
    return AuditService.instance;
  }

  /**
   * Fetches audit logs with optional filters.
   * GET /api/v1/audit-logs (system_admin only — backend enforces RBAC).
   *
   * Returns both the entries and the meta block ({limit, offset, count})
   * so pagination UI can render "showing N of M" without a second call.
   */
  public async getLogs(filters: AuditLogFilters = {}): Promise<AuditLogListResult> {
    const params = new URLSearchParams();
    if (filters.actor_login) params.set("actor_login", filters.actor_login);
    if (filters.action) params.set("action", filters.action);
    if (filters.target_type) params.set("target_type", filters.target_type);
    if (filters.target_id) params.set("target_id", filters.target_id);
    if (filters.command_id) params.set("command_id", filters.command_id);
    if (filters.limit !== undefined) params.set("limit", filters.limit.toString());
    if (filters.offset !== undefined) params.set("offset", filters.offset.toString());

    const queryString = params.toString();
    const path = `/api/v1/audit-logs${queryString ? `?${queryString}` : ""}`;
    const result = await apiClient<ApiResponse<AuditLogEntry[]>>("GET", path);
    return {
      entries: result.data ?? [],
      meta: (result.meta as AuditLogListMeta | undefined) ?? {},
    };
  }
}

export const auditService = AuditService.getInstance();
