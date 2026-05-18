"use client";

import { useEffect, useMemo, useState } from "react";
import {
  ReactFlow,
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  BackgroundVariant,
  type Edge,
  type Node,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { motion } from "framer-motion";
import { Activity, AlertTriangle, ArrowLeft, Globe, Server } from "lucide-react";
import { format, parseISO } from "date-fns";
import { infraService } from "@/lib/services/infra.service";
import type {
  ApiInfraNodeV2,
  ApiInfraServiceV2,
  InfraTopologyV2Meta,
} from "@/lib/services/infra.service";
import { Badge } from "@/components/ui/Badge";
import { Modal } from "@/components/ui/Modal";
import { cn } from "@/lib/utils";
import Link from "next/link";

// Infra topology v2 (HomeLab snapshot 기반) — sprint claude/work_260518-n.
// backend: API-76 /api/v1/infra/services + API-78 /api/v1/infra/topology/v2
// (PR #139, sprint codex/next-step-20260516). 본 페이지는 v2 응답의 nodes 를
// React Flow 로, services 를 사이드바 list 로 노출. degraded_providers 는
// 페이지 헤더의 banner 로 시각화 (ADR-0015/0016 의 운영 의도 정합).

type NodeV2Data = {
  label: string;
  hostname: string;
  status: string;
  environment?: string;
  ipAddress?: string;
};

function nodeStatusVariant(status: string): {
  badge: "success" | "warning" | "danger" | "secondary";
  containerCN: string;
} {
  const norm = status.toLowerCase();
  if (norm === "healthy" || norm === "stable" || norm === "ok") {
    return {
      badge: "success",
      containerCN: "bg-emerald-500/10 border-emerald-500/30 text-emerald-400",
    };
  }
  if (norm === "degraded" || norm === "warning") {
    return {
      badge: "warning",
      containerCN: "bg-amber-500/10 border-amber-500/30 text-amber-400",
    };
  }
  if (norm === "down" || norm === "error" || norm === "failed") {
    return {
      badge: "danger",
      containerCN: "bg-rose-500/10 border-rose-500/30 text-rose-400",
    };
  }
  return {
    badge: "secondary",
    containerCN: "bg-muted/30 border-border text-muted-foreground",
  };
}

function safeFormat(iso: string | null | undefined): string {
  if (!iso) return "—";
  try {
    return format(parseISO(iso), "yyyy-MM-dd HH:mm");
  } catch {
    return iso;
  }
}

export default function AdminTopologyV2Page() {
  const [rawNodes, setRawNodes] = useState<ApiInfraNodeV2[]>([]);
  const [services, setServices] = useState<ApiInfraServiceV2[]>([]);
  const [meta, setMeta] = useState<InfraTopologyV2Meta | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<ApiInfraNodeV2 | null>(null);
  const [nodes, setNodes, onNodesChange] = useNodesState<Node<NodeV2Data>>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  // 1회 로드 — websocket 갱신은 후속 carve out.
  useEffect(() => {
    let cancelled = false;
    (async () => {
      setIsLoading(true);
      try {
        const resp = await infraService.getTopologyV2();
        if (cancelled) return;
        setRawNodes(resp.nodes);
        setServices(resp.services);
        setMeta(resp.meta);

        const flowNodes: Node<NodeV2Data>[] = resp.nodes.map((n, idx) => ({
          id: n.node_id,
          position: {
            x: 80 + (idx % 3) * 280,
            y: 80 + Math.floor(idx / 3) * 200,
          },
          data: {
            label: n.hostname || n.node_id,
            hostname: n.hostname,
            status: n.status,
            environment: n.environment,
            ipAddress: n.ip_address,
          },
          className: cn(
            "glass rounded-2xl p-5 font-black shadow-2xl min-w-[200px] text-center border transition-all duration-500",
            nodeStatusVariant(n.status).containerCN,
          ),
        }));

        const flowEdges: Edge[] = resp.edges.map((e) => ({
          id: e.id,
          source: e.source_id,
          target: e.target_id,
          label: e.label,
          animated: true,
          style: {
            stroke:
              e.status === "stable" || e.status === "healthy"
                ? "#10b981"
                : "#f59e0b",
            strokeWidth: 2,
          },
        }));

        setNodes(flowNodes);
        setEdges(flowEdges);
      } catch (err) {
        if (cancelled) return;
        console.error("[admin/topology-v2] load failed:", err);
        setErrorMsg("토폴로지 v2 데이터를 불러오지 못했습니다.");
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
    // setNodes/setEdges 의 stable identity — Next 가드 lint 통과 위한 minimal dep.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const servicesByNode = useMemo(() => {
    const map: Record<string, ApiInfraServiceV2[]> = {};
    for (const s of services) {
      if (!map[s.node_id]) map[s.node_id] = [];
      map[s.node_id].push(s);
    }
    return map;
  }, [services]);

  const onNodeClick = (_: unknown, node: Node<NodeV2Data>) => {
    const raw = rawNodes.find((n) => n.node_id === node.id) ?? null;
    setSelectedNode(raw);
  };

  const degradedCount = meta?.degraded_providers?.length ?? 0;

  return (
    <div className="space-y-8 h-full flex flex-col pb-10">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }}>
          <div className="flex items-center gap-3 mb-2">
            <Link
              href="/admin"
              className="inline-flex items-center gap-1 text-[10px] font-black text-muted-foreground uppercase tracking-widest hover:text-foreground dark:hover:text-primary-foreground transition-colors"
            >
              <ArrowLeft className="w-3 h-3" />
              v1 Dashboard
            </Link>
          </div>
          <h1 className="text-4xl font-extrabold tracking-tight text-foreground dark:text-primary-foreground mb-2">
            Topology <span className="text-gradient">v2</span>
          </h1>
          <p className="text-muted-foreground text-sm flex items-center gap-2 flex-wrap">
            <Globe className="w-4 h-4 text-primary" />
            HomeLab snapshot 기반 노드/서비스 시각화
            <span className="ml-2 text-[10px] font-mono bg-muted/30 border border-border rounded-md px-2 py-0.5">
              Last snapshot: {meta ? safeFormat(meta.snapshot_at) : "—"}
            </span>
          </p>
        </motion.div>
      </div>

      {/* Degraded providers banner */}
      {degradedCount > 0 && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="glass border border-amber-500/40 bg-amber-500/10 rounded-2xl p-4 flex items-center gap-3"
          role="alert"
        >
          <AlertTriangle className="w-5 h-5 text-amber-400 flex-shrink-0" />
          <div className="flex-1">
            <p className="text-xs font-black text-amber-300 uppercase tracking-widest">
              Degraded providers ({degradedCount})
            </p>
            <p className="text-[11px] text-amber-200/80 font-mono mt-1">
              {meta!.degraded_providers.join(", ")}
            </p>
          </div>
        </motion.div>
      )}

      {/* Error state */}
      {errorMsg && (
        <div className="glass border border-rose-500/40 bg-rose-500/10 rounded-2xl p-4 text-xs text-rose-300 font-bold">
          {errorMsg}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-[1fr_320px] gap-6 flex-1 min-h-[600px]">
        {/* React Flow canvas */}
        <section className="glass rounded-3xl border border-border overflow-hidden relative shadow-2xl">
          {isLoading && (
            <div className="absolute inset-0 z-30 flex flex-col items-center justify-center bg-background/50 backdrop-blur-sm">
              <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin mb-4" />
              <p className="text-foreground/50 dark:text-primary-foreground/50 text-xs font-bold uppercase tracking-[0.2em] animate-pulse">
                Loading Topology v2...
              </p>
            </div>
          )}

          {!isLoading && nodes.length === 0 && !errorMsg && (
            <div className="absolute inset-0 z-20 flex flex-col items-center justify-center gap-3 text-center px-6">
              <Server className="w-12 h-12 text-muted-foreground/30" />
              <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest">
                등록된 infra node 가 없습니다
              </p>
              <p className="text-[10px] text-muted-foreground/60 max-w-md">
                HomeLab agent 가 snapshot 을 ingest 하면 노드/서비스가 자동 표시됩니다.
              </p>
            </div>
          )}

          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onNodeClick={onNodeClick}
            fitView
            colorMode="dark"
          >
            <Controls className="glass border-border rounded-xl overflow-hidden" />
            <MiniMap
              maskColor="rgba(3, 0, 20, 0.8)"
              className="glass border-border rounded-2xl overflow-hidden"
              style={{ background: "transparent" }}
            />
            <Background variant={BackgroundVariant.Lines} gap={30} size={1} color="rgba(255,255,255,0.03)" />
          </ReactFlow>
        </section>

        {/* Services sidebar */}
        <aside className="glass rounded-3xl border border-border overflow-hidden">
          <div className="p-5 border-b border-border/60">
            <h2 className="text-xs font-black text-foreground dark:text-primary-foreground uppercase tracking-widest flex items-center gap-2">
              <Activity className="w-4 h-4 text-primary" />
              Services ({services.length})
            </h2>
          </div>
          <div className="max-h-[600px] overflow-y-auto p-3 space-y-2">
            {services.length === 0 ? (
              <p className="text-[10px] text-muted-foreground/60 text-center py-8 font-bold uppercase tracking-widest">
                no services
              </p>
            ) : (
              services.map((s) => {
                const variant = nodeStatusVariant(s.health_status);
                return (
                  <div
                    key={`${s.node_id}::${s.service_id}`}
                    className="glass-card p-3 flex flex-col gap-1"
                    data-service-id={s.service_id}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-xs font-black text-foreground dark:text-primary-foreground truncate">
                        {s.name}
                      </span>
                      <Badge variant={variant.badge}>{s.health_status}</Badge>
                    </div>
                    <p className="text-[10px] text-muted-foreground font-mono truncate">
                      node: {s.node_id}
                      {s.port ? ` · :${s.port}` : ""}
                      {s.version ? ` · v${s.version}` : ""}
                    </p>
                  </div>
                );
              })
            )}
          </div>
        </aside>
      </div>

      {/* Node detail modal */}
      <Modal
        isOpen={!!selectedNode}
        onClose={() => setSelectedNode(null)}
        title="Node Detail (v2)"
        size="md"
      >
        {selectedNode && (
          <div className="space-y-6">
            <div className="flex items-center justify-between gap-4">
              <div>
                <h4 className="text-xl font-black text-foreground dark:text-primary-foreground tracking-tight">
                  {selectedNode.hostname || selectedNode.node_id}
                </h4>
                <p className="text-xs text-muted-foreground font-mono mt-1">
                  {selectedNode.node_id}
                </p>
              </div>
              <Badge variant={nodeStatusVariant(selectedNode.status).badge} dot>
                {selectedNode.status}
              </Badge>
            </div>

            <div className="grid grid-cols-2 gap-3 text-xs">
              <div className="glass-card p-3">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">
                  Environment
                </p>
                <p className="font-bold text-foreground dark:text-primary-foreground">
                  {selectedNode.environment || "—"}
                </p>
              </div>
              <div className="glass-card p-3">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">
                  IP Address
                </p>
                <p className="font-mono text-foreground dark:text-primary-foreground">
                  {selectedNode.ip_address || "—"}
                </p>
              </div>
              <div className="glass-card p-3 col-span-2">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">
                  Observed At
                </p>
                <p className="font-mono text-foreground dark:text-primary-foreground">
                  {safeFormat(selectedNode.observed_at)}
                </p>
              </div>
            </div>

            {servicesByNode[selectedNode.node_id]?.length ? (
              <div className="space-y-2">
                <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest">
                  Services on this node ({servicesByNode[selectedNode.node_id].length})
                </p>
                <div className="space-y-1.5 max-h-[200px] overflow-y-auto">
                  {servicesByNode[selectedNode.node_id].map((s) => (
                    <div
                      key={s.service_id}
                      className="flex items-center justify-between gap-2 text-xs px-3 py-2 rounded-lg bg-muted/30"
                    >
                      <span className="font-bold text-foreground dark:text-primary-foreground truncate">
                        {s.name}
                      </span>
                      <Badge variant={nodeStatusVariant(s.health_status).badge}>
                        {s.health_status}
                      </Badge>
                    </div>
                  ))}
                </div>
              </div>
            ) : null}
          </div>
        )}
      </Modal>
    </div>
  );
}
