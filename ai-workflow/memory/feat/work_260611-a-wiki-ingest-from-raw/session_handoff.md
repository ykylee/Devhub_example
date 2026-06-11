# Session Handoff — feat/work_260611-a-wiki-ingest-from-raw

- 문서 목적: wiki-ingest-from-raw skill (D-72 Phase 3) 의 본 저장소 측 wrapper 작성. 사용자 directive "raw만 갱신되고 wiki는 업데이트가 안됐네. 스킬로 만들어서 위키에 지식을 등록하는 flow를 자동화하자" 의 본 저장소 측 정공법.
- 범위: `scripts/wiki-ingest-from-raw.sh` (5,987 bytes executable) + `docs/llm-wiki/ingest-skill.md` (6,978 bytes) + 10 file 모순점 정정 (3 commit, 10 file / +13/-13) + main flat memory 3 file + branch memory 4 file. **코드 + docs 신규**.
- 상태: branch `feat/work_260611-a-wiki-ingest-from-raw` 작업 완료, push/PR #552 발행 + 3 commit 추가 push + T-d-72-4 done. PR 머지 pending.
- 최종 수정일: 2026-06-11

## 0. 본 세션 핵심 결과

### 2-step 자동화 flow (정공법)

```
[본 저장소 ~/repos/Devhub_example_omp/]
    ↓ wiki-sync-devhub.sh (raw mirror)
[raw mirror ~/wiki/raw/projects/devhub/]
    ↓ wiki-ingest-from-raw skill (vault 의 wiki page 자동 작성)
[wiki page ~/wiki/projects/devhub/sources/<title>.md]
```

**핵심 정공법**: 본 저장소 = source-of-truth + wrapper. my_harness 측 = skill 의 SSOT (spec + impl). **vault = 공유 자원** (D-72 §11.1, user 2026-06-11 결정).

### 변경 요약 (4 file + 10 file 정정 + 4 file memory)

| 파일 | 변경 | line |
|---|---|---|
| `scripts/wiki-ingest-from-raw.sh` | 신규 executable — 2-step wrapper (wiki-sync-devhub.sh + my_harness 의 run_wiki_ingest.py) | 5,987 bytes |
| `docs/llm-wiki/ingest-skill.md` | 신규 — 본 저장소 측 사용법 가이드 (사용법, 출력, 권한, 운영 SOP, 후속) | 6,978 bytes |
| `docs/wiki/meta/publication-matrix.md` | A 등급 경로 정정 (docs/tech_stack.md → docs/shared/tech_stack.md) — commit 9124e4f0 | 1 line |
| `docs/wiki/README.md` | 관련 문서 link 정정 (publication_matrix.md → publication-matrix.md, style_guide.md → style-guide.md) — commit 9124e4f0 | 1 line |
| `docs/wiki/columns/001-role-priority-ux.md` | frontmatter draft → accepted, 최종 수정일 2026-06-11 — commit ea930f24 | 2 line |
| `docs/wiki/columns/002-system-admin-boundary.md` | frontmatter draft → accepted, 최종 수정일 2026-06-11 — commit ea930f24 | 2 line |
| `docs/wiki/columns/003-adr-lessons-auth-rbac.md` | frontmatter draft → accepted, 최종 수정일 2026-06-11 — commit ea930f24 | 2 line |
| `docs/assessment.md` | broken link 정정 (docs/tech_stack.md → docs/shared/tech_stack.md) — commit f03f491f | 1 line |
| `docs/backend_development_roadmap.md` | broken link 정정 — commit f03f491f | 1 line |
| `docs/backend/requirements_review.md` | broken link 정정 — commit f03f491f | 1 line |
| `docs/learning-session/2026-05-21/01-project-overview.md` | broken link 정정 — commit f03f491f | 1 line |
| `docs/learning-session/2026-05-21/02-roadmap-and-docs.md` | broken link 정정 — commit f03f491f | 1 line |
| `ai-workflow/memory/state.json` (main flat) | M-v1.0 notes append — wiki-ingest skill wrapper | 1 line |
| `ai-workflow/memory/work_backlog.md` (main flat) | 상태 line + 최종 수정일 갱신 | 2 line |
| `ai-workflow/memory/feat/work_260611-a-wiki-ingest-from-raw/{state.json, session_handoff.md, work_backlog.md, backlog/2026-06-11.md}` | branch memory 4 file 신규 + T-d-72-4 done 갱신 | 4 file |

### my_harness 측 SSOT (본 sprint 의 out-of-repo 작성)

| 파일 | 변경 | line |
|---|---|---|
| `~/repos/my_harness/ai-workflow/core/wiki_ingest_skill_spec.md` | 신규 spec — §1~§11 정공법 (목적, 입력 계약, 출력 계약, 동작 절차, 권한, 실패 규칙, 수동 대체) | 10,236 bytes |
| `~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/SKILL.md` | 신규 — purpose/입력/출력/권한/구현 메모/실행 예시 | ~6,000 bytes |
| `~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/scripts/run_wiki_ingest.py` | 신규 — stdlib only, --dry-run/--apply/--project/--source, frontmatter+body+cross-ref 자동 | 19,996 bytes |

### 정공법 핵심

1. **본 저장소 = wrapper (thin), my_harness = skill (thick)**. 본 sprint 의 책임 분리:
   - my_harness = skill 의 SSOT (spec + impl + lint 통합)
   - 본 저장소 = raw source + wrapper (사용자 trigger 시 1회 실행)
   - vault = 공유 자원 (D-72 §11.1, user 2026-06-11 결정)
2. **dry-run PASS (변경 없음)**: 83 source 식별, 0 errors, 0 warnings, cross-ref 자동 매칭 (예: adr-0002 → [[rbac]] 1건). 사용자 검토 후 `--apply`.
3. **2-step wrapper**: `wiki-sync-devhub.sh` (raw mirror, 기존) + `run_wiki_ingest.py` (wiki page 자동 작성, my_harness 신규) — 1 명령 통합.
4. **idempotent**: 이미 ingest 된 source 는 skip. cross-ref 의 `## Related sources` 섹션은 중복 append 방지.
5. **권한 경계**: `raw/` 절대 수정 금지, `schema/` 절대 수정 금지, 자동 lint 결과 자동 머지 금지 (AGENTS.md §6).

### 검증 (본 sprint 의 dry-run + apply 결과)

| 항목 | 값 |
|---|---|
| Source 식별 | 83 file (mirror list 의 82 + raw `_manifest.md` 1) |
| Already ingested | 0 |
| To ingest | 82 (전체, raw/_manifest.md 제외) |
| Skipped | 0 |
| Pages to create (apply 결과) | 82 신규 ingest + 기존 devhub 31 page (cross/cross/) = 113 total |
| Index.md updates | 1 (devhub count 103 → 107) |
| Log.md appends | 82 |
| Cross-ref 매칭 | 자동 (예: `adr-0002` → `[[rbac]]` 1건) |
| Errors (ingest) | 0 |
| Warnings (ingest) | 0 |
| Lint 별도 검증 | errors=8 / warns=62 / pages=124 (follow-up) |
| Report | `~/wiki/_lint/devhub/ingest_2026-06-11.md` |
| Vault commit | `b1599cc ## [2026-06-11] edit \| D-79 LLM Wiki DevHub ingest 반영 (D-72 Phase 3 결과)` |

### Pre-flight / Safety

- **Tier**: 공용 (본 저장소 wrapper = docs + script, 사내 한정 정보 미포함). my_harness 측 skill = my_harness repo (out-of-repo).
- **CI 4/4 PASS 예상** (path-detect → scripts/ + docs/llm-wiki/ + memory 변경 감지, backend skip).
- **vault 권한**: wrapper 가 my_harness skill 의 `vault/AGENTS.md` 부재 시 error_code (강제).

## 1. 다음 세션 directive

1. **PR #552 머지** (사용자 confirm 후, 현재 OPEN + 3 commit 정정 push + T-d-72-4 done).
2. **PR #551 (T-d-72-2 re-sync) 머지** (사용자 confirm 후, vault 갭 해소).
3. **PR #548 (N-13 backend foundation) 머지 결정** (E2E Internal 1 fail 해결 기대, PR #550 spec timing 안정화).
4. **wiki-lint 8 errors / 62 warns follow-up** (별도 sprint — page-level L-rule 위반 분석 + fix).
5. **T-d-72-5 (wiki/cross/) + T-d-72-6 (wiki-lint CI) + D-73/D-74 (my_harness 측 작업)**.
6. **PR A-2 (routing/auto_route.go + voc_handler 통합 + openapi.yaml)** 별도 sprint.
7. **N-6 staging 1주 운영** (사용자 결정).

## 2. 후속 (사용자 결정 영역)

- **PR #552 머지 시점**: 사용자 confirm 후.
- **PR #551 머지 시점**: 사용자 confirm 후.
- **PR #548 머지 시점**: 본 PR + PR #550 머지 후 E2E Internal 1 fail 해결 확인 후.
- **main flat memory 3 file finalize**: PR #552 머지 시점 finalize (merge_pr next action).

## 3. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-11 | 본 sprint — wiki-ingest-from-raw skill (본 저장소 wrapper + my_harness SSOT) + dry-run PASS + branch memory + PR #552 발행 |
| 2026-06-11 | 3 commit 모순점 정정 (fix(wiki) publication-matrix + README / chore(wiki) 3 columns accepted / fix(docs) 5 broken link) + T-d-72-4 --apply (82 wiki page 신규 ingest, vault commit b1599cc) + memory 4 file 갱신 |
