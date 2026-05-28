# Work Backlog: 대칭형 폴더 대개편 (Stage 4)

## 1. 남아있는 Sprint 목표 (Remaining Tasks)

### 1.1 [Stage 4] docs/ 설계/거버넌스 문서 대칭 이관 및 재구성
*   **목적**: 문서와 거버넌스 마크다운 자료들 역시 3대 레이어 및 도메인 단위로 수직 이관하여 대칭 구조 완성.
*   **체크리스트**:
    - [ ] `docs/domain/`, `docs/shared/`, `docs/infrastructure/` 폴더 신설
    - [ ] 기존 마크다운 문서들을 해당하는 도메인/레이어 폴더로 이관 재분류
    - [ ] `docs/README.md` 및 `docs/architecture.md` 종속성 설계서 및 물리 폴더 구조 최신화
    - [ ] 문서 내 상대 링크 정합성 전수 검사 및 수정

### 1.2 최종 통합 검증 및 원격 반영
*   **체크리스트**:
    - [ ] 전체 빌드 및 연동 무결성 최종 검증 (Go Backend + Next.js Frontend + E2E Tests)
    - [ ] 커밋 및 원격 PR 반영 (`git push origin gemini/work_260528-code-cleanup`)

## 2. 우선순위
1.  **Stage 4**: docs/ 마크다운 문서 대칭 이관 (P0)
2.  **최종 검증**: 빌드 및 E2E 테스트 검증 및 원격 PR 반영 (P0)
