# Beta v0.7.33 — OKF v0.1 1-way export (TASK-V0733-OKF-EXPORT)

> **Status**: proposed (TASK-V0733-OKF-EXPORT in-flight)
> 본 release 의 변경. `concepts/okf-open-knowledge-format.md` (C-OKF-1 RESOLVED, 2026-06-16) + ADR-006/007/008 + V-T1 lint + sample OKF bundle 기반.

## 본 release 의 1차 출처

1. **`concepts/okf-open-knowledge-format.md`** (2026-06-16, status: active, verification_status: VERIFIED) — OKF v0.1 spec 의 우리 wiki 정리. primary source: `https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md` (457 lines)
2. **ADR-006 OKF compat frontmatter** (proposed) — wiki → OKF 1-way export 의 정공법
3. **ADR-007 OKF consumer mode** (proposed) — 외부 bundle import 시 loose mode 정의 (deferred to v0.7.34)
4. **ADR-008 in-repo path → URL resolve** (proposed) — `resource` field 자동 채움 (deferred to v0.7.34)
5. **V-T1 title consistency lint** (proposed) — frontmatter `title` ↔ body `# H1` 일치 강제

## 발견 (concepts/okf-open-knowledge-format.md + ADR-006)

### TASK-V0733-OKF-EXPORT-1: wiki ↔ 외부 spec 의 interop 부재
우리 wiki (ai-workflow/wiki/) 는 *closed island* — 외부 spec (OKF, OpenAPI, Frictionless Data Package 등) 와 양방향 호환 없음. wiki 의 frontmatter schema (5 type enum, required status, R-9 archive source) 가 외부 spec 의 loose consumer policy 와 정면 충돌.

**v0.7.33 의 정공법**: ADR-006 채택 — OKF v0.1 spec 권장 5 field (`title` / `description` / `resource` / `tags` / `timestamp`) 를 wiki frontmatter 의 *optional bridge* 로 추가. wiki schema strict 부분 (R-1~R-9, 5 type enum) unchanged. Lint unchanged. 1-way export 만 다룸.

### TASK-V0733-OKF-EXPORT-2: OKF spec 의 loose consumer policy
ADR-006 의 export 1-way 짝 — ADR-007 (loose consumer mode, **v0.7.34 deferred**).

### TASK-V0733-OKF-EXPORT-3: in-repo path 의 OKF `resource` 부재
ADR-008 의 in-repo path → GitHub blob URL auto-resolve (**v0.7.34 deferred**).

## 본 release 의 변경

### 1. `workflow_kit/okf_export.py` (TASK-V0733-OKF-EXPORT-1, 21.7 KB / 626 lines)

**신규 module** (ADR-006 + PoC):

- **API**:
  - `Frontmatter.parse(text: str) -> Frontmatter` — wiki frontmatter 파싱 (5 type enum tolerate, list field, bool field, scalar)
  - `map_frontmatter_to_okf(frontmatter, body) -> OkfMapping` — wiki → OKF spec §4.1 priority order
  - `rewrite_wiki_links_to_okf(body) -> str` — `[[wiki-link]]` → `[text](../path.md#anchor)` (SPEC.md §5.1 bundle-relative)
  - `export_wiki_page(wiki_page, out_path) -> (int, int)` — single page export
  - `export_wiki_to_okf(wiki_root, out_bundle, page_filter) -> ExportReport` — directory batch export
  - `main()` CLI — `python -m workflow_kit.okf_export --wiki <path> --out <bundle> [--include <substr>] [--exclude <substr>] [--json]`

- **Mapping** (wiki → OKF frontmatter):
  - `type` (required, 그대로 emit)
  - `title` (frontmatter 우선, 없으면 body H1 derive, 그 다음 `type` fallback)
  - `description` (frontmatter 우선, 없으면 body 첫 prose paragraph derive, 200 char cap)
  - `last_ingested_from` (URL) → `resource` (in-repo path 면 `resource` 비움, extra `last_ingested_from` key 보존)
  - `tags` ∪ `status:X` ∪ `wiki-type:X` → `tags` (de-dupe, preserve order)
  - `updated` → `timestamp` (ISO 8601 변환: `YYYY-MM-DD` → `YYYY-MM-DDTHH:MM:SSZ`)
  - extensions: `created` / `status` / `related_pages` / `adr_id` / `r9_skip` / `last_ingested_from` (in-repo) — OKF spec §4.1 "consumers SHOULD NOT reject unknown keys" 준수

- **Field ordering**: SPEC.md §4.1 priority — `type → title → description → resource → tags → timestamp → (extensions)`

- **Body rewriting**:
  - `[[path]]` → `[display](../path.md)` (anchor `#anchor` 지원)
  - `## Citations` section 자동 append (in-repo `last_ingested_from` 일 때)
  - `## See Also` section 자동 append (`related_pages` 있을 때)

### 2. 7 smoke test (TASK-V0733-OKF-EXPORT-1, 7/7 PASS, 5-run stable)

`tests/check_okf_export.py` (10.0 KB, importlib-based pattern, 7 smoke test):

1. `test_frontmatter_parse_minimal` — minimal frontmatter (type 만) 파싱
2. `test_frontmatter_parse_full` — 모든 field (list + bool + scalar) 파싱
3. `test_frontmatter_parse_missing_type_raises` — type 부재/empty → InvalidFrontmatterError
4. `test_map_to_okf_field_order` — SPEC.md §4.1 priority order 검증
5. `test_map_to_okf_derives_title_from_body` — frontmatter 부재 시 body H1 + 첫 prose derive
6. `test_rewrite_wiki_links` — `[[path]]` / `[[path#anchor]]` → bundle-relative
7. `test_export_wiki_to_okf_end_to_end` — 1 page export → `type` field + body link rewrite + See Also section

**7/7 PASS** (5-run stable).

### 3. V-T1 title consistency lint (TASK-V0733-OKF-EXPORT-4, **lint PoC only**, **ADR-009 v0.7.34 deferred**)

`tests/check_wiki_title_consistency.py` (8.8 KB, 7 smoke test, 7/7 PASS):

- **Rule**: frontmatter `title` ↔ body `# H1` exact match (post-normalize: strip + collapse whitespace + lowercase)
- **Test list**:
  1. `test_title_matches_h1` — exact match → PASS
  2. `test_title_absent_passes` — frontmatter `title` 부재 + body H1 → PASS (rule §2.1 case 2)
  3. `test_title_mismatch_fails_strict` — mismatch + strict mode → fail
  4. `test_title_mismatch_warns_loose` — mismatch + loose mode → warn (OKF §9 MUST NOT reject)
  5. `test_h1_absent_fails` — body H1 부재 → fail (strict) / warn (loose)
  6. `test_h1_normalization_whitespace` — internal whitespace collapse → PASS
  7. `test_h1_uniqueness` — multiple H1 → first match 만 검사 (current behavior, future enhancement)

- **7/7 PASS** (5-run stable). 단, **lint 는 PoC 단계** — `tests/run_all_checks.py` 에 미통합. ADR-009 채택 시 v0.7.34 에서 정식 등록.

### 4. Sample OKF bundle (TASK-V0733-OKF-EXPORT-5, 6 files / 45 KB)

`docs/samples/okf-bundle-2026-06-16/` — 우리 wiki 의 5 page 를 OKF v0.1 spec 으로 export 한 sample:

- `concepts/okf-open-knowledge-format.md` (19.4 KB)
- `concepts/wiki-source-rule-r9.md` (7.2 KB)
- `decisions/adr-001-3-layer-separation.md` (8.9 KB)
- `entities/workflow-kit.md` (4.8 KB)
- `patterns/r4-anchor-index.md` (1.6 KB)
- `README.md` (5.2 KB) — sample metadata, OKF spec conformance, reproduction, limitations

**OKF spec §4.1 conformance**: 모든 page 가 `type` (non-empty) + 권장 5 field (`title` derive from H1, `description` derive from prose, `tags` from status+wiki-type, `timestamp` ISO 8601) 충족. **§4.1 Extensions**: wiki-native field (`status`, `created`, `related_pages`, `r9_skip`, `last_ingested_from`) 가 extra key 로 보존.

### 5. wiki updates (3 ADR + 2 concept + 1 sample)

**신규 page (4)**:

- `concepts/okf-open-knowledge-format.md` (19.4 KB, status: active, verification_status: VERIFIED) — OKF spec 정리
- `concepts/v-t1-title-consistency-lint.md` (8.5 KB, status: proposed) — V-T1 rule definition
- `decisions/adr-006-okf-compat-frontmatter.md` (10.3 KB, **status: accepted**) — **본 release 에서 status: proposed → accepted 전환**
- `decisions/adr-007-okf-consumer-mode.md` (11.0 KB, status: proposed) — v0.7.34 follow-up
- `decisions/adr-008-in-repo-path-to-url.md` (12.6 KB, status: proposed) — v0.7.34 follow-up

**modification (2)**:
- `ai-workflow/wiki/index.md` — 4 anchor 추가 (okf-open-knowledge-format, v-t1-title-consistency-lint, adr-006, adr-007, adr-008). V-4 45 → 49 entries
- `ai-workflow/wiki/log.md` — 4 신규 entry (C-OKF-1 RESOLVED, OKF follow-up, ADR-007, ADR-008, V-T1 + sample)

**Cumulative wiki page count**: 36 → 45 (entries), 8 page 신규 (1 concept + 1 concept + 1 ADR + 1 ADR + 1 ADR + sample 5 page = 9 surface, 8 wiki + 5 sample)

## 발견된 cross-cutting lesson (v0.7.33)

- **OKF spec 의 loose consumer policy 와 우리 strict lint 의 *additive* 양립**: ADR-006 의 export 1-way 는 strict unchanged, ADR-007 (v0.7.34) 의 loose mode 가 양립 layer 추가. mode flag (strict/loose) 가 *additive* — strict 가 default, loose 가 opt-in.
- **OKF §4.1 priority order 의 *type → title → description → resource → tags → timestamp* 강제**: frontmatter field ordering 이 *semantic*, 단순 cosmetic 아님. OKF consumer (e.g. visualize viewer) 가 priority 따라 field 처리.
- **`last_ingested_from` 의 *dual preservation***: in-repo path 일 때 OKF `resource` 비움 + extra `last_ingested_from` key 보존. 우리 R-9 audit + OKF `resource` semantic 동시 만족.
- **frontmatter `title` 의 *optional bridge* 정신**: ADR-006 의 `title` 이 required 아닌 optional 인 이유 = 기존 wiki 100% page 가 `title` field 미사용. 강제 required 시 모든 page 갱신. *enabler*, not *enforcee* — V-T1 lint 가 점진적 채택 지원.
- **minimal YAML parser 의 *no-dependency* 정신**: `workflow_kit/okf_export.py` 가 PyYAML 의존성 없이 동작 (우리 wiki frontmatter subset 만 — `key: value` + `key: [list]` + `key:\n  - item`). 운영 헌법 ("no optional dep unless necessary") 과 양립.
- **dataclass 의 *sys.modules register* requirement**: `importlib.util.spec_from_file_location` 으로 module load 시 dataclass decorator 가 `sys.modules` 에서 호출 module 을 lookup → 명시적 `sys.modules["module"] = mod` 필요. test framework 의 *non-obvious* 의존성.

## Reference (다른 release note)

- v0.7.32 release note (TASK-V0731-001 + V0732-001 log rotation + metrics aggregation, 본 release 의 *cumulative test 257+ → 271+* 의 1차 출처)
- v0.7.18 release note (destructive subcommand 정공법, 본 release 의 *ADR-007 loose mode 의 opt-in 정신* 의 정공법)
- v0.7.12 release note (4-priority REPO_ROOT, 본 release 의 *no PyYAML dep* 정공법)

## TASK (본 release)

### TASK-V0733-OKF-EXPORT: OKF v0.1 1-way export (proposed → accepted)

- **상태**: in-flight (proposed)
- **commit**: TBD
- **scope**:
  - `workflow_kit/okf_export.py` (21.7 KB / 626 lines, PoC)
  - `tests/check_okf_export.py` (10.0 KB, 7/7 PASS, 5-run stable)
  - `tests/check_wiki_title_consistency.py` (8.8 KB, 7/7 PASS, 5-run stable) — **PoC, 미통합**
  - 4 wiki page (1 concept okf, 1 concept v-t1, 3 ADR 006/007/008)
  - 1 sample bundle (6 files / 45 KB)
- **cumulative test**: 257+ → **271+** (v0.7.32 의 257+ + 7 okf + 7 v-t1)

### Follow-up (v0.7.34+)

- **TASK-V0734-OKF-IMPORT**: ADR-007 의 loose consumer mode 구현 — `workflow_kit/okf_import.py` (신규) + `tests/check_okf_import.py` (mode matrix 8 lint × 2 mode)
- **TASK-V0734-OKF-RESOLVE**: ADR-008 의 in-repo path → URL resolve — `workflow_kit/path_resolver.py` (신규) + `tests/check_path_resolver.py` (5+ test)
- **TASK-V0734-OKF-INDEX**: bundle root `index.md` 자동 emit + `okf_version: "0.1"` frontmatter
- **TASK-V0734-OKF-V-T1-FORMAL**: ADR-009 V-T1 formal adoption + `tests/run_all_checks.py` 통합
- **TASK-V0734-OKF-V-R10**: URL validity lint (HEAD request, stale URL detect) — ADR-010 후보
- **TASK-V0734-OKF-VERSIONING**: `okf_version` 자동 detect + unknown version 시 동작 정책 — ADR-010 후보

## Metric

- v0.7.33 = 1 feat commit (TBD) + 1 chore (version bump, TBD) = 2 commit (TASK-V0733-OKF-EXPORT 의 1 commit 정공법)
- workflow_kit/okf_export.py 신규 (~626 lines, 1 module)
- 2 신규 test file (7 okf + 7 v-t1, 14 smoke)
- 4 신규 wiki page (1 concept + 1 concept + 1 ADR accepted + 2 ADR proposed) + 1 sample bundle (5 page + README)
- 누적 test 257+ → 271+ (v0.7.32 의 257+ + 14 신규)
- 28 release 누적 (v0.7.5~v0.7.33)
- 96 commit code-repo (v0.7.32 까지) + 1-2 commit = **97-98 commit** (estimated)
- wheel + sdist 빌드 + gh release + verify (read-only)
