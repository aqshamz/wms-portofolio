import type { Metadata } from "next";

import { AccessScopesScreen } from "@/features/access-scopes/access-scopes-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Access scopes" };

export default async function AccessScopesPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;

  const permissions = session.user.permissions;
  if (!hasPermission(permissions, PERMISSIONS.MASTER.READ)) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Master data access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for permission to view master data.
        </p>
      </section>
    );
  }

  if (!hasPermission(permissions, PERMISSIONS.SECURITY.READ)) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">
          Account directory access is required
        </h1>
        <p className="mt-2 text-sm">
          Access scopes need Security read permission so an account can be
          selected safely.
        </p>
      </section>
    );
  }

  return (
    <AccessScopesScreen
      canWrite={hasPermission(permissions, PERMISSIONS.MASTER.WRITE)}
    />
  );
}
