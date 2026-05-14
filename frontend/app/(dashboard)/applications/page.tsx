"use client";

import { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { Search, Filter, Box } from "lucide-react";
import { projectService } from "@/lib/services/project.service";
import { Application } from "@/lib/services/project.types";
import { ApplicationTable } from "@/components/project/ApplicationTable";
import { useToast } from "@/components/ui/Toast";
import { useStore } from "@/lib/store";
import { isSystemAdmin } from "@/lib/auth/role-routing";

export default function ApplicationsPage() {
  const [applications, setApplications] = useState<Application[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [query, setQuery] = useState("");
  const { toast } = useToast();
  const actor = useStore(s => s.actor);
  const isAdmin = isSystemAdmin(actor?.role);

  const filteredApps = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return applications;
    return applications.filter((app) =>
      app.name.toLowerCase().includes(q) ||
      app.key.toLowerCase().includes(q) ||
      app.owner_user_id.toLowerCase().includes(q)
    );
  }, [applications, query]);

  const load = async () => {
    setIsLoading(true);
    try {
      const data = await projectService.getApplications();
      setApplications(data);
    } catch (error) {
      console.error("[applications] load failed:", error);
      if ((error as any).status === 501) {
        toast("Backend API not implemented yet (501).", "warning");
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  return (
    <div className="space-y-8 p-8">
      <div className="flex flex-col gap-2 mb-6">
        <h1 className="text-3xl font-black text-foreground tracking-tight uppercase">
          Applications
        </h1>
        <p className="text-sm text-muted-foreground font-medium uppercase tracking-widest opacity-60">
          Governance containers and delivery overview
        </p>
      </div>

      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex items-center gap-4 max-w-2xl">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search applications..."
            className="w-full glass border-border rounded-2xl pl-12 pr-6 py-3.5 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-purple-500/30 transition-all"
          />
        </div>
        <button
          type="button"
          className="glass border-border p-3.5 rounded-2xl text-muted-foreground hover:bg-muted/40 transition-all"
        >
          <Filter className="w-5 h-5" />
        </button>
      </motion.div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-purple-500/20 border-t-purple-500 rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Applications...</p>
        </div>
      ) : (
        <ApplicationTable
          applications={filteredApps}
          onEdit={() => {}} // Read-only for standard users
          onArchive={() => {}}
          onViewRepositories={(app) => {
            toast(`Viewing repositories for ${app.key}`, "glass");
          }}
        />
      )}
    </div>
  );
}
