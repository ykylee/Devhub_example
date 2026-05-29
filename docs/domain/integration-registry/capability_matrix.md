# External Integration Capability Matrix

- 문서 목적: 외부 시스템 연동 후보(Jira/Confluence/Bitbucket/Bamboo/Jenkins/Gitea/Forgejo/HomeLab)의 기능 범위를 공통 capability 키 기준으로 정리하고, MVP 구현 우선순위를 명확히 한다.
- 범위: Provider별 지원 capability, 수집 방식(webhook/pull), 인증 모드, 운영 상태 지표, MVP/후속 구분.
- 대상 독자: Backend/Frontend/운영 담당자, 아키텍처 검토자, QA, AI 에이전트.
- 상태: draft
- 최종 수정일: 2026-05-15
- 관련 문서: [external_system_integration_concept.md](./external_system_integration_concept.md), [system_usecases.md](./system_usecases.md), [../architecture.md](../architecture.md), [../backend_api_contract.md](../backend_api_contract.md)

## 1. 공통 capability 키

| Provider Type | Capability Key |
| --- | --- |
| `alm` | `issue.read`, `epic.read`, `issue.link` |
| `doc` | `page.read`, `space.read`, `doc.link` |
| `scm` | `repo.read`, `pr.read`, `branch.read`, `webhook.ingest` |
| `ci_cd` | `build.read`, `deploy.read`, `job.rerun` |
| `infra` | `node.read`, `service.read`, `snapshot.ingest` |

## 2. Provider별 매트릭스 (초안)

| Provider | Type | 수집 방식 | 인증 모드 후보 | 주요 capability | MVP |
| --- | --- | --- | --- | --- | --- |
| Jira | `alm` | Pull 우선, Webhook 선택 | `oauth2`, `token` | `issue.read`, `epic.read`, `issue.link` | Yes |
| Confluence | `doc` | Pull 우선 | `oauth2`, `token` | `page.read`, `space.read`, `doc.link` | Yes (링크형) |
| Bitbucket | `scm` | Webhook + Pull | `oauth2`, `app_password` | `repo.read`, `pr.read`, `branch.read`, `webhook.ingest` | 후보 |
| Gitea | `scm` | Webhook + Pull | `token` | `repo.read`, `pr.read`, `branch.read`, `webhook.ingest` | Yes |
| Forgejo | `scm` | Webhook + Pull | `token` | `repo.read`, `pr.read`, `branch.read`, `webhook.ingest` | 후보 |
| Bamboo | `ci_cd` | Pull 우선 | `token`, `basic` | `build.read`, `deploy.read` | 후보 |
| Jenkins | `ci_cd` | Webhook(선택) + Pull | `token`, `basic` | `build.read`, `deploy.read`, `job.rerun` | Yes |
| HomeLab Agent | `infra` | Agent Push | `agent` | `node.read`, `service.read`, `snapshot.ingest` | Yes |

## 3. MVP 구현 권장 조합

| 도메인 | 1차 권장 Provider | 이유 |
| --- | --- | --- |
| ALM | Jira | 팀 표준 이슈 트래킹 SoT 후보 |
| 문서 | Confluence | 설계/운영 문서 링크형 통합 우선 |
| SCM | Gitea (우선), Bitbucket/Forgejo (확장) | 기존 webhook 파이프라인 자산 활용 가능 |
| CI/CD | Jenkins (우선), Bamboo (확장) | 홈랩/사내 환경 범용성 |
| Infra | HomeLab Agent | 서버/앱 설치 현황 수집 핵심 |

## 4. 운영 상태 지표 표준

| 지표 | 설명 |
| --- | --- |
| `sync_status` | `requested | verifying | active | degraded | disconnected` |
| `last_sync_at` | 마지막 성공 동기화 시각 |
| `last_error_code` | 마지막 오류 코드 |
| `event_lag_seconds` | 외부 이벤트 발생시각 대비 적재 지연 |
| `snapshot_freshness_seconds` | 홈랩 스냅샷 최신성 |

## 5. 후속 설계 TODO

1. Provider별 rate limit/재시도(backoff) 파라미터를 표준 설정으로 분리한다.
2. `credentials_ref` secret rotation 정책(주기/오류 시나리오)을 ADR 후보로 발급한다.
3. `execution_system` 정책에서 단일 SoT 강제 여부를 프로젝트별 옵션으로 확정한다.

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-15 | 초안 작성 — 외부 연동 provider capability matrix 신설 (MVP 후보/수집 방식/인증 모드/운영 지표 정의). |
