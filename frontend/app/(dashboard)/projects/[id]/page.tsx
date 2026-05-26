"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { 
  ArrowLeft, 
  Briefcase, 
  Calendar, 
  Clock, 
  Plus,
  Target,
  Users,
  Loader2,
  ChevronRight,
  MessageSquare,
  Paperclip,
  TrendingUp
} from "lucide-react";
import { useRouter, useParams } from "next/navigation";
import { Badge } from "@/components/ui/Badge";
import { projectService } from "@/lib/services/project.service";
import type { Project, ProjectRepositoryLink } from "@/lib/services/project.types";
import { identityService, OrgMember } from "@/lib/services/identity.service";
import { 
  Tooltip, 
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell
} from "recharts";

// Mock task distribution
const mockTaskData = [
  { name: "To Do", value: 12, color: "#94a3b8" },
  { name: "In Progress", value: 8, color: "#3b82f6" },
  { name: "Review", value: 5, color: "#8b5cf6" },
  { name: "Done", value: 25, color: "#10b981" },
];

export default function ProjectDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const router = useRouter();
  
  const [project, setProject] = useState<Project | null>(null);
  const [projectRepositories, setProjectRepositories] = useState<ProjectRepositoryLink[]>([]);
  const [users, setUsers] = useState<OrgMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadData = async () => {
      try {
        const [projectData, usersData] = await Promise.all([
          projectService.getProject(id),
          identityService.getUsers()
        ]);
        setProject(projectData);
        setUsers(usersData);
        const links = await projectService.getProjectRepositories(id).catch(() => []);
        setProjectRepositories(links);
      } catch (err) {
        setError("Failed to load project details.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, [id]);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center h-[60vh] gap-4">
        <Loader2 className="w-10 h-10 text-primary animate-spin" />
        <p className="text-muted-foreground animate-pulse font-black uppercase tracking-widest text-[10px]">Assembling Project Roadmap...</p>
      </div>
    );
  }

  if (error || !project) {
    return (
      <div className="text-center py-20 space-y-6">
        <div className="glass-card p-10 max-w-md mx-auto">
          <Briefcase className="w-16 h-16 text-muted-foreground mx-auto mb-4 opacity-20" />
          <h2 className="text-xl font-bold text-foreground dark:text-primary-foreground mb-2">Project Not Found</h2>
          <p className="text-muted-foreground text-sm mb-6">{error || "The requested project roadmap could not be located."}</p>
          <button 
            onClick={() => router.back()}
            className="px-6 py-2 rounded-xl bg-primary text-primary-foreground font-bold text-sm"
          >
            Go Back
          </button>
        </div>
      </div>
    );
  }

  const completionRate = project.status === "closed" ? 100 : (project.status === "active" ? 65 : 10);

  // Find project owner
  const owner = users.find(u => u.id === project.owner_user_id);
  
  // Get other members as contributors
  const contributors = users.filter(u => u.id !== project.owner_user_id);
  
  interface TeamMemberUI {
    name: string;
    role: string;
    status: "Online" | "Busy" | "Offline";
  }
  
  const teamMembers: TeamMemberUI[] = [];
  if (owner) {
    teamMembers.push({
      name: owner.name,
      role: "Owner",
      status: owner.status === "active" ? "Online" : "Offline",
    });
  } else if (project.owner_user_id) {
    teamMembers.push({
      name: `User (${project.owner_user_id})`,
      role: "Owner",
      status: "Online",
    });
  }
  
  // Add up to 3 contributors from backend users
  contributors.slice(0, 3).forEach((c, idx) => {
    const statusVal = c.status === "active" ? (idx % 2 === 0 ? "Online" : "Busy") : "Offline";
    teamMembers.push({
      name: c.name,
      role: "Contributor",
      status: statusVal,
    });
  });

  // Fallback contributors if we have no other users
  if (teamMembers.length <= 1) {
    teamMembers.push(
      { name: "Alex K.", role: "Contributor", status: "Busy" },
      { name: "Sam J.", role: "Contributor", status: "Offline" },
      { name: "Jordan M.", role: "Contributor", status: "Online" }
    );
  }

  // Build milestones dynamically based on project dates
  interface MilestoneUI {
    title: string;
    date: string;
    status: string;
  }
  
  const milestones: MilestoneUI[] = [];
  
  if (project.start_date) {
    milestones.push({
      title: `${project.name} Kickoff`,
      date: new Date(project.start_date).toLocaleDateString("en-US", { month: "short", day: "numeric" }),
      status: "Completed"
    });
  } else {
    milestones.push({
      title: "Project Initiation",
      date: "TBD",
      status: "Completed"
    });
  }

  if (project.due_date) {
    milestones.push({
      title: `${project.name} Target Delivery`,
      date: new Date(project.due_date).toLocaleDateString("en-US", { month: "short", day: "numeric" }),
      status: "Pending"
    });
  } else {
    milestones.push({
      title: "Target Milestones",
      date: "TBD",
      status: "Pending"
    });
  }

  return (
    <div className="space-y-10 pb-20">
      <div className="flex items-center gap-4">
        <button 
          onClick={() => router.back()}
          className="p-3 rounded-xl glass hover:bg-muted/30 transition-all text-muted-foreground group"
        >
          <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
        </button>
        <div className="flex-1">
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-3xl font-black text-foreground dark:text-primary-foreground tracking-tight">{project.name}</h1>
            <Badge variant={project.status === "active" ? "success" : "warning"} dot>{project.status}</Badge>
          </div>
          <p className="text-muted-foreground text-sm flex items-center gap-2">
            <Target className="w-4 h-4" /> {project.key} • <Calendar className="w-4 h-4 ml-2" /> Due: {project.due_date || "TBD"}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button className="flex items-center gap-2 px-4 py-2 rounded-xl bg-primary text-primary-foreground font-bold text-sm hover:opacity-90 transition-opacity shadow-lg shadow-primary/20">
            <Plus className="w-4 h-4" /> Add Task
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-3 space-y-8">
          {/* Progress Banner */}
          <div className="glass-card p-8 relative overflow-hidden group">
            <div className="absolute top-0 right-0 w-64 h-64 bg-primary/5 rounded-full -translate-y-1/2 translate-x-1/2 blur-3xl group-hover:bg-primary/10 transition-colors" />
            <div className="relative z-10">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground">Overall Progress</h3>
                <span className="text-2xl font-black text-primary">{completionRate}%</span>
              </div>
              <div className="h-4 w-full bg-muted/30 rounded-full overflow-hidden border border-border/50">
                <motion.div 
                  initial={{ width: 0 }}
                  animate={{ width: `${completionRate}%` }}
                  transition={{ duration: 1, ease: "easeOut" }}
                  className="h-full bg-gradient-to-r from-primary to-accent"
                />
              </div>
              <div className="grid grid-cols-3 gap-4 mt-8">
                <div>
                  <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Velocity</p>
                  <p className="text-lg font-bold text-foreground dark:text-primary-foreground flex items-center gap-1">
                    <TrendingUp className="w-4 h-4 text-success" /> 14.2 <span className="text-[10px] text-muted-foreground">pts/wk</span>
                  </p>
                </div>
                <div>
                  <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Tasks Done</p>
                  <p className="text-lg font-bold text-foreground dark:text-primary-foreground">25 / 50</p>
                </div>
                <div>
                  <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Days Remaining</p>
                  <p className="text-lg font-bold text-foreground dark:text-primary-foreground">12</p>
                </div>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            <section className="glass-card p-8">
              <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6">Task Distribution</h3>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={mockTaskData}
                      innerRadius={60}
                      outerRadius={80}
                      paddingAngle={5}
                      dataKey="value"
                    >
                      {mockTaskData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip 
                      contentStyle={{ backgroundColor: 'var(--card)', borderRadius: '12px', border: '1px solid var(--border)' }}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <div className="grid grid-cols-2 gap-4 mt-4">
                {mockTaskData.map(item => (
                  <div key={item.name} className="flex items-center gap-2">
                    <div className="w-2 h-2 rounded-full" style={{ backgroundColor: item.color }} />
                    <span className="text-[10px] font-bold text-muted-foreground uppercase">{item.name}: {item.value}</span>
                  </div>
                ))}
              </div>
            </section>

            <section className="glass-card p-8">
              <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6">Recent Activity</h3>
              <div className="space-y-6">
                {[
                  { user: "YK Lee", action: "Completed task", target: "API Authentication", time: "2h ago" },
                  { user: "Alex K.", action: "Commented on", target: "UI Redesign", time: "4h ago" },
                  { user: "Sam J.", action: "Added attachment", target: "Workflow Specs", time: "Yesterday" },
                ].map((act, i) => (
                  <div key={i} className="flex gap-4">
                    <div className="w-8 h-8 rounded-full bg-muted/30 border border-border flex-shrink-0" />
                    <div>
                      <p className="text-xs font-bold text-foreground dark:text-primary-foreground">
                        {act.user} <span className="font-normal text-muted-foreground">{act.action}</span> {act.target}
                      </p>
                      <p className="text-[10px] text-muted-foreground mt-1 uppercase tracking-widest">{act.time}</p>
                    </div>
                  </div>
                ))}
              </div>
            </section>
          </div>

          <section className="glass-card">
            <div className="p-8 border-b border-border/50 flex items-center justify-between">
              <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground">Active Tasks</h3>
              <button className="text-[10px] font-black uppercase text-primary hover:underline">View All</button>
            </div>
            <div className="divide-y divide-border/50">
              {[
                { title: "Implement RBAC Persistence", priority: "High", status: "In Progress", due: "May 20" },
                { title: "Dashboard Responsive Audit", priority: "Medium", status: "Review", due: "May 18" },
                { title: "Legacy Cleanup", priority: "Low", status: "To Do", due: "May 25" },
              ].map((task, i) => (
                <div key={i} className="p-6 flex items-center justify-between hover:bg-muted/5 transition-colors cursor-pointer group">
                  <div className="flex items-center gap-4">
                    <div className="w-2 h-2 rounded-full bg-primary" />
                    <div>
                      <h4 className="text-sm font-bold text-foreground dark:text-primary-foreground group-hover:text-primary transition-colors">{task.title}</h4>
                      <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">Due {task.due} • Priority {task.priority}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <MessageSquare className="w-4 h-4" />
                      <span className="text-[10px] font-bold">2</span>
                      <Paperclip className="w-4 h-4 ml-2" />
                      <span className="text-[10px] font-bold">1</span>
                    </div>
                    <Badge variant="glass">{task.status}</Badge>
                    <ChevronRight className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-all group-hover:translate-x-1" />
                  </div>
                </div>
              ))}
            </div>
          </section>
        </div>

        <div className="space-y-8">
          <section className="glass-card p-8">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6">Linked Repositories (N:M)</h3>
            {projectRepositories.length === 0 ? (
              <p className="text-sm text-muted-foreground">No linked repositories.</p>
            ) : (
              <div className="space-y-3">
                {projectRepositories.map((link) => (
                  <div key={`${link.project_id}-${link.repository_id}`} className="flex items-center justify-between rounded-xl border border-border/50 bg-muted/10 px-4 py-3">
                    <div>
                      <p className="text-xs font-bold text-foreground dark:text-primary-foreground">Repository ID: {link.repository_id}</p>
                      <p className="text-[10px] uppercase tracking-widest text-muted-foreground">Role: {link.role}</p>
                    </div>
                    <Badge variant={link.role === "primary" ? "primary" : "glass"}>{link.role}</Badge>
                  </div>
                ))}
              </div>
            )}
          </section>

          <section className="glass-card p-8">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
              <Users className="w-4 h-4 text-primary" /> Team Members
            </h3>
            <div className="space-y-6">
              {teamMembers.map((member, i) => (
                <div key={i} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="relative">
                      <div className="w-8 h-8 rounded-full bg-muted/40 border border-border" />
                      <div className={`absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-background ${
                        member.status === "Online" ? "bg-success" : 
                        member.status === "Busy" ? "bg-destructive" : "bg-muted-foreground"
                      }`} />
                    </div>
                    <div>
                      <p className="text-xs font-bold text-foreground dark:text-primary-foreground">{member.name}</p>
                      <p className="text-[10px] text-muted-foreground uppercase tracking-widest">{member.role}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
            <button className="w-full mt-8 py-3 rounded-xl bg-muted/30 border border-border text-[10px] font-black uppercase tracking-widest text-muted-foreground hover:bg-muted/50 transition-all">
              Manage Team
            </button>
          </section>

          <section className="glass-card p-8">
            <h3 className="text-sm font-black uppercase tracking-widest text-muted-foreground mb-6 flex items-center gap-2">
              <Clock className="w-4 h-4 text-accent" /> Upcoming Milestones
            </h3>
            <div className="space-y-6">
              {milestones.map((m, i) => (
                <div key={i} className="flex gap-3">
                  <div className="w-1 h-10 rounded-full bg-accent/20" />
                  <div>
                    <p className="text-xs font-bold text-foreground dark:text-primary-foreground">{m.title}</p>
                    <p className="text-[10px] text-muted-foreground mt-1 uppercase tracking-widest">{m.date} • {m.status}</p>
                  </div>
                </div>
              ))}
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
