import { Role } from "./rbac.types";

export interface RbacPolicyResponse {
  status: string;
  data: Role[];
}

class RbacService {
  private baseUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  /**
   * Fetch current default policy (Snapshot API - legacy/readonly)
   */
  async getPolicy(): Promise<any> {
    const response = await fetch(`${this.baseUrl}/api/v1/rbac/policy`);
    if (!response.ok) throw new Error("Failed to fetch RBAC policy");
    const body = await response.json();
    return body.data;
  }

  /**
   * Fetch all roles and their permission matrices (Management API)
   */
  async getPolicies(): Promise<Role[]> {
    try {
      const response = await fetch(`${this.baseUrl}/api/v1/rbac/policies`);
      if (!response.ok) throw new Error("Failed to fetch RBAC policies");
      const body: RbacPolicyResponse = await response.json();
      return body.data;
    } catch (error) {
      console.error("[RbacService] getPolicies failed:", error);
      throw error;
    }
  }

  /**
   * Update policies for specific roles
   */
  async updatePolicies(roles: Partial<Role>[]): Promise<void> {
    const response = await fetch(`${this.baseUrl}/api/v1/rbac/policies`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ roles }),
    });
    if (!response.ok) throw new Error("Failed to update RBAC policies");
  }

  /**
   * Get roles assigned to a specific subject (user)
   */
  async getSubjectRoles(subjectId: string): Promise<string[]> {
    const response = await fetch(`${this.baseUrl}/api/v1/rbac/subjects/${subjectId}/roles`);
    if (!response.ok) throw new Error(`Failed to fetch roles for subject ${subjectId}`);
    const body = await response.json();
    return body.data;
  }

  /**
   * Update roles for a specific subject
   */
  async updateSubjectRoles(subjectId: string, roleIds: string[]): Promise<void> {
    const response = await fetch(`${this.baseUrl}/api/v1/rbac/subjects/${subjectId}/roles`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ roles: roleIds }),
    });
    if (!response.ok) throw new Error(`Failed to update roles for subject ${subjectId}`);
  }
}

export const rbacService = new RbacService();

