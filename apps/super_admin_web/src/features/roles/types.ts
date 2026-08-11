import type { Id } from "@/shared/types/common";

export type RoleScope = "global" | "company" | "department";

export interface PlatformRoleTemplate {
  id: Id;
  key: string;
  label: string;
  scope: RoleScope;
  members: number;
  inheritsFrom: string | null;
  description: string;
}

export interface PermissionTemplate {
  id: Id;
  key: string;
  label: string;
  permissions: string[];
}

export interface ApprovalChain {
  id: Id;
  name: string;
  steps: string[];
  pending: number;
  description: string;
}

export interface RoleInheritanceEdge {
  id: Id;
  childRole: string;
  parentRole: string;
}

export interface TemporaryPermission {
  id: Id;
  subject: string;
  permission: string;
  scope: RoleScope;
  expiresAt: string;
  reason: string;
  status: "active" | "expired" | "revoked";
}

export interface RolesSnapshot {
  generatedAt: string;
  roles: PlatformRoleTemplate[];
  permissionTemplates: PermissionTemplate[];
  approvalChains: ApprovalChain[];
  inheritance: RoleInheritanceEdge[];
  temporaryPermissions: TemporaryPermission[];
}
