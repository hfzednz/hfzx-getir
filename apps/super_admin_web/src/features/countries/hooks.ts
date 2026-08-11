"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchCountries, fetchCountry } from "./api";

export const countryKeys = {
  all: ["platform-countries"] as const,
  list: () => [...countryKeys.all, "list"] as const,
  detail: (id: string) => [...countryKeys.all, "detail", id] as const,
};

export function useCountries() {
  return useQuery({
    queryKey: countryKeys.list(),
    queryFn: fetchCountries,
  });
}

export function useCountry(id: string) {
  return useQuery({
    queryKey: countryKeys.detail(id),
    queryFn: () => fetchCountry(id),
    enabled: Boolean(id),
  });
}
