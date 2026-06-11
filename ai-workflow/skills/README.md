# ai-workflow/skills/ — 본 저장소 (= DevHub) 측 LLM Wiki skill 색인

- **문서 목적**: 본 저장소 의 `ai-workflow/skills/` 디렉터리에 들어가는 skill 들의 색인. 모든 skill 은 **본 저장소 측 thin wrapper** (vault 의 cross-project SSOT 를 호출).
- **밴치마킹 출처**: `~/repos/my_harness/ai-workflow/skills/` (D-86, 2026-06-11, 동일 패턴)
- **SSOT 위치**: `~/wiki/skills/` (vault 의 cross-project SSOT, vault Gitea private)
- **대상 독자**: yklee, Mavis / Mavis Code, 본 저장소 작업 agent
- **상태**: active (D-86 정합, 2026-06-11)
- **최종 수정일**: 2026-06-11
- **관련 문서**:
  - `~/wiki/skills/README.md` (vault 측 SSOT 색인, my_harness 와 동일 형식)
  - `~/wiki/AGENTS.md` v1.5
  - `~/wiki/WORKFLOW.md` (3개 흐름의 통합 정의)
  - `docs/llm-wiki/README.md` (D-72 ~ D-80 의 본 저장소 측 SSOT — `docs/llm-wiki/` 와 역할 분리)

## 1. 본 저장소 측 skill 목록

본 저장소 의 `ai-workflow/skills/` 는 **흐름 1/2/3 (D-86)** 의 본 저장소 측 thin wrapper 만 둠. **flow 정공법 (cross-project)**:

| 흐름 | 본 저장소 skill | SSOT (vault) | trigger |
|---|---|---|---|
| **흐름 1** — 사용자 prompt → raw/ 기록 | `wiki-prompt-log` | `~/wiki/skills/wiki-prompt-log/` | Mavis 자동 (v2.0+) / 수동 |
| **흐름 2** — git 이벤트 → wiki/ 갱신 | `wiki-event-sync` | `~/wiki/skills/wiki-event-sync/` | post-commit hook (v2.0+) / Mavis 반자동 / 수동 |
| **흐름 3** — knowledge 조회 → wiki/ 후보 식별 | `wiki-query-helper` | `~/wiki/skills/wiki-query-helper/` | Mavis 자동 (v2.0+) / 수동 |

각 skill 의 상세:
- `wiki-prompt-log/SKILL.md` + `scripts/wiki-prompt-log`
- `wiki-event-sync/SKILL.md` + `scripts/wiki-event-sync`
- `wiki-query-helper/SKILL.md` + `scripts/wiki-query-helper`

## 2. 본 저장소 측 thin wrapper 정공법

각 skill 의 디렉터리는:
- `SKILL.md` — 본 저장소 측 사용 가이드 (vault SSOT 참조 + 본 저장소 고유 노트)
- `scripts/<skill-name>` — bash thin wrapper (vault 의 `.py` 직접 호출)
- Python 구현은 vault 에만 둠 (cross-project SSOT, my_harness 와 동일)

이유:
- 단일 진실: vault 가 cross-project SSOT (my_harness, devhub, cross 동거)
- 본 저장소 의 wrapper 는 단순 dispatch — 변경 빈도 낮음
- yklee 가 한쪽에서 Python 수정 → 양쪽 repo 의 wrapper 가 자동 효력 발생

## 3. 본 저장소 측 `docs/llm-wiki/` 와의 역할 분리

본 저장소 에는 **두 가지 SSOT 디렉터리**가 의도적으로 공존 (D-72 결정):

| 디렉터리 | 영역 | skill 본문 |
|---|---|---|
| **`docs/llm-wiki/`** (8 file) | D-72 ~ D-80 — vault raw mirror / ingest / D-79 query / D-80 pr-update 의 **본 저장소 측 정공법/가이드** | `scripts/wiki-{sync,ingest-from-raw,query,pr-update}.sh` (D-72 ~ D-80 thin wrapper) |
| **`ai-workflow/skills/`** (이 directory) | D-86 — 흐름 1/2/3 (prompt log / event sync / query helper) 의 **본 저장소 측 thin wrapper** | `scripts/wiki-{prompt-log,event-sync,query-helper}` (D-86 thin wrapper) |

**겹치지 않음**:
- D-72 ~ D-80 = vault 의 raw mirror (7 패턴 docs) + ingest + D-79 query + D-80 pr-update
- D-86 = 흐름 1 (prompt log) + 흐름 2 (git event sync) + 흐름 3 (candidate query)

## 4. 사용법 예시

### 4.1 흐름 1 (prompt log) — Mavis 가 사용자 prompt 기록

```bash
bash ai-workflow/skills/wiki-prompt-log/scripts/wiki-prompt-log \
  --project=devhub --slug=<kebab> --intent="<1~3문장>"
```

### 4.2 흐름 2 (event sync) — commit 직후 자동 또는 수동

```bash
bash ai-workflow/skills/wiki-event-sync/scripts/wiki-event-sync \
  --op=commit --project=devhub --ref=HEAD
```

### 4.3 흐름 3 (query helper) — Mavis 가 자동 또는 사용자 명시

```bash
bash ai-workflow/skills/wiki-query-helper/scripts/wiki-query-helper \
  "Keycloak RBAC" --project=devhub --limit=5
```

## 5. Tier 정책 (2-tier)

본 저장소 의 **3-tier 정책** (AGENTS.md §6) 적용:
- `ai-workflow/skills/**` = **공용** (vault 의 SSOT 와 본 저장소 wrapper 모두 사내 한정 정보 미포함)
- `docs/llm-wiki/**` = **공용** (D-72 결정 — wiki = Gitea private 만, sa-internal/ 격리 불요)

`bash scripts/check-tier-separation.sh` PASS 예상.

## 6. 다음 행동 (forward path)

- v2.0+: Mavis 가 prompt 받을 때 자동 흐름 1/2/3 trigger (현시점 수동 호출만)
- v2.0+: `.git/hooks/post-commit` 으로 흐름 2 자동 dispatch
- v2.0+: BM25 + vector + LLM rerank (qmd) 통합 (흐름 3 v2.0)
- 본 저장소 측 `docs/llm-wiki/ingest-skill.md` 의 `wiki-ingest-from-raw.sh` 와 본 `ai-workflow/skills/wiki-prompt-log` 의 관계:
  - 본 `wiki-prompt-log` = 사용자 prompt → `raw/projects/<project>/prompt/<date>-<slug>.md` 작성
  - `wiki-ingest-from-raw.sh --source <rel> --apply` = 그 raw file 을 `wiki/projects/<project>/sources/<title>.md` 로 ingest
  - **연동**: 본 skill 으로 raw 작성 → 자동 (또는 수동) 으로 ingest dispatch → wiki page 까지 합성. 흐름 1 의 raw 만 작성하고 끝 ❌.
