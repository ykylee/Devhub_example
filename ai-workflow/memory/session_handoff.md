# Session Handoff — main (2026-05-13 EOD)

- 문서 목적: main 브랜치 기준 세션 상태와 다음 작업 진입점을 인계한다.
- 범위: 2026-05-13 머지 4건 (PR #86/#87/#88/#89) 정리 + 진행 중 sprint + 잔여 후보.
- 대상 독자: 후속 에이전트, 프로젝트 리드.
- 브랜치: `main` (HEAD `7fac5bf`, PR #89 squash 직후).
- 최종 수정일: 2026-05-13
- 상태: M1/M2 100% done. GitHub Actions CI 그린 + FU-CI-1..4 모두 처리. 거버넌스 + 추적성 체계 + 1차 종합 매트릭스 도입. 진행 중 sprint `claude/work_260513-d` 가 매트릭스 §5 gap + 문서 메타 헤더 표준화 처리 중.
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [상태 스냅샷](./state.json), [거버넌스](../../docs/governance/README.md), [추적성 매트릭스](../../docs/traceability/report.md).

## 0. 2026-05-13 머지 흐름

```
7fac5bf PR #89 — docs: governance + traceability 체계 도입 + 1차 종합 매트릭스 (claude/work_260513-c)
9268227 PR #88 — docs(adr): ADR-0003 no-docker policy CI scope — drop services: postgres + native PG 15 (claude/work_260513-b)
e86f38f PR #87 — ci: FU-CI-2/3/4 (playwright scope, GHA cache, frontend readiness) + main flat memory sync (claude/work_260513-a)
450cc24 PR #86 — ci: E2E 테스트 안정화 및 GitHub Actions CI 파이프라인 구축 (gemini/prepare-github-action)
```

세부:
- **PR #86** — GitHub Actions CI 1차 도입. backend-unit + frontend-unit + e2e 3잡. 리뷰어 모드 2-pass 로 5 blocker + 5 follow-on fix.
- **PR #87** — FU-CI-2/3/4 처리. playwright install scope (chromium 단일), `actions/cache@v4`, frontend readiness 120s, install split (cache-hit 시 browser install skip).
- **PR #88** — ADR-0003 (no-docker 정책 CI 범위 명문화). `services: postgres:15` sidecar 제거 + pgdg native PG 15 native 설치 step. PG-14 dropcluster + `--port=5432` 강제.
- **PR #89** — `docs/governance/` (README + document-standards.md) + `docs/traceability/` (README + conventions + sync-checklist + report) + PR template + AGENTS/GEMINI 진입점. 1차 종합 매트릭스: REQ-FR 105 + REQ-NFR 26 + ARCH 17 + API 40 + RM 28 + IMPL 79 + UT 47 + TC 37 = 412 항목.

## 1. 진행 중 sprint — `claude/work_260513-d`

PR #89 매트릭스의 §5 gap 분석 + `document-standards.md` §8 우선순위 1+2 문서 메타 헤더 표준화:

- 갭 §5.1 E2E 미커버: command lifecycle / realtime WebSocket / 인프라 토폴로지 / Webhook 처리 — 가능한 TC 후보 등재 + 우선순위.
- 갭 §5.2 ID 부재: backend-ai placeholder, frontend 컴포넌트 Vitest, auth.spec.ts TC 카탈로그 흡수 (직접 fix), RBAC API §12 매핑 정밀화.
- 갭 §5.3 문서↔코드: frontend_integration_requirements §3.8 deprecated 노트 (직접 fix), X-Devhub-Actor 폐기 일정 (후속), RBAC cache 일관성 (후속).
- 메타 헤더 표준화: 누락 4 + 변형 4 + frontend_integration_requirements 부분 보강. 매트릭스 ADR-0003 link 정상화.

## 2. 다음 진입점 — 우선순위 후보

`state.json#next_actions` 참조. 본 sprint 의 후속:

| 후보 | 주요 작업 | 규모 |
| --- | --- | --- |
| document-standards §8 우선순위 3 | 본문 ID 명기 (요구사항/설계 문서에 REQ-FR/ARCH/API 옆 backtick ID) | M |
| document-standards §8 우선순위 4 | deprecated 문서 식별 + 마킹 | S |
| E2E 신규 TC 작성 | TC-CMD-*, TC-INFRA-* — 매트릭스 §5.1 의 후보 구현 | M |
| frontend 컴포넌트 Vitest | Header, Sidebar, AuthGuard 등 | S |
| RBAC API §12 IMPL 정밀 매핑 | endpoint 별 IMPL ID 명시 | S |
| X-Devhub-Actor 폐기 ADR | architecture.md §6.2.3 의 완전 제거 trigger | S |
| RBAC cache 다중 인스턴스 일관성 | M1-DEFER-E | M-L |
| actionlint / workflow lint | ADR-0003 §6 의 후속 ADR | S |
| caller-supplied X-Request-ID validation | 정규식 강제 (work_260512-j 발견) | S |
| ctx 표준 request_id 전파 | client/helper 자동 tagged | M |
| writeRBACServerError → writeServerError 통합 | rbac.go:22 TODO | S |
| M4 진입 | command status WebSocket UI, AI Gardener gRPC, Gitea Hourly Pull worker | L |

## 3. 환경 / 운영 메모

- **CI 환경**: GitHub Actions ubuntu-24.04. native PostgreSQL 15 (pgdg) + native Ory Hydra/Kratos v26.2.0. `services:` 컨테이너 사용 안 함 (ADR-0003).
- **5 프로세스 native 기동** (prod / dev-server): PostgreSQL(시스템 서비스) + Hydra + Kratos + backend-core + frontend.
- **e2e 자동 시드**: `cd frontend && npm run e2e` 한 줄.

## 4. 잔여 결정 대기

- 본 sprint 후속 항목 (위 §2 표) 의 우선순위 결정.
- 운영 진입 전 hygiene: PoC `test/test` 시드 제거 (test-server-deployment.md §10).

## 5. 거버넌스 / 추적성 진입점

- `docs/governance/README.md` — 두 축 (문서 관리 + 추적성) 인덱스.
- `docs/governance/document-standards.md` — 문서 표준 (메타 헤더, lifecycle, 단계별 유형).
- `docs/traceability/conventions.md` — ID 컨벤션.
- `docs/traceability/sync-checklist.md` — 매 PR 동기화 절차.
- `docs/traceability/report.md` — 1차 종합 매트릭스.
- `.github/pull_request_template.md` — PR body 의 "추적성 영향" 섹션 자동 안내.
