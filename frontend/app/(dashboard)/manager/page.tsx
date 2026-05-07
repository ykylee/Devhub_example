"use client";

import { motion } from "framer-motion";
import { 
  AlertTriangle, 
  BarChart3, 
  CheckCircle2, 
  Target, 
  Users,
  Calendar,
  FileText
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useEffect, useState } from "react";
import { Modal } from "@/components/ui/Modal";
import { Badge } from "@/components/ui/Badge";
import { useStore } from "@/lib/store";
import { riskService } from "@/lib/services/risk.service";
import { infraService } from "@/lib/services/infra.service";
import { realtimeService } from "@/lib/services/realtime.service";
import { Metric, Risk } from "@/lib/services/types";

type RiskCreatedEvent = Risk;
type CommandStatusEvent = {
  command_id: string;
  status: string;
};

export default function ManagerDashboard() {
  const [stats, setStats] = useState<Metric[]>([]);
  const [risks, setRisks] = useState<Risk[]>([]);
  const [selectedRisk, setSelectedRisk] = useState<Risk | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const { addToast } = useStore();

  useEffect(() => {
    const loadData = async () => {
      setIsLoading(true);
      try {
        const [metricsData, risksData] = await Promise.all([
          infraService.getMetrics("Manager"),
          riskService.getCriticalRisks()
        ]);
        setStats(metricsData);
        setRisks(risksData);
      } catch (error) {
        console.error("Failed to load dashboard data:", error);
      } finally {
        setIsLoading(false);
      }
    };

    loadData();

    // Subscribe to real-time risk updates
    const unsubscribeRisk = realtimeService.subscribe<RiskCreatedEvent>('risk.critical.created', (event) => {
      const newRisk = event.data;
      setRisks((prev) => [newRisk, ...prev]);
      addToast(`CRITICAL RISK DETECTED: ${newRisk.title}`, "error");
    });

    // Subscribe to command status updates
    const unsubscribeCommand = realtimeService.subscribe<CommandStatusEvent>('command.status.updated', (event) => {
      const { command_id, status } = event.data;
      addToast(`Command ${command_id.substring(0, 8)} updated: ${status}`, "info");
      // Optional: Refetch risks if command affects risk status
      loadData();
    });

    const interval = setInterval(loadData, 30000);
    return () => {
      clearInterval(interval);
      unsubscribeRisk();
      unsubscribeCommand();
    };
  }, [addToast]);

  const handleMitigation = async (plan: { action: string }) => {
    if (!selectedRisk || !selectedRisk.id) return;
    
    try {
      addToast(`Initializing ${plan.action} sequence...`, "info");
      const result = await riskService.applyMitigation(selectedRisk.id, plan.action);
      
      addToast(
        `Action Accepted. Command ID: ${result.command_id.substring(0, 8)}... (Status: ${result.status})`, 
        "success"
      );
      
      setSelectedRisk(null);
    } catch (error) {
      addToast("Failed to initiate mitigation protocol.", "error");
      console.error(error);
    }
  };

  return (
    <div className="space-y-10 pb-20">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <h1 className="text-4xl font-extrabold tracking-tight text-white mb-2">
            Project <span className="text-gradient">Intelligence</span>
          </h1>
          <p className="text-muted-foreground text-lg flex items-center gap-2">
            <Calendar className="w-4 h-4 text-primary" /> Milestone v1.0 • <span className="text-white font-bold">Week 12</span> of 16
          </p>
        </motion.div>

        <motion.div 
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          className="flex items-center gap-3"
        >
          <div className="glass px-4 py-2 rounded-xl border border-white/10 flex items-center gap-2">
            <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
            <span className="text-xs font-bold text-white uppercase tracking-wider">On Track</span>
          </div>
          <button className="bg-primary text-white px-6 py-2 rounded-xl text-sm font-bold hover:bg-primary/90 transition-all shadow-xl flex items-center gap-2">
            <FileText className="w-4 h-4" /> Weekly Report
          </button>
        </motion.div>
      </div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-20 gap-4">
          <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
          <p className="text-muted-foreground font-medium animate-pulse">Aggregating Intelligence...</p>
        </div>
      ) : (
        <>
          {/* KPI Overview */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {stats.map((stat, i) => (
              <motion.div 
                key={i}
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: i * 0.1 }}
                className="glass-card p-6 group"
              >
                <div className="flex items-center justify-between mb-4">
                  <div className={cn("p-2 rounded-xl bg-white/5 border border-white/10", stat.color)}>
                    <Target className="w-5 h-5" />
                  </div>
                  <span className="text-[10px] font-black text-green-400 bg-green-500/10 px-2 py-1 rounded-lg">
                    {stat.trend}
                  </span>
                </div>
                <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest">{stat.label}</p>
                <h3 className="text-3xl font-black text-white mt-1">{stat.value}</h3>
              </motion.div>
            ))}
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Risk Management Section */}
            <div className="lg:col-span-2 space-y-6">
              <section className="glass border-rose-500/20 rounded-3xl overflow-hidden relative">
                <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-rose-500/50 to-transparent"></div>
                <div className="p-8 border-b border-white/5 flex items-center justify-between bg-rose-500/5">
                  <h2 className="text-xl font-bold text-rose-500 flex items-center gap-3">
                    <AlertTriangle className="w-6 h-6 animate-pulse" /> Critical Risk Detection
                  </h2>
                  <span className="text-[10px] font-black uppercase tracking-widest text-rose-500/50">AI-Monitored</span>
                </div>
                
                <div className="divide-y divide-white/5">
                  {risks.map((risk, i) => (
                    <motion.div 
                      key={i}
                      whileHover={{ backgroundColor: "rgba(255,255,255,0.02)" }}
                      className="p-8 flex flex-col md:flex-row md:items-center justify-between gap-6"
                    >
                      <div className="space-y-2">
                        <div className="flex items-center gap-3">
                          <h3 className="text-lg font-bold text-white">{risk.title}</h3>
                          <span className="px-2 py-0.5 rounded-md bg-rose-500/20 text-rose-500 text-[10px] font-black uppercase tracking-tighter">
                            {risk.impact} Impact
                          </span>
                        </div>
                        <p className="text-sm text-muted-foreground max-w-lg">
                          {risk.reason}
                        </p>
                      </div>
                      
                      <div className="flex items-center gap-4">
                        <div className="text-right hidden md:block">
                          <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Assigned to</p>
                          <div className="flex items-center gap-2 justify-end">
                            <div className="w-5 h-5 rounded-full bg-primary/20 border border-primary/20" />
                            <span className="text-xs font-bold text-white">{risk.owner}</span>
                          </div>
                        </div>
                        <button 
                          onClick={() => setSelectedRisk(risk)}
                          className="px-6 py-2.5 rounded-xl bg-white/5 border border-white/10 text-xs font-black uppercase tracking-widest text-white hover:bg-white/10 transition-all"
                        >
                          Details
                        </button>
                      </div>
                    </motion.div>
                  ))}
                </div>
              </section>

          {/* Activity Analytics Placeholder */}
          <section className="glass-card p-8 h-80 flex flex-col items-center justify-center relative group overflow-hidden">
            <div className="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-accent/5 opacity-50 group-hover:opacity-100 transition-opacity" />
            <BarChart3 className="w-16 h-16 text-primary/30 mb-6 group-hover:scale-110 transition-transform duration-500" />
            <h3 className="text-lg font-bold text-white mb-2">Resource Utilization Velocity</h3>
            <p className="text-sm text-muted-foreground max-w-sm text-center">
              Real-time velocity tracking and burndown analytics will be rendered here using the integrated gRPC telemetry stream.
            </p>
          </section>
        </div>

        {/* Right Sidebar Widgets */}
        <div className="space-y-8">
          {/* Team Distribution */}
          <section className="glass-card p-8 space-y-8">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground flex items-center gap-2">
              <Users className="w-4 h-4 text-primary" /> Talent Load Balancing
            </h3>
            
            <div className="space-y-8">
              {[
                { name: "YK Lee", load: 85, status: "Critical", color: "bg-rose-500" },
                { name: "Alex K.", load: 45, status: "Optimal", color: "bg-emerald-500" },
                { name: "Sam J.", load: 92, status: "Overloaded", color: "bg-rose-500" },
                { name: "Jordan M.", load: 60, status: "Optimal", color: "bg-emerald-500" }
              ].map((member, i) => (
                <div key={i} className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-bold text-white">{member.name}</span>
                    <span className={cn("text-[10px] font-black uppercase tracking-tighter", member.load > 80 ? "text-rose-500" : "text-emerald-500")}>
                      {member.load}% Load
                    </span>
                  </div>
                  <div className="h-1.5 w-full bg-white/5 rounded-full overflow-hidden border border-white/5">
                    <motion.div 
                      initial={{ width: 0 }}
                      animate={{ width: `${member.load}%` }}
                      className={cn("h-full rounded-full transition-all duration-1000", member.color)}
                    />
                  </div>
                </div>
              ))}
            </div>
            
            <button className="w-full py-3 rounded-2xl bg-primary/10 border border-primary/20 text-xs font-black uppercase tracking-widest text-primary hover:bg-primary/20 transition-all">
              Optimize Resources
            </button>
          </section>

          {/* Recent Decisions Log */}
          <section className="glass-card p-8 space-y-6">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-accent" /> Decision Audit
            </h3>
            <div className="space-y-6">
              {[
                { title: "Branch Protection Policy", date: "2 days ago", type: "Security" },
                { title: "gRPC IDL Specification", date: "4 days ago", type: "Architecture" }
              ].map((log, i) => (
                <div key={i} className="flex gap-4">
                  <div className="w-1 h-8 rounded-full bg-white/10 mt-1" />
                  <div>
                    <p className="text-xs font-bold text-white">{log.title}</p>
                    <p className="text-[10px] text-muted-foreground mt-1 uppercase tracking-widest">{log.type} • {log.date}</p>
                  </div>
                </div>
              ))}
            </div>
          </section>
          </div>
        </div>
      </>
      )}

      {/* Risk Mitigation Modal */}
      <Modal
        isOpen={!!selectedRisk}
        onClose={() => setSelectedRisk(null)}
        title="Risk Mitigation Protocol"
        size="lg"
      >
        {selectedRisk && (
          <div className="space-y-8">
            <div className="p-6 rounded-2xl bg-rose-500/5 border border-rose-500/20">
              <div className="flex items-center gap-3 mb-4">
                <Badge variant="danger">{selectedRisk.impact} Impact</Badge>
                <h4 className="text-xl font-bold text-white">{selectedRisk.title}</h4>
              </div>
              <p className="text-sm text-muted-foreground leading-relaxed">
                {selectedRisk.reason}
              </p>
            </div>

            <div className="space-y-4">
              <h5 className="text-xs font-black uppercase tracking-widest text-muted-foreground">Proposed Countermeasures</h5>
              <div className="grid grid-cols-1 gap-3">
                {[
                  { title: "Immediate Patching", desc: "Trigger automated rollback to last stable SHA.", action: "Deploy Rollback" },
                  { title: "Scale Up Infrastructure", desc: "Increase runner capacity in Asia-East-1 region.", action: "Execute Scaling" },
                  { title: "Postpone Milestone", desc: "Reschedule v1.0 release by 48 hours for QA.", action: "Update Roadmap" }
                ].map((plan, i) => (
                  <div key={i} className="glass p-4 rounded-xl border border-white/5 flex items-center justify-between group hover:border-primary/30 transition-all">
                    <div>
                      <p className="text-sm font-bold text-white">{plan.title}</p>
                      <p className="text-xs text-muted-foreground">{plan.desc}</p>
                    </div>
                    <button 
                      onClick={() => handleMitigation(plan)}
                      className="px-4 py-2 rounded-lg bg-primary/10 text-primary text-[10px] font-black uppercase tracking-widest hover:bg-primary hover:text-white transition-all"
                    >
                      {plan.action}
                    </button>
                  </div>
                ))}
              </div>
            </div>

            <div className="flex items-center gap-4 pt-4 border-t border-white/5">
              <div className="flex -space-x-2">
                {[1, 2, 3].map(i => (
                  <div key={i} className="w-8 h-8 rounded-full border-2 border-[#030014] bg-muted-foreground/20" />
                ))}
              </div>
              <p className="text-xs text-muted-foreground">
                <span className="text-white font-bold">3 stakeholders</span> are currently reviewing this mitigation plan.
              </p>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
