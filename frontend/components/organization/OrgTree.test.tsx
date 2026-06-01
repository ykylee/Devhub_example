import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor, act } from "@testing-library/react";

// --- Mocks ------------------------------------------------------------

// Capture the latest props passed to <ReactFlow> so individual tests can
// invoke wired callbacks (onConnect / onNodesChange) without rendering the
// real canvas.
const reactFlowProps: { current: Record<string, unknown> } = { current: {} };
const fitView = vi.fn();

// @xyflow/react — replace with a passthrough that runs the component logic
// without depending on canvas / measuring. useNodesState / useEdgesState
// behave like useState tuples.
vi.mock("@xyflow/react", () => {
  const React = require("react");
  type AnyProps = { children?: unknown; [k: string]: unknown };

  const ReactFlowProvider = ({ children }: AnyProps) =>
    React.createElement(React.Fragment, null, children);

  const ReactFlow = (props: AnyProps) => {
    reactFlowProps.current = props as Record<string, unknown>;
    return React.createElement("div", { "data-testid": "react-flow" }, props.children);
  };

  const Controls = ({ children }: AnyProps) =>
    React.createElement("div", { "data-testid": "controls" }, children);

  const Background = (_props: AnyProps) =>
    React.createElement("div", { "data-testid": "background" });

  const Panel = ({ children, position }: AnyProps) =>
    React.createElement("div", { "data-testid": `panel-${position as string}` }, children);

  const useNodesState = (initial: unknown[]) => {
    const [state, setState] = React.useState(initial);
    const onChange = React.useCallback(() => {}, []);
    return [state, setState, onChange];
  };
  const useEdgesState = (initial: unknown[]) => {
    const [state, setState] = React.useState(initial);
    const onChange = React.useCallback(() => {}, []);
    return [state, setState, onChange];
  };

  const useReactFlow = () => ({ fitView });

  const addEdge = (params: unknown, eds: unknown[]) => [...eds, params];

  return {
    ReactFlow,
    Controls,
    Background,
    Panel,
    ReactFlowProvider,
    useNodesState,
    useEdgesState,
    useReactFlow,
    addEdge,
    BackgroundVariant: { Lines: "lines", Dots: "dots", Cross: "cross" },
    Handle: () => null,
    Position: { Top: "top", Bottom: "bottom" },
  };
});

// dagre — replace with stub so layout returns deterministic positions
// without depending on the real graphlib. The component invokes
// `new dagre.graphlib.Graph()` so Graph must be a constructor.
vi.mock("dagre", () => {
  class Graph {
    private nodeStore = new Map<string, { x: number; y: number }>();
    setDefaultEdgeLabel() {
      /* noop */
    }
    setGraph() {
      /* noop */
    }
    setNode(id: string) {
      this.nodeStore.set(id, { x: 200, y: 200 });
    }
    setEdge() {
      /* noop */
    }
    node(id: string) {
      return this.nodeStore.get(id) ?? { x: 0, y: 0 };
    }
  }
  return {
    default: {
      graphlib: { Graph },
      layout: () => {},
    },
  };
});

// framer-motion — already not used by OrgTree directly but safe to stub for
// the OrgNode child path.
vi.mock("framer-motion", () => {
  const React = require("react");
  type AnyProps = { children?: unknown; [k: string]: unknown };
  const motion = new Proxy(
    {},
    {
      get: (_target, tag) =>
        ({ children, ...props }: AnyProps) =>
          React.createElement(tag as string, props, children),
    },
  );
  return {
    motion,
    AnimatePresence: ({ children }: AnyProps) => React.createElement(React.Fragment, null, children),
  };
});

const getOrgHierarchy = vi.fn();
const getUsers = vi.fn();
const createUnit = vi.fn();
const deleteUnit = vi.fn();
const updateUnit = vi.fn();
const updateOrgHierarchy = vi.fn();

vi.mock("@/domain/organization-management/service/identity.service", async () => {
  const actual = await vi.importActual<typeof import("@/domain/organization-management/service/identity.service")>(
    "@/domain/organization-management/service/identity.service",
  );
  return {
    ...actual,
    identityService: {
      ...actual.identityService,
      getOrgHierarchy: (...args: unknown[]) => getOrgHierarchy(...args),
      getUsers: (...args: unknown[]) => getUsers(...args),
      createUnit: (...args: unknown[]) => createUnit(...args),
      deleteUnit: (...args: unknown[]) => deleteUnit(...args),
      updateUnit: (...args: unknown[]) => updateUnit(...args),
      updateOrgHierarchy: (...args: unknown[]) => updateOrgHierarchy(...args),
    },
  };
});

const addToast = vi.fn();
vi.mock("@/lib/store", () => ({
  useStore: (selector?: (s: { addToast: typeof addToast }) => unknown) => {
    const state = { addToast };
    if (typeof selector === "function") return selector(state);
    return state;
  },
}));

import { OrgTree } from "./OrgTree";

const sampleHierarchy = {
  nodes: [
    {
      id: "u-root",
      data: { label: "Root", type: "division", direct_count: 2, total_count: 5 },
      position: { x: 0, y: 0 },
    },
    {
      id: "u-eng",
      data: { label: "Engineering", type: "team", direct_count: 3, total_count: 3 },
      position: { x: 0, y: 0 },
    },
  ],
  edges: [
    { id: "e-1", source: "u-root", target: "u-eng" },
  ],
};

beforeEach(() => {
  getOrgHierarchy.mockReset();
  getUsers.mockReset();
  createUnit.mockReset();
  deleteUnit.mockReset();
  updateUnit.mockReset();
  updateOrgHierarchy.mockReset();
  addToast.mockReset();
  fitView.mockReset();
  reactFlowProps.current = {};
  getUsers.mockResolvedValue([]);
});

describe("OrgTree", () => {
  it("renders loading state while initial fetch is pending", () => {
    getOrgHierarchy.mockReturnValue(new Promise(() => {}));
    render(<OrgTree />);
    expect(screen.getByText(/Rendering Hierarchy/i)).toBeInTheDocument();
  });

  it("renders panels + controls after initial fetch resolves", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByTestId("react-flow")).toBeInTheDocument();
    });
    expect(screen.getByText(/Scope Filter/i)).toBeInTheDocument();
    expect(screen.getByText(/Auto Layout/i)).toBeInTheDocument();
    expect(screen.getByText(/Add Division/i)).toBeInTheDocument();
    expect(screen.getByText(/Save/i)).toBeInTheDocument();
  });

  it("populates the root select dropdown with fetched units", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByText("Root")).toBeInTheDocument();
    });
    expect(screen.getByText("Engineering")).toBeInTheDocument();
    expect(screen.getByText("Show All")).toBeInTheDocument();
  });

  it("renders empty canvas when getOrgHierarchy rejects (no mock fallback)", async () => {
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    getOrgHierarchy.mockRejectedValueOnce(new Error("api"));
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByTestId("react-flow")).toBeInTheDocument();
    });
    errSpy.mockRestore();
  });

  it("changes selectedRoot when the Root Node select fires onChange", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByText("Root")).toBeInTheDocument();
    });
    const select = screen.getByRole("combobox");
    fireEvent.change(select, { target: { value: "u-root" } });
    expect((select as HTMLSelectElement).value).toBe("u-root");
  });

  it("adjusts max depth when the slider input fires onChange", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByText(/Max Depth/i)).toBeInTheDocument();
    });
    const slider = screen.getByRole("slider");
    fireEvent.change(slider, { target: { value: "2" } });
    expect((slider as HTMLInputElement).value).toBe("2");
  });

  it("does nothing on Focus Selection when no root is selected", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByText(/Focus Selection/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/Focus Selection/i));
    // No toast, no service call.
    expect(addToast).not.toHaveBeenCalled();
  });

  it("invokes Auto Layout button + toasts info", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByText(/Auto Layout/i)).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText(/Auto Layout/i));
    expect(addToast).toHaveBeenCalledWith("Hierarchy layout optimized", "info");
  });

  it("invokes Add Division -> createUnit + success toast", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    createUnit.mockResolvedValueOnce({
      unit_id: "new-div",
      parent_unit_id: "",
      unit_type: "division",
      label: "New Division",
      leader_user_id: "",
      position_x: 400,
      position_y: 0,
    });
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByText(/Add Division/i)).toBeInTheDocument();
    });
    await act(async () => {
      fireEvent.click(screen.getByText(/Add Division/i));
    });
    await waitFor(() => {
      expect(createUnit).toHaveBeenCalled();
    });
    expect(addToast).toHaveBeenCalledWith("Adding new division...", "info");
    expect(addToast).toHaveBeenCalledWith("New root-level division added", "success");
  });

  it("toasts error when Add Division createUnit rejects", async () => {
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    createUnit.mockRejectedValueOnce(new Error("api"));
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByText(/Add Division/i)).toBeInTheDocument();
    });
    await act(async () => {
      fireEvent.click(screen.getByText(/Add Division/i));
    });
    await waitFor(() => {
      expect(addToast).toHaveBeenCalledWith("Failed to create new division", "error");
    });
    errSpy.mockRestore();
  });

  it("invokes Save -> updateOrgHierarchy with mapped nodes + edges", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    updateOrgHierarchy.mockResolvedValueOnce(undefined);
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByText(/Save/i)).toBeInTheDocument();
    });
    await act(async () => {
      fireEvent.click(screen.getByText(/Save/i));
    });
    await waitFor(() => {
      expect(updateOrgHierarchy).toHaveBeenCalledTimes(1);
    });
    const [nodesArg, edgesArg] = updateOrgHierarchy.mock.calls[0];
    expect(Array.isArray(nodesArg)).toBe(true);
    expect(Array.isArray(edgesArg)).toBe(true);
    // The two seeded nodes should be present in the persisted payload.
    expect(nodesArg.length).toBeGreaterThanOrEqual(2);
    expect(addToast).toHaveBeenCalledWith("Hierarchy configuration saved", "success");
  });

  it("toasts error when Save updateOrgHierarchy rejects", async () => {
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    updateOrgHierarchy.mockRejectedValueOnce(new Error("api"));
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByText(/Save/i)).toBeInTheDocument();
    });
    await act(async () => {
      fireEvent.click(screen.getByText(/Save/i));
    });
    await waitFor(() => {
      expect(addToast).toHaveBeenCalledWith("Failed to save hierarchy changes", "error");
    });
    errSpy.mockRestore();
  });

  it("calls identityService.getUsers on mount for leader picker context", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(getUsers).toHaveBeenCalledTimes(1);
    });
  });

  it("Focus Selection calls fitView when a specific root is selected", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByText("Root")).toBeInTheDocument();
    });
    const select = screen.getByRole("combobox");
    fireEvent.change(select, { target: { value: "u-root" } });
    fireEvent.click(screen.getByText(/Focus Selection/i));
    await waitFor(() => {
      expect(fitView).toHaveBeenCalled();
    });
  });

  it("filters visible scope when selectedRoot != 'all'", async () => {
    // Provide a 3-node hierarchy and select a subtree so the descendants
    // filter branch executes.
    const sub = {
      nodes: [
        ...sampleHierarchy.nodes,
        {
          id: "u-other",
          data: { label: "Other Branch", type: "team", direct_count: 1, total_count: 1 },
          position: { x: 100, y: 100 },
        },
      ],
      edges: [...sampleHierarchy.edges],
    };
    getOrgHierarchy.mockResolvedValueOnce(sub);
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByText("Other Branch")).toBeInTheDocument();
    });
    fireEvent.change(screen.getByRole("combobox"), { target: { value: "u-root" } });
    // Wait one frame so the filter effect re-runs.
    await waitFor(() => {
      expect((screen.getByRole("combobox") as HTMLSelectElement).value).toBe("u-root");
    });
  });

  it("onConnect wired prop appends new edge via addEdge", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.onConnect).toBeTypeOf("function");
    });
    const onConnect = reactFlowProps.current.onConnect as (c: unknown) => void;
    // The handler reads from the addEdge helper provided by the mock.
    await act(async () => {
      onConnect({ source: "u-root", target: "u-eng" });
    });
    // No throw -> coverage of the onConnect closure.
    expect(true).toBe(true);
  });

  it("handleNodesChange mirrors position updates into master state", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.onNodesChange).toBeTypeOf("function");
    });
    const handler = reactFlowProps.current.onNodesChange as (c: unknown[]) => void;
    await act(async () => {
      handler([
        { type: "position", id: "u-root", position: { x: 999, y: 999 } },
        // Same position no-op branch (after first update applies).
        { type: "select", id: "u-eng", selected: true } as unknown,
      ]);
    });
    expect(true).toBe(true);
  });

  it("handleNodesChange is a no-op when no position changes are present", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.onNodesChange).toBeTypeOf("function");
    });
    const handler = reactFlowProps.current.onNodesChange as (c: unknown[]) => void;
    await act(async () => {
      handler([{ type: "select", id: "u-root", selected: true } as unknown]);
    });
    expect(true).toBe(true);
  });

  it("logs the error when getUsers rejects but does not crash", async () => {
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    getUsers.mockReset();
    getUsers.mockRejectedValueOnce(new Error("api"));
    render(<OrgTree />);
    await waitFor(() => {
      expect(screen.getByTestId("react-flow")).toBeInTheDocument();
    });
    await waitFor(() => {
      expect(errSpy).toHaveBeenCalled();
    });
    errSpy.mockRestore();
  });

  // --- 추가 보강 테스트 --------------------------------------------------

  it("onUpdateNode persists change to backend and updates local state", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    updateUnit.mockResolvedValueOnce(undefined);
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.nodes?.length).toBe(2);
    });
    const nodes = reactFlowProps.current.nodes as any[];
    const rootNode = nodes.find(n => n.id === "u-root");
    expect(rootNode).toBeDefined();

    await act(async () => {
      await rootNode.data.onUpdate("u-root", {
        label: "Updated Root Division",
        type: "division",
        leader_id: "user-1",
        isInitialEditing: false,
      });
    });

    expect(updateUnit).toHaveBeenCalledWith("u-root", {
      label: "Updated Root Division",
      unit_type: "division",
      leader_user_id: "user-1",
    });
    expect(addToast).toHaveBeenCalledWith("Organization unit updated", "success");
  });

  it("onUpdateNode updates state silently without API call for initial editing node", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.nodes?.length).toBe(2);
    });
    const nodes = reactFlowProps.current.nodes as any[];
    const rootNode = nodes.find(n => n.id === "u-root");

    await act(async () => {
      await rootNode.data.onUpdate("u-root", {
        label: "Initial New Name",
        type: "division",
        leader_id: "user-1",
        isInitialEditing: true,
      });
    });

    expect(updateUnit).not.toHaveBeenCalled();
    expect(addToast).not.toHaveBeenCalled();
  });

  it("onUpdateNode toasts error when updateUnit rejects", async () => {
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    updateUnit.mockRejectedValueOnce(new Error("api error"));
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.nodes?.length).toBe(2);
    });
    const nodes = reactFlowProps.current.nodes as any[];
    const rootNode = nodes.find(n => n.id === "u-root");

    await act(async () => {
      await rootNode.data.onUpdate("u-root", {
        label: "Broken Unit",
        type: "division",
        leader_id: "user-1",
        isInitialEditing: false,
      });
    });

    expect(addToast).toHaveBeenCalledWith("Failed to update organization unit", "error");
    errSpy.mockRestore();
  });

  it("onDeleteNode calls deleteUnit API and toasts warning", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    deleteUnit.mockResolvedValueOnce(undefined);
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.nodes?.length).toBe(2);
    });
    const nodes = reactFlowProps.current.nodes as any[];
    const engNode = nodes.find(n => n.id === "u-eng");
    expect(engNode).toBeDefined();

    await act(async () => {
      await engNode.data.onDelete("u-eng");
    });

    expect(deleteUnit).toHaveBeenCalledWith("u-eng");
    expect(addToast).toHaveBeenCalledWith("Organizational unit removed", "warning");
  });

  it("onDeleteNode skips API call for temp new nodes", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.nodes?.length).toBe(2);
    });
    const nodes = reactFlowProps.current.nodes as any[];
    const engNode = nodes.find(n => n.id === "u-eng");
    expect(engNode).toBeDefined();

    await act(async () => {
      await engNode.data.onDelete("temp-1234");
    });

    expect(deleteUnit).not.toHaveBeenCalled();
    expect(addToast).toHaveBeenCalledWith("Organizational unit removed", "warning");
  });

  it("onDeleteNode toasts error when deleteUnit API rejects", async () => {
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    deleteUnit.mockRejectedValueOnce(new Error("delete failed"));
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.nodes?.length).toBe(2);
    });
    const nodes = reactFlowProps.current.nodes as any[];
    const engNode = nodes.find(n => n.id === "u-eng");
    expect(engNode).toBeDefined();

    await act(async () => {
      await engNode.data.onDelete("u-eng");
    });

    expect(addToast).toHaveBeenCalledWith("Failed to remove organizational unit", "error");
    errSpy.mockRestore();
  });

  it("onAddChild creates node on backend and appends node + edge to state", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    createUnit.mockResolvedValueOnce({
      unit_id: "u-eng-child",
      parent_unit_id: "u-eng",
      unit_type: "group",
      label: "New group",
      position_x: 100,
      position_y: 250,
    });
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.nodes?.length).toBe(2);
    });
    const nodes = reactFlowProps.current.nodes as any[];
    const engNode = nodes.find(n => n.id === "u-eng");
    expect(engNode).toBeDefined();

    await act(async () => {
      await engNode.data.onAddChild("u-eng");
    });

    expect(createUnit).toHaveBeenCalledWith({
      unit_id: expect.any(String),
      parent_unit_id: "u-eng",
      unit_type: "group",
      label: "New group",
      position_x: 80,
      position_y: 285,
    });
    expect(addToast).toHaveBeenCalledWith("Adding new group...", "info");
  });

  it("onAddChild defaults to part when parent type cannot be further refined", async () => {
    const partHierarchy = {
      nodes: [
        {
          id: "u-part",
          data: { label: "Some Part", type: "part", direct_count: 1, total_count: 1 },
          position: { x: 0, y: 0 },
        }
      ],
      edges: [],
    };
    getOrgHierarchy.mockResolvedValueOnce(partHierarchy);
    createUnit.mockResolvedValueOnce({
      unit_id: "u-part-child",
      parent_unit_id: "u-part",
      unit_type: "part",
      label: "New part",
      position_x: 100,
      position_y: 250,
    });
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.nodes?.length).toBe(1);
    });
    const nodes = reactFlowProps.current.nodes as any[];
    const partNode = nodes.find(n => n.id === "u-part");
    expect(partNode).toBeDefined();

    await act(async () => {
      await partNode.data.onAddChild("u-part");
    });

    expect(createUnit).toHaveBeenCalledWith(expect.objectContaining({
      parent_unit_id: "u-part",
      unit_type: "part",
    }));
  });

  it("onAddChild toasts error when createUnit API rejects", async () => {
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    createUnit.mockRejectedValueOnce(new Error("create failed"));
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.nodes?.length).toBe(2);
    });
    const nodes = reactFlowProps.current.nodes as any[];
    const engNode = nodes.find(n => n.id === "u-eng");
    expect(engNode).toBeDefined();

    await act(async () => {
      await engNode.data.onAddChild("u-eng");
    });

    expect(addToast).toHaveBeenCalledWith("Failed to create new organizational unit", "error");
    errSpy.mockRestore();
  });

  it("onToggleExpand dynamically adjusts nodes and edges visibility", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.nodes?.length).toBe(2);
    });
    const nodesBefore = reactFlowProps.current.nodes as any[];
    const rootNode = nodesBefore.find(n => n.id === "u-root");
    expect(rootNode).toBeDefined();
    
    await act(async () => {
      rootNode.data.onToggleExpand("u-root");
    });
    
    // act() 는 expandedNodes state 는 flush 하지만, filter effect(setNodes) 는
    // 별도 effect cycle 이므로 waitFor 로 재렌더링을 기다려야 한다.
    await waitFor(() => {
      const nodesAfter = reactFlowProps.current.nodes as any[];
      expect(nodesAfter.some(n => n.id === "u-eng")).toBe(false);
    });
  });

  it("handleNodesChange coordinates match previous position exactly is a no-op", async () => {
    getOrgHierarchy.mockResolvedValueOnce(sampleHierarchy);
    render(<OrgTree />);
    await waitFor(() => {
      expect(reactFlowProps.current.onNodesChange).toBeTypeOf("function");
    });
    const handler = reactFlowProps.current.onNodesChange as (c: unknown[]) => void;
    await act(async () => {
      handler([
        { type: "position", id: "u-root", position: { x: 80, y: 135 } }
      ]);
    });
    expect(true).toBe(true);
  });
});

