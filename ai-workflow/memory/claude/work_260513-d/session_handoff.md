# Session Handoff — claude/work_260513-d (2026-05-13 진행 중)

- 문서 목적: 추적성 갭 분석 + 문서 재정리 sprint 인계.
- 범위: PR #87/#88/#89 머지 후속 정리 + 매트릭스 §5 gap 처리 + 메타 헤더 표준화.
- 최종 수정일: 2026-05-13
- 상태: in_progress. 브랜치 분기 직후.

## 1. 진입 컨텍스트

- 직전 sprint × 3: work_260513-a (PR #87 GHA 최적화), work_260513-b (PR #88 ADR-0003 + native PG 15), work_260513-c (PR #89 거버넌스 + 추적성 + 1차 매트릭스). 모두 머지 완료, main HEAD = `7fac5bf`.
- 본 sprint 는 PR #89 의 `docs/traceability/report.md` §5 gap 항목 + `docs/governance/document-standards.md` §8 우선순위 1+2 (메타 헤더) 를 다룬다.

## 2. 작업 흐름

### 2-1. cleanup

- main flat memory (`ai-workflow/memory/{state.json, session_handoff.md, work_backlog.md}`) 가 PR #87 의 commit (state.json head_commit 450cc24) 까지만 반영. PR #87/#88/#89 누적 sync 필요.
- 매트릭스 §4 의 ADR-0003 가 plain text + ※ — ADR-0003 가 이제 main 에 있으므로 정상 link 로 활성화 가능 (§3 의 ※ 도 함께 정리).

### 2-2. 갭 분석 §5.1 E2E 미커버

| 도메인 | 현황 | 가능한 TC 후보 |
| --- | --- | --- |
| 명령 lifecycle / mitigation | 단위테스트만, e2e UI 흐름 미커버 | TC-CMD-CREATE (대시보드 → service-action 생성 → command_id), TC-CMD-STATUS (상태 조회 → WebSocket 이벤트 수신 → UI 갱신) |
| 실시간 WebSocket | M3 planned. command.status.updated 만 publish | M3 진입 시 e2e — TC-WS-CONNECT, TC-WS-CMD-STATUS, TC-WS-RESILIENCE (re-connect) |
| 인프라 토폴로지 React Flow | 정적 데이터 렌더 | TC-INFRA-RENDER (정적), TC-INFRA-NODE-CLICK (상세 패널), TC-INFRA-GROUP-TOGGLE — UI 인터랙션 |
| Webhook 처리 | 단위테스트로 검증, e2e 외부 영향 어려움 | 후보 없음 — 단위테스트 + 통합테스트 (별도 sprint) 로 충분 |

### 2-3. 갭 분석 §5.2 ID 부재

| 항목 | 처리 |
| --- | --- |
| backend-ai placeholder | 추적 항목 미부여 — M3-04 진입 시 IMPL-ai-XX 발급. 본 sprint 는 매트릭스에 placeholder 행 표기만. |
| frontend 컴포넌트 Vitest 부재 | Header, Sidebar, AuthGuard 등 — 후속 sprint 후보, 우선순위 P2. |
| auth.spec.ts TC-AUTH-01..06 카탈로그 미흡수 | 본 sprint 에서 `docs/tests/test_cases_m2_auth.md` 에 TC-AUTH-01..06 행 추가 — 본문 spec 과 일치하는 TC 명세 등재. |
| RBAC API §12 IMPL 매핑 정밀화 | report.md §2.4 / §3 의 IMPL-rbac-* 가 일부 endpoint 만 cover. 정밀 매핑 보강 (endpoint 별 IMPL ID 명시). |

### 2-4. 갭 분석 §5.3 문서↔코드 불일치

| 항목 | 처리 |
| --- | --- |
| ADR-0001 vs frontend_integration_requirements §3.8 | frontend_integration_requirements §3.8 가 자체 accounts 테이블 endpoint 7종 명시 — ADR-0001 (Ory 도입) 와 불일치. 해당 문서 §3.8 에 deprecated 노트 추가 + ADR-0001 + API §11 (실제 Ory proxy endpoint) 로 redirect. |
| X-Devhub-Actor 폐기 일정 | architecture.md §6.2.3 의 deprecation warning 경로 + 완전 제거 trigger 미정의. 본 sprint 의 결정 없음 — work_backlog 에 별도 결정 후보로 등재. |
| RBAC cache 다중 인스턴스 일관성 | ARCH-13 §12.10 의 미해결 — M1-DEFER-E. 본 sprint 처리 X, work_backlog 에 명시. |

### 2-5. 문서 재정리 — 메타 헤더 표준화

**완전 누락 4 문서** (docs/governance/document-standards.md §2 표준 양식 일괄 적용):
- docs/backend/requirements.md
- docs/backend_requirements_org_hierarchy.md
- docs/org_chart_ux_spec.md
- docs/organizational_hierarchy_spec.md

**변형 4 문서** (누락 항목 보강):
- docs/requirements.md — 목적/범위/대상 독자/관련 문서 추가
- docs/architecture.md — 목적/범위/대상 독자 추가
- docs/backend_api_contract.md — 범위/대상 독자 추가
- docs/tech_stack.md — 목적/범위/대상 독자/최종 수정일 추가

`docs/backend/frontend_integration_requirements.md` — 일부 보강 (범위/대상 독자/최종 수정일).

## 3. 다음 행동

1. branch memory 초기화 (본 파일 + state.json + work_backlog.md).
2. main flat memory sync.
3. 매트릭스 ADR-0003 cleanup.
4. 메타 헤더 표준화 (8+ 문서).
5. 매트릭스 §5 의 처리 결과 반영 (gap 해소 항목 close, 미해소 후속 등재).
6. PR + 2-pass + 머지.
