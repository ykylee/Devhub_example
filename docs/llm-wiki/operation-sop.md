# Phase 1 Operation SOP — Sync + Lint

- **문서 목적**: `~/wiki/` Obsidian vault (out-of-repo, Gitea private) 의 DevHub mirror 운영 SOP. `scripts/wiki-sync-devhub.sh` 의 실행 trigger / frequency / dry-run 절차 + wiki-lint 의 L01~L10 검증 trigger.
- **범위**: Phase 1 (in-repo 변경) + Phase 1 mirror 실행 (out-of-repo) + Phase 3 (mass ingest) 의 forward path + v2.0 (full compile) 진입 시점.
- **대상 독자**: yklee (owner), DevHub 의 LLM agent (wiki page 작성자), my_harness 작업 에이전트 (D-73 wiki-lint 옵션 추가 시).
- **상태**: in_progress (D-72 Phase 1, 2026-06-10)
- **최종 수정일**: 2026-06-10
- **관련 문서**:
  - [`./README.md`](../README.md) (5 file root index)
  - [`./scope-and-rationale.md`](../scope-and-rationale.md) (Phase 1 scope + D-72 Q1~Q6 적용)
  - [`./mirror-list.md`](../mirror-list.md) (Phase 1 source list)
  - [`./lint-config.toml`](../lint-config.toml) (per-project config)
  - `../../scripts/wiki-sync-devhub.sh` (sync script)
  - my_harness 의 [`scratch/devhub_wiki_integration_response/RESPONSE.md`](../../../scratch/devhub_wiki_integration_response/RESPONSE.md)

## 1. Sync trigger (mirror 실행 시점)

| Trigger | 빈도 | 의도 |
| --- | --- | --- |
| **수동 (즉시)** | 사용자 (yklee) 의 명시적 요청 시 | Phase 1 의 첫 mirror + Phase 3 의 mass ingest + 의도적 ADR 갱신 시 |
| **ADR 갱신 후** (수동) | 새 ADR merge 시 (예: ADR-0032 future) | ADR mirror 의 최신성 유지 (raw/ source 1:1) |
| **main flat memory sync 후** (수동) | PR 머지 후 main flat memory finalize commit 직후 | state.json + session_handoff.md + work_backlog.md 의 mirror 최신성 |
| **주기 (선택)** | 매주 1회 또는 매월 1회 (yklee 결정) | mirror drift 방지 |
| **Phase 3 forward** | mass ingest 1회 + 주기 갱신 | 30~50 wiki page 의 raw/ source 정합 |

**현시점 (Phase 1) 의 trigger = "수동 (즉시)" 만**. CI hook (push 시 자동 mirror) 미사용 (D-72 응답 §5 D-71.3 의 consumer 원칙 — wiki 는 derived, mirror 는 명시적 trigger).

## 2. Sync 절차

### 2.1 dry-run mode (검증)

```bash
# Phase 1 의 lint 검증 — dry-run mode 로 source list 정합 확인
cd ~/repos/Devhub_example_omp
bash scripts/wiki-sync-devhub.sh --dry-run
```

**기대 출력**:
```
[wiki-sync-devhub] source root: /Users/yklee/repos/Devhub_example_omp
[wiki-sync-devhub] dry-run: True (no actual mirror)
[wiki-sync-devhub] target vault: /Users/yklee/wiki (Gitea private)
[wiki-sync-devhub] collecting files...

  ADR (31 file):
    docs/adr/0001-idp-selection.md
    docs/adr/0002-rbac-policy-edit-api.md
    ...
    docs/adr/0031-build-tag-policy-review.md

  Governance (5 file):
    docs/governance/code-taxonomy.md
    docs/governance/document-standards.md
    docs/governance/keycloak_admin_responsibility.md
    docs/governance/README.md
    docs/governance/worker_division.md

  Planning (26 file):
    docs/planning/2026-06-12-inbound-source-routing-sprint-plan.md
    ...

  Setup (15 file):
    docs/setup/environment-setup.md
    ...

  Requirements + OpenAPI (2 file):
    docs/requirements.md
    docs/openapi.yaml

  AI-workflow memory (3 file):
    ai-workflow/memory/state.json
    ai-workflow/memory/session_handoff.md
    ai-workflow/memory/work_backlog.md

[wiki-sync-devhub] total: 82 file, 3.5MB (estimated)
[wiki-sync-devhub] dry-run: no changes made
[wiki-sync-devhub] PASS (dry-run)
```

### 2.2 real mode (실제 mirror)

```bash
# 사용자 confirm 후 실행
cd ~/repos/Devhub_example_omp
bash scripts/wiki-sync-devhub.sh
```

**기대 출력**:
```
[wiki-sync-devhub] source root: /Users/yklee/repos/Devhub_example_omp
[wiki-sync-devhub] dry-run: False (real mirror)
[wiki-sync-devhub] target vault: /Users/yklee/wiki
[wiki-sync-devhub] cleaning /Users/yklee/wiki/raw/projects/devhub
[wiki-sync-devhub] collecting files from /Users/yklee/repos/Devhub_example_omp
[wiki-sync-devhub] copying 82 file...
[wiki-sync-devhub] manifest: /Users/yklee/wiki/raw/projects/devhub/_manifest.md updated
[wiki-sync-devhub] DONE
  files: 82
  size:  3.5MB
  manifest: ~/wiki/raw/projects/devhub/_manifest.md
```

### 2.3 vault 부재 시 (smoke test)

```bash
# vault 가 없는 환경 (CI / 신규 owner) 에서 실행
bash scripts/wiki-sync-devhub.sh
```

**기대 출력**:
```
[wiki-sync-devhub] source root: /Users/yklee/repos/Devhub_example_omp
[wiki-sync-devhub] target vault: /Users/yklee/wiki
[wiki-sync-devhub] ERROR: vault not found: /Users/yklee/wiki
[wiki-sync-devhub] hint: 'wiki-init' 명령으로 vault 초기화 (my_harness 측 D-71 §2.2)
[wiki-sync-devhub] exit 1
```

**vault 부재 시 exit 1 + stderr 메시지** — 명시적 실패 (silent no-op 금지).

## 3. Lint trigger (wiki-lint L01~L10 검증)

### 3.1 Phase 1 의 lint

**현시점 (Phase 1) 의 lint = 미실시**. wiki-lint 의 `--project` + `--project-config` 옵션은 my_harness 측 D-73 의 작업. Phase 1 mirror 실행 후 옵션 활성되면 자동 lint 가능.

**Phase 1 의 lint 정책 = "사용자 수동 lint (선택)"**. 권장 trigger:
- `bash scripts/wiki-sync-devhub.sh` 실행 직후 (1차 lint, false positive 검토)
- 매주 1회 (drift 검증)
- Phase 3 의 wiki page 작성 후 (L08 + L10 검증)

### 3.2 Phase 3 의 lint (forward path)

**D-73 (my_harness 측) wiki-lint 옵션 추가 후**:

```bash
# DevHub 의 wiki page 만 lint (my_harness 와 격리)
python3 ~/repos/my_harness/ai-workflow/skills/wiki-lint/scripts/run_wiki_lint.py \
  --vault-path ~/wiki \
  --project devhub \
  --project-config ~/repos/Devhub_example_omp/docs/llm-wiki/lint-config.toml
```

**기대 출력**:
```json
{
  "status": "ok",
  "tool_version": "0.2.0",
  "vault_path": "/Users/yklee/wiki",
  "project": "devhub",
  "examined_at": "2026-06-10T23:30:00",
  "summary": {
    "errors": 0,
    "warns": 0,
    "infos": 0,
    "pages_scanned": 0,
    "rules_executed": ["L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10"]
  },
  "findings": []
}
```

**`pages_scanned: 0`**: Phase 1 의 mirror 만 (raw/) — wiki page 0건. **Phase 3 의 mass ingest 후 wiki page 작성 시 lint 가 실제 검증**.

## 4. Sync + Lint 주기 (forward)

| 단계 | sync 빈도 | lint 빈도 | 비고 |
| --- | --- | --- | --- |
| **Phase 1 (본 PR)** | 1회 (사용자 confirm 후) | 미실시 | mirror + script 의 source-of-truth 정합 |
| **Phase 1.5 (forward)** | ADR 새 merge 시 (수동) | ADR 새 merge 시 (수동) | ADR-0032+ 추가 시 |
| **Phase 3 (mass ingest)** | 1회 (mass ingest) + 매월 1회 (drift 검증) | mass ingest 후 1회 + 매월 1회 | 30~50 wiki page 작성 + lint 검증 |
| **v2.0 (full compile)** | 매주 1회 (자동 hook) | 매주 1회 (CI integration) | LLM 호출 + BM25+vector+MCP |

## 5. Sync 위험 + 대응

| # | 위험 | 영향 | 대응 |
| --- | --- | --- | --- |
| R-d-72-S-1 | mirror 실행 중 `~/wiki/` 의 다른 file 변경 (Obsidian 동기화) | mirror 결과 불완전 + Obsidian 충돌 | mirror 실행 시 vault 잠금 (단, Obsidian 의 atomic write 보장 X — mirror 실행 시간 단축으로 위험 감소) |
| R-d-72-S-2 | `~/wiki/` 의 Gitea push 가 다른 owner 와 충돌 | vault sync conflict | mirror + Gitea push 단일 owner (yklee) 만 — 다중 owner 시 owner 별 raw/ 분리 |
| R-d-72-S-3 | `scripts/wiki-sync-devhub.sh` 의 source list 가 stale (mirror-list.md 와 drift) | 일부 source 누락 | script 가 mirror-list.md 를 source-of-truth 로 정합 (또는 script 의 source list 자동 = `find` 의 glob) |
| R-d-72-S-4 | `target/` 또는 `node_modules/` 의 산출물 mirror | vault 비대화 (수 MB → 수 GB) | script 의 exclude list 정합 (mirror-list.md §3 의 "제외 패턴") |
| R-d-72-S-5 | `infra/idp/_archive_*/` mirror | vault 비대화 + ADR-0001/0009 cross-ref 정합 깨짐 | exclude list 정합 (`_archive_*/` 패턴) |
| R-d-72-S-6 | `docs/wiki/` (Public Wiki) 의 file mirror | 두 wiki 의 cross-pollution | exclude list 정합 (`docs/wiki/` 패턴) |
| R-d-72-S-7 | `docs/llm-wiki/` (LLM Wiki SSOT) 의 file mirror | 중복 (wiki 의 SSOT 가 자기 자신을 mirror) | exclude list 정합 (`docs/llm-wiki/` 패턴) |
| R-d-72-S-8 | `backend-core/` / `frontend/` / `backend-ai/` 의 source code mirror | vault 비대화 + LLM agent 가 코드 정합으로 잘못 사용 (코드 정합은 `code-index-update` 의 영역) | exclude list 정합 (`backend-core/`, `frontend/`, `backend-ai/` 의 source code, 단 mirror-list.md 의 `*.md` 만 docs 하위만) |
| R-d-72-S-9 | `openapi.yaml` 의 사내 한정 정보 (`internal-registry.example.com`, `kc.internal.example.com`, `devhub.example.com`, `172.16.0.0/12`) mirror | Gitea private 만 push 이므로 노출 0, 단 lint L11 (사내 패턴 검출) 자동 검출 권장 | mirror 허용 (D-72 응답 §3 + yklee 결정) + lint L11 의 D-73 작업 (선택) |
| R-d-72-S-10 | `scratch/devhub_wiki_integration_response/RESPONSE.md` (D-72 응답) mirror | 불필요 + D-72 응답은 my_harness 측 작업 — vault 비대화 | exclude list 정합 (`scratch/` 패턴) |

## 6. Mirror + Lint 의 verification

### 6.1 Phase 1 mirror 후 verification (T-d-72-2)

```bash
# 1. mirror 실행 (real)
bash scripts/wiki-sync-devhub.sh

# 2. mirror 결과 검증
ls -la ~/wiki/raw/projects/devhub/
cat ~/wiki/raw/projects/devhub/_manifest.md
find ~/wiki/raw/projects/devhub -type f | wc -l  # 기대: 82

# 3. lint 실행 (D-73 옵션 활성 후)
python3 ~/repos/my_harness/ai-workflow/skills/wiki-lint/scripts/run_wiki_lint.py \
  --vault-path ~/wiki \
  --project devhub \
  --project-config ~/repos/Devhub_example_omp/docs/llm-wiki/lint-config.toml
```

### 6.2 Phase 1 의 dry-run (CI 검증 가능, 본 PR scope)

```bash
# dry-run (실제 mirror X, source list 만 출력)
bash scripts/wiki-sync-devhub.sh --dry-run

# 기대 출력:
# - 82 file 의 source list 출력
# - 0 actual mirror
# - exit 0 (success)
```

**CI integration 가능** (D-72 forward path):
- `ci.yml` 의 workflow-lint job 의 env 에 `bash scripts/wiki-sync-devhub.sh --dry-run` step 추가
- PR 의 source list drift 자동 검증 (mirror-list.md 변경 시 CI fail)

## 7. 알려둘 trade-off (의도적 결정)

- **Phase 1 의 lint = "사용자 수동 lint (선택)"**: wiki-lint 의 `--project` + `--project-config` 옵션은 my_harness 측 D-73 의 작업. 본 PR 은 lint-config.toml 의 source 만 제공. **옵션 추가 후 자동 활성**.
- **CI hook (push 시 자동 mirror) 미사용**: D-72 응답 §5 D-71.3 의 consumer 원칙 — wiki 는 derived, mirror 는 명시적 trigger. 자동 mirror 는 raw/ 의 SSOT 위배 위험.
- **mirror list 의 core subset ~80 file**: D-72 응답 §4 #3 의 "100~200 파일" 의 1/2. domain (66) + architecture + infrastructure + validation (~100 file) 은 Phase 3 (mass ingest) 의 별도 PR. **본 PR 의 lint 검증 + mirror 실행의 검증 가능한 정공법 = 작은 core subset**.
- **mirror 실행 = 본 PR scope 외**: 본 PR 의 lint 검증은 `bash scripts/wiki-sync-devhub.sh --dry-run` 의 source list 정합만. 실제 mirror 는 사용자 (yklee) 가 Phase 1 mirror 실행 시점에 진행. **이유**: mirror 실행은 `~/wiki/` 의 out-of-repo 변경 — 본 PR 의 in-repo 검증 가능 영역 (CI 4/4) 의 검증 범위 외.
- **scratch/ exclude**: `scratch/devhub_wiki_integration_response/RESPONSE.md` (D-72 응답) 는 본 Phase 1 의 reference. **mirror 미포함** (vault 비대화 + D-72 응답은 my_harness 측 작업).
- **`backend-core/` / `frontend/` / `backend-ai/` 의 source code mirror 제외**: vault 비대화 + LLM agent 의 코드 정합은 `code-index-update` 의 영역 (D-72 응답 §4.3 의 "code-index-update 와 wiki-lint 의 대칭 구조"). 단 `docs/` 하위의 의미 있는 file (markdown + yaml + json) 만 mirror.
- **`openapi.yaml` 의 사내 한정 정보 mirror 허용**: D-72 응답 §3 + yklee 결정 (wiki 는 Gitea private 만) 으로 `sa-internal/` 격리 불요. 단 lint L11 (사내 패턴 검출) 의 D-73 작업은 권장 (선택).

## 8. 다음 세션 directive

1. `scripts/wiki-sync-devhub.sh` 작성 (BSD-rsync safe, dry-run + vault-absent no-op + mirror list 동적 + manifest 자동).
2. sprint memory 5 file + main flat memory 3 file.
3. lint 검증 (4종) + script smoke test (dry-run mode).
4. commit + push + PR 발행.
5. main flat memory finalize (post-merge sync).
6. 또는 T-d-72-2 (`bash scripts/wiki-sync-devhub.sh` 1회 실행, 사용자 confirm 후) — mirror 실행 trigger.
7. 또는 다른 sprint (backend-integration matrix / N-10 RBAC E2E 6 TC / release_v1_roadmap §3.5 N-13 정합).
