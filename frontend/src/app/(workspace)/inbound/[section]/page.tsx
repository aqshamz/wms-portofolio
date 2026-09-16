import type { Metadata } from "next";
import {
  AlertTriangle,
  ClipboardCheck,
  ClipboardList,
  PackageCheck,
  RotateCcw,
  ScanSearch,
  Truck,
} from "lucide-react";
import { notFound } from "next/navigation";

import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

const sections = {
  orders: {
    title: "Inbound orders",
    description: "Schedule approved Purchase Order lines for receiving.",
    endpoint: "/api/v1/inbound/orders",
    icon: ClipboardList,
  },
  receipts: {
    title: "Receipts",
    description: "Record arrivals, quantities, lots, and receiving locations.",
    endpoint: "/api/v1/inbound/receipts",
    icon: Truck,
  },
  "quality-inspections": {
    title: "Quality inspections",
    description: "Inspect received inventory before it becomes available.",
    endpoint: "/api/v1/inbound/quality-inspections",
    icon: ClipboardCheck,
  },
  putaway: {
    title: "Putaway tasks",
    description: "Move accepted inventory from receiving into storage.",
    endpoint: "/api/v1/inbound/putaway-tasks",
    icon: PackageCheck,
  },
  quarantine: {
    title: "Quarantine",
    description: "Review held inventory and record disposition decisions.",
    endpoint: "/api/v1/inbound/quarantine-cases",
    icon: AlertTriangle,
  },
  exceptions: {
    title: "Inbound exceptions",
    description: "Investigate operational exceptions across inbound work.",
    endpoint: "/api/v1/inbound/exceptions",
    icon: ScanSearch,
  },
  rework: {
    title: "Rework tasks",
    description: "Execute and complete quarantine rework assignments.",
    endpoint: "/api/v1/inbound/rework-tasks",
    icon: RotateCcw,
  },
} as const;

type InboundSection = keyof typeof sections;

function isInboundSection(value: string): value is InboundSection {
  return value in sections;
}

export async function generateMetadata({
  params,
}: PageProps<"/inbound/[section]">): Promise<Metadata> {
  const { section } = await params;
  return {
    title: isInboundSection(section) ? sections[section].title : "Inbound",
  };
}

export default async function InboundSectionPage({
  params,
}: PageProps<"/inbound/[section]">) {
  const { section } = await params;
  if (!isInboundSection(section)) notFound();

  const session = await getServerSession();
  if (session.status !== "authenticated") return null;
  if (!hasPermission(session.user.permissions, PERMISSIONS.INBOUND.READ)) {
    return (
      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-6 text-amber-950">
        <h1 className="text-xl font-bold">Inbound access is required</h1>
        <p className="mt-2 text-sm">
          Ask an administrator for the INBOUND.READ permission.
        </p>
      </section>
    );
  }

  const definition = sections[section];
  const Icon = definition.icon;

  return (
    <div className="space-y-6">
      <header>
        <p className="text-sm font-semibold text-slate-600">
          Inbound operations
        </p>
        <div className="mt-3 flex items-center gap-3">
          <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <Icon className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              {definition.title}
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              {definition.description}
            </p>
          </div>
        </div>
      </header>

      <section className="rounded-2xl border border-cyan-200 bg-cyan-50 p-6 text-cyan-950">
        <p className="text-xs font-bold tracking-wide uppercase">
          Planned inbound workspace
        </p>
        <h2 className="mt-2 text-lg font-bold">
          Purchase Orders are being completed first
        </h2>
        <p className="mt-2 max-w-2xl text-sm leading-6">
          This navigation destination is ready for the next implementation
          phase. Its API collection starts at{" "}
          <code className="rounded bg-white/70 px-1.5 py-0.5 font-mono text-xs">
            {definition.endpoint}
          </code>
          .
        </p>
      </section>
    </div>
  );
}
