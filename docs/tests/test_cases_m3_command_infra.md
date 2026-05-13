# Test Cases — Command lifecycle + Infra topology E2E (M3 1차)

- 문서 목적: 매트릭스 §5.1 (E2E 미커버 도메인) 의 명령 lifecycle / 인프라 토폴로지 후보 TC 를 카탈로그화한다.
- 범위: TC 등재 + 단계 + 기대 결과 (spec 작성 가이드). 실제 `.spec.ts` 파일 작성은 후속 sprint 의 carve out — 본 문서는 카탈로그 1차.
- 대상 독자: E2E 작성자, QA, 프로젝트 리드.
- 상태: draft (spec ts 미작성, 카탈로그만)
- 최종 수정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-i`.
- 관련 문서: [추적성 매트릭스 §5.1](../traceability/report.md#51-e2e-미커버-도메인), [API 계약 §9 (command/audit)](../backend_api_contract.md#9-commandaudit-계약-초안), [API 계약 §6 (infra/dashboard)](../backend_api_contract.md#6-프론트-snapshot-api-1차).

---

## 1. 검증 대상 기능 (Feature Map)

| ID | 기능 | UI 경로 | 주요 액션 |
| --- | --- | --- | --- |
| **F-CMD-LIFECYCLE** | 명령 lifecycle (생성 → 상태 조회 → WS 갱신) | `/admin` 대시보드 + System Admin only | service action 생성 (dry-run) → command_id 수신 → `GET /api/v1/commands/:id` 또는 WebSocket `command.status.updated` 수신 → UI 상태 갱신 |
| **F-INFRA-TOPOLOGY** | 인프라 토폴로지 React Flow 렌더 + 인터랙션 | `/admin/infra` (또는 dashboard infra panel) | 정적 노드/엣지 렌더, 노드 클릭 시 상세 패널, 그룹 토글 |

---

## 2. 테스트 케이스 상세

### F-CMD-LIFECYCLE

#### TC-CMD-CREATE-01: service action 생성 및 command_id 수신
- **사전조건**:
  - System Admin 로그인 (예: charlie 시드).
  - 백엔드 `commandworker` running.
- **단계**:
  1. `/admin` 대시보드 진입.
  2. service action UI (예: "Restart Runner") 트리거 → `POST /api/v1/admin/service-actions` (API-15).
  3. 응답 `202 Accepted` + `data.command_id` 수신 확인.
  4. UI 상태 표시기 (toast / status panel) 가 "pending" 으로 렌더되는지 확인.
- **기대 결과**: command_id 가 UI 에 노출되고, 백엔드는 dry-run worker 가 `running` → `succeeded` 로 자동 전이.

#### TC-CMD-STATUS-01: command 상태 조회 (REST) + UI 갱신
- **사전조건**: TC-CMD-CREATE-01 직후, command_id 보존.
- **단계**:
  1. `GET /api/v1/commands/:command_id` (API-17) 호출 (또는 UI 가 자동 폴링).
  2. 응답 `data.command_status` 가 `succeeded` 또는 `running` 임을 확인.
  3. UI status 표시기가 동일 상태로 갱신되는지 확인.
- **기대 결과**: REST 폴링 / WebSocket 둘 중 하나로 UI 가 최종 상태 (`succeeded`) 에 도달.

### F-INFRA-TOPOLOGY

#### TC-INFRA-RENDER-01: 정적 노드/엣지 렌더
- **사전조건**: 임의 인증된 사용자 (모든 role 이 `infrastructure:view`).
- **단계**:
  1. `/admin/infra` (또는 dashboard infra panel) 진입.
  2. `GET /api/v1/infra/topology` (API-07 composite) 의 응답 데이터 기반으로 React Flow 가 노드 / 엣지를 렌더.
  3. DOM 에 노드 element 개수가 응답의 `nodes` 길이와 일치.
- **기대 결과**: 응답 데이터로 React Flow 캔버스가 정확히 렌더링.

#### TC-INFRA-NODE-CLICK-01: 노드 클릭 시 상세 패널
- **사전조건**: TC-INFRA-RENDER-01 통과 + 노드 1개 이상.
- **단계**:
  1. 임의 노드 클릭.
  2. 상세 패널 (modal / sidepanel) 이 열리고 노드의 `id`, `type`, `status`, `metadata` 가 표시되는지 확인.
- **기대 결과**: 상세 패널 노출 + 정확한 노드 정보 표시.

#### TC-INFRA-GROUP-TOGGLE-01: 그룹 노드 expand / collapse 토글
- **사전조건**: TC-INFRA-RENDER-01 통과 + 응답에 group 노드 (parent-child 관계) 1개 이상.
- **단계**:
  1. group 노드의 expand/collapse 컨트롤 클릭.
  2. child 노드들이 표시 / 숨김 처리되는지 확인.
- **기대 결과**: 토글 동작이 React Flow 의 child 노드 visibility 에 정확히 반영.

---

## 3. 결정 사항 및 고려 사항

- **권한 제어**: TC-CMD-CREATE-01 / STATUS-01 은 `infrastructure: create` 권한 (default 로 `system_admin` 만). TC-INFRA-* 는 `infrastructure: view` (모든 role).
- **dry-run 강제**: TC-CMD-CREATE-01 에서 `dry_run: true` 만 사용 — live action 은 별도 승인 flow 필요 (본 카탈로그 범위 밖).
- **WebSocket fallback**: TC-CMD-STATUS-01 에서 WebSocket 미연결 시 REST 폴링 fallback 동작 검증은 별도 TC-WS-RESILIENCE-* 후보 (M3 진입 후).
- **Empty state**: `GET /api/v1/infra/topology` 가 empty `{ nodes: [], edges: [] }` 반환 시 UI 가 "No infrastructure data" 같은 안내를 표시하는 별도 TC 후보.

## 4. spec ts 작성 carve out

본 문서는 카탈로그 (테스트 설계) 1차. 실제 `frontend/tests/e2e/command-lifecycle.spec.ts` / `infra-topology.spec.ts` 작성은 후속 sprint 의 별도 PR. 매트릭스 §5.1 의 P2 항목은 본 카탈로그 등재 시점에서 "카탈로그 closed, spec ts open" 으로 정합.

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-i`, C2). |
