<!--
PR template — DevHub Example.
- 본문은 한국어가 default. 필요 시 영어 혼용.
- "카테고리·모듈" 은 `docs/governance/code-taxonomy.md` §1~2 의 명칭 사용.
- "추적성 영향" 섹션은 docs/traceability/sync-checklist.md §3.7 형식에 맞춰 채움.
- PR title prefix 도 카테고리 사용 권장 — 예: `feat(application-lifecycle): ...`.
-->

## 카테고리 · 모듈

<!--
docs/governance/code-taxonomy.md §1~2 참조. 영향 카테고리/모듈 명시.
다중 카테고리는 줄바꿈 또는 콤마.
예:
  - `application-lifecycle` / store/applications, httpapi/applications
  - `auth-session` / lib/auth/refresh-scheduler
-->

- 카테고리:
- 모듈:

## 요약

<!-- 1-3 줄로 본 PR 의 변경 의도 + 영향 범위. -->

## 변경 상세

<!-- 파일별 / 모듈별 변경 요지. 큰 PR 은 섹션 분리. -->

## 추적성 영향

<!--
docs/traceability/sync-checklist.md §3.7 참조.
영향 없으면 "N/A — 의미 변경 없는 리팩토 / 문서 정리" 로 채움.
-->

- 추가:
- 갱신:
- Deprecate:
- 매트릭스: `docs/traceability/report.md` 갱신 사항

## 테스트

<!--
- [x] 로컬 backend `go test ./...`
- [x] 로컬 frontend `npm run test`
- [ ] CI: backend-unit / frontend-unit / e2e 3 잡 SUCCESS
- [ ] 수동 검증 (필요 시 절차 명기)
-->

## 관련 issue / ADR

<!-- 있다면 링크. ADR 추가/변경 시 docs/adr/000X-*.md 명시. -->
