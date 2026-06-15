# Work Backlog — codex/work_260521-a-next-work

- 문서 목적: 현재 브랜치의 workflow 작업 항목과 상태를 관리한다.
- 범위: 23000 단일 포트 접속 검증 + OIDC redirect 정합 리팩토링 + PR 준비
- 대상 독자: 구현 담당자, 리뷰어
- 상태: done
- 최종 수정일: 2026-05-21

## 백로그

| ID | 작업 | 상태 | 비고 |
| --- | --- | --- | --- |
| PKG-URL-01 | 접속 URL/포트 설정 단일화 (`DEVHUB_PUBLIC_BASE_URL`) | done | frontend runtime-config + endpoints fallback 정리 |
| PKG-URL-02 | 배포 preflight 정합 검증 강화 | done | redirect URI == public base URL + basePath 강제 |
| PKG-URL-03 | tailscale `100.90.113.29:23000` 기준 e2e/redirect 검증 | done | Playwright 51 passed, localhost redirect 미재현 |
| PKG-URL-04 | 작업 정리/커밋/PR 준비 | in_progress | 현재 단계 |
