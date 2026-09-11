import type { Metadata } from "next";

import { ReportsOverview } from "@/features/reports/reports-overview";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = {
  title: "Reports",
};

export default async function ReportsPage() {
  const session = await getServerSession();

  if (session.status !== "authenticated") {
    return null;
  }

  const canViewReports =
    session.user.permissions.includes("*") ||
    session.user.permissions.includes("REPORTING.READ") ||
    session.user.permissions.some((permission) =>
      permission.startsWith("REPORT."),
    );

  if (!canViewReports) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <p className="text-sm font-bold tracking-[0.14em] uppercase">
          Access restricted
        </p>
        <h1 className="mt-2 text-2xl font-bold">
          Reports are not assigned to your account
        </h1>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-amber-900">
          Ask an administrator for the specific report permission you need. The
          API will also reject report data requests without that permission.
        </p>
      </section>
    );
  }

  return <ReportsOverview permissions={session.user.permissions} />;
}
