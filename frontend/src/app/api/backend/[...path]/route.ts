import { NextResponse, type NextRequest } from "next/server";

import {
  fetchBackend,
  isSameOriginRequest,
  requestIdFrom,
} from "@/lib/api/server";
import type { ApiResponse } from "@/lib/api/types";
import { clearSessionCookie, sessionToken } from "@/lib/auth/session";

const MUTATING_METHODS = new Set(["POST", "PUT", "PATCH", "DELETE"]);
const FORWARDED_REQUEST_HEADERS = [
  "accept",
  "content-type",
  "idempotency-key",
  "if-match",
  "x-request-id",
];

async function proxyRequest(
  request: NextRequest,
  context: RouteContext<"/api/backend/[...path]">,
) {
  const requestId = requestIdFrom(request);

  if (MUTATING_METHODS.has(request.method) && !isSameOriginRequest(request)) {
    return NextResponse.json<ApiResponse<never>>(
      {
        success: false,
        message: "Cross-origin API requests are not allowed.",
        request_id: requestId,
      },
      { status: 403 },
    );
  }

  const token = await sessionToken();
  if (!token) {
    return NextResponse.json<ApiResponse<never>>(
      {
        success: false,
        message: "Authentication required.",
        request_id: requestId,
      },
      { status: 401 },
    );
  }

  const { path } = await context.params;
  if (
    path.length < 2 ||
    path[0] !== "api" ||
    path[1] !== "v1" ||
    path.some((segment) => segment === "." || segment === "..")
  ) {
    return NextResponse.json<ApiResponse<never>>(
      {
        success: false,
        message: "Only WMS API v1 routes can be proxied.",
        request_id: requestId,
      },
      { status: 400 },
    );
  }

  const headers = new Headers({
    Authorization: `Bearer ${token}`,
    "X-Request-ID": requestId,
  });
  for (const headerName of FORWARDED_REQUEST_HEADERS) {
    const value = request.headers.get(headerName);
    if (value) {
      headers.set(headerName, value);
    }
  }

  const backendPath = `/${path.map(encodeURIComponent).join("/")}${request.nextUrl.search}`;

  try {
    const backendResponse = await fetchBackend(backendPath, {
      method: request.method,
      headers,
      body:
        request.method === "GET" || request.method === "HEAD"
          ? undefined
          : await request.arrayBuffer(),
    });
    const responseHeaders = new Headers();
    for (const headerName of [
      "content-type",
      "content-disposition",
      "x-request-id",
    ]) {
      const value = backendResponse.headers.get(headerName);
      if (value) {
        responseHeaders.set(headerName, value);
      }
    }

    const response = new NextResponse(backendResponse.body, {
      status: backendResponse.status,
      headers: responseHeaders,
    });
    if (backendResponse.status === 401) {
      clearSessionCookie(response);
    }
    return response;
  } catch {
    return NextResponse.json<ApiResponse<never>>(
      {
        success: false,
        message: "The WMS API is temporarily unavailable.",
        request_id: requestId,
      },
      { status: 503 },
    );
  }
}

export const GET = proxyRequest;
export const POST = proxyRequest;
export const PUT = proxyRequest;
export const PATCH = proxyRequest;
export const DELETE = proxyRequest;
export const HEAD = proxyRequest;
