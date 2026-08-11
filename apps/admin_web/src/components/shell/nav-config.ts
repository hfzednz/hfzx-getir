export interface NavItem {
  href: string;
  label: string;
  /** Optional permission soft-gate */
  permission?: string;
}

/** Full nav tree matching ARCHITECTURE.md */
export const NAV_ITEMS: NavItem[] = [
  { href: "/dashboard", label: "Dashboard", permission: "dashboard:read" },
  { href: "/live", label: "Live Ops", permission: "live:read" },
  { href: "/orders", label: "Orders", permission: "orders:read" },
  { href: "/customers", label: "Customers", permission: "customers:read" },
  { href: "/couriers", label: "Couriers", permission: "couriers:read" },
  { href: "/warehouses", label: "Warehouses", permission: "warehouses:read" },
  { href: "/products", label: "Products", permission: "catalog:read" },
  { href: "/inventory", label: "Inventory", permission: "inventory:read" },
  { href: "/delivery", label: "Delivery", permission: "delivery:read" },
  { href: "/campaigns", label: "Campaigns", permission: "campaigns:read" },
  { href: "/pricing", label: "Pricing", permission: "pricing:read" },
  { href: "/loyalty", label: "Loyalty", permission: "loyalty:read" },
  { href: "/crm", label: "CRM", permission: "crm:read" },
  { href: "/support", label: "Support", permission: "support:read" },
  { href: "/finance", label: "Finance", permission: "finance:read" },
  { href: "/analytics", label: "Analytics", permission: "analytics:read" },
  { href: "/ai", label: "AI Command", permission: "ai:read" },
  { href: "/system", label: "System", permission: "system:read" },
  { href: "/rbac", label: "RBAC", permission: "rbac:read" },
  { href: "/audit", label: "Audit Logs", permission: "audit:read" },
  { href: "/notifications", label: "Notifications", permission: "notifications:read" },
  { href: "/monitoring", label: "Monitoring", permission: "monitoring:read" },
  { href: "/reports", label: "Reports", permission: "reports:read" },
];

/** Flat searchable routes for command palette (includes nested paths). */
export const COMMAND_ROUTES: { href: string; label: string; keywords?: string }[] =
  [
    ...NAV_ITEMS.map((i) => ({ href: i.href, label: i.label })),
    { href: "/orders", label: "Order detail", keywords: "orders id" },
    { href: "/customers", label: "Customer detail", keywords: "customers id" },
    { href: "/couriers", label: "Courier detail", keywords: "couriers id" },
    { href: "/warehouses", label: "Warehouse detail", keywords: "warehouses id" },
    { href: "/products", label: "Product detail", keywords: "products id" },
    { href: "/products/import", label: "Product import", keywords: "catalog import" },
    { href: "/delivery/zones", label: "Delivery zones", keywords: "zones geo" },
    { href: "/campaigns", label: "Campaign detail", keywords: "campaigns id" },
    { href: "/support", label: "Support tickets", keywords: "tickets" },
    { href: "/system/flags", label: "Feature flags", keywords: "flags kill switch" },
    { href: "/system/templates", label: "Templates", keywords: "templates locales" },
  ];
