-- Repository 소유권 분리 (SCM ↔ 시스템 연동).
--
-- source: 'scm' = SCM 에서 import/mirror 된 저장소 (clone_url/default_branch/private 등
--   SCM mirror 필드는 sync 가 갱신, 시스템 UI 에서 read-only) | 'system' = 시스템에서
--   생성 (향후 outbound SCM 프로비저닝 대상). NULL = legacy (코드에서 'scm' 로 취급).
-- provider_id: 이 저장소가 연동된 integration_providers(scm) FK. import 시 세팅.
-- description: 시스템 소유 메타데이터. SCM sync 가 절대 덮어쓰지 않는다.
ALTER TABLE repositories
    ADD COLUMN IF NOT EXISTS source      TEXT,
    ADD COLUMN IF NOT EXISTS provider_id UUID REFERENCES integration_providers(provider_id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS description TEXT;

CREATE INDEX IF NOT EXISTS idx_repositories_provider_id ON repositories(provider_id);
