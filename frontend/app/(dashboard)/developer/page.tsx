"use client";

import { motion } from "framer-motion";
import { CheckCircle2, CircleDashed, Clock, GitCommit, GitPullRequest, MessageSquare, PlayCircle, Star, Terminal } from "lucide-react";

export default function DeveloperDashboard() {
  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Welcome back, YK Lee</h1>
          <p className="text-muted-foreground mt-1">Here is what is happening with your projects today.</p>
        </div>
        <div className="flex items-center gap-3">
          <button className="bg-primary text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors shadow-sm">
            Create Issue
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Main Feed */}
        <div className="md:col-span-2 space-y-6">
          {/* Active Work */}
          <div className="bg-card rounded-xl border border-border overflow-hidden shadow-sm">
            <div className="px-6 py-4 border-b border-border bg-card/50">
              <h2 className="font-semibold flex items-center gap-2">
                <Terminal className="w-4 h-4 text-indigo-400" /> Current Work
              </h2>
            </div>
            <div className="divide-y divide-border">
              {[
                { title: "TASK-007 Gitea Webhook 수신부", repo: "devhub-core", status: "In Progress", type: "pr", time: "2h ago" },
                { title: "Refactor Authentication Logic", repo: "devhub-core", status: "Review", type: "pr", time: "5h ago" }
              ].map((item, i) => (
                <div key={i} className="p-6 hover:bg-accent/50 transition-colors group cursor-pointer">
                  <div className="flex items-start justify-between">
                    <div className="flex gap-3">
                      <div className="mt-1">
                        {item.status === "In Progress" ? (
                          <CircleDashed className="w-5 h-5 text-amber-500 animate-[spin_3s_linear_infinite]" />
                        ) : (
                          <GitPullRequest className="w-5 h-5 text-emerald-500" />
                        )}
                      </div>
                      <div>
                        <h3 className="font-medium group-hover:text-primary transition-colors">{item.title}</h3>
                        <p className="text-sm text-muted-foreground mt-1 flex items-center gap-2">
                          <span className="text-xs font-mono bg-accent px-1.5 py-0.5 rounded">{item.repo}</span>
                          • {item.time}
                        </p>
                      </div>
                    </div>
                    <span className="text-xs font-medium px-2 py-1 rounded-full bg-accent text-muted-foreground">
                      {item.status}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* CI/CD Status */}
          <div className="bg-card rounded-xl border border-border overflow-hidden shadow-sm">
            <div className="px-6 py-4 border-b border-border bg-card/50">
              <h2 className="font-semibold flex items-center gap-2">
                <PlayCircle className="w-4 h-4 text-emerald-400" /> Recent Builds
              </h2>
            </div>
            <div className="p-6">
              <div className="space-y-4">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="flex items-center justify-between p-3 rounded-lg border border-border bg-background/50">
                    <div className="flex items-center gap-3">
                      <CheckCircle2 className="w-5 h-5 text-emerald-500" />
                      <div>
                        <p className="text-sm font-medium">Build #10{i} <span className="text-muted-foreground font-normal">for</span> main</p>
                        <p className="text-xs text-muted-foreground mt-0.5">Passed in 2m 14s</p>
                      </div>
                    </div>
                    <GitCommit className="w-4 h-4 text-muted-foreground" />
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Sidebar Widgets */}
        <div className="space-y-6">
          {/* AI Gardener */}
          <motion.div 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="bg-gradient-to-br from-indigo-500/10 via-purple-500/10 to-transparent rounded-xl border border-indigo-500/20 p-6 relative overflow-hidden"
          >
            <div className="absolute top-0 right-0 w-32 h-32 bg-indigo-500/10 rounded-full blur-2xl"></div>
            <h3 className="font-semibold text-indigo-400 flex items-center gap-2">
              <Star className="w-4 h-4" /> AI Gardener
            </h3>
            <p className="text-sm text-muted-foreground mt-3 leading-relaxed">
              You've been working on the Webhook integration for 2 hours. There's a similar implementation in <span className="text-indigo-400 cursor-pointer hover:underline">devhub-legacy</span> that might save you time.
            </p>
            <button className="mt-4 text-xs font-medium bg-indigo-500/20 text-indigo-300 px-3 py-1.5 rounded-md hover:bg-indigo-500/30 transition-colors w-full">
              View Reference
            </button>
          </motion.div>

          {/* Kudos Feed */}
          <div className="bg-card rounded-xl border border-border p-6 shadow-sm">
            <h3 className="font-semibold flex items-center gap-2 mb-4">
              <MessageSquare className="w-4 h-4 text-pink-400" /> Team Kudos
            </h3>
            <div className="space-y-4">
              <div className="p-3 rounded-lg bg-background border border-border text-sm">
                <p><span className="font-medium text-pink-400">@alex</span> gave you kudos for resolving the database deadlock issue quickly! 🎉</p>
                <p className="text-xs text-muted-foreground mt-2">Yesterday</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
