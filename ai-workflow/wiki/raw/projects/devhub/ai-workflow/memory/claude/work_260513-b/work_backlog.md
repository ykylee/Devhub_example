# Work Backlog — claude/work_260513-b

- 문서 목적: 본 sprint 의 작업 목록과 진행 상태.
- 최종 수정일: 2026-05-13

## 진행 중

- [x] ADR-0003 작성 (`docs/adr/0003-no-docker-policy-ci-scope.md`)
- [x] `.github/workflows/ci.yml` 의 `services: postgres:15` 블록 제거 + `Setup PostgreSQL (native)` step 추가
- [x] `docs/setup/test-server-deployment.md` 에 ADR-0003 링크 + §0 정책 줄 갱신
- [x] `ai-workflow/memory/state.json` 의 ci_followups 에서 FU-CI-1 제거 + CI 마일스톤 노트 보강
- [ ] PR 생성 + CI 그린 확인 + 리뷰어 모드 2-pass + 머지 지시 대기

## 미진입 (다음 sprint 후보, ADR-0003 §6 미해결 항목 포함)

- [ ] actionlint / workflow 린트 도입 — `services:` 사용을 정적 검사로 잡는 게이트. 별도 ADR 후보.
- [ ] 사내 self-hosted runner 도입 결정 — pgdg repo mirror 또는 사전 설치 이미지 검토.
- [ ] caller-supplied X-Request-ID validation (정규식 강제). work_260512-j 발견.
- [ ] ctx 표준 request_id 전파.
- [ ] writeRBACServerError → writeServerError 통합 (`rbac.go:22`).
- [ ] M4 진입 후보군.

## 완료

- (PR 머지 후 본 sprint 의 결정/구현 항목들이 여기로 이동.)
