"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchDatabasesSnapshot } from "./api";

export const databasesKeys = {
  all: ["platform-databases"] as const,
  snapshot: () => [...databasesKeys.all, "snapshot"] as const,
};

export function useDatabasesSnapshot() {
  return useQuery({
    queryKey: databasesKeys.snapshot(),
    queryFn: fetchDatabasesSnapshot,
  });
}
