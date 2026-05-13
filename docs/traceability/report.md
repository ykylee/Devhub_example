# Traceability Report — 1차 종합

- 문서 목적: DevHub 의 SDLC 자산 (요구사항 → 설계 → 로드맵 → 구현 → 단위테스트 → E2E) 사이 추적 관계를 단일 매트릭스로 시각화한다.
- 범위: M0–M3 (M4 는 planned 표기). ADR 은 별도 §4 인덱스.
- 대상 독자: 모든 contributor, 후속 리뷰어, 외부 감사.
- 상태: accepted
- 최종 수정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-c`.
- 관련 문서: [`README.md`](./README.md), [`conventions.md`](./conventions.md), [`sync-checklist.md`](./sync-checklist.md), [`../governance/document-standards.md`](../governance/document-standards.md).

## 1. 사용법

- 단계 간 추적 관계를 한 페이지에서 찾을 때 §3 종합 매트릭스 사용.
- 특정 단계의 ID 정의를 찾을 때 §2 단계별 인덱스 사용.
- ADR 은 §4 의 링크.
- 갱신 절차는 [`sync-checklist.md`](./sync-checklist.md) §3.6 참조.

## 2. 단계별 인덱스 (요약)

### 2.1 Requirements (REQ)

- **Functional**: REQ-FR-01 ~ REQ-FR-105 (총 105 항목, `docs/requirements.md` §2–§5 + `docs/backend/requirements.md` §1–§5 + `docs/backend_requirements_org_hierarchy.md` §1–§3 + `docs/frontend_integration_requirements.md` §2–§3 분포).
- **Non-functional**: REQ-NFR-01 ~ REQ-NFR-26 (총 26 항목, 보안/성능/배포 정책 + 운영 hygiene + API 표준).

### 2.2 Design (ARCH / API)

- **Architecture**: ARCH-01 ~ ARCH-17 (총 17 항목, `docs/architecture.md` + `docs/org_chart_ux_spec.md` + `docs/organizational_hierarchy_spec.md` 분포).
- **API contract**: API-01 ~ API-40 (총 40 항목, `docs/backend_api_contract.md` §3–§12 분포).

### 2.3 Roadmap (RM)

- **M0**: RM-M0-01 (1 항목, `X-Devhub-Actor` deprecation).
- **M1**: RM-M1-01 ~ RM-M1-04 (4 항목, RBAC track + API 계약 §11 재작성).
- **M2**: RM-M2-01 ~ RM-M2-16 (16 항목, 인증/계정/조직/UX/audit + CI).
- **M3**: RM-M3-01 ~ RM-M3-07 (7 항목, Sign Up + WebSocket 확장 + AI Gardener — 본 sprint 기준 모두 planned).
- **M4**: 본 sprint 스코프 밖. §3 매트릭스 마지막 행에 planned 표기.

### 2.4 Implementation (IMPL)

- **Backend (`backend-core`)**: IMPL-auth-01..07, rbac-01..04, audit-01..02, account-01..04, org-01..04, command-01..05, serviceaction-01, domain-01..03, dashboard-01, infra-01, store-01..03, gitea-01..02, kratos-01..04, config-01, hrdb-01, realtime-01, me-01, health-01, idp-schema-01 (47 항목).
- **Frontend**: IMPL-frontend-auth-01..06, login-01..03, logout-01, dashboard-01, account-01, admin-01, admin-users-01, admin-org-01, admin-rbac-01, admin-audit-01, org-01, org-comp-01..03, role-01..03, service-auth-01, service-account-01, service-rbac-01, service-audit-01, service-org-01, service-realtime-01, service-api-01, layout-01..02, store-01 (32 항목).
- **Backend AI**: 미사용 (Phase 4 분석에서 placeholder 만 발견).

### 2.5 Unit tests (UT)

- **Backend Go**: UT-httpapi-01..24, rbac-01..03, domain-01..02, auth-01, gitea-01, normalize-01, store-01..03, commandworker-01..02, serviceaction-01, config-01, main-01 (41 파일).
- **Frontend Vitest**: UT-frontend-utils-01, frontend-auth-01..04, frontend-services-01 (6 파일).

### 2.6 E2E TC

- **M2 (28 TC)**: TC-USR-01..06, TC-USR-CRUD-01..03, TC-ACC-01..03, TC-ACC-PROFILE-01, TC-NAV-01..03 + TC-NAV-SIM-01, TC-AUD-01..02, TC-AUTH-NEG-01, TC-AUTH-NOAUTH-01, TC-AUTH-SIGNOUT-REDIR-01, TC-USER-SWITCH-01, TC-RBAC-SUB-01, TC-RBAC-MGR-01, TC-SIGNUP-01..04, TC-PERMISSIONS-SMOKE-01 (`docs/tests/test_cases_m2_auth.md`).
- **M3 (9 TC)**: TC-ORG-LIST-01..02, TC-ORG-UNIT-01..03, TC-ORG-MEM-01..02, TC-ORG-CHART-01..02 (`docs/tests/test_cases_m3_organization.md`).
- **추가** (spec 파일 안에만 정의되어 TC 카탈로그 외): `auth.spec.ts` 의 TC-AUTH-01..06 등 — 향후 TC 카탈로그로 흡수 권장.

## 3. 종합 매트릭스 (도메인 그룹 단위)

| 도메인 | REQ | ARCH / API | ROADMAP | IMPL | UT | TC |
| --- | --- | --- | --- | --- | --- | --- |
| **인증 (auth / OIDC)** | FR-19, 21–24, 65, 67; NFR-3, 18 | ARCH-11, 12; API-19–24, 35 | M1-04, M2-01, 02, 03, 09 | auth-01..07; frontend-auth-01..06; login-01..03; logout-01 | httpapi-01..04; auth-01; frontend-auth-01..04 | TC-AUTH-NEG-01, NOAUTH-01, SIGNOUT-REDIR-01; TC-USER-SWITCH-01 |
| **회원가입 (signup)** | FR-25, 61–63 | ARCH-12; API-23 | M3-01 (planned) | auth-05; frontend-login-03 | httpapi-15 | TC-SIGNUP-01..04 |
| **RBAC** | FR-27, 86; NFR-26 | ARCH-13; API-26–31, 38–40 | M1-01, 02, 03; M2-11 | rbac-01..04; frontend-admin-rbac-01; frontend-service-rbac-01; org-comp-02 | rbac-01..03; domain-01 | TC-RBAC-SUB-01, MGR-01; TC-PERMISSIONS-SMOKE-01 |
| **계정 관리 (account admin + self-service)** | FR-15–18, 20, 22, 23, 26, 61–67; NFR-3, 4, 5, 7, 17, 19, 20 | ARCH-12, 14; API-25, 32, 35 | M2-04, 05, 06 | account-01..04; frontend-account-01; frontend-admin-users-01; frontend-service-account-01 | httpapi-05, 06, 07 | TC-USR-01..06; TC-USR-CRUD-01..03; TC-ACC-01..03; TC-ACC-PROFILE-01 |
| **조직 계층 (organization)** | FR-68–80; NFR-21 | ARCH-15, 16, 17; API-33, 34 | M2-07, 08 | org-01..04; frontend-org-01; frontend-admin-org-01; frontend-org-comp-01..03; frontend-service-org-01 | httpapi-07; store-01 | TC-ORG-LIST-01..02; TC-ORG-UNIT-01..03; TC-ORG-MEM-01..02; TC-ORG-CHART-01..02 |
| **감사 (audit)** | FR-18, 26, 102; NFR-4 | ARCH-14; API-18, 39 | M2-15 (in_progress, PR-M2-AUDIT) | audit-01, 02; kratos-03 | httpapi-19, 24; frontend (Vitest 없음) | TC-AUD-01, 02 |
| **명령 lifecycle (command / mitigation / service action)** | FR-58, 59, 84, 95, 100, 101 | API-15, 16, 17, 36, 37 | M1-01..03 (envelope 정합); M3-02 (replay, planned) | command-01..05; serviceaction-01; realtime-01 | httpapi-09, 13; commandworker-01, 02; serviceaction-01; domain-02 | (E2E 미커버 — gap §5.1) |
| **실시간 (realtime / WebSocket)** | FR-56, 57, 60, 82, 83, 104, 105; NFR-11, 22, 23 | ARCH-05; API-14 | M3-02, 03 (planned) | realtime-01; frontend-service-realtime-01 | httpapi-13 | (E2E 미커버 — gap §5.1) |
| **인프라 토폴로지 (infra)** | FR-12, 13, 97, 98, 99; NFR-12, 16 | ARCH-04, 09; API-06 | M3-02 (planned, infra event publish) | infra-01; frontend-role-03 (gardener) | httpapi-12 | (E2E 미커버 — gap §5.1) |
| **Webhook + 도메인 데이터 (gitea)** | FR-49, 50, 51, 52, 53, 54, 55 | ARCH-06, 07, 08; API-02, 03, 07–13 | (M0 이전 phase 완료) | gitea-01, 02; domain-01..03; normalize | httpapi-10, 14; gitea-01; normalize-01; store-03 | (E2E 미커버 — gap §5.1) |
| **대시보드 / 메트릭 / me** | FR-1–11, 28–36, 81, 85, 88, 89, 96 | ARCH-10; API-05, 36, 38 | (Phase 4 이전 완료) | dashboard-01; me-01; frontend-dashboard-01; frontend-role-01..03; frontend-store-01; frontend-layout-01..02; frontend-service-api-01 | httpapi-08, 11, 22 | TC-NAV-01..03; TC-NAV-SIM-01 |
| **CI / 거버넌스** | NFR-1 (no-docker) | ADR-0001 §5; [ADR-0003](../adr/0003-no-docker-policy-ci-scope.md) | M2-16 (CI 1차, PR #86); FU-CI-1..4 (PR #87); ADR-0003 (PR #88); 거버넌스 + 매트릭스 1차 (PR #89); 갭 분석 + 메타 헤더 표준화 (본 PR / sprint `claude/work_260513-d`) | (build infra: `.github/workflows/ci.yml`, `scripts/ci-setup.sh`, `infra/idp/*.ci.yaml`); `docs/governance/*`, `docs/traceability/*` | (lint 미도입, FU-CI 향후) | (CI run 자체가 검증) |
| **M4 (스코프 밖, planned)** | FR-37–48, 56–57, 60, 90–94 (일부) | API-14 (확장) | M4 항목 (M4 컬럼 별도) | (미진입) | (미진입) | (미진입) |

> Note — 매트릭스 셀의 ID 는 §2 의 단계별 인덱스를 줄여서 표기 (예: `auth-01..07` = `IMPL-auth-01..IMPL-auth-07`). 한 도메인이 여러 단계에 걸쳐 영향을 주므로 정확한 1:1 매핑은 §2 인덱스 + 단계별 문서 본문의 ID 노출 (`document-standards.md` §5) 로 확인.

## 4. ADR 인덱스

| ADR | 제목 | 상태 | 영향 도메인 |
| --- | --- | --- | --- |
| [ADR-0001](../adr/0001-idp-selection.md) | IdP 선정 (Ory Hydra + Kratos) | accepted (2026-05-07) | 인증, 회원가입, 계정 관리 |
| [ADR-0002](../adr/0002-rbac-policy-edit-api.md) | RBAC policy edit API (DB-backed matrix) | accepted (2026-05-08) | RBAC |
| [ADR-0003](../adr/0003-no-docker-policy-ci-scope.md) | No-Docker 정책 CI 범위 명문화 | accepted (2026-05-13, PR #88) | CI / 거버넌스 |

## 5. Gap 요약

### 5.1 E2E 미커버 도메인

후속 sprint 에서 보완 후보. TC 후보 ID 는 등재만 (실제 spec 작성은 별도).

| 도메인 | 현황 | 가능한 TC 후보 | 우선순위 |
| --- | --- | --- | --- |
| 명령 lifecycle / mitigation | 단위테스트 (`UT-httpapi-09`, `UT-commandworker-01..02`, `UT-domain-02`) 만, e2e UI 흐름 미커버 | TC-CMD-CREATE-01 (대시보드 → service-action 생성 → command_id 반환), TC-CMD-STATUS-01 (상태 조회 → UI 갱신) | P2 |
| 실시간 (WebSocket) | M3 planned. 현재는 `command.status.updated` 만 publish | M3 진입 시: TC-WS-CONNECT-01, TC-WS-CMD-STATUS-01, TC-WS-RESILIENCE-01 (re-connect) | P3 (M3 의존) |
| 인프라 토폴로지 React Flow | 정적 데이터 렌더만 e2e 미커버 | TC-INFRA-RENDER-01 (정적 노드/엣지), TC-INFRA-NODE-CLICK-01 (상세 패널), TC-INFRA-GROUP-TOGGLE-01 | P2 |
| Webhook 처리 (gitea HMAC) | 단위테스트 (`UT-httpapi-14`, `UT-gitea-01`) 로 검증, 외부 영향 e2e 어려움 | E2E 후보 없음 — 통합 테스트 (Go test + 모의 webhook server) 가 자연 | P3 (E2E 외 검증으로 충분) |

### 5.2 ID 부재 / 매핑 누락

| 항목 | 상태 | 처리 |
| --- | --- | --- |
| Backend AI (`backend-ai/`) placeholder | open | M3-04 AnalysisService gRPC client 진입 시 IMPL-ai-XX 발급. |
| Frontend 컴포넌트 Vitest (Header, Sidebar, AuthGuard 등) | open | 후속 sprint 후보 (P2). |
| `auth.spec.ts` 의 TC ID 카탈로그 흡수 | **closed (2026-05-13, sprint `claude/work_260513-d`)** | 재검증 결과 spec 안의 TC-AUTH-NEG-01 + TC-AUTH-NOAUTH-01 모두 `test_cases_m2_auth.md` 의 TC 카탈로그에 이미 흡수되어 있음. 1차 분석에서 "01..06 미흡수" 라고 적은 것은 사실과 다름. |
| API §12 RBAC 정책 편집 API 의 IMPL 정밀 매핑 | open | ADR-0002 결정 + 6 PR 분할 머지되었으나 매트릭스의 IMPL-rbac-* 가 일부 endpoint 만 cover. endpoint 별 IMPL ID 정밀 매핑은 후속 sprint. |
| 카탈로그된 TC 가 spec 으로 구현됐는지 역검증 | open | TC-AUD-02 등 일부 TC 가 카탈로그에만 존재할 가능성 — spec 파일 grep 으로 자동 검증할 hygiene 후보. |

### 5.3 문서 ↔ 코드 불일치

| 항목 | 상태 | 처리 |
| --- | --- | --- |
| ADR-0001 vs `frontend_integration_requirements.md` §3.8 | **closed (2026-05-13, sprint `claude/work_260513-d`)** | §3.8 의 "재설계 예정" 노트를 "deprecated" 노트로 명확화 + Phase 13 머지 후 실제 endpoint (API §11.5 / §12.8.2) 로 redirect 링크. |
| X-Devhub-Actor 완전 제거 trigger | open | architecture.md §6.2.3 의 deprecation warning 경로 명시되어 있으나 완전 제거 일정 미정의. 후속 ADR 후보. |
| RBAC cache 다중 인스턴스 일관성 | open | API §12.10 의 미해결 (M1-DEFER-E). pub/sub 또는 polling 도입은 후속 sprint. |

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-c`). Phase 1–6 분석 결과 통합 + 도메인 그룹 13행 매트릭스 + Gap 요약 §5. |
| 2026-05-13 | 리뷰어 모드 2-pass: §3 의 CI/거버넌스 행을 PR 단위 산출 (PR #86 / #87 / #88 / 본 PR) 로 명세화. §4 ADR-0003 행에 ※ 노트 추가 — 본 PR 매트릭스가 PR #88 미머지 상태와 정합하지 않을 가능성 자체 인지. |
| 2026-05-13 | 후속 sprint `claude/work_260513-d`: ADR-0003 가 main 에 머지된 후 §4 ADR-0003 행을 정상 link 로 활성화 + §3 의 ※ 마킹 제거. §5.1 / §5.2 / §5.3 을 표 형식으로 통일 + 상태(open/closed) 컬럼 도입. §5.2 의 auth.spec.ts TC 미흡수 항목과 §5.3 의 frontend_integration_requirements §3.8 deprecation 항목을 closed 처리. 본 sprint 의 메타 헤더 표준화 commit 도 §3 CI/거버넌스 행에 추가. |
| 2026-05-13 | sprint `claude/work_260513-e` (A 묶음, M1 PR-D 정합 마무리): backend-core 의 `writeRBACServerError` → `writeServerError` 통합 (`internal/httpapi/rbac.go` helper 제거 + 11 호출 일괄 치환). `requireRequestID` 미들웨어에 caller-supplied X-Request-ID validation 추가 — 정규식 `^[A-Za-z0-9_-]{1,128}$`, 실패 시 server-generated fallback (work_260512-j 발견 항목 closed). request_id 를 표준 ctx key (`requestIDCtxKey{}`) 에도 stash + `requestIDFromContext` / `logRequestCtx` ctx-aware helper 추가, `kratos_login_client.go` 의 untraced `log.Printf` 2건 + `kratos_identity_resolver.go` 1건을 ctx-aware 로 치환 (logRequest 의 untraced fallback 해소 항목 closed). 단위테스트 11건 추가 (validation 양/음 경로 + ctx 전파 + logRequestCtx percent-safety + 미들웨어 e2e). |
