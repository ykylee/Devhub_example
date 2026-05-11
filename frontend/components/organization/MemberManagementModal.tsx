"use client";

import { useState } from "react";
import { motion } from "framer-motion";
import { X, Users, Search, ChevronRight, ChevronLeft, Check, Crown } from "lucide-react";
import { OrgMember } from "@/lib/services/identity.service";
import { cn } from "@/lib/utils";

interface MemberManagementModalProps {
  unitId: string;
  unitName: string;
  allMembers: OrgMember[];
  currentMemberIds: string[];
  currentLeaderId?: string | null;
  onClose: () => void;
  onSave: (newMemberIds: string[], newLeaderId: string | null) => void | Promise<void>;
  saveError?: string | null;
}

export function MemberManagementModal({
  unitName,
  allMembers,
  currentMemberIds,
  currentLeaderId,
  onClose,
  onSave,
  saveError,
}: MemberManagementModalProps) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set(currentMemberIds));
  const [leaderId, setLeaderId] = useState<string | null>(currentLeaderId ?? null);
  const [searchAvailable, setSearchAvailable] = useState("");
  const [searchCurrent, setSearchCurrent] = useState("");
  const [isSaving, setIsSaving] = useState(false);

  const toggleMember = (id: string) => {
    const newSelected = new Set(selectedIds);
    if (newSelected.has(id)) {
      newSelected.delete(id);
      if (leaderId === id) {
        setLeaderId(null);
      }
    } else {
      newSelected.add(id);
    }
    setSelectedIds(newSelected);
  };

  const handleSaveClick = async () => {
    setIsSaving(true);
    try {
      await onSave(Array.from(selectedIds), leaderId);
    } finally {
      setIsSaving(false);
    }
  };

  const availableMembers = allMembers.filter(
    (m) =>
      !selectedIds.has(m.id) &&
      (m.name.toLowerCase().includes(searchAvailable.toLowerCase()) ||
        m.email.toLowerCase().includes(searchAvailable.toLowerCase())),
  );
  const currentMembers = allMembers.filter(
    (m) =>
      selectedIds.has(m.id) &&
      (m.name.toLowerCase().includes(searchCurrent.toLowerCase()) ||
        m.email.toLowerCase().includes(searchCurrent.toLowerCase())),
  );


  const toggleLeader = (id: string) => {
    setLeaderId((prev) => (prev === id ? null : id));
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />

      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        className="relative w-full max-w-4xl glass bg-[#030014]/90 border border-white/10 rounded-3xl overflow-hidden shadow-2xl flex flex-col max-h-[85vh]"
      >
        <div className="flex items-center justify-between p-6 border-b border-white/10">
          <div>
            <h2 className="text-xl font-black text-white uppercase tracking-tight flex items-center gap-2">
              <Users className="w-5 h-5 text-accent" /> Manage Members
            </h2>
            <p className="text-xs text-muted-foreground mt-1 font-bold">
              Editing roster for <span className="text-primary">{unitName}</span>
            </p>
          </div>
          <button
            onClick={onClose}
            className="p-2 rounded-xl hover:bg-white/5 text-muted-foreground hover:text-white transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="flex-1 flex overflow-hidden p-6 gap-6">
          <div className="flex-1 flex flex-col glass border border-white/10 rounded-2xl overflow-hidden">
            <div className="p-4 border-b border-white/10 bg-white/5">
              <h3 className="text-xs font-black uppercase tracking-widest text-white/60 mb-3">Available Personnel</h3>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
                <input
                  type="text"
                  placeholder="Search available..."
                  value={searchAvailable}
                  onChange={(e) => setSearchAvailable(e.target.value)}
                  className="w-full bg-black/20 border border-white/10 rounded-xl pl-9 pr-3 py-2 text-xs text-white placeholder:text-muted-foreground/50 focus:outline-none focus:border-primary/50"
                />
              </div>
            </div>
            <div className="flex-1 overflow-y-auto p-2 space-y-1 custom-scrollbar">
              {availableMembers.map((member) => (
                <div
                  key={member.id}
                  onClick={() => toggleMember(member.id)}
                  className="flex items-center justify-between p-3 rounded-xl hover:bg-white/5 cursor-pointer group transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-gradient-to-br from-white/10 to-white/5 flex items-center justify-center border border-white/10 text-xs font-bold text-white">
                      {member.name.substring(0, 2).toUpperCase()}
                    </div>
                    <div>
                      <p className="text-sm font-bold text-white group-hover:text-primary transition-colors">{member.name}</p>
                      <p className="text-[10px] text-muted-foreground flex items-center gap-2">
                        {member.email}
                        <span className="bg-white/10 px-1.5 py-0.5 rounded text-[8px] uppercase font-bold text-white/50 border border-white/5">{member.role}</span>
                      </p>
                    </div>
                  </div>
                  <button className="p-1.5 rounded-lg bg-primary/10 text-primary opacity-0 group-hover:opacity-100 transition-opacity">
                    <ChevronRight className="w-4 h-4" />
                  </button>
                </div>
              ))}
              {availableMembers.length === 0 && (
                <div className="h-full flex flex-col items-center justify-center p-6 text-center text-muted-foreground/50">
                  <p className="text-xs font-bold">No matching personnel found</p>
                </div>
              )}
            </div>
          </div>

          <div className="flex-1 flex flex-col glass border border-accent/20 rounded-2xl overflow-hidden relative">
            <div className="absolute inset-0 bg-accent/5 pointer-events-none" />
            <div className="p-4 border-b border-accent/20 bg-accent/10 relative z-10">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-xs font-black uppercase tracking-widest text-accent">Unit Roster</h3>
                <span className="text-[10px] font-black bg-accent text-[#030014] px-2 py-0.5 rounded-full">{selectedIds.size}</span>
              </div>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-accent/50" />
                <input
                  type="text"
                  placeholder="Search current roster..."
                  value={searchCurrent}
                  onChange={(e) => setSearchCurrent(e.target.value)}
                  className="w-full bg-black/20 border border-accent/20 rounded-xl pl-9 pr-3 py-2 text-xs text-white placeholder:text-accent/50 focus:outline-none focus:border-accent"
                />
              </div>
            </div>
            <div className="flex-1 overflow-y-auto p-2 space-y-1 custom-scrollbar relative z-10">
              {currentMembers.map((member) => {
                const isLeader = leaderId === member.id;
                return (
                  <div
                    key={member.id}
                    className={cn(
                      "flex items-center justify-between p-3 rounded-xl group transition-colors border",
                      isLeader
                        ? "bg-amber-500/10 border-amber-500/30"
                        : "border-transparent hover:border-accent/20 hover:bg-accent/10",
                    )}
                  >
                    <button
                      type="button"
                      onClick={() => toggleMember(member.id)}
                      className="p-1.5 rounded-lg bg-red-500/10 text-red-400 opacity-0 group-hover:opacity-100 transition-opacity"
                      title="Remove from unit"
                    >
                      <ChevronLeft className="w-4 h-4" />
                    </button>
                    <div className="flex items-center gap-3 text-right flex-1 justify-end">
                      <div className="text-right">
                        <p className="text-sm font-bold text-white group-hover:text-accent transition-colors flex items-center justify-end gap-2">
                          {isLeader && <Crown className="w-3.5 h-3.5 text-amber-400" />}
                          {member.name}
                        </p>
                        <p className="text-[10px] text-muted-foreground flex items-center justify-end gap-2">
                          <span className="bg-accent/10 px-1.5 py-0.5 rounded text-[8px] uppercase font-bold text-accent/50 border border-accent/20">
                            {member.role}
                          </span>
                          {member.email}
                        </p>
                      </div>
                      <div className="w-8 h-8 rounded-full bg-gradient-to-br from-accent/20 to-accent/5 flex items-center justify-center border border-accent/20 text-xs font-bold text-accent">
                        {member.name.substring(0, 2).toUpperCase()}
                      </div>
                    </div>
                    <button
                      type="button"
                      onClick={() => toggleLeader(member.id)}
                      className={cn(
                        "ml-2 px-2 py-1 rounded-lg text-[9px] font-black uppercase tracking-widest transition-all flex items-center gap-1",
                        isLeader
                          ? "bg-amber-500/20 text-amber-300 border border-amber-500/40"
                          : "bg-white/5 text-muted-foreground border border-white/10 opacity-0 group-hover:opacity-100",
                      )}
                      title={isLeader ? "Remove leader" : "Set as leader"}
                    >
                      <Crown className="w-3 h-3" />
                      {isLeader ? "Leader" : "Set lead"}
                    </button>
                  </div>
                );
              })}
              {currentMembers.length === 0 && (
                <div className="h-full flex flex-col items-center justify-center p-6 text-center text-accent/30">
                  <Users className="w-8 h-8 mb-2 opacity-50" />
                  <p className="text-xs font-bold uppercase tracking-widest">Empty Roster</p>
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="p-6 border-t border-white/10 bg-black/40 flex justify-between items-center gap-4">
          <div className="flex flex-col gap-1 min-w-0">
            <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
              {selectedIds.size} total personnel · leader:{" "}
              <span className={cn("ml-1", leaderId ? "text-amber-300" : "text-muted-foreground")}>
                {leaderId ? allMembers.find((m) => m.id === leaderId)?.name ?? leaderId : "(none)"}
              </span>
            </p>
            {saveError && (
              <p className="text-[10px] font-bold text-red-400 truncate" title={saveError}>
                {saveError}
              </p>
            )}
          </div>
          <div className="flex gap-3">
            <button
              onClick={onClose}
              disabled={isSaving}
              className="px-6 py-2.5 rounded-xl border border-white/10 text-xs font-bold text-white hover:bg-white/5 transition-all uppercase tracking-widest disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Cancel
            </button>
            <button
              onClick={handleSaveClick}
              disabled={isSaving}
              className="px-6 py-2.5 rounded-xl bg-accent text-[#030014] text-xs font-black uppercase tracking-widest hover:bg-accent/90 hover:scale-105 active:scale-95 transition-all flex items-center gap-2 shadow-[0_0_20px_rgba(var(--accent),0.3)] disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
            >
              <Check className="w-4 h-4" /> {isSaving ? "Saving..." : "Save Configuration"}
            </button>
          </div>
        </div>
      </motion.div>
    </div>
  );
}
