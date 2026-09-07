# Study inventory data

This additive development dataset lets you test inventory identity, balance
inquiry, movement history, serial current state, FEFO lots and the immediate
stock-control commands. It is persisted in PostgreSQL; it is not frontend dummy
data.

## Create or reuse the dataset

Start the backend, then run both scripts from another PowerShell window:

```powershell
cd C:\WmsProject\backend
go run .
```

```powershell
cd C:\WmsProject\backend
& .\scripts\seed-study-data.ps1 -BaseUrl http://localhost:8080 | Out-Null
$inventory = (& .\scripts\seed-study-inventory.ps1 | Out-String) | ConvertFrom-Json
$inventory | Select-Object created_identities, reused_identities, posted_movements, replayed_movements
$inventory.balances | Format-Table item_code, lot_number, location_code, inventory_status_code, on_hand_qty, version_no
```

If the acting account is not the bootstrap username, pass
`-ActorUsername YOUR_USERNAME` to the inventory script. The actor must already
exist. The first script uses HTTP master endpoints; the second uses the Go
inventory service because generic opening-balance mutation is intentionally not
public HTTP API.

Both scripts are safe to rerun. Inventory identities are found by their study
natural keys, and six stable posting operation keys replay rather than adding
quantity twice. A clean rerun reports zero created identities, seven reused
identities, zero posted movements, and six replayed movements.

## Dataset contents

| Item / identity | Location | Status | Opening quantity |
| --- | --- | --- | ---: |
| STUDY_COFFEE_250G / lot A / pallet | STUDY_BULK_01 | AVAILABLE | 240 EA |
| STUDY_COFFEE_250G / lot B | STUDY_PICK_01 | AVAILABLE | 96 EA |
| STUDY_COFFEE_1KG / lot A | STUDY_QC_01 | QC_PENDING | 60 EA |
| STUDY_SCANNER / three serials | STUDY_PICK_01 | AVAILABLE | 3 EA |

The two 250 g lots have different expiry dates for FEFO study. The scanner
balance is aggregated at quantity 3 while `serial_inventory` has three individual
current-state rows. The pallet balance demonstrates HU-linked inventory. Use the
non-HU lot-B balance for generic move and transfer commands; generic stock
control intentionally rejects whole-HU relocation.

`STUDY_WH_2` and its `STUDY_BULK_01` location are empty transfer targets. The
manifest printed by the inventory script contains every database-specific ID.

The current local database contains:

| Reference | ID |
| --- | --- |
| Owner | `5e6a104a-51c2-4e94-8643-e339bcd5242f` |
| Primary warehouse | `59113433-3082-4208-87c1-330f4f56fcba` |
| Transfer warehouse | `a75b156f-1dce-4289-b6f5-56043d3be890` |
| Primary bulk location | `80925dd4-eee2-468b-a830-5dfc90f28873` |
| Pick location | `8c4b818a-99fa-4fcb-80d4-150e40c30ac7` |
| QC location | `0646cfe1-0e81-44d1-b392-5ce8edd1a923` |
| Transfer bulk location | `1d90e747-e056-42fa-abca-a86e547324ab` |

IDs differ on a fresh database, so prefer the returned manifest instead of
copying this table elsewhere.

## Test through the API

Log in and set `Authorization: Bearer <token>`, then try:

```text
GET /api/v1/inventory/balances?owner_id=5e6a104a-51c2-4e94-8643-e339bcd5242f&page=1&page_size=100
GET /api/v1/inventory/movements?owner_id=5e6a104a-51c2-4e94-8643-e339bcd5242f&source_document_id=STUDY-OPENING&page=1&page_size=100
GET /api/v1/inventory/serial-states?owner_id=5e6a104a-51c2-4e94-8643-e339bcd5242f&page=1&page_size=100
GET /api/v1/inventory/lots?owner_id=5e6a104a-51c2-4e94-8643-e339bcd5242f&page=1&page_size=100
GET /api/v1/inventory/handling-units?owner_id=5e6a104a-51c2-4e94-8643-e339bcd5242f&page=1&page_size=100
GET /api/v1/stock-control/reason-codes?active=true
```

The verified opening state has four positive balances, six RECEIVE movements,
three serial states and total on-hand quantity 399 EA.

To test a stock command, first copy the latest `balance_id` and `version_no` from
the balance response. For example, move part of lot B from pick face to bulk:

```http
POST /api/v1/stock-control/internal-moves
Authorization: Bearer <token>
Content-Type: application/json

{
  "operation_key": "my-study-move-0001",
  "business_date": "2026-09-07",
  "source_document_id": "MY-STUDY-TEST",
  "reason_code": "RELOCATION",
  "source_balance_id": "BAL-477c02000ee1d1f2ba1ebc36762de788",
  "target_location_id": "80925dd4-eee2-468b-a830-5dfc90f28873",
  "quantity": "10",
  "expected_version": 2,
  "serial_ids": []
}
```

That exact balance/version is the initial local state. If it has already changed,
reload balances and use its current version. Always use a new operation key for
a new business action; repeat the exact same body only when retrying the same
action.

## Inventory boundary

The implemented inventory layer covers lot/serial/HU identities, balances,
immutable movements, serial current state, internal moves, status changes,
adjustments, single-balance count reconciliation, and immediate warehouse
transfer. Reservations/allocation, replenishment tasks, count documents with
approval/freeze, and dispatch/in-transit/receipt workflows belong to outbound,
task, or document orchestration and are not completed inventory APIs yet.
