import type { Metadata } from "next";
import { TransferReceiptsScreen } from "@/features/stock-control/transfer-receipts-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Transfer receipts" };
export default async function TransferReceiptsPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  if (!hasPermission(session.user.permissions, PERMISSIONS.INBOUND.READ))
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6">
        <h1 className="text-xl font-bold">Inbound access is required</h1>
        <p className="mt-2 text-sm">Ask an administrator for INBOUND.READ.</p>
      </section>
    );
  const canReceive = hasPermission(
    session.user.permissions,
    PERMISSIONS.INBOUND.RECEIVE,
  );
  const canPutaway = hasPermission(
    session.user.permissions,
    PERMISSIONS.INBOUND.PUTAWAY,
  );
  if (!canReceive && !canPutaway)
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6">
        <h1 className="text-xl font-bold">
          Transfer receipt access is required
        </h1>
        <p className="mt-2 text-sm">
          Ask an administrator for INBOUND.RECEIVE or INBOUND.PUTAWAY.
        </p>
      </section>
    );
  return (
    <TransferReceiptsScreen
      timezone={session.user.preferred_timezone || "Asia/Jakarta"}
      canReceive={canReceive}
      canPutaway={canPutaway}
    />
  );
}
