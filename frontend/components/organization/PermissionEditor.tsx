"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { Shield, Lock, Eye, Pencil, Crown } from "lucide-react";
import { cn } from "@/lib/utils";
import { rbacService } from "@/lib/services/rbac.service";
import type { RbacPermission, RbacPolicy } from "@/lib/services/types";

const FALLBACK_POLICY: RbacPolicy = {
  roles: [
    { role: "developer", label: "Developer", description: "Developer dashboard access" },
    { role: "manager", label: "Manager", description: "Team operations access" },
    { role: "system_admin", label: "System Admin", description: "System administration access" },
  ],
  resources: [
    { resource: "repositories", label: "Repositories", description: "Repository metadata" },
    { resource: "ci_runs", label: "CI Runs", description: "CI run status and logs" },
    { resource: "risks", label: "Risks", description: "Risk and mitigation workflows" },
    { resource: "commands", label: "Commands", description: "Command lifecycle" },
    { resource: "organization", label: "Organization", description: "Users and org units" },
    { resource: "system_config", label: "System Config", description: "System configuration" },
  ],
  permissions: [
    { permission: "none", label: "None", rank: 0, description: "No access" },
    { permission: "read", label: "Read", rank: 10, description: "Read access" },
    { permission: "write", label: "Write", rank: 20, description: "Write access" },
    { permission: "admin", label: "Admin", rank: 30, description: "Admin access" },
  ],
  matrix: {
    developer: {
      repositories: "read",
      ci_runs: "read",
      risks: "read",
      commands: "none",
      organization: "none",
      system_config: "none",
    },
    manager: {
      repositories: "write",
      ci_runs: "read",
      risks: "write",
      commands: "write",
      organization: "read",
      system_config: "none",
    },
    system_admin: {
      repositories: "admin",
      ci_runs: "admin",
      risks: "admin",
      commands: "admin",
      organization: "admin",
      system_config: "admin",
    },
  },
};

const permissionStyles: Record<RbacPermission, string> = {
  read: "bg-white/5 border-white/20 text-white/70",
  write: "bg-blue-500/10 border-blue-500/30 text-blue-300",
  admin: "bg-accent/15 border-accent/40 text-accent",
  none: "bg-transparent border-white/5 text-muted-foreground/40",
};

const permissionIcon: Record<RbacPermission, typeof Eye | null> = {
  read: Eye,
  write: Pencil,
  admin: Crown,
  none: null,
};

const roleAccent: Record<string, string> = {
  developer: "text-white/80",
  manager: "text-blue-300",
  system_admin: "text-accent",
};

export function PermissionEditor() {
  const [policy, setPolicy] = useState<RbacPolicy>(FALLBACK_POLICY);
  const [policyState, setPolicyState] = useState<"loading" | "api" | "fallback">("loading");

  useEffect(() => {
    let cancelled = false;
    rbacService.getPolicy()
      .then((nextPolicy) => {
        if (!cancelled) {
          setPolicy(nextPolicy);
          setPolicyState("api");
        }
      })
      .catch((error) => {
        console.error("[PermissionEditor] getPolicy error:", error);
        if (!cancelled) {
          setPolicy(FALLBACK_POLICY);
          setPolicyState("fallback");
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
      className="space-y-6"
    >
      <div className="glass border border-yellow-400/20 bg-yellow-400/5 rounded-2xl px-5 py-4 flex items-start gap-3">
        <div className="p-2 rounded-xl bg-yellow-400/10 border border-yellow-400/20 shrink-0">
          <Lock className="w-4 h-4 text-yellow-300" />
        </div>
        <div>
          <p className="text-xs font-black uppercase tracking-widest text-yellow-300">
            {policyState === "api" ? "Read-only policy" : "Default policy"}
          </p>
          <p className="text-[11px] text-muted-foreground mt-0.5">
            {policyState === "api"
              ? "Current access matrix is active. Editing remains disabled."
              : policyState === "loading"
                ? "Loading access matrix."
                : "Default access matrix is active. Editing remains disabled."}
          </p>
        </div>
      </div>

      <div className="glass-card p-6 relative overflow-hidden">
        <div className="absolute -top-20 -right-20 w-64 h-64 blur-3xl opacity-10 pointer-events-none bg-accent" />

        <div className="flex items-center justify-between mb-6 relative z-10">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-2xl bg-accent/10 border border-accent/20">
              <Shield className="w-5 h-5 text-accent" />
            </div>
            <div>
              <h3 className="text-lg font-black text-white uppercase tracking-tight">
                Role-Based <span className="text-accent">Access Control</span>
              </h3>
              <p className="text-[11px] text-muted-foreground font-bold">
                Permissions matrix preview
              </p>
            </div>
          </div>
          <div className="hidden md:flex items-center gap-3 text-[10px] font-black uppercase tracking-widest">
            <LegendChip permission="read" label="Read" />
            <LegendChip permission="write" label="Write" />
            <LegendChip permission="admin" label="Admin" />
          </div>
        </div>

        <div className="overflow-x-auto relative z-10">
          <div className="min-w-[720px]">
            <div
              className="grid gap-2 mb-2"
              style={{ gridTemplateColumns: `200px repeat(${RESOURCES.length}, minmax(120px, 1fr))` }}
            >
              <div className="text-[10px] font-black uppercase tracking-[0.3em] text-muted-foreground px-3 py-2">
                Role / Resource
              </div>
              {policy.resources.map((resource) => (
                <div
                  key={resource.resource}
                  className="text-[10px] font-black uppercase tracking-[0.25em] text-white/60 px-3 py-2 text-center"
                >
                  {resource.label}
                </div>
              ))}
            </div>

            {policy.roles.map((role) => (
              <div
                key={role.role}
                className="grid gap-2 mb-2"
                style={{ gridTemplateColumns: `200px repeat(${policy.resources.length}, minmax(120px, 1fr))` }}
              >
                <div
                  className={cn(
                    "px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-xs font-black uppercase tracking-widest flex items-center",
                    roleAccent[role.role] ?? "text-white/80"
                  )}
                >
                  {role.label}
                </div>
                {policy.resources.map((resource) => {
                  const permission = normalizePermission(policy.matrix[role.role]?.[resource.resource]);
                  return (
                    <PermissionCell
                      key={`${role.role}-${resource.resource}`}
                      permission={permission}
                    />
                  );
                })}
              </div>
            ))}
          </div>
        </div>

        <p className="mt-6 pt-4 border-t border-white/5 text-[10px] font-bold text-muted-foreground/70 uppercase tracking-widest relative z-10">
          Editing disabled until backend policy service is online.
        </p>
      </div>
    </motion.div>
  );
}

function normalizePermission(permission?: string): RbacPermission {
  if (permission === "read" || permission === "write" || permission === "admin" || permission === "none") {
    return permission;
  }
  return "none";
}

function PermissionCell({ permission }: { permission: RbacPermission }) {
  const Icon = permissionIcon[permission];
  const label = permission === "none" ? "—" : permission;
  return (
    <button
      type="button"
      disabled
      aria-disabled="true"
      className={cn(
        "px-3 py-3 rounded-xl border text-[10px] font-black uppercase tracking-widest cursor-not-allowed flex items-center justify-center gap-1.5 transition-colors hover:bg-white/[0.02]",
        permissionStyles[permission]
      )}
    >
      {Icon && <Icon className="w-3 h-3" />}
      {label}
    </button>
  );
}

function LegendChip({ permission, label }: { permission: RbacPermission; label: string }) {
  const Icon = permissionIcon[permission];
  return (
    <span
      className={cn(
        "px-2.5 py-1 rounded-lg border flex items-center gap-1.5",
        permissionStyles[permission]
      )}
    >
      {Icon && <Icon className="w-3 h-3" />}
      {label}
    </span>
  );
}
