import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type {
  AuthProvider,
  GeoIpRule,
  SecurityPolicy,
  SecuritySnapshot,
  SuspiciousSession,
  ThreatAlert,
} from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const mockThreats: ThreatAlert[] = [
  {
    id: "th1",
    severity: "critical",
    title: "Credential stuffing wave",
    detail: "14k failed logins / 5m · EU-WEST identity edge",
    source: "identity-edge",
    status: "open",
    createdAt: new Date(Date.now() - 12 * 60_000).toISOString(),
  },
  {
    id: "th2",
    severity: "high",
    title: "Impossible travel",
    detail: "Session hop TR → BR within 18 minutes",
    source: "session-risk",
    status: "acknowledged",
    createdAt: new Date(Date.now() - 55 * 60_000).toISOString(),
  },
  {
    id: "th3",
    severity: "medium",
    title: "New OAuth client registered",
    detail: "Unknown redirect URI on partner app",
    source: "oauth-registry",
    status: "open",
    createdAt: new Date(Date.now() - 3 * 3600_000).toISOString(),
  },
];

const mockSessions: SuspiciousSession[] = [
  {
    id: "sess_1",
    userEmail: "ops.lead@acme.example",
    deviceId: "dev_9a",
    deviceLabel: "Chrome · Windows",
    ip: "185.22.14.88",
    country: "RU",
    risk: "suspicious",
    reason: "Geo not in allowlist",
    lastSeenAt: new Date(Date.now() - 8 * 60_000).toISOString(),
  },
  {
    id: "sess_2",
    userEmail: "finops@nexora.platform",
    deviceId: "dev_2b",
    deviceLabel: "Safari · macOS",
    ip: "2a02:6b8::1",
    country: "NL",
    risk: "blocked",
    reason: "Token reuse detected",
    lastSeenAt: new Date(Date.now() - 22 * 60_000).toISOString(),
  },
];

const mockRules: GeoIpRule[] = [
  {
    id: "rule_1",
    type: "allow",
    target: "country",
    value: "TR,DE,NL,US,GB",
    label: "Primary ops countries",
    enabled: true,
  },
  {
    id: "rule_2",
    type: "deny",
    target: "country",
    value: "KP,IR",
    label: "Embargoed",
    enabled: true,
  },
  {
    id: "rule_3",
    type: "deny",
    target: "ip_cidr",
    value: "185.22.14.0/24",
    label: "Known botnet range",
    enabled: true,
  },
];

const mockProviders: AuthProvider[] = [
  {
    id: "prov_oidc",
    name: "NEXORA OIDC",
    kind: "oidc",
    enabled: true,
    enforced: true,
    tenantScope: null,
  },
  {
    id: "prov_saml",
    name: "Enterprise SAML",
    kind: "saml",
    enabled: true,
    enforced: false,
    tenantScope: "enterprise",
  },
  {
    id: "prov_google",
    name: "Google OAuth",
    kind: "oauth",
    enabled: true,
    enforced: false,
    tenantScope: null,
  },
  {
    id: "prov_webauthn",
    name: "WebAuthn / passkeys",
    kind: "webauthn",
    enabled: true,
    enforced: false,
    tenantScope: null,
  },
  {
    id: "prov_totp",
    name: "TOTP 2FA",
    kind: "totp",
    enabled: true,
    enforced: true,
    tenantScope: null,
  },
];

const mockPolicies: SecurityPolicy[] = [
  {
    id: "pol_mfa",
    key: "mfa.required",
    name: "MFA required",
    description: "All platform roles must complete MFA",
    value: "true",
    enforced: true,
  },
  {
    id: "pol_session",
    key: "session.max_idle_minutes",
    name: "Session idle timeout",
    description: "Force re-auth after idle",
    value: "30",
    enforced: true,
  },
  {
    id: "pol_password",
    key: "password.min_length",
    name: "Password minimum length",
    description: "Local password policy when applicable",
    value: "14",
    enforced: true,
  },
  {
    id: "pol_device",
    key: "device.trust_required",
    name: "Trusted device required",
    description: "Block unknown devices for platform_owner",
    value: "platform_owner",
    enforced: false,
  },
];

function snapshot(): SecuritySnapshot {
  return {
    loginEvents: [
      {
        id: "le1",
        userEmail: "owner@nexora.platform",
        ip: "81.213.10.22",
        country: "TR",
        city: "Istanbul",
        success: true,
        mfaUsed: true,
        userAgent: "Chrome/128",
        createdAt: new Date(Date.now() - 4 * 60_000).toISOString(),
      },
      {
        id: "le2",
        userEmail: "ops.lead@acme.example",
        ip: "185.22.14.88",
        country: "RU",
        city: "Moscow",
        success: false,
        mfaUsed: false,
        userAgent: "curl/8.4",
        createdAt: new Date(Date.now() - 9 * 60_000).toISOString(),
      },
      {
        id: "le3",
        userEmail: "sre@nexora.platform",
        ip: "52.48.12.9",
        country: "IE",
        city: "Dublin",
        success: true,
        mfaUsed: true,
        userAgent: "Firefox/129",
        createdAt: new Date(Date.now() - 18 * 60_000).toISOString(),
      },
      {
        id: "le4",
        userEmail: "unknown@probe.invalid",
        ip: "103.45.12.8",
        country: "CN",
        city: "Shanghai",
        success: false,
        mfaUsed: false,
        userAgent: "python-requests",
        createdAt: new Date(Date.now() - 21 * 60_000).toISOString(),
      },
    ],
    threats: [...mockThreats],
    sessions: [...mockSessions],
    devices: [
      {
        id: "dev_1",
        userEmail: "owner@nexora.platform",
        label: "YubiKey 5C",
        platform: "WebAuthn",
        trusted: true,
        lastSeenAt: new Date(Date.now() - 4 * 60_000).toISOString(),
        createdAt: "2025-11-01T10:00:00.000Z",
      },
      {
        id: "dev_9a",
        userEmail: "ops.lead@acme.example",
        label: "Chrome · Windows",
        platform: "browser",
        trusted: false,
        lastSeenAt: new Date(Date.now() - 8 * 60_000).toISOString(),
        createdAt: new Date(Date.now() - 2 * 86400_000).toISOString(),
      },
      {
        id: "dev_2b",
        userEmail: "finops@nexora.platform",
        label: "Safari · macOS",
        platform: "browser",
        trusted: true,
        lastSeenAt: new Date(Date.now() - 22 * 60_000).toISOString(),
        createdAt: "2026-01-14T08:00:00.000Z",
      },
    ],
    geoIpRules: [...mockRules],
    providers: [...mockProviders],
    policies: [...mockPolicies],
    failedLogins24h: 14_820,
    openThreats: mockThreats.filter((t) => t.status === "open").length,
    generatedAt: new Date().toISOString(),
  };
}

export async function fetchSecurity(): Promise<SecuritySnapshot> {
  try {
    return await apiClient<SecuritySnapshot>(platformPath("/security"));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return snapshot();
    }
    throw err;
  }
}

export async function acknowledgeThreat(
  threatId: string,
): Promise<ThreatAlert> {
  try {
    return await apiClient<ThreatAlert>(
      platformPath(`/security/threats/${threatId}/ack`),
      { method: "POST", body: {}, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockThreats.findIndex((t) => t.id === threatId);
      if (idx < 0) throw new Error("Threat not found");
      mockThreats[idx] = { ...mockThreats[idx], status: "acknowledged" };
      return mockThreats[idx];
    }
    throw err;
  }
}

export async function revokeSession(
  sessionId: string,
): Promise<SuspiciousSession> {
  try {
    return await apiClient<SuspiciousSession>(
      platformPath(`/security/sessions/${sessionId}/revoke`),
      { method: "POST", body: {}, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockSessions.findIndex((s) => s.id === sessionId);
      if (idx < 0) throw new Error("Session not found");
      mockSessions[idx] = { ...mockSessions[idx], risk: "blocked" };
      return mockSessions[idx];
    }
    throw err;
  }
}

export async function toggleGeoRule(ruleId: string): Promise<GeoIpRule> {
  try {
    return await apiClient<GeoIpRule>(
      platformPath(`/security/geo-ip/${ruleId}/toggle`),
      { method: "POST", body: {}, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockRules.findIndex((r) => r.id === ruleId);
      if (idx < 0) throw new Error("Rule not found");
      mockRules[idx] = { ...mockRules[idx], enabled: !mockRules[idx].enabled };
      return mockRules[idx];
    }
    throw err;
  }
}

export async function toggleProvider(
  providerId: string,
): Promise<AuthProvider> {
  try {
    return await apiClient<AuthProvider>(
      platformPath(`/security/providers/${providerId}/toggle`),
      { method: "POST", body: {}, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockProviders.findIndex((p) => p.id === providerId);
      if (idx < 0) throw new Error("Provider not found");
      mockProviders[idx] = {
        ...mockProviders[idx],
        enabled: !mockProviders[idx].enabled,
      };
      return mockProviders[idx];
    }
    throw err;
  }
}

export async function togglePolicy(
  policyId: string,
): Promise<SecurityPolicy> {
  try {
    return await apiClient<SecurityPolicy>(
      platformPath(`/security/policies/${policyId}/toggle`),
      { method: "POST", body: {}, idempotent: true },
    );
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockPolicies.findIndex((p) => p.id === policyId);
      if (idx < 0) throw new Error("Policy not found");
      mockPolicies[idx] = {
        ...mockPolicies[idx],
        enforced: !mockPolicies[idx].enforced,
      };
      return mockPolicies[idx];
    }
    throw err;
  }
}
