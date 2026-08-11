import { describe, expect, it } from "vitest";
import {
  canCancelOrder,
  canForceCancel,
  canForceComplete,
  canRefund,
  canReassign,
  canReplaceOrder,
} from "../order-rules";
import type { Role } from "../permissions";

function session(roles: Role[]) {
  return { roles };
}

describe("order-rules", () => {
  it("allows cancel for support_lead and above", () => {
    expect(canCancelOrder(session(["support_agent"]))).toBe(false);
    expect(canCancelOrder(session(["support_lead"]))).toBe(true);
    expect(canCancelOrder(session(["admin"]))).toBe(true);
  });

  it("allows force-complete only for admin+", () => {
    expect(canForceComplete(session(["city_ops"]))).toBe(false);
    expect(canForceComplete(session(["admin"]))).toBe(true);
    expect(canForceComplete(session(["super_admin"]))).toBe(true);
  });

  it("mirrors force-cancel to force_complete permission", () => {
    expect(canForceCancel(session(["admin"]))).toBe(true);
    expect(canForceCancel(session(["viewer"]))).toBe(false);
  });

  it("allows refund for support and finance paths", () => {
    expect(canRefund(session(["support_agent"]))).toBe(true);
    expect(canRefund(session(["finance_admin"]))).toBe(true);
    expect(canRefund(session(["viewer"]))).toBe(false);
  });

  it("allows reassign for city_ops", () => {
    expect(canReassign(session(["city_ops"]))).toBe(true);
    expect(canReassign(session(["viewer"]))).toBe(false);
  });

  it("allows replace/write for city_ops write holders", () => {
    expect(canReplaceOrder(session(["city_ops"]))).toBe(true);
    expect(canReplaceOrder(session(["viewer"]))).toBe(false);
  });

  it("denies all rules for null session", () => {
    expect(canCancelOrder(null)).toBe(false);
    expect(canForceComplete(undefined)).toBe(false);
    expect(canRefund(null)).toBe(false);
    expect(canReassign(null)).toBe(false);
    expect(canReplaceOrder(null)).toBe(false);
  });
});
