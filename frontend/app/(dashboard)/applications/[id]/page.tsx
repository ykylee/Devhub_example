"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { 
  Activity, 
  ArrowLeft, 
  Box, 
  Globe, 
  ShieldCheck, 
  Zap,
  Clock,
  ExternalLink,
  Loader2,
  RefreshCcw,
  Settings,
  GitBranch,
  Code2
} from "lucide-react";
import { useRouter, useParams } from "next/navigation";
import { Badge } from "@/components/ui/Badge";
import { cn } from "@/lib/utils";
import { applicationService, Application, ApplicationRollup } from "@/lib/services/application.service";
import { 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  BarChart,
  Bar
} from "recharts";

// Mock historical data for charts
const mockHistoryData = [
  { name: "Day 1", build: 85, quality: 78 },
  { name: "Day 2", build: 88, quality: 80 },
  { name: "Day 3", build: 92, quality: 82 },
  { name: "Day 4", build: 80, quality: 75 },
  { name: "Day 5", build: 95, quality: 85 },
  { name: "Day 6", build: 98, quality: 88 },
  { name: "Day 7", build: 96, quality: 90 },
];

type ApplicationRepositorySummary = {
  repo_provider: string;
  repo_full_name: string;
  role: string;
  sync_status: string;
};

type ApplicationWithRepositories = Application & {
  repositories?: ApplicationRepositorySummary[];
};

export default function ApplicationDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const router = useRouter();
  
  const [app, setApp] = useState<ApplicationWithRepositories | null>(null);
  const [rollup, setRollup] = useState<ApplicationRollup | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadData = async () => {
      try {
        const [appData, rollupData] = await Promise.all([
          applicationService.getApplication(id),
          applicationService.getApplicationRollup(id)
        ]);
        setApp(appData);
        setRollup(rollupData);
      } catch (err) {
        setError("Failed to load application details.");
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
        <p className="text-muted-foreground animate-pulse font-black uppercase tracking-widest text-[10px]">Synchronizing Domain Data...</p>
      </div>
    );
  }

  if (error || !app) {
    return (
      <div className="text-center py-20 space-y-6">
        <div className="glass-card p-10 max-w-md mx-auto">
          <Box className="w-16 h-16 text-muted-foreground mx-auto mb-4 opacity-20" />
          <h2 className="text-xl font-bold text-foreground dark:text-primary-foreground mb-2">Resource Not Found</h2>
          <p className="text-muted-foreground text-sm mb-6">{error || "The requested application could not be located."}</p>
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
          { label: "Build Success", value: `${((rollup?.build_success_rate || 0) * 100).toFixed(1)}%`, icon: Activity, color: "text-emerald-500", trend: "+2.4%" },
          { label: "Quality Score", value: rollup?.quality_score?.toFixed(1) || "N/A", icon: ShieldCheck, color: "text-blue-500", trend: "+0.1" },
          { label: "Critical Warnings", value: rollup?.critical_warning_count.toString() || "0", icon: Zap, color: (rollup?.critical_warning_count || 0) > 0 ? "text-red-500" : "text-green-500", trend: "Stable" },
          { label: "Gate Failures", value: rollup?.quality_gate_failed_count.toString() || "0", icon: Globe, color: "text-purple-500", trend: "0" },
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
              <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground">Historical Performance</h3>
              <p className="text-xs text-muted-foreground">Build success and quality score trends over the last 7 days</p>
            </div>
          </div>
          <div className="h-80 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={mockHistoryData}>
                <defs>
                  <linearGradient id="colorBuild" x1="0" y1="0" x2="0" y2="1">
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
                  dy={10}
                />
                <YAxis hide />
                <Tooltip 
                  contentStyle={{ 
                    backgroundColor: 'var(--card)', 
                    borderRadius: '16px', 
                    border: '1px solid var(--border)',
                    boxShadow: '0 10px 30px rgba(0,0,0,0.1)',
                    backdropFilter: 'blur(10px)',
                  }}
                  itemStyle={{ fontSize: '10px', fontWeight: 900, textTransform: 'uppercase' }}
                  labelStyle={{ fontSize: '12px', fontWeight: 800, color: 'var(--foreground)' }}
                />
                <Area 
                  type="monotone" 
                  dataKey="build" 
                  stroke="var(--primary)" 
                  strokeWidth={3}
                  fillOpacity={1} 
                  fill="url(#colorBuild)" 
                />
                <Area 
                  type="monotone" 
                  dataKey="quality" 
                  stroke="#8b5cf6" 
                  strokeWidth={3}
                  fill="transparent"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </section>

        <section className="glass-card p-8 flex flex-col">
          <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
            <ShieldCheck className="w-4 h-4 text-primary" /> Quality Analysis
          </h3>
          <div className="space-y-6 flex-1">
            <div className="p-4 rounded-xl bg-primary/5 border border-primary/20">
              <p className="text-[10px] font-black text-primary uppercase tracking-widest mb-1">Status</p>
              <h4 className="text-lg font-bold text-foreground dark:text-primary-foreground">Optimal Health</h4>
              <p className="text-xs text-muted-foreground mt-1">No critical roadblocks detected in current build cycle.</p>
            </div>
            
            <div className="space-y-4">
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">PR Distribution</p>
              <div className="h-40">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={Object.entries(rollup?.pull_request_distribution || {}).map(([name, value]) => ({ name, value }))}>
                    <Bar dataKey="value" fill="var(--primary)" radius={[4, 4, 0, 0]} />
                    <XAxis dataKey="name" hide />
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
          {app.repositories?.map((repo, i: number) => (
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
                <Badge variant={repo.sync_status === "synced" ? "success" : "warning"}>{repo.sync_status}</Badge>
                <button className="p-2 rounded-lg hover:bg-muted/30 text-muted-foreground hover:text-primary transition-all">
                  <ExternalLink className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
          {(!app || !app.repositories || app.repositories.length === 0) && (
            <div className="p-20 text-center text-muted-foreground text-sm">
              No repositories linked to this application.
            </div>
          )}
        </div>
      </section>
    </div>
  );
}
