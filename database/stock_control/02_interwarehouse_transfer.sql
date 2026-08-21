-- Inter-warehouse transfer request, dispatch, transit, and receipt operations.
-- Quantities use the item base UOM. Execute one numbered operation at a time.

SET search_path TO wms, public;

-- =============================================================================
-- 1. TRANSFER REQUEST
-- =============================================================================

-- 1.1 Create a draft transfer and generate its varchar ID.
-- $1 owner ID, $2 source warehouse code, $3 target warehouse code,
-- $4 business date, $5 requested transfer timestamp or NULL,
-- $6 notes, $7 actor account ID
WITH context AS (
    SELECT owner.organization_id AS owner_id,
           source.warehouse_id AS source_warehouse_id,
           source.code AS source_warehouse_code,
           target.warehouse_id AS target_warehouse_id,
           dt.document_type_id, initial_status.status_id
    FROM organization owner
    JOIN warehouse_owner source_scope
      ON source_scope.owner_id = owner.organization_id AND source_scope.is_active
    JOIN warehouse source
      ON source.warehouse_id = source_scope.warehouse_id
     AND source.code = $2 AND source.is_active
    JOIN warehouse_owner target_scope
      ON target_scope.owner_id = owner.organization_id AND target_scope.is_active
    JOIN warehouse target
      ON target.warehouse_id = target_scope.warehouse_id
     AND target.code = $3 AND target.is_active
    JOIN document_type dt ON dt.code = 'TRANSFER' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE owner.organization_id = $1
      AND source.warehouse_id <> target.warehouse_id
), numbered AS (
    SELECT context.*,
           generate_document_id('TRANSFER', NULL, source_warehouse_code, $4) AS generated_id
    FROM context
)
INSERT INTO transfer_order (
    transfer_id, document_type_id, status_id, owner_id,
    source_warehouse_id, target_warehouse_id, business_date,
    requested_transfer_at, notes, created_by, updated_by
)
SELECT generated_id, document_type_id, status_id, owner_id,
       source_warehouse_id, target_warehouse_id, $4, $5, $6, $7, $7
FROM numbered
RETURNING *;

-- 1.2 Add a draft transfer line.
-- $1 transfer ID, $2 line number, $3 item code,
-- $4 requested base quantity, $5 requested lot ID or NULL, $6 actor ID
INSERT INTO transfer_order_line (
    transfer_line_id, transfer_id, line_no, item_id,
    requested_qty, uom_id, requested_lot_id, created_by
)
SELECT transfer.transfer_id || '-L-' || lpad($2::text, 4, '0'),
       transfer.transfer_id, $2, item.item_id, $4,
       item.base_uom_id, lot.lot_id, $6
FROM transfer_order transfer
JOIN document_status status ON status.status_id = transfer.status_id
JOIN item
  ON item.owner_id = transfer.owner_id AND item.code = $3 AND item.is_active
LEFT JOIN inventory_lot lot
  ON lot.lot_id = $5 AND lot.owner_id = transfer.owner_id AND lot.item_id = item.item_id
WHERE transfer.transfer_id = $1
  AND status.code = 'DRAFT'
  AND ($5::varchar IS NULL OR lot.lot_id IS NOT NULL)
  AND $4 > 0
RETURNING *;

-- 1.3 Update a draft transfer line.
-- $1 transfer-line ID, $2 requested quantity, $3 requested lot ID or NULL
UPDATE transfer_order_line line
SET requested_qty = $2,
    requested_lot_id = lot.lot_id
FROM transfer_order transfer, document_status status
LEFT JOIN inventory_lot lot ON lot.lot_id = $3
WHERE line.transfer_line_id = $1
  AND transfer.transfer_id = line.transfer_id
  AND status.status_id = transfer.status_id
  AND (lot.lot_id IS NULL OR (
      lot.owner_id = transfer.owner_id AND lot.item_id = line.item_id
  ))
  AND status.code = 'DRAFT'
  AND ($3::varchar IS NULL OR lot.lot_id IS NOT NULL)
  AND $2 > 0
RETURNING line.*;

-- 1.4 Delete a draft line. $1 transfer-line ID
DELETE FROM transfer_order_line line
USING transfer_order transfer, document_status status
WHERE line.transfer_line_id = $1
  AND transfer.transfer_id = line.transfer_id
  AND status.status_id = transfer.status_id
  AND status.code = 'DRAFT'
RETURNING line.*;

-- 1.5 Approve a transfer with at least one line. $1 transfer ID, $2 actor ID
UPDATE transfer_order transfer
SET status_id = approved.status_id,
    approved_at = clock_timestamp(), approved_by = $2,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = transfer.version_no + 1
FROM document_status current_status
JOIN document_status approved
  ON approved.document_type_id = current_status.document_type_id
 AND approved.code = 'APPROVED' AND approved.is_active
WHERE transfer.transfer_id = $1
  AND current_status.status_id = transfer.status_id
  AND current_status.code = 'DRAFT'
  AND EXISTS (
      SELECT 1 FROM transfer_order_line line
      WHERE line.transfer_id = transfer.transfer_id
  )
RETURNING transfer.*;

-- 1.6 Transfer list.
-- $1 owner ID, $2 source warehouse ID or NULL, $3 target warehouse ID or NULL,
-- $4 status code or NULL, $5 date from, $6 date until, $7 limit, $8 offset
SELECT
    transfer.transfer_id,
    transfer.business_date,
    source.code AS source_warehouse_code,
    target.code AS target_warehouse_code,
    status.code AS status_code,
    transfer.requested_transfer_at,
    transfer.completed_at,
    sum(line.requested_qty) AS requested_qty,
    sum(line.dispatched_qty) AS dispatched_qty,
    sum(line.received_qty) AS received_qty,
    sum(line.dispatched_qty - line.received_qty) AS in_transit_qty,
    count(*) OVER () AS total_rows
FROM transfer_order transfer
JOIN warehouse source ON source.warehouse_id = transfer.source_warehouse_id
JOIN warehouse target ON target.warehouse_id = transfer.target_warehouse_id
JOIN document_status status ON status.status_id = transfer.status_id
JOIN transfer_order_line line ON line.transfer_id = transfer.transfer_id
WHERE transfer.owner_id = $1
  AND ($2::uuid IS NULL OR transfer.source_warehouse_id = $2)
  AND ($3::uuid IS NULL OR transfer.target_warehouse_id = $3)
  AND ($4::varchar IS NULL OR status.code = $4)
  AND transfer.business_date BETWEEN $5 AND $6
GROUP BY transfer.transfer_id, transfer.business_date, source.code, target.code,
         status.code, transfer.requested_transfer_at, transfer.completed_at
ORDER BY transfer.business_date DESC, transfer.transfer_id DESC
LIMIT $7 OFFSET $8;

-- =============================================================================
-- 2. SOURCE-WAREHOUSE DISPATCH
-- =============================================================================

-- 2.1 Create a draft dispatch.
-- $1 transfer ID, $2 business date, $3 carrier partner ID or NULL,
-- $4 tracking number, $5 vehicle number, $6 seal number,
-- $7 notes, $8 actor account ID
WITH context AS (
    SELECT transfer.transfer_id, source.code AS warehouse_code,
           dt.document_type_id, initial_status.status_id
    FROM transfer_order transfer
    JOIN document_status transfer_status ON transfer_status.status_id = transfer.status_id
    JOIN warehouse source ON source.warehouse_id = transfer.source_warehouse_id
    JOIN document_type dt ON dt.code = 'TRANSFER_DISPATCH' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    LEFT JOIN business_partner carrier
      ON carrier.partner_id = $3 AND carrier.owner_id = transfer.owner_id AND carrier.is_active
    WHERE transfer.transfer_id = $1
      AND transfer_status.code IN ('APPROVED', 'PARTIALLY_DISPATCHED', 'PARTIALLY_RECEIVED')
      AND ($3::uuid IS NULL OR carrier.partner_id IS NOT NULL)
), numbered AS (
    SELECT context.*,
           generate_document_id('TRANSFER_DISPATCH', NULL, warehouse_code, $2) AS generated_id
    FROM context
)
INSERT INTO transfer_dispatch (
    transfer_dispatch_id, transfer_id, document_type_id, status_id,
    business_date, carrier_partner_id, tracking_number,
    vehicle_number, seal_number, notes, created_by
)
SELECT generated_id, transfer_id, document_type_id, status_id, $2,
       $3, $4, $5, $6, $7, $8
FROM numbered
RETURNING *;

-- 2.2 Add a source balance to a draft dispatch.
-- $1 dispatch ID, $2 line number, $3 transfer-line ID,
-- $4 source balance ID, $5 dispatch base quantity, $6 actor ID
INSERT INTO transfer_dispatch_line (
    transfer_dispatch_line_id, transfer_dispatch_id, transfer_line_id,
    line_no, source_balance_id, dispatched_qty, uom_id, created_by
)
SELECT dispatch.transfer_dispatch_id || '-L-' || lpad($2::text, 4, '0'),
       dispatch.transfer_dispatch_id, transfer_line.transfer_line_id, $2,
       source.balance_id, $5, transfer_line.uom_id, $6
FROM transfer_dispatch dispatch
JOIN document_status dispatch_status ON dispatch_status.status_id = dispatch.status_id
JOIN transfer_order transfer ON transfer.transfer_id = dispatch.transfer_id
JOIN transfer_order_line transfer_line
  ON transfer_line.transfer_line_id = $3
 AND transfer_line.transfer_id = transfer.transfer_id
JOIN inventory_balance source
  ON source.balance_id = $4
 AND source.owner_id = transfer.owner_id
 AND source.warehouse_id = transfer.source_warehouse_id
 AND source.item_id = transfer_line.item_id
 AND source.uom_id = transfer_line.uom_id
WHERE dispatch.transfer_dispatch_id = $1
  AND dispatch_status.code = 'DRAFT'
  AND source.inventory_status_id IN (
      SELECT inventory_status_id FROM inventory_status
      WHERE is_active AND is_allocatable
  )
  AND (transfer_line.requested_lot_id IS NULL
       OR source.lot_id = transfer_line.requested_lot_id)
  AND transfer_line.dispatched_qty + $5 <= transfer_line.requested_qty
  AND source.on_hand_qty - source.reserved_qty >= $5
  AND $5 > 0
RETURNING *;

-- 2.3 Confirm one dispatch line atomically.
-- The dispatch becomes DISPATCHED when all its lines are confirmed. The transfer
-- becomes IN_TRANSIT when every requested line is fully dispatched; otherwise it
-- becomes PARTIALLY_DISPATCHED.
-- $1 dispatch-line ID, $2 movement ID, $3 actor account ID
BEGIN;

SELECT source.balance_id
FROM transfer_dispatch_line line
JOIN inventory_balance source ON source.balance_id = line.source_balance_id
WHERE line.transfer_dispatch_line_id = $1
FOR UPDATE OF source;

WITH context AS (
    SELECT dispatch_line.*, dispatch.transfer_id, dispatch.document_type_id,
           dispatch.business_date, transfer.owner_id,
           transfer.source_warehouse_id, transfer.target_warehouse_id,
           transfer.document_type_id AS transfer_document_type_id,
           transfer.status_id AS transfer_status_id,
           transfer_line.item_id, transfer_line.requested_qty,
           source.location_id, source.lot_id, source.handling_unit_id,
           source.inventory_status_id, source.on_hand_qty, source.reserved_qty
    FROM transfer_dispatch_line dispatch_line
    JOIN transfer_dispatch dispatch
      ON dispatch.transfer_dispatch_id = dispatch_line.transfer_dispatch_id
    JOIN document_status dispatch_status ON dispatch_status.status_id = dispatch.status_id
    JOIN transfer_order transfer ON transfer.transfer_id = dispatch.transfer_id
    JOIN document_status transfer_status ON transfer_status.status_id = transfer.status_id
    JOIN transfer_order_line transfer_line
      ON transfer_line.transfer_line_id = dispatch_line.transfer_line_id
    JOIN inventory_balance source ON source.balance_id = dispatch_line.source_balance_id
    WHERE dispatch_line.transfer_dispatch_line_id = $1
      AND dispatch_status.code = 'DRAFT'
      AND transfer_status.code IN ('APPROVED', 'PARTIALLY_DISPATCHED', 'PARTIALLY_RECEIVED')
      AND dispatch_line.confirmed_at IS NULL
      AND source.on_hand_qty - source.reserved_qty >= dispatch_line.dispatched_qty
      AND transfer_line.dispatched_qty + dispatch_line.dispatched_qty
          <= transfer_line.requested_qty
      AND (
          source.handling_unit_id IS NULL
          OR (source.reserved_qty = 0 AND dispatch_line.dispatched_qty = source.on_hand_qty)
      )
), decremented AS (
    UPDATE inventory_balance source
    SET on_hand_qty = source.on_hand_qty - context.dispatched_qty,
        version_no = source.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE source.balance_id = context.source_balance_id
      AND source.on_hand_qty - source.reserved_qty >= context.dispatched_qty
    RETURNING source.balance_id
), movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, handling_unit_id,
        from_location_id, to_location_id, from_status_id, to_status_id,
        quantity, uom_id, source_document_id, source_line_id, created_by
    )
    SELECT $2, movement_type.movement_type_id, context.owner_id,
           context.source_warehouse_id, context.business_date,
           context.item_id, context.lot_id, context.handling_unit_id,
           context.location_id, NULL, context.inventory_status_id, NULL,
           context.dispatched_qty, context.uom_id,
           context.transfer_dispatch_id, context.transfer_dispatch_line_id, $3
    FROM context CROSS JOIN decremented
    JOIN movement_type ON movement_type.code = 'TRANSFER_OUT' AND movement_type.is_active
    RETURNING movement_id
), confirmed_line AS (
    UPDATE transfer_dispatch_line line
    SET movement_id = movement.movement_id,
        confirmed_at = clock_timestamp(), confirmed_by = $3
    FROM movement
    WHERE line.transfer_dispatch_line_id = $1
    RETURNING line.*
), updated_transfer_line AS (
    UPDATE transfer_order_line line
    SET dispatched_qty = line.dispatched_qty + confirmed.dispatched_qty
    FROM confirmed_line confirmed
    WHERE line.transfer_line_id = confirmed.transfer_line_id
    RETURNING line.*
), finalized_dispatch AS (
    UPDATE transfer_dispatch dispatch
    SET status_id = dispatched_status.status_id,
        dispatched_at = clock_timestamp()
    FROM confirmed_line confirmed, document_status dispatched_status
    WHERE dispatch.transfer_dispatch_id = confirmed.transfer_dispatch_id
      AND dispatched_status.document_type_id = dispatch.document_type_id
      AND dispatched_status.code = 'DISPATCHED'
      AND NOT EXISTS (
          SELECT 1 FROM transfer_dispatch_line other
          WHERE other.transfer_dispatch_id = confirmed.transfer_dispatch_id
            AND other.transfer_dispatch_line_id <> confirmed.transfer_dispatch_line_id
            AND other.confirmed_at IS NULL
      )
    RETURNING dispatch.*
), updated_transfer AS (
    UPDATE transfer_order transfer
    SET status_id = next_status.status_id,
        updated_at = clock_timestamp(), updated_by = $3,
        version_no = transfer.version_no + 1
    FROM finalized_dispatch dispatch, updated_transfer_line current_line,
         document_status next_status
    WHERE transfer.transfer_id = dispatch.transfer_id
      AND current_line.transfer_id = transfer.transfer_id
      AND next_status.document_type_id = transfer.document_type_id
      AND next_status.code = CASE
         WHEN EXISTS (
             SELECT 1 FROM transfer_order_line line
             WHERE line.transfer_id = transfer.transfer_id
               AND line.received_qty > 0
         ) THEN 'PARTIALLY_RECEIVED'
         WHEN current_line.dispatched_qty = current_line.requested_qty
          AND NOT EXISTS (
             SELECT 1 FROM transfer_order_line line
             WHERE line.transfer_id = transfer.transfer_id
               AND line.transfer_line_id <> current_line.transfer_line_id
               AND line.dispatched_qty < line.requested_qty
         ) THEN 'IN_TRANSIT' ELSE 'PARTIALLY_DISPATCHED' END
    RETURNING transfer.transfer_id, next_status.code AS transfer_status_code
), updated_hu AS (
    UPDATE handling_unit hu
    SET current_location_id = NULL
    FROM context
    WHERE hu.handling_unit_id = context.handling_unit_id
    RETURNING hu.handling_unit_id
)
SELECT confirmed_line.*, updated_transfer.transfer_status_code
FROM confirmed_line
LEFT JOIN updated_transfer ON true;

COMMIT;

-- 2.4 In-transit stock inquiry.
-- $1 owner ID, $2 source warehouse ID or NULL, $3 target warehouse ID or NULL,
-- $4 item code or NULL, $5 limit, $6 offset
SELECT
    transfer.transfer_id,
    dispatch.transfer_dispatch_id,
    dispatch.dispatched_at,
    source_warehouse.code AS source_warehouse_code,
    target_warehouse.code AS target_warehouse_code,
    item.code AS item_code,
    lot.lot_number,
    dispatch_line.source_balance_id,
    dispatch_line.dispatched_qty,
    dispatch_line.received_qty,
    dispatch_line.dispatched_qty - dispatch_line.received_qty AS in_transit_qty,
    unit.code AS uom_code,
    dispatch.tracking_number,
    count(*) OVER () AS total_rows
FROM transfer_dispatch_line dispatch_line
JOIN transfer_dispatch dispatch
  ON dispatch.transfer_dispatch_id = dispatch_line.transfer_dispatch_id
JOIN transfer_order transfer ON transfer.transfer_id = dispatch.transfer_id
JOIN warehouse source_warehouse ON source_warehouse.warehouse_id = transfer.source_warehouse_id
JOIN warehouse target_warehouse ON target_warehouse.warehouse_id = transfer.target_warehouse_id
JOIN transfer_order_line transfer_line
  ON transfer_line.transfer_line_id = dispatch_line.transfer_line_id
JOIN item ON item.item_id = transfer_line.item_id
JOIN uom unit ON unit.uom_id = dispatch_line.uom_id
JOIN inventory_balance source_balance ON source_balance.balance_id = dispatch_line.source_balance_id
LEFT JOIN inventory_lot lot ON lot.lot_id = source_balance.lot_id
WHERE transfer.owner_id = $1
  AND ($2::uuid IS NULL OR transfer.source_warehouse_id = $2)
  AND ($3::uuid IS NULL OR transfer.target_warehouse_id = $3)
  AND ($4::varchar IS NULL OR item.code = $4)
  AND dispatch_line.confirmed_at IS NOT NULL
  AND dispatch_line.received_qty < dispatch_line.dispatched_qty
ORDER BY dispatch.dispatched_at, transfer.transfer_id, dispatch_line.line_no
LIMIT $5 OFFSET $6;

-- =============================================================================
-- 3. TARGET-WAREHOUSE RECEIPT
-- =============================================================================

-- 3.1 Create a draft transfer receipt.
-- $1 transfer ID, $2 business date, $3 notes, $4 actor account ID
WITH context AS (
    SELECT transfer.transfer_id, transfer.target_warehouse_id,
           target.code AS warehouse_code, dt.document_type_id,
           initial_status.status_id
    FROM transfer_order transfer
    JOIN document_status transfer_status ON transfer_status.status_id = transfer.status_id
    JOIN warehouse target ON target.warehouse_id = transfer.target_warehouse_id
    JOIN document_type dt ON dt.code = 'TRANSFER_RECEIPT' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE transfer.transfer_id = $1
      AND transfer_status.code IN ('PARTIALLY_DISPATCHED', 'IN_TRANSIT', 'PARTIALLY_RECEIVED')
      AND EXISTS (
          SELECT 1 FROM transfer_dispatch_line dispatch_line
          JOIN transfer_dispatch dispatch
            ON dispatch.transfer_dispatch_id = dispatch_line.transfer_dispatch_id
          WHERE dispatch.transfer_id = transfer.transfer_id
            AND dispatch_line.confirmed_at IS NOT NULL
            AND dispatch_line.received_qty < dispatch_line.dispatched_qty
      )
), numbered AS (
    SELECT context.*,
           generate_document_id('TRANSFER_RECEIPT', NULL, warehouse_code, $2) AS generated_id
    FROM context
)
INSERT INTO transfer_receipt (
    transfer_receipt_id, transfer_id, document_type_id, status_id,
    warehouse_id, business_date, notes, created_by
)
SELECT generated_id, transfer_id, document_type_id, status_id,
       target_warehouse_id, $2, $3, $4
FROM numbered
RETURNING *;

-- 3.2 Add a dispatch line to a draft transfer receipt.
-- $1 receipt ID, $2 line number, $3 dispatch-line ID,
-- $4 target location ID, $5 received base quantity, $6 actor ID
INSERT INTO transfer_receipt_line (
    transfer_receipt_line_id, transfer_receipt_id,
    transfer_dispatch_line_id, line_no, target_location_id,
    received_qty, uom_id, created_by
)
SELECT receipt.transfer_receipt_id || '-L-' || lpad($2::text, 4, '0'),
       receipt.transfer_receipt_id, dispatch_line.transfer_dispatch_line_id,
       $2, location.location_id, $5, dispatch_line.uom_id, $6
FROM transfer_receipt receipt
JOIN document_status receipt_status ON receipt_status.status_id = receipt.status_id
JOIN transfer_dispatch_line dispatch_line
  ON dispatch_line.transfer_dispatch_line_id = $3
JOIN transfer_dispatch dispatch
  ON dispatch.transfer_dispatch_id = dispatch_line.transfer_dispatch_id
 AND dispatch.transfer_id = receipt.transfer_id
JOIN inventory_balance source_balance
  ON source_balance.balance_id = dispatch_line.source_balance_id
JOIN warehouse_location location
  ON location.location_id = $4
 AND location.warehouse_id = receipt.warehouse_id
 AND location.is_active
WHERE receipt.transfer_receipt_id = $1
  AND receipt_status.code = 'DRAFT'
  AND dispatch_line.confirmed_at IS NOT NULL
  AND dispatch_line.received_qty + $5 <= dispatch_line.dispatched_qty
  AND (
      source_balance.handling_unit_id IS NULL
      OR $5 = dispatch_line.dispatched_qty - dispatch_line.received_qty
  )
  AND $5 > 0
RETURNING *;

-- 3.3 Post one transfer-receipt line atomically.
-- $1 receipt-line ID, $2 proposed destination balance ID,
-- $3 movement ID, $4 actor account ID
BEGIN;

SELECT dispatch_line.transfer_dispatch_line_id
FROM transfer_receipt_line receipt_line
JOIN transfer_dispatch_line dispatch_line
  ON dispatch_line.transfer_dispatch_line_id = receipt_line.transfer_dispatch_line_id
WHERE receipt_line.transfer_receipt_line_id = $1
FOR UPDATE OF dispatch_line;

WITH context AS (
    SELECT receipt_line.*, receipt.transfer_id, receipt.document_type_id,
           receipt.business_date, receipt.warehouse_id AS target_warehouse_id,
           transfer.owner_id, transfer.document_type_id AS transfer_document_type_id,
           transfer_line.transfer_line_id, transfer_line.item_id,
           dispatch_line.source_balance_id, dispatch_line.dispatched_qty,
           dispatch_line.received_qty AS previously_received_qty,
           source_balance.lot_id, source_balance.handling_unit_id,
           source_balance.inventory_status_id
    FROM transfer_receipt_line receipt_line
    JOIN transfer_receipt receipt ON receipt.transfer_receipt_id = receipt_line.transfer_receipt_id
    JOIN document_status receipt_status ON receipt_status.status_id = receipt.status_id
    JOIN transfer_order transfer ON transfer.transfer_id = receipt.transfer_id
    JOIN transfer_dispatch_line dispatch_line
      ON dispatch_line.transfer_dispatch_line_id = receipt_line.transfer_dispatch_line_id
    JOIN transfer_order_line transfer_line
      ON transfer_line.transfer_line_id = dispatch_line.transfer_line_id
    JOIN inventory_balance source_balance ON source_balance.balance_id = dispatch_line.source_balance_id
    WHERE receipt_line.transfer_receipt_line_id = $1
      AND receipt_status.code = 'DRAFT'
      AND receipt_line.posted_at IS NULL
      AND dispatch_line.confirmed_at IS NOT NULL
      AND dispatch_line.received_qty + receipt_line.received_qty
          <= dispatch_line.dispatched_qty
      AND (
          source_balance.handling_unit_id IS NULL
          OR receipt_line.received_qty = dispatch_line.dispatched_qty - dispatch_line.received_qty
      )
), destination AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT $2, context.owner_id, context.target_warehouse_id,
           context.target_location_id, context.item_id, context.lot_id,
           context.handling_unit_id, context.inventory_status_id,
           context.received_qty, 0, context.uom_id
    FROM context
    ON CONFLICT (
        owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id
    ) DO UPDATE SET
        on_hand_qty = inventory_balance.on_hand_qty + EXCLUDED.on_hand_qty,
        version_no = inventory_balance.version_no + 1,
        updated_at = clock_timestamp()
    RETURNING *
), movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, handling_unit_id,
        from_location_id, to_location_id, from_status_id, to_status_id,
        quantity, uom_id, source_document_id, source_line_id, created_by
    )
    SELECT $3, movement_type.movement_type_id, context.owner_id,
           context.target_warehouse_id, context.business_date,
           context.item_id, context.lot_id, context.handling_unit_id,
           NULL, context.target_location_id, NULL, context.inventory_status_id,
           context.received_qty, context.uom_id,
           context.transfer_receipt_id, context.transfer_receipt_line_id, $4
    FROM context CROSS JOIN destination
    JOIN movement_type ON movement_type.code = 'TRANSFER_IN' AND movement_type.is_active
    RETURNING movement_id
), posted_line AS (
    UPDATE transfer_receipt_line line
    SET destination_balance_id = destination.balance_id,
        movement_id = movement.movement_id,
        posted_at = clock_timestamp(), posted_by = $4
    FROM destination CROSS JOIN movement
    WHERE line.transfer_receipt_line_id = $1
    RETURNING line.*
), updated_dispatch_line AS (
    UPDATE transfer_dispatch_line line
    SET received_qty = line.received_qty + posted.received_qty
    FROM posted_line posted
    WHERE line.transfer_dispatch_line_id = posted.transfer_dispatch_line_id
      AND line.received_qty + posted.received_qty <= line.dispatched_qty
    RETURNING line.*
), updated_transfer_line AS (
    UPDATE transfer_order_line line
    SET received_qty = line.received_qty + posted.received_qty
    FROM posted_line posted
    JOIN transfer_dispatch_line dispatch_line
      ON dispatch_line.transfer_dispatch_line_id = posted.transfer_dispatch_line_id
    WHERE line.transfer_line_id = dispatch_line.transfer_line_id
      AND line.received_qty + posted.received_qty <= line.dispatched_qty
    RETURNING line.*
), finalized_receipt AS (
    UPDATE transfer_receipt receipt
    SET status_id = received_status.status_id,
        received_at = clock_timestamp()
    FROM posted_line current_line, document_status received_status
    WHERE receipt.transfer_receipt_id = current_line.transfer_receipt_id
      AND received_status.document_type_id = receipt.document_type_id
      AND received_status.code = 'RECEIVED'
      AND NOT EXISTS (
          SELECT 1 FROM transfer_receipt_line other
          WHERE other.transfer_receipt_id = current_line.transfer_receipt_id
            AND other.transfer_receipt_line_id <> current_line.transfer_receipt_line_id
            AND other.posted_at IS NULL
      )
    RETURNING receipt.*
), updated_transfer AS (
    UPDATE transfer_order transfer
    SET status_id = next_status.status_id,
        completed_at = CASE WHEN next_status.code = 'RECEIVED'
                            THEN clock_timestamp() ELSE transfer.completed_at END,
        updated_at = clock_timestamp(), updated_by = $4,
        version_no = transfer.version_no + 1
    FROM finalized_receipt receipt, updated_transfer_line current_line,
         document_status next_status
    WHERE transfer.transfer_id = receipt.transfer_id
      AND current_line.transfer_id = transfer.transfer_id
      AND next_status.document_type_id = transfer.document_type_id
      AND next_status.code = CASE
         WHEN current_line.received_qty = current_line.requested_qty
          AND NOT EXISTS (
             SELECT 1 FROM transfer_order_line line
             WHERE line.transfer_id = transfer.transfer_id
               AND line.transfer_line_id <> current_line.transfer_line_id
               AND line.received_qty < line.requested_qty
         ) THEN 'RECEIVED' ELSE 'PARTIALLY_RECEIVED' END
    RETURNING transfer.transfer_id, next_status.code AS transfer_status_code
), updated_hu AS (
    UPDATE handling_unit hu
    SET warehouse_id = context.target_warehouse_id,
        current_location_id = context.target_location_id
    FROM context
    WHERE hu.handling_unit_id = context.handling_unit_id
    RETURNING hu.handling_unit_id
)
SELECT posted_line.*, updated_transfer.transfer_status_code
FROM posted_line
LEFT JOIN updated_transfer ON true;

COMMIT;

-- 3.4 Transfer trace: order lines, dispatches, receipts, and remaining transit.
-- $1 transfer ID
SELECT
    transfer.transfer_id,
    transfer_status.code AS transfer_status_code,
    transfer_line.line_no AS transfer_line_no,
    item.code AS item_code,
    transfer_line.requested_qty,
    transfer_line.dispatched_qty,
    transfer_line.received_qty,
    transfer_line.dispatched_qty - transfer_line.received_qty AS in_transit_qty,
    dispatch.transfer_dispatch_id,
    dispatch_line.transfer_dispatch_line_id,
    dispatch_line.dispatched_qty AS dispatch_qty,
    dispatch_line.received_qty AS dispatch_received_qty,
    receipt.transfer_receipt_id,
    receipt_line.transfer_receipt_line_id,
    receipt_line.received_qty AS receipt_qty,
    receipt_line.posted_at,
    receipt_line.movement_id
FROM transfer_order transfer
JOIN document_status transfer_status ON transfer_status.status_id = transfer.status_id
JOIN transfer_order_line transfer_line ON transfer_line.transfer_id = transfer.transfer_id
JOIN item ON item.item_id = transfer_line.item_id
LEFT JOIN transfer_dispatch_line dispatch_line
  ON dispatch_line.transfer_line_id = transfer_line.transfer_line_id
LEFT JOIN transfer_dispatch dispatch
  ON dispatch.transfer_dispatch_id = dispatch_line.transfer_dispatch_id
LEFT JOIN transfer_receipt_line receipt_line
  ON receipt_line.transfer_dispatch_line_id = dispatch_line.transfer_dispatch_line_id
LEFT JOIN transfer_receipt receipt
  ON receipt.transfer_receipt_id = receipt_line.transfer_receipt_id
WHERE transfer.transfer_id = $1
ORDER BY transfer_line.line_no, dispatch.business_date,
         dispatch_line.line_no, receipt.business_date, receipt_line.line_no;

-- 3.5 Cancel a transfer only before any physical dispatch is confirmed.
-- $1 transfer ID, $2 actor ID
UPDATE transfer_order transfer
SET status_id = cancelled.status_id,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = transfer.version_no + 1
FROM document_status cancelled
WHERE transfer.transfer_id = $1
  AND cancelled.document_type_id = transfer.document_type_id
  AND cancelled.code = 'CANCELLED' AND cancelled.is_active
  AND NOT EXISTS (
      SELECT 1
      FROM transfer_dispatch dispatch
      JOIN transfer_dispatch_line line
        ON line.transfer_dispatch_id = dispatch.transfer_dispatch_id
      WHERE dispatch.transfer_id = transfer.transfer_id
        AND line.confirmed_at IS NOT NULL
  )
  AND EXISTS (
      SELECT 1 FROM document_status current_status
      WHERE current_status.status_id = transfer.status_id
        AND current_status.code IN ('DRAFT', 'APPROVED', 'PARTIALLY_DISPATCHED')
  )
RETURNING transfer.*;

-- 3.6 Cancel an unposted draft dispatch. $1 dispatch ID
UPDATE transfer_dispatch dispatch
SET status_id = cancelled.status_id
FROM document_status cancelled
WHERE dispatch.transfer_dispatch_id = $1
  AND cancelled.document_type_id = dispatch.document_type_id
  AND cancelled.code = 'CANCELLED' AND cancelled.is_active
  AND NOT EXISTS (
      SELECT 1 FROM transfer_dispatch_line line
      WHERE line.transfer_dispatch_id = dispatch.transfer_dispatch_id
        AND line.confirmed_at IS NOT NULL
  )
  AND EXISTS (
      SELECT 1 FROM document_status current_status
      WHERE current_status.status_id = dispatch.status_id
        AND current_status.code = 'DRAFT'
  )
RETURNING dispatch.*;

-- 3.7 Cancel an unposted draft receipt. $1 receipt ID
UPDATE transfer_receipt receipt
SET status_id = cancelled.status_id
FROM document_status cancelled
WHERE receipt.transfer_receipt_id = $1
  AND cancelled.document_type_id = receipt.document_type_id
  AND cancelled.code = 'CANCELLED' AND cancelled.is_active
  AND NOT EXISTS (
      SELECT 1 FROM transfer_receipt_line line
      WHERE line.transfer_receipt_id = receipt.transfer_receipt_id
        AND line.posted_at IS NOT NULL
  )
  AND EXISTS (
      SELECT 1 FROM document_status current_status
      WHERE current_status.status_id = receipt.status_id
        AND current_status.code = 'DRAFT'
  )
RETURNING receipt.*;
