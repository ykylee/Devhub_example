# Integrated Work Backlog (main, post M1 RBAC)

- 문서 목적: main 브랜치 기준 상위 백로그 인덱스. 세부 sprint backlog 는 브랜치별 메모리 디렉터리 참조.
- 범위: 마일스톤 상태, 최근 머지, 잔여/후속 작업
- 대상 독자: 프로젝트 리드, 후속 에이전트, 트랙 담당자
- 상태: M1 RBAC track (PR-A·F·G1~G6) 머지 완료. M1 잔여 (PR-B/C/D) + DEFER 항목 대기.
- 최종 수정일: 2026-05-08
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [세션 인계](./session_handoff.md), [상태 스냅샷](./state.json), [M1 PR 리뷰 actions](./M1-PR-review-actions.md)

## 1. 마일스톤 진행 상황

| 마일스톤 | 상태 | 종료 일자 | 메모 |
| --- | --- | --- | --- |
| **M0** — 보안 게이트 통과 | ✅ done | 2026-05-08 | PR #14·15·16·17·18·19. SEC-1~4 resolved. T-M0-10 운영 검증 PASS. |
| **M1** — 핵심 기능 contract 정합성 | 🟡 in_progress (RBAC track 완료) | — | RBAC track (T-M1-01·06·07·08 흡수) 완료: PR #20·21·22·23·29·30·31·27. 잔여: T-M1-02·03·04·05 (envelope/types/ws + cmd lifecycle + audit actor + auth_test 보강) → PR-B/C/D. DEFER A~G 는 §3 후속 작업으로 인계. |
| **M2** — 사용자 경험 정합 (Phase 4·5 잔여 + Phase 6/6.1) | 🟡 in_progress (login_action sprint) | — | 로그인 검토 결과 OIDC code flow frontend 4단계 + backend Kratos/Hydra proxy 부재. PR-LOGIN-1 (#33 backend) + PR-LOGIN-2 (#34 frontend form) push 머지 대기. PR-LOGIN-3 (callback + token + httpClient) + PR-LOGIN-4 (logout + /account) 미진입. backlog: `claude/login_action/backlog/2026-05-08.md`. |
| **M3** — Realtime 확장 + 외부 연동 1차 | planned | — | 통합 로드맵 §3.4. |
| **M4** — 운영 / SSO / MFA / 후속 ADR | planned | — | 통합 로드맵 §3.5. ADR-0002 (Gitea SSO) 등. |

## 2. 최근 머지 (M1 RBAC track, 2026-05-08)

| PR | 제목 | 머지 |
| --- | --- | --- |
| #20 | feat(http): SEC-5 mask 5xx errors + M1 sprint plan (PR-A) | `ae8aca1` |
| #21 | docs(adr): ADR-0002 RBAC policy edit API (PR-F) | `950a11f` |
| #22 | docs(api): RBAC §12 rewrite + route mapping (PR-G1) | `1a090a3` |
| #23 | feat(rbac): domain + rbac_policies migration (PR-G2) | `5239a87` |
| #29 | feat(rbac): postgres store + users.role FK (PR-G3 + FIX-A) | `24815b8` |
| #30 | feat(rbac): RBAC handlers (PR-G4 + FIX-B + FIX-C) | `27b6817` |
| #31 | feat(rbac): permission cache + enforcement (PR-G5) | `02eef35` |
| #27 | feat(frontend): RBAC PermissionEditor ↔ backend (PR-G6 + FIX-D) | `e02ba67` |
| #28 | docs(memory): M1 PR review actions tracker | `9bc30c9` |

원본 PR #24, #25, #26 은 stack base 자동 삭제로 close 후 main 위에서 #29/#30/#31 로 재등록.

## 3. M1 잔여 + DEFER 후속 작업

### 3.1 M1 잔여 (P1, 진입 시 분해)

- **T-M1-02 (PR-B)** — API envelope/role wire/필드 정합 (`backend_api_contract.md` ↔ 코드 1:1 강제, 통합 테스트 매트릭스).
- **T-M1-03 (PR-C)** — command lifecycle 6 상태 일관 적용 + dry-run/live 경계 테스트.
- **T-M1-04 (PR-D)** — Audit actor 보강 (`source_ip`, `request_id`, `source_type` + 마이그레이션 + 미들웨어 + 응답 헤더 `X-Request-ID`).
- **T-M1-05** — `auth_test.go` prod 가드 + role 가드 통합 테스트 매트릭스. PR-C 또는 PR-D 에 흡수 가능.
- **T-M1-07 (frontend)** — `frontend/lib/services/types.ts` UI vs wire 분리, 표시 포맷 프론트 이전. PR-B 에 묶거나 단독.
- **T-M1-08** — WebSocket envelope `{schema_version, type, event_id, occurred_at, data}` 코드/문서 정합. PR-B 에 묶거나 단독.

### 3.2 DEFER (M1 PR 리뷰 후속, [상세](./M1-PR-review-actions.md#3-다음-개발로-넘김--defer))

- **M1-DEFER-A** — `rbac_policies` 의 `is_system` ↔ `role_id` 일관성 CHECK 제약 (P2 방어선)
- **M1-DEFER-B** — `requireMinRole`/`roleMeetsMin`/`roleRank` deadcode 정리 + 단위 테스트 제거
- **M1-DEFER-C** — `writeRBACServerError` 임시 helper → `writeServerError` 통합 (PR-G4 의 TODO)
- **M1-DEFER-D** — `DeleteRBACRole` row-lock (다중 인스턴스 race 강화)
- **M1-DEFER-E** — `PermissionCache` 다중 인스턴스 일관성 (pub/sub 또는 polling)
- **M1-DEFER-F** — API contract §12.4 / §12.5 응답 예시 추가
- **M1-DEFER-G** — `MemberTable` role display 회귀 사용자 환경 검증

### 3.2 P1~P2 should-have

- **frontend `/auth/callback`**: Hydra authorization_code → `/oauth2/token` 교환 후 세션 저장 흐름. PR-D 의 자연 후속.
- **frontend `account.service.ts`**: Kratos public flow 호출 helper.
- **frontend `types.ts` 분리**: UI 표시명 vs API wire 타입 분리, 표시 포맷팅을 프론트로 이전.
- **WebSocket envelope 표준화**: `{schema_version, type, event_id, occurred_at, data}` 코드/문서 정합.
- **RBAC policy 편집 API**: 12.x — write/audit 경계 + persistence 또는 *static-default 유지* 결정.

### 3.3 P3 nice-to-have

- ADR-0002 (Gitea SSO 통합) 작성.
- ADR (commits 정규화 테이블 도입 여부).
- ADR (OS service wrapper 결정).

## 4. 운영 / 환경 메모

- Hydra/Kratos PoC: 본 세션에서 e2e 검증 완료. binary 위치 `$env:USERPROFILE\go\bin\` (Windows). 다음 세션에서 재가동은 `hydra serve all --config infra/idp/hydra.yaml --dev` + `kratos serve --config infra/idp/kratos.yaml`.
- 검증용 임시 OIDC client (client_credentials, id `43aa4b74-...`) 는 Hydra 안에 잔존. 보안 위험은 없으나 cleanup 가능.
- backend-core `go test ./...` 는 사내 GoProxy mirror 환경에서 PASS 검증됨.

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-07 | PR #12 통합 sprint 종료 (BLK/SEC/HYG 트래커 등록) |
| 2026-05-08 | PR #13~#19 머지. 통합 로드맵 채택 + M0 sprint 종료. flat memory 갱신. |
| 2026-05-08 | M1 sprint 진입. `claude/m1-sprint-plan` 브랜치 + backlog 초안. 진입 PR 5건 분할 결정. |
| 2026-05-08 | M1 RBAC track 8 PR (#20·21·22·23·29·30·31·27) + 리뷰 트래커 (#28) 머지. FIX A~D 적용. DEFER A~G 백로그 인계. |
| 2026-05-08 | M2 login_action sprint 진입. 로그인 검토 후 PR-LOGIN-1 (#33 backend proxy) + PR-LOGIN-2 (#34 frontend form) push 머지 대기. backlog 작성 (`claude/login_action/backlog/2026-05-08.md`). PR-LOGIN-3·4 미진입. |
