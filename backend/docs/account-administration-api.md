# Account and role administration API

All routes use `/api/v1/security` and require a valid bearer session plus the
`SECURITY.READ` or `SECURITY.WRITE` permission. Every mutation is recorded in
`api_audit_log`.

## Relationship rules

- An account has one status and an optional authentication policy.
- An account can have many direct permissions, roles, owner scopes, and
  warehouse scopes.
- A role has many permissions and can be assigned to many accounts.
- Effective permissions are the union of active direct permissions and the
  permissions from active roles. They are recalculated on every authenticated
  request, so revocation takes effect immediately without waiting for the token
  to expire.
- Deleting an account or role is a soft deactivation. Historical assignments,
  document ownership, inventory movements, billing records, and audit rows are
  never orphaned.
- Disabling an account, resetting its password, changing its authentication
  policy, and reaching the failed-login lock threshold revoke active sessions.
- The final active holder of `SECURITY.WRITE` cannot be disabled or stripped of
  that access, whether the permission is direct or role-derived.
- Mutations using `expected_version` apply optimistic concurrency. Reload the
  account or role after a `409 Conflict`.

New accounts default to `PENDING` when `account_status_id` is omitted. Configure
roles/scopes first, then change the account to `ACTIVE`.

## Authentication and transaction effects

- Login locks the account row and commits the successful-login update together
  with session creation. A concurrent disable or password reset therefore
  cannot create a session after revocation.
- Permission and role changes do not require logout. Effective permissions are
  loaded again for every bearer-authenticated request.
- Owner and warehouse changes do not require logout. Operational scope checks
  query the access tables for every inbound, outbound, and billing request.
- Profile-only changes keep existing sessions. Status, password,
  authentication-policy, failed-login lockout, and explicit revoke actions end
  the affected sessions.
- Account administration never edits inventory balances or document lifecycle
  rows. Existing inbound, stock-control, outbound, and billing transactions
  retain their original atomic behavior.

## Lookup endpoints

```text
GET /api/v1/security/account-statuses
GET /api/v1/security/authentication-policies
GET /api/v1/security/permissions
```

These populate the status, policy, and permission dropdowns in the account UI.

## Account CRUD

```text
POST   /api/v1/security/accounts
GET    /api/v1/security/accounts?search=&status_id=&page=1&page_size=20
GET    /api/v1/security/accounts/:account_id
PUT    /api/v1/security/accounts/:account_id
PATCH  /api/v1/security/accounts/:account_id/status
DELETE /api/v1/security/accounts/:account_id?expected_version=4
POST   /api/v1/security/accounts/:account_id/password-reset
POST   /api/v1/security/accounts/:account_id/unlock
POST   /api/v1/security/accounts/:account_id/revoke-sessions
```

Create a pending local account:

```json
{
  "username": "receiver01",
  "email": "receiver01@example.com",
  "display_name": "Receiving Operator",
  "password": "Change-Receiver-Password-2026!",
  "preferred_timezone": "Asia/Jakarta"
}
```

Update editable identity fields:

```json
{
  "username": "receiver01",
  "email": "receiver01@example.com",
  "display_name": "Senior Receiving Operator",
  "authentication_policy_id": null,
  "preferred_timezone": "Asia/Jakarta",
  "expected_version": 1
}
```

Change status after reading the UUID from `/account-statuses`:

```json
{
  "account_status_id": "ACTIVE-STATUS-UUID",
  "expected_version": 2
}
```

Reset password or unlock a failed-login lock:

```json
{
  "password": "New-Receiver-Password-2026!",
  "expected_version": 3
}
```

```json
{
  "expected_version": 4
}
```

`GET /accounts/:account_id` is the aggregate menu response. It includes core
account fields, status, policy, lock information, roles, direct permissions,
effective permission codes, owner access, warehouse access, and the number of
active sessions. `revoke-sessions` force-logs-out that account with the
`ADMIN_REVOKED` reason.

## Roles

```text
POST   /api/v1/security/roles
GET    /api/v1/security/roles?search=&active=true&page=1&page_size=20
GET    /api/v1/security/roles/:role_id
PUT    /api/v1/security/roles/:role_id
PATCH  /api/v1/security/roles/:role_id/status
PUT    /api/v1/security/roles/:role_id/permissions
DELETE /api/v1/security/roles/:role_id?expected_version=3
```

Create a receiver role using permission UUIDs from `/permissions`:

```json
{
  "code": "RECEIVER",
  "name": "Receiving operator",
  "description": "Can read and process inbound documents",
  "permission_ids": [
    "INBOUND-READ-PERMISSION-UUID",
    "INBOUND-WRITE-PERMISSION-UUID",
    "INVENTORY-READ-PERMISSION-UUID"
  ]
}
```

Replace the complete permission set. An empty array intentionally clears it:

```json
{
  "permission_ids": [
    "INBOUND-READ-PERMISSION-UUID",
    "INBOUND-WRITE-PERMISSION-UUID"
  ],
  "expected_version": 1
}
```

Enable or disable a role:

```json
{
  "is_active": false,
  "expected_version": 2
}
```

## Account relationships

Direct permissions:

```text
GET    /api/v1/security/accounts/:account_id/permissions
POST   /api/v1/security/accounts/:account_id/permissions
DELETE /api/v1/security/accounts/:account_id/permissions/:permission_id
```

Assign a reusable role:

```text
POST   /api/v1/security/accounts/:account_id/roles
DELETE /api/v1/security/accounts/:account_id/roles/:role_id
```

```json
{
  "role_id": "RECEIVER-ROLE-UUID"
}
```

Owner and warehouse scopes:

```text
POST   /api/v1/security/accounts/:account_id/owners
DELETE /api/v1/security/accounts/:account_id/owners/:owner_id
POST   /api/v1/security/accounts/:account_id/warehouses
DELETE /api/v1/security/accounts/:account_id/warehouses/:warehouse_id
```

```json
{
  "owner_id": "OWNER-UUID"
}
```

```json
{
  "warehouse_id": "WAREHOUSE-UUID"
}
```

Owner and warehouse scopes are independent. Revoking an owner does not delete
warehouse access because the same warehouse may serve another owner assigned to
the account. Operational authorization still requires every applicable scope.

## Recommended frontend creation flow

1. Load statuses, authentication policies, permissions, roles, owners, and
   warehouses.
2. Create the account in `PENDING` state.
3. Assign one or more roles. Use direct permissions only for explicit
   exceptions.
4. Assign owner and warehouse scopes.
5. Reload the aggregate account and review `effective_permissions`.
6. Change the status to `ACTIVE` using the latest `version_no`.

This ordering prevents a new user from logging in before their access is fully
configured.
