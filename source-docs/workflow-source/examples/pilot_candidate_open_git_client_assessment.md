# open_git_client 파일럿 후보 평가

- 문서 목적: `open_git_client` 저장소를 표준 AI 워크플로우 첫 실제 파일럿 후보로 볼 수 있는지 평가 결과를 정리한다.
- 범위: 후보 적합성, 강점, 리스크, 바로 다음 실행 액션
- 대상 독자: 저장소 관리자, AI workflow 설계자, 프로젝트 온보딩 담당자
- 상태: sample
- 최종 수정일: 2026-04-22
- 관련 문서: `../templates/pilot_candidate_checklist.md`, `./pilot_candidate_existing_service_repo_example.md`, `./pilot_execution_plan_existing_repo.md`

## 1. 기본 정보

- 대상 저장소:
- `/home/yklee/repos/open_git_client`
- 원격:
- `https://github.com/ykylee/open_git_client.git`
- 현재 브랜치:
- `codeserver`
- 작업 트리 상태:
- 로컬 수정 없음
- 주요 스택:
- `C++20`
- `CMake + Ninja`
- `wxWidgets`
- `libgit2`

## 2. 선정 기준 점검

- 문서와 코드가 모두 있어 `existing` 모드 검증 가치가 큰가:
- `yes`
- 판단 근거:
- `README.md`, `docs/requirements/*`, `docs/design/*`, `src/`, `tests/unit/` 구조가 함께 존재한다.

- 기본 실행 명령과 빠른 테스트 명령을 확인할 수 있는가:
- `yes`
- 판단 근거:
- `README.md` 와 `CMakeLists.txt` 에 `cmake -S . -B build -G Ninja`, `cmake --build build`, `ctest --test-dir build --output-on-failure` 가 명시되어 있다.

- backlog/handoff 성격 문서를 넣어도 운영 마찰이 과도하지 않은가:
- `partial`
- 판단 근거:
- 현재는 `docs/thread-operating-guidelines.md` 같은 문서화 규칙은 있지만, session handoff/work backlog 성격 문서는 없다.

- README 또는 docs 허브가 있어 `code-index-update` 검증 가치가 있는가:
- `yes`
- 판단 근거:
- 루트 `README.md` 와 `docs/requirements`, `docs/design` 구조가 분명하다.

- 첫 파일럿에서 권한, 보안, 외부 의존이 지나치게 많지 않은가:
- `yes`
- 판단 근거:
- 로컬 Linux 기준 빌드/테스트 경로가 문서화돼 있고, 외부 secure runner 의존은 보이지 않는다.

- 변경 빈도가 너무 높지 않아 적용 전후 비교를 남기기 쉬운가:
- `yes`
- 판단 근거:
- 현재 단계가 초기 구현 착수/앱 골격 구성 수준이라 문서 구조와 기본 명령 비교가 상대적으로 안정적이다.

## 3. 강점

- 첫 파일럿에서 `repository_assessment.summary` 의 명령 추정 정확도를 검증하기 좋다.
- `README.md` 가 비교적 잘 정리돼 있어 bootstrap 생성 문서와 실제 구조 비교가 쉽다.
- `CMakeLists.txt` 와 `tests/unit/` 구조 덕분에 validation-plan 초안의 품질을 보기 좋다.
- 요구사항/설계 문서는 있지만 운영 문서는 약해서, backlog/handoff 도입 효과를 관찰하기 좋다.

## 4. 리스크

- 현재 운영 문서 기준선이 `docs/thread-operating-guidelines.md` 정도로 제한되어 있어 handoff/backlog 도입 시 용어 조정이 필요할 수 있다.
- Windows 중심 프로젝트라 Linux 빌드 기준과 실제 주 개발 환경 사이 차이를 profile 문서에서 명시해야 한다.
- `codeserver` 브랜치가 기준 브랜치이므로, 파일럿 시 branch 전략과 문서 반영 위치를 먼저 합의해야 한다.

## 5. 최종 판단

- 최종 판단:
- `go`
- 우선순위:
- 첫 실제 파일럿 1순위 후보
- 선정 이유:
- 코드, 문서, 빌드 명령, 테스트 구조가 모두 있어 `existing` 모드 bootstrap + onboarding runner 검증 가치가 높다.
- 보류하지 않은 이유:
- 외부 보안 제약이 상대적으로 적고, README 기반으로 자동 추정이 크게 어긋날 위험이 낮다.

## 6. 바로 다음 액션

1. `open_git_client` 저장소에 대해 `existing` 모드 bootstrap 대상 경로와 생성물 위치를 정한다.
2. `README.md`, `docs/thread-operating-guidelines.md`, `docs/requirements/SRS.md` 를 기준으로 profile 에 들어갈 문서 경로/검증 포인트를 확정한다.
3. 첫 파일럿에서는 `codex` 하네스를 기준으로 `onboarding_summary -> warnings -> orchestration_plan -> repository_assessment.summary` 소비 흐름을 검증한다.
4. 파일럿 종료 후 `pilot_adoption_record_template.md` 형식으로 실제 마찰 지점을 기록한다.
