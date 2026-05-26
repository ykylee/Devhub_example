"use client";

import { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { identityService, OrgMember } from "@/lib/services/identity.service";
import { MemberTable } from "@/components/organization/MemberTable";
import { defaultRoles, Role } from "@/lib/services/rbac.types";
import { rbacService } from "@/lib/services/rbac.service";
import { FilterBar } from "@/components/ui/FilterBar";
import { useToast } from "@/components/ui/Toast";
import { PendingReviewPanel } from "@/components/admin/users/PendingReviewPanel";

export default function AdminSettingsUsersPage() {
  const [members, setMembers] = useState<OrgMember[]>([]);
  const [unitLeaderIds, setUnitLeaderIds] = useState<string[]>([]);
  const [unitNames, setUnitNames] = useState<Record<string, string>>({});
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
        const unitNameMap: Record<string, string> = {};
        for (const n of hierarchy.nodes) {
          if (n.data.label) unitNameMap[n.id] = n.data.label;
        }
        setUnitNames(unitNameMap);
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
          unitNames={unitNames}
          roles={roles}
          onUpdateMemberRole={handleUpdateRole}
          onMemberCreated={(newMember) => {
            setMembers((prev) => [newMember, ...prev]);
          }}
          onMemberUpdated={(updated) => {
            setMembers((prev) => prev.map((m) => (m.id === updated.id ? updated : m)));
          }}
          onMemberDeleted={(userId) => {
            setMembers((prev) => prev.filter((m) => m.id !== userId));
          }}
        />
      )}
    </div>
  );
}
