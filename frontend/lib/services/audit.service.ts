"use client";

import { apiClient } from "./api-client";
import type { ApiResponse } from "./wire";
import type { AuditLogEntry, AuditLogFilters } from "./audit.types";

class AuditService {
  private static instance: AuditService;

  private constructor() {}

  public static getInstance(): AuditService {
    if (!AuditService.instance) {
      AuditService.instance = new AuditService();
    }
    return AuditService.instance;
  }

  /**
   * Fetches audit logs with optional filters
   * GET /api/v1/audit/logs
   */
  public async getLogs(filters: AuditLogFilters = {}): Promise<AuditLogEntry[]> {
    const params = new URLSearchParams();
    if (filters.action) params.set("action", filters.action);
    if (filters.target_type) params.set("target_type", filters.target_type);
    if (filters.actor_id) params.set("actor_id", filters.actor_id);
    if (filters.since) params.set("since", filters.since);
    if (filters.until) params.set("until", filters.until);
    if (filters.limit) params.set("limit", filters.limit.toString());
    if (filters.offset) params.set("offset", filters.offset.toString());

    const queryString = params.toString();
    const path = `/api/v1/audit/logs${queryString ? `?${queryString}` : ""}`;
    
    const result = await apiClient<ApiResponse<AuditLogEntry[]>>("GET", path);
    return result.data ?? [];
  }

  /**
   * Fetches a single audit log entry by ID
   * GET /api/v1/audit/logs/:id
   */
  public async getLogById(id: string): Promise<AuditLogEntry> {
    const result = await apiClient<ApiResponse<AuditLogEntry>>("GET", `/api/v1/audit/logs/${encodeURIComponent(id)}`);
    return result.data;
  }
}

export const auditService = AuditService.getInstance();
