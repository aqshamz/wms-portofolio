"use client";

import { useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  CircleAlert,
  KeyRound,
  LoaderCircle,
  LockKeyholeOpen,
  LogOut,
  Pencil,
  ShieldCheck,
  Trash2,
  X,
} from "lucide-react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  accountKeys,
  changeAccountStatus,
  getAccount,
  listAccountStatuses,
  resetAccountPassword,
} from "@/features/accounts/account-api";
import {
  passwordResetSchema,
  type PasswordResetValues,
} from "@/features/accounts/account-schema";
import type { AccountSummary } from "@/features/accounts/account-types";

function DialogFrame({
  title,
  description,
  pending,
  onOpenChange,
  children,
}: {
  title: string;
  description: string;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  children: React.ReactNode;
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
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="flex items-start justify-between gap-4 border-b border-slate-200 px-5 py-4">
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
                aria-label="Close dialog"
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

export function AccountStatusDialog({
  account,
  onOpenChange,
}: {
  account: AccountSummary;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [statusId, setStatusId] = useState(account.status.account_status_id);
  const statuses = useQuery({
    queryKey: accountKeys.statuses(),
    queryFn: listAccountStatuses,
  });
  const changeStatus = useMutation({
    mutationFn: () =>
      changeAccountStatus(account.account_id, statusId, account.version_no),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: accountKeys.lists() }),
        queryClient.invalidateQueries({
          queryKey: accountKeys.detail(account.account_id),
        }),
      ]);
      toast.success("Account status updated.");
      onOpenChange(false);
    },
  });
  const options =
    statuses.data?.map((status) => ({
      value: status.account_status_id,
      label: `${status.name}${status.allows_login ? " · Login allowed" : " · Login blocked"}`,
    })) ?? [];

  return (
    <DialogFrame
      title="Change account status"
      description={`Set the login state for ${account.display_name}.`}
      pending={changeStatus.isPending}
      onOpenChange={onOpenChange}
    >
      <div className="p-5">
        <FormField label="Status" htmlFor="account-status-change" required>
          <Select
            id="account-status-change"
            ariaLabel="Account status"
            value={statusId}
            options={options}
            onValueChange={setStatusId}
            placeholder={statuses.isPending ? "Loading statuses…" : "Status"}
            disabled={statuses.isPending || statuses.isError}
            className="mt-2"
          />
        </FormField>
        <p className="mt-3 text-xs text-slate-500">
          Choosing a status that blocks login also revokes active sessions.
        </p>
        {statuses.isError || changeStatus.error ? (
          <p
            role="alert"
            className="mt-4 rounded-xl bg-rose-50 px-4 py-3 text-sm text-rose-900"
          >
            {statuses.error?.message ?? changeStatus.error?.message}
          </p>
        ) : null}
        <div className="mt-6 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
          <Dialog.Close asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </Dialog.Close>
          <Button
            type="button"
            disabled={
              !statusId ||
              statusId === account.status.account_status_id ||
              changeStatus.isPending
            }
            onClick={() => changeStatus.mutate()}
          >
            {changeStatus.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            {changeStatus.isPending ? "Updating…" : "Change status"}
          </Button>
        </div>
      </div>
    </DialogFrame>
  );
}

export function PasswordResetDialog({
  account,
  onOpenChange,
}: {
  account: AccountSummary;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<PasswordResetValues>({
    resolver: zodResolver(passwordResetSchema),
    defaultValues: { password: "", confirmation: "" },
  });
  const resetPassword = useMutation({
    mutationFn: (values: PasswordResetValues) =>
      resetAccountPassword(
        account.account_id,
        values.password,
        account.version_no,
      ),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: accountKeys.lists() }),
        queryClient.invalidateQueries({
          queryKey: accountKeys.detail(account.account_id),
        }),
      ]);
      toast.success("Password reset and active sessions revoked.");
      onOpenChange(false);
    },
  });

  return (
    <DialogFrame
      title="Reset password"
      description={`Set a new password for ${account.display_name}.`}
      pending={resetPassword.isPending}
      onOpenChange={onOpenChange}
    >
      <form
        noValidate
        className="p-5"
        onSubmit={handleSubmit((values) => resetPassword.mutate(values))}
      >
        <div className="space-y-5">
          <FormField
            label="New password"
            htmlFor="password-reset-new"
            required
            error={errors.password?.message}
          >
            <Input
              id="password-reset-new"
              type="password"
              autoFocus
              autoComplete="new-password"
              invalid={Boolean(errors.password)}
              {...register("password")}
            />
          </FormField>
          <FormField
            label="Confirm password"
            htmlFor="password-reset-confirmation"
            required
            error={errors.confirmation?.message}
          >
            <Input
              id="password-reset-confirmation"
              type="password"
              autoComplete="new-password"
              invalid={Boolean(errors.confirmation)}
              {...register("confirmation")}
            />
          </FormField>
        </div>
        <p className="mt-4 text-xs text-slate-500">
          Resetting the password immediately revokes every active session for
          this account.
        </p>
        {resetPassword.error ? (
          <p
            role="alert"
            className="mt-4 rounded-xl bg-rose-50 px-4 py-3 text-sm text-rose-900"
          >
            {resetPassword.error.message}
          </p>
        ) : null}
        <div className="mt-6 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
          <Dialog.Close asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </Dialog.Close>
          <Button type="submit" disabled={resetPassword.isPending}>
            {resetPassword.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            {resetPassword.isPending ? "Resetting…" : "Reset password"}
          </Button>
        </div>
      </form>
    </DialogFrame>
  );
}

export type AccountActionKind = "unlock" | "sessions" | "deactivate";

export function AccountManageDialog({
  account,
  currentAccountId,
  canWrite,
  onEdit,
  onStatus,
  onPassword,
  onAction,
  onOpenChange,
}: {
  account: AccountSummary;
  currentAccountId: string;
  canWrite: boolean;
  onEdit: (account: AccountSummary) => void;
  onStatus: (account: AccountSummary) => void;
  onPassword: (account: AccountSummary) => void;
  onAction: (kind: AccountActionKind, account: AccountSummary) => void;
  onOpenChange: (open: boolean) => void;
}) {
  const detail = useQuery({
    queryKey: accountKeys.detail(account.account_id),
    queryFn: () => getAccount(account.account_id),
  });
  const latest = detail.data ?? account;
  const locked =
    latest.status.code === "LOCKED" ||
    (latest.locked_until && new Date(latest.locked_until) > new Date());

  function continueWith(callback: () => void) {
    onOpenChange(false);
    callback();
  }

  return (
    <DialogFrame
      title="Manage account"
      description={`${latest.display_name} · @${latest.username}`}
      pending={false}
      onOpenChange={onOpenChange}
    >
      {detail.isPending ? (
        <div className="grid min-h-56 place-items-center p-6 text-sm text-slate-600">
          <LoaderCircle className="mb-3 size-6 animate-spin text-cyan-700" />
          Loading security details…
        </div>
      ) : detail.isError ? (
        <div className="p-5">
          <div
            role="alert"
            className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900"
          >
            <div className="flex gap-3">
              <CircleAlert className="mt-0.5 size-5 shrink-0" />
              <p className="text-sm">{detail.error.message}</p>
            </div>
            <Button
              variant="secondary"
              className="mt-4"
              onClick={() => void detail.refetch()}
            >
              Try again
            </Button>
          </div>
        </div>
      ) : (
        <div className="p-5">
          <dl className="grid grid-cols-2 gap-3 rounded-xl bg-slate-50 p-4 text-sm">
            <div>
              <dt className="text-xs font-semibold text-slate-500">Roles</dt>
              <dd className="mt-1 font-bold text-slate-950">
                {detail.data.roles.length}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold text-slate-500">
                Effective permissions
              </dt>
              <dd className="mt-1 font-bold text-slate-950">
                {detail.data.effective_permissions.length}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold text-slate-500">
                Access scope
              </dt>
              <dd className="mt-1 font-bold text-slate-950">
                {detail.data.owner_access.length} owners ·{" "}
                {detail.data.warehouse_access.length} warehouses
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold text-slate-500">
                Active sessions
              </dt>
              <dd className="mt-1 font-bold text-slate-950">
                {detail.data.active_session_count}
              </dd>
            </div>
          </dl>

          {canWrite ? (
            <div className="mt-5 grid gap-2 sm:grid-cols-2">
              <Button
                variant="secondary"
                onClick={() => continueWith(() => onEdit(latest))}
              >
                <Pencil className="size-4" />
                Edit profile
              </Button>
              <Button
                variant="secondary"
                onClick={() => continueWith(() => onStatus(latest))}
              >
                <ShieldCheck className="size-4" />
                Change status
              </Button>
              <Button
                variant="secondary"
                onClick={() => continueWith(() => onPassword(latest))}
              >
                <KeyRound className="size-4" />
                Reset password
              </Button>
              {locked ? (
                <Button
                  variant="secondary"
                  onClick={() => continueWith(() => onAction("unlock", latest))}
                >
                  <LockKeyholeOpen className="size-4" />
                  Unlock account
                </Button>
              ) : null}
              {detail.data.active_session_count > 0 ? (
                <Button
                  variant="secondary"
                  onClick={() =>
                    continueWith(() => onAction("sessions", latest))
                  }
                >
                  <LogOut className="size-4" />
                  Revoke sessions
                </Button>
              ) : null}
              {latest.status.code !== "DISABLED" &&
              latest.account_id !== currentAccountId ? (
                <Button
                  variant="secondary"
                  className="text-rose-700 hover:bg-rose-50"
                  onClick={() =>
                    continueWith(() => onAction("deactivate", latest))
                  }
                >
                  <Trash2 className="size-4" />
                  Deactivate account
                </Button>
              ) : null}
            </div>
          ) : (
            <p className="mt-5 rounded-xl bg-slate-50 px-4 py-3 text-sm text-slate-600">
              You have read-only access to account security details.
            </p>
          )}
        </div>
      )}
    </DialogFrame>
  );
}

const actionCopy: Record<
  AccountActionKind,
  { title: string; description: (name: string) => string; confirm: string }
> = {
  unlock: {
    title: "Unlock account?",
    description: (name) =>
      `Clear failed login attempts and the current lock for ${name}.`,
    confirm: "Unlock account",
  },
  sessions: {
    title: "Revoke active sessions?",
    description: (name) =>
      `${name} will be signed out on every device and must log in again.`,
    confirm: "Revoke sessions",
  },
  deactivate: {
    title: "Deactivate account?",
    description: (name) =>
      `${name} will be unable to log in. Historical relationships are retained.`,
    confirm: "Deactivate account",
  },
};

export function ConfirmAccountActionDialog({
  kind,
  account,
  pending,
  error,
  onConfirm,
  onOpenChange,
}: {
  kind: AccountActionKind;
  account: AccountSummary;
  pending: boolean;
  error: Error | null;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
}) {
  const copy = actionCopy[kind];
  return (
    <DialogFrame
      title={copy.title}
      description={copy.description(account.display_name)}
      pending={pending}
      onOpenChange={onOpenChange}
    >
      <div className="p-5">
        {error ? (
          <p
            role="alert"
            className="mb-4 rounded-xl bg-rose-50 px-4 py-3 text-sm text-rose-900"
          >
            {error.message}
          </p>
        ) : null}
        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <Dialog.Close asChild>
            <Button type="button" variant="secondary" disabled={pending}>
              Cancel
            </Button>
          </Dialog.Close>
          <Button
            type="button"
            disabled={pending}
            className={
              kind === "deactivate" ? "bg-rose-700 hover:bg-rose-800" : ""
            }
            onClick={onConfirm}
          >
            {pending ? <LoaderCircle className="size-4 animate-spin" /> : null}
            {pending ? "Working…" : copy.confirm}
          </Button>
        </div>
      </div>
    </DialogFrame>
  );
}
