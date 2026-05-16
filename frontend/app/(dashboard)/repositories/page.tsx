"use client";

import { motion } from "framer-motion";
import { 
  GitBranch, 
  GitCommit, 
  GitPullRequest, 
  History, 
  Star, 
  Users,
  Code2,
  Lock,
  Unlock
} from "lucide-react";
import { DashboardHeader } from "@/components/ui/DashboardHeader";
import { Badge } from "@/components/ui/Badge";

const mockRepositories = [
  { name: "devhub-core", stars: 124, commits: "2.4k", prs: 12, health: "A+", visibility: "private", language: "Go" },
  { name: "devhub-frontend", stars: 89, commits: "1.8k", prs: 5, health: "A", visibility: "private", language: "TypeScript" },
  { name: "devhub-ai", stars: 45, commits: "800", prs: 2, health: "B", visibility: "internal", language: "Python" },
  { name: "devhub-infra", stars: 32, commits: "400", prs: 0, health: "A+", visibility: "private", language: "Terraform" },
];

export default function RepositoriesStatusPage() {
  return (
    <div className="space-y-10 pb-20">
      <DashboardHeader 
        titlePrefix="Repository"
        titleGradient="Status (저장소 현황)"
        subtitle="Tracking development activity and code health across all internal repositories."
      />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {[
          { label: "Active Commits (24h)", value: "142", icon: GitCommit, color: "text-blue-500" },
          { label: "Open Pull Requests", value: "19", icon: GitPullRequest, color: "text-amber-500" },
          { label: "Contributors", value: "32", icon: Users, color: "text-purple-500" },
        ].map((stat, i) => (
          <motion.div 
            key={stat.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card p-6 flex items-center gap-6"
          >
            <div className={`p-4 rounded-2xl bg-muted/30 border border-border ${stat.color}`}>
              <stat.icon className="w-6 h-6" />
            </div>
            <div>
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">{stat.label}</p>
              <h3 className="text-3xl font-black text-foreground dark:text-primary-foreground">{stat.value}</h3>
            </div>
          </motion.div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {mockRepositories.map((repo, i) => (
          <motion.div
            key={repo.name}
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card p-8 group relative overflow-hidden"
          >
            <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-30 transition-opacity">
              <Code2 className="w-24 h-24 text-primary" />
            </div>

            <div className="flex items-start justify-between mb-8 relative z-10">
              <div className="flex items-center gap-4">
                <div className="p-3 rounded-xl bg-primary/10 border border-primary/20">
                  <GitBranch className="w-6 h-6 text-primary" />
                </div>
                <div>
                  <h3 className="text-xl font-bold text-foreground dark:text-primary-foreground flex items-center gap-2">
                    {repo.name}
                    {repo.visibility === "private" ? <Lock className="w-3 h-3 text-muted-foreground" /> : <Unlock className="w-3 h-3 text-muted-foreground" />}
                  </h3>
                  <p className="text-xs text-muted-foreground font-medium mt-1">Main branch stable • {repo.language}</p>
                </div>
              </div>
              <Badge variant="glass" className="bg-primary/5">{repo.health} Health</Badge>
            </div>

            <div className="grid grid-cols-3 gap-4 relative z-10">
              <div className="space-y-1">
                <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Commits</p>
                <div className="flex items-center gap-2">
                  <History className="w-3 h-3 text-blue-400" />
                  <span className="text-sm font-black text-foreground dark:text-primary-foreground">{repo.commits}</span>
                </div>
              </div>
              <div className="space-y-1">
                <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Stars</p>
                <div className="flex items-center gap-2">
                  <Star className="w-3 h-3 text-yellow-400" />
                  <span className="text-sm font-black text-foreground dark:text-primary-foreground">{repo.stars}</span>
                </div>
              </div>
              <div className="space-y-1">
                <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Open PRs</p>
                <div className="flex items-center gap-2">
                  <GitPullRequest className="w-3 h-3 text-emerald-400" />
                  <span className="text-sm font-black text-foreground dark:text-primary-foreground">{repo.prs}</span>
                </div>
              </div>
            </div>

            <div className="mt-8 pt-6 border-t border-border/50 relative z-10 flex items-center justify-between">
              <div className="flex -space-x-2">
                {[1, 2, 3, 4].map(j => (
                  <div key={j} className="w-6 h-6 rounded-full border-2 border-background bg-muted flex items-center justify-center text-[8px] font-bold">U{j}</div>
                ))}
                <div className="w-6 h-6 rounded-full border-2 border-background bg-primary/20 flex items-center justify-center text-[8px] font-bold text-primary">+8</div>
              </div>
              <button className="text-xs font-black uppercase tracking-widest text-primary hover:underline underline-offset-4">Browse Code</button>
            </div>
          </motion.div>
        ))}
      </div>
    </div>
  );
}
