import type { Id } from "@/shared/types/common";

export type ProductStatus = "active" | "draft" | "archived" | "out_of_stock";

export interface ProductListItem extends Record<string, unknown> {
  id: Id;
  sku: string;
  name: string;
  brand: string;
  category: string;
  status: ProductStatus;
  price: number;
  currency: string;
  inventoryLinked: boolean;
  variantCount: number;
  hasBundle: boolean;
}

export interface ProductVariant {
  id: Id;
  sku: string;
  name: string;
  attributes: Record<string, string>;
  price: number;
  currency: string;
  barcode: string;
}

export interface ProductBundleItem {
  productId: Id;
  sku: string;
  name: string;
  qty: number;
}

export interface ProductAttribute {
  key: string;
  value: string;
}

export interface ProductMedia {
  id: Id;
  url: string;
  alt: string;
  primary: boolean;
}

export interface ProductNutrition {
  servingSize: string;
  calories: number;
  proteinG: number;
  carbsG: number;
  fatG: number;
  allergens: string[];
}

export interface ProductPricing {
  basePrice: number;
  promoPrice: number | null;
  currency: string;
  taxRatePct: number;
}

export interface ProductInventoryLink {
  warehouseId: string;
  warehouseCode: string;
  onHand: number;
  reserved: number;
  safetyStock: number;
}

export interface ProductSupplierMapping {
  supplierId: string;
  supplierName: string;
  supplierSku: string;
  leadTimeDays: number;
}

export interface ProductSeo {
  slug: string;
  metaTitle: string;
  metaDescription: string;
  keywords: string[];
}

export interface ProductDetail {
  id: Id;
  sku: string;
  name: string;
  brand: string;
  category: string;
  status: ProductStatus;
  description: string;
  variants: ProductVariant[];
  bundles: ProductBundleItem[];
  attributes: ProductAttribute[];
  media: ProductMedia[];
  nutrition: ProductNutrition | null;
  pricing: ProductPricing;
  inventoryLinks: ProductInventoryLink[];
  suppliers: ProductSupplierMapping[];
  seo: ProductSeo;
}

export interface ProductListResponse {
  items: ProductListItem[];
  brands: string[];
  categories: string[];
  total: number;
  generatedAt: string;
}

export interface ProductImportPreviewRow extends Record<string, unknown> {
  row: number;
  sku: string;
  name: string;
  brand: string;
  category: string;
  price: string;
  status: "ok" | "warning" | "error";
  message: string;
}

export interface ProductImportPreview {
  fileName: string;
  rows: ProductImportPreviewRow[];
  okCount: number;
  warningCount: number;
  errorCount: number;
}
