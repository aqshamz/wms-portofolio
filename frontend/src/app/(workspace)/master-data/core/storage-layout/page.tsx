import type { Metadata } from "next";

import { StorageLayoutScreen } from "@/features/storage-layout/storage-layout-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Storage layout" };

export default async function StorageLayoutPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;

  const permissions = session.user.permissions;
  if (!hasPermission(permissions, PERMISSIONS.MASTER.READ)) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Storage-layout access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for access to warehouse master data.
        </p>
      </section>
    );
  }

  return (
    <StorageLayoutScreen
      canWrite={hasPermission(permissions, PERMISSIONS.MASTER.WRITE)}
    />
  );
}
