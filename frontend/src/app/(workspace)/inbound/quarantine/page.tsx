import type { Metadata } from "next";
import { QuarantineScreen } from "@/features/quarantine/quarantine-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Quarantine" };
export default async function QuarantinePage() {
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
    <QuarantineScreen
      capabilities={{
        canDispose: hasPermission(
          session.user.permissions,
          PERMISSIONS.INBOUND.QUARANTINE_DISPOSE,
        ),
        timezone: session.user.preferred_timezone || "Asia/Jakarta",
      }}
    />
  );
}
