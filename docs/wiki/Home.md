# DevHub Wiki

DevHub는 개발/관리/시스템 운영을 하나의 제품 흐름으로 연결하는 팀 허브입니다.  
핵심 UX 원칙은 **역할별 기본 진입 우선순위**와 **권한 기반 노출 제어**입니다.

## 핵심 원칙

1. 역할별 UX는 전용 화면 완전 분리보다 **기본 진입 우선순위**로 제공한다.
2. 시스템 운영 기능은 **`system_admin` 권한 사용자에게만 노출**한다.
3. 문서 계약(요구사항/아키텍처/권한 모델)과 런타임 동작이 일치해야 한다.

## 빠른 시작

- [Quickstart](./guides/quickstart)
- [Tech Stack](./reference/tech-stack)
- [Publication Scope](./meta/publication-matrix)
- [Publishing Checklist](./meta/publishing-checklist)

## 추천 읽기

- [역할별 UX를 진입 우선순위로 설계한 이유](./columns/001-role-priority-ux)
- [시스템 관리자 권한 경계를 분리한 이유](./columns/002-system-admin-boundary)
- [Auth/RBAC 의사결정에서 얻은 교훈](./columns/003-adr-lessons-auth-rbac)

## 문서 정책

- 내부 운영 문서(`ai-workflow/memory/*`)는 공개 위키에 직접 게시하지 않습니다.
- 대외 공개 문서는 `docs/wiki/` 편집본을 기준으로 게시합니다.
