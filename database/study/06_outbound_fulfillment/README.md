# Study Guide 06: Outbound — Delivery Order to Store Delivery

The outbound process has three different kinds of records:

| Layer | Meaning | Examples |
|---|---|---|
| Demand | What the client wants delivered | `outbound_order`, `outbound_order_line` |
| Warehouse work | What stock is promised and what staff must do | reservation, wave, pick, staging, check, packing |
| Dispatch and delivery | What left the warehouse and what happened at the store | shipment, delivery, events, returns |

The complete intended flow is:

```text
Client delivery order
  -> outbound validation
  -> release
  -> inventory allocation / reservation
  -> outbound wave
  -> pick tasks
  -> pick execution into staging
  -> outbound staging
  -> outbound check
       -> pass: packing
       -> fail: correction -> recheck
  -> shipment / dispatch
  -> delivery attempt
       -> delivered with POD
       -> failed -> retry or return to depot
```

## 1. Outbound order: the client delivery order

Yes, `outbound_order` is the internal WMS representation of the client's
delivery order (DO).

Important header fields include:

| Field | Purpose |
|---|---|
| `owner_id` | Client whose inventory will be delivered |
| `customer_id` | Commercial customer receiving the goods |
| `warehouse_id` | Warehouse that will fulfil the order |
| `client_delivery_order_no` | Original client DO number; unique per owner |
| `ship_to_partner_id` | Store/customer master, when available |
| `ship_to_*` | Address snapshot used for this particular delivery |
| `requested_ship_at` | Requested dispatch time |
| `status_id` | Current workflow state |

The address fields are intentionally copied onto the order. If the store
master address changes later, an old delivery must still show where it was
originally sent.

`outbound_order_line` contains the requested items. Its quantity columns show
progress through the process:

```text
ordered_qty
  >= allocated_qty
  >= picked_qty
  >= checked_qty
  >= packed_qty
  >= shipped_qty
  >= delivered_qty
```

`requested_lot_no` is optional. When it is populated, allocation may select
only inventory from that lot.

## 2. What is outbound validation?

Outbound validation is the pre-allocation business check. It answers:

> Is this DO complete and valid enough for the warehouse to promise stock and
> begin fulfilment?

It does not move or reserve inventory.

The validation evidence is stored in:

```text
outbound_validation_run
  -> outbound_validation_result_detail
       -> outbound_validation_rule
       -> validation_severity
```

The starter rules are:

| Rule | Severity | Check |
|---|---|---|
| `DO_NUMBER` | Error | Client DO number is present |
| `SHIP_TO` | Error | Ship-to partner/name/address are present |
| `ORDER_LINES` | Error | At least one order line exists |
| `BASE_UOM` | Error | Each line uses the item's base UOM |
| `REQUESTED_SHIP` | Warning | Requested ship time is present |

An `ERROR` blocks processing. A `WARNING` is recorded but does not block it.
If no blocking rule fails, the DO becomes `VALIDATED`; it can then be released
to `RELEASED` for allocation.

This validation history is useful because it records which rules ran, which
passed, which failed, and the message shown to the operator.

Each rule now has a stable `handler_code`. The editable rule code/name,
severity, order, and active flag remain master data, while `handler_code`
selects reviewed executable logic. Adding a row still requires implementing a
handler; business configuration is not treated as executable SQL.

## 3. Inventory reservation is allocation, not a picking list

`inventory_reservation` means:

> This exact quantity from this exact `inventory_balance` is promised to this
> outbound order line.

It is not the warehouse operator's picking list. The actual work instruction
is `pick_task`.

One line can require several reservations because stock may be split by
location, lot, handling unit, or inventory status:

```text
DO line: JUICE, 100 EA

Reservation 1 -> Balance A / LOT-01 / BIN-A01 -> 60 EA
Reservation 2 -> Balance B / LOT-02 / BIN-B04 -> 40 EA
```

When 60 units are reserved from this balance:

```text
Before: on_hand = 100, reserved = 0,  available = 100
After:  on_hand = 100, reserved = 60, available = 40
```

Allocation immediately:

- Inserts `inventory_reservation`.
- Increases `inventory_balance.reserved_qty`.
- Increases `outbound_order_line.allocated_qty`.
- Leaves `on_hand_qty` unchanged because no physical movement happened.

Only inventory statuses with `is_allocatable = true` are candidates. The
picking strategy can order candidates by FEFO, FIFO, location sequence, or
other configured rules.

The current workflow allows a DO into a wave only after every line is fully
allocated.

## 4. Outbound wave and wave orders

`outbound_wave` groups work that should be released to pickers together. A
wave can represent one order, a manual group, a route, a carrier group, or a
cut-off group.

`outbound_wave_order` is only a bridge:

```text
outbound_wave many <-> many outbound_order
```

It identifies which DOs belong to a wave. It is not another delivery order.

When a wave is released, the system generates one `pick_task` for each active
reservation. Therefore the example above creates two pick tasks, one for
60 EA from BIN-A01 and another for 40 EA from BIN-B04.

## 5. Pick task and pick execution

There is deliberately no `pick_task_line` table in the present design. One
`pick_task` is already one work line and points to:

- One reservation.
- One outbound order line.
- One source location.
- One staging target location.
- One planned quantity and UOM.
- An assignee, priority, status, and optional short-pick result.

`pick_execution` records an actual scan/movement. One task can have multiple
execution rows when it is picked partially.

When a pick is confirmed, the database atomically:

1. Decreases source `on_hand_qty`.
2. Decreases source `reserved_qty` by the same quantity.
3. Creates or increases an inventory balance at the staging location.
4. Writes an `inventory_movement` with type `PICK`.
5. Inserts `pick_execution`.
6. Updates picked quantities on the task, reservation, and DO line.

Example:

```text
Before picking from BIN-A01:
  on_hand = 100, reserved = 60, available = 40

After picking 60 into STAGE-01:
  BIN-A01:  on_hand = 40, reserved = 0, available = 40
  STAGE-01: on_hand = 60, reserved = 0
```

A short pick closes the task, records a reason, releases the unpicked reserved
quantity, and reduces `allocated_qty`. The shortage then requires a business
choice: reallocate and create replacement work, accept a partial/backorder, or
cancel the remaining demand.

## 6. Outbound staging

After all pick tasks for one DO and wave are final, the system creates:

```text
outbound_staging
  -> outbound_staging_line
       -> pick_execution
       -> staging inventory balance
```

The header says which DO, wave, warehouse, and staging lane are being
consolidated. Each line proves which pick execution contributed stock.

Staging does not make another physical movement: the stock was already moved
into the staging location by the `PICK` posting. The staging document groups
and confirms those executed picks before checking.

## 7. Outbound checking

Checking verifies the staged goods before packing. The model is:

```text
outbound_check
  -> outbound_check_line
       -> outbound_staging_line
       -> outbound_check_result
```

The starter result codes are:

| Result | Pass? | Meaning |
|---|---:|---|
| `PASS` | Yes | Item and quantity match |
| `SHORT` | No | Checked quantity below expected |
| `OVER` | No | Checked quantity above expected |
| `WRONG_ITEM` | No | Wrong item staged |
| `DAMAGED` | No | Stock damaged before packing |

A check passes only when every line has a passing result and
`checked_qty = expected_qty`. A failed check changes the DO to
`CHECK_FAILED`. Packing cannot start from a failed check.

### What should happen when checking fails?

The result should determine a corrective operation:

| Failure | Recommended response |
|---|---|
| `SHORT` | Find/reallocate replacement stock and pick it, or approve a partial/backorder |
| `OVER` | Return excess stock to the correct location, then recount |
| `WRONG_ITEM` | Move the wrong stock back/hold it and pick the correct item |
| `DAMAGED` | Change damaged stock to a non-allocatable status/quarantine and allocate a replacement |

After correction, create another `outbound_check` whose `parent_check_id`
references the failed check. This preserves the failed attempt and its recheck
history instead of overwriting it.

Every failed line now creates `outbound_check_exception`. Its actions are
recorded independently through `outbound_check_resolution` and linked to the
actual inventory movement or replacement reservation. Supported paths are:

- `SHORT`: remove phantom balance quantity, then replace it or record an
  approved `ACCEPT_SHORT` decision.
- `DAMAGED` / `WRONG_ITEM`: move affected stock to a non-allocatable location
  and status, then allocate and pick replacement stock.
- `OVER`: register the discovered excess directly in non-allocatable stock for
  investigation without reducing the valid staged order quantity.

Replacement pick executions are appended to the existing staging document.
`outbound_staging_line.removed_qty` preserves the original staging evidence
while identifying quantity excluded from the recheck. A recheck is blocked
until all parent-check exceptions have status `RESOLVED`.

## 8. Packing

Packing can be created only from a `PASSED` check.

```text
packing
  -> packing_line
       -> passed outbound_check_line
       -> pick_task
       -> source staging balance
       -> packing balance / handling unit
```

Packing one line moves checked stock from the staging balance into the packing
location or package/handling unit and writes an inventory movement of type
`PACK`.

The document is complete when every passed check line has a corresponding
packing line. The DO then becomes `PACKED`.

`handling_unit_id` can represent a carton, tote, pallet, or other physical
package. It is optional in the current schema.

## 9. Shipment, shipment line, and shipment packing

`shipment` is a manifest for one owner and warehouse. It may contain several
packed DOs through `shipment_order`. It records:

- Carrier service.
- Tracking number.
- Vehicle number.
- Seal number.
- Dispatch time and dispatching account.

`shipment_packing` links a shipment to completed packing documents. It does
not select the carrier and it is not a driver table.

`shipment_line` is created for each packed line that actually leaves the
warehouse. Dispatch:

- Decreases the packing-location inventory balance.
- Writes an `inventory_movement` of type `SHIP`.
- Increases `outbound_order_line.shipped_qty`.

After every attached packing line has shipped, the shipment and DO become
`SHIPPED`. Shipped inventory is no longer part of warehouse
`inventory_balance`.

### Is carrier the driver?

No. `carrier` is the transport company or internal fleet, for example:

```text
Carrier:         INTERNAL_FLEET
Carrier service: JAKARTA_SAME_DAY

Carrier:         JNE
Carrier service: REGULAR
```

A driver is a person assigned to the trip. Drivers and assignments are now
normalized as:

```text
carrier_driver
  driver_id
  carrier_id
  code
  name
  phone_number
  license_number
  is_active

shipment -> shipment_driver -> carrier_driver
```

If drivers already log into the WMS, an optional `account_id` can link the
driver master to `app_account` without assuming that every external driver
needs a WMS login.

`carrier.business_partner_id` optionally links the transport master to the
corresponding business partner, preventing duplicate identity when that
company also participates in partner workflows.

`shipment_order` is the manifest membership bridge, so one truck can carry
several DOs. Each `delivery` still belongs to one DO/store stop, keeping POD
and failure outcomes separate per destination.

## 10. Delivery, POD, failure, retry, and return

`delivery` represents one DO/store stop in a shipped manifest.
`delivery_event` is its chronological operational history, and
`delivery_event_line` records the quantity accepted under each POD event.

Starter event types are:

```text
DEPARTED
ARRIVED_STORE
DELIVERED
DELIVERY_FAILED
RETURNED_TO_DEPOT
```

The event type records what occurred. A failed event also references a master
`delivery_failure_reason`, such as:

- `STORE_CLOSED`
- `RECIPIENT_REJECTED`
- `ADDRESS_NOT_FOUND`
- `DAMAGED_IN_TRANSIT`
- `VEHICLE_ISSUE`
- `OTHER`

On success, `delivery` stores proof of delivery (POD): recipient, recipient
reference, proof reference/URI, time, and optional coordinates. The DO line's
`delivered_qty` becomes its shipped quantity and the DO becomes `DELIVERED`.

### What happens after a failed delivery?

Failure does not immediately put stock back into warehouse inventory. The
goods are still physically on the vehicle or in transport custody. Two paths
are possible:

```text
FAILED -> retry departure/arrival -> DELIVERED

FAILED -> physically return to depot
       -> one or more delivery_return_line rows per delivery line
       -> returned inventory balance + DELIVERY_RETURN movement
       -> RETURNED or PARTIALLY_DELIVERED
```

There is no separate `delivery_return` header in the present model;
`delivery_return_line` belongs directly to the failed delivery.

Each return posting may return part or all of the undelivered remainder.
`outbound_return_policy` supplies the configured location and inventory status
for that owner/warehouse, and the query rejects an allocatable target status.
The returned stock therefore enters `HOLD` (or another configured
non-allocatable status) until inspected.

## 11. Complete quantity example

Suppose the client DO requests 100 EA of `APPLE_JUICE_1L`.

```text
1. DO line created
   ordered = 100

2. Validation passes and DO is released
   no inventory changes

3. Allocation
   reservation A = 60 from BIN-A01 / LOT-01
   reservation B = 40 from BIN-B04 / LOT-02
   allocated = 100
   source on-hand unchanged; source reserved increases by 100

4. Wave release
   two pick tasks are created

5. Pick confirmation
   60 + 40 moves to STAGE-01
   picked = 100
   source on-hand and reserved both decrease by 100

6. Staging
   staging header groups the two pick executions

7. Check
   both lines pass and total checked = 100

8. Packing
   100 moves from staging into packing balances/cartons
   packed = 100

9. Shipment
   carrier service and vehicle are assigned
   dispatch removes 100 from warehouse inventory
   shipped = 100

10a. Successful delivery
    POD captured; delivered = 100

10b. Failed delivery
    failure reason captured; retry or return all lines to depot
```

## 12. Table responsibility summary

| Table | Responsibility |
|---|---|
| `outbound_order` | Client DO header and destination snapshot |
| `outbound_order_line` | Requested item quantities and fulfilment totals |
| `outbound_validation_run` | One validation attempt |
| `outbound_validation_result_detail` | Result for each configured rule |
| `inventory_reservation` | Stock promised from one exact balance |
| `outbound_wave` | Group of picking work |
| `outbound_wave_order` | DO membership in a wave |
| `pick_task` | Picking instruction/work line |
| `pick_execution` | Actual partial or complete physical pick |
| `outbound_staging` | Consolidation header for a DO in staging |
| `outbound_staging_line` | Pick execution included in staging |
| `outbound_check` | One check or recheck attempt |
| `outbound_check_line` | Checked result and affected exception quantity |
| `outbound_check_exception` | One discrepancy requiring correction |
| `outbound_check_resolution` | Correction, replacement, or approved shortage action |
| `packing` | Packing header for one DO |
| `packing_line` | Checked stock packed into a location/HU |
| `shipment` | Dispatch header and transport data |
| `shipment_order` | DOs assigned to a shipment manifest |
| `shipment_driver` | Driver/crew assignments for the shipment |
| `shipment_packing` | Packing documents attached to shipment |
| `shipment_line` | Packed quantities actually dispatched |
| `delivery` | Store-delivery status and POD |
| `delivery_line` | Planned, delivered, and returned quantities per shipped line |
| `delivery_event` | Delivery timeline, including failures |
| `delivery_event_line` | Quantity accepted by one POD event |
| `outbound_return_policy` | Safe non-allocatable return location/status per owner/warehouse |
| `delivery_return_line` | Partial or full undelivered quantity returned to stock |

## 13. Resolved outbound policies

1. Failed checks use immutable exceptions and linked resolution evidence.
2. Short stock may be replaced or accepted short only with recorded approval.
3. Drivers are carrier masters assigned through `shipment_driver`.
4. Carrier may link to a corresponding `business_partner`.
5. A shipment is a multi-DO manifest; delivery/POD remains per DO/store stop.
6. Delivery acceptance and returns may be partial and repeated safely.
7. Returned goods use `outbound_return_policy` and must enter a
   non-allocatable status such as `HOLD`.
8. Validation uses a reviewed handler registry identified by `handler_code`.

These decisions close the previously identified outbound schema gaps. The
backend should call the posting statements as transactions and treat an empty
`RETURNING` result as a rejected state/concurrency transition.

## 14. Related implementation files

- Schema: `database/wms_schema.sql`
- Reference data: `database/master/00_bootstrap_reference_data.sql`
- DO, validation, allocation, and wave queries:
  `database/outbound/01_do_allocation_wave.sql`
- Pick, stage, check, and pack queries:
  `database/outbound/02_pick_stage_check_pack.sql`
- Shipment, delivery, POD, and return queries:
  `database/outbound/03_shipment_delivery.sql`
