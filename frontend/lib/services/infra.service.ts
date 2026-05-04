import { Metric, ServiceNode } from "./types";
import { getMockMetrics } from "../mockData";
import type { UserRole } from "../store";
import { formatBytes } from "../utils";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export class InfraService {
  private static instance: InfraService;

  private constructor() {}

  public static getInstance(): InfraService {
    if (!InfraService.instance) {
      InfraService.instance = new InfraService();
    }
    return InfraService.instance;
  }

  async getMetrics(role: UserRole): Promise<Metric[]> {
    try {
      const roleQuery = encodeURIComponent(role.toLowerCase().replace(' ', '_'));
      const response = await fetch(`${API_BASE}/api/v1/dashboard/metrics?role=${roleQuery}`);
      if (!response.ok) throw new Error('Failed to fetch metrics');
      
      const result = await response.json();
      return result.data.map((m: any) => ({
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
      const response = await fetch(`${API_BASE}/api/v1/infra/nodes`);
      if (!response.ok) throw new Error('Failed to fetch nodes');
      
      const result = await response.json();
      return result.data.map((n: any) => ({
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

  async getTopology(): Promise<{ nodes: ServiceNode[]; edges: any[] }> {
    try {
      const response = await fetch(`${API_BASE}/api/v1/infra/topology`);
      if (!response.ok) throw new Error('Failed to fetch topology');
      
      const result = await response.json();
      const nodes = result.data.nodes.map((n: any) => ({
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
    console.log(`[InfraService] Executing ${action} on ${serviceId}`);
    // This will be implemented in Phase 4: Admin Actions
    await new Promise(resolve => setTimeout(resolve, 800));
    return true;
  }
}

export const infraService = InfraService.getInstance();
