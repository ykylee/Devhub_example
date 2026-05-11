"use client";

import { useEffect, useState } from "react";
import { identityService, OrgMember, OrgNode } from "@/lib/services/identity.service";
import { MemberTable } from "@/components/organization/MemberTable";
import { OrgUnitGrid } from "@/components/organization/OrgUnitGrid";
import { MemberManagementModal } from "@/components/organization/MemberManagementModal";
import { PermissionEditor } from "@/components/organization/PermissionEditor";
import { OrgTree } from "@/components/organization/OrgTree";
import { motion, AnimatePresence } from "framer-motion";
import { Building2, Users, Layers, Shield, Search, Filter, Network } from "lucide-react";
import { cn } from "@/lib/utils";
import { useStore } from "@/lib/store";
import { useRouter } from "next/navigation";

import { defaultRoles, Role } from "@/lib/services/rbac.types";
import { rbacService, RbacError } from "@/lib/services/rbac.service";
import { defaultLandingFor, isSystemAdmin } from "@/lib/auth/role-routing";

type Tab = "members" | "units" | "permissions" | "orgchart";

export default function OrganizationPage() {
  const [activeTab, setActiveTab] = useState<Tab>("members");
  const [members, setMembers] = useState<OrgMember[]>([]);
  const [orgNodes, setOrgNodes] = useState<OrgNode[]>([]);
  const [unitMembers, setUnitMembers] = useState<Record<string, string[]>>({});
  // M1-FIX-D: lazy initializer + deep clone so PermissionMatrix toggles
  // do not mutate the baseline through shared nested permission objects.
  const [roles, setRoles] = useState<Role[]>(() => cloneRoles(defaultRoles));
  const [rolesBaseline, setRolesBaseline] = useState<Role[]>(() => cloneRoles(defaultRoles));
  const [rolesError, setRolesError] = useState<string | null>(null);
  const [savingRoles, setSavingRoles] = useState(false);
  const [managingUnitId, setManagingUnitId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [saveError, setSaveError] = useState<string | null>(null);
  const actor = useStore((s) => s.actor);
  const router = useRouter();
  // Guard on actor.role (real permission) rather than the zustand `role`
  // field, which Header's Switch View can simulate. AuthGuard already
  // redirects unauthenticated visitors away; this is a defence-in-depth
  // bounce for non-admin authenticated users that arrive here directly.
  const allowed = isSystemAdmin(actor?.role);

  useEffect(() => {
    if (actor && !allowed) {
      router.replace(defaultLandingFor(actor.role));
    }
  }, [actor, allowed, router]);

  useEffect(() => {
    if (!allowed) return;
    const fetchData = async () => {
      setIsLoading(true);
      try {
        const [usersData, orgData] = await Promise.all([
          identityService.getUsers(),
          identityService.getOrgHierarchy()
        ]);
        setMembers(usersData);
        setOrgNodes(orgData.nodes);

        // Fetch real members for every unit in parallel so OrgUnitGrid can show
        // accurate counts and the management modal opens with the live roster.
        const memberLists = await Promise.all(
          orgData.nodes.map(async (node): Promise<[string, string[]]> => {
            try {
              const list = await identityService.getUnitMembers(node.id);
              return [node.id, list.map((m) => m.id)];
            } catch (error) {
              console.error(`[organization] getUnitMembers ${node.id} failed:`, error);
              return [node.id, []];
            }
          })
        );
        const fetchedUnitMembers: Record<string, string[]> = {};
        memberLists.forEach(([id, ids]) => {
          fetchedUnitMembers[id] = ids;
        });
        setUnitMembers(fetchedUnitMembers);
      } catch (error) {
        console.error("Failed to fetch organization data:", error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [allowed]);

  useEffect(() => {
    if (!allowed) return;
    const loadPolicies = async () => {
      try {
        const { roles: fetched } = await rbacService.listPolicies();
        setRoles(fetched);
        setRolesBaseline(cloneRoles(fetched));
        setRolesError(null);
      } catch (error) {
        console.error("[organization] listPolicies failed:", error);
        setRolesError("Failed to load RBAC policies; showing defaults.");
      }
    };
    loadPolicies();
  }, [allowed]);

  const isRolesDirty = !rolesEqual(roles, rolesBaseline);

  const handleSavePolicies = async () => {
    setSavingRoles(true);
    setRolesError(null);
    try {
      const baselineById = new Map(rolesBaseline.map(r => [r.id, r]));
      const updates = roles
        .filter((role) => {
          const before = baselineById.get(role.id);
          if (!before) return false; // newly created roles use createPolicy, not bulk PUT
          return !permissionsEqual(before.permissions, role.permissions)
            || (!role.system && (before.name !== role.name || before.description !== role.description));
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
      const message = error instanceof RbacError ? `${error.code}: ${error.message}` : (error instanceof Error ? error.message : "Save failed");
      console.error("[organization] updatePolicies failed:", error);
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
      // Optimistically select the newly created role even if listPolicies returned a slightly different shape.
      void created;
    } catch (error) {
      const message = error instanceof RbacError ? `${error.code}: ${error.message}` : (error instanceof Error ? error.message : "Create failed");
      console.error("[organization] createPolicy failed:", error);
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
      const message = error instanceof RbacError ? `${error.code}: ${error.message}` : (error instanceof Error ? error.message : "Delete failed");
      console.error("[organization] deletePolicy failed:", error);
      setRolesError(message);
    } finally {
      setSavingRoles(false);
    }
  };

  const handleReplaceUnitMembers = async (unitId: string, newMemberIds: string[]) => {
    setSaveError(null);
    try {
      const refreshed = await identityService.replaceUnitMembers(unitId, newMemberIds);
      setUnitMembers((prev) => ({ ...prev, [unitId]: refreshed.map((m) => m.id) }));
      setManagingUnitId(null);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to save members";
      console.error("[organization] replaceUnitMembers failed:", error);
      setSaveError(message);
    }
  };

  const tabs = [
    { id: "members", label: "Members", icon: Users },
    { id: "units", label: "Org Units", icon: Layers },
    { id: "orgchart", label: "Org Chart", icon: Network },
    { id: "permissions", label: "Permissions", icon: Shield },
  ];

  return (
    <div className="space-y-10 pb-20 flex flex-col h-full">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
        >
          <h1 className="text-4xl font-extrabold tracking-tight text-foreground dark:text-white mb-2">
            Organization <span className="text-accent">Governance</span>
          </h1>
          <p className="text-muted-foreground text-lg flex items-center gap-2">
            <Building2 className="w-4 h-4 text-accent" /> DevHub Global • <span className="text-foreground dark:text-white font-bold uppercase tracking-widest text-xs bg-accent/20 px-2 py-0.5 rounded border border-accent/20">Enterprise Tier</span>
          </p>
        </motion.div>

        {/* Tab Navigation */}
        <div className="flex p-1.5 glass border-white/10 rounded-2xl gap-1">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as Tab)}
              className={cn(
                "flex items-center gap-2 px-6 py-2.5 rounded-xl text-xs font-black uppercase tracking-widest transition-all relative",
                activeTab === tab.id ? "text-foreground dark:text-white" : "text-muted-foreground hover:text-foreground dark:hover:text-white"
              )}
            >
              {activeTab === tab.id && (
                <motion.div
                  layoutId="active-tab"
                  className="absolute inset-0 bg-white/10 border border-white/10 rounded-xl"
                  transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                />
              )}
              <tab.icon className={cn("w-4 h-4", activeTab === tab.id ? "text-accent" : "text-muted-foreground")} />
              <span className="relative z-10">{tab.label}</span>
            </button>
          ))}
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 space-y-8">
        {/* Search & Filter Bar (Only for members/units) */}
        {(activeTab === 'members' || activeTab === 'units') && (
          <motion.div 
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex items-center gap-4"
          >
            <div className="relative flex-1">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <input 
                type="text" 
                placeholder={`Search ${activeTab}...`}
                className="w-full glass border-white/10 rounded-2xl pl-12 pr-6 py-3.5 text-sm text-foreground dark:text-white placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-accent/30 transition-all"
              />
            </div>
            <button className="glass border-white/10 p-3.5 rounded-2xl text-muted-foreground hover:text-white transition-all">
              <Filter className="w-5 h-5" />
            </button>
          </motion.div>
        )}

        <div className="relative">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center py-40 gap-4">
              <div className="w-12 h-12 border-4 border-accent/20 border-t-accent rounded-full animate-spin" />
              <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Identity Data...</p>
            </div>
          ) : (
            <AnimatePresence mode="wait">
              <motion.div
                key={activeTab}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                transition={{ duration: 0.3 }}
              >
                {activeTab === 'members' && (
                  <MemberTable 
                    members={members} 
                    roles={roles}
                    onUpdateMemberRole={(memberId, newRoleName) => {
                      setMembers(members.map(m => m.id === memberId ? { ...m, role: newRoleName as OrgMember["role"] } : m));
                    }}
                    onMemberCreated={(newMember) => {
                      setMembers(prev => [newMember, ...prev]);
                    }}
                  />
                )}
                {activeTab === 'units' && (
                  <OrgUnitGrid 
                    nodes={orgNodes} 
                    unitMembers={unitMembers} 
                    onManage={(id) => setManagingUnitId(id)} 
                  />
                )}
                {activeTab === 'orgchart' && <OrgTree />}
                {activeTab === 'permissions' && (
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
                )}
              </motion.div>
            </AnimatePresence>
          )}
        </div>
      </div>

      {/* Member Management Modal */}
      <AnimatePresence>
        {managingUnitId && (
          <MemberManagementModal
            unitId={managingUnitId}
            unitName={orgNodes.find(n => n.id === managingUnitId)?.data.label || 'Unknown Unit'}
            allMembers={members}
            currentMemberIds={unitMembers[managingUnitId] || []}
            onClose={() => { setSaveError(null); setManagingUnitId(null); }}
            onSave={(newMemberIds) => handleReplaceUnitMembers(managingUnitId, newMemberIds)}
            saveError={saveError}
          />
        )}
      </AnimatePresence>
    </div>
  );
}

function cloneRoles(roles: Role[]): Role[] {
  return roles.map((role) => ({
    ...role,
    permissions: clonePermissions(role.permissions),
  }));
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
