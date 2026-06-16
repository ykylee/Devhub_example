# Beta v0.7.35 — OKF formal acceptance bundle (TASK-V0735-OKF-FORMAL)

> **Status**: proposed (TASK-V0735-OKF-FORMAL in-flight)
> 본 release 의 변경. v0.7.33/v0.7.34 의 5 follow-up ADR 의 formal acceptance: ADR-009 (V-T1) + ADR-010 (V-R10) + ADR-011 (OKF version detect). 신규 ADR-012 (V-R10 online layer) + ADR-013 (V-R10 v2 cache) 의 draft. v0.7.36+ follow-up.

## 본 release 의 1차 출처

1. **ADR-009 V-T1 formal adoption** (proposed → **accepted**) — V-T1 lint 의 `tests/run_all_checks.py` 정식 통합
2. **ADR-010 V-R10 URL validity lint** (proposed → **accepted**) — 8 offline check 의 wiki lint suite 정식 등록
3. **ADR-011 OKF version auto-detect** (proposed → **accepted**) — `_parse_okf_version` + `_check_version_compatibility` 의 `okf_import.py` 통합
4. **ADR-012 V-R10 online layer** (신규, **proposed**) — runtime HEAD request 의 `--v-r10-online` flag (v0.7.36 follow-up)
5. **ADR-013 V-R10 v2 cache** (신규, **proposed**) — 24h disk cache + exponential backoff (v0.7.37 follow-up)

## 발견 (v0.7.34 의 3 proposed ADR + 2 신규 ADR)

### TASK-V0735-OKF-FORMAL-1: 3 ADR formal acceptance

v0.7.33/v0.7.34 의 3 proposed ADR (009/010/011) 의 evidence 가 v0.7.34 release 시점부터 모두 충족:

- **ADR-009**: V-T1 PoC 7/7 PASS + `tests/run_all_checks.py` `check_*.py` glob 으로 auto-discovered + 5 lint × 2 mode matrix 의 row 정의
- **ADR-010**: 8 offline check PoC 6/6 PASS + security hardening (https only / no private IP / no localhost / no credentials / no path traversal) + GitHub URL form warn
- **ADR-011**: `_parse_okf_version` + `_check_version_compatibility` 구현 + 5 new test (12/12 PASS) — exact match / major mismatch reject / minor higher warn / missing warn / malformed

**v0.7.35 의 정공법**: 3 ADR 모두 status `proposed` → `accepted` 전환 + `accepted_in: v0.7.35` 명시 + revision log v0.2.0 entry + v0.7.35 release note 동시 release.

### TASK-V0735-OKF-FORMAL-2: V-R10 online layer (ADR-012)

ADR-010 의 6 implementation item 중 online layer (1 item) 가 v0.7.34 에서 미구현. v0.7.35 release 시 ADR draft 만 (proposed, v0.7.36 PoC).

**ADR-012 의 1차 draft**: `check_url_online(url, timeout, retries) -> list[UrlIssue]` 의 API contract + 5+ test (200/3xx/404/5xx/timeout/TLS/DNS) + `--v-r10-online` CLI flag + `GITHUB_TOKEN` integration.

### TASK-V0735-OKF-FORMAL-3: V-R10 v2 cache (ADR-013)

ADR-010 의 6 follow-up 중 "V-R10 v2 — online cache + smart retry" 의 ADR draft. v0.7.35 release 시 ADR draft 만 (proposed, v0.7.37 PoC).

**ADR-013 의 1차 draft**: 24h disk cache (`~/.workflow_kit/url_validity_cache.json` 또는 in-memory), exponential backoff (1s/2s/4s), 5xx retry (max 3 attempt), cache invalidation policy (TTL 만).

## 본 release 의 변경

### 1. `workflow_kit/okf_import.py` enhancement (TASK-V0735-OKF-FORMAL-1, ADR-011)

ADR-011 v0.7.34 PoC 의 formal adoption (acceptance 전환만, code 변경 없음). `ImportReport.version_check` field 가 stable API.

### 2. `tests/check_okf_import.py` enhancement (12/12 PASS, 5-run stable)

ADR-011 v0.7.34 PoC 의 formal adoption (acceptance 전환만, test 변경 없음). 7 → 12 tests.

### 3. wiki updates (3 ADR acceptance + 2 신규 ADR draft)

**Acceptance (proposed → accepted)**:
- `decisions/adr-009-v-t1-formal-adoption.md` — `accepted_in: v0.7.35` + status text v0.2.0 + revision log v0.2.0 entry
- `decisions/adr-010-v-r10-url-validity-lint.md` — `accepted_in: v0.7.35` + status text v0.2.0 + revision log v0.2.0 entry
- `decisions/adr-011-okf-version-auto-detect.md` — `accepted_in: v0.7.35` + status text v0.2.0 + revision log v0.2.0 entry

**신규 (proposed)**:
- `decisions/adr-012-v-r10-online-layer.md` (~10 KB) — V-R10 online HEAD layer 의 formal spec
- `decisions/adr-013-v-r10-v2-cache.md` (~9 KB) — 24h disk cache + exponential backoff 의 formal spec

**Cumulative wiki page count**: 52 → 54 (V-4) — 2 신규 ADR

### 4. version bump (chore)

`workflow_kit/__init__.py` + `pyproject.toml` 의 version `v0.7.34-beta` → `v0.7.35-beta`.

### 5. (deferred to v0.7.36) V-R10 online layer implementation

ADR-012 의 정공법. v0.7.35 release 에서는 ADR draft 만 — code implementation 은 v0.7.36+ 별도 turn. 

### 6. (deferred to v0.7.37) V-R10 v2 cache implementation

ADR-013 의 정공법. v0.7.35 release 에서는 ADR draft 만 — code implementation 은 v0.7.37+ 별도 turn.

## 발견된 cross-cutting lesson (v0.7.35)

- **ADR 의 *proposed → accepted* 전환 의 ceremony**: status field + `accepted_in` field + revision log v0.2.0 entry + status text v0.2.0 의 4 point 일관성 유지. → 향후 ADR-014/015/016 의 accepted 전환 시 동일 4 point 적용.
- **ADR draft 만 release (v0.7.35) vs ADR + code 동시 release (v0.7.33/v0.7.34) 의 trade-off**: v0.7.33/v0.7.34 는 *evidence-first* — code/test 먼저, ADR acceptance 후속. v0.7.35 의 ADR-012/013 draft 는 *decision-first* — ADR draft 선반영, code 별도 turn. 운영 헌법 ("decisions formal") 의 두 가지 정공법.
- **`run_all_checks.py` 의 `check_*.py` glob 의 *auto-discovery* 의 효용성**: ADR-009 의 formal integration 이 *code change 0* 으로 가능. 운영 헌법 ("convention over configuration") 의 검증.
- **`ImportReport.version_check` field 의 *additive* (default None)**: backward-compatible. 기존 caller 영향 0. → 향후 `version_check` 가 `breaking change` 도입 시 default `None` 유지 + `version_check_required: bool = False` 같은 opt-in field 추가.
- **ADR 의 1차 evidence = *test PASS* + *PoC 동작* 의 두 가지 모두 필요**: code 변경 없이 acceptance 만 전환하는 ADR-009/010/011 의 경우, v0.7.34 release 시점의 PoC 가 *evidence* 역할. v0.7.36+ 의 ADR-012/013 code PoC 가 *evidence* 가 되면 acceptance 전환.

## Reference (다른 release note)

- v0.7.34 release note (ADR-006/007/008 accepted + `workflow_kit/okf_import.py` + `path_resolver.py` + 5 ADR PoC, 본 release 의 *3 ADR formal acceptance* 의 1차 출처)
- v0.7.33 release note (ADR-006 채택 + `workflow_kit/okf_export.py` + 5 page PoC + 9/9 unit test, 본 release 의 *전체 follow-up chain* 의 1차 출처)
- v0.7.18 release note (destructive subcommand 정공법, 본 release 의 *ADR formal acceptance ceremony* 의 정공법)

## TASK (본 release)

### TASK-V0735-OKF-FORMAL: OKF formal acceptance bundle (proposed → accepted)

- **상태**: in-flight (proposed)
- **commit**: TBD
- **scope**:
  - 3 ADR acceptance (009/010/011) — frontmatter status/accepted_in + status text + revision log
  - 2 신규 ADR draft (012/013) — 19 KB wiki
  - v0.7.35 release note + version bump
- **cumulative test**: 298+ (v0.7.34 의 298+ — *no new test*, acceptance 만)

### Follow-up (v0.7.36+)

- **TASK-V0736-V-R10-ONLINE**: ADR-012 의 online HEAD layer implementation — `check_url_online()` + `--v-r10-online` flag + 5+ test (mocked HTTP) + `GITHUB_TOKEN` integration
- **TASK-V0737-V-R10-V2**: ADR-013 의 24h cache + exponential backoff — disk cache file + smart retry + 5+ test
- **TASK-V0738-CI-INTEGRATION**: GitHub Actions workflow (`.github/workflows/okf-validate.yml`) — weekly cron + on-PR trigger + `--v-r10-online` + `GITHUB_TOKEN`
- **TASK-V0739-OKF-CONSUMER-DOCS**: `docs/OKF_CONSUMER_GUIDE.md` — 외부 OKF bundle 작성자 가이드
- **TASK-V0740-OKF-CROSS-LINK-LINT**: OKF bundle 의 broken link 자동 detect + report
- **TASK-V0741-V-VERSION-V2**: per-page version (V-Version v2) + vendor extension version policy

## Metric

- v0.7.35 = 1 wiki-ingest commit (TBD) + 1 chore (version bump, TBD) = 2 commit
- 0 신규 module (acceptance + ADR draft 만)
- 0 신규 test (acceptance 만)
- 3 ADR acceptance (status/accepted_in/related_pages 갱신) + 2 ADR 신규 draft (~19 KB)
- 1 release note (v0.7.35.md, ~12 KB)
- 누적 test 298+ (v0.7.34 의 298+ — *no change*)
- 30 release 누적 (v0.7.5~v0.7.35)
- 100 commit code-repo (v0.7.34 까지) + 1-2 commit = **101-102 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
