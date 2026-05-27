-- repositories.scm_provider(provider_key TEXT, #368) ↔ provider_id(FK UUID, #363) 중복 정리.
-- provider_id(FK)를 단일 출처로 통일하고 scm_provider 를 제거한다.
--
-- backfill: scm_provider(provider_key)로 등록된 integration_providers 가 있으면 그 FK 로
-- provider_id 를 채운다 (publish 가 GetIntegrationProviderByKey 로 해석하던 것과 동일 매핑).
UPDATE repositories r
SET provider_id = p.provider_id
FROM integration_providers p
WHERE r.provider_id IS NULL
  AND r.scm_provider IS NOT NULL
  AND r.scm_provider <> ''
  AND p.provider_key = r.scm_provider;

ALTER TABLE repositories DROP COLUMN IF EXISTS scm_provider;
