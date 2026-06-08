# Session Handoff — main (2026-06-08, N-11 정합 진행 중)

- 문서 목적: N-11 (CI e2e + backend-integration job 복원) 의 운영 정합 sprint 260608-a 진행 상태 인계.
- 범위: N-11 의 1차 PR (PR #498 코멘트 갱신) + 2차 PR (메모리 4종 + traceability 정합) 의 2 PR 구조. 잔여 DoD = main 첫 PR 에서 두 job 실 실행 PASS (issue #419).
- 상태: main HEAD `3855e46`, 1차 PR #498 OPEN, 2차 PR 작성 중 (`opencode/work_260608-b-N11-memory-sync`).
- 최종 수정일: 2026-06-08

## 0. N-11 (issue #419) — CI e2e + backend-integration job 복원 운영 정합

### 1) 현황 분석 결과

- `&& false` 2건 (backend-integration, e2e) 은 PR #407 cleanup-recovery (4a1942e) 후속 squash merge 4건 (5f5fdba fan out e2e shards, 9395cd9 restore stable e2e workflow, ce8ce7c fan out stable e2e shards) 으로 **이미 코드 레벨 복원 완료**.
- 본 sprint -a 의 PR (#498) 은 **코멘트만 갱신** (코드 변경 0줄): "비활성화 + 복원 SOP" 안내 → "복원 완료 + 잔여 DoD 명시" 로.
- 잔여 DoD = main HEAD `3855e46` 의 첫 PR 에서 backend-integration + e2e shard 1..3 실 실행 PASS 확인 (issue #419 잔여).

### 2) 2 PR 구조

- **1차 PR #498** (opencode/work_260608-a-N11-ci-restore-verify) — ci.yml 코멘트 갱신. workflow-lint 잡이 1차 CI 검증. **OPEN, 머지 보류**.
- **2차 PR** (opencode/work_260608-b-N11-memory-sync) — 메모리 4종 (state/handoff/work_backlog/release_v1_roadmap) + traceability report.md §6 + §3.5 N-11 + §4.1 + §9 정합. 1차 PR 머지 후 main HEAD XXX 로 state.json head_commit 갱신. **작성 중**.

### 3) 잔여 DoD 모니터링

- 1차 PR 머지 후 main 에 들어오는 첫 PR 의 CI run 에서 `backend-integration` + `e2e (shard 1/3, 2/3, 3/3)` 4개 잡 모두 green 확인.
- 확인 후 별도 housekeeping sprint (예: mvs 또는 opencode) 에서 release_v1_roadmap §3.5 N-11 의 "잔여 DoD" 행을 "✅ resolved" 로 close.

## 1. 2026-06-06 sprint -h 추적성 ID 발급 (PR #490)

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
