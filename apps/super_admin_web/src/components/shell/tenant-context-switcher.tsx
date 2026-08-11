"use client";

import { useMemo } from "react";
import type { Id } from "@/shared/types/common";
import { Select } from "@nexora/ui";
import { useChromeStore } from "@/stores/chrome-store";

export interface TenantOption {
  id: Id;
  name: string;
  code: string;
}

const DEMO_TENANTS: TenantOption[] = [
  { id: "all", name: "All tenants", code: "ALL" },
  { id: "tn_acme", name: "Acme Commerce", code: "ACME" },
  { id: "tn_nova", name: "Nova Quick", code: "NOVA" },
  { id: "tn_helix", name: "Helix Retail", code: "HLX" },
];

/** Optional tenant filter context — not city-ops. */
export function TenantContextSwitcher({
  tenants = DEMO_TENANTS,
}: {
  tenants?: TenantOption[];
}) {
  const tenantContextId = useChromeStore((s) => s.tenantContextId);
  const setTenantContextId = useChromeStore((s) => s.setTenantContextId);

  const value = useMemo(() => {
    if (tenantContextId && tenants.some((t) => t.id === tenantContextId)) {
      return tenantContextId;
    }
    return tenants[0]?.id ?? "all";
  }, [tenantContextId, tenants]);

  return (
    <label className="flex items-center gap-[var(--nx-space-2)]">
      <span className="sr-only">Tenant context</span>
      <Select
        aria-label="Tenant context switcher"
        value={value}
        onChange={(e) => {
          const next = e.target.value;
          setTenantContextId(next === "all" || !next ? null : next);
        }}
        className="min-w-[160px] h-[var(--nx-control-height-sm)] text-[12px]"
      >
        {tenants.map((tenant) => (
          <option key={tenant.id} value={tenant.id}>
            {tenant.name}
          </option>
        ))}
      </Select>
    </label>
  );
}
