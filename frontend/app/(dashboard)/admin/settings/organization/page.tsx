"use client";

import { useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Layers, Network } from "lucide-react";
import { cn } from "@/lib/utils";
import { identityService, OrgMember, OrgNode } from "@/lib/services/identity.service";
import { OrgUnitGrid } from "@/components/organization/OrgUnitGrid";
import { OrgTree } from "@/components/organization/OrgTree";
import { MemberManagementModal } from "@/components/organization/MemberManagementModal";

type View = "units" | "chart";

const views: { id: View; label: string; icon: React.ComponentType<{ className?: string }> }[] = [
  { id: "units", label: "Org Units", icon: Layers },
  { id: "chart", label: "Org Chart", icon: Network },
];

export default function AdminSettingsOrganizationPage() {
  const [view, setView] = useState<View>("units");
  const [members, setMembers] = useState<OrgMember[]>([]);
  const [orgNodes, setOrgNodes] = useState<OrgNode[]>([]);
  const [unitMembers, setUnitMembers] = useState<Record<string, string[]>>({});
  const [unitLeaders, setUnitLeaders] = useState<Record<string, string | null>>({});
  const [managingUnitId, setManagingUnitId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [saveError, setSaveError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      setIsLoading(true);
      try {
        const [usersData, orgData] = await Promise.all([
          identityService.getUsers(),
          identityService.getOrgHierarchy(),
        ]);
        setMembers(usersData);
        setOrgNodes(orgData.nodes);
        // Seed leader map from the hierarchy snapshot. The MemberManagementModal
        // shows / edits leader assignment alongside roster membership.
        const leaders: Record<string, string | null> = {};
        for (const node of orgData.nodes) {
          leaders[node.id] = node.data.leader_id ?? null;
        }
        setUnitLeaders(leaders);

        const memberLists = await Promise.all(
          orgData.nodes.map(async (node): Promise<[string, string[]]> => {
            try {
              const list = await identityService.getUnitMembers(node.id);
              return [node.id, list.map((m) => m.id)];
            } catch (error) {
              console.error(`[admin/settings/organization] getUnitMembers ${node.id} failed:`, error);
              return [node.id, []];
            }
          }),
        );
        const fetched: Record<string, string[]> = {};
        memberLists.forEach(([id, ids]) => {
          fetched[id] = ids;
        });
        setUnitMembers(fetched);
      } catch (error) {
        console.error("[admin/settings/organization] load failed:", error);
      } finally {
        setIsLoading(false);
      }
    };
    load();
  }, []);

  const handleSaveUnitMembers = async (unitId: string, newMemberIds: string[], newLeaderId: string | null) => {
    setSaveError(null);
    try {
      const refreshed = await identityService.replaceUnitMembers(unitId, newMemberIds);
      setUnitMembers((prev) => ({ ...prev, [unitId]: refreshed.map((m) => m.id) }));

      // Persist leader change as a unit PATCH only when it actually moved.
      // Backend audits the unit update; we only need to keep DevHub side
      // consistent with what the modal showed.
      const previousLeaderId = unitLeaders[unitId] ?? null;
      if ((newLeaderId ?? null) !== previousLeaderId) {
        await identityService.updateUnit(unitId, { leader_user_id: newLeaderId ?? "" });
        setUnitLeaders((prev) => ({ ...prev, [unitId]: newLeaderId }));
        // OrgUnitGrid renders the leader badge from node.data.leader_id, so
        // mirror the change into orgNodes too — without this the card keeps
        // the stale leader until the next full reload.
        setOrgNodes((prev) =>
          prev.map((node) =>
            node.id === unitId
              ? { ...node, data: { ...node.data, leader_id: newLeaderId ?? undefined } }
              : node,
          ),
        );
      }

      setManagingUnitId(null);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to save unit";
      console.error("[admin/settings/organization] save unit failed:", error);
      setSaveError(message);
    }
  };

  return (
    <div className="space-y-8">
      <div className="flex p-1.5 glass border-white/10 rounded-2xl gap-1 w-fit">
        {views.map((v) => {
          const isActive = view === v.id;
          return (
            <button
              key={v.id}
              onClick={() => setView(v.id)}
              className={cn(
                "flex items-center gap-2 px-5 py-2 rounded-xl text-[11px] font-bold uppercase tracking-widest transition-all relative",
                isActive ? "text-foreground dark:text-white" : "text-muted-foreground hover:text-foreground dark:hover:text-white",
              )}
            >
              {isActive && (
                <motion.div
                  layoutId="org-view"
                  className="absolute inset-0 bg-white/10 border border-white/10 rounded-xl"
                  transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                />
              )}
              <v.icon className={cn("w-4 h-4 relative z-10", isActive ? "text-accent" : "text-muted-foreground")} />
              <span className="relative z-10">{v.label}</span>
            </button>
          );
        })}
      </div>

      <div className="relative">
        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-32 gap-4">
            <div className="w-12 h-12 border-4 border-accent/20 border-t-accent rounded-full animate-spin" />
            <p className="text-muted-foreground font-bold animate-pulse uppercase tracking-[0.3em] text-[10px]">Loading Organization...</p>
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
              {view === "units" && (
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
      </AnimatePresence>
    </div>
  );
}
