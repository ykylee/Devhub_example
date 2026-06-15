---
type: decision
status: accepted
adr_id: ADR-001
decided_at: 2026-05-02
accepted_in: v0.5.2
alternatives_considered: [single-layer, two-layer-source-state, two-layer-source-knowledge]
related_pages: [concepts/project-architecture, concepts/wiki-source-rule-r9, concepts/memory-3-state-lifecycle]
created: 2026-06-12
updated: 2026-06-12
---

# ADR-001: Source / State / Knowledge 3-layer 분리

## Status

Accepted (v0.5.2, commit `96431f1 refactor(workflow): separate source from runtime layer`). 본 wiki 페이지는 공식 ADR-001 ([docs/architecture/ADR-001-source-state-knowledge-3-layer-separation.md](../../../docs/architecture/ADR-001-source-state-knowledge-3-layer-separation.md)) 의 wiki 형식 mirror.

| 메타 필드 | 값 |
|---|---|
| Supersedes | — |
| Superseded by | — |
| 적용 버전 | v0.5.2 (commit `96431f1`) |
| hotfix 결합 | v0.5.10.1 smart update 정책 |

## Context

v0.5.1 까지 `standard_ai_workflow` 저장소는 단일 레이어로 운영됐다. bootstrap 이 `ai-workflow/` 를 생성할 때 (a) workflow kit 원본 파일, (b) 사용자가 채워야 할 workflow state (handoff, backlog, state.json), (c) 영구 지식 문서 (PROJECT_PROFILE, README 등) 가 한 디렉토리에 섞여 있었다.

이로 인해 세 가지 운영 pain 이 누적됐다.

- **갱신 정책 혼란**: 모든 파일이 git 으로 tracked 되므로, 사용자가 handoff 를 갱신할 때마다 PR 리뷰가 필요했다. workflow state 는 휘발성/진행성 데이터인데 영구 지식과 같은 변경 정책이 적용된 셈.
- **재배포 어려움**: 새 버전을 적용할 때 (`apply_workflow_upgrade`) 가이드 문서와 사용자 state 가 같은 트리에 있어서, source 갱신이 사용자 데이터를 덮어쓸 위험이 있었다.
- **CI 부담**: smoke test 가 모든 markdown 을 검증하므로, 사용자 작성 handoff 까지 검증 대상에 포함되어 false positive 가 잦았다.

[[concepts/project-architecture]] 가 본 ADR + ADR-004 + memory 3-state lifecycle 을 합쳐 통합 architecture view 를 제시한다.

## Decision

저장소를 **3-layer** 로 분리한다. 각 레이어는 위치 / 책임 / 갱신 정책 / git 추적이 명확히 구분된다.

| Layer | 위치 | 책임 | 갱신 정책 | Git tracked |
|---|---|---|---|---|
| **Source (SSOT)** | `workflow-source/` | 모든 skill / MCP / template / helper / doc 의 원본. workflow kit 의 단일 진실 공급원 | PR 리뷰 + merge | ✅ |
| **State (runtime)** | `ai-workflow/memory/` | 세션별 workflow state (handoff, backlog, state.json). 휘발성/진행성 데이터 | 자유롭게 갱신, PR 리뷰 불필요 | ⚠️ selective (`.gitignore` 로 대부분 제외, README 만 tracked) |
| **Knowledge (영구)** | `docs/` | 영구 지식 자산 (PROJECT_PROFILE, CODE_INDEX, RELEASE 등). maintainer 작성 | PR 리뷰 + merge | ✅ |

### 핵심 경계

- **`workflow-source/` 가 SSOT**: 사용자가 bootstrap 으로 받는 모든 파일은 이 디렉토리에서 렌더링됨. `ai-workflow/` 나 `docs/` 파일을 직접 수정해서 다시 bootstrap 하면 덮어써짐.
- **`ai-workflow/memory/` 는 자유 영역**: 사용자가 handoff 를 갱신하거나 backlog 에 task 를 추가할 때 PR 리뷰 없이 commit 가능. [[concepts/memory-3-state-lifecycle]] 의 active state 가 이 레이어에서 운용된다.
- **`docs/` 는 maintainer 영역**: 영구 지식은 PR 리뷰를 거쳐 `main` 머지. `docs/README.md` 가 governance 를 명시.

### Layer 별 owner 와 변경 빈도

| Layer | Owner | 변경 빈도 | Review 정책 |
|---|---|---|---|
| **Source** | maintainer (Sisyphus) | 릴리스 단위 (주~월 단위) | contributor PR + maintainer approve |
| **State** | 사용자 / AI agent | 세션 단위 (수 분~수 시간) | PR 불필요, 자유 commit |
| **Knowledge** | maintainer (Sisyphus) | 분~주 단위 (영구 자산 추가·갱신 시) | maintainer self-review + 24h cool-down |

### PRESERVE_RELATIVE_PATHS

`workflow_kit.constants.PRESERVE_RELATIVE_PATHS` (`workflow_kit/constants.py:47`) 가 `ai-workflow/memory/`, `ai-workflow/WORKFLOW_INDEX.md`, `ai-workflow/README.md` 를 명시. upgrade / bootstrap 시 이 경로는 절대 덮어쓰지 않음. v0.5.10.1 hotfix 의 smart update 정책과 결합 — PRESERVE 경로 파일은 marker/hash 와 무관하게 무조건 보존.

| PRESERVE 경로 | 이유 |
|---|---|
| `ai-workflow/memory/` | 사용자 session state (handoff, backlog, state.json). 자유 영역. |
| `ai-workflow/WORKFLOW_INDEX.md` | workflow 진입 인덱스. 사용자 편집 흔적 가능. |
| `ai-workflow/README.md` | kit 의 진입점. cross-layer 참조 marker 필요. |

### 후속 ADR 과의 관계

본 ADR 의 3-layer 결정은 후속 결정들의 토대가 된다. [[concepts/wiki-source-rule-r9]] 이 본 ADR 의 state layer 정책 위에서 wiki 의 source ingest 규칙을 정의하고, [[concepts/project-architecture]] 가 본 ADR + ADR-004 + memory 3-state lifecycle 을 통합 view 로 묶어 제시한다.

## Consequences

### Positive

- **재배포 안전**: `apply_workflow_upgrade` 가 source 만 갱신하고 사용자 state 는 보존됨.
- **PR 부담 경감**: 사용자의 일상적 workflow state 갱신이 PR 리뷰를 거치지 않음.
- **CI 정확도**: smoke test 가 source 만 검증하므로 false positive 가 줄어듦.
- **책임 명확화**: 각 레이어의 owner / 변경 빈도 / 리뷰 정책이 달라져 운영 부담이 owner 별로 분산됨.
- **Evolution safety**: 3-layer 경계가 명확해져 wiki layer (ADR-004), memory 3-state (R8), read-only MCP (ADR-003) 같은 후속 확장이 layer violation 없이 가능.

### Negative / Trade-offs

- **사용자 학습 비용**: "어떤 파일을 어디 두는지" 사용자가 처음에 헷갈릴 수 있음. `docs/PROJECT_PROFILE.md` 가 이 경계를 명시.
- **cross-layer 의존 명시 필요**: `ai-workflow/README.md` 가 workflow 의 진입점인데, 이 파일은 state layer 에 있지만 workflow 의 source 에서도 참조됨. 양쪽에 동기화된 marker 가 필요할 수 있음 (현 v0.5.10 은 marker 만으로 해결).
- **bootstrap 두 번 실행 시 충돌 가능성**: 사용자가 `ai-workflow/README.md` 를 손으로 수정한 후 다시 bootstrap 하면 smart update 가 marker 가 같고 hash 가 다른 경우 `UPDATED` 결정 → 사용자 편집 손실. `--preserve-user-edits` 옵션은 후속 (v0.5.11+).

## Alternatives Rejected

| Option | Why rejected |
|---|---|
| **Single-layer** (v0.5.1 baseline) | 갱신 정책·재배포·CI 세 pain 이 누적된 상태 그대로. source/state/knowledge 의 책임·갱신 빈도·review 정책이 모두 달라 단일 트리로 강제할 근거 없음. |
| **Two-layer: source + state** | knowledge (`docs/`) 가 source 와 같은 PR review 정책을 그대로 안게 되어, 사용자 프로젝트 runbook 도 contributor PR 게이트를 거치게 됨. maintainer 부담 가중. |
| **Two-layer: source + knowledge** | state (`ai-workflow/memory/`) 가 source 와 같은 변경 정책이 적용되어, 사용자 handoff 갱신마다 PR 리뷰가 강제됨. workflow state 의 휘발성/진행성 성격 무시. |
| **Two-layer: state + knowledge** | source 가 따로 분리되지 않아 `apply_workflow_upgrade` 가 사용자 state 를 덮어쓸 위험이 해소되지 않음. SSOT 부재. |
| **README-only governance** (분류만 governance §1 에 명시) | 디렉토리·정책이 코드 차원에서 강제되지 않아, 신규 contributor 가 governance 를 무시하기 쉬움. layer 경계가 구조적으로 enforcement 되도록 3-layer 분리가 필요. |

## References

- [공식 ADR-001](../../../docs/architecture/ADR-001-source-state-knowledge-3-layer-separation.md) — 본 결정의 정식 기록 (Context / Decision / Consequences / 후속 작업)
- commit `96431f1 refactor(workflow): separate source from runtime layer` — 본 ADR 의 머지 commit
- `workflow_kit/constants.py:47` — `PRESERVE_RELATIVE_PATHS` 정의
- [[concepts/project-architecture]] — 3-Layer + LLM Wiki + Memory 3-State 통합 view
- [[concepts/wiki-source-rule-r9]] — wiki layer 의 source ingest 규칙 (state layer 정책 위에서 동작)
- [[concepts/memory-3-state-lifecycle]] — `ai-workflow/memory/` 의 active/archive/release lifecycle
- `docs/README.md` §1 — 문서 분류 governance
- `workflow-source/core/workflow_state_vs_project_docs.md` — state vs project docs 경계 가이드
- Beta v0.5.10.1 release note — smart update 정책과 PRESERVE 결합
