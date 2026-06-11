# wiki-lint (본 저장소 측 thin wrapper)

- **문서 목적**: LLM Wiki vault 의 무결성 검사 (L01~L10) 의 **본 저장소 측 thin wrapper**. Per-project 지원 (D-72).
- **SSOT**: `~/repos/my_harness/ai-workflow/skills/wiki-lint/SKILL.md` + `scripts/run_wiki_lint.py` (my_harness 의 cross-project SSOT, D-71 ~ D-79)
- **본 wrapper 의 역할**: 본 저장소 (= DevHub) 에서 my_harness 의 Python script 호출. PR #562/563 의 3+13 skill 의 thin wrapper 정공법 정합.
- **대상 독자**: yklee, 본 저장소 작업 agent
- **상태**: active (D-72 per-project 지원, 2026-06-11)
- **최종 수정일**: 2026-06-11
- **관련 문서**:
  - `~/repos/my_harness/ai-workflow/skills/wiki-lint/SKILL.md` (SSOT)
  - `docs/llm-wiki/lint-config.toml` (per-project rule override, DevHub 의 L07 ADR 면제)
  - `docs/llm-wiki/operation-sop.md` §3 (lint SOP)
  - `~/wiki/AGENTS.md` v1.5 (lint 규약)
  - `~/wiki/schema/lint_rules.md` (L01~L10 SSOT)

## 1. 사용법

### 1.1 본 저장소 측 wrapper

```bash
# 본 저장소 repo root 에서:
bash ai-workflow/skills/wiki-lint/scripts/wiki-lint \
  --vault-path ~/wiki \
  --project=devhub \
  --project-config=docs/llm-wiki/lint-config.toml \
  [--rules=L01,L02,...] [--output=json|markdown|both] [--quiet]
```

### 1.2 SSOT 직접 호출 (본 wrapper 가 내부 실행)

```bash
python3 ~/repos/my_harness/ai-workflow/skills/wiki-lint/scripts/run_wiki_lint.py \
  --vault-path ~/wiki --project=devhub --project-config=docs/llm-wiki/lint-config.toml
```

## 2. 옵션 (SSOT 정합)

| 옵션 | 필수 | 설명 |
|---|---|---|
| `--vault-path` | yes | vault 루트 (예: `~/wiki`) |
| `--project` | no | `my-harness` \| `devhub`. 기본: 전체 project 자동 발견 |
| `--project-config` | no | per-project rule override (TOML). 기본: `wiki/projects/<project>/.wiki-lint.toml` 자동 |
| `--rules` | no | 검사할 규칙 ID 콤마 구분 (기본: 전체) |
| `--output` | no | `json` \| `markdown` \| `both` |
| `--quiet` | no | stderr 메시지 최소화 |

## 3. DevHub 의 per-project config

`docs/llm-wiki/lint-config.toml` 의 DevHub 정공법:

- `[project] name = "devhub"`
- `[project] ssot_path = "~/repos/Devhub_example_minimax"`
- `[project] mirror_path = "~/wiki/raw/projects/devhub"`
- `[project] index_path = "~/wiki/index.md"`
- `[project] tier = "sa-private-wiki"`
- `[rules.L07] skip_paths = ["wiki/projects/devhub/sources/ADR-*.md"]` (ADR 의도적 supersede 면제)
- `[rules.L07] skip_if_frontmatter = ["supersedes"]`
- `[rules.L10] devhub_adr_source_pattern = "~/wiki/raw/projects/devhub/docs/adr/*.md"` (1:1 mirror 자동 검증)

## 4. 정책 (SSOT 정합)

- **L01~L10**: 위 DevHub config 적용
- **L11/L12 (사내 패턴 검출 + 사용자 정의)**: D-72 응답 §3 + yklee 결정으로 **미사용** (vault = Gitea private 만)
- **idempotent**: dry-run + apply 모두 가능. apply 시 wiki page 갱신 없음 (read-only + report)

## 5. trigger

- (a) **수동**: 위 1.1 wrapper 또는 직접 Python 호출
- (b) **반자동**: PR merge 직후 / 매월 1회 (drift 검증)
- (c) **CI integration** (forward): `ci.yml` 의 workflow-lint job 의 env 에 step 추가 — D-72 forward path

## 6. 안전 (SSOT 정합)

- `raw/` 절대 수정 안 함
- `wiki/` 페이지 수정 ❌
- stdout JSON / markdown report 만 출력

## 7. 의존

- `python3` 3.10+ (stdlib only)
- `~/repos/my_harness/` (skill SSOT)
- `~/wiki/` vault

## 8. 다음 행동

- 본 저장소 측 CI integration (`ci.yml` 의 workflow-lint job 에 step 추가)
- 또는 vault dry-run 검증 (Phase 1 mirror 후, 사용자 confirm)
- 또는 `wiki/projects/<project>/.wiki-lint.toml` 자동 생성 (per-project default) — my_harness 측 wiki-lint 의 default 동작과 정합
