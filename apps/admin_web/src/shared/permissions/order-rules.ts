import { can, type SessionLike } from "./permissions";

/** Cancel: support_lead+ (orders:cancel). */
export function canCancelOrder(session: SessionLike | null | undefined): boolean {
  return can(session, "orders:cancel");
}

/** Force-complete: admin+ (orders:force_complete). */
export function canForceComplete(
  session: SessionLike | null | undefined,
): boolean {
  return can(session, "orders:force_complete");
}

/** Force-cancel: admin+ (orders:force_complete). */
export function canForceCancel(
  session: SessionLike | null | undefined,
): boolean {
  return can(session, "orders:force_complete");
}

/** Refund: finance/support paths (orders:refund). */
export function canRefund(session: SessionLike | null | undefined): boolean {
  return can(session, "orders:refund");
}

/** Reassign courier: city_ops (orders:reassign). */
export function canReassign(session: SessionLike | null | undefined): boolean {
  return can(session, "orders:reassign");
}

/** Replace items / recreate order: city_ops+ write (orders:write). */
export function canReplaceOrder(
  session: SessionLike | null | undefined,
): boolean {
  return can(session, "orders:write");
}
