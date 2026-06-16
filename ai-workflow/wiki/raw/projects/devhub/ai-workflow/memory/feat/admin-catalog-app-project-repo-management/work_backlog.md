# Work Backlog - feat/admin-catalog-app-project-repo-management

이 백로그는 `feat/admin-catalog-app-project-repo-management` 피처 브랜치에서 수행된 작업 진행도 및 잔여 계획을 기록합니다.

## [Done] Sprint 1: 어드민 카탈로그 리포지토리 인식 고도화 및 자원 연관관계 개선

- [x] **백엔드 (backend-core) 고도화**
  - `PostgresStore` 내의 `ListApplicationRepositories` 메서드 쿼리 개선 (`UNION` 쿼리를 적용하여 직접 연동된 리포지토리와 소속 프로젝트에 소속된 리포지토리의 합집합을 리턴)
  - `PostgresStore` 내의 `CountActiveApplicationRepositories` 메서드 쿼리 개선 (직접 연동 및 간접 연동 리포지토리의 활성 개수 카운트 연동)
  - HTTP handler 테스트 검증용 stub인 `memoryApplicationStore`의 `ListApplicationRepositories` 및 `CountActiveApplicationRepositories` 에도 위와 동일한 성격의 모의 Parity 구현 완료
- [x] **프론트엔드 (frontend) 고도화**
  - `ProjectCreationModal.tsx` 내에서 수정 모드(`isEdit === true`)일 때 비활성화되어 있던 Application 선택 드롭다운 및 Repository 연동 select 영역 잠금 해제
  - 프로젝트 수정 시 connected repositories 변경분을 반영할 수 있도록 `unlinkProjectRepository` 및 `linkProjectRepository` API 병렬 연동 처리 구현
  - `ApplicationCreationModal.tsx` 내에 연관 Projects 및 Repositories 연결/해제 관리 UI(체크박스 목록) 컴포넌트 추가
  - 애플리케이션 정보 수정 시 변경 내역에 따라 `connectRepository`/`disconnectRepository` 및 `updateProject` API 동시 연동 갱신 구현
- [x] **검증 및 PR**
  - 백엔드 단위/통합 테스트 `go test ./...` 실행 및 모두 PASS 확인
  - 프론트엔드 linter 경고 해결 및 `npm run lint` 통과 확인
  - 원격 origin 브랜치 강제 푸시 완료 및 GitHub PR (#395) 생성 완료
