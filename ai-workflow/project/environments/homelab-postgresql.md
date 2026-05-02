# Homelab PostgreSQL 환경 기록

- 문서 목적: DevHub 로컬/홈랩 PostgreSQL 검증 환경 정보를 기록한다.
- 기준일: 2026-05-02
- 상태: active
- 관련 작업: `TASK-007 Gitea Webhook 수신부 및 데이터 모델링 구현`

## 1. 접속 정보

| 항목 | 값 |
| --- | --- |
| Host | `192.168.0.38` |
| Port | `5432` |
| Service | PostgreSQL |
| Database | `devhub` |
| User | `devhub` |
| Password | `.env.local`에 로컬 저장. 문서에는 평문 저장하지 않음 |
| SSL mode | `disable` 기준으로 1차 확인 |

## 2. 확인 결과

- TCP 포트 확인: 성공
  - 명령: `nc -vz 192.168.0.38 5432`
  - 결과: `Connection to 192.168.0.38 port 5432 [tcp/postgresql] succeeded!`
- 인증 확인: 실패
  - 명령: `migrate -path backend-core/migrations -database 'postgres://postgres:<password>@192.168.0.38:5432/postgres?sslmode=disable' version`
  - 결과: `pq: password authentication failed for user "postgres"`
- 정정된 인증 정보 재확인: 실패
  - 기준: 사용자 제공 정정 비밀번호 사용
  - 명령: `migrate -path backend-core/migrations -database 'postgres://postgres:<password>@192.168.0.38:5432/postgres?sslmode=disable' version`
  - 결과: `pq: password authentication failed for user "postgres"`
- DevHub 전용 계정 인증 확인: 성공
  - 기준: 사용자 제공 `devhub` 계정 사용
  - 명령: `migrate -path backend-core/migrations -database 'postgres://devhub:<password>@192.168.0.38:5432/devhub?sslmode=disable' version`
  - 결과: `no migration` (인증과 DB 연결 성공, migration 미적용 상태)
- Migration 적용: 성공
  - 명령: `PATH="$HOME/go/bin:$PATH" MIGRATE_DB_URL='postgres://devhub:<password>@192.168.0.38:5432/devhub?sslmode=disable' make migrate-up`
  - 결과: `1/u create_webhook_events`
- Migration version 확인: 성공
  - 명령: `migrate -path backend-core/migrations -database 'postgres://devhub:<password>@192.168.0.38:5432/devhub?sslmode=disable' version`
  - 결과: `1`

## 3. 현재 판단

- 네트워크와 PostgreSQL 포트는 접근 가능하다.
- 기존 `postgres` 계정 인증은 실패했지만, DevHub 전용 `devhub` 계정 인증은 성공했다.
- `webhook_events` migration은 홈랩 PostgreSQL `devhub` DB에 적용 완료되었다.
- 인증정보는 repo root의 `.env.local`에 저장했고 `.gitignore`로 커밋 제외한다.

## 4. 다음 확인 항목

- `migrate` CLI가 `$HOME/go/bin`에 설치되어 있으므로 shell PATH 또는 Makefile 실행 환경에 반영할지 결정.
- 다음 DB 변경 시 아래 형식으로 migration을 적용한다.

```bash
MIGRATE_DB_URL='postgres://devhub:<password>@192.168.0.38:5432/devhub?sslmode=disable' make migrate-up
```
