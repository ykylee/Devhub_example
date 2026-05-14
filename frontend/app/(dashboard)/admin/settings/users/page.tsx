"use client";

import { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { Search, Filter } from "lucide-react";
import { identityService, OrgMember } from "@/lib/services/identity.service";
import { MemberTable } from "@/components/organization/MemberTable";
import { defaultRoles, Role } from "@/lib/services/rbac.types";
import { rbacService } from "@/lib/services/rbac.service";

import { useToast } from "@/components/ui/Toast";

export default function AdminSettingsUsersPage() {
  const [members, setMembers] = useState<OrgMember[]>([]);
  const [roles, setRoles] = useState<Role[]>(defaultRoles);
  const [isLoading, setIsLoading] = useState(true);
  const [query, setQuery] = useState("");
  const { toast } = useToast();

  const filteredMembers = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return members;
    return members.filter((m) =>
      m.name.toLowerCase().includes(q) ||
      m.email.toLowerCase().includes(q) ||
      m.role.toLowerCase().includes(q)
    );
  }, [members, query]);

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
      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex items-center gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search by name, email, or role..."
            aria-label="Search users"
            className="w-full glass border-border rounded-2xl pl-12 pr-6 py-3.5 text-sm text-foreground text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-accent/30 transition-all"
          />
        </div>
        <button
          type="button"
          disabled
          title="Advanced filters coming soon"
          aria-label="Advanced filters coming soon"
          className="glass border-border p-3.5 rounded-2xl text-muted-foreground/40 cursor-not-allowed transition-all"
        >
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
          members={filteredMembers}
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
