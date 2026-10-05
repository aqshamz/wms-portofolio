import type { Metadata } from "next";
import { WarehouseTransfersScreen } from "@/features/stock-control/warehouse-transfers-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Warehouse transfers" };

export default async function WarehouseTransfersPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  const canRead = hasPermission(
    session.user.permissions,
    PERMISSIONS.INVENTORY.READ,
  );
  const canTransfer = hasPermission(
    session.user.permissions,
    PERMISSIONS.INVENTORY.TRANSFER,
  );
  if (!canRead || !canTransfer) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">
          Warehouse transfer access is required
        </h1>
        <p className="mt-2 text-sm">
          Ask an administrator for INVENTORY.READ and INVENTORY.TRANSFER.
        </p>
      </section>
    );
  }
  return (
    <WarehouseTransfersScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
    />
  );
}
