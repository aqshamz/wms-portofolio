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
  Boxes,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  FolderTree,
  LoaderCircle,
  Pencil,
  Plus,
  Search,
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
  deactivateItem,
  deactivateItemCategory,
  itemCatalogKeys,
  listItemCategories,
  listItems,
  listUOMs,
} from "@/features/item-catalog/item-catalog-api";
import {
  ItemCategoryDialog,
  ItemDialog,
} from "@/features/item-catalog/item-catalog-dialogs";
import type {
  CatalogItem,
  ItemCategory,
} from "@/features/item-catalog/item-catalog-types";
import {
  listOrganizations,
  organizationKeys,
} from "@/features/organizations/organization-api";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 10;
const views = ["items", "categories"] as const;
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
const uomFilters = {
  search: "",
  active: "all" as const,
  page: 1,
  pageSize: 100,
};

type Editor =
  | { kind: "item"; item?: CatalogItem }
  | { kind: "category"; item?: ItemCategory }
  | null;
type DeactivateTarget =
  | { kind: "item"; item: CatalogItem }
  | { kind: "category"; item: ItemCategory };

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(new Date(value));
}

function LoadingState() {
  return (
    <div className="space-y-3 p-5">
      {Array.from({ length: 5 }, (_, index) => (
        <div
          key={index}
          className="h-16 animate-pulse rounded-xl bg-slate-100"
        />
      ))}
    </div>
  );
}

function EmptyState({ view }: { view: (typeof views)[number] }) {
  return (
    <div className="px-5 py-14 text-center">
      {view === "items" ? (
        <Boxes className="mx-auto size-9 text-slate-300" />
      ) : (
        <FolderTree className="mx-auto size-9 text-slate-300" />
      )}
      <h2 className="mt-4 font-bold text-slate-950">
        No {view === "items" ? "items" : "categories"} found
      </h2>
      <p className="mt-1 text-sm text-slate-500">
        Adjust the filters or create the first{" "}
        {view === "items" ? "item" : "category"}.
      </p>
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
            <p className="font-semibold">Item catalog could not be loaded</p>
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
            Deactivate {target?.item.name ?? "record"}?
          </Dialog.Title>
          <Dialog.Description className="mt-2 text-sm leading-6 text-slate-600">
            {target?.kind === "category"
              ? "The category remains on existing records but cannot be assigned to new or updated items. Child categories are not changed."
              : "The item remains in historical records but cannot be selected for new operations."}
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

export function ItemCatalogScreen({ canWrite }: { canWrite: boolean }) {
  const [filters, setFilters] = useQueryStates({
    view: parseAsStringLiteral(views).withDefault("items"),
    owner: parseAsString.withDefault(""),
    category: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    active: parseAsStringLiteral(activeValues).withDefault("all"),
    page: parseAsInteger.withDefault(1),
  });
  const [editor, setEditor] = useState<Editor>(null);
  const [deactivateTarget, setDeactivateTarget] = useState<DeactivateTarget>();
  const queryClient = useQueryClient();
  const owners = useQuery({
    queryKey: organizationKeys.list(ownerFilters),
    queryFn: () => listOrganizations(ownerFilters),
  });
  const categoryLookupFilters = {
    ownerId: filters.owner,
    search: "",
    active: "all" as const,
    page: 1,
    pageSize: 100,
  };
  const categoryLookup = useQuery({
    queryKey: itemCatalogKeys.categoryList(categoryLookupFilters),
    queryFn: () => listItemCategories(categoryLookupFilters),
    enabled: Boolean(filters.owner),
  });
  const uoms = useQuery({
    queryKey: itemCatalogKeys.uomList(uomFilters),
    queryFn: () => listUOMs(uomFilters),
    enabled: filters.view === "items",
  });
  const itemFilters = {
    ownerId: filters.owner,
    categoryId: filters.category,
    search: filters.search,
    active: filters.active,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };
  const categoryFilters = {
    ownerId: filters.owner,
    search: filters.search,
    active: filters.active,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };
  const items = useQuery({
    queryKey: itemCatalogKeys.itemList(itemFilters),
    queryFn: () => listItems(itemFilters),
    enabled: filters.view === "items" && Boolean(filters.owner),
    placeholderData: keepPreviousData,
  });
  const categories = useQuery({
    queryKey: itemCatalogKeys.categoryList(categoryFilters),
    queryFn: () => listItemCategories(categoryFilters),
    enabled: filters.view === "categories" && Boolean(filters.owner),
    placeholderData: keepPreviousData,
  });
  const deactivate = useMutation<
    CatalogItem | ItemCategory,
    Error,
    DeactivateTarget
  >({
    mutationFn: (target) =>
      target.kind === "item"
        ? deactivateItem(target.item.item_id, target.item.updated_at)
        : deactivateItemCategory(target.item.category_id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: itemCatalogKeys.all });
      toast.success(
        deactivateTarget?.kind === "item"
          ? "Item deactivated."
          : "Item category deactivated.",
      );
      setDeactivateTarget(undefined);
    },
  });

  const ownerItems = owners.data?.items ?? [];
  const firstOwnerId = ownerItems[0]?.organization_id;
  useEffect(() => {
    if (!filters.owner && firstOwnerId)
      void setFilters({ owner: firstOwnerId, page: 1 });
  }, [filters.owner, firstOwnerId, setFilters]);

  const ownerOptions = ownerItems.map((owner) => ({
    value: owner.organization_id,
    label: `${owner.name} (${owner.code})`,
  }));
  const currentOwner = ownerItems.find(
    (owner) => owner.organization_id === filters.owner,
  );
  const ownerLabel = currentOwner
    ? `${currentOwner.name} (${currentOwner.code})`
    : "Selected owner";
  const categoryItems = categoryLookup.data?.items ?? [];
  const categoryById = new Map(
    categoryItems.map((category) => [category.category_id, category]),
  );
  const uomItems = uoms.data?.items ?? [];
  const uomById = new Map(uomItems.map((uom) => [uom.uom_id, uom]));
  const categoryOptions = [
    { value: "all", label: "All categories" },
    ...categoryItems.map((category) => ({
      value: category.category_id,
      label: `${category.name} (${category.code})`,
    })),
  ];
  const currentQuery = filters.view === "items" ? items : categories;
  const currentData = filters.view === "items" ? items.data : categories.data;
  const currentPage = currentData?.page ?? Math.max(1, filters.page);
  const totalPages = currentData?.total_pages ?? 0;
  const totalItems = currentData?.total_items ?? 0;
  const hasActiveUOM = uomItems.some((uom) => uom.is_active);
  const createDisabled =
    !filters.owner ||
    (filters.view === "items" && (!hasActiveUOM || uoms.isPending));

  function startCreate() {
    setEditor(
      filters.view === "items" ? { kind: "item" } : { kind: "category" },
    );
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
              <Boxes className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Item catalog
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Maintain owner SKUs, category hierarchy, and inventory control
                settings.
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
            Create {filters.view === "items" ? "item" : "category"}
          </Button>
        ) : null}
      </header>

      <div
        className="flex gap-2 overflow-x-auto pb-1"
        role="tablist"
        aria-label="Item catalog sections"
      >
        {views.map((view) => (
          <button
            key={view}
            type="button"
            role="tab"
            aria-selected={filters.view === view}
            onClick={() =>
              void setFilters({
                view,
                category: "",
                search: "",
                active: "all",
                page: 1,
              })
            }
            className={cn(
              "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold transition-colors",
              filters.view === view
                ? "bg-slate-950 text-white"
                : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
            )}
          >
            {view === "items" ? "Items" : "Categories"}
          </button>
        ))}
      </div>

      <Panel className="p-4 sm:p-5">
        <p className="mb-2 text-xs font-bold tracking-wide text-slate-500 uppercase">
          Owner context
        </p>
        <Select
          ariaLabel="Owner organization"
          value={filters.owner}
          options={ownerOptions}
          onValueChange={(owner) =>
            void setFilters({ owner, category: "", search: "", page: 1 })
          }
          placeholder={
            owners.isPending
              ? "Loading owners…"
              : ownerOptions.length
                ? "Select an owner"
                : "No active organizations"
          }
          disabled={owners.isPending || ownerOptions.length === 0}
          className="max-w-xl"
        />
        {owners.isError ? (
          <p className="mt-2 text-sm text-rose-700">{owners.error.message}</p>
        ) : null}
      </Panel>

      <Panel className="overflow-hidden">
        <form
          className="flex flex-col gap-3 border-b border-slate-200 p-4 lg:flex-row lg:p-5"
          onSubmit={(event) => {
            event.preventDefault();
            const data = new FormData(event.currentTarget);
            void setFilters({
              search: String(data.get("search") ?? "").trim(),
              page: 1,
            });
          }}
        >
          <label className="relative flex-1">
            <span className="sr-only">Search item catalog</span>
            <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
            <input
              key={`${filters.view}-${filters.search}`}
              name="search"
              defaultValue={filters.search}
              placeholder={
                filters.view === "items"
                  ? "Search by item code or name"
                  : "Search categories"
              }
              className="h-11 w-full rounded-xl border border-slate-300 bg-white pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
            />
          </label>
          {filters.view === "items" ? (
            <Select
              ariaLabel="Filter by category"
              value={filters.category || "all"}
              options={categoryOptions}
              onValueChange={(category) =>
                void setFilters({
                  category: category === "all" ? "" : category,
                  page: 1,
                })
              }
              className="lg:w-56"
            />
          ) : null}
          <Select
            ariaLabel="Filter by status"
            value={filters.active}
            options={statusOptions}
            onValueChange={(active) => void setFilters({ active, page: 1 })}
            className="lg:w-40"
          />
          <Button type="submit" variant="secondary">
            Search
          </Button>
          {filters.search ? (
            <Button
              type="button"
              variant="ghost"
              onClick={() => void setFilters({ search: "", page: 1 })}
            >
              Clear
            </Button>
          ) : null}
        </form>

        {!currentQuery.isPending && !currentQuery.isError && filters.owner ? (
          <div className="border-b border-slate-200 px-4 py-3 text-sm text-slate-600 sm:px-5">
            <span className="font-semibold text-slate-950">{totalItems}</span>{" "}
            {filters.view}
          </div>
        ) : null}
        {!filters.owner ? (
          <div className="p-8 text-center text-sm text-slate-500">
            Select an owner to view its item catalog.
          </div>
        ) : currentQuery.isPending ? (
          <LoadingState />
        ) : currentQuery.isError ? (
          <ErrorState
            error={currentQuery.error}
            retry={() => void currentQuery.refetch()}
          />
        ) : filters.view === "items" ? (
          !items.data?.items.length ? (
            <EmptyState view="items" />
          ) : (
            <>
              <div className="hidden overflow-x-auto md:block">
                <table className="w-full min-w-[980px] text-left">
                  <thead className="bg-slate-50 text-xs font-bold tracking-wide text-slate-500 uppercase">
                    <tr>
                      <th className="px-5 py-3">Item</th>
                      <th className="px-5 py-3">Category / UOM</th>
                      <th className="px-5 py-3">Controls</th>
                      <th className="px-5 py-3">Physical</th>
                      <th className="px-5 py-3">Status</th>
                      <th className="px-5 py-3">Updated</th>
                      <th className="px-5 py-3 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-200">
                    {items.data.items.map((item) => {
                      const category = item.category_id
                        ? categoryById.get(item.category_id)
                        : undefined;
                      const uom = uomById.get(item.base_uom_id);
                      return (
                        <tr key={item.item_id} className="hover:bg-slate-50/70">
                          <td className="px-5 py-4">
                            <p className="font-semibold text-slate-950">
                              {item.name}
                            </p>
                            <p className="mt-0.5 font-mono text-xs text-slate-500">
                              {item.code}
                            </p>
                          </td>
                          <td className="px-5 py-4 text-sm text-slate-600">
                            <p>{category ? category.name : "Uncategorized"}</p>
                            <p className="mt-1 text-xs">
                              Base: {uom?.code ?? item.base_uom_id}
                            </p>
                          </td>
                          <td className="px-5 py-4">
                            <div className="flex flex-wrap gap-1">
                              {item.lot_controlled ? (
                                <StatusBadge tone="info">Lot</StatusBadge>
                              ) : null}
                              {item.serial_controlled ? (
                                <StatusBadge tone="warning">Serial</StatusBadge>
                              ) : null}
                              {!item.lot_controlled &&
                              !item.serial_controlled ? (
                                <span className="text-sm text-slate-400">
                                  None
                                </span>
                              ) : null}
                            </div>
                          </td>
                          <td className="px-5 py-4 text-sm text-slate-600">
                            <p>
                              {item.weight
                                ? `${item.weight} weight`
                                : "No weight"}
                            </p>
                            <p className="mt-1 text-xs">
                              {item.volume
                                ? `${item.volume} volume`
                                : "No volume"}
                            </p>
                          </td>
                          <td className="px-5 py-4">
                            <StatusBadge
                              tone={item.is_active ? "success" : "neutral"}
                            >
                              {item.is_active ? "Active" : "Inactive"}
                            </StatusBadge>
                          </td>
                          <td className="px-5 py-4 text-sm text-slate-600">
                            {formatDate(item.updated_at)}
                          </td>
                          <td className="px-5 py-4">
                            <div className="flex justify-end gap-1">
                              {canWrite ? (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() =>
                                    setEditor({ kind: "item", item })
                                  }
                                >
                                  <Pencil className="size-4" />
                                  Edit
                                </Button>
                              ) : null}
                              {canWrite && item.is_active ? (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="text-rose-700"
                                  onClick={() =>
                                    setDeactivateTarget({ kind: "item", item })
                                  }
                                >
                                  Deactivate
                                </Button>
                              ) : null}
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
              <ul className="divide-y divide-slate-200 md:hidden">
                {items.data.items.map((item) => {
                  const category = item.category_id
                    ? categoryById.get(item.category_id)
                    : undefined;
                  const uom = uomById.get(item.base_uom_id);
                  return (
                    <li key={item.item_id} className="p-4">
                      <div className="flex items-start justify-between gap-3">
                        <div>
                          <p className="font-semibold text-slate-950">
                            {item.name}
                          </p>
                          <p className="mt-0.5 font-mono text-xs text-slate-500">
                            {item.code}
                          </p>
                        </div>
                        <StatusBadge
                          tone={item.is_active ? "success" : "neutral"}
                        >
                          {item.is_active ? "Active" : "Inactive"}
                        </StatusBadge>
                      </div>
                      <p className="mt-3 text-sm text-slate-600">
                        {category?.name ?? "Uncategorized"} ·{" "}
                        {uom?.code ?? "Base UOM"}
                      </p>
                      <div className="mt-2 flex gap-1">
                        {item.lot_controlled ? (
                          <StatusBadge tone="info">Lot controlled</StatusBadge>
                        ) : null}
                        {item.serial_controlled ? (
                          <StatusBadge tone="warning">
                            Serial controlled
                          </StatusBadge>
                        ) : null}
                      </div>
                      {canWrite ? (
                        <div className="mt-3 flex justify-end gap-1 border-t border-slate-100 pt-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setEditor({ kind: "item", item })}
                          >
                            <Pencil className="size-4" />
                            Edit
                          </Button>
                          {item.is_active ? (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-rose-700"
                              onClick={() =>
                                setDeactivateTarget({ kind: "item", item })
                              }
                            >
                              Deactivate
                            </Button>
                          ) : null}
                        </div>
                      ) : null}
                    </li>
                  );
                })}
              </ul>
            </>
          )
        ) : !categories.data?.items.length ? (
          <EmptyState view="categories" />
        ) : (
          <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
            {categories.data.items.map((category) => {
              const parent = category.parent_category_id
                ? categoryById.get(category.parent_category_id)
                : undefined;
              return (
                <article
                  key={category.category_id}
                  className="flex min-h-44 flex-col rounded-xl border border-slate-200 p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h2 className="font-bold text-slate-950">
                        {category.name}
                      </h2>
                      <p className="mt-0.5 font-mono text-xs text-slate-500">
                        {category.code}
                      </p>
                    </div>
                    <StatusBadge
                      tone={category.is_active ? "success" : "neutral"}
                    >
                      {category.is_active ? "Active" : "Inactive"}
                    </StatusBadge>
                  </div>
                  <p className="mt-4 flex-1 text-sm text-slate-600">
                    {parent ? (
                      <>
                        Child of{" "}
                        <span className="font-semibold text-slate-800">
                          {parent.name}
                        </span>
                      </>
                    ) : (
                      "Top-level category"
                    )}
                  </p>
                  {canWrite ? (
                    <div className="mt-4 flex justify-end gap-1 border-t border-slate-100 pt-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          setEditor({ kind: "category", item: category })
                        }
                      >
                        <Pencil className="size-4" />
                        Edit
                      </Button>
                      {category.is_active ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-rose-700"
                          onClick={() =>
                            setDeactivateTarget({
                              kind: "category",
                              item: category,
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
        {!currentQuery.isPending && !currentQuery.isError && filters.owner ? (
          <Pagination
            page={currentPage}
            totalPages={totalPages}
            onPage={(page) => void setFilters({ page })}
          />
        ) : null}
      </Panel>

      {editor?.kind === "item" && filters.owner ? (
        <ItemDialog
          ownerId={filters.owner}
          ownerLabel={ownerLabel}
          item={editor.item}
          categories={categoryItems}
          uoms={uomItems}
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
      {editor?.kind === "category" && filters.owner ? (
        <ItemCategoryDialog
          ownerId={filters.owner}
          ownerLabel={ownerLabel}
          category={editor.item}
          categories={categoryItems}
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
