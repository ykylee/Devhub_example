# Work Backlog — opencode/work_260604-g-draft-mgmt

- Branch: `opencode/work_260604-g-draft-mgmt`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04

## 0. 본 sprint 의 목표

admin/catalog 의 draft 상태 repository 별도 관리 + delete/edit 기능. release_v1_roadmap §3.5 NOW backlog 의 마지막 잔여.

## 1. 작업 단위 분해

| ID | 작업 | 상태 |
| --- | --- | --- |
| WB-01 | 브랜치 + memory 디렉터리 set up | done |
| WB-02 | backend store 메서드 (UpdateRepositoryDraft + DeleteRepository) | planned |
| WB-03 | backend handler + router | planned |
| WB-04 | backend tests | planned |
| WB-05 | frontend service 메서드 | planned |
| WB-06 | frontend RepositoryEditModal | planned |
| WB-07 | frontend admin/catalog 페이지 (filter + Edit + real Delete) | planned |
| WB-08 | frontend vitest + tsc | planned |
| WB-09 | E2E test 추가 | planned |
| WB-10 | 커밋 + push + PR | planned |

## 2. backend 변경 설계

### `repositoryDraftStore` interface 확장 (domain.go)

```go
type repositoryDraftStore interface {
    CreateRepositoryDraft(ctx context.Context, key, slug, providerID string) (domain.Repository, error)
    MarkRepositoryDraftPublishRequested(ctx context.Context, repositoryID int64) (domain.Repository, error)
    GetRepositoryByID(ctx context.Context, repositoryID int64) (domain.Repository, error)
    UpdateRepositoryDraft(ctx context.Context, repositoryID int64, params store.RepositoryUpdateDraftParams) (domain.Repository, error) // NEW
    DeleteRepository(ctx context.Context, repositoryID int64) error // NEW
}
```

### `RepositoryUpdateDraftParams` (store/options.go 신규)

```go
type RepositoryUpdateDraftParams struct {
    Key         *string // nil = unchanged
    Slug        *string // nil = unchanged
    ProviderID  *string // nil = unchanged, empty string = unlink (set to NULL)
}
```

### SQL 패턴

**Update**:
```sql
UPDATE repositories
SET
    name = COALESCE($2, name),
    full_name = COALESCE($3, full_name),
    provider_id = CASE
        WHEN $4::boolean THEN $5::uuid  -- explicit set (NULL or uuid)
        ELSE provider_id                -- unchanged
    END,
    updated_at = NOW()
WHERE id = $1 AND repository_status = 'draft'
RETURNING ... (same columns as GetRepositoryByID)
```

**Delete**:
```sql
-- 1) FK 가드: application_repositories / project_repositories 참조 확인
SELECT EXISTS(SELECT 1 FROM application_repositories WHERE repository_id = $1)
    OR EXISTS(SELECT 1 FROM project_repositories WHERE repository_id = $1)
-- 2) draft 만 삭제
DELETE FROM repositories
WHERE id = $1 AND repository_status = 'draft'
```

### FK 가드 정책

- application_repositories / project_repositories 에서 참조 중이면 `ErrConflict` 반환
- HTTP 409 + body `{code: "repository_has_links", linked_applications: N, linked_projects: M}` (단순화 가능)

## 3. frontend 변경 설계

### `RepositoryEditModal.tsx` (신규)

`RepositoryCreationModal.tsx` 패턴 그대로 mirror + `initialData` prop 으로 pre-fill.
- `getSCMProviders()` 는 `integrationService.listProviders({ provider_type: "scm", enabled: true })` 사용 (PR #470 의 codex P2 fix 반영)
- `handleSubmit` 이 `updateRepository(id, { key, slug, provider_key })` 호출
- 성공 시 onUpdated callback 으로 list refresh

### `admin/catalog/page.tsx` 변경

1. **Filter 토글**: repositories 탭 상단에 `All / Drafts only` 버튼 그룹
   - `useState<"all" | "drafts">("all")` + `filteredRepositories` memo 에 적용
2. **Edit 버튼** (draft 한정): `r.status === "draft"` 일 때만 표시
   - 클릭 시 `editingRepository` state set + `RepositoryEditModal` open
3. **Delete 버튼**: 기존 no-op 토스트 → real handler
   - `confirm("정말 삭제할까요?")` 후 `repositoryService.deleteRepository(r.id)` 호출
   - 409 (links) 시 user-friendly toast

### `repositoryService` 확장

```ts
async updateRepository(id: number, data: { key?: string; slug?: string; provider_key?: string }): Promise<Repository> {
    const body = await apiClient<{ status: string; data: Repository }>(
        "PATCH", `${this.baseUrl}/api/v1/repositories/${id}`, data
    );
    return body.data;
}

async deleteRepository(id: number): Promise<void> {
    await apiClient("DELETE", `${this.baseUrl}/api/v1/repositories/${id}`);
}
```

## 4. 검증 기준 (DoD)

- [ ] backend `go test ./...` — 0 errors
- [ ] backend integration test — happy + draft-only + FK 가드 3 케이스 PASS
- [ ] frontend `tsc --noEmit` — 변경 파일 0 errors
- [ ] frontend `vitest` — 추가한 test 0 fail
- [ ] frontend E2E (TC-REPO-DRAFT-LIFECYCLE) — create → edit → delete 흐름 PASS
- [ ] CI: backend-unit / frontend-unit / e2e 3 잡 SUCCESS
- [ ] 수동 검증: admin/catalog > repositories 탭의 All/Drafts 토글 + Edit/Delete 동작

## 5. carry-over (sprint 종료 후)

- **#1 Gitea issue sync (backend)** — Claude 별도 sprint
- **Codex P2 잔여** — catalog 정합
- **v1.0 N-6 staging 운영** + P1-6/P2 carry-over
