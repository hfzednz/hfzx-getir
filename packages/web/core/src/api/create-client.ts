export interface ApiErrorBody {
  error: { code: string; message: string; traceId?: string; retriable?: boolean };
}

export class ApiError extends Error {
  readonly code: string;
  readonly status: number;
  readonly traceId: string;

  constructor(status: number, body: ApiErrorBody["error"]) {
    super(body.message);
    this.name = "ApiError";
    this.status = status;
    this.code = body.code;
    this.traceId = body.traceId ?? "";
  }
}

export interface ClientOptions {
  baseUrl: string;
  tenantId: string;
  getToken?: () => string | null;
  getUserId?: () => string | null;
}

export function createApiClient(opts: ClientOptions) {
  type RequestOpts = Omit<RequestInit, "body"> & {
    body?: unknown;
    idempotencyKey?: string;
  };

  async function request<T>(path: string, init: RequestOpts = {}): Promise<T> {
    const { body, idempotencyKey, headers: initHeaders, ...rest } = init;
    const headers = new Headers(initHeaders);
    if (!headers.has("Accept")) {
      headers.set("Accept", "application/json");
    }
    headers.set("X-Tenant-Id", opts.tenantId);
    if (body !== undefined) {
      headers.set("Content-Type", "application/json");
    }
    const token = opts.getToken?.();
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
    const userId = opts.getUserId?.();
    if (userId) {
      headers.set("X-Nexora-User", userId);
    }
    if (idempotencyKey) {
      headers.set("Idempotency-Key", idempotencyKey);
    }

    const url = path.startsWith("http")
      ? path
      : `${opts.baseUrl}${path.startsWith("/") ? path : `/${path}`}`;

    const res = await fetch(url, {
      ...rest,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });

    const text = await res.text();
    let parsed: unknown = null;
    if (text) {
      try {
        parsed = JSON.parse(text);
      } catch {
        parsed = { error: { code: "invalid_json", message: text } };
      }
    }

    if (!res.ok) {
      const errBody = parsed as ApiErrorBody | null;
      throw new ApiError(
        res.status,
        errBody?.error ?? {
          code: `http_${res.status}`,
          message: res.statusText || "Request failed",
        },
      );
    }

    return parsed as T;
  }

  return { request };
}
