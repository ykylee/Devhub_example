"use client";

import { motion } from "framer-motion";
import { Check, X } from "lucide-react";
import { cn } from "@/lib/utils";

export type ActionType = "view" | "create" | "edit" | "delete";
export type ResourceType = "infrastructure" | "pipelines" | "organization" | "security" | "audit";

export interface PermissionState {
  [resource: string]: {
    [action in ActionType]?: boolean;
  };
}

interface PermissionMatrixProps {
  permissions: PermissionState;
  onChange?: (newPermissions: PermissionState) => void;
  readOnly?: boolean;
}

const resources: { id: ResourceType; label: string }[] = [
  { id: "infrastructure", label: "Infrastructure & Topology" },
  { id: "pipelines", label: "CI/CD Pipelines" },
  { id: "organization", label: "Organization & Members" },
  { id: "security", label: "Risk & Security" },
  { id: "audit", label: "Audit Logs & History" },
];

const actions: { id: ActionType; label: string }[] = [
  { id: "view", label: "View" },
  { id: "create", label: "Create" },
  { id: "edit", label: "Edit" },
  { id: "delete", label: "Delete" },
];

export function PermissionMatrix({ permissions, onChange, readOnly = false }: PermissionMatrixProps) {
  const handleToggle = (resource: ResourceType, action: ActionType) => {
    if (readOnly || !onChange) return;
    
    const newPermissions = { ...permissions };
    if (!newPermissions[resource]) {
      newPermissions[resource] = {};
    }
    
    newPermissions[resource][action] = !newPermissions[resource][action];
    onChange(newPermissions);
  };

  return (
    <div className="overflow-x-auto rounded-2xl border border-white/10 bg-black/20">
      <table className="w-full text-left text-sm whitespace-nowrap">
        <thead className="bg-white/5 uppercase text-xs font-black text-muted-foreground tracking-widest border-b border-white/10">
          <tr>
            <th className="p-4">Resource</th>
            {actions.map((action) => (
              <th key={action.id} className="p-4 text-center">{action.label}</th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-white/10">
          {resources.map((resource) => (
            <tr key={resource.id} className="hover:bg-white/5 transition-colors">
              <td className="p-4 font-bold text-white/80">{resource.label}</td>
              {actions.map((action) => {
                const isGranted = permissions[resource.id]?.[action.id] || false;
                return (
                  <td key={action.id} className="p-4 text-center">
                    <button
                      type="button"
                      disabled={readOnly}
                      onClick={() => handleToggle(resource.id, action.id)}
                      className={cn(
                        "inline-flex items-center justify-center w-8 h-8 rounded-lg transition-all",
                        isGranted
                          ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30"
                          : "bg-white/5 text-white/20 border border-white/5 hover:bg-white/10",
                        readOnly && "cursor-not-allowed opacity-70"
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
