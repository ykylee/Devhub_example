# Work Backlog — feat/work_260609-a-swagger-apikey-expand

- **문서 목적**: 본 sprint 의 todo 목록을 planned → in_progress → done 으로 추적한다.
- **sprint branch**: `feat/work_260609-a-swagger-apikey-expand`
- **최종 갱신일**: 2026-06-09 (session 진행 중)

## 상태 정의

- `planned` — 미착수
- `in_progress` — 작업 중
- `blocked` — 의존성/외부 입력 대기
- `done` — 검증 + 완료 확정

## 진행 중 sprint task (본 PR)

| ID | 상태 | 제목 | 비고 |
| --- | --- | --- | --- |
| T-cb09e16a | done | api_key 미들웨어 + DEVHUB_API_KEY env wire | config.go + auth.go + router.go + main.go + auth_test.go 5 신규 테스트 PASS |
| T-7f1cb4cf | done | voc + 외부 webhook + admin endpoints admin-only 마킹 | (의사결정 단계) — RBAC 가드 미적용 (carve a) |
| T-8a2d9651 | in_progress | openapi.yaml components.schemas 보강 (~30 schema) | 백그라운드 subagent (bg_a3fe74e2) 진행 중. 2612 lines 까지 확장 |
| T-ac17f679 | planned | openapi.yaml P0/P1 paths 보강 (30+ endpoints) | schema 완료 후 sequential |
| T-98120cdf | done | ADR-0029 작성 | docs/adr/0029-api-key-auth-and-swagger-scope.md |
| T-00316f81 | planned | 통합 테스트 + 검증 | go test ./... + yaml lint + curl simulation + lsp_diagnostics |

## carry-over (다음 sprint 후보)

| ID | 우선순위 | 제목 | 출처 |
| --- | --- | --- | --- |
| C-a (ADR-0029 §6 carve a) | P0 | API key caller 의 admin endpoint RBAC 가드 (`enforceRoutePermission` 에 `auth_source != "api_key"` 추가) | ADR-0029 §6 |
| C-b | P1 | API key rotation 정책 SOP (90일 + rolling pattern) | ADR-0029 §6 |
| C-c | P1 | openapi.yaml P2/P3 endpoint 30+ 확장 | ADR-0029 §6 / ADR-0027 §6 (d) |
| C-d | P2 | CI lint gate (openapi.yaml schema validity) | ADR-0029 §6 |
| C-e | P2 | swagger-ui system_admin 가드 미들웨어 | ADR-0027 §6 (c) |
| C-f | P3 | API key 다중 활성 (rolling) 지원 | ADR-0029 §6 |
| C-g | P2 | API key 사용 audit 강화 (callsite 별 X-Request-ID) | ADR-0029 §6 |
