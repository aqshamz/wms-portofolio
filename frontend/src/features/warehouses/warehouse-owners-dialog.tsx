"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Building2, LoaderCircle, Trash2, X } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  listOrganizations,
  organizationKeys,
} from "@/features/organizations/organization-api";
import {
  assignWarehouseOwner,
  deactivateWarehouseOwner,
  listWarehouseOwners,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import type {
  Warehouse,
  WarehouseOwner,
} from "@/features/warehouses/warehouse-types";

const activeOrganizationFilters = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

function DeactivateOwnerDialog({
  owner,
  pending,
  error,
  onConfirm,
  onOpenChange,
}: {
  owner?: WarehouseOwner;
  pending: boolean;
  error: Error | null;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog.Root open={Boolean(owner)} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-[60] bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-[70] w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-white p-5 shadow-2xl focus:outline-none sm:p-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                Deactivate served owner?
              </Dialog.Title>
              <Dialog.Description className="mt-2 text-sm leading-6 text-slate-600">
                {owner?.owner_name} can no longer be selected for new operations
                at this warehouse. Existing records remain available, and the
                relationship can be assigned again later.
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                disabled={pending}
                aria-label="Close owner deactivation confirmation"
                className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100 disabled:opacity-50"
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
              onClick={onConfirm}
              className="bg-rose-700 hover:bg-rose-800 active:bg-rose-900"
            >
              {pending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : (
                <Trash2 className="size-4" />
              )}
              {pending ? "Deactivating…" : "Deactivate owner"}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function WarehouseOwnersDialog({
  warehouse,
  canWrite,
  onOpenChange,
}: {
  warehouse?: Warehouse;
  canWrite: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [ownerId, setOwnerId] = useState("");
  const [deactivateTarget, setDeactivateTarget] = useState<WarehouseOwner>();
  const owners = useQuery({
    queryKey: warehouseKeys.owners(warehouse?.warehouse_id ?? "none"),
    queryFn: () => listWarehouseOwners(warehouse!.warehouse_id),
    enabled: Boolean(warehouse),
  });
  const organizations = useQuery({
    queryKey: organizationKeys.list(activeOrganizationFilters),
    queryFn: () => listOrganizations(activeOrganizationFilters),
    enabled: Boolean(warehouse && canWrite),
  });
  const activeOwnerIds = new Set(
    (owners.data ?? [])
      .filter((owner) => owner.is_active)
      .map((owner) => owner.owner_id),
  );
  const availableOrganizations =
    organizations.data?.items.filter(
      (organization) => !activeOwnerIds.has(organization.organization_id),
    ) ?? [];

  const assign = useMutation({
    mutationFn: () => assignWarehouseOwner(warehouse!.warehouse_id, ownerId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: warehouseKeys.owners(warehouse!.warehouse_id),
      });
      setOwnerId("");
      toast.success("Served owner assigned.");
    },
  });
  const deactivate = useMutation({
    mutationFn: (owner: WarehouseOwner) =>
      deactivateWarehouseOwner(warehouse!.warehouse_id, owner.owner_id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: warehouseKeys.owners(warehouse!.warehouse_id),
      });
      setDeactivateTarget(undefined);
      toast.success("Served owner deactivated.");
    },
  });

  return (
    <>
      <Dialog.Root
        open={warehouse !== undefined}
        onOpenChange={(open) => {
          if (!assign.isPending && !deactivate.isPending) onOpenChange(open);
        }}
      >
        <Dialog.Portal>
          <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
          <Dialog.Content className="fixed top-1/2 left-1/2 z-50 flex max-h-[calc(100vh-2rem)] w-[calc(100%-2rem)] max-w-2xl -translate-x-1/2 -translate-y-1/2 flex-col overflow-hidden rounded-2xl bg-white shadow-2xl focus:outline-none">
            <div className="flex items-start justify-between gap-4 border-b border-slate-200 px-5 py-4 sm:px-6">
              <div>
                <Dialog.Title className="text-lg font-bold text-slate-950">
                  Served owners
                </Dialog.Title>
                <Dialog.Description className="mt-1 text-sm text-slate-600">
                  {warehouse?.name} · {warehouse?.code}
                </Dialog.Description>
              </div>
              <Dialog.Close asChild>
                <button
                  type="button"
                  aria-label="Close served owners"
                  className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
                >
                  <X className="size-5" />
                </button>
              </Dialog.Close>
            </div>

            <div className="overflow-y-auto p-5 sm:p-6">
              <p className="rounded-xl border border-cyan-200 bg-cyan-50 px-4 py-3 text-sm leading-6 text-cyan-950">
                A served owner is an organization whose inventory this warehouse
                may receive, store, and dispatch. Account access scopes are
                configured separately.
              </p>

              {canWrite ? (
                <section className="mt-5 rounded-xl border border-slate-200 p-4">
                  <h2 className="font-bold text-slate-950">Assign owner</h2>
                  <div className="mt-3 flex flex-col gap-2 sm:flex-row">
                    <Select
                      ariaLabel="Organization to assign as served owner"
                      value={ownerId}
                      options={availableOrganizations.map((organization) => ({
                        value: organization.organization_id,
                        label: `${organization.name} (${organization.code})`,
                      }))}
                      placeholder={
                        organizations.isPending
                          ? "Loading organizations…"
                          : availableOrganizations.length === 0
                            ? "All active organizations are assigned"
                            : "Select an organization"
                      }
                      disabled={
                        organizations.isPending ||
                        organizations.isError ||
                        availableOrganizations.length === 0 ||
                        assign.isPending
                      }
                      onValueChange={setOwnerId}
                      className="flex-1"
                    />
                    <Button
                      type="button"
                      disabled={!ownerId || assign.isPending}
                      onClick={() => assign.mutate()}
                    >
                      {assign.isPending ? (
                        <LoaderCircle className="size-4 animate-spin" />
                      ) : null}
                      {assign.isPending ? "Assigning…" : "Assign owner"}
                    </Button>
                  </div>
                  {organizations.isError || assign.error ? (
                    <p
                      role="alert"
                      className="mt-3 rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-900"
                    >
                      {(organizations.error ?? assign.error)?.message}
                    </p>
                  ) : null}
                </section>
              ) : null}

              <section className="mt-5">
                <div className="flex items-center justify-between gap-3">
                  <h2 className="font-bold text-slate-950">
                    Owner relationships
                  </h2>
                  {!owners.isPending && !owners.isError ? (
                    <span className="text-xs font-semibold text-slate-500">
                      {owners.data?.filter((owner) => owner.is_active).length ??
                        0}{" "}
                      active
                    </span>
                  ) : null}
                </div>

                {owners.isPending ? (
                  <div className="mt-3 flex min-h-32 items-center justify-center gap-2 rounded-xl border border-slate-200 text-sm text-slate-600">
                    <LoaderCircle className="size-4 animate-spin" /> Loading
                    owners…
                  </div>
                ) : owners.isError ? (
                  <div
                    role="alert"
                    className="mt-3 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
                  >
                    <p>{owners.error.message}</p>
                    <Button
                      type="button"
                      variant="secondary"
                      size="sm"
                      className="mt-3"
                      onClick={() => owners.refetch()}
                    >
                      Try again
                    </Button>
                  </div>
                ) : owners.data?.length ? (
                  <ul className="mt-3 divide-y divide-slate-200 overflow-hidden rounded-xl border border-slate-200">
                    {owners.data.map((owner) => (
                      <li
                        key={owner.owner_id}
                        className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between"
                      >
                        <div className="flex min-w-0 items-center gap-3">
                          <div className="grid size-10 shrink-0 place-items-center rounded-xl bg-slate-100 text-slate-600">
                            <Building2 className="size-4" />
                          </div>
                          <div className="min-w-0">
                            <p className="truncate font-semibold text-slate-950">
                              {owner.owner_name}
                            </p>
                            <p className="mt-0.5 font-mono text-xs text-slate-500">
                              {owner.owner_code}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-center justify-between gap-3 sm:justify-end">
                          <StatusBadge
                            tone={owner.is_active ? "success" : "neutral"}
                          >
                            {owner.is_active ? "Active" : "Inactive"}
                          </StatusBadge>
                          {canWrite && owner.is_active ? (
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              className="text-rose-700 hover:bg-rose-50 hover:text-rose-800"
                              onClick={() => {
                                deactivate.reset();
                                setDeactivateTarget(owner);
                              }}
                            >
                              Deactivate
                            </Button>
                          ) : null}
                        </div>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <div className="mt-3 rounded-xl border border-dashed border-slate-300 px-5 py-10 text-center">
                    <Building2 className="mx-auto size-8 text-slate-300" />
                    <p className="mt-3 font-semibold text-slate-800">
                      No served owners assigned
                    </p>
                    <p className="mt-1 text-sm text-slate-500">
                      Assign an active organization to enable warehouse
                      operations for its inventory.
                    </p>
                  </div>
                )}
              </section>
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>

      <DeactivateOwnerDialog
        owner={deactivateTarget}
        pending={deactivate.isPending}
        error={deactivate.error}
        onConfirm={() =>
          deactivateTarget && deactivate.mutate(deactivateTarget)
        }
        onOpenChange={(open) => {
          if (!open && !deactivate.isPending) setDeactivateTarget(undefined);
        }}
      />
    </>
  );
}
