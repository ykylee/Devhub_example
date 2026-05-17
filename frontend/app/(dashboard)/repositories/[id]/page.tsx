"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { 
  Activity, 
  ArrowLeft, 
  Code2, 
  GitBranch, 
  GitCommit, 
  GitPullRequest, 
  Globe, 
  ShieldCheck, 
  Users,
  ExternalLink,
  Loader2,
  Lock,
  Unlock,
  History,
  AlertCircle
} from "lucide-react";
import Link from "next/link";
import { useRouter, useParams } from "next/navigation";
import { DashboardHeader } from "@/components/ui/DashboardHeader";
import { Badge } from "@/components/ui/Badge";
import { cn } from "@/lib/utils";
import { repositoryService, Repository, RepositoryActivity } from "@/lib/services/repository.service";
import { 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  BarChart,
  Bar,
  Cell
} from "recharts";

// Mock historical activity
const mockActivityData = [
  { name: "Mon", commits: 12, prs: 2 },
  { name: "Tue", commits: 18, prs: 5 },
  { name: "Wed", commits: 15, prs: 3 },
  { name: "Thu", commits: 25, prs: 8 },
  { name: "Fri", commits: 22, prs: 4 },
  { name: "Sat", commits: 5, prs: 1 },
  { name: "Sun", commits: 3, prs: 0 },
];

export default function RepositoryDetailPage() {
  const params = useParams();
  const idStr = params.id as string;
  const id = parseInt(idStr, 10);
  const router = useRouter();
  
  const [repo, setRepo] = useState<Repository | null>(null);
  const [activity, setActivity] = useState<RepositoryActivity | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadData = async () => {
      try {
        const [repoData, activityData] = await Promise.all([
          repositoryService.getRepository(id),
          repositoryService.getRepositoryActivity(id)
        ]);
        if (!repoData) throw new Error("Repository not found.");
        setRepo(repoData);
        setActivity(activityData);
      } catch (err) {
        setError("Failed to load repository details.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, [id]);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center h-[60vh] gap-4">
        <Loader2 className="w-10 h-10 text-primary animate-spin" />
        <p className="text-muted-foreground animate-pulse font-black uppercase tracking-widest text-[10px]">Mapping Source Intelligence...</p>
      </div>
    );
  }

  if (error || !repo) {
    return (
      <div className="text-center py-20 space-y-6">
        <div className="glass-card p-10 max-w-md mx-auto">
          <Globe className="w-16 h-16 text-muted-foreground mx-auto mb-4 opacity-20" />
          <h2 className="text-xl font-bold text-foreground dark:text-primary-foreground mb-2">Repository Not Found</h2>
          <p className="text-muted-foreground text-sm mb-6">{error || "The requested source repository could not be located."}</p>
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
          { label: "Commit Activity", value: "142", icon: GitCommit, color: "text-foreground", trend: "+12%" },
          { label: "Active PRs", value: activity?.pr_event_count.toString() || "0", icon: GitPullRequest, color: "text-blue-500", trend: "+2" },
          { label: "Build Success", value: `${((activity?.build_success_rate || 0) * 100).toFixed(1)}%`, icon: Activity, color: "text-emerald-500", trend: "Stable" },
          { label: "Contributors", value: activity?.active_contributors.length.toString() || "0", icon: Users, color: "text-purple-500", trend: "Top 1%" },
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
                stat.trend.startsWith('+') ? "text-emerald-500" : 
                stat.trend.startsWith('-') ? "text-rose-500" : 
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
              <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground">Activity Timeline</h3>
              <p className="text-xs text-muted-foreground">Detailed view of commits and PR events over the past week</p>
            </div>
          </div>
          <div className="h-80 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={mockActivityData}>
                <defs>
                  <linearGradient id="colorCommits" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="var(--primary)" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="var(--primary)" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border)" opacity={0.3} />
                <XAxis 
                  dataKey="name" 
                  axisLine={false} 
                  tickLine={false} 
                  tick={{ fill: 'var(--muted-foreground)', fontSize: 10, fontWeight: 700 }}
                />
                <YAxis hide />
                <Tooltip 
                  contentStyle={{ backgroundColor: 'var(--card)', borderRadius: '16px', border: '1px solid var(--border)', boxShadow: '0 10px 30px rgba(0,0,0,0.1)' }}
                />
                <Area 
                  type="monotone" 
                  dataKey="commits" 
                  stroke="var(--primary)" 
                  strokeWidth={3}
                  fillOpacity={1} 
                  fill="url(#colorCommits)" 
                />
                <Area 
                  type="monotone" 
                  dataKey="prs" 
                  stroke="#3b82f6" 
                  strokeWidth={2}
                  fill="transparent"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </section>

        <div className="space-y-8">
          <section className="glass-card p-8">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
              <Users className="w-4 h-4 text-primary" /> Top Contributors
            </h3>
            <div className="space-y-6">
              {(activity?.active_contributors || ["YK Lee", "Alex K.", "Sam J."]).map((user, i) => (
                <div key={i} className="flex items-center justify-between group cursor-pointer">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-muted/40 border border-border group-hover:border-primary/50 transition-all" />
                    <span className="text-sm font-bold text-foreground dark:text-primary-foreground group-hover:text-primary transition-colors">{user}</span>
                  </div>
                  <Badge variant="glass" className="opacity-50 group-hover:opacity-100 transition-opacity">
                    {Math.floor(Math.random() * 50) + 10} Commits
                  </Badge>
                </div>
              ))}
            </div>
          </section>

          <section className="glass-card p-8 border-rose-500/10">
            <h3 className="text-sm font-black uppercase tracking-widest text-rose-500/50 mb-6 flex items-center gap-2">
              <ShieldCheck className="w-4 h-4 text-rose-500" /> Security Status
            </h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between p-3 rounded-xl bg-rose-500/5 border border-rose-500/20">
                <div className="flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 text-rose-500" />
                  <span className="text-xs font-bold text-rose-500">2 Critical Vulnerabilities</span>
                </div>
                <button className="text-[10px] font-black uppercase text-rose-500 hover:underline">Fix</button>
              </div>
              <p className="text-[10px] text-muted-foreground leading-relaxed">
                Automated scan detected dependency vulnerabilities in `package.json`. Immediate patching recommended.
              </p>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
