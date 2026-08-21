# Study Guide 04: Inbound — Purchase Order to Putaway or Quarantine

This guide explains the complete inbound chain:

```text
Purchase order
  -> inbound delivery / ASN
  -> physical receipt
  -> received inventory in QC_PENDING
  -> quality inspection
       -> pass: AVAILABLE -> putaway task -> storage
       -> fail: QUARANTINE -> client decision
            -> ACCEPT -> AVAILABLE -> putaway
            -> REWORK -> rework task -> reinspection
            -> RETURN -> inventory removed and returned
            -> DISPOSE -> inventory removed and disposed
```

## 1. The purpose of each level

| Level | What it represents |
|---|---|
| `purchase_order` | What the owner/client ordered or expects from a vendor/factory |
| `inbound_order` | A planned vendor delivery/ASN against the approved PO |
| `receipt` | The vehicle/delivery that physically arrived and was received |
| `receipt_inventory` | Physical quantity split into independently traceable/QC-processable batches |
| `quality_inspection` | QC process and summarized outcome for one received batch |
| `putaway_task` | Work instruction to move accepted stock into storage |
| `quarantine_case` | Failed stock waiting for a client decision |
| `quarantine_disposition` | Client decision for some or all quarantined quantity |
| `rework_task` | Work required when the disposition is `REWORK` |

## 2. Purchase order

### 2.1 Who creates the PO?

`purchase_order.owner_id` is the client/inventory owner on whose behalf the PO
exists. `created_by` is the actual application account that entered it.

Therefore the account can be:

- A client/owner user; or
- A 3PL user authorized to work for that owner.

Authorization should require:

```text
permission:       INBOUND.PO.CREATE
owner scope:      purchase_order.owner_id
warehouse scope:  purchase_order.warehouse_id
```

### 2.2 Vendor and warehouse selection

`vendor_id` references `business_partner`. The create query accepts only an
active partner belonging to the same owner and classified as `SUPPLIER` or
`FACTORY`.

`warehouse_id` is the expected destination warehouse. The owner/warehouse
combination must be active in `warehouse_owner`.

Example header:

| Field | Example |
|---|---|
| Owner | `CLIENT_FRESH` |
| Vendor | `FACTORY_BDG_01` |
| Warehouse | `WH_JKT_01` |
| Business date | `2026-08-19` |
| Internal `purchase_order_id` | `PO-FACTORY_BDG_01-WH_JKT_01-20260819-000001` |
| Client `purchase_order_no` | `CLIENT-PO-2026-00882` |
| Initial status | `DRAFT` |

The generated `purchase_order_id` is the internal WMS identifier.
`purchase_order_no` is the owner's/client's external business number.

### 2.3 Document type and status

The user normally should not choose the PO document type manually. The create
query resolves:

```text
document_type.code = PURCHASE_ORDER
document_status.is_initial = true
```

The configured starter workflow is:

```text
DRAFT
  -> APPROVED
  -> PARTIALLY_RECEIVED
  -> RECEIVED

DRAFT or APPROVED -> CANCELLED when allowed
```

### 2.4 `purchase_order_line`

Each line identifies one item and ordered quantity.

Example:

| Line | Item | Quantity | UOM | Expected lot | Expected expiry |
|---:|---|---:|---|---|---|
| 1 | `APPLE_JUICE_1L` | 1,000 | `EA` | `BATCH-AUG-01` | `2027-02-01` |
| 2 | `ORANGE_JUICE_1L` | 500 | `EA` | null | null |

`expected_lot_no` and `expected_expiry_date` belong on the PO line when the
owner/vendor already knows them. They are expectations, not the final physical
lot record.

If the vendor has not supplied lot information yet, leave these fields null.
The actual lot number and dates are verified/captured during physical
receiving and stored in `inventory_lot`.

The PO line also supports `vendor_item_code` because the vendor's product code
may differ from the owner's WMS item code.

### 2.5 `version_no`

`version_no` implements optimistic concurrency for a mutable header.

Example:

```text
User A opens PO version 1.
User B approves or updates it, producing version 2.
User A attempts to save using expected version 1.
The UPDATE affects zero rows, so the backend reports that the PO changed.
```

A safe update uses a condition similar to:

```sql
WHERE purchase_order_id = :id
  AND version_no = :expected_version
```

and then applies:

```sql
version_no = version_no + 1
```

It is not a document status and not a daily sequence. It prevents one request
from silently overwriting another request's newer header data.

In the current query catalog, header/status updates increment it, but adding a
line does not. PO lines also have no individual version column. Therefore the
current value is a header-record version, not a guaranteed revision number for
the complete header-and-lines document.

## 3. Inbound order / ASN

### 3.1 When is it created?

In the provided flow, `inbound_order` is created from a PO whose status is
`APPROVED` or `PARTIALLY_RECEIVED`.

It represents the specific delivery the vendor/factory plans to send. It is
often created when:

- The vendor confirms shipment;
- An ASN is received;
- A delivery appointment is scheduled; or
- The 3PL prepares for an expected partial delivery.

It is not merely "the PO was accepted." The approved PO authorizes expected
quantity; the inbound order describes one actual planned arrival.

One PO can be fulfilled by several inbound orders.

Example:

```text
PO quantity:       1,000 EA
Inbound ASN 1:       600 EA expected on 2026-08-20
Inbound ASN 2:       400 EA expected on 2026-08-22
```

The query prevents the sum of inbound-line expected quantities from exceeding
the PO-line ordered quantity.

### 3.2 `inbound_order_line`

This is the list of products and expected quantities for one planned delivery.
It references the relevant PO line and copies:

- Item
- UOM
- Expected lot
- Expected expiry

The inbound header starts in `DRAFT`. After its lines are ready and the
delivery is confirmed, it moves to `RELEASED`. Only a released or partially
received inbound order can create a physical receipt.

## 4. Physical receipt

### 4.1 When is `receipt` created?

Yes: it is created when a released inbound delivery physically arrives at the
warehouse and is processed at a receiving-capable dock.

The header records:

- Arrival/receipt timestamp
- Receiving dock
- Vehicle number
- Seal number
- Delivery note number
- Owner and warehouse

One inbound order can have several receipts when the delivery arrives in
parts or multiple vehicles.

### 4.2 `receipt_line`

`receipt_line` records the physical count at the dock. It does not record the
detailed QC inspection.

Example:

```text
Expected on inbound line: 600 EA
Physically presented:      590 EA
Rejected immediately:       10 EA
Admitted into WMS/QC:       580 EA
```

Stored values are:

```text
received_qty = 590
rejected_qty = 10
```

The immediately rejected quantity may represent visibly wrong, damaged, or
refused goods that never enter WMS inventory. A reason/note process for dock
rejection may be added if stronger audit detail is required.

### 4.3 `receipt_inventory`

`receipt_inventory` does not mean "quantity accepted by QC." It means physical
quantity admitted into WMS custody and registered as stock awaiting QC.

The initial inventory status is normally:

```text
QC_PENDING
```

It excludes `receipt_line.rejected_qty`, but it has not yet passed QC.

Each row is homogeneous by:

```text
item
lot
handling unit/pallet
received location
source/base UOM conversion
initial status
```

This is what allows each pallet or lot to proceed independently.

Example admitted quantity split:

| Received batch | Pallet | Lot | Quantity | Initial status |
|---|---|---|---:|---|
| `RCVI-001` | `PLT-001` | `LOT-A` | 200 EA | `QC_PENDING` |
| `RCVI-002` | `PLT-002` | `LOT-A` | 200 EA | `QC_PENDING` |
| `RCVI-003` | `PLT-003` | `LOT-B` | 180 EA | `QC_PENDING` |

## 5. Quality inspection

### 5.1 What does `quality_inspection` contain?

One inspection belongs to one `receipt_inventory` batch. It records:

- QC process status
- Overall inspection result
- Inspected quantity
- Passed quantity
- Failed quantity
- Inspector
- Inspection time
- Notes
- Optional parent inspection for reinspection history

Example:

| Batch | Result | Passed | Failed | Consequence |
|---|---|---:|---:|---|
| `RCVI-001` | `ACCEPTED` | 200 | 0 | Available and putaway |
| `RCVI-002` | Pending | 0 | 0 | Remains `QC_PENDING` |
| `RCVI-003` | `REJECTED` | 0 | 180 | Quarantine |

### 5.2 Is detailed QC checking stored?

The current table stores only the inspection summary and free-text notes. It
does not yet store individual checklist/measurement results such as:

```text
temperature = 4.2 C
packaging condition = PASS
seal integrity = PASS
remaining shelf life = 165 days
sample defects = 2
photo attachment = ...
```

If the WMS requires structured QC details, the model should be extended before
backend implementation with concepts such as:

```text
inspection_template
inspection_criterion
inspection_template_criterion
quality_inspection_detail
inspection_attachment
```

The detail table would store the observed value, pass/fail result, note, and
possibly defect/reason code for every criterion.

### 5.3 Current pass/fail processing rule

The provided SQL operations complete one homogeneous received batch as either:

- All passed; or
- All failed.

Although the table has `passed_qty`, `failed_qty`, and a `PARTIAL` result
master, the query catalog does not yet implement a partial-QC completion
operation.

If one pallet contains mixed results, physically separate/split it into
independent `receipt_inventory` quantities before completing QC. For example:

```text
Original pallet: 200 EA
Passed portion:  180 EA -> separate batch/HU -> AVAILABLE
Failed portion:   20 EA -> separate batch/HU -> QUARANTINE
```

This prevents one physical handling unit from simultaneously moving to storage
and quarantine without a clear stock split.

## 6. QC pass and putaway

When the whole inspected batch passes, one atomic operation:

1. Decreases/removes the `QC_PENDING` source balance.
2. Creates or increases an `AVAILABLE` balance.
3. Appends a `STATUS_CHANGE` inventory movement.
4. Completes the quality inspection as `PASSED/ACCEPTED`.
5. Creates an `OPEN` `putaway_task`.

The putaway task records:

- Source balance/location
- Target storage location
- Owner, warehouse, item, lot, and handling unit
- Planned quantity
- Priority
- Assigned operator
- Started/completed timestamps

Passing QC makes the stock `AVAILABLE`, but its physical location can still be
the receiving/staging area until the putaway task moves it to storage.

Whether stock should be allocatable before physical putaway is a policy choice.
The current operation changes it to `AVAILABLE` before the putaway movement is
completed. If allocation must wait until storage, use a separate status such as
`PUTAWAY_PENDING` or keep it non-allocatable until task completion.

## 7. QC failure, quarantine, and client decision

When QC fails, the stock does not wait for the owner to decide whether it
should enter quarantine. The failure operation immediately:

1. Moves it from the QC-pending balance to a quarantine location.
2. Changes inventory status to `QUARANTINE`.
3. Completes the inspection as `FAILED/REJECTED`.
4. Opens a `quarantine_case` for the client.

There is no separate quarantine task in the current model. The quarantine case
tracks the stock while awaiting the owner's decision.

The owner can record one or several `quarantine_disposition` decisions whose
quantities do not exceed the case quantity:

| Decision | Result |
|---|---|
| `ACCEPT` | Release selected quantity to `AVAILABLE` and create putaway task |
| `REWORK` | Create a `rework_task`, then perform reinspection |
| `RETURN` | Remove selected quantity through return-to-vendor movement |
| `DISPOSE` | Remove selected quantity through disposal movement |

Example for 180 quarantined units:

```text
100 EA -> REWORK
 50 EA -> RETURN
 30 EA -> ACCEPT and putaway
```

The case becomes partially decided until all 180 units have valid decisions,
then it can close after processing.

## 8. Complete example

### Purchase order

```text
PO:       1,000 EA APPLE_JUICE_1L
Owner:    CLIENT_FRESH
Vendor:   FACTORY_BDG_01
Target:   WH_JKT_01
Status:   APPROVED
```

### First planned arrival

```text
Inbound order: 600 EA
Status: RELEASED
```

### Physical receipt

```text
Presented:        590 EA
Rejected at dock:  10 EA
Entered QC:       580 EA
```

### Received batches

```text
PLT-001: 200 EA, LOT-A
PLT-002: 200 EA, LOT-A
PLT-003: 180 EA, LOT-B
```

### Independent processing

```text
PLT-001 -> QC passed -> AVAILABLE -> putaway -> storage
PLT-002 -> QC pending -> remains at receiving
PLT-003 -> QC failed -> QUARANTINE -> client decision
```

The receipt/inbound/PO progress records physical receipt, while QC and putaway
continue independently for each received batch.

## 9. Corrected interpretation

```text
purchase_order
  = owner expectation/order placed with supplier or factory

inbound_order
  = one planned delivery/ASN against the approved PO

receipt
  = one physical arrival/receiving event

receipt_line
  = physical item count and immediate dock rejection

receipt_inventory
  = quantity admitted into stock, split by lot/HU/location and initially QC_PENDING

quality_inspection
  = summarized QC process and result for one received batch

QC pass
  = AVAILABLE balance plus putaway work

QC fail
  = immediate quarantine plus client decision case

REWORK
  = one possible client decision after quarantine, represented by a rework task
```

## 10. Decisions to confirm before backend implementation

1. Is PO/inbound expected lot information optional or mandatory for particular
   item categories?
2. What reason and evidence must be recorded for dock-rejected quantity?
3. Should PO/inbound receipt progress count all `received_qty`, as the current
   query does, or only `received_qty - rejected_qty`?
4. Does QC require structured templates, criteria, measurements, defects, and
   attachments?
5. Will mixed QC outcomes be physically split before completion, or should a
   partial-QC database operation be implemented?
6. May QC-passed stock become allocatable before putaway finishes?
7. Should adding/updating PO and inbound lines increment the header
   `version_no` so it represents the whole document revision?
8. Which transitions require explicit permissions in
   `document_status_transition.required_permission_id`?

## 11. Source files

- [Inbound tables](../../wms_schema.sql)
- [PO, inbound, receipt, received inventory, and QC queries](../../inbound/01_po_receiving_qc.sql)
- [Quarantine, disposition, rework, and putaway queries](../../inbound/02_quarantine_putaway.sql)
- [Focused inbound process notes](../../inbound/README.md)
- [Inbound ER diagram](../../erd/domains/inbound.mmd)

## 12. Suggested study checklist

- [ ] Explain PO expectation versus ASN/planned delivery.
- [ ] Split one PO across two inbound orders.
- [ ] Explain presented, dock-rejected, and QC-pending quantities.
- [ ] Split one receipt line across several lots or pallets.
- [ ] Trace one passed pallet into putaway and storage.
- [ ] Trace one failed pallet into quarantine and client disposition.
- [ ] Explain `version_no` with two concurrent users.
- [ ] Decide whether structured QC details are required.
- [ ] Resolve the eight inbound policy decisions above.
