# wiki-query (본 저장소 측 thin wrapper)

- **문서 목적**: LLM Wiki vault 의 read-only query (full-text / wikilink / frontmatter key 매칭) 의 **본 저장소 측 thin wrapper**. --file 모드 시 AGENTS.md §2.2 6 step 자동.
- **SSOT**: `~/repos/my_harness/ai-workflow/skills/wiki-query/SKILL.md` + `scripts/run_wiki_query.py` (D-79/D-81 T-d-79-2, handoff §2 의 6 file 중 2)
- **본 wrapper 의 역할**: 본 저장소 (= DevHub) 에서 my_harness 의 Python script 호출. PR #562/563 의 3+13 skill 의 thin wrapper 정공법 정합.
- **대상 독자**: yklee, 본 저장소 작업 agent, Mavis / Mavis Code
- **상태**: active (D-79/D-81, 2026-06-11)
- **최종 수정일**: 2026-06-11
- **관련 문서**:
  - `~/repos/my_harness/ai-workflow/skills/wiki-query/SKILL.md` (SSOT, 5838 lines)
  - `docs/llm-wiki/query-skill.md` (본 저장소 측 사용 가이드, PR #552, handoff §2 의 본 저장소 측 thin wrapper)
  - `~/wiki/AGENTS.md` v1.5 §2.2 (Query 6 step)
  - `ai-workflow/skills/wiki-query-helper/SKILL.md` (D-86 흐름 3 — 본 skill 과 trigger 다름: wiki-query-helper = cross-project 가벼운 후보 추천 (인덱스 + frontmatter grep), wiki-query = devhub/my-harness project 정식 query (vault 의 query/<date>-<topic>.md filing + log/index 갱신))

## 1. 사용법

### 1.1 본 저장소 측 wrapper

```bash
bash ai-workflow/skills/wiki-query/scripts/wiki-query \
  --query "<text>" [--project=devhub|my-harness] \
  [--tag=<tag>] [--type=<type>] [--limit=N] \
  [--format=md|json|plain] [--file|--no-file] [--quiet]
```

### 1.2 SSOT 직접 호출

```bash
python3 ~/repos/my_harness/ai-workflow/skills/wiki-query/scripts/run_wiki_query.py \
  --query "Keycloak RBAC" --project=devhub
```

## 2. 옵션 (SSOT 정합)

| 옵션 | 필수 | 설명 |
|---|---|---|
| `--query` | yes | 검색어 (full-text / wikilink / frontmatter key 매칭) |
| `--vault-path` | no (default `~/wiki`) | vault 루트 |
| `--project` | no (default `devhub`) | `devhub` \| `my-harness` |
| `--tag` | no | frontmatter `tags:` 필터 (AND, 단일) |
| `--type` | no | `concept` \| `entity` \| `topic` \| `source` \| `comparison` \| `query` |
| `--limit` | no (default 20) | 최대 결과 수 (0 이하 = 무제한) |
| `--format` | no (default `md`) | `md` \| `json` \| `plain` |
| `--file` | no | `--no-file` 의 반대. query/ 페이지 자동 file + log.md append + index.md 갱신 |
| `--no-file` | no (default on) | read-only |
| `--quiet` | no | stderr 메시지 최소화 |
| `--output` | no | `json` \| `markdown` \| `both` (lint report 용) |

## 3. 출력 (read-only default, --file 모드는 vault side effect)

### 3.1 stdout (default: md)

`Query: <text>` 헤더 + `## Hits` 섹션 + 각 hit 의 `### [[<title>]]` + frontmatter + excerpt.

### 3.2 stdout (json)

`{query, project, mode, hit_count, results: [...]}`.

### 3.3 vault side effects (--file 모드)

- `~/wiki/wiki/projects/<project>/query/<date>-<topic>.md` (4섹션: 질문 / 사용 컨텍스트 / 답변 / 후속 액션, frontmatter 8 key)
- `~/wiki/log.md` (1 line append, `## [YYYY-MM-DD] query | <topic>`, idempotent)
- `~/wiki/index.md` (query 섹션 1줄 추가, idempotent)

## 4. 정책 (SSOT 정합)

- **raw/ 절대 수정 안 함** (AGENTS.md §6)
- **schema/ 수정 ❌**
- **wiki/concepts/entities/topics/sources/comparisons/** read-only 영역 수정 ❌ (Query 책임 외)
- **AGENTS.md 수정 ❌**
- **다른 project 의 wiki/projects/<other>/ 수정 ❌** (cross-project 분리)
- **vault Gitea remote push ❌** (사용자 수동)
- **자동 lint 결과 자동 머지 ❌**

## 5. trigger

- (a) **수동**: 위 1.1 wrapper 또는 직접 Python 호출
- (b) **반자동**: Mavis 가 사용자 prompt 받으면 (1차로) `wiki-query-helper` 호출 → 후보 식별 후 (2차로) `wiki-query --file` 로 정식 query filing
- (c) **자동 (v2.0+)**: Mavis prompt level "answer 끝에 Filed as [[query/...]]" 자동 trigger

## 6. 안전 (SSOT 정합)

- `raw/` 절대 수정 안 함
- `wiki/` 페이지 read-only (--file 모드만 `query/<date>-<topic>.md` 신규)
- `index.md` / `log.md` 갱신 누락 ❌ (전 단계 verify 후 다음 단계)
- 절대경로 stdout 출력

## 7. 의존

- `python3` 3.10+ (stdlib only)
- `~/repos/my_harness/` (skill SSOT)
- `~/wiki/` vault
- (옵션) `ripgrep` — 미설치 시 `grep` fallback

## 8. 기존 `scripts/wiki-query.sh` 와의 정합

본 저장소 에는 **2 layer 의 wiki-query wrapper** 가 공존:

| Layer | 위치 | Backend |
|---|---|---|
| **기존 thin wrapper** (D-79) | `scripts/wiki-query.sh` (L66) | 9 option dispatch → `run_wiki_query.py` |
| **신규 SKILL wrapper** (PR #562/563 정합) | `ai-workflow/skills/wiki-query/scripts/wiki-query` (이 PR) | `run_wiki_query.py` 직접 호출 |

**역할 분리**: 기존 wrapper = thin shell (옵션 파싱 + 환경변수). 신규 wrapper = 단순 dispatch (skill 만 호출). 둘 다 같은 SSOT 호출.

## 9. 다음 행동

- 본 저장소 측 vault query 실측 (예: `--query "Keycloak RBAC" --file`)
- 또는 Mavis session hook 으로 자동 dispatch
- 또는 `docs/llm-wiki/Mavis-workflow.md` 9번째 문서 작성
