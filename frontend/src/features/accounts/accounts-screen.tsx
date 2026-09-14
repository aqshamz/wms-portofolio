"use client";

import { useState } from "react";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  LoaderCircle,
  MoreHorizontal,
  Pencil,
  Plus,
  Search,
  UserRound,
  UsersRound,
} from "lucide-react";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  accountKeys,
  deactivateAccount,
  listAccounts,
  listAccountStatuses,
  revokeAccountSessions,
  unlockAccount,
} from "@/features/accounts/account-api";
import { AccountFormDialog } from "@/features/accounts/account-form-dialog";
import {
  type AccountActionKind,
  AccountManageDialog,
  AccountStatusDialog,
  ConfirmAccountActionDialog,
  PasswordResetDialog,
} from "@/features/accounts/account-security-dialogs";
import type { AccountSummary } from "@/features/accounts/account-types";

const PAGE_SIZE = 10;

const dateFormatter = new Intl.DateTimeFormat("en-ID", {
  day: "2-digit",
  month: "short",
  year: "numeric",
});

type DialogState =
  | { kind: "create" }
  | { kind: "edit"; account: AccountSummary }
  | { kind: "manage"; account: AccountSummary }
  | { kind: "status"; account: AccountSummary }
  | { kind: "password"; account: AccountSummary }
  | {
      kind: "action";
      action: AccountActionKind;
      account: AccountSummary;
    }
  | null;

function statusTone(account: AccountSummary) {
  if (account.status.code === "ACTIVE") return "success" as const;
  if (account.status.code === "LOCKED") return "danger" as const;
  if (account.status.code === "PENDING") return "warning" as const;
  return "neutral" as const;
}

function isLocked(account: AccountSummary) {
  return (
    account.status.code === "LOCKED" ||
    Boolean(account.locked_until && new Date(account.locked_until) > new Date())
  );
}

function AccountActions({
  canWrite,
  onEdit,
  onManage,
}: {
  canWrite: boolean;
  onEdit: () => void;
  onManage: () => void;
}) {
  return (
    <div className="flex flex-wrap justify-end gap-1">
      {canWrite ? (
        <Button variant="ghost" size="sm" onClick={onEdit}>
          <Pencil className="size-4" />
          Edit
        </Button>
      ) : null}
      <Button variant="ghost" size="sm" onClick={onManage}>
        <MoreHorizontal className="size-4" />
        {canWrite ? "Manage" : "Details"}
      </Button>
    </div>
  );
}

export function AccountsScreen({
  canWrite,
  currentAccountId,
}: {
  canWrite: boolean;
  currentAccountId: string;
}) {
  const queryClient = useQueryClient();
  const [dialog, setDialog] = useState<DialogState>(null);
  const [filters, setFilters] = useQueryStates({
    search: parseAsString.withDefault(""),
    status: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
  });
  const queryFilters = {
    search: filters.search,
    statusId: filters.status,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };
  const accounts = useQuery({
    queryKey: accountKeys.list(queryFilters),
    queryFn: () => listAccounts(queryFilters),
    placeholderData: keepPreviousData,
  });
  const statuses = useQuery({
    queryKey: accountKeys.statuses(),
    queryFn: listAccountStatuses,
  });
  const action = useMutation({
    mutationFn: async (target: {
      kind: AccountActionKind;
      account: AccountSummary;
    }) => {
      if (target.kind === "unlock") {
        await unlockAccount(
          target.account.account_id,
          target.account.version_no,
        );
        return;
      }
      if (target.kind === "deactivate") {
        await deactivateAccount(
          target.account.account_id,
          target.account.version_no,
        );
        return;
      }
      await revokeAccountSessions(target.account.account_id);
    },
    onSuccess: async (_, target) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: accountKeys.lists() }),
        queryClient.invalidateQueries({
          queryKey: accountKeys.detail(target.account.account_id),
        }),
      ]);
      toast.success(
        target.kind === "unlock"
          ? "Account unlocked."
          : target.kind === "deactivate"
            ? "Account deactivated."
            : "Active sessions revoked.",
      );
      setDialog(null);
    },
  });

  const items = accounts.data?.items ?? [];
  const totalItems = accounts.data?.total_items ?? 0;
  const totalPages = accounts.data?.total_pages ?? 0;
  const currentPage = accounts.data?.page ?? queryFilters.page;
  const statusOptions = [
    { value: "all", label: "All statuses" },
    ...(statuses.data?.map((status) => ({
      value: status.account_status_id,
      label: status.name,
    })) ?? []),
  ];

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="flex items-center gap-3">
          <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <UsersRound className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              Accounts
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Manage login identities, authentication state, and sessions.
            </p>
          </div>
        </div>
        {canWrite ? (
          <Button
            className="w-full sm:w-auto"
            onClick={() => setDialog({ kind: "create" })}
          >
            <Plus className="size-4" />
            Create account
          </Button>
        ) : null}
      </header>

      <Panel className="overflow-hidden">
        <form
          className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:p-5 md:flex-row"
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
            <span className="sr-only">Search accounts</span>
            <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
            <input
              key={filters.search}
              name="search"
              defaultValue={filters.search}
              placeholder="Search by username, name, or email"
              className="h-11 w-full rounded-xl border border-slate-300 bg-white pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
            />
          </label>
          <Select
            ariaLabel="Filter accounts by status"
            value={filters.status || "all"}
            options={statusOptions}
            onValueChange={(status) =>
              void setFilters({
                status: status === "all" ? "" : status,
                page: 1,
              })
            }
            disabled={statuses.isPending || statuses.isError}
            className="md:w-48"
          />
          <Button type="submit" variant="secondary">
            Search
          </Button>
          {filters.search || filters.status ? (
            <Button
              type="button"
              variant="ghost"
              onClick={() =>
                void setFilters({ search: "", status: "", page: 1 })
              }
            >
              Clear
            </Button>
          ) : null}
        </form>

        <div className="flex min-h-12 items-center justify-between gap-3 border-b border-slate-200 px-4 py-3 text-sm sm:px-5">
          <p className="font-semibold text-slate-700">
            {accounts.isPending
              ? "Loading accounts…"
              : `${totalItems.toLocaleString()} account${totalItems === 1 ? "" : "s"}`}
          </p>
          {accounts.isFetching && !accounts.isPending ? (
            <span className="inline-flex items-center gap-2 text-xs text-slate-500">
              <LoaderCircle className="size-3.5 animate-spin" />
              Refreshing
            </span>
          ) : null}
        </div>

        {accounts.isPending ? (
          <div className="space-y-3 p-4 sm:p-5">
            {Array.from({ length: 5 }, (_, index) => (
              <div
                key={index}
                className="h-16 animate-pulse rounded-xl bg-slate-100"
              />
            ))}
          </div>
        ) : accounts.isError ? (
          <div className="p-5">
            <div
              role="alert"
              className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900"
            >
              <div className="flex gap-3">
                <CircleAlert className="mt-0.5 size-5 shrink-0" />
                <div>
                  <p className="font-semibold">Accounts could not be loaded</p>
                  <p className="mt-1 text-sm">{accounts.error.message}</p>
                </div>
              </div>
              <Button
                variant="secondary"
                className="mt-4"
                onClick={() => void accounts.refetch()}
              >
                Try again
              </Button>
            </div>
          </div>
        ) : items.length === 0 ? (
          <div className="px-5 py-14 text-center">
            <UserRound className="mx-auto size-9 text-slate-300" />
            <h2 className="mt-4 font-bold text-slate-950">No accounts found</h2>
            <p className="mt-1 text-sm text-slate-500">
              {filters.search || filters.status
                ? "Adjust the search or status filter."
                : "Create the first managed account."}
            </p>
          </div>
        ) : (
          <>
            <div className="hidden overflow-x-auto lg:block">
              <table className="w-full border-collapse text-left">
                <thead>
                  <tr className="bg-slate-50 text-xs font-bold tracking-wide text-slate-500 uppercase">
                    <th className="px-5 py-3">Account</th>
                    <th className="px-5 py-3">Status</th>
                    <th className="px-5 py-3">Authentication</th>
                    <th className="px-5 py-3">Last login</th>
                    <th className="px-5 py-3">Security</th>
                    <th className="px-5 py-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200">
                  {items.map((account) => (
                    <tr
                      key={account.account_id}
                      className="hover:bg-slate-50/70"
                    >
                      <td className="px-5 py-4">
                        <p className="font-semibold text-slate-950">
                          {account.display_name}
                        </p>
                        <p className="mt-0.5 text-xs text-slate-500">
                          @{account.username}
                          {account.email ? ` · ${account.email}` : ""}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <StatusBadge tone={statusTone(account)}>
                          {account.status.name}
                        </StatusBadge>
                      </td>
                      <td className="px-5 py-4 text-sm text-slate-600">
                        <p>
                          {account.authentication_policy_code ||
                            "Default policy"}
                        </p>
                        <p className="mt-0.5 text-xs text-slate-500">
                          {account.preferred_timezone || "No timezone"}
                        </p>
                      </td>
                      <td className="px-5 py-4 text-sm text-slate-600">
                        {account.last_login_at
                          ? dateFormatter.format(
                              new Date(account.last_login_at),
                            )
                          : "Never"}
                      </td>
                      <td className="px-5 py-4 text-sm text-slate-600">
                        {isLocked(account) ? (
                          <span className="font-semibold text-rose-700">
                            Locked
                          </span>
                        ) : account.failed_login_count ? (
                          `${account.failed_login_count} failed attempts`
                        ) : (
                          "Clear"
                        )}
                      </td>
                      <td className="px-5 py-4">
                        <AccountActions
                          canWrite={canWrite}
                          onEdit={() => setDialog({ kind: "edit", account })}
                          onManage={() =>
                            setDialog({ kind: "manage", account })
                          }
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <ul className="divide-y divide-slate-200 lg:hidden">
              {items.map((account) => (
                <li key={account.account_id} className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="truncate font-semibold text-slate-950">
                        {account.display_name}
                      </p>
                      <p className="mt-0.5 truncate text-xs text-slate-500">
                        @{account.username}
                        {account.email ? ` · ${account.email}` : ""}
                      </p>
                    </div>
                    <StatusBadge tone={statusTone(account)}>
                      {account.status.name}
                    </StatusBadge>
                  </div>
                  <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
                    <div>
                      <dt className="text-xs font-semibold text-slate-500">
                        Policy
                      </dt>
                      <dd className="mt-1 text-slate-700">
                        {account.authentication_policy_code || "Default"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-xs font-semibold text-slate-500">
                        Last login
                      </dt>
                      <dd className="mt-1 text-slate-700">
                        {account.last_login_at
                          ? dateFormatter.format(
                              new Date(account.last_login_at),
                            )
                          : "Never"}
                      </dd>
                    </div>
                  </dl>
                  <div className="mt-3 border-t border-slate-100 pt-2">
                    <AccountActions
                      canWrite={canWrite}
                      onEdit={() => setDialog({ kind: "edit", account })}
                      onManage={() => setDialog({ kind: "manage", account })}
                    />
                  </div>
                </li>
              ))}
            </ul>
          </>
        )}

        {!accounts.isPending && !accounts.isError && totalPages > 0 ? (
          <div className="flex flex-col gap-3 border-t border-slate-200 px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5">
            <p className="text-sm text-slate-600">
              Page <span className="font-semibold">{currentPage}</span> of{" "}
              {totalPages}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={currentPage <= 1}
                onClick={() => void setFilters({ page: currentPage - 1 })}
              >
                <ChevronLeft className="size-4" />
                Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={currentPage >= totalPages}
                onClick={() => void setFilters({ page: currentPage + 1 })}
              >
                Next
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </div>
        ) : null}
      </Panel>

      {dialog?.kind === "create" || dialog?.kind === "edit" ? (
        <AccountFormDialog
          open
          accountId={
            dialog.kind === "edit" ? dialog.account.account_id : undefined
          }
          onOpenChange={(open) => {
            if (!open) setDialog(null);
          }}
        />
      ) : null}
      {dialog?.kind === "manage" ? (
        <AccountManageDialog
          account={dialog.account}
          currentAccountId={currentAccountId}
          canWrite={canWrite}
          onEdit={(account) => setDialog({ kind: "edit", account })}
          onStatus={(account) => setDialog({ kind: "status", account })}
          onPassword={(account) => setDialog({ kind: "password", account })}
          onAction={(kind, account) =>
            setDialog({ kind: "action", action: kind, account })
          }
          onOpenChange={(open) => {
            if (!open) setDialog(null);
          }}
        />
      ) : null}
      {dialog?.kind === "status" ? (
        <AccountStatusDialog
          key={dialog.account.account_id}
          account={dialog.account}
          onOpenChange={(open) => {
            if (!open) setDialog(null);
          }}
        />
      ) : null}
      {dialog?.kind === "password" ? (
        <PasswordResetDialog
          account={dialog.account}
          onOpenChange={(open) => {
            if (!open) setDialog(null);
          }}
        />
      ) : null}
      {dialog?.kind === "action" ? (
        <ConfirmAccountActionDialog
          kind={dialog.action}
          account={dialog.account}
          pending={action.isPending}
          error={action.error}
          onConfirm={() =>
            action.mutate({ kind: dialog.action, account: dialog.account })
          }
          onOpenChange={(open) => {
            if (!open && !action.isPending) {
              action.reset();
              setDialog(null);
            }
          }}
        />
      ) : null}
    </div>
  );
}
