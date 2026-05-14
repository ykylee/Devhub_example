"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";
import { 
  Box, 
  GitBranch, 
  FolderKanban, 
  Settings, 
  ArrowLeft, 
  Clock, 
  User, 
  Shield, 
  Eye, 
  Globe,
  Activity,
  BarChart3,
  AlertTriangle
} from "lucide-react";
import { projectService } from "@/lib/services/project.service";
import { Application, ApplicationRepository, ApplicationStatus, ApplicationVisibility } from "@/lib/services/project.types";
import { Badge } from "@/components/ui/Badge";
import { RepositoryTable } from "@/components/project/RepositoryTable";
import { ProjectTable } from "@/components/project/ProjectTable";
import { useToast } from "@/components/ui/Toast";
import { cn } from "@/lib/utils";
import { format } from "date-fns";

type TabType = "overview" | "repositories" | "projects" | "metrics";

export default function ApplicationDetailPage() {
  const { id } = useParams() as { id: string };
  const router = useRouter();
  const { toast } = useToast();
  
  const [application, setApplication] = useState<Application | null>(null);
  const [repositories, setRepositories] = useState<ApplicationRepository[]>([]);
  const [activeTab, setActiveTab] = useState<TabType>("overview");
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const [appData, reposData] = await Promise.all([
          projectService.getApplication(id),
          projectService.getApplicationRepositories(id).catch(() => [] as ApplicationRepository[])
        ]);
        setApplication(appData);
        setRepositories(reposData);
      } catch (err) {
        console.error("[ApplicationDetail] load failed:", err);
        const status = (err as any).status;
        if (status === 501) {
          setError("Backend API is not fully implemented (501). Showing placeholders.");
        } else {
          setError("Failed to load application details.");
        }
      } finally {
        setIsLoading(false);
      }
    };
    load();
  }, [id]);

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center py-32 gap-4">
        <div className="w-12 h-12 border-4 border-purple-500/20 border-t-purple-500 rounded-full animate-spin" />
        <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Application Details...</p>
      </div>
    );
  }

  if (error && !application) {
    return (
      <div className="p-8 space-y-6">
        <button 
          onClick={() => router.back()}
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors text-xs font-bold uppercase tracking-widest"
        >
          <ArrowLeft className="w-4 h-4" /> Back to List
        </button>
        <div className="glass border-orange-500/20 bg-orange-500/5 rounded-3xl p-12 text-center">
          <AlertTriangle className="w-12 h-12 text-orange-400 mx-auto mb-4" />
          <h2 className="text-xl font-black text-foreground uppercase tracking-tight mb-2">Service Note</h2>
          <p className="text-sm text-muted-foreground max-w-md mx-auto">{error}</p>
          <button 
             onClick={() => router.push("/applications")}
             className="mt-6 px-6 py-2 bg-muted/20 border border-border rounded-xl text-xs font-bold text-foreground hover:bg-muted/40 transition-all"
          >
            Return to Applications
          </button>
        </div>
      </div>
    );
  }

  const tabs: { id: TabType; label: string; icon: any }[] = [
    { id: "overview", label: "Overview", icon: Box },
    { id: "repositories", label: "Repositories", icon: GitBranch },
    { id: "projects", label: "Projects", icon: FolderKanban },
    { id: "metrics", label: "Rollup Metrics", icon: BarChart3 },
  ];

  return (
    <div className="space-y-8 p-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col gap-6">
        <button 
          onClick={() => router.back()}
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors text-xs font-bold uppercase tracking-widest self-start"
        >
          <ArrowLeft className="w-4 h-4" /> Back
        </button>

        <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
          <div className="flex items-center gap-6">
            <div className="w-20 h-20 bg-purple-500/10 rounded-3xl flex items-center justify-center border border-purple-500/20 shadow-xl shadow-purple-500/5">
              <Box className="w-10 h-10 text-purple-400" />
            </div>
            <div>
              <div className="flex items-center gap-3">
                <h1 className="text-4xl font-black text-foreground tracking-tighter uppercase">
                  {application?.name || "Application Name"}
                </h1>
                <StatusBadge status={application?.status || "planning"} />
              </div>
              <div className="flex items-center gap-4 mt-2">
                <span className="text-sm font-mono text-muted-foreground bg-muted/50 px-2 py-1 rounded border border-border/40 uppercase tracking-tighter">
                  {application?.key || "ID_MISSING"}
                </span>
                <VisibilityBadge visibility={application?.visibility || "internal"} />
                <div className="flex items-center gap-2 text-muted-foreground">
                  <User className="w-4 h-4 opacity-40" />
                  <span className="text-xs font-bold">{application?.owner_user_id || "Unassigned"}</span>
                </div>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-3">
             <div className="flex flex-col items-end gap-1 px-4 py-2 bg-muted/20 rounded-2xl border border-border/60">
                <span className="text-[9px] font-black text-muted-foreground uppercase tracking-widest opacity-60">Governance Period</span>
                <span className="text-xs font-bold text-foreground">
                  {application?.start_date ? format(new Date(application.start_date), "MMM d, yyyy") : "TBD"}
                  {" — "}
                  {application?.due_date ? format(new Date(application.due_date), "MMM d, yyyy") : "Continuous"}
                </span>
             </div>
             <button className="p-3.5 glass border-border rounded-2xl text-muted-foreground hover:text-foreground transition-all">
               <Settings className="w-5 h-5" />
             </button>
          </div>
        </div>
      </div>

      {/* Tabs Navigation */}
      <nav className="flex p-1.5 glass border-border rounded-2xl gap-1 w-fit">
        {tabs.map((tab) => {
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                "flex items-center gap-2 px-6 py-3 rounded-xl text-xs font-black uppercase tracking-widest transition-all relative",
                isActive ? "text-foreground dark:text-primary-foreground" : "text-muted-foreground hover:text-foreground dark:hover:text-primary-foreground",
              )}
            >
              {isActive && (
                <motion.div
                  layoutId="app-detail-tab"
                  className="absolute inset-0 bg-muted/40 border border-border rounded-xl shadow-inner"
                  transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                />
              )}
              <tab.icon className={cn("w-4 h-4 relative z-10", isActive ? "text-purple-400" : "text-muted-foreground")} />
              <span className="relative z-10">{tab.label}</span>
            </button>
          );
        })}
      </nav>

      {/* Content */}
      <div className="flex-1">
        <AnimatePresence mode="wait">
          <motion.div
            key={activeTab}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
          >
            {activeTab === "overview" && (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                <div className="md:col-span-2 space-y-8">
                  <section className="glass border-border rounded-3xl p-8 space-y-4">
                    <h3 className="text-xs font-black text-muted-foreground uppercase tracking-[0.3em]">Description & Intent</h3>
                    <p className="text-foreground/80 leading-relaxed">
                      {application?.description || "No description provided for this application. Update the meta to include strategic goals and KPI summaries."}
                    </p>
                  </section>
                  
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                     <StatCard label="Total Repositories" value={repositories.length} icon={GitBranch} color="text-pink-400" />
                     <StatCard label="Active Projects" value={0} icon={FolderKanban} color="text-indigo-400" subValue="Mocked" />
                  </div>
                </div>
                
                <div className="space-y-8">
                  <section className="glass border-border rounded-3xl p-6 space-y-6">
                    <h3 className="text-xs font-black text-muted-foreground uppercase tracking-[0.3em]">Quick Links</h3>
                    <div className="space-y-3">
                       <LinkButton label="Jira Board" href="#" icon={ExternalLinkIcon} />
                       <LinkButton label="Confluence Docs" href="#" icon={ExternalLinkIcon} />
                       <LinkButton label="Gitea Organization" href="#" icon={ExternalLinkIcon} />
                    </div>
                  </section>
                  
                  <section className="glass border-border rounded-3xl p-6 space-y-4">
                    <h3 className="text-xs font-black text-muted-foreground uppercase tracking-[0.3em]">Lifecycle</h3>
                    <div className="space-y-4">
                       <TimelineItem date={application?.created_at} label="Application Created" />
                       <TimelineItem date={application?.updated_at} label="Last Meta Update" />
                    </div>
                  </section>
                </div>
              </div>
            )}

            {activeTab === "repositories" && (
              <div className="space-y-6">
                <div className="flex items-center justify-between">
                  <h2 className="text-xl font-black text-foreground uppercase tracking-tight">Connected Repositories</h2>
                  <button className="px-4 py-2 glass border-border rounded-xl text-[10px] font-black uppercase tracking-widest hover:bg-muted/40 transition-all">
                    Add Link
                  </button>
                </div>
                <RepositoryTable repositories={repositories} />
              </div>
            )}

            {activeTab === "projects" && (
              <div className="space-y-6">
                <div className="flex items-center justify-between">
                  <h2 className="text-xl font-black text-foreground uppercase tracking-tight">Operational Projects</h2>
                  <button className="px-4 py-2 glass border-border rounded-xl text-[10px] font-black uppercase tracking-widest hover:bg-muted/40 transition-all">
                    New Project
                  </button>
                </div>
                <ProjectTable projects={[]} />
              </div>
            )}

            {activeTab === "metrics" && (
              <div className="glass border-border rounded-3xl p-12 text-center space-y-4">
                <BarChart3 className="w-12 h-12 text-muted-foreground/20 mx-auto" />
                <h3 className="text-sm font-black text-foreground/60 uppercase tracking-widest">Rollup Metrics Preview</h3>
                <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest opacity-40 max-w-xs mx-auto">
                  Aggregate CI/CD, Quality, and Activity data from all connected repositories.
                </p>
                <div className="pt-8 grid grid-cols-2 md:grid-cols-4 gap-4 max-w-4xl mx-auto opacity-30 grayscale">
                   <StatCard label="Build Success" value="94%" icon={Activity} color="text-green-400" />
                   <StatCard label="Quality Score" value="8.4" icon={Shield} color="text-blue-400" />
                   <StatCard label="Open PRs" value="12" icon={GitBranch} color="text-pink-400" />
                   <StatCard label="Risks" value="2" icon={AlertTriangle} color="text-orange-400" />
                </div>
              </div>
            )}
          </motion.div>
        </AnimatePresence>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: ApplicationStatus }) {
  switch (status) {
    case "active":
      return <Badge variant="success" dot>Active</Badge>;
    case "planning":
      return <Badge variant="primary" dot>Planning</Badge>;
    case "on_hold":
      return <Badge variant="warning" dot>On Hold</Badge>;
    case "closed":
      return <Badge variant="secondary" dot>Closed</Badge>;
    case "archived":
      return <Badge variant="danger" dot>Archived</Badge>;
    default:
      return <Badge variant="glass">{status}</Badge>;
  }
}

function VisibilityBadge({ visibility }: { visibility: ApplicationVisibility }) {
  switch (visibility) {
    case "public":
      return (
        <Badge variant="glass" className="border-blue-500/20 text-blue-400 bg-blue-500/5">
          <Globe className="w-3 h-3" /> Public
        </Badge>
      );
    case "internal":
      return (
        <Badge variant="glass" className="border-emerald-500/20 text-emerald-400 bg-emerald-500/5">
          <Eye className="w-3 h-3" /> Internal
        </Badge>
      );
    case "restricted":
      return (
        <Badge variant="glass" className="border-rose-500/20 text-rose-400 bg-rose-500/5">
          <Shield className="w-3 h-3" /> Restricted
        </Badge>
      );
    default:
      return <Badge variant="glass">{visibility}</Badge>;
  }
}

function StatCard({ label, value, icon: Icon, color, subValue }: any) {
  return (
    <div className="glass border-border rounded-3xl p-6 flex items-center justify-between group">
      <div>
        <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1 opacity-60">{label}</p>
        <div className="flex items-baseline gap-2">
          <p className="text-2xl font-black text-foreground tracking-tighter">{value}</p>
          {subValue && <span className="text-[9px] font-bold text-muted-foreground/40 italic">{subValue}</span>}
        </div>
      </div>
      <div className={cn("w-12 h-12 rounded-2xl bg-muted/20 flex items-center justify-center border border-border/40 group-hover:scale-110 transition-transform", color)}>
        <Icon className="w-6 h-6" />
      </div>
    </div>
  );
}

function LinkButton({ label, icon: Icon, href }: any) {
  return (
    <a 
      href={href} 
      className="flex items-center justify-between p-4 glass border-border/60 rounded-2xl hover:bg-muted/40 transition-all group"
    >
      <span className="text-xs font-bold text-foreground/70 group-hover:text-foreground">{label}</span>
      <Icon className="w-4 h-4 text-muted-foreground/40 group-hover:text-primary transition-colors" />
    </a>
  );
}

function ExternalLinkIcon({ className }: { className?: string }) {
  return (
    <svg 
      xmlns="http://www.w3.org/2000/svg" 
      width="24" height="24" 
      viewBox="0 0 24 24" 
      fill="none" 
      stroke="currentColor" 
      strokeWidth="2" 
      strokeLinecap="round" 
      strokeLinejoin="round" 
      className={className}
    >
      <path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
    </svg>
  );
}

function TimelineItem({ date, label }: any) {
  return (
    <div className="flex items-start gap-3">
      <div className="mt-1 w-1.5 h-1.5 rounded-full bg-purple-500" />
      <div>
        <p className="text-[10px] font-black text-foreground/80 uppercase tracking-tight leading-none">{label}</p>
        <p className="text-[9px] text-muted-foreground mt-1">{date ? format(new Date(date), "yyyy-MM-dd HH:mm") : "TBD"}</p>
      </div>
    </div>
  );
}
