import {
  ArrowRight,
  Boxes,
  CircleAlert,
  ClipboardList,
  PackageCheck,
  PackageOpen,
  ScanLine,
  Truck,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { StatusBadge } from "@/components/ui/status-badge";
import { StatCard } from "@/features/dashboard/stat-card";

interface WorkItem {
  reference: string;
  type: string;
  location: string;
  status: string;
  tone: "info" | "warning" | "danger" | "success";
  age: string;
}

const workQueue: WorkItem[] = [
  {
    reference: "IB-260910-0142",
    type: "Receiving",
    location: "Dock D-03",
    status: "Awaiting QC",
    tone: "warning",
    age: "18 min",
  },
  {
    reference: "WV-260910-0028",
    type: "Wave picking",
    location: "Zone A",
    status: "In progress",
    tone: "info",
    age: "34 min",
  },
  {
    reference: "SC-260910-0007",
    type: "Stock count",
    location: "B-14-02",
    status: "Variance found",
    tone: "danger",
    age: "52 min",
  },
  {
    reference: "SH-260910-0031",
    type: "Shipment",
    location: "Gate G-02",
    status: "Ready",
    tone: "success",
    age: "1 hr",
  },
];

const capacity = [
  { label: "Ambient storage", value: 78, color: "bg-cyan-500" },
  { label: "Pick faces", value: 64, color: "bg-indigo-500" },
  { label: "Receiving docks", value: 42, color: "bg-emerald-500" },
  { label: "Quarantine", value: 18, color: "bg-amber-500" },
];

export function DashboardOverview() {
  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-cyan-700">
            Operations overview
          </p>
          <h1 className="mt-1 text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
            Good morning, Andi
          </h1>
          <p className="mt-1 max-w-2xl text-sm leading-6 text-slate-500 sm:text-base">
            Prioritize exceptions and keep today&apos;s warehouse flow moving.
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="secondary" className="flex-1 sm:flex-none">
            <ScanLine className="size-4" />
            Scan item
          </Button>
          <Button className="flex-1 sm:flex-none">
            View tasks
            <ArrowRight className="size-4" />
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Inbound today"
          value="126 pallets"
          detail="12% ahead of yesterday"
          trend="up"
          icon={PackageOpen}
          tone="cyan"
        />
        <StatCard
          label="Open pick tasks"
          value="84"
          detail="18 due within one hour"
          icon={ClipboardList}
          tone="indigo"
        />
        <StatCard
          label="Outbound ready"
          value="31 loads"
          detail="6 waiting for a vehicle"
          icon={Truck}
          tone="emerald"
        />
        <StatCard
          label="Exceptions"
          value="7"
          detail="3 require manager review"
          trend="down"
          icon={CircleAlert}
          tone="amber"
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1.65fr)_minmax(20rem,0.85fr)]">
        <Panel className="overflow-hidden">
          <div className="flex items-center justify-between gap-4 border-b border-slate-100 px-4 py-4 sm:px-5">
            <div>
              <h2 className="font-bold text-slate-950">Priority work queue</h2>
              <p className="mt-0.5 text-sm text-slate-500">
                Items needing attention across the floor
              </p>
            </div>
            <Button variant="ghost" size="sm" className="hidden sm:inline-flex">
              View all
              <ArrowRight className="size-4" />
            </Button>
          </div>

          <div className="divide-y divide-slate-100">
            {workQueue.map((item) => (
              <div
                key={item.reference}
                className="grid gap-3 px-4 py-4 transition-colors hover:bg-slate-50/70 sm:grid-cols-[minmax(0,1.2fr)_minmax(8rem,0.8fr)_auto] sm:items-center sm:px-5"
              >
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="truncate text-sm font-bold text-slate-950">
                      {item.reference}
                    </p>
                    <StatusBadge tone={item.tone}>{item.status}</StatusBadge>
                  </div>
                  <p className="mt-1 text-sm text-slate-500">
                    {item.type} · {item.location}
                  </p>
                </div>
                <p className="text-sm text-slate-500 sm:text-right">
                  Waiting {item.age}
                </p>
                <Button
                  variant="secondary"
                  size="sm"
                  className="w-full sm:w-auto"
                >
                  Open
                </Button>
              </div>
            ))}
          </div>

          <div className="border-t border-slate-100 p-3 sm:hidden">
            <Button variant="ghost" size="sm" className="w-full">
              View all tasks
              <ArrowRight className="size-4" />
            </Button>
          </div>
        </Panel>

        <Panel className="p-4 sm:p-5">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h2 className="font-bold text-slate-950">Capacity snapshot</h2>
              <p className="mt-0.5 text-sm text-slate-500">
                Current space utilization
              </p>
            </div>
            <div className="grid size-10 place-items-center rounded-xl bg-slate-100 text-slate-700">
              <Boxes className="size-5" />
            </div>
          </div>

          <div className="mt-6 space-y-5">
            {capacity.map((item) => (
              <div key={item.label}>
                <div className="mb-2 flex items-center justify-between gap-4 text-sm">
                  <span className="font-medium text-slate-700">
                    {item.label}
                  </span>
                  <span className="font-bold text-slate-950">
                    {item.value}%
                  </span>
                </div>
                <div className="h-2 overflow-hidden rounded-full bg-slate-100">
                  <div
                    className={`h-full rounded-full ${item.color}`}
                    style={{ width: `${item.value}%` }}
                  />
                </div>
              </div>
            ))}
          </div>

          <div className="mt-6 rounded-xl border border-cyan-100 bg-cyan-50 p-4">
            <div className="flex gap-3">
              <PackageCheck className="mt-0.5 size-5 shrink-0 text-cyan-700" />
              <div>
                <p className="text-sm font-bold text-cyan-950">
                  Space is healthy
                </p>
                <p className="mt-1 text-sm leading-5 text-cyan-800">
                  No zone is projected to exceed 85% capacity today.
                </p>
              </div>
            </div>
          </div>
        </Panel>
      </div>
    </div>
  );
}
