import { apiClient } from "./api-client";
import { DevRequest, DevRequestRegisterPayload, DevRequestStatus } from "./dev_request.types";

type DevRequestQuery = {
  status?: DevRequestStatus | DevRequestStatus[];
  assignee_user_id?: string;
  source_system?: string;
  limit?: number;
  offset?: number;
};

function withQuery(path: string, params?: Record<string, string | number | undefined>): string {
  if (!params) return path;
  const qs = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === "") continue;
    qs.set(k, String(v));
  }
  const raw = qs.toString();
  return raw ? `${path}?${raw}` : path;
}

function joinStatus(status?: DevRequestStatus | DevRequestStatus[]): string | undefined {
  if (!status) return undefined;
  return Array.isArray(status) ? status.join(",") : status;
}

class DevRequestService {
  async list(params?: DevRequestQuery): Promise<{ data: DevRequest[]; total: number }> {
    const path = withQuery("/api/v1/dev-requests", {
      status: joinStatus(params?.status),
      assignee_user_id: params?.assignee_user_id,
      source_system: params?.source_system,
      limit: params?.limit,
      offset: params?.offset,
    });
    const resp = await apiClient<{ data: DevRequest[]; meta?: { total: number } }>("GET", path);
    return { data: resp.data, total: resp.meta?.total ?? resp.data.length };
  }

  async get(id: string): Promise<DevRequest> {
    const resp = await apiClient<{ data: DevRequest }>("GET", `/api/v1/dev-requests/${id}`);
    return resp.data;
  }

  async register(id: string, payload: DevRequestRegisterPayload): Promise<DevRequest> {
    const resp = await apiClient<{ data: { dev_request: DevRequest } }>(
      "POST",
      `/api/v1/dev-requests/${id}/register`,
      payload,
    );
    return resp.data.dev_request;
  }

  async reject(id: string, rejectedReason: string): Promise<DevRequest> {
    const resp = await apiClient<{ data: DevRequest }>(
      "POST",
      `/api/v1/dev-requests/${id}/reject`,
      { rejected_reason: rejectedReason },
    );
    return resp.data;
  }

  async reassign(id: string, assigneeUserID: string): Promise<DevRequest> {
    const resp = await apiClient<{ data: DevRequest }>(
      "PATCH",
      `/api/v1/dev-requests/${id}`,
      { assignee_user_id: assigneeUserID },
    );
    return resp.data;
  }

  async close(id: string): Promise<DevRequest> {
    const resp = await apiClient<{ data: DevRequest }>("DELETE", `/api/v1/dev-requests/${id}`);
    return resp.data;
  }

  // 담당자 dashboard 위젯이 사용. 본인 assignee + pending/in_review 만.
  async getMyPending(): Promise<{ data: DevRequest[]; total: number }> {
    return this.list({ status: ["pending", "in_review"] });
  }
}

export const devRequestService = new DevRequestService();
