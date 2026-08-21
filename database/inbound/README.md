# Focused inbound flow: PO to stored inventory

## Process

```text
Purchase order
  -> inbound delivery / ASN
  -> physical receipt
  -> received batch in QC_PENDING inventory
  -> quality inspection
       -> pass: AVAILABLE -> putaway -> storage
       -> fail: QUARANTINE -> client disposition
            -> ACCEPT: AVAILABLE -> putaway -> storage
            -> REWORK: rework -> reinspection -> pass/fail
            -> RETURN: remove from warehouse inventory
            -> DISPOSE: remove from warehouse inventory
```

## Important model rule

`receipt_line` records the total physical receipt for an item. A
`receipt_inventory` row is a homogeneous batch with one lot, handling unit,
location, UOM, and initial status. If one delivery line contains different lots,
handling units, or expected QC outcomes, create several `receipt_inventory` rows.

QC should normally complete a homogeneous batch as all passed or all failed. If
sampling discovers mixed outcomes, split the physical stock into separate
receipt-inventory batches before completing the inspections.

## Current-inventory rule

`receipt_inventory` is receipt traceability. `inventory_balance` is the current
stock state. QC, quarantine decisions, and putaway move quantities between
balance dimensions and append immutable `inventory_movement` records.

Every balance-changing query must:

1. lock/decrement the source balance only when enough unreserved quantity exists;
2. upsert the destination balance;
3. append an inventory movement;
4. update the operational document/task in the same transaction; and
5. roll back if any expected statement returns zero rows.

## Query catalogs

- `01_po_receiving_qc.sql`: PO, inbound delivery, receipt, received stock, and QC.
- `02_quarantine_putaway.sql`: client disposition, rework, return/dispose, and putaway.

Execute one numbered operation at a time through prepared statements. A group
between `BEGIN` and `COMMIT` is one indivisible backend operation.

