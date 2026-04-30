"use client";

import { Activity, Database, Server, Settings } from "lucide-react";
import { useCallback } from 'react';
import {
  ReactFlow,
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
  BackgroundVariant,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

const initialNodes = [
  { id: '1', position: { x: 250, y: 50 }, data: { label: 'Go Core Service' }, className: 'bg-card border border-border text-foreground rounded-lg p-4 font-semibold shadow-lg min-w-[150px] text-center' },
  { id: '2', position: { x: 100, y: 200 }, data: { label: 'Gitea Server' }, className: 'bg-indigo-500/10 border border-indigo-500/50 text-indigo-400 rounded-lg p-4 font-semibold shadow-lg min-w-[150px] text-center' },
  { id: '3', position: { x: 400, y: 200 }, data: { label: 'Python AI Module' }, className: 'bg-purple-500/10 border border-purple-500/50 text-purple-400 rounded-lg p-4 font-semibold shadow-lg min-w-[150px] text-center' },
  { id: '4', position: { x: 250, y: 350 }, data: { label: 'PostgreSQL' }, className: 'bg-emerald-500/10 border border-emerald-500/50 text-emerald-400 rounded-lg p-4 font-semibold shadow-lg min-w-[150px] text-center' },
];

const initialEdges = [
  { id: 'e1-2', source: '2', target: '1', label: 'Webhook', animated: true, style: { stroke: '#6366f1' } },
  { id: 'e1-3', source: '1', target: '3', label: 'gRPC', animated: true, style: { stroke: '#a855f7' } },
  { id: 'e1-4', source: '1', target: '4', label: 'SQL', animated: false },
  { id: 'e3-4', source: '3', target: '4', label: 'Logs', animated: true },
];

export default function AdminDashboard() {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  const onConnect = useCallback(
    (params: any) => setEdges((eds) => addEdge(params, eds)),
    [setEdges],
  );

  return (
    <div className="space-y-8 animate-in fade-in duration-500 h-full flex flex-col">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">System Infrastructure</h1>
          <p className="text-muted-foreground mt-1">Real-time topology and runner health status.</p>
        </div>
        <div className="flex items-center gap-3">
          <button className="bg-background border border-border text-foreground px-4 py-2 rounded-lg text-sm font-medium hover:bg-accent transition-colors shadow-sm flex items-center gap-2">
            <Settings className="w-4 h-4" /> Global Settings
          </button>
        </div>
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[
          { label: "Core Service", value: "Healthy", icon: Activity, color: "text-emerald-400" },
          { label: "Active Runners", value: "8 / 10", icon: Server, color: "text-indigo-400" },
          { label: "DB Latency", value: "14ms", icon: Database, color: "text-purple-400" },
        ].map((stat, i) => (
          <div key={i} className="bg-card rounded-xl border border-border p-5 shadow-sm flex items-center gap-4">
            <div className={`p-3 rounded-lg bg-background border border-border ${stat.color}`}>
              <stat.icon className="w-5 h-5" />
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">{stat.label}</p>
              <h3 className="text-xl font-bold mt-0.5">{stat.value}</h3>
            </div>
          </div>
        ))}
      </div>

      {/* React Flow Container */}
      <div className="bg-card border border-border rounded-xl shadow-sm flex-1 min-h-[500px] relative overflow-hidden">
        <div className="absolute top-4 left-4 z-10 bg-background/80 backdrop-blur border border-border px-3 py-1.5 rounded-md text-sm font-medium">
          Live Topology
        </div>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          fitView
          className="bg-zinc-950"
        >
          <Controls className="bg-background border-border fill-foreground" />
          <MiniMap 
            nodeColor={(node) => {
              switch (node.id) {
                case '1': return '#fafafa';
                case '2': return '#6366f1';
                case '3': return '#a855f7';
                case '4': return '#10b981';
                default: return '#27272a';
              }
            }}
            maskColor="rgba(0,0,0,0.7)"
            className="bg-card border-border"
          />
          <Background variant={BackgroundVariant.Dots} gap={12} size={1} color="#27272a" />
        </ReactFlow>
      </div>
    </div>
  );
}
