"use client";

import { motion } from "framer-motion";
import { 
  Activity, 
  Box, 
  Cpu, 
  Globe, 
  HardDrive, 
  ShieldCheck, 
  Zap,
  ExternalLink
} from "lucide-react";
import { DashboardHeader } from "@/components/ui/DashboardHeader";
import { Badge } from "@/components/ui/Badge";

const mockApplications = [
  { id: "APP-001", name: "DevHub Core API", status: "Healthy", uptime: "99.98%", version: "v1.2.4", load: "12%", region: "us-east-1" },
  { id: "APP-002", name: "DevHub AI Engine", status: "Healthy", uptime: "99.95%", version: "v0.8.2", load: "45%", region: "us-west-2" },
  { id: "APP-003", name: "Metrics Aggregator", status: "Degraded", uptime: "98.20%", version: "v1.0.1", load: "88%", region: "eu-central-1" },
  { id: "APP-004", name: "Notification Service", status: "Healthy", uptime: "100.00%", version: "v2.1.0", load: "5%", region: "ap-northeast-1" },
];

export default function ApplicationsStatusPage() {
  return (
    <div className="space-y-10 pb-20">
      <DashboardHeader 
        titlePrefix="Application"
        titleGradient="Status (어플리케이션 현황)"
        subtitle="Real-time monitoring of all production and staging application services."
      />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[
          { label: "Total Applications", value: "24", icon: Box, color: "text-blue-500" },
          { label: "Avg. Uptime", value: "99.95%", icon: Activity, color: "text-emerald-500" },
          { label: "Global Traffic", value: "1.2M/req", icon: Globe, color: "text-purple-500" },
          { label: "Security Alerts", value: "0", icon: ShieldCheck, color: "text-green-500" },
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
        {mockApplications.map((app, i) => (
          <motion.div
            key={app.id}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card p-6 flex flex-col md:flex-row items-center justify-between gap-6 group hover:bg-muted/10"
          >
            <div className="flex items-center gap-6 flex-1">
              <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-primary/20 to-accent/20 flex items-center justify-center ring-1 ring-border group-hover:scale-110 transition-transform">
                <Zap className="w-6 h-6 text-primary" />
              </div>
              <div>
                <div className="flex items-center gap-3 mb-1">
                  <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground">{app.name}</h3>
                  <Badge variant={app.status === "Healthy" ? "success" : "warning"} dot>{app.status}</Badge>
                </div>
                <div className="flex items-center gap-4 text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
                  <span className="flex items-center gap-1"><Cpu className="w-3 h-3" /> {app.load} Load</span>
                  <span>•</span>
                  <span className="flex items-center gap-1"><Globe className="w-3 h-3" /> {app.region}</span>
                  <span>•</span>
                  <span>{app.id}</span>
                </div>
              </div>
            </div>

            <div className="flex items-center gap-12 text-right">
              <div>
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Uptime</p>
                <p className="text-lg font-black text-foreground dark:text-primary-foreground">{app.uptime}</p>
              </div>
              <div>
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Version</p>
                <p className="text-sm font-mono font-bold text-accent">{app.version}</p>
              </div>
              <button className="p-3 rounded-xl hover:bg-muted/30 transition-all text-muted-foreground hover:text-primary">
                <ExternalLink className="w-5 h-5" />
              </button>
            </div>
          </motion.div>
        ))}
      </div>
    </div>
  );
}
