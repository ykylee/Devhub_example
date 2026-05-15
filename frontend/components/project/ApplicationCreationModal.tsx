"use client";

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { X, Box, Info, User, Globe, Eye, Lock, Loader2, Calendar, Building2 } from "lucide-react";
import { Application, ApplicationStatus, ApplicationVisibility } from "@/lib/services/project.types";
import { projectService } from "@/lib/services/project.service";
import { cn } from "@/lib/utils";

interface ApplicationCreationModalProps {
  onClose: () => void;
  onCreated: (app: Application) => void;
  initialData?: Partial<Application>;
}

export function ApplicationCreationModal({ onClose, onCreated, initialData }: ApplicationCreationModalProps) {
  const [formData, setFormData] = useState({
    key: initialData?.key || "",
    name: initialData?.name || "",
    description: initialData?.description || "",
    owner_user_id: initialData?.owner_user_id || "",
    leader_user_id: initialData?.leader_user_id || "",
    development_unit_id: initialData?.development_unit_id || "",
    visibility: initialData?.visibility || "internal" as ApplicationVisibility,
    status: initialData?.status || "planning" as ApplicationStatus,
    start_date: initialData?.start_date || "",
    due_date: initialData?.due_date || "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEdit = !!initialData?.id;

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    try {
      let result: Application;
      if (isEdit && initialData.id) {
        // PATCH 시 `key` 는 백엔드(updateApplication)에서 immutable 로 reject 되므로
        // payload 에서 제외한다. (codex PR #114 review P1)
        const { key: _omitKey, ...patchPayload } = formData;
        result = await projectService.updateApplication(initialData.id, patchPayload);
      } else {
        result = await projectService.createApplication(formData);
      }
      onCreated(result);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save application");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-6">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
      />
      
      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        className="relative w-full max-w-2xl glass border-border rounded-3xl shadow-2xl overflow-hidden"
        role="dialog"
        aria-modal="true"
      >
        <div className="p-8 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-purple-500/20 rounded-xl flex items-center justify-center">
              <Box className="w-5 h-5 text-purple-400" />
            </div>
            <div>
              <h2 className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">
                {isEdit ? "Edit" : "Create"} <span className="text-purple-400">Application</span>
              </h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                Governance container for repositories & projects
              </p>
            </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-muted/30 rounded-xl text-muted-foreground transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-8 space-y-6 max-h-[75vh] overflow-y-auto custom-scrollbar">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Application Key</label>
              <div className="relative group">
                <Info className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-primary-foreground/20 group-focus-within:text-purple-400 transition-colors" />
                <input
                  required
                  disabled={isEdit}
                  value={formData.key}
                  onChange={e => setFormData({ ...formData, key: e.target.value.toUpperCase() })}
                  placeholder="E.G. PLATFORM26"
                  maxLength={10}
                  pattern="^[A-Za-z0-9]{1,10}$"
                  className={cn(
                    "w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm font-mono text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50 uppercase",
                    isEdit && "opacity-50 cursor-not-allowed"
                  )}
                />
              </div>
              {!isEdit && (
                <p className="text-[9px] text-muted-foreground px-1 italic">Immutable 10-char alphanumeric ID.</p>
              )}
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Display Name</label>
              <input
                required
                value={formData.name}
                onChange={e => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g. DevHub Platform 2026"
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50"
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Description</label>
            <textarea
              value={formData.description}
              onChange={e => setFormData({ ...formData, description: e.target.value })}
              placeholder="Strategic goals, KPI, and scope summary..."
              rows={3}
              className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50 resize-none"
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Application Leader</label>
              <div className="relative group">
                <User className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-primary-foreground/20 group-focus-within:text-purple-400 transition-colors" />
                <input
                  required
                  value={formData.leader_user_id}
                  onChange={e => setFormData({ ...formData, leader_user_id: e.target.value })}
                  placeholder="e.g. charlie"
                  className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50"
                />
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Development Department</label>
              <div className="relative group">
                <Building2 className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-primary-foreground/20 group-focus-within:text-purple-400 transition-colors" />
                <input
                  required
                  value={formData.development_unit_id}
                  onChange={e => setFormData({ ...formData, development_unit_id: e.target.value })}
                  placeholder="e.g. dept-eng"
                  className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50"
                />
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Owner User (Legacy)</label>
              <div className="relative group">
                <User className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-primary-foreground/20 group-focus-within:text-purple-400 transition-colors" />
                <input
                  required
                  value={formData.owner_user_id}
                  onChange={e => setFormData({ ...formData, owner_user_id: e.target.value })}
                  placeholder="e.g. charlie"
                  className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50"
                />
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Visibility</label>
              <div className="grid grid-cols-3 gap-2">
                {(['public', 'internal', 'restricted'] as ApplicationVisibility[]).map((v) => (
                  <button
                    key={v}
                    type="button"
                    onClick={() => setFormData({ ...formData, visibility: v })}
                    className={cn(
                      "py-2.5 rounded-xl border text-[10px] font-black uppercase tracking-widest transition-all flex flex-col items-center gap-1",
                      formData.visibility === v
                        ? "bg-purple-500/10 border-purple-500/40 text-purple-400 shadow-lg shadow-purple-500/5"
                        : "bg-muted/20 border-border/60 text-muted-foreground hover:bg-muted/40"
                    )}
                  >
                    {v === 'public' && <Globe className="w-3 h-3" />}
                    {v === 'internal' && <Eye className="w-3 h-3" />}
                    {v === 'restricted' && <Lock className="w-3 h-3" />}
                    {v}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Operating Status</label>
              <select
                value={formData.status}
                onChange={e => setFormData({ ...formData, status: e.target.value as ApplicationStatus })}
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50 appearance-none"
              >
                <option value="planning" className="bg-slate-900">Planning</option>
                <option value="active" className="bg-slate-900">Active</option>
                <option value="on_hold" className="bg-slate-900">On Hold</option>
                <option value="closed" className="bg-slate-900">Closed</option>
                <option value="archived" className="bg-slate-900">Archived</option>
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Period (Optional)</label>
              <div className="flex items-center gap-2">
                <div className="relative flex-1">
                  <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground/40" />
                  <input
                    type="date"
                    value={formData.start_date}
                    onChange={e => setFormData({ ...formData, start_date: e.target.value })}
                    className="w-full bg-muted/20 border border-border/40 rounded-xl pl-10 pr-3 py-2 text-xs text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50"
                  />
                </div>
                <span className="text-muted-foreground/40">→</span>
                <div className="relative flex-1">
                  <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground/40" />
                  <input
                    type="date"
                    value={formData.due_date}
                    onChange={e => setFormData({ ...formData, due_date: e.target.value })}
                    className="w-full bg-muted/20 border border-border/40 rounded-xl pl-10 pr-3 py-2 text-xs text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-purple-400/50"
                  />
                </div>
              </div>
            </div>
          </div>

          {error && (
            <motion.div 
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              className="p-4 bg-orange-500/10 border border-orange-500/20 rounded-2xl text-[11px] text-orange-400 font-medium"
            >
              {error}
            </motion.div>
          )}

          <div className="flex gap-4 pt-4 border-t border-border/60">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 glass border-border text-foreground dark:text-primary-foreground font-bold py-4 rounded-2xl hover:bg-muted/30 transition-all uppercase tracking-widest text-[10px]"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="flex-1 bg-purple-600 text-white font-black py-4 px-8 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all shadow-xl shadow-purple-500/20 disabled:opacity-50 uppercase tracking-widest text-[10px] flex items-center justify-center gap-2"
            >
              {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <>{isEdit ? 'Save Changes' : 'Create Application'}</>}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
