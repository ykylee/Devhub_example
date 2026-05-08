# Session Handoff — main (post M0 sprint)

- 브랜치: `main`
- HEAD: `0ea0ce4` (Merge pull request #19)
- 최종 수정일: 2026-05-08
- 상태: M0 sprint 코드/검증 종료. M1 진입 대기.
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [PR-12 액션 트래커](./PR-12-review-actions.md), [M0 sprint 기록](./claude/merge_roadmap/), [상태 스냅샷](./state.json), [상위 backlog](./work_backlog.md)

## 1. 본 세션 활동 요약 (2026-05-08)

PR #12·#13 머지 직후 시점에 시작해 통합 로드맵 채택 → M0 sprint 5 PR 발행·머지 → 운영 검증까지 완료.

| Sub-phase | 결과 |
| --- | --- |
| PR #12 종합 리뷰 → BLK/SEC/HYG 액션 도출 | resolved |
| `source-docs/workflow-source/**` PR #13 분리 | merged |
| 통합 개발 로드맵 작성 (`docs/development_roadmap.md`) | PR #14 merged |
| M0 sprint backlog (T-M0-01..11) | PR #14 |
| **PR-A** SEC-2 부분 + SEC-4 dev-only gate (#15) | merged |
| **PR-B** SEC-3 role guard 미들웨어 + 라우트 (#16) | merged |
| **PR-C** SEC-2 Hydra introspection verifier + prod fail-fast (#17) | merged |
| **PR-D** SEC-1 frontend AuthGuard + `/api/v1/me` (#18) | merged |
| **PR-E** SEC-4 fallback path 제거 + M0 마무리 (#19) | merged |
| T-M0-10 운영 검증 (Hydra/Kratos PoC e2e + valid JWT → 200) | PASS |

## 2. SEC 결함 종합 (M0 종료 시점)

| SEC | 상태 |
| --- | --- |
| SEC-1 frontend mock auth | ✅ resolved (#18) |
| SEC-2 verifier nil + empty Authorization | ✅ resolved (#15 + #17) |
| SEC-3 role 미적용 | ✅ resolved (#16) |
| SEC-4 X-Devhub-Actor fallback | ✅ resolved (#15 dev gate + #19 path 제거) |
| SEC-5 DB 에러 메시지 노출 | tracked → M1 별도 PR |

코드 영역 `// SECURITY (SEC-*)` 마커 0건. 운영 환경에서 Hydra introspection 사이클 실 검증 PASS.

## 3. 운영 환경 정리

- 본 sandbox 의 Hydra (`:4444/:4445`), Kratos (`:4433/:4434`), backend-core (`:8080`) 백그라운드 프로세스 모두 종료.
- DB 의 `hydra` / `kratos` schema 와 마이그레이션은 그대로 유지 (다음 세션 재가동 시 즉시 사용 가능).
- 검증용 임시 OIDC client (`43aa4b74-...`, client_credentials) 는 Hydra 내부에 잔존. 다음 세션에서 정리하거나 무시 (`hydra delete oauth2-client --endpoint http://localhost:4445 <id>`).

## 4. 다음 세션 진입점

### 4.1 우선순위 1 — M1 sprint (통합 로드맵 §3.2)

| Track | 작업 | 비고 |
| --- | --- | --- |
| B·X | API envelope/필드/role wire format 일관 (`backend_api_contract.md` 정합) | M1 DoD #1 |
| B | command lifecycle 6 상태 일관 적용 + dry-run/live 경계 테스트 | M1 DoD #2 |
| B | Audit actor 보강 (`source_ip`, `request_id`, `source_type`) | M1 DoD #3 |
| B | SEC-5 — `writeServerError` 헬퍼 도입 (별도 PR 권장) | M1 DoD #5 |
| B | RBAC policy 편집 API 결정 (write+audit) | M1 DoD #6 |
| F | `frontend/lib/services/types.ts` UI vs wire 타입 분리, 표시 포맷 프론트 이전 | M1 DoD #7 |
| X | WebSocket envelope `{schema_version, type, event_id, occurred_at, data}` 코드/문서 정합 | M1 DoD #8 |

### 4.2 우선순위 2 — M2 잔여 (백엔드 PR-D 의 후속)

- frontend `/auth/callback` 라우트 (Hydra `code` → `/oauth2/token` 교환 → 세션 저장)
- frontend `account.service.ts` 신설 (Kratos public flow 호출)
- backend `/api/v1/admin/identities/*` Kratos admin wrapper

### 4.3 운영 강화

- ADR-0001 §9 Phase 2 (Kratos identity 관리 endpoint, `users.status` 동기화) 진입
- Hydra/Kratos OS service 등록 결정 (NSSM / sc / systemd) — ADR-0002 또는 운영 가이드 보강

## 5. 머지된 PR 흐름 (M0 sprint)

```
0ea0ce4 Merge pull request #19 (PR-E close SEC-4)
e540c45 feat(auth): remove X-Devhub-Actor fallback code path
cf2d55f Merge pull request #18 (PR-D AuthGuard + /api/v1/me)
e4ddb7c feat(auth): replace mock login with /api/v1/me + OIDC redirect
21cd24a Merge pull request #17 (PR-C Hydra verifier)
   ↑ 82769d2 fix Hydra URL Redacted (Codex P1+P2)
   ↑ c9c77c9 fix Validate Env normalize + HydraRoleClaim
   ↑ b51d9f9 feat Hydra introspection
a477ca3 Merge pull request #16 (PR-B role guard)
5a2fec1 Merge pull request #15 (PR-A auth policy)
427d618 Merge pull request #14 (planning)
71233e6 Merge pull request #13 (source-docs split)
8493b63 Merge pull request #12 (integration)
```

## 6. 잔여 환경 제약 / 결정 대기

- **사내 GoProxy mirror 의존**: backend-core `go test ./...` 가 `proxy.golang.org` 도달 가능한 환경에서만 PASS. 사용자 사내 환경에서는 PASS 확인됨.
- **`/auth/callback` 부재**: PR #18 의 `/login` 이 Hydra authorize URL 로 redirect 까지만. 실 OIDC 사이클은 이 라우트 추가까지 미완 — M2 후속.
- **검증용 임시 OIDC client**: 본 세션에서 `client_credentials` grant 의 임시 client 1건을 Hydra 에 등록한 상태로 종료. 보안상 바로 삭제하거나 무시.
- **SEC-5 (DB 에러 노출)**: 별도 cleanup PR 로 분리됨, M1 진입 시 첫 항목으로 처리 권장.
