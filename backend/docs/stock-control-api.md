# Stock-control API

This slice provides safe stock commands and an assigned replenishment workflow
backed by the inventory core.
Every command requires a valid bearer session, derives item/owner/warehouse/lot/HU/
status identity from the source balance, checks `expected_version`, and commits
balance changes plus immutable movement rows atomically. Commands derive their
owner and warehouse from the source balance and enforce the authenticated
account's owner/warehouse grants; unrestricted superadmins retain full access.

Internal moves, status changes, count reconciliations and immediate transfers
are operational postings. Inventory adjustments are approval documents: a
requester creates a DRAFT, a different account approves it, and only approval
posts the inventory movement.

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
POST /api/v1/stock-control/stock-count-reconciliations
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

## Stock-count reconciliation

```json
{
  "operation_key": "count:20260907:0001",
  "business_date": "2026-09-07",
  "source_document_id": "COUNT-0001",
  "reason_code": "COUNT_VARIANCE",
  "notes": "Physical recount confirmed",
  "balance_id": "BAL-...",
  "counted_qty": "9",
  "expected_version": 4,
  "serial_ids": []
}
```

The balance is locked and rechecked before comparison. A variance posts one
`COUNT_CORRECTION` movement; no variance returns `no_variance: true` and no
movement. A reason is required only when there is a variance. This immediate
endpoint reconciles one balance; blind-count assignments and review approval
belong to the later document workflow.

## Inter-warehouse transfer

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
or neither does. It models an immediate transfer, not an in-transit period.
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
