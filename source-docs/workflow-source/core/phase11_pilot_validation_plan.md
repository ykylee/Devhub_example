# Phase 11: Real-world Pilot Validation Plan

- 문서 목적: Phase 10까지 구축된 다중 에이전트 인프라와 스킬들을 실제 프로젝트에 적용하여 실전성을 검증하고 튜닝한다.
- 범위: 파일럿 프로젝트 선정, 검증 시나리오, 성공 기준, 진척 관리
- 대상 독자: 저장소 관리자, AI 에이전트
- 상태: in_progress
- 최종 수정일: 2026-05-04

## 1. 파일럿 대상 및 환경
- **대상 프로젝트**: `Devhub_example_gemini` (또는 유사한 구조의 시뮬레이션 환경)
- **주요 목표**: 문서 정합성 100% 달성, 에이전트 간 위임 루프 안정화

## 2. 검증 시나리오

### 시나리오 A: 문서 기반 정합성 대청소 (Clean Slate)
1. **[steward]**: 프로젝트 내 모든 TASK 문서의 헤더와 관련 문서 링크를 최신 표준으로 업데이트.
2. **[linter]**: `state.json`과 `session_handoff.md`, `backlog` 간의 불일치를 탐지.
3. **[apply]**: Linter의 자동 수정을 통해 `state.json` 동기화.

### 시나리오 B: 에이전트 간 리스크 전파 (Feedback Loop)
1. **[orchestrator]**: 신규 기능 구현을 `code-worker`에게 위임.
2. **[code-worker]**: 구현 중 의존성 충돌 리스크 보고.
3. **[orchestrator]**: 리스크를 인지하고 `validation-worker`에게 로그 분석 위임.
4. **[validation-worker]**: 해결책 제시 및 검증 결과 보고.

### 시나리오 C: 충돌 상황의 지능형 해결 (Contextual Git)
1. **[conflict]**: 강제로 Git 충돌 상황 발생.
2. **[resolver]**: `session_handoff.md`의 현재 문맥을 읽고 자동으로 최적의 버전 선택 제안.

## 3. 성공 기준 (Success Criteria)
- [ ] 모든 자동화 스킬의 실행 결과가 Pydantic 모델을 100% 준수함.
- [ ] Linter 실행 후 `total_issues`가 0으로 수렴함.
- [ ] 에이전트의 결정 근거(`resolution_note`, `risk_identified`)가 사람이 이해하기에 충분히 명확함.

## 4. 진척 관리 (Progress Tracking)

### 단계 1: 파일럿 환경 준비
- [x] Phase 11 브랜치 생성 및 환경 설정
- [x] 파일럿 전용 더미 데이터 세트 구축 (`pilot_env/`)

### 단계 2: 시나리오 실행 및 튜닝
- [x] 시나리오 A (Linter & Steward) 실행 및 피드백 반영
- [x] 시나리오 B (Feedback Loop) 시뮬레이션 실행 (`orchestrator_feedback_loop_demo.py`)
- [x] 시나리오 C (Git Resolver) 실전 데이터 테스트

### 단계 3: 결과 요약 및 Phase 11 종료
- [ ] 파일럿 결과 보고서 작성
- [ ] 다음 단계(Phase 12: 패키지 승격) 이행 판단
