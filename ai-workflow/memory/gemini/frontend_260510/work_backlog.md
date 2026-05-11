# 작업 백로그

## [Planned]
- [ ] 로그인 500 에러 원인 분석 (인프라 기동 후 실제 플로우 테스트)
- [ ] Kratos 아이덴티티 수동 검증 및 정합성 테스트
- [ ] 로컬 개발 환경용 RBAC 초기 데이터 시딩 확장
- [ ] RBAC 서비스 레거시 메서드 제거 및 apiClient 표준화

## [In Progress]
- [ ] Phase 6.1 통합: OIDC 표준 기반 관리자 프로비저닝 안정화

## [Done]
- [x] Kratos Admin Client `CreateIdentity` 구현
- [x] 서버 시작 시 자동 관리자 시딩 로직 (`seedLocalAdmin`) 추가
- [x] `dev-up.sh` 리스타트 및 인프라 실행 경로 수정
- [x] 로컬 PostgreSQL 연동 (계정 정보 수정 포함)
- [x] `MockKratosAdmin` 테스트 빌드 오류 수정 (필드/메서드 추적 기능 확장)
