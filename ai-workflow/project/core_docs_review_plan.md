# 핵심 개발 문서 리뷰 계획

- 목적: 요구사항, 아키텍처, 기술 스택 문서를 집중 리뷰하고 수정/결정 사항을 추적한다.
- 상태: in_progress
- 작성일: 2026-04-29
- 관련 작업: TASK-009 핵심 개발 문서 집중 리뷰
- 대상 문서: `docs/requirements.md`, `docs/architecture.md`, `docs/tech_stack.md`

## 1. 리뷰 범위

이번 리뷰는 제품/시스템 개발의 기준선이 되는 세 문서만 대상으로 한다.

- `docs/requirements.md`: 역할별 요구사항, 운영 원칙, 데이터 원천/연동 전략, 기술 결정의 요구사항 반영 여부
- `docs/architecture.md`: 컴포넌트 경계, 데이터 흐름, 내부/외부 통신 계약, 보안/인증, 구현 상태 표현
- `docs/tech_stack.md`: 확정 스택, 실행 환경, 초기화/검증 명령, 버전 계약, 개발 가이드

리뷰 중 발견한 수정 후보는 이 문서에 먼저 기록하고, 사용자와 합의된 항목만 원문 문서에 반영한다.

## 2. 리뷰 원칙

- 요구사항과 아키텍처가 서로 같은 제품 경계를 말하는지 확인한다.
- 현재 구현된 범위와 예정 범위가 문서에서 섞이지 않도록 구분한다.
- 기술 스택 문서의 명령/버전/포트가 실제 저장소와 충돌하지 않는지 확인한다.
- 모호한 정책은 바로 구현하지 않고 결정 필요 항목으로 남긴다.
- 리뷰 결과는 `planned`, `in_progress`, `blocked`, `done` 중 하나로 관리한다.

## 3. 리뷰 체크리스트

### 3.1 요구사항 문서 (`docs/requirements.md`)

- 상태: done
- 확인 항목:
  - 역할 정의가 개발자, 관리자, QA, 시스템 관리자까지 일관되게 분리되는가.
  - 후보 기능과 확정 결정이 문서 안에서 구분되는가.
  - Gitea 중심 데이터 원천 전략이 기능 요구사항과 연결되는가.
  - 데이터 주권, 조회 권한, 보존 정책, 알림 정책이 구현 가능한 수준으로 정의되는가.
  - 기술 스택 결정 사항이 별도 기술 문서와 중복되거나 충돌하지 않는가.
- 리뷰 결과:
  - 시스템 관리자 역할은 `5.3`에서 핵심 권한으로 등장하지만 `2. 사용자 역할별 요구사항`에는 별도 역할로 정의되어 있지 않다.
  - 상단 역할별 기능은 모두 `후보` 체크리스트로 남아 있고, 아래 `Agenda`는 `결정 사항`으로 작성되어 있어 확정 범위와 후보 범위가 섞여 보인다.
  - `기술 태깅`으로 용어를 바꾸기로 한 운영 원칙과 달리 `5.1`에는 `스텔스 전문가 맵` 표현이 다시 등장한다.
  - 데이터 주권, 조회 권한, 보존 정책, 알림 정책은 방향성은 있으나 역할별 CRUD/조회 권한과 데이터 분류표가 없어 구현 기준으로는 아직 모호하다.
  - 기술 스택 결정이 요구사항 문서에도 길게 포함되어 있고, `Go / Gin or Echo`처럼 기술 스택 문서의 `Go (Gin)` 확정 표현과 충돌하는 표현이 있다.

### 3.2 아키텍처 문서 (`docs/architecture.md`)

- 상태: done
- 확인 항목:
  - 컴포넌트 다이어그램이 현재 스캐폴딩과 목표 아키텍처를 구분하는가.
  - Go Core, Python AI, Frontend, PostgreSQL, Gitea/Gitea Runner 책임 경계가 명확한가.
  - REST, WebSocket/SSE, gRPC, Gitea API/Webhook의 용도가 분리되는가.
  - 데이터 저장/동기화/재처리 흐름이 TASK-007 구현으로 이어질 만큼 구체적인가.
  - 인증, RBAC, 감사 로그가 최소 구현 단위와 후속 범위로 나뉘는가.
- 리뷰 결과:
  - 컴포넌트 다이어그램은 목표 구조를 잘 보여주지만, 현재 구현된 health endpoint와 예정된 Auth/WebSocket/gRPC/Webhook 기능의 상태 구분이 충분하지 않다.
  - Python AI가 PostgreSQL을 직접 조회하는 화살표가 있는데, 요구사항과 기술 스택은 Gitea 이벤트가 Go Core를 먼저 통과한다고 설명한다. 분석 모듈의 DB 직접 접근 권한과 읽기 범위를 명확히 해야 한다.
  - Backend ↔ Frontend 실시간 통신이 `WebSocket 또는 SSE`로 남아 있어 기술 스택 문서의 `WebSocket 우선` 원칙과 결정 수준이 다르다.
  - Gitea Webhook 수신과 Hourly Pull은 언급되지만 서명 검증, 중복 이벤트 처리, 재처리, 실패 큐, 이벤트 원본 저장 기준이 없어 TASK-007 설계 입력으로 부족하다.
  - 인증, RBAC, Audit Log는 방향성만 있고 Gitea SSO 연동 전 최소 인증 방식과 시스템 관리자 권한 격리 기준이 분리되어 있지 않다.

### 3.3 기술 스택 문서 (`docs/tech_stack.md`)

- 상태: done
- 확인 항목:
  - Go, Python, Node, Docker, PostgreSQL, gRPC 도구 버전 계약이 실제 파일과 맞는가.
  - `make init`, `make build`, `make run` 설명이 실제 `Makefile`과 일치하는가.
  - 로컬 실행과 검증 절차가 새 개발자에게 충분히 재현 가능한가.
  - 아직 구현되지 않은 gRPC 서버/DB 마이그레이션/프론트엔드 연동 범위가 예정 상태로 표시되는가.
  - 운영 환경 변수와 비밀값 처리 기준이 과하거나 부족하지 않은가.
- 리뷰 결과:
  - `make init`, `make build`, `make run` 설명은 현재 `Makefile`과 대체로 일치한다.
  - `Node.js v20 이상` 문서와 `frontend/Dockerfile`의 `node:22-alpine`은 호환 범위 안이지만, 재현 가능한 기준 버전으로는 느슨하다.
  - `Python v3.10 이상` 문서와 `backend-ai/Dockerfile`의 `python:3.11-slim`은 호환되지만, `make setup`은 로컬 `python3`를 그대로 사용하므로 로컬 Python이 낮으면 실패할 수 있다.
  - `TanStack Query`, `React Flow`가 확정 스택으로 적혀 있으나 `frontend/package.json`에는 아직 설치되어 있지 않다.
  - `GITEA_TOKEN`은 문서에 필요하다고 되어 있지만 `docker-compose.yml`의 `backend-core` 환경 변수에는 전달되지 않는다.
  - `50051` 포트는 예정 상태로 표기되어 있으나 `docker-compose.yml`에는 이미 노출되어 있어 현재 동작 포트와 예약 포트의 구분이 필요하다.

## 4. 리뷰 산출물

- 리뷰 계획 및 진척: 이 문서
- 수정 후보 목록: 이 문서의 `5. 발견 사항`
- 합의 후 반영 대상: `docs/requirements.md`, `docs/architecture.md`, `docs/tech_stack.md`
- workflow 상태 동기화: `ai-workflow/project/state.json`, `ai-workflow/project/session_handoff.md`, 최신 backlog

## 5. 발견 사항

아직 상세 리뷰 전이다. 아래 형식으로 누적한다.

| ID | 상태 | 우선순위 | 대상 | 요약 | 처리 방향 |
| --- | --- | --- | --- | --- | --- |
| CORE-DOC-001 | done | P1 | `docs/requirements.md` | 시스템 관리자 역할이 기능 요구사항 섹션에는 없고 데이터 연동 섹션에서만 핵심 권한으로 등장함 | `2.4 시스템 관리자` 역할을 추가해 Gitea/Runner/계정/백업/Audit Log 관리 책임을 별도 역할로 명시 |
| CORE-DOC-002 | done | P1 | `docs/requirements.md` | 상단 역할별 기능은 `후보`, 하단 Agenda는 `결정 사항`이라 MVP/확정/후보 범위가 섞여 보임 | `1.1 요구사항 범위 구분`을 추가해 `확정`, `후보`, `MVP 이후` 상태 정의와 체크박스/Agenda 해석 기준을 명시 |
| CORE-DOC-003 | planned | P2 | `docs/requirements.md` | `기술 태깅`으로 용어를 바꾸기로 했지만 `스텔스 전문가 맵` 표현이 다시 사용됨 | 용어를 `기술 태깅 기반 전문가 맵`으로 통일 |
| CORE-DOC-004 | done | P1 | `docs/requirements.md` | 데이터 주권, 조회 권한, 보존, 알림 정책이 방향성 위주라 구현 시 권한/데이터 모델 판단 기준이 부족함 | `4.1 데이터 및 권한 운영 기준` 표를 추가해 데이터 분류별 원천/수정/조회/보존/알림 기준을 명시 |
| CORE-DOC-005 | planned | P2 | `docs/requirements.md` | 기술 스택 결정이 요구사항 문서와 기술 스택 문서에 중복되고, `Go / Gin or Echo`가 `Go (Gin)` 확정 표현과 충돌함 | 요구사항 문서의 기술 스택 섹션은 결정 요약으로 축소하고 상세 계약은 `tech_stack.md`로 위임 |
| CORE-DOC-006 | done | P1 | `docs/architecture.md` | 컴포넌트 다이어그램이 현재 구현과 목표 구현을 한 그림에 섞어 보여줌 | 다이어그램 앞에 `current/planned/external` 상태 기준을 추가하고 각 노드/연결 라벨에 구현 상태를 명시 |
| CORE-DOC-007 | done | P1 | `docs/architecture.md` | Python AI가 PostgreSQL을 직접 조회하는 구조가 Go Core 중심 이벤트 수집 원칙과 충돌할 수 있음 | 초기 구현에서 Python AI의 DB 직접 접근을 금지하고 Go Core를 경유해 분석 입력을 gRPC로 전달하도록 명시 |
| CORE-DOC-008 | planned | P2 | `docs/architecture.md` | Frontend 실시간 통신이 `WebSocket 또는 SSE`로 남아 있어 기술 스택의 `WebSocket 우선`과 결정 수준이 다름 | 기본값을 WebSocket으로 정하고 SSE는 fallback인지 제외인지 결정 |
| CORE-DOC-009 | done | P1 | `docs/architecture.md` | Webhook/Hourly Pull 흐름에 서명 검증, 중복 처리, 실패 재처리, 원본 이벤트 저장 기준이 없음 | `4.2 이벤트 수집 파이프라인`을 추가해 검증, 원본 저장, idempotency, 정규화, 재처리, reconciliation 기준을 명시 |
| CORE-DOC-010 | planned | P2 | `docs/architecture.md` | 인증, RBAC, Audit Log가 목표만 있고 초기 구현 단위가 분리되지 않음 | Gitea SSO 이전 최소 인증/권한 모델과 시스템 관리자 격리 기준을 단계별로 정리 |
| CORE-DOC-011 | done | P1 | `docs/tech_stack.md` | TanStack Query와 React Flow가 확정 스택으로 적혀 있지만 `frontend/package.json`에 설치되어 있지 않음 | 확정 스택이지만 현재 scaffold에는 미설치이며 API/인프라 뷰 구현 시점에 추가하는 `도입 예정` 상태로 문서화 |
| CORE-DOC-012 | planned | P2 | `docs/tech_stack.md` | Node/Python 버전 기준이 Dockerfile, 로컬 실행 명령, 문서 사이에서 재현 가능하게 고정되어 있지 않음 | 기준 버전과 최소 버전을 분리하고 `make setup`의 로컬 런타임 요구를 명시 |
| CORE-DOC-013 | done | P1 | `docs/tech_stack.md` | `GITEA_TOKEN`이 필요하다고 문서화되어 있지만 `docker-compose.yml`에는 backend-core 환경 변수로 전달되지 않음 | `docker-compose.yml`의 backend-core 환경 변수에 `GITEA_TOKEN=${GITEA_TOKEN}`을 추가해 문서의 필수 설정과 실행 계약을 일치시킴 |
| CORE-DOC-014 | planned | P2 | `docs/tech_stack.md` | `50051`은 미구현 예정 포트인데 compose에는 이미 노출되어 있어 현재 동작 포트처럼 보일 수 있음 | 미구현 예약 포트임을 compose 주석/문서에 함께 명시하거나 구현 전까지 노출 제거 검토 |
| CORE-DOC-015 | planned | P2 | `docs/tech_stack.md` | DB 마이그레이션 도구와 검증 명령이 아직 기술 스택 문서에 없다 | PostgreSQL 모델링 시작 전 migration 도구 후보와 실행 명령을 결정 항목으로 추가 |

## 6. 진행 기록

- 2026-04-29 23:35 KST: 리뷰 대상 3개 문서와 workflow 상태를 확인하고 집중 리뷰 계획 문서를 생성함.
- 2026-04-29 23:40 KST: `docs/requirements.md` 리뷰 완료. 역할 정의, 후보/확정 범위, 용어 통일, 권한/데이터 정책, 기술 스택 중복 이슈를 `CORE-DOC-001`~`CORE-DOC-005`로 기록함.
- 2026-04-29 23:46 KST: `docs/architecture.md` 리뷰 완료. 현재/목표 구현 구분, DB 접근 책임, 실시간 통신 결정, Webhook 수집 파이프라인, 인증/RBAC 단계화 이슈를 `CORE-DOC-006`~`CORE-DOC-010`으로 기록함.
- 2026-04-29 23:52 KST: `docs/tech_stack.md` 리뷰 완료. 미설치 frontend dependency, 런타임 버전 기준, 환경 변수, gRPC 예약 포트, migration 도구 공백을 `CORE-DOC-011`~`CORE-DOC-015`로 기록함.
- 2026-04-29 23:56 KST: P1 항목부터 처리 시작. `CORE-DOC-001`을 `in_progress`로 전환하고 시스템 관리자 역할 정의 방식을 검토함.
- 2026-04-29 23:58 KST: `CORE-DOC-001` 반영 완료. `docs/requirements.md`에 `2.4 시스템 관리자` 역할과 확정 기능을 추가함.
- 2026-04-30 00:01 KST: `CORE-DOC-002` 검토 시작. 후보 기능과 확정 결정의 문서 구조 분리 방식을 검토함.
- 2026-04-30 00:03 KST: `CORE-DOC-002` 반영 완료. `docs/requirements.md`에 `1.1 요구사항 범위 구분`을 추가함.
- 2026-04-30 00:03 KST: `CORE-DOC-004` 검토 및 반영 완료. `docs/requirements.md`에 `4.1 데이터 및 권한 운영 기준` 표를 추가함.
- 2026-04-30 00:03 KST: `CORE-DOC-006` 검토 및 반영 완료. `docs/architecture.md`의 시스템 컴포넌트 다이어그램에 상태 표기 기준과 `current/planned/external` 라벨을 추가함.
- 2026-04-30 00:03 KST: `CORE-DOC-007` 검토 및 반영 완료. Python AI의 DB 직접 접근을 초기 구현에서 제외하고 Go Core 경유 데이터 접근 원칙을 `docs/architecture.md`에 명시함.
- 2026-04-30 00:03 KST: `CORE-DOC-009` 검토 및 반영 완료. `docs/architecture.md`에 `4.2 이벤트 수집 파이프라인`을 추가하고 스토리지 구성 섹션을 `4.3`으로 조정함.
- 2026-04-30 00:03 KST: `CORE-DOC-011` 검토 및 반영 완료. `docs/tech_stack.md`에서 TanStack Query/React Flow를 확정 스택이지만 현재 scaffold에는 미설치인 `도입 예정` 상태로 정리함.
- 2026-04-30 00:03 KST: `CORE-DOC-013` 반영 완료. `docker-compose.yml`의 backend-core 환경 변수에 `GITEA_TOKEN` 전달을 추가함.

## 7. 완료 기준

- 세 문서의 리뷰 체크리스트가 모두 `done`으로 전환된다.
- 발견 사항이 `done` 또는 명시적 `blocked` 상태로 정리된다.
- 원문 문서에 반영한 수정은 링크/명령/상태 정합성 검증을 거친다.
- 세션 종료 전 workflow 상태 문서와 backlog가 최신 리뷰 상태를 가리킨다.
