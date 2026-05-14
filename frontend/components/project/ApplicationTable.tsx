"use client";

import { useState } from "react";
import Link from "next/link";
import { 
  Box, 
  MoreHorizontal, 
  ExternalLink, 
  Archive, 
  Edit2, 
  Users, 
  GitBranch, 
  Clock, 
  Eye, 
  Shield, 
  Globe 
} from "lucide-react";
import { format } from "date-fns";
import { motion, AnimatePresence } from "framer-motion";
import { Application, ApplicationStatus, ApplicationVisibility } from "@/lib/services/project.types";
import { Badge } from "@/components/ui/Badge";
import { ActionMenu } from "@/components/ui/ActionMenu";
import { cn } from "@/lib/utils";

interface ApplicationTableProps {
  applications: Application[];
  onEdit: (app: Application) => void;
  onArchive: (app: Application) => void;
  onViewRepositories: (app: Application) => void;
}

export function ApplicationTable({ 
  applications, 
  onEdit, 
  onArchive, 
  onViewRepositories 
}: ApplicationTableProps) {
  return (
    <div className="glass border-border rounded-3xl overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-border/60 bg-muted/20">
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Key & Name</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-center">Status</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-center">Visibility</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Owner</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Period</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40">
            <AnimatePresence mode="popLayout">
              {applications.map((app) => (
                <motion.tr
                  key={app.id}
                  layout
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="group hover:bg-muted/30 transition-colors"
                >
                  <td className="px-6 py-5">
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-xl bg-purple-500/10 flex items-center justify-center border border-purple-500/20">
                        <Box className="w-5 h-5 text-purple-400" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <Link href={`/applications/${app.id}`}>
                            <span className="text-xs font-black text-foreground dark:text-primary-foreground tracking-tight hover:text-purple-400 transition-colors cursor-pointer">{app.name}</span>
                          </Link>
                          <span className="text-[10px] font-mono text-muted-foreground bg-muted/50 px-1.5 py-0.5 rounded border border-border/40 uppercase tracking-tighter">
                            {app.key}
                          </span>
                        </div>
                        <p className="text-[10px] text-muted-foreground mt-1 line-clamp-1 opacity-60">
                          {app.description || "No description provided."}
                        </p>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-5 text-center">
                    <StatusBadge status={app.status} />
                  </td>
                  <td className="px-6 py-5 text-center">
                    <VisibilityBadge visibility={app.visibility} />
                  </td>
                  <td className="px-6 py-5">
                    <div className="flex items-center gap-2">
                      <div className="w-6 h-6 rounded-full bg-accent/20 flex items-center justify-center border border-accent/20">
                        <span className="text-[8px] font-black text-accent uppercase">
                          {app.owner_user_id?.substring(0, 2) || "??"}
                        </span>
                      </div>
                      <span className="text-[11px] font-bold text-foreground/80">{app.owner_user_id}</span>
                    </div>
                  </td>
                  <td className="px-6 py-5 text-[10px] font-medium text-muted-foreground">
                    <div className="flex flex-col gap-1">
                      <div className="flex items-center gap-1.5">
                        <Clock className="w-3 h-3 opacity-40" />
                        <span>{app.start_date ? format(new Date(app.start_date), "MMM d, yyyy") : "TBD"}</span>
                      </div>
                      <div className="flex items-center gap-1.5 ml-4 opacity-50">
                        <span>→ {app.due_date ? format(new Date(app.due_date), "MMM d, yyyy") : "TBD"}</span>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-5 text-right">
                    <ActionMenu
                      title={`${app.name} Actions`}
                      items={[
                        {
                          label: "View Repositories",
                          icon: <GitBranch className="w-4 h-4" />,
                          onClick: () => onViewRepositories(app),
                          primary: true
                        },
                        {
                          label: "Edit Meta",
                          icon: <Edit2 className="w-4 h-4" />,
                          onClick: () => onEdit(app)
                        },
                        {
                          label: "Archive",
                          icon: <Archive className="w-4 h-4" />,
                          onClick: () => onArchive(app),
                          danger: true
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
      {applications.length === 0 && (
        <div className="py-24 flex flex-col items-center justify-center gap-4 text-center">
          <div className="w-16 h-16 rounded-3xl bg-muted/20 flex items-center justify-center border border-border/40">
            <Box className="w-8 h-8 text-muted-foreground/30" />
          </div>
          <div>
            <p className="text-sm font-black text-foreground/60 uppercase tracking-widest">No Applications Found</p>
            <p className="text-[10px] text-muted-foreground font-bold mt-1 uppercase tracking-widest opacity-40">
              Create your first governance container
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: ApplicationStatus }) {
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

function VisibilityBadge({ visibility }: { visibility: ApplicationVisibility }) {
  switch (visibility) {
    case "public":
      return (
        <Badge variant="glass" className="border-blue-500/20 text-blue-400 bg-blue-500/5">
          <Globe className="w-3 h-3" /> Public
        </Badge>
      );
    case "internal":
      return (
        <Badge variant="glass" className="border-emerald-500/20 text-emerald-400 bg-emerald-500/5">
          <Eye className="w-3 h-3" /> Internal
        </Badge>
      );
    case "restricted":
      return (
        <Badge variant="glass" className="border-rose-500/20 text-rose-400 bg-rose-500/5">
          <Shield className="w-3 h-3" /> Restricted
        </Badge>
      );
    default:
      return <Badge variant="glass">{visibility}</Badge>;
  }
}
