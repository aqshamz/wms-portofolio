import type { Metadata } from "next";
import { CycleCountsScreen } from "@/features/stock-control/cycle-counts-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Cycle counts" };
export default async function CycleCountsPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  const permissions = session.user.permissions;
  const allowed =
    hasPermission(permissions, PERMISSIONS.INVENTORY.READ) &&
    (hasPermission(permissions, PERMISSIONS.INVENTORY.COUNT) ||
      hasPermission(permissions, PERMISSIONS.INVENTORY.COUNT_APPROVE));
  if (!allowed)
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        Cycle count access requires INVENTORY.READ plus INVENTORY.COUNT or
        INVENTORY.COUNT_APPROVE.
      </section>
    );
  return (
    <CycleCountsScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
    />
  );
}
