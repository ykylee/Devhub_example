# Beta v0.7.30 — TASK-V0728-001 (Archive Stale Memory Cron Integration)

> **Status**: in-flight
> 본 release 의 변경. v0.7.28 의 archive_stale_memory.py 의 *cron 자동화* — *manual* → *automated*.

## 본 release 의 1차 출처

1. **v0.7.28 release note** (TASK-V0726-004 archive_stale_memory) — 본 release 의 *cron subcommand 가 자동 호출* 의 *tool* 1차 출처
2. **`mavis cron self` 의 시스템 prompt** (cron 정공법) — 본 release 의 *mavis cron create/disable/list* 의 1차 출처
3. **v0.7.18 release note** (destructive subcommand 정공법, --dry-run 필수) — 본 release 의 *destructive subcommand 의 caller opt-in flag* 정공법
4. **v0.7.29 release note** (TASK-V0727-001 post-step 2-phase) — 본 release 의 *post-step 의 caller opt-in flag 의 *5종 정합* 의 *역사* 의 일부

## 발견 (v0.7.28 의 follow-up)

v0.7.28 의 `archive_stale_memory.py` 의 *문제점*: *manual* 호출 필요.
- *caller* 가 매 7 day 마다 *수동* 으로 `archive_stale_memory.py --older-than=30 --apply` 실행
- 운영자가 *잊으면* short SHA dir 가 *누적* → `ai-workflow/memory/` top-level dir 수 *선형 증가*
- 즉, *caller 의 discipline* 이 *필수* (Brittle)

**v0.7.30 의 정공법**: `mavis cron create` 자동 호출 — *automated* 실행. caller 는 *한 번* `archive_stale_memory.py --install-cron` 호출 → 매 interval 마다 *자동* 실행.

## 본 release 의 변경

### 1. `archive_stale_memory.py` 의 3 cron subcommand (TASK-V0728-001)

**3 신규 flag**:
- `--install-cron` — `mavis cron create <agent> <cronName> --schedule=<interval> --prompt=<text>` 자동 호출
- `--uninstall-cron` — `mavis cron disable <agent> <cronName>` 자동 호출
- `--show-cron` — `mavis cron list <agent>` 자동 호출 + `found: True/False` 반환

**2 신규 config flag**:
- `--cron-name` — cron name (default: `archive-memory`)
- `--cron-interval` — schedule interval (default: `7d`, e.g. `7d` / `1d` / `12h`)
- `--agent` — agent name (default: `mavis`)

**3 신규 helper** (`cmd_install_cron` / `cmd_uninstall_cron` / `cmd_show_cron`):
- `cmd_install_cron` — `mavis cron create <agent> <cronName> --schedule=<interval> --prompt=<...>` 자동 호출
  - prompt: `Run: python3 workflow-source/tools/archive_stale_memory.py --older-than=30 --apply --repo-root=<path> (auto-triggered by mavis cron 'archive-memory' every 7d)`
- `cmd_uninstall_cron` — `mavis cron disable <agent> <cronName>` 자동 호출
- `cmd_show_cron` — `mavis cron list <agent>` 의 output 에서 `cronName` 매치 확인. `found: True/False` 반환

**main 의 mode 분기**:
- `--install-cron` → `cmd_install_cron`
- `--uninstall-cron` → `cmd_uninstall_cron`
- `--show-cron` → `cmd_show_cron`
- (existing) `--list` / `--dry-run` / `--apply` / `--cleanup`

### 2. 5 smoke test (5/5 PASS)

`tests/check_v0_7_30_archive_cron.py` (in-process import + mock):
1. `test_install_cron_subprocess` — `--install-cron` 이 `mavis cron create` 호출 (correct args: agent + cronName + `--schedule=7d`)
2. `test_uninstall_cron_subprocess` — `--uninstall-cron` 이 `mavis cron disable` 호출
3. `test_show_cron_subprocess` — `--show-cron` 이 `mavis cron list` 호출 + `found=True/False` 정확
4. `test_cron_name_flag` — `--cron-name=<custom>` 적용
5. `test_cron_interval_flag` — `--cron-interval=<custom>` 적용 (`--schedule=<custom>` 변환)

**5/5 PASS** (232+ → 237+ 누적 test).

### 3. cumulative 237+ test (회귀 0)

- v0.7.29 의 5 신규 test = 232+
- v0.7.30 의 5 신규 test (TASK-V0728-001 5) = **237+**
- 기존 v0.7.28 의 5 test (TASK-V0726-004) 회귀 0
- 기존 `--list` / `--dry-run` / `--apply` / `--cleanup` 정상

## 발견된 cross-cutting lesson (v0.7.30)

- **`mavis cron` 의 *system prompt 정공법***: `mavis cron self <name> --every <interval> --prompt <text>` (TTL auto-cleanup) vs `mavis cron create <agent> <name> --schedule <interval> --prompt <text>` (persistent). 본 release 는 *persistent* (v0.7.18 의 destructive subcommand 와 동일 — caller 의 *명시적* 설치).
- **`mavis cron list <agent>` 의 *JSON output***: `{"tasks": [{"cronName": "..."}]}`. `cronName` 매치로 `found` 결정. **grep + field 정합성** 정공법.
- **`mavis cron disable <agent> <name>` 의 *idempotency***: 이미 disabled 인 cron 도 disable 가능 (no-op). destructive operation 의 *3-checklist* (v0.7.18): `dry-run` 으로 caller 가 명령 list 미리 검토 → `apply` 시 *1+ fail 시 graceful 중단* → `executed` list 로 부분 disable 보고.
- **`--schedule` vs `--every` 의 *용어 차이***: `mavis cron self` 는 `--every` (natural-language interval), `mavis cron create` 는 `--schedule` (cron 형식 또는 natural-language). 본 release 의 *5 smoke* 의 `--schedule=7d` 와 `--schedule=1d` 가 *모두* 정공법.

## Reference (다른 release note)

- v0.7.29 release note (TASK-V0727-001 post-step 2-phase, 본 release 의 *caller opt-in flag 의 5종 정합* 의 *역사* 의 일부)
- v0.7.28 release note (TASK-V0726-004 archive_stale_memory, 본 release 의 *cron subcommand 가 자동 호출* 의 *tool* 1차 출처)
- v0.7.27 release note (TASK-V0726-003 sync_release_hash post-step, 본 release 의 *post-step 의 caller opt-out flag* 정공법의 *역사* 의 일부)
- v0.7.21 release note (--allow-existing-tag, 본 release 의 *caller opt-in flag* 의 정공법)
- v0.7.18 release note (destructive subcommand 정공법, 본 release 의 *destructive subcommand 의 caller opt-in flag* 정공법)
- v0.7.12 release note (4-priority REPO_ROOT, 본 release 의 REPO_ROOT auto-detect 의 1차 출처)

## 1 TASK (본 release)

### TASK-V0728-001: Archive stale memory cron integration

- **commit**: TBD
- **status**: in-flight
- **scope**:
  - `tools/archive_stale_memory.py` (~80 line 추가: 3 cmd + 3 argparse flag + 2 config flag + main mode 분기)
  - `tests/check_v0_7_30_archive_cron.py` (5 smoke, 5/5 PASS)

## Follow-up (v0.7.31+)

- **F-1 (ci-publish, Phase 5)** — GH Actions 자동 release (`.github/workflows/release.yml`)
- **TASK-V0729-001**: `archive_stale_memory.py` 의 *cron prompt* 의 *run-time metrics* — *cron trigger* 마다의 *archive_count*, *skip_count*, *error_count* 의 *log* (post-mortem 분석용)
- **TASK-V0730-001**: `archive_stale_memory.py --install-cron` 의 *idempotency* — *이미* 같은 cron 이 있으면 *skip* (현재는 매번 create 시도 → conflict 가능)

## Metric

- v0.7.30 = 1 feat commit (TBD) + 1 fix(state) (TBD) = 2 commit (TASK-V0727-001 의 2-phase + amend 의 *1 commit* 정공법 적용)
- archive_stale_memory.py 변경 (~80 line: 3 cmd + 3 flag + 2 config)
- 1 신규 test file (5 smoke, 5/5 PASS)
- 누적 test 237+ (v0.7.29 232+ + 5 신규)
- 25 release 누적 (v0.7.5~v0.7.30)
- 89 commit code-repo (v0.7.29 까지) + 2 commit = **91 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
