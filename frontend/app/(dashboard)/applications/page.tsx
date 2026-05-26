"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { 
  Activity, 
  Box, 
  Cpu, 
  Globe, 
  ShieldCheck, 
  Zap,
  ExternalLink
} from "lucide-react";
import Link from "next/link";
import { DashboardHeader } from "@/components/ui/DashboardHeader";
import { Badge } from "@/components/ui/Badge";
import { FilterBar } from "@/components/ui/FilterBar";
import { PageEmpty, PageError, PageLoading } from "@/components/ui/PageState";
import { applicationService, Application, ApplicationRollup } from "@/lib/services/application.service";

interface ApplicationWithRollup extends Application {
  rollup?: ApplicationRollup;
}

export default function ApplicationsStatusPage() {
  const [apps, setApps] = useState<ApplicationWithRollup[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");

  const loadData = async () => {
    try {
      setError(null);
      setLoading(true);
      const fetchedApps = await applicationService.listApplications();
      const appsWithRollups = await Promise.all(
        fetchedApps.map(async (app) => {
          try {
            const rollup = await applicationService.getApplicationRollup(app.id);
            return { ...app, rollup };
          } catch (err) {
            console.error(`Failed to fetch rollup for ${app.id}:`, err);
            return app;
          }
        })
      );
      setApps(appsWithRollups);
    } catch (err) {
      setError("Failed to load applications data.");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      void loadData();
    }, 0);
    return () => clearTimeout(timer);
  }, []);

  const filteredApps = apps.filter((app) => {
    const matchesSearch = app.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
                         app.key.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = statusFilter === "all" || app.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const totalApps = apps.length;
  const avgSuccessRate = apps.length > 0 
    ? (apps.reduce((acc, app) => acc + (app.rollup?.build_success_rate || 0), 0) / apps.length * 100).toFixed(1)
    : "0";
  const totalCritical = apps.reduce((acc, app) => acc + (app.rollup?.critical_warning_count || 0), 0);

  if (loading) {
    return <PageLoading label="Loading applications..." />;
  }

  return (
    <div className="space-y-10 pb-20">
      <DashboardHeader 
        titlePrefix="Application"
        titleGradient="Status (어플리케이션 현황)"
        subtitle="Real-time monitoring of all production and staging application services."
      />

      {error && <PageError message={error} onRetry={() => void loadData()} />}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[
          { label: "Total Applications", value: totalApps.toString(), icon: Box, color: "text-info" },
          { label: "Avg. Build Success", value: `${avgSuccessRate}%`, icon: Activity, color: "text-success" },
          { label: "Critical Warnings", value: totalCritical.toString(), icon: ShieldCheck, color: totalCritical > 0 ? "text-destructive" : "text-success" },
          { label: "Active Regions", value: "Global", icon: Globe, color: "text-purple-500" },
        ].map((stat, i) => (
          <motion.div 
            key={stat.label}
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card p-6"
          >
            <div className="flex items-center justify-between mb-4">
              <div className={`p-2 rounded-xl bg-muted/30 border border-border ${stat.color}`}>
                <stat.icon className="w-5 h-5" />
              </div>
            </div>
            <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">{stat.label}</p>
            <h3 className="text-2xl font-black text-foreground dark:text-primary-foreground">{stat.value}</h3>
          </motion.div>
        ))}
      </div>

      <FilterBar 
        onSearch={setSearchQuery}
        onFilterChange={setStatusFilter}
        activeFilter={statusFilter}
        filterOptions={[
          { label: "All Status", value: "all" },
          { label: "Active", value: "active" },
          { label: "Planning", value: "planning" },
          { label: "On Hold", value: "on_hold" },
          { label: "Archived", value: "archived" },
        ]}
        placeholder="Search applications by name or key..."
      />

      <div className="grid gap-6">
        {filteredApps.map((app, i) => (
          <motion.div
            key={app.id}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card p-6 flex flex-col md:flex-row items-center justify-between gap-6 group hover:bg-muted/10"
          >
            <div className="flex items-center gap-6 flex-1">
              <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-primary/20 to-accent/20 flex items-center justify-center ring-1 ring-border group-hover:scale-110 transition-transform">
                <Zap className="w-6 h-6 text-primary" />
              </div>
              <div>
                <div className="flex items-center gap-3 mb-1">
                  <Link href={`/applications/${app.id}`} className="hover:underline decoration-primary underline-offset-4 decoration-2">
                    <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground">{app.name}</h3>
                  </Link>
                  <Badge variant={app.status === "active" ? "success" : "warning"} dot>{app.status}</Badge>
                </div>
                <div className="flex items-center gap-4 text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                  <span className="flex items-center gap-1"><Cpu className="w-3 h-3" /> Build: {((app.rollup?.build_success_rate || 0) * 100).toFixed(1)}%</span>
                  <span>•</span>
                  <span className="flex items-center gap-1"><ShieldCheck className="w-3 h-3" /> Quality: {app.rollup?.quality_score?.toFixed(1) || "N/A"}</span>
                  <span>•</span>
                  <span>{app.key}</span>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-12 text-right">
              <div>
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Build Success</p>
                <p className="text-lg font-black text-foreground dark:text-primary-foreground">
                  {app.rollup ? `${(app.rollup.build_success_rate * 100).toFixed(1)}%` : "N/A"}
                </p>
              </div>
              <div>
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Quality Score</p>
                <p className="text-sm font-mono font-bold text-accent">
                  {app.rollup?.quality_score?.toFixed(2) || "N/A"}
                </p>
              </div>
              <Link href={`/applications/${app.id}`} className="p-3 rounded-xl hover:bg-muted/30 transition-all text-muted-foreground hover:text-primary">
                <ExternalLink className="w-5 h-5" />
              </Link>
            </div>
          </motion.div>
        ))}
        {filteredApps.length === 0 && !loading && (
          <PageEmpty message="No applications matching your filters" />
        )}
      </div>
    </div>
  );
}
