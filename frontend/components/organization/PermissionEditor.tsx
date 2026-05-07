"use client";

import { motion } from "framer-motion";
import { Shield, Lock, Eye, Pencil, Crown } from "lucide-react";
import { cn } from "@/lib/utils";

type Role = "Developer" | "Manager" | "System Admin";
type Resource =
  | "Repositories"
  | "CI Runs"
  | "Risks"
  | "Commands"
  | "Organization"
  | "System Config";
type Permission = "read" | "write" | "admin" | "none";

const RESOURCES: Resource[] = [
  "Repositories",
  "CI Runs",
  "Risks",
  "Commands",
  "Organization",
  "System Config",
];

const ROLES: Role[] = ["Developer", "Manager", "System Admin"];

const MATRIX: Record<Role, Record<Resource, Permission>> = {
  Developer: {
    Repositories: "read",
    "CI Runs": "read",
    Risks: "read",
    Commands: "none",
    Organization: "none",
    "System Config": "none",
  },
  Manager: {
    Repositories: "write",
    "CI Runs": "read",
    Risks: "write",
    Commands: "write",
    Organization: "read",
    "System Config": "none",
  },
  "System Admin": {
    Repositories: "admin",
    "CI Runs": "admin",
    Risks: "admin",
    Commands: "admin",
    Organization: "admin",
    "System Config": "admin",
  },
};

const permissionStyles: Record<Permission, string> = {
  read: "bg-white/5 border-white/20 text-white/70",
  write: "bg-blue-500/10 border-blue-500/30 text-blue-300",
  admin: "bg-accent/15 border-accent/40 text-accent",
  none: "bg-transparent border-white/5 text-muted-foreground/40",
};

const permissionIcon: Record<Permission, typeof Eye | null> = {
  read: Eye,
  write: Pencil,
  admin: Crown,
  none: null,
};

const roleAccent: Record<Role, string> = {
  Developer: "text-white/80",
  Manager: "text-blue-300",
  "System Admin": "text-accent",
};

export function PermissionEditor() {
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
            Read-only preview
          </p>
          <p className="text-[11px] text-muted-foreground mt-0.5">
            Backend RBAC API pending. Matrix below reflects the proposed default policy.
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
              {RESOURCES.map((resource) => (
                <div
                  key={resource}
                  className="text-[10px] font-black uppercase tracking-[0.25em] text-white/60 px-3 py-2 text-center"
                >
                  {resource}
                </div>
              ))}
            </div>

            {ROLES.map((role) => (
              <div
                key={role}
                className="grid gap-2 mb-2"
                style={{ gridTemplateColumns: `200px repeat(${RESOURCES.length}, minmax(120px, 1fr))` }}
              >
                <div
                  className={cn(
                    "px-4 py-3 rounded-xl border border-white/10 bg-white/5 text-xs font-black uppercase tracking-widest flex items-center",
                    roleAccent[role]
                  )}
                >
                  {role}
                </div>
                {RESOURCES.map((resource) => {
                  const permission = MATRIX[role][resource];
                  return (
                    <PermissionCell
                      key={`${role}-${resource}`}
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

function PermissionCell({ permission }: { permission: Permission }) {
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

function LegendChip({ permission, label }: { permission: Permission; label: string }) {
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
