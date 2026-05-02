import { Metric, ServiceNode } from "./types";
import { getMockMetrics } from "../mockData";
import type { UserRole } from "../store";

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
    // Simulate API latency
    await new Promise(resolve => setTimeout(resolve, 300));
    return getMockMetrics(role);
  }

  async getNodes(): Promise<ServiceNode[]> {
    // In a real app, this would be a gRPC or REST call
    return [
      { id: '1', label: 'Go Core Service', status: 'stable', cpu: '12%', memory: '1.2GB' },
      { id: '2', label: 'Gitea Instance', status: 'stable', cpu: '8%', memory: '0.8GB' },
      { id: '3', label: 'Python AI Engine', status: 'warning', cpu: '45%', memory: '4.2GB' },
      { id: '4', label: 'PostgreSQL Cluster', status: 'stable', cpu: '5%', memory: '2.4GB' },
    ];
  }

  async controlService(serviceId: string, action: string): Promise<boolean> {
    console.log(`[InfraService] Executing ${action} on ${serviceId}`);
    await new Promise(resolve => setTimeout(resolve, 800));
    return true;
  }
}

export const infraService = InfraService.getInstance();
