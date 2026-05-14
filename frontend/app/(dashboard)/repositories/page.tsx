"use client";

import { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { Search, Filter, GitBranch } from "lucide-react";
import { projectService } from "@/lib/services/project.service";
import { ApplicationRepository } from "@/lib/services/project.types";
import { RepositoryTable } from "@/components/project/RepositoryTable";
import { useToast } from "@/components/ui/Toast";

export default function RepositoriesPage() {
  const [repositories, setRepositories] = useState<ApplicationRepository[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [query, setQuery] = useState("");
  const { toast } = useToast();

  const filteredRepos = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return repositories;
    return repositories.filter((repo) =>
      repo.repo_full_name.toLowerCase().includes(q) ||
      repo.repo_provider.toLowerCase().includes(q)
    );
  }, [repositories, query]);

  const load = async () => {
    setIsLoading(true);
    try {
      // In a real app, we might need a separate API to get all repos
      // or fetch per application and flatten. 
      // For now, we'll try to fetch for all apps or show empty if no global API.
      const apps = await projectService.getApplications();
      const allRepos: ApplicationRepository[] = [];
      for (const app of apps) {
        const repos = await projectService.getApplicationRepositories(app.id);
        allRepos.push(...repos);
      }
      setRepositories(allRepos);
    } catch (error) {
      console.error("[repositories] load failed:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  return (
    <div className="space-y-8 p-8">
      <div className="flex flex-col gap-2 mb-6">
        <h1 className="text-3xl font-black text-foreground tracking-tight uppercase">
          Repositories
        </h1>
        <p className="text-sm text-muted-foreground font-medium uppercase tracking-widest opacity-60">
          Execution units and pipeline status
        </p>
      </div>

      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex items-center gap-4 max-w-2xl">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search repositories..."
            className="w-full glass border-border rounded-2xl pl-12 pr-6 py-3.5 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-pink-500/30 transition-all"
          />
        </div>
        <button
          type="button"
          className="glass border-border p-3.5 rounded-2xl text-muted-foreground hover:bg-muted/40 transition-all"
        >
          <Filter className="w-5 h-5" />
        </button>
      </motion.div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-pink-500/20 border-t-pink-500 rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Repositories...</p>
        </div>
      ) : (
        <RepositoryTable
          repositories={filteredRepos}
          showApplicationColumn
        />
      )}
    </div>
  );
}
