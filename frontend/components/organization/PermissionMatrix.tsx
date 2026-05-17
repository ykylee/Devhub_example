"use client";

import { Check, X } from "lucide-react";
import { cn } from "@/lib/utils";

export type ActionType = "view" | "create" | "edit" | "delete";
// Resource list mirrors backend domain.AllResources() (9개). 신규 4종은 sprint
// claude/work_260514-a (ADR-0011 §4.1) 에서 추가 — Application/Repository/Project
// 도메인의 RBAC 1차 enforcement.
export type ResourceType =
  | "infrastructure"
  | "pipelines"
  | "organization"
  | "security"
  | "audit"
  | "applications"
  | "application_repositories"
  | "projects"
  | "scm_providers"
  | "dev_requests"
  | "dev_request_intake_tokens";

export interface PermissionState {
  [resource: string]: {
    [action in ActionType]?: boolean;
  };
}

interface PermissionMatrixProps {
  permissions: PermissionState;
  onChange?: (newPermissions: PermissionState) => void;
  readOnly?: boolean;
  // lockedCells marks per-(resource, action) cells as non-editable, on top of
  // the global readOnly flag. Used to enforce the section 12.0.4 audit
  // invariant in the UI before the request reaches the backend.
  lockedCells?: { [resource: string]: { [action: string]: true } };
}

const resources: { id: ResourceType; label: string }[] = [
  { id: "infrastructure", label: "Infrastructure & Topology" },
  { id: "pipelines", label: "CI/CD Pipelines" },
  { id: "organization", label: "Organization & Members" },
  { id: "security", label: "Risk & Security" },
  { id: "audit", label: "Audit Logs & History" },
  { id: "applications", label: "Applications" },
  { id: "application_repositories", label: "Application Repositories" },
  { id: "projects", label: "Projects" },
  { id: "scm_providers", label: "SCM Providers" },
  { id: "dev_requests", label: "Development Requests (DREQ)" },
  { id: "dev_request_intake_tokens", label: "DREQ Intake Tokens" },
];

const actions: { id: ActionType; label: string }[] = [
  { id: "view", label: "View" },
  { id: "create", label: "Create" },
  { id: "edit", label: "Edit" },
  { id: "delete", label: "Delete" },
];

export function PermissionMatrix({ permissions, onChange, readOnly = false, lockedCells }: PermissionMatrixProps) {
  const isCellLocked = (resource: ResourceType, action: ActionType): boolean => {
    return Boolean(lockedCells?.[resource]?.[action]);
  };

  const handleToggle = (resource: ResourceType, action: ActionType) => {
    if (readOnly || !onChange) return;
    if (isCellLocked(resource, action)) return;

    const newPermissions = { ...permissions };
    if (!newPermissions[resource]) {
      newPermissions[resource] = {};
    }

    newPermissions[resource][action] = !newPermissions[resource][action];
    onChange(newPermissions);
  };

  return (
    <div className="overflow-x-auto rounded-2xl border border-border bg-card/40">
      <table className="w-full text-left text-sm whitespace-nowrap">
        <thead className="bg-muted/30 uppercase text-xs font-black text-muted-foreground tracking-widest border-b border-border">
          <tr>
            <th className="p-4">Resource</th>
            {actions.map((action) => (
              <th key={action.id} className="p-4 text-center">{action.label}</th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {resources.map((resource) => (
            <tr key={resource.id} className="hover:bg-muted/30 transition-colors">
              <td className="p-4 font-bold text-foreground">{resource.label}</td>
              {actions.map((action) => {
                const isGranted = permissions[resource.id]?.[action.id] || false;
                const cellLocked = isCellLocked(resource.id, action.id);
                const disabled = readOnly || cellLocked;
                return (
                  <td key={action.id} className="p-4 text-center">
                    <button
                      type="button"
                      disabled={disabled}
                      title={cellLocked ? "Append-only by system code" : undefined}
                      onClick={() => handleToggle(resource.id, action.id)}
                      className={cn(
                        "inline-flex items-center justify-center w-8 h-8 rounded-lg transition-all",
                        isGranted
                          ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30"
                          : "bg-muted/40 text-muted-foreground border border-border hover:bg-muted/70",
                        disabled && "cursor-not-allowed opacity-70"
                      )}
                    >
                      {isGranted ? <Check className="w-4 h-4" /> : <X className="w-4 h-4" />}
                    </button>
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
