import type { Metadata } from "next";
import { InternalMovementsScreen } from "@/features/stock-control/internal-movements-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Internal movements" };

export default async function InternalMovementsPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  const canRead = hasPermission(
    session.user.permissions,
    PERMISSIONS.INVENTORY.READ,
  );
  const canMove = hasPermission(
    session.user.permissions,
    PERMISSIONS.INVENTORY.MOVE,
  );
  if (!canRead || !canMove) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">
          Internal movement access is required
        </h1>
        <p className="mt-2 text-sm">
          Ask an administrator for INVENTORY.READ and INVENTORY.MOVE.
        </p>
      </section>
    );
  }
  return (
    <InternalMovementsScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
    />
  );
}
