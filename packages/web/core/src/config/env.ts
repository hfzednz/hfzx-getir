export const DEFAULT_TENANT_ID = "11111111-1111-1111-1111-111111111111";

export function tenantId(): string {
  // Next.js only inlines NEXT_PUBLIC_* with static property access (not process.env[key]).
  return process.env.NEXT_PUBLIC_TENANT_ID ?? DEFAULT_TENANT_ID;
}

export function identityUrl(): string {
  return (process.env.NEXT_PUBLIC_IDENTITY_URL ?? "http://localhost:8081").replace(/\/$/, "");
}

export function bffUrl(kind: "customer" | "admin" | "courier" | "warehouse"): string {
  // Static env reads are required for Next.js client bundle inlining.
  const fromEnv =
    kind === "customer"
      ? process.env.NEXT_PUBLIC_BFF_CUSTOMER_URL
      : kind === "admin"
        ? process.env.NEXT_PUBLIC_BFF_ADMIN_URL
        : kind === "courier"
          ? process.env.NEXT_PUBLIC_BFF_COURIER_URL
          : process.env.NEXT_PUBLIC_BFF_WAREHOUSE_URL;
  const defaults: Record<typeof kind, string> = {
    customer: "http://localhost:8111",
    admin: "http://localhost:8114",
    courier: "http://localhost:8112",
    warehouse: "http://localhost:8113",
  };
  return (fromEnv ?? defaults[kind]).replace(/\/$/, "");
}

export function serviceUrl(kind: "finance" | "settlement" | "supplier" | "realtime" | "platform"): string {
  const fromEnv =
    kind === "finance"
      ? process.env.NEXT_PUBLIC_FINANCE_URL
      : kind === "settlement"
        ? process.env.NEXT_PUBLIC_SETTLEMENT_URL
        : kind === "supplier"
          ? process.env.NEXT_PUBLIC_SUPPLIER_URL
          : kind === "realtime"
            ? process.env.NEXT_PUBLIC_REALTIME_URL
            : process.env.NEXT_PUBLIC_PLATFORM_OPS_URL;
  const defaults = {
    finance: "http://localhost:8091",
    settlement: "http://localhost:8092",
    supplier: "http://localhost:8117",
    realtime: "http://localhost:8115",
    platform: "http://localhost:8110",
  } as const;
  return (fromEnv ?? defaults[kind]).replace(/\/$/, "");
}
