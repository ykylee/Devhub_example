# Dogfood 테스트 시나리오

- 문서 목적: dogfood 전용 로컬 시뮬레이션 환경에서 수행할 우선 검증 시나리오를 단계별로 정의한다.
- 범위: smoke 체크, 인증/온보딩, Gitea provider 연결, 플랫폼/프로젝트/저장소 흐름, CI 대시보드 검증, 종료 기준
- 대상 독자: QA, 개발자, AI 에이전트, 기능 검증 수행자
- 상태: draft
- 최종 수정일: 2026-06-05
- 관련 문서: [Dogfood 환경 셋업 가이드](./environment_setup.md), [통합 테스트 시나리오 카탈로그](../planning/integrated_test_scenarios.md), [E2E Test Guide](../setup/e2e-test-guide.md)

## 1. 원칙

dogfood 시나리오는 다음 순서를 따른다.

1. **환경 smoke**: 포트/헬스/OIDC/Gitea 연결 확인
2. **인증 기반 흐름**: 로그인, 역할 라우팅, seed 정합
3. **관리자 설정 흐름**: Gitea provider 등록
4. **도메인 흐름**: 플랫폼, 프로젝트, 저장소 생성 및 연결
5. **검증 흐름**: 대시보드, 이슈/PR 동기화, CI 데이터 표시

## 2. 환경 기준

| 구성 요소 | 주소 |
| --- | --- |
| Frontend | `http://localhost:13000` |
| Backend Core | `http://localhost:18080` |
| Backend AI | `http://localhost:18000` |
| Keycloak | `http://localhost:18180/devhub/auth/keycloak` |
| PostgreSQL | `localhost:15433` |
| Gitea | `https://homelab.ddn777.synology.me/gitea` |

## 3. 사전 준비 데이터

### 3.1 계정

Playwright seed 기준:

| 사용자 | email | 역할 | 기대 landing |
| --- | --- | --- | --- |
| Alice | `alice@example.com` | `developer` | `/developer` |
| Bob | `bob@example.com` | `team_manager` | `/manager` |
| Charlie | `charlie@example.com` | `system_admin` | `/admin` |

### 3.2 외부 연동

- Gitea PAT 준비
- `provider base_url` 은 `https://homelab.ddn777.synology.me/gitea`
- webhook secret 준비

## 4. Phase 0 — Smoke

### SC-DOGFOOD-SMOKE-01: 인프라 헬스

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-SMOKE-01` |
| 우선순위 | P0 |
| 절차 | 1. `curl http://localhost:18080/health` 2. `curl http://localhost:18000/health` 3. `curl -I http://localhost:13000/` 4. `curl http://localhost:18180/devhub/auth/keycloak/realms/devhub/.well-known/openid-configuration` 5. `pg_isready -h localhost -p 15433 -U postgres -d devhub` |
| 기대 결과 | 5개 확인 항목 모두 정상 응답 |

### SC-DOGFOOD-SMOKE-02: Gitea API 연결

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-SMOKE-02` |
| 우선순위 | P0 |
| 절차 | 1. `curl -H "Authorization: token $GITEA_TOKEN" https://homelab.ddn777.synology.me/gitea/api/v1/version` 2. `curl -H "Authorization: token $GITEA_TOKEN" https://homelab.ddn777.synology.me/gitea/api/v1/user` |
| 기대 결과 | Gitea version JSON, user JSON 정상 반환 |

## 5. Phase 1 — 인증과 라우팅

### SC-DOGFOOD-AUTH-01: system_admin 로그인

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-AUTH-01` |
| 우선순위 | P0 |
| 절차 | 1. `http://localhost:13000/login` 접속 2. Keycloak 로그인 3. Charlie 계정으로 로그인 4. `/admin` landing 확인 |
| 기대 결과 | 관리자 대시보드 진입 성공, `/api/v1/me` 응답에서 role=`system_admin` |

### SC-DOGFOOD-AUTH-02: 역할별 landing 확인

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-AUTH-02` |
| 우선순위 | P0 |
| 절차 | 1. Alice 로그인 → `/developer` 2. Bob 로그인 → `/manager` 3. Charlie 로그인 → `/admin` |
| 기대 결과 | 각 역할이 기대한 landing 으로 이동 |

### SC-DOGFOOD-AUTH-03: 권한 제한 확인

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-AUTH-03` |
| 우선순위 | P1 |
| 절차 | 1. Alice 로 `/admin/settings/users` 직접 접근 2. API `GET /api/v1/admin/settings/users` 호출 3. Charlie 로 동일 경로 재검증 |
| 기대 결과 | 일반 사용자 접근은 redirect 또는 403, 관리자만 접근 가능 |

## 6. Phase 2 — Gitea Provider 연결

### SC-DOGFOOD-INT-01: Gitea provider 등록

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-INT-01` |
| 우선순위 | P0 |
| 절차 | 1. Charlie 로그인 2. `/admin/settings/integrations` 이동 3. `Register Provider` 클릭 4. preset 또는 수동으로 아래 입력: provider key, display name, type=`scm`, auth mode=`token`, base URL=`https://homelab.ddn777.synology.me/gitea`, API token=`$GITEA_TOKEN`, webhook secret=`$GITEA_WEBHOOK_SECRET` 5. 저장 |
| 기대 결과 | provider row 생성, sync 가능 상태 표시 |

### SC-DOGFOOD-ONBOARD-01: 신규 계정 온보딩 smoke

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-ONBOARD-01` |
| 우선순위 | P0 |
| 절차 | 1. `./scripts/dogfood-create-user.sh` 로 Keycloak 계정 생성 2. `/login` 로그인 3. `/onboarding` 진입 확인 4. display name 입력 5. `Engineering (dept-eng)` 선택 6. 제출 7. `/developer` 이동 및 `/api/v1/me` 확인 |
| 기대 결과 | 신규 사용자가 온보딩을 완료하고 `review_status=pending_review` 상태로 DevHub 에 등록됨 |

### SC-DOGFOOD-INT-02: Gitea provider sync

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-INT-02` |
| 우선순위 | P0 |
| 절차 | 1. 등록된 provider 에서 `Sync` 실행 2. `/repositories` 또는 관련 API 에서 외부 저장소 목록 확인 |
| 기대 결과 | Gitea 저장소 메타데이터가 DevHub 에 유입됨 |

### SC-DOGFOOD-INT-03: 관리자 Gitea provider 등록 + sync 스모크

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-INT-03` |
| 우선순위 | P0 |
| 절차 | 1. `./scripts/dogfood.sh test-integration-admin` 실행 2. 신규 provider row 생성 확인 3. sync job accepted 확인 4. provider API 에서 `yklee/devhub-simulation` 같은 원격 저장소가 보이는지 확인 5. 생성한 provider 정리 |
| 기대 결과 | 외부 Gitea provider 등록, sync 요청, 원격 저장소 조회, cleanup 이 한 번에 성공 |

## 7. Phase 3 — 플랫폼/프로젝트/저장소 흐름

### SC-DOGFOOD-ORG-01: 조직 단위 생성/수정/삭제와 리더/멤버 변경

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-ORG-01` |
| 우선순위 | P0 |
| 절차 | 1. `./scripts/dogfood.sh test-organization-admin` 실행 2. 조직 단위 생성 3. 이름 수정 4. 리더 변경 5. 멤버 추가/변경 6. 생성한 조직 단위 삭제 |
| 기대 결과 | 조직 관리 화면에서 create/edit/members/delete 흐름이 모두 정상 완료 |

### SC-DOGFOOD-PLAT-01: 플랫폼 생성

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-PLAT-01` |
| 우선순위 | P0 |
| 절차 | 1. Charlie 로그인 2. `/admin/settings/platforms` 이동 3. 새 플랫폼 생성 (`key` 는 10자 이내 영숫자) |
| 기대 결과 | 플랫폼 생성 성공, 상세 페이지 접근 가능 |

### SC-DOGFOOD-SELF-01: self dogfooding 관리자 생성 흐름

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-SELF-01` |
| 우선순위 | P0 |
| 절차 | 1. Charlie 로그인 2. 현재 워크스페이스를 대표하는 플랫폼 생성 3. `gitea` provider 기반 repository draft 생성 4. 플랫폼에 저장소 연결 5. 연결된 저장소를 참조하는 프로젝트 생성 6. 플랫폼/프로젝트 상세 확인 |
| 기대 결과 | platform / repository / project 가 모두 생성되고, 프로젝트 상세에서 연결된 저장소 slug 가 보임 |

### SC-DOGFOOD-PROJ-01: 프로젝트 생성

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-PROJ-01` |
| 우선순위 | P0 |
| 절차 | 1. 위 플랫폼 내부에서 프로젝트 생성 2. owner / visibility / status 입력 |
| 기대 결과 | 프로젝트 생성 성공, 상세 페이지에서 프로젝트 카드 노출 |

### SC-DOGFOOD-REPO-01: 외부 Gitea 저장소 연결 또는 생성

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-REPO-01` |
| 우선순위 | P0 |
| 절차 | 1. Gitea provider 선택 2. existing repo link 또는 outbound create 실행 3. 프로젝트/플랫폼에 저장소 연결 |
| 기대 결과 | repository row 생성, 프로젝트와 연결 확인 |

## 8. Phase 4 — Gitea 작업물 동기화

### SC-DOGFOOD-SCM-01: Issue / PR 동기화

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-SCM-01` |
| 우선순위 | P1 |
| 절차 | 1. Gitea 에서 대상 repo 의 issue 1건, PR 1건 생성 2. DevHub 에서 provider sync 재실행 3. 저장소 상세 또는 관련 API 확인 |
| 기대 결과 | issue / PR 목록이 DevHub 에 보임 |

### SC-DOGFOOD-SCM-02: assignee / 상태 변화 확인

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-SCM-02` |
| 우선순위 | P1 |
| 절차 | 1. Gitea issue assignee 지정 2. close/reopen 또는 PR 상태 변경 3. DevHub sync 재실행 |
| 기대 결과 | assignee 와 상태가 가능한 범위 내에서 DevHub 에 반영됨 |

## 9. Phase 5 — 대시보드와 CI 확인

### SC-DOGFOOD-CI-01: Repository build-runs 기본 표시

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-CI-01` |
| 우선순위 | P1 |
| 절차 | 1. 저장소 상세 페이지 이동 2. Build Runs 섹션 확인 3. API endpoint 존재 여부 확인 |
| 기대 결과 | build run 이력 노출 또는 미구현 여부를 명확히 식별 |

### SC-DOGFOOD-REPO-DASH-01: 저장소 상세 대시보드 역할별 검증

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-REPO-DASH-01` |
| 우선순위 | P0 |
| 절차 | 1. `./scripts/dogfood.sh test-repository-dashboard` 실행 2. developer 시점에서 build/log modal 확인 3. manager 시점에서 manager & governance 탭과 contributor distribution 토글 확인 |
| 기대 결과 | 새 repository dashboard 가 dogfood 환경에서 역할별로 정상 렌더링 |

### SC-DOGFOOD-DASH-01: 플랫폼/프로젝트 대시보드 표시

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-DASH-01` |
| 우선순위 | P1 |
| 절차 | 1. 플랫폼 상세 대시보드 확인 2. 프로젝트 상세 대시보드 확인 3. linked repository / active project / backlog 위젯 확인 |
| 기대 결과 | 현재 구현된 대시보드 위젯이 정상 렌더링 |

### SC-DOGFOOD-DASH-02: self dogfood 결과 대시보드 확인

| 항목 | 내용 |
| --- | --- |
| ID | `TC-DOGFOOD-DASH-02` |
| 우선순위 | P0 |
| 절차 | 1. `./scripts/dogfood.sh test-self-dogfood-dashboard` 실행 2. self dogfood 플랫폼 생성 3. 플랫폼 상세에서 roadmap / repositories 위젯 확인 4. 프로젝트 상세에서 connected repositories / recent activity / active tasks 위젯 확인 |
| 기대 결과 | self dogfooding으로 만든 platform / project / repository 결과가 대시보드 위젯에 반영됨 |

## 10. 권장 실행 순서

```text
Phase 0  Smoke
  -> SC-DOGFOOD-SMOKE-01
  -> SC-DOGFOOD-SMOKE-02

Phase 1  인증
  -> SC-DOGFOOD-AUTH-01
  -> SC-DOGFOOD-AUTH-02
  -> SC-DOGFOOD-AUTH-03

Phase 2  Integration
  -> SC-DOGFOOD-INT-01
  -> SC-DOGFOOD-INT-02
  -> SC-DOGFOOD-INT-03

Phase 3  플랫폼/프로젝트/저장소
  -> SC-DOGFOOD-ORG-01
  -> SC-DOGFOOD-PLAT-01
  -> SC-DOGFOOD-PROJ-01
  -> SC-DOGFOOD-REPO-01

Phase 4  SCM 동기화
  -> SC-DOGFOOD-SCM-01
  -> SC-DOGFOOD-SCM-02

Phase 5  대시보드/CI
  -> SC-DOGFOOD-CI-01
  -> SC-DOGFOOD-REPO-DASH-01
  -> SC-DOGFOOD-DASH-01
  -> SC-DOGFOOD-DASH-02
```

## 11. 종료 기준

다음 조건을 충족하면 dogfood 1차 검증을 완료로 본다.

1. 로그인과 역할별 landing 이 정상 동작한다.
2. Gitea provider 등록이 가능하다.
3. 플랫폼, 프로젝트, 저장소 연결이 가능하다.
4. 외부 Gitea 의 issue 또는 PR 중 최소 1건이 DevHub 에서 확인된다.
5. 플랫폼/프로젝트 대시보드가 오류 없이 렌더링된다.

## 12. 알려진 리스크

| 리스크 | 영향 | 대응 |
| --- | --- | --- |
| 외부 Gitea 상태 변화 | provider 등록, sync 차단 | 토큰/URL 재검증, 필요 시 로컬 Gitea fallback |
| Keycloak health 지연 | 로그인 전 smoke 단계 지연 | 컨테이너 health가 `healthy` 될 때까지 대기 |
| CI run 생성 API 미완성 영역 | build 검증 한계 | dashboard/API 확인 위주로 범위 축소 |
| webhooks 미구성 | 실시간 반영 한계 | manual sync 재실행으로 검증 |

## 13. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-05 | dogfood 전용 분리 포트/외부 Gitea 연동 기준 시나리오 초판 작성. smoke, auth, integration, platform/project/repository, SCM, dashboard/CI 흐름 정리. |
