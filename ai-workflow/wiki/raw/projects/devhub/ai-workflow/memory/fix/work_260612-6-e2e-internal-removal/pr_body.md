# ci(workflow): e2e-internal job 폐기 + ADR-0030/0031 partial supersession (baseline 변경 정공법)

## 결정 (사용자 2026-06-12)

> "e2e-internal은 어차피 사내 환경용 셋팅이면 github action으로 체크할건 아니야 그냥 없애줘."

**본 PR 의 정공법 = housekeeping follow-up + ADR partial supersession (baseline 변경 정공법)**. e2e-internal job 자체 폐기 + ADR-0031 의 정량 측정 baseline 변경 (CI matrix 1쌍 → 단일 matrix). 결론 (runtime injection 유지) 자체는 변동 0건 — runtime injection 결정과 e2e-internal 폐기는 **독립 결정**.

## 카테고리 · 모듈

- 카테고리: `ci` / `governance`
- 모듈: `.github/workflows/`, `scripts/`, `docs/adr/`, `docs/traceability/`

## 변경 요약 (5 file)

| # | 파일 | 변경 | line |
| --- | --- | --- | --- |
| 1 | `.github/workflows/ci.yml` | e2e-internal job 완전 삭제 (line 554-756, ADR-0030 §2.3 + C-i reference 코멘트 포함) | -205 |
| 2 | `scripts/ci-e2e-sync-check.sh` | DEVHUB_BUILD_TIER 의도적 미포함 코멘트 정리 (5 lines) | -5 |
| 3 | `docs/adr/0031-build-tag-policy-review.md` | **partial supersession (baseline 변경 정공법)** — §1.2 / §2.1~3 / §3.1~2 / §4 / §5 / §6.1 / §7.1 / §8 baseline 갱신 + §9 row 추가 | +30/-15 |
| 4 | `docs/adr/0030-sso-integrations-and-auth-session-port.md` | §2.3 row + §9 row 갱신 (partial supersession reference) | +4/-2 |
| 5 | `docs/traceability/report.md` | §6 row 추가 (1 row, 본 sprint 정공법 보고) | +1/-0 |

## 정공법 핵심

### 1. e2e-internal job 자체 폐기

- **사용자 결정 (verbatim)**: "e2e-internal은 어차피 사내 환경용 셋팅이면 github action으로 체크할건 아니야 그냥 없애줘."
- **근본 layer**: e2e-internal = real Keycloak adapter 환경 (sprint -a follow-up PR #540 + #542 의 CI matrix 1쌍) 의 검증. 사내 환경 의존성 (Keycloak container real wire) + GitHub Actions 의 ephemeral 환경 race = **CI 에서 검증할 의미가 약함**.
- **책임 이관**: 사내 staging/prod-smoke 가 real adapter 검증 책임 (운영 SOP 자연 검증).
- **사내 환경 의존성 제거**: GitHub Actions 의 Keycloak container 2개 (port 8180 + 8181) → 1개 (port 8180 만). race 위험 제거 + CI 단순화.

### 2. ADR-0031 partial supersession (baseline 변경 정공법)

**e2e-internal 폐기 후 정량 비교 가능**. 이전 비교 baseline = CI matrix 1쌍 (e2e shard 1/2/3 saovae_stub + e2e-internal real) → 새 baseline = 단일 matrix (e2e shard 1/2/3 saovae_stub default 만).

| 측정 항목 | 폐기 전 | 폐기 후 |
|---|---|---|
| Runtime injection CI matrix | 2 jobs × shard (e2e 1/2/3 saovae + e2e-internal real) | **1 axis (shard only)** — e2e shard 1/2/3 (3 jobs) |
| Runtime injection CI runtime | +15~20min (e2e-internal 1 job) | **0 min 추가** (e2e shard 만) |
| Build tag CI matrix jobs | 4 (e2e 1/2/3 + e2e-internal) | **3 (e2e 1/2/3)** |
| Build tag CI runtime | +15~40min (4 jobs) | **+15~30min (6 vs 3 jobs)** |
| Trade-off (1:N cost ratio) | 1:5+ | **1:3+** (e2e-internal 분 제외 후 build tag cost 가 더 명확히 부각) |

**결론 (변동 0건)**: build tag 의 binary size 절감 (~6KB) 은 무시 가능 수준 (전체 backend-core < 50MB 대비 0.01% 미만). **runtime injection 의 cost 가 build tag 의 cost 보다 본질적으로 작음**. e2e-internal 폐기 후 비교는 폐기 전보다 build tag 의 cost 가 더 명확히 부각됨 (이전엔 e2e-internal 1 job 이 build tag 의 1 set 와 중복 → trade-off 모호; 폐기 후 build tag 가 pure additional cost 가 됨).

### 3. ADR-0030 §2.3 runtime injection 결정 = confirmed 유지

- e2e-internal 폐기와 runtime injection 결정은 **독립 결정**.
- runtime injection 의 사내 staging/prod-smoke 적용은 그대로.
- 사내 staging/prod-smoke 가 real adapter 검증 책임 (e2e-internal 의 role 대체).

## 추적성 영향

- **신규 ID 발급 0건** (housekeeping follow-up 정공법, ADR supersession 정공법 §4.2 정합)
- ADR-0031 §9 변경 이력: 1 row 추가 (2026-06-12 본 sprint, 결론 변동 0건 명시)
- ADR-0030 §9 변경 이력: 1 row 추가 (2026-06-12 partial supersession reference)
- `docs/traceability/report.md` §6: 1 row 추가 (4 item 보고)

## Tier

- [ ] 사외 (GitHub main)
- [ ] 사내 (사내 SCM)
- [x] 공용 (양쪽 동기화)

> **Note**: 본 PR 은 CI config + ADR + script + memory 모두 사내 한정 정보 미포함 → GitHub main push 가능 (현실적으로 public repo). 사내 SCM push 불요.

## 사내 한정 정보 self-check

- [x] 사내 env var 미포함 (e2e-internal job 의 `DEVHUB_KEYCLOAK_*` env var 모두 제거됨)
- [x] 사내 호스트 / IP 대역 미포함
- [x] 사내 한정 경로 (`infrastructure/`, `infra/idp/`, `scripts/setup-keycloak.sh` 등) 변경 없음
- [x] `.env.*` 의 사내 env var 추가/변경 없음

## 검증

```bash
# 1. YAML parse + jobs count
$ python3 -c "import yaml; data = yaml.safe_load(open('.github/workflows/ci.yml')); print(len(data['jobs']))"
9
# e2e-internal 빠짐. 9 jobs 잔여.

# 2. ci-e2e-sync-check
$ bash scripts/ci-e2e-sync-check.sh .github/workflows/ci.yml
E2E-CI sync contract check passed.

# 3. tier-separation
$ bash scripts/check-tier-separation.sh
no changes between origin/main and HEAD

# 4. migration prefix uniqueness
$ bash scripts/check-migration-uniqueness.sh
✅ All migration prefixes are valid and unique!
```

## 후속 (사용자 결정 영역)

- **사내 staging/prod-smoke 의 real adapter 검증 SOP 갱신** (별도 docs, 본 sprint scope 외)
- **v1.1 milestone 진입 시점**: PR #548 의 구현 follow-up sprint (e2e-internal 없이 검증 정합)
- **본 PR 머지 후 main flat memory 4종 sync** (PR 머지 시점 자동)

## Refs

- [ADR-0030 §2.3 runtime injection 결정](https://github.com/ykylee/Devhub_example/blob/main/docs/adr/0030-sso-integrations-and-auth-session-port.md) — confirmed 유지
- [ADR-0031 build tag 정책 재검토 (partial supersession)](https://github.com/ykylee/Devhub_example/blob/main/docs/adr/0031-build-tag-policy-review.md) — baseline 변경 정공법
- [PR #542 (e2e-internal job, MERGED 2026-06-10)](https://github.com/ykylee/Devhub_example/pull/542) — 본 sprint 의 정공법으로 폐기
- [PR #577 (E2E Internal disable, MERGED 2026-06-12)](https://github.com/ykylee/Devhub_example/pull/577) — env var 게이트 (`SKIP_E2E_INTERNAL=true`) — 본 sprint 의 완전 폐기와 정합
- 사용자 결정 (2026-06-12): "e2e-internal은 어차피 사내 환경용 셋팅이면 github action으로 체크할건 아니야 그냥 없애줘"
