# Stock-control API

This slice provides safe stock commands and an assigned replenishment workflow
backed by the inventory core.
Every command requires a valid bearer session, derives item/owner/warehouse/lot/HU/
status identity from the source balance, checks `expected_version`, and commits
balance changes plus immutable movement rows atomically. Commands derive their
owner and warehouse from the source balance and enforce the authenticated
account's owner/warehouse grants; unrestricted superadmins retain full access.

Internal moves, status changes and immediate transfers are operational
postings. Inventory adjustments and cycle counts are controlled documents: a
requester records the work, a different account reviews it, and only approval
posts an inventory correction.

## Endpoints

```text
GET  /api/v1/stock-control/reason-codes?active=true
POST /api/v1/stock-control/internal-moves
GET  /api/v1/stock-control/replenishment-tasks
POST /api/v1/stock-control/replenishment-tasks
GET  /api/v1/stock-control/replenishment-tasks/:id
GET  /api/v1/stock-control/replenishment-tasks/:id/assignees
POST /api/v1/stock-control/replenishment-tasks/:id/assign
POST /api/v1/stock-control/replenishment-tasks/:id/start
POST /api/v1/stock-control/replenishment-tasks/:id/complete
POST /api/v1/stock-control/replenishment-tasks/:id/cancel
POST /api/v1/stock-control/status-changes
GET  /api/v1/stock-control/adjustments
POST /api/v1/stock-control/adjustments
GET  /api/v1/stock-control/adjustments/:id
POST /api/v1/stock-control/adjustments/:id/approve
POST /api/v1/stock-control/adjustments/:id/reject
POST /api/v1/stock-control/adjustments/:id/cancel
GET  /api/v1/stock-control/cycle-counts
POST /api/v1/stock-control/cycle-counts
POST /api/v1/stock-control/cycle-counts/grand
GET  /api/v1/stock-control/cycle-counts/:id
POST /api/v1/stock-control/cycle-counts/:id/count
POST /api/v1/stock-control/cycle-counts/:id/approve
POST /api/v1/stock-control/cycle-counts/:id/reject
POST /api/v1/stock-control/cycle-counts/:id/cancel
GET  /api/v1/stock-control/warehouse-transfer-documents
POST /api/v1/stock-control/warehouse-transfer-documents
GET  /api/v1/stock-control/warehouse-transfer-documents/:id
POST /api/v1/stock-control/warehouse-transfer-documents/:id/approve
POST /api/v1/stock-control/warehouse-transfer-documents/:id/dispatch
POST /api/v1/stock-control/warehouse-transfer-documents/:id/cancel
POST /api/v1/stock-control/warehouse-transfer-documents/:id/receive
POST /api/v1/stock-control/warehouse-transfer-documents/:id/putaway
POST /api/v1/stock-control/warehouse-transfers
```

Use `Content-Type: application/json` and `Authorization: Bearer <token>`.
Read `balance_id` and its current `version_no` from
`GET /api/v1/inventory/balances`. A 400 version error means reload the balance;
do not blindly replace the version.

Every immediate posting command (excluding adjustment requests) has:

```json
{
  "operation_key": "unique-logical-execution-key",
  "business_date": "2026-09-07",
  "source_document_id": "YOUR-BUSINESS-REFERENCE",
  "source_line_id": "OPTIONAL-LINE-REFERENCE",
  "reason_code": "OPTIONAL_STABLE_CODE",
  "notes": "Required when the selected reason says requires_note"
}
```

`operation_key` is the retry/idempotency key. Same key and same normalized data
returns the original movement without applying quantity again. Same key with
different data is rejected. Transfer reserves `:out` and `:in` suffixes, so its
base key is limited to 150 characters.

## Internal move

Internal moves are restricted to `AVAILABLE` inventory moving between active,
unlocked locations whose active location type code is `STORAGE`. Receiving,
checking, pick-face replenishment, quarantine, packing and shipping transitions
remain owned by their dedicated workflows.

```json
{
  "operation_key": "move:20260907:0001",
  "business_date": "2026-09-07",
  "source_document_id": "MOVE-0001",
  "reason_code": "RELOCATION",
  "source_balance_id": "BAL-...",
  "target_location_id": "TARGET-LOCATION-UUID",
  "quantity": "10",
  "expected_version": 2,
  "serial_ids": []
}
```

The target is in the same warehouse and keeps the existing inventory status.
Only unreserved quantity can move. A handling unit can move only as a strict
whole-HU relocation when it has exactly one positive balance, no reservations
and no child handling units; its current location changes atomically with stock.
For a serialized item, post quantity 1 with exactly one `serial_id` per command.
Targets must be active and unlocked and must belong to an active zone and active
location type in the same warehouse.

## Replenishment tasks

Create a manual task from an available reserve balance:

```json
{
  "source_balance_id": "BAL-...",
  "target_location_id": "PICK-FACE-UUID",
  "quantity": "20",
  "priority_code": "NORMAL",
  "notes": "Refill before the afternoon wave",
  "expected_balance_version": 3
}
```

Creation reserves the planned base quantity without moving it. The source must
be active, unlocked, storage-capable, outside a pick face, and have an active
allocatable/pickable status. The target must be an active, unlocked pick face
whose location type allows storage and picking. Serialized tasks require
quantity 1 and one `serial_id`; a handling unit must be replenished as its one
full, unreserved balance with no children.

The lifecycle is `OPEN → ASSIGNED (optional) → IN_PROGRESS → COMPLETED`.
Starting an open task claims it for the caller. An assigned task may be started
only by its assignee, and only the assignee may complete an in-progress task.
Assignee lookup returns active accounts with matching owner/warehouse access and
effective `INVENTORY.MOVE` permission. Completion releases this task's
reservation and atomically posts a `REPLENISHMENT` movement into the pick face.
Cancellation releases the reservation without moving stock; in-progress work
may be cancelled only by its assignee.

## Status change

The UI is available under **Stock control → Inventory status changes** and
requires `INVENTORY.READ` plus `INVENTORY.STATUS_CHANGE`. Select an accessible
warehouse and served owner, filter by source status and location type, then
review and confirm the change. This is an immediate posting, not an adjustment
approval document. Movement history is the audit trail.

Manual status changes are restricted in both the UI and API to `AVAILABLE`,
`HOLD`, `DAMAGED` and `EXPIRED` in active, unlocked `STORAGE` or `PICK_FACE`
locations with an active zone and location type. The target must be active and
different from the source. QC, putaway, quarantine, vendor-return and outbound
workflow statuses cannot be bypassed through this endpoint.

Quantity is in base units and cannot exceed unreserved stock. For serialized
items, select one current serial and quantity 1 per command. A status change
keeps the existing location and HU identity, so loose or HU stock may be split
between statuses without physically relocating it. Reason is required; notes
are required when the selected inventory reason requires them. The UI retains
the exact command and retry key after a failure; refreshing stock requires a new
review. Grand-opname freeze and balance-version checks still apply.

```json
{
  "operation_key": "status:20260907:0001",
  "business_date": "2026-09-07",
  "source_document_id": "STATUS-0001",
  "reason_code": "STATUS_HOLD",
  "notes": "Packaging damaged; awaiting client decision",
  "source_balance_id": "BAL-...",
  "target_inventory_status_id": "TARGET-STATUS-UUID",
  "quantity": "5",
  "expected_version": 2,
  "serial_ids": []
}
```

Stock stays at the same location and changes status. A reason is mandatory.

## Adjustment

Creating a request requires `INVENTORY.ADJUST` and does not change stock:

```json
{
  "business_date": "2026-09-07",
  "direction": "DECREASE",
  "reason_code": "MANUAL_ADJUSTMENT",
  "notes": "Approved physical correction",
  "lines": [
    {
      "balance_id": "BAL-001",
      "quantity": "1",
      "expected_balance_version": 3
    },
    {
      "balance_id": "BAL-002",
      "quantity": "2",
      "serial_id": null,
      "expected_balance_version": 7
    }
  ]
}
```

Direction is exactly `INCREASE` or `DECREASE` and applies to all lines. The
document has one business date, reason and note, and may contain 1–100 distinct
inventory balances from one owner and warehouse. A serialized line carries one
`serial_id` and quantity 1. Adjustment lines are restricted to active,
unlocked locations whose location type is exactly `STORAGE`; receiving,
quarantine, QC, dock, pick-face and shipping balances stay under their dedicated
workflows. A decrease cannot consume reserved quantity.

Approval requires `INVENTORY.ADJUST_APPROVE`:

```json
{
  "expected_version": 1,
  "line_ids": ["IADJ-...-L0001", "IADJ-...-L0003"]
}
```

The requester cannot approve their own request, including when they have the
wildcard permission. Approval rechecks every selected line and posts the
selection atomically; a changed balance causes the whole selected batch to roll
back. Unselected lines remain pending and the document becomes
`PARTIALLY_POSTED`. Reject accepts `expected_version`, `line_ids`, and a
mandatory `reason`. Cancel is restricted to the requester and cancels all
remaining pending lines. Neither rejection nor cancellation reverses lines that
were already posted.

## Cycle counts

Create a document with a business date, absolute tolerance quantity, optional
notes, and 1-100 distinct `balance_ids`. Creation requires
`INVENTORY.COUNT`, captures immutable quantity and version snapshots, and is
restricted to active, unlocked locations whose type is exactly `STORAGE`.
All balances must belong to one owner and warehouse. The counter sees item and
location identity but not the snapshot quantity or variance while the document
is in `DRAFT` or `COUNTING`.

Grand stock opname uses the same blind-count, recount, independent approval and
partial posting workflow, but creation is automatic:

```json
{
  "owner_id": "CUSTOMER-UUID",
  "warehouse_id": "WAREHOUSE-UUID",
  "business_date": "2026-10-05",
  "tolerance_quantity": "0",
  "notes": "Annual customer stock opname"
}
```

`POST /cycle-counts/grand` snapshots every positive balance for that one active
warehouse/served-owner relationship in active, unlocked `STORAGE` and
`PICK_FACE` locations. Receiving, QC, quarantine, dock and transfer
locations are excluded. Only one unfinished grand opname may exist for the same
warehouse and owner; a second customer served by the warehouse requires a
separate grand opname. List ordinary and grand documents independently with
`count_type_code=CYCLE` or `count_type_code=GRAND`.

Creating a grand opname immediately freezes inventory for that entire customer
and warehouse. Finish or cancel existing stock reservations/replenishments first;
creation is rejected while reserved quantities remain. Snapshot creation and
freeze activation share a transaction and an exclusive scope lock, so a stock
write cannot slip between them.

The freeze blocks new stock balances, receipts, putaway, internal moves,
replenishment reservations/releases, adjustments, status changes, transfer
dispatch/receipt, returns, disposal, and outbound allocation/pick/pack/ship in
that scope. It covers all locations, including locations excluded from the
storage count, and items/balances that were not in the original snapshot.
Reads and document preparation remain available. Other customers in the same
warehouse and the same customer in other warehouses are unaffected.

Only approved corrections belonging to that grand opname can change its stock
during the freeze, including serial reconciliation. A client-supplied source
document reference does not grant an exception, and superadmin access does not
bypass the freeze. Blocked writes return HTTP 409 and identify the freezing
document. Retry of an already-posted idempotent command is still allowed because
it does not change stock.

`stock_frozen` in the list/detail response indicates an active grand opname.
Partial posting keeps the freeze. Finalizing all lines, including rejected
lines, or cancelling the remaining lines releases the freeze atomically with
the document transition. Already-posted corrections remain posted on cancellation.
Ordinary cycle counts do not freeze stock. Migration 25 activates freezes for
existing unfinished grand opnames too; if their pre-migration snapshots are
already stale, cancel them and create a fresh grand opname.

Found stock can be recorded on either an open cycle count or an open grand
opname through `POST /cycle-counts/:id/found-stock` (`INVENTORY.COUNT`). Supply
`expected_version`, `item_id`, `location_id`, `inventory_status_id`, positive
`counted_quantity` in base units, and required discovery `notes`. Optional
`lot_id` and `handling_unit_id` identify the actual stock. The item must belong
to the count's customer; the location must be active, unlocked STORAGE (or
PICK_FACE for a grand opname). Lots must belong to the item/customer and are
required for lot-controlled items. Select an open, unnested handling unit
already at that location, or leave it empty for loose stock.

`GET /cycle-counts/:id/found-stock-options` uses the count's authorized scope
and supports `item_id`, `location_id`, `item_search`, `location_search`,
`lot_search`, and `hu_search`. Lookup lists are limited to 100 matches plus the
selected item/location; use search for larger catalogs. A new lot must first
be registered through Inventory lots.

Recording found stock adds a flagged count line and an immutable first count
entry with system quantity zero. It does not create inventory or release the
freeze. Its `balance_id` can be null until posting; `resulting_balance_id`
links the actual balance after approval. `discovery_notes` and
`created_by_display_name` preserve the original discovery across recounts.
An existing zero balance retains its version snapshot. An existing positive
balance or duplicate count identity is rejected: count the existing balance
instead. Stock appearing after discovery prevents posting from stale data.

The count's existing tolerance/recount and partial-review rules apply.
`INVENTORY.COUNT_APPROVE` is required to post or reject found stock, and its
recorder, latest counter, and document creator cannot review their own line.
Serialized discoveries must use whole quantities (up to 5000 base units) and
the serial reconciliation action with the exact found serial numbers.
Posting creates a `COUNT_CORRECTION` movement; rejecting or cancelling the
line creates no stock. Pending found lines keep a grand opname frozen.
Migration 26 adds nullable balance snapshots and found-line identity checks.
A grand opname may start with an empty snapshot to capture found stock; if
none is discovered, cancel the empty document with an explanatory reason.

Record any subset of pending lines through the `count` action using the
document `expected_version` and entries containing `line_id`,
`counted_quantity`, and optional notes. Every submission is retained as an
immutable count attempt. A first count whose absolute variance exceeds the
document tolerance requires a second count. Once all lines are counted and
required recounts are complete, the document enters `REVIEW` and reveals its
snapshots and variances.

Approval requires `INVENTORY.COUNT_APPROVE`, must be performed by a different
account, and accepts `expected_version` plus selected `line_ids`. Each
selected variance is rechecked against the captured balance version and posts
one `COUNT_CORRECTION` movement using reason `COUNT_VARIANCE`. A zero-variance
line is finalized without a movement. Unselected lines remain pending in
`PARTIALLY_POSTED`, so review can be completed in batches. Rejection also
finalizes selected lines without changing stock. Cancellation is limited to the
creator and affects only remaining lines.

Serialized zero-variance lines can be finalized, but a serialized variance is
rejected until it is resolved through serial reconciliation; the quantity
ledger is never changed without matching serial state.

## Inter-warehouse transfer

Normal transfers use `/warehouse-transfer-documents` and move through
`DRAFT -> APPROVED -> IN_TRANSIT -> RECEIVED`. Dispatch posts `TRANSFER_OUT`.
Destination receipt posts `TRANSFER_IN` into `PUTAWAY_PENDING` at an active
receiving-capable location, and the final putaway action moves it to the stored
STORAGE target as `AVAILABLE`. The source and target warehouses must both
actively serve the inventory owner. Receipt and putaway use the destination
account's warehouse scope.

The legacy atomic endpoint below is an administrative exception guarded by
`INVENTORY.TRANSFER_FORCE`; it is not used by the normal UI.

```json
{
  "operation_key": "transfer:20260907:0001",
  "business_date": "2026-09-07",
  "source_document_id": "TRANSFER-0001",
  "source_balance_id": "BAL-...",
  "target_warehouse_id": "TARGET-WAREHOUSE-UUID",
  "target_location_id": "TARGET-LOCATION-UUID",
  "target_inventory_status_id": "TARGET-STATUS-UUID",
  "quantity": "2",
  "expected_version": 5,
  "serial_ids": []
}
```

Source and target warehouses must differ and both must be assigned to the owner.
The command atomically posts `TRANSFER_OUT` and `TRANSFER_IN`; either both commit
or neither does. Source stock must be `AVAILABLE`, both locations must be
active, unlocked locations whose type is exactly `STORAGE`, and the target
inventory status must equal the source status. It models an immediate transfer,
not an in-transit period.
Carrier, dispatch, receiving variance, partial receipts and approval use the
existing transfer-order schema in a future document workflow. HU transfers are
rejected; lot identity is preserved. Serialized units exit and re-enter state in
the same database transaction.

## Responses and rules

Successful commands return 201 with `operation`, `movements`, source/destination
balances where applicable, and `idempotent_replay`. Quantities are base-UOM
strings with six decimals. Read movement details through the inventory API.

- 400: invalid/stale request, incompatible scope, insufficient unreserved stock.
- 401: invalid, expired or revoked app session.
- 404: source balance not found.
- 409: idempotency or unique-data conflict.

Unknown/trailing JSON and request bodies above 1 MiB are rejected. Inventory
reason seeds are repeatable and preserve existing edits. Codes include DAMAGE,
EXPIRY, COUNT_VARIANCE, REPLENISHMENT, RELOCATION, CONSOLIDATION, STATUS_HOLD,
STATUS_RELEASE, MANUAL_ADJUSTMENT and TRANSFER_VARIANCE.

## Verification

```powershell
cd C:\WmsProject\backend
go test ./...
go vet ./...
$env:WMS_INTEGRATION_TEST = '1'
go test -p 1 ./... -count=1
Remove-Item Env:WMS_INTEGRATION_TEST
```

The integration suite executes receipt seed → internal move → status change →
increase/decrease adjustment → count correction → two-sided warehouse transfer
against existing and fresh schemas, then rolls everything back. It verifies
stale-version errors, retries, exact totals and movement counts.

Next: inbound documents can safely call inventory posting for receive/QC/putaway;
outbound can add reservation/allocation before pick/pack/ship.
