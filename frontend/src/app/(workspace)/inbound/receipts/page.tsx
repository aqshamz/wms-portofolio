import type { Metadata } from "next";

import { ReceiptsScreen } from "@/features/receipts/receipts-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Receipts" };

export default async function ReceiptsPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;

  const permissions = session.user.permissions;
  if (!hasPermission(permissions, PERMISSIONS.INBOUND.READ)) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Inbound access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for the INBOUND.READ permission.
        </p>
      </section>
    );
  }

  return (
    <ReceiptsScreen
      canReceive={hasPermission(permissions, PERMISSIONS.INBOUND.RECEIVE)}
      canCancel={hasPermission(permissions, PERMISSIONS.INBOUND.CANCEL)}
      canReadInventory={hasPermission(permissions, PERMISSIONS.INVENTORY.READ)}
    />
  );
}
