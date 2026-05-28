"use client";

import { useState, useEffect } from "react";
import { Search, Bell, User, ChevronDown, Command, Sun, Moon, Settings, X, Menu } from "lucide-react";
import { cn } from "@/shared/utils";
import { motion, AnimatePresence } from "framer-motion";

import { useStore } from "@/lib/store";
import { useRouter } from "next/navigation";
import { authService } from "@/lib/services/auth.service";
import { realtimeService, type ConnectionStatusEvent } from "@/lib/services/realtime.service";
import { devRequestService } from "@/lib/services/dev_request.service";
import { DevRequest } from "@/lib/services/dev_request.types";
import { repositoryService, type Repository } from "@/lib/services/repository.service";
import type { Project } from "@/lib/services/project.types";
import { DevRequestDetailModal } from "@/components/dev-request/DevRequestDetailModal";
import { ProjectCreationModal } from "@/components/project/ProjectCreationModal";

export function Header({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  const { role, actor, notifications, setSidebarOpen } = useStore();
  const router = useRouter();
  const [showDropdown, setShowDropdown] = useState(false);
  const [showNotifications, setShowNotifications] = useState(false);
  const [pendingDreqs, setPendingDreqs] = useState<DevRequest[]>([]);
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [selectedDreq, setSelectedDreq] = useState<DevRequest | null>(null);
  const [showDreqDetail, setShowDreqDetail] = useState(false);
  const [showProjectCreate, setShowProjectCreate] = useState(false);
  const [projectPrefill, setProjectPrefill] = useState<Partial<Project> | null>(null);

  const [isConnected, setIsConnected] = useState(realtimeService.isConnected);
  // 초기 theme 은 paint 전에 layout 의 inline script 가 html 에 적용하므로
  // 여기서는 그 결과(`theme-dark` class 유무)를 읽어 state 와 일치시킨다.
  const [theme, setTheme] = useState<"light" | "dark">(() => {
    if (typeof document === "undefined") return "light";
    return document.documentElement.classList.contains("theme-dark") ? "dark" : "light";
  });

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

  const fetchDreqs = async () => {
    try {
      const res = await devRequestService.list({ status: ["pending", "in_review"], limit: 5 });
      setPendingDreqs(res.data);
      useStore.setState({ notifications: res.total });
    } catch (err) {
      console.error("Failed to fetch pending DREQs for header:", err);
    }
  };

  const fetchRepos = async () => {
    try {
      const repos = await repositoryService.listRepositories();
      setRepositories(repos);
    } catch (err) {
      console.error("Failed to fetch repositories for promotion:", err);
    }
  };

  useEffect(() => {
    // Mount-only: 초기 dreq/repo 페치 + WS subscribe. set-state-in-effect 룰의
    // cascading render 우려는 async fetch boundary + WS callback 이라 해당 없음.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchDreqs();
    fetchRepos();

    const unsubscribeStatus = realtimeService.subscribe<ConnectionStatusEvent>('status.changed', (event) => {
      setIsConnected(event.data.connected);
      fetchDreqs();
    });

    const unsubscribeDreq = realtimeService.subscribe<DevRequest>('dev_request.created', () => {
      fetchDreqs();
    });

    // 레거시 websocketService(AuthGuard) 제거에 따른 알림 핸들러 이관 — ticket WS
    // (realtimeService) 로 일원화. `DEFAULT_EVENT_TYPES` 에 두 이벤트가 이미 포함되어
    // 백엔드 push 를 그대로 수신한다 (realtime.service.ts:14-15). store action 은
    // callback 시점 getState() 로 읽어 effect 의존성/재구독을 피한다.
    const unsubscribeNotification = realtimeService.subscribe<{ message?: string }>('notification.created', (event) => {
      useStore.getState().incrementNotifications();
      useStore.getState().addToast(event.data?.message || "New Notification", "info");
    });

    const unsubscribeRisk = realtimeService.subscribe<{ message?: string }>('risk.critical.created', (event) => {
      useStore.getState().addToast(`CRITICAL: ${event.data?.message || "Risk Detected"}`, "error");
    });

    return () => {
      unsubscribeStatus();
      unsubscribeDreq();
      unsubscribeNotification();
      unsubscribeRisk();
    };
  }, []);

  const handleLogout = async () => {
    setShowDropdown(false);
    await authService.logout();
  };

  return (
    <header className={cn("sticky top-0 z-50 w-full glass border-b border-border/60", className)} {...props}>
      <div className="flex h-16 items-center px-4 lg:px-8 gap-4 lg:gap-8">
        <button 
          onClick={() => setSidebarOpen(true)}
          className="p-2 rounded-xl hover:bg-muted/30 text-muted-foreground lg:hidden"
          aria-label="Open sidebar"
        >
          <Menu className="w-6 h-6" />
        </button>

        <div className="flex-1 flex items-center gap-4">
          <div className="flex items-center gap-2 glass border-border px-3 py-1.5 rounded-xl">
            <div className={cn(
              "w-2 h-2 rounded-full animate-pulse",
              isConnected ? "bg-success shadow-[0_0_8px_rgba(16,185,129,0.5)]" : "bg-destructive shadow-[0_0_8px_rgba(244,63,94,0.5)]"
            )} aria-hidden="true" />
            <span className="text-[10px] font-black text-muted-foreground dark:text-muted-foreground uppercase tracking-widest hidden lg:inline">
              {isConnected ? "Real-time Live" : "Offline"}
            </span>
          </div>
          <div className="relative w-full max-w-lg hidden md:flex items-center group">
            <div className="absolute left-3.5 flex items-center gap-2 pointer-events-none">
              <Search className="h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" aria-hidden="true" />
              <div className="px-1.5 py-0.5 rounded border border-border bg-muted/30 text-[10px] font-mono text-muted-foreground">
                <Command className="w-2 h-2 inline mr-0.5" aria-hidden="true" /> K
              </div>
            </div>
            <input
              type="search"
              placeholder="Search anything..."
              className="flex h-10 w-full rounded-xl border border-border/60 bg-muted/30 px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 focus:bg-muted/40 transition-all pl-24"
              aria-label="Global search"
            />
          </div>
        </div>
        
        <div className="flex items-center gap-3 lg:gap-6">
          <div className="relative">
            <motion.button 
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => {
                setShowNotifications(!showNotifications);
                setShowDropdown(false);
              }}
              className="relative p-2.5 rounded-xl hover:bg-muted/30 text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground transition-all"
              aria-label={`Notifications (${notifications} new)`}
            >
              <Bell className="h-5 w-5" aria-hidden="true" />
              {notifications > 0 && (
                <span className="absolute top-2 right-2 w-2.5 h-2.5 bg-accent rounded-full border-2 border-background flex items-center justify-center text-[7px] font-black text-white">{notifications}</span>
              )}
            </motion.button>

            <AnimatePresence>
              {showNotifications && (
                <motion.div
                  initial={{ opacity: 0, y: 10, scale: 0.95 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: 10, scale: 0.95 }}
                  className="absolute top-full right-0 mt-4 w-80 rounded-2xl glass border border-border p-4 z-[200] shadow-2xl space-y-4"
                >
                  <div className="flex items-center justify-between border-b border-border/40 pb-2">
                    <span className="text-xs font-black uppercase tracking-widest text-foreground dark:text-primary-foreground">Pending Requests</span>
                    <span className="px-2 py-0.5 rounded-full bg-accent/20 text-accent text-[10px] font-black">{pendingDreqs.length}</span>
                  </div>

                  <div className="space-y-2 max-h-64 overflow-y-auto custom-scrollbar">
                    {pendingDreqs.length === 0 ? (
                      <p className="text-center py-6 text-xs text-muted-foreground uppercase tracking-widest font-black opacity-50">No pending requests</p>
                    ) : (
                      pendingDreqs.map((req) => (
                        <div 
                          key={req.id}
                          onClick={() => {
                            setSelectedDreq(req);
                            setShowDreqDetail(true);
                            setShowNotifications(false);
                          }}
                          className="p-3 rounded-xl border border-border/40 bg-muted/10 hover:bg-muted/30 hover:border-primary/30 transition-all cursor-pointer space-y-1 text-left"
                        >
                          <div className="flex items-center justify-between">
                            <span className="text-[9px] font-mono text-muted-foreground uppercase tracking-wider">{req.external_ref || "DREQ"}</span>
                            <span className="text-[8px] font-mono text-muted-foreground">{new Date(req.received_at).toLocaleDateString()}</span>
                          </div>
                          <p className="text-xs font-bold text-foreground dark:text-primary-foreground truncate">{req.title}</p>
                          <p className="text-[9px] text-muted-foreground uppercase tracking-widest truncate">From: {req.requester}</p>
                        </div>
                      ))
                    )}
                  </div>

                  <div className="pt-2 border-t border-border/40 text-center">
                    <button
                      onClick={() => {
                        router.push("/dev-requests");
                        setShowNotifications(false);
                      }}
                      className="text-[10px] font-black text-primary hover:text-foreground dark:hover:text-primary-foreground uppercase tracking-widest transition-colors"
                    >
                      View All Dev Requests
                    </button>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
          
          <div className="h-6 w-px bg-muted/40 hidden sm:block"></div>
          
          <div className="relative">
            {/* Native <button> instead of role="button" motion.div — Playwright trusted click events
                + accessibility tree consistency. e2e regression hotfix (PR #248). */}
            <button
              type="button"
              onClick={() => {
                setShowDropdown(!showDropdown);
                setShowNotifications(false);
              }}
              className="flex items-center gap-3 py-1.5 px-3 rounded-2xl hover:bg-muted/30 transition-all cursor-pointer group"
              aria-haspopup="true"
              aria-expanded={showDropdown}
              aria-label="User menu"
            >
              <div className="w-9 h-9 rounded-xl bg-primary/10 dark:bg-primary/20 flex items-center justify-center border border-border ring-2 ring-primary/20">
                <User className="w-5 h-5 text-primary" aria-hidden="true" />
              </div>
              <div className="flex flex-col hidden sm:flex">
                <span className="text-sm font-semibold leading-none text-foreground dark:text-primary-foreground">{actor?.login || "Guest User"}</span>
                <span className="text-[10px] font-bold text-muted-foreground mt-1 flex items-center gap-1 uppercase tracking-wider">
                  {role || "No Role"} <ChevronDown className={cn("w-3 h-3 transition-transform duration-300", showDropdown && "rotate-180")} aria-hidden="true" />
                </span>
              </div>
            </button>

            <AnimatePresence>
              {showDropdown && (
                <motion.div
                  initial={{ opacity: 0, y: 10, scale: 0.95 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: 10, scale: 0.95 }}
                  // z-[200] (originally z-50) — Sidebar mobile drawer is z-[150];
                  // header sticky stacking context conflict 회피.
                  className="absolute top-full right-0 mt-4 w-56 rounded-2xl glass border border-border p-2 z-[200] shadow-2xl"
                  role="menu"
                >
                  <p className="px-3 pt-2 text-[10px] font-bold text-muted-foreground uppercase tracking-widest opacity-50">
                    Preferences
                  </p>
                  <button
                    onClick={toggleTheme}
                    className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground hover:bg-muted/40 transition-all group"
                    role="menuitem"
                  >
                    <div className="w-8 h-8 rounded-lg bg-muted/20 flex items-center justify-center">
                      {theme === "light" ? <Sun className="w-4 h-4 text-warning" aria-hidden="true" /> : <Moon className="w-4 h-4 text-indigo-400" aria-hidden="true" />}
                    </div>
                    <span className="flex-1 text-left">{theme === "light" ? "Light Mode" : "Dark Mode"}</span>
                    <span className="text-[10px] opacity-40 font-bold uppercase tracking-widest">Switch</span>
                  </button>

                  <div className="h-px bg-muted/30 my-2" />
                  
                  <button 
                    onClick={() => { router.push("/account"); setShowDropdown(false); }}
                    className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground hover:bg-muted/40 transition-all"
                    role="menuitem"
                  >
                    <div className="w-8 h-8 rounded-lg bg-muted/20 flex items-center justify-center">
                      <User className="w-4 h-4" aria-hidden="true" />
                    </div>
                    Account Profile
                  </button>

                  {role === "System Admin" && (
                    <button 
                      onClick={() => { router.push("/admin/settings"); setShowDropdown(false); }}
                      className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground hover:bg-muted/40 transition-all"
                      role="menuitem"
                    >
                      <div className="w-8 h-8 rounded-lg bg-muted/20 flex items-center justify-center">
                        <Settings className="w-4 h-4 text-accent" aria-hidden="true" />
                      </div>
                      System Settings
                    </button>
                  )}

                  <div className="h-px bg-muted/30 my-2" />
                  
                  <button 
                    onClick={handleLogout}
                    className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-destructive hover:bg-destructive/10 transition-all"
                    role="menuitem"
                  >
                    <div className="w-8 h-8 rounded-lg bg-destructive/10 flex items-center justify-center">
                      <X className="w-4 h-4" aria-hidden="true" />
                    </div>
                    Sign Out
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>

      <AnimatePresence>
        {showDreqDetail && selectedDreq && (
          <DevRequestDetailModal
            request={selectedDreq}
            isSystemAdmin={role === "System Admin"}
            onClose={() => {
              setShowDreqDetail(false);
              setSelectedDreq(null);
            }}
            onChanged={() => {
              fetchDreqs();
            }}
            onPromote={(req) => {
              setShowDreqDetail(false);
              setProjectPrefill({
                key: req.external_ref || "",
                name: req.title || "",
                description: req.details || "",
              });
              setShowProjectCreate(true);
            }}
          />
        )}

        {showProjectCreate && (
          <ProjectCreationModal
            repositories={repositories}
            initialData={projectPrefill ?? undefined}
            onClose={() => {
              setShowProjectCreate(false);
              setProjectPrefill(null);
              setSelectedDreq(null);
            }}
            onCreated={(newProj) => {
              if (selectedDreq) {
                devRequestService.register(selectedDreq.id, {
                  target_type: "project",
                  target_id: newProj.id,
                }).then(() => {
                  fetchDreqs();
                }).catch(console.error);
              }
              setShowProjectCreate(false);
              setProjectPrefill(null);
              setSelectedDreq(null);
            }}
          />
        )}
      </AnimatePresence>
    </header>
  );
}
