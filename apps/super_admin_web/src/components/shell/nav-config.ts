export interface NavItem {
  href: string;
  label: string;
  /** Optional permission soft-gate */
  permission?: string;
}

/** Full nav tree matching ARCHITECTURE.md — platform only, no city-ops. */
export const NAV_ITEMS: NavItem[] = [
  { href: "/dashboard", label: "Dashboard", permission: "dashboard:read" },
  { href: "/tenants", label: "Tenants", permission: "tenants:read" },
  { href: "/companies", label: "Companies", permission: "companies:read" },
  { href: "/countries", label: "Countries", permission: "countries:read" },
  { href: "/org", label: "Organization", permission: "org:read" },
  { href: "/roles", label: "Roles", permission: "roles:read" },
  { href: "/config", label: "Config", permission: "config:read" },
  { href: "/flags", label: "Feature flags", permission: "flags:read" },
  { href: "/licenses", label: "Licenses", permission: "licenses:read" },
  { href: "/billing", label: "Billing", permission: "billing:read" },
  { href: "/security", label: "Security", permission: "security:read" },
  { href: "/compliance", label: "Compliance", permission: "compliance:read" },
  { href: "/infra", label: "Infrastructure", permission: "infra:read" },
  { href: "/databases", label: "Databases", permission: "databases:read" },
  { href: "/gateway", label: "API gateway", permission: "gateway:read" },
  { href: "/messaging", label: "Messaging", permission: "messaging:read" },
  { href: "/observability", label: "Observability", permission: "observability:read" },
  { href: "/ai-platform", label: "AI platform", permission: "ai_platform:read" },
  { href: "/analytics", label: "Analytics", permission: "analytics:read" },
  { href: "/disaster-recovery", label: "Disaster recovery", permission: "dr:read" },
  { href: "/deployments", label: "Deployments", permission: "deployments:read" },
  { href: "/monitoring", label: "Monitoring", permission: "monitoring:read" },
  { href: "/notifications", label: "Notifications", permission: "notifications:read" },
  { href: "/audit", label: "Audit", permission: "audit:read" },
  { href: "/reports", label: "Reports", permission: "reports:read" },
];

/** Flat searchable routes for command palette (includes nested detail paths). */
export const COMMAND_ROUTES: { href: string; label: string; keywords?: string }[] =
  [
    ...NAV_ITEMS.map((i) => ({ href: i.href, label: i.label })),
    { href: "/tenants", label: "Tenant detail", keywords: "tenants id isolate" },
    { href: "/companies", label: "Company detail", keywords: "companies legal entity" },
    { href: "/countries", label: "Country detail", keywords: "countries region" },
    { href: "/flags", label: "Kill switches", keywords: "flags kill dual control" },
    {
      href: "/disaster-recovery",
      label: "DR failover",
      keywords: "disaster recovery failover",
    },
    {
      href: "/security",
      label: "Security command center",
      keywords: "security fraud alerts",
    },
    {
      href: "/compliance",
      label: "GDPR KVKK CCPA",
      keywords: "compliance privacy export",
    },
  ];
