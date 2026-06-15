# Wiki-Query Skill (본 저장소 측 사용법 가이드)

- 문서 목적: LLM agent (또는 사용자) 가 `ai-workflow/wiki/` Obsidian vault 를 query 할 때 사용하는 본 저장소 측 wrapper + guide. vault 의 LLM wiki 패턴 (D-71) 의 §2.2 Query 오퍼레이션 6 step 의 자동화 entry point.
- 범위: `scripts/wiki-query.sh` (thin wrapper) + 본 가이드. **my_harness 측 SSOT**: `~/repos/my_harness/ai-workflow/core/wiki_query_skill_spec.md` (D-79) + `~/repos/my_harness/ai-workflow/skills/wiki-query/` (impl).
- 대상 독자: DevHub/my_harness owner (yklee), LLM agent (wiki 의 RAG source 활용자), my_harness 작업 에이전트.
- 상태: draft (D-79 Phase 1, 2026-06-11)
- 최종 수정일: 2026-06-11
- 관련 문서: [`docs/llm-wiki/ingest-skill.md`](./ingest-skill.md) (D-72 Phase 3, 가장 가까운 precedent), [`docs/llm-wiki/README.md`](../llm-wiki/README.md) (D-72 SSOT), `ai-workflow/wiki/AGENTS.md` v1.5 §2.2 Query (vault 의 LLM 운영 규약), `ai-workflow/wiki/schema/page_template.md` (frontmatter 형식), `ai-workflow/wiki/schema/lint_rules.md` (L01~L10).

## 1. 사용법

### 1.1 Read-only (default)

```bash
# 단일 검색어 (full-text + wikilink + frontmatter key 모두 매칭)
bash scripts/wiki-query.sh --query "Keycloak RBAC"

# tag / type filter
bash scripts/wiki-query.sh --query "rbac" --tag rbac --limit 5
bash scripts/wiki-query.sh --query "Keycloak" --type concept

# 출력 형식
bash scripts/wiki-query.sh --query "ADR-0020" --format json --no-file
```

### 1.2 Read + Write (AGENTS.md §2.2 6 step 자동)

```bash
# read 결과로 `ai-workflow/wiki/wiki/projects/<project>/query/<date>-<topic>.md` 신규 + log.md 1 line append
bash scripts/wiki-query.sh --query "ADR-0020 결정 사항" --file
```

### 1.3 출력 예시 (md format, default)

```
# Query: "Keycloak RBAC"
- vault: /home/yklee/wiki
- project: devhub
- filters: tag=N/A, type=N/A, limit=20
- mode: no-file
- results: 7

## Hits

### [[rbac]] (type: concept, tags: [rbac, auth])
- path: wiki/projects/devhub/concepts/rbac.md
- sources: [raw/projects/devhub/docs/governance/code-taxonomy.md, raw/projects/devhub/docs/adr/0002-rbac-policy-edit-api.md]
- last_touched: 2026-06-11
- excerpt: RBAC (Keycloak + cache + row scoping) — DevHub 의 권한 모델은 Keycloak OIDC 의 resource_access claim + backend 의 in-memory cache + SQL row filter 의 3 계층. ADR-0002, ADR-0020 참고.

### [[keycloak]] (type: entity, tags: [keycloak, sso])
- path: wiki/projects/devhub/entities/keycloak.md
- sources: [raw/projects/devhub/docs/setup/keycloak_operations.md, ...]
- last_touched: 2026-06-11
- excerpt: DevHub SSO/IdP (25.0 → 26.0) — single source of truth (ADR-0019).
...
```

### 1.4 출력 예시 (json format)

```json
{
  "query": "Keycloak RBAC",
  "project": "devhub",
  "mode": "no-file",
  "hit_count": 7,
  "results": [
    {
      "title": "rbac",
      "type": "concept",
      "tags": ["rbac", "auth"],
      "path": "wiki/projects/devhub/concepts/rbac.md",
      "sources": ["raw/projects/devhub/docs/governance/code-taxonomy.md", "..."],
      "last_touched": "2026-06-11",
      "excerpt": "RBAC (Keycloak + cache + row scoping) — ..."
    },
    ...
  ]
}
```

## 2. 입력 계약 (input contract)

### 2.1 옵션 (wrapper = `scripts/wiki-query.sh`)

| 옵션 | 필수 | 기본값 | 설명 |
|---|---|---|---|
| `--query <text>` | yes | — | 검색어. full-text / wikilink (`[[<text>]]`) / frontmatter key 모두 매칭. |
| `--project <name>` | no | `devhub` | `devhub` \| `my-harness`. `cross` 미지원. |
| `--tag <tag>` | no | (none) | frontmatter `tags:` 필터 (AND, 단일). |
| `--type <type>` | no | (none) | `concept` \| `entity` \| `topic` \| `source` \| `comparison` \| `query`. |
| `--limit N` | no | `20` | 최대 결과 수. 0 이하 = 무제한. |
| `--format <fmt>` | no | `md` | `md` \| `json` \| `plain`. `json` 은 다른 tool 입력용. |
| `--file` | no | off | `--no-file` 의 반대. query/ 페이지 자동 file + log.md append. |
| `--no-file` | no | on (default) | read-only. agent 가 stdout 결과만 활용. |
| `--quiet` | no | off | stderr 메시지 최소화. |
| `-h`, `--help` | no | — | usage 출력 후 exit 0. |

### 2.2 Exit code

| code | 의미 |
|---|---|
| 0 | success (read 또는 read+write 모두 성공, 0 results 도 success) |
| 1 | read 실패, 또는 write 실패, 또는 my_harness skill 부재, 또는 gh CLI 부재 |
| 2 | invalid option 또는 required option (--query) 부재 |

## 3. 출력 계약 (output contract)

### 3.1 stdout (default: md)

`Query: <text>` 헤더 + `## Hits` 섹션 + 각 hit 의 `### [[<title>]]` + frontmatter + excerpt.

### 3.2 stdout (json)

`{query, project, mode, hit_count, results: [{title, type, tags, path, sources, last_touched, excerpt}]}`.

### 3.3 stdout (plain)

`[<type>] <path> — <excerpt>` 한 줄 per hit.

### 3.4 stderr

- 진행 메시지 (`[wiki-query] step: ...`, `[wiki-query]   vault: ...`, ...) — `--quiet` 시 silent
- error 메시지 (`[wiki-query] error: ...`)

## 4. 권한 (permissions)

본 skill 의 권한 경계는 **`ai-workflow/wiki/AGENTS.md` v1.5 (D-71) 의 §2.2 Query 6 step + §6 금지**:

### 4.1 허용 (read-only default)

- `index.md` 읽기 (후보 페이지 식별, step 1)
- `wiki/**/*` 모든 페이지 읽기 (frontmatter + body)
- `schema/*` 읽기 (frontmatter 형식 참조)
- `raw/**/*` 읽기 (mirror source 검증)

### 4.2 허용 (--file 옵션)

위 4.1 +

- `wiki/projects/<project>/query/<date>-<topic>.md` **신규** 파일 작성 (4섹션: 질문 / 사용 컨텍스트 / 답변 / 후속 액션, frontmatter 8 key)
- `log.md` **append** (`## [<date>] query | <topic>` 1 line, idempotent: 같은 date+topic 가 이미 있으면 skip)
- `index.md` 의 query 섹션에 신규 page 1줄 추가 (있으면 skip)

### 4.3 금지

- `raw/` 수정 (AGENTS.md §6, 절대 금지)
- `wiki/concepts/`, `wiki/entities/`, `wiki/topics/`, `wiki/sources/`, `wiki/comparisons/` 등 read-only 영역 수정 (Query 오퍼레이션의 책임 외)
- `schema/` 수정
- 자동 lint 결과 자동 머지 (AGENTS.md §6)
- `wiki/AGENTS.md` 수정
- 다른 project 의 `wiki/projects/<other>/` 수정 (cross-project는 별도 lint/operation SOP)
- vault Gitea remote push (사용자 수동, AGENTS.md §6.5 정책)

## 5. 6 step 자동화 (--file 모드)

`AGENTS.md` v1.5 §2.2 Query 의 6 step 자동 수행:

| step | action | file |
|---|---|---|
| 1 | `index.md` 읽고 후보 페이지 식별 | `ai-workflow/wiki/index.md` (read) |
| 2 | 관련 `wiki/` 페이지 read + 종합 | `wiki/projects/<project>/{concepts,entities,topics,sources,...}` (read) |
| 3 | 답변 끝에 `Filed as [[query/<date>-<topic>]]` 한 줄 | answer body |
| 4 | `query/` 페이지 본문 4섹션 (질문 / 사용 컨텍스트 / 답변 / 후속 액션) | `wiki/projects/<project>/query/<date>-<topic>.md` (create) |
| 5 | `log.md` `## [<date>] query | <topic>` 1 line | `ai-workflow/wiki/log.md` (append) |
| 6 | 답변은 Obsidian 내부 file → 향후 ingest source 가산 | (`step 4` 의 page 가 source 후보) |

## 6. 실패 규칙 (failure rules)

| 실패 | 처리 |
|---|---|
| `ai-workflow/wiki/AGENTS.md` 부재 | exit 1 + `[wiki-query] error: vault AGENTS.md not found` + hint `wiki-init` (my_harness D-71 §2.2) |
| `ai-workflow/wiki/index.md` 부재 | exit 1 + hint `index.md 필요 (LLM query 의 첫 reading)` |
| `wiki/projects/<project>/query/` 부재 | 자동 생성 (mkdir -p) 후 진행 |
| 0 results | exit 0 + stdout `## Hits` 빈 섹션. `--file` 모드라도 query/ 페이지는 작성 (후속 액션 항목 = "no hits" 명시) |
| `--query` 부재 | exit 2 + usage |
| invalid `--project` / `--format` | exit 2 + usage |
| my_harness skill 부재 | exit 1 + SSOT 경로 안내 |
| `gh` CLI 부재 (wiki-pr-update 만) | exit 1 + `https://cli.github.com` 안내 |
| vault Gitea push 실패 | out of scope (이 skill 은 local vault 만, push 는 사용자 수동) |

## 7. 다음 행동 (Phase 1 → Phase 3)

| ID | 우선순위 | 작업 | 의존 |
| --- | --- | --- | --- |
| T-d-79-1 | P3 | 본 skill 의 본 저장소 측 wrapper 작성 (`scripts/wiki-query.sh` + 본 가이드) | — |
| T-d-79-2 | P3 | my_harness 측 `wiki_query_skill_spec.md` (§1~§11) + `skills/wiki-query/SKILL.md` + `scripts/run_wiki_query.py` 작성 | — |
| T-d-79-3 | P3 | dry-run 검증 (5 query sample: full-text / wikilink / tag / type / json format) | T-d-79-2 |
| T-d-79-4 | P3 | `--file` 옵션 1회 실행 검증 (query/ 페이지 + log.md + index.md side effect) | T-d-79-3 |
| T-d-79-5 | P3 | wiki-lint 통합 (`wiki-lint` skill 의 L01~L10 가 query/ 페이지 검증) | D-74 |
| T-d-79-6 | P3 | v2.0 (BM25 + vector + MCP) — query 결과의 RAG rerank (my_harness 측, 본 skill 의 read-only 가 source) | my_harness v2.0 |

## 8. 다음 sprint 진입

본 skill 의 본 저장소 측 = thin wrapper (D-72 §11.1 정공법). 두뇌는 my_harness 측 SSOT.

본 PR (또는 다음 sprint) 의 작업:

1. **my_harness 측 spec 작성** (`wiki_query_skill_spec.md` §1~§11, ingest-skill_spec.md 패턴) — 본 wrapper 가 dispatch 할 SSOT.
2. **my_harness 측 impl 작성** (`SKILL.md` + `scripts/run_wiki_query.py`, stdlib only, --query/--tag/--type/--limit/--format/--file 옵션).
3. **dry-run 검증** (5 query sample).
4. **`--file` 1회 실행 검증** (vault side effect 3종: query/ 페이지 + log.md + index.md).
5. **PR 발행** (D-72 PR #544 와 동일: `feat(llm-wiki,scripts): wiki-query skill (D-79 Phase 1) — 본 저장소 wrapper + my_harness SSOT`).

## 9. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-11 | 본 sprint — wiki-query skill (D-79) 의 본 저장소 측 wrapper + 본 가이드 작성 (D-72 §11.1 thin wrapper 정공법) |
