import type { Metadata } from "next";
import { InventoryStatusChangesScreen } from "@/features/stock-control/inventory-status-changes-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Inventory status changes" };

export default async function InventoryStatusChangesPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  if (
    !hasPermission(session.user.permissions, PERMISSIONS.INVENTORY.READ) ||
    !hasPermission(
      session.user.permissions,
      PERMISSIONS.INVENTORY.STATUS_CHANGE,
    )
  ) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">
          Inventory status-change access is required
        </h1>
        <p className="mt-2 text-sm">
          Ask an administrator for INVENTORY.READ and INVENTORY.STATUS_CHANGE.
        </p>
      </section>
    );
  }
  return (
    <InventoryStatusChangesScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
    />
  );
}
