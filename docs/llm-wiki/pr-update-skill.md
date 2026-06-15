# Wiki-PR-Update Skill (본 저장소 측 사용법 가이드)

- 문서 목적: PR (`gh pr view <num>`) 의 metadata + touched files → `ai-workflow/wiki/` 의 prs/<num>.md 신규 + log.md append + (optional) touched file re-ingest 의 본 저장소 측 wrapper + guide. PR-auto-update skill 의 thin wrapper.
- 범위: `scripts/wiki-pr-update.sh` (thin wrapper) + 본 가이드. **my_harness 측 SSOT**: `~/repos/my_harness/ai-workflow/core/wiki_pr_update_skill_spec.md` (D-80) + `~/repos/my_harness/ai-workflow/skills/wiki-pr-update/` (impl).
- 대상 독자: DevHub owner (yklee), my_harness 작업 에이전트, `gh` CLI 사용자.
- 상태: draft (D-80 Phase 1, 2026-06-11)
- 최종 수정일: 2026-06-11
- 관련 문서: [`docs/llm-wiki/ingest-skill.md`](./ingest-skill.md) (D-72 Phase 3, 가장 가까운 precedent), [`docs/llm-wiki/query-skill.md`](./query-skill.md) (D-79, 동시 작성), [`docs/llm-wiki/README.md`](../llm-wiki/README.md), `ai-workflow/wiki/AGENTS.md` v1.5 §11.1 D-72 cross-project, [`scripts/wiki-query.sh`](../../scripts/wiki-query.sh) (동시 작성, query page 작성 패턴), [`scripts/wiki-ingest-from-raw.sh`](../../scripts/wiki-ingest-from-raw.sh) (D-72, re-ingest dispatch).

## 1. 사용법

### 1.1 Trigger 조건

본 skill 의 trigger = **local manual, PR merge 후 사용자가 1회 실행**. (옵션 B: Local hook + `gh pr merge` 후 wrapper 수동 실행, 사용자 confirm).

```bash
# PR merge 직후 사용자가 본인 머신에서:
gh pr merge <num> --squash
bash scripts/wiki-pr-update.sh --pr <num> --apply
```

CI hook (GitHub Actions) 은 **사용 안 함** — GitHub hosted runner 는 사내 host (vault `ai-workflow/wiki/`) 접근 불가. 사내 self-hosted runner 도입 시 forward path.

### 1.2 사용 예시

```bash
# 1. dry-run preview (PR metadata + side effect preview, no actual change)
bash scripts/wiki-pr-update.sh --pr 552

# 2. 실제 vault 갱신 (prs/<num>.md + log.md + index.md)
bash scripts/wiki-pr-update.sh --pr 552 --apply

# 3. touched file re-ingest + vault 갱신 (mirror-list 의 source 와 매칭 시)
bash scripts/wiki-pr-update.sh --pr 552 --reingest --apply

# 4. my-harness project (multi-project 지원)
bash scripts/wiki-pr-update.sh --pr <num> --project my-harness --apply
```

### 1.3 출력 예시 (dry-run)

```
[wiki-pr-update] step 1/3: PR metadata extract (#552, project=devhub)
[wiki-pr-update]   state: merged
[wiki-pr-update]   title: feat(wiki): wiki-ingest-from-raw skill (D-72 Phase 3) 본 저장소 측 wrapper
[wiki-pr-update]   head.sha: 43da841f9a4b3e7c2f8b1d6e5c4a9b8d7e6f5a4b
[wiki-pr-update]   mergedAt: 2026-06-11T03:30:00Z
[wiki-pr-update]   touched files: 6
[wiki-pr-update] step 3/3: vault side effects (prs/<num>.md + log.md + index.md)
[wiki-pr-update]   command: python3 ~/repos/my_harness/ai-workflow/skills/wiki-pr-update/scripts/run_wiki_pr_update.py \
                    --vault-path /home/yklee/wiki \
                    --project devhub --pr 552 \
                    --pr-metadata /tmp/pr-update-XXXXXX.json \
                    --touched-files /tmp/touched-XXXXXX.txt
[wiki-pr-update]   preview: 1 file would be created (prs/552.md), 1 line appended (log.md)
[wiki-pr-update] DONE (dry-run, no actual change)
```

## 2. 입력 계약 (input contract)

### 2.1 옵션 (wrapper = `scripts/wiki-pr-update.sh`)

| 옵션 | 필수 | 기본값 | 설명 |
|---|---|---|---|
| `--pr <num>` | yes | — | PR number (1+). |
| `--project <name>` | no | `devhub` | `devhub` \| `my-harness`. |
| `--reingest` | no | off | PR touched file 중 mirror-list source 와 매칭 시 `wiki-ingest-from-raw --source <file> --apply` re-run. |
| `--apply` | no | off | dry-run (default) 의 반대. 실제 vault 갱신. |
| `--quiet` | no | off | stderr 메시지 최소화. |
| `-h`, `--help` | no | — | usage 출력 후 exit 0. |

### 2.2 Exit code

| code | 의미 |
|---|---|
| 0 | success (dry-run 또는 apply 모두 성공, 0 touched files 도 success, 이미 갱신된 PR 도 success) |
| 1 | `gh pr view` 실패, 또는 my_harness skill 부재, 또는 `gh` CLI 부재/미인증, 또는 vault side effect 실패 |
| 2 | invalid option 또는 required option (--pr) 부재 |

## 3. 출력 계약 (output contract)

### 3.1 stdout

본 skill 의 stdout 은 **read-only** (preview 또는 결과) — vault side effect 는 vault file system 에 직접 발생.

### 3.2 stderr

- 진행 메시지 (`[wiki-pr-update] step 1/3 ...`, ...)
- PR metadata 추출 결과 (state, title, head.sha, mergedAt, touched file count)
- vault side effect 결과 (file created, log line appended)
- error 메시지 (`[wiki-pr-update] error: ...`)

### 3.3 vault side effects (--apply 모드)

| 영역 | 변경 |
|---|---|
| `wiki/projects/<project>/prs/<num>.md` | **신규** page (frontmatter 8 key + body: PR metadata + touched files + summary + related sources) |
| `log.md` | 1 line append: `## [<date>] pr-update \| pr-<num>: <title> \| <touched file count> files` |
| `index.md` | `prs` 섹션에 1줄 추가 (idempotent: 이미 있으면 skip) |
| (--reingest) `wiki/projects/<project>/sources/<title>.md` | touched source 마다 re-ingest (wiki-ingest-from-raw 의 --source --apply dispatch) |

## 4. 권한 (permissions)

본 skill 의 권한 경계는 **`ai-workflow/wiki/AGENTS.md` v1.5 (D-71) 의 §11.1 D-72 cross-project + §6 금지 + AGENTS.md 본 저장소 §6.5 PR tier 정책**:

### 4.1 허용

- `wiki/projects/<project>/prs/<num>.md` **신규** 파일 작성
- `log.md` **append** (idempotent, 같은 date+pr_num 가 이미 있으면 skip)
- `index.md` 의 `prs` 섹션 갱신
- (--reingest) `wiki-ingest-from-raw --source <file> --apply` dispatch (D-72 skill 의 권한 상속)

### 4.2 금지

- `raw/` 수정 (AGENTS.md §6, 절대 금지)
- `wiki/concepts/`, `wiki/entities/`, `wiki/topics/`, `wiki/sources/`, `wiki/comparisons/`, `wiki/query/` 등 read-only 영역 수정 (PR-update 의 책임 외)
- `schema/` 수정
- `ai-workflow/wiki/AGENTS.md` 수정
- 다른 project 의 `wiki/projects/<other>/` 수정
- **vault Gitea remote push** (AGENTS.md 본 저장소 §6.5 정책 — 사외/사내 2-tier 형상관리 분리. vault = 사내 in-repo (v0.7.17+). **본 skill 은 local vault 만 갱신, push 는 사용자 수동 결정 영역**)
- 자동 lint 결과 자동 머지 (AGENTS.md §6)
- `main` branch 의 자동 머지 (PR-auto-update ≠ PR-auto-merge)

## 5. Idempotency

| 키 | 동작 |
|---|---|
| `pr-<num>-<head.sha>` | vault state 확인 — 이미 `prs/<num>.md` 가 존재하고 frontmatter `last_touched >= head.sha` 이면 **skip + log "already updated"** (no side effect, exit 0) |
| `pr-<num>` (different sha) | PR 이 force-pushed / amended 된 경우 — 기존 `prs/<num>.md` 의 frontmatter `last_touched` 갱신 + body 의 `head.sha` 필드 갱신 (idempotent re-write) |
| `log.md` 의 `## [<date>] pr-update \| pr-<num>` | 같은 line 이 이미 있으면 skip (idempotent append) |
| `index.md` 의 prs 섹션 | 같은 `[[pr-<num>]]` entry 가 이미 있으면 skip (idempotent update) |

## 6. 실패 규칙 (failure rules)

| 실패 | 처리 |
|---|---|
| `gh` CLI 부재 | exit 1 + `https://cli.github.com` 안내 |
| `gh` 미인증 | exit 1 + `gh auth login` 안내 |
| PR 부재 (closed/merged 안 됨, 또는 다른 repo) | exit 1 + stderr 에 `gh pr view <num>` 의 raw error |
| my_harness skill 부재 (`run_wiki_pr_update.py`) | exit 1 + SSOT 경로 안내 |
| `ai-workflow/wiki/AGENTS.md` 부재 | exit 1 + hint `wiki-init` (my_harness D-71 §2.2) |
| `prs/` 디렉터리 부재 | 자동 생성 (mkdir -p) 후 진행 |
| 0 touched files | 정상 (PR 이 docs-only 가 아닌 경우, 예: workflow 변경만). exit 0 + "no touched files" log |
| `--reingest` 시 touched file 이 mirror-list 와 미매칭 | skip + "not in mirror list" log. 다른 file 들은 정상 진행 |
| vault 권한 오류 (raw/schema 수정 시도) | 절대 발생 안 함 (skill scope 외) |
| 동시성 (2 PR 동시 merge) | vault Git 의 atomic commit 으로 race 회피. log.md append 는 git 의 atomic write 보장 (Obsidian Git plugin + 본인 sole owner) |
| vault Gitea push 실패 | out of scope (local 만, push 는 사용자 수동) |

## 7. 다음 행동 (Phase 1 → Phase 3)

| ID | 우선순위 | 작업 | 의존 |
| --- | --- | --- | --- |
| T-d-80-1 | P3 | 본 skill 의 본 저장소 측 wrapper 작성 (`scripts/wiki-pr-update.sh` + 본 가이드) | — |
| T-d-80-2 | P3 | my_harness 측 `wiki_pr_update_skill_spec.md` (§1~§11) + `skills/wiki-pr-update/SKILL.md` + `scripts/run_wiki_pr_update.py` 작성 | — |
| T-d-80-3 | P3 | dry-run 검증 (PR #552 — 실제 머지된 PR, 6 touched file) | T-d-80-2 |
| T-d-80-4 | P3 | `--apply` 1회 실행 (prs/552.md + log.md 1 line + index.md prs 섹션) | T-d-80-3 |
| T-d-80-5 | P3 | idempotency 검증 (동일 PR 재실행 → skip 확인) | T-d-80-4 |
| T-d-80-6 | P3 | `--reingest` 검증 (PR #552 의 6 touched file 중 mirror-list 매칭 0건 확인 — PR #552 의 4 commit 은 모두 docs/* 변경이지만 mirror-list 의 ADR/governance/planning/setup/requirements/openapi/ai-workflow-memory 중 ADR-0002 가 cross-ref 매칭 1건 가능) | T-d-80-4 |
| T-d-80-7 | P3 | wiki-lint 통합 (`wiki-lint` skill 의 L01~L10 가 prs/<num>.md 검증) | D-74 |
| T-d-80-8 | P3 | CI integration (사내 self-hosted runner 도입 시 forward path — `pull_request: closed` + `workflow_dispatch` trigger) | my_harness v2.0 |
| T-d-80-9 | P3 | main branch 머지 후 `prs/<num>.md` 의 `mergedAt` + `mergeCommitSha` 자동 fill (현재는 dry-run 의 PR JSON 으로 채움) | T-d-80-4 |

## 8. 다음 sprint 진입

본 skill 의 본 저장소 측 = thin wrapper. 두뇌는 my_harness 측 SSOT.

본 PR (또는 다음 sprint) 의 작업:

1. **my_harness 측 spec 작성** (`wiki_pr_update_skill_spec.md` §1~§11, ingest-skill_spec.md 패턴) — 본 wrapper 가 dispatch 할 SSOT.
2. **my_harness 측 impl 작성** (`SKILL.md` + `scripts/run_wiki_pr_update.py`, stdlib only, --pr/--pr-metadata/--touched-files/--apply 옵션 + frontmatter 8 key 작성 + log.md append + index.md prs 섹션 갱신).
3. **dry-run 검증** (PR #552, 6 touched file).
4. **`--apply` 1회 실행** (vault side effect 3종 + prs/552.md body).
5. **idempotency 검증** (동일 PR 재실행).
6. **PR 발행** (`feat(llm-wiki,scripts): wiki-pr-update skill (D-80 Phase 1) — 본 저장소 wrapper + my_harness SSOT`).

## 9. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-11 | 본 sprint — wiki-pr-update skill (D-80) 의 본 저장소 측 wrapper + 본 가이드 작성 (D-72 §11.1 thin wrapper 정공법, local manual trigger) |
