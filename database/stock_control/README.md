# Stock control SQL catalog

This module continues after inbound stock has been stored. It covers:

1. `01_internal_movement_inventory.sql`
   - stock balance, availability, and immutable-ledger inquiry;
   - ad hoc movement, replenishment, consolidation, and relocation;
   - inventory hold, release, and other configured status changes.
2. `02_interwarehouse_transfer.sql`
   - transfer request and approval;
   - source dispatch and transfer-out posting;
   - in-transit inquiry;
   - destination receipt and transfer-in posting.
3. `03_adjustment_stock_count.sql`
   - controlled increase/decrease adjustments;
   - cycle, full, location, and item counts;
   - count review and variance posting.

All quantities in these files are in the item's base UOM. Transaction primary
keys are varchar business IDs; document and movement numbers use the configured
daily-reset `generate_document_id` function. Master records use UUID keys and
stable editable codes.

Before a posting query that accepts a proposed movement or balance ID, generate
and retain the IDs with the same business date used by the document:

```sql
SELECT generate_document_id('MOVEMENT', NULL, :warehouse_code, :business_date);
SELECT generate_document_id('INVENTORY_BALANCE', NULL, :warehouse_code, :business_date);
```

If a destination balance with the same owner/warehouse/location/item/lot/HU/
status already exists, PostgreSQL returns that existing balance ID from the
upsert; the unused proposed number is an acceptable audit-safe sequence gap.

For every explicit `BEGIN` block, the application must confirm that the final
statement returned the expected row. If it returned no row, execute `ROLLBACK`
instead of `COMMIT`. Common causes are an invalid workflow state, stale data,
insufficient available stock, or a concurrent operation that won the lock.

In-transit stock is not placed in a fake warehouse location. Dispatch reduces
the physical source balance; dispatched-minus-received quantity is traceable in
the transfer execution tables until the destination posts its receipt.
