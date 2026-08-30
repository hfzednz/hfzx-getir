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
  /** Called once per 401 so the app can drop the expired session and re-authenticate. */
  onUnauthorized?: () => void;
}

const GENERIC_MESSAGES: Record<number, string> = {
  400: "The request was invalid.",
  401: "Your session has expired. Please sign in again.",
  403: "You are not allowed to perform this action.",
  404: "Not found.",
  409: "This action conflicts with the current state.",
  422: "The submitted data could not be processed.",
  429: "Too many requests. Please wait and try again.",
};

function genericMessage(status: number): string {
  return GENERIC_MESSAGES[status] ?? "Something went wrong. Please try again.";
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

    let res: Response;
    try {
      res = await fetch(url, {
        ...rest,
        // Codespace/Safari tunnels set partitioned cookies; default "same-origin"
        // is not reliable on iOS for *.app.github.dev.
        credentials: "include",
        cache: "no-store",
        signal:
          rest.signal ??
          (typeof AbortSignal !== "undefined" && "timeout" in AbortSignal
            ? AbortSignal.timeout(15_000)
            : undefined),
        headers,
        body: body === undefined ? undefined : JSON.stringify(body),
      });
    } catch (err) {
      if (err instanceof ApiError) throw err;
      const name = err instanceof Error ? err.name : "";
      if (err instanceof TypeError || name === "AbortError" || name === "TimeoutError") {
        throw new ApiError(0, {
          code: "network_error",
          message: "Could not reach the server. Please try again.",
        });
      }
      throw err;
    }

    const text = await res.text();
    let parsed: unknown = null;
    if (text) {
      try {
        parsed = JSON.parse(text);
      } catch {
        // Upstream sent a non-JSON body (HTML error page, panic trace). Never surface it.
        parsed = res.ok ? null : { error: { code: `http_${res.status}` } };
      }
    }

    if (!res.ok) {
      if (res.status === 401) {
        opts.onUnauthorized?.();
      }
      const errBody = parsed as ApiErrorBody | null;
      const code = errBody?.error?.code ?? `http_${res.status}`;
      const upstreamMessage = errBody?.error?.message;
      throw new ApiError(res.status, {
        code,
        // Only trust a message that came from a structured service error envelope,
        // and never relay 5xx internals to the UI.
        message:
          upstreamMessage && res.status < 500
            ? upstreamMessage
            : genericMessage(res.status),
        traceId: errBody?.error?.traceId,
      });
    }

    return parsed as T;
  }

  return { request };
}
