# Beta v0.7.25 — Legacy L2 Page In-Repo Migration (F-6)

> **Status**: in-flight (v0.7.25 의 in-flight work)
> 본 release 의 변경. v0.7.17 의 in-repo redirect 의 *closure* — 외부 wiki 의 legacy L2 page (15 version: v0.1.0 ~ v0.6.3) 를 in-repo 의 L2 SSOT 로 mirror.

## 본 release 의 1차 출처

1. **v0.7.17 release note** — in-repo redirect (외부 wiki → in-repo 의 ai-workflow/memory/active, 본 release 의 *redirect 의 closure* — *legacy* L2 page 의 history 보존)
2. **release_pipeline.py** 의 dry-run/apply 정공법 (v0.7.18, v0.7.21 의 destructive subcommand 정공법) — 본 tool 의 `--dry-run` / `--apply` mode 의 1차 출처
3. **refresh_wiki_memory.py** 의 4-priority REPO_ROOT auto-detect (v0.7.12 정공법) — 본 tool 의 `--repo-root` / env var / `git rev-parse` / legacy fallback 정공법
4. **emit_wiki_l2_body.py** 의 L2 frontmatter 형식 (title / type / tags / sources / last_touched / related / status / contradictions) — 본 tool 의 mirror file 의 frontmatter 의 1차 출처

## 발견

v0.7.17 의 in-repo redirect 후에도 *legacy* L2 page (= 2026-06-13 이전의 releases-*.md file) 는 *외부 wiki* (`/Users/yklee/wiki/wiki/projects/standard-ai-workflow/sources/`) 에만 잔존. in-repo 의 `ai-workflow/memory/release/<v>/` (31 dir) 는 v0.5.5~v0.7.24 까지 cover, **15 version (v0.1.0, v0.2.0, v0.3.0, v0.3.1, v0.3.2, v0.5.0, v0.5.7.1, v0.5.8, v0.5.9.1, v0.5.10.1, v0.5.11, v0.6.0.1, v0.6.1.5, v0.6.2, v0.6.3) 은 *external only***.

*in-repo 의 L2 SSOT* 가 외부 wiki 와 *정합*하지 않음 → v0.7.17 의 in-repo redirect 의 *closure* 필요.

## 본 release 의 변경

### 1. 신규 tool: `tools/migrate_legacy_l2.py` (~280 line)

**목적**: 외부 wiki 의 legacy L2 release page (15 version) → in-repo 의 `ai-workflow/memory/release/_external-wiki-legacy.md` (단일 mirror file) 로 copy.

**subcommand**: `migrate_legacy_l2` (단일 subcommand, v0.7.18+ 의 `release` / `verify` / `rollback` 의 *단일 subcommand* 정신).

**mode**: `--dry-run` / `--apply` (v0.7.18+ 의 dry-run/apply 정공법).

**구현 핵심**:
- `REPO_ROOT` 4-priority 결정 (v0.7.12 의 정공법)
- `get_external_wiki_path()` — `Path.home() / "wiki" / "wiki" / "projects" / "standard-ai-workflow" / "sources"` (legacy location)
- `discover_legacy_release_files(external_wiki)` — `releases-(alpha|beta)-v<X.Y.Z>.md` file list + SHA256
- `is_legacy_version(version, inrepo_releases)` — `version not in inrepo_releases` (정확 매치)
- `build_mirror_frontmatter(legacy_files, migrated_at, commit)` — frontmatter + 1st heading
- `build_mirror_body(legacy_files)` — 15 file 의 1:1 mirror + 목차 (Table of Contents)
- `detect_drift(inrepo_mirror, new_content)` — 기존 mirror 와 new content SHA256 비교
  - `fresh`: 부재 → write
  - `identical`: 동일 → skip (idempotency)
  - `drift`: hash mismatch → *manual review* (덮어쓰기 ❌)

**Mirror file 형식**:
```markdown
---
title: External Wiki Legacy L2 Pages (v0.1.0 ~ v0.6.3)
type: source
tags: [external-wiki, legacy, l2-mirror]
sources:
  - /Users/yklee/wiki/wiki/projects/standard-ai-workflow/sources/releases-alpha-v0.1.0.md
  - ...
migrated_from: /Users/yklee/wiki/wiki/projects/standard-ai-workflow/sources
migrated_at: 2026-06-15T...
commit: TBD
versions: [v0.1.0, v0.2.0, ...]
version_count: 15
last_touched: 2026-06-15
related: []
status: reviewed
contradictions: []
---

# External Wiki Legacy L2 Pages (v0.7.25 F-6 migration)
...
## v0.1.0
**Source**: `/Users/yklee/wiki/.../releases-alpha-v0.1.0.md`
**Size**: 1,318 bytes
**SHA256**: `be61aebe4d3c8cbd...`

```markdown
[원본 file content]
```
```

### 2. 신규 test: `tests/check_v0_7_25_legacy_l2_migration.py` (5/5 PASS)

5 가지 smoke test:
1. `test_dry_run_detect_legacy` — dry-run 이 15 legacy file detect (15 expected, 31 in-repo dir set)
2. `test_apply_writes_mirror` — apply 가 in-repo mirror file 생성 (35,032 bytes, frontmatter + 1st # + 15 ## sections)
3. `test_idempotency_skip` — 2nd apply (동일 external content + 동일 HEAD) 시 `skipped (identical)` 반환 (idempotency)
4. `test_drift_warning` — mirror file 의 `last_touched` 변조 후 2nd apply 시 `drift` 감지 + `skipped (drift — manual review)` (덮어쓰기 ❌, manual review ✅)
5. `test_unknown_args` — `--dry-run` / `--apply` 모두 부재 시 argparse error + "at least one of --dry-run or --apply" 메시지

**5/5 PASS** (202+ → 207+ 누적 test).

### 3. cumulative 207+ test

- v0.7.24 의 5 신규 test (notes-template 5 smoke) = 202+
- v0.7.25 의 5 신규 test (legacy L2 migration 5 smoke) = **207+**
- 누적 test PASS (회귀 0)

## 발견된 cross-cutting lesson (v0.7.25)

- **v0.7.17 의 in-repo redirect 의 *closure***: redirect 가 *신규* L2 page 만 cover, *legacy* L2 page (= 2026-06-13 이전) 는 *외부 wiki* 에만 잔존. *closure* 가 *필수*.
- **mirror 의 *단일 file* 정공법**: 15 file → 1 file (`_external-wiki-legacy.md`). *15 file 1:1 copy* 도 가능하지만, *single source of truth* 가 더 안전. *1 page mirror + frontmatter (versions[])* 가 정공법.
- **idempotency 의 *3-mode* 정공법 (v0.7.18+ 와 동일)**: `fresh` (write) / `identical` (skip) / `drift` (manual review). destructive subcommand 의 *graceful fail* 정공법 (v0.7.18 의 rollback 의 `--dry-run` 필수 + apply 시 graceful fail) 과 *동일 정신* — *idempotent tool* 의 정공법.
- **drift 시 *덮어쓰기 ❌***: mirror file 의 변조 (외부 wiki source 변경 등) 감지 시 *manual review* 가 정답. silent overwrite ❌, error + skip ✅. **destructive 작업의 *3-checklist*** (v0.7.18 의 rollback): dry-run 으로 caller 가 명령 list 미리 검토 → apply 시 1+ fail 시 graceful 중단 → 양쪽 동시 검증.
- **external wiki 의 *legacy file* = 15 file / 20 file (legacy + patch version)**: v0.5.7 (in-repo) ≠ v0.5.7.1 (external), v0.5.10 (in-repo) ≠ v0.5.10.1 (external). *patch version* 차이도 *legacy gap* 의 정의.

## Reference (다른 release note)

- v0.7.24 release note (--notes-template flag, 본 release 의 release note format 의 1차 출처)
- v0.7.23 release note (wiki 운영 cross-link 1-command wrapper)
- v0.7.22 release note (linter symlink-resolve fix, 본 release 의 in-repo storage 의 일관성)
- v0.7.17 release note (in-redirect, 본 release 의 redirect 의 closure)
- v0.7.18 release note (destructive subcommand 정공법, 본 release 의 drift 시 graceful fail 의 1차 출처)
- v0.7.14 release note (cmd_changelog_gen, 본 release 의 frontmatter 형식의 1차 출처)
- v0.7.12 release note (4-priority REPO_ROOT, 본 release 의 REPO_ROOT auto-detect 의 1차 출처)

## 1 TASK (본 release)

### TASK-V0725-001: Legacy L2 page in-repo migration (F-6)

- **commit**: TBD
- **status**: in-flight
- **scope**: tools/migrate_legacy_l2.py (~280 line) + tests/check_v0_7_25_legacy_l2_migration.py (5 smoke) + ai-workflow/memory/release/_external-wiki-legacy.md (mirror file, 35,032 bytes) + state/work_backlog sync

## Follow-up (v0.7.26+)

- **F-1 (ci-publish, Phase 5)** — GH Actions 자동 release (`.github/workflows/release.yml`)
- **F-7 (check_workflow_linter branch detection fix)** — 20-30 line, 즉시 fix
- **TASK-V0725-002**: external wiki 의 *legacy L2 page* 가 update 시 (e.g. 2026-06-13 의 v0.6.3 page 의 frontmatter 변경) *mirror* 자동 갱신 — wiki_emit.py 의 3-step cycle 의 *step 4* 로 통합 가능
- **TASK-V0725-003**: mirror file 의 *section* 별 분리 (e.g. `release/_legacy-from-external-wiki/alpha/` + `beta/`) — *version 별* file 분할 정공법

## Metric

- v0.7.25 = 1 feat commit (TBD) + 1 chore (TBD) = 2 commit
- 1 신규 tool (migrate_legacy_l2.py, ~280 line)
- 1 신규 test file (5 test, 5/5 PASS)
- 1 신규 mirror file (`_external-wiki-legacy.md`, 35,032 bytes, 15 version 1:1 mirror)
- 누적 test 207+ (v0.7.24 202+ + 5 신규)
- 20 release 누적 (v0.7.5~v0.7.25)
- 73 commit code-repo (v0.7.24 까지) + 2 commit = **75 commit**
- wheel + sdist 빌드 + gh release + verify (read-only)
