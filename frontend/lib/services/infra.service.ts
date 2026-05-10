import { ApiMetric, ApiResponse, Metric, ServiceActionCommand, ServiceEdge, ServiceNode } from "./types";
import { getMockMetrics } from "../mockData";
import { useStore, type UserRole } from "../store";
import { formatBytes } from "../utils";

class InfraService {
  private static instance: InfraService;
  private baseUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  private constructor() {}

  public static getInstance(): InfraService {
    if (!InfraService.instance) {
      InfraService.instance = new InfraService();
    }
    return InfraService.instance;
  }

  private getHeaders(): HeadersInit {
    const { actor, role } = useStore.getState();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    
    if (actor?.login) {
      headers["X-Devhub-Actor"] = actor.login;
    }
    if (role) {
      const roleMap: Record<string, string> = {
        "System Admin": "system_admin",
        "Manager": "manager",
        "Developer": "developer"
      };
      headers["X-Devhub-Role"] = roleMap[role] || role.toLowerCase();
    }
    
    return headers;
  }

  async getMetrics(role: UserRole): Promise<Metric[]> {
    try {
      const roleQuery = encodeURIComponent(role.toLowerCase().replace(' ', '_'));
      const response = await fetch(`${this.baseUrl}/api/v1/dashboard/metrics?role=${roleQuery}`, {
        headers: this.getHeaders()
      });
      if (!response.ok) throw new Error('Failed to fetch metrics');
      
      const result = await response.json() as ApiResponse<ApiMetric[]>;
      return result.data.map((m) => ({
        label: m.label,
        value: m.value,
        trend: m.trend,
        color: m.trend_direction === 'up' ? 'text-emerald-500' : 'text-rose-500'
      }));
    } catch (error) {
      console.error('[InfraService] getMetrics error:', error);
      return getMockMetrics(role);
    }
  }

  async getNodes(): Promise<ServiceNode[]> {
    try {
      const response = await fetch(`${this.baseUrl}/api/v1/infra/nodes`, {
        headers: this.getHeaders()
      });
      if (!response.ok) throw new Error('Failed to fetch nodes');
      
      const result = await response.json() as ApiResponse<ApiServiceNode[]>;
      return result.data.map((n) => ({
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
      const response = await fetch(`${this.baseUrl}/api/v1/infra/topology`, {
        headers: this.getHeaders()
      });
      if (!response.ok) throw new Error('Failed to fetch topology');
      
      const result = await response.json() as ApiResponse<{ nodes: ApiServiceNode[]; edges: ServiceEdge[] }>;
      const nodes = result.data.nodes.map((n) => ({
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
      return { nodes, edges: result.data.edges };
    } catch (error) {
      console.error('[InfraService] getTopology error:', error);
      return { nodes: [], edges: [] };
    }
  }

  async controlService(serviceId: string, action: string): Promise<boolean> {
    const actionType = action.toLowerCase().replace(/\s+/g, '_');
    const idempotencyKey = `service-${serviceId}-${actionType}-${Date.now()}`;
    const response = await fetch(`${this.baseUrl}/api/v1/admin/service-actions`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({
        service_id: serviceId,
        action_type: actionType,
        dry_run: true,
        reason: `Manual ${action} from System Admin Dashboard`,
        idempotency_key: idempotencyKey,
      }),
    });
    if (!response.ok) {
      throw new Error('Failed to create service action command');
    }
    const result = await response.json() as ApiResponse<ServiceActionCommand>;
    return result.data.command_status === 'pending';
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
