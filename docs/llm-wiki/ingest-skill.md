# Wiki-Ingest Skill (본 저장소 측 사용법 가이드)

- 문서 목적: `scripts/wiki-ingest-from-raw.sh` 의 본 저장소 (= DevHub) 측 사용법을 정리한다. raw mirror + wiki page 자동 ingest flow 의 1-step entry point.
- 범위: 본 저장소 `~/repos/Devhub_example_omp/` 의 7 패턴 (mirror list §3) source → LLM Wiki vault (`~/wiki/projects/devhub/`) 의 wiki page 자동 작성.
- 대상 독자: DevHub repo 운영자, yklee (owner), my_harness 측 skill 개발자
- 상태: **draft** (D-72, 2026-06-11)
- 최종 수정일: 2026-06-11
- 관련 문서:
  - [./mirror-list.md](./mirror-list.md) (Phase 1 source 7 패턴, 82 file)
  - [./scope-and-rationale.md](./scope-and-rationale.md) (D-72 Q1~Q6 정공법)
  - [./operation-sop.md](./operation-sop.md) (sync + lint SOP)
  - `~/wiki/AGENTS.md` (vault 운영 규약 v1.5, §2.1 Ingest)
  - `~/repos/my_harness/ai-workflow/core/wiki_ingest_skill_spec.md` (my_harness 측 SSOT)
  - `~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/SKILL.md` (my_harness skill 정의)

## 1. 목적

기존의 `scripts/wiki-sync-devhub.sh` 는 **raw mirror 만** 자동화 (mirror list 의 7 패턴 source 를 `~/wiki/raw/projects/devhub/` 로 복사). 본 sprint 의 `scripts/wiki-ingest-from-raw.sh` 는 **raw → wiki page 자동 ingest** 까지 통합:

```
[본 저장소 ~/repos/Devhub_example_omp/]
    ↓ wiki-sync-devhub.sh (raw mirror)
[raw mirror ~/wiki/raw/projects/devhub/]
    ↓ wiki-ingest-from-raw skill (vault 의 wiki page 자동 작성)
[wiki page ~/wiki/projects/devhub/sources/<title>.md]
```

**핵심 정공법**: 사용자가 "raw 만 갱신되고 wiki 는 업데이트가 안됐다" 라고 말할 때, 본 sprint 의 wrapper 가 **2 단계를 1 명령**으로 통합.

## 2. 기대 입력

`scripts/wiki-ingest-from-raw.sh` 의 options:

- `--project <devhub|my-harness>` (필수) — 대상 project
- `--source <rel_path>` — 1 file ingest (raw/ 상대 경로). 미지정 시 `--all`
- `--limit N` — `--all` 시 최대 N건
- `--apply` — 실제 ingest (default = dry-run, 변경 없음)
- `--skip-lint` — post-ingest wiki-lint skip
- `--quiet` — stderr 메시지 최소화
- `-h, --help` — 도움말

## 3. 사용법

### 3.1 dry-run preview (변경 없음, 권장 첫 step)

```bash
bash scripts/wiki-ingest-from-raw.sh --project devhub
```

**출력**:
- step 1: raw mirror dry-run (mirror list 의 82 file)
- step 2: wiki ingest dry-run (각 source 의 frontmatter + body + cross-ref preview)
- JSON stdout (findings 83개, INGEST-01 info)
- Markdown report: `~/wiki/_lint/devhub/ingest_YYYY-MM-DD.md`

### 3.2 실제 ingest (--apply 명시)

```bash
bash scripts/wiki-ingest-from-raw.sh --project devhub --apply
```

**동작**:
- step 1: raw mirror apply (실제 mirror)
- step 2: wiki ingest apply (각 source 의 wiki page 작성 + cross-ref 갱신 + index/log append)
- post-ingest wiki-lint 자동 호출 (errors/warns findings warnings 에 추가)

### 3.3 부분 적용 (preview 의 N건만)

```bash
bash scripts/wiki-ingest-from-raw.sh --project devhub --limit 5 --apply
```

**용도**: CI / 빠른 검증 / 점진적 ingest.

### 3.4 1 file 만

```bash
bash scripts/wiki-ingest-from-raw.sh --project devhub --source docs/adr/0001-idp-selection.md --apply
```

**용도**: 1 file 의 preview/적용 (예: 신규 ADR 1건 발행 직후 ingest).

### 3.5 lint skip (CI 환경)

```bash
bash scripts/wiki-ingest-from-raw.sh --project devhub --apply --skip-lint
```

**용도**: lint 가 외부 의존성 (Keycloak 등) 없는 환경에서 실행 시.

## 4. 출력

### 4.1 stdout JSON

```json
{
  "status": "ok",
  "tool_version": "0.1.0",
  "vault_path": "/home/yklee/wiki",
  "project": "devhub",
  "mode": "dry-run",
  "examined_at": "2026-06-11T02:04:07Z",
  "summary": {
    "sources_total": 83,
    "sources_already_ingested": 0,
    "sources_to_ingest": 83,
    "sources_skipped": 0,
    "pages_to_create": 83,
    "pages_to_update": 0,
    "index_md_updates": 1,
    "log_md_appends": 83
  },
  "findings": [...],
  "errors": [],
  "warnings": []
}
```

### 4.2 vault `_lint/<project>/ingest_YYYY-MM-DD.md` report

- 검사 시각 / 검사자 / 결과 (sources to ingest / errors / warnings)
- ## Preview 표 (source_path → target_page → action → title)
- ## Warnings / ## Errors (해당 시)

## 5. 권한 경계

**wrapper** (`scripts/wiki-ingest-from-raw.sh`) 는 다음 권한을 가진다:
- 실행: `scripts/wiki-sync-devhub.sh` 호출
- 실행: `python3 ~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/scripts/run_wiki_ingest.py` 호출
- PYTHONPATH: 별도 설정 불요 (stdlib only)

**wrapper 가 호출하는** wiki-ingest skill (my_harness 측) 의 권한 경계는 `~/repos/my_harness/ai-workflow/core/wiki_ingest_skill_spec.md` §6 정합.

## 6. my_harness skill 부재 시 에러

`MYHARNESS_ROOT` 환경 변수로 my_harness repo 경로 명시 가능 (default: `~/repos/my_harness`).

```bash
# my_harness repo 가 다른 경로에 있을 시
MYHARNESS_ROOT=~/work/my_harness bash scripts/wiki-ingest-from-raw.sh --project devhub
```

**skill 부재 시**:
```
[wiki-ingest-from-raw] error: wiki-ingest-from-raw skill 미설치: <path>
[wiki-ingest-from-raw]   my_harness 측 SSOT: ~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/
```

## 7. 검증 (본 sprint 의 dry-run 결과)

| 항목 | 값 |
|---|---|
| Source 식별 | 83 file (mirror list 의 82 + raw `_manifest.md` 1) |
| Already ingested | 0 (vault 위의 wiki/projects/devhub/sources/ 가 비어있음) |
| To ingest | 83 (전체) |
| Skipped | 0 (0-byte file 없음) |
| Pages to create | 83 |
| Index.md updates | 1 |
| Log.md appends | 83 |
| Cross-ref 매칭 | 자동 (예: `adr-0002` → `[[rbac]]` 1건) |
| Errors | 0 |
| Warnings | 0 |

## 8. 운영 SOP (D-72 Phase 3 진입 시점)

### 8.1 정기 동기화 (1주 1회 권장)

```bash
# 1. dry-run preview
bash scripts/wiki-ingest-from-raw.sh --project devhub > /tmp/ingest-preview.json 2>&1
# 2. 사용자 검토 (preview 의 83 finding)
cat /tmp/ingest-preview.json | jq '.findings[] | select(.action == "create") | .target_page' | head -20
# 3. 검토 완료 후 apply
bash scripts/wiki-ingest-from-raw.sh --project devhub --apply
# 4. post-ingest wiki-lint 자동 실행
#    (errors 발견 시 ~/wiki/_lint/devhub/ingest_YYYY-MM-DD.md 의 warnings 에 추가됨)
# 5. Obsidian 에서 graph view 확인
# 6. (선택) wiki-sync-ai-workflow.sh + Obsidian Git plugin 으로 vault commit/push
```

### 8.2 신규 ADR 1건 발행 직후

```bash
# 1. dry-run preview
bash scripts/wiki-ingest-from-raw.sh --project devhub --source docs/adr/0032-new-decision.md
# 2. 검토
# 3. apply
bash scripts/wiki-ingest-from-raw.sh --project devhub --source docs/adr/0032-new-decision.md --apply
```

### 8.3 사고 대응 (긴급 wiki page 갱신)

```bash
# raw 의 1 file 만 강제 재-ingest (덮어쓰기)
bash scripts/wiki-ingest-from-raw.sh --project devhub --source docs/setup/keycloak_operations.md --apply
# 이미 ingest 된 page 는 skip (idempotent). 강제 덮어쓰기 v1.1 planned.
```

## 9. 후속 (D-72 Phase 3, planned)

- T-d-72-4: Phase 3 mass ingest (도메인 66 file + architecture + infrastructure + validation = ~100 file)
- T-d-72-5: cross-project 종합 페이지 (`wiki/cross/`)
- T-d-72-6: wiki-lint CI integration (D-74+ 별도 PR)
- v1.1: `--cross-ref-only` mode (기존 page 의 cross-ref 만 갱신)
- v1.1: LLM 호출을 통한 body 자동 요약 (현재는 raw 발췌)
- v1.1: 강제 덮어쓰기 `--force` flag
- v1.1: `wiki/cross/` cross-project 종합 자동 생성

## 10. 관련

- 본 저장소 mirror script: `scripts/wiki-sync-devhub.sh`
- my_harness skill SSOT: `~/repos/my_harness/ai-workflow/core/wiki_ingest_skill_spec.md`
- my_harness skill 구현: `~/repos/my_harness/ai-workflow/skills/wiki-ingest-from-raw/`
- vault 운영 규약: `~/wiki/AGENTS.md` (v1.5, D-71, §2.1 Ingest)
- lint SSOT: `~/wiki/schema/lint_rules.md`
- lint skill: `~/repos/my_harness/ai-workflow/skills/wiki-lint/`
- AGENTS.md (본 저장소): 사외/사내 2-tier 정책 + vault 공유 자원 정책
