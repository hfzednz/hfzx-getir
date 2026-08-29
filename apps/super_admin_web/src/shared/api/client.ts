import type { ApiErrorBody } from "@/shared/types/common";
import { createIdempotencyKey } from "@/shared/lib/idempotency";
import { tenantId } from "@nexora/web-core";
import { useAuthStore } from "@/shared/auth/auth-store";

const DEFAULT_BASE = "http://localhost:8110/v1";

export class ApiError extends Error {
  readonly code: string;
  readonly traceId: string;
  readonly status: number;

  constructor(status: number, body: ApiErrorBody["error"]) {
    super(body.message);
    this.name = "ApiError";
    this.status = status;
    this.code = body.code;
    this.traceId = body.traceId;
  }
}

export interface ApiRequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown;
  /** When true, attaches a fresh Idempotency-Key header. */
  idempotent?: boolean;
  /** Override Idempotency-Key (also sets the header). */
  idempotencyKey?: string;
  token?: string | null;
}

function getBaseUrl(): string {
  if (typeof window !== "undefined") {
    return "/v1";
  }
  return (
    process.env.NEXT_PUBLIC_PLATFORM_OPS_URL?.replace(/\/$/, "") ??
    process.env.NEXT_PUBLIC_BFF_ADMIN_URL?.replace(/\/$/, "") ??
    DEFAULT_BASE
  );
}

/** Prefix a relative path under the platform BFF namespace. */
export function platformPath(path: string): string {
  const normalized = path.startsWith("/") ? path : `/${path}`;
  if (normalized === "/platform" || normalized.startsWith("/platform/")) {
    return normalized;
  }
  return `/platform${normalized}`;
}

function toCamelKey(key: string): string {
  return key.replace(/_([a-z])/g, (_, c: string) => c.toUpperCase());
}

function keysToCamelCase(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(keysToCamelCase);
  }
  if (value !== null && typeof value === "object" && !(value instanceof Date)) {
    const out: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
      out[toCamelKey(k)] = keysToCamelCase(v);
    }
    return out;
  }
  return value;
}

export async function apiClient<T>(
  path: string,
  options: ApiRequestOptions = {},
): Promise<T> {
  const {
    body,
    idempotent,
    idempotencyKey,
    token,
    headers: initHeaders,
    ...rest
  } = options;

  const headers = new Headers(initHeaders);
  if (!headers.has("Accept")) {
    headers.set("Accept", "application/json");
  }
  if (body !== undefined && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  } else {
    const sessionToken = useAuthStore.getState().session?.accessToken;
    if (sessionToken) {
      headers.set("Authorization", `Bearer ${sessionToken}`);
    }
  }
  if (!headers.has("X-Tenant-Id")) {
    headers.set("X-Tenant-Id", tenantId());
  }

  const key = idempotencyKey ?? (idempotent ? createIdempotencyKey() : undefined);
  if (key) {
    headers.set("Idempotency-Key", key);
  }

  const url = path.startsWith("http")
    ? path
    : `${getBaseUrl()}${path.startsWith("/") ? path : `/${path}`}`;

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
      parsed = { error: { code: "invalid_json", message: text, traceId: "" } };
    }
  }

  if (!res.ok) {
    const errBody = parsed as ApiErrorBody | null;
    const error = errBody?.error ?? {
      code: `http_${res.status}`,
      message: res.statusText || "Request failed",
      traceId: "",
    };
    throw new ApiError(res.status, error);
  }

  return keysToCamelCase(parsed) as T;
}

/** Convenience: mint an Idempotency-Key for mutating calls. */
export { createIdempotencyKey };
