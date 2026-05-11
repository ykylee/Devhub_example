# 세션 인계 문서 (2026-05-11)

## 현재 상태
- **완료**: Kratos Admin API 연동, 서버 시작 시 관리자(`test`) 자동 시딩, `dev-up.sh` 인프라 실행 스크립트 안정화.
- **진행 중**: 로그인 시 발생하는 500 에러 원인 분석 및 해결.
- **인프라**: Kratos(4433), Hydra(4444), Backend(8080), Frontend(3000)가 로컬에서 구동 중. DB는 `postgres://yklee@localhost:5432/devhub` 연결 확인.

## 다음 세션 작업 제안
1. **500 에러 디버깅**: 로그인 요청 시 백엔드 로그(`backend.log`)에서 `Internal Server Error`의 상세 스택 트레이스 확인.
2. **Kratos 세션 검증**: Kratos Public API를 통해 `test` 계정의 아이덴티티가 정상적으로 조회되는지 확인 (`kratos identities get`).
3. **E2E 테스트 재수행**: 에러 해결 후 가상 브라우저를 통해 최종 로그인 성공 여부 검증.

## 핵심 컨텍스트
- 모든 사용자 생성은 `Kratos Admin` -> `OrganizationStore.CreateUser` -> `SetKratosIdentityID` 순서로 진행되어야 함.
- `dev-up.sh`는 루트에서 실행되며 각 서비스의 작업 디렉토리를 올바르게 관리함.
