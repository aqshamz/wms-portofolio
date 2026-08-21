# Study Guide 03: Inventory Identity

Inventory identity answers a deceptively simple question:

> Exactly which stock are we talking about?

An item code alone is not enough. The same item may belong to different
clients, exist in different warehouses and bins, come from different lots, be
packed on different pallets, and have different inventory statuses.

## 1. Start with the five layers

| Layer | Question | Main table |
|---|---|---|
| Item | What product is it? | `item` |
| Lot | From which production batch? | `inventory_lot` |
| Serial | Which individual unit? | `serial_number` |
| Handling unit | Inside which physical container? | `handling_unit` |
| Balance | How much of that exact combination exists here? | `inventory_balance` |

The easiest analogy is:

```text
Item          = bottled medicine product
Lot           = manufacturing batch BATCH-20260801-A
Serial        = one individually numbered bottle/device
Handling unit = carton or pallet containing the goods
Balance       = 120 units of that item/lot/pallet/status in one bin
```

## 2. The inventory-balance identity

One `inventory_balance` row represents one quantity bucket identified by:

```text
owner
+ warehouse
+ location
+ item
+ lot (optional)
+ handling unit (optional)
+ inventory status
```

This combination is unique in the schema.

Therefore these are separate balances even when the item is the same:

| Difference | Why a separate balance is required |
|---|---|
| Different owner | Stock belongs to a different client |
| Different warehouse | Stock is in another facility |
| Different location | Stock is in another bin/dock/lane |
| Different lot | Stock has different production/expiry traceability |
| Different handling unit | Stock is physically packed separately |
| Different status | One quantity may be available while another is held |

`on_hand_qty` is the physical/system quantity in the bucket.
`reserved_qty` is the portion promised to outbound demand.

```text
available-to-reserve quantity = on_hand_qty - reserved_qty
```

The balance UOM should be the item's base UOM. Source documents may use boxes
or cartons, but inventory is normalized when it is received.

## 3. `inventory_lot`

A lot represents a manufacturing or supplier batch shared by many physical
units.

Important columns:

| Column | Meaning |
|---|---|
| `lot_id` | Internal WMS varchar identifier |
| `owner_id` | Client that owns the lot-controlled stock |
| `item_id` | Product to which the lot belongs |
| `lot_number` | Manufacturer/vendor batch number |
| `manufacture_date` | Production date |
| `expiry_date` | Expiration date |
| `quality_status_id` | Current recorded quality classification for the lot |

The business uniqueness is:

```text
(owner_id, item_id, lot_number)
```

This means the same visible lot number may exist for another item or another
owner without being treated as the same stock.

A lot does not contain warehouse or location. One lot may be distributed
across several warehouses, locations, handling units, or statuses. Those
differences belong in `inventory_balance`.

Example:

| Field | Value |
|---|---|
| `lot_id` | `LOT-20260819-000001` |
| Owner | `CLIENT_HEALTH` |
| Item | `MED-COLD-100` |
| `lot_number` | `BATCH-20260801-A` |
| `manufacture_date` | `2026-08-01` |
| `expiry_date` | `2027-08-01` |
| Quality | `PENDING` initially, then `PASSED` |

Use a lot when stock needs batch-level expiry, recall, quality, or supplier
traceability. An item with `item.lot_controlled = true` must receive a lot.

## 4. `serial_number`

A serial identifies one individual physical unit.

Important columns:

| Column | Meaning |
|---|---|
| `serial_id` | Internal WMS varchar identifier |
| `owner_id` | Client owning the serialized unit |
| `item_id` | Product represented by the serial |
| `serial_no` | Manufacturer or client-visible serial number |

The business uniqueness is:

```text
(owner_id, item_id, serial_no)
```

Example data for three devices:

| `serial_id` | Item | `serial_no` |
|---|---|---|
| `SER-20260819-000001` | `SCANNER-X1` | `SN-X1-000881` |
| `SER-20260819-000002` | `SCANNER-X1` | `SN-X1-000882` |
| `SER-20260819-000003` | `SCANNER-X1` | `SN-X1-000883` |

Use serial control when every unit must be independently traced, for example:

- Warehouse scanners
- Mobile phones
- Medical devices
- Appliances with warranty numbers
- High-value machinery

An item can be:

| `lot_controlled` | `serial_controlled` | Meaning |
|:---:|:---:|---|
| No | No | Quantity-only commodity |
| Yes | No | Batch-controlled food/material |
| No | Yes | Individually serialized product |
| Yes | Yes | Each serial also belongs to a production batch |

## 5. `handling_unit`

A handling unit is a physical logistics container used to move or store stock.
It is not an item and does not itself represent a quantity.

Starter handling-unit types are:

```text
PALLET
CARTON
TOTE
BIN
```

Important columns:

| Column | Meaning |
|---|---|
| `handling_unit_id` | Internal WMS varchar ID |
| `barcode` | Scannable license-plate number, globally unique |
| `handling_unit_type_id` | Pallet, carton, tote, or bin |
| `warehouse_id` | Warehouse in which the HU exists |
| `owner_id` | Client owning the stock/container scope |
| `parent_handling_unit_id` | Larger HU containing this HU |
| `current_location_id` | Current physical warehouse location |
| `is_closed` | Container has been finalized/sealed by the application process |

Example hierarchy:

```text
Pallet PLT-JKT-000045
  -> Carton CTN-JKT-000301
  -> Carton CTN-JKT-000302
  -> Carton CTN-JKT-000303
```

The pallet is the parent HU. The cartons are child HUs. Scanning the pallet can
allow the application to move all of its child cartons together.

The actual contents of an HU are represented by inventory balances whose
`handling_unit_id` points to that HU. One HU may contain multiple items or lots
if business rules allow mixed containers.

## 6. `receipt_inventory`

`receipt_inventory` records one homogeneous portion of a received line. It is
the bridge between receiving and the initial inventory identity.

Homogeneous means all quantity in the row has the same:

```text
item
source UOM conversion
lot
handling unit
received location
initial inventory status
```

It stores both the supplier-facing quantity and normalized base quantity.

Example:

```text
source_qty     = 10 BOX
conversion     = 12 EA per BOX
base_qty       = 120 EA
lot            = BATCH-20260801-A
handling unit  = PLT-JKT-000045
location       = RCV-DOCK-01
status         = QC_PENDING
```

If five boxes are on one pallet and five boxes are on another pallet, create
two `receipt_inventory` rows because the handling-unit identity differs.

`initial_balance_id` links the received batch to the balance that was created
or increased by receiving.

For serialized stock, `receipt_line_serial` connects each registered serial to
the received batch. Its unique constraint prevents one serial from being
attached to two different receipt batches.

## 7. Complete lot-controlled receiving example

Assume these masters already exist:

### Owner and warehouse

| Entity | Code/value |
|---|---|
| Inventory owner | `CLIENT_HEALTH` |
| Warehouse | `WH_JKT_01` |
| Receiving location | `RCV-DOCK-01` |
| Storage location | `DRY-A-01-01-01` |

### Item and UOM

| Field | Value |
|---|---|
| Item | `MED-COLD-100` |
| Base UOM | `EA` |
| Receiving UOM | `BOX` |
| Conversion | `1 BOX = 12 EA` |
| `lot_controlled` | true |
| `serial_controlled` | false |

### Physical receipt

The warehouse receives:

```text
10 BOX = 120 EA
lot number BATCH-20260801-A
on pallet PLT-JKT-000045
```

The relevant rows look conceptually like this.

#### `inventory_lot`

| `lot_id` | Owner | Item | Lot number | Expiry |
|---|---|---|---|---|
| `LOT-20260819-000001` | `CLIENT_HEALTH` | `MED-COLD-100` | `BATCH-20260801-A` | `2027-08-01` |

#### `handling_unit`

| `handling_unit_id` | Type | Barcode | Warehouse | Location |
|---|---|---|---|---|
| `HU-WHJKT-20260819-000001` | `PALLET` | `PLT-JKT-000045` | `WH_JKT_01` | `RCV-DOCK-01` |

#### `receipt_inventory`

| ID | Source | Base | Lot | HU | Status |
|---|---:|---:|---|---|---|
| `RCVI-...-0001` | `10 BOX` | `120 EA` | `LOT-...000001` | `HU-...000001` | `QC_PENDING` |

#### Initial `inventory_balance`

| Owner | Warehouse | Location | Item | Lot | HU | Status | On hand | Reserved |
|---|---|---|---|---|---|---|---:|---:|
| `CLIENT_HEALTH` | `WH_JKT_01` | `RCV-DOCK-01` | `MED-COLD-100` | `LOT-...000001` | `HU-...000001` | `QC_PENDING` | 120 EA | 0 EA |

After QC passes and putaway completes, the identity changes in two dimensions:

```text
status:   QC_PENDING -> AVAILABLE
location: RCV-DOCK-01 -> DRY-A-01-01-01
```

The resulting balance is:

| Owner | Warehouse | Location | Item | Lot | HU | Status | On hand |
|---|---|---|---|---|---|---|---:|
| `CLIENT_HEALTH` | `WH_JKT_01` | `DRY-A-01-01-01` | `MED-COLD-100` | `LOT-...000001` | `HU-...000001` | `AVAILABLE` | 120 EA |

`inventory_movement` records both changes so the history remains auditable.

## 8. Why the same item creates multiple balances

Suppose the warehouse has 120 EA of `MED-COLD-100`, but the physical reality is:

```text
60 EA, lot A, pallet 1, AVAILABLE
24 EA, lot A, pallet 2, AVAILABLE
12 EA, lot A, pallet 2, HOLD
24 EA, lot B, pallet 3, QC_PENDING
```

This requires four balances:

| Balance | Lot | HU | Status | Quantity |
|---|---|---|---|---:|
| B1 | A | Pallet 1 | `AVAILABLE` | 60 EA |
| B2 | A | Pallet 2 | `AVAILABLE` | 24 EA |
| B3 | A | Pallet 2 | `HOLD` | 12 EA |
| B4 | B | Pallet 3 | `QC_PENDING` | 24 EA |

The item-level total is still 120 EA, but allocation may use only B1 and B2
because only `AVAILABLE` is allocatable and pickable.

## 9. Inventory movement versus balance

These tables serve different purposes:

| Table | Purpose |
|---|---|
| `inventory_balance` | Current state now |
| `inventory_movement` | Immutable history of what changed |

Example putaway movement:

```text
movement type: PUTAWAY
item:          MED-COLD-100
lot:           BATCH-20260801-A
HU:            PLT-JKT-000045
from location: RCV-DOCK-01
to location:   DRY-A-01-01-01
quantity:      120 EA
```

A movement should accompany every balance-changing operation. The balance
answers "where is it now?" while movement answers "how did it get there?"

## 10. Current schema issues to resolve before backend development

The concepts are sound, but the current DDL has several consistency gaps.

### 10.1 Serialized stock has no explicit current-state table

`serial_number` stores identity, and `receipt_line_serial` records its initial
receipt. `inventory_movement` can optionally reference one serial. However,
`inventory_balance` does not include `serial_id`, and no separate table stores
the serial's current warehouse, location, status, lot, or HU.

Consequently, full individual serial tracking after receiving is not yet
guaranteed. Before supporting serial-controlled items, add either:

```text
serial_inventory
```

with one current-state row per serial, or an equivalent balance-to-serial
mapping maintained by every movement transaction.

### 10.2 Owner and item consistency is not fully enforced by DDL

`inventory_lot` and `serial_number` reference `owner_id` and `item_id`
independently. Raw SQL could theoretically combine owner A with an item owned
by owner B. Operational queries validate this, but stronger DDL should use the
existing composite `item(owner_id, item_id)` key.

The same consideration applies to `inventory_balance`.

### 10.3 Handling-unit warehouse/location consistency

`handling_unit.current_location_id` references a location but the DDL does not
itself ensure that the location belongs to the same warehouse as the HU.
Similarly, the parent HU foreign key does not itself ensure that parent and
child have the same owner and warehouse. The receiving queries validate these
conditions, but composite foreign keys would provide stronger protection.

### 10.4 HU current location versus balance location

Both `handling_unit.current_location_id` and `inventory_balance.location_id`
describe location. Every HU movement must update them together, but the
database currently has no constraint that guarantees they remain equal.

The backend transaction design must treat this as one atomic operation, or the
model should choose a single authoritative source and derive the other value.

### 10.5 Closing a handling unit is not enforced by the database

`is_closed = true` indicates a finalized/sealed HU, but no database constraint
prevents adding or removing balances afterward. Application services must
enforce the rule, or database procedures/triggers must be introduced if strong
database enforcement is required.

## 11. Recommended implementation decision

Before starting backend inventory code, decide the required tracking level:

| Requirement | Required model |
|---|---|
| Quantity-only stock | Current balance model is sufficient |
| Lot/expiry traceability | Current lot plus balance model is suitable after stronger owner/item FKs |
| Pallet/carton tracking | Current HU model is suitable after location/parent consistency rules |
| Individual serial tracking | Add explicit current serial state before backend work |

For this WMS, the recommended approach is:

1. Keep `inventory_balance` as aggregated base-UOM quantity by the existing
   identity dimensions.
2. Add a current-state record for every serial-controlled unit.
3. Strengthen composite foreign keys for owner/item and HU warehouse/location.
4. Define whether HU location or balance location is authoritative.
5. Require movements and balance/current-state updates in one database
   transaction.

## 12. Source files

- [Inventory identity DDL](../../wms_schema.sql)
- [Receiving, lot, HU, and serial queries](../../inbound/01_po_receiving_qc.sql)
- [Inventory identity ER diagram](../../erd/domains/inventory_identity.mmd)
- [Inbound ER diagram](../../erd/domains/inbound.mmd)
- [Stock-control ER diagram](../../erd/domains/stock_control.mmd)

## 13. Suggested study checklist

- [ ] Explain item versus lot versus serial versus handling unit.
- [ ] Write the seven dimensions that identify one inventory balance.
- [ ] Explain why one item can have several balance rows.
- [ ] Convert one receiving UOM quantity into base quantity.
- [ ] Split one received line across two lots or handling units.
- [ ] Trace stock from receiving through QC and putaway.
- [ ] Explain current balance versus immutable movement history.
- [ ] Decide whether individual serial tracking is required.
- [ ] Review and resolve the five DDL consistency gaps above.

