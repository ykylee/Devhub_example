"use client";
 
import Link from "next/link";
import { usePathname } from "next/navigation";
import { LayoutDashboard, Users, Server, Settings, Zap, ShieldCheck, X, ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion, AnimatePresence } from "framer-motion";
import { useStore } from "@/lib/store";
import { isSystemAdmin } from "@/lib/auth/role-routing";
import { useState, useEffect } from "react";
 
interface MenuItem {
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  color: string;
}
 
const baseMenu: MenuItem[] = [
  { href: "/developer", icon: LayoutDashboard, label: "Work Status", color: "text-info" },
  { href: "/manager", icon: Users, label: "Quality Status", color: "text-success" },
  { href: "/applications", icon: Zap, label: "Applications", color: "text-purple-400" },
  { href: "/repositories", icon: Server, label: "Repositories", color: "text-cyan-400" },
  { href: "/projects", icon: Settings, label: "Projects", color: "text-pink-400" },
];
 
const systemMenu: MenuItem[] = [
  { href: "/admin", icon: Server, label: "Sys Admin Dashboard", color: "text-accent" },
];
 
const systemBottomMenu: MenuItem = {
  href: "/admin/settings",
  icon: Settings,
  label: "Sys Admin Settings",
  color: "text-accent",
};
 
export function Sidebar({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  const pathname = usePathname();
  const { actor, isSidebarOpen, setSidebarOpen, isSidebarCollapsed, setSidebarCollapsed } = useStore();
  const [mounted, setMounted] = useState(false);
 
  useEffect(() => {
    setMounted(true);
  }, []);
 
  const collapsed = mounted ? isSidebarCollapsed : false;
  const showSystem = isSystemAdmin(actor?.role);
 
  return (
    <>
      {/* Mobile Backdrop */}
      <AnimatePresence>
        {isSidebarOpen && (
          <motion.div 
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setSidebarOpen(false)}
            className="fixed inset-0 bg-background/80 backdrop-blur-sm z-[140] lg:hidden"
            aria-hidden="true"
          />
        )}
      </AnimatePresence>
 
      {/* z-[150] only when the mobile drawer is *active*. On lg+ (sticky desktop sidebar),
          drop to z-[10] so the Header dropdown menu (z-50) stacks above it — otherwise
          desktop user_id click → dropdown menu never appears in front. */}
      <aside className={cn(
        "glass border-r border-border h-screen flex flex-col transition-all duration-300",
        "fixed inset-y-0 left-0 w-72 lg:sticky lg:top-0 lg:translate-x-0",
        collapsed ? "lg:w-20" : "lg:w-64",
        "z-[150] lg:z-[10]",
        isSidebarOpen ? "translate-x-0 shadow-2xl" : "-translate-x-full lg:translate-x-0",
        className
      )} {...props}>
        <div className="p-6 flex flex-col h-full overflow-y-auto">
          <div className={cn("flex items-center justify-between mb-10 px-2", collapsed && "lg:justify-center lg:px-0")}>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-gradient-to-br from-primary to-accent rounded-xl shadow-lg flex items-center justify-center ring-2 ring-border/70 shrink-0">
                <Zap className="w-6 h-6 text-primary-foreground fill-current" aria-hidden="true" />
              </div>
              <span className={cn(
                "text-2xl font-bold tracking-tighter text-gradient transition-all duration-300 origin-left",
                collapsed ? "lg:opacity-0 lg:w-0 lg:scale-0 lg:pointer-events-none" : "opacity-100 w-auto scale-100"
              )}>DevHub</span>
            </div>
            <button 
              onClick={() => setSidebarOpen(false)}
              className="p-2 rounded-xl hover:bg-muted/30 text-muted-foreground lg:hidden"
              aria-label="Close sidebar"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
 
          <nav className="space-y-2 flex-1">
            {collapsed ? (
              <div className="border-t border-border/30 my-4" />
            ) : (
              <p className="px-4 text-[10px] font-bold text-muted-foreground uppercase tracking-[0.2em] mb-4 opacity-50">
                Main Navigation
              </p>
            )}
            {baseMenu.map((item) => renderMenuItem(item, pathname, collapsed, () => setSidebarOpen(false)))}
 
            {showSystem && (
              <>
                {collapsed ? (
                  <div className="border-t border-border/30 my-4" />
                ) : (
                  <p className="px-4 pt-4 text-[10px] font-bold text-muted-foreground uppercase tracking-[0.2em] mb-2 opacity-50 flex items-center gap-2">
                    <ShieldCheck className="w-3 h-3 text-accent" aria-hidden="true" />
                    System (Admin only)
                  </p>
                )}
                {systemMenu.map((item) => renderMenuItem(item, pathname, collapsed, () => setSidebarOpen(false)))}
              </>
            )}
          </nav>
 
          <div className="mt-auto pt-6 border-t border-border/60">
            {showSystem && (
              <div className="mb-3">
                {renderMenuItem(systemBottomMenu, pathname, collapsed, () => setSidebarOpen(false))}
              </div>
            )}
            <Link href="/account" onClick={() => setSidebarOpen(false)} aria-label="Account Settings" title={collapsed ? "Account" : undefined}>
              <div className={cn(
                "flex items-center gap-3 rounded-xl px-4 py-3 text-sm font-medium text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground hover:bg-muted/30 transition-all",
                collapsed && "lg:justify-center lg:px-0"
              )}>
                <Settings className="h-5 w-5 shrink-0" aria-hidden="true" />
                <span className={cn(
                  "transition-all duration-300 origin-left",
                  collapsed ? "lg:opacity-0 lg:w-0 lg:scale-0 lg:pointer-events-none" : "opacity-100 w-auto scale-100"
                )}>Account</span>
              </div>
            </Link>
            
            {collapsed ? (
              <div className="mt-4 flex justify-center py-2 glass rounded-xl border border-border/60 animate-fade-in" title="System Online (v0.1.0)">
                <div className="w-2.5 h-2.5 bg-success rounded-full animate-pulse" aria-hidden="true" />
              </div>
            ) : (
              <div className="mt-4 px-4 py-3 glass rounded-xl border border-border/60 text-[10px] text-muted-foreground flex items-center justify-between">
                <span className="flex items-center gap-2">
                  <div className="w-1.5 h-1.5 bg-success rounded-full animate-pulse" aria-hidden="true" />
                  System Online
                </span>
                <span className="opacity-50 italic">v0.1.0</span>
              </div>
            )}
 
            {/* Collapsible Sidebar Toggle Button */}
            <div className="hidden lg:flex justify-end pt-4">
              <button
                onClick={() => setSidebarCollapsed(!collapsed)}
                className="w-8 h-8 rounded-full border border-border bg-background/50 hover:bg-muted/50 backdrop-blur-md flex items-center justify-center text-muted-foreground hover:text-foreground shadow-sm transition-all focus:outline-none"
                aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
                title={collapsed ? "Expand sidebar" : "Collapse sidebar"}
              >
                {collapsed ? (
                  <ChevronRight className="w-4 h-4" />
                ) : (
                  <ChevronLeft className="w-4 h-4" />
                )}
              </button>
            </div>
          </div>
        </div>
      </aside>
    </>
  );
}
 
function renderMenuItem(item: MenuItem, pathname: string, collapsed: boolean, onClick?: () => void) {
  const isActive =
    item.href === "/admin"
      ? pathname === "/admin"
      : pathname === item.href || pathname.startsWith(`${item.href}/`);
 
  return (
    <Link key={item.href} href={item.href} onClick={onClick} aria-label={item.label} title={collapsed ? item.label : undefined}>
      <motion.div
        whileHover={collapsed ? {} : { x: 4 }}
        whileTap={{ scale: 0.98 }}
        className={cn(
          "relative flex items-center gap-3 rounded-xl px-4 py-3 text-sm font-medium transition-all group overflow-hidden",
          isActive
            ? "text-primary dark:text-primary-foreground bg-primary/10 dark:bg-muted/40 border border-primary/20 dark:border-border"
            : "text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground hover:bg-muted/30",
          collapsed && "lg:justify-center lg:px-0"
        )}
      >
        {isActive && (
          <motion.div
            layoutId="active-pill"
            className="absolute inset-0 bg-gradient-to-r from-primary/10 to-accent/10 dark:from-primary/20 dark:to-accent/20 opacity-50"
            transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
          />
        )}
        <item.icon className={cn("h-5 w-5 transition-colors z-10 shrink-0", isActive ? item.color : "group-hover:text-foreground dark:group-hover:text-primary-foreground")} aria-hidden="true" />
        <span className={cn(
          "z-10 transition-all duration-300 origin-left",
          collapsed ? "lg:opacity-0 lg:w-0 lg:scale-0 lg:pointer-events-none" : "opacity-100 w-auto scale-100"
        )}>{item.label}</span>
      </motion.div>
    </Link>
  );
}
