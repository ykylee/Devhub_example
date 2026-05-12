# 세션 인계 문서 (2026-05-12 work continuation slot — IdP migrate automation)

## 세션 목표

직전 슬롯 `claude/work_260512-c` 의 Planned #1 — kratos / hydra `migrate sql` 자동화 평가 + 적용. 사용자 선택: **옵션 A 부분 적용** (schema 생성 + IdP migrate 까지 자동화, OIDC client 등록은 scope 외).

## 평가 결과

| 단계 | 도구 | 멱등성 | dev-up 통합 |
| --- | --- | --- | --- |
| schema 생성 | `backend-core/cmd/idp-apply-schemas` Go 헬퍼 + `infra/idp/sql/001_create_idp_schemas.sql` (`CREATE SCHEMA IF NOT EXISTS`) | ✓ | ✓ 추가 |
| kratos migrate | `kratos migrate sql up --yes <DSN>` | ✓ (applied version skip) | ✓ 추가 |
| hydra migrate | `hydra migrate sql up --yes <DSN>` | ✓ | ✓ 추가 |
| OIDC client 등록 | `infra/idp/scripts/register-devhub-client.ps1` | ✓ (delete-then-create) | ✗ scope 외 (hydra 가동 후 별도 단계) |

## 픽스

### `dev-up.ps1` / `dev-up.sh`
backend-core migrate (단계 1) 직후 새 단계 1b 삽입:
- Schema apply: `go run ./cmd/idp-apply-schemas -dsn $DbUrl -sql ...` (psql 의존 없음 — 결함 가능성 회피)
- Kratos migrate: `kratos migrate sql up --yes <kratos DSN>` (결함 #4 의 `idp_dsn` 헬퍼 재사용)
- Hydra migrate: `hydra migrate sql up --yes <hydra DSN>`
- 환경변수 `DEVHUB_SKIP_IDP_MIGRATE=1` 로 우회 가능
- kratos/hydra 가 PATH 에 없으면 warning 후 해당 migrate 만 skip (graceful)

### `infra/idp/README.md` §2 머리
"dev-up.ps1 / dev-up.sh 가 §2 + §3 을 자동 실행. 일상 개발은 `./dev-up.sh` 하나로 충분. 본 절차는 디버깅/수동 실행 reference" 안내 박스 추가.

## 실머신 검증

이미 schema / migrate 가 적용된 환경에서 dev-up.ps1 실행:
- `Schemas present: hydra, kratos` (apply 헬퍼 출력)
- Kratos migrate plan 출력 (모든 행 `Applied` — 멱등 skip)
- Hydra migrate `Successfully applied migrations!`
- 5/5 서비스 ready
- backend `/health` 200
- dev-down: backend PID-kill + frontend port-sweep 패턴 유지 (직전 슬롯 결함 #2 픽스 그대로)
- 종료 후 6 포트 free, `.pids/` 비움

## 다음 슬롯 출발점 후보

1. **이번 묶음 commit + PR + 본인 리뷰 모드 머지** — `dev-up.ps1` / `dev-up.sh` / `infra/idp/README.md` + 본 메모리 3 파일. base=`gemini/frontend_260510`.
2. **OIDC client 등록 자동화** — `register-devhub-client.ps1` 를 dev-up 의 hydra ready 이후로 통합할지. 이걸 통합하면 fresh DB → 진짜 zero-touch dev-up. 단, scripts/ 가 PS-only 면 sh 평행 구현 필요.
3. **`gemini/frontend_260510` → `main` 머지 전략 검토** — 이 feature branch 위에 #63-#71 + 본 PR 누적. 통합 시점 결정.
4. **register-devhub-client.ps1 의 bash 평행** — PoC 자산이 PS 만 있는지, sh 도 필요한지 평가.

## 환경 정리 상태 (현 시점)

- 6 포트 모두 free
- `.pids/` 비움
- log 잔재 제거
- `dev-bin/backend-core.exe` 그대로 (gitignore 되어 영향 없음)
- 작업 트리: `dev-up.ps1` / `dev-up.sh` / `infra/idp/README.md` + 본 슬롯 3 파일
