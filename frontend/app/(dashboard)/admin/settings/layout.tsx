"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { motion } from "framer-motion";
import { Users, Network, Shield } from "lucide-react";
import { cn } from "@/lib/utils";
import { useStore } from "@/lib/store";
import { defaultLandingFor, isSystemAdmin } from "@/lib/auth/role-routing";

const subTabs = [
  { href: "/admin/settings/users", label: "Users", icon: Users },
  { href: "/admin/settings/organization", label: "Organization", icon: Network },
  { href: "/admin/settings/permissions", label: "Permissions", icon: Shield },
  /* 데모 중 혼선을 피하기 위해 아래 프로토타입 메뉴들은 노출에서 제외합니다.
  { href: "/admin/settings/applications", label: "Applications", icon: Box },
  { href: "/admin/settings/dev-requests", label: "Dev Requests", icon: Inbox },
  { href: "/admin/settings/dev-request-tokens", label: "Intake Tokens", icon: Key },
  { href: "/admin/settings/integrations", label: "Integrations", icon: Plug },
  { href: "/admin/settings/integration-bindings", label: "Bindings", icon: Link2 },
  { href: "/admin/settings/audit", label: "Audit", icon: FileText },
  */
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
    <div className="pb-20 flex flex-col h-full">
      {/* Mobile Title Block */}
      <div className="md:hidden mb-6">
        <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }}>
          <h1 className="text-3xl font-black text-foreground dark:text-primary-foreground tracking-tighter uppercase">
            System <span className="text-accent">Settings</span>
          </h1>
          <p className="text-muted-foreground font-bold text-[10px] uppercase tracking-widest mt-1.5">
            system_admin only · Access Control
          </p>
        </motion.div>
      </div>

      <div className="flex flex-col md:grid md:grid-cols-12 gap-8 items-start">
        {/* Navigation Sidebar/Top Header */}
        <aside className="w-full md:col-span-3 flex flex-col gap-4">
          {/* Desktop Title Block */}
          <div className="hidden md:block mb-4">
            <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }}>
              <h1 className="text-2xl font-black text-foreground dark:text-primary-foreground tracking-tighter uppercase leading-tight">
                System <span className="text-accent">Settings</span>
              </h1>
              <p className="text-muted-foreground font-bold text-[10px] uppercase tracking-widest mt-2 leading-relaxed">
                Access Control Console
              </p>
            </motion.div>
          </div>

          <div className="hidden md:block">
            <p className="text-[10px] font-black text-muted-foreground/60 uppercase tracking-widest mb-1 px-3">
              Access Control
            </p>
          </div>

          <nav className="flex md:flex-col p-1.5 glass border-border rounded-2xl gap-1 w-full overflow-hidden">
            {subTabs.map((tab) => {
              const isActive = pathname === tab.href || pathname.startsWith(`${tab.href}/`);
              return (
                <Link
                  key={tab.href}
                  href={tab.href}
                  className={cn(
                    "flex flex-1 md:flex-initial items-center justify-center md:justify-start gap-2.5 px-4 py-3 rounded-xl text-xs font-black uppercase tracking-widest transition-all relative",
                    isActive 
                      ? "text-foreground dark:text-primary-foreground" 
                      : "text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground",
                  )}
                >
                  {isActive && (
                    <motion.div
                      layoutId="settings-sub-tab"
                      className="absolute inset-0 bg-muted/40 border border-border/80 rounded-xl"
                      transition={{ type: "spring", bounce: 0.15, duration: 0.5 }}
                    />
                  )}
                  <tab.icon className={cn("w-4 h-4 relative z-10 transition-colors", isActive ? "text-accent" : "text-muted-foreground")} />
                  <span className="relative z-10 hidden sm:inline md:inline">{tab.label}</span>
                </Link>
              );
            })}
          </nav>
        </aside>

        {/* Content Area */}
        <main className="w-full md:col-span-9">
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
