/**
 * Application / Project lifecycle status → Badge variant 매핑.
 *
 * 백엔드 enum (`domain.ApplicationStatus` = `domain.ProjectStatus`) 의 5종 상태를
 * UI Badge 색으로 일관되게 변환한다. 기존 `status === "active" ? "success" : "warning"`
 * binary 매핑 (active 외 4종이 모두 노랑) 의 시각 구분 부족 해결.
 *
 * 매핑 의도:
 *  - active   = 정상 운영       → success (녹색)
 *  - planning = 진행 예정/계획   → primary (파랑)
 *  - on_hold  = 일시 정지        → warning (노랑)
 *  - closed   = 정상 종료        → secondary (회색)
 *  - archived = 아카이브 (보존)  → glass (흐림)
 *  - 그 외 (방어적 fallback)    → secondary
 *
 * 동일 enum 을 쓰는 Application detail / list, Project detail 모두 이 helper 를 통해
 * 매핑한다. 추가 enum 값이 백엔드에 도입되면 본 함수 + 본 함수의 unit test 만 수정.
 */
export type LifecycleBadgeVariant =
  | "success"
  | "primary"
  | "warning"
  | "secondary"
  | "glass";

export function lifecycleStatusBadgeVariant(status: string): LifecycleBadgeVariant {
  switch (status) {
    case "active":
      return "success";
    case "planning":
      return "primary";
    case "on_hold":
      return "warning";
    case "closed":
      return "secondary";
    case "archived":
      return "glass";
    default:
      return "secondary";
  }
}
