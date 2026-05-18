# Session Handoff — adr0018-reverse-proxy

- **상태**: ✅ Done (구현 및 검증 완료)
- **최종 수정일**: 2026-05-18
- **현재 브랜치**: `adr0018-reverse-proxy`

---

## 1. 개요 및 완결 사항
본 세션에서는 **ADR-0018: 단일 외부 포트 역프록시 구성 및 `/devhub` sub-path prefix 아키텍처**를 완전히 구현하고 최종 검증을 완료했습니다.

1. **아키텍처 결정서 수립**: [0018-single-port-reverse-proxy-policy.md](file:///Users/yklee/repos/Devhub_example_gemini/docs/adr/0018-single-port-reverse-proxy-policy.md) (ADR-0018)를 작성하여, 외부 접속 단일 포트(80/443), `/devhub` 서브패스 격리, Same-Origin CORS 무력화 이점, 쿠키 고립, TLS 종료 및 proxy trust, cutover/rollback SOP 가이드라인을 명문화했습니다.
2. **Nginx 설정 골격 구축**: [devhub.conf](file:///Users/yklee/repos/Devhub_example_gemini/infra/nginx/devhub.conf)를 신규 작성하여 HSTS/보안헤더 탑재, `/devhub/api/` strip proxy, Kratos/Hydra public endpoints 및 Next.js/WebSocket default 라우팅을 완비했습니다.
3. **프론트엔드 dynamic basePath 및 endpoints 런타임 이식성 장착**:
   - [next.config.ts](file:///Users/yklee/repos/Devhub_example_gemini/frontend/next.config.ts): `NEXT_PUBLIC_BASE_PATH` 감지 시 `basePath: "/devhub"`를 동적으로 지정하고 로컬 개발 모드(basePath가 없을 때)에서만 rewrites(CORS bypass)가 켜지게 이중 가드를 구축했습니다.
   - [endpoints.ts](file:///Users/yklee/repos/Devhub_example_gemini/frontend/lib/config/endpoints.ts): OIDC_REDIRECT_URI와 WS_BASE_URL 등이 브라우저 런타임(`window.location.origin` 및 `window.location.host`)을 기반으로 실시간으로 동적 조립되도록 고이식성 dynamic resolution을 장착하여, 빌드 시점에 도메인이 영구 inline화 되는 문제를 해소했습니다.
4. **배포 가이드라인 보강**: [test-server-deployment.md](file:///Users/yklee/repos/Devhub_example_gemini/docs/setup/test-server-deployment.md) 파일에 Nginx reverse proxy 연동 가이드를 수립했습니다.
5. **추적성 동기화**: [report.md](file:///Users/yklee/repos/Devhub_example_gemini/docs/traceability/report.md)에 ADR-0018 인덱스 등록 및 본 스프린트 변경 이력을 추가하여 거버넌스를 완수했습니다.

---

## 2. 검증 완료 내역
- **TypeScript 타입 분석**: `npx tsc --noEmit` -> **100% PASS** (에러 0건)
- **Frontend Vitest**: `npx vitest run` -> **100% PASS** (48개 케이스 모두 성공)
- **Backend Unit/Integration Tests**: `go test -v ./internal/...` -> **100% PASS** (전원 통과)

---

## 3. 다음 작업 지시사항 (Next Directives)
1. **브랜치 병합**: `adr0018-reverse-proxy` 피처 브랜치의 변경 사항을 `main` 브랜치에 안전하게 Merge(또는 Pull Request) 합니다.
2. **Keycloak SSO 통합**: 다음 마일스톤인 ADR-0019 (Keycloak SSO Federation) 분석 및 연동 작업을 순차적으로 기획/진행할 수 있습니다.
