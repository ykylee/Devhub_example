# wiki-query-helper (본 저장소 측 thin wrapper)

- **문서 목적**: 흐름 3 (WORKFLOW.md §4) 의 **본 저장소 측 thin wrapper**. LLM agent 가 사용자 질문 받으면 `wiki-query-helper` 로 **후보 페이지** 빠르게 식별.
- **SSOT**: `~/wiki/skills/wiki-query-helper/SKILL.md` + `~/wiki/skills/wiki-query-helper/scripts/wiki-query-helper.py` (vault, cross-project)
- **본 wrapper 의 역할**: 본 저장소 (= DevHub) 에서 vault 의 SSOT Python script 호출. my_harness 와 동일 패턴 (D-86, 2026-06-11).
- **대상 독자**: yklee, Mavis / Mavis Code, 본 저장소 작업 agent
- **상태**: active (D-86 정합, 2026-06-11)
- **최종 수정일**: 2026-06-11
- **관련 문서**:
  - `~/wiki/skills/wiki-query-helper/SKILL.md` (SSOT)
  - `~/wiki/AGENTS.md` v1.5
  - `~/wiki/WORKFLOW.md` §4 (흐름 3)
  - `docs/llm-wiki/query-skill.md` (D-79 본 저장소 wrapper, D-79 = devhub 측 SSOT dispatch / 본 skill = cross-project 가벼운 후보 추천)
  - `~/repos/my_harness/ai-workflow/skills/wiki-query-helper/` (밴치마킹 출처)

## 1. 사용법

### 1.1 LLM agent 자동 (Mavis / Mavis Code 룰)

Mavis 가 사용자 prompt 받을 때:
- **vault 조회 필요 = 사용자 prompt 가 DevHub / Keycloak / RBAC / ADR / 이 저장소 관련**
- Mavis 가 1차로 `wiki-query-helper "<추출 검색어>" --project=devhub` 실행
- 후보 page 의 path/title/snippet 을 system context 에 inject
- 답변 끝에 `Filed as [[query/YYYY-MM-DD-<topic>]]` 표기 (선택)

### 1.2 사용자 명시 호출

```bash
bash ai-workflow/skills/wiki-query-helper/scripts/wiki-query-helper \
  "<query>" [--project=devhub] [--limit=10]
```

### 1.3 JSON 출력

```json
{
  "query": "...",
  "project": "...",
  "index_hits_count": ...,
  "candidates": [
    {
      "path": "/Users/yklee/wiki/wiki/projects/devhub/...",
      "title": "...",
      "type": "...",
      "last_touched": "...",
      "snippet": "..."
    }
  ]
}
```

## 2. 옵션 (SSOT §3 정합)

| 옵션 | 필수 | 설명 |
|---|---|---|
| `query` (positional) | yes | 검색어 |
| `--project` | no (default `cross`) | `my-harness` \| `devhub` \| `cross` |
| `--limit` | no (default `10`) | 최대 후보 수 |

## 3. 동작 (SSOT §4 정합)

1. `~/wiki/index.md` 의 section 별 entry 매칭 (대소문자 무시 substring)
2. 매칭 안 되면 `grep -rl "<query>" ~/wiki/wiki/projects/<project>/` — frontmatter title/related + 본문 첫 200자 추출
3. 후보 페이지 list + 절대경로 + title + type + last_touched + snippet (50자)
4. stdout JSON 출력

## 4. 정책 (SSOT §6 정합)

- **fallback**: 후보 0건이면 raw/ 만 보고 즉시 답 ❌. **"raw 만 있고 wiki 미합성 — ingest 먼저 할까?" 권고**
- **LRU/최근 갱신**: `last_touched` desc 정렬
- **타입 가중치**: query 타입 > source 타입 > 그 외
- **cross-project**: `wiki/cross/` 도 후보 포함

## 5. trigger (SSOT §2 정합)

- (a) LLM agent 가 "모르면 wiki 먼저" 룰에 따라 호출
- (b) 사용자 명시 호출: 위 1.2

## 6. 안전 (SSOT §7 정합)

- `raw/` 절대 수정 안 함
- `wiki/` 페이지 read-only
- stdout JSON 만, 부수효과 ❌

## 7. 의존 (SSOT §8 정합)

- `python3` 3.10+ (stdlib only)
- `ripgrep` (옵션, 미설치 시 `grep` fallback)
- `~/wiki/` vault

## 8. 향후 (SSOT §9 정합)

- v2.0: BM25 + 벡터 임베딩 + LLM 리랭킹 (qmd)
- v2.0: query 결과 → `query/YYYY-MM-DD-<topic>.md` 자동 filing
- v2.0: Mavis prompt level "answer 끝에 Filed as [[query/...]]" 자동 trigger

## 9. 다음 행동

- Mavis 가 prompt 받을 때 자동 trigger (현시점 수동 호출만 — 자동 trigger 는 v2.0+ forward path)
- 본 skill 과 `docs/llm-wiki/query-skill.md` 의 D-79 wrapper 의 **역할 분리**:
  - 본 skill (wiki-query-helper) = **cross-project 가벼운 후보 추천** (인덱스 + frontmatter grep), json 형식
  - D-79 (wiki-query) = **devhub/my-harness project 정식 query** (vault 의 `query/<date>-<topic>.md` filing + log/index 갱신), AGENTS.md §2.2 6 step 자동
  - **호출 우선순위**: LLM agent 는 본 skill 로 후보 식별 → 후보 1건 이상이면 read; D-79 는 사용자 명시 filing (--file 모드)
