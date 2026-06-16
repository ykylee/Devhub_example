# Work Backlog — codex/work_260612-579-ci-rearchitecture

- 문서 목적: #579 CI 재구성 및 flaky 복구 스프린트의 현재 진행상황을 관리한다.
- 상태: **complete**
- 최종 수정일: 2026-06-13

## 완료

- CI 재구성 설계 문서 작성
- workflow 3분리 구현
  - `ci.yml`
  - `e2e-regression.yml`
  - `e2e-quarantine.yml`
- manifest 기반 spec selection 도입
- signout/login flaky 복구용 fixture/spec 조정
- workflow sync / selector / focused type check 검증

## 완료 결과

1. PR #580 merged
2. 새 workflow 구조의 GitHub Actions 동작 확인
3. PR #579 최신 main 반영 후 재평가 통과 확인

## 메모

- `voc-auto-routing.spec.ts` 는 현재 branch baseline 에 존재하지 않으므로 quarantine manifest 에서 자동 skip 되도록 selector 가 처리한다.
- full frontend `tsc --noEmit` 실패는 기존 선행 오류라서 이번 PR의 accept/reject 신호로 직접 사용하지 않는다.
