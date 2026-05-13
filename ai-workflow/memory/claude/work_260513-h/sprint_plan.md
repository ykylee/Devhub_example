# Sprint Plan — claude/work_260513-h (B4: X-Devhub-Actor 폐기 ADR)

- 문서 목적: ADR-0004 발급으로 X-Devhub-Actor 헤더 폐기 사실의 ex-post-facto 명문화. 매트릭스 §5.3 "X-Devhub-Actor 완전 제거 trigger" open 항목 closing.
- 범위: 신규 ADR-0004 + 기존 문서 잔재 정리 + me.go 주석 1줄. 코드 동작 변경 0 (SEC-4 가 이미 처리).
- 진입 base: main HEAD `594be74` (PR #93 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. 현황 분석

### 코드 (이미 처리됨)

- `auth.go::authenticateActor` 가 X-Devhub-Actor 헤더를 처리하지 않음. M0 SEC-4 resolution 으로 이미 제거.
- 테스트 4 파일 (audit_test.go, commands_test.go, auth_test.go, me_test.go) 이 "X-Devhub-Actor must be ignored" SEC-4 회귀 방지 검증 — 본 sprint 보존.

### 문서 (잔재)

| 위치 | 잔재 |
| --- | --- |
| `docs/architecture.md` line 174 | "X-Devhub-Actor 헤더는 폐기 예정 폴백으로만 유지합니다 (폐기 시점은 ADR-0001 §8 미해결 항목)." — 이미 제거됐는데 "유지" 라고 적혀 있음 |
| `docs/adr/0001-idp-selection.md` §8 #4 | "Phase 13 P1 7단계 도입 시점에 X-Devhub-Actor 폴백은 유지하되 deprecation warning 로그를 남긴다. ... 별도 phase 에서 완전 제거." — trigger 충족 사실 미반영 |
| `backend-core/internal/httpapi/me.go` line 16 | 주석에 "or the dev fallback is active and no X-Devhub-Actor was supplied" — X-Devhub-Actor 언급 잔재 |

## 2. ADR-0004 의 결정

- **컨텍스트**: ADR-0001 §8 #4 가 X-Devhub-Actor 폐기를 "Bearer token verifier + 모든 backend handler 가 token 경로로 actor 도출되도록 전환된 후 별도 phase 에서 완전 제거" 로 결정.
- **trigger 충족 사실**: SEC-4 (M0, 2026-05-07/08) 에서 prod 코드의 X-Devhub-Actor 처리 자체를 제거. M1 의 Bearer token verifier 도입 (BearerTokenVerifier interface + authenticateActor middleware) 으로 actor 도출 경로 표준화. SEC-4 가 별도 phase 없이 즉시 완전 제거 path 를 채택한 것.
- **결정 (2026-05-13)**: X-Devhub-Actor 폐기 완료 선언. ADR-0001 §8 #4 의 trigger 가 SEC-4 시점에 이미 충족됐고 본 ADR 이 그 사실의 ex-post-facto 명문화.
- **Consequences**:
  - 회귀 방지 테스트 (X-Devhub-Actor → 401 / actor=system) 유지 — security regression 방지.
  - architecture.md / ADR-0001 §8 #4 / me.go 주석 잔재 정리.
  - 새 코드는 X-Devhub-Actor 헤더에 대한 어떤 처리도 추가하지 않는다.

## 3. 작업 항목

| 항목 | 위치 | 결과 |
| --- | --- | --- |
| ADR-0004 발급 | `docs/adr/0004-x-devhub-actor-removal.md` | 신규 ADR (표준 §1~§7 양식) |
| architecture.md 잔재 정리 | `docs/architecture.md` §6.2.X | line 174 갱신 — ADR-0004 참조로 |
| ADR-0001 §8 #4 인라인 갱신 | `docs/adr/0001-idp-selection.md` §8 #4 | "후속 결정 (2026-05-13, ADR-0004): trigger 충족, closed" 추가 |
| me.go 주석 정리 | `backend-core/internal/httpapi/me.go` line 16 | "X-Devhub-Actor" 언급 제거 |
| 매트릭스 §4 ADR 인덱스 | `docs/traceability/report.md` §4 | ADR-0004 행 추가 |
| 매트릭스 §5.3 closed | `docs/traceability/report.md` §5.3 | "X-Devhub-Actor 완전 제거 trigger" → closed |
| 매트릭스 §6 변경 이력 | `docs/traceability/report.md` §6 | 본 sprint 한 줄 |
| main flat memory sync | `ai-workflow/memory/*` | PR #93 흡수 + 본 sprint IN PROGRESS |

## 4. 검증

- 코드 동작 변경 0 (me.go 주석만 변경). `go test ./...` PASS.
- frontend 영향 0.

## 5. 미진입 / 다음 sprint 후보

- B1 추가 도메인 (account / org / command / audit / infra) 본문 ID 노출
- B2 deprecated 문서 식별 + 마킹
- C1 frontend 컴포넌트 Vitest
- C2 E2E 신규 TC
- D5 actionlint
- M3 / M4 진입
