# Work Backlog - gemini/gitea-integration-enhancement

- [x] [PLANNING] Gitea 연계 기능 고도화 구현 계획 수립 및 승인 받기
- [x] [RESEARCH] 홈랩 Gitea API 호출 테스트 및 기존 SCM 동기화 로직 구조 분석
- [x] [BE] Gitea SCM 동기화 로직 및 Sync Worker 구현 및 main.go 와이어링 완료
- [x] [BE] integration_sync_jobs 큐잉 폴링 및 상태 관리 핸들러 DB Store 연계 완료
- [x] [FE/BE] Gitea 연동 설정(인증토큰 등) 관리 UI/API 보완 (has_credentials 마운트 및 API 응답 패치) 완료
- [ ] [VERIFY] 홈랩 Gitea Webhook 및 Sync 모듈 실환경 수동/E2E 검증 수행 (타 작업 서비스 기동 후 검증 예정으로 대기)
