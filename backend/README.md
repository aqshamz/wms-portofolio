# WMS API

Initial backend foundation using Gin, GORM, and PostgreSQL.

## Run locally

1. Copy `.env.example` to `.env` and review the credentials.
2. Make sure PostgreSQL is listening on port `7654`.
3. Run `go run .`.
4. Check `GET http://localhost:8080/health`.

When `DB_AUTO_CREATE=true`, startup connects to the PostgreSQL `postgres`
administrative database and creates the configured `wms` database if it does
not exist. It also creates the `wms` schema. GORM table migration is deliberately
not enabled until the entity models are implemented and reviewed against
`../database/wms_schema.sql`.

## Package rules

- `routes`: endpoint registration only.
- `controller`: HTTP request validation and response handling, grouped by domain.
- `services`: business logic, grouped by domain.
- `repository`: the only application layer allowed to query GORM; one file per table.
- `models`: database entity structs; one file per table.
- `dto`: JSON request and response structs, grouped by domain.
- `middleware`: authentication, authorization, logging, and other guards.
- `utils`: small application-wide helpers without domain business logic.
- `config`: environment and infrastructure configuration.

Dependencies must flow in this direction:

`routes -> controller -> services -> repository -> models`

DTOs may be used by controllers and services. Controllers and services must not
query the database directly.

## Authentication

Startup migrates only these authentication tables: `account_status`,
`authentication_policy`, `session_revocation_reason`, `app_account`, and
`app_session`. With the local `.env`, an initial development administrator is
created once:

- username: `admin`
- password: `Admin@123456`

Change this password before using the API outside local development. The
bootstrap process never overwrites an existing account.

Authentication endpoints:

- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/logout-all`

Login request:

```json
{
  "identifier": "admin",
  "password": "Admin@123456"
}
```

For protected endpoints, copy `data.token` from the login response into the
request header:

```text
Authorization: Bearer <token>
```

Only the SHA-256 hash of this opaque token is stored in PostgreSQL.

## Master data: organizations and warehouses

All master endpoints require the Bearer token described above.

- `POST /api/v1/master/organizations`
- `GET /api/v1/master/organizations`
- `GET /api/v1/master/organizations/:id`
- `PUT /api/v1/master/organizations/:id`
- `PATCH /api/v1/master/organizations/:id/deactivate`
- `POST /api/v1/master/warehouses`
- `GET /api/v1/master/warehouses`
- `GET /api/v1/master/warehouses/:id`
- `PUT /api/v1/master/warehouses/:id`
- `PATCH /api/v1/master/warehouses/:id/deactivate`

Create an organization first:

```json
{
  "code": "ORG01",
  "name": "Main Organization",
  "timezone_name": "Asia/Jakarta",
  "city": "Jakarta",
  "country_code": "ID"
}
```

Then use its `organization_id` as the warehouse `operator_id`:

```json
{
  "operator_id": "<organization_id>",
  "code": "WH01",
  "name": "Main Warehouse",
  "timezone_name": "Asia/Jakarta",
  "city": "Jakarta",
  "country_code": "ID"
}
```

Update and deactivate requests must include the latest `updated_at` value as
`expected_updated_at`. A stale value returns HTTP `409`, preventing one user
from silently overwriting another user's changes. List endpoints accept
`search`, `active`, `page`, and `page_size`; warehouse lists also accept
`operator_id`.

### Warehouse structure and access scope

The remaining warehouse foundation is available under `/api/v1/master` and
requires authentication:

- `POST|GET /warehouses/:id/owners`
- `PATCH /warehouses/:id/owners/:owner_id/deactivate`
- `POST|GET /warehouses/:id/zones`
- `PUT|PATCH /zones/:id` (`PATCH` uses `/deactivate`)
- `POST|GET /warehouses/:id/locations`
- `GET|PUT|PATCH /locations/:id` (`PATCH` uses `/deactivate`)
- `POST|GET /location-types`
- `PUT /location-types/:id`
- `POST /owners/:owner_id/account-access`
- `POST /warehouse-access/:warehouse_id`
- `GET /accounts/:account_id/owner-access`
- `DELETE /accounts/:account_id/owner-access/:owner_id`
- `GET /accounts/:account_id/warehouse-access`
- `DELETE /accounts/:account_id/warehouse-access/:warehouse_id`

Account-access grant requests contain the target account:

```json
{
  "account_id": "<account_id>"
}
```

Location capacities use JSON strings to preserve PostgreSQL `numeric(20,6)`
precision:

```json
{
  "zone_id": "<zone_id>",
  "location_type_id": "<location_type_id>",
  "code": "A01-B01-L01-P01",
  "barcode": "LOC-A01-B01-L01-P01",
  "pick_sequence": 10,
  "max_weight": "1000.500000",
  "max_volume": "20.125000",
  "is_pick_face": true
}
```

Startup seeds the standard location types `RECEIVING`, `STORAGE`, `PICK_FACE`,
`STAGING`, `CHECKING`, `PACKING`, `SHIPPING`, and `QUARANTINE`. Location codes
are unique per warehouse; a location's zone is enforced to belong to that same
warehouse. Owner assignments and physical master records use soft deactivation.
Account scope grants are join records and are removed when revoked.

### Business partners, items, UOMs, and classifications

Catalog master data is now available under `/api/v1/master`: business partners
with multiple types, owner-scoped item categories/items, global UOMs, item
conversions/barcodes, inventory statuses, quality statuses, and inspection
results. Each catalog table has its own model and repository, with JSON DTOs
kept separate.

See [Catalog API and testing guide](docs/catalog-api.md) for endpoints, request
bodies, validation rules, reference seeds, and automated tests.

These routes validate database sessions and can enforce module permissions when
RBAC is enabled. Owner/warehouse access remains an additional data-scope guard.
Inventory stock transactions and quality workflows are separate modules.

### Operational configuration

Document types/statuses/transitions, versioned document-number rules, atomic
daily counters/ID allocation, picking and putaway strategies/rules, and task
types/statuses/transitions/priorities are available under `/api/v1/master`.

See [Operational configuration API and tests](docs/operational-configuration-api.md).
Counters are read-only except through ID allocation; allocation is a consuming
POST, not a preview. These APIs configure workflows and strategies but do not
execute warehouse transactions or implement role-based authorization.

### Study dataset

See [Study master data](docs/study-master-data.md) for a persisted coffee-warehouse
example, real local UUIDs, setup order, and API requests to explore the data.
With the development API running, use `scripts/seed-study-data.ps1` to create
missing `STUDY_` records without resetting existing entries. This opt-in script
is not run on application startup and creates no inventory transactions.
Then run `scripts/seed-study-inventory.ps1` for idempotent lots, a pallet, serials,
opening balances and RECEIVE movements. See
[Study inventory data](docs/study-inventory-data.md).

### Inventory identity

Lots, serial identities and handling units now support create/list/get under
`/api/v1/inventory`, with session guards, scope validation and database constraints.
Handling-unit types support master CRUD under `/api/v1/master/handling-unit-types`.
See [Inventory identity API and testing examples](docs/inventory-identity-api.md).
Identity creation does not receive stock or change quantities. Movement, QC,
container changes and individual serial current state remain separate work.

### Inventory core

Current balances, immutable movements and serial current-state pointers now form
the shared inventory posting foundation. HTTP exposes authenticated inquiry only;
domain services use an atomic, idempotent internal posting method. See
[Inventory core architecture and inquiry API](docs/inventory-core-api.md).
Generic balance/movement mutation is intentionally not exposed. Reservations and
document-backed stock-control workflows remain separate orchestration layers.

### Stock control

Authenticated immediate commands now cover internal moves, status changes,
adjustments, single-balance count reconciliation, and atomic inter-warehouse
transfer. They require optimistic balance versions and create idempotent core
movements. See [Stock-control API](docs/stock-control-api.md). These are posting
commands; maker/checker and dispatch/in-transit/receipt document workflows remain
distinct future orchestration over the same core.

### Inbound part 1

Purchase orders, inbound orders/ASNs, and receiving are implemented through the
point where accepted stock is posted as `QC_PENDING`. The flow supports partial
planning and receipt, dock rejection, lot/serial/HU identity, receiving-UOM
conversion, optimistic document versions, atomic completion, and idempotent
inventory posting. See [Inbound part 1 API](docs/inbound-part-1-api.md).

### Inbound part 2

Quality inspection now splits received stock into `PUTAWAY_PENDING` and/or
`QUARANTINE`. Authenticated putaway tasks move passed inventory into configured
storage as `AVAILABLE`; quarantine dispositions support accept-to-storage,
return-to-vendor, and disposal with partial-case tracking. See
[Inbound part 2 API](docs/inbound-part-2-api.md).

### Inbound part 3

The inbound lifecycle is completed with configurable PO-line receiving
tolerances, multiple-receipt variance tracking, short closure, safe document
cancellation, untouched-receipt reversal, putaway assignment/retargeting,
putaway cancellation and reversal, and quarantine rework with child
reinspection lineage. Every exception and reversal remains auditable in the
inventory ledger and inbound exception inquiry. See
[Inbound part 3 API](docs/inbound-part-3-api.md).

### Outbound part 1

Client delivery orders now support recorded validation, release, FEFO/FIFO
allocation, inventory reservations, multi-order waves, pick-task execution,
short-pick closure, idempotent `PICK` movements, and staging confirmation. Stock
is reserved without changing on-hand quantity and is moved only by physical pick
confirmation. See [Outbound part 1 API](docs/outbound-part-1-api.md). Verification,
packing, shipment, and delivery continue in outbound part 2.

### Outbound part 2

Completed staging now continues through line-level picked-stock checks, packing
with idempotent `PACK` movements, multi-order shipment manifests with
idempotent `SHIP` movements, delivery events, proof of delivery, delivery
failure, and controlled return-to-depot stock. A return policy routes undelivered
goods into a non-allocatable inventory status. See
[Outbound part 2 API](docs/outbound-part-2-api.md).

### Outbound completion

Outbound is now closed for the planned backend scope: draft-order maintenance,
executable QC replacement picks, carrier/service/driver setup, primary-driver
shipment enforcement, safe pre-execution cancellations, serial/HU state
continuity, and bearer-account owner/warehouse authorization are implemented.
See [Outbound completion API](docs/outbound-completion-api.md). Billing can build
on the final shipped/delivered quantities and immutable movement history.

## Billing

Billing is implemented as a separate owner/warehouse-scoped module. It provides
contracts, versioned movement/manual rate cards, idempotent billable-event
collection, calculation/review/reopen/cancel controls, invoice snapshots,
partial payments, and credit-note settlement. See [Billing API](docs/billing-api.md)
for the complete study flow and request examples.

## Hardening

The API now includes module RBAC, permission administration, exact-origin CORS,
request/body/rate limits, request IDs, security headers, mutation audit records,
versioned startup migrations, graceful HTTP shutdown, and route-level OpenAPI
documentation at `/docs`. Development keeps RBAC disabled until an administrator
is bootstrapped; production configuration fails closed. See
[Backend hardening](docs/hardening.md) for setup and verification.

## Account administration

Security administration now provides account create/list/detail/update,
status/deactivation, password reset, unlock, direct permissions, reusable roles,
role permissions, and owner/warehouse access. Account detail returns the full
aggregate needed by the frontend menu. Deactivation preserves operational
history and revokes sessions. See
[Account and role administration API](docs/account-administration-api.md).
