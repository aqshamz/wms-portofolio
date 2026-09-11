import "server-only";

import { cache } from "react";
import { cookies } from "next/headers";
import type { NextResponse } from "next/server";

import { fetchBackend, readApiResponse } from "@/lib/api/server";
import type { ApiResponse } from "@/lib/api/types";
import {
  authenticatedUserSchema,
  type AuthenticatedUser,
} from "@/lib/auth/schemas";

export const SESSION_COOKIE = "wms_session";

export type SessionState =
  | { status: "authenticated"; user: AuthenticatedUser }
  | { status: "unauthenticated"; reason: "missing" | "invalid" }
  | { status: "unavailable"; message: string };

export async function sessionToken() {
  return (await cookies()).get(SESSION_COOKIE)?.value;
}

export function setSessionCookie(
  response: NextResponse,
  token: string,
  expiresAt: string,
) {
  response.cookies.set({
    name: SESSION_COOKIE,
    value: token,
    expires: new Date(expiresAt),
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    priority: "high",
  });
}

export function clearSessionCookie(response: NextResponse) {
  response.cookies.set({
    name: SESSION_COOKIE,
    value: "",
    expires: new Date(0),
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
  });
}

export const getServerSession = cache(async (): Promise<SessionState> => {
  const token = await sessionToken();

  if (!token) {
    return { status: "unauthenticated", reason: "missing" };
  }

  try {
    const response = await fetchBackend("/api/v1/auth/me", {
      headers: { Authorization: `Bearer ${token}` },
    });

    if (response.status === 401) {
      return { status: "unauthenticated", reason: "invalid" };
    }

    const payload = await readApiResponse<AuthenticatedUser>(response);
    if (!response.ok || !payload?.success || payload.data === undefined) {
      return {
        status: "unavailable",
        message:
          payload?.message ?? "The authentication service is unavailable.",
      };
    }

    const user = authenticatedUserSchema.safeParse(payload.data);
    if (!user.success) {
      return {
        status: "unavailable",
        message: "The authentication service returned an unexpected response.",
      };
    }

    return { status: "authenticated", user: user.data };
  } catch {
    return {
      status: "unavailable",
      message: "The authentication service is temporarily unreachable.",
    };
  }
});

export function authResponse<T>(
  success: boolean,
  message: string,
  data?: T,
): ApiResponse<T> {
  return { success, message, data };
}
