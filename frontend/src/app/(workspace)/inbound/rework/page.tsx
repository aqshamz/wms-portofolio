import type { Metadata } from "next";
import { ReworkScreen } from "@/features/rework/rework-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Rework tasks" };
export default async function ReworkPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  if (!hasPermission(session.user.permissions, PERMISSIONS.INBOUND.READ))
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Inbound access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for the INBOUND.READ permission.
        </p>
      </section>
    );
  return (
    <ReworkScreen
      capabilities={{
        accountId: session.user.account_id,
        canRework: hasPermission(
          session.user.permissions,
          PERMISSIONS.INBOUND.REWORK,
        ),
        timezone: session.user.preferred_timezone || "Asia/Jakarta",
      }}
    />
  );
}
