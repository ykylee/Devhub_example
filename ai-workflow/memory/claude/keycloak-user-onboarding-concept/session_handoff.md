# Session Handoff — claude/keycloak-user-onboarding-concept (2026-05-20)

- 문서 목적: Keycloak 사용자 onboarding 컨셉 단계 의사결정 확정 내역 인계.
- 진입 base: main HEAD `63e0157`.
- 최종 수정일: 2026-05-20
- 상태: done

## 이번 세션 결과

- `docs/domain/onboarding/concept.md` 컨셉 결정사항 다수 확정.
- open question 항목 중 5.2/5.3/5.4/5.5/5.6/8/9/10/11/12를 결정 상태로 전환.

## 핵심 결정 요약

1. onboarding 완료 기준은 `onboarding_completed_at`.
2. lazy auto-create 폐기, user row는 onboarding 완료 시 생성/등록.
3. role은 onboarding 입력 불가 (Keycloak role 매핑 또는 기본값 정책).
4. gating은 Backend 강제 + Frontend redirect 보조.
5. organization picker는 hybrid, 검색 2글자/최대 20개/조직명 표시.
6. onboarding 완료 후 관리자 검토 단계 운영(`pending_review`/`reviewed`).
7. `pending_review`는 무소속 제한 접근(할당 리소스 + 공통 메뉴).
8. `/account` self-service 소속 변경 허용, 변경 시 `pending_review` 재진입.
9. i18n은 한국어 고정 + display_name 단일 필드, 영문 필드 확장 여지 유지.
10. 모바일 반응형은 범위 제외, a11y/검증 UX는 필수.
11. 테스트 시드 세트 확정(test 계정 5종 + org fixture bulk).

## 다음 단계

- requirements 문서로 컨셉 결정사항 전개(REQ/AC/edge case 명시).
- API 계약/상태 모델(`pending_review`) 및 접근제어 allowlist 상세 설계.
- traceability row 발급(REQ/ARCH/API/IMPL/UT/TC).
