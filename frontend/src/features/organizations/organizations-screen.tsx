"use client";

import { useState } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import {
  Building2,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  LoaderCircle,
  Pencil,
  Plus,
  Search,
} from "lucide-react";
import Link from "next/link";
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
  listOrganizations,
  organizationKeys,
} from "@/features/organizations/organization-api";
import { OrganizationDeactivateDialog } from "@/features/organizations/organization-deactivate-dialog";
import { OrganizationFormDialog } from "@/features/organizations/organization-form-dialog";
import type { Organization } from "@/features/organizations/organization-types";

const PAGE_SIZE = 10;
const activeOptions = ["all", "active", "inactive"] as const;
const statusFilterOptions = [
  { value: "all", label: "All statuses" },
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
] as const;

type EditorState =
  { mode: "create" } | { mode: "edit"; organizationId: string } | null;

const dateFormatter = new Intl.DateTimeFormat("en-ID", {
  day: "2-digit",
  month: "short",
  year: "numeric",
});

function locationLabel(organization: Organization) {
  return (
    [organization.city, organization.province, organization.country_code]
      .filter(Boolean)
      .join(", ") || "Not provided"
  );
}

function RowActions({
  organization,
  canWrite,
  onEdit,
  onDeactivate,
}: {
  organization: Organization;
  canWrite: boolean;
  onEdit: () => void;
  onDeactivate: () => void;
}) {
  if (!canWrite) return null;

  return (
    <div className="flex flex-wrap items-center justify-end gap-1">
      <Button variant="ghost" size="sm" onClick={onEdit}>
        <Pencil className="size-4" />
        Edit
      </Button>
      {organization.is_active ? (
        <Button
          variant="ghost"
          size="sm"
          onClick={onDeactivate}
          className="text-rose-700 hover:bg-rose-50 hover:text-rose-800"
        >
          Deactivate
        </Button>
      ) : null}
    </div>
  );
}

function LoadingRows() {
  return (
    <div className="space-y-3 p-4 sm:p-5">
      {Array.from({ length: 5 }, (_, index) => (
        <div
          key={index}
          className="h-16 animate-pulse rounded-xl bg-slate-100"
        />
      ))}
    </div>
  );
}

export function OrganizationsScreen({ canWrite }: { canWrite: boolean }) {
  const [filters, setFilters] = useQueryStates({
    search: parseAsString.withDefault(""),
    active: parseAsStringLiteral(activeOptions).withDefault("all"),
    page: parseAsInteger.withDefault(1),
  });
  const [editor, setEditor] = useState<EditorState>(null);
  const [deactivateTarget, setDeactivateTarget] = useState<Organization>();
  const queryFilters = {
    search: filters.search,
    active: filters.active,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };
  const organizations = useQuery({
    queryKey: organizationKeys.list(queryFilters),
    queryFn: () => listOrganizations(queryFilters),
    placeholderData: keepPreviousData,
  });

  const items = organizations.data?.items ?? [];
  const totalItems = organizations.data?.total_items ?? 0;
  const totalPages = organizations.data?.total_pages ?? 0;
  const currentPage = organizations.data?.page ?? queryFilters.page;

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <Link
            href="/master-data/core"
            className="inline-flex min-h-9 items-center gap-2 text-sm font-semibold text-slate-600 hover:text-slate-950"
          >
            <ChevronLeft className="size-4" />
            Core masters
          </Link>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <Building2 className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Organizations
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Owners and operators represented in warehouse transactions.
              </p>
            </div>
          </div>
        </div>
        {canWrite ? (
          <Button
            className="w-full sm:w-auto"
            onClick={() => setEditor({ mode: "create" })}
          >
            <Plus className="size-4" />
            Create organization
          </Button>
        ) : null}
      </header>

      <Panel className="overflow-hidden">
        <div className="border-b border-slate-200 p-4 sm:p-5">
          <form
            className="flex flex-col gap-3 md:flex-row"
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
              <span className="sr-only">Search organizations</span>
              <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
              <input
                key={filters.search}
                name="search"
                defaultValue={filters.search}
                placeholder="Search by code or name"
                className="h-11 w-full rounded-xl border border-slate-300 bg-white pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
              />
            </label>
            <Select
              ariaLabel="Filter organizations by status"
              value={filters.active}
              options={statusFilterOptions}
              onValueChange={(active) => void setFilters({ active, page: 1 })}
              className="md:w-40"
            />
            <Button type="submit" variant="secondary">
              Search
            </Button>
            {filters.search ? (
              <Button
                type="button"
                variant="ghost"
                onClick={() => {
                  void setFilters({ search: "", page: 1 });
                }}
              >
                Clear
              </Button>
            ) : null}
          </form>
        </div>

        <div className="flex min-h-12 items-center justify-between gap-3 border-b border-slate-200 px-4 py-3 text-sm sm:px-5">
          <p className="font-semibold text-slate-700">
            {organizations.isPending
              ? "Loading organizations…"
              : `${totalItems.toLocaleString()} organization${totalItems === 1 ? "" : "s"}`}
          </p>
          {organizations.isFetching && !organizations.isPending ? (
            <span className="inline-flex items-center gap-2 text-xs text-slate-500">
              <LoaderCircle className="size-3.5 animate-spin" />
              Refreshing
            </span>
          ) : null}
        </div>

        {organizations.isPending ? (
          <LoadingRows />
        ) : organizations.isError ? (
          <div className="p-5">
            <div
              role="alert"
              className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900"
            >
              <div className="flex gap-3">
                <CircleAlert className="mt-0.5 size-5 shrink-0" />
                <div>
                  <p className="font-semibold">
                    Organizations could not be loaded
                  </p>
                  <p className="mt-1 text-sm">{organizations.error.message}</p>
                </div>
              </div>
              <Button
                variant="secondary"
                className="mt-4"
                onClick={() => organizations.refetch()}
              >
                Try again
              </Button>
            </div>
          </div>
        ) : items.length === 0 ? (
          <div className="px-5 py-14 text-center">
            <Building2 className="mx-auto size-9 text-slate-300" />
            <h2 className="mt-4 font-bold text-slate-950">
              No organizations found
            </h2>
            <p className="mt-1 text-sm text-slate-500">
              {filters.search || filters.active !== "all"
                ? "Adjust the search or status filter."
                : "Create the first organization to begin configuring warehouses."}
            </p>
          </div>
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="w-full border-collapse text-left">
                <thead>
                  <tr className="bg-slate-50 text-xs font-bold tracking-wide text-slate-500 uppercase">
                    <th className="px-5 py-3">Organization</th>
                    <th className="px-5 py-3">Location</th>
                    <th className="px-5 py-3">Timezone</th>
                    <th className="px-5 py-3">Status</th>
                    <th className="px-5 py-3">Updated</th>
                    <th className="px-5 py-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200">
                  {items.map((organization) => (
                    <tr
                      key={organization.organization_id}
                      className="hover:bg-slate-50/70"
                    >
                      <td className="px-5 py-4">
                        <p className="font-semibold text-slate-950">
                          {organization.name}
                        </p>
                        <p className="mt-0.5 font-mono text-xs text-slate-500">
                          {organization.code}
                        </p>
                      </td>
                      <td className="px-5 py-4 text-sm text-slate-600">
                        {locationLabel(organization)}
                      </td>
                      <td className="px-5 py-4 text-sm text-slate-600">
                        {organization.timezone_name}
                      </td>
                      <td className="px-5 py-4">
                        <StatusBadge
                          tone={organization.is_active ? "success" : "neutral"}
                        >
                          {organization.is_active ? "Active" : "Inactive"}
                        </StatusBadge>
                      </td>
                      <td className="px-5 py-4 text-sm text-slate-600">
                        {dateFormatter.format(
                          new Date(organization.updated_at),
                        )}
                      </td>
                      <td className="px-5 py-4">
                        <RowActions
                          organization={organization}
                          canWrite={canWrite}
                          onEdit={() =>
                            setEditor({
                              mode: "edit",
                              organizationId: organization.organization_id,
                            })
                          }
                          onDeactivate={() => setDeactivateTarget(organization)}
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <ul className="divide-y divide-slate-200 md:hidden">
              {items.map((organization) => (
                <li key={organization.organization_id} className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="truncate font-semibold text-slate-950">
                        {organization.name}
                      </p>
                      <p className="mt-0.5 font-mono text-xs text-slate-500">
                        {organization.code}
                      </p>
                    </div>
                    <StatusBadge
                      tone={organization.is_active ? "success" : "neutral"}
                    >
                      {organization.is_active ? "Active" : "Inactive"}
                    </StatusBadge>
                  </div>
                  <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
                    <div>
                      <dt className="text-xs font-semibold text-slate-500">
                        Location
                      </dt>
                      <dd className="mt-1 text-slate-700">
                        {locationLabel(organization)}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-xs font-semibold text-slate-500">
                        Timezone
                      </dt>
                      <dd className="mt-1 text-slate-700">
                        {organization.timezone_name}
                      </dd>
                    </div>
                  </dl>
                  <div className="mt-3 border-t border-slate-100 pt-2">
                    <RowActions
                      organization={organization}
                      canWrite={canWrite}
                      onEdit={() =>
                        setEditor({
                          mode: "edit",
                          organizationId: organization.organization_id,
                        })
                      }
                      onDeactivate={() => setDeactivateTarget(organization)}
                    />
                  </div>
                </li>
              ))}
            </ul>
          </>
        )}

        {!organizations.isPending &&
        !organizations.isError &&
        totalPages > 0 ? (
          <div className="flex flex-col gap-3 border-t border-slate-200 px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5">
            <p className="text-sm text-slate-600">
              Page{" "}
              <span className="font-semibold text-slate-900">
                {currentPage}
              </span>{" "}
              of {totalPages}
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

      {editor ? (
        <OrganizationFormDialog
          open
          organizationId={
            editor.mode === "edit" ? editor.organizationId : undefined
          }
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
      {deactivateTarget ? (
        <OrganizationDeactivateDialog
          organization={deactivateTarget}
          onOpenChange={(open) => {
            if (!open) setDeactivateTarget(undefined);
          }}
        />
      ) : null}
    </div>
  );
}
