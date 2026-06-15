# Session Handoff — claude/work_260515-b

- 브랜치: `claude/work_260515-b`
- Base: `main` @ `cbc36b0` (PR #116 머지 직후)
- 날짜: 2026-05-15
- 상태: in_progress
- 목적: PR #114 가 추가한 frontend 모달/테이블 컴포넌트의 `text-primary-foreground` 직접 사용 53건을 PR #115 의 token 정책으로 sweep.

## 배경

PR #115 (b669bc7) 가 `text-foreground` / `text-foreground dark:text-primary-foreground` token 정책을 정착시켰지만, 같은 날 머지된 PR #114 (25f97ba) 의 신규 컴포넌트들이 정책 외 패턴으로 작성됨. light theme 에서 하얀 텍스트가 흰 배경 위에 출력되는 가독성 회귀 위험이 있어 즉시 sweep.

## sweep 대상

| 파일 | 건수 |
| --- | --- |
| ApplicationCreationModal.tsx | 24 |
| ProjectCreationModal.tsx | 18 |
| RepositoryLinkModal.tsx | 8 |
| ApplicationTable.tsx | 1 |
| ProjectTable.tsx | 1 |
| RepositoryTable.tsx | 1 |
| **합계** | **53** |

## sweep 정책

PR #115 가 정착시킨 패턴 그대로 적용:
- **기본 본문 / 라벨 / 입력 텍스트**: `text-foreground`
- **다크 강조가 필요한 헤딩/주요 라벨**: `text-foreground dark:text-primary-foreground`
- **보조/플레이스홀더/도움말**: `text-muted-foreground` 유지
- **`text-primary-foreground/40` 같은 opacity 변형**: `text-muted-foreground` 또는 적절한 tone token 으로 치환

판단 기준: PR #115 의 `components/organization/UserCreationModal.tsx` 가 reference (라벨은 muted-foreground, 본문 입력은 foreground/dark:primary-foreground 패턴).

## 작업 순서

1. ApplicationCreationModal sweep (24건)
2. ProjectCreationModal sweep (18건)
3. RepositoryLinkModal sweep (8건)
4. 3개 Table 파일 단건씩
5. `npx tsc --noEmit` (변경 파일) + `npx next build` 검증
6. commit + push + PR open + CI 그린 확인
7. 본인 4단계 리뷰 (diff 재검토 → 코멘트 → 필요 시 보강 → squash merge)

## 후속 (본 sprint 머지 후 다음 세션 후보)

- ADR-0011 §4.2 enforceRowOwnership helper
- critical 임계치 외부화 + Count*CriticalWarnings 성능 분리
- Repository commit activity ingest
- Project active→closed 가드 정책
- M4 RM-M4-XX 본격 진입
- traceability follow-up (TC-NAV-* row + endpoints 모듈 ARCH 한 줄)
