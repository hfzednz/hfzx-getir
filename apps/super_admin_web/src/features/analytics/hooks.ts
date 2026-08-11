"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { fetchAnalyticsSnapshot } from "./api";
import type { AnalyticsScope } from "./types";

export const analyticsKeys = {
  all: ["platform-analytics"] as const,
  snapshot: (scope: AnalyticsScope, scopeId: string | null) =>
    [...analyticsKeys.all, "snapshot", scope, scopeId] as const,
};

export function useAnalyticsSnapshot() {
  const [scope, setScope] = useState<AnalyticsScope>("worldwide");
  const [scopeId, setScopeId] = useState<string | null>(null);

  const query = useQuery({
    queryKey: analyticsKeys.snapshot(scope, scopeId),
    queryFn: () => fetchAnalyticsSnapshot(scope, scopeId),
  });

  return { ...query, scope, setScope, scopeId, setScopeId };
}
