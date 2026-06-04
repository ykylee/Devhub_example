-- 000038: 외부 연동 provider 의 endpoint/base URL 저장 (등록 UI 고도화 #2)
-- UI 로 등록한 provider 가 외부 시스템 주소(outbound sync 대상)를 담을 수 있도록
-- integration_providers 에 base_url 컬럼 추가. nullable — webhook 전용 provider 는 미사용.

ALTER TABLE integration_providers
  ADD COLUMN IF NOT EXISTS base_url TEXT;
