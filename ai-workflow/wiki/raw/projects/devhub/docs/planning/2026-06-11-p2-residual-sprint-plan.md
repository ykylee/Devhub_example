# P2 잔여 5건 일괄 처리 Sprint Plan

- **작성일**: 2026-06-11
- **목적**: v0.1.0 출시 이후 P2 잔여 이슈 5건을 일괄 처리하기 위한 통합 스프린트 계획 문서. 각 이슈는 본 계획에 따라 후속 스프린트에서 구현됩니다.
- **대상 이슈**: #386, #382, #381, #380, #383
- **워커**: Claude (backend/CI), Gemini (frontend)

## 1. Issue #386: Admin Catalog 모달 AnimatePresence key 수정
- **도메인**: `application-lifecycle` / `ui-foundation`
- **목표**: `/admin/catalog` 의 탭 전환 불가 현상(오버레이 DOM 잔존) 해결
- **작업 범위**:
  - `app/(dashboard)/admin/catalog/page.tsx` 의 `<AnimatePresence>` 내 조건부 자식들에 고유 `key` 속성 부여.
  - 모달 root 닫힘 시 언마운트 정상 동작 확인.
- **예상 소요**: 1시간
- **검증**: `frontend` vitest 수행 및 수동 브라우저 UI 검증.

## 2. Issue #382: Migration Prefix Uniqueness CI guard 강화
- **도메인**: `database-migration` / `ci`
- **목표**: 마이그레이션 파일 접두어(e.g., `000042_`) 중복 발생 시 CI가 사전에 차단하도록 가드 강화.
- **작업 범위**:
  - `scripts/check-migration-uniqueness.sh` 등 쉘 스크립트에 `uniq -d` 검사 강화.
  - `.github/workflows/ci.yml` 에 admin bypass 방지용 main push 트리거 검사 추가 또는 가이드 링크 명시.
- **예상 소요**: 1시간
- **검증**: 의도적으로 중복된 migration 파일 커밋 후 CI Job 실패 확인.

## 3. Issue #381: 프론트 단위테스트 보강 6종
- **도메인**: `ui-foundation` / `test`
- **목표**: 프론트엔드 핵심 서비스/컴포넌트의 단위테스트 보강 (최소 6종).
- **작업 범위**:
  - `integration.service` (SCM repo list/import/create 등)
  - `repository.service` (createRepositoryDraft / publish)
  - `integration-provider-presets` 추가분 커버
  - 신규 모달 3종: `CreateScmRepositoryModal`, `ImportRepositoriesModal` 등 렌더링/콜백 테스트.
- **예상 소요**: 2시간
- **검증**: `npm run test` 통과, coverage 상승 확인.

## 4. Issue #380: N-2/N-3 Repository UT & SCM E2E
- **도메인**: `repository-integration`
- **목표**: draft→publish 흐름과 SCM import/create/publish happy-path 무테스트 부채 해결.
- **작업 범위**:
  - Backend UT: `createRepositoryDraft` (API-91), `requestRepositoryPublish` (API-92)
  - Frontend E2E: SCM import → create → publish 로 이어지는 `tests/e2e` Playwright 스펙 작성 (TC-REPO-LIFECYCLE-01).
- **예상 소요**: 3시간
- **검증**: `go test` 통과 및 e2e spec (`npx playwright test`) 통과.

## 5. Issue #383: X-3 Envelope Encryption (Key Management)
- **도메인**: `integration-registry` / `security`
- **목표**: 평문 secret 암호화 (credentials_ref / api_token / auth_secret) at-rest.
- **작업 범위**:
  - `internal/crypt` 기반 암호화 적용 완료(sprint-april)의 후속인 DEK/KEK rotation 키 관리 API 및 ADR 문서 반영.
  - 관련 DB 스키마 업데이트 (필요시 migration).
- **예상 소요**: 3~4시간
- **검증**: Backend UT (Envelope test), key rotation 통합 테스트.

## 후속 지침
- 각 과제는 본 계획을 기반으로 독립된 하위 스프린트/PR로 진행합니다.
- 기존 이슈들은 본 스프린트 계획 수립으로 `closed` 처리되며, 구현 시 본 문서의 항목을 참조합니다.