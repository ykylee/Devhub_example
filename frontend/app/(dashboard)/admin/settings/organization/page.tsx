"use client";

import { useState, useEffect } from "react";
import { Plus, LayoutGrid, List, Network } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { OrgUnitTable } from "@/components/organization/OrgUnitTable";
import { OrgUnitGrid } from "@/components/organization/OrgUnitGrid";
import { OrgTree } from "@/components/organization/OrgTree";
import { MemberManagementModal } from "@/components/organization/MemberManagementModal";
import { UnitManagementModal } from "@/components/organization/UnitManagementModal";
import { identityService, OrgNode, OrgMember } from "@/lib/services/identity.service";

export default function AdminSettingsOrganizationPage() {
  const [view, setView] = useState<"list" | "grid" | "chart">("list");
  const [orgNodes, setOrgNodes] = useState<OrgNode[]>([]);
  const [members, setMembers] = useState<OrgMember[]>([]);
  const [unitMembers, setUnitMembers] = useState<Record<string, string[]>>({});
  const [unitLeaders, setUnitLeaders] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [managingUnitId, setManagingUnitId] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [editingUnitId, setEditingUnitId] = useState<string | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const [{ nodes }, allMembers] = await Promise.all([
        identityService.getOrgHierarchy(),
        identityService.getUsers(),
      ]);

      setOrgNodes(nodes);
      setMembers(allMembers);

      const memberMap: Record<string, string[]> = {};
      allMembers.forEach((m) => {
        // Source of truth for unit roster is appointments. current_dept_id is
        // a primary-view field and can diverge from per-unit membership.
        const unitIds = new Set<string>();
        m.appointments.forEach((a) => {
          if (a.dept_id) unitIds.add(a.dept_id);
        });
        if (unitIds.size === 0 && m.current_dept_id) {
          unitIds.add(m.current_dept_id);
        }
        unitIds.forEach((unitId) => {
          if (!memberMap[unitId]) memberMap[unitId] = [];
          memberMap[unitId].push(m.id);
        });
      });

      const userIdSet = new Set(allMembers.map((m) => m.id));
      const leaderMap: Record<string, string> = {};
      const leaderPatches: Array<{ unitId: string; leaderId: string }> = [];
      nodes.forEach((node) => {
        const rawLeader = node.data.leader_id;
        const unitMemberIds = memberMap[node.id] || [];

        if (rawLeader && userIdSet.has(rawLeader)) {
          leaderMap[node.id] = rawLeader;
          return;
        }

        // Legacy/mock-like IDs (e.g. u1/u2) or empty leader are normalized
        // to a real registered user in the same unit when available.
        if (unitMemberIds.length > 0) {
          const fallbackLeader = unitMemberIds[0];
          leaderMap[node.id] = fallbackLeader;
          if (rawLeader !== fallbackLeader) {
            leaderPatches.push({ unitId: node.id, leaderId: fallbackLeader });
          }
        }
      });

      if (leaderPatches.length > 0) {
        await Promise.all(
          leaderPatches.map(({ unitId, leaderId }) =>
            identityService.updateUnit(unitId, { leader_user_id: leaderId }).catch((err) => {
              console.warn(`[organization] failed to patch leader for ${unitId}:`, err);
            }),
          ),
        );
      }

      setUnitMembers(memberMap);
      setUnitLeaders(leaderMap);
      setError(null);
    } catch (err) {
      console.error("Failed to load organization data:", err);
      setError("Failed to connect to the organization service. Please try again later.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleCreateUnit = async (data: any) => {
    try {
      await identityService.createUnit(data);
      setIsCreating(false);
      loadData();
    } catch (err) {
      console.error("Failed to create unit:", err);
    }
  };

  const handleUpdateUnit = async (data: any) => {
    try {
      await identityService.updateUnit(data.unit_id, data);
      setEditingUnitId(null);
      loadData();
    } catch (err) {
      console.error("Failed to update unit:", err);
    }
  };

  const handleDeleteUnit = async (unitId: string) => {
    if (confirm("Are you sure you want to delete this unit? This cannot be undone.")) {
      try {
        await identityService.deleteUnit(unitId);
        loadData();
      } catch (err) {
        console.error("Failed to delete unit:", err);
      }
    }
  };

  const handleSaveUnitMembers = async (unitId: string, memberIds: string[], leaderId: string | null) => {
    try {
      await identityService.replaceUnitMembers(unitId, memberIds);
      // Also update leader if changed (this might be a separate API call in reality)
      if (leaderId) {
        await identityService.updateUnit(unitId, { leader_user_id: leaderId });
      }
      setManagingUnitId(null);
      loadData();
    } catch (err) {
      setSaveError("Failed to save member changes.");
    }
  };

  return (
    <div className="h-full flex flex-col">
      <div className="p-8 pb-4 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-black text-foreground tracking-tight">Organization</h1>
          <p className="text-sm text-muted-foreground font-medium">Manage units, teams, and hierarchical structure</p>
        </div>

        <div className="flex items-center gap-4">
          <div className="flex items-center bg-muted/30 rounded-2xl p-1 border border-border">
            <button
              onClick={() => setView("list")}
              className={`p-2 rounded-xl transition-all ${view === "list" ? "bg-primary text-primary-foreground shadow-lg shadow-primary/20" : "text-muted-foreground hover:text-foreground"}`}
              aria-label="List View"
            >
              <List className="w-4 h-4" />
            </button>
            <button
              onClick={() => setView("grid")}
              className={`p-2 rounded-xl transition-all ${view === "grid" ? "bg-primary text-primary-foreground shadow-lg shadow-primary/20" : "text-muted-foreground hover:text-foreground"}`}
              aria-label="Grid View"
            >
              <LayoutGrid className="w-4 h-4" />
            </button>
            <button
              onClick={() => setView("chart")}
              className={`p-2 rounded-xl transition-all ${view === "chart" ? "bg-primary text-primary-foreground shadow-lg shadow-primary/20" : "text-muted-foreground hover:text-foreground"}`}
              aria-label="Org Chart"
            >
              <Network className="w-4 h-4" />
            </button>
          </div>

          <button
            onClick={() => setIsCreating(true)}
            className="flex items-center gap-2 bg-accent hover:bg-accent/80 text-accent-foreground px-6 py-2.5 rounded-2xl font-black text-xs uppercase tracking-widest transition-all shadow-lg shadow-accent/20"
          >
            <Plus className="w-4 h-4" />
            Create Unit
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-8 pt-0 relative">
        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-32 gap-4">
            <div className="w-12 h-12 border-4 border-accent/20 border-t-accent rounded-full animate-spin" />
            <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">
              Initializing Organization Data...
            </p>
          </div>
        ) : error ? (
          <div className="glass border-red-500/20 bg-red-500/5 rounded-3xl p-12 text-center">
            <p className="text-red-400 font-bold mb-2">Service Unavailable</p>
            <p className="text-xs text-muted-foreground mb-6">{error}</p>
            <button
              onClick={() => loadData()}
              className="px-6 py-2 bg-muted/20 border border-border rounded-xl text-xs font-bold text-foreground hover:bg-muted/40 transition-all"
            >
              Retry Connection
            </button>
          </div>
        ) : (
          <AnimatePresence mode="wait">
            <motion.div
              key={view}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              transition={{ duration: 0.3 }}
            >
              {view === "list" && (
                <OrgUnitTable
                  nodes={orgNodes}
                  members={members}
                  unitLeaders={unitLeaders}
                  unitMembers={unitMembers}
                  onManage={(id) => setManagingUnitId(id)}
                  onEdit={(id) => setEditingUnitId(id)}
                  onDelete={(id) => handleDeleteUnit(id)}
                />
              )}
              {view === "grid" && (
                <OrgUnitGrid nodes={orgNodes} unitMembers={unitMembers} onManage={(id) => setManagingUnitId(id)} />
              )}
              {view === "chart" && <OrgTree />}
            </motion.div>
          </AnimatePresence>
        )}
      </div>

      <AnimatePresence>
        {managingUnitId && (
          <MemberManagementModal
            unitId={managingUnitId}
            unitName={orgNodes.find((n) => n.id === managingUnitId)?.data.label || "Unknown Unit"}
            allMembers={members}
            currentMemberIds={unitMembers[managingUnitId] || []}
            currentLeaderId={unitLeaders[managingUnitId] ?? null}
            onClose={() => {
              setSaveError(null);
              setManagingUnitId(null);
            }}
            onSave={(newMemberIds, newLeaderId) => handleSaveUnitMembers(managingUnitId, newMemberIds, newLeaderId)}
            saveError={saveError}
          />
        )}

        {isCreating && (
          <UnitManagementModal
            mode="create"
            availableParents={orgNodes}
            availableLeaders={members}
            onClose={() => setIsCreating(false)}
            onSave={handleCreateUnit}
          />
        )}

        {editingUnitId && (
          <UnitManagementModal
            mode="edit"
            initialData={{
              unit_id: editingUnitId,
              label: orgNodes.find((n) => n.id === editingUnitId)?.data.label,
              unit_type: orgNodes.find((n) => n.id === editingUnitId)?.data.type as any,
              parent_unit_id: "",
              leader_user_id: unitLeaders[editingUnitId] ?? "",
            }}
            availableParents={orgNodes}
            availableLeaders={members}
            onClose={() => setEditingUnitId(null)}
            onSave={handleUpdateUnit}
          />
        )}
      </AnimatePresence>
    </div>
  );
}
