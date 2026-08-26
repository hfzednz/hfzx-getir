import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type { OrgSnapshot, PlatformPerson } from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

function mockSnapshot(): OrgSnapshot {
  const people: PlatformPerson[] = [
    {
      id: "p1",
      name: "Hafize Platform",
      email: "owner@nexora.platform",
      kind: "platform_admin",
      orgUnitId: "org_root",
      orgUnitName: "NEXORA Platform",
      status: "active",
    },
    {
      id: "p2",
      name: "Selin SRE",
      email: "sre@nexora.platform",
      kind: "employee",
      orgUnitId: "dept_sre",
      orgUnitName: "SRE",
      status: "active",
    },
    {
      id: "p3",
      name: "Mert FinOps",
      email: "finops@nexora.platform",
      kind: "manager",
      orgUnitId: "dept_finops",
      orgUnitName: "FinOps",
      status: "active",
    },
    {
      id: "p4",
      name: "External Auditor Co",
      email: "audit@external.example",
      kind: "auditor",
      orgUnitId: null,
      orgUnitName: null,
      status: "active",
    },
    {
      id: "p5",
      name: "Cloud Partner",
      email: "partner@cloud.example",
      kind: "partner",
      orgUnitId: null,
      orgUnitName: null,
      status: "invited",
    },
    {
      id: "p6",
      name: "Payment Supplier",
      email: "ops@payments.example",
      kind: "supplier",
      orgUnitId: null,
      orgUnitName: null,
      status: "active",
    },
    {
      id: "p7",
      name: "Guest Analyst",
      email: "guest@external.example",
      kind: "external_user",
      orgUnitId: "team_compliance",
      orgUnitName: "Compliance reviews",
      status: "disabled",
    },
  ];

  return {
    generatedAt: new Date().toISOString(),
    organizations: [
      {
        id: "org_root",
        name: "NEXORA Platform",
        parentId: null,
        type: "organization",
        headcount: 86,
      },
    ],
    departments: [
      {
        id: "dept_sre",
        name: "SRE",
        parentId: "org_root",
        type: "department",
        headcount: 18,
      },
      {
        id: "dept_finops",
        name: "FinOps",
        parentId: "org_root",
        type: "department",
        headcount: 12,
      },
      {
        id: "dept_security",
        name: "Security",
        parentId: "org_root",
        type: "department",
        headcount: 14,
      },
      {
        id: "dept_compliance",
        name: "Compliance",
        parentId: "org_root",
        type: "department",
        headcount: 9,
      },
    ],
    teams: [
      {
        id: "team_platform_ctrl",
        name: "Platform control plane",
        parentId: "dept_sre",
        type: "team",
        headcount: 6,
      },
      {
        id: "team_compliance",
        name: "Compliance reviews",
        parentId: "dept_compliance",
        type: "team",
        headcount: 4,
      },
      {
        id: "team_identity",
        name: "Identity & access",
        parentId: "dept_security",
        type: "team",
        headcount: 5,
      },
    ],
    people,
  };
}

export async function fetchOrgSnapshot(): Promise<OrgSnapshot> {
  try {
    return await apiClient<OrgSnapshot>(platformPath("/org"));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockSnapshot();
    }
    throw err;
  }
}
