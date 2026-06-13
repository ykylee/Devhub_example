import { apiClient } from "@/shared/api/api-client";
import type {
  IntegrationSyncJob,
  IntegrationSyncJobStatusSummaryResponse,
  ListIntegrationSyncJobsOptions,
  ListIntegrationSyncJobsResponse,
} from "@/domain/integration-registry/schema/integration.types";

/**
 * AdminX1Service — X-1 System Admin 운영 대시보드 (RM-M4-07) 의 admin
 * 운영 endpoint 3종 백엔드. system_admin 일임 (routePermissionTable 의
 * integration_sync_jobs resource gate). 3 method:
 *   - listSyncJobs (API-104) — sync job 큐/상태 + limit/offset
 *   - getSyncJob (API-105) — 단건 조회
 *   - getStatusSummary (API-106) — dashboard summary (4 status 별 count)
 */
class AdminX1Service {
  /** API-104 — `GET /api/v1/admin/integrations/sync-jobs?status=&limit=&offset=` */
  async listSyncJobs(opts: ListIntegrationSyncJobsOptions = {}): Promise<ListIntegrationSyncJobsResponse> {
    const params = new URLSearchParams();
    if (opts.status) params.set("status", opts.status);
    if (opts.limit) params.set("limit", String(opts.limit));
    if (opts.offset) params.set("offset", String(opts.offset));
    const qs = params.toString();
    const path = qs
      ? `/api/v1/admin/integrations/sync-jobs?${qs}`
      : "/api/v1/admin/integrations/sync-jobs";
    return await apiClient<ListIntegrationSyncJobsResponse>("GET", path);
  }

  /** API-105 — `GET /api/v1/admin/integrations/sync-jobs/:jobID` */
  async getSyncJob(jobID: string): Promise<IntegrationSyncJob> {
    return await apiClient<IntegrationSyncJob>(
      "GET",
      `/api/v1/admin/integrations/sync-jobs/${encodeURIComponent(jobID)}`,
    );
  }

  /** API-106 — `GET /api/v1/admin/integrations/summary` */
  async getStatusSummary(): Promise<IntegrationSyncJobStatusSummaryResponse> {
    return await apiClient<IntegrationSyncJobStatusSummaryResponse>(
      "GET",
      "/api/v1/admin/integrations/summary",
    );
  }
}

export const adminX1Service = new AdminX1Service();
