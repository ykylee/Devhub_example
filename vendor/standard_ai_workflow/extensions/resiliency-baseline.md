# Resiliency Baseline Extension (v0.7.2, 4종)

- 문서 목적: standard_ai_workflow v0.7.2 의 resiliency-baseline extension (4종 중 1). AIDLC 의 16 rule 중 *우리 운영에 applicable 한 8 rule* 적응.
- 범위: 8 RES-WF rule (RES-WF-01~08) + workflow_kit.common.resiliency helper + 8 smoke test
- 상태: stable (v0.7.2 도입)
- 최종 수정일: 2026-06-13
- 관련 문서: [`../SCHEMA.md`](../SCHEMA.md) (extension system SSOT)
- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/resiliency/baseline/resiliency-baseline.md` (490 line, 16 rule, commit `b19c819`, 2026-06-08)

## §1 왜 Resiliency Baseline 이 필요한가

AIDLC 의 resiliency-baseline 16 rule 중 8 rule 은 **cloud workload** (HA / DR / Incident Response / Multi-region) — 우리 local runtime 에 N/A. 나머지 8 rule 은 *cross-cutting* 으로 우리 workflow 에 applicable.

본 1차 출시: **resiliency-baseline 1종 (4종 중 1)**, 8 rule 의 *우리 적응*.

## §2 8 Rule 정의 (우리 적응)

### 2.1 Rule RES-WF-01: Critical Workload Identification

**Rule**: workflow_kit 의 *critical component* 식별. ai-workflow core (state.json, audit.md, handoff.md) 가 critical.

**Verification**:
- `workflow_kit.common.resiliency.identify_critical_components()` 자동 식별
- critical list 가 `ai-workflow/memory/active/critical_components.json` 에 기록
- smoke test: `test_critical_workload_identification`

### 2.2 Rule RES-WF-02: Change Management

**Rule**: 모든 변경 (commit / PR / state doc edit) 이 *추적 가능*. git + PR review + wiki R-1~R9.

**Verification**:
- 모든 commit 이 PR merge 또는 직접 push (둘 다 추적)
- PR 의 *required review* ≥ 1
- state doc 변경 시 commit hash 명시 (wiki log commit hash 정책, yklee 2026-06-13)
- smoke test: `test_change_management_tracking`

### 2.3 Rule RES-WF-03: Audit Log Observability

**Rule**: audit log 가 *observability dashboard* (Grafana / Obsidian Canvas) 에 자동 export. v0.7.0 step 10 + 위키 vault 자동 propagate.

**Verification**:
- `audit.md` 가 vault 의 `~/wiki/log.md` 에 자동 propagate (wiki-event-sync)
- dashboard 의 Compliance Summary 가 audit log 의 events 기반 자동 emit
- smoke test: `test_audit_log_observability`

### 2.4 Rule RES-WF-04: State.json Observability

**Rule**: state.json 의 *변경* 이 *observability metric* 으로 emit. commit 별 state.json diff.

**Verification**:
- `state.json` 의 last_modified 가 commit hash 명시
- state.json diff 가 `state_history.jsonl` 에 누적
- smoke test: `test_state_json_observability`

### 2.5 Rule RES-WF-05: workflow_kit Health Check

**Rule**: `workflow_kit.__version__` + smoke test 6종 PASS rate 가 *health metric*. 5/5 미만 시 health=critical.

**Verification**:
- `tools/health_check.py` 가 `__version__` + smoke test 결과 emit
- 5/5 PASS = healthy / 3-4/5 = degraded / ≤ 2/5 = critical
- smoke test: `test_workflow_kit_health_check`

### 2.6 Rule RES-WF-06: Git Remote + Vault Backup

**Rule**: 모든 변경이 *2 곳* 에 propagate. GitHub (primary) + Gitea vault (backup).

**Verification**:
- GitHub commit push 자동 (workflow_kit 가 `git push` 호출)
- Gitea vault 자동 sync (wiki-event-sync)
- 2 곳 push 실패 시 alert
- smoke test: `test_dual_remote_backup`

### 2.7 Rule RES-WF-07: Backlog Recovery

**Rule**: 작업 손실 시 `mavis-trash` + `git restore` + `state.json` recovery 절차.

**Verification**:
- `mavis-trash <path>` 가 OS Trash 로 이동 (recoverable)
- `git restore <path>` 가 commit 복원
- `state.json` 의 last_modified 기반 recovery
- recovery 절차 문서화 (`docs/RECOVERY.md`)
- smoke test: `test_backlog_recovery_procedure`

### 2.8 Rule RES-WF-08: Session Handoff Recovery

**Rule**: `session_handoff.md` 가 *cross-session continuity* 보장. session 중단 시 다음 session 이 handoff 로 즉시 resume.

**Verification**:
- session 종료 시 handoff 자동 emit (state + backlog + next_documents)
- 다음 session 의 session-start 가 handoff 읽음
- smoke test: `test_session_handoff_recovery`

## §3 Compliance Summary

| Rule ID | Title | Status | Notes |
|---|---|---|---|
| RES-WF-01 | Critical Workload Identification | ✅ | workflow_kit core 자동 식별 |
| RES-WF-02 | Change Management | ✅ | git + PR + wiki log |
| RES-WF-03 | Audit Log Observability | ✅ | vault 자동 propagate |
| RES-WF-04 | State.json Observability | ✅ | state_history.jsonl |
| RES-WF-05 | workflow_kit Health Check | ✅ | 6 smoke test PASS rate |
| RES-WF-06 | Git Remote + Vault Backup | ✅ | GitHub + Gitea |
| RES-WF-07 | Backlog Recovery | ✅ | mavis-trash + git restore |
| RES-WF-08 | Session Handoff Recovery | ✅ | cross-session continuity |

## §4 8/16 N/A Rule (AIDLC 의 cloud-specific)

| Rule | N/A 이유 |
|---|---|
| RES-09 (RTO/RPO targets) | cloud workload 한정 |
| RES-10 (Multi-region topology) | cloud workload 한정 |
| RES-11 (Active-Active) | cloud workload 한정 |
| RES-12 (Backup frequency SLA) | git 자동화 충분 |
| RES-13 (CI/CD pipeline) | GitHub Actions / Gitea Actions |
| RES-14 (Resiliency testing) | smoke test 가 부분 커버리지 |
| RES-15 (Incident response) | v0.7.3+ ADR 후보 |
| RES-16 (Postmortem) | wiki log 가 유사 역할 |

## §5 우리 runtime helper (workflow_kit.common.resiliency)

```python
def identify_critical_components() -> list[str]:
    """workflow_kit core component 자동 식별."""

def health_check() -> dict:
    """__version__ + smoke test 6종 PASS rate 기반 health metric."""

def state_history_diff(commit_a: str, commit_b: str) -> dict:
    """두 commit 의 state.json diff."""
```

## §6 한계 / 예외

- **AIDLC 의 6 pillar (Business Goals, Change Mgmt, Observability, HA, DR, Continuous Improvement)** 중 *HA, DR, Incident* N/A — 우리 local runtime
- **Multi-cloud backup**: 우리 storage 가 GitHub + Gitea 만. *3rd cloud* backup 은 v0.7.3+ ADR

## §7 Follow-up (v0.7.3+)

- v0.7.3: Incident response process (RES-15) — 우리 운영에 맞게 *경량 SOP*
- v0.7.3: Postmortem template (RES-16) — wiki log 와 통합
- v0.7.3: workflow_kit.common.resiliency runtime helper 본 구현

## §8 References

- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/resiliency/baseline/resiliency-baseline.md` (490 line, 16 rule, commit `b19c819`)
- 우리 SSOT: `extensions/SCHEMA.md` (v0.7.0 step 7, 200 line)
- 우리 wiki: `ai-workflow/wiki/concepts/extension-system.md` (210 line)
- 우리 검증: `tests/check_extension_system.py` (23 test PASS, v0.7.0)
- 우리 검증 (본 1차 출시): `tests/check_resiliency_baseline.py` (8 test PASS, v0.7.2)
- v0.7.1+ roadmap: `extensions/v0.7.1-roadmap.md` (8/16 rule 적응 결정)
