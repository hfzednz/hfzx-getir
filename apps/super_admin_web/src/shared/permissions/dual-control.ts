import type { Id } from "@/shared/types/common";
import type { PlatformSession } from "@/shared/auth/session";
import { can } from "@/shared/permissions/platform-permissions";

/** Actions that require a second distinct approver before execution. */
export type DualControlAction =
  | "kill_switch"
  | "tenant_suspend"
  | "tenant_delete"
  | "dr_failover"
  | "secret_rotate"
  | "license_override";

const DUAL_CONTROL_ACTIONS = new Set<DualControlAction>([
  "kill_switch",
  "tenant_suspend",
  "tenant_delete",
  "dr_failover",
  "secret_rotate",
  "license_override",
]);

export interface DualControlProposal {
  id: Id;
  action: DualControlAction;
  requesterId: Id;
  status: "pending" | "approved" | "rejected" | "executed";
  payload?: Record<string, unknown>;
  createdAt: string;
}

export function requiresDualControl(action: string): action is DualControlAction {
  return DUAL_CONTROL_ACTIONS.has(action as DualControlAction);
}

/**
 * Approver must hold dual_control:approve, must differ from the requester,
 * and the proposal must still be pending.
 */
export function canApprove(
  session: PlatformSession | null | undefined,
  proposal: DualControlProposal,
): boolean {
  if (!session) return false;
  if (proposal.status !== "pending") return false;
  if (session.userId === proposal.requesterId) return false;
  if (!can(session, "dual_control:approve")) return false;
  if (!requiresDualControl(proposal.action)) return false;
  return true;
}
