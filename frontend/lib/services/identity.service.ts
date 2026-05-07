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

function mapBackendUser(u: any): OrgMember {
  return {
    id: u.user_id,
    name: u.display_name,
    email: u.email,
    role: ROLE_BACKEND_TO_UI[u.role] ?? "Developer",
    status: u.status,
    primary_dept_id: u.primary_unit_id ?? "",
    current_dept_id: u.current_unit_id ?? "",
    is_seconded: !!u.is_seconded,
    appointments: (u.appointments ?? []).map((a: any) => ({
      dept_id: a.unit_id,
      role: a.appointment_role,
    })),
    joined_at: typeof u.joined_at === "string" ? u.joined_at.slice(0, 10) : u.joined_at,
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
      const result = await response.json();
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

  async getOrgHierarchy() {
    try {
      const response = await fetch(`/api/v1/organization/hierarchy`);
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const result = await response.json();
      const units: any[] = result.data?.units ?? [];
      const edges: any[] = result.data?.edges ?? [];
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
      const result = await response.json();
      return (result.data ?? []).map(mapBackendUser);
    } catch (error) {
      console.error('[IdentityService] getUnitMembers error:', error);
      return [];
    }
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
