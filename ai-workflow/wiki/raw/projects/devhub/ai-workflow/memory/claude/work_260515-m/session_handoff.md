# Session Handoff — claude/work_260515-m

- 문서 목적: 본 sprint 의 작업 상태와 다음 작업 진입점 인계
- 범위: DREQ carve out 1/4 — RBAC-ADR + Promote-Tx
- 최종 수정일: 2026-05-15
- 상태: in_progress

## 본 sprint 배경

2026-05-15 final EOD (PR #127) 머지 직후 진입. DREQ 도메인 1차 완성 (컨셉 + ADR-0012 + Backend + Frontend) 가 끝났고, 사용자 지시로 carve out 4건 묶음 진입. 본 sprint 는 그 1번째 묶음 (RBAC-ADR + Promote-Tx 병행).

## 산출물

### ADR-0013
`docs/adr/0013-dreq-rbac-row-scoping.md` — dev_requests resource 의 RBAC row-scoping 정책 명문화. ADR-0011 §4.2 의 dev_requests resource 적용 사례.

핵심 결정:
- system_admin: 전체 행 일임 (view/create/edit/delete)
- pmo_manager: 전체 행 위양 (view/edit, 시스템 운영 위임 권한)
- assignee owner-self (action='view'/'register'/'reject' 의 본인 의뢰 행만)
- audit: deny 시 `auth.row_denied` 이벤트 + payload `{actor_role, owner_user_id, resource='dev_requests', denied_reason='owner_mismatch'}`
- 활성화 wire-up: `getDevRequest`, `registerDevRequest`, `rejectDevRequest` 3 handler (이미 wire 됨, ADR 가 사후 명문화)
- carve out: `delete` (close) 와 `reassign` 은 system_admin only (route gate 가 처리)

### DREQ-Promote-Tx
신규 store 메서드 2개:
- `RegisterDevRequestWithNewApplication(ctx, drID, app, primaryRepo)` — pool.BeginTx → INSERT applications + INSERT application_repositories (옵션) → UPDATE dev_requests status='registered' + target → Commit
- `RegisterDevRequestWithNewProject(ctx, drID, project)` — pool.BeginTx → INSERT projects → UPDATE dev_requests status='registered' + target → Commit

handler `registerDevRequest` 분기:
- 기존: `target_id` 직접 매핑 (legacy, MarkDevRequestRegistered)
- 신규: `target_payload` (application 또는 project 정보) → 트랜잭션 메서드 호출

## 다음 sprint (2/4)

DREQ-Admin-UI:
- backend admin endpoint: `POST /api/v1/dev-request-tokens` (system_admin only, plain 1회 노출 패턴), `DELETE /api/v1/dev-request-tokens/:id` (revoke)
- frontend page: `/admin/settings/dev-request-tokens` (accounts_admin password issuance 패턴 따름)
