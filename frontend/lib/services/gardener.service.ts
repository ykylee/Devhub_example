export interface CommandResponse {
  command_id: string;
  status: string;
}

export interface Suggestion {
  id: string;
  title: string;
  description: string;
  type: "optimization" | "security" | "scaling";
  impact: "low" | "medium" | "high";
  auto_fixable: boolean;
  created_at: string;
}

import { apiClient } from "./api-client";


export class GardenerService {
  private static instance: GardenerService;

  private constructor() {}

  public static getInstance(): GardenerService {
    if (!GardenerService.instance) {
      GardenerService.instance = new GardenerService();
    }
    return GardenerService.instance;
  }

  async getSuggestions(): Promise<Suggestion[]> {
    try {
      const result = await apiClient<{ data: Suggestion[] }>("GET", "/api/v1/gardener/suggestions");
      return result.data;
    } catch (error) {
      console.error('[GardenerService] getSuggestions error:', error);
      // Fallback to mock data for Phase 4 prototyping
      return [
        {
          id: "sug-001",
          title: "Scale up Go Core Service",
          description: "Traffic is increasing. Recommend adding 1 more instance for high availability.",
          type: "scaling",
          impact: "medium",
          auto_fixable: true,
          created_at: new Date().toISOString()
        },
        {
          id: "sug-002",
          title: "Idle Node Cleanup",
          description: "Python AI Engine (Node 3) has been idle for 2 hours. Consider shutting it down to save costs.",
          type: "optimization",
          impact: "low",
          auto_fixable: true,
          created_at: new Date().toISOString()
        }
      ];
    }
  }

  async applySuggestion(suggestionId: string): Promise<CommandResponse> {
    try {
      const result = await apiClient<{ data: { command_id: string; command_status: string } }>("POST", `/api/v1/gardener/suggestions/${suggestionId}/apply`);
      return {
        command_id: result.data.command_id,
        status: result.data.command_status
      };
    } catch (error) {
      console.error('[GardenerService] applySuggestion error:', error);
      throw error;
    }
  }
}

export const gardenerService = GardenerService.getInstance();
