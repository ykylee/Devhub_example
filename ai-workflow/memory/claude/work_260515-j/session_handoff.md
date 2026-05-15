# Session Handoff — claude/work_260515-j (DREQ-Frontend)

- 브랜치: `claude/work_260515-j`
- Base: `main` @ `333edc9`
- 날짜: 2026-05-15
- 상태: in_progress
- 목적: DREQ 도메인 frontend 1차 — 담당자 dashboard 위젯 + 관리 페이지 + 상세 modal. Backend API-59..65 와 연계.

## 작업 분해

1. (in progress) 메모리 초기화
2. (planned) lib/services/dev_request.types.ts + .service.ts (7 메서드)
3. (planned) /admin/settings/dev-requests page + sidebar entry
4. (planned) DevRequestTable + DevRequestDetailModal
5. (planned) MyPendingDevRequestsWidget — developer/manager dashboard
6. (planned) tsc + next build + PR + CI + self-review + merge

## 패턴 reference (sprint b/c/d 정착)

- `lib/services/project.service.ts` — apiClient 패턴 + endpoints 모듈
- `components/organization/UserCreationModal.tsx` — token 정책 (text-foreground / dark:primary-foreground)
- `app/(dashboard)/admin/settings/applications/page.tsx` — 목록 페이지 패턴
- `components/project/ApplicationCreationModal.tsx` — 모달 패턴 (단, edit 시 immutable key 제외 PR #119 hotfix 패턴)
- `parseISO` 사용 (date-only fields)

## carve out

- intake token 발급 UI → DREQ-Admin-UI
- Promote 단일 트랜잭션 → DREQ-Promote-Tx
- e2e + Vitest → DREQ-E2E
