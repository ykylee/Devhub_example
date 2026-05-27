"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { 
  GitBranch, 
  GitPullRequest, 
  Globe, 
  ExternalLink,
  Code2,
  Users,
  Activity
} from "lucide-react";
import Link from "next/link";
import { DashboardHeader } from "@/components/ui/DashboardHeader";
import { Badge } from "@/components/ui/Badge";
import { FilterBar } from "@/components/ui/FilterBar";
import { PageEmpty, PageError, PageLoading } from "@/components/ui/PageState";
import { repositoryService, Repository, RepositoryActivity, RepositoryBuildRun } from "@/lib/services/repository.service";

interface RepositoryWithActivity extends Repository {
  activity?: RepositoryActivity;
  latestBuild?: RepositoryBuildRun;
  unresolvedFailedBranches?: string[];
  buildHealth?: "success" | "failed" | "unknown";
}

function evaluateBuildHealth(runs: RepositoryBuildRun[]): { buildHealth: "success" | "failed" | "unknown"; unresolvedFailedBranches: string[] } {
  const terminalByBranch = new Map<string, "success" | "failed" | "unknown">();
  for (const run of runs) {
    const branch = run.branch?.trim() || "(unknown)";
    if (terminalByBranch.has(branch)) continue;
    if (run.status === "success") {
      terminalByBranch.set(branch, "success");
      continue;
    }
    if (run.status === "failed") {
      terminalByBranch.set(branch, "failed");
      continue;
    }
    if (run.status === "cancelled" || run.status === "skipped" || run.status === "unknown") {
      terminalByBranch.set(branch, "unknown");
    }
  }

  const unresolvedFailedBranches = [...terminalByBranch.entries()]
    .filter(([, status]) => status === "failed")
    .map(([branch]) => branch);

  if (unresolvedFailedBranches.length > 0) {
    return { buildHealth: "failed", unresolvedFailedBranches };
  }
  const hasSuccess = [...terminalByBranch.values()].some((status) => status === "success");
  if (hasSuccess) {
    return { buildHealth: "success", unresolvedFailedBranches: [] };
  }
  return { buildHealth: "unknown", unresolvedFailedBranches: [] };
}

export default function RepositoriesStatusPage() {
  const [repos, setRepos] = useState<RepositoryWithActivity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [visibilityFilter, setVisibilityFilter] = useState("all");

  const loadData = async () => {
    try {
      setError(null);
      setLoading(true);
      const fetchedRepos = await repositoryService.listRepositories();
      const reposWithActivity = await Promise.all(
        fetchedRepos.map(async (repo) => {
          try {
            const [activity, recentRuns] = await Promise.all([
              repositoryService.getRepositoryActivity(repo.id),
              repositoryService.getRepositoryBuildRuns(repo.id, { limit: 50 }),
            ]);
            const { buildHealth, unresolvedFailedBranches } = evaluateBuildHealth(recentRuns);
            return { ...repo, activity, latestBuild: recentRuns[0], buildHealth, unresolvedFailedBranches };
          } catch (err) {
            console.error(`Failed to fetch activity for ${repo.id}:`, err);
            return repo;
          }
        })
      );
      setRepos(reposWithActivity);
    } catch (err) {
      setError("Failed to load repositories data.");
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

  const filteredRepos = repos.filter((repo) => {
    const matchesSearch = repo.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
                         repo.owner_login.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesVisibility = visibilityFilter === "all" || 
                             (visibilityFilter === "private" ? repo.private : !repo.private);
    return matchesSearch && matchesVisibility;
  });

  const totalRepos = repos.length;
  const activePRs = repos.reduce((acc, repo) => acc + (repo.activity?.pr_event_count || 0), 0);
  const totalContributors = new Set(repos.flatMap(repo => repo.activity?.active_contributors || [])).size;
  const latestBuildSuccessCount = repos.filter((repo) => repo.buildHealth === "success").length;
  const latestBuildFailureCount = repos.filter((repo) => repo.buildHealth === "failed").length;
  const latestBuildUnknownCount = repos.filter((repo) => repo.buildHealth !== "success" && repo.buildHealth !== "failed").length;

  if (loading) {
    return <PageLoading label="Loading repositories..." />;
  }

  return (
    <div className="space-y-10 pb-20">
      <DashboardHeader 
        titlePrefix="Repository"
        titleGradient="Activity (저장소 활동성)"
        subtitle="Operational status and activity metrics across all integrated SCM repositories."
      />

      {error && <PageError message={error} onRetry={() => void loadData()} />}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[
          { label: "Total Repositories", value: totalRepos.toString(), icon: Globe, color: "text-foreground" },
          { label: "Active PRs (30d)", value: activePRs.toString(), icon: GitPullRequest, color: "text-info" },
          { label: "Total Contributors", value: totalContributors.toString(), icon: Users, color: "text-purple-500" },
          { label: "Latest Build Success", value: latestBuildSuccessCount.toString(), icon: Activity, color: "text-success" },
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
        onFilterChange={setVisibilityFilter}
        activeFilter={visibilityFilter}
        filterOptions={[
          { label: "All Repos", value: "all" },
          { label: "Public", value: "public" },
          { label: "Private", value: "private" },
        ]}
        placeholder="Search repositories by name or owner..."
      />

      <div className="grid gap-6">
        {filteredRepos.map((repo, i) => (
          <motion.div
            key={repo.id}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card p-6 flex flex-col md:flex-row items-center justify-between gap-6 group hover:bg-muted/10"
          >
            <div className="flex items-center gap-6 flex-1">
              <div className="w-12 h-12 rounded-2xl bg-muted/30 border border-border flex items-center justify-center group-hover:scale-110 transition-transform">
                <Code2 className="w-6 h-6 text-muted-foreground group-hover:text-primary" />
              </div>
              <div>
                <div className="flex items-center gap-3 mb-1">
                  <Link href={`/repositories/${repo.id}`} className="hover:underline decoration-primary underline-offset-4 decoration-2">
                    <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground">{repo.name}</h3>
                  </Link>
                  <Badge variant={repo.private ? "secondary" : "glass"}>{repo.private ? "Private" : "Public"}</Badge>
                </div>
                <div className="flex items-center gap-4 text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                  <span className="flex items-center gap-1"><GitBranch className="w-3 h-3" /> {repo.default_branch}</span>
                  <span>•</span>
                  <span className="flex items-center gap-1"><Users className="w-3 h-3" /> {repo.activity?.active_contributors.length || 0} contributors</span>
                  <span>•</span>
                  <span>{repo.owner_login}</span>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-12 text-right">
              <div>
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">PR Events</p>
                <p className="text-lg font-black text-foreground dark:text-primary-foreground">{repo.activity?.pr_event_count || 0}</p>
              </div>
              <div>
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Latest Build</p>
                <p className={`text-sm font-mono font-bold ${
                  repo.buildHealth === "success"
                    ? "text-success"
                    : repo.buildHealth === "failed"
                      ? "text-destructive"
                      : "text-muted-foreground"
                }`}>
                  {repo.buildHealth === "success" ? "Success" : repo.buildHealth === "failed" ? "Failed" : "N/A"}
                </p>
                {repo.buildHealth === "failed" && (repo.unresolvedFailedBranches?.length ?? 0) > 0 && (
                  <p className="text-[10px] text-destructive/80">
                    unresolved: {repo.unresolvedFailedBranches?.join(", ")}
                  </p>
                )}
              </div>
              <div className="flex items-center gap-2">
                <Link 
                  href={`/repositories/${repo.id}`}
                  className="p-3 rounded-xl hover:bg-primary/10 text-primary transition-all"
                  title="Internal Report"
                >
                  <Activity className="w-5 h-5" />
                </Link>
                <a 
                  href={repo.html_url} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="p-3 rounded-xl hover:bg-muted/30 transition-all text-muted-foreground hover:text-primary"
                  title="SCM Link"
                >
                  <ExternalLink className="w-5 h-5" />
                </a>
              </div>
            </div>
          </motion.div>
        ))}
        {filteredRepos.length === 0 && !loading && (
          <PageEmpty message="No repositories matching your filters" />
        )}
      </div>
      <p className="text-[10px] text-muted-foreground uppercase tracking-widest">
        Latest Build Summary: Success {latestBuildSuccessCount} / Failed {latestBuildFailureCount} / Unknown {latestBuildUnknownCount}
      </p>
    </div>
  );
}
