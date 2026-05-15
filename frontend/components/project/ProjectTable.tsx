"use client";

import Link from "next/link";
import { FolderKanban, Clock, User, ChevronRight } from "lucide-react";
import { format } from "date-fns";
import { motion, AnimatePresence } from "framer-motion";
import { Project, ProjectStatus } from "@/lib/services/project.types";
import { Badge } from "@/components/ui/Badge";
import { ActionMenu } from "@/components/ui/ActionMenu";

interface ProjectTableProps {
  projects: Project[];
  onViewDetails?: (project: Project) => void;
}

export function ProjectTable({ 
  projects, 
  onViewDetails 
}: ProjectTableProps) {
  return (
    <div className="glass border-border rounded-3xl overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-border/60 bg-muted/20">
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Key & Project Name</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-center">Status</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Owner</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Timeline</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40">
            <AnimatePresence mode="popLayout">
              {projects.map((project) => (
                <motion.tr
                  key={project.id}
                  layout
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="group hover:bg-muted/30 transition-colors"
                >
                  <td className="px-6 py-5">
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-xl bg-indigo-500/10 flex items-center justify-center border border-indigo-500/20">
                        <FolderKanban className="w-5 h-5 text-indigo-400" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <Link href={`/projects/${project.id}`}>
                            <span className="text-xs font-black text-foreground dark:text-primary-foreground tracking-tight hover:text-indigo-400 transition-colors cursor-pointer">
                              {project.name}
                            </span>
                          </Link>
                          <span className="text-[9px] font-mono text-muted-foreground bg-muted/50 px-1.5 py-0.5 rounded border border-border/40 uppercase">
                            {project.key}
                          </span>
                        </div>
                        <p className="text-[10px] text-muted-foreground mt-1 opacity-60">
                          Repo ID: {project.repository_id}
                        </p>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-5 text-center">
                    <ProjectStatusBadge status={project.status} />
                  </td>
                  <td className="px-6 py-5">
                    <div className="flex items-center gap-2">
                      <User className="w-3.5 h-3.5 text-muted-foreground/40" />
                      <span className="text-[11px] font-bold text-foreground/80">{project.owner_user_id}</span>
                    </div>
                  </td>
                  <td className="px-6 py-5 text-[10px] font-medium text-muted-foreground">
                    <div className="flex items-center gap-1.5">
                      <Clock className="w-3 h-3 opacity-40" />
                      <span>
                        {project.start_date ? format(new Date(project.start_date), "MMM d") : "TBD"}
                        {" → "}
                        {project.due_date ? format(new Date(project.due_date), "MMM d") : "TBD"}
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-5 text-right">
                    <ActionMenu
                      title={`${project.name} Actions`}
                      items={[
                        {
                          key: "view-details",
                          label: "View Details",
                          icon: <ChevronRight className="w-4 h-4" />,
                          onClick: () => onViewDetails?.(project),
                        },
                        {
                          key: "edit-project",
                          label: "Edit Project",
                          icon: <FolderKanban className="w-4 h-4" />,
                          onClick: () => {}
                        }
                      ]}
                    />
                  </td>
                </motion.tr>
              ))}
            </AnimatePresence>
          </tbody>
        </table>
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
