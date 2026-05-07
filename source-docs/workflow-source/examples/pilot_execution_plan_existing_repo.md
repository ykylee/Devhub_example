# 파일럿 실행 계획 예시

- 문서 목적: 실제 파일럿 후보 저장소를 선정한 뒤 첫 적용을 어떤 순서로 진행할지 실행 계획 예시를 제공한다.
- 범위: 후보 선정, 사전 확인, 실행 순서, 산출물, 완료 판단 기준
- 대상 독자: 저장소 관리자, AI workflow 설계자, 프로젝트 온보딩 담당자
- 상태: sample
- 최종 수정일: 2026-04-22
- 관련 문서: `../templates/pilot_candidate_checklist.md`, `../templates/pilot_adoption_record_template.md`, `./pilot_adoption_existing_repo_example.md`

## 1. 파일럿 목표

- `existing` 모드 bootstrap 과 onboarding runner 가 실제 작업 중인 저장소에서도 첫 세션 기준선 복원에 유용한지 확인한다.
- 하네스가 `onboarding_summary`, `warnings`, `orchestration_plan`, `repository_assessment.summary` 를 실사용 가능한 순서로 읽을 수 있는지 본다.
- 자동 추정값 중 사람이 반복적으로 수정하는 항목을 수집해 다음 개선 후보를 좁힌다.

## 2. 후보 선정 판단

- 후보 저장소:
- `Existing Repo` 가상 예시
- 선정 이유:
- 코드와 docs 가 함께 있고 기본 명령을 비교적 쉽게 확인할 수 있다.
- `existing` 모드 검증 가치:
- `repository_assessment.md`, onboarding runner, 하네스 소비 예시를 함께 검증하기 좋다.
- 제외하지 않은 이유:
- 보안 승인 없이는 기본 실행이 불가능한 저장소가 아니고, 문서 허브와 테스트 명령이 최소 수준으로 존재한다.

## 3. 적용 전 준비

- 담당자 지정:
- workflow maintainer 1명
- 저장소 담당자 1명
- 사용할 하네스:
- `codex`
- 먼저 확인할 항목:
- `README.md`
- `docs/README.md`
- 실제 설치/실행/테스트 명령
- 운영 문서 또는 runbook 위치
- 첫 세션에서 확정할 항목:
- project profile 의 문서 경로
- 승인 규칙
- 환경 제약
- backlog/handoff 정착 방식

## 4. 실행 순서

1. `pilot_candidate_checklist.md` 기준으로 후보 적합성을 다시 한 번 확인한다.
2. `bootstrap_workflow_kit.py --adoption-mode existing` 를 실행해 초안 문서를 생성한다.
3. `repository_assessment.md` 의 추정 명령과 문서 경로를 실제 저장소와 대조한다.
4. `run_existing_project_onboarding.py` 를 실행해 첫 세션용 JSON 요약을 만든다.
5. 하네스에서는 `onboarding_summary.recommended_next_steps -> warnings -> orchestration_plan -> repository_assessment.summary` 순서로 읽는다.
6. profile/handoff/backlog 문서에 실제 운영 규칙을 반영해 사람이 확정해야 하는 값을 채운다.
7. `pilot_adoption_record_template.md` 형식으로 적용 기록을 남긴다.

## 5. 기대 산출물

- `ai-workflow/memory/PROJECT_PROFILE.md`
- `ai-workflow/memory/session_handoff.md`
- `ai-workflow/memory/work_backlog.md`
- `ai-workflow/memory/backlog/YYYY-MM-DD.md`
- `ai-workflow/memory/repository_assessment.md`
- onboarding runner 결과 JSON
- 파일럿 적용 기록 문서

## 6. 체크 포인트

- bootstrap 초안 문서 경로가 실제 저장소 구조와 크게 어긋나지 않는가
- onboarding runner 가 latest backlog 미발견 상황에서도 의미 있는 요약을 남기는가
- `repository_assessment.summary` 의 기본 명령이 완전히 틀리지 않고 수정 가능한 수준인가
- 하네스가 `warnings` 와 `orchestration_plan` 을 첫 세션 분배 입력으로 쓰기 쉬운가
- backlog/handoff 문서가 실제 운영 습관에 자연스럽게 들어갈 수 있는가

## 7. 완료 판단 기준

- 첫 세션 시작 전에 무엇을 읽고 무엇을 확정할지 사용자 입장에서 명확해진다.
- 자동 생성 문서와 runner 출력이 사람 검토를 돕는 수준까지는 올라온다.
- 반복적으로 틀리는 추정값과 부족한 가이드가 파일럿 기록에 구체적으로 남는다.
- 다음 파일럿 후보에 적용할 수정 포인트를 3개 이내로 압축할 수 있다.
