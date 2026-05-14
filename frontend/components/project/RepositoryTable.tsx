"use client";

import Link from "next/link";
import { GitBranch, ExternalLink, Activity, ShieldCheck, AlertCircle, Unlink } from "lucide-react";
import { format } from "date-fns";
import { motion, AnimatePresence } from "framer-motion";
import { ApplicationRepository, ApplicationRepositorySyncStatus, SyncErrorCode } from "@/lib/services/project.types";
import { Badge } from "@/components/ui/Badge";
import { ActionMenu } from "@/components/ui/ActionMenu";

interface RepositoryTableProps {
  repositories: ApplicationRepository[];
  onDisconnect?: (repo: ApplicationRepository) => void;
  showApplicationColumn?: boolean;
}

export function RepositoryTable({ 
  repositories, 
  onDisconnect,
  showApplicationColumn = false 
}: RepositoryTableProps) {
  return (
    <div className="glass border-border rounded-3xl overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-border/60 bg-muted/20">
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Provider & Repository</th>
              {showApplicationColumn && (
                <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Application</th>
              )}
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-center">Sync Status</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-center">Role</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest">Last Sync</th>
              <th className="px-6 py-5 text-[10px] font-black text-muted-foreground uppercase tracking-widest text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40">
            <AnimatePresence mode="popLayout">
              {repositories.map((repo) => (
                <motion.tr
                  key={`${repo.repo_provider}/${repo.repo_full_name}`}
                  layout
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="group hover:bg-muted/30 transition-colors"
                >
                  <td className="px-6 py-5">
                    <div className="flex items-center gap-4">
                      <div className="w-10 h-10 rounded-xl bg-pink-500/10 flex items-center justify-center border border-pink-500/20">
                        <GitBranch className="w-5 h-5 text-pink-400" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <Link href={`/repositories/${repo.repo_provider}/${repo.repo_full_name}`}>
                            <span className="text-xs font-black text-foreground dark:text-primary-foreground tracking-tight hover:text-pink-400 transition-colors cursor-pointer">
                              {repo.repo_full_name}
                            </span>
                          </Link>
                        </div>
                        <p className="text-[10px] text-muted-foreground mt-1 uppercase tracking-widest opacity-60">
                          {repo.repo_provider}
                        </p>
                      </div>
                    </div>
                  </td>
                  {showApplicationColumn && (
                    <td className="px-6 py-5">
                       <span className="text-[11px] font-bold text-muted-foreground">{repo.application_id}</span>
                    </td>
                  )}
                  <td className="px-6 py-5 text-center">
                    <SyncStatusBadge status={repo.sync_status} errorCode={repo.sync_error_code} />
                  </td>
                  <td className="px-6 py-5 text-center">
                    <Badge variant={repo.role === 'primary' ? 'primary' : 'glass'}>
                      {repo.role}
                    </Badge>
                  </td>
                  <td className="px-6 py-5 text-[10px] font-medium text-muted-foreground">
                    <div className="flex items-center gap-1.5">
                      <Activity className="w-3 h-3 opacity-40" />
                      <span>{repo.last_sync_at ? format(new Date(repo.last_sync_at), "MMM d, HH:mm") : "Never"}</span>
                    </div>
                  </td>
                  <td className="px-6 py-5 text-right">
                    <ActionMenu
                      title={`${repo.repo_full_name} Actions`}
                      items={[
                        {
                          label: "View PRs",
                          icon: <GitBranch className="w-4 h-4" />,
                          onClick: () => {},
                          primary: true
                        },
                        {
                          label: "View Metrics",
                          icon: <Activity className="w-4 h-4" />,
                          onClick: () => {}
                        },
                        onDisconnect ? {
                          label: "Disconnect",
                          icon: <Unlink className="w-4 h-4" />,
                          onClick: () => onDisconnect(repo),
                          danger: true
                        } : null
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

function SyncStatusBadge({ status, errorCode }: { status: ApplicationRepositorySyncStatus, errorCode?: SyncErrorCode }) {
  switch (status) {
    case "active":
      return <Badge variant="success" dot>Active</Badge>;
    case "requested":
    case "verifying":
      return <Badge variant="primary" dot className="animate-pulse">Syncing</Badge>;
    case "degraded":
      return (
        <div className="flex flex-col items-center gap-1">
          <Badge variant="warning" dot>Degraded</Badge>
          {errorCode && <span className="text-[8px] font-mono text-amber-500 opacity-70">{errorCode}</span>}
        </div>
      );
    case "disconnected":
      return <Badge variant="secondary">Disconnected</Badge>;
    default:
      return <Badge variant="glass">{status}</Badge>;
  }
}
