import type { Id } from "@/shared/types/common";

export type ThreatSeverity = "critical" | "high" | "medium" | "low" | "info";
export type SessionRisk = "normal" | "suspicious" | "blocked";
export type AuthProviderKind = "oidc" | "saml" | "oauth" | "webauthn" | "totp";

export interface LoginEvent {
  id: Id;
  userEmail: string;
  ip: string;
  country: string;
  city: string;
  success: boolean;
  mfaUsed: boolean;
  userAgent: string;
  createdAt: string;
}

export interface ThreatAlert {
  id: Id;
  severity: ThreatSeverity;
  title: string;
  detail: string;
  source: string;
  status: "open" | "acknowledged" | "resolved";
  createdAt: string;
}

export interface SuspiciousSession {
  id: Id;
  userEmail: string;
  deviceId: Id;
  deviceLabel: string;
  ip: string;
  country: string;
  risk: SessionRisk;
  reason: string;
  lastSeenAt: string;
}

export interface RegisteredDevice {
  id: Id;
  userEmail: string;
  label: string;
  platform: string;
  trusted: boolean;
  lastSeenAt: string;
  createdAt: string;
}

export interface GeoIpRule {
  id: Id;
  type: "allow" | "deny";
  target: "country" | "ip_cidr";
  value: string;
  label: string;
  enabled: boolean;
}

export interface AuthProvider {
  id: Id;
  name: string;
  kind: AuthProviderKind;
  enabled: boolean;
  enforced: boolean;
  tenantScope: string | null;
}

export interface SecurityPolicy {
  id: Id;
  key: string;
  name: string;
  description: string;
  value: string;
  enforced: boolean;
}

export interface SecuritySnapshot {
  loginEvents: LoginEvent[];
  threats: ThreatAlert[];
  sessions: SuspiciousSession[];
  devices: RegisteredDevice[];
  geoIpRules: GeoIpRule[];
  providers: AuthProvider[];
  policies: SecurityPolicy[];
  failedLogins24h: number;
  openThreats: number;
  generatedAt: string;
}
