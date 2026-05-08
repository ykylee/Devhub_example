# 파일럿 후보 단기 목록 예시

- 문서 목적: 첫 파일럿에 적합한 저장소 후보를 어떻게 1~2개로 좁힐지 예시를 제공한다.
- 범위: 후보 비교, 1순위 선정 이유, 보류 이유, 바로 다음 액션
- 대상 독자: 저장소 관리자, AI workflow 설계자, 프로젝트 온보딩 담당자
- 상태: sample
- 최종 수정일: 2026-04-22
- 관련 문서: `../templates/pilot_candidate_checklist.md`, `./pilot_execution_plan_existing_repo.md`, `./pilot_adoption_existing_repo_example.md`

## 1. 후보 목록

### 후보 A: 서비스 운영형 Node 저장소

- 저장소 성격:
- API 서버와 `docs/` 운영 문서를 함께 가진 작업 중인 서비스 저장소
- 예상 기본 명령:
- `npm install`, `npm run dev`, `npm test`, `npm run test:unit`, `npm run test:smoke`
- 장점:
- `existing` 모드 bootstrap 검증 가치가 크다.
- `repository_assessment.summary` 와 `onboarding_summary.inferred_commands` 비교가 쉽다.
- README 와 docs 허브가 있어 `code-index-update` 추천을 보기 좋다.
- backlog/handoff 도입 후 변화가 눈에 띄기 쉽다.
- 주의점:
- 운영 승인 규칙과 staging 접근 제약은 자동 추정이 어렵다.

### 후보 B: 리서치 운영형 Python 저장소

- 저장소 성격:
- 실험 결과, 배포 리포트, dataset 문서를 함께 가진 평가 운영형 저장소
- 예상 기본 명령:
- `uv sync`, `uv run pytest -q`, `uv run ruff check .`, 리포트 dry-run 계열 명령
- 장점:
- 문서 정합성과 산출물 추적이 중요해 workflow 문서의 가치가 크다.
- `validation-plan`, `code-index-update`, 문서 허브 갱신 흐름이 잘 드러난다.
- 주의점:
- secure runner 또는 데이터 접근 제약이 있으면 첫 파일럿 난도가 올라간다.
- 서비스 운영형보다 하네스 소비 순서 검증보다 문서 정합성 검증 비중이 커진다.

## 2. 1순위 선정

- 1순위 후보:
- 후보 A: 서비스 운영형 Node 저장소
- 선정 이유:
- 첫 파일럿 목표가 `existing` 모드 onboarding runner 와 하네스 소비 순서를 검증하는 데 있기 때문이다.
- `repository_assessment.md` 의 추정 명령을 실제 명령과 대조하기 쉽다.
- docs 허브와 운영 문서가 함께 있어 `code-index-update` 와 `warnings` 의 효용을 보기 좋다.
- 보안/데이터 제약이 리서치 저장소보다 덜할 가능성이 높아 첫 파일럿 리스크가 낮다.

## 3. 보류 이유

- 후보 B 를 2순위로 둔 이유:
- workflow 문서 가치 자체는 높지만, secure runner 나 데이터 접근 제약 때문에 첫 파일럿에서 변수 관리가 어려울 수 있다.
- 첫 파일럿은 마찰 지점 수집이 목적이므로, 하네스 소비와 onboarding 요약 확인이 쉬운 저장소가 더 적합하다.

## 4. 바로 다음 액션

1. 실제 후보 저장소가 후보 A 성격에 맞는지 `pilot_candidate_checklist.md` 기준으로 다시 점검한다.
2. 담당자 2명 이상이 확보되는지 확인한다.
3. `pilot_execution_plan_existing_repo.md` 순서대로 첫 파일럿 실행 계획을 잡는다.
4. 파일럿 종료 후 `pilot_adoption_record_template.md` 형식으로 결과를 남긴다.
