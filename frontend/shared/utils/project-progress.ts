import type { ProjectTaskItem } from "@/domain/platform-lifecycle/schema/project.types";

// Task 완료율 기반 project progress 계산. 단순 count 기반 (모든 task 동일 가중).
// 향후 story-point 기반 가중치 추가 시 이 함수를 확장.
//
// @param tasks 프로젝트의 전체 task 목록 (모든 status 포함)
// @returns 0-100 (반올림), 또는 null (task 없음 = neutral state)
export function computeProjectProgress(tasks: ProjectTaskItem[]): number | null {
  const total = tasks.length;
  if (total === 0) return null;
  const doneCount = tasks.filter((t) => t.status === "done").length;
  return Math.round((doneCount / total) * 100);
}
