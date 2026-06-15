- 문서 목적: 본 sprint 의 todo (planned → in_progress → done) + carry-over 추적.
- sprint branch: feat/work_260610-c-i-e2e-internal-job
- 대상 독자: 후속 에이전트, PR reviewer, owner
- 상태: in_progress (사용자 confirm 후 commit + push + PR 발행 대기)
- 최종 수정일: 2026-06-10
- 관련 문서: session_handoff.md, backlog/2026-06-10.md, state.json, pr_body.md

## 상태 정의

- planned — 미착수
- in_progress — 작업 중
- blocked — 의존성/외부 입력 대기
- done — 검증 + 완료 확정

## 본 sprint (PR, C-i 풀번들) task

| ID | 상태 | 우선순위 | 제목 | 비고 |
| --- | --- | --- | --- | --- |
| T-c-i-e2e-internal-1 | done | P2 | ci.yml e2e-internal job 신규 (+202 lines) | 23 step (PG 15 + Keycloak port 8181 + apply migrations + validate contract + Start Backend DEVHUB_BUILD_TIER=internal + Start Frontend + Wait + Run E2E Tests + Upload Report + Upload Logs). e2e shard 1/2/3 변경 0 |
| T-c-i-e2e-internal-2 | done | P2 | ci-e2e-sync-check.sh DEVHUB_BUILD_TIER 의도적 미포함 + comment + verify all 4 lint PASS | required_e2e_tokens 변경 0, comment 5 lines. tier-separation/openapi/migration/e2e-sync 모두 PASS |
| T-c-i-e2e-internal-3 | pending | P2 | commit + push + gh pr create --body-file pr_body.md | 사용자 confirm 대기 |
| T-c-i-e2e-internal-4 | planned | P2 | main flat memory 4종 sync (post-merge) | 본 PR 머지 후 자동 sync (PR body 의 head_commit + memory directive 의 머지 후 갱신 절차) |

## carry-over (다음 sprint 후보)

| ID | 우선순위 | 제목 | 비고 |
| --- | --- | --- | --- |
| C-j | P3 | build tag 정책 재검토 (runtime injection ↔ build tag 전환) | ADR-0030 §2.3 의 trade-off 재평가 |
| backend-integration DEVHUB_BUILD_TIER matrix | P3 | backend-integration job (현재 && false) 의 DEVHUB_BUILD_TIER=internal matrix | sprint -a follow-up 본 PR #539 의 backend-integration 재활성화 + DEVHUB_BUILD_TIER matrix |
| release_v1_roadmap §3.5 N-13 정합 | P3 | 본 PR 의 N-13 row status 갱신 (C-i done) | housekeeping |

## 사이드 점검 (본 PR 진행 시 발견)

- `docs/adr/0030-sso-integrations-and-auth-session-port.md` §2.3 runtime injection 의 정공법 = 본 PR 의 e2e-internal job. ADR 변경 불요 (이미 §2.3 + §5 timeline 1.1a/1.1b accepted/done 정합).
- `ci.yml` 의 e2e shard 1/2/3 의 env block 의 saovae_stub default 유지 — 본 PR 의 의도 (real adapter 검증은 e2e-internal 만).
- `ci-e2e-sync-check.sh` 의 contract = e2e helper 가 token 사용 OR ci.yml 의 e2e step env block 에 token 존재. DEVHUB_BUILD_TIER 는 e2e-internal job 의 env block 에만 노출되므로 script token 에 미포함이 정합.
- frontend e2e env 의 DEVHUB_BUILD_TIER 추가 불요 — frontend 의 e2e logic 가 backend 의 build tier 무관 (auth, RBAC, CRUD 모두 backend API 만 호출).
