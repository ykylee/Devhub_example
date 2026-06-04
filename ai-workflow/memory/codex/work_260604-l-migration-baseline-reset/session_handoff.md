# Session Handoff — codex/work_260604-l-migration-baseline-reset

- 문서 목적: migration baseline reset 브랜치의 구현/검증 상태와 PR 직전 진입점을 인계한다.
- 범위: backend-core migration 체인 재편, legacy archive, 핵심 문서 정리
- 대상 독자: 후속 에이전트, 리뷰어
- 상태: in_progress
- 최종 수정일: 2026-06-04

## 이번 세션 요약

- `docs/planning/migration_baseline_reset_plan_2026-06-04.md` 에 baseline reset 실행 계획을 먼저 문서화했다.
- `backend-core/migrations` active 체인을 아래 3단계로 재구성했다.
  - `000001_initial_schema`
  - `000002_seed_system_rbac`
  - `000003_seed_bootstrap_catalog`
- 기존 `000001_create_webhook_events` ~ `000047_normalize_team_manager_display_name` SQL 파일은 `backend-core/migrations-legacy-20260604/` 로 이동해 historical reference 로 보존했다.
- baseline replay 결과를 기준으로 `org_units`, `users`, `unit_appointments`, `scm_providers`, `rbac_policies` bootstrap seed 를 유지했다.
- 핵심 안내 문서를 새 active 체인 기준으로 갱신했다.
  - `backend-core/migrations/README.md`
  - `docs/governance/code-taxonomy.md`
  - `docs/traceability/report.md`

## 검증 결과

- 빈 DB 기준 `migrate -path backend-core/migrations ... up` 성공 (`version=3`, `dirty=false`)
- 빈 DB 기준 `migrate ... down -all` round-trip 성공
- `bash ./scripts/check-migration-uniqueness.sh` PASS
- `cd backend-core && go test -count=1 -run 'TestIntegration_' ./internal/store/...` PASS
- `cd backend-core && go test ./...` PASS
- frontend test/build 는 미실행

## 다음 세션 첫 작업

1. PR 생성: "기존 DB upgrade 포기 + DB 재생성 전제" 를 본문에 명확히 적는다.
2. 리뷰 대비: legacy archive 유지 이유와 bootstrap seed 유지 이유를 설명한다.
3. 필요 시 historical 문서 중 migration 번호 참조 잔여분을 추가 정리한다.
