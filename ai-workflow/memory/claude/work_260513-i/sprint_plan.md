# Sprint Plan — claude/work_260513-i (대형 묶음: B1~D5)

- 문서 목적: B1 추가 도메인 + B2 + C1 + C2 (카탈로그) + D5 를 한 sprint / 한 PR 으로 통합 진행.
- 범위: docs (다수) + frontend (3 Vitest) + CI workflow (actionlint job) + ADR-0005.
- 진입 base: main HEAD `ceb0f6f` (PR #94 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. 작업 항목

| # | 항목 | 위치 | 규모 | 코드 변경 |
|---|------|------|------|----------|
| B1-account | 계정 관리 도메인 ID 노출 + IMPL 정의 | `backend_api_contract.md` §11.4, accounts admin endpoint + `report.md` §2.2 / §2.4 / §3 | S+ | 0 |
| B1-org | 조직 도메인 | `backend_api_contract.md` (organization endpoint) + `report.md` | S+ | 0 |
| B1-command | 명령 lifecycle | `backend_api_contract.md` §9 + `report.md` | S+ | 0 |
| B1-audit | 감사 (command cross-cut) | `backend_api_contract.md` §9 audit endpoint + `report.md` | S | 0 |
| B1-infra | 인프라 토폴로지 | `backend_api_contract.md` §6 + `report.md` | S+ | 0 |
| B2 | deprecated 문서 식별 + 마킹 | `backend/requirements_review.md` 등 후보 grep | S | 0 |
| C1 | frontend Vitest (Header / Sidebar / AuthGuard) | `frontend/src/components/**/*.test.tsx` | S+ | 신규 테스트 파일 3개 |
| C2 | TC 카탈로그 (TC-CMD-* / TC-INFRA-*) | `docs/tests/test_cases_*.md` | S | spec ts carve out |
| D5 | actionlint + ADR-0005 | `docs/adr/0005-*.md` + `.github/workflows/ci.yml` | S+ | workflow 변경 |

## 2. ID 매핑 (sprint 결정)

도메인별 매핑은 작업 진입 시 매트릭스 §2.2 / §2.4 의 기존 ID 범위를 우선 따르고, 본문 위치 명시.

매트릭스 §3 의 도메인 행 ID 범위 참조:
- 계정 관리: API-25, 32, 35 (35 = §11.5.1 cross-cut from auth, 이미 PR #93 에서 정의)
- 조직: API-33, 34
- 명령 lifecycle: API-15, 16, 17, 36, 37
- 감사: API-18, 39 (39 = §12.9 cross-cut from RBAC, 이미 PR #92 에서 정의)
- 인프라: API-06, 07 (+ API-05 dashboard)

## 3. 검증

- backend test: 0 (코드 변경 frontend 만)
- frontend test: Vitest 신규 + 기존 PASS
- CI: backend-unit / frontend-unit / e2e / actionlint (신규)

## 4. C2 carve out

E2E spec ts 의 실제 작성 (`*.spec.ts`) 은 본 sprint 의 carve out — TC 카탈로그 (`test_cases_*.md`) 에 행 추가만. 실제 spec 은 다음 sprint 의 별도 PR.

## 5. 미진입 / 다음 sprint 후보

- C2 의 실제 .spec.ts 작성
- M1-DEFER-E RBAC cache 다중 인스턴스 일관성
- M3 / M4 진입
