"use client";

import { motion } from "framer-motion";
import { Users, Building2, Layers, Shield, ArrowRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { OrgNode } from "@/lib/services/identity.service";

interface OrgUnitGridProps {
  nodes: OrgNode[];
  unitMembers: Record<string, string[]>;
  onManage: (id: string) => void;
}

export function OrgUnitGrid({ nodes, unitMembers, onManage }: OrgUnitGridProps) {
  const typeIcons = {
    division: Building2,
    team: Users,
    group: Layers,
    part: Shield,
    company: Building2,
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-xl font-black text-white uppercase tracking-tight">Organization <span className="text-accent">Units</span></h3>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {nodes.map((node, index) => {
          const type = node.data.type as keyof typeof typeIcons;
          const Icon = typeIcons[type] || Building2;
          const currentMemberCount = unitMembers[node.id]?.length || 0;

          return (
            <motion.div
              key={node.id}
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: index * 0.05 }}
              className={cn(
                "glass-card p-6 hover:border-accent/30 transition-all duration-500 group relative overflow-hidden",
                type === 'division' ? "border-primary/20" : 
                type === 'team' ? "border-accent/20" : "border-white/5"
              )}
            >
              {/* Subtle background glow based on type */}
              <div className={cn(
                "absolute -top-10 -right-10 w-32 h-32 blur-3xl opacity-20 pointer-events-none transition-opacity group-hover:opacity-40",
                type === 'division' ? "bg-primary" : 
                type === 'team' ? "bg-accent" : "bg-white"
              )} />

              <div className="flex justify-between items-start mb-4 relative z-10">
                <div className={cn(
                  "p-3 rounded-2xl border group-hover:scale-110 transition-transform",
                  type === 'division' ? "bg-primary/10 border-primary/20 text-primary" :
                  type === 'team' ? "bg-accent/10 border-accent/20 text-accent" :
                  "bg-white/5 border-white/10 text-white/50"
                )}>
                  <Icon className="w-6 h-6" />
                </div>
                <div className="text-right">
                  <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Type</p>
                  <p className={cn(
                    "text-sm font-black uppercase tracking-widest",
                    type === 'division' ? "text-primary" :
                    type === 'team' ? "text-accent" : "text-white/80"
                  )}>{type}</p>
                </div>
              </div>

              <div className="relative z-10">
                <h4 className="text-lg font-black text-white mb-2 truncate" title={node.data.label}>
                  {node.data.label}
                </h4>
                {node.data.leader_id && (
                  <div className="flex items-center gap-1.5 text-[10px] font-bold text-orange-400 mb-6 bg-orange-400/10 w-fit px-2 py-0.5 rounded-full border border-orange-400/20">
                    Leader: {node.data.leader_id}
                  </div>
                )}
                {!node.data.leader_id && (
                  <div className="h-[22px] mb-6" /> // Placeholder to maintain height
                )}
              </div>

              <div className="flex items-center justify-between pt-4 border-t border-white/5 relative z-10">
                <div className="flex items-center gap-2">
                  <div className="flex -space-x-2">
                    {Array.from({ length: Math.min(currentMemberCount, 3) }).map((_, i) => (
                      <div key={i} className="w-6 h-6 rounded-full border-2 border-[#030014] bg-white/20" />
                    ))}
                    {currentMemberCount === 0 && (
                      <div className="w-6 h-6 rounded-full border-2 border-[#030014] bg-white/5 flex items-center justify-center">
                        <Users className="w-3 h-3 text-white/20" />
                      </div>
                    )}
                  </div>
                  <span className="text-[10px] font-bold text-muted-foreground">
                    {currentMemberCount} members
                  </span>
                </div>
                
                <button 
                  onClick={() => onManage(node.id)}
                  className="text-[10px] font-black uppercase tracking-widest text-accent hover:text-white transition-colors flex items-center gap-1 group/btn px-3 py-1.5 rounded-lg bg-accent/10 hover:bg-accent/20"
                >
                  Manage <ArrowRight className="w-3 h-3 group-hover/btn:translate-x-1 transition-transform" />
                </button>
              </div>
            </motion.div>
          );
        })}
      </div>
    </div>
  );
}
