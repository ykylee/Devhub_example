# Beta v0.7.36 — V-R10 online + cache formal adoption (TASK-V0736-OKF-ONLINE-CACHE)

> **Status**: proposed (TASK-V0736-OKF-ONLINE-CACHE in-flight)
> 본 release 의 변경. v0.7.35/v0.7.36 의 2 follow-up ADR (012/013) 의 formal acceptance. v0.7.37+ 의 4 follow-up ADR draft (014/015/016/017) — V-R10 v3 (cache LRU + file lock + GHA actions/cache) + V-R11 (body audit) + V-R12 (commit-pinned URL).

## 본 release 의 1차 출처

1. **ADR-012 V-R10 online layer** (proposed → **accepted**) — 8 online case PoC (200/3xx/404/410/5xx/429/timeout/TLS/DNS) 의 formal adoption
2. **ADR-013 V-R10 v2 cache** (proposed → **accepted**) — 24h disk cache + exponential backoff (1s/2s/4s) + max 3 retries 의 formal adoption
3. **ADR-014 V-R10 v3 cache LRU** (신규, **proposed**) — cache size cap (10MB) + LRU eviction (v0.7.37 follow-up)
4. **ADR-015 V-R10 v3 file lock** (신규, **proposed**) — `fcntl.flock` for concurrent access safety (v0.7.37 follow-up)
5. **ADR-016 GHA actions/cache** (신규, **proposed**) — cross-PR cache sharing via GHA `actions/cache` step (v0.7.37 follow-up)
6. **ADR-017 V-R11 body content audit** (신규, **proposed**) — URL 의 *body content* 검증 (HTML renderable, no phishing) (v0.7.37 follow-up)
7. **ADR-018 V-R12 commit-pinned URL** (신규, **proposed**) — `path_resolver` 의 `vcs_commit` field + commit-pinned URL (v0.7.37 follow-up)

## 발견 (v0.7.35/v0.7.36 의 2 follow-up ADR formal + 5 신규 ADR draft)

### TASK-V0736-OKF-ONLINE-CACHE-1: 2 ADR formal acceptance

v0.7.35 의 2 proposed ADR (012/013) 의 evidence 가 v0.7.36 release 시점부터 모두 충족:

- **ADR-012**: `check_url_online()` 8 case PoC 6/6 PASS + `.github/workflows/okf-validate.yml` 의 `GITHUB_TOKEN` integration + weekly cron + on-PR trigger
- **ADR-013**: `check_url_with_cache()` 24h disk cache + smart retry PoC 4/4 PASS (total 16/16 with offline + online) + `~/.workflow_kit/url_validity_cache.json` JSON format

**v0.7.36 의 정공법**: 2 ADR 모두 status `proposed` → `accepted` 전환 + `accepted_in: v0.7.36` 명시 + revision log v0.2.0 entry + v0.7.36 release note 동시 release.

### TASK-V0736-OKF-ONLINE-CACHE-2: V-R10 v3 follow-up ADR bundle (5 ADR)

v0.7.36 release 시 5 follow-up ADR 의 *decision draft* 만 작성 (v0.7.37 PoC). 본 release 의 *single commit* 의 5 page 신규:

- **ADR-014** (V-R10 v3 cache LRU): cache size cap (10MB) + LRU eviction policy
- **ADR-015** (V-R10 v3 file lock): `fcntl.flock` 기반 concurrent access safety
- **ADR-016** (GHA actions/cache): cross-PR cache sharing via GHA `actions/cache@v4` step
- **ADR-017** (V-R11 body audit): URL 의 *body content* 검증 (HTML renderable, no phishing, Content-Type check)
- **ADR-018** (V-R12 commit-pinned URL): `path_resolver` 의 `vcs_commit` field + commit-pinned URL format (`blob/<sha>/<path>`)

## 본 release 의 변경

### 1. 2 ADR acceptance (TASK-V0736-OKF-ONLINE-CACHE-1)

- `decisions/adr-012-v-r10-online-layer.md` — status: proposed → **accepted** + status text v0.2.0 + revision log v0.2.0
- `decisions/adr-013-v-r10-v2-cache.md` — status: proposed → **accepted** + status text v0.2.0 + revision log v0.2.0

### 2. 5 신규 ADR draft (TASK-V0736-OKF-ONLINE-CACHE-2)

- `decisions/adr-014-v-r10-v3-cache-lru.md` (~9 KB) — V-R10 v3 cache size cap + LRU
- `decisions/adr-015-v-r10-v3-file-lock.md` (~8 KB) — V-R10 v3 file lock + concurrent access
- `decisions/adr-016-gha-actions-cache.md` (~9 KB) — GHA actions/cache for cross-PR
- `decisions/adr-017-v-r11-body-audit.md` (~10 KB) — V-R11 body content audit
- `decisions/adr-018-v-r12-commit-pinned-url.md` (~9 KB) — V-R12 commit-pinned URL

**Cumulative wiki page count**: 55 → 60 (V-4) — 5 신규 ADR

### 3. version confirmation

version `v0.7.36-beta` 유지 (v0.7.36 = online + cache formal adoption; v0.7.37 = v3 follow-up bundle). ADR-012/013 의 formal adoption 은 v0.7.36 의 scope.

## 발견된 cross-cutting lesson (v0.7.36)

- **ADR 의 *proposed → accepted* 전환 의 ceremony 재확인**: status field + `accepted_in` field + status text v0.2.0 + revision log v0.2.0 entry 의 4 point 일관성. v0.7.35 release 의 3 ADR (009/010/011) 와 동일 ceremony.
- **5 follow-up ADR 의 *draft only* vs *draft + PoC* 의 trade-off**: 본 release 의 5 ADR (014-018) 은 *draft only* — code/test 없이 decision 만. v0.7.37 release 시 *PoC + decision* 동시 release. 운영 헌법 ("decisions formal" + "evidence-first") 의 두 가지 정공법.
- **Cumulative ADR count 의 *traceability***: 8 ADR (006/007/008/009/010/011/012/013) accepted + 5 ADR (014-018) proposed = **13 total**. 향후 ADR-019/020 의 naming conflict 회피.
- **`workflow_kit/url_validity.py` 의 *layer* 구조**: 1) offline (ADR-010, 8 check) + 2) online (ADR-012, 8 case) + 3) cache (ADR-013, 24h disk). 본 release 의 formal adoption 으로 *3-layer* canonical. v0.7.37 의 ADR-014/015 가 *layer 3 enhancement* (cache size cap + LRU + file lock).
- **`GITHUB_TOKEN` integration 의 *v0.7.36 release 시점* 의 5000 req/h rate limit**: 우리 wiki 의 ~100 URL = ~1.2% of hourly limit. *at most* 1% per CI run. cache 도입 (ADR-013) 으로 *first run 후 0%* — cache hit 시 *0 GitHub API call*.

## Reference (다른 release note)

- v0.7.35 release note (3 ADR acceptance + 2 ADR draft + release note, 본 release 의 *2 ADR acceptance* 의 1차 출처)
- v0.7.34 release note (ADR-006/007/008 accepted + 3 module, 본 release 의 *cumulative 13 ADR* 의 1차 출처)
- v0.7.33 release note (ADR-006 채택 + 5 page PoC, 본 release 의 *전체 follow-up chain* 의 1차 출처)
- v0.7.18 release note (destructive subcommand 정공법, 본 release 의 *5 ADR draft* 의 정공법)

## TASK (본 release)

### TASK-V0736-OKF-ONLINE-CACHE: 2 ADR acceptance + 5 ADR draft (proposed → accepted)

- **상태**: in-flight (proposed)
- **commit**: TBD
- **scope**:
  - 2 ADR acceptance (012/013) — frontmatter status/accepted_in + status text + revision log
  - 5 신규 ADR draft (014/015/016/017/018) — ~45 KB wiki
  - v0.7.36 release note
- **cumulative test**: 349+ (v0.7.35 의 349+ — *no new test*, formal acceptance + ADR draft only)

### Follow-up (v0.7.37+)

- **TASK-V0737-V-R10-V3-LRU**: ADR-014 의 cache size cap (10MB) + LRU eviction implementation
- **TASK-V0737-V-R10-V3-LOCK**: ADR-015 의 `fcntl.flock` for concurrent access safety
- **TASK-V0737-GHA-CACHE**: ADR-016 의 `.github/workflows/okf-validate.yml` 에 `actions/cache@v4` step 추가
- **TASK-V0737-V-R11-BODY**: ADR-017 의 `check_url_body()` (HTML renderable, no phishing) + 5+ test
- **TASK-V0737-V-R12-PINNED**: ADR-018 의 `path_resolver.py` 의 `vcs_commit` field + commit-pinned URL format
- **TASK-V0738-OKF-CONSUMER-DOCS**: `docs/OKF_CONSUMER_GUIDE.md` — 외부 OKF bundle 작성자 가이드
- **TASK-V0739-OKF-CROSS-LINK-LINT**: OKF bundle 의 broken link 자동 detect + report
- **TASK-V0740-V-VERSION-V2**: per-page version (V-Version v2) + vendor extension version policy
- **TASK-V0741-V-R13-SEMANTIC**: V-R13 semantic URL verification (URL 의 *의미* (specific commit, specific line) 가 unchanged)

## Metric

- v0.7.36 = 1 wiki-ingest commit (TBD) = 1 commit
- 0 신규 module (acceptance + ADR draft only)
- 0 신규 test (formal acceptance only)
- 2 ADR acceptance (status/accepted_in/related_pages 갱신) + 5 ADR 신규 draft (~45 KB)
- 1 release note (v0.7.36.md, ~10 KB)
- 누적 test 349+ (v0.7.35 의 349+ — *no change*)
- 31 release 누적 (v0.7.5~v0.7.36)
- 102 commit code-repo (v0.7.35 까지) + 1 commit = **103 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
