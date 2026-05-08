import { Role, RbacPolicyMeta } from "./rbac.types";

interface ListPoliciesEnvelope {
  status: string;
  data: Role[];
  meta: RbacPolicyMeta;
}

interface RoleEnvelope {
  status: string;
  data: Role;
}

interface SubjectRolesEnvelope {
  status: string;
  data: string[];
  meta: { subject_id: string; single_role_mode: boolean };
}

interface DeleteEnvelope {
  status: string;
  data: { role_id: string };
}

export interface ListPoliciesResult {
  roles: Role[];
  meta: RbacPolicyMeta;
}

class RbacService {
  private baseUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  // GET /api/v1/rbac/policies — section 12.2.
  async listPolicies(): Promise<ListPoliciesResult> {
    const response = await fetch(`${this.baseUrl}/api/v1/rbac/policies`);
    if (!response.ok) {
      throw new Error(`listPolicies failed: ${response.status}`);
    }
    const body: ListPoliciesEnvelope = await response.json();
    return { roles: body.data, meta: body.meta };
  }

  // POST /api/v1/rbac/policies — section 12.4. id must be `custom-{slug}`.
  async createPolicy(role: Pick<Role, "id" | "name" | "description" | "permissions">): Promise<Role> {
    const response = await fetch(`${this.baseUrl}/api/v1/rbac/policies`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(role),
    });
    if (!response.ok) {
      throw await rbacError(response, "createPolicy");
    }
    const body: RoleEnvelope = await response.json();
    return body.data;
  }

  // PUT /api/v1/rbac/policies — section 12.3. Bulk update permissions and/or
  // metadata. System roles only accept permission changes; metadata changes on
  // them are rejected with 422 system_role_immutable.
  async updatePolicies(roles: Array<Pick<Role, "id"> & Partial<Pick<Role, "name" | "description" | "permissions">>>): Promise<ListPoliciesResult> {
    const response = await fetch(`${this.baseUrl}/api/v1/rbac/policies`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ roles }),
    });
    if (!response.ok) {
      throw await rbacError(response, "updatePolicies");
    }
    const body: ListPoliciesEnvelope = await response.json();
    return { roles: body.data, meta: body.meta };
  }

  // DELETE /api/v1/rbac/policies/:role_id — section 12.5. System roles return
  // 422 system_role_not_deletable; in-use custom roles return 422 role_in_use.
  async deletePolicy(roleId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/api/v1/rbac/policies/${encodeURIComponent(roleId)}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw await rbacError(response, "deletePolicy");
    }
    const body: DeleteEnvelope = await response.json();
    if (body.status !== "deleted") {
      throw new Error(`deletePolicy unexpected status: ${body.status}`);
    }
  }

  // GET /api/v1/rbac/subjects/:subject_id/roles — section 12.6. Single-role
  // mode: response array length is 0 (only when subject not found, which
  // surfaces as 404) or 1.
  async getSubjectRoles(subjectId: string): Promise<string[]> {
    const response = await fetch(`${this.baseUrl}/api/v1/rbac/subjects/${encodeURIComponent(subjectId)}/roles`);
    if (!response.ok) {
      throw await rbacError(response, "getSubjectRoles");
    }
    const body: SubjectRolesEnvelope = await response.json();
    return body.data;
  }

  // PUT /api/v1/rbac/subjects/:subject_id/roles — section 12.7. Single-role
  // mode requires exactly one entry; the helper enforces it client-side too.
  async setSubjectRole(subjectId: string, roleId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/api/v1/rbac/subjects/${encodeURIComponent(subjectId)}/roles`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ roles: [roleId] }),
    });
    if (!response.ok) {
      throw await rbacError(response, "setSubjectRole");
    }
  }
}

// RbacError preserves the contract section 12 error code (e.g. role_in_use,
// audit_invariant_violation) so the UI can branch on a stable identifier.
export class RbacError extends Error {
  status: number;
  code: string;

  constructor(message: string, status: number, code: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

async function rbacError(response: Response, op: string): Promise<RbacError> {
  let code = "unknown";
  let message = `${op} failed: ${response.status}`;
  try {
    const body = await response.json();
    if (body?.code) code = body.code;
    if (body?.error) message = body.error;
  } catch {
    // body was not JSON; keep the default message.
  }
  return new RbacError(message, response.status, code);
}

export const rbacService = new RbacService();
