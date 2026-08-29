import { afterEach, describe, expect, it, vi } from "vitest";
import { RoleNotAllowedError, verifyOtp } from "@nexora/web-core";

// roles: ["courier"]
const COURIER_TOKEN = `h.${Buffer.from(JSON.stringify({ roles: ["courier"] })).toString("base64url")}.s`;

function stubVerify() {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () =>
      new Response(
        JSON.stringify({ accessToken: COURIER_TOKEN, principalId: "p1", expiresIn: 3600 }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    ),
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("verifyOtp role enforcement", () => {
  it("rejects a session whose roles do not match the app", async () => {
    stubVerify();
    await expect(
      verifyOtp("ch1", "000000", "+900000000000", "identity", ["customer"]),
    ).rejects.toBeInstanceOf(RoleNotAllowedError);
  });

  it("accepts a session that carries one of the expected roles", async () => {
    stubVerify();
    const session = await verifyOtp("ch1", "000000", "+900000000000", "identity", [
      "courier",
    ]);
    expect(session.roles).toContain("courier");
    expect(session.accessToken).toBe(COURIER_TOKEN);
  });

  it("accepts any role when the app declares no expectation", async () => {
    stubVerify();
    const session = await verifyOtp("ch1", "000000", "+900000000000", "identity", []);
    expect(session.roles).toEqual(["courier"]);
  });
});
