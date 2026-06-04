# Work Backlog — gemini/work_260604-a-platform-dashboard

- 문서 목적: 2026-06-04 플랫폼 대시보드 개편 스프린트의 작업 백로그 명시
- 범위: 플랫폼 대시보드 뷰 개편 및 로컬 테스트 검증 완결
- 대상 독자: 개발자, AI 에이전트, 운영자
- 상태: active
- 최종 수정일: 2026-06-05
- 관련 문서: [PROJECT_PROFILE.md](../../PROJECT_PROFILE.md), [session_handoff.md](./session_handoff.md)

이 문서는 `gemini/work_260604-a-platform-dashboard` 브랜치 세션 동안 검토하거나 진행한 작업 백로그를 명시합니다.

## 1. 현재 스프린트 작업
- [x] 플랫폼 대시보드 레이아웃을 빌드 중심에서 플랫폼/프로젝트 관리 및 품질 지표 중심으로 개편 완료 (2026-06-04)
- [x] 빌드 결과 상세 로그 뷰 제거 및 빈 히스토리 그래프 대응용 Empty State 폴백 구축 완료 (2026-06-04)
- [x] 로컬 backend 및 frontend 테스트 스위트 정상 통과 검증 및 빌드 확인 완료 (2026-06-04)
- [x] 원격 PR [#479](https://github.com/ykylee/Devhub_example/pull/479) 생성 완료 (2026-06-04)
- [x] 프로젝트 대시보드 3대 페르소나별 컨셉 설계, 요구사항 명세 및 유스케이스 정의 완료 (2026-06-04)
- [x] Codex 자동화 PR 리뷰 보완 (깨진 빌드 상태 알림 복원 및 실패 런 정보 + 로그 URL 딥링크 뷰 복원 - REQ-FR-APPDASH-001 충족) 완료 (2026-06-05)
- [x] User 2차 리뷰 피드백 보완 (API-98 개칭을 통한 ID 중복 충돌 해결, 변경된 대시보드 UI 레이아웃에 맞춤 요구사항 및 유스케이스 갱신, PROJDASH 문서의 역할 2차원 RBAC 표준 동기화) 완료 (2026-06-05)

## 2. 잔여 로드맵 연계 백로그 (향후 후속 Task)
* **P1-2/BE**: `history_trend` 데이터가 제공되도록 일별 메트릭 집계(build_runs / quality_snapshots) 백엔드 크론 로직 구축 (Claude/OpenCode 협업)
* **P2/BE**: `projects_progress.progress_percent`가 하드코딩 `0`이 아닌 연동된 Task Item(Jira/Gitea 등)의 스토리 포인트 기반으로 실시간 계산되도록 백엔드 비즈니스 로직 연계
* **P2/FE**: Keycloak 사용자 display name의 Name Resolver 어댑터를 통한 leader/assignee 노출 고도화
