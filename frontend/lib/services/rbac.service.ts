import type { ApiResponse, RbacPolicy } from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export class RbacService {
  private static instance: RbacService;

  private constructor() {}

  public static getInstance(): RbacService {
    if (!RbacService.instance) {
      RbacService.instance = new RbacService();
    }
    return RbacService.instance;
  }

  async getPolicy(): Promise<RbacPolicy> {
    const response = await fetch(`${API_BASE}/api/v1/rbac/policy`);
    if (!response.ok) {
      throw new Error("Failed to fetch RBAC policy");
    }
    const result = await response.json() as ApiResponse<RbacPolicy>;
    return result.data;
  }

  async replacePolicy(reason: string, matrix: RbacPolicy["matrix"], policyVersion?: string): Promise<RbacPolicy> {
    const response = await fetch(`${API_BASE}/api/v1/rbac/policy`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "X-Devhub-Actor": "yklee",
        "X-Devhub-Role": "system_admin",
      },
      body: JSON.stringify({
        policy_version: policyVersion,
        reason,
        matrix,
      }),
    });
    if (!response.ok) {
      throw new Error("Failed to replace RBAC policy");
    }
    const result = await response.json() as ApiResponse<RbacPolicy>;
    return result.data;
  }
}

export const rbacService = RbacService.getInstance();
