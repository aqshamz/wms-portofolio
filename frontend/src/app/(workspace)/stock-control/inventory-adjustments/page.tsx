import type { Metadata } from "next";
import { InventoryAdjustmentsScreen } from "@/features/stock-control/inventory-adjustments-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Inventory adjustments" };

export default async function InventoryAdjustmentsPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  const permissions = session.user.permissions;
  const canRead = hasPermission(permissions, PERMISSIONS.INVENTORY.READ);
  const canParticipate =
    hasPermission(permissions, PERMISSIONS.INVENTORY.ADJUST) ||
    hasPermission(permissions, PERMISSIONS.INVENTORY.ADJUST_APPROVE);
  if (!canRead || !canParticipate) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">
          Inventory adjustment access is required
        </h1>
        <p className="mt-2 text-sm">
          Ask an administrator for INVENTORY.READ plus INVENTORY.ADJUST or
          INVENTORY.ADJUST_APPROVE.
        </p>
      </section>
    );
  }
  return (
    <InventoryAdjustmentsScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
    />
  );
}
