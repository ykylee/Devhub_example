# Work Backlog — feat/work_260610-d-72-wiki-phase-1

- 문서 목적: 본 sprint 의 todo (planned → in_progress → done) + carry-over 추적.
- 범위: 본 sprint = 4 task (T-d-72-1..4) + 8 carry-over (T-d-72-5 mirror 실행 / D-73 wiki-lint 옵션 / D-74 _lint/ 셋업 / Phase 3 mass ingest / wiki/cross/ / v2.0 / N-13 정합 / wiki-lint CI). 본 PR 의 scope = T-d-72-1 (script + 5 file 작성) + T-d-72-2 (lint 4종 all PASS + script smoke test) + T-d-72-3 (commit + push + PR 발행) + T-d-72-4 (main flat memory sync, post-merge).
- sprint branch: feat/work_260610-d-72-wiki-phase-1
- 대상 독자: 후속 에이전트, PR reviewer, owner
- 상태: in_progress (사용자 confirm 후 commit + push + PR 발행 대기)
- 최종 수정일: 2026-06-10
- 관련 문서: session_handoff.md, backlog/2026-06-10.md, state.json, pr_body.md
## 상태 정의

- planned — 미착수
- in_progress — 작업 중
- blocked — 의존성/외부 입력 대기
- done — 검증 + 완료 확정

## 본 sprint (PR, D-72 Phase 1 풀번들) task

| ID | 상태 | 우선순위 | 제목 | 비고 |
| --- | --- | --- | --- | --- |
| T-d-72-1 | done | P3 | scripts/wiki-sync-devhub.sh 작성 (BSD-rsync safe, 7 source 패턴, --dry-run + vault 부재 no-op) | my_harness 의 wiki-sync-ai-workflow.sh 와 동일 pattern. 82 file source list 출력 정상 (dry-run), vault 부재 시 명시적 error + exit 1 |
| T-d-72-2 | done | P3 | docs/llm-wiki/ 5 file + lint 4종 all PASS + script smoke test (dry-run + vault 부재) | 5 file 신규 (README / scope-and-rationale / mirror-list / lint-config / operation-sop) |
| T-d-72-3 | pending | P3 | commit + push + gh pr create --body-file pr_body.md + gh pr merge --squash --delete-branch | 사용자 confirm 대기 |
| T-d-72-4 | planned | P3 | main flat memory 3 file sync (post-merge) | 본 PR 머지 후 자동 sync |

## carry-over (다음 sprint / 사용자 trigger)

| ID | 우선순위 | 제목 | 비고 |
| --- | --- | --- | --- |
| T-d-72-5 | P3 | `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real, dry-run 아닌) → `~/wiki/raw/projects/devhub/` 에 ~82 file mirror + `_manifest.md` 자동 생성 | 사용자 confirm 후 |
| D-73 | P3 | my_harness 측: wiki-lint skill 에 `--project` + `--project-config` 옵션 추가 | 본 저장소 의 lint-config.toml 활성 |
| D-74 | P3 | my_harness 측: my_harness 의 `_lint/my-harness/` + 본 저장소 의 `_lint/devhub/` 디렉터리 셋업 | D-73 후 |
| Phase 3 | P3 | mass ingest 별도 PR — domain (66) + architecture + infrastructure + validation (~100 file) mirror + 30~50 wiki page 작성 | T-d-72-5 + D-73 + D-74 후 |
| wiki/cross/ | P3 | cross-project 종합 (my_harness 의 LLM Wiki 패턴 ↔ DevHub 의 ADR-0030 runtime injection) | Phase 3 후속 |
| v2.0 | P3 | full compile (LLM 호출 + BM25+vector+MCP) | my_harness 의 v2.0 경험 보고 진입 |
| N-13 정합 | P3 | release_v1_roadmap.md §3.5 N-13 row status = done 마킹 | housekeeping |
| wiki-lint CI | P3 | `ci.yml` 의 별도 lint job 또는 e2e shard 의 lint step 추가 | D-73 + D-74 후 |

## 사이드 점검 (본 PR 진행 시 발견)

- 기존 `docs/wiki/` (Public Wiki, mtime 2026-05-20) 와 본 Phase 1 의 `docs/llm-wiki/` (LLM Wiki SSOT) 의 분리 = 두 wiki 의 audience + source-of-truth 가 다름. `docs/wiki/` (Public, GitHub Wiki 게시 source) vs `docs/llm-wiki/` (LLM, `~/wiki/` out-of-repo). cross-link 없음.
- ADR file naming = `0001-idp-selection.md` (4-digit prefix + title), NOT `ADR-0001-...`. mirror-list.md §1.1 + script 의 `0[0-9][0-9][0-9]-*.md` 패턴 정합.
- mirror list 의 `openapi.yaml` 의 사내 한정 정보 (`internal-registry.example.com`, `kc.internal.example.com`, `devhub.example.com`, `172.16.0.0/12`) 포함되지만 D-72 응답 §3 + yklee 결정으로 Gitea private 만 push 이므로 lint L11 미사용.
- 스크립트 의 `list_sources` helper = dry-run + real 양쪽 source list 의 single source-of-truth. 중복 코드 회피.
- lint-config.toml 의 `[rules.L07].skip_if_frontmatter = ["supersedes"]` = frontmatter 의 `supersedes:` 가 있으면 L07 skip (DevHub 의 의도적 supersede 정공법).
- `scratch/devhub_wiki_integration_response/RESPONSE.md` (D-72 응답) 는 본 Phase 1 의 reference. mirror list 에 미포함 (vault 비대화 + D-72 응답은 my_harness 측 작업).
- `backend-core/` / `frontend/` / `backend-ai/` 의 source code mirror 제외. LLM agent 의 코드 정합은 `code-index-update` 의 영역. 단 `docs/` 하위의 의미 있는 file (markdown + yaml + json) 만 mirror.
