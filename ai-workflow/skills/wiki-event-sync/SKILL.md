# wiki-event-sync (본 저장소 측 thin wrapper)

- **문서 목적**: 흐름 2 (WORKFLOW.md §3) 의 **본 저장소 측 thin wrapper**. git 이벤트 (commit / push / PR / merge / release) 가 발생하면 영향 받은 wiki 페이지의 `last_touched` + `related` + Observations 갱신.
- **SSOT**: `~/wiki/skills/wiki-event-sync/SKILL.md` + `~/wiki/skills/wiki-event-sync/scripts/wiki-event-sync.py` (vault, cross-project)
- **본 wrapper 의 역할**: 본 저장소 (= DevHub) 에서 vault 의 SSOT Python script 호출. my_harness 와 동일 패턴 (D-86, 2026-06-11).
- **대상 독자**: yklee, Mavis / Mavis Code, 본 저장소 작업 agent
- **상태**: active (D-86 정합, 2026-06-11)
- **최종 수정일**: 2026-06-11
- **관련 문서**:
  - `~/wiki/skills/wiki-event-sync/SKILL.md` (SSOT)
  - `~/wiki/AGENTS.md` v1.5
  - `~/wiki/WORKFLOW.md` §3 (흐름 2)
  - `docs/llm-wiki/pr-update-skill.md` (D-80 본 저장소 wrapper, 본 skill 과 trigger 다름 — D-80 = PR-merge 후 vault 갱신, 본 skill = commit/push/PR-merge 통합)
  - `~/repos/my_harness/ai-workflow/skills/wiki-event-sync/` (밴치마킹 출처)

## 1. 사용법

### 1.1 commit op

```bash
bash ai-workflow/skills/wiki-event-sync/scripts/wiki-event-sync \
  --op=commit --project=devhub --ref=<sha> [--dry-run]
```

git 저장소 안에서 실행 (caller cwd 가 본 저장소 또는 sub-directory).

### 1.2 PR op

```bash
bash ai-workflow/skills/wiki-event-sync/scripts/wiki-event-sync \
  --op=pr --project=devhub --ref=<PR#> --repo=ykylee/Devhub_example [--dry-run]
```

`gh` CLI 인증 필요.

### 1.3 release / push / merge op

현시점 SSOT 는 commit + pr 만 구현 (v0.1.0). 나머지는 v0.2.0+ forward.

## 2. 옵션 (SSOT §3 정합)

| 옵션 | 필수 | 설명 |
|---|---|---|
| `--op` | yes | `commit` \| `push` \| `pr` \| `merge` \| `release` |
| `--project` | yes | `my-harness` \| `devhub` \| `cross` |
| `--ref` | yes | sha (short or full) \| PR# \| tag name |
| `--repo` | PR op 필수 | `owner/name` (e.g. `ykylee/Devhub_example`) |
| `--intent` | no | 사용자 의도 (옵션, 기본 = commit message 1줄) |
| `--dry-run` | no | 실제 작성 없이 preview 만 |

## 3. 출력 (commit 기준, SSOT §4 정합)

1. 영향 받은 wiki 페이지 절대경로 list
2. `~/wiki/log.md` 에 `## [YYYY-MM-DD] edit | <sha-prefix> <description>` append
3. 영향 받은 페이지의 frontmatter `last_touched: YYYY-MM-DD` 갱신

## 4. 정책 (SSOT §5 정합)

- **idempotent**: 동일 sha 가 이미 wiki 에 반영되어 있으면 skip (re-run OK)
- **push 무결성**: 로컬 git ref 가 remote 와 다르면 abort (`git fetch` 권고)
- **skip 규칙**: `[skip-wiki]` / `[no-wiki]` commit message → log 한 줄만, page 갱신 ❌
- **외부 영향**: DevHub PR → cross-project wiki 갱신 가능 (사용자 confirm 필요)

## 5. trigger (SSOT §2 정합)

- (a) **수동**: 위 1.1 wrapper 또는 직접 Python 호출
- (b) **반자동**: Mavis session 내 commit 직후 Mavis 가 호출
- (c) **자동 (v2.0+)**: `.git/hooks/post-commit` + GitHub webhook (PR/release) — forward path

## 6. 안전 (SSOT §6 정합)

- `raw/` 절대 수정 안 함
- `wiki/` 페이지 삭제 ❌
- `index.md` / `log.md` 갱신 누락 ❌
- 절대경로 stdout 출력

## 7. 본 저장소 측 trigger 통합 (v2.0+)

`.git/hooks/post-commit` 또는 `post-merge` 에 본 wrapper 자동 등록:

```bash
# post-commit
#!/usr/bin/env bash
cd ~/repos/Devhub_example_minimax
bash ai-workflow/skills/wiki-event-sync/scripts/wiki-event-sync \
  --op=commit --project=devhub --ref=HEAD || true
```

GitHub Actions webhook 으로 PR event 수신 시 dispatch (사내 self-hosted runner 도입 시 forward).

## 8. 다음 행동

- 본 저장소 측 post-commit hook 등록 (선택, 사용자 confirm 후) — v2.0 forward path
- 또는 사용자 명시 호출 — 현시점 권장
- 본 skill 과 `docs/llm-wiki/pr-update-skill.md` 의 D-80 wrapper 는 **trigger 차이**: 본 skill = commit/push/PR 통합 (vault 의 영향 받은 page 식별), D-80 = PR-merge 후 vault 의 `prs/<num>.md` 신규 + log/index 갱신. 보완 관계.
