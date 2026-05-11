# 세션 인계 문서 (2026-05-12)

## 현재 상태
- **완료**: `MockKratosAdmin` 테스트 빌드 오류 수정. 전체 백엔드 테스트(`go test ./...`) 및 프론트엔드 빌드(`next build`) 통과.
- **진행 중**: 로그인 시 발생하는 500 에러 원인 분석 (Kratos/Hydra 인프라 기동 후 실제 검증 필요).
- **인프라**: Kratos(4433), Hydra(4444), Backend(8080), Frontend(3000) 로컬 구동 구성 유지.

## 이번 세션 수정 내역
- `kratos_admin_client.go`의 `MockKratosAdmin`에 `PasswordResets`, `StateChanges`, `DeletedIDs`, `FindError`, `FindCalls`, `FindIDOverride` 필드 추가.
- Mock ID 접두사를 `mock-id-`에서 `mock-k-id-`로 통일하여 테스트 assertion 일치.
- 모든 Mock 메서드에 호출 추적/에러 주입 로직 구현.

## 다음 세션 작업 제안
1. **500 에러 디버깅**: `dev-up.sh`로 인프라 기동 후 실제 로그인 플로우 테스트. 백엔드 로그에서 상세 스택 트레이스 확인.
2. **Kratos 세션 검증**: `test` 계정 아이덴티티 정상 조회 확인.
3. **RBAC 서비스 정비**: `rbac.service.ts` 레거시 메서드 제거 및 `apiClient` 표준화.

## 핵심 컨텍스트
- 모든 사용자 생성은 `Kratos Admin` → `OrganizationStore.CreateUser` → `SetKratosIdentityID` 순서로 진행.
- `dev-up.sh`는 루트에서 실행되며 각 서비스의 작업 디렉토리를 올바르게 관리.
