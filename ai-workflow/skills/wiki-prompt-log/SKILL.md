# wiki-prompt-log (본 저장소 측 thin wrapper)

- **문서 목적**: 흐름 1 (WORKFLOW.md §2) 의 **본 저장소 측 thin wrapper**. 사용자 프롬프트 + LLM 결정 + 산출물을 `~/wiki/raw/projects/devhub/prompt/` (또는 `my-harness`, `cross`) 에 immutable 기록.
- **SSOT**: `~/wiki/skills/wiki-prompt-log/SKILL.md` + `~/wiki/skills/wiki-prompt-log/scripts/wiki-prompt-log.py` (vault 의 cross-project SSOT)
- **본 wrapper 의 역할**: 본 저장소 (= DevHub) 에서 vault 의 SSOT Python script 호출. my_harness 와 동일 패턴 (D-86, 2026-06-11).
- **대상 독자**: yklee, Mavis / Mavis Code, 본 저장소 작업 agent
- **상태**: active (D-86 정합, 2026-06-11)
- **최종 수정일**: 2026-06-11
- **관련 문서**:
  - `~/wiki/skills/wiki-prompt-log/SKILL.md` (SSOT)
  - `~/wiki/AGENTS.md` v1.5 (vault 운영 규약)
  - `~/wiki/WORKFLOW.md` §2 (흐름 1)
  - `docs/llm-wiki/ingest-skill.md` (raw → wiki page ingest, 본 저장소 측 가이드)
  - `~/repos/my_harness/ai-workflow/skills/wiki-prompt-log/` (밴치마킹 출처)

## 1. 사용법

### 1.1 본 저장소 측 wrapper (= `scripts/wiki-prompt-log`)

```bash
# 본 저장소 repo root 에서:
bash ai-workflow/skills/wiki-prompt-log/scripts/wiki-prompt-log \
  --project=devhub \
  --slug=<kebab-case> \
  --intent="<1~3 문장 의도>" \
  [--context="<item>"] [--decision="<1줄>"] \
  [--artifacts="<item>"] [--followup="<item>"] \
  [--dry-run]
```

본 wrapper 는 단순히 vault 의 SSOT Python 을 호출. 모든 옵션/동작/출력은 SSOT 가 정의.

### 1.2 vault SSOT 직접 호출 (= 본 wrapper 가 내부적으로 실행하는 명령)

```bash
python3 ~/wiki/skills/wiki-prompt-log/scripts/wiki-prompt-log.py \
  --project=devhub --slug=<slug> --intent="..."
```

## 2. 옵션 (SSOT 정합)

| 옵션 | 필수 | 설명 |
|---|---|---|
| `--project` | yes | `my-harness` \| `devhub` \| `cross` |
| `--slug` | yes | kebab-case, 4~80자 |
| `--intent` | yes | 1~3문장 의도 |
| `--context` | no | 컨텍스트 항목 (반복 가능) |
| `--decision` | no | 결정 1줄 |
| `--artifacts` | no | 산출물 (반복 가능) — commit hash, file list, PR link 등 |
| `--followup` | no | 후속 작업 후보 (반복 가능) |
| `--dry-run` | no | 실제 작성 없이 preview 만 |

## 3. 출력

1. `~/wiki/raw/projects/<project>/prompt/YYYY-MM-DD-<slug>.md` (frontmatter + 본문)
2. `~/wiki/raw/projects/<project>/prompt/_manifest.md` (append)
3. `~/wiki/log.md` (append, `## [YYYY-MM-DD] prompt | <project>/<slug>`)

## 4. 정책 (SSOT §5 정합)

- **idempotent**: 동일 slug 이미 존재 시 abort (덮어쓰기 ❌)
- **slug 검증**: kebab-case, 4~80자
- **날짜 검증**: KST 기준 오늘, 미래 날짜 ❌
- **raw 만 기록**: wiki 합성은 별도 ingest (흐름 2 의 일부, `wiki-ingest-from-raw.sh --source <rel> --apply`)

## 5. trigger

- **Mavis 자동**: LLM agent 가 작업 종료/세션 종료 시 hook 으로 호출 (v2.0+)
- **수동**: 위 1.1 wrapper 또는 직접 Python 호출
- **반자동**: Mavis session 내 "이제 정리하자" 시 Mavis 가 호출

## 6. 안전 (SSOT §6 정합)

- `raw/` 절대 수정 안 함 (기존 파일 보존)
- `_manifest.md` 는 `## [...]` 한 줄 append 만
- `log.md` 는 `## [YYYY-MM-DD] <op> | <title>` 형식 강제
- 실행 후 절대경로 3개 stdout 출력

## 7. 의존

- `python3` 3.10+ (stdlib only)
- `~/wiki/` vault (`raw/projects/<project>/prompt/` 디렉터리 사전 존재 필요)
- (vault Gitea remote push 는 사용자 수동, 본 skill scope 외)

## 8. 다음 행동

- 본 저장소 측 session hook (`.git/hooks/post-commit` 또는 Mavis session hook) 으로 자동 dispatch — **v2.0+ forward path**
- 또는 사용자가 명시 호출 — **현시점 권장**
- 본 저장소 측 `docs/llm-wiki/` 와의 역할 분리: 본 skill = prompt log (raw/ 작성), `wiki-ingest-from-raw.sh` = raw → wiki page 자동 ingest (병행)
