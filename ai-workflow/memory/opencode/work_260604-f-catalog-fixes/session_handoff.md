# Session Handoff — opencode/work_260604-f-catalog-fixes

- Branch: `opencode/work_260604-f-catalog-fixes`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04
- Sprint: admin/catalog 3건 결함 수정 (release_v1_roadmap §3.5 의 NOW backlog 후속)

## 🎯 Current Focus

사용자 보고 5건 중 frontend 3건:
1. **Bug A** (단일 root cause) — `frontend/domain/application-lifecycle/view/ProjectCreationModal.tsx:49` 가 `Repository` 타입 객체의 `repo_provider` 를 하드코딩 `"github"` 로 설정. 이게:
   - **#2** "project edit 의 repositories 가 실제 연동과 불일치" 원인
   - **#3** "gitea 저장소가 (github) 로 표시" 원인
   - 수정: `r.provider_key` (= backend `integration_providers.provider_key` JOIN derive) 로 대체
2. **Bug B** — `frontend/components/project/RepositoryCreationModal.tsx` 의 SCM Provider Key 가 free-text input. `getSCMProviders()` 기반 dropdown 으로 교체 (**#4** "repository 생성 시 scm 을 등록된 목록에서 선택").

Scope out (별도 sprint 후보):
- **#1** Gitea issue sync (backend) — per-provider webhook 이 이벤트 저장만 하고 Process() 호출 안함. backend 변경.
- **#5** Draft 관리 기능 — admin/catalog 의 'Drafts' 탭 + repository delete/edit. 기능 추가 작업.

## 📊 Work Status

- [WB-01] 브랜치 + memory 디렉터리 set up: done
- [WB-02] Bug A: ProjectCreationModal.tsx:49 hardcoded "github" fallback 제거: planned
- [WB-03] Bug B: RepositoryCreationModal SCM dropdown 교체: planned
- [WB-04] lsp_diagnostics + frontend type check + next build 검증: planned
- [WB-05] 커밋 + push + PR + state/handoff final: planned

## ⏭️ Next Actions

- 본 sprint 종료 후 (PR 머지 후):
  1. **#1 Gitea issue sync (backend)** — Claude 별도 sprint. `POST /api/v1/integration/providers/:provider_id/webhook` 가 EventProcessor.Process() 호출하도록 보강.
  2. **#5 Draft 관리 기능** — 사용자 우선순위 확인 후 별도 sprint. admin/catalog 의 'Drafts' 탭 + repository 별 delete/edit handler.
  3. **Codex P2 잔여** (catalog mismatch `application_repositories.repo_provider` ↔ `repositories.provider_id`) — 별도 governance 결정 후 backend migration.

## ⚠️ Risks & Blockers

- `getSCMProviders()` 가 backend `scm_providers` (구 4종 하드코드: gitea/github/bitbucket/forgejo) 와 `integration_providers` (신규) 의 양쪽을 다루는지 불명확. dropdown 의 enabled provider 가 expected set 인지 확인 필요. 시드 데이터 점검 후 필요시 backend reconciliation.
- dropdown 선택 후 `provider_key` 가 빈 string 으로 저장되지 않도록 `required` + 빈 값 거부 검증 필요.

## 📁 Key Files

- `frontend/components/project/RepositoryCreationModal.tsx` (BUG B: 14-18, 36-40, 107-120)
- `frontend/domain/application-lifecycle/view/ProjectCreationModal.tsx:49` (BUG A: hardcoded "github")
- `frontend/domain/application-lifecycle/service/project.service.ts:49-52` (`getSCMProviders()` — reuse)
- `frontend/domain/repository-integration/view/RepositoryLinkModal.tsx:96-114` (REFERENCE: dropdown pattern)
