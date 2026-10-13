import type { Metadata } from "next";
import { OutboundOrdersScreen } from "@/features/outbound-orders/outbound-orders-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Delivery orders" };
export default async function OutboundOrdersPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  const permissions = session.user.permissions;
  if (!hasPermission(permissions, PERMISSIONS.OUTBOUND.READ))
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Outbound access is required</h1>
        <p className="mt-2 text-sm">Ask an administrator for OUTBOUND.READ.</p>
      </section>
    );
  return (
    <OutboundOrdersScreen
      canPlan={hasPermission(permissions, PERMISSIONS.OUTBOUND.PLAN)}
      canCancel={hasPermission(permissions, PERMISSIONS.OUTBOUND.CANCEL)}
      canCloseShort={hasPermission(
        permissions,
        PERMISSIONS.OUTBOUND.CLOSE_SHORT,
      )}
    />
  );
}
