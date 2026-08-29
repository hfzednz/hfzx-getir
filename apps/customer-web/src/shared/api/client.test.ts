import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, createApiClient } from "@nexora/web-core";

function respond(status: number, body: string, contentType = "application/json") {
  return new Response(body, { status, headers: { "Content-Type": contentType } });
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("api client error handling", () => {
  it("calls onUnauthorized exactly once for a 401", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        respond(401, JSON.stringify({ error: { code: "unauthorized", message: "missing access token" } })),
      ),
    );
    const onUnauthorized = vi.fn();
    const api = createApiClient({ baseUrl: "", tenantId: "t1", onUnauthorized });

    await expect(api.request("/v1/customer/home")).rejects.toBeInstanceOf(ApiError);
    expect(onUnauthorized).toHaveBeenCalledTimes(1);
  });

  it("does not call onUnauthorized for a 403", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        respond(403, JSON.stringify({ error: { code: "forbidden", message: "insufficient role" } })),
      ),
    );
    const onUnauthorized = vi.fn();
    const api = createApiClient({ baseUrl: "", tenantId: "t1", onUnauthorized });

    await expect(api.request("/v1/admin/dashboard")).rejects.toMatchObject({ status: 403 });
    expect(onUnauthorized).not.toHaveBeenCalled();
  });

  it("never surfaces a non-JSON upstream body", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        respond(500, "<html>goroutine 1 [running]: panic at order-service</html>", "text/html"),
      ),
    );
    const api = createApiClient({ baseUrl: "", tenantId: "t1" });

    await expect(api.request("/v1/customer/home")).rejects.toMatchObject({
      status: 500,
      message: "Something went wrong. Please try again.",
    });
  });

  it("hides 5xx service messages but keeps 4xx service messages", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        respond(502, JSON.stringify({ error: { code: "upstream", message: "recommendation-service dial tcp refused" } })),
      ),
    );
    const api = createApiClient({ baseUrl: "", tenantId: "t1" });
    await expect(api.request("/v1/customer/home")).rejects.toMatchObject({
      message: "Something went wrong. Please try again.",
    });

    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        respond(409, JSON.stringify({ error: { code: "conflict", message: "order already completed" } })),
      ),
    );
    const conflictApi = createApiClient({ baseUrl: "", tenantId: "t1" });
    await expect(conflictApi.request("/v1/customer/checkout/place", { method: "POST", body: {} })).rejects.toMatchObject(
      { status: 409, message: "order already completed" },
    );
  });

  it("sends tenant, bearer token and idempotency key", async () => {
    const fetchMock = vi.fn(async (_url: string, _init: RequestInit) =>
      respond(200, JSON.stringify({ ok: true })),
    );
    vi.stubGlobal("fetch", fetchMock);
    const api = createApiClient({
      baseUrl: "",
      tenantId: "11111111-1111-1111-1111-111111111111",
      getToken: () => "tok",
    });

    await api.request("/v1/customer/checkout/place", {
      method: "POST",
      body: {},
      idempotencyKey: "key-1",
    });

    const headers = fetchMock.mock.calls[0]![1].headers as Headers;
    expect(headers.get("X-Tenant-Id")).toBe("11111111-1111-1111-1111-111111111111");
    expect(headers.get("Authorization")).toBe("Bearer tok");
    expect(headers.get("Idempotency-Key")).toBe("key-1");
  });
});
