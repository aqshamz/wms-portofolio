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
  X,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  accountKeys,
  getAccount,
  listAccounts,
} from "@/features/accounts/account-api";
import type {
  AccountPermission,
  AccountRole,
} from "@/features/accounts/account-types";
import {
  assignAccountRole,
  grantAccountPermission,
  listPermissions,
  listRoles,
  permissionKeys,
  revokeAccountPermission,
  revokeAccountRole,
} from "@/features/permissions/permission-api";
import type { Permission } from "@/features/permissions/permission-types";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 10;
const activeRoleFilters = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
const assignmentViews = ["roles", "direct", "effective"] as const;
type AssignmentView = (typeof assignmentViews)[number];

type RevokeTarget =
  | { kind: "role"; item: AccountRole }
  | { kind: "permission"; item: AccountPermission };

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

function RevokeDialog({
  target,
  pending,
  error,
  onConfirm,
  onOpenChange,
}: {
  target: RevokeTarget;
  pending: boolean;
  error: Error | null;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
}) {
  const label = target.kind === "role" ? target.item.name : target.item.name;
  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!pending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-white p-5 shadow-2xl focus:outline-none sm:p-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                Revoke {target.kind}?
              </Dialog.Title>
              <Dialog.Description className="mt-2 text-sm text-slate-600">
                Remove {label} from this account. Effective permissions may
                change immediately.
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close revoke confirmation"
                className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>
          <p className="mt-4 rounded-xl bg-amber-50 px-4 py-3 text-xs text-amber-900">
            The system will reject this change if it would remove the final
            active security administrator.
          </p>
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
              <Button type="button" variant="secondary">
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
              {pending ? "Revoking…" : "Revoke"}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function AccountPermissionAssignments({
  canWrite,
  currentAccountId,
}: {
  canWrite: boolean;
  currentAccountId: string;
}) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [filters, setFilters] = useQueryStates({
    accountSearch: parseAsString.withDefault(""),
    accountPage: parseAsInteger.withDefault(1),
    account: parseAsString.withDefault(""),
    assignment: parseAsString.withDefault("roles"),
  });
  const [grantTarget, setGrantTarget] = useState({
    accountId: "",
    view: "",
    value: "",
  });
  const [revokeTarget, setRevokeTarget] = useState<RevokeTarget>();
  const currentView: AssignmentView = assignmentViews.includes(
    filters.assignment as AssignmentView,
  )
    ? (filters.assignment as AssignmentView)
    : "roles";
  const accountFilters = {
    search: filters.accountSearch,
    statusId: "",
    page: Math.max(1, filters.accountPage),
    pageSize: PAGE_SIZE,
  };
  const accounts = useQuery({
    queryKey: accountKeys.list(accountFilters),
    queryFn: () => listAccounts(accountFilters),
    placeholderData: keepPreviousData,
  });
  const firstAccountId = accounts.data?.items[0]?.account_id ?? "";

  useEffect(() => {
    if (!filters.account && firstAccountId) {
      void setFilters({ account: firstAccountId });
    }
  }, [filters.account, firstAccountId, setFilters]);

  const account = useQuery({
    queryKey: accountKeys.detail(filters.account),
    queryFn: () => getAccount(filters.account),
    enabled: Boolean(filters.account),
  });
  const roles = useQuery({
    queryKey: permissionKeys.roleList(activeRoleFilters),
    queryFn: () => listRoles(activeRoleFilters),
    enabled: Boolean(filters.account),
  });
  const permissions = useQuery({
    queryKey: permissionKeys.available(),
    queryFn: listPermissions,
    enabled: Boolean(filters.account),
  });

  const assignedRoleIds = useMemo(
    () => new Set((account.data?.roles ?? []).map((item) => item.role_id)),
    [account.data?.roles],
  );
  const directPermissionIds = useMemo(
    () =>
      new Set(
        (account.data?.direct_permissions ?? []).map(
          (item) => item.permission_id,
        ),
      ),
    [account.data?.direct_permissions],
  );
  const roleOptions =
    roles.data?.items
      .filter((role) => !assignedRoleIds.has(role.role_id))
      .map((role) => ({
        value: role.role_id,
        label: `${role.name} (${role.code})`,
      })) ?? [];
  const permissionOptions =
    permissions.data
      ?.filter((item) => !directPermissionIds.has(item.permission_id))
      .map((item) => ({
        value: item.permission_id,
        label: `${item.name} (${item.code})`,
      })) ?? [];
  const selection =
    grantTarget.accountId === filters.account &&
    grantTarget.view === currentView
      ? grantTarget.value
      : "";

  async function refreshAccount() {
    await queryClient.invalidateQueries({
      queryKey: accountKeys.detail(filters.account),
    });
    if (filters.account === currentAccountId) router.refresh();
  }

  const grant = useMutation({
    mutationFn: ({ kind, id }: { kind: "role" | "permission"; id: string }) =>
      kind === "role"
        ? assignAccountRole(filters.account, id)
        : grantAccountPermission(filters.account, id),
    onSuccess: async (_, target) => {
      await refreshAccount();
      setGrantTarget({ accountId: "", view: "", value: "" });
      toast.success(
        target.kind === "role" ? "Role assigned." : "Permission granted.",
      );
    },
  });
  const revoke = useMutation({
    mutationFn: (target: RevokeTarget) =>
      target.kind === "role"
        ? revokeAccountRole(filters.account, target.item.role_id)
        : revokeAccountPermission(filters.account, target.item.permission_id),
    onSuccess: async (_, target) => {
      await refreshAccount();
      toast.success(
        target.kind === "role" ? "Role revoked." : "Permission revoked.",
      );
      setRevokeTarget(undefined);
    },
  });

  const accountItems = accounts.data?.items ?? [];
  const totalPages = accounts.data?.total_pages ?? 0;
  const currentPage = accounts.data?.page ?? accountFilters.page;
  const effectivePermissions = useMemo(() => {
    const catalog = new Map(
      (permissions.data ?? []).map((item) => [item.code, item]),
    );
    return (account.data?.effective_permissions ?? []).map(
      (code): Permission =>
        catalog.get(code) ?? {
          permission_id: code,
          code,
          name: code,
          module_code: code.split(".")[0] || "OTHER",
        },
    );
  }, [account.data?.effective_permissions, permissions.data]);
  const effectiveGroups = effectivePermissions.reduce((groups, item) => {
    const values = groups.get(item.module_code) ?? [];
    values.push(item);
    groups.set(item.module_code, values);
    return groups;
  }, new Map<string, Permission[]>());

  function permissionSources(permission: Permission) {
    const sources: string[] = [];
    if (
      account.data?.direct_permissions.some(
        (item) => item.permission_id === permission.permission_id,
      )
    ) {
      sources.push("Direct");
    }
    for (const role of account.data?.roles ?? []) {
      const definition = roles.data?.items.find(
        (item) => item.role_id === role.role_id,
      );
      if (
        role.is_active &&
        definition?.permissions.some(
          (item) => item.permission_id === permission.permission_id,
        )
      ) {
        sources.push(role.name);
      }
    }
    return sources.join(", ") || "Role derived";
  }

  return (
    <div className="grid items-start gap-5 xl:grid-cols-[minmax(19rem,0.8fr)_minmax(0,1.5fr)]">
      <Panel className="overflow-hidden">
        <form
          className="border-b border-slate-200 p-4"
          onSubmit={(event) => {
            event.preventDefault();
            const data = new FormData(event.currentTarget);
            void setFilters({
              accountSearch: String(
                data.get("accountPermissionSearch") ?? "",
              ).trim(),
              accountPage: 1,
              account: "",
            });
          }}
        >
          <label className="relative block">
            <span className="sr-only">Search accounts</span>
            <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
            <input
              key={filters.accountSearch}
              name="accountPermissionSearch"
              defaultValue={filters.accountSearch}
              placeholder="Search accounts"
              className="h-11 w-full rounded-xl border border-slate-300 bg-white pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
            />
          </label>
          <div className="mt-3 flex gap-2">
            <Button type="submit" variant="secondary" className="flex-1">
              Search
            </Button>
            {filters.accountSearch ? (
              <Button
                type="button"
                variant="ghost"
                onClick={() =>
                  void setFilters({
                    accountSearch: "",
                    accountPage: 1,
                    account: "",
                  })
                }
              >
                Clear
              </Button>
            ) : null}
          </div>
        </form>
        <div className="flex min-h-12 items-center justify-between border-b border-slate-200 px-4 py-3 text-sm">
          <p className="font-semibold text-slate-700">
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
          <div className="p-4">
            <p className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
              {accounts.error.message}
            </p>
          </div>
        ) : accountItems.length === 0 ? (
          <div className="px-5 py-12 text-center">
            <UsersRound className="mx-auto size-9 text-slate-300" />
            <h2 className="mt-4 font-bold text-slate-950">No accounts found</h2>
          </div>
        ) : (
          <ul className="divide-y divide-slate-200">
            {accountItems.map((item) => (
              <li key={item.account_id}>
                <button
                  type="button"
                  aria-pressed={filters.account === item.account_id}
                  onClick={() => void setFilters({ account: item.account_id })}
                  className={cn(
                    "flex min-h-20 w-full items-center gap-3 px-4 py-3 text-left hover:bg-slate-50",
                    filters.account === item.account_id &&
                      "bg-cyan-50 hover:bg-cyan-50",
                  )}
                >
                  <span
                    className={cn(
                      "grid size-10 shrink-0 place-items-center rounded-full bg-slate-100 font-bold text-slate-600",
                      filters.account === item.account_id &&
                        "bg-cyan-600 text-white",
                    )}
                  >
                    {item.display_name.slice(0, 1).toUpperCase()}
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-sm font-bold text-slate-950">
                      {item.display_name}
                    </span>
                    <span className="mt-0.5 block truncate text-xs text-slate-500">
                      @{item.username}
                    </span>
                  </span>
                </button>
              </li>
            ))}
          </ul>
        )}
        {!accounts.isPending && !accounts.isError && totalPages > 0 ? (
          <div className="flex items-center justify-between border-t border-slate-200 px-4 py-3">
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
                  void setFilters({
                    accountPage: currentPage - 1,
                    account: "",
                  })
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
                  void setFilters({
                    accountPage: currentPage + 1,
                    account: "",
                  })
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
                Choose an account to inspect its effective permissions.
              </p>
            </div>
          </div>
        ) : account.isPending ? (
          <LoadingState label="Loading assignments…" />
        ) : account.isError ? (
          <div className="p-5">
            <div
              role="alert"
              className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900"
            >
              <div className="flex gap-3">
                <CircleAlert className="mt-0.5 size-5 shrink-0" />
                <p className="text-sm">{account.error.message}</p>
              </div>
            </div>
          </div>
        ) : (
          <>
            <div className="border-b border-slate-200 p-4 sm:p-5">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-xs font-bold tracking-wide text-slate-500 uppercase">
                    Account permissions
                  </p>
                  <h2 className="mt-1 text-xl font-bold text-slate-950">
                    {account.data.display_name}
                  </h2>
                  <p className="mt-1 text-sm text-slate-500">
                    @{account.data.username}
                  </p>
                </div>
                <StatusBadge
                  tone={
                    account.data.status.allows_login ? "success" : "neutral"
                  }
                >
                  {account.data.status.name}
                </StatusBadge>
              </div>
              <div
                className="mt-5 flex gap-2 overflow-x-auto"
                role="tablist"
                aria-label="Account permission assignments"
              >
                {assignmentViews.map((view) => (
                  <button
                    key={view}
                    type="button"
                    role="tab"
                    aria-selected={currentView === view}
                    onClick={() => {
                      grant.reset();
                      setGrantTarget({ accountId: "", view: "", value: "" });
                      void setFilters({ assignment: view });
                    }}
                    className={cn(
                      "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold",
                      currentView === view
                        ? "bg-slate-950 text-white"
                        : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
                    )}
                  >
                    {view === "roles"
                      ? `Roles (${account.data.roles.length})`
                      : view === "direct"
                        ? `Direct (${account.data.direct_permissions.length})`
                        : `Effective (${account.data.effective_permissions.length})`}
                  </button>
                ))}
              </div>
            </div>

            {currentView !== "effective" && canWrite ? (
              <div className="border-b border-slate-200 bg-slate-50/70 p-4 sm:p-5">
                <div className="flex flex-col gap-2 sm:flex-row">
                  <Select
                    ariaLabel={
                      currentView === "roles"
                        ? "Role to assign"
                        : "Direct permission to grant"
                    }
                    value={selection}
                    options={
                      currentView === "roles" ? roleOptions : permissionOptions
                    }
                    onValueChange={(value) =>
                      setGrantTarget({
                        accountId: filters.account,
                        view: currentView,
                        value,
                      })
                    }
                    placeholder={
                      currentView === "roles"
                        ? roles.isPending
                          ? "Loading active roles…"
                          : roles.isError
                            ? "Active roles could not be loaded"
                            : roleOptions.length
                              ? "Select a role"
                              : "All active roles assigned"
                        : permissions.isPending
                          ? "Loading permissions…"
                          : permissions.isError
                            ? "Permissions could not be loaded"
                            : permissionOptions.length
                              ? "Select a permission"
                              : "All permissions granted directly"
                    }
                    disabled={
                      currentView === "roles"
                        ? roles.isPending ||
                          roles.isError ||
                          roleOptions.length === 0
                        : permissions.isPending ||
                          permissions.isError ||
                          permissionOptions.length === 0
                    }
                    className="flex-1"
                  />
                  <Button
                    disabled={!selection || grant.isPending}
                    onClick={() =>
                      grant.mutate({
                        kind: currentView === "roles" ? "role" : "permission",
                        id: selection,
                      })
                    }
                  >
                    {grant.isPending ? (
                      <LoaderCircle className="size-4 animate-spin" />
                    ) : null}
                    {currentView === "roles" ? "Assign role" : "Grant directly"}
                  </Button>
                </div>
                {grant.error ? (
                  <p className="mt-3 rounded-xl bg-rose-50 p-3 text-sm text-rose-900">
                    {grant.error.message}
                  </p>
                ) : null}
                {currentView === "roles" && roles.isError ? (
                  <div
                    role="alert"
                    className="mt-3 flex flex-col gap-2 rounded-xl bg-rose-50 p-3 text-sm text-rose-900 sm:flex-row sm:items-center sm:justify-between"
                  >
                    <span>{roles.error.message}</span>
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => void roles.refetch()}
                    >
                      Retry roles
                    </Button>
                  </div>
                ) : null}
                {currentView === "direct" && permissions.isError ? (
                  <div
                    role="alert"
                    className="mt-3 flex flex-col gap-2 rounded-xl bg-rose-50 p-3 text-sm text-rose-900 sm:flex-row sm:items-center sm:justify-between"
                  >
                    <span>{permissions.error.message}</span>
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => void permissions.refetch()}
                    >
                      Retry permissions
                    </Button>
                  </div>
                ) : null}
              </div>
            ) : null}

            {currentView === "roles" ? (
              account.data.roles.length ? (
                <ul className="divide-y divide-slate-200">
                  {account.data.roles.map((role) => (
                    <li
                      key={role.role_id}
                      className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5"
                    >
                      <div>
                        <div className="flex flex-wrap items-center gap-2">
                          <p className="font-semibold text-slate-950">
                            {role.name}
                          </p>
                          {!role.is_active ? (
                            <StatusBadge tone="neutral">Inactive</StatusBadge>
                          ) : null}
                        </div>
                        <p className="mt-0.5 font-mono text-xs text-slate-500">
                          {role.code}
                        </p>
                      </div>
                      {canWrite ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="self-end text-rose-700 sm:self-auto"
                          onClick={() =>
                            setRevokeTarget({ kind: "role", item: role })
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
                  <ShieldCheck className="mx-auto size-9 text-slate-300" />
                  <h3 className="mt-4 font-bold text-slate-950">
                    No roles assigned
                  </h3>
                </div>
              )
            ) : currentView === "direct" ? (
              account.data.direct_permissions.length ? (
                <ul className="divide-y divide-slate-200">
                  {account.data.direct_permissions.map((permission) => (
                    <li
                      key={permission.permission_id}
                      className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5"
                    >
                      <div>
                        <p className="font-semibold text-slate-950">
                          {permission.name}
                        </p>
                        <p className="mt-0.5 font-mono text-xs text-slate-500">
                          {permission.code}
                        </p>
                      </div>
                      {canWrite ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="self-end text-rose-700 sm:self-auto"
                          onClick={() =>
                            setRevokeTarget({
                              kind: "permission",
                              item: permission,
                            })
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
                  <KeyRound className="mx-auto size-9 text-slate-300" />
                  <h3 className="mt-4 font-bold text-slate-950">
                    No direct permissions
                  </h3>
                  <p className="mt-1 text-sm text-slate-500">
                    Prefer roles; use direct grants only for exceptions.
                  </p>
                </div>
              )
            ) : effectivePermissions.length ? (
              <div className="space-y-4 p-4 sm:p-5">
                {[...effectiveGroups.entries()].map(([module, values]) => (
                  <section
                    key={module}
                    className="overflow-hidden rounded-xl border border-slate-200"
                  >
                    <h3 className="border-b border-slate-100 bg-slate-50 px-3 py-2 text-xs font-bold tracking-wide text-slate-700 uppercase">
                      {module}
                    </h3>
                    <ul className="divide-y divide-slate-100">
                      {values.map((permission) => (
                        <li
                          key={permission.code}
                          className="flex flex-col gap-1 px-3 py-3 sm:flex-row sm:items-center sm:justify-between"
                        >
                          <div>
                            <p className="text-sm font-semibold text-slate-900">
                              {permission.name}
                            </p>
                            <p className="font-mono text-[11px] text-slate-500">
                              {permission.code}
                            </p>
                          </div>
                          <p className="text-xs text-slate-500">
                            Via {permissionSources(permission)}
                          </p>
                        </li>
                      ))}
                    </ul>
                  </section>
                ))}
              </div>
            ) : (
              <div className="px-5 py-14 text-center">
                <KeyRound className="mx-auto size-9 text-slate-300" />
                <h3 className="mt-4 font-bold text-slate-950">
                  No effective permissions
                </h3>
              </div>
            )}
          </>
        )}
      </Panel>

      {revokeTarget ? (
        <RevokeDialog
          target={revokeTarget}
          pending={revoke.isPending}
          error={revoke.error}
          onConfirm={() => revoke.mutate(revokeTarget)}
          onOpenChange={(open) => {
            if (!open && !revoke.isPending) {
              revoke.reset();
              setRevokeTarget(undefined);
            }
          }}
        />
      ) : null}
    </div>
  );
}
