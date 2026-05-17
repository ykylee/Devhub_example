# PR Draft (codex/next-step-20260516)

## 제목(안)

`feat(infra/integration): HomeLab pull adapter 고도화 + snapshot 영속화 + Prometheus 관측성 추가`

## 본문(초안)

## 요약

HomeLab infra integration 경로를 push/pull 혼합 모델로 확장했다.  
snapshot 영속화와 adapter 경계를 도입하고, pull scheduler + HTTP puller(retry/backoff) + Prometheus `/metrics` 계측을 추가해 외부 서비스/서버 운영 가시성을 확보했다.

## 변경 상세

- backend-core
  - API-73 verifier strategy 확장 (`provider_sdk:<provider>:<secret>`)
  - API-76/77/78 baseline + ingest 영속화/조회 hydrate 경로 보강
  - HomeLab adapter 계층 도입 (`internal/integrations/adapters/*`)
    - normalize + health policy
    - file puller
    - HTTP puller + retry/backoff
    - pull loop scheduler
  - Prometheus 지표 추가 + `/metrics` endpoint 노출
  - 관련 config/env 확장
- migration/store
  - `000029_infra_service_snapshots` 추가
  - `SaveInfraSnapshot/LoadLatestInfraSnapshot` 구현
- docs
  - architecture/api_contract/traceability 동기화
  - integration 테스트 문서 + 실행 리포트 갱신
  - HomeLab pull 전략 및 Prometheus alert/dashboard 초안 추가

## 추적성 영향

- 추가:
  - API-76, API-77, API-78 구현/테스트/영속화 연계 항목
  - IMPL-int-store-02, IMPL-int-migration-02 (영속화 계층)
- 갱신:
  - API-73 verifier strategy (`provider_sdk`) 관련 row
  - Integration role/resilience test TC/UT 매핑 row
  - HomeLab adapter/observability 관련 matrix row
- Deprecate:
  - (없음)
- 매트릭스:
  - `docs/traceability/report.md` 갱신 (integration + homelab + persistence + observability 반영)

## 테스트

- [x] 로컬 backend `go test ./...`
- [ ] 로컬 frontend `npm run test` (이번 변경 범위 외)
- [ ] CI: backend-unit / frontend-unit / e2e SUCCESS
- [ ] 수동 검증
  - [ ] `/api/v1/infra/services/snapshot` ingest 후 API-76/78 조회 반영 확인
  - [ ] `DEVHUB_HOMELAB_PULL_ENABLED=1` + file/http pull 동작 확인
  - [ ] `/metrics`에서 HomeLab 지표 노출/변화 확인

## 관련 issue / ADR

- Traceability/Architecture/API docs 동기화 포함
