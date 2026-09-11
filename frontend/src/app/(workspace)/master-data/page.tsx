import type { Metadata } from "next";

import { MasterDataOverview } from "@/features/master-data/master-data-overview";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Master data" };

export default async function MasterDataPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;

  const permissions = session.user.permissions;
  const canView = hasPermission(permissions, PERMISSIONS.MASTER.READ);

  if (!canView) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <p className="text-sm font-bold tracking-[0.14em] uppercase">
          Access restricted
        </p>
        <h1 className="mt-2 text-2xl font-bold">
          Master data is not assigned to your account
        </h1>
        <p className="mt-2 text-sm leading-6">
          Ask an administrator for master-data access if you need to maintain
          warehouse configuration.
        </p>
      </section>
    );
  }

  const canWrite = hasPermission(permissions, PERMISSIONS.MASTER.WRITE);

  return <MasterDataOverview canWrite={canWrite} />;
}
