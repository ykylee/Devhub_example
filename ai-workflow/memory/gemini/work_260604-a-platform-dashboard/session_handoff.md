# Session Handoff — gemini/work_260604-a-platform-dashboard (2026-06-05)

- 문서 목적: 2026-06-04 플랫폼 대시보드 개편(프로젝트 중심 뷰) 스프린트의 작업 완료 및 2026-06-05 Codex & User PR 리뷰 피드백 반영 보완 작업에 따른 세션 인계.
- 범위: 플랫폼 대시보드 컴포넌트의 레이아웃 재배치, 지표 개편, CI 중심에서 관리/품질 중심으로의 전환, 예외 처리 UI 구축, 깨진 빌드 알림 목록 뷰 복원, API ID 충돌(API-98) 해결, 2차원 RBAC 사양 동기화.
- 대상 독자: 개발자, AI 에이전트, 운영자
- 상태: done
- 최종 수정일: 2026-06-05
- 관련 문서: [PROJECT_PROFILE.md](../../PROJECT_PROFILE.md), [work_backlog.md](./work_backlog.md)

## 1. 최근 완결된 작업
* **플랫폼 대시보드 개편 (`frontend/app/(dashboard)/platforms/[id]/page.tsx`)**:
  * **Stat Cards 지표 변경**: 기존 빌드 실행 및 성공률 통계 중심에서 활성 프로젝트 수(`Active Projects`), 플랫폼 품질 지수(`Platform Quality`), 대기 중인 DREQ 목록(`Pending Requests`), 연동된 컴포넌트(`Linked Components`)로 개편하여 플랫폼 레벨의 관심사로 초점을 맞췄습니다.
  * **Sub-Project Delivery & Roadmap 재배치**: 서브 프로젝트들의 진척률과 로드맵 카드를 화면 좌측 최상단으로 배치하여 즉각적인 로드맵 진척 파악이 가능하게 하였습니다.
  * **DREQ Backlog 영역 확장**: 플랫폼 단위에서 관리하는 요구사항 및 과제 의뢰(DREQ) 백로그 영역의 가로폭과 콘텐츠 영역을 확장하고, 'Promote Project' 버튼을 고도화하였습니다.
  * **CI 빌드 로그 뷰 제거**: 플랫폼 레벨의 혼잡도를 최소화하기 위해 타겟 브랜치 빌드 실패 결과에 종속적인 상세 빌드 로그는 대시보드 뷰에서 제거하였습니다.
  * **Empty State 폴백 처리**: backend api에 의해 `history_trend` 데이터가 제공되지 않을 경우(`[]`인 경우) 그래프 영역에 밋밋하게 비는 UI 대신 세련된 Glassmorphism 가이드 폴백 화면을 렌더링하도록 예외 처리하였습니다.
  * **Codex PR 리뷰 보완 (깨진 빌드 상태 알림 복원 - REQ-FR-APPDASH-001)**:
    * 플랫폼 Header 영역에 타겟 브랜치 빌드 헬스 지표 배지(`Healthy 🟢` / `Build Broken 🔴` / `Build Unknown ⚪`)를 추가하였습니다.
    * 연결된 리포지토리의 최신 타겟 브랜치 빌드가 깨진(`broken`) 경우, Overview stats 바로 아래에 시인성이 강한 Red Neon 펄스 애니메이션이 가미된 Alert banner를 복원하였습니다.
    * 해당 Alert banner 내부에 실패한 빌드 목록(`repo_slug`, `branch`, `build_number`, `error_snippet`)과 빌드 로그로 직접 이동할 수 있는 `Log URL` [로그 진입 딥링크] 버튼을 재배치 렌더링하여 MVP 요구사항의 퇴행을 방지하였습니다.

* **테스트 및 빌드 안정성 검증**:
  * **백엔드 테스트**: `go test ./...` 실행 결과 100% 정상 통과함을 확인하였습니다.
  * **프론트엔드 유닛 테스트**: `npm run test` 실행 결과 1015개 모든 유닛 테스트가 정상 통과함을 확인하였습니다.
  * **프론트엔드 프로덕션 컴파일**: `npm run build`를 통해 빌드가 정상적으로 완료됨을 검증하였습니다.

* **PR 발행 및 2차 리뷰 피드백 보완 완료**:
  * `opencode/work_260604-k-v1-harden`를 target base 브랜치로 하는 Pull Request [#479](https://github.com/ykylee/Devhub_example/pull/479)에 연계 반영 완료하였습니다.
  * **이슈 1: API ID 중복 충돌 해결**
    * 기존 Task Item Ingestion API의 ID `API-94`와의 중복 충돌을 해결하기 위해, 프로젝트 상세 대시보드 API ID를 신규 미할당 ID인 **`API-98`**로 개칭하였습니다.
    * 관련 문서([api.md](../../../../docs/domain/platform-lifecycle/api.md), [report.md](../../../../docs/traceability/report.md)) 전역에서 프로젝트 대시보드 연계 API ID를 `API-98`로 동기화하였습니다.
  * **이슈 2: 대시보드 UI 개편과 요구사항/유스케이스 문서 불일치 정합성 보완**
    * 플랫폼 대시보드 최상단 카드에서 빌드 헬스가 빠지고 다른 성격의 메트릭(Active Projects 등)이 배치된 UI 변동에 맞춰, 요구사항([requirements.md](../../../../docs/domain/platform-lifecycle/requirements.md) REQ-FR-APPDASH-001) 및 시스템 유스케이스([system_usecases.md](../../../../docs/planning/system_usecases.md) UC-APPDASH-01)의 문구를 갱신하여 현실 UI 명세(대시보드 상단 헬스 배지 및 broken build 전용 Alert 배너 영역)와 정합성을 일치시켰습니다.
  * **이슈 3: PROJDASH 문서의 역할 매핑 2D RBAC 동기화**
    * 기존의 개념 설계, 요구사항, 아키텍처 문서에서 Keycloak의 단방향 legacy 역할명(`contributor` 등)을 언급하던 부분을 2차원 RBAC 표준 명칭(System Role: `developer/team_manager/system_admin` 및 Resource Role: `project_member/project_leader`)에 맞춰 명확하게 정정하였습니다.
    * `project_leader`의 식별 기준은 토큰 클레임이 아닌 `project_members.project_role = 'lead'` 데이터에 있음을 아키텍처 및 요구사항에 명문화하여 구현 정합성을 확보하였습니다.

* **프로젝트 상세 대시보드(PROJDASH) 컨셉, 요구사항, 유스케이스, 아키텍처 및 API 명세 정립**:
  * **3대 페르소나별 뷰 모델 설계**: 개발자(My Work/빌드 확인), 프로젝트 리더(PR 통합/Stale/Blocker 중재), 조직 관리자(인력 부하/Forecast/거버넌스)의 페르소나별 최적화 뷰 스펙 도출.
  * **컨셉 상세 설계서 작성**: [project_dashboard_concept.md](../../../../docs/domain/platform-lifecycle/project_dashboard_concept.md) 파일을 신규 생성하여, 3단 스위처(3-Way Persona Switcher) 및 HSL Red Neon Indicator 등의 UI/UX 인터랙션 스펙 상세화 기술.
  * **요구사항 갱신**: [requirements.md](../../../../docs/domain/platform-lifecycle/requirements.md) 파일에 `REQ-FR-PROJDASH-001..006` 및 `REQ-NFR-PROJDASH-001..002` 신규 요구사항 8건을 최종 등록하고 이력 로그 업데이트.
  * **시스템 유스케이스 추가**: [system_usecases.md](../../../../docs/planning/system_usecases.md) 파일에 `UC-PROJDASH-01`부터 `06`까지의 유스케이스를 추가하고 상세 흐름 시나리오(Actor, 사전조건, 기본/예외 흐름, 사후조건)를 구체화하여 명세화.
  * **도메인 아키텍처 정의 추가**: [architecture.md](../../../../docs/domain/platform-lifecycle/architecture.md) 파일에 `ARCH-PROJDASH-01..03`을 추가하여 3단 스위처의 OIDC/RBAC 2차 가드 제어, Gitea 병열 스캔 PR 분석, 마감 지연 리스크($R_{\text{SLA}}$) 및 멤버 부하($L_u$) 산정 수식 모델 규정.
  * **API 계약 추가**: [api.md](../../../../docs/domain/platform-lifecycle/api.md) 파일에 `API-98 (GET /api/v1/projects/{project_id}/dashboard)` 명세를 추가하고 OIDC/RBAC 인증, 쿼리 매개변수(`persona`) 분기 및 페르소나별 다형적 응답 예시 JSON 구조화.
  * **원격 푸시 준비**: 모든 문서 정합성 및 링크 무결성 검증을 완료하고 Pull Request [#479](https://github.com/ykylee/Devhub_example/pull/479)에 연계 푸시 완료.

## 2. 다음 세션 참고 정보
* UI 레이아웃 리팩토링, 폴백 뷰 처리, 깨진 빌드 경고 알림 복원 및 API ID/RBAC 정합성 보완은 로컬 검증 및 유닛 테스트 상으로 모두 완결되었습니다.
* 백엔드의 `history_trend` 일별 집계 및 `progress_percent` 스토리포인트 기반 실 계산 로직은 v1.1 혹은 v2 스프린트의 후속 백엔드 고도화 태스크로 이어집니다.
