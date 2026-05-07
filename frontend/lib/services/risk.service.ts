import { ApiResponse, Risk, ServiceActionCommand } from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export class RiskService {
  private static instance: RiskService;

  private constructor() {}

  public static getInstance(): RiskService {
    if (!RiskService.instance) {
      RiskService.instance = new RiskService();
    }
    return RiskService.instance;
  }

  async getCriticalRisks(): Promise<Risk[]> {
    try {
      const response = await fetch(`${API_BASE}/api/v1/risks/critical`);
      if (!response.ok) throw new Error('Failed to fetch critical risks');
      
      const result = await response.json() as ApiResponse<ApiRisk[]>;
      return result.data.map((r) => ({
        id: r.id,
        title: r.title,
        reason: r.reason,
        impact: r.impact,
        status: r.status,
        owner: r.owner_login,
        owner_login: r.owner_login,
        suggested_actions: r.suggested_actions,
        created_at: r.created_at
      }));
    } catch (error) {
      console.error('[RiskService] getCriticalRisks error:', error);
      return [
        { 
          title: "Gitea Migration Blocked", 
          reason: "Access token expiration and scope mismatch detected in logs.",
          impact: "High", 
          status: "Action Required",
          owner: "Alex K."
        },
        { 
          title: "Frontend CI Pipeline Delay", 
          reason: "Average build time increased by 45% in last 24h.",
          impact: "Medium", 
          status: "Investigation",
          owner: "YK Lee"
        }
      ];
    }
  }

  async applyMitigation(riskId: string, action: string): Promise<{ command_id: string; status: string }> {
    try {
      const idempotencyKey = `mitigation-${riskId}-${action}-${Date.now()}`;
      const response = await fetch(`${API_BASE}/api/v1/risks/${riskId}/mitigations`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          // TODO: Replace hardcoded actor ID with actual authenticated user in Phase 4
          'X-Devhub-Actor': 'yklee' // Hardcoded for now as per API contract draft
        },
        body: JSON.stringify({
          action_type: action,
          reason: 'Manual mitigation from Manager Dashboard',
          idempotency_key: idempotencyKey
        })
      });

      if (!response.ok) throw new Error('Failed to apply mitigation');
      
      const result = await response.json() as ApiResponse<ServiceActionCommand>;
      return {
        command_id: result.data.command_id,
        status: result.data.command_status
      };
    } catch (error) {
      console.error('[RiskService] applyMitigation error:', error);
      throw error;
    }
  }
}

export const riskService = RiskService.getInstance();

interface ApiRisk {
  id: string;
  title: string;
  reason: string;
  impact: string;
  status: string;
  owner_login: string;
  suggested_actions: string[];
  created_at: string;
}
