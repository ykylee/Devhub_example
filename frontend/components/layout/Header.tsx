"use client";

import { useState, useEffect } from "react";
import { Search, Bell, User, ChevronDown, Command } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion, AnimatePresence } from "framer-motion";

import { useStore, type UserRole } from "@/lib/store";
import { useRouter } from "next/navigation";
import { authService } from "@/lib/services/auth.service";
import { realtimeService, type ConnectionStatusEvent } from "@/lib/services/realtime.service";
import { ThemeToggle } from "./ThemeToggle";

export function Header({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  const { role, actor, setRole, notifications, clearNotifications } = useStore();
  const [showDropdown, setShowDropdown] = useState(false);
  const [isConnected, setIsConnected] = useState(realtimeService.isConnected);
  const router = useRouter();

  useEffect(() => {
    const unsubscribe = realtimeService.subscribe<ConnectionStatusEvent>('status.changed', (event) => {
      setIsConnected(event.data.connected);
    });
    return () => unsubscribe();
  }, []);

  const handleRoleChange = (newRole: UserRole) => {
    setRole(newRole);
    setShowDropdown(false);
    
    // Role-based navigation
    const pathMap: Record<UserRole, string> = {
      "Developer": "/developer",
      "Manager": "/manager",
      "System Admin": "/admin"
    };
    router.push(pathMap[newRole]);
  };

  const handleLogout = async () => {
    setShowDropdown(false);
    await authService.logout();
  };

  return (
    <header className={cn("sticky top-0 z-50 w-full glass border-b border-border/60", className)} {...props}>
      <div className="flex h-16 items-center px-8 gap-8">
        <div className="flex-1 flex items-center gap-4">
          <div className="flex items-center gap-2 glass border-border px-3 py-1.5 rounded-xl">
            <div className={cn(
              "w-2 h-2 rounded-full animate-pulse",
              isConnected ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" : "bg-rose-500 shadow-[0_0_8px_rgba(244,63,94,0.5)]"
            )} />
            <span className="text-[10px] font-black text-muted-foreground dark:text-primary-foreground/50 uppercase tracking-widest hidden lg:inline">
              {isConnected ? "Real-time Live" : "Offline"}
            </span>
          </div>
          <div className="relative w-full max-w-lg hidden md:flex items-center group">
            <div className="absolute left-3.5 flex items-center gap-2 pointer-events-none">
              <Search className="h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
              <div className="px-1.5 py-0.5 rounded border border-border bg-muted/30 text-[10px] font-mono text-muted-foreground">
                <Command className="w-2 h-2 inline mr-0.5" /> K
              </div>
            </div>
            <input
              type="search"
              placeholder="Search anything..."
              className="flex h-10 w-full rounded-xl border border-border/60 bg-muted/30 px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 focus:bg-muted/40 transition-all pl-24"
            />
          </div>
        </div>
        
        <div className="flex items-center gap-6">
          <ThemeToggle />
          <motion.button 
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={clearNotifications}
            className="relative p-2.5 rounded-xl hover:bg-muted/30 text-muted-foreground hover:text-primary-foreground transition-all"
          >
            <Bell className="h-5 w-5" />
            {notifications > 0 && (
              <span className="absolute top-2 right-2 w-2 h-2 bg-accent rounded-full border-2 border-background"></span>
            )}
          </motion.button>
          
          <div className="h-6 w-px bg-muted/40"></div>
          
          <div className="relative">
            <motion.div 
              onClick={() => setShowDropdown(!showDropdown)}
              className="flex items-center gap-3 py-1.5 px-3 rounded-2xl hover:bg-muted/30 transition-all cursor-pointer group"
            >
              <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-primary/20 to-accent/20 flex items-center justify-center border border-border ring-2 ring-primary/20">
                <User className="w-5 h-5 text-primary-foreground" />
              </div>
              <div className="flex flex-col hidden sm:flex">
                <span className="text-sm font-semibold leading-none text-foreground dark:text-primary-foreground">{actor?.login || "Guest User"}</span>
                <span className="text-[10px] font-bold text-muted-foreground mt-1 flex items-center gap-1 uppercase tracking-wider">
                  {role || "No Role"} <ChevronDown className={cn("w-3 h-3 transition-transform duration-300", showDropdown && "rotate-180")} />
                </span>
              </div>
            </motion.div>

            <AnimatePresence>
              {showDropdown && (
                <motion.div
                  initial={{ opacity: 0, y: 10, scale: 0.95 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: 10, scale: 0.95 }}
                  className="absolute top-full right-0 mt-4 w-56 rounded-2xl glass border border-border p-2 z-50 shadow-2xl"
                >
                  <p className="px-3 pt-2 text-[10px] font-bold text-muted-foreground uppercase tracking-widest opacity-50">
                    Switch View
                  </p>
                  <p className="px-3 pb-2 text-[9px] text-muted-foreground/60 leading-tight normal-case">
                    Menu preview only — actual permissions follow server actor.role.
                  </p>
                  {(["Developer", "Manager", "System Admin"] as UserRole[]).map((r) => (
                    <button
                      key={r}
                      onClick={() => handleRoleChange(r)}
                      className={cn(
                        "flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-all hover:bg-muted/40 group",
                        role === r ? "text-primary bg-primary/10" : "text-muted-foreground hover:text-primary-foreground"
                      )}
                    >
                      <div className={cn("w-2 h-2 rounded-full transition-all", role === r ? "bg-primary scale-100" : "bg-transparent scale-0")} />
                      {r}
                    </button>
                  ))}
                  <div className="h-px bg-muted/30 my-2" />
                  <button 
                    onClick={() => { router.push("/account"); setShowDropdown(false); }}
                    className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-muted-foreground hover:text-primary-foreground hover:bg-muted/40 transition-all"
                  >
                    Account Settings
                  </button>
                  <button 
                    onClick={handleLogout}
                    className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-rose-400 hover:bg-rose-400/10 transition-all"
                  >
                    Sign Out
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>
    </header>
  );
}
