"use client";

import { useState } from "react";
import { motion } from "framer-motion";
import { UserCircle, KeyRound, AlertCircle, Check } from "lucide-react";
import { useStore } from "@/lib/store";
import { accountService } from "@/lib/services/account.service";
import { useToast } from "@/components/ui/Toast";

export default function AccountPage() {
  const { role } = useStore();
  const { toast } = useToast();
  const [isUpdating, setIsUpdating] = useState(false);
  
  // Password Form State
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  const handlePasswordUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newPassword !== confirmPassword) {
      toast("New passwords do not match", "error");
      return;
    }
    
    setIsUpdating(true);
    try {
      await accountService.updateMyPassword(currentPassword, newPassword);
      toast("Password updated successfully", "success");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (error) {
      toast(error instanceof Error ? error.message : "Failed to update password", "error");
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      <div className="flex items-center gap-4">
        <div className="p-3 bg-primary/20 rounded-2xl border border-primary/30">
          <UserCircle className="w-8 h-8 text-primary" />
        </div>
        <div>
          <h1 className="text-3xl font-black text-white tracking-tight">Account Settings</h1>
          <p className="text-muted-foreground text-sm font-medium">Manage your identity and security preferences.</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {/* Profile Info Card */}
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="col-span-1 glass rounded-3xl p-6 border-white/10 h-fit"
        >
          <div className="flex flex-col items-center text-center space-y-4 mb-6">
            <div className="w-24 h-24 rounded-full bg-gradient-to-br from-primary/20 to-accent/20 border border-white/10 flex items-center justify-center relative">
              <UserCircle className="w-12 h-12 text-white/50" />
              <div className="absolute -bottom-1 -right-1 bg-green-500 w-5 h-5 rounded-full border-2 border-[#030014]" />
            </div>
            <div>
              <h3 className="text-xl font-bold text-white">Mock User</h3>
              <p className="text-sm text-primary">{role}</p>
            </div>
          </div>
          
          <div className="space-y-4">
            <div className="bg-white/5 rounded-xl p-3 border border-white/5">
              <p className="text-[10px] uppercase tracking-widest text-muted-foreground font-bold mb-1">Login ID</p>
              <p className="text-sm text-white font-medium flex items-center justify-between">
                mockuser
                <span className="text-[10px] text-accent border border-accent/20 bg-accent/10 px-2 py-0.5 rounded-full uppercase font-bold">Read Only</span>
              </p>
            </div>
            <div className="bg-white/5 rounded-xl p-3 border border-white/5">
              <p className="text-[10px] uppercase tracking-widest text-muted-foreground font-bold mb-1">Account Status</p>
              <p className="text-sm text-white font-medium flex items-center gap-2">
                <Check className="w-4 h-4 text-green-400" /> Active
              </p>
            </div>
          </div>
        </motion.div>

        {/* Security / Password Card */}
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="col-span-1 md:col-span-2 glass rounded-3xl p-8 border-white/10"
        >
          <div className="flex items-center gap-3 mb-6 pb-6 border-b border-white/10">
            <KeyRound className="w-6 h-6 text-accent" />
            <h2 className="text-xl font-black text-white">Security & Password</h2>
          </div>

          <form onSubmit={handlePasswordUpdate} className="space-y-6 max-w-md">
            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest ml-1">Current Password</label>
              <input
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                className="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:ring-2 focus:ring-accent/50 transition-all"
                placeholder="Enter current password"
                required
              />
              <p className="text-[10px] text-muted-foreground ml-1">For demo, use: &apos;password&apos;</p>
            </div>

            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest ml-1">New Password</label>
              <input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:ring-2 focus:ring-accent/50 transition-all"
                placeholder="Enter new password"
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-[10px] font-black text-muted-foreground uppercase tracking-widest ml-1">Confirm New Password</label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:ring-2 focus:ring-accent/50 transition-all"
                placeholder="Confirm new password"
                required
              />
            </div>

            <button
              type="submit"
              disabled={isUpdating}
              className="px-6 py-3 bg-accent text-[#030014] font-black rounded-xl hover:scale-105 active:scale-95 transition-all shadow-[0_0_20px_rgba(var(--accent),0.3)] disabled:opacity-50 disabled:cursor-not-allowed uppercase tracking-widest text-xs flex items-center justify-center gap-2"
            >
              {isUpdating ? "Updating..." : "Update Password"}
            </button>
          </form>

          <div className="mt-8 pt-6 border-t border-white/5 flex gap-4 p-4 bg-orange-500/10 border border-orange-500/20 rounded-xl">
            <AlertCircle className="w-5 h-5 text-orange-400 shrink-0" />
            <p className="text-xs text-orange-200/80 leading-relaxed">
              <strong>Note:</strong> Password changes require re-authentication on all active sessions. 
              If your account is managed via Single Sign-On (SSO), this panel may be read-only depending on your organization&apos;s policy.
            </p>
          </div>
        </motion.div>
      </div>
    </div>
  );
}
