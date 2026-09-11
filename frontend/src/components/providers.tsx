"use client";

import { useEffect, useState, type ReactNode } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { NuqsAdapter } from "nuqs/adapters/next/app";
import { Toaster } from "sonner";
import { useRouter } from "next/navigation";

import { SESSION_EXPIRED_EVENT } from "@/lib/api/client";
import { createQueryClient } from "@/lib/query-client";

export function Providers({ children }: { children: ReactNode }) {
  const [queryClient] = useState(createQueryClient);
  const router = useRouter();

  useEffect(() => {
    function handleSessionExpired() {
      router.replace("/login?reason=session-expired");
      router.refresh();
    }

    window.addEventListener(SESSION_EXPIRED_EVENT, handleSessionExpired);
    return () =>
      window.removeEventListener(SESSION_EXPIRED_EVENT, handleSessionExpired);
  }, [router]);

  return (
    <NuqsAdapter>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      <Toaster richColors closeButton position="top-right" />
    </NuqsAdapter>
  );
}
