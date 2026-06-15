# 작업 백로그

슬롯 종료. 모든 Planned 항목 완료.

## [Planned]
(없음)

## [In Progress]
(없음)

## [Done — PR #67]
- [x] PS dev-up.ps1 / dev-down.ps1 신규 (ASCII 영어, 메모리 정책)
- [x] dev-up.sh / dev-down.sh 하드닝 — set -euo pipefail, wait_for_port, migrate-up 자동 호출
- [x] M2 audit consumer UI — `/admin/settings/audit` page + service 와이어 contract 정정
- [x] signout.spec ERR_ABORTED flake 안정화 (Next 16 turbopack)
- [x] Kratos credentials_identifier 기반 O(1) FindIdentityByUserID fast path + httptest 3 신규
- [x] password-change.spec, signout.spec 회귀 재검증 (resolver slow path 회복 확인)

## [Done — 이전 세션 (gemini/frontend_260510 검토 → 정리)]
- [x] C1-C3 Kratos admin client revert + httptest (PR #63)
- [x] H3-H5 / M1 / M3-M5 follow-ups + seedLocalAdmin 단위 테스트 (PR #64)
- [x] N1 actor role error surface + GetUser sentinel (PR #65)
- [x] N2 Kratos 0-based pagination (PR #66)
