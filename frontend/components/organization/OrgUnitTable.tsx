"use client";

import { useState } from "react";
import { Users, Shield, Settings, UserPlus, Trash2, Search, Filter } from "lucide-react";
import { OrgNode } from "@/lib/services/identity.service";
import { cn } from "@/lib/utils";

interface OrgUnitTableProps {
  nodes: OrgNode[];
  unitMembers: Record<string, string[]>;
  onManage: (unitId: string) => void;
  onEdit: (unitId: string) => void;
  onDelete: (unitId: string) => void;
}

export function OrgUnitTable({ nodes, unitMembers, onManage, onEdit, onDelete }: OrgUnitTableProps) {
  const [search, setSearch] = useState("");

  const filteredNodes = nodes.filter(
    (node) =>
      node.data.label.toLowerCase().includes(search.toLowerCase()) ||
      node.data.type.toLowerCase().includes(search.toLowerCase()),
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search units by name or type..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-white/5 border border-white/10 rounded-xl pl-10 pr-4 py-2 text-xs text-white focus:outline-none focus:border-accent/50 transition-all placeholder:text-muted-foreground/50"
          />
        </div>
        <div className="flex items-center gap-2 text-[10px] font-bold text-muted-foreground uppercase tracking-widest bg-white/5 px-3 py-2 rounded-lg border border-white/10">
          <Filter className="w-3 h-3" />
          {filteredNodes.length} units found
        </div>
      </div>

      <div className="glass border-white/10 rounded-3xl overflow-hidden">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-white/10 bg-white/5">
              <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground">Unit Name</th>
              <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground">Type</th>
              <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground">Leader</th>
              <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground text-center">Members</th>
              <th className="px-6 py-4 text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/5">
            {filteredNodes.map((node) => {
              const memberCount = unitMembers[node.id]?.length || 0;
              return (
                <tr key={node.id} className="group hover:bg-white/[0.02] transition-colors">
                  <td className="px-6 py-5">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-primary/20 to-primary/5 flex items-center justify-center border border-primary/10">
                        <Shield className="w-4 h-4 text-primary" />
                      </div>
                      <div>
                        <p className="text-sm font-bold text-white group-hover:text-primary transition-colors">{node.data.label}</p>
                        <p className="text-[10px] text-muted-foreground font-medium uppercase tracking-wider">{node.id}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-5">
                    <span className="px-2.5 py-1 rounded-lg bg-white/5 border border-white/10 text-[9px] font-black uppercase tracking-widest text-white/60 group-hover:text-white transition-colors">
                      {node.data.type}
                    </span>
                  </td>
                  <td className="px-6 py-5">
                    {node.data.leader_id ? (
                      <div className="flex items-center gap-2">
                        <div className="w-6 h-6 rounded-full bg-accent/20 border border-accent/20 flex items-center justify-center text-[10px] font-bold text-accent">
                          {node.data.leader_id.substring(0, 1).toUpperCase()}
                        </div>
                        <span className="text-xs font-bold text-white/80">{node.data.leader_id}</span>
                      </div>
                    ) : (
                      <span className="text-[10px] font-bold text-muted-foreground/30 italic">No Leader</span>
                    )}
                  </td>
                  <td className="px-6 py-5">
                    <div className="flex items-center justify-center gap-2">
                      <Users className="w-3.5 h-3.5 text-muted-foreground" />
                      <span className="text-sm font-black text-white">{memberCount}</span>
                    </div>
                  </td>
                  <td className="px-6 py-5">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => onManage(node.id)}
                        className="p-2 rounded-xl bg-primary/10 text-primary border border-primary/20 hover:bg-primary/20 transition-all opacity-0 group-hover:opacity-100 flex items-center gap-2"
                        title="Manage Members"
                      >
                        <UserPlus className="w-3.5 h-3.5" />
                        <span className="text-[9px] font-black uppercase">Members</span>
                      </button>
                      <button
                        onClick={() => onEdit(node.id)}
                        className="p-2 rounded-xl bg-white/5 text-white/60 border border-white/10 hover:bg-white/10 hover:text-white transition-all opacity-0 group-hover:opacity-100"
                        title="Edit Unit"
                      >
                        <Settings className="w-3.5 h-3.5" />
                      </button>
                      <button
                        onClick={() => onDelete(node.id)}
                        className="p-2 rounded-xl bg-red-500/10 text-red-400 border border-red-500/20 hover:bg-red-500/20 transition-all opacity-0 group-hover:opacity-100"
                        title="Delete Unit"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })}
            {filteredNodes.length === 0 && (
              <tr>
                <td colSpan={5} className="px-6 py-20 text-center">
                  <div className="flex flex-col items-center gap-3 opacity-30">
                    <Search className="w-10 h-10" />
                    <p className="text-xs font-black uppercase tracking-widest">No matching units found</p>
                  </div>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
