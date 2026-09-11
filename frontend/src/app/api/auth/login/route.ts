import { NextResponse } from "next/server";

import { loginSchema } from "@/features/auth/login-schema";
import {
  fetchBackend,
  isSameOriginRequest,
  readApiResponse,
  requestIdFrom,
} from "@/lib/api/server";
import type { ApiResponse } from "@/lib/api/types";
import { setSessionCookie } from "@/lib/auth/session";
import {
  loginResultSchema,
  type AuthenticatedUser,
  type LoginResult,
} from "@/lib/auth/schemas";

export async function POST(request: Request) {
  const requestId = requestIdFrom(request);

  if (!isSameOriginRequest(request)) {
    return NextResponse.json<ApiResponse<never>>(
      {
        success: false,
        message: "Cross-origin login requests are not allowed.",
        request_id: requestId,
      },
      { status: 403 },
    );
  }

  let requestBody: unknown;
  try {
    requestBody = await request.json();
  } catch {
    return NextResponse.json<ApiResponse<never>>(
      {
        success: false,
        message: "Enter a username or email and password.",
        request_id: requestId,
      },
      { status: 400 },
    );
  }

  const credentials = loginSchema.safeParse(requestBody);
  if (!credentials.success) {
    return NextResponse.json<ApiResponse<never>>(
      {
        success: false,
        message: "Enter a valid username or email and password.",
        error: credentials.error.flatten().fieldErrors,
        request_id: requestId,
      },
      { status: 400 },
    );
  }

  try {
    const backendResponse = await fetchBackend("/api/v1/auth/login", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "X-Request-ID": requestId,
        "User-Agent": request.headers.get("user-agent") ?? "NEXA-WMS-Web",
      },
      body: JSON.stringify(credentials.data),
    });
    const payload = await readApiResponse<LoginResult>(backendResponse);

    if (
      !backendResponse.ok ||
      !payload?.success ||
      payload.data === undefined
    ) {
      return NextResponse.json<ApiResponse<never>>(
        {
          success: false,
          message: payload?.message ?? "Unable to sign in.",
          error: payload?.error,
          request_id: payload?.request_id ?? requestId,
        },
        { status: backendResponse.status },
      );
    }

    const result = loginResultSchema.safeParse(payload.data);
    if (!result.success) {
      return NextResponse.json<ApiResponse<never>>(
        {
          success: false,
          message:
            "The authentication service returned an unexpected response.",
          request_id: payload.request_id ?? requestId,
        },
        { status: 502 },
      );
    }

    const response = NextResponse.json<ApiResponse<AuthenticatedUser>>({
      success: true,
      message: "Welcome back.",
      data: result.data.user,
      request_id: payload.request_id ?? requestId,
    });
    setSessionCookie(response, result.data.token, result.data.expires_at);
    return response;
  } catch {
    return NextResponse.json<ApiResponse<never>>(
      {
        success: false,
        message: "The authentication service is temporarily unavailable.",
        request_id: requestId,
      },
      { status: 503 },
    );
  }
}
