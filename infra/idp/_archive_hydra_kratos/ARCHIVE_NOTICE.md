# Archived — Ory Hydra / Ory Kratos PoC 자산

본 디렉터리는 ADR-0001 (Ory Hydra + Kratos 선정, 2026-05-07) 시기에 생성된 자산으로,
[ADR-0019 — Keycloak 단일화](../../../docs/adr/0019-keycloak-only-idp.md) 결정 (2026-05-19) 으로
DevHub 의 active IdP stack 에서 폐기되었다.

## 보존 사유
- ADR-0001 의 historical context 와 의사결정 reversal 의 추적성 보존 (immutable history)
- Hydra/Kratos PoC 시점의 사내 환경 특수 제약 (`ENVIRONMENT_NOTES.md`) 의 일부는 현재의 Keycloak 운영에도 환경 정보로서 참고 가치 보유

## 사용 금지 사항
- 본 디렉터리 내 어떤 yaml / sh / ps1 / jsonnet / json 도 active 운영 자산이 아니다
- `kratos serve` / `hydra serve` 등의 가동은 ADR-0019 위배
- 운영 IdP 환경 구성은 다음을 따른다:
  - [`docs/adr/0019-keycloak-only-idp.md`](../../../docs/adr/0019-keycloak-only-idp.md)
  - [`docs/setup/keycloak_operations.md`](../../../docs/setup/keycloak_operations.md)
  - [`docs/setup/single_port_deployment.md`](../../../docs/setup/single_port_deployment.md)
  - 로컬 dev 모드 realm: `infra/idp/keycloak-realm.dev.json`
  - 외부 Keycloak 운영 모드 realm 템플릿: `infra/idp/keycloak-realm.prod.json`

## 파일 목록
- `README.md` — Hydra/Kratos PoC setup 가이드 (deprecation banner 포함)
- `ENVIRONMENT_NOTES.md` — 사내 환경 특수 제약 메모 (일부 내용은 Keycloak 운영에도 참고)
- `hydra.yaml`, `hydra.ci.yaml`, `hydra.deploy.yaml` — Ory Hydra 설정
- `kratos.yaml`, `kratos.ci.yaml`, `kratos.deploy.yaml` — Ory Kratos 설정
- `kratos_webhooks/` — Kratos webhook 스크립트
- `scripts/install-binaries.ps1` — Hydra/Kratos 바이너리 설치 헬퍼

본 디렉터리는 향후 ADR-0001 / ADR-0019 의 reader 참조용으로만 보존되며, 별도 carve sprint 가 결정하면 일괄 삭제될 수 있다.
