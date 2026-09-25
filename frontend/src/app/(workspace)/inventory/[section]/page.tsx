import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { InventoryBalancesScreen } from "@/features/inventory/inventory-screen";
import { InventoryMovementsScreen } from "@/features/inventory/movement-screen";
import { InventorySerialStatesScreen } from "@/features/inventory/serial-state-screen";
import { InventoryLotsScreen } from "@/features/inventory/lot-screen";
import { InventorySerialsScreen } from "@/features/inventory/serial-screen";
import { InventoryHandlingUnitsScreen } from "@/features/inventory/handling-unit-screen";
import { getServerSession } from "@/lib/auth/session";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";

export const metadata: Metadata = { title: "Inventory" };
export default async function InventoryPage({
  params,
}: {
  params: Promise<{ section: string }>;
}) {
  const { section } = await params;
  if (
    ![
      "balances",
      "movements",
      "serials",
      "serial-numbers",
      "lots",
      "handling-units",
    ].includes(section)
  )
    notFound();
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  if (!hasPermission(session.user.permissions, PERMISSIONS.INVENTORY.READ))
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Inventory access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for the INVENTORY.READ permission.
        </p>
      </section>
    );
  return section === "balances" ? (
    <InventoryBalancesScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
    />
  ) : section === "movements" ? (
    <InventoryMovementsScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
    />
  ) : section === "serials" ? (
    <InventorySerialStatesScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
    />
  ) : section === "serial-numbers" ? (
    <InventorySerialsScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
    />
  ) : section === "lots" ? (
    <InventoryLotsScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
    />
  ) : (
    <InventoryHandlingUnitsScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
      canCreate={hasPermission(
        session.user.permissions,
        PERMISSIONS.INVENTORY.IDENTITY,
      )}
    />
  );
}
