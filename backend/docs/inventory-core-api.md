# Inventory core

The inventory core is the shared stock foundation for stock control, inbound
and outbound. It owns current balances, the immutable movement ledger and the
current-state pointer for serialized units.

It intentionally has **no generic public posting endpoint**. A caller must first
validate a real business operation—receipt, QC result, move, adjustment, pick or
shipment—and then call `inventory.Service.PostMovement` inside the backend.
Allowing clients to directly edit balances or manufacture movements would bypass
document status, approvals and business rules.

## Data model

One `inventory_balance` row is the current base-UOM quantity for exactly:

```text
owner + warehouse + location + item + optional lot
+ optional handling unit + inventory status
```

NULL lot/HU values participate in uniqueness (`NULLS NOT DISTINCT`). A balance
stores non-negative `on_hand_qty` and `reserved_qty`; reserved cannot exceed on
hand. `available_qty` is returned as `on_hand_qty - reserved_qty`. Quantity
precision is `numeric(20,6)`. Zero balances are retained for traceability and
hidden from list queries unless `include_zero=true`.

`inventory_movement` is append-only application history. Each successful core
posting records its source/destination location and status, positive base-UOM
quantity, business date, source document, actor and optional traceability IDs.
The API has no update or delete route for movements.

`serial_inventory` contains one current-state row per serialized unit currently
in stock. It points at its balance; location, lot, handling unit, status and UOM
are derived through that balance instead of being duplicated. Shipping/removing
the serial deletes its current-state pointer but keeps its serial identity and
movement history.

## Read-only HTTP API

All endpoints require `Authorization: Bearer <session token>`.

| Endpoint | Purpose |
| --- | --- |
| `GET /api/v1/inventory/balances` | Paginated current balance inquiry |
| `GET /api/v1/inventory/balances/:id` | Balance detail |
| `GET /api/v1/inventory/movements` | Paginated immutable movement ledger |
| `GET /api/v1/inventory/movements/:id` | Movement detail |
| `GET /api/v1/inventory/serial-states` | Serialized units currently in stock |
| `GET /api/v1/inventory/serial-states/:serial_id` | One serial's current state |
| `GET /api/v1/inventory/movement-types?active=true` | Posting type references |

The previous lot, serial identity and handling-unit endpoints remain available.

### Balance filters

`page`, `page_size`, `owner_id`, `warehouse_id`, `location_id`, `item_id`,
`lot_id`, `handling_unit_id`, `inventory_status_id`, `search`, `include_zero`.

Example using the study owner and warehouse:

```http
GET /api/v1/inventory/balances?owner_id=5e6a104a-51c2-4e94-8643-e339bcd5242f&warehouse_id=59113433-3082-4208-87c1-330f4f56fcba&include_zero=false
Authorization: Bearer <token>
```

The collection is initially empty. Identity creation does not create stock.

### Movement filters

`page`, `page_size`, `owner_id`, `warehouse_id`, `item_id`, `movement_type_id`,
`source_document_id`, `operation_key`, `search`, `occurred_from`, `occurred_until`.
Occurrence filters are RFC3339 timestamps, the lower bound is inclusive, and the
upper bound is exclusive. Example: `2026-09-07T00:00:00+07:00`.

### Serial-state filters

`page`, `page_size`, `owner_id`, `warehouse_id`, `location_id`, `item_id`,
`lot_id`, `handling_unit_id`, `inventory_status_id`, `search`.

All collection endpoints use `page=1`, `page_size=20` by default and return
`data.items`, `total_items`, and `total_pages`. UUID filters are validated.
Unknown/repeated parameters, invalid booleans and invalid timestamps return 400.
The current session guards still do not enforce account owner/warehouse scopes;
that authorization layer must be added before production multi-tenant exposure.

## Internal posting contract

Domain services call the exported Go method, not HTTP:

```go
result, err := inventoryService.PostMovement(ctx, inventorydto.PostingRequest{
    OperationKey:     "receipt:RCPT-0001:line:1:execution:1",
    MovementTypeCode: "RECEIVE",
    OwnerID:           ownerID,
    WarehouseID:       warehouseID,
    BusinessDate:      "2026-09-07",
    ItemID:            itemID,
    LotID:             &lotID,
    To: &inventorydto.BalanceDimension{
        LocationID:       receivingLocationID,
        InventoryStatusID: qcPendingStatusID,
    },
    Quantity:         "24",
    SourceDocumentID: receiptID,
    SourceLineID:     &receiptLineID,
}, actorAccountID)
```

This example means 24 base units entered the receiving/QC_PENDING bucket. Source
documents may be expressed in BOX, but their service must convert to the item's
base UOM before posting. The core always uses `item.base_uom_id`; callers cannot
substitute another UOM.

Posting shapes:

| Shape | Meaning |
| --- | --- |
| `From=nil`, `To=value` | Quantity enters WMS stock (receipt/positive correction) |
| `From=value`, `To=nil` | Quantity leaves WMS stock (ship/dispose/negative correction) |
| Both present | Location or status transfer |

At least one source/destination must exist. If both exist, location or status
must actually change. Quantity must be positive and fit `numeric(20,6)`.

The service verifies active owner/item/warehouse/warehouse-owner assignment,
active movement type, master scope, item control flags, lot ownership, target
location/status state, HU scope/location and unreserved source availability.
Source rows may reference a master later deactivated, allowing controlled recovery;
new destinations must use active/unlocked targets.

### Idempotency and concurrency

`OperationKey` is mandatory, stable for one logical execution, globally unique,
and persisted with a SHA-256 fingerprint of normalized posting data. Retrying the
same key and same data returns the original movement with `IdempotentReplay=true`
without changing stock. Reusing the key for different data returns an error.

Use a key more specific than the document line because one line may post several
times, for example:

```text
<document-type>:<document-id>:<line-id>:<execution-id>
```

The repository takes a transaction-scoped PostgreSQL advisory lock per
owner/warehouse/item. It then locks relevant balance and serial-state rows,
updates/creates balances, changes serial state, and inserts the movement in one
database transaction. Any failure rolls back all parts. This prevents lost
updates and makes inverse operations use the same lock order.

Database constraints independently enforce unique identities, owner/item scope,
warehouse/location scope, lot/HU scope, valid quantities, serial-to-balance scope,
and unique operation keys. Migrations are additive; conflicting legacy data makes
startup fail for review rather than being silently repaired or deleted.

### Serialized posting rules

Because the existing movement schema has one `serial_id` per movement, a
serial-controlled posting must use exactly one serial and quantity exactly 1.
Post multiple units as separate operations with separate keys. On entry, the
serial must not already be in inventory; on transfer it must occupy the source
balance; on exit its state pointer is removed. Non-serial items reject serial IDs.

A lot-controlled item requires its matching owner/item lot on every posting.
A non-lot-controlled item rejects a lot. Generic posting cannot relocate an HU;
moving an HU must be implemented as a stock-control workflow that validates and
moves the complete HU hierarchy/content atomically. Generic status changes at
the HU's current location remain possible. Receiving into a closed HU is rejected.

Reservations are deliberately not mutated here yet. A debit cannot reduce on
hand below existing reserved quantity. Outbound reservation/allocation will add
a separate atomic reservation operation rather than pretending it is a movement.

## Movement-type seeds

Startup repeatably seeds without overwriting existing rows:

`RECEIVE`, `PUTAWAY`, `PICK`, `PACK`, `SHIP`, `TRANSFER`, `INTERNAL_MOVE`,
`TRANSFER_OUT`, `TRANSFER_IN`, `ADJUSTMENT`, `STATUS_CHANGE`,
`COUNT_CORRECTION`, `RETURN_TO_VENDOR`, `DISPOSE`, `DELIVERY_RETURN`, and
`OUTBOUND_CHECK_CORRECTION`.

## Verification

```powershell
cd C:\WmsProject\backend
go test ./...
go vet ./...
$env:WMS_INTEGRATION_TEST = '1'
go test -p 1 ./... -count=1
Remove-Item Env:WMS_INTEGRATION_TEST
```

Integration tests run in rollback/test-owned schemas. They cover entry,
status transfer, exit, decimal normalization, reserved quantity protection,
serial-state lifecycle, reference/control validation, read queries, repeatable
migration, operation-key replay and mismatch. The committed concurrency test
starts 12 distinct postings against one balance and 12 concurrent retries of one
logical posting; it verifies exact quantity and movement counts, then drops only
its uniquely named test schema.

Next, stock control should add document-backed internal moves, status changes,
adjustments and stock counts. Those services will call this core only after their
workflow/approval checks pass.
