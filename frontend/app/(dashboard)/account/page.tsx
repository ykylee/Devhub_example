"use client";

import { motion } from "framer-motion";
import { User, Mail, Shield, Key, Loader2, Save, AlertCircle, CheckCircle2 } from "lucide-react";
import { useState } from "react";
import { useStore } from "@/lib/store";
import { accountService, SettingsFlowError } from "@/lib/services/account.service";

export default function AccountPage() {
  const { actor } = useStore();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);
  const [reauthURL, setReauthURL] = useState<string | null>(null);

  const handlePasswordUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newPassword !== confirmPassword) {
      setMessage({ type: 'error', text: "New passwords do not match." });
      return;
    }

    setSubmitting(true);
    setMessage(null);
    setReauthURL(null);
    try {
      await accountService.updateMyPassword(currentPassword, newPassword);
      setMessage({ type: 'success', text: "Password updated successfully." });
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err) {
      if (err instanceof SettingsFlowError && err.code === "REAUTH_REQUIRED") {
        setReauthURL(err.redirectURL ?? "/login");
        setMessage({
          type: 'error',
          text: "Re-authentication required to change your password. Please sign in again.",
        });
      } else if (err instanceof SettingsFlowError) {
        setMessage({ type: 'error', text: err.message });
      } else {
        setMessage({ type: 'error', text: err instanceof Error ? err.message : "Failed to update password." });
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="space-y-8 max-w-4xl mx-auto pb-12">
      {/* Header Section */}
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-black text-foreground dark:text-white tracking-tighter uppercase">
          Account <span className="text-primary">Settings</span>
        </h1>
        <p className="text-muted-foreground font-bold text-xs uppercase tracking-widest">
          Manage your DevHub identity and security preferences
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {/* Profile Info */}
        <motion.div 
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          className="md:col-span-1 space-y-6"
        >
          <div className="glass border-white/10 rounded-3xl p-8 text-center space-y-4">
            <div className="w-24 h-24 mx-auto rounded-2xl bg-gradient-to-br from-primary to-accent flex items-center justify-center border border-white/20 shadow-2xl ring-4 ring-primary/10">
              <User className="w-12 h-12 text-white" />
            </div>
            <div className="space-y-1">
              <h2 className="text-xl font-bold text-foreground dark:text-white tracking-tight">{actor?.login || "Guest User"}</h2>
              <p className="text-xs font-bold text-primary uppercase tracking-widest">{actor?.role || "Developer"}</p>
            </div>
            <div className="pt-4 border-t border-white/5 space-y-3">
              <div className="flex items-center gap-3 text-xs text-muted-foreground justify-center">
                <Mail className="w-3.5 h-3.5" />
                <span className="font-medium">{actor?.login}@example.com</span>
              </div>
              <div className="flex items-center gap-3 text-[10px] text-white/40 justify-center uppercase tracking-widest font-black">
                <Shield className="w-3 h-3" />
                <span>Subject: {actor?.subject?.slice(0, 8) || "N/A"}...</span>
              </div>
            </div>
          </div>

          <div className="glass border-white/10 rounded-3xl p-6 space-y-4">
            <h3 className="text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em] px-2">Session Info</h3>
            <div className="space-y-3">
              <div className="flex items-center justify-between px-2">
                <span className="text-[10px] font-bold text-white/50 uppercase tracking-widest">Source</span>
                <span className="text-[10px] font-mono text-primary font-bold uppercase">{actor?.source || "Local"}</span>
              </div>
              <div className="flex items-center justify-between px-2">
                <span className="text-[10px] font-bold text-white/50 uppercase tracking-widest">MFA Status</span>
                <span className="text-[10px] font-bold text-emerald-400 uppercase">Disabled</span>
              </div>
            </div>
          </div>
        </motion.div>

        {/* Security / Password Form */}
        <motion.div 
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          className="md:col-span-2"
        >
          <div className="glass border-white/10 rounded-3xl overflow-hidden shadow-2xl">
            <div className="bg-white/5 px-8 py-6 border-b border-white/10 flex items-center gap-4">
              <div className="p-2.5 rounded-xl bg-primary/20 border border-primary/20">
                <Key className="w-5 h-5 text-primary" />
              </div>
              <div>
                <h3 className="text-lg font-bold text-foreground dark:text-white tracking-tight">Security Credentials</h3>
                <p className="text-xs text-muted-foreground">Update your authentication password</p>
              </div>
            </div>
            
            <form onSubmit={handlePasswordUpdate} className="p-8 space-y-6">
              <div className="space-y-4">
                <div className="space-y-2">
                  <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">
                    Current Password
                  </label>
                  <input
                    type="password"
                    required
                    value={currentPassword}
                    onChange={(e) => setCurrentPassword(e.target.value)}
                    className="w-full bg-white/50 dark:bg-black/40 border border-white/10 rounded-2xl px-4 py-3 text-foreground dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-muted-foreground/50"
                    placeholder="Enter current password"
                  />
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">
                      New Password
                    </label>
                    <input
                      type="password"
                      required
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                      className="w-full bg-white/50 dark:bg-black/40 border border-white/10 rounded-2xl px-4 py-3 text-foreground dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-muted-foreground/50"
                      placeholder="Min 8 characters"
                    />
                  </div>
                  <div className="space-y-2">
                    <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest px-1">
                      Confirm New Password
                    </label>
                    <input
                      type="password"
                      required
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      className="w-full bg-white/50 dark:bg-black/40 border border-white/10 rounded-2xl px-4 py-3 text-foreground dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-muted-foreground/50"
                      placeholder="Repeat new password"
                    />
                  </div>
                </div>
              </div>

              {message && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  className={cn(
                    "flex flex-col gap-3 p-4 rounded-2xl border text-sm font-medium",
                    message.type === 'success'
                      ? "bg-emerald-500/10 border-emerald-500/20 text-emerald-400"
                      : "bg-rose-500/10 border-rose-500/20 text-rose-400"
                  )}
                >
                  <div className="flex items-center gap-3">
                    {message.type === 'success' ? <CheckCircle2 className="w-5 h-5" /> : <AlertCircle className="w-5 h-5" />}
                    {message.text}
                  </div>
                  {reauthURL && (
                    <button
                      type="button"
                      onClick={() => window.location.assign(reauthURL)}
                      className="self-start px-4 py-2 rounded-xl bg-rose-500/20 hover:bg-rose-500/30 text-rose-100 text-xs font-bold uppercase tracking-widest transition-all"
                    >
                      Sign In Again
                    </button>
                  )}
                </motion.div>
              )}

              <div className="pt-4 border-t border-white/5 flex justify-end items-center gap-4">
                <p className="text-[9px] text-muted-foreground font-bold uppercase tracking-widest max-w-[200px]">
                  Password must include uppercase, lowercase, and symbols.
                </p>
                <button
                  type="submit"
                  disabled={submitting || !currentPassword || !newPassword}
                  className="bg-primary text-white font-black px-8 py-3.5 rounded-2xl hover:scale-105 active:scale-95 transition-all flex items-center gap-2 shadow-lg shadow-primary/20 disabled:opacity-50 disabled:hover:scale-100 uppercase tracking-widest text-xs"
                >
                  {submitting ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Updating...
                    </>
                  ) : (
                    <>
                      <Save className="w-4 h-4" />
                      Save Changes
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>

          <div className="mt-8 glass border-white/10 rounded-3xl p-8 border-l-4 border-l-amber-500/50">
            <div className="flex items-start gap-4">
              <div className="p-2.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-500">
                <Shield className="w-5 h-5" />
              </div>
              <div className="space-y-1">
                <h4 className="text-foreground dark:text-white font-bold tracking-tight">Two-Factor Authentication</h4>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  Enhance your account security by adding an extra layer of protection. This feature is currently in preview and will be available in v0.6.0.
                </p>
                <button className="mt-4 text-[10px] font-black text-amber-500 uppercase tracking-widest hover:underline opacity-50 cursor-not-allowed">
                  Setup MFA Coming Soon
                </button>
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  );
}

function cn(...inputs: (string | boolean | undefined | null)[]) {
  return inputs.filter(Boolean).join(" ");
}
