# Session Handoff — deepseek/test-scenarios-20260601

- **에이전트**: Sisyphus (Reasonix/deepseek-v4-flash)
- **세션 시작**: 2026-06-01
- **상태**: 🔄 작업 중 — Phase 5 DREQ E2E 완료, Phase 6 권한 모델 컨셉 완료, TC 16개 작성

## 완료된 작업

### Phase 1~4: E2E 통합 테스트 (선행)
- Phase 1: Keycloak 온보딩/RBAC — 4/4 완료
- Phase 2: App/Project/Repo/Gitea — 5/5 완료 (로컬 Gitea 추가)
- Phase 3: Gitea PR/Issue/Sync — 5/5 완료 (증분 sync 미동작 확인)
- Phase 4: CI Build — DB INSERT 조회 (생성 API 부재 확인)
- 로컬 Gitea 설정: `docker-compose.yml`에 `gitea:1.22` 추가
- 테스트 리포트: `docs/planning/integrated_test_report_20260601.md`
- 로드맵 갱신: `docs/planning/release_v1_roadmap.md` — P0-4, P1-6, P1-7 신규 carve
- Sprint Plan: `docs/planning/sprint-plan-20260601.md` — 3 sprint 구성

### Phase 5: DREQ E2E 테스트 (완료)
- DREQ 전체 lifecycle 시나리오 TC 10개 작성 (`docs/domain/dev-request/test_cases.md`)
- Go handler 통합 테스트 (`dev_requests_e2e_test.go`) — intake → read → register → reject → close
- `BypassEnforceRoutePermission` 도입 (intake endpoint bypass)
- 0 BUG, 전수 TC pass
- traceability `report.md` 갱신 (TC-078 ~ TC-087)
- Git commit: `7b54f0d`

### Phase 6: Role-Based View Scope (완료)
- **Role model concept**: `docs/planning/role-access-concept.md`
  - Two-Dimensional RBAC 정의: 3 system roles (developer/team_manager/system_admin) × 4 resource roles (project_member/project_leader/application_leader/org_head)
  - Enforcement architecture 설계 (enforceRowReadScope 추가)
  - 4-Phase 점진적 구현 계획
  - 4개 background agent 활용 사전 조사 (RBAC, project_members, org hierarchy, enforcement)
- **Test cases**: `docs/domain/application-lifecycle/test_cases.md` — TC-088 ~ TC-103 (16개)
  - Baseline membership scope (TC-088~090)
  - Project leader management info (TC-091~092)
  - Application leader dashboard (TC-093~094)
  - Org head subtree scope (TC-095~096)
  - Team manager team-wide scope (TC-097~099)
  - System admin unrestricted (TC-100)
  - Scope merging (TC-101)
  - Negative tests (TC-102~103)

## 핵심 작업 파일

- `docs/planning/role-access-concept.md` — 역할 모델 컨셉
- `docs/domain/application-lifecycle/test_cases.md` — view scope TC 16개
- `docs/domain/dev-request/test_cases.md` — DREQ E2E TC 10개
- `docs/planning/integrated_test_report_20260601.md` — 통합 테스트 리포트
- `docs/planning/release_v1_roadmap.md` — 로드맵 갱신
- `docs/planning/sprint-plan-20260601.md` — sprint plan
- `backend-core/internal/httpapi/dev_requests_e2e_test.go` — DREQ E2E 통합 테스트

## 미완료 / 후속 필요

| 항목 | 우선순위 | Phase | 비고 |
|------|---------|-------|------|
| Phase 1: Matrix 확장 + Row Filter | P0 | 6-P1 | developer에 apps/projects:view 추가, project_members 기반 row filter |
| Phase 2: Resource Role Enforcement | P1 | 6-P2 | enforceRowReadScope, management info gate |
| Phase 3: team_manager + org_head | P1 | 6-P3 | 신규 role 추가 + subtree scope |
| Phase 4: 정책 안정화 | P2 | 6-P4 | manager/pmo_manager migration |
| CI Run 생성 API (P0-4) | **P0** | sprint -h | Claude 처리 |
| Sign-out endpoint (P1-6) | P1 | sprint -i | Claude 처리 |
| Keycloak role sync (P1-1) | P1 | sprint -h | Claude 처리 |
| Manager role RBAC 검증 (N-10) | P2 | — | mgr-user-b API 검증 |
| Gitea Webhook 증분 sync (BUG-06) | P3 | v1.1 | 연기 |

## 현재 환경 상태

- colima docker-compose 7개 서비스 운영 중 (Gitea 포함)
- Keycloak: `localhost:8180/devhub/auth/keycloak`
- Gitea: `localhost:3300` (admin yklee/yklee12!)
- DB: `localhost:5433` (postgres/your_password)
- Provider ID: `7c235b75-4246-4f40-9267-202222f07ddf`
- Application: `f0a18b05-92e6-45d6-ba00-4d2228550208` (TESTAPP01)
- Project: `bd4e187e-7267-407e-9a9d-a7963ac7464c` (ALPHA-SPRINT-1)
- Repo: `1` (testapp-alpha-repo)
