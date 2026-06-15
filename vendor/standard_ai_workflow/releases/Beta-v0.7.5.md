# Beta v0.7.5 — Extension Audit + Wiki 운영 자동화 (2026-06-14)

> v0.7.4 의 CLI wrapper + decorator 의 *운영 layer* 보강.
> 4 sub-cat dispatcher runtime test 보강 (Extension audit) + wiki raw mirror / L2 dense 자동화 tool 정식화.

## 핵심 추가 (2 follow-up)

### 1. Extension audit (4 신규 runtime test)

v0.7.3 의 4 sub-cat dispatcher (security-auth / testing-property-based / performance-memory / resiliency) 의 *runtime test* 가 *1건* (test_evaluate_all_baselines) 만 추가됐던 갭 해소. v0.7.5 에서 4 dispatcher 의 rule_id 정합 + expected count 검증 test 4종 추가.

```python
# tests/check_baselines_compliance.py (12 → 16 test)
def test_evaluate_security_auth_baseline() -> None:
    """evaluate_compliance(baseline='security-auth') → 6 SEC-AUTH-01~06."""
    # rule_id set 검증

def test_evaluate_testing_pbt_baseline() -> None:
    """evaluate_compliance(baseline='testing-property-based') → 6 PBT-WF-01~06."""

def test_evaluate_performance_memory_baseline() -> None:
    """evaluate_compliance(baseline='performance-memory') → 6 PERF-MEM-01~06."""

def test_evaluate_resiliency_baseline() -> None:
    """evaluate_compliance(baseline='resiliency') → 8 RES-WF-01~08."""
```

**dispatcher 정합 결과**:
- security-auth: SEC-AUTH-01~06 (6 rule, keyring chmod 600 / token rotation / OAuth scope / 2FA / entropy / audit log)
- testing-property-based: PBT-WF-01~06 (6 rule, test_prop_ / round_trip / invariant / idempotency / generator / shrink)
- performance-memory: PERF-MEM-01~06 (6 rule, tracemalloc+RSS / leak / GC pause / ref cycle / cProfile / regression)
- resiliency: RES-WF-01~08 (8 rule, doctor / structured log / metrics / run_id / 5-tuple error / SIGTERM / max iter / health snapshot)

**trigger**: v0.7.3 commit `d03348a` 의 "12/12 baseline compliance test" 가 4 helper × 3 helper_test 만 카운트. dispatcher 자체의 *어떤 rule 이 emit* 되는지 미검증 상태였음. 본 patch 로 full loop 완성 (v0.7.1-roadmap §1 sub-cat 도입 → v0.7.2 4종 본 구현 → v0.7.3 dispatcher → v0.7.5 본 patch).

### 2. Wiki 운영 자동화 (refresh_wiki_memory.py 정식화)

v0.7.0~v0.7.4 의 5 release / 35+ commit 의 raw mirror 갭을 *release 별* 자동 갱신하는 CLI tool. 2026-06-14 session 에서 1회용 helper (`/tmp/refresh_saw_wiki.py`) 로 검증한 로직을 정식 tool 로 승격.

```bash
# dry-run: 어떤 file 이 어떻게 갱신될지 미리 보기
python3 tools/refresh_wiki_memory.py --refresh-raw --dry-run --json

# raw mirror 4 file 갱신 (state.json / work_backlog.md / wiki/log.md / memory/log.md)
python3 tools/refresh_wiki_memory.py --refresh-raw --apply

# vault L2 stub 4 file dense 재emit
python3 tools/refresh_wiki_memory.py --emit-l2 --apply

# 특정 release 만 (e.g. v0.7.4 후속 patch 시)
python3 tools/refresh_wiki_memory.py --refresh-raw --apply --since=2026-06-14
```

**Tool 구조 (433 line)**:
- argparse 5 flag: `--refresh-raw` / `--emit-l2` / `--since` / `--dry-run` / `--json`
- 1차 출처 `REPO_ROOT = ~/repos/standard_ai_workflow_minimax` (git)
- 2차 출처 `VAULT_ROOT = ~/wiki` (raw mirror + L2 sources/)
- sub-routine: `collect_commits` (git log) → `categorize` (release 별 bucket) → `pick_feat_commit` (reversed feat) → 4 `update_*` (raw mirror) → `reemit_l2_stubs` (L2 dense)
- 분리 정책: `--refresh-raw` 는 raw mirror, `--emit-l2` 는 vault L2 (R-3 단계 분리, `emit_wiki_l2_body.py` 와의 cross-ref)

**Smoke test (10 test, 246 line)**:
- `test_cli_subcommand_parsing` / `test_cli_no_subcommand_errors`: argparse 검증
- `test_collect_and_categorize`: 13 release bucket (v0.6.1~v0.7.4 + unreleased)
- `test_pick_feat_commit_priority`: reversed feat 우선 (v0.7.0 step 1 매칭)
- `test_pick_feat_commit_fallback`: feat 부재 시 첫 commit
- `test_update_state_json_dry_returns_8_lines`: v0.6.4~v0.7.4 8 entry format
- `test_update_work_backlog_dry_returns_5_blocks`: 5 release anchor
- `test_update_wiki_log_dry_returns_5_entries`: 5 release entry + head/commits/range
- `test_update_memory_log_dry_returns_1_entry`: sync backfill
- `test_reemit_l2_stubs_dry_returns_4_non_empty`: 4 stub 모두 100+ bytes

## 검증

- 신규 test: 4 (sub-cat dispatcher) + 10 (refresh_wiki_memory) = **14 신규 test**
- 회귀 test: 0 (16 + 12 + 13 + 7 + 13 = 61+ PASS 유지)
- **누적 75+ test PASS** (v0.7.4 200+ + 14 신규)

## Commit

| Hash | Subject |
|---|---|
| `9e1f206` | test(v0.7.5): 4 sub-cat dispatcher runtime test 보강 (12 → 16) |
| `0741775` | feat(v0.7.5): refresh_wiki_memory tool 정식화 + 10 smoke test (Wiki 운영 자동화) |
| `c2a75f8` | chore(v0.7.5): version bump 0.7.4 → 0.7.5 + release note |

## 다음 (v0.7.6 / v0.8.0 후보)

- **Release pipeline 정식화** — `workflow doctor` 의 release validator hook + PyPI 자동 publish + GH release note 자동 generate
- **Extension 2차 확장** — observability / docs-quality / dependency-audit sub-cat 추가 (v0.7.1-roadmap §1 follow-up)
- **Wiki 운영 cross-link** — `emit_wiki_l2_body.py` 와 `refresh_wiki_memory.py` 의 1-command 통합 (`scripts/release_post.sh`)

## Reference

- [v0.7.4 release note](Beta-v0.7.4.md) (직전)
- [v0.7.1-roadmap.md](../extensions/v0.7.1-roadmap.md) §1 sub-cat 도입
- [tools/score_wiki_trend.py](../tools/score_wiki_trend.py) (v0.7.1, commit 별 score tracking — cross-ref)
- [tools/emit_wiki_l2_body.py](../tools/emit_wiki_l2_body.py) (L2 sources/ 본문 emit — 다음 step)
- [tests/check_baselines_compliance.py](../tests/check_baselines_compliance.py) (12 → 16 test)
- [tests/check_refresh_wiki_memory.py](../tests/check_refresh_wiki_memory.py) (10 test, 본 release)
