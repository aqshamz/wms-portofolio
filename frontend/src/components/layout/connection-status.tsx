"use client";

import { useCallback, useEffect, useState } from "react";
import { WifiOff } from "lucide-react";

type ConnectionState = "checking" | "healthy" | "unavailable";

export function ConnectionStatus() {
  const [status, setStatus] = useState<ConnectionState>("checking");

  const checkConnection = useCallback(async () => {
    if (!navigator.onLine) {
      setStatus("unavailable");
      return;
    }

    try {
      const response = await fetch("/api/health", {
        cache: "no-store",
        signal: AbortSignal.timeout(4_000),
      });

      setStatus(response.ok ? "healthy" : "unavailable");
    } catch {
      setStatus("unavailable");
    }
  }, []);

  useEffect(() => {
    const initialCheck = window.setTimeout(checkConnection, 0);
    const interval = window.setInterval(checkConnection, 30_000);
    window.addEventListener("online", checkConnection);
    window.addEventListener("offline", checkConnection);

    return () => {
      window.clearTimeout(initialCheck);
      window.clearInterval(interval);
      window.removeEventListener("online", checkConnection);
      window.removeEventListener("offline", checkConnection);
    };
  }, [checkConnection]);

  if (status !== "unavailable") {
    return null;
  }

  return (
    <div
      role="status"
      className="flex min-h-9 items-center gap-2 rounded-lg border border-rose-200 bg-rose-50 px-3 text-xs font-semibold text-rose-800"
      title="Your session is preserved. Actions will resume when the connection returns."
    >
      <WifiOff className="size-4" />
      <span className="hidden sm:inline">Connection unavailable</span>
      <span className="sm:hidden">Offline</span>
    </div>
  );
}
