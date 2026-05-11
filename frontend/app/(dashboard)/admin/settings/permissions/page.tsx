"use client";

import { useEffect, useState } from "react";
import { PermissionEditor } from "@/components/organization/PermissionEditor";
import { defaultRoles, Role } from "@/lib/services/rbac.types";
import { rbacService, RbacError } from "@/lib/services/rbac.service";

export default function AdminSettingsPermissionsPage() {
  // M1-FIX-D: lazy initializer + deep clone so PermissionMatrix toggles
  // do not mutate the baseline through shared nested permission objects.
  const [roles, setRoles] = useState<Role[]>(() => cloneRoles(defaultRoles));
  const [rolesBaseline, setRolesBaseline] = useState<Role[]>(() => cloneRoles(defaultRoles));
  const [rolesError, setRolesError] = useState<string | null>(null);
  const [savingRoles, setSavingRoles] = useState(false);

  useEffect(() => {
    const load = async () => {
      try {
        const { roles: fetched } = await rbacService.listPolicies();
        setRoles(fetched);
        setRolesBaseline(cloneRoles(fetched));
        setRolesError(null);
      } catch (error) {
        console.error("[admin/settings/permissions] listPolicies failed:", error);
        setRolesError("Failed to load RBAC policies; showing defaults.");
      }
    };
    load();
  }, []);

  const isRolesDirty = !rolesEqual(roles, rolesBaseline);

  const handleSavePolicies = async () => {
    setSavingRoles(true);
    setRolesError(null);
    try {
      const baselineById = new Map(rolesBaseline.map((r) => [r.id, r]));
      const updates = roles
        .filter((role) => {
          const before = baselineById.get(role.id);
          if (!before) return false;
          return (
            !permissionsEqual(before.permissions, role.permissions) ||
            (!role.system && (before.name !== role.name || before.description !== role.description))
          );
        })
        .map((role) => {
          const payload: { id: string; permissions: Role["permissions"]; name?: string; description?: string } = {
            id: role.id,
            permissions: role.permissions,
          };
          if (!role.system) {
            payload.name = role.name;
            payload.description = role.description;
          }
          return payload;
        });
      if (updates.length === 0) {
        setSavingRoles(false);
        return;
      }
      const { roles: refreshed } = await rbacService.updatePolicies(updates);
      setRoles(refreshed);
      setRolesBaseline(cloneRoles(refreshed));
    } catch (error) {
      const message = error instanceof RbacError ? `${error.code}: ${error.message}` : error instanceof Error ? error.message : "Save failed";
      console.error("[admin/settings/permissions] updatePolicies failed:", error);
      setRolesError(message);
    } finally {
      setSavingRoles(false);
    }
  };

  const handleCreateRole = async (role: Role) => {
    setSavingRoles(true);
    setRolesError(null);
    try {
      const created = await rbacService.createPolicy({
        id: role.id,
        name: role.name,
        description: role.description,
        permissions: role.permissions,
      });
      const { roles: refreshed } = await rbacService.listPolicies();
      setRoles(refreshed);
      setRolesBaseline(cloneRoles(refreshed));
      void created;
    } catch (error) {
      const message = error instanceof RbacError ? `${error.code}: ${error.message}` : error instanceof Error ? error.message : "Create failed";
      console.error("[admin/settings/permissions] createPolicy failed:", error);
      setRolesError(message);
    } finally {
      setSavingRoles(false);
    }
  };

  const handleDeleteRole = async (roleId: string) => {
    setSavingRoles(true);
    setRolesError(null);
    try {
      await rbacService.deletePolicy(roleId);
      const { roles: refreshed } = await rbacService.listPolicies();
      setRoles(refreshed);
      setRolesBaseline(cloneRoles(refreshed));
    } catch (error) {
      const message = error instanceof RbacError ? `${error.code}: ${error.message}` : error instanceof Error ? error.message : "Delete failed";
      console.error("[admin/settings/permissions] deletePolicy failed:", error);
      setRolesError(message);
    } finally {
      setSavingRoles(false);
    }
  };

  return (
    <PermissionEditor
      roles={roles}
      setRoles={setRoles}
      onSave={handleSavePolicies}
      onCreate={handleCreateRole}
      onDelete={handleDeleteRole}
      saving={savingRoles}
      errorMessage={rolesError}
      isDirty={isRolesDirty}
    />
  );
}

function cloneRoles(roles: Role[]): Role[] {
  return roles.map((role) => ({ ...role, permissions: clonePermissions(role.permissions) }));
}

function clonePermissions(permissions: Role["permissions"]): Role["permissions"] {
  const out: Role["permissions"] = {};
  for (const resource of Object.keys(permissions)) {
    out[resource] = { ...permissions[resource] };
  }
  return out;
}

function rolesEqual(a: Role[], b: Role[]): boolean {
  if (a.length !== b.length) return false;
  const byId = new Map(b.map((r) => [r.id, r]));
  for (const role of a) {
    const other = byId.get(role.id);
    if (!other) return false;
    if (role.name !== other.name || role.description !== other.description || role.system !== other.system) return false;
    if (!permissionsEqual(role.permissions, other.permissions)) return false;
  }
  return true;
}

function permissionsEqual(a: Role["permissions"], b: Role["permissions"]): boolean {
  const keys = new Set([...Object.keys(a), ...Object.keys(b)]);
  for (const resource of keys) {
    const aRes = a[resource] || {};
    const bRes = b[resource] || {};
    for (const action of ["view", "create", "edit", "delete"] as const) {
      if (Boolean(aRes[action]) !== Boolean(bRes[action])) return false;
    }
  }
  return true;
}
