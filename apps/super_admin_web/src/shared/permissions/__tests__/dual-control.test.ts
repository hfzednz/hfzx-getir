import { describe, expect, it } from "vitest";
import {
  canApprove,
  requiresDualControl,
  type DualControlProposal,
} from "@/shared/permissions/dual-control";
import type { PlatformSession } from "@/shared/auth/session";
import { permissionsForRoles } from "@/shared/permissions/platform-permissions";
import { platformPath } from "@/shared/api/client";

function session(
  overrides: Partial<PlatformSession> & Pick<PlatformSession, "userId" | "roles">,
): PlatformSession {
  return {
    email: "a@nexora.local",
    displayName: "Tester",
    permissions: permissionsForRoles(overrides.roles),
    mfaVerified: true,
    webauthnVerified: false,
    ...overrides,
  };
}

describe("platformPath", () => {
  it("prefixes /platform", () => {
    expect(platformPath("/tenants")).toBe("/platform/tenants");
    expect(platformPath("tenants")).toBe("/platform/tenants");
    expect(platformPath("/platform/flags")).toBe("/platform/flags");
  });
});

describe("dual-control", () => {
  const proposal: DualControlProposal = {
    id: "prop_1",
    action: "kill_switch",
    requesterId: "usr_a",
    status: "pending",
    createdAt: new Date().toISOString(),
  };

  it("requires dual control for sensitive actions", () => {
    expect(requiresDualControl("kill_switch")).toBe(true);
    expect(requiresDualControl("tenant_delete")).toBe(true);
    expect(requiresDualControl("tenant_suspend")).toBe(true);
    expect(requiresDualControl("dr_failover")).toBe(true);
    expect(requiresDualControl("secret_rotate")).toBe(true);
    expect(requiresDualControl("license_override")).toBe(true);
    expect(requiresDualControl("flags:read")).toBe(false);
    expect(requiresDualControl("deployments:rollback")).toBe(false);
  });

  it("blocks self-approval", () => {
    const requester = session({
      userId: "usr_a",
      roles: ["platform_owner"],
    });
    expect(canApprove(requester, proposal)).toBe(false);
  });

  it("allows distinct approver with dual_control:approve", () => {
    const approver = session({
      userId: "usr_b",
      roles: ["platform_security"],
    });
    expect(canApprove(approver, proposal)).toBe(true);
  });

  it("denies viewer without approve permission", () => {
    const viewer = session({
      userId: "usr_c",
      roles: ["platform_viewer"],
    });
    expect(canApprove(viewer, proposal)).toBe(false);
  });

  it("blocks approval when proposal is not pending", () => {
    const approver = session({
      userId: "usr_b",
      roles: ["platform_sre"],
    });
    expect(
      canApprove(approver, { ...proposal, status: "executed" }),
    ).toBe(false);
    expect(
      canApprove(approver, { ...proposal, status: "rejected" }),
    ).toBe(false);
  });

  it("allows SRE to approve DR failover from another requester", () => {
    const drProposal: DualControlProposal = {
      id: "prop_dr",
      action: "dr_failover",
      requesterId: "usr_owner",
      status: "pending",
      payload: { from: "eu-west-1", to: "eu-south-1" },
      createdAt: new Date().toISOString(),
    };
    const sre = session({
      userId: "usr_sre",
      roles: ["platform_sre"],
    });
    expect(canApprove(sre, drProposal)).toBe(true);
    expect(
      canApprove(session({ userId: "usr_owner", roles: ["platform_owner"] }), drProposal),
    ).toBe(false);
  });

  it("allows security to approve secret rotation", () => {
    const secretProposal: DualControlProposal = {
      id: "prop_sec",
      action: "secret_rotate",
      requesterId: "usr_sre",
      status: "pending",
      createdAt: new Date().toISOString(),
    };
    expect(
      canApprove(
        session({ userId: "usr_sec", roles: ["platform_security"] }),
        secretProposal,
      ),
    ).toBe(true);
  });

  it("denies null session", () => {
    expect(canApprove(null, proposal)).toBe(false);
    expect(canApprove(undefined, proposal)).toBe(false);
  });
});
