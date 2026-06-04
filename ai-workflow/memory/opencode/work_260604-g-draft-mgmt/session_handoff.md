# Session Handoff — opencode/work_260604-g-draft-mgmt

- Branch: `opencode/work_260604-g-draft-mgmt`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04
- Sprint: #5 Draft 관리 (release_v1_roadmap §3.5 의 NOW backlog 마지막 frontend 잔여)

## 🎯 Current Focus

사용자 5건 결함 보고 중 마지막 잔여. admin/catalog 의 repository 중 `status=draft` 인 것만 별도 관리 (filter + delete + edit).

**Scope** (사용자 확정):
- backend: `PATCH /api/v1/repositories/:id` + `DELETE /api/v1/repositories/:id` (draft 한정, FK reference 가드)
- frontend: `RepositoryEditModal` + `All/Drafts only` 토글 + 진짜 delete handler + Edit 버튼 (draft 한정)

## 📊 Work Status

- [WB-01] 브랜치 + memory 디렉터리 set up: done
- [WB-02] backend store 메서드 (UpdateRepositoryDraft + DeleteRepository) 추가: planned
- [WB-03] backend handler + router 등록: planned
- [WB-04] backend tests (memory mock + integration): planned
- [WB-05] frontend service 메서드 (update/delete): planned
- [WB-06] frontend RepositoryEditModal: planned
- [WB-07] frontend admin/catalog 페이지 (filter + Edit + real Delete): planned
- [WB-08] frontend vitest + tsc 검증: planned
- [WB-09] E2E TC 추가: planned
- [WB-10] 커밋 + push + PR: planned

## ⏭️ Next Actions

- 본 sprint 종료 후 (PR 머지 후):
  1. **#1 Gitea issue sync (backend)** — Claude 별도 sprint
  2. **Codex P2 잔여** — `scm_providers` ↔ `integration_providers` catalog 정합
  3. **v1.0 N-6 staging 운영** + P1-6/P2 carry-over

## ⚠️ Risks & Blockers

- `UpdateRepositoryDraft` 의 key/slug unique constraint 처리: 변경 시 동일 key/slug 가 이미 존재하면 `ErrConflict` 반환. frontend 에서 user-friendly error 표시 필요.
- `DeleteRepository` FK 가드: `application_repositories` / `project_repositories` 참조 시 `ErrConflict` 반환 + 409. frontend 에서 user-friendly error ("이 저장소는 N 개 application/project 에 연결되어 있어 먼저 unlink 필요") 표시 필요.
- `provider_id` UUID 직접 저장 vs `provider_key` resolve: 기존 `CreateRepositoryDraft` 가 `provider_key` → `provider_id` 변환은 handler 에서. 일관성 위해 `UpdateRepositoryDraft` 도 동일 패턴.
- E2E seed 가 충분히 복잡한 draft 시나리오를 다루는지 확인 필요. 단순 happy path 만 추가.

## 📁 Key Files (변경 대상)

**Backend**:
- `backend-core/internal/domain/application-lifecycle/repository/projects.go` (line 253 이후 — 새 메서드 추가)
- `backend-core/internal/httpapi/domain.go:40-44` (`repositoryDraftStore` interface 확장)
- `backend-core/internal/httpapi/domain.go:148+` (handler 추가)
- `backend-core/internal/httpapi/router.go:329-330` (route 추가)
- `backend-core/internal/httpapi/applications_test.go` (memory mock 확장)
- `backend-core/internal/domain/application-lifecycle/repository/projects_creation_integration_test.go` 또는 신규 (integration test)

**Frontend**:
- `frontend/domain/repository-integration/service/repository.service.ts` (메서드 추가)
- `frontend/components/project/RepositoryEditModal.tsx` (신규)
- `frontend/app/(dashboard)/admin/catalog/page.tsx` (filter + Edit + Delete 교체)
- `frontend/domain/repository-integration/service/repository.service.test.ts` (test 추가)
- `frontend/tests/e2e/repositories-draft-lifecycle.spec.ts` (신규)
