# Work Backlog — claude/work_26_05_11-d

- 문서 목적: 본 sprint 의 backlog 인덱스
- 범위: TDD foundation — Vitest + Playwright + 첫 시나리오
- 상태: in_progress (DEC-1/2/3/4 확정. PR-T1 진입 가능)
- 최종 수정일: 2026-05-11

## 활성 backlog

- [2026-05-11 — sprint 진입 + 개발 계획 + 진척 추적](./backlog/2026-05-11.md) ← **단일 source-of-truth**

## 진척 요약 (자세한 체크리스트는 위 backlog §5)

결정 4건 확정 (2026-05-11, 권장안 모두 채택): DEC-1=Vitest / DEC-2=Playwright / DEC-3=A 사용자 native / DEC-4=B 별도 sprint CI.

- PR-T1 (Makefile + baseline): done
- PR-T2 (Vitest + 첫 단위 테스트): done
- PR-T3 (Playwright + 첫 e2e 시나리오): done — 5/6 PASS, password-change skip
- PR-T3 fix-ups (PR #58 안): video=off, identity seed schema 정정, DSN env override 가이드, idp-apply-schemas -query, 002_seed_e2e_users.sql, password-change skip
- PR-T3.5 (e2e seed helper 자동화): planned — pre-e2e hook 으로 Kratos identity 3건 + DevHub users 3행 시드 (현재는 수동 절차). idempotent. 별도 sprint 가능

## 후속 sprint 후보 (이 sprint 외부)

- **PR-L4** (Kratos session 정합) — 현재 로그인이 api-mode 라 `ory_kratos_session` cookie 가 없어 `/account` 비밀번호 변경(PR-L3)이 실패. 두 가지 경로 중 결정 필요: (a) backend `/api/v1/account/password` proxy 가 session_token 으로 settings 수행, (b) 로그인을 browser-mode redirect 로 전환. e2e password-change.spec 는 skip 상태 — PR-L4 종료 시 unskip. 우선순위: M2 이전 hygiene.

## 인계 / 상태

- [세션 인계](./session_handoff.md)
- [상태 스냅샷](./state.json)
- [직전 sprint (CLOSED, PR #57)](../work_26_05_11-c/session_handoff.md)
