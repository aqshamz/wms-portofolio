import type { Metadata } from "next";

import { UnitsPackagingScreen } from "@/features/units-packaging/units-packaging-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Units and packaging" };

export default async function UnitsPackagingPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;

  const permissions = session.user.permissions;
  if (!hasPermission(permissions, PERMISSIONS.MASTER.READ)) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Packaging access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for access to master catalog data.
        </p>
      </section>
    );
  }

  return (
    <UnitsPackagingScreen
      canWrite={hasPermission(permissions, PERMISSIONS.MASTER.WRITE)}
    />
  );
}
