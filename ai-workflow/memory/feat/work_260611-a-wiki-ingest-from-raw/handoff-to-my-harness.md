# Handoff to my_harness — D-79 (wiki-query) + D-80 (wiki-pr-update) Skill 작성 의뢰

- 문서 목적: 본 저장소 (= DevHub) 측에서 작성한 D-79 + D-80 skill 의 **thin wrapper 4 file** 의 정공법 정합을 위해, my_harness 측 owner/agent 가 작성해야 할 **SSOT spec 2 file + impl 4 file** 의 정밀한 작성 가이드. 본 메시지는 sprint `feat/work_260611-a-wiki-ingest-from-raw` 의 후속 작업으로 my_harness 측에 일임.
- 범위: D-72 §11.1 thin-wrapper 정공법 — 본 저장소 = thin wrapper (DONE), my_harness = SSOT spec + impl (이 메시지의 작성 대상). **본 메시지 작성 = 본 저장소 측 housekeeping**, my_harness 측 작성은 별도 작업.
- 대상 독자: my_harness 측 owner (yklee), my_harness 측 작업 agent (Codex/Claude 등), my_harness 측 skill catalog 운영자.
- 상태: draft (D-79/D-80 handoff 의뢰, 2026-06-11)
- 최종 수정일: 2026-06-11
- 관련 문서: [`./state.json`](./state.json) (sprint state), [`./session_handoff.md`](./session_handoff.md), [`./work_backlog.md`](./work_backlog.md), [`./backlog/2026-06-11.md`](./backlog/2026-06-11.md), [`../../../../docs/llm-wiki/query-skill.md`](../../../../docs/llm-wiki/query-skill.md) (D-79 wrapper guide), [`../../../../docs/llm-wiki/pr-update-skill.md`](../../../../docs/llm-wiki/pr-update-skill.md) (D-80 wrapper guide), [`../../../../scripts/wiki-query.sh`](../../../../scripts/wiki-query.sh) (D-79 wrapper), [`../../../../scripts/wiki-pr-update.sh`](../../../../scripts/wiki-pr-update.sh) (D-80 wrapper), [`~/wiki/AGENTS.md` v1.5 (D-71, 2026-06-10)](file:///home/yklee/wiki/AGENTS.md), [`~/repos/my_harness/ai-workflow/core/wiki_ingest_skill_spec.md` (D-72, 278 lines, §1-§11)](file:///home/yklee/repos/my_harness/ai-workflow/core/wiki_ingest_skill_spec.md) (가장 가까운 precedent, verbatim §1-§11 구조 사용)

## 0. 본 메시지의 의도

본 저장소 (= DevHub) 측은 2026-06-11 에 D-79 (`vault-query`) + D-80 (`pr-vault-update`) 의 **thin wrapper 4 file** 을 작성 완료 (commit `15ca106f`, PR #552 update):

- `scripts/wiki-query.sh` (192 lines, executable, 옵션 9개)
- `scripts/wiki-pr-update.sh` (150 lines, executable, 옵션 5개)
- `docs/llm-wiki/query-skill.md` (230+ lines, §1-§9)
- `docs/llm-wiki/pr-update-skill.md` (230+ lines, §1-§9)

본 wrapper 들은 `MYHARNESS_ROOT=$HOME/repos/my_harness` + `VAULT_ROOT=$HOME/wiki` 환경 변수 가정 하에 my_harness 측 SSOT spec/impl 의 Python script 를 dispatch. wrapper 는 **my_harness 측 skill 부재 시 exit 1 + SSOT 경로 안내** — 따라서 my_harness 측 작성 즉시 wrapper 동작 가능.

D-72 §11.1 정공법 (본 저장소 = thin wrapper, my_harness = SSOT) 의 일관성 유지를 위해, **my_harness 측에 다음 6 file 작성 의뢰**:

| # | artifact | 위치 | 본 저장소 wrapper 와의 관계 |
|---|---|---|---|
| 1 | D-79 spec | `~/repos/my_harness/ai-workflow/core/wiki_query_skill_spec.md` | 본 wrapper 의 SSOT (입출력 계약, 권한, 실패 규칙) |
| 2 | D-79 SKILL.md | `~/repos/my_harness/ai-workflow/skills/wiki-query/SKILL.md` | (선택) wiki-lint 의 SKILL.md 와 동일 패턴 |
| 3 | D-79 impl | `~/repos/my_harness/ai-workflow/skills/wiki-query/scripts/run_wiki_query.py` | 본 wrapper 가 dispatch 하는 두뇌 |
| 4 | D-80 spec | `~/repos/my_harness/ai-workflow/core/wiki_pr_update_skill_spec.md` | 본 wrapper 의 SSOT |
| 5 | D-80 SKILL.md | `~/repos/my_harness/ai-workflow/skills/wiki-pr-update/SKILL.md` | (선택) |
| 6 | D-80 impl | `~/repos/my_harness/ai-workflow/skills/wiki-pr-update/scripts/run_wiki_pr_update.py` | 본 wrapper 가 dispatch 하는 두뇌 |

## 1. 작성 패턴 (verbatim) — D-72 wiki_ingest_skill_spec.md 의 §1-§11 구조 그대로

D-72 의 my_harness 측 spec (`wiki_ingest_skill_spec.md`, 278 lines, 11 sections) 가 가장 가까운 precedent. **각 spec 의 §1-§11 structure 를 verbatim 으로 따를 것**.

### 1.1 Frontmatter (5 key, 9 lines)

```yaml
# <Title> Skill Spec

- 문서 목적: <skill-name> skill 의 입력/출력 계약, 동작 순서, 권한 경계, 실패 규칙을 정의한다.
- 범위: <what the skill does in 1-2 sentences>
- 대상 독자: AI agent 설계자, skill 구현자, vault 운영자, DevHub / my_harness 프로젝트 멤버
- 상태: **draft** (D-XX, YYYY-MM-DD)
- 최종 수정일: YYYY-MM-DD
- 관련 문서: ~/wiki/AGENTS.md, ~/wiki/schema/lint_rules.md, ./workflow_skill_catalog.md, ./session_start_skill_spec.md, ../skills/<skill-name>/SKILL.md
```

### 1.2 Section headers (§1-§11, verbatim)

```
## 1. 목적
## 2. 선행 원칙
## 3. 입력 계약
### 3.1 필수 입력
### 3.2 선택 입력
### 3.3 입력 해석 규칙
## 4. 출력 계약
### 4.1 JSON (stdout)
### 4.2 Markdown (vault _lint/<project>/...md)
## 5. 동작 절차
### 5.1 사전 검증 (validate)
### 5.2 source 식별 (collect)
### 5.3 page 작성 (render)
### 5.4 cross-ref 갱신 (cross-link)
### 5.5 index/log 갱신 (manifest)
### 5.6 lint (post-ingest)
### 5.7 최종 출력
## 6. 권한 경계
## 7. 판단 규칙
## 8. 실패 및 경고 규칙
### 8.1 실패로 처리할 조건
### 8.2 경고로 처리할 조건
### 8.3 실패 시 최소 출력
## 9. 권한과 수정 제한
## 10. 수동 대체 절차
## 11. 구현 체크리스트
## 다음에 읽을 문서
```

## 2. D-79 wiki-query-skill — 입력/출력 계약 (본 turn wrapper 와 정합)

### 2.1 필수 입력 (§3.1)

| 필드 | CLI | 설명 |
|---|---|---|
| `--query` | str | 검색어. full-text / wikilink / frontmatter key 모두 매칭. |
| `--vault-path` | str | vault 루트 경로. 기본 `~/wiki`. |
| `--project` | str | devhub \| my-harness. cross 미지원 (정합성). |

### 2.2 선택 입력 (§3.2)

| 필드 | CLI | 기본값 | 설명 |
|---|---|---|---|
| `--tag` | str | (none) | frontmatter `tags:` 필터 (AND, 단일). |
| `--type` | enum | (none) | `concept` \| `entity` \| `topic` \| `source` \| `comparison` \| `query` \| `meta` (AGENTS.md §3 정합). |
| `--limit` | int | 20 | 최대 결과 수. 0 이하 = 무제한. |
| `--format` | enum | `md` | `md` \| `json` \| `plain`. (D-72 spec 의 `--output` 와 별도 — wrapper 의 `--format` 은 결과 출력 형식, spec 의 `--output` 은 JSON/Markdown report 출력) |
| `--file` | flag | off | query/ 페이지 자동 file + log.md 1 line append (AGENTS.md §2.2 6 step 자동). default = read-only. |
| `--quiet` | flag | off | stderr 메시지 최소화. |
| `--output` | enum | `json` | `json` \| `markdown` \| `both` (D-72 wiki_ingest_skill_spec.md §3.2 의 `--output` 정합, lint report 용). |

### 2.3 출력 JSON (§4.1)

```json
{
  "ok": true,
  "query": "Keycloak RBAC",
  "project": "devhub",
  "mode": "no-file",
  "tool_version": "0.1.0",
  "examined_at": "2026-06-11T...",
  "hit_count": 7,
  "results": [
    {
      "title": "rbac",
      "type": "concept",
      "tags": ["rbac", "auth"],
      "path": "wiki/projects/devhub/concepts/rbac.md",
      "sources": ["raw/projects/devhub/docs/governance/code-taxonomy.md"],
      "last_touched": "2026-06-11",
      "excerpt": "RBAC (Keycloak + cache + row scoping) — ...",
      "links": [],
      "backlinks": ["keycloak", "devhub-auth-session"]
    }
  ],
  "warnings": [],
  "errors": []
}
```

### 2.4 Markdown 출력 (§4.2)

`~/wiki/_lint/<project>/query_<date>.md` 에 Markdown 형식 결과 (D-72 ingest 의 `ingest_<date>.md` 와 동일 위치). mode 가 `--file` 이면 추가 side effect:
- `wiki/projects/<project>/query/<date>-<topic>.md` **신규** (4섹션: 질문 / 사용 컨텍스트 / 답변 / 후속 액션, frontmatter 8 key)
- `log.md` **append** (`## [<date>] query | <topic>` 1 line, idempotent)

## 3. D-79 동작 절차 (§5)

### 5.1 사전 검증

- `--vault-path` 디렉터리 존재 확인
- `~/wiki/AGENTS.md` v1.5 존재 확인 (없으면 exit 1 + `wiki-init` hint)
- `--project` whitelist (devhub | my-harness)
- `--file` 모드 시 `wiki/projects/<project>/query/` 디렉터리 auto-create

### 5.2 source 식별 — 4 query primitive (background 5 의 E 정합)

`~/wiki/AGENTS.md` §2.2 Query step 1 ("index.md 먼저 읽고 후보 페이지 식별") 기반:

1. **Tag list** — `rg '\#[a-zA-Z0-9_-]+' --only-matching` (frontmatter + inline)
2. **Full-text search** — `rg -w '<query>' --line-number --context 1 --json` (content)
3. **Wikilink traversal** — `rg '\[\[([^\]|]+)(?:\|[^\]]+)?\]\]' --only-matching` (graph)
4. **Frontmatter read** — frontmatter 8 key 파싱 (title/type/tags/sources/last_touched/related/status/contradictions)

### 5.3 page 작성 — `--file` 모드일 때만

1. `wiki/projects/<project>/query/<YYYY-MM-DD>-<slug>.md` **신규** (frontmatter 8 key + body 4섹션)
2. body 의 4섹션:
   - `# 질문` — 입력 `--query` 원문
   - `# 사용 컨텍스트` — agent 가 받은 context (전제, 가정)
   - `# 답변` — `index.md` + 후보 wiki pages 종합
   - `# 후속 액션` — 권장 follow-up (관련 source 갱신, 새 topic 추가 등)

### 5.4 cross-ref 갱신 — read-only 모드일 때는 skip

`--file` 모드 시에만 query/ page 의 `related: [[<hit>]]` 자동 populate (idempotent: 이미 있으면 skip).

### 5.5 index/log 갱신 — `--file` 모드일 때만

- `index.md` 의 query 섹션에 신규 page 1줄 추가 (idempotent)
- `log.md` append (`## [<date>] query | <topic>` 1 line, idempotent)

### 5.6 lint — read-only 모드일 때는 skip

`--file` 모드 시 wiki-lint skill 호출 (L01~L10, L07 ADR skip). 통과 시 done, 실패 시 `_lint/<project>/query_<date>.md` 에 report.

### 5.7 최종 출력

JSON to stdout (--output=json|markdown|both). Markdown = `_lint/<project>/query_<date>.md`. Done.

## 4. D-79 권한 경계 (§6)

| 권한 | Read | Write | Forbidden |
|---|---|---|---|
| `raw/**` | ✓ (read-only, AGENTS.md §6) | ❌ (절대 금지) | raw 수정 |
| `schema/**` | ✓ | ❌ | schema 수정 |
| `wiki/**/*` (read) | ✓ | ❌ | read-only 영역 (concepts/entities/topics/sources/comparisons) 수정 |
| `wiki/projects/<project>/query/` (--file) | n/a | ✓ (신규 + 4섹션 작성) | read-only 영역은 write 불요 |
| `log.md` (--file) | n/a | ✓ (append, idempotent) | n/a |
| `index.md` (--file) | n/a | ✓ (query 섹션 갱신) | n/a |
| `AGENTS.md` | n/a | ❌ (AGENTS.md §6) | AGENTS.md 수정 |
| 다른 project 의 `wiki/projects/<other>/` | n/a | ❌ | cross-project 수정 |

## 5. D-79 실패 규칙 (§8)

### 8.1 실패로 처리할 조건

- `--vault-path` 디렉터리 부재 → exit 1 + stderr
- `~/wiki/AGENTS.md` 부재 → exit 1 + `wiki-init` hint
- `--query` 부재 → exit 2 (usage)
- `--project` whitelist 미준수 → exit 2
- `query/` 디렉터리 생성 실패 (--file) → exit 1
- `log.md` append 실패 (--file) → exit 1
- 0 results + `--file` 모드 → **warning** (query/ 페이지는 "no hits" 명시 후 작성)

### 8.2 경고로 처리할 조건

- `--limit` > 기본 권장 (20) → stderr 경고
- `--tag` / `--type` 미매칭 결과 0건 → stderr warning (results=0)
- frontmatter 8 key 일부 누락 (lint L01) → stderr warning

### 8.3 실패 시 최소 출력

stdout: `{ok: false, error: "<error_code>", message: "..."}` (JSON)
stderr: 상세 error log
exit code: 0 (정상) | 1 (실패) | 2 (invalid option)

## 6. D-79 수동 대체 절차 (§10)

`run_wiki_query.py` 가 부재하거나 오류 시:

1. Obsidian GUI 에서 직접 query (Ctrl+Shift+F, `tag:` / `path:` / `line:` prefix)
2. 또는 `rg '<query>' ~/wiki --type md` 수동 검색
3. 결과를 Obsidian 의 `query/` 페이지에 manual 작성

## 7. D-79 구현 체크리스트 (§11)

D-72 wiki_ingest_skill_spec.md 의 §11 + background 5 의 E (Obsidian vault query research) 기반:

- [ ] Python stdlib only (no PyYAML, no requests) — wikilink/frontmatter regex 파싱
- [ ] `subprocess.run(['rg', ...])` for full-text search (zero deps, instant)
- [ ] `re` module for frontmatter parsing (`^---\n(.*?)\n---\n?` multiline)
- [ ] `re` module for wikilink extraction (`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
- [ ] `pathlib` for vault path resolution
- [ ] `argparse` for CLI (D-72 wiki-ingest 의 `argparse.ArgumentParser` 패턴)
- [ ] `dataclasses` for Finding/Result (D-72 의 `@dataclass` 패턴)
- [ ] exit codes 0/1/2 (D-72 정합)
- [ ] JSON output schema: `{ok, query, project, mode, tool_version, examined_at, hit_count, results[], warnings[], errors[]}` (D-72 정합)
- [ ] Markdown report: `~/wiki/_lint/<project>/query_<date>.md` (D-72 위치 정합)
- [ ] `--file` mode: query/ 페이지 + log.md + index.md side effect (AGENTS.md §2.2 6 step 정합)
- [ ] idempotency: log.md 같은 line 있으면 skip, query/ 같은 파일 있으면 skip
- [ ] fail-open: vault absent → warning + exit 0 (read-only), 또는 --file 모드도 warning + exit 0 권장
- [ ] (v2.0) BM25 + vector + RRF (FTS5 + sqlite-vec) — `flowing-abyss/obsidian-hybrid-search` (MIT) 패턴

**권장 prior art 참고**:
- `mikeyobrien/rho` (361⭐, MIT) — FTS5 + ripgrep fallback, content-hash incremental indexing, pure regex frontmatter/wikilink parser
- `flowing-abyss/obsidian-hybrid-search` (66⭐, MIT) — FTS5 + trigram + sqlite-vec + RRF, MCP stdio server
- `Roasbeef/obsidian-claude-code` vault-search (207⭐, MIT) — sqlite-vec + ChromaDB + SQLite notes/vec_chunks schema

## 8. D-80 wiki-pr-update-skill — 입력/출력 계약 (본 turn wrapper 와 정합)

### 8.1 필수 입력

| 필드 | CLI | 설명 |
|---|---|---|
| `--pr` | int | PR number. |
| `--vault-path` | str | `~/wiki`. |
| `--project` | str | devhub \| my-harness. |
| `--pr-metadata` | file | `gh pr view <num> --json ...` 의 JSON file. wrapper 가 전달. |

### 8.2 선택 입력

| 필드 | CLI | 기본값 | 설명 |
|---|---|---|---|
| `--touched-files` | file | (none) | `gh pr diff <num> --name-only` 의 output file. wrapper 가 전달. |
| `--apply` | flag | off | 실제 vault 갱신. default = dry-run. |
| `--quiet` | flag | off | stderr 메시지 최소화. |
| `--output` | enum | `json` | json \| markdown \| both. |
| `--reingest` | flag | off | PR touched file 중 mirror-list source 와 매칭 시 `wiki-ingest-from-raw --source <file> --apply` re-run. (wrapper 의 `--reingest` 와 정합) |

### 8.3 출력 JSON

```json
{
  "ok": true,
  "pr_number": 552,
  "pr_title": "feat(wiki): wiki-ingest-from-raw skill (D-72 Phase 3)",
  "head_sha": "43da841f...",
  "vault_path": "/home/yklee/wiki",
  "project": "devhub",
  "mode": "apply",
  "tool_version": "0.1.0",
  "examined_at": "2026-06-11T...",
  "summary": {
    "touched_files": 6,
    "vault_source_files": 0,
    "pages_created": 1,
    "index_md_updates": 1,
    "log_md_appends": 1,
    "idempotent_skip": false
  },
  "created": [
    "wiki/projects/devhub/prs/552.md"
  ],
  "appended": [
    "log.md (1 line)"
  ],
  "warnings": [],
  "errors": []
}
```

### 8.4 Markdown report

`~/wiki/_lint/<project>/pr_update_<date>.md` 에 side effect 상세.

## 9. D-80 동작 절차 (§5)

### 5.1 사전 검증

- `--vault-path` 디렉터리 + `~/wiki/AGENTS.md` v1.5 존재 확인
- `--pr-metadata` JSON file 파싱 가능 확인
- `--touched-files` 가 있으면 file 존재 확인
- `gh` CLI 부재 시 → exit 1 (wrapper 가 사전 검증, my_harness 측 skill 은 가정 가능)

### 5.2 PR metadata 추출

`--pr-metadata` JSON file 에서 (background 4 의 D 정합):
- `number` (= --pr 값과 일치 확인)
- `title`
- `state` (open | closed)
- `merged` (true | false)
- `mergedAt` (merged 시 timestamp)
- `mergeCommitSha` (merged 시)
- `head.sha` (idempotency key 의 head 부분)
- `head.ref` (source branch)
- `base.ref` (target branch)
- `author.login`
- `url`
- `body`
- `labels[]`
- `files[].filename` (--touched-files 가 없을 때)

### 5.3 vault source 매핑

`--touched-files` 의 각 file 에 대해 D-72 mirror-list.md 의 source 패턴 매칭 (7 patterns):

```python
# 7 패턴 (D-72 mirror-list.md §3)
SOURCE_PATTERNS = [
    r"^docs/adr/0\d{3}-.*\.md$",
    r"^docs/governance/.*\.md$",
    r"^docs/planning/.*\.md$",
    r"^docs/setup/.*\.md$",
    r"^docs/requirements\.md$",
    r"^docs/openapi\.yaml$",
    r"^ai-workflow/memory/(state\.json|session_handoff\.md|work_backlog\.md)$",
]
```

매칭된 file 마다 `wiki-ingest-from-raw --source <file> --apply` dispatch (--reingest 일 때만, wrapper 가 분기).

### 5.4 page 작성 — `prs/<num>.md` 신규

`wiki/projects/<project>/prs/<num>.md` **신규** (frontmatter 8 key + body, AGENTS.md §3 정합):

- frontmatter:
  ```yaml
  ---
  title: "PR #<num>: <title>"
  type: pr
  tags: [pr, project-devhub]
  pr_number: <num>
  author: <author.login>
  state: <state>
  merged_at: <mergedAt or none>
  head_sha: <head.sha>
  sources: [raw/projects/devhub/docs/...touched files...]
  last_touched: <YYYY-MM-DD>
  related: [[[<touched source title>]], ...]
  status: draft
  contradictions: [none]
  ---
  ```
- body:
  - `# PR #<num>: <title>` H1
  - `**Author**: @<login>` + `**State**: <state>` + `**Branches**: <head> → <base>` + `**URL**: [<num>](<url>)`
  - `## PR body` (body 의 처음 50 줄)
  - `## Touched vault sources` (matched file list)
  - `<!-- IdemKey: pr-<num>-<head.sha> -->` HTML 주석 (lint skip marker)

### 5.5 cross-ref 갱신

`touched file` 의 wiki page 가 이미 존재하면 (D-79 query --type source --tag project-devhub), `related: [[pr-<num>]]` 추가 (idempotent).

### 5.6 log.md append

`## [<YYYY-MM-DD>] pr-update | pr-<num>: <title> | <author>` 1 line append (idempotency key `pr-<num>-<head.sha>` 가 이미 있으면 skip). AGENTS.md §2.2 / naming.md §4 의 op type: 본 skill 은 `pr-update` (신규 op) 또는 기존 `edit` 재사용 — **권장 = `pr-update`** (idempotency 명확).

### 5.7 index.md 갱신

`prs` 섹션에 `[[pr-<num>]]` entry 추가 (idempotent).

### 5.8 lint (post-update)

wiki-lint skill 호출 (L01~L10, L07 ADR skip). 통과 시 done, 실패 시 `_lint/<project>/pr_update_<date>.md` 에 report.

## 10. D-80 권한 경계 (§6)

| 권한 | Read | Write | Forbidden |
|---|---|---|---|
| `raw/**` | ✓ (--touched-files verify) | ❌ (절대 금지) | raw 수정 |
| `schema/**` | ✓ | ❌ | schema 수정 |
| `wiki/projects/<project>/prs/` (--apply) | n/a | ✓ (신규 prs/<num>.md) | read-only 영역 (concepts/entities/topics/sources) 수정 |
| `log.md` (--apply) | n/a | ✓ (append) | n/a |
| `index.md` (--apply) | n/a | ✓ (prs 섹션) | n/a |
| `AGENTS.md` | n/a | ❌ | AGENTS.md 수정 |
| 다른 project 의 `wiki/projects/<other>/` | n/a | ❌ | cross-project |
| **vault Gitea remote push** | n/a | ❌ (사용자 수동, AGENTS.md §6.5 정책) | 자동 push 금지 |

## 11. D-80 Idempotency (§7)

**Key**: `pr-<num>-<head.sha>`

- 이미 `prs/<num>.md` 가 존재하고 frontmatter `last_touched >= head.sha` 이면 **skip + "already updated"** (no side effect, exit 0)
- `--touched-files` 가 다른 SHA 와 함께 (force-push / rebase) 이면 **idempotent re-write** (frontmatter `last_touched` + `head.sha` 갱신)
- log.md 같은 line 있으면 skip (idempotent append)
- index.md 같은 `[[pr-<num>]]` 있으면 skip (idempotent update)

## 12. D-80 실패 규칙 (§8)

### 8.1 실패로 처리할 조건

- `--vault-path` 디렉터리 부재 → exit 1
- `~/wiki/AGENTS.md` 부재 → exit 1 + `wiki-init` hint
- `--pr` 부재 또는 invalid → exit 2
- `--pr-metadata` file 부재 또는 invalid JSON → exit 1
- `--touched-files` file 부재 → exit 1
- 0 touched files → **warning** (정상 — PR 이 docs-only 가 아닌 경우)
- vault write 실패 (permission 등) → exit 1

### 8.2 경고로 처리할 조건

- `--reingest` 일 때 mirror-list 매칭 0건 → warning
- wiki-lint L01~L10 위반 → warning (--apply 모드)
- 이미 updated (idempotent skip) → "already updated" log

### 8.3 실패 시 최소 출력

stdout: `{ok: false, error_code: "VAULT_ABSENT" or "INVALID_PR", ...}`
exit code: 0 (정상 or already-updated) | 1 (실패) | 2 (invalid option)

## 13. D-80 수동 대체 절차 (§10)

`run_wiki_pr_update.py` 부재 시:

1. `gh pr view <num> --json ...` 로 metadata 수동 추출
2. Obsidian 에서 `wiki/projects/<project>/prs/<num>.md` 수동 작성 (frontmatter 8 key + body)
3. `log.md` 수동 append
4. `index.md` 수동 갱신

## 14. D-80 구현 체크리스트 (§11)

- [ ] Python stdlib only
- [ ] `argparse` for CLI
- [ ] `json` for --pr-metadata parsing
- [ ] `pathlib` for vault path
- [ ] `subprocess` for `wiki-ingest-from-raw` re-ingest dispatch (--reingest 모드)
- [ ] Datetime/strftime for `log.md` timestamp
- [ ] Idempotency: `pr-<num>-<head.sha>` key + skip-if-exists
- [ ] Fail-open: vault absent → warning + skip (--apply 모드는 warning + exit 0 권장)
- [ ] JSON output schema: `{ok, pr_number, pr_title, head_sha, vault_path, project, mode, tool_version, examined_at, summary, created, appended, warnings, errors}`
- [ ] Markdown report: `~/wiki/_lint/<project>/pr_update_<date>.md`
- [ ] mirror-list 매칭 regex (7 patterns)
- [ ] frontmatter 8 key 정확히 채움 (title, type, tags, pr_number, author, state, merged_at, head_sha, sources, last_touched, related, status, contradictions)

## 15. 검증 절차 (my_harness 측 작성 완료 후)

### 15.1 D-79 검증

1. **dry-run 5 sample**:
   - `--query "Keycloak RBAC" --no-file` (full-text)
   - `--query "ADR-0020" --tag rbac --limit 5` (tag filter)
   - `--query "keycloak" --type concept` (type filter)
   - `--query "keycloak" --format json` (JSON output)
   - `--query "ADR-0020 결정 사항" --file` (--file 모드, 6 step 자동)
2. **verify**:
   - 5 sample 모두 exit 0
   - 0 errors, 0 warnings
   - JSON schema 정합
   - `--file` sample 의 `wiki/projects/<project>/query/<date>-<topic>.md` + `log.md` 1 line + `index.md` 갱신 확인

### 15.2 D-80 검증

1. **dry-run PR #552** (이미 머지된 PR, 6 touched file):
   - `bash scripts/wiki-pr-update.sh --pr 552` (dry-run)
   - 6 touched file 중 mirror-list 매칭 확인 (예상: 0건 — PR #552 의 4 commit 은 모두 docs/* 변경이지만 mirror-list 의 ADR/governance/planning/setup/requirements/openapi/ai-workflow-memory 패턴 중 cross-ref 매칭 1건 가능)
2. **apply PR #552**:
   - `bash scripts/wiki-pr-update.sh --pr 552 --apply`
   - `wiki/projects/<project>/prs/552.md` 신규
   - `log.md` 1 line append
   - `index.md` 의 prs 섹션 갱신
3. **idempotency 검증**:
   - 동일 `--pr 552 --apply` 재실행 → skip + "already updated"
4. **--reingest 검증** (선택):
   - `bash scripts/wiki-pr-update.sh --pr 552 --reingest --apply`
   - touched file 중 mirror-list 매칭 시 re-ingest dispatch

## 16. 의존 + 순서

```
T-d-79-1, T-d-79-2:  my_harness spec 2 file 작성
  ↓
T-d-79-3, T-d-79-4:  my_harness impl 4 file 작성 (run_*.py + SKILL.md)
  ↓
T-d-79-5, T-d-79-6:  dry-run 검증 + --file 검증
  ↓
T-d-80-1, T-d-80-2:  my_harness spec + impl 작성 (D-79 와 동일 패턴)
  ↓
T-d-80-3..6:  dry-run + apply + idempotency + reingest 검증
  ↓
(optional) workflow_skill_catalog.md 갱신 (D-79/D-80 row 추가)
```

## 17. 다음 step + 사용자 confirm 영역

| # | 작업 | 비고 |
|---|---|---|
| 1 | **본 메시지 (handoff-to-my-harness.md) 검토** | 본 저장소 측 housekeeping 의 일환 |
| 2 | **my_harness 측 owner/agent 가 spec 2 file 작성** | `~/repos/my_harness/ai-workflow/core/wiki_query_skill_spec.md` + `wiki_pr_update_skill_spec.md` |
| 3 | **my_harness 측 owner/agent 가 impl 4 file 작성** | `~/repos/my_harness/ai-workflow/skills/wiki-query/SKILL.md + scripts/run_wiki_query.py` + `wiki-pr-update/SKILL.md + scripts/run_wiki_pr_update.py` |
| 4 | **본 저장소 wrapper 4 file 의 dry-run 검증** (T-d-79-3, T-d-79-4) | my_harness 측 작성 완료 후 |
| 5 | **workflow_skill_catalog.md 갱신** (my_harness 측) | D-79/D-80 row 추가 |
| 6 | **PR #552 머지** (본 저장소) | 사용자 confirm 별도 |

본 메시지 끝. 작성 시 의문점은 본 저장소 PR #552 코멘트 또는 `ai-workflow/memory/feat/work_260611-a-wiki-ingest-from-raw/session_handoff.md` 에 follow-up.

## 18. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-11 | 본 sprint — D-79 + D-80 thin wrapper 4 file (commit `15ca106f`, PR #552) 완료 후 my_harness 측 SSOT 작성 의뢰 handoff message 작성. D-72 §11.1 thin-wrapper 정공법 + D-72 wiki_ingest_skill_spec.md 의 §1-§11 verbatim 구조 + background 5 (Obsidian query + PR hook research) 정합. |
