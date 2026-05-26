"use client";

import { useCallback, useEffect, useState } from "react";
import { motion } from "framer-motion";
import { 
  Activity, 
  ArrowLeft, 
  Globe,
  ShieldCheck, 
  Zap,
  Clock,
  ExternalLink,
  RefreshCcw,
  Settings,
  GitBranch,
  Code2
} from "lucide-react";
import { useRouter, useParams } from "next/navigation";
import { Badge } from "@/components/ui/Badge";
import { cn } from "@/lib/utils";
import { applicationService, Application, ApplicationRollup } from "@/lib/services/application.service";
import { projectService } from "@/lib/services/project.service";
import { ApplicationRepository } from "@/lib/services/project.types";
import { toUserErrorMessage } from "@/lib/services/error-message";
import { PageError, PageLoading } from "@/components/ui/PageState";
import { 
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip
} from "recharts";

export default function ApplicationDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const router = useRouter();
  
  const [app, setApp] = useState<Application | null>(null);
  const [rollup, setRollup] = useState<ApplicationRollup | null>(null);
  const [repositories, setRepositories] = useState<ApplicationRepository[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    try {
      setError(null);
      setLoading(true);
      const [appData, rollupData, reposData] = await Promise.all([
        applicationService.getApplication(id),
        applicationService.getApplicationRollup(id),
        projectService.getApplicationRepositories(id),
      ]);
      setApp(appData);
      setRollup(rollupData);
      setRepositories(reposData);
    } catch (err) {
      setError(toUserErrorMessage(err, "Failed to load application details."));
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    const timer = setTimeout(() => {
      void loadData();
    }, 0);
    return () => clearTimeout(timer);
  }, [loadData]);

  if (loading) {
    return <PageLoading label="Loading application details..." />;
  }

  if (error || !app) {
    return (
      <div className="space-y-6">
        <PageError message={error || "The requested application could not be located."} onRetry={() => void loadData()} />
        <div>
          <button
            onClick={() => router.back()}
            className="px-6 py-2 rounded-xl bg-primary text-primary-foreground font-bold text-sm"
          >
            Go Back
          </button>
        </div>
      </div>
    );
  }

  const buildSuccessPct = ((rollup?.build_success_rate || 0) * 100).toFixed(1);
  const qualityScore = rollup?.quality_score?.toFixed(1) || "N/A";
  const criticalWarnings = rollup?.critical_warning_count || 0;
  const gateFailures = rollup?.quality_gate_failed_count || 0;
  const healthTitle =
    criticalWarnings > 0 || gateFailures > 0
      ? "Needs Attention"
      : "Optimal Health";
  const healthDescription =
    criticalWarnings > 0 || gateFailures > 0
      ? `Critical warnings ${criticalWarnings}, gate failures ${gateFailures}.`
      : "No critical roadblocks detected in current build cycle.";
  const qualityTrend = rollup?.quality_score !== undefined ? `${qualityScore}` : "N/A";

  return (
    <div className="space-y-10 pb-20">
      <div className="flex items-center gap-4">
        <button 
          onClick={() => router.back()}
          className="p-3 rounded-xl glass hover:bg-muted/30 transition-all text-muted-foreground group"
        >
          <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
        </button>
        <div className="flex-1">
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-3xl font-black text-foreground dark:text-primary-foreground tracking-tight">{app.name}</h1>
            <Badge variant={app.status === "active" ? "success" : "warning"} dot>{app.status}</Badge>
          </div>
          <p className="text-muted-foreground text-sm flex items-center gap-2">
            <Clock className="w-4 h-4" /> Updated {new Date(app.updated_at).toLocaleDateString()} • {app.key}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button className="p-3 rounded-xl glass border border-border/50 hover:bg-muted/30 text-muted-foreground transition-all">
            <RefreshCcw className="w-5 h-5" />
          </button>
          <button className="p-3 rounded-xl glass border border-border/50 hover:bg-muted/30 text-muted-foreground transition-all">
            <Settings className="w-5 h-5" />
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[
          { label: "Build Success", value: `${buildSuccessPct}%`, icon: Activity, color: "text-success", trend: "Current" },
          { label: "Quality Score", value: qualityScore, icon: ShieldCheck, color: "text-info", trend: qualityTrend },
          { label: "Critical Warnings", value: String(criticalWarnings), icon: Zap, color: criticalWarnings > 0 ? "text-destructive" : "text-success", trend: criticalWarnings > 0 ? "Open" : "None" },
          { label: "Gate Failures", value: String(gateFailures), icon: Globe, color: "text-purple-500", trend: gateFailures > 0 ? "Open" : "None" },
        ].map((stat, i) => (
          <motion.div 
            key={stat.label}
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card p-6 flex flex-col justify-between"
          >
            <div className="flex items-center justify-between mb-4">
              <div className={cn("p-2 rounded-xl bg-muted/30 border border-border", stat.color)}>
                <stat.icon className="w-5 h-5" />
              </div>
              <span className={cn("text-[10px] font-black uppercase tracking-tighter", 
                stat.trend.startsWith('+') ? "text-success" : 
                stat.trend.startsWith('-') ? "text-destructive" : 
                "text-muted-foreground"
              )}>
                {stat.trend}
              </span>
            </div>
            <div>
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em] mb-1">{stat.label}</p>
              <h3 className="text-2xl font-black text-foreground dark:text-primary-foreground">{stat.value}</h3>
            </div>
          </motion.div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <section className="lg:col-span-3 glass-card p-8 flex flex-col">
          <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
            <ShieldCheck className="w-4 h-4 text-primary" /> Quality Analysis
          </h3>
          <div className="space-y-6 flex-1">
            <div className="p-4 rounded-xl bg-primary/5 border border-primary/20">
              <p className="text-[10px] font-black text-primary uppercase tracking-widest mb-1">Status</p>
              <h4 className="text-lg font-bold text-foreground dark:text-primary-foreground">{healthTitle}</h4>
              <p className="text-xs text-muted-foreground mt-1">{healthDescription}</p>
            </div>
            
            <div className="space-y-4">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">PR Distribution</p>
              <div className="h-40">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={Object.entries(rollup?.pull_request_distribution || {}).map(([name, value]) => ({ name, value }))}>
                    <Bar dataKey="value" fill="var(--primary)" radius={[4, 4, 0, 0]} />
                    <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fill: "var(--muted-foreground)", fontSize: 10, fontWeight: 700 }} />
                    <YAxis axisLine={false} tickLine={false} tick={{ fill: "var(--muted-foreground)", fontSize: 10, fontWeight: 700 }} />
                    <Tooltip 
                      cursor={{fill: 'transparent'}}
                      contentStyle={{ backgroundColor: 'var(--card)', borderRadius: '12px', border: '1px solid var(--border)' }}
                    />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>
        </section>
      </div>

      <section className="glass-card overflow-hidden">
        <div className="p-8 border-b border-border/50">
          <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground flex items-center gap-2">
            <GitBranch className="w-5 h-5 text-muted-foreground" /> Linked Repositories
          </h3>
        </div>
        <div className="divide-y divide-border/50">
          {repositories.map((repo, i: number) => (
            <div key={i} className="p-6 flex items-center justify-between hover:bg-muted/5 transition-colors group">
              <div className="flex items-center gap-4">
                <div className="w-10 h-10 rounded-xl bg-muted/30 border border-border flex items-center justify-center group-hover:scale-110 transition-transform">
                  <Code2 className="w-5 h-5 text-muted-foreground" />
                </div>
                <div>
                  <h4 className="text-sm font-bold text-foreground dark:text-primary-foreground">{repo.repo_full_name}</h4>
                  <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">{repo.repo_provider} • {repo.role}</p>
                </div>
              </div>
              <div className="flex items-center gap-6">
                <Badge
                  variant={
                    repo.sync_status === "active"
                      ? "success"
                      : repo.sync_status === "degraded"
                        ? "warning"
                        : "secondary"
                  }
                >
                  {repo.sync_status}
                </Badge>
                <button className="p-2 rounded-lg hover:bg-muted/30 text-muted-foreground hover:text-primary transition-all">
                  <ExternalLink className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
          {repositories.length === 0 && (
            <div className="p-20 text-center text-muted-foreground text-sm">
              No repositories linked to this application.
            </div>
          )}
        </div>
      </section>
    </div>
  );
}
