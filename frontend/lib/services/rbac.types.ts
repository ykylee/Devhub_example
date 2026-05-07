import { PermissionState } from "@/components/organization/PermissionMatrix";

export interface Role {
  id: string;
  name: string;
  description: string;
  permissions: PermissionState;
}

export const defaultRoles: Role[] = [
  { 
    id: "sysadmin",
    name: "System Admin", 
    description: "Full access to all resources and platform settings.",
    permissions: {
      infrastructure: { view: true, create: true, edit: true, delete: true },
      pipelines: { view: true, create: true, edit: true, delete: true },
      organization: { view: true, create: true, edit: true, delete: true },
      security: { view: true, create: true, edit: true, delete: true },
      audit: { view: true, create: true, edit: true, delete: true },
    }
  },
  { 
    id: "manager",
    name: "Manager", 
    description: "Team and project level management with operational oversight.",
    permissions: {
      infrastructure: { view: true, create: false, edit: true, delete: false },
      pipelines: { view: true, create: true, edit: true, delete: true },
      organization: { view: true, create: true, edit: true, delete: false },
      security: { view: true, create: true, edit: true, delete: false },
      audit: { view: true, create: false, edit: false, delete: false },
    }
  },
  { 
    id: "developer",
    name: "Developer", 
    description: "Source code and CI/CD development access.",
    permissions: {
      infrastructure: { view: true, create: false, edit: false, delete: false },
      pipelines: { view: true, create: true, edit: false, delete: false },
      organization: { view: true, create: false, edit: false, delete: false },
      security: { view: false, create: false, edit: false, delete: false },
      audit: { view: false, create: false, edit: false, delete: false },
    }
  }
];
