"use client";

import { useCallback, useEffect, useState } from "react";
import { motion } from "framer-motion";
import { 
  Activity, 
  ArrowLeft, 
  GitBranch, 
  GitPullRequest, 
  Globe, 
  Users,
  ExternalLink,
  Lock,
  Unlock,
  ShieldCheck
} from "lucide-react";
import { useRouter, useParams } from "next/navigation";
import { Badge } from "@/components/ui/Badge";
import { cn } from "@/lib/utils";
import { repositoryService, Repository, RepositoryActivity } from "@/lib/services/repository.service";
import { toUserErrorMessage } from "@/lib/services/error-message";
import { PageError, PageLoading } from "@/components/ui/PageState";

export default function RepositoryDetailPage() {
  const params = useParams();
  const idStr = params.id as string;
  const id = parseInt(idStr, 10);
  const router = useRouter();
  
  const [repo, setRepo] = useState<Repository | null>(null);
  const [activity, setActivity] = useState<RepositoryActivity | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    try {
      setError(null);
      setLoading(true);
      const [repoData, activityData] = await Promise.all([
        repositoryService.getRepository(id),
        repositoryService.getRepositoryActivity(id)
      ]);
      if (!repoData) throw new Error("Repository not found.");
      setRepo(repoData);
      setActivity(activityData);
    } catch (err) {
      setError(toUserErrorMessage(err, "Failed to load repository details."));
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
    return <PageLoading label="Loading repository details..." />;
  }

  if (error || !repo) {
    return (
      <div className="space-y-6">
        <PageError message={error || "The requested source repository could not be located."} onRetry={() => void loadData()} />
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

  const prEvents = activity?.pr_event_count ?? 0;
  const buildRuns = activity?.build_run_count ?? 0;
  const contributors = activity?.active_contributors.length ?? 0;
  const buildSuccessPct = ((activity?.build_success_rate || 0) * 100).toFixed(1);

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
            <h1 className="text-3xl font-black text-foreground dark:text-primary-foreground tracking-tight">{repo.name}</h1>
            <Badge variant={repo.private ? "secondary" : "glass"}>
              {repo.private ? <Lock className="w-3 h-3 mr-1" /> : <Unlock className="w-3 h-3 mr-1" />}
              {repo.private ? "Private" : "Public"}
            </Badge>
          </div>
          <p className="text-muted-foreground text-sm flex items-center gap-2">
            <Globe className="w-4 h-4" /> {repo.full_name} • <GitBranch className="w-4 h-4 ml-2" /> {repo.default_branch}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <a 
            href={repo.html_url} 
            target="_blank" 
            rel="noopener noreferrer"
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-muted/30 border border-border text-xs font-bold hover:bg-muted/50 transition-all"
          >
            <ExternalLink className="w-4 h-4" /> View on SCM
          </a>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[
          { label: "PR Events", value: String(prEvents), icon: GitPullRequest, color: "text-info", trend: "Current" },
          { label: "Build Runs", value: String(buildRuns), icon: Activity, color: "text-foreground", trend: "Current" },
          { label: "Build Success", value: `${buildSuccessPct}%`, icon: ShieldCheck, color: "text-success", trend: "Current" },
          { label: "Contributors", value: String(contributors), icon: Users, color: "text-purple-500", trend: "Current" },
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
        <section className="lg:col-span-2 glass-card p-8">
          <div className="flex items-center justify-between mb-8">
            <div>
              <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground">Activity Window</h3>
              <p className="text-xs text-muted-foreground">Current backend-provided activity summary for this repository.</p>
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="rounded-2xl border border-border/60 bg-muted/10 p-5">
              <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-2">From</p>
              <p className="text-sm font-bold text-foreground dark:text-primary-foreground">
                {activity?.window_from ? new Date(activity.window_from).toLocaleString() : "-"}
              </p>
            </div>
            <div className="rounded-2xl border border-border/60 bg-muted/10 p-5">
              <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-2">To</p>
              <p className="text-sm font-bold text-foreground dark:text-primary-foreground">
                {activity?.window_to ? new Date(activity.window_to).toLocaleString() : "-"}
              </p>
            </div>
            <div className="rounded-2xl border border-border/60 bg-muted/10 p-5">
              <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-2">PR Events</p>
              <p className="text-xl font-black text-foreground dark:text-primary-foreground">{activity?.pr_event_count ?? 0}</p>
            </div>
            <div className="rounded-2xl border border-border/60 bg-muted/10 p-5">
              <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground mb-2">Build Runs</p>
              <p className="text-xl font-black text-foreground dark:text-primary-foreground">{activity?.build_run_count ?? 0}</p>
            </div>
          </div>
        </section>

        <div className="space-y-8">
          <section className="glass-card p-8">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
              <Users className="w-4 h-4 text-primary" /> Top Contributors
            </h3>
            <div className="space-y-6">
              {(activity?.active_contributors || []).map((user, i) => (
                <div key={i} className="flex items-center justify-between group cursor-pointer">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-muted/40 border border-border group-hover:border-primary/50 transition-all" />
                    <span className="text-sm font-bold text-foreground dark:text-primary-foreground group-hover:text-primary transition-colors">{user}</span>
                  </div>
                  <Badge variant="glass" className="opacity-50 group-hover:opacity-100 transition-opacity">
                    Active
                  </Badge>
                </div>
              ))}
              {(!activity?.active_contributors || activity.active_contributors.length === 0) && (
                <p className="text-sm text-muted-foreground">No active contributors in current window.</p>
              )}
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
