import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type {
  ComplianceSnapshot,
  PrivacyRequest,
  RetentionPolicy,
} from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

let mockRequests: PrivacyRequest[] = [
  {
    id: "pr_1",
    type: "export",
    regime: "gdpr",
    subjectEmail: "customer.a@example.com",
    tenantName: "ACME Quick Commerce",
    status: "in_progress",
    dueAt: new Date(Date.now() + 18 * 86400_000).toISOString(),
    createdAt: new Date(Date.now() - 5 * 86400_000).toISOString(),
    assignee: "compliance@nexora.platform",
  },
  {
    id: "pr_2",
    type: "delete",
    regime: "kvkk",
    subjectEmail: "musteri@ornek.tr",
    tenantName: "Nova Market",
    status: "verifying",
    dueAt: new Date(Date.now() + 10 * 86400_000).toISOString(),
    createdAt: new Date(Date.now() - 2 * 86400_000).toISOString(),
    assignee: null,
  },
  {
    id: "pr_3",
    type: "access",
    regime: "ccpa",
    subjectEmail: "user.ca@example.com",
    tenantName: "Orbit Enterprise",
    status: "received",
    dueAt: new Date(Date.now() + 35 * 86400_000).toISOString(),
    createdAt: new Date(Date.now() - 1 * 86400_000).toISOString(),
    assignee: null,
  },
  {
    id: "pr_4",
    type: "delete",
    regime: "gdpr",
    subjectEmail: "erase.me@example.de",
    tenantName: "ACME Quick Commerce",
    status: "completed",
    dueAt: new Date(Date.now() - 2 * 86400_000).toISOString(),
    createdAt: new Date(Date.now() - 28 * 86400_000).toISOString(),
    assignee: "compliance@nexora.platform",
  },
];

const mockRetention: RetentionPolicy[] = [
  {
    id: "ret_1",
    dataClass: "Order history",
    regime: "gdpr",
    retentionDays: 730,
    legalHoldExempt: true,
    autoDelete: true,
  },
  {
    id: "ret_2",
    dataClass: "Customer PII",
    regime: "kvkk",
    retentionDays: 365,
    legalHoldExempt: true,
    autoDelete: false,
  },
  {
    id: "ret_3",
    dataClass: "Marketing consent logs",
    regime: "ccpa",
    retentionDays: 1095,
    legalHoldExempt: false,
    autoDelete: true,
  },
  {
    id: "ret_4",
    dataClass: "Audit trail",
    regime: "global",
    retentionDays: 2555,
    legalHoldExempt: true,
    autoDelete: false,
  },
];

function snapshot(): ComplianceSnapshot {
  return {
    regimes: [
      {
        regime: "gdpr",
        label: "GDPR (EU/EEA)",
        status: "compliant",
        openRequests: mockRequests.filter(
          (r) => r.regime === "gdpr" && r.status !== "completed" && r.status !== "rejected",
        ).length,
        retentionAligned: true,
        dpaSigned: true,
        notes: "SCCs v2021 in force for subprocessors",
      },
      {
        regime: "kvkk",
        label: "KVKK (Türkiye)",
        status: "gaps",
        openRequests: mockRequests.filter(
          (r) => r.regime === "kvkk" && r.status !== "completed" && r.status !== "rejected",
        ).length,
        retentionAligned: false,
        dpaSigned: true,
        notes: "VERBİS registry refresh due Q3",
      },
      {
        regime: "ccpa",
        label: "CCPA / CPRA (California)",
        status: "compliant",
        openRequests: mockRequests.filter(
          (r) => r.regime === "ccpa" && r.status !== "completed" && r.status !== "rejected",
        ).length,
        retentionAligned: true,
        dpaSigned: true,
        notes: "Do-not-sell honoured via preference centre",
      },
    ],
    retention: [...mockRetention],
    consents: [
      {
        id: "c1",
        subjectId: "sub_1001",
        purpose: "Marketing email",
        regime: "gdpr",
        status: "granted",
        channel: "app",
        updatedAt: new Date(Date.now() - 10 * 86400_000).toISOString(),
      },
      {
        id: "c2",
        subjectId: "sub_1002",
        purpose: "Location for delivery",
        regime: "kvkk",
        status: "granted",
        channel: "checkout",
        updatedAt: new Date(Date.now() - 3 * 86400_000).toISOString(),
      },
      {
        id: "c3",
        subjectId: "sub_1003",
        purpose: "Sale of personal info",
        regime: "ccpa",
        status: "denied",
        channel: "privacy_centre",
        updatedAt: new Date(Date.now() - 1 * 86400_000).toISOString(),
      },
      {
        id: "c4",
        subjectId: "sub_1004",
        purpose: "Push notifications",
        regime: "gdpr",
        status: "withdrawn",
        channel: "app",
        updatedAt: new Date(Date.now() - 6 * 3600_000).toISOString(),
      },
    ],
    requests: [...mockRequests],
    generatedAt: new Date().toISOString(),
  };
}

export async function fetchCompliance(): Promise<ComplianceSnapshot> {
  try {
    return await apiClient<ComplianceSnapshot>(platformPath("/compliance"));
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return snapshot();
    }
    throw err;
  }
}

export async function advancePrivacyRequest(input: {
  requestId: string;
  status: PrivacyRequest["status"];
  assignee?: string | null;
}): Promise<PrivacyRequest> {
  try {
    return await apiClient<PrivacyRequest>(
      platformPath(`/compliance/requests/${input.requestId}`),
      { method: "PATCH", body: input, idempotent: true },
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockRequests.findIndex((r) => r.id === input.requestId);
      if (idx < 0) throw new Error("Request not found");
      mockRequests[idx] = {
        ...mockRequests[idx],
        status: input.status,
        assignee:
          input.assignee !== undefined
            ? input.assignee
            : mockRequests[idx].assignee,
      };
      return mockRequests[idx];
    }
    throw err;
  }
}

export async function createPrivacyRequest(input: {
  type: PrivacyRequest["type"];
  regime: PrivacyRequest["regime"];
  subjectEmail: string;
  tenantName: string;
}): Promise<PrivacyRequest> {
  try {
    return await apiClient<PrivacyRequest>(
      platformPath("/compliance/requests"),
      { method: "POST", body: input, idempotent: true },
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const dueDays = input.type === "delete" || input.type === "export" ? 30 : 45;
      const created: PrivacyRequest = {
        id: `pr_${Date.now()}`,
        type: input.type,
        regime: input.regime,
        subjectEmail: input.subjectEmail,
        tenantName: input.tenantName,
        status: "received",
        dueAt: new Date(Date.now() + dueDays * 86400_000).toISOString(),
        createdAt: new Date().toISOString(),
        assignee: null,
      };
      mockRequests = [created, ...mockRequests];
      return created;
    }
    throw err;
  }
}

export async function updateRetention(
  policyId: string,
  patch: Partial<Pick<RetentionPolicy, "retentionDays" | "autoDelete">>,
): Promise<RetentionPolicy> {
  try {
    return await apiClient<RetentionPolicy>(
      platformPath(`/compliance/retention/${policyId}`),
      { method: "PATCH", body: patch, idempotent: true },
    );
  } catch (err) {
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = mockRetention.findIndex((r) => r.id === policyId);
      if (idx < 0) throw new Error("Policy not found");
      mockRetention[idx] = { ...mockRetention[idx], ...patch };
      return mockRetention[idx];
    }
    throw err;
  }
}
