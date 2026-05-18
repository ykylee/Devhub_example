# Work Backlog — codex/keycloak-only-refactor-plan

- 문서 목적: Keycloak 단일화 리팩토링 브랜치의 작업 백로그를 관리한다.
- 범위: 구현 단계별 TODO, 상태 추적
- 대상 독자: 구현 담당자, 리뷰어
- 상태: active
- 최종 수정일: 2026-05-18

## 백로그

| ID | 작업 | 상태 | 비고 |
| --- | --- | --- | --- |
| KC-PLAN-01 | Keycloak 단일화 실행 계획 문서화 | done | `docs/planning/keycloak_only_refactor_execution_plan.md` |
| KC-BOOT-01 | 브랜치 memory 초기화(state/handoff/backlog) | done | 본 브랜치 디렉터리 생성 완료 |
| KC-PR-A | config/provider 스켈레톤 구현 | done | 완료 커밋: `b8eb2ba` |
| KC-PR-B | Keycloak JWT/JWKS verifier 전환 | done | `internal/auth/keycloak_verifier.go` + verifier tests + main wiring |
| KC-PR-C | account/admin API Keycloak Admin 연동 | in_progress | keycloak 모드 Kratos mock 주입 방지까지 반영 |
| KC-PR-D | frontend auth/logout flow 전환 | planned | OIDC discovery 기반 |
| KC-PR-E | identity 컬럼 일반화 마이그레이션 | planned | `kratos_identity_id` -> `idp_subject` |
| KC-PR-F | 테스트/문서/traceability 동기화 | planned | e2e + 계약 문서 갱신 |
