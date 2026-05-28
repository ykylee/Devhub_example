"use client";

import { useEffect, useState, useCallback } from "react";
import { motion } from "framer-motion";
import {
  Briefcase, 
  Calendar, 
  CheckCircle2, 
  Clock, 
  Layout, 
  MoreHorizontal,
  Target,
  Users,
  Trash2,
} from "lucide-react";
import Link from "next/link";
import { DashboardHeader } from "@/components/ui/DashboardHeader";
import { Badge } from "@/components/ui/Badge";
import { FilterBar } from "@/components/ui/FilterBar";
import { PageEmpty, PageError, PageLoading } from "@/components/ui/PageState";
import { projectService } from "@/lib/services/project.service";
import type { Project } from "@/lib/services/project.types";
import { repositoryService } from "@/lib/services/repository.service";

export default function ProjectsStatusPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");

  const refresh = useCallback(async () => {
    try {
      setError(null);
      setLoading(true);
      const repos = await repositoryService.listRepositories();
      const allProjects = await projectService.listAllProjects(repos.map(r => r.id), { include_archived: statusFilter === "archived" });
      setProjects(allProjects);
    } catch (err) {
      setError("Failed to load projects data.");
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [statusFilter]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const handleDelete = useCallback(async (projectId: string, currentStatus: string) => {
    const isArchived = currentStatus === "archived";
    const confirmMsg = isArchived
      ? "Are you sure you want to permanently delete this project? This action cannot be undone."
      : "Are you sure you want to archive this project?";
    
    if (!confirm(confirmMsg)) return;

    try {
      setError(null);
      await projectService.archiveProject(projectId, isArchived);
      await refresh();
    } catch (err) {
      setError(isArchived ? "Failed to delete project." : "Failed to archive project.");
      console.error(err);
    }
  }, [refresh]);

  const filteredProjects = projects.filter((project) => {
    const matchesSearch = project.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
                         project.key.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         (project.description || "").toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = statusFilter === "all" || project.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const totalProjects = projects.length;
  const activeProjects = projects.filter(p => p.status === "active").length;
  const planningProjects = projects.filter(p => p.status === "planning").length;
  const closedProjects = projects.filter(p => p.status === "closed" || p.status === "archived").length;

  if (loading) {
    return <PageLoading label="Loading projects..." />;
  }

  return (
    <div className="space-y-10 pb-20">
      <DashboardHeader 
        titlePrefix="Project"
        titleGradient="Milestones (과제 현황)"
        subtitle="Tracking development projects, milestones, and delivery timelines."
      />

      {error && <PageError message={error} onRetry={() => void refresh()} />}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[
          { label: "Total Projects", value: totalProjects.toString(), icon: Briefcase, color: "text-info" },
          { label: "Active", value: activeProjects.toString(), icon: Target, color: "text-success" },
          { label: "Planning", value: planningProjects.toString(), icon: Clock, color: "text-warning" },
          { label: "Completed", value: closedProjects.toString(), icon: CheckCircle2, color: "text-purple-500" },
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
          { label: "All Projects", value: "all" },
          { label: "Active", value: "active" },
          { label: "Planning", value: "planning" },
          { label: "On Hold", value: "on_hold" },
          { label: "Archived", value: "archived" },
          { label: "Closed", value: "closed" },
        ]}
        placeholder="Search projects by name, key, or description..."
      />

      <div className="grid gap-6">
        {filteredProjects.map((project, i) => (
          <motion.div
            key={project.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.05 }}
            className="glass-card p-6 group hover:border-primary/50 transition-colors"
          >
            <div className="flex flex-col md:flex-row gap-6">
              <div className="flex-1">
                <div className="flex items-center gap-3 mb-2">
                  <Badge variant={
                    project.status === "active" ? "success" : 
                    project.status === "planning" ? "warning" : 
                    "secondary"
                  } dot>{project.status}</Badge>
                  <Link href={`/projects/${project.id}`} className="hover:underline decoration-primary underline-offset-4 decoration-2">
                    <h3 className="text-xl font-bold text-foreground dark:text-primary-foreground">{project.name}</h3>
                  </Link>
                </div>
                <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
                  {project.description || "No description provided."}
                </p>
                <div className="flex flex-wrap items-center gap-6 text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                  <span className="flex items-center gap-1.5"><Users className="w-3.5 h-3.5" /> {project.owner_user_id}</span>
                  <span className="flex items-center gap-1.5"><Calendar className="w-3.5 h-3.5" /> Due: {project.due_date || "TBD"}</span>
                  <span className="flex items-center gap-1.5"><Layout className="w-3.5 h-3.5" /> {project.key}</span>
                </div>
              </div>

              <div className="flex flex-col justify-between items-end gap-4 min-w-[200px]">
                <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
                  {/* Progress bar logic - if status is closed, 100%, else estimation? 
                      For now, using status as a proxy */}
                  <div 
                    className={`h-full transition-all duration-1000 ${
                      project.status === "active" ? "bg-primary w-2/3" : 
                      project.status === "closed" ? "bg-success w-full" : 
                      "bg-muted-foreground/30 w-1/4"
                    }`}
                  />
                </div>
                <div className="flex items-center gap-2">
                  <button 
                    onClick={() => void handleDelete(project.id, project.status)}
                    className="p-2 rounded-lg hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors"
                    title={project.status === "archived" ? "Permanently Delete" : "Archive Project"}
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                  <button className="p-2 rounded-lg hover:bg-muted/30 transition-colors">
                    <MoreHorizontal className="w-5 h-5 text-muted-foreground" />
                  </button>
                  <Link 
                    href={`/projects/${project.id}`}
                    className="px-4 py-2 rounded-lg bg-muted/30 border border-border text-xs font-bold hover:bg-muted/50 transition-all"
                  >
                    View Details
                  </Link>
                </div>
              </div>
            </div>
          </motion.div>
        ))}
        {filteredProjects.length === 0 && !loading && (
          <PageEmpty message="No projects matching your filters" />
        )}
      </div>
    </div>
  );
}
