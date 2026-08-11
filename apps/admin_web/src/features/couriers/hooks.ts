"use client";

import { useQuery } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import { fetchCourierDetail, fetchCouriers } from "./api";

export const courierKeys = {
  all: ["couriers"] as const,
  list: (cityId: string | null) => [...courierKeys.all, "list", cityId] as const,
  detail: (id: string) => [...courierKeys.all, "detail", id] as const,
};

export function useCouriers() {
  const cityId = useChromeStore((s) => s.cityId);
  return useQuery({
    queryKey: courierKeys.list(cityId),
    queryFn: () => fetchCouriers(cityId),
  });
}

export function useCourierDetail(id: string) {
  return useQuery({
    queryKey: courierKeys.detail(id),
    queryFn: () => fetchCourierDetail(id),
    enabled: Boolean(id),
  });
}
