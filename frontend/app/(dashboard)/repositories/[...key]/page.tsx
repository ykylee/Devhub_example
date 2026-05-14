"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";
import { 
  GitBranch, 
  ArrowLeft, 
  Activity, 
  Clock, 
  ExternalLink, 
  ShieldCheck, 
  AlertTriangle,
  History,
  FileCode2,
  CheckCircle2,
  XCircle,
  Timer
} from "lucide-react";
import { projectService } from "@/lib/services/project.service";
import { ApplicationRepository, ApplicationRepositorySyncStatus, SyncErrorCode } from "@/lib/services/project.types";
import { Badge } from "@/components/ui/Badge";
import { useToast } from "@/components/ui/Toast";
import { cn } from "@/lib/utils";
import { format } from "date-fns";

export default function RepositoryDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  
  // Extract key from [...key] array
  const keyArray = params.key as string[];
  const provider = keyArray[0];
  const fullName = keyArray.slice(1).join("/");
  
  const [repository, setRepository] = useState<ApplicationRepository | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      setIsLoading(true);
      setError(null);
      try {
        // We don't have a direct "get single repo" API in §13, 
        // but we can list repos for an app or use the global list if it existed.
        // For now, let's look for this repo in all apps or just show placeholder if 501.
        
        // Mocking the search for this specific repo across apps
        const apps = await projectService.getApplications();
        let found = false;
        for (const app of apps) {
           const repos = await projectService.getApplicationRepositories(app.id);
           const match = repos.find(r => r.repo_provider === provider && r.repo_full_name === fullName);
           if (match) {
             setRepository(match);
             found = true;
             break;
           }
        }
        
        if (!found) {
           // Fallback to placeholder if backend APIs are 501
           setError("Repository not found or API implementation pending.");
        }
      } catch (err) {
        console.error("[RepositoryDetail] load failed:", err);
        setError("Failed to load repository details.");
      } finally {
        setIsLoading(false);
      }
    };
    load();
  }, [provider, fullName]);

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center py-32 gap-4">
        <div className="w-12 h-12 border-4 border-pink-500/20 border-t-pink-500 rounded-full animate-spin" />
        <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Repository Details...</p>
      </div>
    );
  }

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
            <div className="w-20 h-20 bg-pink-500/10 rounded-3xl flex items-center justify-center border border-pink-500/20 shadow-xl shadow-pink-500/5">
              <GitBranch className="w-10 h-10 text-pink-400" />
            </div>
            <div>
              <div className="flex items-center gap-3">
                <h1 className="text-4xl font-black text-foreground tracking-tighter uppercase">
                  {fullName || "Repository Name"}
                </h1>
                <SyncStatusBadge status={repository?.sync_status || "requested"} errorCode={repository?.sync_error_code} />
              </div>
              <div className="flex items-center gap-4 mt-2">
                <span className="text-xs font-black text-muted-foreground uppercase tracking-widest bg-muted/30 px-2 py-1 rounded border border-border/40">
                  {provider || "UNKNOWN"}
                </span>
                <Badge variant={repository?.role === 'primary' ? 'primary' : 'glass'}>
                  {repository?.role || "shared"}
                </Badge>
                <div className="flex items-center gap-2 text-muted-foreground">
                  <Activity className="w-4 h-4 opacity-40" />
                  <span className="text-[10px] font-bold uppercase tracking-tight">
                    Last Sync: {repository?.last_sync_at ? format(new Date(repository.last_sync_at), "MMM d, HH:mm") : "Never"}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-3">
             <a 
               href="#" 
               target="_blank" 
               className="flex items-center gap-2 px-6 py-3 glass border-border rounded-2xl text-[10px] font-black uppercase tracking-widest hover:bg-muted/40 transition-all shadow-lg"
             >
               View on {provider} <ExternalLink className="w-3.5 h-3.5" />
             </a>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        <div className="md:col-span-2 space-y-8">
          {/* Recent Pipelines */}
          <section className="glass border-border rounded-3xl overflow-hidden">
             <div className="p-6 border-b border-border/40 bg-muted/10 flex items-center justify-between">
                <h3 className="text-xs font-black text-muted-foreground uppercase tracking-[0.3em]">Recent Pipelines (CI/CD)</h3>
                <Badge variant="glass">Planned API-53</Badge>
             </div>
             <div className="divide-y divide-border/40 opacity-40 grayscale pointer-events-none">
                <PipelineRow status="success" branch="main" sha="f327ff2" time="12m ago" duration="4m 20s" />
                <PipelineRow status="failed" branch="feat/auth" sha="460aa01" time="2h ago" duration="1m 05s" />
                <PipelineRow status="success" branch="main" sha="642d976" time="5h ago" duration="4m 15s" />
             </div>
             <div className="p-6 text-center border-t border-border/40">
                <p className="text-[9px] font-bold text-muted-foreground uppercase tracking-widest italic">
                  Aggregate build runs from external provider...
                </p>
             </div>
          </section>

          {/* Code Quality */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
             <QualityCard label="Static Analysis" score="8.4" trend="+0.2" tool="SonarQube" color="text-blue-400" />
             <QualityCard label="Unit Test Coverage" score="92%" trend="-1.5%" tool="Vitest" color="text-emerald-400" />
          </div>
        </div>

        <div className="space-y-8">
           {/* Connection Details */}
           <section className="glass border-border rounded-3xl p-6 space-y-6">
              <h3 className="text-xs font-black text-muted-foreground uppercase tracking-[0.3em]">Connection Health</h3>
              <div className="space-y-4">
                 <HealthItem label="Webhook Delivery" status="healthy" />
                 <HealthItem label="Token Authority" status="healthy" />
                 <HealthItem label="Branch Protection" status="warning" note="Unsigned commits allowed" />
              </div>
           </section>

           {/* Repository Metadata */}
           <section className="glass border-border rounded-3xl p-6 space-y-4">
              <h3 className="text-xs font-black text-muted-foreground uppercase tracking-[0.3em]">Repository Meta</h3>
              <div className="space-y-3">
                 <div className="flex justify-between items-center">
                    <span className="text-[10px] text-muted-foreground font-bold uppercase tracking-tight">External ID</span>
                    <span className="text-[10px] font-mono text-foreground">{repository?.external_repo_id || "N/A"}</span>
                 </div>
                 <div className="flex justify-between items-center">
                    <span className="text-[10px] text-muted-foreground font-bold uppercase tracking-tight">Linked At</span>
                    <span className="text-[10px] font-mono text-foreground">
                      {repository?.linked_at ? format(new Date(repository.linked_at), "yyyy-MM-dd") : "TBD"}
                    </span>
                 </div>
              </div>
           </section>
        </div>
      </div>
    </div>
  );
}

function SyncStatusBadge({ status, errorCode }: { status: ApplicationRepositorySyncStatus, errorCode?: SyncErrorCode }) {
  switch (status) {
    case "active":
      return <Badge variant="success" dot>Active</Badge>;
    case "requested":
    case "verifying":
      return <Badge variant="primary" dot className="animate-pulse">Syncing</Badge>;
    case "degraded":
      return <Badge variant="warning" dot>Degraded</Badge>;
    case "disconnected":
      return <Badge variant="secondary">Disconnected</Badge>;
    default:
      return <Badge variant="glass">{status}</Badge>;
  }
}

function PipelineRow({ status, branch, sha, time, duration }: any) {
  return (
    <div className="flex items-center justify-between p-4 px-6 hover:bg-muted/20 transition-colors">
       <div className="flex items-center gap-4">
          {status === 'success' ? <CheckCircle2 className="w-5 h-5 text-green-500" /> : <XCircle className="w-5 h-5 text-rose-500" />}
          <div>
             <div className="flex items-center gap-2">
                <span className="text-xs font-bold text-foreground">#{Math.floor(Math.random() * 1000)}</span>
                <span className="text-[10px] font-mono text-muted-foreground bg-muted/50 px-1.5 py-0.5 rounded uppercase">{branch}</span>
                <span className="text-[10px] font-mono text-muted-foreground opacity-50">{sha}</span>
             </div>
             <p className="text-[9px] text-muted-foreground mt-0.5">{time}</p>
          </div>
       </div>
       <div className="flex items-center gap-2 text-muted-foreground">
          <Timer className="w-3.5 h-3.5 opacity-40" />
          <span className="text-[10px] font-bold uppercase">{duration}</span>
       </div>
    </div>
  );
}

function QualityCard({ label, score, trend, tool, color }: any) {
  return (
    <div className="glass border-border rounded-3xl p-6 space-y-4 group">
       <div className="flex items-center justify-between">
          <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest opacity-60">{label}</p>
          <span className="text-[9px] font-black text-muted-foreground bg-muted/40 px-2 py-0.5 rounded-full border border-border/40">{tool}</span>
       </div>
       <div className="flex items-baseline justify-between">
          <h4 className={cn("text-3xl font-black tracking-tighter", color)}>{score}</h4>
          <span className={cn("text-[10px] font-bold", trend.startsWith('+') ? 'text-green-500' : 'text-rose-500')}>
            {trend}
          </span>
       </div>
       <div className="w-full h-1 bg-muted/30 rounded-full overflow-hidden">
          <motion.div 
             initial={{ width: 0 }}
             animate={{ width: '80%' }}
             className={cn("h-full", color.replace('text-', 'bg-'))}
          />
       </div>
    </div>
  );
}

function HealthItem({ label, status, note }: any) {
  return (
    <div className="flex items-start justify-between gap-4">
       <div className="flex flex-col gap-0.5">
          <span className="text-[11px] font-bold text-foreground/80">{label}</span>
          {note && <span className="text-[9px] text-muted-foreground opacity-60 leading-tight">{note}</span>}
       </div>
       {status === 'healthy' ? <ShieldCheck className="w-4 h-4 text-green-500" /> : <AlertTriangle className="w-4 h-4 text-orange-400" />}
    </div>
  );
}
