# 세션 인계 — claude/merge_roadmap

- 상태: M0 sprint 코드 측 종료, T-M0-10 운영 검증만 잔존
- 생성일: 2026-05-08
- 최종 수정일: 2026-05-08

## 1. 현재 작업 요약

- 현재 기준선: `main` HEAD `cf2d55f` (PR #18 머지 직후, M0 sprint 의 PR-A~D 머지 완료)
- 현재 주 작업 축: **PR #19 (PR-E)** 가 SEC-4 fallback 코드 path 자체 제거 + M0 sprint 마무리 메타 갱신으로 sprint 종료

## 2. M0 sprint 결과 요약

### 2.1 머지 commit 요약

| PR | 제목 | 커밋 |
| --- | --- | --- |
| #14 | docs(roadmap): integrated development roadmap + M0 sprint planning | `427d618` |
| #15 | fix(auth): close empty-Authorization pass-through, gate X-Devhub-Actor (PR-A) | `5a2fec1` |
| #16 | feat(authz): role guard middleware + protected route mapping (PR-B) | `a477ca3` |
| #17 | feat(auth): Hydra introspection verifier + prod fail-fast (PR-C) | `21cd24a` |
| #18 | feat(auth): replace mock login with /api/v1/me + OIDC redirect (PR-D) | `cf2d55f` |
| #19 (대기) | feat(auth): remove X-Devhub-Actor fallback code path; close M0 sprint (PR-E) | — |

### 2.2 SEC 결함 해소

| SEC | 상태 | 해소 PR |
| --- | --- | --- |
| SEC-1 mock auth (frontend) | ✅ resolved | #18 |
| SEC-2 verifier nil + empty Auth pass-through | ✅ resolved | #15 + #17 |
| SEC-3 role 미적용 | ✅ resolved | #16 |
| SEC-4 X-Devhub-Actor fallback | ✅ resolved | #15 (gate) + #19 (path 자체 제거) |
| SEC-5 DB 에러 노출 | tracked → M1 | — |

### 2.3 코드 내 SECURITY 마커

- backend-core 영역: **0건**
- frontend 영역: **0건**
- ai-workflow/memory 영역: 추적 메타라 잔존 (의도)

## 3. 잔여 작업

### T-M0-10 — ADR-0001 §9 Phase 1 운영 검증 (사용자 환경)

- Hydra binary :4444/:4445, Kratos binary :4433/:4434 가동
- PostgreSQL `hydra` / `kratos` schema 분리 검증
- `devhub-frontend` OIDC client 등록 검증
- 발급 access token 의 introspection 결과가 `auth.HydraIntrospectionVerifier` 에서 검증 가능
- 프론트엔드 `/login` → Hydra `/oauth2/auth` → Kratos UI → consent → callback → `/api/v1/me` 사이클 1회

검증 결과는 본 인계 또는 PR 코멘트로 첨부.

## 4. 다음 sprint 후보 (M1)

통합 로드맵 §3.2 M1 의 DoD 진입.

- envelope/lifecycle/role wire format 일관 적용
- audit actor 보강 (`source_ip`, `request_id`)
- RBAC policy 편집 API 결정
- types.ts UI/wire 분리
- WebSocket envelope 표준화 코드 적용
- SEC-5 (DB 에러 노출 마스킹) 별도 PR

## 5. 환경별 검증 현황

- 검증 완료 호스트: 사용자 사내 (PR #12·17·18·19 backend `go test ./...` PASS), 본 sandbox (frontend build/tsc PASS)
- infra/idp/ scaffold: 본 세션에서 실 구동 검증 미수행. T-M0-10 에서 사용자 환경 검증.

## 6. 현재 진행 중

- PR #19 (claude/m0-marker-sweep) 발행 + 머지 대기.

## 7. 차단 항목

- T-M0-10 운영 검증만 사용자 환경 의존. 코드 측 차단 없음.
