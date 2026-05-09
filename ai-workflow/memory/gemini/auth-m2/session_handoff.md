# Session Handoff: M3 사용자 및 조직 관리 UI 안정화 완료

## 현재 상태 요약
- **브랜치**: `gemini/auth-m2`
- **핵심 성과**: 
    1. 인사 DB 연동 기반 셀프 가입(Sign Up) 플로우 완성.
    2. 조직도(Org Chart) 편집 및 영속화(Save) 기능 구현 완료.
    3. **조직도 UI 최적화**: 화면 번쩍거림 현상 해결, 모든 연결선 실선 표준화, 호버 액션 버튼 복원.
    4. 시스템/AI 계정 생성을 지원하는 유연한 사용자 관리(Admin User CRUD) 기능 구현.

## 진행 중인 작업 (Pending)
- **RBAC 권한 동기화**: 조직도 API(`/api/v1/organization/hierarchy`) 호출 시 발생하는 401/403 이슈 해결을 위한 백엔드 권한 정책 검토 필요.
- 마일스톤 4(실시간 대시보드) 설계를 위한 배경 조사.

## 다음 세션 권장 작업
1. **RBAC 정책 검증**: 백엔드 `internal/httpapi/permissions.go`에서 조직도 관련 권한 매핑이 실제 사용자 역할과 일치하는지 확인.
2. 마일스톤 4: `RealtimeHub`를 통한 지표 스트리밍 구현 착수.
3. `gemini/auth-m2`의 변경 사항을 `main`에 병합하기 전 최종 검토.

## 특이 사항
- 조직도 UI(`OrgNode.tsx`)에서 애니메이션 충돌 방지를 위해 `layout` 프로퍼티와 CSS `transition-all`을 제거함. 스타일 수정 시 Framer Motion의 `animate` 속성만 사용할 것을 권장.
- 현재 조직도 API는 RBAC 미들웨어에 의해 보호되고 있으며, 로컬 테스트 시 해당 권한이 올바르게 부여되었는지 확인이 필요함.
