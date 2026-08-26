import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient } from "@/shared/api/client";
import type {
  ProductDetail,
  ProductImportPreview,
  ProductListItem,
  ProductListResponse,
} from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const MOCK_LIST: ProductListItem[] = [
  {
    id: "prd_1",
    sku: "SKU-MLK-012",
    name: "Full fat milk 1L",
    brand: "Sütaş",
    category: "Dairy",
    status: "active",
    price: 42_90,
    currency: "TRY",
    inventoryLinked: true,
    variantCount: 1,
    hasBundle: false,
  },
  {
    id: "prd_2",
    sku: "SKU-ICE-044",
    name: "Magnum Classic",
    brand: "Magnum",
    category: "Frozen",
    status: "active",
    price: 55_00,
    currency: "TRY",
    inventoryLinked: true,
    variantCount: 3,
    hasBundle: false,
  },
  {
    id: "prd_3",
    sku: "SKU-BND-001",
    name: "Breakfast bundle",
    brand: "NEXORA",
    category: "Bundles",
    status: "active",
    price: 189_90,
    currency: "TRY",
    inventoryLinked: true,
    variantCount: 0,
    hasBundle: true,
  },
  {
    id: "prd_4",
    sku: "SKU-WTR-003",
    name: "Still water 0.5L",
    brand: "Erikli",
    category: "Beverages",
    status: "draft",
    price: 12_50,
    currency: "TRY",
    inventoryLinked: false,
    variantCount: 2,
    hasBundle: false,
  },
  {
    id: "prd_5",
    sku: "SKU-SNK-220",
    name: "Protein bar cocoa",
    brand: "Quest",
    category: "Snacks",
    status: "out_of_stock",
    price: 68_00,
    currency: "TRY",
    inventoryLinked: true,
    variantCount: 1,
    hasBundle: false,
  },
];

function mockDetail(id: string): ProductDetail {
  const base =
    MOCK_LIST.find((p) => p.id === id) ??
    ({ ...MOCK_LIST[0], id, sku: id.toUpperCase() } satisfies ProductListItem);

  return {
    id: base.id,
    sku: base.sku,
    name: base.name,
    brand: base.brand,
    category: base.category,
    status: base.status,
    description: `${base.name} — catalog entry for ops and storefront.`,
    variants:
      base.variantCount > 0
        ? [
            {
              id: "var_1",
              sku: `${base.sku}-A`,
              name: `${base.name} · Default`,
              attributes: { size: "default" },
              price: base.price,
              currency: base.currency,
              barcode: "8690000000012",
            },
            ...(base.variantCount > 1
              ? [
                  {
                    id: "var_2",
                    sku: `${base.sku}-B`,
                    name: `${base.name} · Large`,
                    attributes: { size: "large" },
                    price: base.price + 800,
                    currency: base.currency,
                    barcode: "8690000000013",
                  },
                ]
              : []),
          ]
        : [],
    bundles: base.hasBundle
      ? [
          {
            productId: "prd_1",
            sku: "SKU-MLK-012",
            name: "Full fat milk 1L",
            qty: 1,
          },
          {
            productId: "prd_4",
            sku: "SKU-WTR-003",
            name: "Still water 0.5L",
            qty: 2,
          },
        ]
      : [],
    attributes: [
      { key: "unit", value: "1 pcs" },
      { key: "storage", value: base.category === "Frozen" ? "frozen" : "ambient" },
      { key: "origin", value: "TR" },
    ],
    media: [
      {
        id: "media_1",
        url: "/placeholder-product.jpg",
        alt: base.name,
        primary: true,
      },
    ],
    nutrition:
      base.category === "Bundles"
        ? null
        : {
            servingSize: "100g",
            calories: 210,
            proteinG: 4.2,
            carbsG: 28,
            fatG: 8.1,
            allergens: ["milk"],
          },
    pricing: {
      basePrice: base.price,
      promoPrice: base.status === "active" ? base.price - 500 : null,
      currency: base.currency,
      taxRatePct: 10,
    },
    inventoryLinks: base.inventoryLinked
      ? [
          {
            warehouseId: "wh_07",
            warehouseCode: "WH-07",
            onHand: 120,
            reserved: 14,
            safetyStock: 40,
          },
          {
            warehouseId: "wh_14",
            warehouseCode: "WH-14",
            onHand: 86,
            reserved: 6,
            safetyStock: 30,
          },
        ]
      : [],
    suppliers: [
      {
        supplierId: "sup_1",
        supplierName: "Metro Cash",
        supplierSku: `M-${base.sku}`,
        leadTimeDays: 2,
      },
    ],
    seo: {
      slug: base.name.toLowerCase().replace(/\s+/g, "-"),
      metaTitle: `${base.name} | NEXORA`,
      metaDescription: `Buy ${base.name} with quick delivery.`,
      keywords: [base.brand, base.category, "quick commerce"],
    },
  };
}

export async function fetchProducts(
  cityId: string | null,
): Promise<ProductListResponse> {
  try {
    const q = cityId ? `?cityId=${encodeURIComponent(cityId)}` : "";
    return await apiClient<ProductListResponse>(`/admin/catalog/products${q}`);
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    await delay();
    return {
      items: MOCK_LIST,
      brands: [...new Set(MOCK_LIST.map((p) => p.brand))],
      categories: [...new Set(MOCK_LIST.map((p) => p.category))],
      total: MOCK_LIST.length,
      generatedAt: new Date().toISOString(),
    };
  }
}

export async function fetchProductDetail(id: string): Promise<ProductDetail> {
  try {
    return await apiClient<ProductDetail>(`/admin/catalog/products/${id}`);
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    await delay();
    return mockDetail(id);
  }
}

export async function previewProductImport(
  fileName: string,
): Promise<ProductImportPreview> {
  try {
    return await apiClient<ProductImportPreview>(
      "/admin/catalog/products/import/preview",
      { method: "POST", body: { fileName }, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    await delay(400);
    return {
      fileName,
      okCount: 3,
      warningCount: 1,
      errorCount: 1,
      rows: [
        {
          row: 2,
          sku: "SKU-NEW-101",
          name: "Organic eggs 10pcs",
          brand: "Organik",
          category: "Dairy",
          price: "89.90",
          status: "ok",
          message: "Ready to import",
        },
        {
          row: 3,
          sku: "SKU-NEW-102",
          name: "Almond milk 1L",
          brand: "Alpro",
          category: "Dairy",
          price: "72.50",
          status: "ok",
          message: "Ready to import",
        },
        {
          row: 4,
          sku: "SKU-MLK-012",
          name: "Full fat milk 1L",
          brand: "Sütaş",
          category: "Dairy",
          price: "42.90",
          status: "warning",
          message: "SKU exists — will update",
        },
        {
          row: 5,
          sku: "",
          name: "Broken row",
          brand: "",
          category: "",
          price: "",
          status: "error",
          message: "Missing SKU",
        },
        {
          row: 6,
          sku: "SKU-NEW-103",
          name: "Sparkling water 1L",
          brand: "Beypazarı",
          category: "Beverages",
          price: "18.00",
          status: "ok",
          message: "Ready to import",
        },
      ],
    };
  }
}
