<!--
PR template — DevHub Example.
- 본문은 한국어가 default. 필요 시 영어 혼용.
- "카테고리·모듈" 은 `docs/governance/code-taxonomy.md` §1~2 의 명칭 사용.
- "추적성 영향" 섹션은 docs/traceability/sync-checklist.md §3.7 형식에 맞춰 채움.
- PR title prefix 도 카테고리 사용 권장 — 예: `feat(application-lifecycle): ...`.
- "Tier" 섹션은 `docs/governance/worker_division.md` §6 의 2-tier 정책에 맞춰 채움.
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

## Tier

<!--
docs/governance/worker_division.md §6 (사외/사내 2-tier 분업) 참조.
본 PR 의 push 대상 tier 명시. 다중 tier 도 가능 (예: 사외+공용).

Tier 정의:
  - 사외: GitHub main push. 사내 인프라 의존 없음.
  - 사내: 사내 SCM push. 사내 호스트/시크릿/사내 IdP 팀 SOP.
  - 공용: 양쪽 byte-identical 유지 필수 (governance/agent/추적성 ID).
-->

- [ ] 사외 (GitHub main)
- [ ] 사내 (사내 SCM)
- [ ] 공용 (양쪽 동기화)

## 사내 한정 정보 self-check (사외 PR 인 경우)

<!--
사외 PR 인 경우, 본 PR 의 변경에 다음 패턴이 포함되지 않았는지 확인. 매칭 시 자동 flag.
- DEVHUB_KEYCLOAK_*, GITEA_URL, HR_EXPORT_CMD 등 사내 env var
- internal-registry.example.com, kc.internal.example.com, devhub.example.com, sahub.example.com 등 사내 호스트
- 172.16.0.0/12, 10.x, 192.168.x 사내 IP 대역
- infrastructure/, infra/idp/, scripts/setup-keycloak.sh, docker-compose.{local,test,deploy,colima}.yml 등 사내 경로
- .env.deploy, .env.test, frontend/.env.example 의 사내 env var 추가/변경
-->

- [ ] 사내 env var 미포함
- [ ] 사내 호스트명/IP 미포함
- [ ] 사내 한정 경로 변경 없음
- [ ] 사내 한정 env 파일 변경 없음

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
