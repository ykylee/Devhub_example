# Homelab PostgreSQL 환경 기록

- 문서 목적: DevHub 로컬/홈랩 PostgreSQL 검증 환경 정보를 기록한다.
- 기준일: 2026-05-02
- 상태: blocked
- 관련 작업: `TASK-007 Gitea Webhook 수신부 및 데이터 모델링 구현`

## 1. 접속 정보

| 항목 | 값 |
| --- | --- |
| Host | `192.168.0.38` |
| Port | `5432` |
| Service | PostgreSQL |
| Database | 확인 필요 |
| User | `postgres` |
| Password | 사용자 제공 값 사용. 문서에는 평문 저장하지 않음 |
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

## 3. 현재 판단

- 네트워크와 PostgreSQL 포트는 접근 가능하다.
- 사용자 제공 계정 정보와 정정된 인증 정보 모두 현재 인증이 통과하지 않는다.
- DB명 확인 전에 사용자/password 인증 단계에서 실패하므로, PostgreSQL 사용자 비밀번호 또는 `pg_hba.conf`/인증 정책 확인이 필요하다.

## 4. 다음 확인 항목

- 홈랩 PostgreSQL의 실제 `postgres` 사용자 비밀번호 확인.
- DevHub 전용 database 이름(`devhub` 또는 별도 이름) 생성 여부 확인.
- DevHub migration 적용용 접속 문자열 확정:

```bash
MIGRATE_DB_URL='postgres://postgres:<password>@192.168.0.38:5432/<database>?sslmode=disable' make migrate-up
```
