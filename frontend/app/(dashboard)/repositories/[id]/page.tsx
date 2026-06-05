"use client";

import { useParams } from "next/navigation";
import { RepositoryDashboardView } from "@/domain/repository-integration/view/RepositoryDashboardView";

export default function RepositoryDetailPage() {
  const params = useParams();
  const idStr = params.id as string;
  const id = parseInt(idStr, 10);
  
  return <RepositoryDashboardView repoId={id} />;
}
