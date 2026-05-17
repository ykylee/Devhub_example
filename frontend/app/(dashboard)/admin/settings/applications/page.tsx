"use client";

import { useEffect, useState, useMemo } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Plus } from "lucide-react";
import { projectService } from "@/lib/services/project.service";
import { Application } from "@/lib/services/project.types";
import { ApplicationTable } from "@/components/project/ApplicationTable";
import { ApplicationCreationModal } from "@/components/project/ApplicationCreationModal";
import { FilterBar } from "@/components/ui/FilterBar";
import { useToast } from "@/components/ui/Toast";

const STATUS_OPTIONS = [
  { label: "All Status", value: "all" },
  { label: "Active", value: "active" },
  { label: "Planning", value: "planning" },
  { label: "On Hold", value: "on_hold" },
  { label: "Closed", value: "closed" },
  { label: "Archived", value: "archived" },
];

export default function AdminSettingsApplicationsPage() {
  const [applications, setApplications] = useState<Application[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [activeStatus, setActiveStatus] = useState("all");
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingApp, setEditingApp] = useState<Application | null>(null);
  const { toast } = useToast();

  const filteredApplications = useMemo(() => {
    const q = query.trim().toLowerCase();
    return applications.filter((app) => {
      const matchesQuery = !q || 
        app.name.toLowerCase().includes(q) ||
        app.key.toLowerCase().includes(q) ||
        app.description.toLowerCase().includes(q);
      
      const matchesStatus = activeStatus === "all" || app.status === activeStatus;
      
      return matchesQuery && matchesStatus;
    });
  }, [applications, query, activeStatus]);

  const refresh = async () => {
    setIsLoading(true);
    try {
      const data = await projectService.getApplications();
      setApplications(data);
    } catch (error) {
      console.error("[admin/settings/applications] load failed:", error);
      if ((error as { status?: number })?.status === 501) {
        toast("Backend API not implemented yet (501). Showing empty list.", "warning");
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    refresh();
  }, []);

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
        <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex-1">
          <FilterBar 
            onSearch={setQuery}
            onFilterChange={setActiveStatus}
            activeFilter={activeStatus}
            filterOptions={STATUS_OPTIONS}
            placeholder="Search by name, key, or owner..."
          />
        </motion.div>

        <motion.button
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          onClick={handleCreate}
          className="bg-primary text-primary-foreground px-6 py-3.5 rounded-2xl font-black uppercase tracking-widest text-[10px] flex items-center gap-2 shadow-xl shadow-primary/20 transition-all border border-primary/20"
        >
          <Plus className="w-4 h-4" />
          New Application
        </motion.button>
      </div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-32 gap-4">
          <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
          <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Applications...</p>
        </div>
      ) : (
        <ApplicationTable
          applications={filteredApplications}
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
