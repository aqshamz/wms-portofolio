"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import {
  ChevronLeft,
  ChevronRight,
  ClipboardCheck,
  Eye,
  LoaderCircle,
  Plus,
  Search,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  listWarehouseOwners,
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import { inspectionKeys, listInspections } from "./quality-inspection-api";
import { inspectionState, inspectionTone } from "./quality-inspection-types";
import { InspectionCreateDialog } from "./inspection-create-dialog";
import { InspectionDetailDialog } from "./inspection-detail-dialog";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
const statuses = [
  { value: "all", label: "All statuses" },
  { value: "PENDING", label: "Pending" },
  { value: "PASSED", label: "Accepted" },
  { value: "FAILED", label: "Failed / partially accepted" },
  { value: "WAIVED", label: "Cancelled" },
];

export function QualityInspectionsScreen({
  canQC,
  canCancel,
}: {
  canQC: boolean;
  canCancel: boolean;
}) {
  const [filters, setFilters] = useQueryStates({
    owner: parseAsString.withDefault(""),
    warehouse: parseAsString.withDefault(""),
    status: parseAsString.withDefault("all"),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
    inspection: parseAsString.withDefault(""),
  });
  const [searchDraft, setSearchDraft] = useState(filters.search);
  const [creating, setCreating] = useState(false);
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(activeWarehouses),
    queryFn: () => listWarehouses(activeWarehouses),
  });
  const warehouseOwners = useQuery({
    queryKey: warehouseKeys.owners(filters.warehouse || "none"),
    queryFn: () => listWarehouseOwners(filters.warehouse),
    enabled: Boolean(filters.warehouse),
  });
  const owners = useMemo(
    () => (warehouseOwners.data ?? []).filter((owner) => owner.is_active),
    [warehouseOwners.data],
  );
  const owner = owners.find((row) => row.owner_id === filters.owner);
  const warehouse = warehouses.data?.items.find(
    (row) => row.warehouse_id === filters.warehouse,
  );
  useEffect(() => {
    if (!filters.warehouse && warehouses.data?.items.length === 1)
      void setFilters({
        warehouse: warehouses.data.items[0].warehouse_id,
        page: 1,
      });
  }, [filters.warehouse, warehouses.data?.items, setFilters]);
  useEffect(() => {
    if (!filters.owner && owners.length === 1)
      void setFilters({ owner: owners[0].owner_id, page: 1 });
  }, [filters.owner, owners, setFilters]);
  useEffect(() => {
    if (warehouseOwners.isSuccess && filters.owner && !owner)
      void setFilters({ owner: "", page: 1 });
  }, [warehouseOwners.isSuccess, filters.owner, owner, setFilters]);
  const queryFilters = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    status: filters.status === "all" ? "" : filters.status,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const inspections = useQuery({
    queryKey: inspectionKeys.list(queryFilters),
    queryFn: () => listInspections(queryFilters),
    enabled: Boolean(owner && warehouse),
  });
  const rows = inspections.data?.items ?? [];
  const page = inspections.data?.page ?? queryFilters.page;
  const totalPages = inspections.data?.total_pages ?? 0;
  const error = warehouses.error ?? warehouseOwners.error ?? inspections.error;
  const select = (id: string) => void setFilters({ inspection: id });

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-slate-600">
            Inbound operations
          </p>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <ClipboardCheck className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Quality inspections
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Inspect accepted receipt batches before putaway or quarantine.
              </p>
            </div>
          </div>
        </div>
        {canQC ? (
          <Button
            className="w-full sm:w-auto"
            disabled={!owner || !warehouse}
            onClick={() => setCreating(true)}
          >
            <Plus className="size-4" />
            Start inspection
          </Button>
        ) : null}
      </header>
      <section className="rounded-2xl border border-slate-200 bg-white p-4 sm:p-5">
        <div className="flex flex-wrap gap-2 text-xs font-semibold">
          <span className="rounded-full bg-cyan-50 px-3 py-1.5 text-cyan-900">
            Receipt completed · QC pending
          </span>
          <span className="px-1 py-1.5 text-slate-400">→</span>
          <span className="rounded-full bg-emerald-50 px-3 py-1.5 text-emerald-900">
            Passed → Putaway task
          </span>
          <span className="rounded-full bg-rose-50 px-3 py-1.5 text-rose-900">
            Failed → Quarantine case
          </span>
        </div>
        <p className="mt-3 text-sm text-slate-600">
          One inspection per receipt batch, with quantities in base units.
          Completed results are immutable. Cancelling a pending inspection
          creates a replacement; it never skips QC.
        </p>
        {!canQC ? (
          <p className="mt-3 rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
            This account cannot start or complete inspections. INBOUND.QC is
            required; INBOUND.CANCEL separately controls cancellation.
          </p>
        ) : null}
      </section>
      <Panel className="overflow-hidden">
        <form
          className="grid gap-3 border-b border-slate-200 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-[1fr_1fr_1fr_2fr_auto]"
          onSubmit={(event) => {
            event.preventDefault();
            void setFilters({ search: searchDraft.trim(), page: 1 });
          }}
        >
          <Select
            ariaLabel="Filter by warehouse"
            value={filters.warehouse}
            options={(warehouses.data?.items ?? []).map((row) => ({
              value: row.warehouse_id,
              label: `${row.name} (${row.code})`,
            }))}
            placeholder={
              warehouses.isPending ? "Loading warehouses…" : "Select warehouse"
            }
            onValueChange={(id) => {
              setCreating(false);
              void setFilters({
                warehouse: id,
                owner: "",
                inspection: "",
                page: 1,
              });
            }}
          />
          <Select
            ariaLabel="Filter by owner"
            value={filters.owner}
            options={owners.map((row) => ({
              value: row.owner_id,
              label: `${row.owner_name} (${row.owner_code})`,
            }))}
            disabled={!warehouse || warehouseOwners.isPending}
            placeholder="Select served owner"
            onValueChange={(id) => {
              setCreating(false);
              void setFilters({ owner: id, inspection: "", page: 1 });
            }}
          />
          <Select
            ariaLabel="Filter inspection status"
            value={filters.status}
            options={statuses}
            onValueChange={(status) => void setFilters({ status, page: 1 })}
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search inspections"
              placeholder="Inspection, receipt, item or lot"
              maxLength={160}
              value={searchDraft}
              onChange={(event) => setSearchDraft(event.target.value)}
            />
          </div>
          <Button type="submit" variant="secondary">
            Search
          </Button>
        </form>
        {error ? (
          <p
            role="alert"
            className="m-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {error.message}
          </p>
        ) : !owner || !warehouse ? (
          <div className="grid min-h-56 place-items-center p-6 text-center">
            <div>
              <ClipboardCheck className="mx-auto size-8 text-slate-400" />
              <p className="mt-3 font-semibold">
                Select a warehouse and served owner
              </p>
              <p className="mt-1 text-sm text-slate-600">
                Inspections are limited to your configured access scope.
              </p>
            </div>
          </div>
        ) : inspections.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" />
            Loading inspections…
          </div>
        ) : !rows.length ? (
          <div className="grid min-h-56 place-items-center p-6 text-center">
            <div>
              <p className="font-semibold">No quality inspections found</p>
              <p className="mt-1 text-sm text-slate-600">
                Complete a receipt, then start an inspection for its accepted
                batches.
              </p>
            </div>
          </div>
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="min-w-full text-left text-sm">
                <thead className="bg-slate-50 text-xs font-bold text-slate-500 uppercase">
                  <tr>
                    {[
                      "Inspection / receipt",
                      "Item / lot",
                      "Received location",
                      "Base quantity",
                      "Result",
                      "Action",
                    ].map((heading) => (
                      <th key={heading} className="px-5 py-3">
                        {heading}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {rows.map((row) => (
                    <tr
                      key={row.inspection_id}
                      className="border-t border-slate-100 hover:bg-slate-50"
                    >
                      <td className="px-5 py-4">
                        <p className="font-semibold text-slate-950">
                          {row.inspection_id}
                        </p>
                        <p className="mt-1 text-xs text-slate-500">
                          {row.receipt_id}
                          {row.parent_inspection_id ? " · Reinspection" : ""}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <p>{row.item_code}</p>
                        <p className="mt-1 text-xs text-slate-500">
                          {row.lot_number || "No lot"}
                        </p>
                      </td>
                      <td className="px-5 py-4">{row.location_code}</td>
                      <td className="px-5 py-4">
                        <p>{row.inspected_qty}</p>
                        {row.inspected_at ? (
                          <p className="mt-1 text-xs text-slate-500">
                            Pass {row.passed_qty} / fail {row.failed_qty}
                          </p>
                        ) : null}
                      </td>
                      <td className="px-5 py-4">
                        <StatusBadge tone={inspectionTone(row)}>
                          {inspectionState(row)}
                        </StatusBadge>
                      </td>
                      <td className="px-5 py-4">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => select(row.inspection_id)}
                        >
                          <Eye className="size-4" />
                          View
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <div className="divide-y divide-slate-100 md:hidden">
              {rows.map((row) => (
                <article key={row.inspection_id} className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="font-semibold break-all text-slate-950">
                        {row.inspection_id}
                      </p>
                      <p className="mt-1 text-xs break-all text-slate-500">
                        {row.receipt_id}
                      </p>
                    </div>
                    <StatusBadge tone={inspectionTone(row)}>
                      {inspectionState(row)}
                    </StatusBadge>
                  </div>
                  <dl className="mt-3 grid grid-cols-2 gap-3 text-sm">
                    <div>
                      <dt className="text-xs text-slate-500">Item / lot</dt>
                      <dd>
                        {row.item_code}
                        <span className="block text-xs text-slate-500">
                          {row.lot_number}
                        </span>
                      </dd>
                    </div>
                    <div>
                      <dt className="text-xs text-slate-500">
                        Base quantity / location
                      </dt>
                      <dd>
                        {row.inspected_qty} · {row.location_code}
                      </dd>
                    </div>
                  </dl>
                  <Button
                    className="mt-4 w-full"
                    variant="secondary"
                    onClick={() => select(row.inspection_id)}
                  >
                    <Eye className="size-4" />
                    View inspection
                  </Button>
                </article>
              ))}
            </div>
          </>
        )}
        {owner &&
        warehouse &&
        inspections.data &&
        inspections.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {inspections.data.total_items} inspections · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || inspections.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" />
                Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || inspections.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>
      {creating && owner && warehouse ? (
        <InspectionCreateDialog
          ownerId={owner.owner_id}
          warehouseId={warehouse.warehouse_id}
          scopeLabel={`${owner.owner_name} · ${warehouse.name}`}
          onOpenChange={setCreating}
          onCreated={(id) => {
            setCreating(false);
            select(id);
          }}
        />
      ) : null}
      {filters.inspection ? (
        <InspectionDetailDialog
          key={filters.inspection}
          inspectionId={filters.inspection}
          canQC={canQC}
          canCancel={canCancel}
          onSelect={select}
          onOpenChange={(open) => {
            if (!open) void setFilters({ inspection: "" });
          }}
        />
      ) : null}
    </div>
  );
}
