# 파일럿 적용 기록 예시

- 문서 목적: 표준 AI 워크플로우를 작업 중인 실제 저장소에 시범 적용할 때 어떤 기록을 남기면 되는지 예시를 제공한다.
- 범위: 파일럿 적용 전후 비교, 운영 피드백, 후속 수정 포인트 예시
- 대상 독자: 저장소 관리자, AI workflow 설계자, 프로젝트 온보딩 담당자
- 상태: sample
- 최종 수정일: 2026-04-22
- 관련 문서: `../templates/pilot_adoption_record_template.md`, `../templates/pilot_candidate_checklist.md`, `../core/workflow_adoption_entrypoints.md`

## 1. 적용 대상

- 적용일:
- `2026-04-22`
- 대상 저장소:
- `Existing Repo` 가상 예시
- 저장소 성격:
- Node 기반 서비스 애플리케이션과 운영 문서를 함께 가진 작업 중인 프로젝트
- 기본 스택:
- `node`
- 하네스:
- `codex`
- 적용 모드:
- `existing`
- 적용 담당:
- workflow maintainer

## 2. 적용 전 상태

- 기존 운영 문서 위치:
- `README.md`, `docs/README.md` 는 있었지만 handoff/backlog 문서는 없었다.
- 기존 세션/handoff 관행:
- 이슈 댓글과 PR 설명에 산발적으로 남겨 세션 기준선 복원이 어려웠다.
- 기존 backlog 관행:
- 날짜별 작업 로그는 없고, 작업은 이슈 보드와 채팅에 분산되어 있었다.
- 확인한 기본 명령:
- 설치:
- `npm install`
- 실행:
- `npm run dev`
- 빠른 테스트:
- `npm test`
- 격리 테스트:
- `npm run test:unit`
- smoke 확인:
- `npm run test:smoke`
- 적용 전에 이미 있던 제약:
- 운영 승인 규칙과 staging 접근 조건은 repository assessment 에서 자동 추정할 수 없어 사람이 보강해야 했다.

## 3. 적용 범위

- 생성/복사한 문서:
- `ai-workflow/memory/PROJECT_PROFILE.md`
- `ai-workflow/memory/session_handoff.md`
- `ai-workflow/memory/work_backlog.md`
- `ai-workflow/memory/backlog/2026-04-22.md`
- `ai-workflow/memory/repository_assessment.md`
- 적용한 하네스 오버레이:
- `AGENTS.md`
- `.codex/config.toml.example`
- 실행한 스크립트:
- `python3 scripts/bootstrap_workflow_kit.py --adoption-mode existing`
- `python3 scripts/run_existing_project_onboarding.py ...`
- 참고한 core 문서:
- `core/workflow_adoption_entrypoints.md`
- `core/existing_project_onboarding_contract.md`
- `core/output_schema_guide.md`

## 4. 적용 중 관찰

- 잘 맞았던 점:
- repository assessment 가 기본 명령과 상위 디렉터리 구조를 빠르게 초안화해 첫 세션 진입 비용을 줄였다.
- 수정이 필요했던 경로/명칭:
- docs 홈은 맞았지만 운영 문서 위치와 backlog 용어는 팀 실제 표현에 맞게 바꿔야 했다.
- 헷갈렸던 입력:
- `repository_assessment.summary` 는 유용했지만 승인 규칙, staging 제약 같은 운영 정보는 별도 보강이 필요했다.
- 과도하거나 불필요했던 규칙:
- 첫 파일럿 단계에서는 worker 분배 메타데이터가 약간 과할 수 있었지만 하네스 소비 예시로는 도움이 됐다.
- 부족했던 가이드:
- 하네스가 onboarding 결과를 어떤 순서로 읽어야 하는지 예시가 없으면 처음에는 헷갈릴 수 있었다.
- 수동 확인이 많이 필요했던 지점:
- profile 문서의 예외 규칙, 승인 규칙, 실제 운영 문서 경로 확정은 사람이 직접 마무리해야 했다.

## 5. 적용 후 상태

- 최종 project profile 경로:
- `ai-workflow/memory/PROJECT_PROFILE.md`
- 최종 session handoff 경로:
- `ai-workflow/memory/session_handoff.md`
- 최종 work backlog 경로:
- `ai-workflow/memory/work_backlog.md`
- 최신 backlog 경로:
- `ai-workflow/memory/backlog/2026-04-22.md`
- 실제 첫 세션에서 사용한 runner 또는 skill:
- `run_existing_project_onboarding.py`
- 하네스에서 우선 읽은 출력:
- `onboarding_summary.recommended_next_steps`
- `warnings`
- `orchestration_plan`
- `repository_assessment.summary`
- 남은 미정 항목:
- 승인 규칙, 운영 리뷰 주체, staging 접근 제약, backlog 정착 방식

## 6. 검증 결과

- 실행한 smoke/test:
- `python3 tests/check_bootstrap.py`
- `python3 tests/check_existing_project_onboarding.py`
- `python3 tests/check_docs.py`
- 실행하지 못한 항목과 사유:
- 실제 대상 저장소가 아니라 가상 예시라 하네스 내부 운영 습관까지는 검증하지 못했다.
- 출력 계약 문제 여부:
- 대표 sample 과 runner 출력 구조는 일치했다.
- 문서 링크/메타데이터 문제 여부:
- 생성 문서 링크와 메타데이터는 기본 smoke 기준을 통과했다.
- runner 또는 하네스 연결 문제 여부:
- assessment 가 포함된 onboarding sample 이 없으면 하네스 소비 예시가 약했으나, 샘플 보강 후 명확해졌다.

## 7. 적용 전후 비교

- 도입 전 대비 좋아진 점:
- 첫 세션에서 읽을 문서와 다음 작업 순서가 JSON 요약으로 정리돼 진입이 빨라졌다.
- 여전히 수동인 부분:
- 팀 승인 규칙, 운영 예외, 문서 위치 최종 확정은 사람이 채워야 한다.
- 예상보다 비용이 컸던 부분:
- 실제 운영 문서 체계와 template/backlog 용어를 맞추는 데 소규모 조정이 필요했다.
- 다음 파일럿 전에 꼭 고칠 점:
- assessment 포함 onboarding sample 과 하네스 소비 예시를 더 전면에 배치한다.

## 8. 최종 판단

- 재적용 의향:
- `yes`
- 권장 도입 범위:
- 기존 코드와 docs 구조가 있고 기본 테스트 명령을 확인할 수 있는 저장소의 첫 도입
- 권장 제외 범위:
- 승인 체계가 불명확하고 운영 문서가 거의 없는 저장소의 즉시 전면 도입
- 다음 후속 작업:
- 실제 파일럿 후보 저장소 1개를 선정해 같은 포맷으로 운영 피드백을 남긴다.
