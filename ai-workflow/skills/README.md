# ai-workflow/skills/ — 본 저장소 (= DevHub) 측 skill 색인 (D-86 폐기 반영, 2026-06-11)

- **문서 목적**: 본 저장소 의 `ai-workflow/skills/` 디렉터리에 들어가는 skill 들의 색인.
- **SSOT 위치 (wiki 관련)**: `~/wiki/skills/` (vault 의 cross-project SSOT, vault Gitea private) — **2026-06-11 사용자 결정: wiki 관련 skill SSOT 는 `~/wiki` 에서 일괄 관리**.
- **D-86 폐기 (2026-06-11)**: 본래 본 디렉터리는 흐름 1/2/3 (`wiki-prompt-log` / `wiki-event-sync` / `wiki-query-helper`) 의 thin wrapper 를 두는 D-86 정공법을 따랐음. **2026-06-11 D-86 thin wrapper 3종 폐기 결정**으로 모두 삭제. 흐름 1/2/3 호출 시 `~/wiki/skills/<name>/SKILL.md` + `scripts/<name>.py` SSOT 직접 호출.
- **밴치마킹 출처**: `~/repos/my_harness/ai-workflow/skills/` (D-86 동일 패턴. my_harness 측 동일 폐기는 별도 작업, 본 저장소 scope 외).
- **대상 독자**: yklee, Mavis / Mavis Code, 본 저장소 작업 agent
- **상태**: active (D-86 폐기 반영, 2026-06-11)
- **최종 수정일**: 2026-06-11
- **관련 문서**:
  - `~/wiki/skills/README.md` (vault 측 SSOT 색인, wiki 관련 skill 의 정공법)
  - `~/wiki/AGENTS.md` v1.5
  - `~/wiki/WORKFLOW.md` (3개 흐름의 통합 정의)
  - `docs/llm-wiki/README.md` (D-72 ~ D-80 의 본 저장소 측 SSOT — `docs/llm-wiki/` 와 역할 분리)
  - `docs/llm-wiki/{ingest-skill,query-skill,pr-update-skill,lint-skill}.md` (D-72 ~ D-80 가이드)

## 1. 본 저장소 측 skill 목록 (2026-06-11 갱신)

본 저장소 의 `ai-workflow/skills/` 는 **wiki 비관련 + wiki-관련 정공법 도구** 의 본 저장소 측 thin wrapper / consumer 만 둠. **흐름 1/2/3 (D-86) 의 본 저장소 측 thin wrapper 는 2026-06-11 폐기**.

### 1.1 wiki-* skill (vault SSOT 직접 호출 — 본 저장소 wrapper 없음)

| Skill | 영역 | SSOT |
|---|---|---|
| `wiki-event-sync` | 흐름 2 (git 이벤트 → wiki 갱신) | `~/wiki/skills/wiki-event-sync/` |
| `wiki-prompt-log` | 흐름 1 (사용자 prompt → raw 기록) | `~/wiki/skills/wiki-prompt-log/` |
| `wiki-query-helper` | 흐름 3 (wiki 후보 추천) | `~/wiki/skills/wiki-query-helper/` |
| `wiki-source-sync` | 흐름 1.5 (raw → wiki frontmatter 자동) | `~/wiki/skills/wiki-source-sync/` |

### 1.2 wiki-* skill (D-72 ~ D-80 잔여 — my_harness SSOT 의 본 저장소 측 thin wrapper)

| Skill | 영역 | SSOT | 본 저장소 측 형태 |
|---|---|---|---|
| `wiki-ingest-from-raw` | D-72 ~ D-80 — raw → wiki page 합성 | `~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/` (my_harness SSOT) | **thin wrapper** (D-72 정공법) |
| `wiki-lint` | D-72 — wiki lint (L01~L10) | `~/repos/my_harness/ai-workflow/skills/wiki-lint/` (my_harness SSOT) | **thin wrapper** (D-72 정공법) |
| `wiki-pr-update` | D-80 — PR-merge 후 vault 갱신 | `~/repos/my_harness/ai-workflow/skills/wiki-pr-update/` (my_harness SSOT) | **thin wrapper** (D-80 정공법) |
| `wiki-query` | D-79 — vault 정식 query (--file filing) | `~/repos/my_harness/ai-workflow/skills/wiki-query/` (my_harness SSOT) | **thin wrapper** (D-79 정공법) |

각 thin wrapper 의 상세: `wiki-{ingest-from-raw,lint,pr-update,query}/SKILL.md` + `scripts/<name>`.

## 2. 본 저장소 측 thin wrapper 정공법 (D-72 ~ D-80 잔여 4종)

각 skill 의 디렉터리는:
- `SKILL.md` — 본 저장소 측 사용 가이드 (my_harness SSOT 참조 + 본 저장소 고유 노트)
- `scripts/<skill-name>` — bash thin wrapper (my_harness 의 `.py` 직접 호출)
- Python 구현은 my_harness 에만 둠 (cross-project SSOT)

이유:
- 단일 진실: my_harness 가 cross-project SSOT (my_harness, devhub, cross 동거)
- 본 저장소 의 wrapper 는 단순 dispatch — 변경 빈도 낮음
- yklee 가 my_harness 측 Python 수정 → 본 저장소 wrapper 자동 효력

**D-86 thin wrapper 폐기 사유 (2026-06-11)**: wiki 관련 SSOT 의 단일 source-of-truth 위치를 `~/wiki/skills/` 로 확정. my_harness 측 wrapper 와 본 저장소 측 wrapper 의 중복 유지 비용 (양쪽 PR 정합, push drift, my_harness 측 Python 수정 → 두 repo 의 wrapper 갱신) 이 단일 SSOT 의 이점보다 작다고 판단. 흐름 1/2/3 호출은 `~/wiki/skills/<name>/scripts/<name>.py` 직접 호출로 단순화.

## 3. 본 저장소 측 `docs/llm-wiki/` 와의 역할 분리

본 저장소 에는 **두 가지 SSOT 디렉터리**가 의도적으로 공존 (D-72 결정):

| 디렉터리 | 영역 | skill 본문 |
|---|---|---|
| **`docs/llm-wiki/`** (8 file) | D-72 ~ D-80 — vault raw mirror / ingest / D-79 query / D-80 pr-update 의 **본 저장소 측 정공법/가이드** | `scripts/wiki-{sync,ingest-from-raw,query,pr-update}.sh` (D-72 ~ D-80 thin wrapper) |
| **`ai-workflow/skills/`** (이 directory) | D-72 ~ D-80 잔여 — ingest / lint / pr-update / query 의 **본 저장소 측 thin wrapper** (D-86 thin wrapper 3종은 2026-06-11 폐기) | `scripts/wiki-{ingest-from-raw,lint,pr-update,query}` (D-72 ~ D-80 thin wrapper) |

**겹치지 않음**:
- D-72 ~ D-80 = vault 의 raw mirror (7 패턴 docs) + ingest + D-79 query + D-80 pr-update + lint
- D-86 (**2026-06-11 폐기**) = 흐름 1 (prompt log) + 흐름 2 (git event sync) + 흐름 3 (candidate query) — **SSOT 는 `~/wiki/skills/` 단일**

## 4. 사용법 예시 (D-86 폐기 후)

### 4.1 흐름 1 (prompt log) — Mavis 가 사용자 prompt 기록

```bash
# SSOT 직접 호출 (D-86 thin wrapper 폐기 후):
python3 ~/wiki/skills/wiki-prompt-log/scripts/wiki-prompt-log.py \
  --project=devhub --slug=<kebab> --intent="<1~3문장>"
```

### 4.2 흐름 2 (event sync) — commit 직후 자동 또는 수동

```bash
# SSOT 직접 호출 (D-86 thin wrapper 폐기 후):
python3 ~/wiki/skills/wiki-event-sync/scripts/wiki-event-sync.py \
  --op=commit --project=devhub --ref=HEAD
```

### 4.3 흐름 3 (query helper) — Mavis 가 자동 또는 사용자 명시

```bash
# SSOT 직접 호출 (D-86 thin wrapper 폐기 후):
python3 ~/wiki/skills/wiki-query-helper/scripts/wiki-query-helper.py \
  "Keycloak RBAC" --project=devhub --limit=5
```

### 4.4 흐름 1.5 (source sync) — raw 변경 후 wiki frontmatter 자동

```bash
# SSOT 직접 호출:
python3 ~/wiki/skills/wiki-source-sync/scripts/wiki-source-sync.py \
  --project=devhub [--dry-run] [--auto-fix]
```

### 4.5 D-72 ~ D-80 잔여 4종 (그대로, my_harness SSOT dispatch)

```bash
bash ai-workflow/skills/wiki-ingest-from-raw/scripts/wiki-ingest-from-raw --source <rel> --apply
bash ai-workflow/skills/wiki-lint/scripts/wiki-lint [--fix]
bash ai-workflow/skills/wiki-pr-update/scripts/wiki-pr-update --pr=<num> [--apply]
bash ai-workflow/skills/wiki-query/scripts/wiki-query --query "<text>" [--file]
```

## 5. Tier 정책 (2-tier)

본 저장소 의 **2-tier 정책** (AGENTS.md §6) 적용:
- `ai-workflow/skills/**` = **공용** (my_harness SSOT 와 본 저장소 wrapper 모두 사내 한정 정보 미포함)
- `docs/llm-wiki/**` = **공용** (D-72 결정 — wiki = Gitea private 만, sa-internal/ 격리 불요)

`bash scripts/check-tier-separation.sh` PASS 예상.

## 6. 다음 행동 (forward path)

### 6.1 my_harness 측 (별도 작업, 본 저장소 scope 외)

- **my_harness 측 D-86 정합** (사용자 결정 시점): my_harness 의 3종 D-86 thin wrapper 도 본 저장소 와 동시에 폐기.
- **my_harness 측 SSOT 이전** (사용자 결정 시점): D-72 ~ D-80 의 my_harness SSOT 를 `~/wiki` 로 이전 시 본 저장소 의 4종 thin wrapper SSOT 참조 경로 갱신.

### 6.2 흐름 1/2/3 forward path (v2.0+)

- v2.0+: Mavis 가 prompt 받을 때 자동 흐름 1/2/3 trigger (현시점 수동 호출만)
- v2.0+: `.git/hooks/post-commit` 으로 흐름 2 자동 dispatch
- v2.0+: BM25 + vector + LLM rerank (qmd) 통합 (흐름 3 v2.0)

### 6.3 흐름 1 ↔ 4.5 (wiki-ingest-from-raw) 연동

본 저장소 측 `docs/llm-wiki/ingest-skill.md` 의 `wiki-ingest-from-raw.sh` 와 `ai-workflow/skills/wiki-ingest-from-raw` 의 관계:
- `ai-workflow/skills/wiki-ingest-from-raw` = 본 저장소 측 thin wrapper (my_harness SSOT dispatch)
- `wiki-ingest-from-raw.sh --source <rel> --apply` = 운영 entry point (thin shell + 환경변수)
- **연동**: 흐름 1 의 `wiki-prompt-log` (raw 작성) → 자동 (또는 수동) 으로 `wiki-ingest-from-raw` dispatch → wiki page 까지 합성. 흐름 1 의 raw 만 작성하고 끝 ❌.
