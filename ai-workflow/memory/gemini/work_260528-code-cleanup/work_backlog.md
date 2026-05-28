# Work Backlog: 대칭형 폴더 대개편 (Stage 3 & Stage 4)

## 1. 남아있는 Sprint 목표 (Remaining Tasks)

### 1.1 [Stage 3] Next.js Frontend Domain별 격리 및 컴포넌트 이관 (Thin Shell 라우터화)
*   **목적**: 프론트엔드의 컴포넌트 및 서비스들도 백엔드처럼 도메인 단위로 수직 이관하여 대칭 구조 완성.
*   **체크리스트**:
    - [ ] `frontend/domain/` 하위 10대 도메인별 폴더 및 서브 계층 (`view/`, `service/`, `schema/`) 신설
    - [ ] 기존 flat 디렉토리 (`components/`, `lib/`) 내 UI 컴포넌트, 훅, DTO 타입, API 호출 함수들을 해당 도메인 컴포넌트로 분리 및 이관
    - [ ] Next.js `app/` 라우팅 폴더의 페이지 파일들을 얇은 렌더링 쉘(Thin Shell)로 정비 및 임포트 경로 전수 교정
    - [ ] Frontend 컴파일 타입 검증 (`npx tsc --noEmit`) 및 Next.js 프로덕션 빌드 (`npm run build`) 무결성 검증

### 1.2 [Stage 4] docs/ 설계/거버넌스 문서 대칭 이관 및 재구성
*   **체크리스트**:
    - [ ] `docs/domain/`, `docs/shared/`, `docs/infrastructure/` 폴더 신설
    - [ ] 기존 마크다운 문서들을 해당하는 도메인/레이어 폴더로 이관 재분류
    - [ ] `docs/README.md` 및 `docs/architecture.md` 종속성 설계서 및 물리 폴더 구조 최신화
    - [ ] 문서 내 상대 링크 정합성 전수 검사 및 수정

### 1.3 최종 통합 검증 및 원격 반영
*   **체크리스트**:
    - [ ] 전체 빌드 및 연동 무결성 최종 검증 (Go Backend + Next.js Frontend + E2E Tests)
    - [ ] 커밋 및 원격 PR 반영 (`git push origin gemini/work_260528-code-cleanup`)

## 2. 우선순위
1.  **Stage 3**: Next.js Frontend Domain 격리 (P0)
2.  **Stage 4**: docs/ 마크다운 문서 대칭 이관 (P1)
3.  **최종 검증**: 빌드 및 E2E 테스트 검증 및 원격 PR 반영 (P0)
