# Session Handoff — claude/work_260513-g (2026-05-13)

- 문서 목적: B1 auth 도메인 확장 (RBAC 다음 1차) sprint 인계.
- 범위: docs only. 코드 변경 0.
- 진입 base: main HEAD `a73dba1` (PR #92 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. 작업 흐름

| 단계 | 결과 |
| --- | --- |
| 0. branch + sprint memory 초기화 | DONE |
| 1. backend_api_contract.md §11.3 / §11.5 / §11.5.1 ID 노출 | pending |
| 2. report.md §2.2 auth API 매핑 서브 표 | pending |
| 3. report.md §2.4 IMPL-auth-XX 책임 정의 서브 표 | pending |
| 4. report.md §3 인증 / 회원가입 / 계정 관리 행 ID 범위 정리 | pending |
| 5. report.md §6 변경 이력 + main flat sync (PR #92 흡수) | pending |
| 6. PR + 2-pass + squash merge | pending |

## 2. 핵심 결정

- API-19 = §11.3 Bearer token 경계 (verifier interface), API-20..24 = §11.5 의 5 self-service endpoint, API-35 = §11.5.1 self-service password change (cross-cut: 인증 행 + 계정 관리 행).
- §11.2 Hydra 표준 endpoint 는 외부 의존성 — API-XX 미부여 (`conventions.md` §5.2).
- §11.4 admin identity wrapper 는 planned — M3 진입 시 ID 발급.
- IMPL-auth-01..02 는 verifier/actor wiring, IMPL-auth-03..07 은 5 endpoint handler. POST /api/v1/account/password 의 IMPL 은 account 도메인 (본 sprint 스코프 밖, account_password.go).

## 3. 회귀 위험

- 코드 변경 0. 매트릭스 row + §11 본문 ID 노출만.
