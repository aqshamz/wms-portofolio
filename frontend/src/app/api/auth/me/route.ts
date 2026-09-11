import { NextResponse } from "next/server";

import type { ApiResponse } from "@/lib/api/types";
import { clearSessionCookie, getServerSession } from "@/lib/auth/session";
import type { AuthenticatedUser } from "@/lib/auth/schemas";

export async function GET() {
  const session = await getServerSession();

  if (session.status === "authenticated") {
    return NextResponse.json<ApiResponse<AuthenticatedUser>>({
      success: true,
      message: "Authenticated user.",
      data: session.user,
    });
  }

  const response = NextResponse.json<ApiResponse<never>>(
    {
      success: false,
      message:
        session.status === "unavailable"
          ? session.message
          : "Authentication required.",
    },
    { status: session.status === "unavailable" ? 503 : 401 },
  );

  if (session.status === "unauthenticated") {
    clearSessionCookie(response);
  }

  return response;
}
