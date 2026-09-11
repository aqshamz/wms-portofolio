import type { ApiResponse } from "@/lib/api/types";

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly details?: unknown,
    public readonly requestId?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

type RequestOptions = Omit<RequestInit, "body"> & {
  body?: unknown;
};

export const SESSION_EXPIRED_EVENT = "wms:session-expired";

export async function apiRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");
  headers.set("X-Request-ID", crypto.randomUUID());

  if (options.body !== undefined) {
    headers.set("Content-Type", "application/json");
  }

  const proxyPath = `/api/backend${path.startsWith("/") ? path : `/${path}`}`;
  const response = await fetch(proxyPath, {
    ...options,
    headers,
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });

  const payload = (await response.json()) as ApiResponse<T>;

  if (!response.ok || !payload.success) {
    if (response.status === 401 && typeof window !== "undefined") {
      window.dispatchEvent(new Event(SESSION_EXPIRED_EVENT));
    }

    throw new ApiError(
      payload.message || "The request could not be completed.",
      response.status,
      payload.error,
      payload.request_id,
    );
  }

  return payload.data as T;
}
