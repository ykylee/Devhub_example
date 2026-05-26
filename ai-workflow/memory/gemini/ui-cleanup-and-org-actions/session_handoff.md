# Session Handoff - gemini/ui-cleanup-and-org-actions

## 1. 현재 상태
- **브랜치**: `gemini/ui-cleanup-and-org-actions`
- **목표**: 
  - 무의미한 UI 상태를 지닌 `Work Status`, `Quality Status`, `Sys Admin Dashboard` 화면을 사이드바 메뉴에서 아카이브 및 비노출 처리
  - Organization 설정 리스트 내 `ActionMenu` 컴포넌트를 명시적인 수정/삭제 버튼으로 교체하여 Users 관리와 동일한 UX 제공
- **현 단계**: 구현 및 빌드 검증 완료 (`done`)

## 2. 작업 완료 세부사항
- `Sidebar.tsx`에서 3개 임시 대시보드 메뉴 삭제
- `/developer/page.tsx`, `/manager/page.tsx`, `/admin/page.tsx` 각 페이지 진입 시 `/projects`로 redirect 되도록 처리 완료
- `OrgUnitTable.tsx`에서 반응하지 않던 `ActionMenu`를 제거하고 개별 `Edit3` (수정) 및 `Trash2` (삭제) 아이콘 버튼을 직접 배치하여 Users 테이블(`MemberTable.tsx`)과 동일한 일관된 디자인 제공
- Next.js Turbopack 빌드 검증 (`npm run build`) 통과 확인 완료
