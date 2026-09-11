import type { Metadata } from "next";

import { WarehousesScreen } from "@/features/warehouses/warehouses-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Warehouses" };

export default async function WarehousesPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;

  const permissions = session.user.permissions;
  if (!hasPermission(permissions, PERMISSIONS.MASTER.READ)) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Warehouse access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for access to warehouse master data.
        </p>
      </section>
    );
  }

  return (
    <WarehousesScreen
      canWrite={hasPermission(permissions, PERMISSIONS.MASTER.WRITE)}
    />
  );
}
