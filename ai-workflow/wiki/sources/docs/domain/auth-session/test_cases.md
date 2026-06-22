---
title: test_cases
type: source
tags: [domain, test_cases.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/auth-session/test_cases.md]
git_commit: 046e0c81
git_branch: chore/260622-wiki-drift-cleanup-4
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:22:35Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# Test Cases — M2 Auth/Account (Keycloak OIDC Baseline)

- 문서 목적: M2 인증/계정 영역의 현재 유효 E2E/UT 검증 범위를 정의한다.
- 범위: 로그인/로그아웃, AuthGuard, 계정(비밀번호 변경), 관리자 계정 화면의 핵심 회귀.
- 대상 독자: QA, e2e 작성자, 릴리즈 검증 담당.
- 상태: accepted
- 최종 수정일: 2026-05-20
- 관련 문서: [E2E 가이드](../setup/e2e-test-guide.md), [API 계약 §11](../backend_api_contract.md), [아키텍처 §6.2](../architecture.md)

## 1. 기준

- 인증 흐름은 Keycloak OIDC 기준이다.
- `/api/v1/auth/*`, `login_challenge`, `logout_challenge` 기반 시나리오는 제거되었다.
- `/auth/signup` 은 전환 기간 동안 비활성화 상태이며, self-signup 성공 시나리오는 현행 범위가 아니다.

## 2. 유효 E2E 시나리오

| 영역 | spec 파일 | 핵심 TC |
| --- | --- | --- |
| role landing + auth guard | `frontend/tests/e2e/auth.spec.ts` | developer/manager/system_admin 기본 랜딩, 비로그인 보호 경로 리다이렉트, 로그인 실패 유지 |
| sign out | `frontend/tests/e2e/signout.spec.ts` | Sign Out 후 재인증 요구, 보호 페이지 재진입 차단, 사용자 전환 |
| account password | `frontend/tests/e2e/account.spec.ts` | 본인 비밀번호 변경, 안내 문구 노출, mismatch 검증 |
| signup guard | `frontend/tests/e2e/signup.spec.ts` | `/auth/signup` 비활성화 안내 + `/login` 링크 동작 |
| admin users search | `frontend/tests/e2e/admin-users-search.spec.ts` | 사용자 검색/필터 UI smoke |
| admin permissions smoke | `frontend/tests/e2e/admin-permissions.spec.ts` | 권한 매트릭스 진입 smoke |

## 3. 실행

```sh
cd frontend
npm run e2e
```

사전 조건은 `docs/setup/e2e-test-guide.md` 를 따른다.

## 4. 핵심 합격 기준

1. 로그인 성공 시 role 기반 기본 랜딩이 일관되다.
2. 비로그인 또는 로그아웃 이후 보호 경로 접근이 항상 재인증으로 귀결된다.
3. `/account` 비밀번호 변경 성공/실패 UX가 명확하다.
4. `/auth/signup` 접근 시 self-signup 비활성화 안내가 노출된다.

## 5. Legacy 정리 메모

아래 항목은 **현행 기준에서 제외**한다.

- Kratos webhook 기반 audit 상세 검증(`kratos-audit-webhook.spec.ts`) 중심 TC
- `/api/v1/auth/signup` 성공/실패/cleanup round-trip TC
- `login_challenge`/`logout_challenge` 전제의 URL 단언

필요 시 historical 목적의 회귀 기록은 git history에서 확인한다.
