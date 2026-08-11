import { describe, expect, it } from "vitest";
import {
  can,
  permissionsForRoles,
  ROLE_PERMISSIONS,
  type Permission,
  type Role,
} from "../permissions";

describe("permissionsForRoles", () => {
  it("unions permissions across multiple roles", () => {
    const perms = permissionsForRoles(["viewer", "finance_analyst"]);
    expect(perms).toContain("dashboard:read");
    expect(perms).toContain("finance:export");
    expect(perms).toContain("reports:export");
  });

  it("gives super_admin system:flags and rbac:write", () => {
    const perms = permissionsForRoles(["super_admin"]);
    expect(perms).toContain("system:flags");
    expect(perms).toContain("rbac:write");
    expect(perms).toContain("finance:payout:approve");
  });

  it("does not give admin kill-switch flag permission", () => {
    const perms = permissionsForRoles(["admin"]);
    expect(perms).toContain("system:write");
    expect(perms).not.toContain("system:flags");
  });
});

describe("can", () => {
  it("returns false for null session", () => {
    expect(can(null, "orders:read")).toBe(false);
  });

  it("checks role map", () => {
    const session = { roles: ["city_ops" as Role] };
    expect(can(session, "orders:reassign")).toBe(true);
    expect(can(session, "orders:force_complete")).toBe(false);
  });

  it("honors explicit permissions override list", () => {
    const session = {
      roles: ["viewer" as Role],
      permissions: ["orders:cancel" as Permission],
    };
    expect(can(session, "orders:cancel")).toBe(true);
  });
});

describe("ROLE_PERMISSIONS integrity", () => {
  it("includes READ_OPS baseline for every role", () => {
    const roles = Object.keys(ROLE_PERMISSIONS) as Role[];
    for (const role of roles) {
      expect(ROLE_PERMISSIONS[role]).toContain("dashboard:read");
      expect(ROLE_PERMISSIONS[role]).toContain("analytics:read");
      expect(ROLE_PERMISSIONS[role]).toContain("monitoring:read");
    }
  });

  it("only super_admin has system:flags", () => {
    const withFlags = (Object.entries(ROLE_PERMISSIONS) as [Role, Permission[]][])
      .filter(([, perms]) => perms.includes("system:flags"))
      .map(([role]) => role);
    expect(withFlags).toEqual(["super_admin"]);
  });
});
