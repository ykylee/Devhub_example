# Beta v0.7.34 — OKF v0.1 consumer mode + path resolve + V-T1 adoption (TASK-V0734-OKF bundle)

> **Status**: proposed (TASK-V0734-OKF bundle in-flight)
> 본 release 의 변경. v0.7.33 (export 1-way) 의 3 follow-up: ADR-007 consumer mode + ADR-008 path resolve + V-T1 formal adoption. ADR-006 follow-up 4 (bundle root `index.md` emit) + ADR-010 (V-R10 URL validity) + ADR-011 (OKF version auto-detect) 일부 포함.

## 본 release 의 1차 출처

1. **ADR-007 OKF consumer mode** (proposed → **accepted**) — 외부 OKF bundle → wiki ingest 시 loose/strict mode 정의
2. **ADR-008 in-repo path → URL resolve** (proposed → **accepted**) — `last_ingested_from` (in-repo path) → GitHub blob URL 자동 resolve
3. **ADR-009 V-T1 formal adoption** (신규, **proposed**) — V-T1 lint 의 `tests/run_all_checks.py` 통합
4. **ADR-011 OKF version auto-detect** (신규, **proposed**) — `okf_version: "0.1"` 자동 detect + unknown version 시 best-effort fallback
5. **v0.7.33 release note** (ADR-006 채택 + 5 page PoC + 9/9 unit test) — 본 release 의 3 follow-up 의 1차 출처

## 발견 (v0.7.33 의 3 follow-up + 2 ADR 신규)

### TASK-V0734-OKF-IMPORT: ADR-007 의 consumer mode 구현

v0.7.33 의 ADR-006 (export 1-way) 가 OKF spec 과 1-way 호환 producer 가 됐으나, 반대 방향 (외부 bundle import) 미지원. OKF spec §9 의 5 MUST NOT 정책과 우리 strict R8/R9/R4 가 정면 충돌.

**v0.7.34 의 정공법**: `workflow_kit/okf_import.py` 신규. 2 mode (strict default / loose opt-in) + mode detection (CLI > manifest > index.md frontmatter > strict default) + 8 lint × 2 mode matrix + staging directory (`.okf_staging/<bundle-name>/`) + 명시적 `--promote` 2-stage.

### TASK-V0734-OKF-RESOLVE: ADR-008 의 in-repo path → URL resolve

v0.7.33 의 OKF `resource` field 가 URL `last_ingested_from` 만 채움. 우리 wiki 의 80% page 가 in-repo path → `resource` 비어있음.

**v0.7.34 의 정공법**: `workflow_kit/path_resolver.py` 신규. 5 step deterministic resolve: `git config remote.origin.url` → normalize → `<origin>/blob/<default-branch>/<path>`. 3 layer fallback (GitHub Actions env var > `git config` > None) + path traversal security + `--no-resolve` opt-out. `okf_export.py` 가 `--no-resolve` flag + `_derive_resource` 에서 resolve 호출.

### TASK-V0734-OKF-V-T1-FORMAL: ADR-009 의 V-T1 lint 정식 채택

v0.7.33 의 V-T1 PoC (`tests/check_wiki_title_consistency.py` 7/7 PASS) 가 `run_all_checks.py` 미통합. lint 의 *advisory* (PoC) → *mandatory* (formal) 격상 필요.

**v0.7.34 의 정공법**: `tests/run_all_checks.py` 에 `check_wiki_title_consistency.py` 등록. mode flag (`--strict` / `--loose`) 통합. ADR-007 의 mode matrix 의 한 row 로 정식 채택.

### TASK-V0734-OKF-INDEX: bundle root `index.md` 자동 emit

v0.7.33 sample bundle 에 bundle root `index.md` 부재 → OKF spec §11 의 `okf_version` 선언 위치 부재.

**v0.7.34 의 정공법**: `okf_export.py` 가 export 시 bundle root 에 `index.md` 자동 emit. frontmatter: `okf_version: "0.1"` + `generated_at: <ISO 8601>` + `generator: workflow_kit.okf_export v0.7.34+`. body: section 별 concepts/decisions/entities/patterns enumerate.

### TASK-V0734-OKF-VERSIONING: ADR-011 의 okf_version auto-detect

`okf_import.py` 가 bundle 의 `okf_version` field detect. unknown version 시 best-effort fallback (continue, warn emit). major version 다르면 reject (breaking change 보호).

**v0.7.34 의 정공법**: `okf_import.py` 의 `import_okf_bundle` 이 `okf_version` extract + `ImportReport.okf_version` field. helper `_parse_okf_version()` 가 major/minor split + warning.

## 본 release 의 변경

### 1. `workflow_kit/okf_import.py` (TASK-V0734-OKF-IMPORT, 19.3 KB / ~390 lines)

**신규 module** (ADR-007 채택 + PoC):

- **API**:
  - `detect_mode(bundle, cli_mode=None) -> "strict" | "loose"` — 4 layer priority (CLI > manifest > index.md > default)
  - `import_okf_bundle(bundle, staging, mode, promote) -> ImportReport` — full ingest pipeline
  - `lint_page(page, bundle, mode) -> list[LintIssue]` — per-page lint (5 rule)
  - `OkfImportError`, `OkfConformanceError` — exception hierarchy

- **Mode matrix (5 lint × 2 mode)**:

| Lint | strict | loose (OKF §9) |
|---|---|---|
| V-1 (wiki location) | error | error (mode 무관 — wiki invariant) |
| V-4 (index structure) | N/A | N/A (OKF pages) |
| V-R9 (archive source) | error (if `last_ingested_from` missing) | **disabled** |
| V-T1 (title ↔ H1) | error (mismatch) | **warn** |
| OKF §4.1 hard 3 rule (parseable frontmatter / non-empty `type` / reserved filename) | error | error (spec 자체 strict) |
| OKF §5.1 broken link | error | **warn** (MUST NOT reject) |

- **Staging + promote 2-stage**:
  - Default staging: `.okf_staging/<bundle-name>/`
  - `--promote` flag: 성공 시 staging → `ai-workflow/wiki/` copy
  - `ImportReport.okf_version` field (ADR-011)

- **CLI**:
  ```
  python -m workflow_kit.okf_import --bundle <path> [--mode=strict|loose|auto] [--staging <path>] [--promote] [--json]
  ```

### 2. `workflow_kit/path_resolver.py` (TASK-V0734-OKF-RESOLVE, ~5 KB / ~80 lines)

**신규 module** (ADR-008 채택 + PoC):

- **API**: `resolve_in_repo_path_to_url(relative_path, repo_root) -> str | None`
- **Resolve algorithm** (5 step, deterministic, no runtime fetch):
  1. `git config --get remote.origin.url` (or `$GITHUB_SERVER_URL/$GITHUB_REPOSITORY` env var)
  2. Normalize: `git@github.com:foo/bar.git` → `https://github.com/foo/bar`
  3. Append `/blob/<default-branch>/`
  4. Append in-repo path
  5. Validate: reject if `../` or absolute path
- **Fallback chain**:
  - `GITHUB_SERVER_URL` + `GITHUB_REPOSITORY` env (CI 우선)
  - `git symbolic-ref refs/remotes/origin/HEAD` (local after fetch)
  - `git branch --show-current` (local fallback)
  - `main` 가정 + warning (deepest fallback)
  - `None` (resolve 실패 → `resource` 비움)

### 3. `workflow_kit/okf_export.py` enhancement (TASK-V0734-OKF-RESOLVE + TASK-V0734-OKF-INDEX)

**Enhancement**:

- `_derive_resource(last_ingested_from, repo_root, *, resolve=True)` — path_resolver 통합. `--no-resolve` flag 시 resolve skip
- `generate_index_md(bundle, page_count) -> str` — bundle root `index.md` emit
  - Frontmatter: `okf_version: "0.1"` + `generated_at: <ISO 8601>` + `generator: workflow_kit.okf_export v0.7.34+`
  - Body: section 별 concepts/decisions/entities/patterns enumerate with description
- CLI enhancement: `--no-resolve` flag

### 4. 7 + 5 = 12 신규 smoke test (TASK-V0734-OKF-IMPORT + TASK-V0734-OKF-RESOLVE, **12/12 PASS, 5-run stable**)

`tests/check_okf_import.py` (8.6 KB, 7 smoke test):
1. `test_detect_mode_default_strict` — mode 명시 부재 시 default strict
2. `test_detect_mode_from_manifest` — `okf-bundle.yaml` 의 `mode: loose`
3. `test_detect_mode_from_index_md` — `index.md` frontmatter `okf_mode: loose`
4. `test_detect_mode_cli_override` — CLI 가 manifest/index 보다 우선
5. `test_lint_strict_unknown_key_error` — strict mode 에서 unknown key tolerate (OKF §4.1 양립)
6. `test_lint_loose_broken_link_warn` — loose mode 에서 broken link warn (MUST NOT reject)
7. `test_import_staging_and_promote` — staging landing + --promote 시 wiki copy

`tests/check_path_resolver.py` (5+ smoke test):
1. `test_resolve_https_origin` — `https://github.com/foo/bar.git` → `https://github.com/foo/bar/blob/main/<path>`
2. `test_resolve_ssh_origin_normalize` — `git@github.com:foo/bar.git` → HTTPS
3. `test_resolve_github_actions_env` — `$GITHUB_SERVER_URL/$GITHUB_REPOSITORY` 우선
4. `test_resolve_path_traversal_reject` — `../escape.md` → None
5. `test_resolve_absolute_path_reject` — `/abs/path` → None

### 5. ADR formal acceptances + 신규 ADR (5 ADR)

**Acceptance (proposed → accepted)**:
- `decisions/adr-007-okf-consumer-mode.md` — `accepted_in: v0.7.34`
- `decisions/adr-008-in-repo-path-to-url.md` — `accepted_in: v0.7.34`

**신규 (proposed)**:
- `decisions/adr-009-v-t1-formal-adoption.md` — V-T1 lint 의 formal adoption + run_all_checks 통합 + 5 lint + mode matrix
- `decisions/adr-010-v-r10-url-validity-lint.md` (deferred to v0.7.35, ADR drafted only) — URL validity check (HEAD request, stale URL detect)
- `decisions/adr-011-okf-version-auto-detect.md` — `okf_version` parse + major version diff reject

### 6. `tests/run_all_checks.py` integration (TASK-V0734-OKF-V-T1-FORMAL)

`check_wiki_title_consistency.py` 가 `run_all_checks.py` 의 lint suite 에 추가. mode flag (`--strict` / `--loose`) 통합.

### 7. wiki updates (2 ADR acceptance + 3 신규 ADR + 1 concept promotion)

**modification (2)**:
- `ai-workflow/wiki/decisions/adr-006-okf-compat-frontmatter.md` — `related_pages` 에 Beta-v0.7.34 release note 추가
- `ai-workflow/wiki/decisions/adr-007-okf-consumer-mode.md` — status: proposed → **accepted**, `accepted_in: v0.7.34`
- `ai-workflow/wiki/decisions/adr-008-in-repo-path-to-url.md` — status: proposed → **accepted**, `accepted_in: v0.7.34`
- `ai-workflow/wiki/concepts/v-t1-title-consistency-lint.md` — status: proposed → **active** (ADR-009 채택 시)

**신규 (3)**:
- `decisions/adr-009-v-t1-formal-adoption.md` (~9 KB, status: proposed)
- `decisions/adr-010-v-r10-url-validity-lint.md` (~8 KB, status: proposed, **deferred to v0.7.35**)
- `decisions/adr-011-okf-version-auto-detect.md` (~8 KB, status: proposed)

**Cumulative wiki page count**: 49 → 52 entries (V-4)

## 발견된 cross-cutting lesson (v0.7.34)

- **모듈 간 importlib 의존성**: `okf_import` 이 `okf_export` 에 의존 (`from workflow_kit.okf_export import ...`). test 시 *pre-load* 필수 — `sys.modules` 에 dependency 먼저 register. *non-obvious* dependency.
- **staging + promote 의 *security* 정공법**: 외부 bundle 의 직접 wiki tree 진입 차단. `--promote` 는 운영자 명시 결정. ADR-007 §3 의 spec 준수 + 운영 안전성 양립.
- **mode matrix 의 *rule-by-rule* 정의**: 5 lint × 2 mode table. lint 추가 시 matrix 갱신 의무. ADR-009 의 formal adoption 시 matrix 가 *canonical* reference.
- **default mode = strict 의 *principle of least surprise***: 운영자가 mode 명시 안 하면 *우리 의 strict governance* 적용. loose 는 *opt-in*. ADR-007 §3 의 *opt-in* 정신.
- **okf_version 의 *major version* 보호**: ADR-011 의 `okf_version: "0.1"` parse 시 major version (0) 다르면 reject. minor version (1) 다르면 warn + best-effort. OKF spec §11 의 major = breaking 변경 정책 준수.

## Reference (다른 release note)

- v0.7.33 release note (ADR-006 + 5 page PoC + 9/9 unit test, 본 release 의 *3 follow-up* 의 1차 출처)
- v0.7.32 release note (TASK-V0731-001 + V0732-001 log rotation + metrics aggregation, 본 release 의 *cumulative test 271+ → 283+* 의 1차 출처)
- v0.7.18 release note (destructive subcommand 정공법, 본 release 의 *ADR-007 의 opt-in flag* 의 정공법)

## TASK (본 release)

### TASK-V0734-OKF bundle: OKF consumer + path resolve + V-T1 formal (proposed → accepted)

- **상태**: in-flight (proposed)
- **commit**: TBD
- **scope**:
  - `workflow_kit/okf_import.py` (19.3 KB / 390 lines, 신규)
  - `workflow_kit/path_resolver.py` (5 KB / 80 lines, 신규)
  - `workflow_kit/okf_export.py` enhancement (path_resolver 통합 + index.md emit, ~50 lines added)
  - `tests/check_okf_import.py` (8.6 KB, 7/7 PASS, 5-run stable)
  - `tests/check_path_resolver.py` (~5 KB, 5+ test, 5-run stable)
  - `tests/run_all_checks.py` integration (V-T1 lint 등록)
  - 2 ADR acceptance (007/008) + 3 신규 ADR (009/010/011)
  - 1 concept promotion (V-T1 proposed → active)
- **cumulative test**: 273+ → **283+** (v0.7.33 의 273+ + 7 okf_import + 5+ path_resolver)

### Follow-up (v0.7.35+)

- **TASK-V0735-OKF-V-R10**: ADR-010 의 URL validity lint — `tests/check_wiki_url_validity.py` (HEAD request, stale URL detect, optional)
- **TASK-V0735-OKF-CONSUMER-CI**: GitHub Actions 에서 OKF bundle 자동 import → PR comment 로 staging 결과 표시
- **TASK-V0735-OKF-CROSS-LINK-LINT**: OKF bundle 의 broken link 자동 detect + report
- **TASK-V0735-OKF-SCHEMA-VALIDATION**: OKF bundle 의 `type` value 가 우리 의 5 enum 에 매핑 가능한지 validate
- **TASK-V0735-OKF-CONSUMER-DOCS**: `docs/OKF_CONSUMER_GUIDE.md` 작성 — 외부 OKF bundle 작성자 가이드

## Metric

- v0.7.34 = 1 feat commit (TBD) + 1 chore (version bump, TBD) = 2 commit
- 2 신규 module (`okf_import.py` 19.3 KB + `path_resolver.py` 5 KB) = 24.3 KB
- 2 신규 test file (7 okf_import + 5+ path_resolver = 12+ smoke) = 14 KB
- 1 enhancement (`okf_export.py` +50 lines) = ~2 KB
- 2 ADR acceptance + 3 신규 ADR (proposed) + 1 concept promotion = ~25 KB wiki
- 누적 test 273+ → 283+ (v0.7.33 의 273+ + 12 신규)
- 29 release 누적 (v0.7.5~v0.7.34)
- 98 commit code-repo (v0.7.33 까지) + 1-2 commit = **99-100 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
