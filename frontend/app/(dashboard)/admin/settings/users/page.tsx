"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { Search, Filter } from "lucide-react";
import { identityService, OrgMember } from "@/lib/services/identity.service";
import { MemberTable } from "@/components/organization/MemberTable";
import { defaultRoles, Role } from "@/lib/services/rbac.types";
import { rbacService } from "@/lib/services/rbac.service";

export default function AdminSettingsUsersPage() {
  const [members, setMembers] = useState<OrgMember[]>([]);
  const [roles, setRoles] = useState<Role[]>(defaultRoles);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      setIsLoading(true);
      try {
        const [usersData, policy] = await Promise.all([
          identityService.getUsers(),
          rbacService.listPolicies().catch(() => ({ roles: defaultRoles })),
        ]);
        setMembers(usersData);
        setRoles(policy.roles);
      } catch (error) {
        console.error("[admin/settings/users] load failed:", error);
      } finally {
        setIsLoading(false);
      }
    };
    load();
  }, []);

  return (
    <div className="space-y-8">
      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex items-center gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search users..."
            className="w-full glass border-white/10 rounded-2xl pl-12 pr-6 py-3.5 text-sm text-foreground dark:text-white placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-accent/30 transition-all"
          />
        </div>
        <button className="glass border-white/10 p-3.5 rounded-2xl text-muted-foreground hover:text-white transition-all">
          <Filter className="w-5 h-5" />
        </button>
      </motion.div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-accent/20 border-t-accent rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Users...</p>
        </div>
      ) : (
        <MemberTable
          members={members}
          roles={roles}
          onUpdateMemberRole={(memberId, newRoleName) => {
            setMembers((prev) => prev.map((m) => (m.id === memberId ? { ...m, role: newRoleName as OrgMember["role"] } : m)));
          }}
          onMemberCreated={(newMember) => {
            setMembers((prev) => [newMember, ...prev]);
          }}
        />
      )}
    </div>
  );
}
