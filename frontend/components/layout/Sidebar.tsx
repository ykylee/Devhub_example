"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { LayoutDashboard, Users, Server, Settings, Zap, Network, ShieldCheck } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import { useStore } from "@/lib/store";
import { isSystemAdmin } from "@/lib/auth/role-routing";

interface MenuItem {
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  color: string;
}

const baseMenu: MenuItem[] = [
  { href: "/developer", icon: LayoutDashboard, label: "Developer", color: "text-blue-400" },
  { href: "/manager", icon: Users, label: "Manager", color: "text-emerald-400" },
];

const systemMenu: MenuItem[] = [
  { href: "/admin", icon: Server, label: "Sys Admin Dashboard", color: "text-orange-400" },
  { href: "/admin/settings", icon: Settings, label: "Sys Admin Settings", color: "text-orange-400" },
  // /organization is folded under Sys Admin until PR-S2 retires the route
  // in favour of /admin/settings/organization.
  { href: "/organization", icon: Network, label: "Organization", color: "text-purple-400" },
];

export function Sidebar({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  const pathname = usePathname();
  const actor = useStore((s) => s.actor);
  const showSystem = isSystemAdmin(actor?.role);

  return (
    <aside className={cn("glass border-r border-white/10 w-64 h-screen sticky top-0 flex flex-col z-40", className)} {...props}>
      <div className="p-6 flex flex-col h-full">
        <div className="flex items-center gap-3 mb-10 px-2">
          <div className="w-10 h-10 bg-gradient-to-br from-primary to-accent rounded-xl shadow-lg flex items-center justify-center ring-2 ring-white/10">
            <Zap className="w-6 h-6 text-white fill-current" />
          </div>
          <span className="text-2xl font-bold tracking-tighter text-gradient">DevHub</span>
        </div>

        <nav className="space-y-2 flex-1">
          <p className="px-4 text-[10px] font-bold text-muted-foreground uppercase tracking-[0.2em] mb-4 opacity-50">
            Main Navigation
          </p>
          {baseMenu.map((item) => renderMenuItem(item, pathname))}

          {showSystem && (
            <>
              <p className="px-4 pt-4 text-[10px] font-bold text-muted-foreground uppercase tracking-[0.2em] mb-2 opacity-50 flex items-center gap-2">
                <ShieldCheck className="w-3 h-3 text-orange-400" />
                System (Admin only)
              </p>
              {systemMenu.map((item) => renderMenuItem(item, pathname))}
            </>
          )}
        </nav>

        <div className="mt-auto pt-6 border-t border-white/5">
          <Link href="/account">
            <div className="flex items-center gap-3 rounded-xl px-4 py-3 text-sm font-medium text-muted-foreground hover:text-foreground dark:hover:text-white hover:bg-white/5 transition-all">
              <Settings className="h-5 w-5" />
              <span>Account</span>
            </div>
          </Link>
          <div className="mt-4 px-4 py-3 glass rounded-xl border border-white/5 text-[10px] text-muted-foreground flex items-center justify-between">
            <span className="flex items-center gap-2">
              <div className="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse" />
              System Online
            </span>
            <span className="opacity-50 italic">v0.1.0</span>
          </div>
        </div>
      </div>
    </aside>
  );
}

function renderMenuItem(item: MenuItem, pathname: string) {
  // Active state matches /admin/settings exactly (not /admin) so the two
  // system items don't both light up when the user is on /admin/settings.
  const isActive =
    item.href === "/admin"
      ? pathname === "/admin"
      : pathname === item.href || pathname.startsWith(`${item.href}/`);

  return (
    <Link key={item.href} href={item.href}>
      <motion.div
        whileHover={{ x: 4 }}
        whileTap={{ scale: 0.98 }}
        className={cn(
          "relative flex items-center gap-3 rounded-xl px-4 py-3 text-sm font-medium transition-all group overflow-hidden",
          isActive
            ? "text-primary dark:text-white bg-primary/10 dark:bg-white/10 border border-primary/20 dark:border-white/10"
            : "text-muted-foreground hover:text-foreground dark:hover:text-white hover:bg-white/5"
        )}
      >
        {isActive && (
          <motion.div
            layoutId="active-pill"
            className="absolute inset-0 bg-gradient-to-r from-primary/10 to-accent/10 dark:from-primary/20 dark:to-accent/20 opacity-50"
            transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
          />
        )}
        <item.icon className={cn("h-5 w-5 transition-colors z-10", isActive ? item.color : "group-hover:text-foreground dark:group-hover:text-white")} />
        <span className="z-10">{item.label}</span>
      </motion.div>
    </Link>
  );
}
