import type { Id } from "@/shared/types/common";
import type { Role } from "@/shared/permissions/permissions";

export interface AdminSession {
  userId: Id;
  email: string;
  displayName: string;
  roles: Role[];
  permissions: string[];
  cityIds: Id[];
  mfaVerified: boolean;
}
