import type { SeriesPoint } from "@/shared/lib/charts";

export interface GatewayRoute {
  id: string;
  path: string;
  methods: string;
  upstream: string;
  version: string;
  auth: "oidc" | "api_key" | "oauth" | "public" | "mtls";
  rateLimitRpm: number;
  status: "active" | "canary" | "disabled";
}

export interface ApiKeyRecord {
  id: string;
  name: string;
  prefix: string;
  tenantId: string | null;
  scopes: string;
  rateLimitRpm: number;
  lastUsedAt: string | null;
  status: "active" | "revoked" | "expired";
  expiresAt: string | null;
}

export interface OAuthClient {
  id: string;
  clientId: string;
  name: string;
  grantTypes: string;
  redirectUris: string;
  scopes: string;
  status: "active" | "disabled";
  createdAt: string;
}

export interface RateLimitPolicy {
  id: string;
  name: string;
  scope: "global" | "tenant" | "route" | "key";
  target: string;
  limitRpm: number;
  burst: number;
  status: "enforced" | "shadow" | "disabled";
}

export interface ApiVersion {
  id: string;
  version: string;
  status: "stable" | "deprecated" | "sunset" | "beta";
  trafficPct: number;
  sunsetAt: string | null;
  routes: number;
}

export interface ServiceDiscoveryEntry {
  id: string;
  service: string;
  instances: number;
  healthy: number;
  registry: "consul" | "k8s" | "eureka" | "dns";
  endpoint: string;
  status: "healthy" | "degraded" | "down";
}

export interface GatewaySnapshot {
  generatedAt: string;
  kpis: {
    rps: number;
    p99Ms: number;
    errorRatePct: number;
    activeKeys: number;
    oauthClients: number;
    discoveredServices: number;
  };
  trafficSeries: SeriesPoint[];
  errorSeries: SeriesPoint[];
  usageByRoute: SeriesPoint[];
  config: {
    edgeRegion: string;
    tlsMinVersion: string;
    jwtIssuer: string;
    requestTimeoutMs: number;
    bodyLimitMb: number;
    corsMode: string;
    wafEnabled: boolean;
  };
  routes: GatewayRoute[];
  apiKeys: ApiKeyRecord[];
  oauthClients: OAuthClient[];
  rateLimits: RateLimitPolicy[];
  versions: ApiVersion[];
  discovery: ServiceDiscoveryEntry[];
}
