"use client";

import { GardenerFeed } from "@/components/dashboard/GardenerFeed";
import { motion } from "framer-motion";
import { Brain, ShieldCheck, Zap } from "lucide-react";
import { cn } from "@/lib/utils";
import { useStore } from "@/lib/store";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function GardenerPage() {
  const { role: userRole } = useStore();
  const router = useRouter();

  useEffect(() => {
    if (userRole !== "System Admin") {
      router.push("/developer");
    }
  }, [userRole, router]);

  if (userRole !== "System Admin") return null;

  return (
    <div className="space-y-10 pb-10 h-full flex flex-col">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
        >
          <h1 className="text-4xl font-extrabold tracking-tight text-foreground dark:text-primary-foreground mb-2">
            AI <span className="text-gradient">Gardener</span>
          </h1>
          <p className="text-muted-foreground text-lg flex items-center gap-2">
            <Brain className="w-4 h-4 text-primary" /> Autonomous System Optimization • <span className="text-foreground dark:text-primary-foreground font-bold uppercase tracking-widest text-xs bg-primary/20 px-2 py-0.5 rounded border border-primary/20">AI Active</span>
          </p>
        </motion.div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 flex-1">
        {/* Left Column: Feed */}
        <div className="lg:col-span-2 space-y-8">
          <section className="glass rounded-3xl p-8 border border-border shadow-2xl relative overflow-hidden">
             {/* Decorative glow */}
            <div className="absolute top-0 right-0 w-64 h-64 bg-primary/5 rounded-full blur-[100px] -mr-32 -mt-32" />
            
            <GardenerFeed />
          </section>
        </div>

        {/* Right Column: AI Stats & Info */}
        <div className="space-y-6">
          <motion.div 
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="glass-card p-6"
          >
            <h3 className="text-xs font-black uppercase tracking-[0.2em] text-muted-foreground mb-6">Autonomous Stats</h3>
            <div className="space-y-6">
              {[
                { label: "Optimization Score", value: "94/100", icon: Zap, color: "text-amber-400" },
                { label: "Security Posture", value: "Locked", icon: ShieldCheck, color: "text-emerald-400" },
              ].map((stat, i) => (
                <div key={i} className="flex items-center gap-4">
                  <div className={cn("p-2 rounded-lg bg-muted/30 border border-border", stat.color)}>
                    <stat.icon className="w-4 h-4" />
                  </div>
                  <div>
                    <p className="text-[10px] font-bold text-muted-foreground uppercase">{stat.label}</p>
                    <p className="text-lg font-black text-foreground dark:text-primary-foreground">{stat.value}</p>
                  </div>
                </div>
              ))}
            </div>
          </motion.div>

          <motion.div 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
            className="glass rounded-2xl p-6 border-border/60"
          >
            <h3 className="text-xs font-black uppercase tracking-[0.2em] text-foreground/50 dark:text-primary-foreground/50 mb-4">How it works</h3>
            <p className="text-xs text-muted-foreground leading-relaxed">
              AI Gardener analyzes real-time telemetry from Gitea, Go Core, and Infrastructure nodes to identify bottlenecks and security risks before they impact users.
            </p>
          </motion.div>
        </div>
      </div>
    </div>
  );
}
