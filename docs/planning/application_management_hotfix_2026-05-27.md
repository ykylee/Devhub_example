# Application 관리 Hotfix 계획 (2026-05-27)

- 문서 목적: Application 등록/수정 UI에서 확인된 오류와 개선 요구를 구현 전 합의 가능한 작업 단위로 정리한다.
- 범위: Application 등록 다이얼로그 + 관련 API 검증 + 검색 UX + 스타일 이슈
- 대상 독자: 구현 담당자, 리뷰어, QA
- 상태: draft
- 최종 수정일: 2026-05-27

## 1) 관찰된 이슈

1. Application 등록 실패
- 증상: `key must match ...` 오류로 등록 실패
- 재현 입력: `key=DEVHUB`
- 확인 결과: 백엔드 검증 정규식이 `^[A-Za-z0-9]{10}$` (정확히 10자)로 고정되어 있어 6자 key가 거부됨
- 프론트 입력 안내는 `maxLength=10`/패턴 `1~10자`라 백엔드와 규칙 불일치

2. Key 중복 체크
- 현재 상태: 서버 측 중복 방어 존재
- 근거: `CreateApplication`에서 `store.ErrConflict` 발생 시 `409 application_key_conflict` 반환
- 결론: 기능은 존재하나, 사용자 입장에서는 제출 후에만 인지 가능

3. Application Leader / Development Department 입력 UX
- 현재: 단순 텍스트 입력
- 요구: 키워드 검색 기반 선택 가능해야 함

4. Owner User (Legacy) 필드
- 현재: 폼에서 필수 입력
- 요구: 더 이상 사용하지 않는 항목 제거

5. Operating status 드롭다운 라이트 모드 가독성
- 현재: option 배경/글자색 조합 문제로 검은 배경+검은 글씨가 발생

## 2) 구현 방향

1. key 규칙 정합
- 백엔드 `Application create` 검증을 `1~10자 영문/숫자`로 완화
- 오류 메시지/테스트를 새 규칙에 맞게 갱신
- Application 등록 폼 안내 문구를 백엔드 규칙과 동일하게 수정

2. 중복 체크 UX 보강
- 서버 409 중복 체크를 단일 source-of-truth 로 사용
- 프론트는 사전 목록 비교를 하지 않고, 생성 API 의 409 응답 시 중복 메시지를 노출

3. Leader/Department 검색형 입력
- `ComboBox` 적용
- Leader: `/api/v1/users` 목록 기반 검색
- Development Department: `/api/v1/organization/hierarchy` 노드 기반 검색

4. Legacy Owner 제거
- 폼 UI에서 Owner 입력 제거
- 생성 payload의 `owner_user_id`는 `leader_user_id`를 기본값으로 동기화하여 기존 API 요구사항과 호환
- 수정(PATCH)에서도 `owner_user_id`를 `leader_user_id`와 동기화하여 row ownership drift 방지

5. Operating status 스타일 수정
- option 강제 다크 배경 클래스 제거
- 라이트/다크 모두 대비가 유지되도록 기본 테마 색상 사용

## 3) 변경 예상 파일

- `backend-core/internal/httpapi/applications.go`
- `backend-core/internal/httpapi/applications_test.go`
- `backend-core/migrations/000035_relax_applications_key_format.{up,down}.sql`
- `frontend/components/project/ApplicationCreationModal.tsx`
- `frontend/app/(dashboard)/admin/settings/organization/page.tsx`
- `frontend/components/organization/MemberTable.tsx`

## 4) 검증 계획

- Backend: `cd backend-core && go test ./internal/httpapi -run Application`
- Frontend: `cd frontend && npm run lint`
- 수동 확인:
  - key=`DEVHUB` 등록 성공
  - 동일 key 재등록 시 중복 메시지 확인
  - Leader/Department 검색 선택 동작
  - Owner 입력 필드 미노출
  - Operating status 드롭다운 라이트 모드 가독성 확인
