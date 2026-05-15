"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { motion } from "framer-motion";
import { Users, Network, Shield, FileText, Box, Inbox, Key } from "lucide-react";
import { cn } from "@/lib/utils";
import { useStore } from "@/lib/store";
import { defaultLandingFor, isSystemAdmin } from "@/lib/auth/role-routing";

const subTabs = [
  { href: "/admin/settings/users", label: "Users", icon: Users },
  { href: "/admin/settings/organization", label: "Organization", icon: Network },
  { href: "/admin/settings/permissions", label: "Permissions", icon: Shield },
  { href: "/admin/settings/applications", label: "Applications", icon: Box },
  { href: "/admin/settings/dev-requests", label: "Dev Requests", icon: Inbox },
  { href: "/admin/settings/dev-request-tokens", label: "Intake Tokens", icon: Key },
  { href: "/admin/settings/audit", label: "Audit", icon: FileText },
];

export default function AdminSettingsLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const actor = useStore((s) => s.actor);
  const allowed = isSystemAdmin(actor?.role);

  // Defence-in-depth: AuthGuard already gates /admin/* on actor.role, but
  // this re-check stops a stale render path from leaking the layout shell.
  useEffect(() => {
    if (actor && !allowed) {
      router.replace(defaultLandingFor(actor.role));
    }
  }, [actor, allowed, router]);

  if (!allowed) return null;

  return (
    <div className="space-y-10 pb-20 flex flex-col h-full">
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }}>
          <h1 className="text-3xl font-black text-foreground dark:text-primary-foreground tracking-tighter uppercase">
            System <span className="text-orange-400">Settings</span>
          </h1>
          <p className="text-muted-foreground font-bold text-xs uppercase tracking-widest mt-2">
            system_admin only · users / organization / permissions
          </p>
        </motion.div>

        <nav className="flex p-1.5 glass border-border rounded-2xl gap-1">
          {subTabs.map((tab) => {
            const isActive = pathname === tab.href || pathname.startsWith(`${tab.href}/`);
            return (
              <Link
                key={tab.href}
                href={tab.href}
                className={cn(
                  "flex items-center gap-2 px-5 py-2.5 rounded-xl text-xs font-black uppercase tracking-widest transition-all relative",
                  isActive ? "text-foreground dark:text-primary-foreground" : "text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground",
                )}
              >
                {isActive && (
                  <motion.div
                    layoutId="settings-sub-tab"
                    className="absolute inset-0 bg-muted/40 border border-border rounded-xl"
                    transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                  />
                )}
                <tab.icon className={cn("w-4 h-4 relative z-10", isActive ? "text-orange-400" : "text-muted-foreground")} />
                <span className="relative z-10">{tab.label}</span>
              </Link>
            );
          })}
        </nav>
      </div>

      <div className="flex-1">{children}</div>
    </div>
  );
}
