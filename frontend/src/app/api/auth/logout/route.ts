import { NextResponse } from "next/server";

import {
  fetchBackend,
  isSameOriginRequest,
  requestIdFrom,
} from "@/lib/api/server";
import type { ApiResponse } from "@/lib/api/types";
import { clearSessionCookie, sessionToken } from "@/lib/auth/session";

export async function POST(request: Request) {
  const requestId = requestIdFrom(request);

  if (!isSameOriginRequest(request)) {
    return NextResponse.json<ApiResponse<never>>(
      {
        success: false,
        message: "Cross-origin logout requests are not allowed.",
        request_id: requestId,
      },
      { status: 403 },
    );
  }

  const token = await sessionToken();
  if (token) {
    try {
      await fetchBackend("/api/v1/auth/logout", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "X-Request-ID": requestId,
        },
      });
    } catch {
      // Local logout must still succeed if the API is temporarily unavailable.
    }
  }

  const response = NextResponse.json<ApiResponse<never>>({
    success: true,
    message: "Signed out on this device.",
    request_id: requestId,
  });
  clearSessionCookie(response);
  return response;
}
