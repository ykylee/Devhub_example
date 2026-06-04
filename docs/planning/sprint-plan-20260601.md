# Sprint Plan: Test-Driven v1.0 Gap Closure

- **문서 목적**: 2026-06-01 통합 테스트 결과 기반 v1.0 GAP 해소를 위한 sprint 계획 수립
- **범위**: P0~P1 신규 carve + v1.0 sprint 순서 재조정
- **대상 독자**: Claude, Codex, Gemini, Sisyphus
- **상태**: active
- **최종 수정일**: 2026-06-01
- **관련 문서**: `release_v1_roadmap.md` (§3.1~3.5 P0/P1 carve), `integrated_test_report_20260601.md` (§8~10 평가)

---

## 1. 배경

2026-06-01 브랜치 `deepseek/test-scenarios-20260601`에서 4개 Phase(Keycloak 온보딩 → App/Project/Repo/Gitea → Gitea PR/Issue → CI 빌드) E2E 통합 테스트 수행 결과:

- **P0 Blocker (1건)**: CI Run 생성 API 부재 (ISSUE-05) — CI 기능의 실질적 사용 불가
- **P1 권장 (3건)**: Sign-out endpoint 미구현 (BUG-03), Repository build-runs endpoint 미구현 (ISSUE-04), Keycloak role sync 누락 (BUG-02, P1-1에 포함)
- **Manager role RBAC 검증 누락**: mgr-user-b password grant 실패 → 재생성 후 확인 필요

본 sprint plan은 이 GAP을 해소하고 v1.0 로드맵(release_v1_roadmap.md)에 신규 carve를 통합한다.

---

## 2. Sprint 구성

### Sprint -h (즉시 진입): P0-4 CI Run API + P1-1 Role Sync

| 작업 | ID | 영역 | 워커 | 예상 기간 |
|------|----|------|------|----------|
| `POST /api/v1/ci-runs` endpoint 구현 | P0-4 (ISSUE-05) | Backend (Go) | Claude | 2일 |
| CI Run 생성 validation (status: queued/running/success/failed/cancelled/skipped/unknown) | P0-4 | Backend | Claude | 포함 |
| CI Run 생성 시 `repository_id`, `branch`, `commit_sha` 필수 검증 | P0-4 | Backend | Claude | 포함 |
| P1-1 Keycloak event listener — `devhub_role` → `users.role` sync | P1-1 (BUG-02) | Backend (Go) | Claude | 3일 |

### Sprint -i: P1-6 Sign-out + P1-7 Build-runs + P1-2 JWKS

| 작업 | ID | 영역 | 워커 | 예상 기간 |
|------|----|------|------|----------|
| `POST /api/v1/auth/logout` endpoint (token 폐기 + session 종료) | P1-6 (BUG-03) | Backend | Claude | 1일 |
| `GET /api/v1/repos/{id}/build-runs` (ci_runs 기반) | P1-7 (ISSUE-04) | Backend | Claude | 1일 |
| Frontend build-runs widget (repo detail page) | P1-7 | Frontend | Gemini | 1일 |
| P1-2 JWKS stale-while-error expiry case 확장 | P1-2 | Backend | Claude | 2일 |

### Sprint -j (-k 병합 가능): UI 마무리 + 검증

| 작업 | ID | 영역 | 워커 | 예상 기간 |
|------|----|------|------|----------|
| P2-1 governance SOP + P0-2 UI polish 마무리 | P2-1, P0-2 | Infra/Frontend | Codex+Gemini | 3일 |
| Manager role RBAC E2E 재검증 (mgr-user-b) | N-10 | 테스트 | Sisyphus | 1일 |
| v1.0 종합 E2E 회귀 테스트 | - | 모두 | 전 워커 | 2일 |
| staging 1주 운영 검증 | N-6 | 사내 | 사용자 | 7일 |

---

## 3. 워커 분담

### Claude (Backend) — 3 sprint 순차 진행

```
sprint -h: P0-4 (CI Run API) → P1-1 (role sync)  [5일]
sprint -i: P1-6 (Sign-out) → P1-7 (Build-runs) → P1-2 (JWKS)  [4일]
sprint -j: 필요시 hotfix + 회귀 테스트 지원
```

### Gemini (Frontend)

```
sprint -g/h: P0-3 Playwright screenshot + P0-2 UI polish 진입
sprint -i: P1-7 Build-runs widget (repo detail page)
sprint -j: UI polish 마무리
```

### Codex (Infra/CI)

```
sprint -g/h: P0-3 Playwright screenshot CI + CI artifact upload
sprint -i/j: P2-1 governance SOP
```

### Sisyphus (Test)

```
sprint -i: N-10 Manager role RBAC 검증
sprint -k: v1.0 E2E 회귀 테스트
```

---

## 4. 의존성 및 리스크

| 의존성 | 영향 | 완화 |
|--------|------|------|
| P0-4 (CI Run API) → sprint -g 완료 불필요 | 독립 진행 가능 | sprint -h에서 Claude backend 단독 처리 |
| P1-7 (Build-runs) → FE widget은 Gemini 대기 | BE 우선 구현 | Claude BE endpoint 먼저, Gemini는 sprint -i에서 widget |
| P1-1 (role sync) → BUG-02 + P1-6 (Sign-out) 독립 | 독립 | 두 작업 병렬 가능 |
| Manager role RBAC (N-10) → mgr-user-b 재생성 완료 | 완료 | 이미 Keycloak에서 credentials 포함 재생성 완료 (2026-06-01) |

---

## 5. 완료 정의 (DoD)

| 항목 | 기준 |
|------|------|
| P0-4 CI Run API | `POST /api/v1/ci-runs` → 201 + DB 저장 + `GET /api/v1/ci-runs` 조회 |
| P1-1 Role sync | Keycloak `devhub_role` 변경 → DevHub `users.role` 자동 반영 |
| P1-6 Sign-out | `POST /api/v1/auth/logout` → token expired + 200 |
| P1-7 Build-runs | `GET /api/v1/repos/{id}/build-runs` → `ci_runs` 기반 결과 반환 |
| N-10 Manager RBAC | mgr-user-b로 `/api/v1/platforms` 조회 → developer와 다른 scope 확인 |

---

## 6. 변경 이력

| 일자 | 변경 |
|------|------|
| 2026-06-01 | 최초 작성 — 통합 테스트 결과 기반 v1.0 GAP sprint plan |
