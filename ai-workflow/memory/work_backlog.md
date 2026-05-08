# Integrated Work Backlog (main, post M0)

- 문서 목적: main 브랜치 기준 상위 백로그 인덱스. 세부 sprint backlog 는 브랜치별 메모리 디렉터리 참조.
- 상태: M0 종료, M1 대기
- 최종 수정일: 2026-05-08
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [세션 인계](./session_handoff.md), [상태 스냅샷](./state.json)

## 1. 마일스톤 진행 상황

| 마일스톤 | 상태 | 종료 일자 | 메모 |
| --- | --- | --- | --- |
| **M0** — 보안 게이트 통과 | ✅ done | 2026-05-08 | PR #14·15·16·17·18·19. SEC-1~4 resolved. T-M0-10 운영 검증 PASS. |
| **M1** — 핵심 기능 contract 정합성 | 🟡 in_progress | — | sprint planning branch `claude/m1-sprint-plan`, backlog 초안 작성 완료 (T-M1-01..08). 진입 PR 순서: PR-A SEC-5 → PR-F RBAC ADR → PR-B contract → PR-C cmd lifecycle → PR-D audit. |
| **M2** — 사용자 경험 정합 (Phase 4·5 잔여 + Phase 6/6.1) | planned | — | 통합 로드맵 §3.3. M0 의 잔여 (`/auth/callback`, `account.service.ts`) 도 흡수. |
| **M3** — Realtime 확장 + 외부 연동 1차 | planned | — | 통합 로드맵 §3.4. |
| **M4** — 운영 / SSO / MFA / 후속 ADR | planned | — | 통합 로드맵 §3.5. ADR-0002 (Gitea SSO) 등. |

## 2. 최근 머지 (M0 sprint)

| PR | 제목 | merge commit |
| --- | --- | --- |
| #14 | docs(roadmap): integrated development roadmap + M0 sprint planning | `427d618` |
| #15 | fix(auth): close empty-Auth pass-through, gate X-Devhub-Actor (PR-A) | `5a2fec1` |
| #16 | feat(authz): role guard middleware + protected route mapping (PR-B) | `a477ca3` |
| #17 | feat(auth): Hydra introspection verifier + prod fail-fast (PR-C) | `21cd24a` |
| #18 | feat(auth): replace mock login with /api/v1/me + OIDC redirect (PR-D) | `cf2d55f` |
| #19 | feat(auth): close SEC-4 + finish M0 sprint marker sweep (PR-E) | `0ea0ce4` |

## 3. M1 진입 후보 작업

다음 sprint 의 작업 분해는 진입 시점에 브랜치별 메모리에서 수행. 현 시점 후보 일람:

### 3.1 P0~P1 must-have

- **SEC-5 — DB 에러 노출 마스킹**: `organization.go`, `commands.go`, `audit.go` 등 5xx 응답 본문을 일반화하는 `writeServerError` 헬퍼 도입. SEC-5 단독 cleanup PR.
- **API 계약 정합**: 모든 핸들러 응답 envelope/필드/role wire 가 `docs/backend_api_contract.md` 와 일치하도록 검증·정정.
- **command lifecycle**: 6 상태값(`pending|running|succeeded|failed|rejected|cancelled`) 일관 적용 + dry-run/live 경계 테스트.
- **Audit actor 보강**: `source_ip`, `request_id`, `source_type` 필드 추가.

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
