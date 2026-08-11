import { describe, expect, it } from "vitest";
import {
  can,
  hasPlatformRole,
  isPlatformRole,
  permissionsForRoles,
  ROLE_PERMISSIONS,
  type Permission,
} from "@/shared/permissions/platform-permissions";

describe("platform-permissions", () => {
  it("recognizes platform roles", () => {
    expect(isPlatformRole("platform_owner")).toBe(true);
    expect(isPlatformRole("platform_sre")).toBe(true);
    expect(isPlatformRole("city_ops")).toBe(false);
    expect(hasPlatformRole({ roles: ["platform_viewer"] })).toBe(true);
    expect(hasPlatformRole({ roles: [] })).toBe(false);
    expect(hasPlatformRole(null)).toBe(false);
  });

  it("grants viewers read-only coverage of platform modules", () => {
    const perms = permissionsForRoles(["platform_viewer"]);
    const reads: Permission[] = [
      "dashboard:read",
      "dr:read",
      "deployments:read",
      "monitoring:read",
      "notifications:read",
      "audit:read",
      "reports:read",
    ];
    for (const p of reads) {
      expect(perms).toContain(p);
    }
    expect(perms).not.toContain("dr:execute");
    expect(perms).not.toContain("deployments:write");
    expect(perms).not.toContain("dual_control:approve");
    expect(perms).not.toContain("reports:export");
  });

  it("grants SRE DR execute, deployments, and dual-control approve", () => {
    const sre = { roles: ["platform_sre" as const] };
    expect(can(sre, "dr:execute")).toBe(true);
    expect(can(sre, "deployments:write")).toBe(true);
    expect(can(sre, "deployments:rollback")).toBe(true);
    expect(can(sre, "notifications:write")).toBe(true);
    expect(can(sre, "dual_control:approve")).toBe(true);
    expect(can(sre, "tenants:delete")).toBe(false);
  });

  it("grants finops reports export but not DR execute", () => {
    const finops = { roles: ["platform_finops" as const] };
    expect(can(finops, "reports:export")).toBe(true);
    expect(can(finops, "billing:write")).toBe(true);
    expect(can(finops, "dr:execute")).toBe(false);
    expect(can(finops, "dual_control:approve")).toBe(false);
  });

  it("grants compliance audit read and export", () => {
    const compliance = { roles: ["platform_compliance" as const] };
    expect(can(compliance, "audit:read")).toBe(true);
    expect(can(compliance, "reports:export")).toBe(true);
    expect(can(compliance, "compliance:export")).toBe(true);
    expect(can(compliance, "dual_control:approve")).toBe(true);
  });

  it("platform_owner includes dual-control and DR execute", () => {
    const ownerPerms = ROLE_PERMISSIONS.platform_owner;
    expect(ownerPerms).toContain("dr:execute");
    expect(ownerPerms).toContain("deployments:rollback");
    expect(ownerPerms).toContain("dual_control:approve");
    expect(ownerPerms).toContain("reports:export");
  });

  it("honors explicit permissions override on session", () => {
    const session = {
      roles: ["platform_viewer" as const],
      permissions: ["dr:execute"],
    };
    expect(can(session, "dr:execute")).toBe(true);
    expect(can(session, "deployments:write")).toBe(false);
  });

  it("denies when session is missing", () => {
    expect(can(null, "monitoring:read")).toBe(false);
    expect(can(undefined, "audit:read")).toBe(false);
  });

  it("unions permissions across multiple roles", () => {
    const perms = permissionsForRoles(["platform_finops", "platform_sre"]);
    expect(perms).toContain("billing:write");
    expect(perms).toContain("dr:execute");
    expect(perms).toContain("dual_control:approve");
  });
});
