"use client";

import React, { memo, useState } from 'react';
import { Handle, Position, Node, NodeProps } from '@xyflow/react';
import { Plus, Minus, Edit3, Crown, Check, X, Building2, Users, Layers, Shield } from 'lucide-react';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';

type OrgNodeData = {
  [key: string]: unknown;
  label?: string;
  type?: string;
  leader_id?: string;
  direct_count?: number;
  total_count?: number;
  allowedTypes?: string[];
  isInitialEditing?: boolean;
  onAddChild?: (id: string) => void;
  onDelete?: (id: string) => void;
  onUpdate?: (id: string, data: Partial<OrgNodeData>) => void;
  onToggleExpand?: (id: string) => void;
  isExpanded?: boolean;
  hasChildren?: boolean;
};

export const OrgNode = memo(({ id, data, selected }: NodeProps<Node<OrgNodeData>>) => {
  const [isEditing, setIsEditing] = useState(data.isInitialEditing || false);
  const [editedName, setEditedName] = useState(data.label || "");
  const [editedLeader, setEditedLeader] = useState(data.leader_id || "");
  const [editedType, setEditedType] = useState(data.type || "team");
  
  const typeIcons = {
    division: Building2,
    team: Users,
    group: Layers,
    part: Shield,
    company: Building2,
  };

  const Icon = typeIcons[editedType as keyof typeof typeIcons] || Building2;

  const handleSave = (e: React.MouseEvent) => {
    e.stopPropagation();
    data.onUpdate?.(id, { 
      label: editedName, 
      leader_id: editedLeader, 
      type: editedType,
      isInitialEditing: false 
    });
    setIsEditing(false);
  };

  const handleCancel = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (data.isInitialEditing) {
      data.onDelete?.(id);
    } else {
      setEditedName(data.label || "");
      setEditedLeader(data.leader_id || "");
      setEditedType(data.type || "team");
      setIsEditing(false);
    }
  };

  return (
    <div 
      className={cn(
        "relative group",
        isEditing ? "z-50" : "z-10"
      )}
    >
      <Handle type="target" position={Position.Top} className="!bg-white/20 !w-3 !h-3 !border-white/40" />
      
      <motion.div
        initial={false}
        animate={{
          width: isEditing ? 360 : 240,
          height: isEditing ? 280 : 140,
        }}
        className={cn(
          "glass rounded-2xl border p-4 flex flex-col justify-between shadow-2xl",
          selected ? "border-blue-500/50 ring-2 ring-blue-500/20" : "border-white/10 group-hover:border-white/30"
        )}
      >
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3 w-full">
            <div className={cn(
              "p-2 rounded-xl border shrink-0",
              (isEditing ? editedType : data.type) === 'division' ? "bg-primary/20 border-primary/30 text-primary" :
              (isEditing ? editedType : data.type) === 'team' ? "bg-accent/20 border-accent/30 text-accent" :
              "bg-white/5 border-white/10 text-white/50"
            )}>
              <Icon className="w-4 h-4" />
            </div>
            
            {!isEditing ? (
              <div className="flex flex-col truncate flex-1">
                <span className="text-[10px] font-black text-white/40 uppercase tracking-widest">{data.type}</span>
                <h4 className="text-sm font-bold text-white truncate">{data.label}</h4>
              </div>
            ) : (
              <div className="flex flex-col gap-3 flex-1">
                <div className="flex gap-2">
                   <div className="flex-1">
                    <span className="text-[10px] font-black text-primary uppercase tracking-widest">Type</span>
                    <select 
                      value={editedType}
                      onChange={(e) => setEditedType(e.target.value)}
                      className="themed-select !py-1 !text-[11px] w-full mt-1"
                    >
                      {(data.allowedTypes || ['division', 'team', 'group', 'part']).map((t) => (
                        <option key={t} value={t} className="bg-popover text-popover-foreground">
                          {t.charAt(0).toUpperCase() + t.slice(1)}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="flex-[2]">
                    <span className="text-[10px] font-black text-primary uppercase tracking-widest">Name</span>
                    <input 
                      autoFocus
                      value={editedName}
                      onChange={(e) => setEditedName(e.target.value)}
                      className="bg-white/5 border border-white/10 rounded-lg px-2 py-1.5 text-sm font-bold text-white focus:outline-none focus:border-primary/50 w-full mt-1"
                      onClick={(e) => e.stopPropagation()}
                    />
                  </div>
                </div>
                <div>
                  <span className="text-[10px] font-black text-orange-400 uppercase tracking-widest">Leader ID</span>
                  <input 
                    value={editedLeader}
                    onChange={(e) => setEditedLeader(e.target.value)}
                    placeholder="Enter Leader ID (e.g. u1)"
                    className="bg-white/5 border border-white/10 rounded-lg px-2 py-1.5 text-xs font-bold text-white focus:outline-none focus:border-orange-500/50 w-full mt-1"
                    onClick={(e) => e.stopPropagation()}
                  />
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Member Counts & Leader Info */}
        {!isEditing && (
          <div className="flex flex-col gap-2 mt-2">
            <div className="flex items-center justify-between border-t border-white/5 pt-2">
               {data.leader_id ? (
                <div className="flex items-center gap-1.5 text-[9px] font-black text-orange-400 bg-orange-400/10 px-2 py-1 rounded-full border border-orange-400/20">
                  <Crown className="w-2.5 h-2.5" />
                  {data.leader_id}
                </div>
              ) : (
                <div className="text-[9px] font-black text-white/20 italic">No Leader</div>
              )}
              <div className="flex items-center gap-3">
                <div className="text-right">
                  <p className="text-[7px] font-black text-white/20 uppercase tracking-tighter">Direct</p>
                  <p className="text-[10px] font-bold text-white/80 leading-none">{data.direct_count || 0}</p>
                </div>
                <div className="text-right border-l border-white/10 pl-3">
                  <p className="text-[7px] font-black text-white/20 uppercase tracking-tighter">Total</p>
                  <p className="text-[10px] font-bold text-primary leading-none">{data.total_count || 0}</p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Action Buttons */}
        <div 
          className={cn(
            "flex items-center justify-end gap-1 mt-auto pt-2 border-t border-white/5 transition-all duration-300 transform",
            (selected || isEditing) ? "opacity-100 translate-y-0" : "opacity-0 translate-y-2 pointer-events-none group-hover:opacity-100 group-hover:translate-y-0 group-hover:pointer-events-auto"
          )}
        >
          {!isEditing ? (
            <>
              <button 
                onClick={(e) => { e.stopPropagation(); data.onDelete?.(id); }}
                className="p-1.5 rounded-lg bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-colors"
                title="Delete"
              >
                <Minus className="w-3.5 h-3.5" />
              </button>
              <button 
                onClick={(e) => { e.stopPropagation(); setIsEditing(true); }}
                className="p-1.5 rounded-lg bg-white/5 text-white/50 hover:text-white transition-colors"
                title="Edit"
              >
                <Edit3 className="w-3.5 h-3.5" />
              </button>
              <button 
                onClick={(e) => { e.stopPropagation(); data.onAddChild?.(id); }}
                className="p-1.5 rounded-lg bg-primary/20 text-primary hover:bg-primary/30 transition-colors"
                title="Add Child"
              >
                <Plus className="w-3.5 h-3.5" />
              </button>
              {data.hasChildren && (
                <button 
                  onClick={(e) => { e.stopPropagation(); data.onToggleExpand?.(id); }}
                  className={cn(
                    "p-1.5 rounded-lg transition-colors",
                    data.isExpanded ? "bg-white/10 text-white" : "bg-accent/20 text-accent"
                  )}
                  title={data.isExpanded ? "Collapse" : "Expand"}
                >
                  <Layers className={cn("w-3.5 h-3.5", !data.isExpanded && "animate-pulse")} />
                </button>
              )}
            </>
          ) : (
            <>
              <button 
                onClick={handleCancel}
                className="flex items-center gap-1 px-2 py-1.5 rounded-lg bg-white/5 text-white/50 text-[10px] font-black uppercase"
              >
                <X className="w-3 h-3" /> {data.isInitialEditing ? 'Discard' : 'Cancel'}
              </button>
              <button 
                onClick={handleSave}
                className="flex items-center gap-1 px-2 py-1.5 rounded-lg bg-primary text-white text-[10px] font-black uppercase"
              >
                <Check className="w-3 h-3" /> Save
              </button>
            </>
          )}
        </div>
      </motion.div>

      <Handle type="source" position={Position.Bottom} className="opacity-0" />
    </div>
  );
});

OrgNode.displayName = 'OrgNode';
