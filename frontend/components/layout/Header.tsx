"use client";

import { useState, useEffect } from "react";
import { Search, Bell, User, ChevronDown, Command, Sun, Moon, Settings, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion, AnimatePresence } from "framer-motion";

import { useStore, type UserRole } from "@/lib/store";
import { useRouter } from "next/navigation";
import { authService } from "@/lib/services/auth.service";
import { realtimeService, type ConnectionStatusEvent } from "@/lib/services/realtime.service";

export function Header({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  const { role, actor, setRole, notifications, clearNotifications } = useStore();
  const [showDropdown, setShowDropdown] = useState(false);
  const [isConnected, setIsConnected] = useState(realtimeService.isConnected);
  // 초기 theme 은 paint 전에 layout 의 inline script 가 html 에 적용하므로
  // 여기서는 그 결과(`theme-dark` class 유무)를 읽어 state 와 일치시킨다.
  const [theme, setTheme] = useState<"light" | "dark">(() => {
    if (typeof document === "undefined") return "light";
    return document.documentElement.classList.contains("theme-dark") ? "dark" : "light";
  });
  const router = useRouter();

  const toggleTheme = () => {
    const newTheme = theme === "light" ? "dark" : "light";
    setTheme(newTheme);
    if (newTheme === "dark") {
      document.documentElement.classList.add("theme-dark");
    } else {
      document.documentElement.classList.remove("theme-dark");
    }
    localStorage.setItem("devhub-theme", newTheme);
  };

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
            <span className="text-[10px] font-black text-muted-foreground dark:text-muted-foreground uppercase tracking-widest hidden lg:inline">
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
          <motion.button 
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={clearNotifications}
            className="relative p-2.5 rounded-xl hover:bg-muted/30 text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground transition-all"
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
              <div className="w-9 h-9 rounded-xl bg-primary/10 dark:bg-primary/20 flex items-center justify-center border border-border ring-2 ring-primary/20">
                <User className="w-5 h-5 text-primary" />
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
                    Preferences
                  </p>
                  <button
                    onClick={toggleTheme}
                    className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground hover:bg-muted/40 transition-all group"
                  >
                    <div className="w-8 h-8 rounded-lg bg-muted/20 flex items-center justify-center">
                      {theme === "light" ? <Sun className="w-4 h-4 text-amber-500" /> : <Moon className="w-4 h-4 text-indigo-400" />}
                    </div>
                    <span className="flex-1 text-left">{theme === "light" ? "Light Mode" : "Dark Mode"}</span>
                    <span className="text-[10px] opacity-40 font-bold uppercase tracking-widest">Switch</span>
                  </button>

                  <div className="h-px bg-muted/30 my-2" />
                  
                  <button 
                    onClick={() => { router.push("/account"); setShowDropdown(false); }}
                    className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground hover:bg-muted/40 transition-all"
                  >
                    <div className="w-8 h-8 rounded-lg bg-muted/20 flex items-center justify-center">
                      <User className="w-4 h-4" />
                    </div>
                    Account Profile
                  </button>

                  {role === "System Admin" && (
                    <button 
                      onClick={() => { router.push("/admin/settings"); setShowDropdown(false); }}
                      className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground hover:bg-muted/40 transition-all"
                    >
                      <div className="w-8 h-8 rounded-lg bg-muted/20 flex items-center justify-center">
                        <Settings className="w-4 h-4 text-orange-400" />
                      </div>
                      System Settings
                    </button>
                  )}

                  <div className="h-px bg-muted/30 my-2" />
                  
                  <button 
                    onClick={handleLogout}
                    className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-rose-400 hover:bg-rose-400/10 transition-all"
                  >
                    <div className="w-8 h-8 rounded-lg bg-rose-400/10 flex items-center justify-center">
                      <X className="w-4 h-4" />
                    </div>
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
