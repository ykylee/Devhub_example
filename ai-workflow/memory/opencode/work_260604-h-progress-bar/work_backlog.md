# Work Backlog — opencode/work_260604-h-progress-bar

- Branch: `opencode/work_260604-h-progress-bar`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04

## 0. 본 sprint 의 목표

project milestones 페이지 (`/projects`) 의 list progress bar 가 실제 task 완료율을 반영하도록 수정. 현재는 status 기반 hardcoded width.

## 1. 작업 단위 분해

| ID | 작업 | 상태 |
| --- | --- | --- |
| WB-01 | 브랜치 + memory 디렉터리 set up | done |
| WB-02 | projects/page.tsx 수정 | planned |
| WB-03 | vitest 추가 | planned |
| WB-04 | tsc + vitest 검증 | planned |
| WB-05 | 커밋 + push + PR | planned |

## 2. 구현 설계

### progress 계산

```ts
// per project
const tasks = await projectService.getProjectTasks(projectId, [
  "todo", "in_progress", "review", "done"
]);
const total = tasks.length;
const doneCount = tasks.filter(t => t.status === "done").length;
const progress = total > 0 ? Math.round((doneCount / total) * 100) : null;
// null = "No tasks yet", 0-100 = 실제 완료율
```

### State

```ts
const [taskProgress, setTaskProgress] = useState<Record<string, number | null>>({});
// null = 아직 fetch 안 됨, number | null = 결과
```

### Fetch

```ts
useEffect(() => {
  if (filteredProjects.length === 0) return;
  const ids = filteredProjects.map(p => p.id);
  void Promise.all(
    ids.map(id => projectService.getProjectTasks(id, [...ALL_STATUSES])
      .then(tasks => ({ id, total: tasks.length, done: tasks.filter(t => t.status === "done").length }))
      .catch(() => ({ id, total: 0, done: 0 })))
  ).then(results => {
    const map: Record<string, number | null> = {};
    for (const r of results) {
      map[r.id] = r.total > 0 ? Math.round((r.done / r.total) * 100) : null;
    }
    setTaskProgress(map);
  });
}, [filteredProjects]);
```

### Render

```tsx
const renderProgress = (project: Project) => {
  const progress = taskProgress[project.id];
  if (progress === undefined) {
    return <div className="h-2 w-full bg-muted rounded-full animate-pulse" />; // loading
  }
  if (progress === null) {
    return <div className="text-[10px] text-muted-foreground">No tasks</div>;
  }
  return (
    <>
      <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
        <div className="h-full bg-primary transition-all duration-1000" style={{ width: `${progress}%` }} />
      </div>
      <span className="text-[10px] font-bold text-muted-foreground">{progress}%</span>
    </>
  );
};
```

## 3. 검증 기준 (DoD)

- [ ] 변경 파일 1개 (`projects/page.tsx`)
- [ ] vitest 추가 (progress 계산 순수 함수 unit test)
- [ ] tsc 0 errors
- [ ] vitest 0 fail
- [ ] CI frontend-unit PASS
- [ ] 수동 검증: /projects 페이지에서 task 있는/없는 프로젝트 progress bar 정확

## 4. carry-over (sprint 종료 후)

- **#1 Gitea issue sync (backend)** — Claude
- **Application progress_percent** (backend data_gap) — 별도
- **Codex P2 잔여** + **v1.0 N-6/P1-6/P2**
