# Study Guide 05: Stock Control

Stock control is easier to understand when its tables are separated into three
layers:

| Layer | Purpose | Main tables |
|---|---|---|
| Current state | What stock exists now? | `inventory_balance` |
| Operational documents | What work/change has been requested and approved? | Internal move, transfer, status change, adjustment, and count tables |
| Audit ledger | What physical or status change actually happened? | `inventory_movement` |

The fundamental rule is:

```text
Operational document requests the work.
Posting/execution changes inventory_balance.
inventory_movement records the permanent audit event.
```

## 1. `inventory_balance`: current stock state

One balance represents one exact stock identity:

```text
owner
+ warehouse
+ location
+ item
+ lot
+ handling unit
+ inventory status
```

### 1.1 `on_hand_qty`

`on_hand_qty` is the complete current quantity in that one balance bucket. It
includes quantity that has been reserved.

It is not the total item quantity across the whole warehouse. To get the total,
sum all relevant balances.

Example balance:

```text
Owner:       CLIENT_FRESH
Warehouse:   WH_JKT_01
Location:    DRY-A-01-01
Item:        APPLE_JUICE_1L
Lot:         LOT-A
Status:      AVAILABLE
On hand:     1,000 EA
Reserved:      200 EA
Available:     800 EA
```

The formula is:

```text
available_qty = on_hand_qty - reserved_qty
```

### 1.2 `reserved_qty`

Yes: in the current schema, reserved quantity is primarily stock booked by
outbound allocation for a delivery/outbound order.

The detail is stored in `inventory_reservation`:

```text
outbound_order_line
  -> inventory_reservation
  -> inventory_balance
```

Several reservations can contribute to one balance's summarized
`reserved_qty`. A reservation promises stock but does not physically move it.

Internal moves, transfers, status changes, and negative adjustments normally
use only:

```text
on_hand_qty - reserved_qty
```

so they cannot accidentally consume stock already promised to outbound.

## 2. Internal movement

Internal movement means moving stock inside the same warehouse:

- Bin to bin
- Aisle to aisle
- Zone to zone
- Reserve storage to pick face
- Consolidating several locations into one
- Relocating stock because of maintenance or space optimization

It is not an inter-warehouse transfer.

### 2.1 `internal_move_order`

This is the request/header. `internal_move_type_id` explains the operational
purpose.

Starter types are:

| Code | Meaning |
|---|---|
| `AD_HOC` | Operator-requested location move |
| `REPLENISHMENT` | Reserve storage to forward pick location |
| `CONSOLIDATION` | Combine compatible stock into fewer locations/HUs |
| `RELOCATION` | Planned relocation; starter configuration requires approval |

The type is not the detailed item movement. It classifies why the overall move
is being performed.

### 2.2 `internal_move_order_line`

The line identifies the exact work:

```text
source_balance_id
target_location_id
planned_qty
completed_qty
assigned_to
```

The source balance already identifies owner, warehouse, source location, item,
lot, HU, status, and UOM. Therefore these fields do not need to be copied onto
the order line.

Example:

```text
Move type:       REPLENISHMENT
Source balance:  APPLE_JUICE, LOT-A, DRY-A-01-01
Target location: PICK-A-01-01
Planned qty:     100 EA
```

### 2.3 `internal_move_execution`

This records the actual execution of all or part of one line.

Posting an execution atomically:

1. Decreases the source balance.
2. Creates/increases the destination balance.
3. Creates an `inventory_movement` with type `INTERNAL_MOVE`.
4. Records source and destination balance IDs in
   `internal_move_execution`.
5. Increases the line's `completed_qty`.
6. Completes the header when every line is complete.

The inventory status is preserved. Only the location changes.

One line can be executed partially several times. If a balance belongs to a
handling unit, the current queries require moving the complete unreserved HU
balance so the database does not leave one HU in two locations.

## 3. Inter-warehouse transfer

An inter-warehouse transfer moves one owner's stock from one warehouse to
another.

It is not putaway:

```text
Putaway/internal move = location change inside one warehouse
Transfer              = source warehouse to target warehouse
```

Both warehouses must be configured for the owner in `warehouse_owner`.

### 3.1 `transfer_order`

This is the approved request to move stock between warehouses.

Example:

```text
Owner:            CLIENT_FRESH
Source warehouse: WH_JKT_01
Target warehouse: WH_BDG_01
Requested date:   2026-08-20
```

Starter status flow:

```text
DRAFT
  -> APPROVED
  -> PARTIALLY_DISPATCHED / IN_TRANSIT
  -> PARTIALLY_RECEIVED
  -> RECEIVED
```

### 3.2 `transfer_order_line`

The line requests:

```text
item
requested quantity
optional requested lot
```

It also accumulates dispatched and received quantities.

It does not choose the exact source balance yet. One requested item line may be
fulfilled from several locations, lots, or HUs when dispatch is prepared.

### 3.3 `transfer_dispatch` and `transfer_dispatch_line`

`transfer_dispatch` represents one physical shipment leaving the source
warehouse. It can record carrier, tracking number, vehicle, and seal.

Each dispatch line selects an exact source balance and quantity. Confirming it:

1. Decreases the physical source balance.
2. Creates a `TRANSFER_OUT` movement.
3. Increases dispatched progress on the order line.
4. Marks the quantity as in transit.

One transfer order can have several dispatches. This supports partial truck
loads or separate shipping dates.

In-transit quantity is calculated as:

```text
dispatch_line.dispatched_qty - dispatch_line.received_qty
```

It is not stored in a fake warehouse location or normal inventory balance.

### 3.4 `transfer_receipt` and `transfer_receipt_line`

The target warehouse creates a transfer receipt against confirmed dispatch
lines. This is separate from the inbound vendor `receipt` table.

Each line records:

```text
which dispatch line arrived
target warehouse location
received quantity
```

Posting it:

1. Creates/increases the target inventory balance.
2. Preserves item, lot, HU, and inventory status from the source balance.
3. Creates a `TRANSFER_IN` movement.
4. Decreases the calculated in-transit remainder.
5. Updates transfer received progress.

Example:

```text
Requested:  300 EA
Dispatched: 200 EA
Received:   190 EA
In transit:  10 EA
Not yet dispatched from order: 100 EA
```

For a handling unit, the current flow requires the remaining dispatched HU
quantity to be received together, then updates the HU's warehouse and current
location.

### 3.5 Does transfer receipt perform QC?

No. The current transfer-receipt flow preserves the source inventory status.
For example, `AVAILABLE` stock at the source becomes `AVAILABLE` at the target.

It does not create vendor-style `receipt`, `receipt_line`,
`receipt_inventory`, or `quality_inspection` rows.

If the receiving warehouse must inspect transferred stock, a transfer-QC flow
must be designed. Options include receiving into a non-allocatable
`TRANSFER_QC_PENDING` status and adding an inspection source applicable to
transfer receipts.

## 4. Inventory status change

`inventory_status_change` is a separate controlled document for changing stock
condition without changing physical quantity or location.

Examples:

```text
AVAILABLE -> HOLD
HOLD      -> AVAILABLE
AVAILABLE -> DAMAGED
AVAILABLE -> EXPIRED
```

The header records owner, warehouse, reason, approval, and posting status. Each
`inventory_status_change_line` selects a source balance, target status, and
quantity.

Posting a line:

1. Decreases the source-status balance.
2. Creates/increases the same identity under the target status.
3. Creates a `STATUS_CHANGE` movement.

It does not happen automatically after every other stock-control operation.

## 5. Inventory adjustment

An adjustment deliberately changes book quantity when no normal operational
transaction explains the difference.

Examples:

- Unrecorded loss found during investigation
- Authorized correction of a migration/input error
- Found stock that must be added outside a stock count
- Confirmed shrinkage

Starter adjustment types are:

| Code | `quantity_effect` | Meaning |
|---|---:|---|
| `INCREASE` | +1 | Add quantity |
| `DECREASE` | -1 | Remove quantity |

`inventory_adjustment` is the approval/reason header.
`inventory_adjustment_line` identifies the balance identity and positive
adjustment magnitude.

The type's `quantity_effect` decides whether the quantity is added or removed.
For a decrease, the line references an existing source balance. For an
increase, it can create a new balance identity.

Posting creates an `ADJUSTMENT` movement. It does not create an
`inventory_status_change` document.

Use a status change, not an adjustment, when the quantity still exists but is
damaged or held:

```text
Still physically present but damaged -> AVAILABLE to DAMAGED status change
Physically missing/lost              -> DECREASE adjustment
```

## 6. Stock count

A stock count is a physical verification process. It is not another current
balance table.

```text
inventory_balance = what the system currently believes
stock_count        = the controlled event used to compare belief with reality
```

Starter count types are:

| Code | Purpose |
|---|---|
| `CYCLE` | Recurring count of selected stock |
| `FULL` | Entire in-scope warehouse; starter config requests freeze |
| `LOCATION` | Selected warehouse locations |
| `ITEM` | Selected items across locations |

### 6.1 Count process

```text
Create count
  -> snapshot selected inventory balances
  -> start counting
  -> enter physical quantities
  -> review variances
  -> recount if necessary
  -> post corrections
```

Each `stock_count_line` stores:

```text
system_qty         = balance quantity at snapshot time
system_version_no  = balance version at snapshot time
counted_qty        = physical quantity entered by counter
variance           = counted_qty - system_qty
```

Example:

```text
System quantity: 500 EA
Physical count:  497 EA
Variance:         -3 EA
```

After review, posting changes the balance to 497 and creates a
`COUNT_CORRECTION` movement for 3 EA.

If physical stock exists but no balance was in the snapshot, a found-stock line
starts with `system_qty = 0` and can create a new balance after review.

The snapshot version prevents posting when another transaction changed the
balance after counting began. The count must then be refreshed or repeated.

### 6.2 Inventory freeze limitation

`stock_count.freeze_inventory` is currently only recorded as a flag. Other
movement, allocation, and posting queries do not check it. Therefore the
database does not yet truly freeze stock during a full count.

Before backend implementation, either:

- Enforce active count locks in every balance-changing operation; or
- Use a dedicated inventory/location lock table checked by those operations;
  or
- Rely on optimistic version conflicts and require recounts, while renaming the
  flag so it does not imply a real freeze.

## 7. `inventory_movement`: universal audit ledger

`inventory_movement` is not only for outbound. It is the immutable audit
history for every inventory-changing process.

Examples of `movement_type`:

| Process | Movement type |
|---|---|
| Vendor receipt | `RECEIVE` |
| Putaway | `PUTAWAY` |
| Internal location move | `INTERNAL_MOVE` |
| Transfer dispatch | `TRANSFER_OUT` |
| Transfer receipt | `TRANSFER_IN` |
| Status hold/release | `STATUS_CHANGE` |
| Manual quantity adjustment | `ADJUSTMENT` |
| Stock-count variance | `COUNT_CORRECTION` |
| Outbound picking | `PICK` |
| Shipment | `SHIP` |
| Return to vendor | `RETURN_TO_VENDOR` |
| Disposal | `DISPOSE` |
| Delivery return | `DELIVERY_RETURN` |

There is no `inventory_movement_line` table in the current model. One
`inventory_movement` row is already one atomic item/lot/HU/quantity movement.

`source_document_id` and `source_line_id` identify the operational document and
line that caused it. For example:

```text
source document = internal_move_order
source line     = internal_move_order_line
movement type   = INTERNAL_MOVE
```

For an inter-warehouse transfer, there are two movements:

```text
Source warehouse: TRANSFER_OUT
Target warehouse: TRANSFER_IN
```

The time between them is represented by the dispatch/receipt quantities.

## 8. What each successful process creates

| Successful process | Changes balance? | Creates movement? | Creates status-change document? |
|---|:---:|:---:|:---:|
| Internal move execution | Yes, location | `INTERNAL_MOVE` | No |
| Transfer dispatch | Yes, removes source | `TRANSFER_OUT` | No |
| Transfer receipt | Yes, adds target | `TRANSFER_IN` | No |
| Status-change posting | Yes, status | `STATUS_CHANGE` | It is the status-change document |
| Adjustment posting | Yes, quantity | `ADJUSTMENT` | No |
| Count-variance posting | Yes, quantity | `COUNT_CORRECTION` | No |
| Outbound pick | Yes, location/quantity | `PICK` | No |

All successful stock operations create or update `inventory_balance` and append
an `inventory_movement`; they do not all create `inventory_status_change`.

## 9. Complete examples

### 9.1 Internal replenishment

```text
Before:
  Reserve location DRY-A-01: 1,000 EA
  Pick face PICK-A-01:          50 EA

Request:
  REPLENISHMENT of 200 EA from reserve to pick face

After execution:
  Reserve location: 800 EA
  Pick face:        250 EA
  Movement: INTERNAL_MOVE 200 EA
```

### 9.2 Inter-warehouse transfer

```text
Transfer order: WH_JKT_01 -> WH_BDG_01, 300 EA
Dispatch 1: 200 EA -> TRANSFER_OUT
Receipt 1:  190 EA -> TRANSFER_IN

Current progress:
  100 EA not yet dispatched
   10 EA in transit
  190 EA received at target
```

### 9.3 Stock status hold

```text
Before: 100 EA AVAILABLE
Action:  Hold 20 EA because packaging is under investigation

After:
  80 EA AVAILABLE
  20 EA HOLD
  Movement: STATUS_CHANGE 20 EA
```

### 9.4 Stock count correction

```text
System:   500 EA
Counted:  497 EA
Variance:  -3 EA

After reviewed posting:
  Balance: 497 EA
  Movement: COUNT_CORRECTION 3 EA decrease
```

## 10. Corrected mental model

```text
inventory_balance
  = current quantity now

inventory_reservation
  = part of current balance promised to outbound

internal_move_order + line
  = request to move stock inside one warehouse

internal_move_execution
  = actual completed quantity and linked ledger event

transfer_order
  = request to move stock between warehouses

transfer_dispatch
  = physical stock leaving source warehouse

transfer_receipt
  = physical stock received at target warehouse

inventory_status_change
  = controlled AVAILABLE/HOLD/DAMAGED/etc. reclassification

inventory_adjustment
  = controlled book-quantity increase or decrease

stock_count
  = physical verification and variance correction

inventory_movement
  = universal immutable audit ledger for all modules
```

## 11. Decisions to confirm before backend implementation

1. Should transferred stock retain its source status, or enter a target
   `TRANSFER_QC_PENDING` status?
2. Is transfer QC required at some or all target warehouses?
3. How will full-count inventory freezes be enforced by every balance-changing
   operation?
4. May reserved inventory be internally replenished while preserving its
   reservation, or should only unreserved quantity move as currently designed?
5. Who may approve relocation, transfers, status changes, adjustments, and
   count variances?
6. Should zero-quantity balance rows be retained for traceability or cleaned up
   periodically?
7. Should `source_document_id`/`source_line_id` remain polymorphic text links,
   or should a stronger source-reference mechanism be added?

## 12. Source files

- [Stock-control tables](../../wms_schema.sql)
- [Balance inquiry, internal movement, and status change](../../stock_control/01_internal_movement_inventory.sql)
- [Inter-warehouse transfer](../../stock_control/02_interwarehouse_transfer.sql)
- [Adjustments and stock counts](../../stock_control/03_adjustment_stock_count.sql)
- [Stock-control process notes](../../stock_control/README.md)
- [Stock-control ER diagram](../../erd/domains/stock_control.mmd)

## 13. Suggested study checklist

- [ ] Calculate on-hand, reserved, and available quantities.
- [ ] Trace one internal move from order line through execution and movement.
- [ ] Trace one partial warehouse transfer through dispatch, transit, and receipt.
- [ ] Explain why transfer receipt is different from vendor receipt.
- [ ] Change part of one balance from `AVAILABLE` to `HOLD`.
- [ ] Explain adjustment versus status change.
- [ ] Compare system quantity with counted quantity and post a variance.
- [ ] List the movement type produced by each stock-control operation.
- [ ] Resolve the seven stock-control policy decisions above.

