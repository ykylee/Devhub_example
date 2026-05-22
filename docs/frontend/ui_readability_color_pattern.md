# UI Readability & Color Pattern (Light/Dark)

## 목적
라이트/다크 테마에서 텍스트 가독성을 일관되게 확보하기 위한 기본 규칙을 정의한다.
신규 UI 추가/기존 UI 변경 시 본 문서를 우선 참조한다.

## 핵심 원칙
- 본문 텍스트는 배경 대비 4.5:1 이상을 기본 목표로 한다.
- 12px 이하(또는 얇은 폰트) 텍스트는 대비 7:1을 권장한다.
- 상태 색상(`success`, `warning`, `accent`)은 "텍스트"와 "배경/배지" 용도를 분리한다.
- 라이트 모드에서 `text-primary-foreground/*`를 일반 텍스트 색으로 직접 사용하지 않는다.

## 토큰 정책
- 라이트 모드 토큰은 텍스트 용도로도 충분한 명도/채도를 가져야 한다.
- 2026-05-22 기준 라이트 토큰:
  - `--foreground: #0b1220`
  - `--accent: #0369a1`
  - `--success: #15803d`
  - `--warning: #b45309`
- 다크 모드 토큰은 기존 의도를 유지하되, 라이트와 혼용되는 클래스(`dark:` 없는 클래스)에서 저대비가 발생하지 않게 분기한다.

## 클래스 사용 규칙
- 기본 텍스트:
  - `text-foreground`
  - 보조: `text-foreground/70` 이상 권장 (`/50` 이하 지양)
- 아이콘/플레이스홀더:
  - 라이트: `text-foreground/35~55`
  - 다크: `dark:text-primary-foreground/20~40`
- 상태 텍스트:
  - `text-success`, `text-warning`, `text-accent` 사용 가능
  - 배지 배경(`bg-*/10`) 위에서도 텍스트 대비가 충분한지 확인
- 금지/주의:
  - 라이트 화면에서 `text-primary-foreground/10~40` 단독 사용 금지
  - `text-white`는 `bg-primary`, `bg-accent` 등 진한 배경 버튼에만 사용

## 구현 패턴
- 라이트/다크 분기가 필요한 경우:
  - `text-foreground/45 dark:text-primary-foreground/30`
  - `text-foreground/35 dark:text-primary-foreground/20`
- 입력창 선행 아이콘 권장 패턴:
  - `text-foreground/35 dark:text-primary-foreground/20 group-focus-within:text-<semantic>`

## PR 체크리스트
- 라이트 모드에서 주요 페이지(대시보드/관리자/폼/모달) 시각 확인
- 12px 이하 텍스트, 배지, 뮤트 텍스트, 상태 컬러 텍스트 우선 점검
- `text-white`, `text-primary-foreground/*`, `text-foreground/50 이하` 사용 위치 재검토
- 가능하면 자동 점검(contrast audit) 결과를 첨부

