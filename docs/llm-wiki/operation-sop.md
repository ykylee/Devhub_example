# Phase 1 + Phase 1.5 Operation SOP — Sync + Lint + Wiki-only Maintenance

- **문서 목적**: `~/wiki/` Obsidian vault (out-of-repo, Gitea private) 의 DevHub mirror 운영 SOP. `scripts/wiki-sync-devhub.sh` 의 실행 trigger / frequency / dry-run 절차 + wiki-lint 의 L01~L10 검증 trigger + **위키만으로 코드 maintenance 가능 정공법** (Phase 1.5 추가).
- **범위**: Phase 1 (in-repo 변경) + **Phase 1.5 (소스코드 + workflow + scripts + branch memory + traceability mirror scope, 본 저장소 한정 + 위키만으로 maintenance 가능)** + Phase 1 mirror 실행 (out-of-repo) + Phase 3 (mass ingest) 의 forward path + v2.0 (full compile) 진입 시점.
- **대상 독자**: yklee (owner), DevHub 의 LLM agent (wiki page 작성자 + **코드 maintenance 작업 시 RAG source**), my_harness 작업 에이전트 (D-73 wiki-lint 옵션 추가 시).
- **상태**: active (D-72 Phase 1 + Phase 1.5, 2026-06-13)
- **최종 수정일**: 2026-06-13 (Phase 1.5 추가: source code + workflow + scripts + branch memory + traceability mirror scope, lint-config.toml 갱신, 위키 본문 1:1 mirror 정공법)
- **관련 문서**:
  - [`./README.md`](../README.md) (5 file root index)
  - [`./scope-and-rationale.md`](../scope-and-rationale.md) (Phase 1+1.5 scope + D-72 Q1~Q6 적용)
  - [`./mirror-list.md`](../mirror-list.md) (**Phase 1+1.5 source list, 13 패턴 ~140 file**)
  - [`./lint-config.toml`](../lint-config.toml) (**Phase 1.5 갱신: L02 broken link + L10 raw source 0 정합**)
  - `../../scripts/wiki-sync-devhub.sh` (**Phase 1.5 sync script, 13 패턴**)
  - my_harness 의 [`scratch/devhub_wiki_integration_response/RESPONSE.md`](../../../scratch/devhub_wiki_integration_response/RESPONSE.md)

## 0. 위키만으로 코드 maintenance 가능 정공법 (Phase 1.5 의 핵심, 2026-06-13)

**Phase 1.5 의 의의**: 본 저장소 의 source code (backend Go + frontend TS + workflow yml + shell scripts) 의 **maintenance critical subset ~28 file** 이 wiki mirror scope 에 포함되어, **위키만으로 코드 maintenance 가능**.

**위키의 4가지 layer 정공법**:

| Layer | Wiki 의 역할 | SSOT |
|---|---|---|
| **L1: ADR / governance / planning** | decision rationale + cross-ref | `docs/adr/*.md` + `docs/governance/*.md` + `docs/planning/*.md` |
| **L2: Setup / Requirements / OpenAPI** | 운영 SOP + REQ + API contract | `docs/setup/*.md` + `docs/requirements.md` + `docs/openapi.yaml` |
| **L3: Source code (Phase 1.5)** | backend Go + frontend e2e + workflow + scripts 의 **maintenance critical subset** | `backend-core/internal/...` + `frontend/tests/e2e/...` + `.github/workflows/*.yml` + `scripts/*.sh` |
| **L4: Traceability + ID slot** | REQ/ARCH/API/RM/IMPL/UT/TC ID + supersede chain | `docs/traceability/{report,conventions,sync-checklist}.md` |
| **L5: Branch memory (Phase 1.5)** | sprint 의 결정 + 정공법 + cross-ref | `ai-workflow/memory/<agent>/<branch>/*` |

**위키 1:1 mirror 의 verification**:
```bash
# 1. mirror byte-identical 검증
cd ~/repos/Devhub_example_minimax
python3 << 'EOF'
import os, hashlib
scope_files = []
for d in ['docs/adr', 'docs/governance', 'docs/planning', 'docs/setup']:
    for root, _, files in os.walk(d):
        for f in files:
            if f.endswith('.md'):
                scope_files.append(os.path.join(root, f))
scope_files += ['docs/requirements.md', 'docs/openapi.yaml',
                'ai-workflow/memory/state.json', 'ai-workflow/memory/session_handoff.md', 'ai-workflow/memory/work_backlog.md']
# Phase 1.5 추가: backend Go 화이트리스트 + frontend e2e 화이트리스트 + workflows + scripts + traceability
phase_15_files = [
    'backend-core/internal/domain/application-lifecycle/routing/auto_route.go',
    'backend-core/internal/domain/dev-request/view/voc_handler.go',
    'backend-core/main.go',
    'backend-core/internal/auth/keycloak_verifier.go',
    'backend-core/internal/httpapi/keycloak_admin_client.go',
    'backend-core/internal/sso-integrations/keycloak/saovae_stub.go',
    'backend-core/internal/domain/auth-session/integration/ports.go',
    'backend-core/internal/domain/auth-session/view/auth.go',
    'backend-core/internal/domain/auth-session/view/handler.go',
    'backend-core/internal/audit/middleware.go',
    'backend-core/internal/rbac/policy_store.go',
    'backend-core/internal/store/postgres/repository_ops.go',
    'frontend/tests/e2e/fixtures.ts', 'frontend/tests/e2e/signout.spec.ts',
    'frontend/tests/e2e/voc-auto-routing.spec.ts',
    'frontend/tests/e2e-manifests/smoke.txt', 'frontend/tests/e2e-manifests/quarantine.txt',
    'frontend/lib/auth/tokenStore.ts', 'frontend/lib/auth/apiClient.ts', 'frontend/lib/auth/role-routing.ts',
    '.github/workflows/ci.yml', '.github/workflows/e2e-regression.yml', '.github/workflows/e2e-quarantine.yml',
    'scripts/wiki-sync-devhub.sh', 'scripts/select-playwright-specs.sh',
    'scripts/ci-e2e-sync-check.sh', 'scripts/check-migration-uniqueness.sh',
    'docs/traceability/README.md', 'docs/traceability/conventions.md',
    'docs/traceability/report.md', 'docs/traceability/sync-checklist.md',
]
scope_files += phase_15_files
diff = 0
for f in scope_files:
    if not os.path.exists(f):
        print(f"MISSING: {f}"); diff += 1; continue
    raw = f'/Users/yklee/wiki/raw/projects/devhub/{f}'
    if not os.path.exists(raw):
        print(f"MISSING in mirror: {f}"); diff += 1; continue
    if hashlib.md5(open(f, 'rb').read()).hexdigest() != hashlib.md5(open(raw, 'rb').read()).hexdigest():
        print(f"DIFF: {f}"); diff += 1
print(f"Total: {len(scope_files)}, Diff: {diff}, Identical: {len(scope_files) - diff}")
EOF
```

**기대**: `Total: ~140, Diff: 0, Identical: ~140` (Phase 1 85 + Phase 1.5 ~55 = ~140).

**위키 1:1 mirror 의 source-of-truth**:
- **위키의 모든 page** = 본 저장소 의 raw file 의 **1:1 byte-identical mirror** (= "source-of-truth = raw/, wiki = read-only")
- 위키의 frontmatter + index 페이지 = 별도 metadata (wiki vault only, raw/ 미포함)

**위키 1:1 mirror 의 한계 (의도적 제외)**:
- `backend-core/cmd/`, `backend-core/migrations/`, `backend-core/test/` (생성물/테스트, ~5 file)
- `frontend/src/`, `frontend/components/`, `frontend/app/`, `frontend/node_modules/`, `frontend/.next/`, `frontend/dist/` (~10000+ file, 빌드 산출물)
- `infra/idp/_archive_*/` (immutable archive, ~5 file)
- `docs/wiki/` (Public Wiki, ~5 file, LLM Wiki 와 cross-link 없음)

**위키의 향후 진입 시점 (forward path)**:
- **Phase 1.5 의 mirror scope 확장**: 새 backend domain / 새 frontend e2e spec / 새 workflow 추가 시, `docs/llm-wiki/mirror-list.md` §1.7 갱신 + `scripts/wiki-sync-devhub.sh` 의 화이트리스트 갱신 + `docs/llm-wiki/lint-config.toml` 갱신.
- **Phase 3 (mass ingest)**: domain (66) + architecture (1) + infrastructure + validation (~100 file) — 본 저장소 의 wiki page 작성 시점의 별도 PR. **본 문서 갱신 시점 trigger 조건 충족 (2026-06-13)**.

## 1. Sync trigger (mirror 실행 시점)

| Trigger | 빈도 | 의도 |
| --- | --- | --- |
| **수동 (즉시)** | 사용자 (yklee) 의 명시적 요청 시 | Phase 1 의 첫 mirror + Phase 1.5 의 첫 mirror + Phase 3 의 mass ingest + 의도적 ADR 갱신 시 |
| **PR 머지 후** (수동) | 본 저장소 측 PR 머지 후 main flat memory finalize commit 직후 | state.json + session_handoff.md + work_backlog.md 의 mirror 최신성 |
| **ADR 갱신 후** (수동) | 새 ADR merge 시 (예: ADR-0032 future) | ADR mirror 의 최신성 유지 (raw/ source 1:1) |
| **Source code 갱신 후** (수동, Phase 1.5 신규) | backend Go / frontend e2e / workflow / script 의 mirror scope 내 file 변경 시 | wiki 1:1 mirror 정합 + 위키만으로 maintenance 가능 |
| **Branch memory 갱신 후** (수동, Phase 1.5 신규) | active branch memory 의 state.json / session_handoff.md / work_backlog.md / backlog/*.md / pr_body.md 변경 시 | sprint 의 결정 + 정공법 + cross-ref 의 mirror 최신성 |
| **주기 (선택)** | 매주 1회 또는 매월 1회 (yklee 결정) | mirror drift 방지 |
| **Phase 3 forward** | mass ingest 1회 + 주기 갱신 | 30~50 wiki page 의 raw/ source 정합 |

**현시점 (Phase 1+1.5) 의 trigger = "수동 (즉시)" + "PR 머지 후"**. CI hook (push 시 자동 mirror) 미사용 (D-72 응답 §5 D-71.3 의 consumer 원칙 — wiki 는 derived, mirror 는 명시적 trigger).

## 2. Sync 절차

### 2.1 dry-run mode (검증)

```bash
cd ~/repos/Devhub_example_minimax
bash scripts/wiki-sync-devhub.sh --dry-run
```

**기대 출력 (Phase 1.5 갱신)**:
```
[wiki-sync-devhub] source root: /Users/yklee/repos/Devhub_example_minimax
[wiki-sync-devhub] dry-run: True (no actual mirror)
[wiki-sync-devhub] target vault: /Users/yklee/wiki (Gitea private)
[wiki-sync-devhub] collecting files (dry-run)...

  === Phase 1 (docs subset, 7 패턴) ===
  ADR (31 file) ...
  Governance (5 file) ...
  Planning (29 file) ...
  Setup (15 file) ...
  Requirements + OpenAPI (2 file) ...
  AI-workflow memory (main flat, 3 file) ...

  === Phase 1.5 (source code + workflow + scripts + branch memory + traceability, 6 패턴) ===
  Workflows (.github/workflows/*.yml, 4 file)
  Scripts (화이트리스트, 4 file)
  Backend critical Go (화이트리스트, ~12 file)
  Frontend e2e critical (화이트리스트, 8 file)
  Traceability (4 file)
  Branch memory (active + 30일 이내 CLOSED, ~700+ file, sprint 한정)

[wiki-sync-devhub] total: ~840 file (estimated)
[wiki-sync-devhub] dry-run: no changes made
[wiki-sync-devhub] PASS (dry-run)
```

### 2.2 real mode (실제 mirror)

```bash
# 사용자 confirm 후 실행
cd ~/repos/Devhub_example_minimax
bash scripts/wiki-sync-devhub.sh
```

**기대 출력 (Phase 1.5 갱신)**:
```
[wiki-sync-devhub] DONE
  files: ~840
  size:  ~6.6MB
  manifest: /Users/yklee/wiki/raw/projects/devhub/_manifest.md
```

### 2.3 incremental mode (--no-clean)

```bash
# mirror 의 incremental 갱신 (DEST 의 기존 file 유지 + 변경분만 추가/덮어쓰기)
cd ~/repos/Devhub_example_minimax
bash scripts/wiki-sync-devhub.sh --no-clean
```

**용도**: wiki mirror 가 일관성 없는 상태 (예: 부분 mirror + manual edit) 일 때, incremental 갱신.

### 2.4 vault 부재 시 (smoke test)

```bash
# vault 가 없는 환경 (CI / 신규 owner) 에서 실행
bash scripts/wiki-sync-devhub.sh
```

**기대 출력**:
```
[wiki-sync-devhub] ERROR: vault not found: /Users/yklee/wiki
[wiki-sync-devhub] hint: 'wiki-init' 명령으로 vault 초기화 (my_harness 측 D-71 §2.2)
[wiki-sync-devhub] exit 1
```

**vault 부재 시 exit 1 + stderr 메시지** — 명시적 실패 (silent no-op 금지).

## 3. Lint trigger (wiki-lint L01~L10 검증)

### 3.1 Phase 1 + 1.5 의 lint

**현시점 (Phase 1+1.5) 의 lint = 위키의 본문 1:1 mirror 정공법 + 사용자 수동 lint (선택)**. wiki-lint 의 `--project` + `--project-config` 옵션은 my_harness 측 D-73 의 작업.

**Phase 1+1.5 의 lint 정책**:
- **위키 본문 1:1 mirror 정공법 (Phase 1.5 핵심)**: 위키의 `sources/adr-0028-...md` page 의 본문 = raw/ 의 `docs/adr/0028-...md` 의 1:1 byte-identical mirror (frontmatter + raw/ 본문). 본 정공법으로 L10 (raw/ source 0) 면제 (mirror 1:1) + L02 (broken wiki link) PASS (raw/ 의 source code link 가 mirror scope 내).
- **사용자 수동 lint (선택)**: 권장 trigger:
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
  --project-config ~/repos/Devhub_example_minimax/docs/llm-wiki/lint-config.toml
```

**기대 출력**:
```json
{
  "status": "ok",
  "tool_version": "0.2.0",
  "vault_path": "/Users/yklee/wiki",
  "project": "devhub",
  "examined_at": "2026-06-13T09:30:00",
  "summary": {
    "errors": 0,
    "warns": 0,
    "infos": 0,
    "pages_scanned": 96,
    "rules_executed": ["L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10"]
  },
  "findings": []
}
```

**`pages_scanned: 96`**: Phase 1+1.5 의 mirror (raw/ + wiki) — wiki page 96건. **L02 + L07 + L10 + L08 모두 PASS** (Phase 1.5 mirror scope 의 source code + script + workflow 의 raw/ 1:1 mirror + L07 ADR 면제 + L10 raw/ source 0 면제 + L08 index.md 갱신 완료).

## 4. Sync + Lint 주기 (forward)

| 단계 | sync 빈도 | lint 빈도 | 비고 |
| --- | --- | --- | --- |
| **Phase 1 (2026-06-10)** | 1회 (사용자 confirm 후) | 미실시 | mirror + script 의 source-of-truth 정합 |
| **Phase 1.5 (2026-06-13, 본 갱신)** | 1회 (본 sprint 의 22 file + branch memory + maintenance critical ~28 file, total ~140 file) | 사용자 수동 lint (선택) | 본 저장소 한정 + 위키만으로 코드 maintenance 가능 |
| **Phase 1.5 forward** | PR 머지 후 (수동) | PR 머지 후 (수동) | source code + workflow + script 의 mirror scope 내 file 변경 시 |
| **Phase 3 (mass ingest)** | 1회 (mass ingest) + 매월 1회 (drift 검증) | mass ingest 후 1회 + 매월 1회 | 30~50 wiki page 작성 + lint 검증 |
| **v2.0 (full compile)** | 매주 1회 (자동 hook) | 매주 1회 (CI integration) | LLM 호출 + BM25+vector+MCP |

## 5. Sync 위험 + 대응 (Phase 1.5 갱신)

| # | 위험 | 영향 | 대응 |
| --- | --- | --- | --- |
| R-d-72-S-1 | mirror 실행 중 `~/wiki/` 의 다른 file 변경 (Obsidian 동기화) | mirror 결과 불완전 + Obsidian 충돌 | mirror 실행 시 vault 잠금 (단, Obsidian 의 atomic write 보장 X — mirror 실행 시간 단축으로 위험 감소) |
| R-d-72-S-2 | `~/wiki/` 의 Gitea push 가 다른 owner 와 충돌 | vault sync conflict | mirror + Gitea push 단일 owner (yklee) 만 — 다중 owner 시 owner 별 raw/ 분리 |
| R-d-72-S-3 | `scripts/wiki-sync-devhub.sh` 의 source list 가 stale (mirror-list.md 와 drift) | 일부 source 누락 | script 의 화이트리스트 (`backend_files` + `frontend_files`) + `find` glob 동적 정합 |
| R-d-72-S-4 | `target/` 또는 `node_modules/` 의 산출물 mirror | vault 비대화 (수 MB → 수 GB) | script 의 exclude list 정합 (mirror-list.md §2 의 "제외 패턴") |
| R-d-72-S-5 | `infra/idp/_archive_*/` mirror | vault 비대화 + ADR-0001/0009 cross-ref 정합 깨짐 | exclude list 정합 (`_archive_*/` 패턴) |
| R-d-72-S-6 | `docs/wiki/` (Public Wiki) 의 file mirror | 두 wiki 의 cross-pollution | exclude list 정합 (`docs/wiki/` 패턴) |
| R-d-72-S-7 | `docs/llm-wiki/` (LLM Wiki SSOT) 의 file mirror | 중복 (wiki 의 SSOT 가 자기 자신을 mirror) | exclude list 정합 (`docs/llm-wiki/` 패턴) |
| **R-d-72-S-8a** | **Phase 1.5 추가 — backend bulk source code (`cmd/`, `migrations/`, `test/`, `internal/store/postgres/*.go` 등) 의 mirror 누락** | **위키만으로 backend maintenance 시 일부 file 의 detail 미확인 가능** | **mirror-list.md §1.7.1 + script 의 `backend_files` 화이트리스트 정합. 본 갱신 시점 12 file 화이트리스트 (PR #579 + ADR-0030/0031 정공법의 maintenance critical). forward: 새 backend file 추가 시 PR 본문에 mirror scope 추가 요청 + mirror-list 갱신.** |
| **R-d-72-S-8b** | **Phase 1.5 추가 — frontend bulk source code (`src/`, `components/`, `app/`, `node_modules/`, `.next/`) 의 mirror 누락** | **위키만으로 frontend maintenance 시 page/component 의 detail 미확인 가능** | **mirror-list.md §1.7.2 + script 의 `frontend_files` 화이트리스트 정합. 본 갱신 시점 8 file (e2e + lib/auth). forward: 새 frontend page/component 추가 시 PR 본문에 mirror scope 추가 요청.** |
| **R-d-72-S-8c** | **Phase 1.5 추가 — branch memory 의 archive 미정공법** | **active branch 외 archive branch memory 가 mirror 에 포함되어 vault 비대화** | **mirror-list.md §1.7.4 정공법: active + 30일 이내 CLOSED branch 만. forward: 30일 후 `mavis-trash` 권장.** |
| R-d-72-S-9 | `openapi.yaml` 의 사내 한정 정보 mirror | Gitea private 만 push 이므로 노출 0, 단 lint L11 (사내 패턴 검출) 자동 검출 권장 | mirror 허용 (D-72 응답 §3 + yklee 결정) + lint L11 의 D-73 작업 (선택) |
| R-d-72-S-10 | `scratch/devhub_wiki_integration_response/RESPONSE.md` (D-72 응답) mirror | 불필요 + D-72 응답은 my_harness 측 작업 — vault 비대화 | exclude list 정합 (`scratch/` 패턴) |
| **R-d-72-S-11 (Phase 1.5 신규)** | **위키 본문이 raw/ 와 1:1 mirror 가 아닌 stub (L10 면제만으로 wiki 정상 처리)** | **위키만으로 코드 maintenance 시 detail 미확인** | **본 SOP §0 정공법 — 위키의 `sources/adr-0028-...md` page 의 본문 = raw/ 의 `docs/adr/0028-...md` 의 1:1 byte-identical mirror. frontmatter 만 추가, 본문은 raw/ 본문 그대로.** |
| **R-d-72-S-12 (Phase 1.5 신규)** | **위키의 source code link 가 raw/ 에 부재 (L02 위반)** | **위키의 link 깨짐** | **mirror-list.md §1.7.1 + §1.7.2 의 화이트리스트가 raw/ 의 source code link 와 1:1 정합. 본 갱신 시점 12 backend + 8 frontend file 의 raw/ mirror 정상 → 위키의 L02 link PASS.** |
| **R-d-72-S-13 (Phase 1.5 신규)** | **`//` (double slash) file path (release/v0.5.7//backlog/...) 본 세션 발견 (2026-06-13)** | **cp 자체는 작동하나 wikilink 검증 시 path error 가능** | **follow-up: branch_dir 의 trailing slash 처리 fix. 본 세션은 mirror 정상 작동, follow-up 정공법.** |

## 6. Mirror + Lint 의 verification

### 6.1 Phase 1+1.5 mirror 후 verification (T-d-72-2 + 본 갱신)

```bash
# 1. mirror 실행 (real)
bash scripts/wiki-sync-devhub.sh

# 2. mirror 결과 검증
ls -la ~/wiki/raw/projects/devhub/
cat ~/wiki/raw/projects/devhub/_manifest.md
find ~/wiki/raw/projects/devhub -type f | wc -l  # 기대: ~840

# 3. 위키 1:1 mirror byte-identical 검증 (Phase 1.5 추가)
python3 << 'EOF'
# 본 SOP §0 의 verification script 실행
EOF

# 4. lint 실행 (D-73 옵션 활성 후)
python3 ~/repos/my_harness/ai-workflow/skills/wiki-lint/scripts/run_wiki_lint.py \
  --vault-path ~/wiki \
  --project devhub \
  --project-config ~/repos/Devhub_example_minimax/docs/llm-wiki/lint-config.toml
```

### 6.2 dry-run (CI 검증 가능)

```bash
# dry-run (실제 mirror X, source list 만 출력)
bash scripts/wiki-sync-devhub.sh --dry-run
# 기대 출력:
# - ~840 file 의 source list 출력
# - 0 actual mirror
# - exit 0 (success)
```

**CI integration 가능** (D-72 forward path):
- `ci.yml` 의 workflow-lint job 의 env 에 `bash scripts/wiki-sync-devhub.sh --dry-run` step 추가
- PR 의 source list drift 자동 검증 (mirror-list.md 변경 시 CI fail)

## 7. 알려둘 trade-off (의도적 결정, Phase 1.5 갱신)

- **Phase 1+1.5 의 lint = "사용자 수동 lint (선택)"**: wiki-lint 의 `--project` + `--project-config` 옵션은 my_harness 측 D-73 의 작업. 본 PR 은 lint-config.toml 의 source 만 제공. **옵션 추가 후 자동 활성**.
- **CI hook (push 시 자동 mirror) 미사용**: D-72 응답 §5 D-71.3 의 consumer 원칙 — wiki 는 derived, mirror 는 명시적 trigger. 자동 mirror 는 raw/ 의 SSOT 위배 위험.
- **mirror list 의 core subset ~140 file (Phase 1+1.5)**: D-72 응답 §4 #3 의 "100~200 파일" 정합. 본 세션 scope = Phase 1+1.5 = docs + source code maintenance critical + branch memory. **Phase 3 의 domain (66) + architecture (1) + infrastructure + validation (~100 file) = 별도 PR**.
- **mirror 실행 = 본 PR scope 외**: 본 PR 의 lint 검증은 `bash scripts/wiki-sync-devhub.sh --dry-run` 의 source list 정합만. 실제 mirror 는 사용자 (yklee) 가 Phase 1+1.5 mirror 실행 시점에 진행. **이유**: mirror 실행은 `~/wiki/` 의 out-of-repo 변경 — 본 PR 의 in-repo 검증 가능 영역 (CI 4/4) 의 검증 범위 외.
- **scratch/ exclude**: `scratch/devhub_wiki_integration_response/RESPONSE.md` (D-72 응답) 는 본 Phase 1 의 reference. **mirror 미포함** (vault 비대화 + D-72 응답은 my_harness 측 작업).
- **`backend-core/` / `frontend/` / `backend-ai/` 의 source code bulk mirror 제외**: vault 비대화 + LLM agent 의 코드 정합은 **Phase 1.5 의 maintenance critical subset (~28 file)** 만 mirror. **forward: 새 backend file (e.g. 새 domain) / 새 frontend page 추가 시 PR 본문에 mirror scope 추가 요청**.
- **`openapi.yaml` 의 사내 한정 정보 mirror 허용**: D-72 응답 §3 + yklee 결정 (wiki 는 Gitea private 만) 으로 `sa-internal/` 격리 불요. 단 lint L11 (사내 패턴 검출) 의 D-73 작업은 권장 (선택).
- **위키의 본문 1:1 mirror 정공법 (Phase 1.5 핵심)**: 위키의 `sources/adr-0028-...md` page 의 본문 = raw/ 의 `docs/adr/0028-...md` 의 1:1 byte-identical mirror. **frontmatter 만 추가, 본문은 raw/ 본문 그대로**. 본 정공법으로 L10 (raw/ source 0) 면제 + L02 (broken wiki link) PASS + 위키만으로 코드 maintenance 가능.
- **본 저장소 한정 (Phase 1.5 정공법)**: my_harness 측 wiki 일임 결정 (session §10) 해제 후 본 저장소 측에서 진행. my_harness 측 D-73 (wiki-lint 옵션 추가) 의 결과는 본 정공법과 독립.

## 8. 다음 세션 directive

1. **Phase 1.5 의 mirror 실행** (T-d-72-2 + 본 갱신): `bash scripts/wiki-sync-devhub.sh` 1회 실행 → `~/wiki/raw/projects/devhub/` 에 ~840 file / ~6.6M mirror + `_manifest.md`. **본 sprint 의 22 file + maintenance critical ~28 file + branch memory ~700+ file = 위키만으로 코드 maintenance 가능 정공법**.
2. **wiki-source-sync + wiki-event-sync 호출** (raw → wiki 갱신): 본 sprint 의 4 PR + memory finalize + 3 ADR 갱신 + Phase 1.5 scope 의 file 변경분.
3. **lint 검증** (wiki-lint 의 L01~L10, D-73 옵션 활성 후): 사용자 수동 lint (선택) 또는 D-73 옵션 추가 시 자동.
4. **vault Gitea private push** (yklee 수동): `~/wiki/` 의 변경분을 my_harness 측 Gitea private 으로 push.
5. **commit + push + PR 발행** (Phase 1.5 scope): 본 sprint scope = `docs/llm-wiki/mirror-list.md` (Phase 1.5 추가) + `docs/llm-wiki/lint-config.toml` (Phase 1.5 갱신) + `docs/llm-wiki/operation-sop.md` (Phase 1.5 갱신) + `scripts/wiki-sync-devhub.sh` (Phase 1.5 sync script).
6. **main flat memory finalize** (post-merge sync): state.json + session_handoff.md + work_backlog.md 의 mirror 갱신.
7. 또는 **Phase 3 (mass ingest) trigger** (forward path): domain (66) + architecture (1) + infrastructure + validation (~100 file). 본 저장소 측의 별도 PR. **위키 1:1 mirror 정공법 + maintenance critical subset 의 mirror scope 확장**.

## 9. 향후 작업 지침 (위키 1:1 mirror 정공법 영구화, 2026-06-13 추가)

**위키만으로 코드 maintenance 가능** 하도록 하기 위해, **모든 신규 PR**은 다음을 충족해야 한다:

1. **소스코드 변경 PR** (`backend-core/`, `frontend/`, `.github/workflows/`, `scripts/`, `docs/traceability/`, `ai-workflow/memory/`):
   - 본 저장소 측 PR 본문 + branch memory 의 `pr_body.md` 에 다음 명시:
     - 변경 file path
     - 변경 line / block summary (1-2 line)
     - cross-reference (이전 PR / ADR / ID)
   - **mirror scope 가 확장될 가능성 있는 신규 file** (예: backend 새 도메인 / frontend 새 e2e spec / workflow 신규) 추가 시, PR 본문 + pr_body.md 에 다음 명시:
     - 신규 file path
     - **wiki mirror scope 추가 요청** (mirror-list.md §1.7 갱신 + lint-config.toml 갱신 + scripts/wiki-sync-devhub.sh 의 화이트리스트 갱신)
2. **PR 머지 후** 위키 mirror 갱신:
   - `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real mode) → `~/wiki/raw/projects/devhub/` 의 1:1 mirror 갱신
   - `python3 ~/wiki/skills/wiki-source-sync/scripts/wiki-source-sync.py --project=devhub --scope=all --auto-fix` (raw → wiki)
   - `python3 ~/wiki/skills/wiki-event-sync/scripts/wiki-event-sync.py --op=commit --project=devhub --ref=<merge-commit-sha> --intent="..."` (wiki event sync)
3. **위키의 1:1 mirror 정공법 자동 검증**:
   - 본 SOP §0 의 verification script 실행 (Total: ~140, Diff: 0, Identical: ~140)
   - 실패 시 즉시 fix (위키 갱신 + lint-config.toml 갱신)
4. **AGENTS.md 정합** (이미 `## 문서 tier 라벨` 섹션 + 본 §9 정공법 정합): 모든 신규 sprint 진입 시 본 §9 의 지침 우선 확인.
5. **branch memory mirror 정공법** (Phase 1.5 §1.7.4): active branch memory 의 state.json / session_handoff.md / work_backlog.md / backlog/*.md / pr_body.md 의 5 file 모두 mirror. **PR 머지 후 30일 이내 CLOSED branch = mirror 유지**, **30일 후 = `mavis-trash` 권장**.

**본 §9 의 정공법으로 위키 1:1 mirror 영구화**. 위키 만으로 코드 maintenance 가능 보장.
