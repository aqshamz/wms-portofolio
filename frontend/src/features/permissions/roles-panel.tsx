"use client";

import { useState } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import {
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  KeyRound,
  LoaderCircle,
  Pencil,
  Plus,
  Search,
  ShieldCheck,
} from "lucide-react";
import {
  parseAsInteger,
  parseAsString,
  parseAsStringLiteral,
  useQueryStates,
} from "nuqs";

import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  listRoles,
  permissionKeys,
} from "@/features/permissions/permission-api";
import {
  RoleFormDialog,
  RolePermissionsDialog,
  RoleStatusDialog,
} from "@/features/permissions/role-dialogs";
import type {
  ActiveFilter,
  Role,
} from "@/features/permissions/permission-types";

const PAGE_SIZE = 10;
const activeFilters = ["all", "active", "inactive"] as const;
const activeOptions = [
  { value: "all", label: "All statuses" },
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
] as const;

type Editor =
  | { kind: "create" }
  | { kind: "edit"; role: Role }
  | { kind: "permissions"; role: Role }
  | { kind: "status"; role: Role }
  | null;

export function RolesPanel({ canWrite }: { canWrite: boolean }) {
  const [editor, setEditor] = useState<Editor>(null);
  const [filters, setFilters] = useQueryStates({
    roleSearch: parseAsString.withDefault(""),
    roleActive: parseAsStringLiteral(activeFilters).withDefault("all"),
    rolePage: parseAsInteger.withDefault(1),
  });
  const queryFilters = {
    search: filters.roleSearch,
    active: filters.roleActive as ActiveFilter,
    page: Math.max(1, filters.rolePage),
    pageSize: PAGE_SIZE,
  };
  const roles = useQuery({
    queryKey: permissionKeys.roleList(queryFilters),
    queryFn: () => listRoles(queryFilters),
    placeholderData: keepPreviousData,
  });
  const items = roles.data?.items ?? [];
  const totalItems = roles.data?.total_items ?? 0;
  const totalPages = roles.data?.total_pages ?? 0;
  const currentPage = roles.data?.page ?? queryFilters.page;

  return (
    <>
      <Panel className="overflow-hidden">
        <div className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:items-center sm:p-5">
          <form
            className="flex flex-1 flex-col gap-3 md:flex-row"
            onSubmit={(event) => {
              event.preventDefault();
              const data = new FormData(event.currentTarget);
              void setFilters({
                roleSearch: String(data.get("roleSearch") ?? "").trim(),
                rolePage: 1,
              });
            }}
          >
            <label className="relative flex-1">
              <span className="sr-only">Search roles</span>
              <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
              <input
                key={filters.roleSearch}
                name="roleSearch"
                defaultValue={filters.roleSearch}
                placeholder="Search roles"
                className="h-11 w-full rounded-xl border border-slate-300 bg-white pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
              />
            </label>
            <Select
              ariaLabel="Filter roles by status"
              value={filters.roleActive}
              options={activeOptions}
              onValueChange={(roleActive) =>
                void setFilters({ roleActive, rolePage: 1 })
              }
              className="md:w-40"
            />
            <Button type="submit" variant="secondary">
              Search
            </Button>
            {filters.roleSearch || filters.roleActive !== "all" ? (
              <Button
                type="button"
                variant="ghost"
                onClick={() =>
                  void setFilters({
                    roleSearch: "",
                    roleActive: "all",
                    rolePage: 1,
                  })
                }
              >
                Clear
              </Button>
            ) : null}
          </form>
          {canWrite ? (
            <Button onClick={() => setEditor({ kind: "create" })}>
              <Plus className="size-4" />
              Create role
            </Button>
          ) : null}
        </div>

        <div className="flex min-h-12 items-center justify-between border-b border-slate-200 px-4 py-3 text-sm sm:px-5">
          <p className="font-semibold text-slate-700">
            {roles.isPending
              ? "Loading roles…"
              : `${totalItems.toLocaleString()} role${totalItems === 1 ? "" : "s"}`}
          </p>
          {roles.isFetching && !roles.isPending ? (
            <LoaderCircle className="size-4 animate-spin text-slate-400" />
          ) : null}
        </div>

        {roles.isPending ? (
          <div className="space-y-3 p-4 sm:p-5">
            {Array.from({ length: 4 }, (_, index) => (
              <div
                key={index}
                className="h-20 animate-pulse rounded-xl bg-slate-100"
              />
            ))}
          </div>
        ) : roles.isError ? (
          <div className="p-5">
            <div
              role="alert"
              className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900"
            >
              <div className="flex gap-3">
                <CircleAlert className="mt-0.5 size-5 shrink-0" />
                <div>
                  <p className="font-semibold">Roles could not be loaded</p>
                  <p className="mt-1 text-sm">{roles.error.message}</p>
                </div>
              </div>
              <Button
                variant="secondary"
                className="mt-4"
                onClick={() => void roles.refetch()}
              >
                Try again
              </Button>
            </div>
          </div>
        ) : items.length === 0 ? (
          <div className="px-5 py-14 text-center">
            <ShieldCheck className="mx-auto size-9 text-slate-300" />
            <h2 className="mt-4 font-bold text-slate-950">No roles found</h2>
            <p className="mt-1 text-sm text-slate-500">
              Adjust the filters or create a reusable permission role.
            </p>
          </div>
        ) : (
          <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
            {items.map((role) => (
              <article
                key={role.role_id}
                className="flex flex-col rounded-xl border border-slate-200 p-4"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <h2 className="truncate font-bold text-slate-950">
                      {role.name}
                    </h2>
                    <p className="mt-0.5 font-mono text-xs text-slate-500">
                      {role.code}
                    </p>
                  </div>
                  <StatusBadge tone={role.is_active ? "success" : "neutral"}>
                    {role.is_active ? "Active" : "Inactive"}
                  </StatusBadge>
                </div>
                <p className="mt-3 min-h-10 text-sm text-slate-600">
                  {role.description || "No description"}
                </p>
                <div className="mt-3 flex flex-wrap gap-1.5">
                  {[
                    ...new Set(
                      role.permissions.map((item) => item.module_code),
                    ),
                  ]
                    .slice(0, 5)
                    .map((module) => (
                      <span
                        key={module}
                        className="rounded-full bg-cyan-50 px-2 py-1 text-xs font-semibold text-cyan-800"
                      >
                        {module}
                      </span>
                    ))}
                </div>
                <p className="mt-3 text-xs font-semibold text-slate-500">
                  {role.permissions.length} permission
                  {role.permissions.length === 1 ? "" : "s"}
                </p>
                <div className="mt-auto flex flex-wrap justify-end gap-1 border-t border-slate-100 pt-3">
                  {canWrite ? (
                    <>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setEditor({ kind: "edit", role })}
                      >
                        <Pencil className="size-4" />
                        Edit
                      </Button>
                    </>
                  ) : null}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setEditor({ kind: "permissions", role })}
                  >
                    <KeyRound className="size-4" />
                    Permissions
                  </Button>
                  {canWrite ? (
                    <Button
                      variant="ghost"
                      size="sm"
                      className={role.is_active ? "text-rose-700" : ""}
                      onClick={() => setEditor({ kind: "status", role })}
                    >
                      {role.is_active ? "Deactivate" : "Activate"}
                    </Button>
                  ) : null}
                </div>
              </article>
            ))}
          </div>
        )}

        {!roles.isPending && !roles.isError && totalPages > 0 ? (
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
                onClick={() => void setFilters({ rolePage: currentPage - 1 })}
              >
                <ChevronLeft className="size-4" />
                Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={currentPage >= totalPages}
                onClick={() => void setFilters({ rolePage: currentPage + 1 })}
              >
                Next
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </div>
        ) : null}
      </Panel>

      {editor?.kind === "create" || editor?.kind === "edit" ? (
        <RoleFormDialog
          roleId={editor.kind === "edit" ? editor.role.role_id : undefined}
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
      {editor?.kind === "permissions" ? (
        <RolePermissionsDialog
          key={`${editor.role.role_id}-${editor.role.version_no}`}
          role={editor.role}
          canWrite={canWrite}
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
      {editor?.kind === "status" ? (
        <RoleStatusDialog
          role={editor.role}
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
    </>
  );
}
