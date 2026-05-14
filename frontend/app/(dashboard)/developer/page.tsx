"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { 
  CheckCircle2,
  AlertTriangle,
  CircleDashed, 
  GitPullRequest, 
  MessageSquare, 
  PlayCircle, 
  Star, 
  Terminal, 
  Zap, 
  Coffee, 
  ArrowRight,
  Info
} from "lucide-react";
import { useStore } from "@/lib/store";
import { mockBuildLogs } from "@/lib/mockData";
import { Modal } from "@/components/ui/Modal";
import { Badge } from "@/components/ui/Badge";
import { cn } from "@/lib/utils";
import { infraService } from "@/lib/services/infra.service";
import { Metric } from "@/lib/services/types";
import { useEffect } from "react";

export default function DeveloperDashboard() {
  const { isDeepFocus, setDeepFocus, addToast } = useStore();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [stats, setStats] = useState<Metric[]>([]);

  useEffect(() => {
    const loadData = async () => {
      try {
        const metricsData = await infraService.getMetrics("Developer");
        setStats(metricsData);
      } catch (error) {
        console.error("Failed to load metrics:", error);
      }
    };
    loadData();
    const interval = setInterval(loadData, 10000);
    return () => clearInterval(interval);
  }, []);

  const toggleFocus = () => {
    const nextState = !isDeepFocus;
    setDeepFocus(nextState);
    if (nextState) {
      addToast("Deep Focus mode activated. Flow state protection on.", "success");
    } else {
      addToast("Focus session ended.", "info");
    }
  };

  return (
    <div className="space-y-10 pb-20">
      {/* Header Section */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
        >
          <h1 className="text-4xl font-extrabold tracking-tight text-foreground dark:text-primary-foreground mb-2">
            Developer <span className="text-gradient">Workspace</span>
          </h1>
          <p className="text-muted-foreground text-lg flex items-center gap-2">
            Welcome back, <span className="text-foreground dark:text-primary-foreground font-bold">YK Lee</span> • <Badge variant="success" dot>Active Now</Badge>
          </p>
        </motion.div>

        <motion.div 
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          className="flex items-center gap-4"
        >
          <button 
            onClick={toggleFocus}
            className={cn(
              "flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-bold transition-all border",
              isDeepFocus 
                ? "bg-primary text-primary-foreground border-primary shadow-[0_0_20px_rgba(139,92,246,0.5)]" 
                : "glass border-border text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground"
            )}
          >
            {isDeepFocus ? <Zap className="w-4 h-4 fill-current" /> : <Coffee className="w-4 h-4" />}
            {isDeepFocus ? "Deep Focus Active" : "Start Deep Focus"}
          </button>
          <button 
            onClick={() => setIsModalOpen(true)}
            className="glass text-foreground dark:text-primary-foreground px-6 py-2 rounded-xl text-sm font-bold hover:bg-muted/40 transition-all border border-border flex items-center gap-2"
          >
            <Info className="w-4 h-4" /> Project Info
          </button>
        </motion.div>
      </div>

      {/* Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        {stats.map((stat, i) => (
          <motion.div 
            key={i}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card p-6 flex flex-col justify-between"
          >
            <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">{stat.label}</p>
            <div className="flex items-center justify-between">
              <h3 className="text-2xl font-black text-foreground dark:text-primary-foreground">{stat.value}</h3>
              <span className={cn("text-[10px] font-black uppercase tracking-tighter", stat.color)}>
                {stat.trend}
              </span>
            </div>
          </motion.div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Main Content Area */}
        <div className="lg:col-span-2 space-y-8">
          
          {/* Work Stream */}
          <section className="space-y-4">
            <div className="flex items-center justify-between px-2">
              <h2 className="text-xl font-bold text-foreground dark:text-primary-foreground flex items-center gap-2">
                <Terminal className="w-5 h-5 text-primary" /> Active Stream
              </h2>
              <button className="text-xs font-bold text-muted-foreground hover:text-primary transition-colors flex items-center gap-1">
                View All <ArrowRight className="w-3 h-3" />
              </button>
            </div>
            
            <div className="grid gap-4">
              {[
                { 
                  id: "TASK-007", 
                  title: "Gitea Webhook Receiver Implementation", 
                  repo: "devhub-core", 
                  status: "In Progress", 
                  type: "feature",
                  progress: 65,
                  time: "Updated 12m ago"
                },
                { 
                  id: "PR-124", 
                  title: "Refactor Auth Middleware for gRPC", 
                  repo: "devhub-ai", 
                  status: "Review", 
                  type: "refactor",
                  progress: 100,
                  time: "Updated 2h ago"
                }
              ].map((task, i) => (
                <motion.div
                  key={task.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: i * 0.1 }}
                  className="glass-card p-6 group cursor-pointer relative overflow-hidden"
                >
                  <div className="flex items-start justify-between mb-4">
                    <div className="flex gap-4">
                      <div className="mt-1">
                        {task.status === "In Progress" ? (
                          <div className="relative">
                            <CircleDashed className="w-6 h-6 text-primary animate-[spin_4s_linear_infinite]" />
                            <div className="absolute inset-0 bg-primary/20 blur-lg rounded-full animate-pulse" />
                          </div>
                        ) : (
                          <GitPullRequest className="w-6 h-6 text-accent" />
                        )}
                      </div>
                      <div>
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-[10px] font-black text-primary uppercase tracking-widest">{task.id}</span>
                          <span className="text-[10px] text-primary-foreground/30">•</span>
                          <span className="text-[10px] font-bold text-muted-foreground uppercase">{task.repo}</span>
                        </div>
                        <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground group-hover:text-primary transition-colors">
                          {task.title}
                        </h3>
                      </div>
                    </div>
                    <div className={cn(
                      "px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-tighter",
                      task.status === "In Progress" ? "bg-primary/20 text-primary border border-primary/20" : "bg-accent/20 text-accent border border-accent/20"
                    )}>
                      {task.status}
                    </div>
                  </div>
                  
                  <div className="space-y-2">
                    <div className="flex justify-between text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                      <span>Progress</span>
                      <span>{task.progress}%</span>
                    </div>
                    <div className="h-1.5 w-full bg-muted/30 rounded-full overflow-hidden">
                      <motion.div 
                        initial={{ width: 0 }}
                        animate={{ width: `${task.progress}%` }}
                        className={cn(
                          "h-full transition-all duration-1000",
                          task.status === "In Progress" ? "bg-primary" : "bg-accent"
                        )}
                      />
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </section>

          {/* Infrastructure Health / Builds */}
          <section className="space-y-4">
            <h2 className="text-xl font-bold text-foreground dark:text-primary-foreground flex items-center gap-2 px-2">
              <PlayCircle className="w-5 h-5 text-accent" /> Deployment Pipeline
            </h2>
            <div className="glass-card divide-y divide-border/60">
              {mockBuildLogs.map((build) => (
                <div key={build.id} className="p-5 flex items-center justify-between hover:bg-muted/20 transition-colors">
                  <div className="flex items-center gap-4">
                    <div className={cn(
                      "w-10 h-10 rounded-xl flex items-center justify-center border",
                      build.status === "Passed" ? "bg-green-500/10 border-green-500/20" : "bg-rose-500/10 border-rose-500/20"
                    )}>
                      {build.status === "Passed" ? (
                        <CheckCircle2 className="w-5 h-5 text-green-500" />
                      ) : (
                        <AlertTriangle className="w-5 h-5 text-rose-500" />
                      )}
                    </div>
                    <div>
                      <p className="text-sm font-bold text-foreground dark:text-primary-foreground">{build.title}</p>
                      <p className="text-xs text-muted-foreground mt-0.5">{build.status} • {build.time} • <span className="font-mono">8a2f1b4</span></p>
                    </div>
                  </div>
                  <motion.button 
                    whileHover={{ scale: 1.1 }}
                    className="p-2 rounded-lg hover:bg-muted/30 text-muted-foreground transition-all"
                  >
                    <ArrowRight className="w-4 h-4" />
                  </motion.button>
                </div>
              ))}
            </div>
          </section>
        </div>

        {/* Sidebar Widgets */}
        <div className="space-y-8">
          {/* AI Gardener Widget */}
          <motion.div 
            whileHover={{ y: -5 }}
            className="relative p-8 rounded-3xl overflow-hidden group"
          >
            <div className="absolute inset-0 bg-gradient-to-br from-primary/20 via-accent/20 to-transparent opacity-50 transition-opacity group-hover:opacity-70" />
            <div className="absolute inset-0 glass opacity-50" />
            
            <div className="relative z-10 space-y-4">
              <div className="flex items-center gap-2">
                <div className="p-2 rounded-lg bg-muted/40 border border-border/80">
                  <Star className="w-4 h-4 text-yellow-400 fill-current" />
                </div>
                <span className="text-sm font-black uppercase tracking-widest text-foreground dark:text-primary-foreground">AI Gardener</span>
              </div>
              
              <p className="text-sm text-foreground/80 dark:text-primary-foreground/80 leading-relaxed font-medium">
                &quot;I noticed you&apos;re implementing a <span className="text-primary font-bold">Webhook secret validator</span>. There&apos;s a battle-tested helper in the <span className="underline decoration-accent underline-offset-4 cursor-pointer">shared-utils</span> package that handles Gitea signatures.&quot;
              </p>
              
              <button className="w-full py-3 rounded-2xl bg-card text-card-foreground text-xs font-black uppercase tracking-widest hover:bg-card/90 transition-all shadow-xl">
                Adopt Suggestion
              </button>
            </div>
          </motion.div>

          {/* Social / Kudos */}
          <div className="glass-card p-6 space-y-6">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground flex items-center gap-2">
              <MessageSquare className="w-4 h-4 text-accent" /> Recent Recognition
            </h3>
            <div className="space-y-4">
              <div className="p-4 rounded-2xl bg-muted/30 border border-border/60 space-y-3">
                <div className="flex items-center gap-2">
                  <div className="w-6 h-6 rounded-full bg-blue-500 flex items-center justify-center text-[10px] font-bold text-primary-foreground">A</div>
                  <span className="text-xs font-bold text-foreground dark:text-primary-foreground">@alex_dev</span>
                </div>
                <p className="text-xs text-muted-foreground dark:text-muted-foreground leading-relaxed">
                  &quot;Amazing work on the gRPC migration! The performance gains are already visible in the staging logs. Keep it up! 🚀&quot;
                </p>
              </div>
            </div>
          </div>

          {/* System Status Mini */}
          <div className="glass-card p-6">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-4">Infrastructure</h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-foreground/70 dark:text-primary-foreground/70">Gitea Server</span>
                <span className="text-[10px] font-black text-green-500 uppercase">Operational</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-foreground/70 dark:text-primary-foreground/70">AI Runner</span>
                <span className="text-[10px] font-black text-green-500 uppercase">Operational</span>
              </div>
              <div className="flex items-center justify-between opacity-50">
                <span className="text-xs font-bold text-foreground/70 dark:text-primary-foreground/70">Metrics DB</span>
                <span className="text-[10px] font-black text-amber-500 uppercase">Maintenance</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <AnimatePresence>
        {isDeepFocus && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-[100] glass flex items-center justify-center p-6 text-center"
          >
            <div className="max-w-md space-y-8">
              <motion.div
                animate={{ 
                  scale: [1, 1.1, 1],
                  rotate: [0, 5, -5, 0]
                }}
                transition={{ duration: 4, repeat: Infinity }}
                className="w-24 h-24 bg-primary/20 rounded-full flex items-center justify-center mx-auto border border-primary/40 shadow-[0_0_50px_rgba(139,92,246,0.3)]"
              >
                <Zap className="w-12 h-12 text-primary fill-current" />
              </motion.div>
              <div className="space-y-2">
                <h2 className="text-4xl font-black text-foreground dark:text-primary-foreground uppercase tracking-tighter">Deep Focus Mode</h2>
                <p className="text-muted-foreground dark:text-muted-foreground">Notifications are silenced. DevHub is protecting your flow state.</p>
              </div>
              <button 
                onClick={() => setDeepFocus(false)}
                className="px-8 py-3 rounded-2xl bg-card text-card-foreground font-black uppercase tracking-widest hover:bg-card/90 transition-all shadow-2xl"
              >
                Exit Session
              </button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
      <Modal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)}
        title="Project Intelligence Summary"
        size="lg"
      >
        <div className="space-y-6">
          <div className="grid grid-cols-2 gap-4">
            <div className="glass-card p-4">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Current Milestone</p>
              <p className="text-lg font-bold text-primary-foreground">v1.0.0-beta-2</p>
            </div>
            <div className="glass-card p-4">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Deployment Status</p>
              <Badge variant="success" dot>Stable</Badge>
            </div>
          </div>
          
          <div className="space-y-3">
            <h4 className="text-sm font-bold text-primary-foreground">Technical Ecosystem</h4>
            <div className="flex flex-wrap gap-2">
              {["Go Core", "Python AI Engine", "Next.js 15", "Tailwind 4", "gRPC", "PostgreSQL"].map(tech => (
                <Badge key={tech} variant="secondary">{tech}</Badge>
              ))}
            </div>
          </div>
          
          <div className="p-4 rounded-2xl bg-primary/5 border border-primary/10">
            <p className="text-xs text-primary/80 leading-relaxed italic">
              &quot;AI Gardener has analyzed your current PR. No conflicts detected with the Gitea migration branch. Proceed with confidence.&quot;
            </p>
          </div>
          
          <button 
            onClick={() => setIsModalOpen(false)}
            className="w-full py-4 rounded-2xl bg-card text-card-foreground text-xs font-black uppercase tracking-widest hover:bg-card/90 transition-all"
          >
            Acknowledge & Continue
          </button>
        </div>
      </Modal>
    </div>
  );
}
