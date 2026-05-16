"use client";

import { motion } from "framer-motion";
import { 
  Target, 
  Calendar, 
  Users, 
  Layers, 
  Trophy,
  ArrowUpRight,
  Clock,
  Layout
} from "lucide-react";
import { DashboardHeader } from "@/components/ui/DashboardHeader";
import { Badge } from "@/components/ui/Badge";

const mockProjects = [
  { name: "OIDC Standardization", status: "In Progress", progress: 85, milestone: "v0.5.0-beta", team: 5, health: "on-track" },
  { name: "Real-time Metrics Pipeline", status: "In Progress", progress: 40, milestone: "v0.6.0-alpha", team: 3, health: "warning" },
  { name: "Organization Governance UI", status: "Done", progress: 100, milestone: "v0.4.0", team: 4, health: "on-track" },
  { name: "DevHub Mobile App", status: "Planning", progress: 5, milestone: "v1.0.0-backlog", team: 2, health: "on-track" },
];

export default function ProjectsStatusPage() {
  return (
    <div className="space-y-10 pb-20">
      <DashboardHeader 
        titlePrefix="Project"
        titleGradient="Status (과제 현황)"
        subtitle="Strategic overview of project milestones, delivery velocity, and roadmap progress."
      />

      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        {[
          { label: "Active Projects", value: "12", icon: Target },
          { label: "Next Milestone", value: "Jun 15", icon: Calendar },
          { label: "Team Velocity", value: "84 pts", icon: ArrowUpRight },
          { label: "Total Artifacts", value: "1.2k", icon: Layers },
        ].map((stat, i) => (
          <motion.div 
            key={stat.label}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card p-6"
          >
            <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-2 flex items-center gap-2">
              <stat.icon className="w-3 h-3 text-primary" /> {stat.label}
            </p>
            <h3 className="text-2xl font-black text-foreground dark:text-primary-foreground">{stat.value}</h3>
          </motion.div>
        ))}
      </div>

      <div className="grid gap-6">
        <h2 className="text-xl font-bold text-foreground dark:text-primary-foreground flex items-center gap-2 px-2">
          <Layout className="w-5 h-5 text-primary" /> Strategic Roadmap
        </h2>
        
        {mockProjects.map((project, i) => (
          <motion.div
            key={project.name}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card p-8 group"
          >
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-8">
              <div className="flex items-center gap-4">
                <div className={`w-12 h-12 rounded-2xl flex items-center justify-center ring-1 ring-border shadow-lg transition-transform group-hover:rotate-12 ${
                  project.status === "Done" ? "bg-green-500/10" : "bg-primary/10"
                }`}>
                  {project.status === "Done" ? <Trophy className="w-6 h-6 text-green-500" /> : <Target className="w-6 h-6 text-primary" />}
                </div>
                <div>
                  <h3 className="text-xl font-bold text-foreground dark:text-primary-foreground">{project.name}</h3>
                  <div className="flex items-center gap-3 mt-1">
                    <Badge variant={project.health === "on-track" ? "success" : "warning"} dot>{project.status}</Badge>
                    <span className="text-xs text-muted-foreground flex items-center gap-1 font-medium">
                      <Clock className="w-3 h-3" /> {project.milestone}
                    </span>
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-8">
                <div className="text-right">
                  <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Squad Size</p>
                  <div className="flex items-center justify-end gap-2">
                    <Users className="w-4 h-4 text-muted-foreground" />
                    <span className="text-lg font-black text-foreground dark:text-primary-foreground">{project.team}</span>
                  </div>
                </div>
                <button className="px-6 py-3 rounded-2xl bg-primary text-primary-foreground text-xs font-black uppercase tracking-widest hover:bg-primary/90 transition-all shadow-xl">
                  Details
                </button>
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex justify-between text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em]">
                <span>Overall Progress</span>
                <span>{project.progress}%</span>
              </div>
              <div className="h-3 w-full bg-muted/30 rounded-full overflow-hidden p-0.5 border border-border/50">
                <motion.div 
                  initial={{ width: 0 }}
                  animate={{ width: `${project.progress}%` }}
                  transition={{ duration: 1, delay: i * 0.2 }}
                  className={`h-full rounded-full transition-all ${
                    project.status === "Done" ? "bg-green-500" : "bg-gradient-to-r from-primary to-accent"
                  }`}
                />
              </div>
            </div>
          </motion.div>
        ))}
      </div>
    </div>
  );
}
