# 개발 문서 리뷰 및 정리 계획 (2026-04-29)

- 목적: 현재 작성된 개발 문서 전체(`README.md`, `docs/`)의 불일치와 구현 대비 과장 표현을 추적하고 순차적으로 해결한다.
- 상태: done
- 범위: `README.md`, `docs/README.md`, `docs/requirements.md`, `docs/architecture.md`, `docs/tech_stack.md`, `docs/assessment.md`
- 관련 작업: `TASK-008 개발 문서 리뷰 결과 정리 및 수정`

## 1. 리뷰 요약

현재 개발 문서는 제품 방향과 주요 기술 선택을 설명하는 데에는 충분하지만, 일부 문서가 실제 스캐폴딩 상태보다 구현이 더 완료된 것처럼 읽히거나, 서로 다른 문서가 다른 계약을 말하는 문제가 있다.

우선 수정 순서는 실행 가능성에 직접 영향을 주는 환경/명령 문서, 구현 상태 오해를 만드는 아키텍처 문서, stale 문서, 탐색성 개선 순으로 둔다.

## 2. 해결 항목

### DOC-REVIEW-001 초기화 순서와 proto 의존성 정리

- 상태: done
- 우선순위: P1
- 대상 문서: `docs/tech_stack.md`, 필요 시 `Makefile`
- 문제:
  - `docs/tech_stack.md`는 `make proto`를 `make setup`보다 먼저 실행하라고 안내한다.
  - `make proto`의 Python 생성 단계는 `grpc_tools`가 먼저 설치되어 있어야 한다.
  - Go 쪽 `protoc-gen-go`, `protoc-gen-go-grpc` 설치도 prerequisites에 빠져 있다.
- 완료 기준:
  - 새 환경에서 따라 할 설치 순서가 실제 의존성과 맞는다.
  - `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, Python gRPC tools 준비 조건이 문서화된다.
- 처리 결과:
  - `Makefile`에 `init`, `proto-tools` 타깃을 추가하고 초기화 순서를 `setup -> proto-tools -> proto`로 정리함.
  - `docs/tech_stack.md`에 Go/Python proto 도구 사전 조건과 단계별 실행 순서를 명시함.

### DOC-REVIEW-002 Go 버전 계약 정리

- 상태: done
- 우선순위: P1
- 대상 문서: `docs/tech_stack.md`, `backend-core/go.mod`, `backend-core/Dockerfile`
- 문제:
  - 문서는 Go v1.21 이상을 안내한다.
  - `backend-core/go.mod`는 `go 1.26.2`를 요구한다.
  - `backend-core/Dockerfile`은 `golang:1.24-alpine`을 사용한다.
- 완료 기준:
  - 문서, `go.mod`, Dockerfile의 Go 버전 계약이 하나로 맞춰진다.
  - Docker build 재현성을 해치는 버전 불일치가 제거된다.
- 처리 결과:
  - `docs/tech_stack.md`, `backend-core/go.mod`, `backend-core/Dockerfile`의 Go 계약을 `1.26.2` 기준으로 맞춤.

### DOC-REVIEW-003 내부 통신 방식 확정 표현 통일

- 상태: done
- 우선순위: P2
- 대상 문서: `docs/requirements.md`, `docs/architecture.md`, `docs/tech_stack.md`
- 문제:
  - `requirements.md`에는 Go Core와 Python AI 통신이 `gRPC 또는 REST API`로 남아 있다.
  - `architecture.md`와 `tech_stack.md`는 gRPC를 확정 스택으로 설명한다.
- 완료 기준:
  - 내부 통신의 기본 계약을 gRPC로 통일한다.
  - REST API가 fallback 또는 운영 API인지 여부를 명확히 구분한다.
- 처리 결과:
  - `docs/requirements.md`, `docs/architecture.md`, `docs/tech_stack.md`에서 Go Core ↔ Python AI 기본 계약을 gRPC로 통일함.
  - REST/HTTP는 외부 API, 프론트엔드 API, 상태 확인 endpoint 용도로 구분함.

### DOC-REVIEW-004 gRPC 구현 상태와 예정 범위 구분

- 상태: done
- 우선순위: P2
- 대상 문서: `docs/architecture.md`, `docs/tech_stack.md`
- 문제:
  - 문서상 Go Core와 Python AI의 gRPC/ProtoBuf 통신이 현재 구현된 구조처럼 읽힌다.
  - 실제 `backend-ai/main.py`는 FastAPI HTTP 서버만 실행하고, `50051` gRPC 서버는 아직 없다.
- 완료 기준:
  - 현재 구현 완료 범위와 예정 범위가 문서에서 분리된다.
  - `50051` 포트는 예정 또는 TODO 상태임이 명확히 표시된다.
- 처리 결과:
  - `docs/architecture.md`와 `docs/tech_stack.md`에 현재 FastAPI health endpoint만 구현되어 있고 gRPC 서버는 후속 구현 범위임을 명시함.

### DOC-REVIEW-005 저장소 분석 리포트 최신화

- 상태: done
- 우선순위: P2
- 대상 문서: `docs/assessment.md`
- 문제:
  - 저장소 분석 리포트가 `Devhub Example Gemini`, `unknown`, 소스/문서/테스트 디렉터리 없음, TODO 명령 상태로 남아 있다.
  - 현재 저장소에는 Go/Python/Next.js/Docker/docs/ai-workflow 구조가 들어와 있다.
- 완료 기준:
  - 분석 대상, 감지 스택, 디렉터리 구조, 실행/검증 명령이 현재 저장소와 맞는다.
  - stale 리포트가 온보딩 문서로 다시 신뢰 가능한 상태가 된다.
- 처리 결과:
  - `docs/assessment.md`의 프로젝트명, 감지 스택, 주요 경로, 실행/검증 명령을 현재 저장소 기준으로 갱신함.

### DOC-REVIEW-006 요구사항 문서 섹션 번호 정리

- 상태: done
- 우선순위: P3
- 대상 문서: `docs/requirements.md`
- 문제:
  - `## 5. 핵심 기획 아젠다` 아래 섹션이 `4.1`, `4.2`, `4.3`으로 번호가 맞지 않는다.
  - 이후 `## 9. 기술 스택 결정 사항`으로 점프한다.
- 완료 기준:
  - 섹션 번호가 문서 구조와 일치한다.
  - 향후 리뷰와 링크 참조가 안정적으로 가능하다.
- 처리 결과:
  - `docs/requirements.md`의 Agenda 섹션을 `5.x`, 기술 스택/아키텍처 섹션을 `6`, `7`로 정리함.

## 3. 권장 처리 순서

1. DOC-REVIEW-001, DOC-REVIEW-002: done
2. DOC-REVIEW-003, DOC-REVIEW-004: done
3. DOC-REVIEW-005: done
4. DOC-REVIEW-006: done

## 4. 검증 기준

- 문서 링크가 깨지지 않는다.
- 문서의 실행 명령이 현재 `Makefile` 및 실제 의존성과 충돌하지 않는다.
- 구현 완료 범위와 예정 범위가 분리되어 읽힌다.
- 수정 후 `state.json`, `session_handoff.md`, 최신 backlog가 현재 작업 상태와 동기화된다.
