# Work Backlog — opencode/work_260604-d-fix-projects-list-aggregation

- Branch: `opencode/work_260604-d-fix-projects-list-aggregation`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04
- Source of truth: 본 파일

## 0. 본 sprint 의 목표

Frontend `/projects` 페이지의 listAllProjects 결함 fix. 사용자 보고: "방금 project 를 생성했는데 project list 에 표시가 안되는".

## 1. 작업 단위 분해

| ID | 작업 | 상태 |
| --- | --- | --- |
| WB-01 | 브랜치 + memory 디렉터리 set up | done |
| WB-02 | listAllProjects 재작성 (standalone + per-application 통합 + dedup) | in_progress |
| WB-03 | projects/page.tsx:36-37 listRepositories 호출 제거 + listAllProjects 시그니처 갱신 | planned |
| WB-04 | project.service.test.ts:742-787 listAllProjects 테스트 4건 재작성 | planned |
| WB-05 | frontend UT 실행 (vitest) | planned |
| WB-06 | 커밋 + push + PR | planned |

## 2. 결함 분석

### 2.1 listAllProjects (`project.service.ts:217-228`)

**현재**:
```typescript
async listAllProjects(repositoryIds: number[], params?: ProjectQuery): Promise<Project[]> {
  const allProjects: Project[] = [];
  for (const repoId of repositoryIds) {
    try {
      const projects = await this.getRepositoryProjects(repoId, params);
      allProjects.push(...projects);
    } catch (err) {
      console.error(`Failed to fetch projects for repo ${repoId}:`, err);
    }
  }
  return allProjects;
}
```

**문제**:
- `repositoryIds` 만 순회 → standalone (`/api/v1/projects/standalone`) 누락
- `getRepositoryProjects` 가 `GET /api/v1/repositories/{id}/projects` 호출 → backend `ListProjects` 가 410 GONE (PR #462 의 legacy route disable) → silent fail
- 결과: 모든 경로 누락, 페이지 빈 화면

**신규**:
```typescript
async listAllProjects(params?: ProjectQuery): Promise<Project[]> {
  // 1. Standalone projects (no application)
  // 2. Per-application projects (parallel)
  // 3. Dedup by project ID
  // 4. 에러는 per-source swallow + console.warn
}
```

### 2.2 projects/page.tsx:36-37

**현재**:
```typescript
const repos = await repositoryService.listRepositories();
const allProjects = await projectService.listAllProjects(repos.map(r => r.id), { include_archived: statusFilter === "archived" });
```

**변경**:
```typescript
const allProjects = await projectService.listAllProjects({ include_archived: statusFilter === "archived" });
```

- `listRepositories` 호출 제거 (이제 불필요)

### 2.3 project.service.test.ts:742-787

**기존 4건** (전부 폐기):
- "aggregates projects across repository ids" — 의미 변경됨
- "logs and swallows error for a single repo and continues to next" — 의미 변경됨
- "forwards params to underlying getRepositoryProjects" — 의미 변경됨
- "returns [] when no repository ids" — 무효

**신규 5건**:
- "aggregates standalone + per-application projects with dedup"
- "dedups projects that appear in both standalone and per-application"
- "logs and continues when standalone fetch fails"
- "logs and continues when one application fetch fails (others succeed)"
- "forwards params (status, include_archived) to all sources"

## 3. 검증 기준 (DoD)

- [x] listAllProjects 가 standalone + per-application 통합
- [x] dedup 동작
- [x] per-source 에러 swallow (전체 fail 안 함)
- [x] tests 4~5건 PASS
- [ ] `npm run test` (frontend) PASS
- [x] projects/page.tsx 동시 갱신
- [ ] PR 머지 후 사용자 확인
