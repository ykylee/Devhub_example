"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  Users, 
  Network, 
  Shield, 
  Box, 
  Inbox, 
  Key, 
  Plug, 
  Link2, 
  FileText, 
  ChevronDown,
  ChevronLeft,
  ChevronRight
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useStore } from "@/lib/store";
import { defaultLandingFor, isSystemAdmin } from "@/lib/auth/role-routing";

const categories = [
  {
    id: "access-control",
    label: "Access Control",
    items: [
      { href: "/admin/settings/users", label: "Users", icon: Users },
      { href: "/admin/settings/organization", label: "Organization", icon: Network },
      { href: "/admin/settings/permissions", label: "Permissions", icon: Shield },
    ],
  },
  {
    id: "app-requests",
    label: "App & Requests",
    items: [
      { href: "/admin/settings/applications", label: "Applications", icon: Box },
      { href: "/admin/settings/dev-requests", label: "Dev Requests", icon: Inbox },
      { href: "/admin/settings/dev-request-tokens", label: "Intake Tokens", icon: Key },
    ],
  },
  {
    id: "integrations-audit",
    label: "Integrations & Audit",
    items: [
      { href: "/admin/settings/integrations", label: "Integrations", icon: Plug },
      { href: "/admin/settings/integration-bindings", label: "Bindings", icon: Link2 },
      { href: "/admin/settings/audit", label: "Audit", icon: FileText },
    ],
  },
];

const allTabs = categories.flatMap((cat) => cat.items);

export default function AdminSettingsLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const actor = useStore((s) => s.actor);
  const allowed = isSystemAdmin(actor?.role);
  const [isOpen, setIsOpen] = useState(false);
  const [isCollapsed, setIsCollapsed] = useState(false);

  // SSR-safe localStorage hydration: server 에서는 default(false) 로 render,
  // client mount 이후 저장된 값으로 sync. setState-in-effect 룰의 cascading
  // render 우려는 mount-only effect 라 해당 없음.
  useEffect(() => {
    if (typeof window !== "undefined") {
      const saved = localStorage.getItem("admin-settings-sidebar-collapsed");
      if (saved !== null) {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setIsCollapsed(saved === "true");
      }
    }
  }, []);

  const handleToggleCollapse = () => {
    const nextVal = !isCollapsed;
    setIsCollapsed(nextVal);
    if (typeof window !== "undefined") {
      localStorage.setItem("admin-settings-sidebar-collapsed", String(nextVal));
    }
  };

  // Defense-in-depth: AuthGuard already gates /admin/* on actor.role, but
  // this re-check stops a stale render path from leaking the layout shell.
  useEffect(() => {
    if (actor && !allowed) {
      router.replace(defaultLandingFor(actor.role));
    }
  }, [actor, allowed, router]);

  if (!allowed) return null;

  // Active Tab/Category Resolution
  const activeTab = allTabs.find(
    (tab) => pathname === tab.href || pathname.startsWith(`${tab.href}/`)
  ) || allTabs[0];

  const activeCategory = categories.find(
    (cat) => cat.items.some((item) => item.href === activeTab.href)
  )?.label || "Access Control";

  return (
    <div className="pb-20 flex flex-col h-full">
      {/* Mobile Title Block */}
      <div className="md:hidden mb-5">
        <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }}>
          <h1 className="text-3xl font-black text-foreground dark:text-primary-foreground tracking-tighter uppercase">
            System <span className="text-accent">Settings</span>
          </h1>
          <p className="text-muted-foreground font-bold text-[10px] uppercase tracking-widest mt-1.5">
            system_admin only · {activeCategory}
          </p>
        </motion.div>
      </div>

      {/* Mobile Glassmorphism Dropdown Selector */}
      <div className="md:hidden relative mb-6 z-30">
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="w-full flex items-center justify-between px-5 py-4 glass border border-border/80 rounded-2xl shadow-lg transition-all duration-300 hover:border-accent/40 active:scale-[0.98]"
        >
          <div className="flex items-center gap-3">
            <div className="p-2 bg-accent/10 rounded-xl text-accent">
              <activeTab.icon className="w-5 h-5" />
            </div>
            <div className="text-left">
              <p className="text-[9px] font-black text-muted-foreground/60 uppercase tracking-widest leading-none mb-1">
                {activeCategory}
              </p>
              <p className="text-sm font-black text-foreground uppercase tracking-widest leading-none">
                {activeTab.label}
              </p>
            </div>
          </div>
          <motion.div
            animate={{ rotate: isOpen ? 180 : 0 }}
            transition={{ type: "spring", stiffness: 200, damping: 15 }}
            className="text-muted-foreground"
          >
            <ChevronDown className="w-5 h-5" />
          </motion.div>
        </button>

        <AnimatePresence>
          {isOpen && (
            <motion.div
              initial={{ opacity: 0, y: -15, scale: 0.98 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: -15, scale: 0.98 }}
              transition={{ duration: 0.2, ease: "easeOut" }}
              className="absolute top-[calc(100%+8px)] left-0 right-0 glass border border-border rounded-2xl shadow-2xl z-50 p-4 backdrop-blur-xl bg-background/90"
            >
              <div className="flex flex-col gap-4 max-h-[60vh] overflow-y-auto pr-1">
                {categories.map((category) => (
                  <div key={category.id} className="space-y-2">
                    <p className="text-[9px] font-black text-muted-foreground/50 uppercase tracking-widest px-2.5">
                      {category.label}
                    </p>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-1">
                      {category.items.map((tab) => {
                        const isActive = pathname === tab.href || pathname.startsWith(`${tab.href}/`);
                        return (
                          <Link
                            key={tab.href}
                            href={tab.href}
                            onClick={() => setIsOpen(false)}
                            className={cn(
                              "flex items-center gap-3 px-4 py-3 rounded-xl text-xs font-black uppercase tracking-widest transition-all duration-200 relative border",
                              isActive
                                ? "bg-accent/15 text-accent border-accent/25"
                                : "text-muted-foreground hover:bg-muted/30 hover:text-foreground border-transparent",
                            )}
                          >
                            <tab.icon className={cn("w-4 h-4", isActive ? "text-accent" : "text-muted-foreground")} />
                            <span>{tab.label}</span>
                          </Link>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Grid Layout */}
      <div className="flex flex-col md:grid md:grid-cols-12 gap-8 items-start w-full">
        {/* Navigation Sidebar */}
        <aside className={cn(
          "hidden md:flex flex-col gap-6 w-full transition-all duration-300",
          isCollapsed ? "md:col-span-1" : "md:col-span-3"
        )}>
          {/* Desktop Title Block */}
          <div className={cn("mb-2 flex items-center gap-3 w-full", isCollapsed ? "justify-center" : "justify-between")}>
            {!isCollapsed && (
              <motion.div initial={{ opacity: 0, x: -10 }} animate={{ opacity: 1, x: 0 }} transition={{ duration: 0.2 }}>
                <h1 className="text-2xl font-black text-foreground dark:text-primary-foreground tracking-tighter uppercase leading-tight">
                  System <span className="text-accent">Settings</span>
                </h1>
                <p className="text-muted-foreground font-bold text-[9px] uppercase tracking-widest mt-1.5 leading-relaxed">
                  System Control Console
                </p>
              </motion.div>
            )}
            <button
              onClick={handleToggleCollapse}
              className="p-2.5 glass border border-border/80 hover:border-accent/40 rounded-xl text-muted-foreground hover:text-accent transition-all duration-300 active:scale-95 shadow-md flex items-center justify-center"
              title={isCollapsed ? "Expand Sidebar" : "Collapse Sidebar"}
            >
              {isCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
            </button>
          </div>

          <div className="flex flex-col gap-5 w-full">
            {categories.map((category) => (
              <div key={category.id} className="flex flex-col gap-2 w-full">
                {isCollapsed ? (
                  <div className="border-t border-border/40 my-1 mx-2" />
                ) : (
                  <p className="text-[10px] font-black text-muted-foreground/50 uppercase tracking-widest px-3.5 mb-0.5">
                    {category.label}
                  </p>
                )}
                <nav className={cn(
                  "flex flex-col p-1.5 glass border-border/60 rounded-2xl gap-0.5 w-full overflow-hidden transition-all duration-300",
                  isCollapsed ? "items-center" : ""
                )}>
                  {category.items.map((tab) => {
                    const isActive = pathname === tab.href || pathname.startsWith(`${tab.href}/`);
                    return (
                      <Link
                        key={tab.href}
                        href={tab.href}
                        title={tab.label}
                        className={cn(
                          "flex items-center gap-3 rounded-xl text-xs font-black uppercase tracking-widest transition-all duration-300 relative group overflow-hidden w-full",
                          isCollapsed ? "p-3.5 justify-center" : "px-4 py-3",
                          isActive
                            ? "text-foreground dark:text-primary-foreground"
                            : "text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground",
                        )}
                      >
                        {isActive && (
                          <motion.div
                            layoutId="settings-sub-tab-active"
                            className="absolute inset-0 bg-muted/40 border border-border/80 rounded-xl"
                            transition={{ type: "spring", bounce: 0.15, duration: 0.5 }}
                          />
                        )}
                        <tab.icon className={cn("w-4 h-4 relative z-10 transition-colors duration-300", isActive ? "text-accent" : "text-muted-foreground group-hover:text-foreground")} />
                        {!isCollapsed && <span className="relative z-10">{tab.label}</span>}
                      </Link>
                    );
                  })}
                </nav>
              </div>
            ))}
          </div>
        </aside>

        {/* Content Area */}
        <main className={cn(
          "w-full transition-all duration-300",
          isCollapsed ? "md:col-span-11" : "md:col-span-9"
        )}>
          <motion.div
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, ease: "easeOut" }}
            className="w-full"
          >
            {children}
          </motion.div>
        </main>
      </div>
    </div>
  );
}
