"use client";

import { useEffect, useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  AlertTriangle,
  Barcode as BarcodeIcon,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  LoaderCircle,
  PackageOpen,
  Pencil,
  Plus,
  Scale,
  Search,
  Star,
  X,
} from "lucide-react";
import Link from "next/link";
import {
  parseAsInteger,
  parseAsString,
  parseAsStringLiteral,
  useQueryStates,
} from "nuqs";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  itemCatalogKeys,
  listItems,
} from "@/features/item-catalog/item-catalog-api";
import {
  listOrganizations,
  organizationKeys,
} from "@/features/organizations/organization-api";
import {
  deactivateHandlingUnitType,
  deactivateItemBarcode,
  deactivateItemUOM,
  deactivateUOM,
  listHandlingUnitTypes,
  listItemBarcodes,
  listItemUOMs,
  listUOMs,
  setPrimaryItemBarcode,
  unitsPackagingKeys,
} from "@/features/units-packaging/units-packaging-api";
import {
  BarcodeDialog,
  HandlingUnitDialog,
  ItemUOMDialog,
  UOMDialog,
} from "@/features/units-packaging/units-packaging-dialogs";
import type {
  HandlingUnitType,
  ItemBarcode,
  ItemUOM,
  UOM,
  UnitsPackagingView,
} from "@/features/units-packaging/units-packaging-types";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 10;
const views = ["uoms", "handling-units", "item-uoms", "barcodes"] as const;
const activeValues = ["all", "active", "inactive"] as const;
const statusOptions = [
  { value: "all", label: "All statuses" },
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
] as const;
const ownerFilters = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
const uomLookupFilters = {
  search: "",
  active: "all" as const,
  page: 1,
  pageSize: 100,
};

const viewLabels: Record<UnitsPackagingView, string> = {
  uoms: "Units of measure",
  "handling-units": "Handling units",
  "item-uoms": "Item UOMs",
  barcodes: "Barcodes",
};

type Editor =
  | { kind: "uom"; item?: UOM }
  | { kind: "handling"; item?: HandlingUnitType }
  | { kind: "item-uom"; item?: ItemUOM }
  | { kind: "barcode"; item?: ItemBarcode }
  | null;
type DeactivateTarget =
  | { kind: "uom"; item: UOM }
  | { kind: "handling"; item: HandlingUnitType }
  | { kind: "item-uom"; item: ItemUOM }
  | { kind: "barcode"; item: ItemBarcode };

function LoadingState() {
  return (
    <div className="space-y-3 p-5">
      {Array.from({ length: 5 }, (_, index) => (
        <div
          key={index}
          className="h-24 animate-pulse rounded-xl bg-slate-100"
        />
      ))}
    </div>
  );
}

function ErrorState({ error, retry }: { error: Error; retry: () => void }) {
  return (
    <div className="p-5">
      <div
        role="alert"
        className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900"
      >
        <div className="flex gap-3">
          <CircleAlert className="mt-0.5 size-5 shrink-0" />
          <div>
            <p className="font-semibold">Packaging data could not be loaded</p>
            <p className="mt-1 text-sm">{error.message}</p>
          </div>
        </div>
        <Button variant="secondary" className="mt-4" onClick={retry}>
          Try again
        </Button>
      </div>
    </div>
  );
}

function EmptyState({ view }: { view: UnitsPackagingView }) {
  return (
    <div className="px-5 py-14 text-center">
      <PackageOpen className="mx-auto size-9 text-slate-300" />
      <h2 className="mt-4 font-bold text-slate-950">
        No {viewLabels[view].toLowerCase()} found
      </h2>
      <p className="mt-1 text-sm text-slate-500">
        Adjust the filters or create the first record.
      </p>
    </div>
  );
}

function Pagination({
  page,
  totalPages,
  onPage,
}: {
  page: number;
  totalPages: number;
  onPage: (page: number) => void;
}) {
  if (totalPages <= 0) return null;
  return (
    <div className="flex flex-col gap-3 border-t border-slate-200 px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5">
      <p className="text-sm text-slate-600">
        Page <span className="font-semibold text-slate-900">{page}</span> of{" "}
        {totalPages}
      </p>
      <div className="flex gap-2">
        <Button
          variant="secondary"
          size="sm"
          disabled={page <= 1}
          onClick={() => onPage(page - 1)}
        >
          <ChevronLeft className="size-4" />
          Previous
        </Button>
        <Button
          variant="secondary"
          size="sm"
          disabled={page >= totalPages}
          onClick={() => onPage(page + 1)}
        >
          Next
          <ChevronRight className="size-4" />
        </Button>
      </div>
    </div>
  );
}

function DeactivateDialog({
  target,
  pending,
  error,
  onConfirm,
  onOpenChange,
}: {
  target?: DeactivateTarget;
  pending: boolean;
  error?: Error | null;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
}) {
  const label = !target
    ? "record"
    : target.kind === "barcode"
      ? target.item.barcode
      : target.kind === "item-uom"
        ? "item UOM"
        : target.item.name;
  const message =
    target?.kind === "item-uom"
      ? "Barcodes using this UOM must be deactivated first. The base UOM cannot be deactivated."
      : target?.kind === "barcode"
        ? "The barcode remains visible in history and loses primary status."
        : "The record remains available for historical references but cannot be selected for new configuration.";
  return (
    <Dialog.Root
      open={target !== undefined}
      onOpenChange={(open) => {
        if (!pending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content
          role="alertdialog"
          className="fixed top-1/2 left-1/2 z-50 w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-white p-5 shadow-2xl focus:outline-none sm:p-6"
        >
          <div className="flex items-start justify-between">
            <div className="grid size-11 place-items-center rounded-xl bg-amber-100 text-amber-700">
              <AlertTriangle className="size-5" />
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close confirmation"
                className="grid size-9 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-4" />
              </button>
            </Dialog.Close>
          </div>
          <Dialog.Title className="mt-5 text-lg font-bold text-slate-950">
            Deactivate {label}?
          </Dialog.Title>
          <Dialog.Description className="mt-2 text-sm leading-6 text-slate-600">
            {message}
          </Dialog.Description>
          {error ? (
            <p
              role="alert"
              className="mt-4 rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
            >
              {error.message}
            </p>
          ) : null}
          <div className="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <Dialog.Close asChild>
              <Button type="button" variant="secondary">
                Keep active
              </Button>
            </Dialog.Close>
            <Button
              type="button"
              className="bg-rose-700 hover:bg-rose-800"
              disabled={pending}
              onClick={onConfirm}
            >
              {pending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              {pending ? "Deactivating…" : "Deactivate"}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function UnitsPackagingScreen({ canWrite }: { canWrite: boolean }) {
  const [filters, setFilters] = useQueryStates({
    view: parseAsStringLiteral(views).withDefault("uoms"),
    owner: parseAsString.withDefault(""),
    item: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    active: parseAsStringLiteral(activeValues).withDefault("all"),
    page: parseAsInteger.withDefault(1),
  });
  const [searchDraft, setSearchDraft] = useState(filters.search);
  const [editor, setEditor] = useState<Editor>(null);
  const [deactivateTarget, setDeactivateTarget] = useState<DeactivateTarget>();
  const queryClient = useQueryClient();
  const itemScoped =
    filters.view === "item-uoms" || filters.view === "barcodes";
  const referenceFilters = {
    search: filters.search,
    active: filters.active,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };
  const uomList = useQuery({
    queryKey: unitsPackagingKeys.uomList(referenceFilters),
    queryFn: () => listUOMs(referenceFilters),
    enabled: filters.view === "uoms",
    placeholderData: keepPreviousData,
  });
  const handlingList = useQuery({
    queryKey: unitsPackagingKeys.handlingList(referenceFilters),
    queryFn: () => listHandlingUnitTypes(referenceFilters),
    enabled: filters.view === "handling-units",
    placeholderData: keepPreviousData,
  });
  const uomLookup = useQuery({
    queryKey: unitsPackagingKeys.uomList(uomLookupFilters),
    queryFn: () => listUOMs(uomLookupFilters),
    enabled: itemScoped,
  });
  const owners = useQuery({
    queryKey: organizationKeys.list(ownerFilters),
    queryFn: () => listOrganizations(ownerFilters),
    enabled: itemScoped,
  });
  const itemLookupFilters = {
    ownerId: filters.owner,
    categoryId: "",
    search: "",
    active: "all" as const,
    page: 1,
    pageSize: 100,
  };
  const itemLookup = useQuery({
    queryKey: itemCatalogKeys.itemList(itemLookupFilters),
    queryFn: () => listItems(itemLookupFilters),
    enabled: itemScoped && Boolean(filters.owner),
  });
  const itemUOMs = useQuery({
    queryKey: unitsPackagingKeys.itemUOMs(filters.item),
    queryFn: () => listItemUOMs(filters.item),
    enabled: itemScoped && Boolean(filters.item),
  });
  const barcodes = useQuery({
    queryKey: unitsPackagingKeys.barcodes(filters.item),
    queryFn: () => listItemBarcodes(filters.item),
    enabled: filters.view === "barcodes" && Boolean(filters.item),
  });

  const ownerItems = owners.data?.items ?? [];
  const firstOwnerId = ownerItems[0]?.organization_id;
  useEffect(() => {
    if (itemScoped && !filters.owner && firstOwnerId)
      void setFilters({ owner: firstOwnerId, item: "" });
  }, [filters.owner, firstOwnerId, itemScoped, setFilters]);
  const items = itemLookup.data?.items ?? [];
  const firstItemId = items[0]?.item_id;
  useEffect(() => {
    if (itemScoped && filters.owner && !filters.item && firstItemId)
      void setFilters({ item: firstItemId });
  }, [filters.item, filters.owner, firstItemId, itemScoped, setFilters]);

  const currentItem = items.find((item) => item.item_id === filters.item);
  const allUOMs = uomLookup.data?.items ?? [];
  const uomById = new Map(allUOMs.map((uom) => [uom.uom_id, uom]));
  const visibleItemUOMs = (itemUOMs.data ?? []).filter((unit) => {
    const matchesStatus =
      filters.active === "all" ||
      unit.is_active === (filters.active === "active");
    const uom = uomById.get(unit.uom_id);
    const needle = filters.search.toLowerCase();
    return (
      matchesStatus &&
      (!needle ||
        unit.conversion_to_base.toLowerCase().includes(needle) ||
        uom?.code.toLowerCase().includes(needle) ||
        uom?.name.toLowerCase().includes(needle))
    );
  });
  const visibleBarcodes = (barcodes.data ?? []).filter((barcode) => {
    const matchesStatus =
      filters.active === "all" ||
      barcode.is_active === (filters.active === "active");
    const uom = barcode.uom_id ? uomById.get(barcode.uom_id) : undefined;
    const needle = filters.search.toLowerCase();
    return (
      matchesStatus &&
      (!needle ||
        barcode.barcode.toLowerCase().includes(needle) ||
        uom?.code.toLowerCase().includes(needle))
    );
  });

  const deactivate = useMutation<
    UOM | HandlingUnitType | ItemUOM | ItemBarcode,
    Error,
    DeactivateTarget
  >({
    mutationFn: (target) => {
      if (target.kind === "uom") return deactivateUOM(target.item.uom_id);
      if (target.kind === "handling")
        return deactivateHandlingUnitType(target.item.handling_unit_type_id);
      if (target.kind === "item-uom")
        return deactivateItemUOM(filters.item, target.item.item_uom_id);
      return deactivateItemBarcode(filters.item, target.item.item_barcode_id);
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: unitsPackagingKeys.all }),
        queryClient.invalidateQueries({ queryKey: itemCatalogKeys.uomLists() }),
        queryClient.invalidateQueries({
          queryKey: itemCatalogKeys.item(filters.item),
        }),
      ]);
      toast.success("Record deactivated.");
      setDeactivateTarget(undefined);
    },
  });
  const makePrimary = useMutation({
    mutationFn: (barcode: ItemBarcode) =>
      setPrimaryItemBarcode(filters.item, barcode.item_barcode_id),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: unitsPackagingKeys.barcodes(filters.item),
        }),
        queryClient.invalidateQueries({
          queryKey: itemCatalogKeys.item(filters.item),
        }),
      ]);
      toast.success("Primary barcode updated.");
    },
    onError: (error) => toast.error(error.message),
  });

  const ownerOptions = ownerItems.map((owner) => ({
    value: owner.organization_id,
    label: `${owner.name} (${owner.code})`,
  }));
  const itemOptions = items.map((item) => ({
    value: item.item_id,
    label: `${item.name} (${item.code})${item.is_active ? "" : " · Inactive"}`,
  }));
  const globalData = filters.view === "uoms" ? uomList.data : handlingList.data;
  const globalQuery = filters.view === "uoms" ? uomList : handlingList;
  const childQuery = filters.view === "barcodes" ? barcodes : itemUOMs;
  const currentError = itemScoped
    ? (owners.error ??
      itemLookup.error ??
      itemUOMs.error ??
      barcodes.error ??
      uomLookup.error)
    : globalQuery.error;
  const currentPending = itemScoped
    ? itemLookup.isPending ||
      (Boolean(filters.item) &&
        (itemUOMs.isPending ||
          (filters.view === "barcodes" && barcodes.isPending)))
    : globalQuery.isPending;
  const activeAssignedIds = new Set(
    (itemUOMs.data ?? []).map((unit) => unit.uom_id),
  );
  const canCreateItemUOM =
    Boolean(currentItem?.is_active) &&
    allUOMs.some((uom) => uom.is_active && !activeAssignedIds.has(uom.uom_id));
  const createDisabled = itemScoped
    ? !currentItem ||
      !currentItem.is_active ||
      (filters.view === "item-uoms" && !canCreateItemUOM)
    : false;

  function startCreate() {
    if (filters.view === "uoms") setEditor({ kind: "uom" });
    else if (filters.view === "handling-units") setEditor({ kind: "handling" });
    else if (filters.view === "item-uoms") setEditor({ kind: "item-uom" });
    else setEditor({ kind: "barcode" });
  }

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <Link
            href="/master-data/catalog"
            className="inline-flex min-h-9 items-center gap-2 text-sm font-semibold text-slate-600 hover:text-slate-950"
          >
            <ChevronLeft className="size-4" />
            Catalog
          </Link>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <PackageOpen className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Units and packaging
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Define global units, item conversions, package dimensions, and
                scannable identifiers.
              </p>
            </div>
          </div>
        </div>
        {canWrite ? (
          <Button
            className="w-full sm:w-auto"
            disabled={createDisabled}
            onClick={startCreate}
          >
            <Plus className="size-4" />
            Create{" "}
            {filters.view === "uoms"
              ? "UOM"
              : filters.view === "handling-units"
                ? "handling unit"
                : filters.view === "item-uoms"
                  ? "item UOM"
                  : "barcode"}
          </Button>
        ) : null}
      </header>

      <div
        className="flex gap-2 overflow-x-auto pb-1"
        role="tablist"
        aria-label="Units and packaging sections"
      >
        {views.map((view) => (
          <button
            key={view}
            type="button"
            role="tab"
            aria-selected={filters.view === view}
            onClick={() => {
              setSearchDraft("");
              void setFilters({ view, search: "", active: "all", page: 1 });
            }}
            className={cn(
              "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold transition-colors",
              filters.view === view
                ? "bg-slate-950 text-white"
                : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
            )}
          >
            {viewLabels[view]}
          </button>
        ))}
      </div>

      {itemScoped ? (
        <Panel className="grid gap-4 p-4 sm:p-5 lg:grid-cols-2">
          <div>
            <p className="mb-2 text-xs font-bold tracking-wide text-slate-500 uppercase">
              Owner context
            </p>
            <Select
              ariaLabel="Owner organization"
              value={filters.owner}
              options={ownerOptions}
              onValueChange={(owner) => {
                setSearchDraft("");
                void setFilters({ owner, item: "", search: "" });
              }}
              placeholder={
                owners.isPending ? "Loading owners…" : "Select an owner"
              }
              disabled={owners.isPending || !ownerOptions.length}
            />
          </div>
          <div>
            <p className="mb-2 text-xs font-bold tracking-wide text-slate-500 uppercase">
              Item context
            </p>
            <Select
              ariaLabel="Catalog item"
              value={filters.item}
              options={itemOptions}
              onValueChange={(item) => {
                setSearchDraft("");
                void setFilters({ item, search: "", active: "all" });
              }}
              placeholder={
                itemLookup.isPending
                  ? "Loading items…"
                  : itemOptions.length
                    ? "Select an item"
                    : "No items for this owner"
              }
              disabled={
                !filters.owner || itemLookup.isPending || !itemOptions.length
              }
            />
          </div>
          {currentItem && !currentItem.is_active ? (
            <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-900 lg:col-span-2">
              This item is inactive. Its packaging remains visible, but new or
              reactivated records are not allowed.
            </p>
          ) : null}
        </Panel>
      ) : null}

      <Panel className="overflow-hidden">
        <form
          className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:p-5"
          onSubmit={(event) => {
            event.preventDefault();
            void setFilters({ search: searchDraft.trim(), page: 1 });
          }}
        >
          <label className="relative flex-1">
            <span className="sr-only">Search packaging</span>
            <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
            <input
              value={searchDraft}
              onChange={(event) => setSearchDraft(event.target.value)}
              placeholder={`Search ${viewLabels[filters.view].toLowerCase()}`}
              className="h-11 w-full rounded-xl border border-slate-300 bg-white pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
            />
          </label>
          <Select
            ariaLabel="Filter by status"
            value={filters.active}
            options={statusOptions}
            onValueChange={(active) => void setFilters({ active, page: 1 })}
            className="sm:w-40"
          />
          <Button type="submit" variant="secondary">
            Search
          </Button>
          {filters.search ? (
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                setSearchDraft("");
                void setFilters({ search: "", page: 1 });
              }}
            >
              Clear
            </Button>
          ) : null}
        </form>

        {itemScoped && !filters.item ? (
          <div className="p-8 text-center text-sm text-slate-500">
            Select an owner and item to manage its packaging.
          </div>
        ) : currentPending ? (
          <LoadingState />
        ) : currentError ? (
          <ErrorState
            error={currentError}
            retry={() => {
              if (itemScoped) {
                void itemLookup.refetch();
                void childQuery.refetch();
                void uomLookup.refetch();
              } else void globalQuery.refetch();
            }}
          />
        ) : filters.view === "uoms" ? (
          !uomList.data?.items.length ? (
            <EmptyState view="uoms" />
          ) : (
            <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
              {uomList.data.items.map((uom) => (
                <article
                  key={uom.uom_id}
                  className="flex min-h-44 flex-col rounded-xl border border-slate-200 p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h2 className="font-bold text-slate-950">{uom.name}</h2>
                      <p className="mt-0.5 font-mono text-xs text-slate-500">
                        {uom.code}
                      </p>
                    </div>
                    <StatusBadge tone={uom.is_active ? "success" : "neutral"}>
                      {uom.is_active ? "Active" : "Inactive"}
                    </StatusBadge>
                  </div>
                  <div className="mt-4 flex flex-1 items-center gap-3 rounded-xl bg-slate-50 p-3">
                    <Scale className="size-5 text-cyan-800" />
                    <p className="text-sm text-slate-600">
                      <span className="font-bold text-slate-950">
                        {uom.decimal_scale}
                      </span>{" "}
                      decimal places
                    </p>
                  </div>
                  {canWrite ? (
                    <div className="mt-4 flex justify-end gap-1 border-t border-slate-100 pt-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setEditor({ kind: "uom", item: uom })}
                      >
                        <Pencil className="size-4" />
                        Edit
                      </Button>
                      {uom.is_active ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-rose-700"
                          onClick={() =>
                            setDeactivateTarget({ kind: "uom", item: uom })
                          }
                        >
                          Deactivate
                        </Button>
                      ) : null}
                    </div>
                  ) : null}
                </article>
              ))}
            </div>
          )
        ) : filters.view === "handling-units" ? (
          !handlingList.data?.items.length ? (
            <EmptyState view="handling-units" />
          ) : (
            <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
              {handlingList.data.items.map((unit) => (
                <article
                  key={unit.handling_unit_type_id}
                  className="flex min-h-48 flex-col rounded-xl border border-slate-200 p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h2 className="font-bold text-slate-950">{unit.name}</h2>
                      <p className="mt-0.5 font-mono text-xs text-slate-500">
                        {unit.code}
                      </p>
                    </div>
                    <StatusBadge tone={unit.is_active ? "success" : "neutral"}>
                      {unit.is_active ? "Active" : "Inactive"}
                    </StatusBadge>
                  </div>
                  <dl className="mt-4 grid flex-1 grid-cols-2 gap-3 text-sm">
                    <div className="rounded-xl bg-slate-50 p-3">
                      <dt className="text-xs text-slate-500">Max weight</dt>
                      <dd className="mt-1 font-semibold text-slate-900">
                        {unit.max_weight ?? "Not limited"}
                      </dd>
                    </div>
                    <div className="rounded-xl bg-slate-50 p-3">
                      <dt className="text-xs text-slate-500">Max volume</dt>
                      <dd className="mt-1 font-semibold text-slate-900">
                        {unit.max_volume ?? "Not limited"}
                      </dd>
                    </div>
                  </dl>
                  {canWrite ? (
                    <div className="mt-4 flex justify-end gap-1 border-t border-slate-100 pt-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          setEditor({ kind: "handling", item: unit })
                        }
                      >
                        <Pencil className="size-4" />
                        Edit
                      </Button>
                      {unit.is_active ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-rose-700"
                          onClick={() =>
                            setDeactivateTarget({
                              kind: "handling",
                              item: unit,
                            })
                          }
                        >
                          Deactivate
                        </Button>
                      ) : null}
                    </div>
                  ) : null}
                </article>
              ))}
            </div>
          )
        ) : filters.view === "item-uoms" ? (
          !visibleItemUOMs.length ? (
            <EmptyState view="item-uoms" />
          ) : (
            <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
              {visibleItemUOMs.map((unit) => {
                const uom = uomById.get(unit.uom_id);
                const isBase = unit.uom_id === currentItem?.base_uom_id;
                return (
                  <article
                    key={unit.item_uom_id}
                    className="flex min-h-56 flex-col rounded-xl border border-slate-200 p-4"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <h2 className="font-bold text-slate-950">
                          {uom?.name ?? "Unit"}
                        </h2>
                        <p className="mt-0.5 font-mono text-xs text-slate-500">
                          {uom?.code ?? unit.uom_id}
                        </p>
                      </div>
                      <div className="flex flex-wrap justify-end gap-1">
                        {isBase ? (
                          <StatusBadge tone="info">Base</StatusBadge>
                        ) : null}
                        <StatusBadge
                          tone={unit.is_active ? "success" : "neutral"}
                        >
                          {unit.is_active ? "Active" : "Inactive"}
                        </StatusBadge>
                      </div>
                    </div>
                    <p className="mt-4 text-sm text-slate-600">
                      <span className="font-bold text-slate-950">
                        1 {uom?.code ?? "unit"}
                      </span>{" "}
                      = {unit.conversion_to_base} base units
                    </p>
                    <div className="mt-3 flex flex-wrap gap-1">
                      {unit.is_receiving_uom ? (
                        <StatusBadge tone="info">Receiving</StatusBadge>
                      ) : null}
                      {unit.is_picking_uom ? (
                        <StatusBadge tone="warning">Picking</StatusBadge>
                      ) : null}
                    </div>
                    <p className="mt-3 flex-1 text-xs text-slate-500">
                      L/W/H:{" "}
                      {[unit.length, unit.width, unit.height]
                        .map((value) => value ?? "—")
                        .join(" / ")}{" "}
                      · Weight: {unit.weight ?? "—"}
                    </p>
                    {canWrite ? (
                      <div className="mt-4 flex justify-end gap-1 border-t border-slate-100 pt-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() =>
                            setEditor({ kind: "item-uom", item: unit })
                          }
                        >
                          <Pencil className="size-4" />
                          Edit
                        </Button>
                        {unit.is_active && !isBase ? (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-rose-700"
                            onClick={() =>
                              setDeactivateTarget({
                                kind: "item-uom",
                                item: unit,
                              })
                            }
                          >
                            Deactivate
                          </Button>
                        ) : null}
                      </div>
                    ) : null}
                  </article>
                );
              })}
            </div>
          )
        ) : !visibleBarcodes.length ? (
          <EmptyState view="barcodes" />
        ) : (
          <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
            {visibleBarcodes.map((barcode) => {
              const uom = barcode.uom_id
                ? uomById.get(barcode.uom_id)
                : undefined;
              return (
                <article
                  key={barcode.item_barcode_id}
                  className="flex min-h-48 flex-col rounded-xl border border-slate-200 p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <BarcodeIcon className="size-5 text-cyan-800" />
                    <div className="flex flex-wrap justify-end gap-1">
                      {barcode.is_primary ? (
                        <StatusBadge tone="warning">
                          <Star className="size-3" />
                          Primary
                        </StatusBadge>
                      ) : null}
                      <StatusBadge
                        tone={barcode.is_active ? "success" : "neutral"}
                      >
                        {barcode.is_active ? "Active" : "Inactive"}
                      </StatusBadge>
                    </div>
                  </div>
                  <p className="mt-4 font-mono text-base font-bold break-all text-slate-950">
                    {barcode.barcode}
                  </p>
                  <p className="mt-2 flex-1 text-sm text-slate-600">
                    {uom ? `${uom.name} (${uom.code})` : "Item-wide barcode"}
                  </p>
                  {canWrite ? (
                    <div className="mt-4 flex flex-wrap justify-end gap-1 border-t border-slate-100 pt-2">
                      {barcode.is_active &&
                      !barcode.is_primary &&
                      currentItem?.is_active ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          disabled={makePrimary.isPending}
                          onClick={() => makePrimary.mutate(barcode)}
                        >
                          <Star className="size-4" />
                          Make primary
                        </Button>
                      ) : null}
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          setEditor({ kind: "barcode", item: barcode })
                        }
                      >
                        <Pencil className="size-4" />
                        Edit
                      </Button>
                      {barcode.is_active ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-rose-700"
                          onClick={() =>
                            setDeactivateTarget({
                              kind: "barcode",
                              item: barcode,
                            })
                          }
                        >
                          Deactivate
                        </Button>
                      ) : null}
                    </div>
                  ) : null}
                </article>
              );
            })}
          </div>
        )}

        {!itemScoped && !globalQuery.isPending && !globalQuery.isError ? (
          <Pagination
            page={globalData?.page ?? referenceFilters.page}
            totalPages={globalData?.total_pages ?? 0}
            onPage={(page) => void setFilters({ page })}
          />
        ) : null}
      </Panel>

      {editor?.kind === "uom" ? (
        <UOMDialog
          uom={editor.item}
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
      {editor?.kind === "handling" ? (
        <HandlingUnitDialog
          handlingUnit={editor.item}
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
      {editor?.kind === "item-uom" && currentItem ? (
        <ItemUOMDialog
          item={currentItem}
          itemUOM={editor.item}
          itemUOMs={itemUOMs.data ?? []}
          uoms={allUOMs}
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
      {editor?.kind === "barcode" && currentItem ? (
        <BarcodeDialog
          item={currentItem}
          barcode={editor.item}
          itemUOMs={itemUOMs.data ?? []}
          uoms={allUOMs}
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
      <DeactivateDialog
        target={deactivateTarget}
        pending={deactivate.isPending}
        error={deactivate.error}
        onConfirm={() => {
          if (deactivateTarget) deactivate.mutate(deactivateTarget);
        }}
        onOpenChange={(open) => {
          if (!open) {
            deactivate.reset();
            setDeactivateTarget(undefined);
          }
        }}
      />
    </div>
  );
}
