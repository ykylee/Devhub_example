# Session Handoff — claude/work_26_05_11-d (TDD foundation)

- 문서 목적: `claude/work_26_05_11-d` sprint 의 세션 간 상태 인계
- 범위: TDD 기반 마련 — Frontend 단위 테스트 인프라 (Vitest) + E2E (Playwright) 인프라 + 첫 시나리오 세트
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 상태: in_progress (결정 4건 확정 2026-05-11. PR-T1 진입 가능)
- 브랜치: `claude/work_26_05_11-d` (HEAD `fe27845`)
- 최종 수정일: 2026-05-11

## 0. 현재 기준선

- main HEAD `4e831a3` — 직전 sprint (work_26_05_11-c, M1 PR-D) closure 직후. M1 sprint 100% 완결.
- 본 brand 의 baseline commit `fe27845` 는 work_26_05_11-c closure 의 state.json 갱신 cherry-pick.
- 사용자가 다음 작업을 **TDD 기반 마련** 으로 선택. 회귀 안전망 구축이 본질.

## 1. 작업 축 (PR-T1/T2/T3)

[`./backlog/2026-05-11.md`](./backlog/2026-05-11.md) §3 가 단일 source-of-truth.

| PR | 작업 | 규모 |
| --- | --- | --- |
| PR-T1 | Makefile `test`/`test-race`/`test-coverage` + sprint baseline | S |
| PR-T2 | Frontend Vitest 인프라 + 첫 단위 테스트 (lib/auth/ 5 파일, ≥20 케이스) | M |
| PR-T3 | Playwright E2E 인프라 + 첫 시나리오 7건 + e2e 가이드 | L |

## 2. 결정 (확정 2026-05-11, 권장안 모두 채택)

- **DEC-1=Vitest**: Next.js 16 + Turbopack 호환, ESM 친화
- **DEC-2=Playwright**: 표준, 멀티 브라우저, Hydra/Kratos 통합 가능
- **DEC-3=A 사용자 native E2E**: 실 운영 흐름과 동일, no-docker 정책 부합
- **DEC-4=B 별도 sprint CI**: 본 sprint 는 인프라 + 첫 시나리오에 집중

## 3. 진입 순서

1. **PR-T1** — Makefile 3 target + baseline 4 문서 (PR 머지)
2. **PR-T2** — Vitest 설치 + 5 단위 테스트 파일 (≥20 케이스, lib/auth/ ≥80% coverage)
3. **PR-T3** — Playwright 설치 + 7 시나리오 + e2e 가이드 (사용자 환경에서 1회 PASS 검증)

## 4. 위험 / 운영 의존

- **PR-T2 npm 의존성 추가**: Vitest + RTL + jsdom + user-event 등. lock file 변경량 큼.
- **PR-T3 사용자 환경 의존**: Hydra/Kratos native + DevHub OIDC client + Kratos identity 3 role seed. 운영자가 사전 준비 필요 (e2e 가이드에 명시).
- **CI 부재**: 본 sprint 범위 밖. 다음 sprint 첫 작업이 GitHub Actions.

## 5. 다음에 읽을 문서

- [본 sprint backlog](./backlog/2026-05-11.md)
- [배포 가이드 §6-7 (E2E 사전 조건)](../../../../docs/setup/test-server-deployment.md)
- [API 계약](../../../../docs/backend_api_contract.md)
- [직전 sprint closure (work_26_05_11-c)](../work_26_05_11-c/session_handoff.md)
