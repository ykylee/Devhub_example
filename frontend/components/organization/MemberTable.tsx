"use client";

import { OrgMember } from "@/lib/services/identity.service";
import { motion, AnimatePresence } from "framer-motion";
import { UserPlus, Mail, Shield, ArrowRightLeft, Crown, Key, UserX, KeyRound, Bot, Copy, Check } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { cn } from "@/lib/utils";
import { useState, useEffect } from "react";
import { useStore } from "@/lib/store";
import { accountService } from "@/lib/services/account.service";
import { useToast } from "@/components/ui/Toast";
import { Modal } from "@/components/ui/Modal";
import { ActionMenu } from "@/components/ui/ActionMenu";

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
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [adminActionResult, setAdminActionResult] = useState<{
    title: string;
    details: { label: string; value: string }[];
  } | null>(null);
  const [copied, setCopied] = useState<string | null>(null);

  const handleCopy = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    setCopied(label);
    setTimeout(() => setCopied(null), 2000);
  };

  const handleAdminAction = async (action: 'issue' | 'reset' | 'disable', member: OrgMember) => {
    try {
      if (action === 'issue') {
        const loginId = member.email.split('@')[0];
        const { tempPassword } = await accountService.issueAccount(member.id, loginId, true);
        setAdminActionResult({
          title: "Account Issued Successfully",
          details: [
            { label: "Login ID", value: loginId },
            { label: "Temporary Password", value: tempPassword }
          ]
        });
      } else if (action === 'reset') {
        const { tempPassword } = await accountService.forceResetPassword(member.id);
        setAdminActionResult({
          title: "Password Reset Successful",
          details: [
            { label: "New Temporary Password", value: tempPassword }
          ]
        });
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
        {adminActionResult && (
          <Modal
            isOpen={!!adminActionResult}
            onClose={() => setAdminActionResult(null)}
            title={adminActionResult.title}
            size="sm"
          >
            <div className="space-y-4">
              <p className="text-xs text-muted-foreground">Please share these credentials with the user securely. They will only be shown once.</p>
              <div className="space-y-3">
                {adminActionResult.details.map((detail, idx) => (
                  <div key={idx} className="glass-card p-4 space-y-1 relative group">
                    <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">{detail.label}</p>
                    <div className="flex items-center justify-between">
                      <p className="text-sm font-mono font-bold text-primary-foreground break-all">{detail.value}</p>
                      <button 
                        onClick={() => handleCopy(detail.value, detail.label)}
                        className="p-1.5 rounded-lg hover:bg-muted/40 text-muted-foreground hover:text-primary-foreground transition-all ml-2"
                      >
                        {copied === detail.label ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
              <button 
                onClick={() => setAdminActionResult(null)}
                className="w-full py-3 bg-primary text-primary-foreground rounded-xl text-xs font-bold hover:bg-primary/90 transition-all shadow-lg mt-4"
              >
                Close & Confirm
              </button>
            </div>
          </Modal>
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
                  className="glass group hover:bg-muted/30 transition-all duration-300 rounded-2xl"
                >
                  <td className="px-6 py-4 rounded-l-2xl">
                    <div className="flex items-center gap-3">
                      <div className="relative">
                        <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary/20 to-accent/20 flex items-center justify-center border border-border">
                          <span className="font-black text-foreground">{member.name.charAt(0)}</span>
                        </div>
                        {isLeader && (
                          <div className="absolute -top-1 -right-1 bg-orange-500 rounded-full p-0.5 border border-background">
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
                        member.role === 'System Admin' ? "text-orange-400" : 
                        member.role === 'Manager' ? "text-emerald-400" : "text-blue-400"
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
                        <span className="text-xs font-bold text-foreground/90 dark:text-primary-foreground/90">{member.current_dept_id}</span>
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
                    {currentUserRole === "System Admin" && (
                      <ActionMenu
                        title="User Actions"
                        items={[
                          {
                            key: "issue",
                            label: "Issue Account",
                            onClick: () => handleAdminAction("issue", member),
                            icon: <Key className="w-3.5 h-3.5 text-accent" />,
                          },
                          {
                            key: "reset",
                            label: "Force Reset Password",
                            onClick: () => handleAdminAction("reset", member),
                            icon: <KeyRound className="w-3.5 h-3.5 text-orange-400" />,
                          },
                          {
                            key: "revoke",
                            label: "Revoke Account",
                            onClick: () => handleAdminAction("disable", member),
                            icon: <UserX className="w-3.5 h-3.5" />,
                            tone: "danger",
                          },
                        ]}
                      />
                    )}
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
