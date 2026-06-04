# Work Backlog — opencode/work_260604-b-opencode-areas

- Branch: `opencode/work_260604-b-opencode-areas`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04
- Source of truth: 본 파일 (브랜치별 memory 디렉터리 표준)

## 0. 본 sprint 의 목표

OpenCode 워커의 영역/책임/작업 스타일/인계 SOP 을 `worker_division.md` §1.4 본문으로 정의. 본 sprint 는 governance 정합 + 메모리 set up 에 한정. 본 워커의 첫 실제 carve(예: N-10) 는 본 sprint 종료 후 별도 sprint 에서 진행.

## 1. 작업 단위 분해

| ID | 작업 | 영향 파일 | 의존 | 상태 |
| --- | --- | --- | --- | --- |
| WB-01 | 브랜치 + memory 디렉터리 set up | `ai-workflow/memory/opencode/work_260604-b-opencode-areas/{state.json,session_handoff.md,work_backlog.md,backlog/2026-06-04.md}` | — | done |
| WB-02 | 다른 워커 분담 + 최근 PR 패턴 분석 | (read-only) | WB-01 | in_progress |
| WB-03 | opencode 차별 영역 후보 도출 + 우선순위 사용자 confirm | (분석) | WB-02 | planned |
| WB-04 | §1.4 본문 초안 + AGENTS.md 보강 | `docs/governance/worker_division.md`, `AGENTS.md` | WB-03 | planned |
| WB-05 | release_v1_roadmap §5.1 OpenCode 행 갱신 | `docs/planning/release_v1_roadmap.md` | WB-04 | planned |
| WB-06 | 커밋 + push + PR + state/handoff final | (git + memory dir) | WB-05 | planned |

## 2. 분석 기준 (WB-02)

다음 자료를 직독:
- `docs/governance/worker_division.md` §1.1~§1.3 (Claude/Codex/Gemini) — 각 워커의 "주요 책임/작업 스타일" 절
- `AGENTS.md` 의 "Codex 전용 메모" + "Reasonix 전용 메모" — 메타 운영 패턴
- 최근 5 PR 의 `gh pr view --json body,files` — 워커별 분담 흔적
- `release_v1_roadmap.md` §3.5 NOW + LATER + v1.1/v2 항목 — 미선점 영역 후보

## 3. 후보 영역 (1차 식별)

| 영역 | 기존 워커 | 미선점/부족 | opencode 적합도 |
| --- | --- | --- | --- |
| AI/ML service (`backend-ai/`) | (없음) | v2 P3 만 — 후순위였음 | ⭐⭐⭐ |
| Cross-cutting 리팩토링 (multi-file, multi-layer) | Claude 일부 | bounded scope 분리 가능 | ⭐⭐⭐ |
| Migration & data tooling (backfill, schema 검증) | Claude 일부 | 별도 carve 가능 | ⭐⭐ |
| Observability/Performance (Prometheus, alert, dashboard) | Codex 일부 | 정밀화 + 검증 부족 | ⭐⭐ |
| Test infrastructure (UT/E2E fixture, framework) | Gemini 일부 | 보강 필요 | ⭐ |

## 4. 검증 기준 (DoD)

- [ ] §1.4 본문: 주요 책임 / 작업 스타일 / 인계 SOP / 누적 이력 placeholder
- [ ] AGENTS.md OpenCode 전용 메모에 도구/스킬/외부 contribution 정책 보강
- [ ] release_v1_roadmap §5.1 OpenCode 행이 본 sprint 결과로 갱신
- [ ] `git status` 의도치 않은 파일 변경 없음
- [ ] PR 생성 + 사용자 리뷰
