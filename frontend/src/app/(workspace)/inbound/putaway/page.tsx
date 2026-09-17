import type { Metadata } from "next";
import { PutawayScreen } from "@/features/putaway/putaway-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Putaway tasks" };
export default async function PutawayPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  const permissions = session.user.permissions;
  if (!hasPermission(permissions, PERMISSIONS.INBOUND.READ))
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Inbound access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for the INBOUND.READ permission.
        </p>
      </section>
    );
  return (
    <PutawayScreen
      capabilities={{
        accountId: session.user.account_id,
        timezone: session.user.preferred_timezone || "Asia/Jakarta",
        canPutaway: hasPermission(permissions, PERMISSIONS.INBOUND.PUTAWAY),
        canAssign: hasPermission(permissions, PERMISSIONS.INBOUND.ASSIGN),
        canCancel: hasPermission(permissions, PERMISSIONS.INBOUND.CANCEL),
      }}
    />
  );
}
