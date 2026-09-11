import type { ReactNode } from "react";
import { redirect } from "next/navigation";

import { AuthProvider } from "@/components/auth/auth-provider";
import { ServiceUnavailable } from "@/components/auth/service-unavailable";
import { AppShell } from "@/components/layout/app-shell";
import { getServerSession } from "@/lib/auth/session";

export default async function WorkspaceLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await getServerSession();

  if (session.status === "unauthenticated") {
    redirect(
      session.reason === "invalid" ? "/api/auth/session-expired" : "/login",
    );
  }

  if (session.status === "unavailable") {
    return <ServiceUnavailable message={session.message} />;
  }

  return (
    <AuthProvider user={session.user}>
      <AppShell>{children}</AppShell>
    </AuthProvider>
  );
}
