# ai-workflow/wiki/raw/ Mirror Manifest (in-repo SSOT, 2026-06-15)

> 본 file 은 v0.7.17 in-repo wiki redirect 의 *raw mirror* 의 provenance. wiki 운영의 SSOT = `scripts/wiki-sync-devhub.sh` 의 mirror list + 본 file.
> raw mirror 자체 (`ai-workflow/wiki/raw/projects/devhub/`, 8.0M, 964 file) 는 `.gitignore` 으로 *ignore* — commit 부담 회피. 운영은 `bash scripts/wiki-sync-devhub.sh` 로 local regenerate.
> 후속 결정: wiki raw mirror 의 full commit vs ignore (SSoT out-of-raw vs in-raw) — 미정, follow-up PR.

## Mirror source

- **DEST**: `ai-workflow/wiki/raw/projects/devhub/` (in-repo, v0.7.17 redirect)
- **SRC**: 본 저장소 의 15 패턴 source (mirror list = `docs/llm-wiki/mirror-list.md` §2)
- **sync tool**: `scripts/wiki-sync-devhub.sh` (BSD-rsync safe, 15 패턴)
- **vault path 의 redirect**: `VAULT_ROOT="${SRC}/ai-workflow/wiki"` (이전 `${HOME}/wiki`)

## 본 file vs raw 의 _manifest.md

- **본 file** (`ai-workflow/wiki/RAW_MIRROR_MANIFEST.md`, raw/ 바깥) = *운영 가이드 + provenance*. 항상 commit.
- **raw 의 _manifest.md** (`ai-workflow/wiki/raw/projects/devhub/_manifest.md`) = `scripts/wiki-sync-devhub.sh` 의 자동 생성. *각 mirror 실행* 의 git state + version + timestamp + size. ignore. local regenerate.

## v0.7.17 raw mirror (2026-06-15, 1차)

- **mirror count**: 964 file
- **size**: 8.0M
- **commit (source)**: `dd20266a` (PR #600 머지, feat(vendor): standard_ai_workflow v0.7.17 import + wiki in-repo redirect)
- **branch (mirror)**: `feat/wiki-mirror-bootstrap`
- **version**: system=v0.1.1-alpha, workflow=v0.5.11-beta
- **15 패턴 (mirror list scope)**:
  1. `docs/adr/0[0-9][0-9][0-9]-*.md` (ADR, ~31 file)
  2. `docs/governance/*.md` (~5 file)
  3. `docs/planning/*.md` (~27 file)
  4. `docs/setup/*.md` (~15 file)
  5. `docs/requirements.md`
  6. `docs/openapi.yaml`
  7. `ai-workflow/memory/{state.json, session_handoff.md, work_backlog.md}` (main flat 3 file)
  8. `.github/workflows/*.yml`
  9. `scripts/*.sh` (화이트리스트)
  10. `backend-core/internal/...` (maintenance critical Go, ~25 file)
  11. `frontend/tests/e2e/...` + `frontend/lib/...` + `frontend/domain/...` (e2e critical, ~14 file)
  12. `docs/traceability/{README,conventions,report,sync-checklist}.md` (4 file)
  13. `ai-workflow/memory/<agent>/<branch>/*` (active + 30일 이내 CLOSED branch memory)
  14. `docs/domain/**/*.md` (~66 file)
  15. `docs/{architecture,infrastructure,validation}/*.md` (~12 file)

## Mirror 운영

```bash
# Dry-run (어떤 file 이 mirror 대상인지)
bash scripts/wiki-sync-devhub.sh --dry-run

# Real mirror (DEST 의 clean 후 1:1 byte copy + manifest 자동)
bash scripts/wiki-sync-devhub.sh

# Incremental (DEST 의 기존 file 유지, --no-clean)
bash scripts/wiki-sync-devhub.sh --no-clean
```

## Follow-up (별도 PR)

- **wiki raw mirror 의 full commit 결정** (in-raw vs out-of-raw SSOT). 8M / 964 file — 1 commit 으로 가능하지만 PR 머지 부담 큼. git LFS / submodule / 별도 wiki-only branch 등 옵션 검토.
- vendor 의 emit_wiki_l2_body.py 의 *devhub project mode* adapter (L1 page 자동 emit, A안 5 page 의 220+ 확장)
- A안 5 page 의 운영 검증 후 전체 L1 + L2 page emit (v0.7.17 in-repo 의 wiki 가 *모든 운영의 1차 출처* 가 되도록)
- 7 script 의 my_harness wiki-* skill 호출의 in-repo vendor tool redirect
