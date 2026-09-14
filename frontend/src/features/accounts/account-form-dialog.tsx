"use client";

import { useEffect } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, X } from "lucide-react";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  accountKeys,
  createAccount,
  getAccount,
  listAccountStatuses,
  listAuthenticationPolicies,
  updateAccount,
} from "@/features/accounts/account-api";
import {
  accountFormSchema,
  emptyAccountForm,
  type AccountFormValues,
} from "@/features/accounts/account-schema";
import type {
  AccountDetail,
  CreateAccountRequest,
  UpdateAccountRequest,
} from "@/features/accounts/account-types";
import { ApiError } from "@/lib/api/client";

function optional(value: string) {
  const normalized = value.trim();
  return normalized || undefined;
}

function editValues(account: AccountDetail): AccountFormValues {
  return {
    mode: "edit",
    username: account.username,
    email: account.email ?? "",
    display_name: account.display_name,
    password: "",
    account_status_id: account.status.account_status_id,
    authentication_policy_id: account.authentication_policy_id ?? "",
    preferred_timezone: account.preferred_timezone ?? "",
  };
}

function describedBy(id: string, error?: string) {
  return error ? `${id}-error` : undefined;
}

export function AccountFormDialog({
  open,
  accountId,
  onOpenChange,
}: {
  open: boolean;
  accountId?: string;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = Boolean(accountId);
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: accountKeys.detail(accountId ?? "new"),
    queryFn: () => getAccount(accountId!),
    enabled: open && editing,
  });
  const statuses = useQuery({
    queryKey: accountKeys.statuses(),
    queryFn: listAccountStatuses,
    enabled: open,
  });
  const policies = useQuery({
    queryKey: accountKeys.policies(),
    queryFn: listAuthenticationPolicies,
    enabled: open,
  });
  const {
    control,
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<AccountFormValues>({
    resolver: zodResolver(accountFormSchema),
    defaultValues: emptyAccountForm,
  });
  const save = useMutation({
    mutationFn: (values: AccountFormValues) => {
      const common = {
        username: values.username.trim(),
        email: optional(values.email)?.toLowerCase(),
        display_name: values.display_name.trim(),
        authentication_policy_id: optional(values.authentication_policy_id),
        preferred_timezone: optional(values.preferred_timezone),
      };
      if (values.mode === "create") {
        const request: CreateAccountRequest = {
          ...common,
          password: values.password,
          account_status_id: optional(values.account_status_id),
        };
        return createAccount(request);
      }
      if (!detail.data) throw new Error("Account details are not loaded.");
      const request: UpdateAccountRequest = {
        ...common,
        expected_version: detail.data.version_no,
      };
      return updateAccount(detail.data.account_id, request);
    },
    onSuccess: async (account) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: accountKeys.lists() }),
        queryClient.invalidateQueries({
          queryKey: accountKeys.detail(account.account_id),
        }),
      ]);
      toast.success(editing ? "Account updated." : "Account created.");
      onOpenChange(false);
    },
  });

  useEffect(() => {
    if (open && !editing) reset(emptyAccountForm);
  }, [editing, open, reset]);

  useEffect(() => {
    if (open && detail.data) reset(editValues(detail.data));
  }, [detail.data, open, reset]);

  const statusOptions =
    statuses.data?.map((status) => ({
      value: status.account_status_id,
      label: status.name,
    })) ?? [];
  const policyOptions = [
    { value: "default", label: "Default policy" },
    ...(policies.data?.map((policy) => ({
      value: policy.authentication_policy_id,
      label: `${policy.name}${policy.is_default ? " (default)" : ""}`,
    })) ?? []),
  ];

  return (
    <Dialog.Root
      open={open}
      onOpenChange={(nextOpen) => {
        if (!save.isPending) onOpenChange(nextOpen);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[92vh] w-[calc(100%-2rem)] max-w-2xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                {editing ? "Edit account" : "Create account"}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                {editing
                  ? "Update identity, authentication policy, and timezone."
                  : "Create login credentials and an initial account profile."}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close account form"
                className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>

          {editing && detail.isPending ? (
            <div className="grid min-h-72 place-items-center p-6 text-sm text-slate-600">
              <LoaderCircle className="mb-3 size-6 animate-spin text-cyan-700" />
              Loading account…
            </div>
          ) : editing && detail.isError ? (
            <div className="p-6">
              <p
                role="alert"
                className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
              >
                {detail.error.message}
              </p>
              <Button
                className="mt-4"
                variant="secondary"
                onClick={() => void detail.refetch()}
              >
                Try again
              </Button>
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
                  className="mb-5 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-900"
                >
                  <p>{save.error.message}</p>
                  {editing &&
                  save.error instanceof ApiError &&
                  save.error.status === 409 ? (
                    <Button
                      type="button"
                      size="sm"
                      variant="secondary"
                      className="mt-3"
                      onClick={async () => {
                        save.reset();
                        await detail.refetch();
                      }}
                    >
                      Reload latest data
                    </Button>
                  ) : null}
                </div>
              ) : null}

              <div className="grid gap-5 sm:grid-cols-2">
                <FormField
                  label="Username"
                  htmlFor="account-username"
                  required
                  error={errors.username?.message}
                >
                  <Input
                    id="account-username"
                    autoFocus
                    autoComplete="off"
                    placeholder="warehouse.user"
                    invalid={Boolean(errors.username)}
                    aria-describedby={describedBy(
                      "account-username",
                      errors.username?.message,
                    )}
                    {...register("username")}
                  />
                </FormField>
                <FormField
                  label="Display name"
                  htmlFor="account-display-name"
                  required
                  error={errors.display_name?.message}
                >
                  <Input
                    id="account-display-name"
                    placeholder="Warehouse User"
                    invalid={Boolean(errors.display_name)}
                    aria-describedby={describedBy(
                      "account-display-name",
                      errors.display_name?.message,
                    )}
                    {...register("display_name")}
                  />
                </FormField>
                <FormField
                  label="Email"
                  htmlFor="account-email"
                  error={errors.email?.message}
                >
                  <Input
                    id="account-email"
                    type="email"
                    autoComplete="off"
                    placeholder="user@example.com"
                    invalid={Boolean(errors.email)}
                    aria-describedby={describedBy(
                      "account-email",
                      errors.email?.message,
                    )}
                    {...register("email")}
                  />
                </FormField>
                {editing ? (
                  <div />
                ) : (
                  <FormField
                    label="Initial password"
                    htmlFor="account-password"
                    required
                    error={errors.password?.message}
                  >
                    <Input
                      id="account-password"
                      type="password"
                      autoComplete="new-password"
                      invalid={Boolean(errors.password)}
                      aria-describedby={describedBy(
                        "account-password",
                        errors.password?.message,
                      )}
                      {...register("password")}
                    />
                  </FormField>
                )}
                {!editing ? (
                  <FormField
                    label="Initial status"
                    htmlFor="account-status"
                    error={errors.account_status_id?.message}
                  >
                    <Controller
                      control={control}
                      name="account_status_id"
                      render={({ field }) => (
                        <Select
                          id="account-status"
                          ariaLabel="Initial account status"
                          value={field.value}
                          options={statusOptions}
                          onValueChange={field.onChange}
                          placeholder="Pending (default)"
                          disabled={statuses.isPending || statuses.isError}
                          className="mt-2"
                        />
                      )}
                    />
                  </FormField>
                ) : null}
                <FormField
                  label="Authentication policy"
                  htmlFor="account-policy"
                  error={errors.authentication_policy_id?.message}
                >
                  <Controller
                    control={control}
                    name="authentication_policy_id"
                    render={({ field }) => (
                      <Select
                        id="account-policy"
                        ariaLabel="Authentication policy"
                        value={field.value || "default"}
                        options={policyOptions}
                        onValueChange={(value) =>
                          field.onChange(value === "default" ? "" : value)
                        }
                        placeholder="Default policy"
                        disabled={policies.isPending || policies.isError}
                        className="mt-2"
                      />
                    )}
                  />
                </FormField>
                <FormField
                  label="Preferred timezone"
                  htmlFor="account-timezone"
                  error={errors.preferred_timezone?.message}
                >
                  <Input
                    id="account-timezone"
                    list="account-timezone-options"
                    placeholder="Asia/Jakarta"
                    invalid={Boolean(errors.preferred_timezone)}
                    {...register("preferred_timezone")}
                  />
                  <datalist id="account-timezone-options">
                    <option value="Asia/Jakarta" />
                    <option value="Asia/Makassar" />
                    <option value="Asia/Jayapura" />
                    <option value="UTC" />
                  </datalist>
                </FormField>
              </div>

              {!editing ? (
                <p className="mt-5 rounded-xl bg-amber-50 px-4 py-3 text-sm text-amber-900">
                  A new account defaults to Pending unless you explicitly select
                  another status.
                </p>
              ) : null}

              <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
                <Dialog.Close asChild>
                  <Button type="button" variant="secondary">
                    Cancel
                  </Button>
                </Dialog.Close>
                <Button
                  type="submit"
                  disabled={
                    save.isPending || statuses.isPending || policies.isPending
                  }
                >
                  {save.isPending ? (
                    <LoaderCircle className="size-4 animate-spin" />
                  ) : null}
                  {save.isPending
                    ? "Saving…"
                    : editing
                      ? "Save changes"
                      : "Create account"}
                </Button>
              </div>
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
