import type { Id } from "@/shared/types/common";

export interface OrgUnit {
  id: Id;
  name: string;
  parentId: Id | null;
  type: "organization" | "department" | "team";
  headcount: number;
}

export type PlatformPersonKind =
  | "platform_admin"
  | "employee"
  | "manager"
  | "external_user"
  | "partner"
  | "supplier"
  | "auditor";

export interface PlatformPerson {
  id: Id;
  name: string;
  email: string;
  kind: PlatformPersonKind;
  orgUnitId: Id | null;
  orgUnitName: string | null;
  status: "active" | "invited" | "disabled";
}

export interface OrgSnapshot {
  generatedAt: string;
  organizations: OrgUnit[];
  departments: OrgUnit[];
  teams: OrgUnit[];
  people: PlatformPerson[];
}
