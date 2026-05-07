"use client";

import { motion } from "framer-motion";
import { Users, Mail, Calendar, Building2, ArrowRightLeft } from "lucide-react";
import { OrgMember } from "@/lib/services/identity.service";
import { cn } from "@/lib/utils";

interface MemberTableProps {
  members: OrgMember[];
}

const roleStyles: Record<OrgMember["role"], string> = {
  Developer: "bg-white/5 border-white/10 text-white/80",
  Manager: "bg-blue-500/10 border-blue-500/20 text-blue-300",
  "System Admin": "bg-accent/15 border-accent/30 text-accent",
};

const statusStyles: Record<OrgMember["status"], string> = {
  active: "bg-emerald-500/10 border-emerald-500/20 text-emerald-300",
  pending: "bg-yellow-500/10 border-yellow-500/20 text-yellow-300",
  deactivated: "bg-red-500/10 border-red-500/20 text-red-300",
};

const statusDot: Record<OrgMember["status"], string> = {
  active: "bg-emerald-400",
  pending: "bg-yellow-400",
  deactivated: "bg-red-400",
};

export function MemberTable({ members }: MemberTableProps) {
  if (members.length === 0) {
    return (
      <div className="glass border border-white/10 rounded-3xl py-24 flex flex-col items-center justify-center text-center gap-3">
        <div className="p-4 rounded-2xl bg-white/5 border border-white/10">
          <Users className="w-8 h-8 text-muted-foreground" />
        </div>
        <p className="text-xs font-black uppercase tracking-[0.3em] text-white/60">
          No personnel found
        </p>
        <p className="text-[11px] text-muted-foreground/70">
          Adjust filters or invite members to populate the roster.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-xl font-black text-white uppercase tracking-tight">
          Personnel <span className="text-accent">Roster</span>
        </h3>
        <span className="text-[10px] font-black uppercase tracking-[0.3em] text-muted-foreground">
          {members.length} {members.length === 1 ? "member" : "members"}
        </span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
        {members.map((member, index) => {
          const initials = member.name.substring(0, 2).toUpperCase();
          return (
            <motion.div
              key={member.id}
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.04, duration: 0.3 }}
              className="glass-card p-5 hover:border-accent/30 transition-all duration-500 group relative overflow-hidden"
            >
              <div className="absolute -top-12 -right-12 w-32 h-32 blur-3xl opacity-10 pointer-events-none group-hover:opacity-25 transition-opacity bg-accent" />

              <div className="flex items-start justify-between gap-3 relative z-10">
                <div className="flex items-center gap-3 min-w-0">
                  <div className="w-11 h-11 rounded-2xl bg-gradient-to-br from-accent/30 to-primary/20 border border-white/10 flex items-center justify-center text-sm font-black text-white shrink-0">
                    {initials}
                  </div>
                  <div className="min-w-0">
                    <p className="text-sm font-black text-white truncate" title={member.name}>
                      {member.name}
                    </p>
                    <p className="text-[11px] text-muted-foreground flex items-center gap-1 truncate">
                      <Mail className="w-3 h-3 shrink-0" />
                      <span className="truncate">{member.email}</span>
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-1.5 shrink-0">
                  <span
                    className={cn(
                      "w-1.5 h-1.5 rounded-full animate-pulse",
                      statusDot[member.status]
                    )}
                  />
                  <span
                    className={cn(
                      "px-2 py-0.5 rounded-full border text-[9px] font-black uppercase tracking-widest",
                      statusStyles[member.status]
                    )}
                  >
                    {member.status}
                  </span>
                </div>
              </div>

              <div className="mt-4 flex flex-wrap gap-2 relative z-10">
                <span
                  className={cn(
                    "px-2.5 py-1 rounded-lg border text-[10px] font-black uppercase tracking-widest",
                    roleStyles[member.role]
                  )}
                >
                  {member.role}
                </span>
                {member.is_seconded && (
                  <span className="px-2.5 py-1 rounded-lg border border-orange-400/30 bg-orange-400/10 text-orange-300 text-[10px] font-black uppercase tracking-widest flex items-center gap-1">
                    <ArrowRightLeft className="w-3 h-3" /> Seconded
                  </span>
                )}
              </div>

              <div className="mt-5 pt-4 border-t border-white/5 grid grid-cols-2 gap-3 relative z-10">
                <div>
                  <p className="text-[9px] font-black uppercase tracking-[0.25em] text-muted-foreground mb-1">
                    Primary
                  </p>
                  <p className="text-[11px] font-bold text-white/80 flex items-center gap-1.5 truncate">
                    <Building2 className="w-3 h-3 text-accent shrink-0" />
                    <span className="truncate">{member.primary_dept_id}</span>
                  </p>
                </div>
                <div>
                  <p className="text-[9px] font-black uppercase tracking-[0.25em] text-muted-foreground mb-1">
                    Current
                  </p>
                  <p
                    className={cn(
                      "text-[11px] font-bold flex items-center gap-1.5 truncate",
                      member.is_seconded ? "text-orange-300" : "text-white/80"
                    )}
                  >
                    <Building2 className="w-3 h-3 shrink-0" />
                    <span className="truncate">{member.current_dept_id}</span>
                  </p>
                </div>
              </div>

              <div className="mt-3 flex items-center gap-1.5 text-[10px] text-muted-foreground/80 relative z-10">
                <Calendar className="w-3 h-3" />
                <span className="font-bold">Joined {member.joined_at}</span>
              </div>
            </motion.div>
          );
        })}
      </div>
    </div>
  );
}
