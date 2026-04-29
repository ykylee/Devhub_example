# DevHub 기술 스택 및 환경 설정 가이드

- **작성일:** 2026-04-29
- **상태:** Finalized
- **관련 문서:** [아키텍처 설계서](./architecture.md), [요구사항 정의서](./requirements.md)

## 1. 확정 기술 스택 (Technology Stack)

DevHub은 Gitea 연동, AI 분석, 실시간 대시보드 제공을 위해 다음과 같은 하이브리드 스택을 사용합니다.

### 1.1 Backend
- **Core Service (Main):** **Go (Gin)**
    - 역할: Gitea API/Webhook 연동, 시스템 제어 로직, 권한 관리.
    - 특징: 고성능 비동기 처리, Gitea와의 언어적 정합성.
- **AI/Analysis Module:** **Python (FastAPI)**
    - 역할: 빌드 로그 분석(AI 가드너), 리스크 탐지 모델 구동.
    - 특징: 풍부한 AI/데이터 분석 생태계 활용. 현재 스캐폴딩은 FastAPI health endpoint와 gRPC 의존성 준비까지 완료되었으며, 실제 gRPC 서버 구현은 후속 작업 범위입니다.
- **Internal Communication:** **gRPC** (Go ↔ Python, 확정 기본 계약)
    - REST/HTTP는 외부 API 및 상태 확인용 endpoint에 사용하며, Go Core와 Python AI 간 분석 요청/응답의 기본 계약은 gRPC로 둡니다.

### 1.2 Frontend
- **Framework:** **Next.js (React 19, App Router)**
    - 역할: 역할별 대시보드, 실시간 상태 시각화.
- **Styling:** **Tailwind CSS**
- **Data Fetching:** **TanStack Query (React Query)** (도입 예정)
    - 상태: 확정 스택이나 현재 scaffold에는 미설치. API 연동 구현 시 추가.
- **Interactive UI:** **React Flow** (인프라 구성도용, 도입 예정)
    - 상태: 확정 스택이나 현재 scaffold에는 미설치. 시스템 관리자 인프라 뷰 구현 시 추가.

### 1.3 Database
- **Main DB:** **PostgreSQL (v15+)**
    - 역할: 정형 데이터 및 JSONB 기반 비정형 분석 결과 저장.

---

## 2. 개발 환경 설정 (Environment Setup)

### 2.1 사전 요구 사항 (Prerequisites)
- **Docker & Docker Compose**
- **Go** (v1.26.2, `backend-core/go.mod` 및 Docker builder 이미지 기준)
- **Python** (v3.10 이상)
- **Node.js** (v20 이상)
- **protoc** (gRPC 컴파일러)
- **Go protoc plugins:** `protoc-gen-go`, `protoc-gen-go-grpc` (`make proto-tools`로 설치)
- **Python gRPC tools:** `grpcio`, `grpcio-tools` (`make setup`으로 `backend-ai/requirements.txt` 설치)

### 2.2 프로젝트 초기화 (Initialization)
루트 디렉토리에서 제공된 `Makefile`을 사용하여 의존성을 설치합니다.

```bash
# 각 프로젝트 의존성과 proto 생성 도구 설치 후 proto 파일 컴파일
make init
```

단계별로 실행할 경우 다음 순서를 사용합니다.

```bash
make setup
make proto-tools
make proto
```

### 2.3 로컬 실행 (Running Locally)
Docker Compose를 사용하여 전체 시스템(DB, Backend, Frontend)을 한 번에 구동합니다.

```bash
# 서비스 빌드 및 실행
make build
make run
```

### 2.4 주요 서비스 포트 정보
- **Frontend:** `http://localhost:3000`
- **Backend Core:** `http://localhost:8080`
- **Backend AI:** `http://localhost:8000`
- **Backend AI gRPC:** `localhost:50051` (예정, 현재 서버 구현 전)
- **PostgreSQL:** `localhost:5432`

---

## 3. 개발 가이드라인

### 3.1 Gitea 연동 설정
`.env` 파일(또는 환경 변수)에 Gitea 서버 정보를 설정해야 합니다.
```env
GITEA_URL=http://your-gitea-server.com
GITEA_TOKEN=your-access-token
```

### 3.2 아키텍처 원칙
- 모든 Gitea 이벤트는 **Go Core**를 통해 먼저 수신되며, 필요한 경우에만 **Python AI**로 전달됩니다.
- 프론트엔드와 백엔드 간의 실시간 연동은 **WebSocket**을 우선적으로 사용합니다.
- 데이터 보존 정책(1개월)을 준수하기 위해 DB 파티셔닝 또는 스케줄링된 삭제 로직을 사용합니다.
