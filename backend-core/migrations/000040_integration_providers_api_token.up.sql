-- 000040: 외부 연동 provider 의 outbound API token (PAT) 슬롯 (Gitea 연동 #3 token slot).
-- credentials_ref 는 inbound webhook 서명 시크릿 전용이라, outbound sync(REST pull)용
-- API token 을 담을 별도 컬럼이 없었다 (sync worker 가 env GITEA_TOKEN 사용). UI 로
-- 등록한 provider 가 자체 API token 을 보관해 향후 per-provider sync 에 사용한다.
-- nullable — webhook 전용 provider 는 미사용.

ALTER TABLE integration_providers
  ADD COLUMN IF NOT EXISTS api_token TEXT;
