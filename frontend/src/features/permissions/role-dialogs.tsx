"use client";

import { useEffect, useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, X } from "lucide-react";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import {
  changeRoleStatus,
  createRole,
  getRole,
  listPermissions,
  permissionKeys,
  replaceRolePermissions,
  updateRole,
} from "@/features/permissions/permission-api";
import { PermissionChecklist } from "@/features/permissions/permission-checklist";
import {
  emptyRoleForm,
  roleFormSchema,
  type RoleFormValues,
} from "@/features/permissions/permission-schema";
import type { Role } from "@/features/permissions/permission-types";
import { ApiError } from "@/lib/api/client";

function optional(value: string) {
  const normalized = value.trim();
  return normalized || undefined;
}

function DialogFrame({
  title,
  description,
  pending,
  onOpenChange,
  children,
  wide,
}: {
  title: string;
  description: string;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  children: React.ReactNode;
  wide?: boolean;
}) {
  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!pending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content
          className={`fixed top-1/2 left-1/2 z-50 max-h-[92vh] w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none ${wide ? "max-w-3xl" : "max-w-lg"}`}
        >
          <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                {title}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                {description}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close role dialog"
                className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function RoleFormDialog({
  roleId,
  onOpenChange,
}: {
  roleId?: string;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = Boolean(roleId);
  const queryClient = useQueryClient();
  const role = useQuery({
    queryKey: permissionKeys.role(roleId ?? "new"),
    queryFn: () => getRole(roleId!),
    enabled: editing,
  });
  const permissions = useQuery({
    queryKey: permissionKeys.available(),
    queryFn: listPermissions,
    enabled: !editing,
  });
  const {
    control,
    register,
    reset,
    handleSubmit,
    formState: { errors },
  } = useForm<RoleFormValues>({
    resolver: zodResolver(roleFormSchema),
    defaultValues: emptyRoleForm,
  });
  const save = useMutation({
    mutationFn: (values: RoleFormValues) => {
      if (!editing) {
        return createRole({
          code: values.code.trim().toUpperCase(),
          name: values.name.trim(),
          description: optional(values.description),
          permission_ids: values.permission_ids,
        });
      }
      if (!role.data) throw new Error("Role details are not loaded.");
      return updateRole(role.data.role_id, {
        name: values.name.trim(),
        description: optional(values.description),
        expected_version: role.data.version_no,
      });
    },
    onSuccess: async (savedRole) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: permissionKeys.roleLists() }),
        queryClient.invalidateQueries({
          queryKey: permissionKeys.role(savedRole.role_id),
        }),
      ]);
      toast.success(editing ? "Role updated." : "Role created.");
      onOpenChange(false);
    },
  });

  useEffect(() => {
    if (!editing) reset(emptyRoleForm);
  }, [editing, reset]);

  useEffect(() => {
    if (role.data) {
      reset({
        code: role.data.code,
        name: role.data.name,
        description: role.data.description ?? "",
        permission_ids: role.data.permissions.map((item) => item.permission_id),
      });
    }
  }, [reset, role.data]);

  return (
    <DialogFrame
      title={editing ? "Edit role" : "Create role"}
      description={
        editing
          ? "Update the role name and description."
          : "Define a reusable bundle of application permissions."
      }
      pending={save.isPending}
      onOpenChange={onOpenChange}
      wide={!editing}
    >
      {editing && role.isPending ? (
        <div className="grid min-h-64 place-items-center p-6 text-sm text-slate-600">
          <LoaderCircle className="mb-3 size-6 animate-spin text-cyan-700" />
          Loading role…
        </div>
      ) : editing && role.isError ? (
        <div className="p-5">
          <p className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
            {role.error.message}
          </p>
        </div>
      ) : (
        <form
          noValidate
          className="p-5 sm:p-6"
          onSubmit={handleSubmit((values) => save.mutate(values))}
        >
          {save.error ? (
            <div
              role="alert"
              className="mb-5 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
            >
              <p>{save.error.message}</p>
              {editing &&
              save.error instanceof ApiError &&
              save.error.status === 409 ? (
                <Button
                  type="button"
                  variant="secondary"
                  size="sm"
                  className="mt-3"
                  onClick={async () => {
                    save.reset();
                    await role.refetch();
                  }}
                >
                  Reload latest role
                </Button>
              ) : null}
            </div>
          ) : null}
          <div className="grid gap-5 sm:grid-cols-2">
            <FormField
              label="Code"
              htmlFor="role-code"
              required
              error={errors.code?.message}
            >
              <Input
                id="role-code"
                placeholder="WAREHOUSE_ADMIN"
                readOnly={editing}
                autoFocus
                invalid={Boolean(errors.code)}
                {...register("code")}
              />
            </FormField>
            <FormField
              label="Name"
              htmlFor="role-name"
              required
              error={errors.name?.message}
            >
              <Input
                id="role-name"
                placeholder="Warehouse administrator"
                invalid={Boolean(errors.name)}
                {...register("name")}
              />
            </FormField>
            <div className="sm:col-span-2">
              <FormField
                label="Description"
                htmlFor="role-description"
                error={errors.description?.message}
              >
                <textarea
                  id="role-description"
                  rows={3}
                  className="mt-2 w-full rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                  {...register("description")}
                />
              </FormField>
            </div>
            {!editing ? (
              <div className="sm:col-span-2">
                <FormField
                  label="Initial permissions"
                  htmlFor="role-permissions"
                >
                  {permissions.isPending ? (
                    <div className="mt-2 grid min-h-36 place-items-center rounded-xl border border-slate-200 text-sm text-slate-500">
                      Loading permissions…
                    </div>
                  ) : permissions.isError ? (
                    <p className="mt-2 rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
                      {permissions.error.message}
                    </p>
                  ) : (
                    <Controller
                      control={control}
                      name="permission_ids"
                      render={({ field }) => (
                        <div id="role-permissions" className="mt-2">
                          <PermissionChecklist
                            permissions={permissions.data}
                            selected={field.value}
                            onChange={field.onChange}
                          />
                        </div>
                      )}
                    />
                  )}
                </FormField>
              </div>
            ) : null}
          </div>
          <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
            <Dialog.Close asChild>
              <Button type="button" variant="secondary">
                Cancel
              </Button>
            </Dialog.Close>
            <Button
              type="submit"
              disabled={save.isPending || (!editing && permissions.isPending)}
            >
              {save.isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              {save.isPending
                ? "Saving…"
                : editing
                  ? "Save changes"
                  : "Create role"}
            </Button>
          </div>
        </form>
      )}
    </DialogFrame>
  );
}

export function RolePermissionsDialog({
  role,
  canWrite,
  onOpenChange,
}: {
  role: Role;
  canWrite: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [selected, setSelected] = useState(
    role.permissions.map((item) => item.permission_id),
  );
  const permissions = useQuery({
    queryKey: permissionKeys.available(),
    queryFn: listPermissions,
  });
  const save = useMutation({
    mutationFn: () =>
      replaceRolePermissions(role.role_id, selected, role.version_no),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: permissionKeys.roleLists() }),
        queryClient.invalidateQueries({
          queryKey: permissionKeys.role(role.role_id),
        }),
        queryClient.invalidateQueries({ queryKey: ["accounts", "detail"] }),
      ]);
      toast.success("Role permissions updated.");
      onOpenChange(false);
    },
  });

  return (
    <DialogFrame
      title="Role permissions"
      description={`${role.name} · ${selected.length} selected`}
      pending={save.isPending}
      onOpenChange={onOpenChange}
      wide
    >
      <div className="p-5 sm:p-6">
        {permissions.isPending ? (
          <div className="grid min-h-56 place-items-center text-sm text-slate-600">
            <LoaderCircle className="mb-3 size-6 animate-spin text-cyan-700" />
            Loading permissions…
          </div>
        ) : permissions.isError ? (
          <p className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
            {permissions.error.message}
          </p>
        ) : (
          <PermissionChecklist
            permissions={permissions.data}
            selected={selected}
            onChange={setSelected}
            disabled={!canWrite || save.isPending}
          />
        )}
        {save.error ? (
          <p
            role="alert"
            className="mt-4 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {save.error.message}
          </p>
        ) : null}
        <div className="mt-6 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
          <Dialog.Close asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </Dialog.Close>
          {canWrite ? (
            <Button
              type="button"
              disabled={save.isPending || permissions.isPending}
              onClick={() => save.mutate()}
            >
              {save.isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              {save.isPending ? "Saving…" : "Save permission set"}
            </Button>
          ) : null}
        </div>
      </div>
    </DialogFrame>
  );
}

export function RoleStatusDialog({
  role,
  onOpenChange,
}: {
  role: Role;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const save = useMutation({
    mutationFn: () =>
      changeRoleStatus(role.role_id, !role.is_active, role.version_no),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: permissionKeys.roleLists() }),
        queryClient.invalidateQueries({
          queryKey: permissionKeys.role(role.role_id),
        }),
        queryClient.invalidateQueries({ queryKey: ["accounts", "detail"] }),
      ]);
      toast.success(role.is_active ? "Role deactivated." : "Role activated.");
      onOpenChange(false);
    },
  });

  return (
    <DialogFrame
      title={role.is_active ? "Deactivate role?" : "Activate role?"}
      description={
        role.is_active
          ? "Assignments are retained, but this role stops contributing permissions."
          : "Existing assignments will immediately contribute this role's permissions."
      }
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      <div className="p-5">
        {save.error ? (
          <p
            role="alert"
            className="mb-4 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {save.error.message}
          </p>
        ) : null}
        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <Dialog.Close asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </Dialog.Close>
          <Button
            type="button"
            className={role.is_active ? "bg-rose-700 hover:bg-rose-800" : ""}
            disabled={save.isPending}
            onClick={() => save.mutate()}
          >
            {save.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            {save.isPending
              ? "Saving…"
              : role.is_active
                ? "Deactivate role"
                : "Activate role"}
          </Button>
        </div>
      </div>
    </DialogFrame>
  );
}
