# Work Backlog — opencode/work_260604-a-opencode-workflow-bootstrap

- Branch: `opencode/work_260604-a-opencode-workflow-bootstrap`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04
- Source of truth: 본 파일 (브랜치별 memory 디렉터리 표준)

## 0. 본 sprint 의 목표

OpenCode 워커 prefix 를 7개 governance/workflow 문서 + memory 4종 + 위치 패턴 표에 정합 반영. 본 워커의 영역/스타일 정의는 후속 sprint 의 backlog 로 분리.

## 1. 작업 단위 분해

| ID | 작업 | 영향 파일 | 의존 | 상태 |
| --- | --- | --- | --- | --- |
| WB-01 | 브랜치 생성 + checkout | (git) | main | done |
| WB-02 | memory 디렉터리 set up | `ai-workflow/memory/opencode/work_260604-a-opencode-workflow-bootstrap/{state.json,session_handoff.md,work_backlog.md,backlog/2026-06-04.md}` | WB-01 | done |
| WB-03 | AGENTS.md "OpenCode 전용 메모" 섹션 신설 + 최종 수정일 갱신 | `AGENTS.md` | — | done |
| WB-04 | worker_division.md §1.4 신설 + §2.5 worker 목록/예시/예외 갱신 + §6 변경 이력 | `docs/governance/worker_division.md` | — | done |
| WB-05 | release_v1_roadmap.md §5.1 분담 표 행 추가 + §6.3 label 추가 + §9 변경 이력 | `docs/planning/release_v1_roadmap.md` | — | done |
| WB-06 | MEMORY_GOVERNANCE.md §0 prefix 예시 추가 | `ai-workflow/MEMORY_GOVERNANCE.md` | — | done |
| WB-07 | WORKFLOW_INDEX.md §1 예시 추가 | `ai-workflow/WORKFLOW_INDEX.md` | — | done |
| WB-08 | ai-workflow/README.md §2/§7 갱신 | `ai-workflow/README.md` | — | done |
| WB-09 | docs/planning/README.md §3 표 행 추가 | `docs/planning/README.md` | — | done |
| WB-10 | 검증: `git status`/`git diff`/`grep` 으로 모든 변경 일관성 확인 + state/handoff final 갱신 | (this dir) | WB-02..09 | done |

## 2. 적용 규칙 (이번 sprint 만)

- 본 sprint 는 governance/문서 갱신만 수행 — 코드/빌드 영향 없음
- 모든 변경은 한국어 사용자 보고 + 영문 ID/경로 유지
- `state.json` + `session_handoff.md` 의 `updated_at` 은 PR 마감 시점에 1회 갱신

## 3. 검증 기준 (DoD)

- [x] 7개 governance/workflow 문서에 opencode 섹션/예시/행/링크 추가 완료
- [x] `ai-workflow/memory/opencode/work_260604-a-opencode-workflow-bootstrap/` 4종 파일 존재
- [x] `git status` 상 의도치 않은 파일 변경 없음 (untracked `ai-workflow/memory/deepseek/work_260602/`, `devhub-backend/` 사전 그대로)
- [x] 본 문서(`work_backlog.md`) 의 모든 WB-* 가 done
- [x] 변경 이력 (각 문서 §변경 이력) row 추가

## 4. carry-over / 후속 sprint 후보

- **opencode-areas** — `worker_division.md` §1.4 본문 채우기 (영역/도구/스킬/외부 contribution 정책)
- **N-10 Manager RBAC 검증** — NOW backlog. mgr-user-b 재생성 + 권한 scope 확인
- **X-1 System Admin dashboard** — v1.1 진입 준비
- **v1.0 staging 1주 운영 (N-6)** — 사용자 동반 작업
