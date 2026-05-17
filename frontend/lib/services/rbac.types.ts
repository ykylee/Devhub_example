import { PermissionState } from "@/components/organization/PermissionMatrix";

// Role mirrors the backend GET /api/v1/rbac/policies wire shape from
// docs/backend_api_contract.md section 12.2. The id strings are aligned with
// backend domain.AppRole values ("developer" | "manager" | "system_admin"),
// not the legacy frontend-only "sysadmin" alias.
export interface Role {
  id: string;
  name: string;
  description: string;
  system: boolean;
  permissions: PermissionState;
}

export interface RbacPolicyMeta {
  policy_version: string;
  source: string;
  editable: boolean;
  system_roles: string[];
}

// SYSTEM_ROLE_IDS lists the immutable role ids the backend seeds. UI uses this
// to gate destructive operations on system roles even before the matrix loads.
export const SYSTEM_ROLE_IDS: readonly string[] = ["developer", "manager", "system_admin", "pmo_manager"];

export function isSystemRole(roleId: string): boolean {
  return SYSTEM_ROLE_IDS.includes(roleId);
}

// defaultRoles is the offline fallback used when the backend has not been
// reached yet (initial page paint, dev without rbac_policies seeded). Matrices
// match docs/backend_api_contract.md section 12.1 default policy so the UI is
// internally consistent even without a network round-trip.
// 4 신규 resource (applications / application_repositories / projects / scm_providers)
// 는 sprint claude/work_260514-a (ADR-0011 §4.1) 에서 추가. system_admin 일임 정책 —
// developer/manager 는 모든 axis false, system_admin 만 모든 axis true. backend
// migration 000018 seed 와 정합.
export const defaultRoles: Role[] = [
  {
    id: "system_admin",
    name: "System Admin",
    description: "시스템 설정, 조직/사용자 관리, 운영 command 관리 권한",
    system: true,
    permissions: {
      infrastructure:           { view: true, create: true, edit: true, delete: true },
      pipelines:                { view: true, create: true, edit: true, delete: true },
      organization:             { view: true, create: true, edit: true, delete: true },
      security:                 { view: true, create: true, edit: true, delete: true },
      audit:                    { view: true, create: false, edit: false, delete: false },
      applications:             { view: true, create: true, edit: true, delete: true },
      application_repositories: { view: true, create: true, edit: true, delete: true },
      projects:                 { view: true, create: true, edit: true, delete: true },
      scm_providers:            { view: true, create: true, edit: true, delete: true },
      dev_requests:             { view: true, create: true, edit: true, delete: true },
      dev_request_intake_tokens: { view: true, create: true, edit: true, delete: true },
    },
  },
  {
    id: "manager",
    name: "Manager",
    description: "팀 운영, risk triage, 승인 전 command 생성 권한",
    system: true,
    permissions: {
      infrastructure:           { view: true, create: false, edit: false, delete: false },
      pipelines:                { view: true, create: false, edit: false, delete: false },
      organization:             { view: true, create: false, edit: false, delete: false },
      security:                 { view: true, create: true,  edit: false, delete: false },
      audit:                    { view: true, create: false, edit: false, delete: false },
      applications:             { view: false, create: false, edit: false, delete: false },
      application_repositories: { view: false, create: false, edit: false, delete: false },
      projects:                 { view: false, create: false, edit: false, delete: false },
      scm_providers:            { view: false, create: false, edit: false, delete: false },
      dev_requests:             { view: true, create: false, edit: false, delete: false },
      dev_request_intake_tokens: { view: false, create: false, edit: false, delete: false },
    },
  },
  {
    id: "developer",
    name: "Developer",
    description: "개발자 대시보드, 본인 관련 repository/CI/risk 조회 권한",
    system: true,
    permissions: {
      infrastructure:           { view: true, create: false, edit: false, delete: false },
      pipelines:                { view: true, create: false, edit: false, delete: false },
      organization:             { view: true, create: false, edit: false, delete: false },
      security:                 { view: true, create: false, edit: false, delete: false },
      audit:                    { view: false, create: false, edit: false, delete: false },
      applications:             { view: false, create: false, edit: false, delete: false },
      application_repositories: { view: false, create: false, edit: false, delete: false },
      projects:                 { view: false, create: false, edit: false, delete: false },
      scm_providers:            { view: false, create: false, edit: false, delete: false },
      dev_requests:             { view: true, create: false, edit: false, delete: false },
      dev_request_intake_tokens: { view: false, create: false, edit: false, delete: false },
    },
  },
  {
    id: "pmo_manager",
    name: "PMO Manager",
    description: "Application 수정 + Project 운영/멤버 관리 위양. 시스템/계정/RBAC 변경 금지.",
    system: true,
    permissions: {
      infrastructure:           { view: true, create: false, edit: false, delete: false },
      pipelines:                { view: true, create: false, edit: false, delete: false },
      organization:             { view: true, create: false, edit: false, delete: false },
      security:                 { view: true, create: false, edit: false, delete: false },
      audit:                    { view: true, create: false, edit: false, delete: false },
      applications:             { view: true, create: false, edit: true,  delete: false },
      application_repositories: { view: true, create: false, edit: false, delete: false },
      projects:                 { view: true, create: true,  edit: true,  delete: true },
      scm_providers:            { view: true, create: false, edit: false, delete: false },
      dev_requests:             { view: true, create: false, edit: true,  delete: false },
      dev_request_intake_tokens: { view: false, create: false, edit: false, delete: false },
    },
  },
];

// AUDIT_LOCKED_ACTIONS encodes the section 12.0.4 invariant: audit:create/edit/
// delete must always be false. PermissionEditor passes this to PermissionMatrix
// to render those cells as locked even before the backend rejects a write.
export const AUDIT_LOCKED_ACTIONS: { [resource: string]: { [action: string]: true } } = {
  audit: { create: true, edit: true, delete: true },
};
