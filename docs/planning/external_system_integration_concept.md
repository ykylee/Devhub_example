# 외부 시스템 연동 도메인 컨셉

- 문서 목적: DevHub 에서 ALM/SCM/CI-CD/문서관리/홈랩 인프라를 통합 운영하기 위한 **외부 시스템 연동 도메인**의 컨셉을 정의하고, 후속 요구사항(REQ) → 유스케이스(UC) → 설계(ARCH/API) 고도화의 기준선을 만든다.
- 범위: 컨셉 단계. 연동 대상군 정의, 공통 연동 모델, 시스템별 후보 정책(Jira/Confluence/Bitbucket/Bamboo/Jenkins/Gitea/Forgejo), 홈랩 인프라 관리 범위, SoT 경계, MVP scope, out-of-scope, 미해결 항목, 후속 단계 hook.
- 대상 독자: 프로젝트 리드, Backend/Frontend/Auth/운영 담당자, 홈랩 운영자, AI agent, 리뷰어.
- 상태: draft
- 최종 수정일: 2026-05-15
- 관련 문서: [project_management_concept.md](./project_management_concept.md), [development_request_concept.md](./development_request_concept.md), [system_usecases.md](./system_usecases.md), [system_erd.md](./system_erd.md), [backend_api_contract.md](../backend_api_contract.md), [development_roadmap.md](../development_roadmap.md), [traceability/report.md](../traceability/report.md)

## 1. 컨셉 정리 배경

- 현재 DevHub 는 Application/Repository/Project 계층과 DREQ 도메인을 통해 내부 운영 모델의 1차 골격을 확보했다.
- 다음 단계는 "실행 데이터가 발생하는 외부 시스템"을 연동해, 계획-개발-배포-운영의 흐름을 하나의 운영 시야로 묶는 것이다.
- 특히 팀의 실운영 환경은 다음 두 축이 동시에 존재한다:
  - 업무 시스템 축: Jira/Confluence/Bitbucket/Bamboo/Jenkins/Gitea/Forgejo
  - 인프라 축: 홈랩 서버/서비스 현황(노드 상태, 설치 앱, 버전, 헬스, 의존 관계)
- 본 컨셉은 기술 연결 자체보다 **운영 경계와 책임 분리**를 우선 정의한다:
  - 어디가 Source of Truth(SoT) 인가
  - 무엇을 DevHub 에 복제/정규화할 것인가
  - 어떤 권한으로 누가 조회/수정할 수 있는가

## 2. 도메인 정의

### 2.1 연동 도메인 목표

외부 시스템별 원천 데이터를 DevHub 공통 모델로 정규화해 다음을 제공한다.

1. 운영 가시성: 프로젝트/리포지토리/파이프라인/문서/인프라 상태를 단일 화면에서 조회
2. 추적성: 요구사항-작업-코드-배포-운영 이벤트를 연계
3. 통제성: 권한 기반 조회/수정, 감사 로그, 장애 시 격리 대응

### 2.2 연동 대상군 (1차 후보)

| 그룹 | 후보 시스템 | DevHub 관점 역할 |
| --- | --- | --- |
| ALM | Jira | 이슈/에픽/상태 전이 원천 |
| 문서 | Confluence | 운영/설계 문서 원천 |
| SCM | Bitbucket, Gitea, Forgejo | 저장소/PR/브랜치 활동 원천 |
| CI/CD | Bamboo, Jenkins | 빌드/배포 실행 상태 원천 |
| Infra/HomeLab | Linux 서버, VM, 컨테이너 런타임, 앱 프로세스 | 노드/서비스 상태 원천 |

### 2.3 핵심 운영 원칙

1. 원천 불변 원칙: 외부 시스템이 SoT 인 데이터는 DevHub 에서 직접 수정하지 않는다(미러+정규화).
2. 제어 분리 원칙: 조회 경로(read)와 제어 경로(write/action)를 분리한다.
3. 장애 격리 원칙: 특정 어댑터 실패가 전체 도메인 장애로 전파되지 않도록 provider 단위로 격리한다.
4. 감사 일관성 원칙: 연동 생성/변경/실패/복구 이벤트를 공통 audit namespace로 기록한다.

## 3. SoT(원천) 경계

| 데이터 영역 | SoT | DevHub 역할 |
| --- | --- | --- |
| 이슈/에픽 상태 | Jira (또는 조직이 채택한 ALM) | 동기화/조회/링크 |
| 저장소/PR/빌드 이벤트 | SCM/CI 원천 시스템 | 정규화/집계/뷰 제공 |
| 운영 문서 본문 | Confluence | 링크/메타/상태 참조 |
| 인프라 상태(노드/서비스 헬스) | 홈랩 수집기(Agent) | 토폴로지/상태 대시보드/이력 |
| DevHub 내부 엔터티(Application/Project/DREQ) | DevHub | 워크플로우 오케스트레이션 |

## 4. 연동 모델 (공통)

### 4.1 공통 어댑터 모델

- Provider Registry: `provider_key` + `capabilities` + `enabled` + `auth_mode`
- Adapter Contract: `pull`, `webhook_ingest`, `normalize`, `health_check`
- Normalized Event: `provider`, `resource_type`, `external_id`, `event_type`, `occurred_at`, `payload`
- Sync State: `requested | verifying | active | degraded | disconnected`

### 4.2 홈랩 인프라 모델

홈랩은 "연동 대상"이면서 동시에 "운영 대상"이다.

- Node: 서버/VM/컨테이너 호스트 단위 (CPU, MEM, DISK, load, uptime, reachability)
- Service: 노드 위 앱 단위 (name, version, port, health, last_deploy, owner)
- Edge: 서비스 간 의존 (API, DB, Queue, Webhook)
- Snapshot + Event: 현재 상태 + 변경 이력(`infra.node.updated`, `infra.service.updated`)

### 4.3 DREQ Intake Token 과의 경계

- `dev_request_intake_tokens` 는 DREQ 외부 수신을 위한 **도메인 특화 Push-only intake 채널**로 취급한다.
- 1차 정책은 Integration Provider Registry 와 분리 관리한다.
  - 이유: DREQ intake 는 일반 provider catalog 수명주기(조회/동기화/reconcile)보다 인증·수신 경계가 우선이기 때문.
- 단, 운영 통합 관점에서는 "external intake endpoint" 메타를 공통 관제 화면에서 함께 노출할 수 있도록 확장 여지를 둔다.

## 5. 시스템별 초기 정책 (컨셉 수준)

### 5.1 Jira / Confluence

- Jira: 실행 작업의 SoT. Application/Project 와 매핑된 key 정책만 DevHub 에서 관리.
- Confluence: 운영 문서의 SoT. DevHub 는 문서 링크/스페이스/최종 갱신 시각 중심으로 참조.

### 5.2 Bitbucket / Gitea / Forgejo

- SCM 은 동등 provider 원칙으로 취급한다.
- Repository/PR/activity 정규화 스키마는 공통, provider 특화 필드는 payload 확장으로 보존.

### 5.3 Bamboo / Jenkins

- CI/CD 결과(성공/실패/지속시간/브랜치/커밋)를 공통 build_run 모델로 정규화한다.
- 배포 단계 정보(deploy target/env)는 후속 스키마 확장 후보로 분리한다.

## 6. MVP scope (컨셉 후속 1차 구현 범위)

### 6.1 In-scope (MVP)

- Integration 등록/조회/수정/비활성화의 공통 도메인 흐름
- SCM 1계열 + CI 1계열 + ALM/문서 링크형 연동(최소 1개씩) PoC
- 홈랩 Node/Service 인벤토리 + 상태 조회 대시보드(읽기 중심)
- provider 상태/오류 코드/최근 동기화 시각 노출
- 기본 RBAC + audit 로그

### 6.2 Out-of-scope (후속)

- 양방향 강제 동기화(DevHub 에서 외부 시스템 본문 직접 변경)
- 복잡한 워크플로우 오케스트레이션(승인 체계, 릴리즈 체인 자동화)
- 멀티 테넌트 완전 분리
- 고급 비용/용량 예측, AI 기반 자동 최적화

## 7. 리스크 및 미해결 항목

1. 조직별 SoT 정책 충돌(Jira 중심 vs SCM 중심)의 표준 운영안 필요
2. 홈랩 수집 방식 선택(Agent Push vs 중앙 Polling) 결정 필요
3. provider 인증정보 저장/회전 정책(Secret Vault 연계) 결정 필요
4. 이벤트 유실/중복 처리(idempotency key, replay policy) 설계 필요
5. 연동 장애 시 사용자 UX(부분 장애 표기, degraded 정책) 기준 필요

## 8. 후속 단계 진입 hook

| 단계 | 산출물 | 진입 조건 |
| --- | --- | --- |
| Req sprint | REQ-FR-INT-* / REQ-NFR-INT-* 발급, SoT 정책 명문화 | 본 컨셉 머지 |
| Usecase sprint | UC-INT-* (연동 등록/동기화/장애대응/홈랩 모니터링) | Req 초안 머지 |
| Design sprint | ARCH-INT-* + API-INT-* + ERD 확장 (integration/provider/node/service/event) | Usecase 정리 |
| Implementation sprint 1 | 어댑터 프레임워크 + 2~3 provider PoC + 홈랩 상태 수집기 1종 | Design 머지 |
| Implementation sprint 2 | 운영 고도화(알림, 재동기화, 보안/권한 세분화) | 1차 운영 피드백 |

## 9. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-15 | 1차 작성 — 외부 시스템 연동 도메인 컨셉 초안. 후보 시스템군(Jira/Confluence/Bitbucket/Bamboo/Jenkins/Gitea/Forgejo) + 홈랩 인프라 관리 범위 + SoT 경계 + MVP/out-of-scope + 후속 단계 hook 정의. |
