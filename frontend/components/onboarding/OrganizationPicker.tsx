"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Search, ChevronRight, Building2, X } from "lucide-react";
import { onboardingService, type OrgSearchResult } from "@/domain/onboarding/service/onboarding.service";
import { identityService, type OrgNode, type OrgEdge } from "@/domain/organization-management/service/identity.service";

interface Props {
  value: string;
  onChange: (unitId: string, label?: string) => void;
  disabled?: boolean;
  allowTree?: boolean;
  "data-testid"?: string;
}

type TreeNode = {
  node: OrgNode;
  children: TreeNode[];
};

function buildTree(nodes: OrgNode[], edges: OrgEdge[]): TreeNode[] {
  const byId = new Map<string, TreeNode>();
  nodes.forEach((n) => byId.set(n.id, { node: n, children: [] }));
  const childIds = new Set<string>();
  edges.forEach((e) => {
    const parent = byId.get(e.source);
    const child = byId.get(e.target);
    if (parent && child) {
      parent.children.push(child);
      childIds.add(child.node.id);
    }
  });
  return nodes.filter((n) => !childIds.has(n.id)).map((n) => byId.get(n.id)!).filter(Boolean);
}

function TreeBranch({
  node,
  depth,
  value,
  onPick,
}: {
  node: TreeNode;
  depth: number;
  value: string;
  onPick: (id: string, label: string) => void;
}) {
  const [open, setOpen] = useState(depth < 1);
  const hasKids = node.children.length > 0;
  const selected = value === node.node.id;
  return (
    <div>
      <div
        className={`flex items-center gap-2 px-2 py-1.5 rounded cursor-pointer text-sm ${
          selected ? "bg-primary/15 text-primary" : "hover:bg-card/50"
        }`}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
      >
        {hasKids ? (
          <button
            type="button"
            onClick={() => setOpen((v) => !v)}
            className="p-0.5 hover:bg-card/50 rounded"
            data-testid={`org-tree-toggle-${node.node.id}`}
          >
            <ChevronRight className={`w-3 h-3 transition-transform ${open ? "rotate-90" : ""}`} />
          </button>
        ) : (
          <span className="w-4 inline-block" />
        )}
        <Building2 className="w-3 h-3 text-muted-foreground" />
        <button
          type="button"
          onClick={() => onPick(node.node.id, node.node.data.label)}
          className="flex-1 text-left truncate"
          data-testid={`org-tree-node-${node.node.id}`}
        >
          {node.node.data.label}
        </button>
      </div>
      {open && hasKids && node.children.map((c) => (
        <TreeBranch key={c.node.id} node={c} depth={depth + 1} value={value} onPick={onPick} />
      ))}
    </div>
  );
}

export function OrganizationPicker({ value, onChange, disabled, allowTree = true, "data-testid": testId }: Props) {
  const [mode, setMode] = useState<"search" | "tree">("search");
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<OrgSearchResult[]>([]);
  const [searching, setSearching] = useState(false);
  const [searchError, setSearchError] = useState<string | null>(null);
  const [tree, setTree] = useState<TreeNode[]>([]);
  const [treeError, setTreeError] = useState<string | null>(null);
  const [selectedLabel, setSelectedLabel] = useState<string>("");
  const debounce = useRef<ReturnType<typeof setTimeout> | null>(null);

  // typeahead query → 2자 이상이면 검색. setState 는 항상 timeout 콜백 내에서만 호출한다
  // (effect body 내 직접 setState 는 cascading render 트리거).
  useEffect(() => {
    if (debounce.current) clearTimeout(debounce.current);
    const trimmed = query.trim();
    if (trimmed.length < 2) {
      debounce.current = setTimeout(() => {
        setResults([]);
        setSearchError(null);
        setSearching(false);
      }, 0);
      return () => {
        if (debounce.current) clearTimeout(debounce.current);
      };
    }
    debounce.current = setTimeout(async () => {
      setSearching(true);
      try {
        const out = await onboardingService.searchOrganizations(trimmed, 20);
        setResults(out);
        setSearchError(null);
      } catch (err) {
        setSearchError(err instanceof Error ? err.message : "검색 실패");
      } finally {
        setSearching(false);
      }
    }, 250);
    return () => {
      if (debounce.current) clearTimeout(debounce.current);
    };
  }, [query]);

  // tree mode 진입 시 hierarchy 로드 (lazy)
  useEffect(() => {
    if (mode !== "tree" || tree.length > 0) return;
    (async () => {
      try {
        const h = await identityService.getOrgHierarchy();
        setTree(buildTree(h.nodes, h.edges));
        setTreeError(null);
      } catch (err) {
        setTreeError(err instanceof Error ? err.message : "조직도 로드 실패");
      }
    })();
  }, [mode, tree.length]);

  const picked = useMemo(() => {
    if (!value) return null;
    const fromResults = results.find((r) => r.unit_id === value);
    if (fromResults) return fromResults.name;
    return selectedLabel || value;
  }, [value, results, selectedLabel]);

  function pick(unitId: string, label?: string) {
    if (label) setSelectedLabel(label);
    onChange(unitId, label);
  }

  function clearPick() {
    setSelectedLabel("");
    onChange("", undefined);
  }

  return (
    <div className="space-y-2" data-testid={testId}>
      <div className="flex gap-2 text-xs">
        <button
          type="button"
          onClick={() => setMode("search")}
          disabled={disabled}
          className={`px-3 py-1.5 rounded font-bold uppercase tracking-wider ${
            mode === "search" ? "bg-primary text-primary-foreground" : "bg-card hover:bg-card/70"
          }`}
          data-testid="org-picker-mode-search"
        >
          검색
        </button>
        {allowTree && (
          <button
            type="button"
            onClick={() => setMode("tree")}
            disabled={disabled}
            className={`px-3 py-1.5 rounded font-bold uppercase tracking-wider ${
              mode === "tree" ? "bg-primary text-primary-foreground" : "bg-card hover:bg-card/70"
            }`}
            data-testid="org-picker-mode-tree"
          >
            트리
          </button>
        )}
      </div>

      {picked && (
        <div className="flex items-center gap-2 bg-primary/10 border border-primary/30 rounded px-3 py-2 text-sm" data-testid="org-picker-selected">
          <Building2 className="w-4 h-4 text-primary" />
          <span className="flex-1 font-medium">{picked}</span>
          <span className="text-xs text-muted-foreground font-mono">{value}</span>
          {!disabled && (
            <button
              type="button"
              onClick={clearPick}
              className="p-1 hover:bg-card/70 rounded"
              data-testid="org-picker-clear"
              aria-label="조직 선택 해제"
            >
              <X className="w-3 h-3" />
            </button>
          )}
        </div>
      )}

      {mode === "search" && (
        <div className="space-y-2">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="조직명 검색 (2자 이상)"
              disabled={disabled}
              className="w-full pl-10 pr-3 py-2 bg-card border border-border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              data-testid="org-picker-search-input"
            />
          </div>
          {searching && <p className="text-xs text-muted-foreground">검색 중…</p>}
          {searchError && <p className="text-xs text-red-500" role="alert">{searchError}</p>}
          {!searching && results.length > 0 && (
            <ul className="max-h-64 overflow-y-auto bg-card/50 border border-border rounded divide-y divide-border" data-testid="org-picker-results">
              {results.map((r) => (
                <li key={r.unit_id}>
                  <button
                    type="button"
                    onClick={() => pick(r.unit_id, r.name)}
                    className="w-full px-3 py-2 text-left hover:bg-card/80 flex items-center gap-2"
                    data-testid={`org-picker-result-${r.unit_id}`}
                  >
                    <Building2 className="w-3 h-3 text-muted-foreground" />
                    <span className="flex-1 truncate">{r.name}</span>
                    <span className="text-xs text-muted-foreground font-mono">{r.unit_id}</span>
                  </button>
                </li>
              ))}
            </ul>
          )}
          {!searching && query.trim().length >= 2 && results.length === 0 && !searchError && (
            <p className="text-xs text-muted-foreground">일치하는 조직이 없습니다.</p>
          )}
        </div>
      )}

      {allowTree && mode === "tree" && (
        <div className="bg-card/50 border border-border rounded max-h-72 overflow-y-auto py-1" data-testid="org-picker-tree">
          {treeError && <p className="text-xs text-red-500 px-3 py-2" role="alert">{treeError}</p>}
          {!treeError && tree.length === 0 && <p className="text-xs text-muted-foreground px-3 py-2">조직도 로드 중…</p>}
          {tree.map((root) => (
            <TreeBranch key={root.node.id} node={root} depth={0} value={value} onPick={pick} />
          ))}
        </div>
      )}
    </div>
  );
}
