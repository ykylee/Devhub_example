"use client";

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { X, Building2, Check, AlertCircle } from "lucide-react";
import { OrgUnit, CreateUnitPayload, UpdateUnitPayload, OrgNode, OrgMember } from "@/lib/services/identity.service";
import { cn } from "@/lib/utils";

interface UnitManagementModalProps {
  mode: "create" | "edit";
  initialData?: Partial<OrgUnit>;
  availableParents: OrgNode[];
  availableLeaders: OrgMember[];
  onClose: () => void;
  onSave: (payload: CreateUnitPayload | UpdateUnitPayload) => Promise<void>;
}

const UNIT_TYPES = [
  { id: "division", label: "Division" },
  { id: "team", label: "Team" },
  { id: "group", label: "Group" },
  { id: "part", label: "Part" },
  { id: "company", label: "Company" },
];

export function UnitManagementModal({
  mode,
  initialData,
  availableParents,
  availableLeaders,
  onClose,
  onSave,
}: UnitManagementModalProps) {
  const [label, setLabel] = useState(initialData?.label || "");
  const [type, setType] = useState<OrgUnit["unit_type"]>(initialData?.unit_type || "team");
  const [parentId, setParentId] = useState(initialData?.parent_unit_id || "");
  const [leaderId, setLeaderId] = useState(initialData?.leader_user_id || "");
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!label.trim()) {
      setError("Unit name is required");
      return;
    }

    setIsSaving(true);
    setError(null);

    try {
      if (mode === "create") {
        await onSave({
          unit_id: `unit-${Date.now()}`, // Simple ID generation for POC
          label,
          unit_type: type,
          parent_unit_id: parentId || undefined,
          leader_user_id: leaderId || undefined,
        } as CreateUnitPayload);
      } else {
        await onSave({
          label,
          unit_type: type,
          parent_unit_id: parentId || undefined,
          leader_user_id: leaderId || "",
        } as UpdateUnitPayload);
      }
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save unit");
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="absolute inset-0 bg-background/70 backdrop-blur-sm"
        onClick={onClose}
      />

      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        className="relative w-full max-w-lg glass bg-background/90 border border-border rounded-3xl overflow-hidden shadow-2xl"
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
      >
        <div className="flex items-center justify-between p-6 border-b border-border">
          <h2 id="modal-title" className="text-xl font-black text-foreground uppercase tracking-tight flex items-center gap-2">
            <Building2 className="w-5 h-5 text-accent" />
            {mode === "create" ? "Create New Unit" : "Edit Unit Details"}
          </h2>
          <button
            onClick={onClose}
            className="p-2 rounded-xl hover:bg-muted/40 text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Close"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {error && (
            <div className="p-4 rounded-2xl bg-red-500/10 border border-red-500/20 flex items-start gap-3 animate-shake">
              <AlertCircle className="w-4 h-4 text-red-400 mt-0.5" />
              <p className="text-xs font-bold text-red-400">{error}</p>
            </div>
          )}

          <div className="space-y-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground ml-1">
              Unit Name
            </label>
            <input
              type="text"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              placeholder="e.g. Infrastructure Team"
              className="w-full bg-muted/30 border border-border rounded-xl px-4 py-3 text-sm text-foreground focus:outline-none focus:border-accent/50 transition-all"
              autoFocus
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground ml-1">
                Unit Type
              </label>
              <select
                value={type}
                onChange={(e) => setType(e.target.value as OrgUnit["unit_type"])}
                className="themed-select w-full"
              >
                {UNIT_TYPES.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-2">
              <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground ml-1">
                Parent Unit
              </label>
              <select
                value={parentId}
                onChange={(e) => setParentId(e.target.value)}
                className="themed-select w-full"
              >
                <option value="">None (Root)</option>
                {availableParents
                  .filter((p) => p.id !== initialData?.unit_id)
                  .map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.data.label}
                    </option>
                  ))}
              </select>
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground ml-1">
              Unit Leader
            </label>
            <select
              value={leaderId}
              onChange={(e) => setLeaderId(e.target.value)}
              className="themed-select w-full"
            >
              <option value="">None</option>
              {availableLeaders.map((m) => (
                <option key={m.id} value={m.id}>
                  {m.name} ({m.id})
                </option>
              ))}
            </select>
          </div>

          <div className="pt-4 flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-6 py-3 rounded-xl border border-border text-xs font-bold text-foreground hover:bg-muted/40 transition-all uppercase tracking-widest"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSaving}
              className="flex-[1.5] px-6 py-3 rounded-xl bg-accent text-accent-foreground text-xs font-black uppercase tracking-widest hover:bg-accent/90 hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2 shadow-[0_0_20px_rgba(var(--accent),0.3)]"
            >
              <Check className="w-4 h-4" />
              {isSaving ? "Saving..." : mode === "create" ? "Create Unit" : "Save Changes"}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
