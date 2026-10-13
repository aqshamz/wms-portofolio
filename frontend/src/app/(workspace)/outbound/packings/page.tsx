import type { Metadata } from "next";
import { FulfillmentScreen } from "@/features/outbound-fulfillment/fulfillment-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";
export const metadata: Metadata = { title: "Packing" };
export default async function PackingPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  const permissions = session.user.permissions;
  if (!hasPermission(permissions, PERMISSIONS.OUTBOUND.READ))
    return <p>Ask an administrator for OUTBOUND.READ.</p>;
  return (
    <FulfillmentScreen
      mode="packings"
      canWrite={hasPermission(permissions, PERMISSIONS.OUTBOUND.PACK)}
      canCancel={hasPermission(permissions, PERMISSIONS.OUTBOUND.CANCEL)}
      canCreateHU={hasPermission(permissions, PERMISSIONS.INVENTORY.IDENTITY)}
    />
  );
}
