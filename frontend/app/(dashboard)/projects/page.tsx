"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { 
  Briefcase, 
  Calendar, 
  CheckCircle2, 
  Clock, 
  Layout, 
  MoreHorizontal,
  Plus,
  Target,
  Users,
  Loader2
} from "lucide-react";
import { DashboardHeader } from "@/components/ui/DashboardHeader";
import { Badge } from "@/components/ui/Badge";
import { projectService, Project } from "@/lib/services/project.service";
import { repositoryService } from "@/lib/services/repository.service";

export default function ProjectsStatusPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadData = async () => {
      try {
        const repos = await repositoryService.listRepositories();
        const allProjects = await projectService.listAllProjects(repos.map(r => r.id));
        setProjects(allProjects);
      } catch (err) {
        setError("Failed to load projects data.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, []);

  const totalProjects = projects.length;
  const activeProjects = projects.filter(p => p.status === "active").length;
  const planningProjects = projects.filter(p => p.status === "planning").length;
  const closedProjects = projects.filter(p => p.status === "closed" || p.status === "archived").length;

  if (loading) {
    return (
      <div className="flex items-center justify-center h-[60vh]">
        <Loader2 className="w-10 h-10 text-primary animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-10 pb-20">
      <DashboardHeader 
        titlePrefix="Project"
        titleGradient="Milestones (과제 현황)"
        subtitle="Tracking development projects, milestones, and delivery timelines."
      >
        <button className="flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-primary-foreground font-bold text-sm hover:opacity-90 transition-opacity">
          <Plus className="w-4 h-4" /> New Project
        </button>
      </DashboardHeader>

      {error && (
        <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[
          { label: "Total Projects", value: totalProjects.toString(), icon: Briefcase, color: "text-blue-500" },
          { label: "Active", value: activeProjects.toString(), icon: Target, color: "text-emerald-500" },
          { label: "Planning", value: planningProjects.toString(), icon: Clock, color: "text-amber-500" },
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

      <div className="grid gap-6">
        {projects.map((project, i) => (
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
                  <h3 className="text-xl font-bold text-foreground dark:text-primary-foreground">{project.name}</h3>
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
                      project.status === "closed" ? "bg-emerald-500 w-full" : 
                      "bg-muted-foreground/30 w-1/4"
                    }`}
                  />
                </div>
                <div className="flex items-center gap-2">
                  <button className="p-2 rounded-lg hover:bg-muted/30 transition-colors">
                    <MoreHorizontal className="w-5 h-5 text-muted-foreground" />
                  </button>
                  <button className="px-4 py-2 rounded-lg bg-muted/30 border border-border text-xs font-bold hover:bg-muted/50 transition-all">
                    View Details
                  </button>
                </div>
              </div>
            </div>
          </motion.div>
        ))}
        {projects.length === 0 && !loading && (
          <div className="text-center py-20 glass-card">
            <p className="text-muted-foreground">No projects found.</p>
          </div>
        )}
      </div>
    </div>
  );
}
