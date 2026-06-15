# Session Handoff — fix/work_260612-6-e2e-internal-removal (2026-06-12, e2e-internal job 폐기 + ADR partial supersession)

- 문서 목적: 사용자 결정 (2026-06-12, "e2e-internal 은 사내 환경용 셋팅, GitHub Action 으로 체크 불요") 기반 e2e-internal job 완전 폐기 sprint 의 정공법 + ADR-0030/0031 partial supersession (baseline 변경 정공법) 상태 인계.
- 범위: `.github/workflows/ci.yml` (e2e-internal job 삭제, line 554-756) + `scripts/ci-e2e-sync-check.sh` (DEVHUB_BUILD_TIER 코멘트 정리) + `docs/adr/0031-build-tag-policy-review.md` (partial supersession, baseline 변경) + `docs/adr/0030-sso-integrations-and-auth-session-port.md` (§2.3 + §9 갱신) + `docs/traceability/report.md` (§6 row 추가).
- 상태: in_progress. main HEAD `379e894` (PR #577 E2E Internal disable + 후속 memory finalize).
- 최종 수정일: 2026-06-12 (sprint `fix/work_260612-6-e2e-internal-removal`)

## 0. 본 sprint 결정 사항 (사용자 2026-06-12)

### 사용자 결정 (verbatim)

> "e2e-internal은 어차피 사내 환경용 셋팅이면 github action으로 체크할건 아니야 그냥 없애줘."

### 결론

1. **e2e-internal job 자체 폐기** — `ci.yml` 에서 `e2e-internal:` job + 23 step + 202 lines 완전 삭제.
2. **ADR-0030 §2.3 의 runtime injection 결정 = 유지** (e2e-internal 폐기와 독립). 사내 staging/prod-smoke 가 real adapter 검증 책임.
3. **ADR-0031 partial supersession (baseline 변경 정공법)** — e2e-internal 폐기 후 정량 비교 가능. 결론 (옵션 2 runtime injection 유지) 변동 0건.

## 1. 변경 요약 (5 file)

### 1.1 `.github/workflows/ci.yml` (CI 설정)

- `e2e-internal:` job (line 554-756) 완전 삭제. ADR-0030 §2.3 + C-i reference 코멘트 (line 554-560) 포함.
- 9 jobs 잔여 (changed-paths, workflow-lint, migration-prefix-lint, openapi-yaml-lint, backend-unit, backend-integration, frontend-unit, e2e-build, e2e). e2e shard 1/2/3 (saovae_stub default) 만.
- file 756 → 551 lines (-205).
- YAML parse + workflow-lint PASS.

### 1.2 `scripts/ci-e2e-sync-check.sh` (sync contract)

- `DEVHUB_BUILD_TIER` 의도적 미포함 코멘트 (5 lines) 정리. e2e shard 1/2/3 env block 의 DEVHUB_BUILD_TIER 미설정 정합은 그대로 유지.
- 사내 한정 정보 변경 없음.

### 1.3 `docs/adr/0031-build-tag-policy-review.md` (ADR partial supersession)

**baseline 변경 정공법** 적용:
- §1 배경: "2026-06-12, e2e-internal job 폐기 결정" reference.
- §1.2 표: PR #542 row → "**2026-06-12 폐기** (사용자 결정)" 마킹.
- §2.1 표: CI matrix = **1 axis (shard only)** — e2e shard 1/2/3 saovae_stub default 만. CI runtime = e2e shard 1/2/3 (3 jobs × 4-5min) baseline. Keycloak container 1개.
- §2.2 표: CI matrix 2배 = **e2e shard 1/2/3 × 2 tags = 6 jobs**. CI runtime = **+15~30min** (6 vs 3 jobs). 
- §2.3 비교 표: CI matrix jobs **4 → 3**. trade-off 정합.
- §3.1 옵션 1 마지막 row: "현재 PR #542 의 e2e-internal job 이 build tag 의 CI matrix 의 1/2 와 동일" → **삭제** (baseline 변경).
- §3.2 옵션 2: "현재 PR #542 1쌍" → "**e2e shard 1/2/3 saovae_stub default 만**".
- §4 결정 근거 7번 (new): "e2e-internal 폐기 결정 정합 — 사용자는 e2e-internal 을 사내 환경용 셋팅이라 명시. CI 의 1 axis (build tier) 가 사라짐. 사내 staging/prod-smoke 가 real adapter 검증 책임".
- §5 재검토 trigger 3번: "현재 2 axes" → "**2026-06-12 기준 1 axis (shard only)** — e2e-internal 폐기 후".
- §6.1 표의 `.github/workflows/ci.yml (e2e-internal job)` row → **삭제** (job 자체가 없어짐).
- §6.2: "PR #542 의 1쌍 matrix 가 현시점 optimal" → "**e2e shard 1/2/3 (saovae_stub default) 가 현시점 optimal**".
- §7.1 risks 3번: "현시점 2 axes (build tier × shard)" → "**2026-06-12 기준 1 axis (shard only)** — e2e-internal 폐기 후".
- §7.1 risks 4번 (new): "e2e-internal 폐기 결정의 운영 위험 — e2e shard 의 saovae_stub default 검증만으로 real adapter production 회귀 미검증. 사내 staging/prod-smoke 가 검증 책임 (의도). 별도 risk 아님".
- §8 supersession: "본 ADR 은 ADR-0030 을 supersede 하지 않음" + "2026-06-12 partial supersession (baseline 변경 정공법)" 절 추가.
- §9 변경 이력: 1 row 추가 (2026-06-12 본 sprint, 결론 변동 0건 명시).

### 1.4 `docs/adr/0030-sso-integrations-and-auth-session-port.md` (ADR 본체 reference 갱신)

- §2.3 row: "**2026-06-12 partial supersession (baseline 변경 정공법)**" reference 추가 (e2e-internal 폐기 = runtime injection 결정과 독립).
- §9 변경 이력: 1 row 추가 (2026-06-12 partial supersession 정공법).

### 1.5 `docs/traceability/report.md` (§6 row 추가)

- §6 변경 이력 본 row 신규 — 4 item 보고 (CI / script / ADR-0031 partial supersession / ADR-0030 reference 갱신) + 신규 ID 발급 0건 + Tier 공용.

## 2. Trade-off (e2e-internal 폐기 결정)

| 항목 | 폐기 전 (PR #542 baseline) | 폐기 후 (본 sprint) |
|---|---|---|
| CI matrix | e2e shard 1/2/3 + e2e-internal (1쌍) | e2e shard 1/2/3 만 (단일) |
| CI runtime | +15~20min (e2e-internal 1 job) | 0 min 추가 (e2e shard 만) |
| Keycloak container | 2개 (port 8180 + 8181) | 1개 (port 8180 만) |
| Real adapter 검증 책임 | GitHub Actions (e2e-internal) | 사내 staging/prod-smoke (운영 SOP) |
| 사내 환경 의존성 | GitHub Action 에 사내 Keycloak 의존 (race 위험) | 운영 환경 자연 검증 |

**결론**: 사내 환경 의존성을 GitHub Actions 에서 사내 운영 환경으로 이동. **race 위험 제거 + CI 단순화**.

## 3. 후속 (사용자 결정 영역)

- e2e-internal 폐기 결정의 운영 SOP 갱신: 사내 staging/prod-smoke 가 real adapter 검증 절차 (별도 docs). 본 sprint scope 외.
- ADR-0030 / ADR-0031 본문 외 다른 doc 의 e2e-internal reference 정합: 본 sprint 에서 5 file 변경 (1차). 잔여 N-13 follow-up 결정의 "E2E Internal 1 fail" historical record (PR #548 close) 는 손대지 않음 (별개 historical context).
- v1.1 milestone 진입 시점: PR #548 의 구현 follow-up 별도 sprint (e2e-internal 없이 검증 정합).

## 4. 검증 (PR 머지 직전)

- [ ] `python3 -c "import yaml; data = yaml.safe_load(open('.github/workflows/ci.yml')); print(len(data['jobs']))"` = 9 (e2e-internal 빠짐)
- [ ] `bash scripts/ci-e2e-sync-check.sh .github/workflows/ci.yml` PASS
- [ ] `python3 -m pip install --user --quiet "PyYAML>=6.0,<7" && bash scripts/check-openapi-yaml-lint.sh` PASS (openapi 변경 없음, sanity check)
- [ ] `bash scripts/check-tier-separation.sh` PASS
- [ ] `git diff --stat` = 5 file 변경 (ci.yml + script + 2 ADR + traceability)
- [ ] `git log -1 --format='%an %ae'` = 본 세션 author (force push 방지)
