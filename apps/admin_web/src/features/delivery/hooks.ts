"use client";

import { useQuery } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import {
  fetchDeliverySnapshot,
  fetchDeliveryZoneDetail,
  fetchDeliveryZones,
} from "./api";

export const deliveryKeys = {
  all: ["delivery"] as const,
  snapshot: (cityId: string | null) =>
    [...deliveryKeys.all, "snapshot", cityId] as const,
  zones: (cityId: string | null) =>
    [...deliveryKeys.all, "zones", cityId] as const,
  zoneDetail: (id: string) => [...deliveryKeys.all, "zone", id] as const,
};

export function useDeliverySnapshot() {
  const cityId = useChromeStore((s) => s.cityId);
  return useQuery({
    queryKey: deliveryKeys.snapshot(cityId),
    queryFn: () => fetchDeliverySnapshot(cityId),
    refetchInterval: 20_000,
  });
}

export function useDeliveryZones() {
  const cityId = useChromeStore((s) => s.cityId);
  return useQuery({
    queryKey: deliveryKeys.zones(cityId),
    queryFn: () => fetchDeliveryZones(cityId),
  });
}

export function useDeliveryZoneDetail(id: string) {
  return useQuery({
    queryKey: deliveryKeys.zoneDetail(id),
    queryFn: () => fetchDeliveryZoneDetail(id),
    enabled: Boolean(id),
  });
}
