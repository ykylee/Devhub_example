# Work Backlog — mvs/work_260608-i-488-sign-out (N-8 / P1-6)

- 문서 목적: 본 sprint 의 작업 항목 진행 상태와 결정 기록을 추적한다.
- 범위: `POST /api/v1/auth/logout` 신규 endpoint — access token 폐기 + session 종료.
- 대상 독자: 본 sprint 진입자 (Mavis / MiniMax-Code), 후속 리뷰어 (Codex 외부 리뷰 + Claude self-review).
- 상태: in_progress
- 최종 수정일: 2026-06-08
- 관련 문서: [session_handoff.md](./session_handoff.md), [state.json](./state.json), [Issue #488](https://github.com/ykylee/Devhub_example/issues/488), [release_v1_roadmap.md §3.2 P1-6 + §3.5 N-8](../../../../docs/planning/release_v1_roadmap.md).
- 스프린트 목표: v1.0 release 안정성 carve (P1-6 / N-8) 해소. 세션 관리 기본.

## 진행 상태

- [x] 이전 sprint N-7 (PR #494) 머지 완료 + state.json done 갱신
- [x] 브랜치 mvs/work_260608-i-488-sign-out 생성 (origin/main @ 7609000 base)
- [x] sprint scaffold 3 파일 초기화 (state.json + work_backlog.md + session_handoff.md)
- [ ] Issue #488 본문 + 2차 정정 housekeeping comment 정독
- [ ] 기존 auth handler 패턴 탐색 (JWKS, token verify, middleware)
- [ ] IdentityAdmin.LogoutUserSession 활용 가능 여부 검증
- [ ] `POST /api/v1/auth/logout` handler 구현
- [ ] refresh token rotate 포함 여부 결정 (사용자 확인)
- [ ] route 등록 + RBAC + audit emit
- [ ] UT 4-5건 (TC-AUTH-LOGOUT-01..05)
- [ ] go test ./... + go build PASS
- [ ] commit + push + PR 생성

## 결정 기록

- **2026-06-08 sprint-i 진입** — N-7 (PR #494) 머지 후 자연스러운 다음 단계. release_v1_roadmap §4.1 의 sprint-i 의 P1-6 + P1-2 + P0-2 동시 중 P1-6 가 가장 blocker 적.
- **2026-06-08 정식 ID = API-99** — issue #488 housekeeping comment 의 2차 정정에 따라 (1차 API-99 → 2차 API-99 유지, ID 슬롯 충돌 해결 완료). sprint -h ID 슬롯 (API-98/99/100) 중 99 가 본 carve.
- **(예정) refresh token rotate** — issue #488 본문 spec 이 "access token 폐기 + session 종료" 만 명시. refresh token rotate 포함 여부는 사용자 확인 필요 (spec 옵션).
- **(예정) LogoutUserSession 재사용** — RouterConfig.IdentityAdmin.LogoutUserSession (sprint ADR-0020 sub-carve E) 가 이미 존재. Keycloak admin client 의 logout session API. 본 handler 가 본 메서드 호출.

## 다음 sprint 진입 후보 (본 sprint 머지 후)

1. **P1-2 sub-carve D** (JWKS stale-while-error expiry) — Claude sprint-i 동시
2. **P0-2 UI polish 마무리** — Gemini 주도
3. **N-9 frontend widget** (repository build-runs, P1-7 GET 기반) — Gemini
4. **N-11 CI e2e+backend-integration job 복원** (#419) — Codex + 사용자

## 의존성

- N-7 (#486) — main 머지 완료 ✓
- PR #418 (refactor stabilize) — main 머지 완료 ✓
- IdentityAdmin 인터페이스 (ADR-0020 sub-carve E, sprint -n) — main 머지 완료 ✓
- Keycloak OIDC verifier + JWKS cache (`internal/auth/keycloak_verifier.go`) — done ✓
