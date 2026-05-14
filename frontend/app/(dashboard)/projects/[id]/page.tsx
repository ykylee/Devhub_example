"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";
import { 
  FolderKanban, 
  GitBranch, 
  Box, 
  Settings, 
  ArrowLeft, 
  Clock, 
  User, 
  Users,
  Shield, 
  Globe,
  ExternalLink,
  CheckCircle2,
  AlertCircle
} from "lucide-react";
import { projectService } from "@/lib/services/project.service";
import { Project, ProjectMember, ProjectIntegration, ProjectStatus } from "@/lib/services/project.types";
import { Badge } from "@/components/ui/Badge";
import { useToast } from "@/components/ui/Toast";
import { cn } from "@/lib/utils";
import { format } from "date-fns";

export default function ProjectDetailPage() {
  const { id } = useParams() as { id: string };
  const router = useRouter();
  const { toast } = useToast();
  
  const [project, setProject] = useState<Project | null>(null);
  const [members, setMembers] = useState<ProjectMember[]>([]);
  const [integrations, setIntegrations] = useState<ProjectIntegration[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      setIsLoading(true);
      setError(null);
      try {
        // Since many of these are planned (API-56, API-58), we handle 501
        const projectData = await projectService.getProject(id);
        setProject(projectData);
        
        // Members and integrations fetching might fail with 501
        try {
           // We might need to add getProjectMembers and getProjectIntegrations to service
        } catch (e) {}
      } catch (err) {
        console.error("[ProjectDetail] load failed:", err);
        const status = (err as any).status;
        if (status === 501) {
          setError("Backend API for Project Detail is not implemented yet (501).");
        } else {
          setError("Failed to load project details.");
        }
      } finally {
        setIsLoading(false);
      }
    };
    load();
  }, [id]);

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center py-32 gap-4">
        <div className="w-12 h-12 border-4 border-indigo-500/20 border-t-indigo-500 rounded-full animate-spin" />
        <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Project Details...</p>
      </div>
    );
  }

  if (error && !project) {
    return (
      <div className="p-8 space-y-6">
        <button 
          onClick={() => router.back()}
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors text-xs font-bold uppercase tracking-widest"
        >
          <ArrowLeft className="w-4 h-4" /> Back
        </button>
        <div className="glass border-indigo-500/20 bg-indigo-500/5 rounded-3xl p-12 text-center">
          <AlertCircle className="w-12 h-12 text-indigo-400 mx-auto mb-4" />
          <h2 className="text-xl font-black text-foreground uppercase tracking-tight mb-2">Operational Note</h2>
          <p className="text-sm text-muted-foreground max-w-md mx-auto">{error}</p>
          <button 
             onClick={() => router.push("/projects")}
             className="mt-6 px-6 py-2 bg-muted/20 border border-border rounded-xl text-xs font-bold text-foreground hover:bg-muted/40 transition-all"
          >
            Return to Projects
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8 p-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col gap-6">
        <button 
          onClick={() => router.back()}
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors text-xs font-bold uppercase tracking-widest self-start"
        >
          <ArrowLeft className="w-4 h-4" /> Back
        </button>

        <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
          <div className="flex items-center gap-6">
            <div className="w-20 h-20 bg-indigo-500/10 rounded-3xl flex items-center justify-center border border-indigo-500/20 shadow-xl shadow-indigo-500/5">
              <FolderKanban className="w-10 h-10 text-indigo-400" />
            </div>
            <div>
              <div className="flex items-center gap-3">
                <h1 className="text-4xl font-black text-foreground tracking-tighter uppercase">
                  {project?.name || "Project Name"}
                </h1>
                <ProjectStatusBadge status={project?.status || "planning"} />
              </div>
              <div className="flex items-center gap-4 mt-2">
                <span className="text-sm font-mono text-muted-foreground bg-muted/50 px-2 py-1 rounded border border-border/40 uppercase tracking-tighter">
                  {project?.key || "ID_MISSING"}
                </span>
                <div className="flex items-center gap-2 text-muted-foreground">
                  <User className="w-4 h-4 opacity-40" />
                  <span className="text-xs font-bold">{project?.owner_user_id || "Unassigned"}</span>
                </div>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-3">
             <div className="flex flex-col items-end gap-1 px-4 py-2 bg-muted/20 rounded-2xl border border-border/60">
                <span className="text-[9px] font-black text-muted-foreground uppercase tracking-widest opacity-60">Timeline</span>
                <span className="text-xs font-bold text-foreground">
                  {project?.start_date ? format(new Date(project.start_date), "MMM d, yyyy") : "TBD"}
                  {" — "}
                  {project?.due_date ? format(new Date(project.due_date), "MMM d, yyyy") : "Continuous"}
                </span>
             </div>
             <button className="p-3.5 glass border-border rounded-2xl text-muted-foreground hover:text-foreground transition-all">
               <Settings className="w-5 h-5" />
             </button>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        <div className="md:col-span-2 space-y-8">
          {/* Main Info */}
          <section className="glass border-border rounded-3xl p-8 space-y-6">
             <div className="flex items-center justify-between border-b border-border/40 pb-4">
                <h3 className="text-xs font-black text-muted-foreground uppercase tracking-[0.3em]">Context & Hierarchy</h3>
                <div className="flex items-center gap-4">
                   <div className="flex items-center gap-1.5 px-3 py-1 bg-purple-500/10 rounded-full border border-purple-500/20">
                      <Box className="w-3 h-3 text-purple-400" />
                      <span className="text-[10px] font-black text-purple-400 uppercase tracking-tight">
                        App: {project?.application_id || "N/A"}
                      </span>
                   </div>
                   <div className="flex items-center gap-1.5 px-3 py-1 bg-pink-500/10 rounded-full border border-pink-500/20">
                      <GitBranch className="w-3 h-3 text-pink-400" />
                      <span className="text-[10px] font-black text-pink-400 uppercase tracking-tight">
                        Repo ID: {project?.repository_id}
                      </span>
                   </div>
                </div>
             </div>

             <div className="space-y-4">
                <h3 className="text-xs font-black text-muted-foreground uppercase tracking-[0.3em]">Project Description</h3>
                <p className="text-foreground/80 leading-relaxed text-sm">
                  {project?.description || "No description provided for this project unit."}
                </p>
             </div>
          </section>

          {/* Members */}
          <section className="glass border-border rounded-3xl p-8 space-y-6">
             <div className="flex items-center justify-between">
                <h3 className="text-xs font-black text-muted-foreground uppercase tracking-[0.3em]">Assigned Members</h3>
                <Badge variant="glass">{members.length} Total</Badge>
             </div>
             
             <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {members.length > 0 ? members.map(m => (
                  <div key={m.user_id} className="flex items-center gap-3 p-3 bg-muted/20 rounded-2xl border border-border/60">
                    <div className="w-8 h-8 rounded-full bg-accent/20 flex items-center justify-center border border-accent/20">
                       <User className="w-4 h-4 text-accent" />
                    </div>
                    <div>
                      <p className="text-xs font-bold text-foreground">{m.user_id}</p>
                      <p className="text-[9px] font-black text-muted-foreground uppercase tracking-widest">{m.project_role}</p>
                    </div>
                  </div>
                )) : (
                  <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest opacity-40 col-span-2 text-center py-4 italic">
                    No explicit memberships found.
                  </p>
                )}
             </div>
          </section>
        </div>

        <div className="space-y-8">
           {/* Integrations */}
           <section className="glass border-border rounded-3xl p-6 space-y-6">
              <h3 className="text-xs font-black text-muted-foreground uppercase tracking-[0.3em]">Integrations</h3>
              <div className="space-y-4">
                 <IntegrationStatus label="Jira execution_system" status="connected" />
                 <IntegrationStatus label="Confluence documentation" status="missing" />
              </div>
              <button className="w-full py-3 glass border-border rounded-2xl text-[9px] font-black uppercase tracking-widest hover:bg-muted/40 transition-all">
                Configure Integrations
              </button>
           </section>

           {/* Activities */}
           <section className="glass border-border rounded-3xl p-6 space-y-4">
              <h3 className="text-xs font-black text-muted-foreground uppercase tracking-[0.3em]">Recent Activity</h3>
              <div className="space-y-6">
                 <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest opacity-40 text-center py-8">
                   No recent PR activity recorded.
                 </p>
              </div>
           </section>
        </div>
      </div>
    </div>
  );
}

function ProjectStatusBadge({ status }: { status: ProjectStatus }) {
  switch (status) {
    case "active":
      return <Badge variant="success" dot>Active</Badge>;
    case "planning":
      return <Badge variant="primary" dot>Planning</Badge>;
    case "on_hold":
      return <Badge variant="warning" dot>On Hold</Badge>;
    case "closed":
      return <Badge variant="secondary" dot>Closed</Badge>;
    case "archived":
      return <Badge variant="danger" dot>Archived</Badge>;
    default:
      return <Badge variant="glass">{status}</Badge>;
  }
}

function IntegrationStatus({ label, status }: { label: string, status: "connected" | "missing" | "error" }) {
  return (
    <div className="flex items-center justify-between group">
      <span className="text-[11px] font-bold text-foreground/70">{label}</span>
      {status === "connected" && <CheckCircle2 className="w-4 h-4 text-green-500" />}
      {status === "missing" && <div className="w-4 h-4 rounded-full border-2 border-muted-foreground/20" />}
      {status === "error" && <AlertCircle className="w-4 h-4 text-rose-500" />}
    </div>
  );
}
