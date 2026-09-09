# Outbound part 1: delivery order through staging

This slice implements the first half of outbound fulfillment:

```text
OUTBOUND: DRAFT -> VALIDATED -> RELEASED -> PARTIALLY_ALLOCATED | ALLOCATED
                                                        |
                                                        v
                                                    WAVED -> PICKING -> STAGED

RESERVATION: ACTIVE -> PARTIALLY_PICKED -> CONSUMED
OUTBOUND_WAVE: DRAFT -> RELEASED -> IN_PROGRESS -> COMPLETED
PICK_TASK: OPEN -> IN_PROGRESS -> COMPLETED
OUTBOUND_STAGING: OPEN -> COMPLETED
```

Part 2 will continue from `STAGED` with picked-stock verification, packing,
shipping, delivery, and negative delivery outcomes.

## Important behavior

- Delivery-order quantities always use each item's base UOM.
- Validation writes an immutable run and one result per configured rule. Blocking
  failures leave the order in `DRAFT`.
- Allocation follows the best owner/warehouse picking strategy, including FEFO,
  FIFO, and location sequence rules. It selects only active, allocatable,
  pickable stock in an active pick-enabled location.
- Allocation increases `reserved_qty`; it does not reduce `on_hand_qty` and does
  not create an inventory movement.
- `allow_partial: false` makes an inventory shortage roll back the complete
  allocation. `allow_partial: true` keeps available reservations and returns the
  exact shortage.
- A wave accepts only fully `ALLOCATED` orders from the same owner and warehouse.
  Releasing it creates one pick task per active reservation.
- Starting the first task advances the wave to `IN_PROGRESS` and its order to
  `PICKING`.
- Confirming a pick atomically reduces source on-hand and reserved quantities,
  increases stock at the wave staging location, writes a `PICK` movement, writes
  a pick execution, consumes reservation quantity, and updates task/order totals.
- `operation_key` makes a pick confirmation safe to retry. Reusing the key with
  different data returns HTTP `409`.
- `expected_version` protects order and wave transitions. Pick confirmation uses
  `expected_balance_version` to reject stale stock reads.
- A short pick releases its unpicked reservation and requires an active
  `OUTBOUND` reason such as `SHORT_PICK`.
- Completing the staging document advances the delivery order to `STAGED`.
- Every route requires a valid, non-revoked Bearer session.

## Routes

```text
POST /api/v1/outbound/orders
GET  /api/v1/outbound/orders?owner_id=<uuid>&warehouse_id=<uuid>&status_code=<code>&search=<text>&page=1&page_size=20
GET  /api/v1/outbound/orders/:id
POST /api/v1/outbound/orders/:id/validate
POST /api/v1/outbound/orders/:id/release
POST /api/v1/outbound/orders/:id/allocate
POST /api/v1/outbound/orders/:id/cancel
GET  /api/v1/outbound/validation-runs/:id

GET  /api/v1/outbound/reservations?owner_id=<uuid>&warehouse_id=<uuid>&status_code=<code>&page=1&page_size=20
POST /api/v1/outbound/reservations/:id/release

POST /api/v1/outbound/waves
GET  /api/v1/outbound/waves?owner_id=<uuid>&warehouse_id=<uuid>&status_code=<code>&page=1&page_size=20
GET  /api/v1/outbound/waves/:id
POST /api/v1/outbound/waves/:id/release
POST /api/v1/outbound/waves/:id/cancel

GET  /api/v1/outbound/pick-tasks?owner_id=<uuid>&warehouse_id=<uuid>&status_code=<code>&assignee_id=<uuid>&page=1&page_size=20
GET  /api/v1/outbound/pick-tasks/:id
POST /api/v1/outbound/pick-tasks/:id/start
POST /api/v1/outbound/pick-tasks/:id/confirm
POST /api/v1/outbound/pick-tasks/:id/short-close

POST /api/v1/outbound/stagings
GET  /api/v1/outbound/stagings?owner_id=<uuid>&warehouse_id=<uuid>&status_code=<code>&page=1&page_size=20
GET  /api/v1/outbound/stagings/:id
POST /api/v1/outbound/stagings/:id/complete
```

List endpoints require both `owner_id` and `warehouse_id`. Unknown JSON fields, extra JSON values,
oversized bodies, unknown query parameters, and repeated query parameters are
rejected.

## Test with the study data

Start the API and run both additive study-data scripts:

```powershell
cd C:\WmsProject\backend
go run .
```

```powershell
cd C:\WmsProject\backend
.\scripts\seed-study-data.ps1 -BaseUrl http://localhost:8080 | Out-Null
.\scripts\seed-study-inventory.ps1 | Out-Null
```

Log in through `POST /api/v1/auth/login`, copy `data.token`, and send:

```text
Authorization: Bearer <token>
Content-Type: application/json
```

Discover current IDs rather than copying UUIDs from another database:

```text
GET /api/v1/master/organizations?search=STUDY_OWNER&active=true
GET /api/v1/master/warehouses?search=STUDY_WH&active=true
GET /api/v1/master/business-partners?owner_id=<owner_uuid>&search=STUDY_CUSTOMER&active=true
GET /api/v1/master/business-partners?owner_id=<owner_uuid>&search=STUDY_STORE&active=true
GET /api/v1/master/items?owner_id=<owner_uuid>&search=STUDY_COFFEE_250G&active=true
GET /api/v1/master/picking-strategies?owner_id=<owner_uuid>&warehouse_id=<warehouse_uuid>&search=STUDY_FEFO&active=true
GET /api/v1/master/warehouses/<warehouse_uuid>/locations
GET /api/v1/inventory/balances?owner_id=<owner_uuid>&warehouse_id=<warehouse_uuid>&item_id=<item_uuid>&page=1&page_size=100
```

Use the non-HU `AVAILABLE` balance for lot B at `STUDY_PICK_01`. Use
`STUDY_STAGE_01` as the staging location.

### 1. Create and validate the delivery order

```http
POST /api/v1/outbound/orders

{
  "owner_id": "<owner_uuid>",
  "customer_id": "<STUDY_CUSTOMER_uuid>",
  "ship_to_partner_id": "<STUDY_STORE_uuid>",
  "warehouse_id": "<warehouse_uuid>",
  "business_date": "2026-09-08",
  "client_delivery_order_no": "CLIENT-DO-STUDY-001",
  "customer_order_no": "SO-STUDY-001",
  "requested_ship_at": "2026-09-08T15:00:00+07:00",
  "notes": "Outbound part 1 study",
  "lines": [
    {
      "item_id": "<item_uuid>",
      "ordered_qty": "10",
      "requested_lot_no": "STUDY-250G-2026-08-B"
    }
  ]
}
```

Save `data.outbound_id` and `data.version_no`, then validate:

```http
POST /api/v1/outbound/orders/<outbound_id>/validate

{
  "expected_version": 1,
  "notes": "Initial validation"
}
```

The result should be `PASSED` with five rule results. Reload the order; it is
`VALIDATED` with version 2.

### 2. Release and allocate

```http
POST /api/v1/outbound/orders/<outbound_id>/release

{
  "expected_version": 2
}
```

```http
POST /api/v1/outbound/orders/<outbound_id>/allocate

{
  "expected_version": 3,
  "picking_strategy_id": "<STUDY_FEFO_uuid>",
  "allow_partial": false
}
```

Save the returned reservation ID. Verify the order is `ALLOCATED`,
`allocated_qty` is 10, and `shortage_qty` is 0. Inventory inquiry should show
the selected source still has the same on-hand quantity but its reserved
quantity increased by 10.

### 3. Create and release a wave

```http
POST /api/v1/outbound/waves

{
  "owner_id": "<owner_uuid>",
  "warehouse_id": "<warehouse_uuid>",
  "business_date": "2026-09-08",
  "wave_type_code": "SINGLE_ORDER",
  "picking_strategy_id": "<STUDY_FEFO_uuid>",
  "outbound_ids": ["<outbound_id>"],
  "notes": "Study wave"
}
```

```http
POST /api/v1/outbound/waves/<wave_id>/release

{
  "expected_version": 1,
  "staging_location_id": "<STUDY_STAGE_01_uuid>",
  "priority_code": "NORMAL"
}
```

The response reports one pick task. Retrieve it from the wave or from the
pick-task list and save its `balance_version`.

### 4. Pick into staging

Starting a task has no request body:

```http
POST /api/v1/outbound/pick-tasks/<pick_task_id>/start
```

Confirm the physical pick:

```http
POST /api/v1/outbound/pick-tasks/<pick_task_id>/confirm

{
  "picked_qty": "10",
  "expected_balance_version": <balance_version>,
  "business_date": "2026-09-08",
  "operation_key": "study-pick-client-do-001"
}
```

The source balance now has on-hand and reserved quantities reduced by 10. A
balance at `STUDY_STAGE_01` has quantity 10, the reservation and task are final,
and the wave is `COMPLETED`. Repeating the exact confirmation returns the same
movement and execution.

For a physical shortage, confirm any quantity actually picked first, then close
the remainder:

```http
POST /api/v1/outbound/pick-tasks/<pick_task_id>/short-close

{
  "reason_code": "SHORT_PICK",
  "notes": "Two units missing at the source location"
}
```

### 5. Record and confirm staging

```http
POST /api/v1/outbound/stagings

{
  "outbound_id": "<outbound_id>",
  "wave_id": "<wave_id>",
  "notes": "Counted at outbound staging"
}
```

Complete it with no request body:

```http
POST /api/v1/outbound/stagings/<staging_id>/complete
```

The staging document is now `COMPLETED` and the delivery order is `STAGED`,
ready for outbound part 2.

## Automated verification

```powershell
cd C:\WmsProject\backend
go test ./...
```

The full PostgreSQL test runs the workflow in both the current schema and a
fresh temporary schema, with every change rolled back afterward:

```powershell
$env:WMS_INTEGRATION_TEST='1'
go test -p 1 ./services/outbound -count=1 -v
```
