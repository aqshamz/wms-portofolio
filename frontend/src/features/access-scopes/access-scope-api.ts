import { apiRequest } from "@/lib/api/client";
import type {
  AccountListFilters,
  AccountPage,
  AccountSummary,
  OwnerAccess,
  WarehouseAccess,
} from "@/features/access-scopes/access-scope-types";

export const accessScopeKeys = {
  all: ["access-scopes"] as const,
  accounts: (filters: AccountListFilters) =>
    [...accessScopeKeys.all, "accounts", filters] as const,
  account: (accountId: string) =>
    [...accessScopeKeys.all, "account", accountId] as const,
  ownerAccess: (accountId: string) =>
    [...accessScopeKeys.all, "owners", accountId] as const,
  warehouseAccess: (accountId: string) =>
    [...accessScopeKeys.all, "warehouses", accountId] as const,
};

export function accountListPath(filters: AccountListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return `/api/v1/security/accounts?${query}`;
}

export function listAccounts(filters: AccountListFilters) {
  return apiRequest<AccountPage>(accountListPath(filters));
}

export function getAccount(accountId: string) {
  return apiRequest<AccountSummary>(`/api/v1/security/accounts/${accountId}`);
}

export function listOwnerAccess(accountId: string) {
  return apiRequest<OwnerAccess[]>(
    `/api/v1/master/accounts/${accountId}/owner-access`,
  );
}

export function grantOwnerAccess(accountId: string, ownerId: string) {
  return apiRequest<void>(`/api/v1/master/owners/${ownerId}/account-access`, {
    method: "POST",
    body: { account_id: accountId },
  });
}

export function revokeOwnerAccess(accountId: string, ownerId: string) {
  return apiRequest<void>(
    `/api/v1/master/accounts/${accountId}/owner-access/${ownerId}`,
    { method: "DELETE" },
  );
}

export function listWarehouseAccess(accountId: string) {
  return apiRequest<WarehouseAccess[]>(
    `/api/v1/master/accounts/${accountId}/warehouse-access`,
  );
}

export function grantWarehouseAccess(accountId: string, warehouseId: string) {
  return apiRequest<void>(`/api/v1/master/warehouse-access/${warehouseId}`, {
    method: "POST",
    body: { account_id: accountId },
  });
}

export function revokeWarehouseAccess(accountId: string, warehouseId: string) {
  return apiRequest<void>(
    `/api/v1/master/accounts/${accountId}/warehouse-access/${warehouseId}`,
    { method: "DELETE" },
  );
}
