import type { Metadata } from "next";

import { InventoryClassificationsScreen } from "@/features/inventory-classifications/inventory-classifications-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Inventory classifications" };

export default async function InventoryClassificationsPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;

  const permissions = session.user.permissions;
  if (!hasPermission(permissions, PERMISSIONS.MASTER.READ)) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Classification access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for access to master catalog data.
        </p>
      </section>
    );
  }

  return (
    <InventoryClassificationsScreen
      canWrite={hasPermission(permissions, PERMISSIONS.MASTER.WRITE)}
    />
  );
}
