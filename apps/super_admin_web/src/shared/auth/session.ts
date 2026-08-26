import type { Id } from "@/shared/types/common";
import type { Role } from "@/shared/permissions/platform-permissions";

export interface PlatformSession {
  userId: Id;
  email: string;
  displayName: string;
  roles: Role[];
  permissions: string[];
  mfaVerified: boolean;
  webauthnVerified: boolean;
  accessToken?: string;
}
