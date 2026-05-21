# Session Handoff — codex/work_260521-a-next-work

- 문서 목적: 23000 단일 포트 배포 검증/리팩토링 결과를 인계한다.
- 범위: OIDC redirect 정합, 배포 env 계약, E2E 검증
- 대상 독자: 후속 에이전트, 리뷰어
- 상태: in_progress
- 최종 수정일: 2026-05-21

## 이번 세션 요약

- `DEVHUB_PUBLIC_BASE_URL`를 배포 계약의 기준값으로 도입했다.
- frontend runtime-config/클라이언트 fallback이 public base URL 기준으로 redirect URI를 생성하도록 정리했다.
- deploy preflight에서 redirect URI 정합(`public_base + basePath + /auth/callback`)을 fail-fast로 검증하도록 보강했다.
- tailscale `http://100.90.113.29:23000` 기준 Playwright E2E를 실행했고 `51 passed`를 확인했다.
- nginx/keycloak 로그 기준으로 `localhost` 리디렉션 흔적 없이 `100.90.113.29:23000`로만 OIDC flow가 수행됨을 확인했다.

## 다음 세션 첫 작업

1. 본 브랜치 PR 생성/리뷰 반영 후 merge.
2. stage/prod 실환경 `.env`에 `DEVHUB_PUBLIC_BASE_URL` 반영.
3. deploy preflight + smoke/E2E 재검증.
