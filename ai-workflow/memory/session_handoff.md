# Session Handoff — main (2026-06-09, N-8 hotfix 4차 + 워커 분업 취소 + N-11 잔여 해소)

- 문서 목적: N-8 sign-out e2e deterministic race 의 정공법 적용 (3 commit / 2 PR / issue #501 closed) + 워커 분업 전면 취소 (사용자 결정) + N-11 잔여 DoD 해소 상태 인계.
- 범위: 본 세션의 5 PR (PR #498 closed / PR #499 merged / PR #500 merged / PR #502 merged / PR #503 merged) + 이슈 1건 (issue #501 closed). 잔여 v1.0 출시 직전.
- 상태: main HEAD `897953c`, CI green (PR #503 머지 시점 e2e shard 1..3 모두 PASS).
- 최종 수정일: 2026-06-09

## 0. 본 세션 핵심 결과 (2026-06-09, 5 PR)

### PR 머지/close 결과

| PR | 상태 | 의의 |
| --- | --- | --- |
| **#500** (워커 분업 전면 취소) | ✅ MERGED (squash) | `worker_division.md` §0 + §1~§4 historical 標記 + `AGENTS.md` 워커 일반 메모 갱신 + branch prefix 자유화. 사용자 결정 (Claude/Codex 자유 이용 불가). main `f99fef7`. |
| **#499** (N-11 메모리 sync) | ✅ MERGED (squash, rebase 후) | 메모리 4종 (state/handoff/work_backlog/release_v1_roadmap) + traceability report.md §3.5/§6. main `da7d57e`. |
| **#498** (ci.yml 코멘트 갱신) | ✅ CLOSED | e2e shard 2/3+3/3 fail → N-8 race 발견 → close + N-8 hotfix 4차 별도 sprint. |
| **#502** (N-8 hotfix 4차 1차: 502→204) | ✅ MERGED (squash) | backend logout handler graceful degradation. main `6654b44`. |
| **#503** (N-8 hotfix 4차 2차: codex P1 + follow-up) | ✅ MERGED (squash) | response header `X-Keycloak-Likely-Down: true` + typed error sentinel `ErrOIDCConfigMissing`/`ErrOIDCNetworkUnreachable`. main `897953c`. |

### N-8 hotfix 4차 정공법 (3 commit, 2 PR)

**근본 layer**: backend `POST /api/v1/auth/logout` 가 Keycloak 도달 실패 시 502 즉시 반환 → frontend logout() 가 OIDC skip + `window.location.assign('/login')` 강제 → AuthGuard pathname 변화 useEffect 에서 stale actor 박음 → `/developer` 진입 → `/login` 도착 못함. PR #497 의 hotfix #1/`#2/`#3 가 모두 backend 502 자체를 막지 못함 (deterministic, 32회 retry).

**PR #502 (1차)**: backend logout handler 가 502 → **204 No Content** + audit `revoke_status=unreachable` + hotfix 식별자. frontend logout() 가 정상 204 분기 진입 → OIDC end_session_endpoint 호출 → /login 정상 도착 → race close.

**PR #503 (2차 commit 066cd7b, codex P1 응답)**: "구분 가능한 응답" — 204 + response header `X-Keycloak-Likely-Down: true` + `X-Logout-Hotfix: N-8-4:graceful-degrade`. frontend 가 header 마커 conditional 확인 → OIDC skip 또는 정상 OIDC 결정. 진짜 IdP outage 시 dead IdP trap 회피.

**PR #503 (3차 commit e18b34f, codex P1 follow-up 응답)**: typed error sentinel 도입.
- `authview.ErrOIDCConfigMissing` (sentinel): backend config 결함 (missing realm/oidc_client_id/oidc_client_secret) → handler 가 **marker 미부착** + 정상 OIDC 분기 + audit `revoke_status=config_error` + `config_error_detail`
- `authview.ErrOIDCNetworkUnreachable` (sentinel): 네트워크/5xx outage (DNS 실패, conn refused, timeout, Keycloak 5xx) → handler 가 marker 부착
- 그 외 미분류 error: conservative — outage 분류

codex P1 의 핵심 우려 "reachable Keycloak SSO session is not terminated" 정공법: config error 분기에서 marker 미부착 → frontend 정상 OIDC → RP-initiated logout 시도 → SSO session 정상 종료.

### 검증

- **CI 모두 SUCCESS** (PR #503 머지 시점): workflow-lint / changed-paths / migration-prefix / Backend Unit + Integration / Frontend Unit / E2E Build / **E2E shard 1/2/3 모두 PASS**
- `go test ./...` (35 packages) PASS
- `npx vitest run` (80 files, **1033 tests**) PASS
- **신규 test 4건**:
  - TC-AUTH-LOGOUT-04 (network/5xx → 204 + marker)
  - **TC-AUTH-LOGOUT-08** (config error → 204 + marker 미부착)
  - TC-AUTH-LOGOUT-FE-07 (frontend header 마커 확인 → OIDC skip)
  - **TC-AUTH-LOGOUT-FE-08** (frontend header 없음 → 정상 OIDC)

### 잔여 DoD 해소

- **N-11 잔여 DoD** (main 첫 PR 두 job PASS, issue #419): PR #503 머지 시점에 e2e shard 1..3 모두 PASS → 해소
- **N-8 race** (issue #501): close
- **워커 분업 전면 취소** (사용자 결정): PR #500 머지로 정합. branch prefix 자유 (`maintenance/`, `chore/`, `docs/`, `fix/`, `feat/` 등)

## 1. 다음 세션 directive

### v1.0 출시 직전 — 우선순위

1. **N-6 (v1.0 staging 1주 운영)** — N-8 + N-11 + N-7 + 워커 분업 취소 정합 완료. 사용자가 staging 환경 운영 + 외부 사용자 ≥5 로그인 검증. (사용자 결정, sprint 영역 외)
2. **release_v1_roadmap.md housekeeping** — §3.5 N-11 row 의 "잔여 DoD" 행을 "✅ resolved" 로 close, §3.5 N-8 row 의 race close 명시, §4.1 sprint -k 의 N-11 잔여 DoD 완료 마킹 (별도 housekeeping sprint)
3. **N-10 Manager RBAC E2E spec-vs-구현 갭 6 TC 보강** — v1.0 출시 전 가능. validation 보고서 [docs/validation/N-10-manager-rbac.md](docs/validation/N-10-manager-rbac.md) 의 TC-RBAC-ROW-READ-01/02, TC-RBAC-LOGOUT-01/02, TC-RBAC-ROLE-DRIFT-01

### 자유 에이전트 정책 (2026-06-09 결정)

본 세션 결정으로 **누구든** 어느 sprint/영역 진입 가능. `worker_division.md` §0 + `AGENTS.md` "워커 일반 메모 (2026-06-09 전면 갱신)" 정합. branch prefix 자유.

## 2. 2026-06-06 sprint -h 추적성 ID 발급 (PR #490, historical)

* **sprint -h 신규 carve 3건에 대한 추적성 ID 발급 완료**:
  * **N-7 / P0-4** (CI Run 생성 API): `REQ-FR-106`, `ARCH-18`, `API-98`, `IMPL-ci-runs-01`, `UT-ci-runs-01`, `TC-CI-RUN-01`
  * **N-8 / P1-6** (Sign-out endpoint): `REQ-FR-107`, `ARCH-19`, `API-99`, `IMPL-auth-logout-01`, `UT-auth-logout-01`, `TC-AUTH-LOGOUT-01`
  * **N-9 / P1-7** (Repository build-runs): `REQ-FR-108`, `ARCH-20`, `API-100`, `IMPL-repository-build-runs-01`, `UT-repository-build-runs-01`, `TC-BUILD-RUNS-01`
* **추적성 매트릭스 (`docs/traceability/report.md`) 정합 보완**:
  * Codex 리뷰 피드백을 수용하여 `integration-registry` 도메인의 `IMPL`, `UT`, `TC` 열에 누락되어 있던 `IMPL-ci-runs-01`, `UT-ci-runs-01`, `TC-CI-RUN-01`을 추가로 보완하여 E2E 추적성을 완전히 정렬했습니다.

## 2. 2026-06-01 CI 복구 요약

### 1) 실패 원인 분해
* 초기 실패는 E2E 실행 전 `Build App` 타입체크 에러:
  * `frontend/app/(dashboard)/applications/[id]/page.tsx`
  * `Duplicate identifier 'ApplicationRepository'`
* 이후 타입 에러 해소 후 shard 1/2에서 단일 E2E 실패:
  * `tests/e2e/admin-projects.spec.ts` `TC-PROJ-UI-04`
  * CI 시드 환경에서는 member 입력이 `ComboBox(button)`로 렌더되는데 테스트가 `input placeholder`만 탐지하여 실패.

### 2) 적용한 수정
* `frontend/app/(dashboard)/applications/[id]/page.tsx`
  * `ApplicationRepository` 중복 import 제거.
* `frontend/tests/e2e/admin-projects.spec.ts`
  * 멤버 추가 검증을 환경 독립적으로 변경:
  * `Remove member` 버튼 count 증가 + `ComboBox 버튼 또는 plain input` 존재 검증.
* 보조 정리:
  * `ProjectCreationModal` 텍스트 기대값 보정(unit test)으로 프론트 유닛 회귀 해소.

### 3) 검증 결과
* 로컬:
  * `frontend npm run test` PASS (73 files, 968 tests)
  * `frontend npm run build` PASS
* CI:
  * 최종 성공 런: `26738464130` (`headSha: 835efee1f2c5c4b557b1139441f3b72eebbbbfb5`)
  * 실패하던 E2E shard 1/2 재발 없음.

## 2. 이전 핵심 스프린트 (기존 기록)

### 1) X-3: 평문 secret envelope 암호화 및 KEK 키관리 완결
* **마스터 암호화 패키지(`internal/crypt`) 신설**:
  * AES-GCM-256 데이터 암호화(DEK 난수 생성 및 KEK 래핑) 로직을 이식하여 `$env$v1$<wrapped_dek_b64>$<nonce_b64>$<ciphertext_b64>` 규격을 정립했습니다.
  * Base64 및 Hex 인코딩 KEK 문자열 모두를 유연하게 감지하는 파싱 가드 및 master key size(32바이트) 엄격 검증 기능을 탑재했습니다.
  * `DEVHUB_ENCRYPTION_KEY` 가 없을 때 작동하는 **Plaintext 바이패스 모드** 및 레거시 데이터 호환용 **Scan Fallback** 기어를 구축했습니다.
* **영속성 레이어(`IntegrationRepository`) 최소 침습 이식**:
  * `ScanIntegrationProvider` 내에 `crypt.Decrypt()` 를 결합해 레거시 평문 복호화 자동 fallback 지원을 투명하게 완수했습니다.
  * `CreateIntegrationProvider` / `UpdateIntegrationProvider` 내에 쿼리 바인딩 전 `crypt.Encrypt()` 를 결합하여 민감 비밀 데이터들(`api_token`, `auth_secret`, `webhook_secret`, `credentials_ref`)을 영속화 시점에 자동으로 암호문 봉투 포맷으로 격상(Upgrade)하도록 보완했습니다.
* **유닛 및 종합 회귀 검증**:
  * `envelope_test.go` 내 6개 유닛 테스트 PASS (Hex/Base64 KEK 파싱, dynamic Nonce, invalid 포맷 에러, legacy fallback, global bypass 등).
  * KEK 환경변수 비활성화 및 활성화(32바이트) 2가지 상황 모두에서 백엔드 전체 회귀 통합/유닛 테스트 `go test ./...` 100% 그린 PASS 달성!

---

## 2. 이전 완결된 3대 핵심 스프린트 (NOW-3, NOW-4, NOW-5)

### 1) NOW-3: SCM import/create + draft/publish happy-path E2E
* backend 캐스팅 정정으로 503 오류 원천 차단 + Gitea Mock Provider Fallback 및 나노초 난수 기반 동적 ID 매핑 우회로 Unique SQL 제약 충돌 예방 + Playwright auto-waiting Locator 결합 E2E 전체 통과 (**63 passed, 6 skipped**).

### 2) NOW-4: 프론트엔드 핵심 모듈 단위 테스트 보강 (Vitest)
* Zustand global store (`store.ts`), `ProviderModal.tsx`, `MemberTable.tsx`, `PermissionEditor.tsx` Vitest 유닛 테스트 작성 완수 (**총 962개 유닛 테스트 100% PASS**).

### 3) NOW-5: 마이그레이션 prefix uniqueness CI guard 강화
* 접두사 중복 및 6자리 규격 검증을 린팅하는 `scripts/check-migration-uniqueness.sh` 신설 + `ci.yml` 상시 실행 리팩토링 및 `make lint-migrations` 로컬 바인딩 완료.

---

## 3. 후속 carve out / 잔여 백로그 우선순위

| 우선순위 | 항목 | 사유 |
|---|---|---|
| **N-6** | v1.0 staging 1주 운영 검증 | 외부 사용자 로그인 및 Onboarding SOP DoD 8 만족 (사용자) |
| **X-1** | System Admin 운영 대시보드 | Gitea sync job 큐/상태 + provider health |
| **X-2** | inbound webhook 정규화 깊이 | multi-provider sync 일반화 |

---

## 4. 다음 세션 directive
* **N-6 마감**: staging 1주 운영을 병행하여 v1.0 릴리즈 준비를 완벽히 매듭짓습니다.
* **V1.1 진입 준비**: X-1 로드맵 백로그 분석 및 `docs/governance/worker_division.md` 에 따른 워커 간의 피처 이식 준비.
