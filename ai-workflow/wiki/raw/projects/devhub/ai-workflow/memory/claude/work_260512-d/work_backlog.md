# 작업 백로그

`claude/work_260512-d` 슬롯 — 직전 슬롯 work_260512-c (PR #71) Planned #1 처리 (kratos/hydra migrate 자동화).

## [Planned]
1. OIDC client 등록 (`register-devhub-client.ps1`) 의 dev-up 통합 — fresh DB 진짜 zero-touch 까지. PS 만 있는지 sh 평행 필요한지 평가 동반.
2. `gemini/frontend_260510` → `main` 머지 전략 검토 — 이 라인 위에 #63-#71 + 본 PR (예상 #72) 까지 누적. main 통합 시점 / 방식 결정.

## [In Progress]
(없음)

## [Done — 이번 세션]

### 평가
- [x] kratos / hydra migrate sql 의 syntax + idempotency 확인 (Ory v26 `migrate sql up --yes <DSN>` 가 표준 form, applied version skip)
- [x] schema 생성 (`CREATE SCHEMA IF NOT EXISTS`) 의 idempotency 확인
- [x] backend-core/cmd/idp-apply-schemas Go 헬퍼가 psql 의존 없이 SQL 적용 가능 확인 (pgx/v5 재사용)
- [x] 사용자와 옵션 합의: 옵션 A 부분 적용 (schema + IdP migrate, OIDC client 는 scope 외)

### 구현
- [x] dev-up.ps1: backend-core migrate 직후 단계 1b — schema apply + kratos migrate + hydra migrate. `DEVHUB_SKIP_IDP_MIGRATE=1` 우회. kratos/hydra missing 시 graceful skip with warning.
- [x] dev-up.sh: 동일 의미론 bash 평행 — `( cd backend-core && go run ./cmd/idp-apply-schemas ... )` subshell, `kratos migrate sql up --yes "$(idp_dsn ...)"`.
- [x] infra/idp/README.md §2 머리에 자동화 안내 박스 추가.

### 실머신 검증
- [x] 사전: 6 포트 free, `.pids/` 없음
- [x] dev-up: schema apply Done, kratos migrate plan 모두 Applied (idempotent skip), hydra Successfully applied migrations, 5/5 서비스 ready, backend `/health` 200
- [x] dev-down: backend PID-kill + frontend port-sweep 패턴 그대로 (#71 의 결함 #2 픽스 무회귀)
- [x] 종료 후 6 포트 free, `.pids/` 비움

## [Carried over]
- 직전 슬롯 Planned #2 (gemini/frontend_260510 → main 머지 전략) — 본 슬롯 Planned #2 로 이월
