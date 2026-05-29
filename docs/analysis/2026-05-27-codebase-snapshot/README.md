# 코드베이스 현황 분석 스냅샷 — 2026-05-27

- 문서 목적: 2026-05-27 시점 (main HEAD `cf19c94`, 마지막 기능 커밋 `99d6edc` / PR #373) DevHub 코드베이스의 전수 분석과 SDLC 자산 정합 점검, 향후 개발 방향 도출을 단일 폴더에 모은다.
- 범위: 백엔드(Go/Python) + 프론트엔드(Next.js) + 인프라/마이그레이션 + 테스트 + 문서(컨셉→요구사항→유스케이스→설계→구현→단위테스트→E2E) + 로드맵.
- 대상 독자: 프로젝트 리드, 백엔드/프론트엔드/인프라 워커, 후속 작업자, 외부 감사.
- 상태: accepted (분석 스냅샷)
- 작성 sprint: `claude/work_260527-codebase-review-roadmap-refresh`
- 기준 커밋: main `cf19c94` (PR #374 housekeeping), 마지막 기능 커밋 `99d6edc` (PR #373).

## 문서 구성

| # | 문서 | 내용 | 대응 요청 단계 |
| --- | --- | --- | --- |
| 00 | (본 README) | 인덱스 + 요약 지표 | — |
| 01 | [`01_codebase_state_analysis.md`](./01_codebase_state_analysis.md) | 코드베이스 전수 상태 분석 (백엔드/프론트/인프라/테스트) | 1 |
| 02 | [`02_sdlc_chain_status.md`](./02_sdlc_chain_status.md) | 컨셉→요구사항→유스케이스→설계→구현→단위테스트→E2E 구성 현황 + 갭 | 2 |
| 03 | [`03_frontend_summary.md`](./03_frontend_summary.md) | 프론트엔드 개발 사항 정리 | 5 |
| 04 | [`04_backend_summary.md`](./04_backend_summary.md) | 백엔드 개발 사항 정리 | 6 |
| 05 | [`05_fe_be_balance.md`](./05_fe_be_balance.md) | 프론트엔드 ↔ 백엔드 개발 균형 분석 | 7 |
| 06 | [`06_future_direction.md`](./06_future_direction.md) | 향후 개발 방향 + 도출 아이템 (백로그) | 8 |

### 항목별 상세 분석 (code/ + docs-review/)

사용자 지시("각 항목 별로 상세 분석해서 항목 별 별도 문서로 작성 + 분석 토대로 각 항목 갱신")에 따른 항목별 산출물.

**코드 항목 분석** (`code/`):

| 문서 | 대상 |
| --- | --- |
| [`code/backend/httpapi.md`](./code/backend/httpapi.md) | 핸들러 42 + router + permissions + 인증 미들웨어 체인 + 라우트 전수표 |
| [`code/backend/store.md`](./code/backend/store.md) | PostgresStore 도메인별 메서드 + 트랜잭션 패턴 + in-memory fake parity |
| [`code/backend/domain.md`](./code/backend/domain.md) | aggregate/enum/상태머신/검증 helper |
| [`code/backend/migrations.md`](./code/backend/migrations.md) | 000001..000045 + up/down 정합 |
| [`code/backend/support_packages.md`](./code/backend/support_packages.md) | gitea/audit/auth/commandworker/serviceaction/config/devrequest/normalize/hrdb/adapters/main.go/backend-ai |
| [`code/frontend/pages.md`](./code/frontend/pages.md) | 33 page + layout |
| [`code/frontend/components.md`](./code/frontend/components.md) | 50 컴포넌트 |
| [`code/frontend/services.md`](./code/frontend/services.md) | 18 service + api-client/realtime/websocket |
| [`code/frontend/frontend_platform.md`](./code/frontend/frontend_platform.md) | auth/config/storage/store + 테스트 + 스택 |
| [`code/infra/infra_ci.md`](./code/infra/infra_ci.md) | CI/nginx/Keycloak SPI/Prometheus/proto |
| **[`code/code_findings_backlog.md`](./code/code_findings_backlog.md)** | **위 분석에서 도출된 통합 발견 백로그 (H/M/C/L/I 우선순위 + N/X/E 매핑)** |

**개발문서 항목 분석** (`docs-review/`) — 각 문서의 현재 상태·누락·불일치 분석 + 적용한 갱신 로그:

| 문서 | 대상 (분석 + 갱신 적용) |
| --- | --- |
| [`docs-review/requirements_analysis.md`](./docs-review/requirements_analysis.md) | requirements.md → §5.8 REQ-FR-REPO + §5.6 REQ-FR-INT-013..015 + §2.5 정정 배너 |
| [`docs-review/architecture_analysis.md`](./docs-review/architecture_analysis.md) | architecture.md → §10 ARCH-REPO + §8.7 ARCH-INT-07 + §6.5 |
| [`docs-review/usecases_analysis.md`](./docs-review/usecases_analysis.md) | system_usecases.md → §2.14 UC-REPO + §2.12 UC-INT-15..18 |
| [`docs-review/api_contract_analysis.md`](./docs-review/api_contract_analysis.md) | backend_api_contract.md → API-91/92 draft/publish + API-70 env fallback 정정 |
| [`docs-review/roadmaps_analysis.md`](./docs-review/roadmaps_analysis.md) | backend/frontend 세부 로드맵 stale 정정 (Hydra/accounts/Phase) |
| [`docs-review/adr_planning_analysis.md`](./docs-review/adr_planning_analysis.md) | ADR 24 + planning 33 정합 점검 + 갱신 권고 |

본 분석을 토대로 갱신된 외부 문서 (요청 단계 3·4·8):

- [추적성 매트릭스](../../traceability/report.md) — ADR-0022/0023 인덱스 보강 + API-87..90 + migration 000034..000045 + repository draft/publish·SCM 통합 도메인 행 추가 + 헤더/범위/날짜 갱신.
- [통합 개발 로드맵](../../development_roadmap.md) — M5/M6/M7 closing 사실 정합 + 변경 이력 보강.
- [v1.0 릴리즈 로드맵](../../planning/release_v1_roadmap.md) — Onboarding 전 carve 완료 + 2026-05-26/27 신규 작업(Gitea·SCM 양방향·project v2·auth_mode) 반영 + 향후 carve 재정렬.
- [프론트엔드 세부 로드맵](../../frontend_development_roadmap.md) / [백엔드 세부 로드맵](../../../docs/backend_development_roadmap.md) — stale 상태(Hydra/Kratos·`/api/v1/accounts/*`) 정정 + Phase 현황 갱신.

## 핵심 요약 지표 (2026-05-27)

| 영역 | 지표 | 값 |
| --- | --- | --- |
| 백엔드 (Go) | 구현 파일 (non-test) | 89 |
| 백엔드 (Go) | 단위/통합 테스트 파일 | 72 |
| 백엔드 (Go) | HTTP 라우트 (대략) | ~100 |
| 백엔드 (Go) | 마이그레이션 (up) | 45 (000001..000045) |
| 백엔드 (Python) | backend-ai | 스켈레톤 (`/health` 만, gRPC TODO) |
| 프론트엔드 | 페이지 (`page.tsx`) | 33 |
| 프론트엔드 | 컴포넌트 (`.tsx`, non-test) | 50 |
| 프론트엔드 | 서비스 (`*.service.ts`) | 18 |
| 프론트엔드 | 단위 테스트 (Vitest) | 10 파일 |
| 프론트엔드 | E2E 스펙 (Playwright) | 28 |
| 설계 | ADR | 24 (0001..0024) |
| 누적 | 머지 PR | 374 |

## 한눈 결론

- **백엔드는 v1.0 scope 기준 기능 완성** 상태에 도달했고, 2026-05-26~27 에는 v1.0 scope 를 넘어선 **Gitea SCM 동기화 + 외부 연동(provider/binding) 깊이 + SCM↔시스템 repository 양방향 연동**으로 확장 중이다.
- **프론트엔드는 운영 UI 전환(mock 제거)** 을 1차 완료하고 admin catalog/onboarding/topology v2 까지 따라왔으나, **테스트(특히 단위)와 신규 backend 기능의 UI 노출 깊이에서 backend 에 뒤처진다**.
- **문서(추적성·로드맵)는 2026-05-16~21 시점에 정지**해 있어, 그 이후 약 1주 + 100+ PR 의 작업(Onboarding IMPL 완료, ADR-0022/0023/0024, Gitea/SCM 통합, project v2)이 반영되지 않은 **drift** 가 가장 큰 부채다 — 본 PR 이 이를 정합화한다.
