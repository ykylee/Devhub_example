"use client";

import { identityService, OrgMember } from "@/lib/services/identity.service";
import { motion, AnimatePresence } from "framer-motion";
import { UserPlus, Mail, Shield, ArrowRightLeft, Crown, Bot, Edit3, Trash2 } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { cn } from "@/lib/utils";
import { useState } from "react";
import { useToast } from "@/components/ui/Toast";
import { UserCreationModal } from "./UserCreationModal";
import { UserEditModal } from "./UserEditModal";
import { DestructiveConfirmModal } from "@/components/ui/DestructiveConfirmModal";
import { Role } from "@/lib/services/rbac.types";

interface MemberTableProps {
  members: OrgMember[];
  unitLeaderIds?: string[];
  unitNames?: Record<string, string>;
  roles: Role[];
  onUpdateMemberRole: (memberId: string, newRoleName: string) => void;
  onMemberCreated?: (user: OrgMember) => void;
  onMemberUpdated?: (user: OrgMember) => void;
  onMemberDeleted?: (userId: string) => void;
}

export function MemberTable({ members, unitLeaderIds = [], unitNames = {}, roles, onUpdateMemberRole, onMemberCreated, onMemberUpdated, onMemberDeleted }: MemberTableProps) {
  const lookupUnitName = (id: string | null | undefined): string => {
    if (!id) return "-";
    return unitNames[id] || id;
  };
  const { toast } = useToast();
  const unitLeaderIdSet = new Set(unitLeaderIds);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingMember, setEditingMember] = useState<OrgMember | null>(null);
  const [deletingMember, setDeletingMember] = useState<OrgMember | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleConfirmDelete = async () => {
    if (!deletingMember) return;
    const target = deletingMember;
    setIsDeleting(true);
    try {
      await identityService.deleteUser(target.id);
      onMemberDeleted?.(target.id);
      toast(`Member '${target.name}' deleted`, "success");
      setDeletingMember(null);
    } catch (error) {
      console.error("[MemberTable] delete failed:", error);
      toast("Failed to delete member", "error");
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">Organization <span className="text-primary">Members</span></h3>
        <button 
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 rounded-xl text-xs font-bold hover:bg-primary/90 transition-all shadow-lg"
        >
          <UserPlus className="w-4 h-4" /> Invite Member
        </button>
      </div>

      <AnimatePresence>
        {showCreateModal && (
          <UserCreationModal 
            roles={roles}
            onClose={() => setShowCreateModal(false)}
            onCreated={(user) => {
              onMemberCreated?.(user);
              toast("Member created successfully", "success");
            }}
          />
        )}
      </AnimatePresence>

      <div className="overflow-x-auto overflow-y-visible">
        <table className="w-full text-left border-separate border-spacing-y-3">
          <thead>
            <tr className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-4">
              <th className="px-6 py-2">User</th>
              <th className="px-6 py-2">Role</th>
              <th className="px-6 py-2">Department</th>
              <th className="px-6 py-2">Status</th>
              <th className="px-6 py-2 text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {members.map((member, index) => {
              const isLeader = member.appointments.some(a => a.role === 'leader') || unitLeaderIdSet.has(member.id);
              const isDualLeader = member.appointments.filter(a => a.role === 'leader').length > 1;
              const displayDeptId =
                member.current_dept_id ||
                member.appointments[0]?.dept_id ||
                member.primary_dept_id ||
                "";
              const displayDept = lookupUnitName(displayDeptId);

              return (
                <motion.tr
                  key={member.id}
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: index * 0.05 }}
                  className="glass group hover:bg-muted/30 transition-all duration-300 rounded-2xl"
                >
                  <td className="px-6 py-4 rounded-l-2xl">
                    <div className="flex items-center gap-3">
                      <div className="relative">
                        <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary/20 to-accent/20 flex items-center justify-center border border-border">
                          <span className="font-black text-foreground">{member.name.charAt(0)}</span>
                        </div>
                        {isLeader && (
                          <div className="absolute -top-1 -right-1 bg-accent rounded-full p-0.5 border border-background">
                            <Crown className="w-2.5 h-2.5 text-primary-foreground" />
                          </div>
                        )}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <p className="text-sm font-bold text-foreground">{member.name}</p>
                          {isDualLeader && (
                            <Badge variant="warning" className="text-[8px] py-0 px-1 uppercase">Dual</Badge>
                          )}
                        </div>
                        <p className="text-[10px] text-muted-foreground flex items-center gap-1">
                          <Mail className="w-3 h-3" /> {member.email}
                        </p>
                        {member.type === 'system' && (
                          <div className="mt-1 inline-flex items-center gap-1 bg-accent/10 border border-accent/20 px-1.5 py-0.5 rounded text-[8px] font-black text-accent uppercase tracking-tighter">
                            <Bot className="w-2.5 h-2.5" /> System / AI
                          </div>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <Shield className={cn("w-3 h-3", 
                        member.role === 'System Admin' ? "text-accent" : 
                        member.role === 'Manager' ? "text-success" : "text-info"
                      )} />
                      <select
                        value={member.role}
                        onChange={(e) => onUpdateMemberRole(member.id, e.target.value)}
                        className="themed-select !py-1 !text-[11px] !rounded-lg"
                      >
                        {roles.map(r => (
                          <option key={r.id} value={r.name} className="bg-popover text-popover-foreground">{r.name}</option>
                        ))}
                      </select>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex flex-col gap-1">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-bold text-foreground/90">{displayDept}</span>
                        {member.is_seconded && (
                          <div className="flex items-center gap-1 bg-info/10 border border-info/20 px-1.5 py-0.5 rounded text-[8px] font-black text-info uppercase">
                            <ArrowRightLeft className="w-2 h-2" /> Seconded
                          </div>
                        )}
                      </div>
                      {member.is_seconded && member.primary_dept_id && (
                        <p className="text-[9px] text-muted-foreground italic">Original: {lookupUnitName(member.primary_dept_id)}</p>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <Badge variant={member.status === 'active' ? 'success' : member.status === 'pending' ? 'warning' : 'danger'} dot>
                      {member.status}
                    </Badge>
                  </td>
                  <td className="px-6 py-4 text-right rounded-r-2xl">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        type="button"
                        onClick={() => setEditingMember(member)}
                        className="p-1.5 rounded-lg bg-muted/30 hover:bg-muted/50 text-muted-foreground hover:text-foreground transition-colors"
                        title="Edit member"
                        aria-label={`Edit ${member.name}`}
                      >
                        <Edit3 className="w-3.5 h-3.5" />
                      </button>
                      <button
                        type="button"
                        onClick={() => setDeletingMember(member)}
                        className="p-1.5 rounded-lg bg-destructive/10 hover:bg-destructive/20 text-destructive transition-colors"
                        title="Delete member"
                        aria-label={`Delete ${member.name}`}
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </td>
                </motion.tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <AnimatePresence>
        {editingMember && (
          <UserEditModal
            initial={editingMember}
            roles={roles}
            onClose={() => setEditingMember(null)}
            onUpdated={(updated) => {
              onMemberUpdated?.(updated);
              toast(`Member '${updated.name}' updated`, "success");
            }}
          />
        )}
      </AnimatePresence>

      <DestructiveConfirmModal
        isOpen={!!deletingMember}
        onClose={() => { if (!isDeleting) setDeletingMember(null); }}
        onConfirm={handleConfirmDelete}
        title="Delete Member"
        description={deletingMember ? `'${deletingMember.name}' (${deletingMember.id}) 사용자를 삭제합니다. 되돌릴 수 없습니다.` : ""}
        confirmText={isDeleting ? "Deleting..." : "Delete Member"}
      />
    </div>
  );
}
