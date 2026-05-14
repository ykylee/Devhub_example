"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { ShieldCheck, Lock, Plus, Trash2, ShieldAlert, Eye, Pencil, Crown, Shield, Save } from "lucide-react";
import { cn } from "@/lib/utils";
import { PermissionMatrix, PermissionState } from "./PermissionMatrix";
import { Role, AUDIT_LOCKED_ACTIONS } from "@/lib/services/rbac.types";

interface PermissionEditorProps {
  roles: Role[];
  setRoles: (roles: Role[]) => void;
  onSave?: () => Promise<void> | void;
  onCreate?: (role: Role) => Promise<void> | void;
  onDelete?: (roleId: string) => Promise<void> | void;
  saving?: boolean;
  errorMessage?: string | null;
  isDirty?: boolean;
}

export function PermissionEditor({
  roles,
  setRoles,
  onSave,
  onCreate,
  onDelete,
  saving = false,
  errorMessage,
  isDirty = false,
}: PermissionEditorProps) {
  const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null);

  const selectedRole = roles.find(r => r.id === selectedRoleId);

  const handleCreateRole = async () => {
    const newRole: Role = {
      id: `custom-${Date.now()}`,
      name: "New Custom Role",
      description: "Define custom access policies for this role.",
      system: false,
      permissions: {},
    };
    if (onCreate) {
      await onCreate(newRole);
    } else {
      setRoles([...roles, newRole]);
    }
    setSelectedRoleId(newRole.id);
  };

  const handleDeleteRole = async (id: string) => {
    const target = roles.find(r => r.id === id);
    if (target?.system) return;
    if (onDelete) {
      await onDelete(id);
    } else {
      setRoles(roles.filter(r => r.id !== id));
    }
    if (selectedRoleId === id) setSelectedRoleId(null);
  };

  const handleUpdatePermissions = (newPermissions: PermissionState) => {
    if (!selectedRoleId) return;
    setRoles(roles.map(r =>
      r.id === selectedRoleId ? { ...r, permissions: newPermissions } : r
    ));
  };

  return (
    <div className="space-y-10">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-xl bg-primary/20 border border-primary/30">
            <Lock className="w-5 h-5 text-primary" />
          </div>
          <div>
            <h3 className="text-xl font-black text-foreground uppercase tracking-tight">RBAC <span className="text-primary">Policies</span></h3>
            <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest mt-1">
              Role-Based Access Control Configuration
            </p>
          </div>
        </div>

        <div className="flex items-center gap-6">
          <div className="hidden md:flex items-center gap-3 text-[10px] font-black uppercase tracking-widest">
            <LegendChip type="read" label="Read" />
            <LegendChip type="write" label="Write" />
            <LegendChip type="admin" label="Admin" />
          </div>
          {onSave && (
            <button
              onClick={() => onSave()}
              disabled={saving || !isDirty}
              className={cn(
                "flex items-center gap-2 px-4 py-2 font-bold rounded-xl transition-all text-sm",
                saving || !isDirty
                  ? "bg-muted/50 text-muted-foreground cursor-not-allowed"
                  : "bg-emerald-500 text-emerald-50 hover:bg-emerald-500/90"
              )}
            >
              <Save className="w-4 h-4" />
              {saving ? "Saving…" : "Save Permissions"}
            </button>
          )}
          <button
            onClick={handleCreateRole}
            disabled={saving}
            className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground font-bold rounded-xl hover:bg-primary/90 transition-all text-sm disabled:opacity-50"
          >
            <Plus className="w-4 h-4" />
            Create Role
          </button>
        </div>
      </div>

      {errorMessage && (
        <div className="rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300">
          {errorMessage}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
        <div className="lg:col-span-5 space-y-4 max-h-[600px] overflow-y-auto pr-2 custom-scrollbar">
          <AnimatePresence>
            {roles.map((role) => {
              const isSelected = selectedRoleId === role.id;
              return (
                <motion.div
                  key={role.id}
                  layout
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                  onClick={() => setSelectedRoleId(role.id)}
                  className={cn(
                    "rounded-2xl p-6 relative overflow-hidden cursor-pointer transition-all border",
                    isSelected
                      ? "bg-primary/10 border-primary/50 shadow-[0_0_30px_rgba(var(--primary),0.15)]"
                      : "glass border-border hover:border-border/80"
                  )}
                >
                  <div className={cn("absolute top-0 left-0 w-1.5 h-full",
                    role.id === 'system_admin' ? "bg-orange-500" :
                    role.id === 'manager' ? "bg-emerald-500" :
                    isSelected ? "bg-primary" : "bg-blue-500"
                  )} />

                  <div className="flex flex-col gap-3 ml-2">
                    <div className="flex items-center justify-between">
                      <h4 className={cn("text-lg font-black", isSelected ? "text-primary" : "text-foreground")}>
                        {role.name}
                      </h4>
                      {!role.system && (
                        <button
                          onClick={(e) => { e.stopPropagation(); handleDeleteRole(role.id); }}
                          disabled={saving}
                          className="p-1.5 rounded-lg text-muted-foreground hover:text-red-400 hover:bg-red-400/10 transition-colors disabled:opacity-30"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground leading-relaxed line-clamp-2">
                      {role.description}
                    </p>
                    <div className="flex items-center gap-2 mt-2">
                      <ShieldCheck className={cn("w-4 h-4", isSelected ? "text-primary/70" : "text-muted-foreground")} />
                      <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                        {Object.keys(role.permissions).length} Resources Configured
                      </span>
                      {role.system && (
                        <span className="text-[10px] font-bold text-orange-400/70 uppercase tracking-wider ml-1">
                          System
                        </span>
                      )}
                    </div>
                  </div>
                </motion.div>
              );
            })}
          </AnimatePresence>
        </div>

        <div className="lg:col-span-7">
          <AnimatePresence mode="wait">
            {selectedRole ? (
              <motion.div
                key={selectedRole.id}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                className="glass rounded-3xl p-8 border-border h-full flex flex-col"
              >
                <div className="mb-8">
                  <div className="flex items-center gap-3 mb-2">
                    <Shield className="w-5 h-5 text-accent" />
                    <h3 className="text-2xl font-black text-foreground">{selectedRole.name} Matrix</h3>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Configure fine-grained access policies for <strong className="text-foreground">{selectedRole.name}</strong>.
                    {" Audit logs are append-only — write actions on the audit row stay locked."}
                  </p>
                </div>

                <div className="flex-1">
                  <PermissionMatrix
                    permissions={selectedRole.permissions}
                    onChange={handleUpdatePermissions}
                    lockedCells={AUDIT_LOCKED_ACTIONS}
                  />
                </div>
              </motion.div>
            ) : (
              <motion.section
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="glass rounded-3xl p-10 border-dashed border-border text-center h-full flex flex-col items-center justify-center"
              >
                <ShieldAlert className="w-16 h-16 text-primary-foreground/10 mb-4" />
                <h4 className="text-xl font-black text-muted-foreground mb-2">Select a Role</h4>
                <p className="text-sm text-muted-foreground max-w-sm">
                  Choose a role from the left panel to inspect or edit its permission matrix.
                </p>
              </motion.section>
            )}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
}

function LegendChip({ type, label }: { type: 'read' | 'write' | 'admin'; label: string }) {
  const icons = {
    read: Eye,
    write: Pencil,
    admin: Crown
  };
  const styles = {
    read: "bg-muted/40 border-border text-foreground",
    write: "bg-blue-500/10 border-blue-500/30 text-blue-300",
    admin: "bg-accent/15 border-accent/40 text-accent"
  };
  const Icon = icons[type];

  return (
    <span className={cn("px-2.5 py-1 rounded-lg border flex items-center gap-1.5", styles[type])}>
      <Icon className="w-3 h-3" />
      {label}
    </span>
  );
}
