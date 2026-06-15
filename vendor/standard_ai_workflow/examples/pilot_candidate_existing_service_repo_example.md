# 파일럿 후보 체크리스트 예시

- 문서 목적: 첫 파일럿 1순위 후보를 정했을 때 선정 체크리스트를 어떻게 채우는지 예시를 제공한다.
- 범위: 후보 적합성 판단, 리스크, 바로 진행 가능한 다음 액션
- 대상 독자: 저장소 관리자, AI workflow 설계자, 프로젝트 온보딩 담당자
- 상태: sample
- 최종 수정일: 2026-04-22
- 관련 문서: `../templates/pilot_candidate_checklist.md`, `./pilot_candidate_shortlist_example.md`, `./pilot_execution_plan_existing_repo.md`

## 1. 후보 기본 정보

- 후보 이름:
- 서비스 운영형 Node 저장소 가상 예시
- 후보 유형:
- 작업 중인 프로젝트, `existing` 모드 대상
- 예상 하네스:
- `codex`
- 예상 기본 스택:
- `node`

## 2. 선정 기준 점검

- 문서와 코드가 모두 있어 `existing` 모드 검증 가치가 큰가:
- `yes`
- 기본 실행 명령과 빠른 테스트 명령을 확인할 수 있는가:
- `yes`
- backlog/handoff 성격 문서를 넣어도 운영 마찰이 과도하지 않은가:
- `partial`
- README 또는 docs 허브가 있어 `code-index-update` 검증 가치가 있는가:
- `yes`
- 첫 파일럿에서 권한, 보안, 외부 의존이 지나치게 많지 않은가:
- `yes`
- 변경 빈도가 너무 높지 않아 적용 전후 비교를 남기기 쉬운가:
- `yes`

## 3. 리스크 점검

- 승인 규칙이 자동 추정으로 해결되는가:
- `no`
- 운영 문서 위치가 완전히 고정돼 있는가:
- `partial`
- 첫 세션에서 사람이 직접 채워야 할 값이 많은가:
- `yes`
- 보류가 필요한 수준의 blocker 가 있는가:
- `no`

## 4. 선정 판단

- 최종 판단:
- `go`
- 선정 이유:
- onboarding runner, `repository_assessment.summary`, 하네스 소비 순서, docs 허브 재확인 흐름을 함께 보기 좋은 가장 무난한 유형이다.
- 첫 파일럿에서 확인할 핵심:
- `repository_assessment.md` 추정 명령 정합성
- `onboarding_summary.recommended_next_steps` 의 실사용성
- `warnings` 와 `orchestration_plan` 의 하네스 소비 편의성
- backlog/handoff 문서 도입 마찰

## 5. 바로 다음 액션

1. 담당자 2명 이상을 확보한다.
2. 실제 설치/실행/테스트 명령을 확정한다.
3. `pilot_execution_plan_existing_repo.md` 순서대로 첫 파일럿 실행 계획을 잡는다.
4. 파일럿 종료 후 `pilot_adoption_record_template.md` 형식으로 결과를 남긴다.
