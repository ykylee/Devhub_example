"use client";

import { Activity, Settings, Shield, Globe, Zap, Terminal } from "lucide-react";
import { 
  ReactFlow,
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
  BackgroundVariant,
  type Connection,
  type Node,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { getMockMetrics } from "@/lib/mockData";
import { motion } from "framer-motion";
import { useStore } from "@/lib/store";
import { useState } from "react";
import { Modal } from "@/components/ui/Modal";
import { Badge } from "@/components/ui/Badge";

const initialNodes = [
  { 
    id: '1', 
    position: { x: 250, y: 50 }, 
    data: { label: 'Go Core Service' }, 
    className: 'glass-card border border-white/20 text-white rounded-2xl p-6 font-black shadow-2xl min-w-[200px] text-center' 
  },
  { 
    id: '2', 
    position: { x: 50, y: 250 }, 
    data: { label: 'Gitea Instance' }, 
    className: 'glass bg-blue-500/10 border border-blue-500/30 text-blue-400 rounded-2xl p-6 font-black shadow-2xl min-w-[200px] text-center' 
  },
  { 
    id: '3', 
    position: { x: 450, y: 250 }, 
    data: { label: 'Python AI Engine' }, 
    className: 'glass bg-purple-500/10 border border-purple-500/30 text-purple-400 rounded-2xl p-6 font-black shadow-2xl min-w-[200px] text-center' 
  },
  { 
    id: '4', 
    position: { x: 250, y: 450 }, 
    data: { label: 'PostgreSQL Cluster' }, 
    className: 'glass bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 rounded-2xl p-6 font-black shadow-2xl min-w-[200px] text-center' 
  },
];

const initialEdges = [
  { id: 'e1-2', source: '2', target: '1', label: 'WEBHOOK', animated: true, style: { stroke: '#3b82f6', strokeWidth: 2 } },
  { id: 'e1-3', source: '1', target: '3', label: 'gRPC', animated: true, style: { stroke: '#a855f7', strokeWidth: 2 } },
  { id: 'e1-4', source: '1', target: '4', label: 'SQL', animated: true, style: { stroke: '#10b981', strokeWidth: 2 } },
  { id: 'e3-4', source: '3', target: '4', label: 'VECTOR', animated: true, style: { stroke: '#a855f7', strokeWidth: 1, opacity: 0.5 } },
];


export default function AdminDashboard() {
  const [nodes, , onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const stats = getMockMetrics("System Admin");
  const { addToast } = useStore();

  const onConnect = (params: Connection) => setEdges((eds) => addEdge(params, eds));

  const onNodeClick = (_: React.MouseEvent, node: Node) => {
    setSelectedNode(node);
  };

  const handleAction = (action: string) => {
    addToast(`${selectedNode?.data?.label} : ${action} command sent to runner.`, "info");
    setSelectedNode(null);
  };

  return (
    <div className="space-y-10 h-full flex flex-col pb-10">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
        >
          <h1 className="text-4xl font-extrabold tracking-tight text-white mb-2">
            System <span className="text-gradient">Infrastructure</span>
          </h1>
          <p className="text-muted-foreground text-lg flex items-center gap-2">
            <Globe className="w-4 h-4 text-primary" /> Global Cluster Status • <span className="text-white font-bold uppercase tracking-widest text-xs bg-green-500/20 px-2 py-0.5 rounded border border-green-500/20">All Systems Nominal</span>
          </p>
        </motion.div>

        <motion.div 
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          className="flex items-center gap-3"
        >
          <button className="glass border-white/10 text-white px-6 py-2.5 rounded-xl text-xs font-black uppercase tracking-widest hover:bg-white/10 transition-all flex items-center gap-2">
            <Shield className="w-4 h-4 text-accent" /> Security Audit
          </button>
          <button className="glass border-white/10 text-white px-6 py-2.5 rounded-xl text-xs font-black uppercase tracking-widest hover:bg-white/10 transition-all flex items-center gap-2">
            <Settings className="w-4 h-4 text-primary" /> Config
          </button>
        </motion.div>
      </div>

      {/* Real-time Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        {stats.map((stat, i) => (
          <motion.div 
            key={i}
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card p-6 flex items-center gap-5"
          >
            <div className={cn("p-3 rounded-2xl bg-white/5 border border-white/10", stat.color)}>
              <Activity className="w-6 h-6" />
            </div>
            <div>
              <p className="text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em]">{stat.label}</p>
              <h3 className="text-2xl font-black text-white mt-1">{stat.value}</h3>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Main Visualization Container */}
      <section className="flex-1 min-h-[600px] glass rounded-3xl border border-white/10 overflow-hidden relative shadow-2xl">
        <div className="absolute top-6 left-6 z-20 flex items-center gap-4">
          <div className="glass border-white/20 px-4 py-2 rounded-2xl flex items-center gap-3">
            <div className="w-2 h-2 bg-primary rounded-full animate-pulse shadow-[0_0_10px_rgba(139,92,246,1)]" />
            <span className="text-[10px] font-black text-white uppercase tracking-widest">Live Topology Stream</span>
          </div>
          
          <div className="flex gap-2">
            {["Node View", "Edge View", "Log View"].map((tab) => (
              <button key={tab} className="glass border-white/5 px-3 py-1.5 rounded-xl text-[10px] font-bold text-muted-foreground hover:text-white transition-all">
                {tab}
              </button>
            ))}
          </div>
        </div>

        <div className="absolute inset-0 z-0 opacity-40">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={onNodeClick}
            fitView
            colorMode="dark"
          >
            <Controls className="glass border-white/10 rounded-xl overflow-hidden" />
            <MiniMap 
              nodeColor={(node) => {
                switch (node.id) {
                  case '1': return '#8b5cf6';
                  case '2': return '#3b82f6';
                  case '3': return '#a855f7';
                  case '4': return '#10b981';
                  default: return '#27272a';
                }
              }}
              maskColor="rgba(3, 0, 20, 0.8)"
              className="glass border-white/10 rounded-2xl overflow-hidden"
              style={{ background: 'transparent' }}
            />
            <Background variant={BackgroundVariant.Lines} gap={30} size={1} color="rgba(255,255,255,0.03)" />
          </ReactFlow>
        </div>

        {/* Node Detail Modal */}
        <Modal
          isOpen={!!selectedNode}
          onClose={() => setSelectedNode(null)}
          title="Service Intelligence"
          size="md"
        >
          {selectedNode && (
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <div>
                  <h4 className="text-2xl font-black text-white uppercase tracking-tighter">
                    {selectedNode.data.label}
                  </h4>
                  <p className="text-xs text-muted-foreground font-mono mt-1">ID: {selectedNode.id} • Cluster-Asia-01</p>
                </div>
                <Badge variant="success" dot>Operational</Badge>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="glass-card p-4">
                  <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">CPU Usage</p>
                  <p className="text-xl font-bold text-white">12.4%</p>
                </div>
                <div className="glass-card p-4">
                  <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mb-1">Memory</p>
                  <p className="text-xl font-bold text-white">1.2 GB</p>
                </div>
              </div>

              <div className="space-y-3">
                <h5 className="text-xs font-black uppercase tracking-widest text-muted-foreground">Admin Actions</h5>
                <div className="grid grid-cols-2 gap-3">
                  <button 
                    onClick={() => handleAction("Restart")}
                    className="py-3 rounded-xl glass border-white/10 text-xs font-bold text-white hover:bg-white/10 transition-all flex items-center justify-center gap-2"
                  >
                    <Zap className="w-4 h-4 text-primary" /> Restart
                  </button>
                  <button 
                    onClick={() => handleAction("Log Stream")}
                    className="py-3 rounded-xl glass border-white/10 text-xs font-bold text-white hover:bg-white/10 transition-all flex items-center justify-center gap-2"
                  >
                    <Terminal className="w-4 h-4 text-accent" /> Log Stream
                  </button>
                </div>
              </div>
            </div>
          )}
        </Modal>
        
        {/* Decorative corner glow */}
        <div className="absolute -bottom-20 -right-20 w-80 h-80 bg-primary/10 rounded-full blur-[120px] pointer-events-none" />
        <div className="absolute -top-20 -left-20 w-80 h-80 bg-accent/10 rounded-full blur-[120px] pointer-events-none" />
      </section>
    </div>
  );
}
