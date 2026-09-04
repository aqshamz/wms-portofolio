# Catalog master-data API

This slice adds business partners and their types, UOMs, item categories, items,
item conversions/barcodes, inventory statuses, quality statuses, and inspection
results. It does **not** implement stock balances, inventory movements, inspections,
or quarantine workflows.

## Start and authenticate

From PowerShell:

```powershell
cd C:\WmsProject\backend
go run .
```

Use `http://localhost:8080` (or the configured port), not `0.0.0.0` in your API
client. Restart an already-running API to load these routes. Startup runs catalog
migrations and inserts standard reference codes; it does not create example
partners or items.

Login with your existing account:

```http
POST /api/v1/auth/login
Content-Type: application/json

{"identifier":"<username or email>","password":"<your password>"}
```

Copy `data.token` from login. Every endpoint below requires:

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

Authentication continues to validate the database session on every request.
These endpoints currently have **session authentication only**. Role permissions
and owner/warehouse access enforcement are not implemented by this slice; a
logged-in account is not restricted by its account-scope assignments yet.

## Endpoints

Prefix all paths with `/api/v1/master`.

| Resource | Collection path |
| --- | --- |
| Partner types | `/partner-types` |
| Business partners | `/business-partners` |
| Global UOMs | `/uoms` |
| Owner-scoped category hierarchy | `/item-categories` |
| Items | `/items` |
| Inventory classifications | `/inventory-statuses` |
| Quality classifications | `/quality-statuses` |
| Inspection result classifications | `/inspection-results` |

Every collection supports:

- `POST /` relative to the collection: create an active record.
- `GET /`: paginated list.
- `GET /:id`: detail.
- `PUT /:id`: update mutable fields.
- `PATCH /:id/deactivate`: soft deactivate.

Additional relationship endpoints:

| Method | Path | Purpose |
| --- | --- | --- |
| GET, POST | `/business-partners/:id/types` | List/assign partner types |
| DELETE | `/business-partners/:id/types/:partner_type_id` | Remove type assignment |
| GET, POST | `/items/:id/uoms` | List/create conversions |
| PUT | `/items/:id/uoms/:item_uom_id` | Update conversion |
| PATCH | `/items/:id/uoms/:item_uom_id/deactivate` | Deactivate conversion |
| GET, POST | `/items/:id/barcodes` | List/create barcodes |
| PUT | `/items/:id/barcodes/:barcode_id` | Update barcode |
| PATCH | `/items/:id/barcodes/:barcode_id/primary` | Make active barcode primary |
| PATCH | `/items/:id/barcodes/:barcode_id/deactivate` | Deactivate barcode |

`item_uom_id` identifies the conversion row, **not** the global `uom_id`.
Child list endpoints return arrays, including inactive records. Item detail includes
`uoms` and `barcodes`; partner detail includes `partner_types`.

Top-level lists accept `page` (default 1), `page_size` (default 20, maximum 100),
`search` (code/name), and `active=true|false`. Additional filters:

- Business partners: `owner_id`, `partner_type_code`.
- Item categories: `owner_id`. Results are flat records with `parent_category_id`.
- Items: `owner_id`, `category_id` (exact category, not descendants).

Responses use the existing `success/message/data` envelope. Paginated `data`
contains `items`, `page`, `page_size`, `total_items`, and `total_pages`.

## Manual test sequence

Replace angle-bracket placeholders with IDs returned by the API. An owner is an
existing active organization from `GET /api/v1/master/organizations`.

### 1. Read reference data

```http
GET /api/v1/master/uoms
GET /api/v1/master/partner-types
GET /api/v1/master/inventory-statuses
GET /api/v1/master/quality-statuses
GET /api/v1/master/inspection-results
```

Startup seeds UOMs `EA, BOX, CTN, PLT, KG, G, L, M`; partner types
`SUPPLIER, FACTORY, CUSTOMER, STORE, CARRIER, OTHER`; inventory statuses
`QC_PENDING, AVAILABLE, HOLD, QUARANTINE, DAMAGED, EXPIRED`; quality statuses
`PENDING, PASSED, FAILED, WAIVED`; inspection results
`ACCEPTED, PARTIAL, REJECTED`.

Seeds are insert-only by code. They preserve existing names, flags, and
deactivations. `AVAILABLE` is initially allocatable/pickable; `ACCEPTED` and
`PARTIAL` are initially accepted results. These are classifications, not a
workflow engine.

### 2. Create a business partner and assign types

```http
POST /api/v1/master/business-partners

{
  "owner_id": "<organization_id>",
  "code": "SUP001",
  "name": "Packaging Supplier",
  "legal_name": "Packaging Supplier Ltd",
  "email": "sales@example.com",
  "country_code": "ID"
}
```

```http
POST /api/v1/master/business-partners/<partner_id>/types

{"partner_type_id":"<SUPPLIER partner_type_id>"}
```

Repeat with another type if needed. Reassigning the same type is idempotent.
List/filter using `/business-partners?owner_id=<id>&partner_type_code=SUPPLIER`.

### 3. Create a category and an item

```http
POST /api/v1/master/item-categories

{
  "owner_id": "<organization_id>",
  "code": "PACKAGING",
  "name": "Packaging",
  "parent_category_id": null
}
```

```http
POST /api/v1/master/items

{
  "owner_id": "<organization_id>",
  "category_id": "<category_id>",
  "code": "SKU001",
  "name": "Packaging Tape",
  "base_uom_id": "<EA uom_id>",
  "weight": "0.250000",
  "volume": "0.001000",
  "lot_controlled": true,
  "serial_controlled": false,
  "shelf_life_days": 365,
  "minimum_receive_days": 90
}
```

The item and its active base conversion `1` are created in one transaction.
Retrieve `GET /items/<item_id>` to see the conversion row.

### 4. Add a box conversion and barcodes

```http
POST /api/v1/master/items/<item_id>/uoms

{
  "uom_id": "<BOX uom_id>",
  "conversion_to_base": "12",
  "length": "30",
  "width": "20",
  "height": "15",
  "weight": "3",
  "is_receiving_uom": true,
  "is_picking_uom": false
}
```

This means **1 BOX = 12 base units**. Dimension units are not defined in the
current schema; use one agreed measurement convention throughout your operation.

```http
POST /api/v1/master/items/<item_id>/barcodes

{
  "uom_id": "<BOX uom_id>",
  "barcode": "SKU001-BOX",
  "is_primary": true
}
```

Barcode `uom_id` is optional; if supplied, the UOM must be assigned and active
for this item. Barcodes are globally unique. To switch the primary, send
`PATCH /items/<item_id>/barcodes/<barcode_id>/primary` with no body.
Creating another primary barcode switches it atomically as well.

### 5. Update and deactivate safely

PUT is a full update of mutable fields, not a partial PATCH. Supply required
fields and `is_active`; omitted nullable fields are cleared. Codes, owner IDs,
and item base UOMs are immutable. A conversion's UOM is immutable too.
Unknown fields (including immutable fields) are rejected.

Item and business-partner updates/deactivations require the latest
`updated_at` as `expected_updated_at`:

```http
PUT /api/v1/master/items/<item_id>

{
  "name": "Packaging Tape - Updated",
  "category_id": "<category_id>",
  "weight": "0.250000",
  "volume": "0.001000",
  "lot_controlled": true,
  "serial_controlled": false,
  "shelf_life_days": 365,
  "minimum_receive_days": 90,
  "is_active": true,
  "expected_updated_at": "<latest updated_at>"
}
```

```http
PATCH /api/v1/master/items/<item_id>/deactivate

{"expected_updated_at":"<latest updated_at from PUT response>"}
```

Other catalog deactivate endpoints take no body. Deactivation is not cascading:
historical relationships remain. Restore a record through its PUT endpoint
with `is_active: true` and valid related references.

Reference create examples:

```json
{"code":"PAL","name":"Pallet","decimal_scale":0}
```

Send to `POST /uoms`. Partner-type and quality-status create requests use
`code/name/description`. Inventory-status requests additionally support
`is_allocatable/is_pickable`; inspection-result requests support `is_accepted`.

## Rules and expected failures

- Numeric measurements/conversions use **JSON strings**, never floating-point
  JSON numbers; maximum 14 integer digits and 6 fractional digits.
- Conversions must be positive. Dimensions, weight, and volume are non-negative.
- The base conversion row cannot be deactivated or changed from `1`.
- Deactivate active barcodes referencing a conversion before deactivating it.
- Category parent and item category must belong to the same owner. Self-parenting
  and indirect category cycles are rejected.
- Partner/item/category codes are unique per owner; reference codes are global.
- Minimum receive days cannot exceed supplied shelf-life days.
- At most one active barcode per item can be primary. Deactivating it clears
  primary; no replacement is selected automatically.
- New assignments require active related records.

Status codes: `201` create; `200` successful reads/updates; `400` invalid JSON,
IDs, or business rules; `401` invalid/missing session; `404` missing records or
child ID belonging to another item; `409` duplicates or stale timestamps.

Try duplicate partner codes/barcodes, changing the base conversion to `2`,
reusing an old timestamp, using a category from another owner, or selecting
an inactive primary barcode. These should fail without partially changing data.

## Automated tests

```powershell
cd C:\WmsProject\backend
go test ./...
go vet ./...
```

Database tests are opt-in and read `backend/.env` (environment values take
precedence). They use the configured existing PostgreSQL database:

```powershell
$env:WMS_INTEGRATION_TEST = '1'
go test ./... -count=1
Remove-Item Env:WMS_INTEGRATION_TEST
```

The integration suite tests the existing schema and a transaction-local empty
schema, repeats migrations and seeding, and rolls back all fixture/schema
changes even on failure. It requires schema-creation privileges for the empty
schema test. Do not run development migrations/tests against production.

Unit/HTTP tests cover numeric validation, strict request binding, and session
guards on every new route. Database tests cover partner types, owner/category
validation, audit timestamps, pagination totals, exact decimals, false boolean
flags, base conversions, barcode switching, and transactional rollback.

## Structure

The existing module grouping is preserved. Catalog endpoints live in
`routes/catalog_routes.go`; the master catalog controller handles JSON only;
`services/master/catalog_*` contains business rules and reference seeds.
Every table has one entity in `models/master` and one typed repository in
`repository/master`; shared query/transaction plumbing stays in repository.
Request/response structs remain separate in `dto/master`.
