# Inventory identity API

Inventory identity is now implemented for lots, serials and handling units.
These records answer **which batch/unit/container?**, not **how much stock?**
Creating an identity does not receive stock, create balances, or post movements.

## Start and authenticate

Restart the backend from `C:\WmsProject\backend` using `go run .`. Startup applies
additive migrations and seeds handling-unit types PALLET, CARTON, TOTE and BIN.
The existing study records are preserved. The identity tables currently start
empty; automated test fixtures were rolled back.

Use your configured API port, normally `http://localhost:8080`.
Log in using `POST /api/v1/auth/login` with `identifier` and `password`, then put
`data.token` in `Authorization: Bearer <token>` for every endpoint below.
For POST/PUT requests, set `Content-Type: application/json`.

These routes use the same database-session authentication as masters. Logout
revokes that session and subsequent identity requests return 401. Account role
permissions and owner/warehouse access-scope enforcement are still not implemented;
do not expose this backend as a production multi-tenant API yet.

## Endpoints

| Resource | Collection | Supported operations |
| --- | --- | --- |
| Lot | `/api/v1/inventory/lots` | POST, GET list, GET `/:id` |
| Serial identity | `/api/v1/inventory/serials` | POST, GET list, GET `/:id` |
| Handling unit | `/api/v1/inventory/handling-units` | POST, GET list, GET `/:id` |
| HU type master | `/api/v1/master/handling-unit-types` | POST, GET list, GET/PUT `/:id`, PATCH `/:id/deactivate` |

Internal IDs are server-generated: `LOT-<32 random hex characters>`,
`SER-<32 random hex characters>`, and `HU-<32 random hex characters>`. They are
varchar keys, matching the schema, not master UUIDs or daily document numbers.
No document counter is consumed. Do not supply IDs or audit fields in POST bodies.

Lot/serial/barcode values are trimmed and case-sensitive; internal case is not
changed. Business uniqueness is `(owner_id,item_id,lot_number)` for lots,
`(owner_id,item_id,serial_no)` for serials, and global `barcode` for handling units.
Duplicates return 409, not an update. On an uncertain network result, GET by exact
number/barcode before retrying; check the existing metadata matches your intent.

## 1. Create a coffee lot using the study masters

`POST /api/v1/inventory/lots`

```json
{
  "owner_id": "5e6a104a-51c2-4e94-8643-e339bcd5242f",
  "item_id": "b9d199ad-343e-480e-a5a5-c8af97d3ddee",
  "lot_number": "COFFEE-20260904-A",
  "manufacture_date": "2026-09-04",
  "expiry_date": "2027-09-04"
}
```

Expected: **201** with `data.lot_id`, audit fields and the dates as `YYYY-MM-DD`.
The owner is STUDY_OWNER and the item is STUDY_COFFEE_250G. These UUIDs are only
valid in the current study database; discover them again for a different DB.

Both owner and item must be active, the item must belong to that owner, and
`lot_controlled` must be true. Optional dates must be valid dates; expiry cannot
precede manufacture. No expiry is invented and no receiving shelf-life decision
is performed during identity registration; those belong to receiving later.

Optional `quality_status_id` accepts an active quality-status UUID obtained from
`GET /api/v1/master/quality-statuses`. Omitting it leaves quality unclassified
(`null`), not automatically PASSED. There is no QC transition endpoint in this slice.

Retrieve it with either:

```text
GET /api/v1/inventory/lots/<returned-lot-id>
GET /api/v1/inventory/lots?owner_id=5e6a104a-51c2-4e94-8643-e339bcd5242f&item_id=b9d199ad-343e-480e-a5a5-c8af97d3ddee&lot_number=COFFEE-20260904-A
```

A lot has no warehouse/location: the same batch can later occupy several bins.

## 2. Register a pallet and nested carton

First GET `/api/v1/master/handling-unit-types?active=true` and take the PALLET UUID.
In this development database it is `5db8fc77-b62c-4722-8dfc-c37adfa623d0`.

`POST /api/v1/inventory/handling-units`

```json
{
  "owner_id": "5e6a104a-51c2-4e94-8643-e339bcd5242f",
  "warehouse_id": "59113433-3082-4208-87c1-330f4f56fcba",
  "handling_unit_type_id": "5db8fc77-b62c-4722-8dfc-c37adfa623d0",
  "current_location_id": "33b8d9d5-96c7-48f9-b0ad-3698f902d8ca",
  "barcode": "STUDY-PLT-0001"
}
```

This creates an empty pallet at STUDY_RCV_01, not goods on a pallet.
Save `data.handling_unit_id`. Then create its carton using the CARTON type:

```json
{
  "owner_id": "5e6a104a-51c2-4e94-8643-e339bcd5242f",
  "warehouse_id": "59113433-3082-4208-87c1-330f4f56fcba",
  "handling_unit_type_id": "f15b32f1-9949-4fa8-9d2c-2277b41dcb85",
  "parent_handling_unit_id": "REPLACE_WITH_RETURNED_PALLET_ID",
  "barcode": "STUDY-CTN-0001"
}
```

If a parent is supplied and location is omitted/null, the child inherits the
parent location. Explicit location must agree with the parent and all ancestors.
A root HU may have no location yet. Parent, owner and warehouse must agree;
parent and all ancestors must be open. Cycles and more than 64 ancestors are
rejected. A supplied location must be in the warehouse, active and unlocked,
with an active zone and location type. Owner, operator, warehouse, HU type and
warehouse-owner assignment must all be active.

`is_closed` defaults to false and cannot be provided or changed through these
identity APIs. Moving, reparenting, sealing/reopening and contents changes are
reserved for audited inventory workflows, to avoid bypassing balance updates.
The type's capacity fields are master configuration, not capacity enforcement
for physical contents in this slice.

## 3. Register an individual serial

The coffee study items are NOT serial-controlled, so a serial registration for
them correctly returns 400. To test serials, first create a separate item using
`POST /api/v1/master/items`:

```json
{
  "owner_id": "5e6a104a-51c2-4e94-8643-e339bcd5242f",
  "code": "STUDY_SCANNER",
  "name": "Study - Barcode Scanner",
  "base_uom_id": "6df21b70-0d00-4406-b8d1-71aeca1b62fa",
  "lot_controlled": false,
  "serial_controlled": true
}
```

Use its returned `item_id` with `POST /api/v1/inventory/serials`:

```json
{
  "owner_id": "5e6a104a-51c2-4e94-8643-e339bcd5242f",
  "item_id": "REPLACE_WITH_RETURNED_SCANNER_ITEM_UUID",
  "serial_no": "SCANNER-SN-000001"
}
```

Expected: **201** with `data.serial_id`. This only registers identity. It does
not mean the unit is received, in stock, at a location, or assigned to a lot/HU.
Even if an item has both control flags enabled, serial-to-lot membership and
individual current-state tracking require a later `serial_inventory` (or
equivalent) design integrated with movement transactions.

## Filtering and errors

All inventory lists accept `page` (1..1000000), `page_size` (1..100, default 20),
`owner_id` and `search`. Results use `data.items`, `total_items`, `total_pages`.
`search` is PostgreSQL case-insensitive substring matching (`%` and `_` act as
wildcards). Use the exact filters for scanner lookups:

- Lots: `item_id`, `lot_number`.
- Serials: `item_id`, `serial_no`.
- HUs: `warehouse_id`, `current_location_id`, `parent_handling_unit_id`, `barcode`,
  `closed=true|false`.

Unknown or repeated inventory list parameters are rejected. Identity POSTs
reject unknown fields, trailing JSON, malformed UUIDs and oversized bodies.
No generic PUT/DELETE/deactivate exists for identities: traceability keys and
metadata are immutable in this implementation. Inactive master references do
not hide historical identities from GET.

| Status | Meaning |
| --- | --- |
| 400 | Invalid fields, missing/inactive references, wrong scope or control flags |
| 401 | Missing/expired/revoked bearer session |
| 404 | Requested identity ID does not exist |
| 409 | Duplicate lot/serial business key or HU barcode |
| 500 | Unexpected internal failure (database details not exposed) |

## Implementation and verification

One model and typed repository per table. Master HU types stay under `master`;
identity routes/controllers/services/DTOs/models/repositories stay under
`inventory`. All database access stays in repositories. Creation uses database
transactions and shared locks on referenced records. Database unique/composite
foreign keys independently enforce duplicate and owner/item/location/parent
scope constraints; incompatible legacy rows cause migration failure, not cleanup.

```powershell
cd C:\WmsProject\backend
go test ./...
go vet ./...
$env:WMS_INTEGRATION_TEST = '1'
go test -p 1 ./... -count=1
Remove-Item Env:WMS_INTEGRATION_TEST
```

PostgreSQL integration tests require the development database and schema-create
permissions. New inventory tests run against existing and fresh schemas, rolling
back every fixture and migration. Route tests check all nine identity endpoints
and all master routes require sessions. Real-API smoke checks also verify revoked
sessions and rejection of serial registration for a non-serial-controlled item.

Next: inventory balances, movement ledger and explicit serial current state,
followed by receiving workflows that update them atomically. Do not create stock
by directly inserting or editing balances.
