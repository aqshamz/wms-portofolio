import { NextResponse } from "next/server";

import { clearSessionCookie } from "@/lib/auth/session";

export function GET(request: Request) {
  const loginUrl = new URL("/login", request.url);
  loginUrl.searchParams.set("reason", "session-expired");

  const response = NextResponse.redirect(loginUrl);
  clearSessionCookie(response);
  return response;
}
