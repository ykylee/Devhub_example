# Beta v0.7.28 — TASK-V0726-004 (Detached HEAD Memory Dir Cleanup)

> **Status**: in-flight
> 본 release 의 변경. F-7 (v0.7.26) 의 detached HEAD → short SHA fix 의 *closure* — age-based auto-archive.

## 본 release 의 1차 출처

1. **v0.7.17 release note** (in-repo redirect, 외부 wiki → in-repo 의 ai-workflow/memory/) — 본 release 의 *memory dir SSOT* 의 1차 출처
2. **v0.7.26 release note** (F-7 branch detection, detached HEAD → short SHA) — 본 release 의 *cleanup 의 target* 의 1차 출처 (short SHA dir)
3. **v0.7.18 release note** (destructive subcommand 정공법, --dry-run 필수 + apply 시 graceful fail) — 본 release 의 *destructive operation* (file move) 의 정공법
4. **v0.7.25 release note** (F-6 closure, SHA256 hash 기반 idempotency) — 본 release 의 *idempotency* 의 1차 출처

## 발견 (v0.7.26 의 follow-up)

F-7 (v0.7.26) 의 fix 로 detached HEAD 시 memory dir name = 7-char short SHA (e.g. `ai-workflow/memory/aabb123/state.md`). *시간이 지나면*:

- *옛* commit 으로 checkout (e.g. CI 의 detached HEAD at `<sha>`) 시 *새* short SHA dir 자동 생성
- 30+ days 가 지나면 *옛* dir 들이 *누적* → `ai-workflow/memory/` 의 *top-level dir 수* 가 *선형 증가*
- *same short SHA* 로 *다른 시점* 에 checkout 시 *collision* 가능 (이론적: 16M commit 중 1% collision)
- *workflow_kit.common.paths:workflow_branch_dir()* 가 매번 *새 dir* 생성 → state 파일 / backlog 파일 *분산*

**해결**: age-based auto-archive. N day 이전의 short SHA dir → `archive/<date>/<sha>/` 로 move.

## 본 release 의 변경

### 1. 신규 tool: `tools/archive_stale_memory.py` (~250 line)

**subcommand**: `--list` (모든 short SHA dir list) / `--older-than=N --dry-run` (candidates list) / `--older-than=N --apply` (실제 archive) / `--cleanup` (apply alias).

**mode**: `--dry-run` / `--apply` / `--cleanup` (v0.7.18+ 의 dry-run/apply 정공법).

**구현 핵심**:
- `REPO_ROOT` 4-priority 결정 (v0.7.12 의 정공법)
- `get_memory_dir(repo_root)` — `ai-workflow/memory/` 의 path
- `is_short_sha_dir(path)` — `re.compile(r"^[a-f0-9]{7}$")` 의 7자 hex match. named branch (e.g. 'main', 'feat/x', 'release-v1.0.0') 는 *불일치* (correct)
- `get_dir_age_days(path)` — mtime 기준 (UTC)
- `build_archive_path(memory_dir, sha, archive_date)` — `archive/<date>/<sha>/`
- `check_already_archived(memory_dir, sha)` — archive/ 하위 *모든* dir 검색. idempotency.
- `sha256_of_dir(path)` — file 들의 SHA256 concat. v0.7.25 의 F-6 의 *1차 출처* 정공법.

**Archive target**: `ai-workflow/memory/archive/<YYYY-MM-DD>/<sha>/` (date-based, 같은 날 *여러* archive 시 *같은* dir).

**Safety**:
- `--dry-run` 시 file 이동 0, archive dir 생성 0
- `--apply` 시 *candidate* 가 *옛* dir 만 (N day 이전)
- *이미* archive 된 dir skip (idempotency)
- archive target 이미 존재 시 skip (collision 회피)
- `shutil.move` 가 *cross-device* 일 때 *fallback* (cp + rm)

### 2. 5 smoke test (5/5 PASS)

`tests/check_v0_7_28_archive_stale_memory.py`:
1. `test_short_sha_dir_detected` — 7-char hex dir name → short SHA dir 로 detect. named branch (e.g. 'main', 'feat/x') 는 *불일치* (correct). `deadbeef` (8자) 와 `old1234` (l = hex 아님) 도 *불일치*.
2. `test_age_filter` — N day 이전 dir 만 archive 후보. too-recent skip (5 day vs 40 day, --older-than=30).
3. `test_already_archived_skip` — sha 가 archive/ 하위에 *이미* 있으면 skip. idempotency.
4. `test_dry_run_no_move` — dry-run 시 file 이동 0, archive dir 생성 0.
5. `test_apply_archives_old_dir` — apply 시 옛 dir 이 archive/<date>/<sha>/ 로 move. content 보존, SHA256 동일성.

**5/5 PASS** (222+ → 227+ 누적 test).

### 3. cumulative 227+ test (회귀 0)

- v0.7.27 의 5 신규 test = 222+
- v0.7.28 의 5 신규 test (TASK-V0726-004 5) = **227+**
- 기존 `version-bump --dry-run` 정상 (회귀 0)
- 기존 `archive_stale_memory --list` 정상 (회귀 0)

## 발견된 cross-cutting lesson (v0.7.28)

- **short SHA dir pattern 의 *2-축 검증***: 길이 (7자) + charset (`[a-f0-9]`). 단일 axis (길이만 또는 charset 만) 는 *false positive* 위험. **2-축 검증** = regex `^[a-f0-9]{7}$` (정공법).
- **idempotency 의 *2-level* 검증**: (1) `check_already_archived` (path-based) + (2) `sha256_of_dir` (content-based). v0.7.25 의 F-6 의 *1차 출처* 정공법 — *path-based* 만으로는 *불충분* (다른 dir 의 *같은* sha 가 archive 되면 *collision*).
- **date-based archive path**: `archive/<YYYY-MM-DD>/<sha>/` (vs. `archive/<sha>/`). 같은 날 *여러* archive 시 *같은* dir — *manageable*. 다른 날 *같은* sha 시 *다른* dir — *history* 보존. **date-based 정공법** = *time* + *sha* 의 2-axis.
- **mtime vs ctime**: dir 의 mtime = *last modification* (file 추가/삭제 시 update). ctime = *last status change* (metadata 변경 시 update). **mtime** 이 *사용자* 의 *의도* (file touch) 와 *가장* 부합. **mtime 정공법** (vs ctime/atime).
- **`shutil.move` 의 *cross-device***: 같은 device 내 = `mv` (atomic). 다른 device = `cp` + `rm` (fallback). **자동 처리** (Python 표준 library 의 내장 동작) — caller 가 *별도 처리* 불필요.

## Reference (다른 release note)

- v0.7.27 release note (TASK-V0726-003 sync_release_hash post-step, 본 release 의 *post-step 의 caller opt-out flag* 정공법의 *역사* 의 일부)
- v0.7.26 release note (F-7 branch detection fix, 본 release 의 *cleanup 의 target* 의 1차 출처)
- v0.7.25 release note (F-6 closure, 본 release 의 *SHA256 idempotency* 의 1차 출처)
- v0.7.18 release note (destructive subcommand 정공법, 본 release 의 *destructive operation* (file move) 의 정공법)
- v0.7.17 release note (in-repo redirect, 본 release 의 *memory dir SSOT* 의 1차 출처)
- v0.7.12 release note (4-priority REPO_ROOT, 본 release 의 REPO_ROOT auto-detect 의 1차 출처)

## 1 TASK (본 release)

### TASK-V0726-004: Detached HEAD memory dir age-based auto-archive

- **commit**: TBD
- **status**: in-flight
- **scope**:
  - `tools/archive_stale_memory.py` (~250 line, 4-priority REPO_ROOT + 3 subcommand + 5 helper)
  - `tests/check_v0_7_28_archive_stale_memory.py` (5 smoke, 5/5 PASS)
  - Beta-v0.7.28.md + state/work_backlog sync

## Follow-up (v0.7.29+)

- **F-1 (ci-publish, Phase 5)** — GH Actions 자동 release (`.github/workflows/release.yml`)
- **TASK-V0727-001**: post-step 의 *chore commit 통합* — post-step 의 sync_hash 가 *chore commit* 의 file 들을 stage + amend 하는 *단일 commit* 정공법
- **TASK-V0728-001**: `archive_stale_memory.py` 의 *cron* 자동 호출 (e.g. `mavis cron self archive-memory --every 7d --prompt "..."`) — *manual* → *automated*

## Metric

- v0.7.28 = 1 feat commit (TBD) + 1 chore (TBD) = 2 commit
- 1 신규 tool (archive_stale_memory.py, ~250 line)
- 1 신규 test file (5 smoke, 5/5 PASS)
- 누적 test 227+ (v0.7.27 222+ + 5 신규)
- 23 release 누적 (v0.7.5~v0.7.28)
- 83 commit code-repo (v0.7.27 까지) + 2 commit = **85 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
