# Session Handoff — deepseek/e2e_ci_cleanup

- 목적: E2E 테스트 + CI 구성 현행화 (skip 테스트 해소, 문서 정합, 버그픽스)
- 상태: PR 준비 완료
- 최종 수정일: 2026-05-26
- 워커: deepseek (codex 영역: infra/CI/build)

## 변경 요약

### 1. ci-e2e-sync-check.sh 버그픽스
- `required_workflow_tokens`에서 문장형 중복 항목 `"DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET is required for e2e global setup"` 제거
- 동일 env var가 5번째 항목으로 이미 검사되고 있어 중복이었음

### 2. describe.skip 테스트 해소
- **infra-topology.spec.ts**: `/admin` 페이지 아카이브로 TC-INFRA-RENDER-01 무효 → 파일 삭제 (admin-topology-v2.spec.ts가 대체)
- **dashboard-retry-empty-state.spec.ts**: admin dashboard 아카이브로 admin test 제거, developer/manager test만 유지, describe.skip 해제

### 3. 문서 정합
- `test_cases_m7_onboarding.md`: `onboarding.spec.ts` → `onboarding-first-login.spec.ts` (12개 참조 일괄 교체)
- `test_cases_m4_integration.md`: 상태 draft → accepted
- `e2e_testing_strategy.md`: stale `scripts/ci-setup.sh` 참조 → `.github/workflows/ci.yml` inline step 참조로 정정

### 4. client ID 기본값 통일
- `global-setup.ts`: `devhub-e2e-seeder` → `devhub-backend`
- `onboarding-first-login.spec.ts`: `devhub-e2e-seeder` → `devhub-backend`
- fixtures.ts의 기본값 `devhub-backend`와 일치

### 5. CI workflow stale 주석 정정
- `ci.yml` 2건: `ci-setup.sh` 참조 → 실제 코드 구조에 맞게 정정

## 다음 directive
- PR 생성 및 머지
- 필요 시 `docs/setup/e2e-test-guide.md`도 `devhub-e2e-seeder`→`devhub-backend` 문서 갱신 고려
