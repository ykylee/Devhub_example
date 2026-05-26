"use client";

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { X, Edit3, Mail, Shield, Loader2 } from "lucide-react";
import { identityService, OrgMember } from "@/lib/services/identity.service";
import { Role } from "@/lib/services/rbac.types";

interface UserEditModalProps {
  initial: OrgMember;
  roles: Role[];
  onClose: () => void;
  onUpdated: (user: OrgMember) => void;
}

export function UserEditModal({ initial, roles, onClose, onUpdated }: UserEditModalProps) {
  const [formData, setFormData] = useState({
    display_name: initial.name,
    email: initial.email,
    role: initial.role,
    status: initial.status,
  });
  const [submitting, setSubmitting] = useState(false);
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
    setError(null);
    setSubmitting(true);
    try {
      const updated = await identityService.updateUser(initial.id, {
        display_name: formData.display_name,
        email: formData.email,
        role: formData.role,
        status: formData.status,
      });
      onUpdated(updated);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update user");
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
        className="relative w-full max-w-lg glass border-border rounded-3xl shadow-2xl overflow-hidden"
        role="dialog"
        aria-modal="true"
        aria-labelledby="user-edit-modal-title"
      >
        <div className="p-8 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-primary/20 rounded-xl flex items-center justify-center">
              <Edit3 className="w-5 h-5 text-primary" />
            </div>
            <div>
              <h2 id="user-edit-modal-title" className="text-xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tight">Edit <span className="text-primary">Member</span></h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">{initial.id}</p>
            </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-muted/30 rounded-xl text-muted-foreground transition-colors" aria-label="Close">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-8 space-y-6 max-h-[70vh] overflow-y-auto custom-scrollbar">
          <div className="space-y-2">
            <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Display Name</label>
            <input
              required
              value={formData.display_name}
              onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
              className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-primary/50"
            />
          </div>

          <div className="space-y-2">
            <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Email</label>
            <div className="relative group">
              <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/35 group-focus-within:text-primary transition-colors" />
              <input
                required
                type="email"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-primary/50"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Role</label>
              <div className="relative group">
                <Shield className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground/35 group-focus-within:text-primary transition-colors" />
                <select
                  value={formData.role}
                  onChange={(e) => setFormData({ ...formData, role: e.target.value as OrgMember["role"] })}
                  className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 appearance-none"
                >
                  {roles.map((r) => (
                    <option key={r.id} value={r.name} className="bg-slate-900">{r.name}</option>
                  ))}
                </select>
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">Status</label>
              <select
                value={formData.status}
                onChange={(e) => setFormData({ ...formData, status: e.target.value as OrgMember["status"] })}
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-foreground dark:text-primary-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 appearance-none"
              >
                <option value="active" className="bg-slate-900">active</option>
                <option value="pending" className="bg-slate-900">pending</option>
                <option value="deactivated" className="bg-slate-900">deactivated</option>
              </select>
            </div>
          </div>

          {error && (
            <div className="p-4 bg-accent/10 border border-accent/20 rounded-2xl text-[11px] text-accent font-medium">
              {error}
            </div>
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
              className="flex-1 bg-primary text-primary-foreground font-black py-4 px-8 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all shadow-xl shadow-primary/20 disabled:opacity-50 uppercase tracking-widest text-[10px] flex items-center justify-center gap-2"
            >
              {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : "Save Changes"}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
