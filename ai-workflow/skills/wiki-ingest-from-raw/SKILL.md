# wiki-ingest-from-raw (본 저장소 측 thin wrapper)

- **문서 목적**: raw/projects/<project>/ → wiki/projects/<project>/sources/ 자동 ingest 의 **본 저장소 측 thin wrapper**. D-72 Phase 3 의 진짜 통합 wrapper.
- **SSOT**: `~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/scripts/run_wiki_ingest.py` (D-79 의 6 file 중 1, D-72 §11.1 정공법)
- **본 wrapper 의 역할**: 본 저장소 (= DevHub) 에서 my_harness 의 Python script 호출. PR #562/563 의 3+13 skill 의 thin wrapper 정공법 정합.
- **대상 독자**: yklee, 본 저장소 작업 agent, Mavis / Mavis Code
- **상태**: active (D-72 Phase 3, 2026-06-11)
- **최종 수정일**: 2026-06-11
- **관련 문서**:
  - `~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/` (SSOT, D-79 작성)
  - `docs/llm-wiki/ingest-skill.md` (본 저장소 측 사용 가이드, PR #552)
  - `~/wiki/AGENTS.md` v1.5 §2.1 (Ingest 6 step)
  - `~/repos/my_harness/ai-workflow/core/wiki_ingest_skill_spec.md` (D-72 §1-§11, §278 lines)

## 1. 사용법

### 1.1 본 저장소 측 wrapper (= `scripts/wiki-ingest-from-raw`)

```bash
# 본 저장소 repo root 에서:
bash ai-workflow/skills/wiki-ingest-from-raw/scripts/wiki-ingest-from-raw \
  --project=devhub \
  [--source <rel_path>] [--all] [--limit N] \
  [--apply] [--skip-lint] [--output=json|markdown|both] [--quiet]
```

### 1.2 SSOT 직접 호출 (본 wrapper 가 내부 실행)

```bash
python3 ~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/scripts/run_wiki_ingest.py \
  --project=devhub --all
```

## 2. 옵션 (SSOT 정합)

| 옵션 | 필수 | 설명 |
|---|---|---|
| `--vault-path` | no (default `~/wiki`) | vault 루트 |
| `--project` | yes | `devhub` \| `my-harness` |
| `--source` | no | 1 file ingest (raw/ 상대 경로) |
| `--all` | no | project 의 모든 source 일괄 ingest |
| `--limit` | no | `--all` 시 최대 N건 |
| `--apply` | no (default dry-run) | 실제 ingest (default = dry-run, 변경 없음) |
| `--skip-lint` | no | post-ingest wiki-lint skip |
| `--output` | no | `json` \| `markdown` \| `both` |
| `--quiet` | no | stderr 메시지 최소화 |

## 3. 출력

- `~/wiki/wiki/projects/<project>/sources/<title>.md` (frontmatter 8 key + body)
- `~/wiki/log.md` (1 line append, `## [YYYY-MM-DD] ingest | <project>/<source>`)
- `~/wiki/index.md` (sources 섹션 1줄 추가, idempotent)
- cross-ref 자동 populate (`related: [[<related source>]]`)
- stdout JSON / markdown report

## 4. 정책 (SSOT 정합)

- **idempotent**: 동일 source 의 frontmatter `last_touched >= ...` 면 skip
- **raw 절대 수정 안 함**
- **wiki/concepts/, entities/, topics/, comparisons/ 등 read-only 영역 수정 ❌** (Ingest 의 책임 외)
- **schema/ 수정 ❌**
- **AGENTS.md 수정 ❌**
- **다른 project 의 wiki/projects/<other>/ 수정 ❌** (cross-project 분리)
- **vault Gitea remote push ❌** (사용자 수동)

## 5. trigger

- (a) **수동**: 위 1.1 wrapper 또는 직접 Python 호출
- (b) **반자동**: PR merge 후 `--reingest` 모드 (wiki-pr-update 의 dispatch)
- (c) **자동 (v2.0+)**: Mavis session hook (flows 1 → 2/3 chain)

## 6. 안전 (SSOT 정합)

- `raw/` 절대 수정 안 함
- `wiki/` 페이지 삭제 ❌
- `index.md` / `log.md` 갱신 누락 ❌
- 절대경로 stdout 출력

## 7. 의존

- `python3` 3.10+ (stdlib only)
- `~/repos/my_harness/` (skill SSOT)
- `~/wiki/` vault
- 본 저장소 의 `scripts/wiki-sync-devhub.sh` (raw mirror 의 source, step 1)

## 8. 기존 `scripts/wiki-ingest-from-raw.sh` 와의 정합

본 저장소 에는 **2 layer 의 wiki-ingest-from-raw wrapper** 가 공존:

| Layer | 위치 | Backend |
|---|---|---|
| **기존 thin wrapper** (D-72 Phase 3) | `scripts/wiki-ingest-from-raw.sh` (L57) | `wiki-sync-devhub.sh` (raw) + `run_wiki_ingest.py` (ingest) — 2-step 통합 |
| **신규 SKILL wrapper** (PR #562/563 정합) | `ai-workflow/skills/wiki-ingest-from-raw/scripts/wiki-ingest-from-raw` (이 PR) | `run_wiki_ingest.py` 만 — ingest step 만 (raw 는 caller) |

**역할 분리**: 기존 wrapper = raw + ingest 2-step 통합 (운영). 신규 wrapper = ingest step 만 (자동화/세밀 제어). 둘 다 같은 SSOT 호출.

## 9. 다음 행동

- 본 저장소 측 vault dry-run (Phase 1 mirror 후 사용자 confirm)
- 또는 Mavis session hook 으로 자동 dispatch (raw 작성 → 즉시 ingest chain)
- 또는 `docs/llm-wiki/Mavis-workflow.md` 9번째 문서 작성 (3축 통합 SSOT)
