export interface OrgMember {
  id: string;
  name: string;
  email: string;
  primary_dept_id: string;
  current_dept_id: string;
  is_seconded: boolean;
  role: "Developer" | "Manager" | "System Admin";
  status: "active" | "pending" | "deactivated";
  appointments: {
    dept_id: string;
    role: 'leader' | 'member';
  }[];
  joined_at: string;
}

export interface Team {
  id: string;
  name: string;
  description: string;
  member_count: number;
  project_count: number;
  lead_id: string;
}

export interface OrgNode {
  id: string;
  type?: 'division' | 'team' | 'group' | 'part' | 'company' | 'input';
  data: {
    label: string;
    type: string;
    leader_id?: string;
    direct_count?: number;
    total_count?: number;
  };
  position: { x: number; y: number };
}

export interface OrgEdge {
  id: string;
  source: string;
  target: string;
  animated?: boolean;
}

type JsonObject = Record<string, unknown>;

interface ApiResponse<T> {
  data?: T;
  error?: string;
}

interface BackendAppointment {
  unit_id: string;
  appointment_role: 'leader' | 'member';
}

interface BackendUser {
  user_id: string;
  display_name: string;
  email: string;
  role: string;
  status: OrgMember["status"];
  primary_unit_id?: string;
  current_unit_id?: string;
  is_seconded?: boolean;
  appointments?: BackendAppointment[];
  joined_at?: string;
}

interface BackendUnit {
  unit_id: string;
  parent_unit_id?: string;
  unit_type: OrgUnit["unit_type"];
  label: string;
  leader_user_id?: string;
  position_x?: number;
  position_y?: number;
  direct_count?: number;
  total_count?: number;
}

interface BackendEdge {
  source_unit_id: string;
  target_unit_id: string;
}

const DEPT_PRIORITY = {
  'division': 4,
  'team': 3,
  'group': 2,
  'part': 1,
  'company': 5
};

const ROLE_BACKEND_TO_UI: Record<string, OrgMember["role"]> = {
  developer: "Developer",
  manager: "Manager",
  system_admin: "System Admin",
};

const ROLE_UI_TO_BACKEND: Record<OrgMember["role"], string> = {
  Developer: "developer",
  Manager: "manager",
  "System Admin": "system_admin",
};

export interface CreateUserPayload {
  user_id: string;
  email: string;
  display_name: string;
  role: OrgMember["role"];
  status: OrgMember["status"];
  primary_dept_id?: string;
  current_dept_id?: string;
  is_seconded?: boolean;
  joined_at?: string;
}

export interface UpdateUserPayload {
  email?: string;
  display_name?: string;
  role?: OrgMember["role"];
  status?: OrgMember["status"];
  primary_dept_id?: string;
  current_dept_id?: string;
  is_seconded?: boolean;
  joined_at?: string;
}

export interface OrgUnit {
  unit_id: string;
  parent_unit_id: string;
  unit_type: "company" | "division" | "team" | "group" | "part";
  label: string;
  leader_user_id: string;
  position_x: number;
  position_y: number;
  direct_count?: number;
  total_count?: number;
}

export interface CreateUnitPayload {
  unit_id: string;
  parent_unit_id?: string;
  unit_type: OrgUnit["unit_type"];
  label: string;
  leader_user_id?: string;
  position_x?: number;
  position_y?: number;
}

export interface UpdateUnitPayload {
  parent_unit_id?: string;
  unit_type?: OrgUnit["unit_type"];
  label?: string;
  leader_user_id?: string;
  position_x?: number;
  position_y?: number;
}

export class IdentityServiceError extends Error {
  constructor(public status: number, public payload: unknown, message: string) {
    super(message);
    this.name = "IdentityServiceError";
  }
}

function isJsonObject(value: unknown): value is JsonObject {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

async function jsonRequest<T>(method: string, path: string, body?: unknown): Promise<T> {
  const response = await fetch(path, {
    method,
    headers: body !== undefined ? { "Content-Type": "application/json" } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  let parsed: unknown = null;
  const text = await response.text();
  if (text.length > 0) {
    try {
      parsed = JSON.parse(text);
    } catch {
      parsed = { raw: text };
    }
  }
  if (!response.ok) {
    const errMessage = isJsonObject(parsed) && typeof parsed.error === "string" ? parsed.error : `HTTP ${response.status}`;
    throw new IdentityServiceError(response.status, parsed, errMessage);
  }
  return parsed as T;
}

function mapBackendUnit(u: BackendUnit): OrgUnit {
  return {
    unit_id: u.unit_id,
    parent_unit_id: u.parent_unit_id ?? "",
    unit_type: u.unit_type,
    label: u.label,
    leader_user_id: u.leader_user_id ?? "",
    position_x: u.position_x ?? 0,
    position_y: u.position_y ?? 0,
    direct_count: u.direct_count,
    total_count: u.total_count,
  };
}

function mapBackendUser(u: BackendUser): OrgMember {
  return {
    id: u.user_id,
    name: u.display_name,
    email: u.email,
    role: ROLE_BACKEND_TO_UI[u.role] ?? "Developer",
    status: u.status,
    primary_dept_id: u.primary_unit_id ?? "",
    current_dept_id: u.current_unit_id ?? "",
    is_seconded: !!u.is_seconded,
    appointments: (u.appointments ?? []).map((a) => ({
      dept_id: a.unit_id,
      role: a.appointment_role,
    })),
    joined_at: typeof u.joined_at === "string" ? u.joined_at.slice(0, 10) : (u.joined_at ?? ""),
  };
}

export class IdentityService {
  private static instance: IdentityService;

  private constructor() {}

  public static getInstance(): IdentityService {
    if (!IdentityService.instance) {
      IdentityService.instance = new IdentityService();
    }
    return IdentityService.instance;
  }

  async getUsers(): Promise<OrgMember[]> {
    try {
      const response = await fetch(`/api/v1/users`);
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const result = await response.json() as ApiResponse<BackendUser[]>;
      return (result.data ?? []).map(mapBackendUser);
    } catch (error) {
      console.error('[IdentityService] getUsers error, falling back to mock:', error);
      return this.mockUsers();
    }
  }

  async getTeams(): Promise<Team[]> {
    return [
      {
        id: "team-infra",
        name: "Infrastructure",
        description: "Cloud resources and Kubernetes management",
        member_count: 5,
        project_count: 3,
        lead_id: "u1"
      },
      {
        id: "team-frontend",
        name: "Frontend",
        description: "DevHub Web interface and Mobile apps",
        member_count: 8,
        project_count: 2,
        lead_id: "u3"
      }
    ];
  }

  async getOrgHierarchy(): Promise<{ nodes: OrgNode[]; edges: OrgEdge[] }> {
    try {
      const response = await fetch(`/api/v1/organization/hierarchy`);
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const result = await response.json() as ApiResponse<{ units?: BackendUnit[]; edges?: BackendEdge[] }>;
      const units = result.data?.units ?? [];
      const edges = result.data?.edges ?? [];
      const nodes = units.map((u) => ({
        id: u.unit_id,
        ...(u.unit_id === 'org-root' ? { type: 'input' as const } : {}),
        data: {
          label: u.label,
          type: u.unit_type,
          leader_id: u.leader_user_id || undefined,
          direct_count: u.direct_count,
          total_count: u.total_count,
        },
        position: { x: u.position_x ?? 0, y: u.position_y ?? 0 },
      }));
      const mappedEdges = edges.map((e) => ({
        id: `e-${e.source_unit_id}-${e.target_unit_id}`,
        source: e.source_unit_id,
        target: e.target_unit_id,
        animated: e.source_unit_id === 'org-root',
      }));
      return { nodes, edges: mappedEdges };
    } catch (error) {
      console.error('[IdentityService] getOrgHierarchy error, falling back to mock:', error);
      return this.mockHierarchy();
    }
  }

  async getUnitMembers(unitId: string): Promise<OrgMember[]> {
    try {
      const response = await fetch(`/api/v1/organization/units/${encodeURIComponent(unitId)}/members`);
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const result = await response.json() as ApiResponse<BackendUser[]>;
      return (result.data ?? []).map(mapBackendUser);
    } catch (error) {
      console.error('[IdentityService] getUnitMembers error:', error);
      return [];
    }
  }

  async createUser(payload: CreateUserPayload): Promise<OrgMember> {
    const body = {
      user_id: payload.user_id,
      email: payload.email,
      display_name: payload.display_name,
      role: ROLE_UI_TO_BACKEND[payload.role],
      status: payload.status,
      primary_unit_id: payload.primary_dept_id ?? "",
      current_unit_id: payload.current_dept_id ?? "",
      is_seconded: !!payload.is_seconded,
      joined_at: payload.joined_at ?? "",
    };
    const result = await jsonRequest<ApiResponse<BackendUser>>("POST", `/api/v1/users`, body);
    if (!result.data) throw new IdentityServiceError(500, result, "missing user payload");
    return mapBackendUser(result.data);
  }

  async updateUser(userId: string, payload: UpdateUserPayload): Promise<OrgMember> {
    const body: Record<string, unknown> = {};
    if (payload.email !== undefined) body.email = payload.email;
    if (payload.display_name !== undefined) body.display_name = payload.display_name;
    if (payload.role !== undefined) body.role = ROLE_UI_TO_BACKEND[payload.role];
    if (payload.status !== undefined) body.status = payload.status;
    if (payload.primary_dept_id !== undefined) body.primary_unit_id = payload.primary_dept_id;
    if (payload.current_dept_id !== undefined) body.current_unit_id = payload.current_dept_id;
    if (payload.is_seconded !== undefined) body.is_seconded = payload.is_seconded;
    if (payload.joined_at !== undefined) body.joined_at = payload.joined_at;
    const result = await jsonRequest<ApiResponse<BackendUser>>(
      "PATCH",
      `/api/v1/users/${encodeURIComponent(userId)}`,
      body,
    );
    if (!result.data) throw new IdentityServiceError(500, result, "missing user payload");
    return mapBackendUser(result.data);
  }

  async deleteUser(userId: string): Promise<void> {
    await jsonRequest<ApiResponse<unknown>>("DELETE", `/api/v1/users/${encodeURIComponent(userId)}`);
  }

  async getUnit(unitId: string): Promise<OrgUnit> {
    const result = await jsonRequest<ApiResponse<BackendUnit>>("GET", `/api/v1/organization/units/${encodeURIComponent(unitId)}`);
    if (!result.data) throw new IdentityServiceError(500, result, "missing unit payload");
    return mapBackendUnit(result.data);
  }

  async createUnit(payload: CreateUnitPayload): Promise<OrgUnit> {
    const body = {
      unit_id: payload.unit_id,
      parent_unit_id: payload.parent_unit_id ?? "",
      unit_type: payload.unit_type,
      label: payload.label,
      leader_user_id: payload.leader_user_id ?? "",
      position_x: payload.position_x ?? 0,
      position_y: payload.position_y ?? 0,
    };
    const result = await jsonRequest<ApiResponse<BackendUnit>>("POST", `/api/v1/organization/units`, body);
    if (!result.data) throw new IdentityServiceError(500, result, "missing unit payload");
    return mapBackendUnit(result.data);
  }

  async updateUnit(unitId: string, payload: UpdateUnitPayload): Promise<OrgUnit> {
    const body: Record<string, unknown> = {};
    if (payload.parent_unit_id !== undefined) body.parent_unit_id = payload.parent_unit_id;
    if (payload.unit_type !== undefined) body.unit_type = payload.unit_type;
    if (payload.label !== undefined) body.label = payload.label;
    if (payload.leader_user_id !== undefined) body.leader_user_id = payload.leader_user_id;
    if (payload.position_x !== undefined) body.position_x = payload.position_x;
    if (payload.position_y !== undefined) body.position_y = payload.position_y;
    const result = await jsonRequest<ApiResponse<BackendUnit>>(
      "PATCH",
      `/api/v1/organization/units/${encodeURIComponent(unitId)}`,
      body,
    );
    if (!result.data) throw new IdentityServiceError(500, result, "missing unit payload");
    return mapBackendUnit(result.data);
  }

  async deleteUnit(unitId: string): Promise<void> {
    await jsonRequest<ApiResponse<unknown>>("DELETE", `/api/v1/organization/units/${encodeURIComponent(unitId)}`);
  }

  async replaceUnitMembers(unitId: string, userIds: string[]): Promise<OrgMember[]> {
    const result = await jsonRequest<ApiResponse<BackendUser[]>>(
      "PUT",
      `/api/v1/organization/units/${encodeURIComponent(unitId)}/members`,
      { user_ids: userIds },
    );
    return (result.data ?? []).map(mapBackendUser);
  }

  calculatePrimaryDept(member: OrgMember, nodes: OrgNode[]): string {
    if (member.appointments.length <= 1) {
      return member.current_dept_id;
    }

    // 1. Get leader appointments
    const leadDepts = member.appointments
      .filter(a => a.role === 'leader')
      .map(a => nodes.find(n => n.id === a.dept_id))
      .filter(n => !!n) as OrgNode[];

    if (leadDepts.length === 0) return member.current_dept_id;

    // 2. Sort by priority (Division > Team > Group > Part)
    leadDepts.sort((a, b) => {
      const prioA = DEPT_PRIORITY[a.data.type as keyof typeof DEPT_PRIORITY] || 0;
      const prioB = DEPT_PRIORITY[b.data.type as keyof typeof DEPT_PRIORITY] || 0;
      if (prioA !== prioB) return prioB - prioA;

      // 3. If same priority, count children (simulated for now)
      return 0;
    });

    return leadDepts[0].id;
  }

  private mockUsers(): OrgMember[] {
    return [
      {
        id: "u1",
        name: "YK Lee",
        email: "yklee@example.com",
        role: "System Admin",
        status: "active",
        primary_dept_id: "dept-eng",
        current_dept_id: "dept-eng",
        is_seconded: false,
        appointments: [
          { dept_id: "org-root", role: "leader" },
          { dept_id: "dept-eng", role: "leader" }
        ],
        joined_at: "2026-01-15"
      },
      {
        id: "u2",
        name: "Alex Kim",
        email: "alex@example.com",
        role: "Manager",
        status: "active",
        primary_dept_id: "dept-prod",
        current_dept_id: "team-ux",
        is_seconded: true,
        appointments: [
          { dept_id: "dept-prod", role: "leader" }
        ],
        joined_at: "2026-02-01"
      },
      {
        id: "u3",
        name: "Sam Jones",
        email: "sam@example.com",
        role: "Developer",
        status: "active",
        primary_dept_id: "team-infra",
        current_dept_id: "team-infra",
        is_seconded: false,
        appointments: [
          { dept_id: "team-infra", role: "member" }
        ],
        joined_at: "2026-05-01"
      }
    ];
  }

  private mockHierarchy() {
    return {
      nodes: [
        { id: 'org-root', type: 'input' as const, data: { label: 'DevHub Global', type: 'division', leader_id: 'u1', direct_count: 5, total_count: 150 }, position: { x: 400, y: 0 } },
        { id: 'dept-eng', data: { label: 'Engineering', type: 'division', leader_id: 'u1', direct_count: 10, total_count: 85 }, position: { x: 200, y: 150 } },
        { id: 'dept-prod', data: { label: 'Product', type: 'division', leader_id: 'u2', direct_count: 8, total_count: 65 }, position: { x: 600, y: 150 } },
        { id: 'team-infra', data: { label: 'Infrastructure', type: 'team', leader_id: 'u1', direct_count: 12, total_count: 24 }, position: { x: 50, y: 300 } },
        { id: 'team-frontend', data: { label: 'Frontend', type: 'team', leader_id: 'u3', direct_count: 15, total_count: 15 }, position: { x: 350, y: 300 } },
        { id: 'team-ux', data: { label: 'UX Strategy', type: 'team', leader_id: 'u2', direct_count: 6, total_count: 6 }, position: { x: 600, y: 300 } },
        { id: 'part-security', data: { label: 'Security Part', type: 'part', direct_count: 4, total_count: 4 }, position: { x: 50, y: 450 } },
      ],
      edges: [
        { id: 'e-root-eng', source: 'org-root', target: 'dept-eng', animated: true },
        { id: 'e-root-prod', source: 'org-root', target: 'dept-prod', animated: true },
        { id: 'e-eng-infra', source: 'dept-eng', target: 'team-infra' },
        { id: 'e-eng-front', source: 'dept-eng', target: 'team-frontend' },
        { id: 'e-prod-ux', source: 'dept-prod', target: 'team-ux' },
        { id: 'e-infra-sec', source: 'team-infra', target: 'part-security' },
      ]
    };
  }
}

export const identityService = IdentityService.getInstance();
