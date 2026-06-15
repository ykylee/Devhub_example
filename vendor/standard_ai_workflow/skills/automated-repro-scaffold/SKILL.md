# Skill: automated-repro-scaffold

- 문서 목적: 버그 재현용 독립 샌드박스 스크립트 자동 생성 스킬의 사용법과 기능을 설명한다.
- 범위: 스킬 실행 방법, 주요 기능, 프로토타입 실행 예시
- 대상 독자: AI 에이전트, 개발자
- 상태: prototype
- 최종 수정일: 2026-04-27
- 관련 문서: `core/automated_repro_scaffold_skill_spec.md`, `core/workflow_skill_catalog.md`

## 실행 방법 (Prototype)

```bash
python3 skills/automated-repro-scaffold/automated_repro_scaffold.py \
  --report "bug_report.md" \
  --output "tests/repro_issue_1.py"
```

## 주요 기능
- **이슈 분석**: 텍스트 기반 버그 리포트에서 재현 핵심 로직 추출.
- **코드 스캐폴딩**: `unittest` 기반의 독립 실행형 테스트 코드 생성.
- **의존성 주입**: 필요한 경우 가짜 데이터(Mock)나 최소한의 환경 설정 포함.

## 판단 기준 (Checklist)
- [ ] 스크립트가 다른 파일에 의존하지 않고 독립적으로 실행 가능한가?
- [ ] 버그 리포트의 에러 상황을 `assert` 또는 예외 발생으로 재현하는가?
- [ ] 실행 명령어가 명확히 제공되는가?
