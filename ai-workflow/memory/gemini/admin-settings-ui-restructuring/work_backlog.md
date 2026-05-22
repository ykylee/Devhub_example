# Work Backlog - Admin Settings UI Restructuring

## 📋 Backlog Items

### 1. 설계 및 준비
- [x] main 최신화 및 현황 파악
- [x] `gemini/admin-settings-ui-restructuring` 작업용 브랜치 신규 생성 및 전환
- [x] 구현 계획서(`implementation_plan.md`) 작성 및 설계 구성안 확정
- [x] 사용자 승인 및 task.md 수립

### 2. UI/UX 레이아웃 리팩토링 및 Collapsible Sidebar (layout.tsx)
- [x] 9개 메뉴의 3대 카테고리(`Access Control`, `App & Requests`, `Integrations & Audit`) 그룹화 데이터 구성
- [x] 데스크톱 뷰: 12컬럼 그리드 적용 및 좌측 세로형 사이드바 마크업 구현
- [x] 모바일 뷰: 상단 글래스모피즘 아코디언 드롭다운 셀렉터 마크업 및 열기/닫기 동작 구현
- [x] `framer-motion` 물리 스프링 이동 연출 최적화 (세로 전환 대응)
- [x] 사이드바 접기 토글 버튼 (`ChevronLeft` / `ChevronRight`) Glassmorphism 적용
- [x] 축소 시 헤더 실선화, 아이콘 중앙 정렬 및 호버 툴팁(`title` 속성) 탑재
- [x] SSR Hydration 무오류 디자인 기반 `localStorage` 상태 영속화

### 3. 검증 및 빌드
- [x] Next.js 빌드 성공 여부 검증 (`npx tsc --noEmit` 캐시 클리어 후 로컬 확인 통과)
- [x] Docker stack 배포 및 접속 URL 정합 확인
- [x] Playwright E2E 테스트 실행 및 탭 전환 및 가드 작동 여부 확인 (DB IP `172.24.0.2` & Seeder Secret 정합 적용하여 `2 passed` 확인)
- [x] PR #298과의 병합 충돌 검토, preflight 안전성 통합 및 최종 E2E 재검증 완수 (`fe7ad09`)
- [x] 최종 변경 사항 커밋 및 원격 브랜치 푸시 완료
