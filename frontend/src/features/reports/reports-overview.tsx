import type { LucideIcon } from "lucide-react";
import {
  ArrowRight,
  Boxes,
  Building2,
  Calculator,
  PackageOpen,
  Truck,
} from "lucide-react";

import { Panel } from "@/components/ui/panel";

interface ReportDefinition {
  name: string;
  description: string;
}

interface ReportArea {
  id: string;
  name: string;
  description: string;
  icon: LucideIcon;
  reports: ReportDefinition[];
}

const reportAreas: ReportArea[] = [
  {
    id: "master-data",
    name: "Master data",
    description: "Warehouse structure, partners, items, and configuration.",
    icon: Building2,
    reports: [
      {
        name: "Organization and warehouse coverage",
        description: "Operational coverage and active warehouse structure.",
      },
      {
        name: "Location utilization",
        description: "Capacity and current inventory occupancy by location.",
      },
      {
        name: "Business partner directory",
        description: "Owners, vendors, customers, and service partners.",
      },
      {
        name: "Item catalog",
        description: "Categories, base UOM, packaging, and control settings.",
      },
      {
        name: "Workflow configuration",
        description: "Statuses, transitions, numbering, and strategies.",
      },
    ],
  },
  {
    id: "inbound",
    name: "Inbound",
    description: "From purchase-order planning through putaway.",
    icon: PackageOpen,
    reports: [
      {
        name: "Purchase order fulfillment",
        description: "Ordered, received, outstanding, and variance by PO line.",
      },
      {
        name: "Receipt and QC summary",
        description: "Daily receiving volume and quality outcomes.",
      },
      {
        name: "QC inspection exceptions",
        description: "Inspection details, failures, and unresolved variance.",
      },
      {
        name: "Quarantine aging",
        description: "Open cases, age, and client disposition progress.",
      },
      {
        name: "Putaway productivity",
        description: "Task completion, backlog, and processing time.",
      },
      {
        name: "Vendor inbound performance",
        description: "Receipt timeliness, shortages, and quality trends.",
      },
    ],
  },
  {
    id: "stock-control",
    name: "Stock control",
    description: "Inventory position, movement, accuracy, and transfers.",
    icon: Boxes,
    reports: [
      {
        name: "Current stock on hand",
        description: "Balance detail by item, lot, status, and location.",
      },
      {
        name: "Stock status summary",
        description: "Available, reserved, quarantine, and blocked quantities.",
      },
      {
        name: "Inventory movement ledger",
        description: "Immutable transaction history and references.",
      },
      {
        name: "Internal movement completion",
        description: "Movement volume, completion, and operator activity.",
      },
      {
        name: "Warehouse transfer progress",
        description: "Dispatch, receipt, transit time, and variance.",
      },
      {
        name: "Adjustments and status changes",
        description: "Authorized corrections with reason and audit context.",
      },
      {
        name: "Stock count variance",
        description: "Expected versus counted quantity and reconciliation.",
      },
    ],
  },
  {
    id: "outbound",
    name: "Outbound",
    description: "Delivery-order fulfillment through final delivery.",
    icon: Truck,
    reports: [
      {
        name: "Delivery order fulfillment",
        description:
          "Ordered, allocated, picked, shipped, and delivered by line.",
      },
      {
        name: "Allocation and picking exceptions",
        description: "Short allocation, short pick, and unresolved tasks.",
      },
      {
        name: "Wave productivity",
        description: "Wave progress, throughput, and task performance.",
      },
      {
        name: "Pack and shipment reconciliation",
        description: "Checked, packed, and shipped quantities by order.",
      },
      {
        name: "Store delivery performance",
        description: "Delivery timeliness and proof-of-delivery completion.",
      },
      {
        name: "Delivery failures and returns",
        description: "Failure events and returned inventory outcomes.",
      },
    ],
  },
  {
    id: "billing",
    name: "Billing",
    description: "Contract coverage, revenue, invoices, and settlement.",
    icon: Calculator,
    reports: [
      {
        name: "Contract and rate coverage",
        description: "Active commercial terms and missing rate configuration.",
      },
      {
        name: "Billable event status",
        description: "Collected, pending, rated, and excluded service events.",
      },
      {
        name: "Rated revenue by service",
        description: "Calculated charges grouped by service and period.",
      },
      {
        name: "Billing run reconciliation",
        description: "Run totals, exceptions, review, and posting status.",
      },
      {
        name: "Invoice aging",
        description: "Open receivables, due dates, and outstanding balances.",
      },
      {
        name: "Payments and allocation",
        description: "Receipts and their settlement against invoices.",
      },
      {
        name: "Credits and adjustments",
        description: "Credit notes and changes to invoiced charges.",
      },
    ],
  },
];

export function ReportsOverview() {
  return (
    <div className="space-y-8">
      <header>
        <p className="text-sm font-semibold text-cyan-700">Reporting</p>
        <h1 className="mt-1 text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
          Reports
        </h1>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-500 sm:text-base">
          Choose an operational area, then open the report needed for review or
          export.
        </p>
      </header>

      <nav
        aria-label="Report categories"
        className="flex gap-2 overflow-x-auto pb-1"
      >
        {reportAreas.map((area) => (
          <a
            key={area.id}
            href={`#${area.id}`}
            className="shrink-0 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-slate-700 shadow-sm transition-colors hover:border-cyan-300 hover:text-cyan-800 focus-visible:ring-2 focus-visible:ring-cyan-500 focus-visible:outline-none"
          >
            {area.name}
          </a>
        ))}
      </nav>

      <div className="space-y-8">
        {reportAreas.map((area) => {
          const Icon = area.icon;

          return (
            <section key={area.id} id={area.id} className="scroll-mt-24">
              <div className="mb-4 flex items-start gap-3">
                <div className="grid size-10 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
                  <Icon className="size-5" />
                </div>
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <h2 className="text-lg font-bold text-slate-950">
                      {area.name}
                    </h2>
                    <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-600">
                      {area.reports.length} reports
                    </span>
                  </div>
                  <p className="mt-0.5 text-sm text-slate-500">
                    {area.description}
                  </p>
                </div>
              </div>

              <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                {area.reports.map((report) => (
                  <Panel
                    key={report.name}
                    className="group flex min-h-36 flex-col p-4 transition-colors hover:border-cyan-300 sm:p-5"
                  >
                    <h3 className="font-bold text-slate-950">{report.name}</h3>
                    <p className="mt-2 flex-1 text-sm leading-5 text-slate-500">
                      {report.description}
                    </p>
                    <button
                      type="button"
                      className="mt-4 flex min-h-9 items-center gap-2 self-start text-sm font-semibold text-cyan-800 focus-visible:ring-2 focus-visible:ring-cyan-500 focus-visible:outline-none"
                    >
                      Open report
                      <ArrowRight className="size-4 transition-transform group-hover:translate-x-0.5" />
                    </button>
                  </Panel>
                ))}
              </div>
            </section>
          );
        })}
      </div>
    </div>
  );
}
