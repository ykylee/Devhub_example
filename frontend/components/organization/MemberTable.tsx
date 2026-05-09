"use client";

import { OrgMember } from "@/lib/services/identity.service";
import { motion, AnimatePresence } from "framer-motion";
import { MoreHorizontal, UserPlus, Mail, Shield, ArrowRightLeft, Crown, Key, UserX, KeyRound } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { cn } from "@/lib/utils";
import { useState } from "react";
import { useStore } from "@/lib/store";
import { accountService } from "@/lib/services/account.service";
import { useToast } from "@/components/ui/Toast";

import { Role } from "@/lib/services/rbac.types";
import { UserCreationModal } from "./UserCreationModal";

interface MemberTableProps {
  members: OrgMember[];
  roles: Role[];
  onUpdateMemberRole: (memberId: string, newRoleName: string) => void;
  onMemberCreated?: (user: OrgMember) => void;
}

export function MemberTable({ members, roles, onUpdateMemberRole, onMemberCreated }: MemberTableProps) {
  const { role: currentUserRole } = useStore();
  const { toast } = useToast();
  const [openActionId, setOpenActionId] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);

  const handleAdminAction = async (action: 'issue' | 'reset' | 'disable', member: OrgMember) => {
    setOpenActionId(null);
    try {
      if (action === 'issue') {
        const { tempPassword } = await accountService.issueAccount(member.id, member.email.split('@')[0], true);
        alert(`Account Issued for ${member.name}.\nLogin ID: ${member.email.split('@')[0]}\nTemp Password: ${tempPassword}\n\nPlease communicate this securely.`);
      } else if (action === 'reset') {
        const { tempPassword } = await accountService.forceResetPassword(member.id);
        alert(`Password Reset for ${member.name}.\nTemp Password: ${tempPassword}\n\nPlease communicate this securely.`);
      } else if (action === 'disable') {
        if (confirm(`Are you sure you want to disable ${member.name}'s account?`)) {
          await accountService.disableAccount(member.id, "Admin requested");
          toast("Account disabled", "success");
        }
      }
    } catch {
      toast("Failed to perform action", "error");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-xl font-black text-foreground dark:text-white uppercase tracking-tight">Organization <span className="text-primary">Members</span></h3>
        <button 
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 bg-primary text-white px-4 py-2 rounded-xl text-xs font-bold hover:bg-primary/90 transition-all shadow-lg"
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

      <div className="overflow-x-auto">
        <table className="w-full text-left border-separate border-spacing-y-3">
          <thead>
            <tr className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-4">
              <th className="px-6 py-2">User</th>
              <th className="px-6 py-2">Role</th>
              <th className="px-6 py-2">Department</th>
              <th className="px-6 py-2">Status</th>
              <th className="px-6 py-2 text-right">Action</th>
            </tr>
          </thead>
          <tbody>
            {members.map((member, index) => {
              const isLeader = member.appointments.some(a => a.role === 'leader');
              const isDualLeader = member.appointments.filter(a => a.role === 'leader').length > 1;

              return (
                <motion.tr
                  key={member.id}
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: index * 0.05 }}
                  className="glass group hover:bg-white/5 transition-all duration-300 rounded-2xl overflow-hidden"
                >
                  <td className="px-6 py-4 rounded-l-2xl">
                    <div className="flex items-center gap-3">
                      <div className="relative">
                        <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary/20 to-accent/20 flex items-center justify-center border border-white/10">
                          <span className="font-black text-foreground dark:text-white">{member.name.charAt(0)}</span>
                        </div>
                        {isLeader && (
                          <div className="absolute -top-1 -right-1 bg-orange-500 rounded-full p-0.5 border border-[#030014]">
                            <Crown className="w-2.5 h-2.5 text-white" />
                          </div>
                        )}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <p className="text-sm font-bold text-foreground dark:text-white">{member.name}</p>
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
                        member.role === 'System Admin' ? "text-orange-400" : 
                        member.role === 'Manager' ? "text-emerald-400" : "text-blue-400"
                      )} />
                      <select
                        value={member.role}
                        onChange={(e) => onUpdateMemberRole(member.id, e.target.value)}
                        className="bg-black/20 border border-white/10 rounded-lg text-xs font-medium text-foreground/80 dark:text-white/80 focus:ring-1 focus:ring-primary/50 focus:outline-none p-1 transition-colors hover:border-white/20"
                      >
                        {roles.map(r => (
                          <option key={r.id} value={r.name} className="bg-slate-900">{r.name}</option>
                        ))}
                      </select>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex flex-col gap-1">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-bold text-foreground/90 dark:text-white/90">{member.current_dept_id}</span>
                        {member.is_seconded && (
                          <div className="flex items-center gap-1 bg-blue-500/10 border border-blue-500/20 px-1.5 py-0.5 rounded text-[8px] font-black text-blue-400 uppercase">
                            <ArrowRightLeft className="w-2 h-2" /> Seconded
                          </div>
                        )}
                      </div>
                      {member.is_seconded && (
                        <p className="text-[9px] text-muted-foreground italic">Original: {member.primary_dept_id}</p>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <Badge variant={member.status === 'active' ? 'success' : member.status === 'pending' ? 'warning' : 'danger'} dot>
                      {member.status}
                    </Badge>
                  </td>
                  <td className="px-6 py-4 text-right rounded-r-2xl relative">
                    <button 
                      onClick={() => setOpenActionId(openActionId === member.id ? null : member.id)}
                      className="p-2 hover:bg-white/10 rounded-lg transition-colors text-muted-foreground hover:text-white"
                    >
                      <MoreHorizontal className="w-5 h-5" />
                    </button>

                    <AnimatePresence>
                      {openActionId === member.id && (
                        <>
                          <div 
                            className="fixed inset-0 z-40" 
                            onClick={() => setOpenActionId(null)}
                          />
                          <motion.div
                            initial={{ opacity: 0, scale: 0.95, y: -10 }}
                            animate={{ opacity: 1, scale: 1, y: 0 }}
                            exit={{ opacity: 0, scale: 0.95, y: -10 }}
                            className="absolute right-8 top-12 z-50 w-48 glass bg-[#030014]/90 border border-white/10 rounded-xl overflow-hidden shadow-2xl py-1"
                          >
                            <div className="px-3 py-2 border-b border-white/10">
                              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest text-left">Actions</p>
                            </div>
                            
                            {currentUserRole === "System Admin" && (
                              <div className="py-1">
                                <button 
                                  onClick={() => handleAdminAction('issue', member)}
                                  className="w-full flex items-center gap-2 px-3 py-2 text-xs text-foreground dark:text-white hover:bg-white/5 transition-colors text-left"
                                >
                                  <Key className="w-3.5 h-3.5 text-accent" /> Issue Account
                                </button>
                                <button 
                                  onClick={() => handleAdminAction('reset', member)}
                                  className="w-full flex items-center gap-2 px-3 py-2 text-xs text-foreground dark:text-white hover:bg-white/5 transition-colors text-left"
                                >
                                  <KeyRound className="w-3.5 h-3.5 text-orange-400" /> Force Reset Password
                                </button>
                                <button 
                                  onClick={() => handleAdminAction('disable', member)}
                                  className="w-full flex items-center gap-2 px-3 py-2 text-xs text-red-400 hover:bg-red-400/10 transition-colors text-left"
                                >
                                  <UserX className="w-3.5 h-3.5" /> Revoke Account
                                </button>
                              </div>
                            )}
                          </motion.div>
                        </>
                      )}
                    </AnimatePresence>
                  </td>
                </motion.tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
