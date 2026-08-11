"use client";

import { useQuery } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import { fetchWarehouseDetail, fetchWarehouses } from "./api";

export const warehouseKeys = {
  all: ["warehouses"] as const,
  list: (cityId: string | null) =>
    [...warehouseKeys.all, "list", cityId] as const,
  detail: (id: string) => [...warehouseKeys.all, "detail", id] as const,
};

export function useWarehouses() {
  const cityId = useChromeStore((s) => s.cityId);
  return useQuery({
    queryKey: warehouseKeys.list(cityId),
    queryFn: () => fetchWarehouses(cityId),
  });
}

export function useWarehouseDetail(id: string) {
  return useQuery({
    queryKey: warehouseKeys.detail(id),
    queryFn: () => fetchWarehouseDetail(id),
    enabled: Boolean(id),
  });
}
