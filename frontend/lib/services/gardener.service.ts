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

const API_BASE = "";

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
      const response = await fetch(`${API_BASE}/api/v1/gardener/suggestions`);
      if (!response.ok) throw new Error('Failed to fetch suggestions');
      
      const result = await response.json();
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
      const response = await fetch(`${API_BASE}/api/v1/gardener/suggestions/${suggestionId}/apply`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Devhub-Actor': 'yklee'
        }
      });

      if (!response.ok) throw new Error('Failed to apply suggestion');
      
      const result = await response.json();
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
