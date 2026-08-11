export interface Department {
  id: string;
  name: string;
  headcount: number;
  roles: string[];
}

export interface RoleRow {
  id: string;
  key: string;
  label: string;
  members: number;
  description: string;
}

export interface PermissionMatrixCell {
  role: string;
  permission: string;
  granted: boolean;
}

export interface CustomPermission {
  id: string;
  key: string;
  description: string;
  createdBy: string;
}

export interface TemporaryGrant {
  id: string;
  user: string;
  permission: string;
  expiresAt: string;
  reason: string;
  status: "active" | "expired" | "revoked";
}

export interface ApprovalWorkflow {
  id: string;
  name: string;
  steps: number;
  pending: number;
  description: string;
}

export interface RbacSnapshot {
  generatedAt: string;
  departments: Department[];
  roles: RoleRow[];
  matrix: PermissionMatrixCell[];
  matrixPermissions: string[];
  matrixRoles: string[];
  customPermissions: CustomPermission[];
  temporaryGrants: TemporaryGrant[];
  approvals: ApprovalWorkflow[];
}
