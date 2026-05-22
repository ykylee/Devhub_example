# Work Backlog — codex/work_260521-c-db-docker-option

- 문서 목적: 현재 브랜치의 workflow 작업 항목과 상태를 관리한다.
- 범위: host build packaging + nginx ingress 정합
- 대상 독자: 구현 담당자, 리뷰어
- 상태: in_progress
- 최종 수정일: 2026-05-22

## 백로그

| ID | 작업 | 상태 | 비고 |
| --- | --- | --- | --- |
| BUILD-ARTIFACT-01 | host build 스크립트 추가 | done | `scripts/build-artifacts.sh` |
| BUILD-ARTIFACT-02 | runtime-only Dockerfile 로 전환 | done | backend-core/backend-ai/frontend |
| BUILD-ARTIFACT-03 | deploy wrapper 에 host build 연결 | done | `deploy-from-env.sh` |
| BUILD-ARTIFACT-04 | nginx 23000 HTTP redirect 제거 | done | `absolute_redirect off` 포함 |
| BUILD-ARTIFACT-05 | e2e Keycloak HTTPS 요구 원인 정리 | done | `sslRequired=none` + master realm 갱신 |
| BUILD-ARTIFACT-06 | compose nginx TLS 제거 | done | 443 포트/인증서 마운트 삭제 |
| BUILD-ARTIFACT-07 | backend-core Go 버전 1.25.9 고정 | done | `go.mod` + docs 정합 |
| BUILD-ARTIFACT-08 | host-run e2e 의 DB 접근 경로 정리 | in_progress | `db` 호스트명 host 해석 실패 |
| BUILD-ARTIFACT-09 | backend-core Alpine ca-certificates 제거 | done | 2026-05-21 (`a216514` post-rebase) |
| BUILD-ARTIFACT-10 | frontend public copy 제거 | done | 2026-05-21 (`a1d3ae4` post-rebase) |
| BUILD-ARTIFACT-11 | deploy-from-env.sh push 액션 제거 | done | 로컬 전용 build/deploy 흐름 단순화 |
| BUILD-ARTIFACT-12 | 외부 접속 기준 주소/포트 분리 | done | 2026-05-21 (`61e0937` post-rebase, `PUBLIC_ACCESS_*` + `NGINX_HTTP_PORT=3000`) |
| BUILD-ARTIFACT-13 | main rebase + force-push | done | 2026-05-22 origin/main `1239f3c` 위로, 3 commit drop (PR #282 흡수분), 백업 ref `backup/pre-rebase-main-260522` |
| BUILD-ARTIFACT-14 | nginx conf ↔ template auto-sync | done | 2026-05-22 `3ae46cf` — `scripts/nginx-conf-sync.sh` (envsubst render, `--check`/`--fix`) + `deploy-preflight.sh` 가 `[0/3]` 단계에서 `--fix` 자동 실행 + `infra/nginx/devhub.deploy.conf` `.gitignore` 추가 + stale 버전 git rm |
| BUILD-ARTIFACT-15 | Keycloak image 25.0 pin | done | 2026-05-22 `42b18b1` — compose / CI / dev-up.sh 3 위치 (active code only). docs 의 historical 26.0 언급은 immutable 보존 |
| BUILD-ARTIFACT-16 | ADR-0022 draft 발행 | done | 2026-05-22 `docs/adr/0022-keycloak-version-pin-25-0.md` draft (§3.1 retreat 사유 placeholder — 사용자 finalize 후 Accepted 승격) |
| BUILD-ARTIFACT-17 | Python image 3.12 pin | done | 2026-05-22 `7626a8c` — backend-ai/Dockerfile 3.13 → 3.12 + tech_stack.md / environment-setup.md prerequisite 3.12 + build-artifacts.sh ABI 주의 주석. host `python3` 도 3.12 권장 (site-packages ABI 호환) |

## 잔여 작업

- **BUILD-ARTIFACT-08** in_progress. `idp-apply-schemas` host-run 시 docker 내부 호스트명 `db` 해석 실패. DSN/포트 노출 분리 또는 host-network 모드 검토 필요.
- ADR-0022 §3.1 retreat 사유 사용자 finalize → Accepted 승격.
- 사용자 사내 환경 redeploy + :13000 smoke test (preflight 가 nginx conf 자동 sync + Keycloak 25.0 image pull).
- PR 생성 여부 결정 (현재 push 만 완료).
