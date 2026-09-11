import "server-only";

import type { ApiResponse } from "@/lib/api/types";

const DEFAULT_API_URL = "http://localhost:8080";
const API_TIMEOUT_MS = 8_000;

export function backendUrl(path: string) {
  const baseUrl =
    process.env.WMS_API_URL ??
    process.env.NEXT_PUBLIC_API_URL ??
    DEFAULT_API_URL;

  return new URL(path, baseUrl.endsWith("/") ? baseUrl : `${baseUrl}/`);
}

export async function fetchBackend(path: string, init: RequestInit = {}) {
  return fetch(backendUrl(path), {
    ...init,
    cache: "no-store",
    signal: init.signal ?? AbortSignal.timeout(API_TIMEOUT_MS),
  });
}

export async function readApiResponse<T>(response: Response) {
  try {
    return (await response.json()) as ApiResponse<T>;
  } catch {
    return null;
  }
}

export function requestIdFrom(request: Request) {
  return request.headers.get("x-request-id") ?? crypto.randomUUID();
}

export function isSameOriginRequest(request: Request) {
  const origin = request.headers.get("origin");
  return origin === null || origin === new URL(request.url).origin;
}
