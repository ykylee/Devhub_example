# wiki-pr-update (본 저장소 측 thin wrapper)

- **문서 목적**: PR (`gh pr view <num>`) 의 metadata + touched files → `~/wiki/` 의 prs/<num>.md 신규 + log.md append + (optional) touched file re-ingest 의 **본 저장소 측 thin wrapper**.
- **SSOT**: `~/repos/my_harness/ai-workflow/skills/wiki-pr-update/SKILL.md` + `scripts/run_wiki_pr_update.py` (D-80/D-81 T-d-80-2, handoff §3 의 6 file 중 2)
- **본 wrapper 의 역할**: 본 저장소 (= DevHub) 에서 my_harness 의 Python script 호출. PR #562/563 의 3+13 skill 의 thin wrapper 정공법 정합.
- **대상 독자**: yklee, 본 저장소 작업 agent
- **상태**: active (D-80/D-81, 2026-06-11)
- **최종 수정일**: 2026-06-11
- **관련 문서**:
  - `~/repos/my_harness/ai-workflow/skills/wiki-pr-update/SKILL.md` (SSOT, 6012 lines)
  - `docs/llm-wiki/pr-update-skill.md` (본 저장소 측 사용 가이드, PR #552, handoff §3 의 본 저장소 측 thin wrapper)
  - `~/wiki/AGENTS.md` v1.5 §11.1 D-72 cross-project
  - `ai-workflow/skills/wiki-event-sync/SKILL.md` (D-86 흐름 2 의 본 저장소 측 thin wrapper — 본 skill 과 trigger 다름: wiki-event-sync = commit/push 영향 page 식별 + last_touched 갱신, wiki-pr-update = PR-merge 후 vault 의 prs/<num>.md 신규)

## 1. 사용법

### 1.1 본 저장소 측 wrapper

```bash
# 본 저장소 repo root 에서:
bash ai-workflow/skills/wiki-pr-update/scripts/wiki-pr-update \
  --pr=<num> [--project=devhub|my-harness] \
  [--reingest] [--apply] [--gh-fetch] \
  [--output=json|markdown|both] [--quiet]
```

### 1.2 SSOT 직접 호출 (본 wrapper 가 내부 실행)

```bash
python3 ~/repos/my_harness/ai-workflow/skills/wiki-pr-update/scripts/run_wiki_pr_update.py \
  --pr 552 --project=devhub
```

## 2. 옵션 (SSOT 정합)

| 옵션 | 필수 | 설명 |
|---|---|---|
| `--pr` | yes | GitHub PR number |
| `--vault-path` | no (default `~/wiki`) | vault 루트 |
| `--project` | no (default `devhub`) | `devhub` \| `my-harness` |
| `--pr-metadata` | no | `gh pr view --json` output file (wrapper 가 전달 가능) |
| `--touched-files` | no | `gh pr diff --name-only` output file |
| `--apply` | no (default dry-run) | 실제 vault 갱신 |
| `--reingest` | no | PR touched file 중 mirror-list source 와 매칭 시 `wiki-ingest-from-raw --source <file> --apply` re-run |
| `--gh-fetch` | no | wrapper 없이 직접 `gh pr view` 호출 (Python 이 gh CLI 호출) |
| `--quiet` | no | stderr 메시지 최소화 |
| `--output` | no | `json` \| `markdown` \| `both` |

## 3. 출력 (vault side effect, --apply 모드)

- `~/wiki/wiki/projects/<project>/prs/<num>.md` (frontmatter 8 key + body)
- `~/wiki/log.md` (1 line append, `## [YYYY-MM-DD] pr-update | pr-<num>: <title>`)
- `~/wiki/index.md` (prs 섹션 1줄 추가, idempotent)
- (--reingest) `wiki-ingest-from-raw --source <file> --apply` dispatch

## 4. 정책 (SSOT 정합)

- **idempotent**: `pr-<num>-<head.sha>` key — 이미 `prs/<num>.md` 가 존재하고 `last_touched >= head.sha` 이면 skip
- **raw/ 절대 수정 안 함**
- **wiki/concepts/entities/topics/sources/comparisons/query/** read-only 영역 수정 ❌
- **schema/ 수정 ❌**
- **AGENTS.md 수정 ❌**
- **다른 project 의 wiki/projects/<other>/ 수정 ❌**
- **vault Gitea remote push ❌** (사용자 수동, AGENTS.md §6.5 정책)
- **main branch 자동 머지 ❌** (PR-auto-update ≠ PR-auto-merge)

## 5. trigger

- (a) **수동**: 위 1.1 wrapper 또는 직접 Python 호출
- (b) **반자동**: Mavis session 내 commit 직후 Mavis 가 호출 (`--reingest` 모드)
- (c) **CI integration** (forward, v2.0+): 사내 self-hosted runner 도입 시 `pull_request: closed` + `workflow_dispatch`

## 6. 안전 (SSOT 정합)

- `raw/` 절대 수정 안 함
- `wiki/` 페이지 삭제 ❌
- `index.md` / `log.md` 갱신 누락 ❌
- 절대경로 stdout 출력

## 7. 의존

- `python3` 3.10+ (stdlib only)
- `gh` CLI 2.46+ (PR op; `gh pr view` + `gh pr diff` 호출)
- `~/repos/my_harness/` (skill SSOT)
- `~/wiki/` vault

## 8. 기존 `scripts/wiki-pr-update.sh` 와의 정합

본 저장소 에는 **2 layer 의 wiki-pr-update wrapper** 가 공존:

| Layer | 위치 | Backend |
|---|---|---|
| **기존 thin wrapper** (D-80) | `scripts/wiki-pr-update.sh` (L66) | `gh pr view` + `run_wiki_pr_update.py` — `--reingest` dispatch |
| **신규 SKILL wrapper** (PR #562/563 정합) | `ai-workflow/skills/wiki-pr-update/scripts/wiki-pr-update` (이 PR) | `run_wiki_pr_update.py` 만 — gh CLI 호출은 skill 내장 |

**역할 분리**: 기존 wrapper = gh CLI + skill 통합 (운영). 신규 wrapper = skill 만 (Python 이 직접 gh 호출). 둘 다 같은 SSOT 호출.

## 9. 다음 행동

- 본 저장소 측 PR 머지 시 1회 실행 검증 (PR #560 또는 #561)
- 또는 Mavis session hook 으로 자동 dispatch
- 또는 `docs/llm-wiki/Mavis-workflow.md` 9번째 문서 작성
