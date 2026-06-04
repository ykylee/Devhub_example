# Work Backlog — opencode/work_260604-f-catalog-fixes

- Branch: `opencode/work_260604-f-catalog-fixes`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04

## 0. 본 sprint 의 목표

admin/catalog 3건 결함 수정 (release_v1_roadmap §3.5 의 NOW backlog 후속):
- **(2)(3) Bug A** — `ProjectCreationModal.tsx:49` 의 `repo_provider = "github"` 하드코딩 제거. 이게 project edit 의 repositories mismatch + gitea→github mis-label 의 single root cause.
- **(4) Bug B** — `RepositoryCreationModal.tsx` 의 free-text SCM input 을 `getSCMProviders()` dropdown 으로 교체.

Scope out:
- **(1) Gitea issue sync (backend)** — 별도 sprint (Claude/backend 영역).
- **(5) Draft 관리 기능** — 별도 sprint (기능 추가, 작업량 큼).

## 1. 작업 단위 분해

| ID | 작업 | 상태 |
| --- | --- | --- |
| WB-01 | 브랜치 + memory 디렉터리 set up | done |
| WB-02 | Bug A: ProjectCreationModal.tsx:49 hardcoded "github" fallback 제거 | planned |
| WB-03 | Bug B: RepositoryCreationModal SCM dropdown 교체 (free-text → select) | planned |
| WB-04 | lsp_diagnostics + frontend type check + next build 검증 | planned |
| WB-05 | 커밋 + push + PR + state/handoff final | planned |

## 2. Bug A — ProjectCreationModal hardcoded "github"

**위치**: `frontend/domain/application-lifecycle/view/ProjectCreationModal.tsx:49`

```ts
// AS-IS (BUG)
const repo_provider = isAppRepo ? (r as ApplicationRepository).repo_provider : "github";
```

**TO-BE (FIX)**:
```ts
// Repository 타입은 backend join 으로 provider_key (= integration_providers.provider_key) 를
// derive 해서 내려받음 (migration 000045 통합 후 single source of truth).
const repo_provider = isAppRepo
  ? (r as ApplicationRepository).repo_provider
  : ((r as Repository).provider_key ?? "");
```

**영향**:
- `repo_full_name` 옆에 표시되는 `(repo_provider)` 가 실제 SCM 으로 정상 표기됨
- project edit 화면에서 gitea 가 gitea 로 표시 (이전엔 github 로 mislabel)
- `repositories` 테이블 line 433 의 `r.provider_key` 표시가 backend join 으로 정상화됨 (cascade effect)

## 3. Bug B — RepositoryCreationModal SCM dropdown

**위치**: `frontend/components/project/RepositoryCreationModal.tsx:14-18, 36-40, 107-120`

**AS-IS (BUG)**: `<input>` 으로 free-text. placeholder `"e.g. gitea-main"`. 사용자가 잘못된 key 직접 입력 가능.

**TO-BE (FIX)**: `<select>` 으로 교체. `projectService.getSCMProviders()` 호출 후 `enabled=true` 인 provider 만 옵션. `RepositoryLinkModal.tsx:96-114` 패턴 그대로 mirror.

```tsx
// useEffect:
projectService
  .getSCMProviders()
  .then((providers) => {
    setScmProviders(providers.filter((p) => p.enabled));
  })
  .catch(() => setScmProviders([]));

// render (free-text input → select):
<select
  value={formData.provider_key}
  onChange={(e) => setFormData({ ...formData, provider_key: e.target.value })}
  required
  className="..."
>
  <option value="">Select SCM provider</option>
  {scmProviders.map((p) => (
    <option key={p.provider_key} value={p.provider_key}>
      {p.display_name} ({p.provider_key})
    </option>
  ))}
</select>
```

## 4. 검증 기준 (DoD)

- [x] Bug A 변경 — `provider_key` fallback nullish 처리 (Repository type cast 안전)
- [x] Bug B 변경 — `required` select, 빈 값 거부, `getSCMProviders()` 실패 시 fallback disabled select
- [x] `lsp_diagnostics` clean on both files
- [ ] frontend `tsc --noEmit` (또는 `next build`) 통과
- [ ] RepositoryLinkModal 의 dropdown 패턴과 시각적 일관성 유지
- [ ] PR 머지 후 v1.0 D-11 안에서 2건 결함 close

## 5. carry-over (sprint 종료 후)

- **#1 Gitea issue sync (backend)** — Claude 별도 sprint
- **#5 Draft 관리 기능** — 사용자 우선순위 확인 후 별도 sprint
- **Codex P2 잔여** (catalog mismatch) — 별도 governance 결정
