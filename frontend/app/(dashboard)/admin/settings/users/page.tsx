"use client";

import { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { identityService, OrgMember } from "@/lib/services/identity.service";
import { MemberTable } from "@/components/organization/MemberTable";
import { defaultRoles, Role } from "@/lib/services/rbac.types";
import { rbacService } from "@/lib/services/rbac.service";
import { FilterBar } from "@/components/ui/FilterBar";
import { useToast } from "@/components/ui/Toast";
import { ExternalLink, Info } from "lucide-react";
import { getKCAdminConsoleUrl } from "@/lib/config/endpoints";
import { useStore } from "@/lib/store";
import { PendingReviewPanel } from "@/components/admin/users/PendingReviewPanel";

export default function AdminSettingsUsersPage() {
  const { role: currentUserRole } = useStore();
  const [members, setMembers] = useState<OrgMember[]>([]);
  const [unitLeaderIds, setUnitLeaderIds] = useState<string[]>([]);
  const [roles, setRoles] = useState<Role[]>(defaultRoles);
  const [isLoading, setIsLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [activeRole, setActiveRole] = useState("all");
  const { toast } = useToast();

  const roleOptions = useMemo(() => {
    const options = [{ label: "All Roles", value: "all" }];
    roles.forEach(r => {
      options.push({ label: r.name, value: r.name });
    });
    return options;
  }, [roles]);

  const filteredMembers = useMemo(() => {
    const q = query.trim().toLowerCase();
    return members.filter((m) => {
      const matchesQuery = !q || 
        m.name.toLowerCase().includes(q) ||
        m.email.toLowerCase().includes(q) ||
        m.role.toLowerCase().includes(q);
      
      const matchesRole = activeRole === "all" || m.role === activeRole;
      
      return matchesQuery && matchesRole;
    });
  }, [members, query, activeRole]);

  useEffect(() => {
    const load = async () => {
      setIsLoading(true);
      try {
        const [usersData, policy, hierarchy] = await Promise.all([
          identityService.getUsers(),
          rbacService.listPolicies().catch(() => ({ roles: defaultRoles })),
          identityService.getOrgHierarchy().catch(() => ({ nodes: [], edges: [] })),
        ]);
        setMembers(usersData);
        setRoles(policy.roles);
        const leaders = hierarchy.nodes
          .map((n) => n.data.leader_id)
          .filter((v): v is string => Boolean(v));
        setUnitLeaderIds(Array.from(new Set(leaders)));
      } catch (error) {
        console.error("[admin/settings/users] load failed:", error);
      } finally {
        setIsLoading(false);
      }
    };
    load();
  }, []);

  const handleUpdateRole = async (memberId: string, newRoleName: string) => {
    try {
      // Optimistic UI update
      setMembers((prev) => 
        prev.map((m) => (m.id === memberId ? { ...m, role: newRoleName as OrgMember["role"] } : m))
      );

      await identityService.updateUser(memberId, { 
        role: newRoleName as OrgMember["role"] 
      });
      
      toast(`User role updated to ${newRoleName}`, "success");
    } catch (error) {
      console.error("[admin/settings/users] handleUpdateRole failed:", error);
      toast("Failed to update user role", "error");
      
      // Rollback on failure
      const refreshedUsers = await identityService.getUsers();
      setMembers(refreshedUsers);
    }
  };

  return (
    <div className="space-y-8">
      {currentUserRole === "System Admin" && (() => {
        const kcAdminUrl = getKCAdminConsoleUrl();
        return (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="glass-card p-4 border-accent/20 bg-accent/5 flex items-start gap-3"
          >
            <div className="p-2 bg-accent/20 rounded-lg">
              <Info className="w-4 h-4 text-accent" />
            </div>
            <div className="flex-1 space-y-1">
              <p className="text-xs font-bold text-foreground dark:text-primary-foreground uppercase tracking-tight">Account Management moved to Keycloak</p>
              <p className="text-[10px] text-muted-foreground">For security and audit consistency, administrative account operations (Issue, Password Reset, Disable) are now managed through the central IdP console.</p>
              {kcAdminUrl ? (
                <a
                  href={kcAdminUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1.5 text-[10px] font-black text-accent hover:text-accent/80 transition-colors uppercase tracking-widest pt-2 group"
                >
                  Open Keycloak Admin Console <ExternalLink className="w-3 h-3 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
                </a>
              ) : (
                <p className="text-[10px] text-muted-foreground/70 italic pt-2">
                  Keycloak Admin Console URL unavailable — set <span className="font-mono">NEXT_PUBLIC_OIDC_ISSUER_URL</span> or <span className="font-mono">NEXT_PUBLIC_KC_ADMIN_URL</span>.
                </p>
              )}
            </div>
          </motion.div>
        );
      })()}

      {!isLoading && (
        <PendingReviewPanel
          members={members}
          onReviewed={(userId) => {
            setMembers((prev) =>
              prev.map((m) =>
                m.id === userId ? { ...m, review_status: "reviewed" } : m,
              ),
            );
            toast("검토를 확정했습니다.", "success");
          }}
        />
      )}

      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
        <FilterBar
          onSearch={setQuery}
          onFilterChange={setActiveRole}
          activeFilter={activeRole}
          filterOptions={roleOptions}
          placeholder="Search by name, email, or role..."
          searchLabel="Search users"
        />
      </motion.div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-accent/20 border-t-accent rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Users...</p>
        </div>
      ) : (
        <MemberTable
          members={filteredMembers}
          unitLeaderIds={unitLeaderIds}
          roles={roles}
          onUpdateMemberRole={handleUpdateRole}
          onMemberCreated={(newMember) => {
            setMembers((prev) => [newMember, ...prev]);
          }}
        />
      )}
    </div>
  );
}
