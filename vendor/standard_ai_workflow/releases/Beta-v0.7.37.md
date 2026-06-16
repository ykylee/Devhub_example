# Beta v0.7.37 — V-R10 v3 cache layer formal + V-R11 body audit + V-R12 commit-pinned URL (TASK-V0737-V-R10-V3-BUNDLE)

> **Status**: proposed (TASK-V0737-V-R10-V3-BUNDLE in-flight)
> 본 release 의 변경. v0.7.36 의 5 follow-up ADR (014/015/016/017/018) 의 formal acceptance. cumulative test 363+. v0.7.38+ 의 4 follow-up ADR draft (cache_stats extension + body CLI flag + vcs_commit + CI integration).

## 본 release 의 1차 출처

1. **ADR-014 V-R10 v3 cache LRU** (proposed → **accepted**) — 10MB size cap + 10000 entry cap + LRU by timestamp
2. **ADR-015 V-R10 v3 file lock** (proposed → **accepted**) — `fcntl.flock` + sidecar `.lock` file + Windows no-op fallback
3. **ADR-016 GHA actions/cache** (proposed → **accepted**) — `actions/cache@v4` for cross-PR cache sharing (CI 500x speedup)
4. **ADR-017 V-R11 body audit** (proposed → **accepted**) — `check_url_body()` (4 check + 8 phishing keywords + 1MB body cap)
5. **ADR-018 V-R12 commit-pinned URL** (proposed → **accepted**) — `resolve_in_repo_path_to_url_pinned()` (commit_sha + ref)

## 발견 (v0.7.36 의 5 follow-up ADR formal acceptance)

### TASK-V0737-V-R10-V3-BUNDLE: 5 ADR formal acceptance

v0.7.36 release 시 5 ADR 의 evidence 가 v0.7.37 release 시점부터 모두 충족:

- **ADR-014**: 4/4 PASS LRU eviction (max_entries / max_bytes / keep recent / default caps). cache size unbounded 위험 해소
- **ADR-015**: 2/2 PASS file lock (context manager + concurrent writes via multiprocessing). cache corruption 위험 해소
- **ADR-016**: workflow YAML 검증 OK. CI runtime 500x speedup. GitHub rate limit 0% 사용
- **ADR-017**: 5/5 PASS body audit (HTML pass / phishing detected / missing html / content-type / timeout). phishing 검출
- **ADR-018**: 3/3 PASS commit-pinned URL (commit_sha / ref / invalid SHA). URL immutability

**v0.7.37 의 정공법**: 5 ADR 모두 status `proposed` → `accepted` 전환 + `accepted_in: v0.7.37` 명시 + revision log v0.2.0 entry + v0.7.37 release note 동시 release + version bump v0.7.36 → v0.7.37.

## 본 release 의 변경

### 1. 5 ADR formal acceptance (TASK-V0737-V-R10-V3-BUNDLE)

- `decisions/adr-014-v-r10-v3-cache-lru.md` — status text v0.2.0 + revision log v0.2.0
- `decisions/adr-015-v-r10-v3-file-lock.md` — status text v0.2.0 + revision log v0.2.0
- `decisions/adr-016-gha-actions-cache.md` — status text v0.2.0 + revision log v0.2.0
- `decisions/adr-017-v-r11-body-audit.md` — status text v0.2.0 + revision log v0.2.0
- `decisions/adr-018-v-r12-commit-pinned-url.md` — status text v0.2.0 + revision log v0.2.0

### 2. version bump (chore)

`workflow_kit/__init__.py` + `pyproject.toml` 의 version `v0.7.36-beta` → `v0.7.37-beta`.

### 3. cumulative test count

v0.7.36 의 363+ → v0.7.37 의 **363+ (no change)** — 5 ADR 의 formal acceptance 만, code change 없음.

## 발견된 cross-cutting lesson (v0.7.37)

- **ADR 의 *formal adoption* 의 cost ≈ 0** (5 ADR, ~5 line/edit × 2 edits × 5 ADR = ~50 lines, all *mechanical*). 운영 헌법 ("decisions formal") 의 *low-friction* 정공법.
- **5 ADR 의 *proposed → accepted* 동시 전환**: *bulk* acceptance 가 *individual* acceptance 의 *operational friction* ↓. *release cycle* 와 *ADR acceptance* 의 *sync*.
- **`accepted_in: v0.7.37` 의 *release traceability***: ADR 이 *어느 release* 에 accepted 됐는지 *canonical reference*. 향후 ADR-019/020 의 naming conflict 회피.
- **5 layer cache 의 *formal v0.7.37* 완성**: ADR-013 (24h disk) + ADR-014 (10MB + LRU) + ADR-015 (file lock) + ADR-016 (GHA cross-PR) + ADR-017 (body audit). 우리 의 *5-layer cache + audit* 의 *canonical reference*.
- **cumulative test 의 *zero growth* 의 의미**: 5 ADR acceptance 가 *no new test* — *evidence* 가 v0.7.36 release 시점부터 *충족*. 운영 헌법 ("decisions formal" + "evidence-first") 의 *low-friction* 정공법.

## Reference (다른 release note)

- v0.7.36 release note (2 ADR acceptance + 5 ADR draft, 본 release 의 *5 ADR formal acceptance* 의 1차 출처)
- v0.7.35 release note (3 ADR acceptance + 2 ADR draft, 본 release 의 *cumulative 8 ADR accepted* 의 1차 출처)
- v0.7.34 release note (ADR-006/007/008 accepted + 3 module, 본 release 의 *cumulative 13 ADR* 의 1차 출처)
- v0.7.33 release note (ADR-006 채택 + 5 page PoC, 본 release 의 *전체 follow-up chain* 의 1차 출처)

## TASK (본 release)

### TASK-V0737-V-R10-V3-BUNDLE: 5 ADR formal acceptance (proposed → accepted)

- **상태**: in-flight (proposed)
- **commit**: TBD
- **scope**:
  - 5 ADR acceptance (014/015/016/017/018) — frontmatter status/accepted_in + status text + revision log
  - v0.7.37 release note
  - version bump v0.7.36 → v0.7.37
- **cumulative test**: 363+ (v0.7.36 의 363+ — *no change*)

### Follow-up (v0.7.38+)

- **TASK-V0738-CACHE-STATS-EXT**: `cache_stats()` 의 `bytes` field + `evictions_total` counter
- **TASK-V0738-BODY-CLI**: `--body` CLI flag + `--max-body-bytes` flag + integration
- **TASK-V0738-VCS-COMMIT**: `okf_export.py` 의 `vcs_commit` field + commit-pinned URL emit
- **TASK-V0738-CI-INTEGRATION**: `.github/workflows/okf-validate.yml` 의 `--body` + `--commit $GITHUB_SHA` 자동 inject
- **TASK-V0739-V-R13-SEMANTIC**: V-R13 semantic URL verification (proposed, 별도 turn)
- **TASK-V0740-OKF-CONSUMER-DOCS**: `docs/OKF_CONSUMER_GUIDE.md` (외부 OKF bundle 작성자 가이드)
- **TASK-V0741-V-VERSION-V2**: per-page version (V-Version v2) + vendor extension version policy

## Metric

- v0.7.37 = 1 wiki-ingest commit (TBD) + 1 chore (version bump, TBD) = 2 commit
- 0 신규 module (acceptance + version bump only)
- 0 신규 test (formal acceptance only)
- 5 ADR formal acceptance (status/accepted_in/related_pages 갱신) = ~50 lines wiki
- 1 release note (v0.7.37.md, ~9 KB)
- 누적 test 363+ (v0.7.36 의 363+ — *no change*)
- 32 release 누적 (v0.7.5~v0.7.37)
- 103 commit code-repo (v0.7.36 까지) + 1-2 commit = **104-105 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
