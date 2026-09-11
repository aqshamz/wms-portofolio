import { NextResponse } from "next/server";

import { fetchBackend, readApiResponse } from "@/lib/api/server";
import type { ApiResponse } from "@/lib/api/types";

export async function GET() {
  try {
    const backendResponse = await fetchBackend("/health");
    const payload = await readApiResponse<unknown>(backendResponse);

    return NextResponse.json<ApiResponse<unknown>>(
      payload ?? {
        success: false,
        message: "The API returned an unexpected health response.",
      },
      { status: backendResponse.status },
    );
  } catch {
    return NextResponse.json<ApiResponse<never>>(
      { success: false, message: "The API is temporarily unavailable." },
      { status: 503 },
    );
  }
}
