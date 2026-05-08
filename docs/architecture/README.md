# Architecture Documentation

- 문서 목적: `docs/architecture/` 디렉터리의 진입점. 시스템 설계 자료를 종류별로 안내한다.
- 범위: 시스템 전체 설계 본문, ADR (Architecture Decision Records), 컴포넌트별 설계 자료, 통합 로드맵의 아키텍처 측 매핑
- 대상 독자: 개발자, 설계자, 후속 ADR 작성자
- 상태: stable
- 최종 수정일: 2026-05-08 (TBD 스텁 → 인덱스 채움)
- 관련 문서: [../README.md](../README.md), [../development_roadmap.md](../development_roadmap.md), [../PROJECT_PROFILE.md](../PROJECT_PROFILE.md)

## 0. 자료의 위치

DevHub 의 아키텍처 자료는 다음 4 영역에 분산되어 있다.

| 영역 | 위치 | 역할 |
| --- | --- | --- |
| **시스템 전체 설계 본문** | [`../architecture.md`](../architecture.md) | 컴포넌트, 통신 방식, 데이터 흐름, 보안 모델, UI/UX 전략 등 — 아키텍처의 *narrative* source-of-truth |
| **ADR (Architecture Decision Records)** | [`../adr/`](../adr/) | 단일 결정 단위의 영구 기록. 결정 시점·근거·대안·결과를 보존 |
| **컴포넌트별 설계 자료** | 본 디렉터리 (`docs/architecture/`) | 시스템 본문에 통합되기 전의 컴포넌트별 설계 초안 / 보조 자료 (현재 비어 있음) |
| **API 계약 / 데이터 모델** | [`../backend_api_contract.md`](../backend_api_contract.md), [`../backend_requirements_org_hierarchy.md`](../backend_requirements_org_hierarchy.md), [`../organizational_hierarchy_spec.md`](../organizational_hierarchy_spec.md) | API · 데이터 형식 — 아키텍처에서 파생된 contract |

> 일반 개발자는 본문([`../architecture.md`](../architecture.md))과 통합 로드맵([`../development_roadmap.md`](../development_roadmap.md))만 읽으면 충분하다. 본 README 는 자료 위치 안내와 ADR 인덱스다.

## 1. 시스템 전체 설계

[`../architecture.md`](../architecture.md) 를 source-of-truth 로 본다. 주요 섹션:

| § | 내용 |
| --- | --- |
| §2 시스템 컴포넌트 구조 | Frontend / Backend Core / Backend AI / Data / IdP / External 의 mermaid 다이어그램과 상태(`current` / `planned` / `external`) |
| §3 서비스 간 통신 | REST snapshot + WebSocket (브라우저↔서버), gRPC (Go Core ↔ Python AI), Webhook (Gitea → Go Core) |
| §4 데이터 전략 | Webhook 정규화, Hourly Pull reconciliation, 충돌 해소 정책, 보존 정책 |
| §5 UI/UX 시각화 | 역할별 대시보드, 실시간 피드백 |
| §6 보안 모델 | RBAC 단계화 (Phase 1~4), 인증 경계, audit 정책 |

## 2. ADR 인덱스

[`../adr/`](../adr/) 하위 단일 결정 기록.

| 번호 | 제목 | 상태 | 결정 요약 |
| --- | --- | --- | --- |
| [ADR-0001](../adr/0001-idp-selection.md) | IdP 선택 | accepted (2026-05-07) | DevHub 자체 accounts 폐기, **Ory Hydra (OAuth 2.0 / OIDC)** + **Ory Kratos (identity)** 도입. Kratos 가 credential master, `users` 테이블은 organizational metadata. PostgreSQL schema 분리(`hydra`, `kratos`), first-party silent consent, 1차 MFA 미도입, Gitea SSO 는 별도 ADR-0002 예정. |

향후 ADR 후보 (통합 로드맵 §3.5 M4 진입 시 작성):

- **ADR-0002** Gitea SSO 통합 (RBAC Phase 4)
- **ADR-0003** commits 정규화 테이블 도입 여부 (M3 진입 시 결정)
- **ADR-0004** OS service wrapper (NSSM / sc / systemd) 운영 진입 시점 결정

## 3. 통합 로드맵 매핑

본 디렉터리의 자료는 [통합 개발 로드맵](../development_roadmap.md) 의 트랙·마일스톤과 다음과 같이 연결된다.

| 통합 로드맵 트랙 | 본 디렉터리에서 다루는 자료 |
| --- | --- |
| **A — Auth & IdP** | [ADR-0001](../adr/0001-idp-selection.md). 향후 ADR-0002 (Gitea SSO) 가 추가 예정. 본문 [`../architecture.md`](../architecture.md) §6 RBAC 단계화. |
| **X — Cross / Contract** | API envelope / role wire / command lifecycle 표준은 [`../backend_api_contract.md`](../backend_api_contract.md). WebSocket envelope, AccountStatus invariant, 데이터 보존 정책 등은 [`../architecture.md`](../architecture.md) §4·§6. |
| **B — Backend** | [`../architecture.md`](../architecture.md) §2 컴포넌트, §3 통신, §4 데이터. 백엔드 트랙 세부는 [`../../ai-workflow/memory/backend_development_roadmap.md`](../../ai-workflow/memory/backend_development_roadmap.md). |
| **F — Frontend** | [`../architecture.md`](../architecture.md) §5 UI/UX. 프론트엔드 트랙 세부는 [`../frontend_development_roadmap.md`](../frontend_development_roadmap.md). |
| **AI** | [`../architecture.md`](../architecture.md) §2-Backend AI / §3 gRPC. 1차 PoC 는 통합 로드맵 M3 에 진입. |
| **O — Operations** | [`../setup/environment-setup.md`](../setup/environment-setup.md) (docker / native 분기). 컨테이너 자산은 git 추적 외부. |

## 4. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-04-30 | 초기 TBD 스텁 생성 |
| 2026-05-08 | TBD 스텁 → 디렉터리 진입점 + ADR 인덱스 + 통합 로드맵 매핑 (claude/merge_roadmap 브랜치) |

## 5. 신규 자료 작성 규칙

본 디렉터리에 신규 컴포넌트별 설계 문서를 추가할 때:

1. 파일명은 영문 소문자 + `_` 또는 `-` (예: `commandworker_design.md`).
2. 본문 시작에 표준 frontmatter (목적/범위/대상/상태/최종 수정일/관련 문서) 를 포함한다.
3. *결정* 단위의 영구 기록은 본 디렉터리가 아닌 [`../adr/`](../adr/) 에 ADR 형태로 작성한다.
4. 추가 후 본 README 의 §3 매핑 표를 갱신한다.
