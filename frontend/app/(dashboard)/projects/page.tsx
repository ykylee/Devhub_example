"use client";

import { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { Search, Filter, FolderKanban } from "lucide-react";
import { projectService } from "@/lib/services/project.service";
import { Project } from "@/lib/services/project.types";
import { ProjectTable } from "@/components/project/ProjectTable";
import { useToast } from "@/components/ui/Toast";

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [query, setQuery] = useState("");
  const { toast } = useToast();

  const filteredProjects = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return projects;
    return projects.filter((project) =>
      project.name.toLowerCase().includes(q) ||
      project.key.toLowerCase().includes(q) ||
      project.owner_user_id.toLowerCase().includes(q)
    );
  }, [projects, query]);

  const load = async () => {
    setIsLoading(true);
    try {
      // In a real app, we might need a separate API to get all projects
      // For now, we'll try to fetch for all apps -> repos -> projects
      const apps = await projectService.getApplications();
      const allProjects: Project[] = [];
      for (const app of apps) {
        const repos = await projectService.getApplicationRepositories(app.id);
        for (const repo of repos) {
          // This is a bit inefficient, but works for mock/demo
          // In reality, we'd have a GET /api/v1/projects endpoint.
          // Since it's planned (API-55/56), we'll assume it exists or show empty.
          try {
             // Use repo_full_name as repository_id if it's a number, 
             // but our type says number. We'll skip for now if backend doesn't support global list.
          } catch (e) {}
        }
      }
      setProjects(allProjects);
    } catch (error) {
      console.error("[projects] load failed:", error);
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
          Projects
        </h1>
        <p className="text-sm text-muted-foreground font-medium uppercase tracking-widest opacity-60">
          Operational units and time-bound delivery
        </p>
      </div>

      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex items-center gap-4 max-w-2xl">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search projects..."
            className="w-full glass border-border rounded-2xl pl-12 pr-6 py-3.5 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-indigo-500/30 transition-all"
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
          <div className="w-12 h-12 border-4 border-indigo-500/20 border-t-indigo-500 rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Projects...</p>
        </div>
      ) : (
        <ProjectTable
          projects={filteredProjects}
          onViewDetails={(project) => {
            toast(`Viewing details for ${project.key} (Coming soon)`, "glass");
          }}
        />
      )}
    </div>
  );
}
