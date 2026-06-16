# Branch Cleanup SOP

- 문서 목적: 원격/로컬 branch 일괄 정리 (cleanup) 작업의 분류 기준, 백업 절차, 실행 명령, 안전장치를 단일 SOP 로 정의한다.
- 범위: `origin` 원격 + 로컬 `refs/heads/` 의 모든 branch 중 작업 종료된 것을 일괄 정리하는 작업. 신규 branch 작업 진입 시점의 hygiene PR 도 포함.
- 대상 독자: 모든 contributor (사람 + AI agent), sprint owner.
- 상태: accepted
- 최종 수정일: 2026-06-16
- 관련 문서: [`./worker_division.md`](./worker_division.md) §6 (사외/사내 2-tier), [`./document-standards.md`](./document-standards.md) (메타 헤더 형식), [`../traceability/sync-checklist.md`](../traceability/sync-checklist.md) (cleanup 후 PR 의 추적성 갱신).

## 1. 동기

저장소의 branch 수는 시간이 지남에 따라 누적된다. 다음을 방지하기 위해 주기적 정리가 필요하다.

- 원격 `origin` 의 stale ref 누적 (`git branch -r` 가 100+ line).
- 로컬 `refs/heads/` 의 작업 종료된 branch 누적 (`git for-each-ref refs/heads/` 40+ line).
- `git fetch --prune` 만으로는 머지/closed PR 의 origin ref 가 정리되지 않음.
- 로컬에서 `--delete --remote` (push) + `git branch -D` 가 일관성 없이 부분 적용되어 origin/local 비대칭 발생.

본 SOP 는 위의 비대칭 해소 + 안전한 일괄 처리 절차를 단일화한다.

## 2. 분류 기준 (3-tier)

정리 대상 branch 를 다음 3-tier 로 분류한다. 각 tier 의 안전도와 권장 조치는 §3 참조.

| Tier | 정의 | 판별 방법 |
| --- | --- | --- |
| **A. main-ancestor** | `git merge-base --is-ancestor <branch> origin/main` = true | main branch tip 의 ancestor = 본 branch 의 commit 이 main 에 이미 포함됨 |
| **B. PR merged** | PR 의 `state = MERGED` (gh API) | `gh pr list --state all --json headRefName,mergedAt` 결과 |
| **C. closed (not merged)** | PR 의 `state = CLOSED` 이고 `mergedAt = null` | PR 닫혔으나 머지 안 됨 (작업 중단 결정) |
| **D. open PR** | PR 의 `state = OPEN` | **정리 금지** (작업 진행 중) |
| **E. no PR** | `gh pr list` 결과에 `<branch>` 가 없음 | A/B/C 모두 아닐 때, **사전 확인 필요** (E.1 / E.2 결정) |

### 2.1 E 의 추가 분류

| Sub-tier | 정의 | 권장 |
| --- | --- | --- |
| **E.1 stale (remote 존재, 30일+ 묵음)** | origin ref 존재 + 마지막 commit > 30일 + PR 없음 | 사용자 결정 — 일반적으로 safe delete |
| **E.2 truly orphaned (no PR + no remote)** | origin ref 없음 + PR 없음 + 로컬에만 존재 | 사용자 결정 — 의도적 backup 가능성 (예: `backup/...` prefix) |

## 3. 조치 매트릭스

| Tier | 안전도 | 원격 (`origin`) | 로컬 (`refs/heads/`) |
| --- | --- | --- | --- |
| A. main-ancestor | ✅ safe delete | `git push origin --delete` | `git branch -D` |
| B. PR merged | ✅ safe delete | `git push origin --delete` | `git branch -D` |
| C. closed (not merged) | ✅ safe delete (작업 중단 결정 반영) | `git push origin --delete` | `git branch -D` |
| D. open PR | ❌ **금지** | 건드리지 않음 | 건드리지 않음 |
| E.1 stale | ⚠ 사용자 확인 | 사용자 결정 후 delete | 사용자 결정 후 delete |
| E.2 orphaned | ⚠ 사용자 확인 (보존 의도 가능) | — (없음) | 사용자 결정 후 delete (또는 보존) |

## 4. 안전장치

### 4.1 Backup 정책

정리 **반드시** 사전에 backup 을 dump 한다.

#### 4.1.1 원격 branch backup

`git push origin --delete` 직전에 각 branch 의 (sha, author date, subject) 를 dump.

저장 위치: `.git/branch-backup/<UTC-timestamp>/branches.tsv`

형식: `<branch>\t<commit_sha>\t<YYYY-MM-DD>\t<subject>`

복구: `bash .git/branch-backup/<UTC-timestamp>/restore.sh` (대화형, `<sha>:refs/heads/<branch>` push)

#### 4.1.2 로컬 branch backup

`git branch -D` 직전에 동일 형식으로 dump. 단, 로컬 backup 은 origin 이 없는 branch (E.2 truly orphaned) 의 SHA 보존이 주목적.

저장 위치: `.git/branch-backup/<UTC-timestamp>-local/branches.tsv`

복구: `bash .git/branch-backup/<UTC-timestamp>-local/restore.sh` (대화형, `git branch <branch> <sha>`)

### 4.2 Prerequisite check

`git push origin --delete` / `git branch -D` 실행 **전** 다음을 확인:

1. `git status` 가 clean (uncommitted change 없음).
2. `git rev-parse HEAD` = `git rev-parse origin/main` (또는 명시적 ff/merge 완료).
3. `gh auth status` 가 authenticated.
4. 정리 대상 list 가 `wc -l` 으로 0 보다 큼 + dry-run 출력으로 사용자가 1회 확인.

### 4.3 Dry-run

`git push --delete` 는 `--dry-run` 을 지원하지 않음. 다음 절차로 dry-run 효과:

1. 분류 script 가 만든 `<list>.txt` 를 사용자에게 출력.
2. `git log --oneline -n 3 origin/<branch>` 로 tip 의 마지막 3 commit + author 출력.
3. 사용자가 1회 OK 한 후에만 실제 delete 실행.

## 5. 실행 절차

### 5.1 원격 branch 일괄 정리

```bash
# 1. Prerequisite
git fetch --all --prune
gh auth status

# 2. Build classification TSV
#    A: main ancestor
#    B: PR merged
#    C: closed (not merged)
#    D: open PR  (제외)
#    E: no PR   (사전 확인 후 결정)
git for-each-ref --format='%(refname:short)' refs/remotes/origin/ \
  | grep -v '^origin/HEAD$' > /tmp/origin-branches.txt
# ... (분류 script — `gh pr list --state all --json ...` 와 merge-base 결합)

# 3. Backup dump
TS=$(date -u +%Y%m%dT%H%M%SZ)
mkdir -p ".git/branch-backup/$TS"
awk '{...}' > ".git/branch-backup/$TS/branches.tsv"

# 4. Show user
echo "Delete list:"; cat /tmp/list-final-delete.txt

# 5. Execute
xargs git push origin --delete < /tmp/list-final-delete.txt

# 6. Prune tracking refs
git remote prune origin
```

### 5.2 로컬 branch 일괄 정리

```bash
# 1. Prerequisite: same as 5.1
# 2. Build classification — 동일 절차, source = refs/heads/ (main 제외)
# 3. Backup dump — .git/branch-backup/<TS>-local/
# 4. Show user
# 5. Execute
xargs git branch -D < /tmp/list-local-delete.txt
```

### 5.3 검증

```bash
# working tree clean
git status --short --branch

# main 정합
git rev-list --left-right --count HEAD...origin/main   # 0 0

# branch count
git for-each-ref --format='%(refname:short)' refs/heads/ | wc -l
git for-each-ref --format='%(refname:short)' refs/remotes/origin/ \
  | grep -v '^origin/HEAD$' | wc -l
```

## 6. gitignore 처리

`.git/branch-backup/` 는 `.git/` 내부 디렉터리이므로 git 의 자체 store 영역에 해당 — **자동으로 추적 불가**. 별도 `.gitignore` 항목 불필요.

(검증: `git ls-files | grep branch-backup` → 빈 결과.)

## 7. Cleanup 의 sprint 통합

본 SOP 의 일괄 정리는 다음 시점 중 하나로 trigger 한다.

1. **Hygiene sprint 시작 시점** — `docs/planning/release_v1_roadmap.md` 의 P3 잔여 chunk 에 포함.
2. **v0.x → v0.y major-minor 전환 직전** — 이전 milestone 의 branch 일괄 정리.
3. **저장소 branch count 가 100+ 누적 시점** — 자동 trigger (linter 또는 maintainer 판단).

## 8. 한계 + 알려진 함정

- `git push --delete --dry-run` 미지원 → §4.3 절차로 dry-run 효과.
- `gh pr list --limit` 의 기본값 (30) → 본 SOP 의 일괄 처리 시 `--limit 1500` 권장 (현재 PR 574 개).
- 5xx+ PR 의 경우 `--limit` 을 더 올리거나 graphql API 로 paginate 필요.
- backup restore.sh 는 `origin` remote 의 write 권한 필요 (PR §0 "사외/사내 2-tier" 에서 사외 repo 의 push 만 가능).
- backup 디렉터리는 host-local 이므로 **다른 머신으로 이관 시 별도 동기화** 필요 (rsync 또는 tar).

## 9. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-16 | 1차 작성. 원격 78 + 로컬 41 일괄 cleanup (2026-06-16 sprint hygiene) 의 SOP 표준화. |
