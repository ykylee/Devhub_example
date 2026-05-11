"use client";

import { motion } from "framer-motion";
import { Construction, Settings as SettingsIcon } from "lucide-react";

// Placeholder until PR-S2 lands the real /admin/settings layout (users /
// organization / permissions sub-routes). Sidebar already routes here for
// system_admin users in PR-S1; landing on a useful page is PR-S2's job.
export default function AdminSettingsPlaceholderPage() {
  return (
    <div className="min-h-[60vh] flex items-center justify-center p-8">
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        className="glass border-white/10 rounded-3xl p-10 max-w-xl w-full text-center space-y-6 shadow-2xl"
      >
        <div className="inline-flex p-4 rounded-2xl bg-orange-500/10 border border-orange-500/20">
          <Construction className="w-8 h-8 text-orange-400" />
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-black text-foreground dark:text-white tracking-tighter uppercase">
            System <span className="text-orange-400">Settings</span>
          </h1>
          <p className="text-sm text-muted-foreground">
            User management, organization, and permission consoles will move under
            this route in PR-S2. For now they live at the legacy <code className="font-mono text-xs px-1 py-0.5 rounded bg-white/10">/organization</code> page.
          </p>
        </div>
        <div className="flex items-center gap-3 justify-center text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
          <SettingsIcon className="w-3 h-3" />
          system_admin only
        </div>
      </motion.div>
    </div>
  );
}
