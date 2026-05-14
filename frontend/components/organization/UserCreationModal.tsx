"use client";

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { X, UserPlus, Mail, Shield, Building2, Loader2, Search, Bot, User, Key } from "lucide-react";
import { identityService, OrgMember } from "@/lib/services/identity.service";
import { Role } from "@/lib/services/rbac.types";
import { cn } from "@/lib/utils";

interface UserCreationModalProps {
  onClose: () => void;
  onCreated: (user: OrgMember) => void;
  roles: Role[];
}

export function UserCreationModal({ onClose, onCreated, roles }: UserCreationModalProps) {
  const [formData, setFormData] = useState({
    user_id: "",
    email: "",
    display_name: "",
    role: "Developer",
    status: "active",
    type: "human" as "human" | "system",
    password: "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [lookingUp, setLookingUp] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const handleLookup = async () => {
    if (!formData.user_id) return;
    setLookingUp(true);
    setError(null);
    try {
      const data = await identityService.lookupHR(formData.user_id);
      setFormData(prev => ({
        ...prev,
        email: data.email,
        type: "human"
      }));
    } catch {
      setError("Not found in HR database. You can still enter details manually.");
    } finally {
      setLookingUp(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    try {
      const created = await identityService.createUser({
        user_id: formData.user_id,
        email: formData.email,
        display_name: formData.display_name,
        role: formData.role as OrgMember["role"],
        status: formData.status as OrgMember["status"],
        type: formData.type,
        password: formData.password || undefined,
      });
      onCreated(created);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create user");
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
        className="relative w-full max-w-xl glass border-border rounded-3xl shadow-2xl overflow-hidden"
        role="dialog"
        aria-modal="true"
        aria-labelledby="user-modal-title"
      >
        <div className="p-8 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-primary/20 rounded-xl flex items-center justify-center">
              <UserPlus className="w-5 h-5 text-primary" />
            </div>
            <div>
              <h2 id="user-modal-title" className="text-xl font-black text-primary-foreground uppercase tracking-tight">Add <span className="text-primary">Member</span></h2>
              <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Register human or system account</p>
            </div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-muted/30 rounded-xl text-muted-foreground transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-8 space-y-6 max-h-[70vh] overflow-y-auto custom-scrollbar">
          {/* Account Type Selection */}
          <div className="space-y-3">
            <label className="text-[10px] font-black text-primary-foreground/40 uppercase tracking-widest px-1">Account Type</label>
            <div className="grid grid-cols-2 gap-4">
              <button
                type="button"
                onClick={() => setFormData({ ...formData, type: 'human' })}
                className={cn(
                  "flex items-center gap-3 p-4 rounded-2xl border transition-all",
                  formData.type === 'human' 
                    ? "bg-primary/10 border-primary text-primary-foreground shadow-lg shadow-primary/10" 
                    : "bg-muted/30 border-border text-muted-foreground hover:bg-muted/40"
                )}
              >
                <User className="w-5 h-5" />
                <div className="text-left">
                  <p className="text-xs font-black uppercase tracking-widest">Human</p>
                  <p className="text-[10px] opacity-60">Verified personnel</p>
                </div>
              </button>
              <button
                type="button"
                onClick={() => setFormData({ ...formData, type: 'system' })}
                className={cn(
                  "flex items-center gap-3 p-4 rounded-2xl border transition-all",
                  formData.type === 'system' 
                    ? "bg-accent/10 border-accent text-primary-foreground shadow-lg shadow-accent/10" 
                    : "bg-muted/30 border-border text-muted-foreground hover:bg-muted/40"
                )}
              >
                <Bot className="w-5 h-5" />
                <div className="text-left">
                  <p className="text-xs font-black uppercase tracking-widest">System/AI</p>
                  <p className="text-[10px] opacity-60">Automation & bots</p>
                </div>
              </button>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-primary-foreground/40 uppercase tracking-widest px-1">User ID / System ID</label>
              <div className="relative group flex gap-2">
                <div className="relative flex-1">
                  <Building2 className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-primary-foreground/20 group-focus-within:text-primary transition-colors" />
                  <input
                    required
                    value={formData.user_id}
                    onChange={e => setFormData({ ...formData, user_id: e.target.value })}
                    placeholder="e.g. yklee or bot-gardener"
                    className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-primary-foreground focus:outline-none focus:ring-1 focus:ring-primary/50"
                  />
                </div>
                {formData.type === 'human' && (
                  <button
                    type="button"
                    onClick={handleLookup}
                    disabled={lookingUp || !formData.user_id}
                    className="glass border-border px-4 rounded-2xl hover:bg-muted/40 transition-all text-primary-foreground disabled:opacity-30"
                    title="Fetch from HR DB"
                  >
                    {lookingUp ? <Loader2 className="w-4 h-4 animate-spin" /> : <Search className="w-4 h-4" />}
                  </button>
                )}
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-primary-foreground/40 uppercase tracking-widest px-1">Display Name</label>
              <input
                required
                value={formData.display_name}
                onChange={e => setFormData({ ...formData, display_name: e.target.value })}
                placeholder="e.g. YK Lee"
                className="w-full bg-muted/30 border border-border rounded-2xl px-4 py-3 text-sm text-primary-foreground focus:outline-none focus:ring-1 focus:ring-primary/50"
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-[10px] font-black text-primary-foreground/40 uppercase tracking-widest px-1">Email Address</label>
            <div className="relative group">
              <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-primary-foreground/20 group-focus-within:text-primary transition-colors" />
              <input
                required
                type="email"
                value={formData.email}
                onChange={e => setFormData({ ...formData, email: e.target.value })}
                placeholder="e.g. gardener@devhub.internal"
                className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-primary-foreground focus:outline-none focus:ring-1 focus:ring-primary/50"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-primary-foreground/40 uppercase tracking-widest px-1">Initial Role</label>
              <div className="relative group">
                <Shield className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-primary-foreground/20 group-focus-within:text-primary transition-colors" />
                <select
                  value={formData.role}
                  onChange={e => setFormData({ ...formData, role: e.target.value })}
                  className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-primary-foreground focus:outline-none focus:ring-1 focus:ring-primary/50 appearance-none"
                >
                  {roles.map(r => <option key={r.id} value={r.name} className="bg-slate-900">{r.name}</option>)}
                </select>
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-[10px] font-black text-primary-foreground/40 uppercase tracking-widest px-1">Account Password (Optional)</label>
              <div className="relative group">
                <Key className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-primary-foreground/20 group-focus-within:text-primary transition-colors" />
                <input
                  type="password"
                  value={formData.password}
                  onChange={e => setFormData({ ...formData, password: e.target.value })}
                  placeholder="Set to enable immediate login"
                  className="w-full bg-muted/30 border border-border rounded-2xl pl-12 pr-4 py-3 text-sm text-primary-foreground focus:outline-none focus:ring-1 focus:ring-primary/50"
                />
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

          {/* Dev Helper */}
          <div className="p-4 bg-muted/30 rounded-2xl border border-border/60">
            <p className="text-[9px] font-black text-primary-foreground/20 uppercase tracking-widest mb-2 flex items-center gap-1">
              <span className="w-1.5 h-1.5 rounded-full bg-amber-500/50" /> Dev Info: Mock HR Personnel
            </p>
            <div className="flex flex-wrap gap-2">
              {['yklee', 'akim', 'sjones'].map(id => (
                <button 
                  key={id}
                  type="button"
                  onClick={() => setFormData(prev => ({ ...prev, user_id: id, type: 'human' }))}
                  className="px-2 py-1 bg-muted/30 hover:bg-muted/40 rounded-lg text-[9px] text-primary-foreground/40 transition-all border border-border/60"
                >
                  {id}
                </button>
              ))}
            </div>
          </div>

          <div className="flex gap-4 pt-4 border-t border-border/60">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 glass border-border text-primary-foreground font-bold py-4 rounded-2xl hover:bg-muted/30 transition-all uppercase tracking-widest text-[10px]"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="flex-1 bg-primary text-primary-foreground font-black py-4 px-8 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all shadow-xl shadow-primary/20 disabled:opacity-50 uppercase tracking-widest text-[10px] flex items-center justify-center gap-2"
            >
              {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <>Create {formData.type === 'human' ? 'Member' : 'System Account'}</>}
            </button>
          </div>
        </form>
      </motion.div>
    </div>
  );
}
