import { ApiMetric, ApiResponse, Metric, ServiceActionCommand, ServiceEdge, ServiceNode } from "@/shared/api/types";
import { type UserRole } from "@/lib/store";
import { formatBytes } from "@/shared/utils";
import { apiClient } from "@/shared/api/api-client";
import { API_BASE_URL } from "@/shared/config/endpoints";

class InfraService {
  private static instance: InfraService;
  private baseUrl = API_BASE_URL;

  private constructor() {}

  public static getInstance(): InfraService {
    if (!InfraService.instance) {
      InfraService.instance = new InfraService();
    }
    return InfraService.instance;
  }
  async getMetrics(role: UserRole): Promise<Metric[]> {
    const roleQuery = encodeURIComponent(role.toLowerCase().replace(' ', '_'));
    const result = await apiClient<ApiResponse<ApiMetric[]>>("GET", `${this.baseUrl}/api/v1/dashboard/metrics?role=${roleQuery}`);
    return result.data!.map((m) => ({
      label: m.label,
      value: m.value,
      trend: m.trend,
      color: m.trend_direction === 'up' ? 'text-emerald-500' : 'text-rose-500'
    }));
  }
  async getNodes(): Promise<ServiceNode[]> {
    try {
      const result = await apiClient<ApiResponse<ApiServiceNode[]>>("GET", `${this.baseUrl}/api/v1/infra/nodes`);
      return result.data!.map((n) => ({
        id: n.id,
        label: n.label,
        status: n.status,
        cpu: n.cpu_percent ? `${n.cpu_percent.toFixed(1)}%` : '0%',
        memory: n.memory_bytes ? formatBytes(n.memory_bytes) : '0 B',
        cpu_percent: n.cpu_percent,
        memory_bytes: n.memory_bytes,
        kind: n.kind,
        region: n.region,
        updated_at: n.updated_at
      }));
    } catch (error) {
      console.error('[InfraService] getNodes error:', error);
      return [
        { id: '1', label: 'Go Core Service', status: 'stable', cpu: '12%', memory: '1.2GB' },
        { id: '2', label: 'Gitea Instance', status: 'stable', cpu: '8%', memory: '0.8GB' },
        { id: '3', label: 'Python AI Engine', status: 'warning', cpu: '45%', memory: '4.2GB' },
        { id: '4', label: 'PostgreSQL Cluster', status: 'stable', cpu: '5%', memory: '2.4GB' },
      ];
    }
  }
  async getTopology(): Promise<{ nodes: ServiceNode[]; edges: ServiceEdge[] }> {
    try {
      const result = await apiClient<ApiResponse<{ nodes: ApiServiceNode[]; edges: ServiceEdge[] }>>("GET", `${this.baseUrl}/api/v1/infra/topology`);
      const nodes = result.data!.nodes.map((n) => ({
        id: n.id,
        label: n.label,
        status: n.status,
        cpu: n.cpu_percent ? `${n.cpu_percent.toFixed(1)}%` : '0%',
        memory: n.memory_bytes ? formatBytes(n.memory_bytes) : '0 B',
        cpu_percent: n.cpu_percent,
        memory_bytes: n.memory_bytes,
        kind: n.kind,
        region: n.region,
        updated_at: n.updated_at
      }));
      return { nodes, edges: result.data!.edges };
    } catch (error) {
      console.error('[InfraService] getTopology error:', error);
      return { nodes: [], edges: [] };
    }
  }
  async controlService(serviceId: string, action: string): Promise<boolean> {
    const actionType = action.toLowerCase().replace(/\s+/g, '_');
    const idempotencyKey = `service-${serviceId}-${actionType}-${Date.now()}`;
    const result = await apiClient<ApiResponse<ServiceActionCommand>>("POST", `${this.baseUrl}/api/v1/admin/service-actions`, {
        service_id: serviceId,
        action_type: actionType,
        dry_run: true,
        reason: `Manual ${action} from System Admin Dashboard`,
        idempotency_key: idempotencyKey,
      });
    return result.data!.command_status === 'pending';
  }

  /** API-76 + API-78 — infra topology v2 (HomeLab snapshot 기반).
   *  sprint claude/work_260518-n. backend response (`/api/v1/infra/topology/v2`):
   *  - data: { nodes: ApiInfraNodeV2[], edges: ApiServiceEdgeV2[], services: ApiInfraServiceV2[] }
   *  - meta: { snapshot_at: string, degraded_providers: string[] }
   */
  async getTopologyV2(): Promise<InfraTopologyV2Response> {
    const result = await apiClient<{
      data: { nodes: ApiInfraNodeV2[]; edges: ApiServiceEdgeV2[]; services: ApiInfraServiceV2[] };
      meta: InfraTopologyV2Meta;
    }>("GET", `${this.baseUrl}/api/v1/infra/topology/v2`);
    return {
      nodes: result.data.nodes,
      edges: result.data.edges,
      services: result.data.services,
      meta: result.meta,
    };
  }

  public formatBytes(bytes: number): string {
    return formatBytes(bytes);
  }
}

export const infraService = InfraService.getInstance();


interface ApiServiceNode {
  id: string;
  label: string;
  status: "stable" | "warning" | "down";
  cpu_percent?: number;
  memory_bytes?: number;
  kind?: string;
  region?: string;
  updated_at?: string;
}

// v2 (HomeLab snapshot 기반) — backend infra_integrations.go 의 schema mirror.
// sprint claude/work_260518-n. v1 (ApiServiceNode) 와 schema 다름:
// v1 = { id, label, status, cpu_percent, memory_bytes, ... }
// v2 = { node_id, hostname, ip_address, environment, status, metrics, observed_at }

export interface ApiInfraNodeV2 {
  node_id: string;
  hostname: string;
  ip_address?: string;
  environment?: string;
  status: string;
  metrics?: Record<string, unknown>;
  observed_at?: string;
}

export interface ApiInfraServiceV2 {
  service_id: string;
  node_id: string;
  name: string;
  version?: string;
  port?: number;
  health_status: string;
  observed_at?: string;
}

export interface ApiServiceEdgeV2 {
  id: string;
  source_id: string;
  target_id: string;
  label?: string;
  status?: string;
}

export interface InfraTopologyV2Meta {
  snapshot_at: string;
  degraded_providers: string[];
}

export interface InfraTopologyV2Response {
  nodes: ApiInfraNodeV2[];
  edges: ApiServiceEdgeV2[];
  services: ApiInfraServiceV2[];
  meta: InfraTopologyV2Meta;
}
