---
type: topic
status: active
last_ingested_from: vendor/standard_ai_workflow/harnesses/minimax-code/README.md + vendor/standard_ai_workflow/core/global_workflow_standard.md + vendor/standard_ai_workflow/core/orchestrator_subagent_contract_v1.md + vendor/standard_ai_workflow/core/workflow_task_modes.md
related_pages: [concepts/devhub-overview, decisions/v0.7.17-import, patterns/in-repo-redirect]
created: 2026-06-15
updated: 2026-06-15
active_since: 2026-06-15
active_reason: "v0.7.17 vendor import (PR #600) + v0.7.17 SSOT 동기화 (PR #601) 의 본 저장소 측 SSOT"
git_commit: baf1cf24
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:37:45Z
mirror_dirty: |
---

# Standard AI Workflow Vendor (L1 topic, in-repo)

## TL;DR

`ykylee/standard_ai_workflow` 는 Mavis (MiniMax Code) 의 표준화 워크플로우를 발매하는 패키지. v0.7.17 (commit 4d09dee) 가 본 저장소 의 *vendor import* 기준. *발매자* (standard_ai_workflow) vs *소비자* (DevHub) 의 관계, *본 저장소 의 SSOT 1차 출처* (vendor 의 40+ core spec + DevHub 의 1차 layer 운영 문서). Mavis 의 4 sub-agent role + 6 task mode + SSOT 1차 출처 (maturity_matrix.json) 가 본 저장소 의 운영 패턴.

## 1. vendor 구조 (28M → 5.7M git pack, 746 file tracked)

```
vendor/standard_ai_workflow/
├── ai-workflow/             # vendor 의 자체 active memory + wiki mini
│   ├── memory/{active,archive,codex,gemini,plans,release}/
│   └── wiki/{concepts,decisions,entities,patterns,topics,sources}/
├── core/                    # 40+ spec doc (SSOT 1차 출처)
│   ├── global_workflow_standard.md
│   ├── orchestrator_subagent_contract_v1.md
│   ├── workflow_task_modes.md
│   ├── workflow_agent_topology.md
│   ├── workflow_harness_distribution.md
│   ├── maturity_matrix.json
│   ├── mcp_installation_by_harness.md
│   ├── phase{5,6,9,11}_*.md
│   └── ...
├── examples/                # 290K
├── extensions/
├── global-snippets/
├── harnesses/               # 6 harness overlay
│   ├── _template/
│   ├── antigravity/
│   ├── codex/
│   ├── gemini-cli/
│   ├── minimax-code/        # Mavis (MiniMax Code) overlay
│   │   ├── README.md
│   │   └── apply_guide.md
│   ├── opencode/
│   └── pi-dev/
├── mcp_servers/
├── prompts/                 # 12K
│   ├── code_worker_prompt.md
│   ├── doc_worker_prompt.md
│   └── validation_worker_prompt.md
├── pyproject.toml           # v0.7.17
├── releases/                # 392K
│   ├── Beta-v0.7.17.md
│   └── ...
├── schemas/
├── scripts/                 # 344K
├── skills/
├── templates/
├── tests/                   # 840K (v0.7.17 11 smoke + 70+ check)
├── workflow_kit/            # 640K
└── CHANGELOG.md
```

## 2. 발매자 (vendor) vs 소비자 (DevHub) 의 의미 분리

| 차원 | standard_ai_workflow (발매자) | DevHub (소비자) |
|---|---|---|
| `ai-workflow/` 의미 | 도구/스킬 발매 패키지 (workflow-source 의 sibling) | 프로젝트 운영 메모리/스크립트/스킬/테스트 인프라 (AGENTS.md workflow layer) |
| `workflow-source/` 의미 | raw source code (build target) | (없음) |
| `vendor/standard_ai_workflow/` 의미 | (없음) | 발매자 패키지의 격리 import (read-only reference + 도구 사용) |
| 1차 출처 SSOT | `core/maturity_matrix.json` | `AGENTS.md` + `ai-workflow/MEMORY_GOVERNANCE.md` + `ai-workflow/minimax_code_workflow.md` + `ai-workflow/memory/PROJECT_PROFILE.md` |
| Wiki 위치 | `ai-workflow/wiki/` (in-repo mini) | `ai-workflow/wiki/` (in-repo 운영) + `~/wiki/` (사용 안 함, v0.7.17 결정) |
| 외부 vault `~/wiki/` | (vendor 도구 자체 가 외부 vault 호출 ❌) | ❌ (v0.7.17 redirect 결정) |

## 3. Mavis 운영 패턴 (vendor minimax-code harness)

### 3.1 진입 파일

- `AGENTS.md` (Codex/OpenCode 와 공통): 워크플로우 규칙 요약
- `MiniMax.md` (MiniMax Code 전용): 메인 orchestrator 운영 원칙 + 한국어 보고 규칙

DevHub 의 경우 *AGENTS.md 가 본 저장소 workflow layer* (worker_division.md, v0.1.0 roadmap, 2-tier 형상관리, 워커별 메모 Historical) 로 흡수. Mavis 운영 패턴 = `ai-workflow/minimax_code_workflow.md` (PR #601 v0.7.17 동기화, 12 섹션, 352 line, 상태 active).

### 3.2 워커 오버레이 (`.minimax/agents/`, vendor)

| 파일 | 역할 | DevHub 측 |
|---|---|---|
| `workflow-orchestrator.md` | 메인 orchestrator 페르소나 | Mavis root session (mini-mavis-team skill) |
| `workflow-worker.md` | 워커 공통 운영 계약 | (Mavis 의 sub-agent 가 본 패턴 따름) |
| `workflow-doc-worker.md` | 문서 정합성 / 메타데이터 / 카탈로그 동기화 | 본 저장소 의 doc-sync skill |
| `workflow-code-worker.md` | 코드 구현 / 정밀 리팩토링 | 본 저장소 의 robust_patcher + code-index-update skill |
| `workflow-validation-worker.md` | 테스트/스모크 실행 + 결과 기록 | 본 저장소 의 validation-plan skill + check_*.py |

### 3.3 sub-agent contract v1 (vendor)

- **4 role**: orchestrator (메인) / doc-worker / code-worker / validation-worker / workflow-worker (임시)
- **위임 입력/출력 스키마** (`orchestrator_subagent_contract_v1.md` §4-§5): delegation_id + issued_at + task_type (4 enum) + brief + constraints + inputs + expected_outputs + validation + deadline_hint + required_model_tier
- **멀티 컴포넌트 fan-out** (§4.2 v0.5.7): sub_tasks + parent_delegation_id + sub_id
- **본 저장소 매핑**: mavis-team prompt 6섹션 (TASK/EXPECTED OUTCOME/REQUIRED TOOLS/MUST DO/MUST NOT DO/CONTEXT) ↔ contract v1 field 1:1

## 4. 작업 모드 6종 (vendor workflow_task_modes.md)

| 모드 | 목적 | 본 저장소 매핑 |
|---|---|---|
| Analysis | 코드베이스 구조/의존성/로직 파악 | code-index-update + mavis-team explore |
| Requirements | 사용자 니즈 정의/제약 | mavis-team general + ask_user |
| Design | 신규 기능 청사진/아키텍처 | docs/governance/adr/ + mavis-team general (main-tier) |
| Planning | 실행 가능한 태스크로 분해 | backlog-update skill + ai-workflow/memory/feat/*/work_backlog.md |
| Implementation | 실제 코드 작성/단위 테스트 | mavis-team mini-coder-max + validation-plan |
| Refactoring | 기능 유지/코드 품질 개선 | mavis-team mini-coder-max + merge-doc-reconcile |

## 5. vendor 의 release cadence

- **현재 vendor import**: v0.7.17 (commit 4d09dee, 2026-06-15)
- **vendor 의 latest (2026-06-15 시점)**: v0.7.21 (commit fa329b1) — 4 release ahead
- **갱신 SOP**: vendor 의 새 release 시 `git pull` + 본 저장소 의 vendor/ 갱신 + 16/16 smoke 회귀 + 메모리 anchor 갱신
- **update_strategy** (`.upstream-url`): `manual (릴리즈 노트 + breaking change 발생 시 gh pr create --base main 후 vendor/ 갱신)`

## 6. follow-up (별도 PR)

- vendor v0.7.18~v0.7.21 의 release note 흡수 (1 PR)
- 7 script 의 my_harness wiki-* skill 호출의 in-repo vendor tool redirect
- vendor 의 emit_wiki_l2_body.py 의 *devhub project mode* adapter (L1/L2 page 자동 emit)
- v0.7.15 atomic_write / v0.7.15 changelog-gen / v0.7.16 workflow-doctor config thresholds 의 DevHub 자체 도구 도입
