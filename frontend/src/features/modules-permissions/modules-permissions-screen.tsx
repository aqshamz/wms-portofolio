"use client";

import { useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  AlertTriangle,
  Blocks,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  KeyRound,
  LoaderCircle,
  Pencil,
  Plus,
  Search,
  ShieldCheck,
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
import { AppModuleDialog } from "@/features/modules-permissions/app-module-dialog";
import {
  deactivateAppModule,
  listAppModules,
  listWorkflowPermissions,
  modulesPermissionsKeys,
} from "@/features/modules-permissions/modules-permissions-api";
import type {
  AppModule,
  ModulesPermissionsView,
} from "@/features/modules-permissions/modules-permissions-types";
import { cn } from "@/lib/utils";

const views = ["modules", "permissions"] as const;
const activeValues = ["all", "active", "inactive"] as const;
const PAGE_SIZE = 12;
const statusOptions = [
  { value: "all", label: "All statuses" },
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
] as const;
const allModulesFilter = {
  search: "",
  active: "all" as const,
  page: 1,
  pageSize: 100,
};

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
        Page <span className="font-semibold text-slate-950">{page}</span> of{" "}
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

function LoadingState() {
  return (
    <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
      {Array.from({ length: 6 }, (_, index) => (
        <div
          key={index}
          className="h-48 animate-pulse rounded-xl bg-slate-100"
        />
      ))}
    </div>
  );
}

function ErrorState({
  view,
  error,
  retry,
}: {
  view: ModulesPermissionsView;
  error: Error;
  retry: () => void;
}) {
  return (
    <div className="p-5">
      <div
        role="alert"
        className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900"
      >
        <div className="flex gap-3">
          <CircleAlert className="mt-0.5 size-5 shrink-0" />
          <div>
            <p className="font-semibold">
              {view === "modules"
                ? "Application modules could not be loaded"
                : "Workflow permissions could not be loaded"}
            </p>
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

function DeactivateDialog({
  item,
  pending,
  error,
  onConfirm,
  onOpenChange,
}: {
  item?: AppModule;
  pending: boolean;
  error?: Error | null;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog.Root
      open={item !== undefined}
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
            Deactivate {item?.name ?? "module"}?
          </Dialog.Title>
          <Dialog.Description className="mt-2 text-sm leading-6 text-slate-600">
            The module remains in historical records, but it will no longer be
            available as an active workflow grouping. Existing permission
            definitions are not deleted.
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

export function ModulesPermissionsScreen({ canWrite }: { canWrite: boolean }) {
  const [filters, setFilters] = useQueryStates({
    view: parseAsStringLiteral(views).withDefault("modules"),
    search: parseAsString.withDefault(""),
    active: parseAsStringLiteral(activeValues).withDefault("all"),
    module: parseAsString.withDefault("all"),
    page: parseAsInteger.withDefault(1),
  });
  const [editor, setEditor] = useState<{ item?: AppModule }>();
  const [deactivateTarget, setDeactivateTarget] = useState<AppModule>();
  const queryClient = useQueryClient();
  const listFilters = {
    search: filters.search,
    active: filters.active,
    moduleCode: filters.module === "all" ? undefined : filters.module,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };
  const modules = useQuery({
    queryKey: modulesPermissionsKeys.moduleList(listFilters),
    queryFn: () => listAppModules(listFilters),
    enabled: filters.view === "modules",
    placeholderData: keepPreviousData,
  });
  const permissions = useQuery({
    queryKey: modulesPermissionsKeys.permissionList(listFilters),
    queryFn: () => listWorkflowPermissions(listFilters),
    enabled: filters.view === "permissions",
    placeholderData: keepPreviousData,
  });
  const moduleOptionsQuery = useQuery({
    queryKey: modulesPermissionsKeys.moduleList(allModulesFilter),
    queryFn: () => listAppModules(allModulesFilter),
    enabled: filters.view === "permissions",
  });
  const deactivate = useMutation({
    mutationFn: (item: AppModule) => deactivateAppModule(item.module_id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: modulesPermissionsKeys.all,
      });
      toast.success("Application module deactivated.");
      setDeactivateTarget(undefined);
    },
  });
  const currentQuery = filters.view === "modules" ? modules : permissions;
  const currentData =
    filters.view === "modules" ? modules.data : permissions.data;
  const moduleItems = modules.data?.items ?? [];
  const permissionItems = permissions.data?.items ?? [];
  const currentPage = currentData?.page ?? listFilters.page;
  const totalPages = currentData?.total_pages ?? 0;
  const moduleOptions = [
    { value: "all", label: "All modules" },
    ...(moduleOptionsQuery.data?.items ?? []).map((module) => ({
      value: module.code,
      label: `${module.name} (${module.code})`,
    })),
  ];

  function switchView(view: ModulesPermissionsView) {
    void setFilters({
      view,
      search: "",
      active: "all",
      module: "all",
      page: 1,
    });
  }

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <Link
            href="/master-data/operational"
            className="inline-flex min-h-9 items-center gap-2 text-sm font-semibold text-slate-600 hover:text-slate-950"
          >
            <ChevronLeft className="size-4" />
            Operational setup
          </Link>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <Blocks className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Modules and workflow permissions
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Organize application capabilities and inspect permission codes
                used by workflow transitions.
              </p>
            </div>
          </div>
        </div>
        {canWrite && filters.view === "modules" ? (
          <Button className="w-full sm:w-auto" onClick={() => setEditor({})}>
            <Plus className="size-4" />
            Create module
          </Button>
        ) : null}
      </header>

      <div
        className="flex gap-2 overflow-x-auto pb-1"
        role="tablist"
        aria-label="Module and permission sections"
      >
        <button
          type="button"
          role="tab"
          aria-selected={filters.view === "modules"}
          onClick={() => switchView("modules")}
          className={cn(
            "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold transition-colors",
            filters.view === "modules"
              ? "bg-slate-950 text-white"
              : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
          )}
        >
          Application modules
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={filters.view === "permissions"}
          onClick={() => switchView("permissions")}
          className={cn(
            "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold transition-colors",
            filters.view === "permissions"
              ? "bg-slate-950 text-white"
              : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
          )}
        >
          Workflow permissions
        </button>
      </div>

      {filters.view === "permissions" ? (
        <div className="flex gap-3 rounded-2xl border border-cyan-200 bg-cyan-50 p-4 text-cyan-950">
          <ShieldCheck className="mt-0.5 size-5 shrink-0" />
          <div>
            <p className="text-sm font-bold">Read-only permission catalog</p>
            <p className="mt-1 text-sm leading-6">
              These codes are defined by the backend and selected when setting
              document or task transitions. Manage role access from Accounts →
              Permissions.
            </p>
          </div>
        </div>
      ) : null}

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
          <label className="relative min-w-0 flex-1">
            <span className="sr-only">
              Search {filters.view === "modules" ? "modules" : "permissions"}
            </span>
            <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
            <input
              key={`${filters.view}-${filters.search}`}
              name="search"
              defaultValue={filters.search}
              placeholder={
                filters.view === "modules"
                  ? "Search modules by code or name"
                  : "Search permissions by code or name"
              }
              className="h-11 w-full rounded-xl border border-slate-300 bg-white pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
            />
          </label>
          {filters.view === "permissions" ? (
            <Select
              ariaLabel="Filter permissions by module"
              value={filters.module}
              options={moduleOptions}
              onValueChange={(module) => void setFilters({ module, page: 1 })}
              disabled={moduleOptionsQuery.isPending}
              className="lg:w-64"
            />
          ) : null}
          <Select
            ariaLabel="Filter by active status"
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

        {!currentQuery.isPending && !currentQuery.isError ? (
          <div className="border-b border-slate-200 px-4 py-3 text-sm text-slate-600 sm:px-5">
            <span className="font-semibold text-slate-950">
              {currentData?.total_items ?? 0}
            </span>{" "}
            {filters.view === "modules" ? "modules" : "workflow permissions"}
          </div>
        ) : null}

        {currentQuery.isPending ? (
          <LoadingState />
        ) : currentQuery.isError ? (
          <ErrorState
            view={filters.view}
            error={currentQuery.error}
            retry={() => void currentQuery.refetch()}
          />
        ) : filters.view === "modules" ? (
          moduleItems.length ? (
            <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
              {moduleItems.map((item) => (
                <article
                  key={item.module_id}
                  className="flex min-h-44 flex-col rounded-xl border border-slate-200 p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h2 className="font-bold text-slate-950">{item.name}</h2>
                      <p className="mt-0.5 font-mono text-xs text-slate-500">
                        {item.code}
                      </p>
                    </div>
                    <StatusBadge tone={item.is_active ? "success" : "neutral"}>
                      {item.is_active ? "Active" : "Inactive"}
                    </StatusBadge>
                  </div>
                  <div className="mt-5 flex flex-1 items-center gap-3 rounded-xl bg-slate-50 p-3">
                    <div className="grid size-9 place-items-center rounded-lg bg-white text-slate-600 shadow-sm">
                      <span className="text-sm font-bold">
                        {item.display_order}
                      </span>
                    </div>
                    <div>
                      <p className="text-xs font-semibold tracking-wide text-slate-500 uppercase">
                        Display order
                      </p>
                      <p className="mt-0.5 text-sm text-slate-700">
                        Lower values appear first.
                      </p>
                    </div>
                  </div>
                  {canWrite ? (
                    <div className="mt-4 flex justify-end gap-1 border-t border-slate-100 pt-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setEditor({ item })}
                      >
                        <Pencil className="size-4" />
                        Edit
                      </Button>
                      {item.is_active ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-rose-700"
                          onClick={() => setDeactivateTarget(item)}
                        >
                          Deactivate
                        </Button>
                      ) : null}
                    </div>
                  ) : null}
                </article>
              ))}
            </div>
          ) : (
            <div className="px-5 py-14 text-center">
              <Blocks className="mx-auto size-9 text-slate-300" />
              <h2 className="mt-4 font-bold text-slate-950">
                No modules found
              </h2>
              <p className="mt-1 text-sm text-slate-500">
                Adjust the filters or create the first application module.
              </p>
            </div>
          )
        ) : permissionItems.length ? (
          <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
            {permissionItems.map((permission) => (
              <article
                key={permission.permission_id}
                className="flex min-h-48 flex-col rounded-xl border border-slate-200 p-4"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="grid size-9 shrink-0 place-items-center rounded-lg bg-cyan-50 text-cyan-800">
                    <KeyRound className="size-4" />
                  </div>
                  <StatusBadge
                    tone={permission.is_active ? "success" : "neutral"}
                  >
                    {permission.is_active ? "Active" : "Inactive"}
                  </StatusBadge>
                </div>
                <h2 className="mt-4 font-bold text-slate-950">
                  {permission.name}
                </h2>
                <p className="mt-1 font-mono text-xs break-all text-slate-500">
                  {permission.code}
                </p>
                <p className="mt-3 flex-1 text-sm leading-6 text-slate-600">
                  {permission.description || "No description"}
                </p>
                <div className="mt-4 border-t border-slate-100 pt-3">
                  <span className="rounded-full bg-slate-100 px-2.5 py-1 font-mono text-xs font-semibold text-slate-700">
                    {permission.module_code}
                  </span>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <div className="px-5 py-14 text-center">
            <KeyRound className="mx-auto size-9 text-slate-300" />
            <h2 className="mt-4 font-bold text-slate-950">
              No workflow permissions found
            </h2>
            <p className="mt-1 text-sm text-slate-500">
              Adjust the search, module, or active-status filter.
            </p>
          </div>
        )}

        {!currentQuery.isPending && !currentQuery.isError ? (
          <Pagination
            page={currentPage}
            totalPages={totalPages}
            onPage={(page) => void setFilters({ page })}
          />
        ) : null}
      </Panel>

      {editor ? (
        <AppModuleDialog
          item={editor.item}
          onOpenChange={(open) => {
            if (!open) setEditor(undefined);
          }}
        />
      ) : null}
      <DeactivateDialog
        item={deactivateTarget}
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
