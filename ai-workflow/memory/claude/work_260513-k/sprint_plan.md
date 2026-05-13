# Sprint Plan — claude/work_260513-k (M3/M4 drift 정합)

- 문서 목적: M3/M4 마일스톤 정의의 3 source (`docs/development_roadmap.md`, `docs/traceability/report.md`, `ai-workflow/memory/backend_development_roadmap.md` + `state.json`) 사이 drift 를 해소한다.
- 진입 base: main HEAD `f551e6a` (PR #96 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. drift 현황

### 1.1 development_roadmap.md §3 (source-of-truth 의도)

- **M2**: 인증 및 계정 기반 완성 (done)
- **M3**: 사용자 및 조직 관리 — User CRUD UI + Org CRUD/계층 편집/리더 변경 영속화 + Sign Up (셀프 가입) + 인사 DB 연동 + CI/CD GHA
- **M4**: 실시간 대시보드 + AI Gardener + 과제 추적 + 시스템 관리

문제:
- §3 line 83 + 84 에 M3 헤더 중복.
- M3 의 "사용자/조직 관리" 항목은 사실상 M2 1차 완성 sprint (`claude/login_usermanagement_finish`, PR #85) 가 흡수 — roadmap 본문 미반영.

### 1.2 매트릭스 §2.3 (추적성 RM)

- "RM-M3-01..07 (7 항목, Sign Up + WebSocket 확장 + AI Gardener — 모두 planned)"
- §3 도메인 행 인용:
  - 회원가입 → `RM-M3-01`
  - 명령 lifecycle → `RM-M3-02 (replay, planned)`
  - 실시간 (WebSocket) → `RM-M3-02, 03 (planned)`
  - 인프라 → `RM-M3-02 (infra event publish, planned)`
  - AI → `RM-M3-04` (backend-ai placeholder)

문제: 매트릭스의 RM-M3 정의가 development_roadmap.md 의 M3 ("사용자/조직 관리") 가 아니라 M4 ("실시간 + AI") 항목을 가리킨다. cross-cut drift.

### 1.3 backend_development_roadmap.md §5 + Phase

- §5 [P1] **M3**: User CRUD + Sign Up + Org Hierarchy — development_roadmap 과 정합.
- §5 [P2] **M4**: 실시간 (Phase 8 잔여) + AI (Phase 9) — development_roadmap 과 정합.
- §5 [P3] **M4**: Gitea Reconciliation (Phase 10) + Task + Admin — development_roadmap 과 정합.

따라서 backend_roadmap 은 development_roadmap 정의와 정합. 매트릭스만 drift.

### 1.4 state.json

- `next_actions.m4_entry_candidates` 변수명 — 내용 = WebSocket 확장 + AI Gardener + Gitea Pull + (frontend) command status WS UI.
- 이 항목들은 development_roadmap §3 / backend_roadmap §5 [P2] 의 **M4** 와 정합. 변수명 정확.
- 그러나 매트릭스 §3 인용은 이 항목들을 `RM-M3-XX` 로 표기 — drift.

## 2. 결정 — source-of-truth = `development_roadmap.md` §3

- M3 = 사용자/조직 관리 + Sign Up + 인사 DB + CI/CD GHA
- M4 = 실시간 + AI Gardener + 과제 추적 + 시스템 관리

매트릭스 § 2.3 + § 3 의 RM-M3-XX 인용을 다음과 같이 재매핑:

| 기존 매트릭스 표기 | drift 정합 후 |
| --- | --- |
| 회원가입 `RM-M3-01` | `RM-M3-01` (유지 — Sign Up 은 development_roadmap M3 의 핵심 잔여) |
| 명령 lifecycle `RM-M3-02 (replay, planned)` | `RM-M4-XX` (실시간 도메인, M4) |
| 실시간 (WebSocket) `RM-M3-02, 03` | `RM-M4-XX` |
| 인프라 `RM-M3-02 (infra event publish)` | `RM-M4-XX` |
| AI Gardener `RM-M3-04` | `RM-M4-XX` |

## 3. 작업 항목

| # | 항목 | 위치 |
|---|------|------|
| 1 | development_roadmap.md §3 M3 헤더 중복 + 본문 정리 (M2 흡수 + 잔여 명확화) | docs/development_roadmap.md |
| 2 | 매트릭스 §2.3 RM-M3 + RM-M4 정의 명시 (development_roadmap 정합) | docs/traceability/report.md §2.3 |
| 3 | 매트릭스 §3 의 도메인 행 RM-M3 → RM-M4 갱신 (명령 / 실시간 / 인프라 / AI) | docs/traceability/report.md §3 |
| 4 | state.json `m4_entry_candidates` 정합 + `m3_entry_candidates` 분리 (Sign Up + 인사 DB) | ai-workflow/memory/state.json |
| 5 | session_handoff.md / work_backlog.md M3/M4 표기 정합 | ai-workflow/memory/*.md |
| 6 | backend_development_roadmap.md §5 의 phase ↔ M3/M4 매핑 명확화 (Phase 8/9/10 ↔ M4) | ai-workflow/memory/backend_development_roadmap.md |
| 7 | 매트릭스 §6 변경 이력 + main flat sync (PR #96 흡수) | docs/traceability/report.md §6 |
| 8 | PR + 2-pass + merge | |

## 4. 검증

- 코드 변경 0 (docs only)
- `go test ./...` PASS sanity
- 매트릭스 §3 의 모든 RM-MX-XX 인용이 정합 source-of-truth 와 일치
