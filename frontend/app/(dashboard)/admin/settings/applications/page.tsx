"use client";

import { useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Search, Filter, Plus } from "lucide-react";
import { projectService } from "@/lib/services/project.service";
import { Application } from "@/lib/services/project.types";
import { ApplicationTable } from "@/components/project/ApplicationTable";
import { ApplicationCreationModal } from "@/components/project/ApplicationCreationModal";
import { useToast } from "@/components/ui/Toast";

export default function AdminSettingsApplicationsPage() {
  const [applications, setApplications] = useState<Application[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingApp, setEditingApp] = useState<Application | null>(null);
  const { toast } = useToast();

  const refresh = async (searchQuery: string = query) => {
    setIsLoading(true);
    try {
      const q = searchQuery.trim();
      const data = await projectService.getApplications(q ? { q } : undefined);
      setApplications(data);
    } catch (error) {
      console.error("[admin/settings/applications] load failed:", error);
      // For demo/design purposes, if 501, we might show empty or mock
      if ((error as { status?: number })?.status === 501) {
        toast("Backend API not implemented yet (501). Showing empty list.", "warning");
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const q = query.trim();
        const data = await projectService.getApplications(q ? { q } : undefined);
        if (!cancelled) setApplications(data);
      } catch (error) {
        console.error("[admin/settings/applications] load failed:", error);
        if (!cancelled && (error as { status?: number })?.status === 501) {
          toast("Backend API not implemented yet (501). Showing empty list.", "warning");
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [toast, query]);

  const handleCreate = () => {
    setEditingApp(null);
    setShowCreateModal(true);
  };

  const handleEdit = (app: Application) => {
    setEditingApp(app);
    setShowCreateModal(true);
  };

  const handleArchive = async (app: Application) => {
    if (!confirm(`Are you sure you want to archive ${app.name}?`)) return;
    try {
      await projectService.archiveApplication(app.id);
      toast(`Application ${app.name} archived`, "success");
      refresh();
    } catch {
      toast("Failed to archive application", "error");
    }
  };

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex items-center gap-4 flex-1">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search by name, key, or owner..."
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

        <motion.button
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          onClick={handleCreate}
          className="bg-purple-600 text-white px-6 py-3.5 rounded-2xl font-black uppercase tracking-widest text-[10px] flex items-center gap-2 shadow-xl shadow-purple-500/20 transition-all"
        >
          <Plus className="w-4 h-4" />
          New Application
        </motion.button>
      </div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-purple-500/20 border-t-purple-500 rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Applications...</p>
        </div>
      ) : (
        <ApplicationTable
          applications={applications}
          onEdit={handleEdit}
          onArchive={handleArchive}
          onViewRepositories={(app) => {
            toast(`Viewing repositories for ${app.key} (Coming soon)`, "info");
          }}
        />
      )}

      <AnimatePresence>
        {showCreateModal && (
          <ApplicationCreationModal
            initialData={editingApp || undefined}
            onClose={() => setShowCreateModal(false)}
            onCreated={(newApp) => {
              toast(`Application ${newApp.name} ${editingApp ? 'updated' : 'created'}`, "success");
              refresh();
            }}
          />
        )}
      </AnimatePresence>
    </div>
  );
}
