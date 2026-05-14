"use client";

import { useState, useCallback, useEffect, useRef } from 'react';
import {
  ReactFlow,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  Node,
  Edge,
  NodeChange,
  BackgroundVariant,
  Panel,
  ReactFlowProvider,
  useReactFlow,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { identityService, OrgNode as OrgNodeModel, OrgUnit } from '@/lib/services/identity.service';
import { Plus, Save, ZoomIn, Building2, LayoutTemplate } from 'lucide-react';
import { useStore } from '@/lib/store';
import { OrgNode } from './OrgNode';
import dagre from 'dagre';

const nodeTypes = {
  org: OrgNode,
};

type OrgTreeNodeData = {
  label?: string;
  type?: string;
  direct_count?: number;
  total_count?: number;
};

const edgeStyle = {
  strokeDasharray: '0',
  stroke: 'var(--border)',
  strokeWidth: 2,
};

const getLayoutedElements = (nodes: Node[], edges: Edge[]) => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  dagreGraph.setGraph({ rankdir: 'TB', nodesep: 70, ranksep: 100 });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: 240, height: 130 });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - 120,
        y: nodeWithPosition.y - 65,
      },
    };
  });

  return { nodes: layoutedNodes, edges };
};

function OrgTreeContent() {
  const [allNodes, setAllNodes] = useState<Node[]>([]);
  const [allEdges, setAllEdges] = useState<Edge[]>([]);
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [maxDepth, setMaxDepth] = useState(4);
  const [selectedRoot, setSelectedRoot] = useState<string>('all');
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  
  const addToast = useStore(state => state.addToast);
  const { fitView } = useReactFlow();

  const recalculateMemberCounts = useCallback((nodes: Node[], edges: Edge[]) => {
    const nodeMap = new Map(nodes.map(n => [n.id, { ...n, data: { ...n.data } }]));
    
    const calculateTotal = (id: string): number => {
      const node = nodeMap.get(id);
      if (!node) return 0;
      
      const childEdges = edges.filter(e => e.source === id);
      const childrenTotal = childEdges.reduce((sum, edge) => sum + calculateTotal(edge.target), 0);
      
      const nodeData = node.data as OrgTreeNodeData;
      const total = (nodeData.direct_count || 0) + childrenTotal;
      node.data = { ...node.data, total_count: total };
      return total;
    };

    const roots = nodes.filter(n => !edges.some(e => e.target === n.id));
    roots.forEach(r => calculateTotal(r.id));

    return Array.from(nodeMap.values());
  }, []);

  const onToggleExpand = useCallback((id: string) => {
    setExpandedNodes((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  // Optimized Filter logic
  useEffect(() => {
    // 1. Root selection
    let nodesInScope = allNodes;
    let edgesInScope = allEdges;

    if (selectedRoot !== 'all') {
      const getDescendants = (id: string, currentEdges: Edge[]): string[] => {
        const children = currentEdges.filter(e => e.source === id).map(e => e.target);
        return [id, ...children.flatMap(childId => getDescendants(childId, currentEdges))];
      };
      const allowedIds = getDescendants(selectedRoot, allEdges);
      nodesInScope = allNodes.filter(n => allowedIds.includes(n.id));
      edgesInScope = allEdges.filter(e => allowedIds.includes(e.source) && allowedIds.includes(e.target));
    }

    // 2. Depth calculation
    const nodeDepths = new Map<string, number>();
    const rootIds = nodesInScope.filter(n => !edgesInScope.some(e => e.target === n.id)).map(n => n.id);
    const assignDepth = (id: string, depth: number) => {
      nodeDepths.set(id, depth);
      edgesInScope.filter(e => e.source === id).forEach(e => assignDepth(e.target, depth + 1));
    };
    rootIds.forEach(id => assignDepth(id, 1));

    // 3. Visibility logic (Respect Expansion & MaxDepth)
    const visibleNodeIds = new Set<string>();
    const traverse = (id: string, parentVisible: boolean) => {
      const depth = nodeDepths.get(id) || 1;
      const isVisible = parentVisible && depth <= maxDepth;
      
      if (isVisible) {
        visibleNodeIds.add(id);
        const isExpanded = expandedNodes.has(id);
        edgesInScope.filter(e => e.source === id).forEach(e => traverse(e.target, isExpanded));
      }
    };
    rootIds.forEach(id => traverse(id, true));

    // 4. Final Elements with decorated data
    const finalNodes = nodesInScope
      .filter(n => visibleNodeIds.has(n.id))
      .map(n => ({
        ...n,
        data: {
          ...n.data,
          isExpanded: expandedNodes.has(n.id),
          hasChildren: allEdges.some(e => e.source === n.id),
          onToggleExpand
        }
      }));

    const finalEdges = edgesInScope.filter(e => visibleNodeIds.has(e.source) && visibleNodeIds.has(e.target));

    setNodes(finalNodes);
    setEdges(finalEdges);
  }, [allNodes, allEdges, maxDepth, selectedRoot, expandedNodes, setNodes, setEdges, onToggleExpand]);

  const onUpdateNode = useCallback(async (id: string, newData: Partial<{ label: string; type: string; leader_id: string; isInitialEditing: boolean }>) => {
    try {
      // 1. Persist to backend if not a temporary initial edit
      if (!newData.isInitialEditing) {
        await identityService.updateUnit(id, {
          label: newData.label,
          unit_type: newData.type as OrgUnit["unit_type"],
          leader_user_id: newData.leader_id
        });
      }

      // 2. Update local state
      setAllNodes((nds) => {
        const updatedNodes = nds.map((node) => 
          node.id === id ? { ...node, data: { ...node.data, ...newData } } : node
        );
        return recalculateMemberCounts(updatedNodes, allEdges);
      });
      
      if (!newData.isInitialEditing) {
        addToast("Organization unit updated", "success");
      }
    } catch (error) {
      console.error("[OrgTree] Failed to update unit:", error);
      addToast("Failed to update organization unit", "error");
    }
  }, [allEdges, addToast, recalculateMemberCounts]);


  const onDeleteNode = useCallback(async (id: string) => {
    try {
      // 1. Persist deletion to backend (except for unsaved new nodes)
      if (!id.startsWith('temp-')) {
        await identityService.deleteUnit(id);
      }

      // 2. Update local state
      setAllNodes((nds) => {
        const filteredNodes = nds.filter((node) => node.id !== id);
        const filteredEdges = allEdges.filter((edge) => edge.source !== id && edge.target !== id);
        return recalculateMemberCounts(filteredNodes, filteredEdges);
      });
      setAllEdges((eds) => eds.filter((edge) => edge.source !== id && edge.target !== id));
      
      setExpandedNodes(prev => {
        const next = new Set(prev);
        next.delete(id);
        return next;
      });

      addToast("Organizational unit removed", "warning");
    } catch (error) {
      console.error("[OrgTree] Failed to delete unit:", error);
      addToast("Failed to remove organizational unit", "error");
    }
  }, [allEdges, addToast, recalculateMemberCounts]);

  const nodesRef = useRef(allNodes);
  const addChildRef = useRef<(parentId: string) => void>(() => {});
  useEffect(() => {
    nodesRef.current = allNodes;
  }, [allNodes]);

  const onAddChild = useCallback(async (parentId: string) => {
    const parentNode = nodesRef.current.find(n => n.id === parentId);
    if (!parentNode) return;

    const typeHierarchy = ['division', 'team', 'group', 'part'];
    const parentIdx = typeHierarchy.indexOf(parentNode.data.type as string);
    const allowedTypes = typeHierarchy.slice(parentIdx + 1);
    if (allowedTypes.length === 0) allowedTypes.push('part');
    const nextType = allowedTypes[0] as OrgUnit["unit_type"];

    try {
      // 1. Create on backend first to get a real ID
      addToast(`Adding new ${nextType}...`, "info");
      
      const newUnit = await identityService.createUnit({
        unit_id: `unit-${Date.now()}`, // Backend usually generates this, but we follow contract PR-B1
        parent_unit_id: parentId,
        unit_type: nextType,
        label: `New ${nextType}`,
        position_x: parentNode.position.x,
        position_y: parentNode.position.y + 150
      });

      const newNode: Node = {
        id: newUnit.unit_id,
        type: 'org',
        data: { 
          label: newUnit.label, 
          type: newUnit.unit_type,
          allowedTypes,
          isInitialEditing: true,
          direct_count: 0,
          total_count: 0,
          onAddChild: (id: string) => addChildRef.current(id),
          onDelete: onDeleteNode,
          onUpdate: onUpdateNode,
          onToggleExpand
        },
        position: { x: newUnit.position_x, y: newUnit.position_y },
      };

      const newEdge: Edge = {
        id: `e-${parentId}-${newUnit.unit_id}`,
        source: parentId,
        target: newUnit.unit_id,
      };

      setAllNodes((nds) => {
        const newNodes = nds.concat(newNode);
        const newEdges = allEdges.concat(newEdge);
        return recalculateMemberCounts(newNodes, newEdges);
      });
      setAllEdges((eds) => eds.concat(newEdge));
      
      window.requestAnimationFrame(() => fitView({ duration: 800 }));
    } catch (error) {
      console.error("[OrgTree] Failed to create child unit:", error);
      addToast("Failed to create new organizational unit", "error");
    }
  }, [allEdges, addToast, onDeleteNode, onUpdateNode, onToggleExpand, fitView, recalculateMemberCounts]);

  useEffect(() => {
    addChildRef.current = onAddChild;
  }, [onAddChild]);

  // Initial fetch only
  useEffect(() => {
    let isMounted = true;
    const fetchData = async () => {
      try {
        const data = await identityService.getOrgHierarchy();
        if (!isMounted) return;

        const processedNodes: Node[] = data.nodes.map(node => ({
          ...node,
          type: 'org',
          data: {
            ...node.data,
            onAddChild,
            onDelete: onDeleteNode,
            onUpdate: onUpdateNode
          }
        }));


        const processedEdges = data.edges.map(edge => ({
          ...edge,
          style: edgeStyle,
        }));

        const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
          processedNodes,
          processedEdges
        );

        const calculatedNodes = recalculateMemberCounts(layoutedNodes, layoutedEdges);
        setAllNodes(calculatedNodes);
        setAllEdges(layoutedEdges);
        // Default to expanded tree so the chart matches the list view density.
        setExpandedNodes(new Set(calculatedNodes.map((n) => n.id)));
      } catch (error) {
        console.error("Failed to load org hierarchy, using enhanced mock:", error);
        // Fallback with enhanced mock nodes
        const mock = identityService.mockHierarchy();

        const enhancedNodes: Node[] = mock.nodes.map((node: { id: string; type?: string; data: Record<string, unknown>; position: { x: number; y: number } }) => ({
          ...node,
          type: 'org',
          data: {
            ...node.data,
            onAddChild,
            onDelete: onDeleteNode,
            onUpdate: onUpdateNode
          }
        }));

        const enhancedEdges = mock.edges.map((edge: { id: string; source: string; target: string; animated?: boolean }) => ({
          ...edge,
          style: edgeStyle,
        }));
        
        const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
          enhancedNodes,
          enhancedEdges
        );
        const calculatedNodes = recalculateMemberCounts(layoutedNodes, layoutedEdges);
        setAllNodes(calculatedNodes);
        setAllEdges(layoutedEdges);
        setExpandedNodes(new Set(calculatedNodes.map((n) => n.id)));
      } finally {
        if (isMounted) setIsLoading(false);
      }
    };
    fetchData();
    return () => { isMounted = false; };
  }, [onAddChild, onDeleteNode, onUpdateNode, onToggleExpand, recalculateMemberCounts]);

  const onLayout = useCallback(() => {
    const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(allNodes, allEdges);
    setAllNodes([...layoutedNodes]);
    setAllEdges([...layoutedEdges]);
    window.requestAnimationFrame(() => fitView({ duration: 800 }));
    addToast("Hierarchy layout optimized", "info");
  }, [allNodes, allEdges, fitView, addToast]);

  const onConnect = useCallback(
    (params: Connection) => setAllEdges((eds) => addEdge(params, eds)),
    []
  );

  // Mirror React Flow's internal NodeChange position updates back into the
  // master `allNodes` list. Without this, dragging a node only updates the
  // view-state copy and Save would persist the unchanged source coordinates.
  const handleNodesChange = useCallback(
    (changes: NodeChange<Node>[]) => {
      onNodesChange(changes);
      const positionUpdates = new Map<string, { x: number; y: number }>();
      for (const change of changes) {
        if (change.type === 'position' && change.position) {
          positionUpdates.set(change.id, change.position);
        }
      }
      if (positionUpdates.size === 0) return;
      setAllNodes((prev) =>
        prev.map((node) => {
          const next = positionUpdates.get(node.id);
          if (!next) return node;
          if (next.x === node.position.x && next.y === node.position.y) return node;
          return { ...node, position: next };
        }),
      );
    },
    [onNodesChange],
  );

  const addRootNode = () => {
    const id = `node-${Date.now()}`;
    const newNode: Node = {
      id,
      type: 'org',
      data: { 
        label: 'New Division', 
        type: 'division',
        allowedTypes: ['division', 'team', 'group', 'part'],
        isInitialEditing: true,
        direct_count: 0,
        total_count: 0,
        onAddChild,
        onDelete: onDeleteNode,
        onUpdate: onUpdateNode,
        onToggleExpand
      },
      position: { x: 400, y: 0 },
    };
    setAllNodes((nds) => nds.concat(newNode));
    setExpandedNodes((prev) => {
      const next = new Set(prev);
      next.add(id);
      return next;
    });
    addToast("New root-level division added", "success");
  };

  if (isLoading) return (
    <div className="flex flex-col items-center justify-center h-[600px] glass rounded-3xl animate-pulse">
      <Building2 className="w-12 h-12 text-accent/20 mb-4" />
      <p className="text-xs font-black uppercase tracking-[0.3em] text-muted-foreground">Rendering Hierarchy...</p>
    </div>
  );

  const defaultEdgeOptions = {
    animated: false,
    style: edgeStyle,
  };

  return (
    <div className="h-[750px] glass rounded-3xl border border-border overflow-hidden relative shadow-2xl">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodesChange={handleNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        defaultEdgeOptions={defaultEdgeOptions}
        fitView
      >
        <Background variant={BackgroundVariant.Lines} gap={40} size={1} color="rgba(255,255,255,0.03)" />
        <Controls className="glass border-border rounded-xl overflow-hidden" />
        
        <Panel position="top-left" className="flex flex-col gap-4">
          <div className="glass border-border/50 p-4 rounded-2xl flex flex-col gap-3 min-w-[200px]">
            <p className="text-[10px] font-black text-primary uppercase tracking-widest">Scope Filter</p>
            
            <div className="flex flex-col gap-2">
              <label className="text-[9px] font-bold text-muted-foreground uppercase">Root Node</label>
              <select 
                value={selectedRoot}
                onChange={(e) => setSelectedRoot(e.target.value)}
                className="themed-select"
              >
                <option value="all">Show All</option>
                {allNodes.map(n => {
                  const nodeData = n.data as OrgTreeNodeData;
                  return (
                    <option key={n.id} value={n.id}>{nodeData.label}</option>
                  );
                })}
              </select>
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-[9px] font-bold text-muted-foreground uppercase">Jump to Unit</label>
              <button
                onClick={() => {
                  if (selectedRoot === 'all') return;
                  const target = nodes.find(n => n.id === selectedRoot);
                  if (target) {
                    fitView({ nodes: [target], duration: 800, padding: 0.5 });
                  }
                }}
                className="bg-primary/10 border border-primary/20 rounded-lg px-2 py-1.5 text-[10px] font-black uppercase text-primary hover:bg-primary/20 transition-all text-left"
              >
                Focus Selection
              </button>
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-[9px] font-bold text-muted-foreground uppercase flex justify-between">
                <span>Max Depth</span>
                <span className="text-primary">{maxDepth}</span>
              </label>
              <input 
                type="range"
                min="1"
                max="5"
                step="1"
                value={maxDepth}
                onChange={(e) => setMaxDepth(parseInt(e.target.value))}
                className="accent-primary"
              />
            </div>
          </div>
        </Panel>

        <Panel position="top-right" className="flex gap-2">
          <button 
            onClick={onLayout}
            className="glass border-border text-foreground px-4 py-2 rounded-xl text-[10px] font-black uppercase tracking-widest hover:bg-muted/40 transition-all flex items-center gap-2"
          >
            <LayoutTemplate className="w-3 h-3 text-emerald-400" /> Auto Layout
          </button>
          <button 
            onClick={addRootNode}
            className="glass border-border bg-primary/20 text-foreground px-4 py-2 rounded-xl text-[10px] font-black uppercase tracking-widest hover:bg-primary/30 transition-all flex items-center gap-2"
          >
            <Plus className="w-3 h-3" /> Add Division
          </button>
          <button 
            onClick={async () => {
              try {
                // Map React Flow nodes back to OrgNode domain model
                const orgNodes: OrgNodeModel[] = allNodes.map(n => ({
                  id: n.id,
                  position: n.position,
                  data: {
                    label: n.data.label as string,
                    type: n.data.type as string,
                    direct_count: n.data.direct_count as number,
                    total_count: n.data.total_count as number
                  }
                }));
                const orgEdges = allEdges.map(e => ({
                  source: e.source,
                  target: e.target
                }));

                await identityService.updateOrgHierarchy(orgNodes, orgEdges);

                addToast("Hierarchy configuration saved", "success");
              } catch (error) {
                console.error("[OrgTree] Failed to save hierarchy:", error);
                addToast("Failed to save hierarchy changes", "error");
              }
            }}
            className="glass border-border text-foreground px-4 py-2 rounded-xl text-[10px] font-black uppercase tracking-widest hover:bg-muted/40 transition-all flex items-center gap-2"
          >
            <Save className="w-3 h-3 text-accent" /> Save
          </button>
        </Panel>

        <Panel position="bottom-left" className="glass border-border px-4 py-2 rounded-2xl">
          <p className="text-[10px] font-black text-muted-foreground uppercase tracking-widest flex items-center gap-2">
            <ZoomIn className="w-3 h-3" /> Filtered Scope • Hover card for actions
          </p>
        </Panel>
      </ReactFlow>
    </div>
  );
}

export function OrgTree() {
  return (
    <ReactFlowProvider>
      <OrgTreeContent />
    </ReactFlowProvider>
  );
}
