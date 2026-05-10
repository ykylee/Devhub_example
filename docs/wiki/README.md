# DevHub Public Wiki Source

- 문서 목적: GitHub Wiki에 게시할 대외 공유용 칼럼/가이드 원고를 관리한다.
- 범위: 공개 대상 칼럼, 공개형 가이드, 공개형 레퍼런스 요약본
- 대상 독자: 외부 개발자, 잠재 사용자, 기술 블로그 독자
- 상태: active
- 최종 수정일: 2026-05-10
- 관련 문서: [분류 기준](./meta/publication_matrix.md), [작성 가이드](./meta/style_guide.md), [내부 docs 인덱스](../docs/DOCUMENT_INDEX.md)

## 운영 원칙

1. `docs/`는 내부 source-of-truth로 유지한다.
2. `docs/wiki/`는 대외 공개용 편집본만 관리한다.
3. Wiki 게시는 `docs/wiki/`에서만 수행하고, `docs/`를 직접 복사 게시하지 않는다.
4. 보안/운영 민감 정보(내부 경로, 계정/환경 상세, 임시 의사결정)는 게시 전 제거한다.

## 디렉터리

- `columns/`: 칼럼형 아티클 초안
- `guides/`: 대외용 사용/도입 가이드
- `reference/`: 공개 가능한 계약/개념 요약
- `meta/`: 게시 정책, 분류표, 편집 규칙
