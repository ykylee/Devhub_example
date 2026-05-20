# Tech Stack

- 문서 목적: DevHub의 공개 가능한 기술 스택 개요를 제공한다.
- 범위: Backend/Frontend/Data 구성과 핵심 아키텍처 노트.
- 대상 독자: 외부 개발자, 기술 검토자.
- 상태: accepted
- 최종 수정일: 2026-05-20

## Backend

- **Go (Gin)**: API, 권한 경계, 외부 연동의 중심 서비스
- **Python (FastAPI + gRPC)**: AI 분석 및 추천 처리
- **gRPC (내부 통신)**: Go ↔ Python 분석 계약

## Frontend

- **Next.js (App Router)**: 대시보드/관리 UI
- **Tailwind CSS**: 스타일링
- **실시간 채널**: WebSocket 기반 상태 업데이트

## Data

- **PostgreSQL**: 정형 데이터 + JSONB 기반 이벤트/요약 저장

## Architectural Notes

- 브라우저 ↔ 서버 실시간 계약은 WebSocket을 우선 사용
- 시스템 운영 기능은 권한 정책으로 분리
- 인증/권한 모델은 ADR 기반으로 단계적으로 진화
