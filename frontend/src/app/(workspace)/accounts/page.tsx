import type { Metadata } from "next";

import { AccountsScreen } from "@/features/accounts/accounts-screen";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Accounts" };

export default async function AccountsPage() {
  const session = await getServerSession();
  if (session.status !== "authenticated") return null;

  const permissions = session.user.permissions;
  if (!hasPermission(permissions, PERMISSIONS.SECURITY.READ)) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Account access is required</h1>
        <p className="mt-2 text-sm">
          Ask a security administrator for permission to view accounts.
        </p>
      </section>
    );
  }

  return (
    <AccountsScreen
      currentAccountId={session.user.account_id}
      canWrite={hasPermission(permissions, PERMISSIONS.SECURITY.WRITE)}
    />
  );
}
