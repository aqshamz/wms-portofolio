import type { Metadata } from "next";
import { VendorReturnScreen } from "@/features/vendor-returns/vendor-return-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Return to vendor" };

export default async function VendorReturnsPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  const permissions = session.user.permissions;
  if (!hasPermission(permissions, PERMISSIONS.OUTBOUND.READ))
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Outbound access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for the OUTBOUND.READ permission.
        </p>
      </section>
    );
  return (
    <VendorReturnScreen
      capabilities={{
        timezone: session.user.preferred_timezone || "Asia/Jakarta",
        canComplete: hasPermission(
          permissions,
          PERMISSIONS.OUTBOUND.RETURN_TO_VENDOR,
        ),
        canCancel: hasPermission(permissions, PERMISSIONS.OUTBOUND.CANCEL),
      }}
    />
  );
}
