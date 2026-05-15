import { Role, RbacPolicyMeta } from "./rbac.types";
import { apiClient, ApiError } from "./api-client";
import { API_BASE_URL } from "../config/endpoints";

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
  // Keep RBAC calls same-origin by default so browser clients use Next.js
  // rewrite/proxy instead of resolving localhost in user environments.
  private baseUrl = API_BASE_URL;



  // GET /api/v1/rbac/policies — section 12.2.
  async listPolicies(): Promise<ListPoliciesResult> {
    try {
      const body = await apiClient<ListPoliciesEnvelope>("GET", `${this.baseUrl}/api/v1/rbac/policies`);
      return { roles: body.data, meta: body.meta };
    } catch (err) {
      throw await rbacErrorFromApi(err, "listPolicies");
    }
  }

  // POST /api/v1/rbac/policies — section 12.4. id must be `custom-{slug}`.
  async createPolicy(role: Pick<Role, "id" | "name" | "description" | "permissions">): Promise<Role> {
    try {
      const body = await apiClient<RoleEnvelope>("POST", `${this.baseUrl}/api/v1/rbac/policies`, role);
      return body.data;
    } catch (err) {
      throw await rbacErrorFromApi(err, "createPolicy");
    }
  }

  // PUT /api/v1/rbac/policies — section 12.3. Bulk update permissions and/or
  // metadata. System roles only accept permission changes; metadata changes on
  // them are rejected with 422 system_role_immutable.
  async updatePolicies(roles: Array<Pick<Role, "id"> & Partial<Pick<Role, "name" | "description" | "permissions">>>): Promise<ListPoliciesResult> {
    try {
      const body = await apiClient<ListPoliciesEnvelope>("PUT", `${this.baseUrl}/api/v1/rbac/policies`, { roles });
      return { roles: body.data, meta: body.meta };
    } catch (err) {
      throw await rbacErrorFromApi(err, "updatePolicies");
    }
  }

  // DELETE /api/v1/rbac/policies/:role_id — section 12.5. System roles return
  // 422 system_role_not_deletable; in-use custom roles return 422 role_in_use.
  async deletePolicy(roleId: string): Promise<void> {
    try {
      const body = await apiClient<DeleteEnvelope>("DELETE", `${this.baseUrl}/api/v1/rbac/policies/${encodeURIComponent(roleId)}`);
      if (body.status !== "deleted") {
        throw new Error(`deletePolicy unexpected status: ${body.status}`);
      }
    } catch (err) {
      throw await rbacErrorFromApi(err, "deletePolicy");
    }
  }

  // GET /api/v1/rbac/subjects/:subject_id/roles — section 12.6. Single-role
  // mode: response array length is 0 (only when subject not found, which
  // surfaces as 404) or 1.
  async getSubjectRoles(subjectId: string): Promise<string[]> {
    try {
      const body = await apiClient<SubjectRolesEnvelope>("GET", `${this.baseUrl}/api/v1/rbac/subjects/${encodeURIComponent(subjectId)}/roles`);
      return body.data;
    } catch (err) {
      throw await rbacErrorFromApi(err, "getSubjectRoles");
    }
  }

  // PUT /api/v1/rbac/subjects/:subject_id/roles — section 12.7. Single-role
  // mode requires exactly one entry; the helper enforces it client-side too.
  async setSubjectRole(subjectId: string, roleId: string): Promise<void> {
    try {
      await apiClient("PUT", `${this.baseUrl}/api/v1/rbac/subjects/${encodeURIComponent(subjectId)}/roles`, { roles: [roleId] });
    } catch (err) {
      throw await rbacErrorFromApi(err, "setSubjectRole");
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

async function rbacErrorFromApi(err: unknown, op: string): Promise<RbacError> {
  if (err instanceof ApiError) {
    let code = "unknown";
    if (err.payload && typeof err.payload === 'object' && 'code' in (err.payload as Record<string, unknown>)) {
      code = (err.payload as Record<string, string>).code;
    }
    return new RbacError(err.message, err.status, code);
  }
  const message = err instanceof Error ? err.message : 'Unknown error';
  return new RbacError(`${op} failed: ${message}`, 500, "unknown");
}

export const rbacService = new RbacService();
