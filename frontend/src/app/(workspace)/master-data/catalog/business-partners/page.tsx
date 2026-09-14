import type { Metadata } from "next";

import { BusinessPartnersScreen } from "@/features/business-partners/business-partners-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Business partners" };

export default async function BusinessPartnersPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;

  const permissions = session.user.permissions;
  const canView = hasPermission(permissions, PERMISSIONS.MASTER.READ);

  if (!canView) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Catalog access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for access to master catalog data.
        </p>
      </section>
    );
  }

  const canWrite = hasPermission(permissions, PERMISSIONS.MASTER.WRITE);
  return <BusinessPartnersScreen canWrite={canWrite} />;
}
