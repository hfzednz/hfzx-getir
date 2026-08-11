"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { useChromeStore } from "@/stores/chrome-store";
import {
  fetchProductDetail,
  fetchProducts,
  previewProductImport,
} from "./api";

export const productKeys = {
  all: ["products"] as const,
  list: (cityId: string | null) =>
    [...productKeys.all, "list", cityId] as const,
  detail: (id: string) => [...productKeys.all, "detail", id] as const,
};

export function useProducts() {
  const cityId = useChromeStore((s) => s.cityId);
  return useQuery({
    queryKey: productKeys.list(cityId),
    queryFn: () => fetchProducts(cityId),
  });
}

export function useProductDetail(id: string) {
  return useQuery({
    queryKey: productKeys.detail(id),
    queryFn: () => fetchProductDetail(id),
    enabled: Boolean(id),
  });
}

export function useProductImportPreview() {
  return useMutation({
    mutationFn: (fileName: string) => previewProductImport(fileName),
  });
}
