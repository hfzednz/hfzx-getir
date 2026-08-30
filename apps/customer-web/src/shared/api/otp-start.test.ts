import { afterEach, describe, expect, it, vi } from "vitest";
import { startOtp } from "@nexora/web-core";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("startOtp customer-bff", () => {
  it("POSTs same-origin /v1/customer/auth/otp/start and reads challengeId", async () => {
    const fetchMock = vi.fn(async (url: string, init?: RequestInit) => {
      expect(url).toBe("/v1/customer/auth/otp/start");
      expect(init?.method).toBe("POST");
      expect(JSON.parse(String(init?.body))).toEqual({ phone: "+905551112233" });
      expect(new Headers(init?.headers).get("X-Tenant-Id")).toBeTruthy();
      expect(init?.credentials).toBe("include");
      return new Response(JSON.stringify({ challengeId: "ch-1", expiresIn: 300 }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    });
    vi.stubGlobal("fetch", fetchMock);

    const res = await startOtp("+905551112233", "customer-bff");
    expect(res.challengeId).toBe("ch-1");
  });

  it("throws a user-facing error when the body has no challenge id", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response(JSON.stringify({ expiresIn: 300 }), { status: 200 })),
    );
    await expect(startOtp("+905551112233", "customer-bff")).rejects.toThrow(
      /Could not start verification/,
    );
  });

  it("throws a user-facing error when the network fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new TypeError("Failed to fetch");
      }),
    );
    await expect(startOtp("+905551112233", "customer-bff")).rejects.toThrow(
      /Could not reach the server/,
    );
  });
});
