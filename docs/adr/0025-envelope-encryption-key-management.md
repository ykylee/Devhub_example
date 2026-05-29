# ADR-0025: 데이터베이스 내 자격증명 비밀 데이터 봉투 암호화 및 키 관리 정책

## 1. 상태
- **상태**: Accepted
- **작성일**: 2026-05-29
- **수정일**: 2026-05-29
- **결정 근거 sprint**: `gemini/work_260529-a-envelope-encryption`
- **관련 문서**: [`backend-core/internal/domain/integration-registry/repository/integration_registry.go`](../../backend-core/internal/domain/integration-registry/repository/integration_registry.go), [v1.0 릴리즈 로드맵 §3.5 신규 도출 백로그](../planning/release_v1_roadmap.md), [Session Handoff EOD](../../ai-workflow/memory/session_handoff.md).

---

## 2. 컨텍스트

### 2.1 보안 취약점 식별
현재 DevHub의 외부 연동(SCM Gitea, HomeLab 등) 시 활용되는 자격증명 데이터(`api_token`, `auth_secret`, webhook 서명 검증용 `credentials_ref` 내의 secret 등)가 데이터베이스(`integration_providers` 테이블)의 텍스트 컬럼에 평문(Plaintext) 상태로 적재되고 있습니다.
만약 데이터베이스 백업 파일이 유출되거나, SQL 인젝션 공격 또는 DB 물리 계정이 탈취될 경우, 연동된 모든 외부 시스템의 쓰기/관리자 권한을 포함하는 Access Token이 Plaintext 상태로 외부 노출되는 매우 치명적인 보안 위협이 존재합니다.

### 2.2 봉투 암호화(Envelope Encryption) 아키텍처의 강점
비밀 데이터를 직접 마스터 키로 암호화(Direct Encryption)할 경우, 마스터 키가 유출되었을 때 모든 데이터가 동시에 복호화되며 대량 암호문 분석 공격에 취약해집니다.
이를 해결하기 위해 **봉투 암호화(Envelope Encryption)** 아키텍처를 도입합니다:
* **KEK (Key Encryption Key)**: 서버가 기동 환경변수(`DEVHUB_ENCRYPTION_KEY`)로부터 주입받는 단일 마스터 키입니다.
* **DEK (Data Encryption Key)**: 매번 암호화 요청이 발생할 때마다 `crypto/rand` 난수 생성기로 독립 생성되는 고유 암호화 키입니다.
* 데이터 자체는 **DEK**를 키로 AES-GCM-256 대칭 키 알고리즘을 사용해 고유 Nonce와 함께 암호화하며, 비밀 데이터 복호화의 열쇠가 되는 이 DEK 자체를 마스터 키(**KEK**)로 암호화하여 데이터와 한 봉투로 직렬화해 보관합니다.

---

## 3. 결정

### 3.1 암호화 및 복호화 메커니즘
* **알고리즘**: 대칭 키 암호화의 표준 규격이자 내부 무결성 검증을 포함하는 **AES-GCM-256**을 전격 채택합니다.
* **직렬화 규격 (Envelope Format)**: 
  데이터베이스 TEXT 컬럼에 기록되는 최종 암호문은 레거시 평문과 명확히 식별되고 자체 봉투 데이터를 포함하도록 아래의 직렬화 포맷을 강제합니다:
  ```
  $env$v1$<wrapped_dek_b64>$<nonce_b64>$<ciphertext_b64>
  ```
  * `$env$v1$`: 봉투 암호문임을 나타내는 식별용 접두사.
  * `<wrapped_dek_b64>`: 마스터 KEK로 AES-GCM 암호화 및 Nonce가 패키징된 데이터 암호화 키의 Base64 스트링.
  * `<nonce_b64>`: 암호화에 사용된 고유 12바이트 Nonce의 Base64 스트링.
  * `<ciphertext_b64>`: 데이터 본문의 AES-GCM 암호문 Base64 스트링.

### 3.2 마스터 키 주입 및 환경별 이중 분기 정책
* **KEK 주입**: 환경변수 `DEVHUB_ENCRYPTION_KEY`를 통해 32바이트 대칭 키 데이터를 Hex 또는 Base64 디코딩 가능한 문자열로 주입받습니다.
* **이중 분기 (Environment Bounded Fallback)**:
  * **Production / Staging**: `DEVHUB_ENCRYPTION_KEY`가 생략되었거나 32바이트 규격 미달 시, 보안 정합성을 위해 **서버 부팅 단계에서 즉시 Panic 및 프로세스 차단**을 실행합니다.
  * **Local / Test / CI**: 로컬 개발 편의성 및 유닛 테스트 정합을 위해, 환경변수가 공란("")일 시 자동으로 **Plaintext 바이패스(Bypass) 모드**가 가동되도록 분기 처리합니다.

### 3.3 레거시 호환 및 점진적 자동 격상 (Auto-Upgrade)
기존 DB에 존재하던 평문 비밀들과의 매끄러운 징검다리를 확립하기 위해 다음 자동 대응 기어를 탑재합니다:
* **Scan Fallback**: 데이터베이스 조회(`Scan`) 시, 가져온 컬럼의 값이 `$env$v1$` 접두사로 시작하지 않는 Plaintext인 경우, 복호화 절차를 투명하게 건너뛰고 평문 문자열 그대로 복원합니다.
* **Save Auto-Upgrade**: Plaintext 상태로 읽어들인 레코드라도, 이후 변경(`Create` 또는 `Update`)되어 저장되는 시점에는 마스터 키가 활성화되어 있을 경우 강제적으로 Envelope 암호문 포맷으로 격상하여 DB에 씁니다.

### 3.4 영속성 레이어 최소 침습 가드
비즈니스 로직(SCM sync 등)에 침투적인 코드 churn을 배제하기 위해, DB 입출력을 전담하는 `IntegrationRepository` 내의 `Scan` 및 `Create`/`Update` 시점에 투명 암복호화 필터를 끼워 넣습니다. 

---

## 4. 결과

### 4.1 긍정적 효과
* **상용 등급 보안성**: DB 백업이나 물리 데이터가 유출되더라도, 환경변수로 격리 보관된 KEK가 부재하다면 비밀 정보를 절대로 복구할 수 없습니다.
* **최소 침습 구현**: SCM 연동 등 기존 비즈니스 로직 소스 코드를 단 한 줄도 손대지 않아, 무결성과 안정성이 100% 보존됩니다.
* **매끄러운 데이터 마이그레이션**: 별도의 DB 무중단 마이그레이션 쿼리를 돌릴 필요 없이, 일상적인 CRUD 동작 중 평문 데이터가 자동으로 봉투 암호문으로 자연 격상(Auto-Upgrade)됩니다.

### 4.2 부정적 효과
* **암복호화 오버헤드**: 미세한 AES-GCM 연산 비용이 추가되나, Provider 저장/조회 빈도(분당 수회 이내)를 감안할 때 실 서비스 Uptime 성능 영향도는 0%에 수렴합니다.
* **마스터 키 관리 의존성**: Production 기동 시 `DEVHUB_ENCRYPTION_KEY` 주입 관리를 위한 사내 운영 SOP 갱신 부담이 발생합니다.

---

## 5. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-29 | 1차 발행 (Accepted). KEK 기반 AES-GCM-256 봉투 암호화 및 자동 격상 정책 수립. | `gemini/work_260529-a-envelope-encryption` |
