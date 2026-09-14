"use client";

import { useEffect, useMemo, useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  Building2,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  KeyRound,
  LoaderCircle,
  Search,
  ShieldCheck,
  Trash2,
  UserRound,
  UsersRound,
  Warehouse as WarehouseIcon,
  X,
} from "lucide-react";
import Link from "next/link";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  accessScopeKeys,
  getAccount,
  grantOwnerAccess,
  grantWarehouseAccess,
  listAccounts,
  listOwnerAccess,
  listWarehouseAccess,
  revokeOwnerAccess,
  revokeWarehouseAccess,
} from "@/features/access-scopes/access-scope-api";
import type {
  OwnerAccess,
  ScopeKind,
  WarehouseAccess,
} from "@/features/access-scopes/access-scope-types";
import {
  listOrganizations,
  organizationKeys,
} from "@/features/organizations/organization-api";
import {
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 10;
const scopeKinds = ["owners", "warehouses"] as const;
const activeLookupFilters = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

const dateFormatter = new Intl.DateTimeFormat("en-ID", {
  day: "2-digit",
  month: "short",
  year: "numeric",
});

type RevokeTarget =
  | { kind: "owners"; item: OwnerAccess }
  | { kind: "warehouses"; item: WarehouseAccess };

function ErrorState({
  title,
  message,
  retry,
}: {
  title: string;
  message: string;
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
            <p className="font-semibold">{title}</p>
            <p className="mt-1 text-sm">{message}</p>
          </div>
        </div>
        <Button variant="secondary" className="mt-4" onClick={retry}>
          Try again
        </Button>
      </div>
    </div>
  );
}

function LoadingState({ label }: { label: string }) {
  return (
    <div className="grid min-h-48 place-items-center p-6 text-sm text-slate-600">
      <div className="text-center">
        <LoaderCircle className="mx-auto mb-3 size-6 animate-spin text-cyan-700" />
        {label}
      </div>
    </div>
  );
}

function RevokeAccessDialog({
  target,
  pending,
  error,
  onConfirm,
  onOpenChange,
}: {
  target?: RevokeTarget;
  pending: boolean;
  error: Error | null;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
}) {
  const label =
    target?.kind === "owners"
      ? target.item.owner_name
      : target?.item.warehouse_name;

  return (
    <Dialog.Root open={Boolean(target)} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-white p-5 shadow-2xl focus:outline-none sm:p-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                Revoke access?
              </Dialog.Title>
              <Dialog.Description className="mt-2 text-sm text-slate-600">
                This account will no longer be allowed to work with {label}.
                Existing operational records are not deleted.
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                disabled={pending}
                className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
                aria-label="Close revoke access confirmation"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>
          {error ? (
            <p
              role="alert"
              className="mt-4 rounded-xl bg-rose-50 px-4 py-3 text-sm text-rose-900"
            >
              {error.message}
            </p>
          ) : null}
          <div className="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <Dialog.Close asChild>
              <Button type="button" variant="secondary" disabled={pending}>
                Cancel
              </Button>
            </Dialog.Close>
            <Button
              type="button"
              disabled={pending}
              className="bg-rose-700 hover:bg-rose-800"
              onClick={onConfirm}
            >
              {pending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : (
                <Trash2 className="size-4" />
              )}
              {pending ? "Revoking…" : "Revoke access"}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function AccessScopesScreen({ canWrite }: { canWrite: boolean }) {
  const queryClient = useQueryClient();
  const [filters, setFilters] = useQueryStates({
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
    account: parseAsString.withDefault(""),
    scope: parseAsString.withDefault("owners"),
  });
  const [grantTarget, setGrantTarget] = useState({ accountId: "", value: "" });
  const [revokeTarget, setRevokeTarget] = useState<RevokeTarget>();
  const currentScope: ScopeKind = scopeKinds.includes(
    filters.scope as ScopeKind,
  )
    ? (filters.scope as ScopeKind)
    : "owners";
  const accountFilters = {
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };

  const accounts = useQuery({
    queryKey: accessScopeKeys.accounts(accountFilters),
    queryFn: () => listAccounts(accountFilters),
    placeholderData: keepPreviousData,
  });
  const firstAccountId = accounts.data?.items[0]?.account_id ?? "";

  useEffect(() => {
    if (!filters.account && firstAccountId) {
      void setFilters({ account: firstAccountId });
    }
  }, [filters.account, firstAccountId, setFilters]);

  const selectedAccount = useQuery({
    queryKey: accessScopeKeys.account(filters.account),
    queryFn: () => getAccount(filters.account),
    enabled: Boolean(filters.account),
  });
  const ownerAccess = useQuery({
    queryKey: accessScopeKeys.ownerAccess(filters.account),
    queryFn: () => listOwnerAccess(filters.account),
    enabled: Boolean(filters.account),
  });
  const warehouseAccess = useQuery({
    queryKey: accessScopeKeys.warehouseAccess(filters.account),
    queryFn: () => listWarehouseAccess(filters.account),
    enabled: Boolean(filters.account),
  });
  const organizations = useQuery({
    queryKey: organizationKeys.list(activeLookupFilters),
    queryFn: () => listOrganizations(activeLookupFilters),
    enabled: Boolean(filters.account) && canWrite,
  });
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(activeLookupFilters),
    queryFn: () => listWarehouses(activeLookupFilters),
    enabled: Boolean(filters.account) && canWrite,
  });

  const ownerIds = useMemo(
    () => new Set((ownerAccess.data ?? []).map((item) => item.owner_id)),
    [ownerAccess.data],
  );
  const warehouseIds = useMemo(
    () =>
      new Set((warehouseAccess.data ?? []).map((item) => item.warehouse_id)),
    [warehouseAccess.data],
  );
  const ownerOptions =
    organizations.data?.items
      .filter((item) => !ownerIds.has(item.organization_id))
      .map((item) => ({
        value: item.organization_id,
        label: `${item.name} (${item.code})`,
      })) ?? [];
  const warehouseOptions =
    warehouses.data?.items
      .filter((item) => !warehouseIds.has(item.warehouse_id))
      .map((item) => ({
        value: item.warehouse_id,
        label: `${item.name} (${item.code})`,
      })) ?? [];
  const grantSelection =
    grantTarget.accountId === filters.account ? grantTarget.value : "";

  const grant = useMutation({
    mutationFn: ({ kind, id }: { kind: ScopeKind; id: string }) =>
      kind === "owners"
        ? grantOwnerAccess(filters.account, id)
        : grantWarehouseAccess(filters.account, id),
    onSuccess: async (_, target) => {
      await queryClient.invalidateQueries({
        queryKey:
          target.kind === "owners"
            ? accessScopeKeys.ownerAccess(filters.account)
            : accessScopeKeys.warehouseAccess(filters.account),
      });
      setGrantTarget({ accountId: "", value: "" });
      toast.success(
        target.kind === "owners"
          ? "Owner access granted."
          : "Warehouse access granted.",
      );
    },
  });
  const revoke = useMutation({
    mutationFn: (target: RevokeTarget) =>
      target.kind === "owners"
        ? revokeOwnerAccess(filters.account, target.item.owner_id)
        : revokeWarehouseAccess(filters.account, target.item.warehouse_id),
    onSuccess: async (_, target) => {
      await queryClient.invalidateQueries({
        queryKey:
          target.kind === "owners"
            ? accessScopeKeys.ownerAccess(filters.account)
            : accessScopeKeys.warehouseAccess(filters.account),
      });
      toast.success(
        target.kind === "owners"
          ? "Owner access revoked."
          : "Warehouse access revoked.",
      );
      setRevokeTarget(undefined);
    },
  });

  const accountItems = accounts.data?.items ?? [];
  const totalPages = accounts.data?.total_pages ?? 0;
  const currentPage = accounts.data?.page ?? accountFilters.page;
  const activeScopeQuery =
    currentScope === "owners" ? ownerAccess : warehouseAccess;
  const grantOptions =
    currentScope === "owners" ? ownerOptions : warehouseOptions;
  const lookupPending =
    currentScope === "owners" ? organizations.isPending : warehouses.isPending;

  function changeScope(scope: ScopeKind) {
    setGrantTarget({ accountId: "", value: "" });
    grant.reset();
    void setFilters({ scope });
  }

  return (
    <div className="space-y-6">
      <header>
        <Link
          href="/master-data/core"
          className="inline-flex min-h-9 items-center gap-2 text-sm font-semibold text-slate-600 hover:text-slate-950"
        >
          <ChevronLeft className="size-4" />
          Core masters
        </Link>
        <div className="mt-3 flex items-center gap-3">
          <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <ShieldCheck className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              Access scopes
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Limit each account to the owners and warehouses it may operate.
            </p>
          </div>
        </div>
      </header>

      <div className="grid items-start gap-5 xl:grid-cols-[minmax(19rem,0.8fr)_minmax(0,1.5fr)]">
        <Panel className="overflow-hidden">
          <form
            className="border-b border-slate-200 p-4"
            onSubmit={(event) => {
              event.preventDefault();
              const data = new FormData(event.currentTarget);
              void setFilters({
                search: String(data.get("search") ?? "").trim(),
                page: 1,
                account: "",
              });
            }}
          >
            <label className="relative block">
              <span className="sr-only">Search accounts</span>
              <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
              <input
                key={filters.search}
                name="search"
                defaultValue={filters.search}
                placeholder="Search accounts"
                className="h-11 w-full rounded-xl border border-slate-300 bg-white pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
              />
            </label>
            <div className="mt-3 flex gap-2">
              <Button type="submit" variant="secondary" className="flex-1">
                Search
              </Button>
              {filters.search ? (
                <Button
                  type="button"
                  variant="ghost"
                  onClick={() =>
                    void setFilters({ search: "", page: 1, account: "" })
                  }
                >
                  Clear
                </Button>
              ) : null}
            </div>
          </form>
          <div className="flex min-h-12 items-center justify-between border-b border-slate-200 px-4 py-3">
            <p className="text-sm font-semibold text-slate-700">
              {accounts.isPending
                ? "Loading accounts…"
                : `${(accounts.data?.total_items ?? 0).toLocaleString()} accounts`}
            </p>
            {accounts.isFetching && !accounts.isPending ? (
              <LoaderCircle className="size-4 animate-spin text-slate-400" />
            ) : null}
          </div>

          {accounts.isPending ? (
            <LoadingState label="Loading accounts…" />
          ) : accounts.isError ? (
            <ErrorState
              title="Accounts could not be loaded"
              message={accounts.error.message}
              retry={() => void accounts.refetch()}
            />
          ) : accountItems.length === 0 ? (
            <div className="px-5 py-12 text-center">
              <UsersRound className="mx-auto size-9 text-slate-300" />
              <h2 className="mt-4 font-bold text-slate-950">
                No accounts found
              </h2>
              <p className="mt-1 text-sm text-slate-500">
                Adjust the account search and try again.
              </p>
            </div>
          ) : (
            <ul className="divide-y divide-slate-200">
              {accountItems.map((account) => (
                <li key={account.account_id}>
                  <button
                    type="button"
                    aria-pressed={filters.account === account.account_id}
                    onClick={() =>
                      void setFilters({ account: account.account_id })
                    }
                    className={cn(
                      "flex min-h-20 w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-slate-50",
                      filters.account === account.account_id &&
                        "bg-cyan-50 hover:bg-cyan-50",
                    )}
                  >
                    <span
                      className={cn(
                        "grid size-10 shrink-0 place-items-center rounded-full bg-slate-100 font-bold text-slate-600",
                        filters.account === account.account_id &&
                          "bg-cyan-600 text-white",
                      )}
                    >
                      {account.display_name.slice(0, 1).toUpperCase()}
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block truncate text-sm font-bold text-slate-950">
                        {account.display_name}
                      </span>
                      <span className="mt-0.5 block truncate text-xs text-slate-500">
                        @{account.username}
                      </span>
                    </span>
                    <span
                      className={cn(
                        "size-2 shrink-0 rounded-full",
                        account.status.allows_login
                          ? "bg-emerald-500"
                          : "bg-slate-300",
                      )}
                      title={account.status.name}
                    />
                  </button>
                </li>
              ))}
            </ul>
          )}

          {!accounts.isPending && !accounts.isError && totalPages > 0 ? (
            <div className="flex items-center justify-between gap-3 border-t border-slate-200 px-4 py-3">
              <p className="text-xs text-slate-600">
                Page {currentPage} of {totalPages}
              </p>
              <div className="flex gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label="Previous account page"
                  disabled={currentPage <= 1}
                  onClick={() =>
                    void setFilters({ page: currentPage - 1, account: "" })
                  }
                >
                  <ChevronLeft className="size-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label="Next account page"
                  disabled={currentPage >= totalPages}
                  onClick={() =>
                    void setFilters({ page: currentPage + 1, account: "" })
                  }
                >
                  <ChevronRight className="size-4" />
                </Button>
              </div>
            </div>
          ) : null}
        </Panel>

        <Panel className="overflow-hidden">
          {!filters.account ? (
            <div className="grid min-h-96 place-items-center px-5 py-14 text-center">
              <div>
                <UserRound className="mx-auto size-10 text-slate-300" />
                <h2 className="mt-4 font-bold text-slate-950">
                  Select an account
                </h2>
                <p className="mt-1 text-sm text-slate-500">
                  Choose an account to view and maintain its access scope.
                </p>
              </div>
            </div>
          ) : selectedAccount.isPending ? (
            <LoadingState label="Loading account scope…" />
          ) : selectedAccount.isError ? (
            <ErrorState
              title="Account could not be loaded"
              message={selectedAccount.error.message}
              retry={() => void selectedAccount.refetch()}
            />
          ) : (
            <>
              <div className="border-b border-slate-200 p-4 sm:p-5">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <p className="text-xs font-bold tracking-wide text-slate-500 uppercase">
                      Account scope
                    </p>
                    <h2 className="mt-1 text-xl font-bold text-slate-950">
                      {selectedAccount.data.display_name}
                    </h2>
                    <p className="mt-1 text-sm text-slate-500">
                      @{selectedAccount.data.username}
                      {selectedAccount.data.email
                        ? ` · ${selectedAccount.data.email}`
                        : ""}
                    </p>
                  </div>
                  <StatusBadge
                    tone={
                      selectedAccount.data.status.allows_login
                        ? "success"
                        : "neutral"
                    }
                  >
                    {selectedAccount.data.status.name}
                  </StatusBadge>
                </div>
                <div
                  className="mt-5 flex gap-2 overflow-x-auto"
                  role="tablist"
                  aria-label="Access scope types"
                >
                  <button
                    type="button"
                    role="tab"
                    aria-selected={currentScope === "owners"}
                    onClick={() => changeScope("owners")}
                    className={cn(
                      "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold",
                      currentScope === "owners"
                        ? "bg-slate-950 text-white"
                        : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
                    )}
                  >
                    Owner access ({ownerAccess.data?.length ?? 0})
                  </button>
                  <button
                    type="button"
                    role="tab"
                    aria-selected={currentScope === "warehouses"}
                    onClick={() => changeScope("warehouses")}
                    className={cn(
                      "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold",
                      currentScope === "warehouses"
                        ? "bg-slate-950 text-white"
                        : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
                    )}
                  >
                    Warehouse access ({warehouseAccess.data?.length ?? 0})
                  </button>
                </div>
              </div>

              <div className="border-b border-slate-200 bg-slate-50/70 p-4 sm:p-5">
                <div className="flex gap-3 rounded-xl border border-cyan-200 bg-cyan-50 p-3 text-sm text-cyan-950">
                  <KeyRound className="mt-0.5 size-4 shrink-0" />
                  <p>
                    Operational access requires both an owner and warehouse
                    grant, plus an active relationship between that owner and
                    warehouse.
                  </p>
                </div>
                {canWrite ? (
                  <div className="mt-4 flex flex-col gap-2 sm:flex-row">
                    <Select
                      ariaLabel={
                        currentScope === "owners"
                          ? "Owner to grant"
                          : "Warehouse to grant"
                      }
                      value={grantSelection}
                      options={grantOptions}
                      onValueChange={(value) =>
                        setGrantTarget({ accountId: filters.account, value })
                      }
                      placeholder={
                        lookupPending
                          ? `Loading ${currentScope}…`
                          : grantOptions.length === 0
                            ? `All active ${currentScope} are granted`
                            : `Select ${currentScope === "owners" ? "an owner" : "a warehouse"}`
                      }
                      disabled={lookupPending || grantOptions.length === 0}
                      className="flex-1"
                    />
                    <Button
                      disabled={!grantSelection || grant.isPending}
                      onClick={() =>
                        grant.mutate({
                          kind: currentScope,
                          id: grantSelection,
                        })
                      }
                    >
                      {grant.isPending ? (
                        <LoaderCircle className="size-4 animate-spin" />
                      ) : null}
                      {grant.isPending ? "Granting…" : "Grant access"}
                    </Button>
                  </div>
                ) : null}
                {grant.error ? (
                  <p
                    role="alert"
                    className="mt-3 rounded-xl bg-rose-50 px-4 py-3 text-sm text-rose-900"
                  >
                    {grant.error.message}
                  </p>
                ) : null}
              </div>

              {activeScopeQuery.isPending ? (
                <LoadingState label={`Loading ${currentScope} access…`} />
              ) : activeScopeQuery.isError ? (
                <ErrorState
                  title={`${currentScope === "owners" ? "Owner" : "Warehouse"} access could not be loaded`}
                  message={activeScopeQuery.error.message}
                  retry={() => void activeScopeQuery.refetch()}
                />
              ) : currentScope === "owners" ? (
                ownerAccess.data?.length ? (
                  <ul className="divide-y divide-slate-200">
                    {ownerAccess.data.map((item) => (
                      <li
                        key={item.owner_id}
                        className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5"
                      >
                        <div className="flex min-w-0 items-center gap-3">
                          <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-cyan-50 text-cyan-800">
                            <Building2 className="size-5" />
                          </span>
                          <div className="min-w-0">
                            <p className="truncate font-semibold text-slate-950">
                              {item.owner_name}
                            </p>
                            <p className="mt-0.5 text-xs text-slate-500">
                              <span className="font-mono">
                                {item.owner_code}
                              </span>{" "}
                              · Granted{" "}
                              {dateFormatter.format(new Date(item.granted_at))}
                            </p>
                          </div>
                        </div>
                        {canWrite ? (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="self-end text-rose-700 sm:self-auto"
                            onClick={() =>
                              setRevokeTarget({ kind: "owners", item })
                            }
                          >
                            Revoke
                          </Button>
                        ) : null}
                      </li>
                    ))}
                  </ul>
                ) : (
                  <div className="px-5 py-14 text-center">
                    <Building2 className="mx-auto size-9 text-slate-300" />
                    <h3 className="mt-4 font-bold text-slate-950">
                      No owner access
                    </h3>
                    <p className="mt-1 text-sm text-slate-500">
                      Grant at least one owner before this account handles
                      owner-scoped operations.
                    </p>
                  </div>
                )
              ) : warehouseAccess.data?.length ? (
                <ul className="divide-y divide-slate-200">
                  {warehouseAccess.data.map((item) => (
                    <li
                      key={item.warehouse_id}
                      className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5"
                    >
                      <div className="flex min-w-0 items-center gap-3">
                        <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-cyan-50 text-cyan-800">
                          <WarehouseIcon className="size-5" />
                        </span>
                        <div className="min-w-0">
                          <p className="truncate font-semibold text-slate-950">
                            {item.warehouse_name}
                          </p>
                          <p className="mt-0.5 text-xs text-slate-500">
                            <span className="font-mono">
                              {item.warehouse_code}
                            </span>{" "}
                            · Granted{" "}
                            {dateFormatter.format(new Date(item.granted_at))}
                          </p>
                        </div>
                      </div>
                      {canWrite ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="self-end text-rose-700 sm:self-auto"
                          onClick={() =>
                            setRevokeTarget({ kind: "warehouses", item })
                          }
                        >
                          Revoke
                        </Button>
                      ) : null}
                    </li>
                  ))}
                </ul>
              ) : (
                <div className="px-5 py-14 text-center">
                  <WarehouseIcon className="mx-auto size-9 text-slate-300" />
                  <h3 className="mt-4 font-bold text-slate-950">
                    No warehouse access
                  </h3>
                  <p className="mt-1 text-sm text-slate-500">
                    Grant at least one warehouse before this account performs
                    warehouse operations.
                  </p>
                </div>
              )}
            </>
          )}
        </Panel>
      </div>

      <RevokeAccessDialog
        target={revokeTarget}
        pending={revoke.isPending}
        error={revoke.error}
        onConfirm={() => {
          if (revokeTarget) revoke.mutate(revokeTarget);
        }}
        onOpenChange={(open) => {
          if (!open && !revoke.isPending) {
            revoke.reset();
            setRevokeTarget(undefined);
          }
        }}
      />
    </div>
  );
}
