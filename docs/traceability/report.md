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
| **CI / 거버넌스 (본 sprint 산출)** | NFR-1 (no-docker) | (ADR-0001 §5; ADR-0003) | M2-16 (CI 1차); 본 sprint (FU-CI-1..4 + 거버넌스) | (build infra: `.github/workflows/ci.yml`, `scripts/ci-setup.sh`, `infra/idp/*.ci.yaml`); `docs/governance/*`, `docs/traceability/*` | (lint 미도입, FU-CI 향후) | (CI run 자체가 검증) |
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

다음 도메인은 현재 E2E TC 가 없거나 매우 적음 — 후속 sprint 에서 보완 후보:

- **명령 lifecycle / mitigation**: 단위테스트는 있으나 e2e UI 흐름 (대시보드 → 명령 실행 → 결과 확인) 미커버.
- **실시간 (WebSocket)**: backend infra/ci/risk event publish 가 M3 planned, e2e 자연 후속.
- **인프라 토폴로지**: React Flow UI 의 인터랙션 (zoom, pan, group) e2e 미커버.
- **Webhook 처리**: HMAC 서명 검증 + idempotency 는 단위테스트로 검증, 외부 영향 e2e 어려움.

### 5.2 ID 부재 / 매핑 누락 (분석 중 발견)

- **Backend AI (`backend-ai/`)**: 현재 placeholder 만 존재, 추적 항목 미부여. M3-04 AnalysisService gRPC client 진입 시 IMPL-ai-XX 발급 필요.
- **Frontend 컴포넌트 테스트**: 다수 모듈 (Header, Sidebar, AuthGuard 등) 의 Vitest 단위테스트 부재.
- **`auth.spec.ts` 의 TC-AUTH-01..06**: spec 파일 안에만 정의, `test_cases_m2_auth.md` 의 TC 카탈로그에 흡수되지 않음. 후속 hygiene.
- **API §12 RBAC 정책 편집 API 구현 상태**: ADR-0002 결정 + 6 PR 분할 머지되었으나 매트릭스의 IMPL-rbac-* 가 일부 endpoint 만 cover. 정밀 매핑 보강 후보.

### 5.3 문서 ↔ 코드 불일치 (분석 중 발견)

- **ADR-0001 vs frontend_integration_requirements §3.8**: Phase 13 의 자체 accounts 테이블 endpoint 7종이 frontend_integration_requirements 에 그대로 기술되어 있어 ADR-0001 Ory 도입과 불일치. 후속 정리 sprint.
- **X-Devhub-Actor 폐기 일정**: ARCH-13 의 deprecation warning 경로는 `architecture.md` §6.2.3 에 명시되었으나 완전 제거 trigger 미정의.
- **RBAC cache 다중 인스턴스 일관성**: ARCH-13 §12.10 의 미해결 항목. M1-DEFER-E.

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-c`). Phase 1–6 분석 결과 통합 + 도메인 그룹 13행 매트릭스 + Gap 요약 §5. |
