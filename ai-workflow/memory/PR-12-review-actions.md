# PR #12 종합 리뷰 액션 플랜

- 문서 목적: PR #12 (`test/backend-integration` → `main`) 에 대한 self-review, Codex 자동 리뷰, claude code 종합 리뷰를 통합하여 머지 전 처리해야 할 수정 사항을 한 곳에 정리한다.
- 범위: PR #12 코멘트 종합, 우선순위 분류, 검증 체크리스트
- 대상 독자: PR 작성자(ykylee), 리뷰어, 다음 세션 담당자
- 상태: draft
- 최종 수정일: 2026-05-08
- 관련 문서: PR https://github.com/ykylee/Devhub_example/pull/12, `ai-workflow/memory/state.json`, `ai-workflow/memory/session_handoff.md`, `CLAUDE.md`

## 1. 입력 자료 출처

| 출처 | 작성자 | 일시 | 결론 |
| --- | --- | --- | --- |
| Self review (issue comment 4398244076) | ykylee | 2026-05-07 | Approve, Next steps 3건 제안 |
| Codex review (review 4245103190) | chatgpt-codex-connector[bot] | 2026-05-07 | Comment, P1 1건 + P2 1건 |
| Codex line comment 3202430736 | bot | 2026-05-07 | P1 — `organization.go:416` audit 실패 시 500 |
| Codex line comment 3202430740 | bot | 2026-05-07 | P2 — `next.config.ts:3` localhost fallback 제거 |
| 종합 review (issue comment 4401940017) | claude code | 2026-05-08 | Comment, 블로커 3건 외 다수 |

## 2. PR 메타

- 규모: 353 files, +32,087 / -3,187, commits 8
- mergeable: CLEAN, review decision 없음
- CI: `gh pr checks 12` → no checks reported (워크플로우 미가동)

## 3. 액션 항목

### 3.1 블로커 — 머지 전 반드시 처리

| ID | 위치 | 문제 | 권장 조치 | 상태 |
| --- | --- | --- | --- | --- |
| BLK-1 | 정책 | 사용자 메모리 `feedback_no_docker.md` 와 PR 의 docker 우선 방향이 충돌 | **결정 (2026-05-08): 컨테이너 자산은 git 추적 외부, git 에는 docker/native 분기 가이드만.** 적용 — `.gitignore` DEV ENVIRONMENT 섹션 추가, `docker-compose.yml` 및 3개 `Dockerfile` `git rm --cached`, `docs/setup/environment-setup.md` 신규, README/`docs/README.md` 진입점 추가, `Makefile` `build`/`run` 가이드 위임, `CLAUDE.md` 실행 기본값 갱신, 메모리(`feedback_no_docker.md`) How to apply 보강. | resolved |
| BLK-2 | `frontend/next.config.ts:3` | `BACKEND_API_URL` 미설정 시 기본값이 `http://backend-core:8080` (compose 내부 호스트명) → native 환경 DNS 실패 (Codex P2) | **적용 (2026-05-08): default 를 `http://localhost:8080` 로 복원.** 영향 검토 — `BACKEND_API_URL` 은 next.config.ts 1곳만, client-side 서비스(`rbac`/`realtime`/`websocket`)는 이미 `NEXT_PUBLIC_*` default 가 localhost 라 일관. compose 모드 사용자는 가이드 §3 에 따라 `BACKEND_API_URL` 을 명시 주입하므로 영향 없음. | resolved |
| BLK-3 | `backend-core/internal/httpapi/organization.go` 7곳 (createUser/updateUser/deleteUser/createOrgUnit/updateOrgUnit/deleteOrgUnit/replaceUnitMembers) | 본 mutation 성공 후 `recordAudit` 실패만으로 500 반환 → 클라이언트 재시도 시 중복/충돌 (Codex P1 + 동일 패턴 6곳 전수 점검 결과) | **적용 (2026-05-08): best-effort + 로그.** `audit.go` 에 `recordAuditBestEffort` 헬퍼 추가 (`log.Printf` 사용, backend-core 표준 logger 컨벤션 따름), organization.go 의 7곳 호출을 헬퍼로 교체. 응답 흐름은 `addAuditMeta` 의 기존 `AuditID==""` 가드로 자연 흡수. `gofmt` 통과. organization 다른 핸들러에 동일 패턴 잔존 없음 (grep 검증). 컴파일 검증은 PR 작성자 환경에서 수행 필요(사내 GOPROXY 제약). | resolved |

### 3.2 보안 — backlog 등록 후 phase 종료 전 해소

| ID | 위치 | 문제 | 권장 조치 |
| --- | --- | --- | --- |
| SEC-1 | `frontend/components/layout/AuthGuard.tsx` | "Basic Mock Auth Check" — `useStore().role` 유무로만 redirect, 실제 인증 아님 | Phase 5.2 종료 전 실제 토큰 검증으로 교체. backlog 항목 등록. |
| SEC-2 | `backend-core/internal/httpapi/auth.go` | `BearerTokenVerifier` 가 nil 이면 인증 우회 (`X-Devhub-Auth: bearer_unverified` 헤더만 부착) | `main.go` 에서 prod 빌드 시 verifier 가 항상 주입되는지 확인하고, 누락 시 즉시 fail-fast. |

### 3.3 위생 — 분리 PR 또는 일괄 정리

| ID | 위치 | 문제 | 권장 조치 |
| --- | --- | --- | --- |
| HYG-1 | `.DS_Store` | macOS 부산물 커밋 | `.gitignore` 추가 후 파일 제거. |
| HYG-2 | `ai-workflow/memory/session_handoff.md.bak` | 백업 파일 51줄 포함 | 의도가 아니면 제거. |
| HYG-3 | 루트 `workflow-source` (1줄 신규 파일) | 정체 불명 | 작성자가 용도 확인 후 유지/삭제 결정. |
| HYG-4 | `tests/devhub_temp_source/standard-ai-workflow-antigravity-v0.4.1-beta.zip` 외 | release artifact 가 git 본체에 포함 | GitHub Releases / LFS 또는 별도 저장소로 분리. |
| HYG-5 | `ai-workflow/memory/backlog/2026-05-04.md`, `2026-05-06.md` 삭제 | 백로그 이력 손실 위험 | 브랜치별 디렉터리로 이관됐는지 확인. 누락 시 복원. |

### 3.4 인프라 — 정책 결정에 종속

| ID | 위치 | 문제 | 권장 조치 | 상태 |
| --- | --- | --- | --- | --- |
| INF-1 | `docker-compose.yml`, `frontend/next.config.ts` | docker 의존을 기본 가정으로 만든 변경 | BLK-1 결정으로 `docker-compose.yml` 은 git 추적 해제됨. `frontend/next.config.ts` 만 BLK-2 로 처리. | resolved (compose 측) / open (next.config 측 → BLK-2) |
| INF-2 | DB 포트 5432 → 5433 매핑 | 로컬 PG 와의 충돌 회피 — 정책 유지 시 의미 없음 | `docker-compose.yml` 자체가 git 외부로 이동. 환경별 포트 매핑은 사용자 로컬 자산에서 관리. | resolved |

### 3.5 PR 분할 권장

353 files 단일 머지는 회귀 추적·롤백이 사실상 불가. 다음 단위로 분할:

1. backend-core 신규 API + 테스트 (auth, rbac, realtime, audit, commands, organization 변경)
2. frontend Phase 5.2 (login/account/gardener/AuthGuard/websocket/rbac UI)
3. infra (docker-compose, next.config) — BLK-1 결정에 종속
4. `ai-workflow/memory/**` 브랜치별 메모리 정리
5. `source-docs/workflow-source/**` 대량 자료 (단독 PR 권장)
6. `tests/devhub_temp_source/**` 분리·제거

분할이 어렵다면 file-path 기준으로 reviewer 를 분담시키도록 review 코멘트를 영역별로 정리.

### 3.6 self-review 의 follow-up 흡수

PR 작성자가 이미 적은 next steps 도 동일 트래커에 흡수.

| ID | 항목 | 매핑 |
| --- | --- | --- |
| FU-1 | DB 비밀번호 등 `.env` 분리 | INF 영역 — docker 트랙 결정 후 처리 |
| FU-2 | API 통신 오류 UI 피드백 강화 | frontend Phase 5.2 backlog |
| FU-3 | docker 빌드 최적화 (volumes 바인드 마운트) | INF 영역 — BLK-1 결정 후 |

PR description 의 향후 작업도 동일 트래커:

- RBAC write API 연동 → backend backlog
- WebSocket 실시간 명령 상태 스트레스 테스트 → 검증 backlog
- main 머지 → 본 문서의 모든 블로커 해소 조건

## 4. 머지 전 검증 체크리스트

PR 코멘트로 결과를 첨부.

- [ ] BLK-1 의사결정 기록 (정책 유지/갱신, 어디에 기록했는지)
- [ ] BLK-2 next.config fallback 수정 후 native dev 환경에서 `/api/*` 통신 확인
- [ ] BLK-3 organization.go audit 실패 패턴 통일 후 update/delete 핸들러 점검
- [ ] SEC-1, SEC-2 backlog 항목 등록 (위치/ID 첨부)
- [ ] HYG-1~5 처리 또는 별도 cleanup PR 발행
- [ ] `cd backend-core && go test ./...` 결과 첨부
- [ ] `cd frontend && npm run build && npx tsc --noEmit` 결과 첨부
- [ ] `pytest ai-workflow/tests/check_docs.py` 결과 첨부
- [ ] PR 분할 결정 기록 (분할 시 신규 PR 번호, 미분할 시 사유)

## 5. 다음 세션 인계 메모

- 본 문서는 main 브랜치 flat 위치(`ai-workflow/memory/PR-12-review-actions.md`)에 둔다. PR #12 가 들고 오는 `ai-workflow/memory/plans/` 디렉터리와 충돌하지 않기 위함.
- PR #12 머지 시 본 문서 status 를 `done` 으로 갱신하거나, 미해결 항목만 `ai-workflow/memory/work_backlog.md` 로 옮기고 본 문서는 archive.
- 신규 코멘트가 추가되면 §1 표에 row 를 덧붙이고 §3 액션 항목을 갱신한다.
