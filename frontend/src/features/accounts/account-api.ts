import { apiRequest } from "@/lib/api/client";
import type {
  AccountDetail,
  AccountListFilters,
  AccountPage,
  AccountStatus,
  AuthenticationPolicy,
  CreateAccountRequest,
  UpdateAccountRequest,
} from "@/features/accounts/account-types";

const accountBasePath = "/api/v1/security/accounts";

export const accountKeys = {
  all: ["accounts"] as const,
  lists: () => [...accountKeys.all, "list"] as const,
  list: (filters: AccountListFilters) =>
    [...accountKeys.lists(), filters] as const,
  details: () => [...accountKeys.all, "detail"] as const,
  detail: (accountId: string) => [...accountKeys.details(), accountId] as const,
  statuses: () => [...accountKeys.all, "statuses"] as const,
  policies: () => [...accountKeys.all, "policies"] as const,
};

export function accountListPath(filters: AccountListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.statusId) query.set("status_id", filters.statusId);
  return `${accountBasePath}?${query}`;
}

export function listAccounts(filters: AccountListFilters) {
  return apiRequest<AccountPage>(accountListPath(filters));
}

export function getAccount(accountId: string) {
  return apiRequest<AccountDetail>(`${accountBasePath}/${accountId}`);
}

export function listAccountStatuses() {
  return apiRequest<AccountStatus[]>("/api/v1/security/account-statuses");
}

export function listAuthenticationPolicies() {
  return apiRequest<AuthenticationPolicy[]>(
    "/api/v1/security/authentication-policies",
  );
}

export function createAccount(request: CreateAccountRequest) {
  return apiRequest<AccountDetail>(accountBasePath, {
    method: "POST",
    body: request,
  });
}

export function updateAccount(
  accountId: string,
  request: UpdateAccountRequest,
) {
  return apiRequest<AccountDetail>(`${accountBasePath}/${accountId}`, {
    method: "PUT",
    body: request,
  });
}

export function changeAccountStatus(
  accountId: string,
  accountStatusId: string,
  expectedVersion: number,
) {
  return apiRequest<AccountDetail>(`${accountBasePath}/${accountId}/status`, {
    method: "PATCH",
    body: {
      account_status_id: accountStatusId,
      expected_version: expectedVersion,
    },
  });
}

export function deactivateAccount(accountId: string, expectedVersion: number) {
  const query = new URLSearchParams({
    expected_version: String(expectedVersion),
  });
  return apiRequest<AccountDetail>(`${accountBasePath}/${accountId}?${query}`, {
    method: "DELETE",
  });
}

export function resetAccountPassword(
  accountId: string,
  password: string,
  expectedVersion: number,
) {
  return apiRequest<void>(`${accountBasePath}/${accountId}/password-reset`, {
    method: "POST",
    body: { password, expected_version: expectedVersion },
  });
}

export function unlockAccount(accountId: string, expectedVersion: number) {
  return apiRequest<AccountDetail>(`${accountBasePath}/${accountId}/unlock`, {
    method: "POST",
    body: { expected_version: expectedVersion },
  });
}

export function revokeAccountSessions(accountId: string) {
  return apiRequest<void>(`${accountBasePath}/${accountId}/revoke-sessions`, {
    method: "POST",
  });
}
