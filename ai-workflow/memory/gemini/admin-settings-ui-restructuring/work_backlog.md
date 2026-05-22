# Work Backlog - Admin Settings UI Restructuring

## 📋 Backlog Items

### 1. 설계 및 준비
- [x] main 최신화 및 현황 파악
- [x] `gemini/admin-settings-ui-restructuring` 작업용 브랜치 신규 생성 및 전환
- [x] 구현 계획서(`implementation_plan.md`) 작성 및 설계 구성안 확정
- [ ] 사용자 승인 및 task.md 수립

### 2. UI/UX 레이아웃 리팩토링 (layout.tsx)
- [ ] 9개 메뉴의 3대 카테고리(`Access Control`, `App & Requests`, `Integrations & Audit`) 그룹화 데이터 구성
- [ ] 데스크톱 뷰: 12컬럼 그리드 적용 및 좌측 세로형 사이드바 마크업 구현
- [ ] 모바일 뷰: 상단 글래스모피즘 아코디언 드롭다운 셀렉터 마크업 및 열기/닫기 동작 구현
- [ ] `framer-motion` 물리 스프링 이동 연출 최적화 (세로 전환 대응)

### 3. 검증 및 빌드
- [ ] Next.js 빌드 성공 여부 검증 (`npm run build` 또는 `npx tsc --noEmit`)
- [ ] Playwright E2E 테스트 실행 및 탭 전환 및 가드 작동 여부 확인
- [ ] 변경 사항 커밋 및 푸시
