"use client";

import { useState, useEffect, useCallback } from "react";
import { 
  Activity, 
  Terminal, 
  LayoutDashboard, 
  SlidersHorizontal,
  Loader2,
  ArrowLeft,
  Lock,
  Unlock,
  Globe,
  GitBranch,
  ExternalLink
} from "lucide-react";
import { useRouter } from "next/navigation";
import { useStore } from "@/lib/store";
import { 
  repositoryService, 
  Repository, 
  RepositoryActivity, 
  RepositoryDashboardData 
} from "@/domain/repository-integration/service/repository.service";
import { PageError } from "@/shared/ui-foundation/components/PageState";
import { Badge } from "@/shared/ui-foundation/components/Badge";
import { DeveloperView } from "./DeveloperView";
import { ManagerView } from "./ManagerView";

interface RepositoryDashboardViewProps {
  repoId: number;
}

export function RepositoryDashboardView({ repoId }: RepositoryDashboardViewProps) {
  const { role } = useStore();
  const router = useRouter();
  
  const [repo, setRepo] = useState<Repository | null>(null);
  const [activity, setActivity] = useState<RepositoryActivity | null>(null);
  const [dashboardData, setDashboardData] = useState<RepositoryDashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Decide active view mode: "developer" | "manager"
  const [viewMode, setViewMode] = useState<"developer" | "manager">("developer");

  // Initial role detection to set default tab
  useEffect(() => {
    if (role === "Developer") {
      setViewMode("developer");
    } else if (role === "Manager" || role === "System Admin") {
      setViewMode("manager");
    }
  }, [role]);

  const loadData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      
      const [repoData, activityData, extraData] = await Promise.all([
        repositoryService.getRepository(repoId),
        repositoryService.getRepositoryActivity(repoId),
        repositoryService.getRepositoryDashboardData(repoId)
      ]);
      
      if (!repoData) throw new Error("Repository not found.");
      
      setRepo(repoData);
      setActivity(activityData);
      setDashboardData(extraData);
    } catch (err) {
      console.error(err);
      setError("Failed to fetch repository dashboard metrics. Please check network logs.");
    } finally {
      setLoading(false);
    }
  }, [repoId]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  if (loading) {
    return (
      <div className="h-96 flex flex-col items-center justify-center gap-3">
        <Loader2 className="w-10 h-10 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground animate-pulse font-bold">Aggregating workspace telemetry...</p>
      </div>
    );
  }

  if (error || !repo || !dashboardData) {
    return (
      <div className="space-y-6">
        <PageError message={error || "Failed to load repository details."} onRetry={() => void loadData()} />
      </div>
    );
  }

  return (
    <div className="space-y-10 pb-20">
      {/* Repository Metadata Header */}
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
          <p className="text-muted-foreground text-xs flex items-center gap-2">
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

      {/* SubHeader view controllers */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 glass border-border rounded-2xl bg-muted/10">
        <div>
          <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Dashboard Perspective</span>
          <h2 className="text-sm font-bold text-foreground dark:text-primary-foreground mt-0.5">
            Viewing as <span className="text-primary">{role || "Developer"}</span>
          </h2>
        </div>

        {/* Tab switch buttons */}
        <div className="flex items-center gap-2 p-1 bg-muted/20 border border-border/60 rounded-xl">
          <button
            onClick={() => setViewMode("developer")}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-black uppercase tracking-wider transition-all ${
              viewMode === "developer"
                ? "bg-primary text-primary-foreground shadow-lg shadow-primary/20"
                : "text-muted-foreground hover:bg-muted/30"
            }`}
          >
            <Terminal className="w-3.5 h-3.5" /> Developer
          </button>
          <button
            onClick={() => setViewMode("manager")}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-black uppercase tracking-wider transition-all ${
              viewMode === "manager"
                ? "bg-primary text-primary-foreground shadow-lg shadow-primary/20"
                : "text-muted-foreground hover:bg-muted/30"
            }`}
          >
            <LayoutDashboard className="w-3.5 h-3.5" /> Manager & Governance
          </button>
        </div>
      </div>

      {/* Render sub view conditionally with transition animation */}
      <div className="min-h-[500px]">
        {viewMode === "developer" ? (
          <DeveloperView repo={repo} activity={activity} dashboardData={dashboardData} />
        ) : (
          <ManagerView repo={repo} activity={activity} dashboardData={dashboardData} />
        )}
      </div>
    </div>
  );
}
