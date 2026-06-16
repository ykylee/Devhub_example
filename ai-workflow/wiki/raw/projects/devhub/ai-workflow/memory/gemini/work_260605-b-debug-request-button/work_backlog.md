# Work Backlog — gemini/work_260605-b-debug-request-button

- 문서 목적: 의뢰 수신 기능 테스트를 위한 디버그 버튼 구현 스프린트의 작업 백로그 명시.
- 범위: 백엔드 인증 우회 추가, 프론트엔드 API 추가, UI 연동, 그리고 빌드/단위 테스트 검증.
- 대상 독자: 개발자, AI 에이전트, 운영자
- 상태: done
- 최종 수정일: 2026-06-05
- 관련 문서: [PROJECT_PROFILE.md](../../PROJECT_PROFILE.md), [session_handoff.md](./session_handoff.md)

이 문서는 `gemini/work_260605-b-debug-request-button` 브랜치 스프린트의 작업 백로그 상태를 기록합니다.

## 1. 스프린트 작업
- [done] 원격 main 최신 상태 로컬 main에 동기화 완료 (2026-06-05)
- [done] gemini/work_260605-b-debug-request-button 피처 브랜치 분기 완료 (2026-06-05)
- [done] 백엔드 RequireIntakeToken 미들웨어에 DevFallbackEnabled 우회(더미 컨텍스트 삽입) 추가 완료 (2026-06-05)
- [done] 프론트엔드 devRequestService에 createDebugDreq API 추가 완료 (2026-06-05)
- [done] Header.tsx 내 개발 모드 전용 Debug DREQ 버튼 UI 구현 및 연동 완료 (2026-06-05)
- [done] 백엔드 단위 테스트(go test) 및 프론트엔드 빌드(npm run build) 검증 완료 (2026-06-05)

## 2. 잔여 로드맵 연계 백로그
* **N-3**: SCM import/create + draft/publish happy-path E2E 검증 보강
* **P0-3**: Playwright screenshot mode 도입 연계
