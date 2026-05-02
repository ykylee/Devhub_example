import { Risk } from "./types";

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
    await new Promise(resolve => setTimeout(resolve, 400));
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

  async applyMitigation(riskTitle: string, action: string): Promise<boolean> {
    console.log(`[RiskService] Applying ${action} to risk: ${riskTitle}`);
    await new Promise(resolve => setTimeout(resolve, 1000));
    return true;
  }
}

export const riskService = RiskService.getInstance();
