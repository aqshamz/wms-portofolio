# Stock-control API

This slice provides safe, immediate stock commands backed by the inventory core.
Every command requires a valid bearer session, derives item/owner/warehouse/lot/HU/
status identity from the source balance, checks `expected_version`, and commits
balance changes plus immutable movement rows atomically.

These commands are operational posting commands, not approval documents. The
draft/approve/execute tables in `database/stock_control` remain the future
workflow layer for businesses that require maker/checker approval. Until role and
scope authorization exists, do not expose these endpoints as a production
multi-tenant API.

## Endpoints

```text
GET  /api/v1/stock-control/reason-codes?active=true
POST /api/v1/stock-control/internal-moves
POST /api/v1/stock-control/status-changes
POST /api/v1/stock-control/adjustments
POST /api/v1/stock-control/stock-count-reconciliations
POST /api/v1/stock-control/warehouse-transfers
```

Use `Content-Type: application/json` and `Authorization: Bearer <token>`.
Read `balance_id` and its current `version_no` from
`GET /api/v1/inventory/balances`. A 400 version error means reload the balance;
do not blindly replace the version.

Every command has:

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
Only unreserved quantity can move. Generic HU relocation is rejected because a
whole-HU workflow must move all nested contents and container locations together.
For a serialized item, post quantity 1 with exactly one `serial_id` per command.

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

```json
{
  "operation_key": "adjust:20260907:0001",
  "business_date": "2026-09-07",
  "source_document_id": "ADJUST-0001",
  "reason_code": "MANUAL_ADJUSTMENT",
  "notes": "Approved physical correction",
  "balance_id": "BAL-...",
  "direction": "DECREASE",
  "quantity": "1",
  "expected_version": 3,
  "serial_ids": []
}
```

Direction is exactly `INCREASE` or `DECREASE`. A reason is mandatory. An increase
applies only to an existing identity bucket; inbound should create first stock.
A decrease cannot consume reserved quantity.

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
