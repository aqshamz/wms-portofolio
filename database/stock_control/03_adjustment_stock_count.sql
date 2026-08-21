-- Authorized inventory adjustments and physical stock counts.
-- Quantities use item base UOM. Execute one numbered operation at a time.

SET search_path TO wms, public;

-- =============================================================================
-- 1. INVENTORY ADJUSTMENT
-- =============================================================================

-- 1.1 Active adjustment types; quantity_effect is configured master data.
SELECT code, name, description, quantity_effect, requires_approval
FROM inventory_adjustment_type
WHERE is_active
ORDER BY name;

-- 1.2 Create a draft adjustment and generate its varchar ID.
-- $1 owner ID, $2 warehouse code, $3 business date,
-- $4 inventory reason-code, $5 notes, $6 actor account ID
WITH context AS (
    SELECT owner.organization_id AS owner_id, warehouse.warehouse_id,
           warehouse.code AS warehouse_code, reason.reason_code_id,
           dt.document_type_id, initial_status.status_id
    FROM organization owner
    JOIN warehouse_owner scope ON scope.owner_id = owner.organization_id AND scope.is_active
    JOIN warehouse ON warehouse.warehouse_id = scope.warehouse_id
                   AND warehouse.code = $2 AND warehouse.is_active
    JOIN reason_code reason ON reason.module_code = 'INVENTORY'
                           AND reason.code = $4 AND reason.is_active
    JOIN document_type dt ON dt.code = 'INVENTORY_ADJUSTMENT' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE owner.organization_id = $1
), numbered AS (
    SELECT context.*,
           generate_document_id('INVENTORY_ADJUSTMENT', NULL, warehouse_code, $3) AS generated_id
    FROM context
)
INSERT INTO inventory_adjustment (
    inventory_adjustment_id, document_type_id, status_id,
    owner_id, warehouse_id, business_date, reason_code_id,
    notes, created_by, updated_by
)
SELECT generated_id, document_type_id, status_id, owner_id, warehouse_id,
       $3, reason_code_id, $5, $6, $6
FROM numbered
RETURNING *;

-- 1.3 Add a DECREASE (or any configured negative-effect) line from a balance.
-- $1 adjustment ID, $2 line number, $3 adjustment-type code,
-- $4 source balance ID, $5 base quantity, $6 notes, $7 actor ID
INSERT INTO inventory_adjustment_line (
    inventory_adjustment_line_id, inventory_adjustment_id, line_no,
    inventory_adjustment_type_id, source_balance_id,
    location_id, item_id, lot_id, handling_unit_id, inventory_status_id,
    adjustment_qty, uom_id, notes, created_by
)
SELECT adjustment.inventory_adjustment_id || '-L-' || lpad($2::text, 4, '0'),
       adjustment.inventory_adjustment_id, $2, adjustment_type.inventory_adjustment_type_id,
       source.balance_id, source.location_id, source.item_id, source.lot_id,
       source.handling_unit_id, source.inventory_status_id,
       $5, source.uom_id, $6, $7
FROM inventory_adjustment adjustment
JOIN document_status status ON status.status_id = adjustment.status_id
JOIN inventory_adjustment_type adjustment_type
  ON adjustment_type.code = $3 AND adjustment_type.is_active
 AND adjustment_type.quantity_effect = -1
JOIN inventory_balance source
  ON source.balance_id = $4
 AND source.owner_id = adjustment.owner_id
 AND source.warehouse_id = adjustment.warehouse_id
WHERE adjustment.inventory_adjustment_id = $1
  AND status.code = 'DRAFT'
  AND source.on_hand_qty - source.reserved_qty >= $5
  AND $5 > 0
RETURNING *;

-- 1.4 Add an INCREASE (or any configured positive-effect) line.
-- $1 adjustment ID, $2 line number, $3 adjustment-type code,
-- $4 location ID, $5 item code, $6 lot ID or NULL,
-- $7 handling-unit ID or NULL, $8 inventory-status code,
-- $9 base quantity, $10 notes, $11 actor ID
INSERT INTO inventory_adjustment_line (
    inventory_adjustment_line_id, inventory_adjustment_id, line_no,
    inventory_adjustment_type_id, source_balance_id,
    location_id, item_id, lot_id, handling_unit_id, inventory_status_id,
    adjustment_qty, uom_id, notes, created_by
)
SELECT adjustment.inventory_adjustment_id || '-L-' || lpad($2::text, 4, '0'),
       adjustment.inventory_adjustment_id, $2, adjustment_type.inventory_adjustment_type_id,
       NULL, location.location_id, item.item_id, lot.lot_id, hu.handling_unit_id,
       inventory_status.inventory_status_id, $9, item.base_uom_id, $10, $11
FROM inventory_adjustment adjustment
JOIN document_status status ON status.status_id = adjustment.status_id
JOIN inventory_adjustment_type adjustment_type
  ON adjustment_type.code = $3 AND adjustment_type.is_active
 AND adjustment_type.quantity_effect = 1
JOIN warehouse_location location
  ON location.location_id = $4
 AND location.warehouse_id = adjustment.warehouse_id AND location.is_active
JOIN item ON item.owner_id = adjustment.owner_id AND item.code = $5 AND item.is_active
JOIN inventory_status ON inventory_status.code = $8 AND inventory_status.is_active
LEFT JOIN inventory_lot lot
  ON lot.lot_id = $6 AND lot.owner_id = adjustment.owner_id AND lot.item_id = item.item_id
LEFT JOIN handling_unit hu
  ON hu.handling_unit_id = $7
 AND hu.owner_id = adjustment.owner_id AND hu.warehouse_id = adjustment.warehouse_id
WHERE adjustment.inventory_adjustment_id = $1
  AND status.code = 'DRAFT'
  AND ($6::varchar IS NULL OR lot.lot_id IS NOT NULL)
  AND ($7::varchar IS NULL OR hu.handling_unit_id IS NOT NULL)
  AND (hu.handling_unit_id IS NULL OR hu.current_location_id IS NULL
       OR hu.current_location_id = location.location_id)
  AND $9 > 0
RETURNING *;

-- 1.5 Delete a draft adjustment line. $1 line ID
DELETE FROM inventory_adjustment_line line
USING inventory_adjustment adjustment, document_status status
WHERE line.inventory_adjustment_line_id = $1
  AND adjustment.inventory_adjustment_id = line.inventory_adjustment_id
  AND status.status_id = adjustment.status_id
  AND status.code = 'DRAFT'
RETURNING line.*;

-- 1.6 Approve an adjustment. $1 adjustment ID, $2 actor ID
UPDATE inventory_adjustment adjustment
SET status_id = approved.status_id,
    approved_at = clock_timestamp(), approved_by = $2,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = adjustment.version_no + 1
FROM document_status current_status
JOIN document_status approved
  ON approved.document_type_id = current_status.document_type_id
 AND approved.code = 'APPROVED' AND approved.is_active
WHERE adjustment.inventory_adjustment_id = $1
  AND current_status.status_id = adjustment.status_id
  AND current_status.code = 'DRAFT'
  AND EXISTS (
      SELECT 1 FROM inventory_adjustment_line line
      WHERE line.inventory_adjustment_id = adjustment.inventory_adjustment_id
  )
RETURNING adjustment.*;

-- 1.7 Post one approved adjustment line atomically.
-- Repeat until all lines post; the last line closes the document.
-- $1 adjustment-line ID, $2 proposed resulting balance ID,
-- $3 movement ID, $4 actor account ID
BEGIN;

SELECT source.balance_id
FROM inventory_adjustment_line line
JOIN inventory_balance source ON source.balance_id = line.source_balance_id
WHERE line.inventory_adjustment_line_id = $1
FOR UPDATE OF source;

WITH context AS (
    SELECT line.*, adjustment.document_type_id, adjustment.owner_id,
           adjustment.warehouse_id, adjustment.business_date,
           adjustment.reason_code_id, adjustment_type.quantity_effect
    FROM inventory_adjustment_line line
    JOIN inventory_adjustment adjustment
      ON adjustment.inventory_adjustment_id = line.inventory_adjustment_id
    JOIN document_status status ON status.status_id = adjustment.status_id
    JOIN inventory_adjustment_type adjustment_type
      ON adjustment_type.inventory_adjustment_type_id = line.inventory_adjustment_type_id
     AND adjustment_type.is_active
    LEFT JOIN inventory_balance source ON source.balance_id = line.source_balance_id
    WHERE line.inventory_adjustment_line_id = $1
      AND status.code = 'APPROVED'
      AND line.posted_at IS NULL
      AND (
          (adjustment_type.quantity_effect = 1 AND line.source_balance_id IS NULL)
          OR
          (adjustment_type.quantity_effect = -1
           AND source.on_hand_qty - source.reserved_qty >= line.adjustment_qty)
      )
), decreased AS (
    UPDATE inventory_balance balance
    SET on_hand_qty = balance.on_hand_qty - context.adjustment_qty,
        version_no = balance.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE context.quantity_effect = -1
      AND balance.balance_id = context.source_balance_id
      AND balance.on_hand_qty - balance.reserved_qty >= context.adjustment_qty
    RETURNING balance.*
), increased AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT $2, context.owner_id, context.warehouse_id, context.location_id,
           context.item_id, context.lot_id, context.handling_unit_id,
           context.inventory_status_id, context.adjustment_qty, 0, context.uom_id
    FROM context
    WHERE context.quantity_effect = 1
    ON CONFLICT (
        owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id
    ) DO UPDATE SET
        on_hand_qty = inventory_balance.on_hand_qty + EXCLUDED.on_hand_qty,
        version_no = inventory_balance.version_no + 1,
        updated_at = clock_timestamp()
    RETURNING *
), resulting_balance AS (
    SELECT * FROM decreased
    UNION ALL
    SELECT * FROM increased
), movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, handling_unit_id,
        from_location_id, to_location_id, from_status_id, to_status_id,
        quantity, uom_id, source_document_id, source_line_id,
        reason_code_id, notes, created_by
    )
    SELECT $3, movement_type.movement_type_id, context.owner_id,
           context.warehouse_id, context.business_date, context.item_id,
           context.lot_id, context.handling_unit_id,
           CASE WHEN context.quantity_effect = -1 THEN context.location_id END,
           CASE WHEN context.quantity_effect =  1 THEN context.location_id END,
           CASE WHEN context.quantity_effect = -1 THEN context.inventory_status_id END,
           CASE WHEN context.quantity_effect =  1 THEN context.inventory_status_id END,
           context.adjustment_qty, context.uom_id,
           context.inventory_adjustment_id,
           context.inventory_adjustment_line_id,
           context.reason_code_id, context.notes, $4
    FROM context CROSS JOIN resulting_balance
    JOIN movement_type ON movement_type.code = 'ADJUSTMENT' AND movement_type.is_active
    RETURNING movement_id
), posted_line AS (
    UPDATE inventory_adjustment_line line
    SET resulting_balance_id = resulting_balance.balance_id,
        movement_id = movement.movement_id,
        posted_at = clock_timestamp(), posted_by = $4
    FROM resulting_balance CROSS JOIN movement
    WHERE line.inventory_adjustment_line_id = $1
    RETURNING line.*
), closed_document AS (
    UPDATE inventory_adjustment adjustment
    SET status_id = posted_status.status_id,
        posted_at = clock_timestamp(), updated_at = clock_timestamp(),
        updated_by = $4, version_no = adjustment.version_no + 1
    FROM posted_line current_line, document_status posted_status
    WHERE adjustment.inventory_adjustment_id = current_line.inventory_adjustment_id
      AND posted_status.document_type_id = adjustment.document_type_id
      AND posted_status.code = 'POSTED'
      AND NOT EXISTS (
          SELECT 1 FROM inventory_adjustment_line other
          WHERE other.inventory_adjustment_id = current_line.inventory_adjustment_id
            AND other.inventory_adjustment_line_id <> current_line.inventory_adjustment_line_id
            AND other.posted_at IS NULL
    )
    RETURNING adjustment.inventory_adjustment_id
), updated_hu AS (
    UPDATE handling_unit hu
    SET current_location_id = context.location_id
    FROM context, posted_line
    WHERE context.quantity_effect = 1
      AND hu.handling_unit_id = context.handling_unit_id
      AND (hu.current_location_id IS NULL OR hu.current_location_id = context.location_id)
    RETURNING hu.handling_unit_id
)
SELECT posted_line.*,
       EXISTS (SELECT 1 FROM closed_document) AS document_completed
FROM posted_line;

COMMIT;

-- 1.8 Adjustment list and totals.
-- $1 owner ID, $2 warehouse ID, $3 status or NULL,
-- $4 date from, $5 date until, $6 limit, $7 offset
SELECT adjustment.inventory_adjustment_id, adjustment.business_date,
       status.code AS status_code, reason.code AS reason_code,
       count(line.inventory_adjustment_line_id) AS line_count,
       sum(CASE WHEN adjustment_type.quantity_effect = 1
                THEN line.adjustment_qty ELSE 0 END) AS increase_qty,
       sum(CASE WHEN adjustment_type.quantity_effect = -1
                THEN line.adjustment_qty ELSE 0 END) AS decrease_qty,
       adjustment.approved_at, adjustment.posted_at,
       count(*) OVER () AS total_rows
FROM inventory_adjustment adjustment
JOIN document_status status ON status.status_id = adjustment.status_id
JOIN reason_code reason ON reason.reason_code_id = adjustment.reason_code_id
LEFT JOIN inventory_adjustment_line line
  ON line.inventory_adjustment_id = adjustment.inventory_adjustment_id
LEFT JOIN inventory_adjustment_type adjustment_type
  ON adjustment_type.inventory_adjustment_type_id = line.inventory_adjustment_type_id
WHERE adjustment.owner_id = $1
  AND adjustment.warehouse_id = $2
  AND ($3::varchar IS NULL OR status.code = $3)
  AND adjustment.business_date BETWEEN $4 AND $5
GROUP BY adjustment.inventory_adjustment_id, adjustment.business_date,
         status.code, reason.code, adjustment.approved_at, adjustment.posted_at
ORDER BY adjustment.business_date DESC, adjustment.inventory_adjustment_id DESC
LIMIT $6 OFFSET $7;

-- 1.9 Cancel an adjustment before any line posts.
-- $1 adjustment ID, $2 actor ID
UPDATE inventory_adjustment adjustment
SET status_id = cancelled.status_id,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = adjustment.version_no + 1
FROM document_status cancelled
WHERE adjustment.inventory_adjustment_id = $1
  AND cancelled.document_type_id = adjustment.document_type_id
  AND cancelled.code = 'CANCELLED' AND cancelled.is_active
  AND NOT EXISTS (
      SELECT 1 FROM inventory_adjustment_line line
      WHERE line.inventory_adjustment_id = adjustment.inventory_adjustment_id
        AND line.posted_at IS NOT NULL
  )
  AND EXISTS (
      SELECT 1 FROM document_status current_status
      WHERE current_status.status_id = adjustment.status_id
        AND NOT current_status.is_final
  )
RETURNING adjustment.*;

-- =============================================================================
-- 2. PHYSICAL / CYCLE STOCK COUNT
-- =============================================================================

-- 2.1 Active count types.
SELECT code, name, description, requires_freeze
FROM stock_count_type
WHERE is_active
ORDER BY name;

-- 2.2 Create a draft count and generate its varchar ID.
-- A type configured with requires_freeze automatically marks the count frozen.
-- $1 owner ID, $2 warehouse code, $3 business date,
-- $4 count-type code, $5 inventory reason-code,
-- $6 notes, $7 actor account ID
WITH context AS (
    SELECT owner.organization_id AS owner_id, warehouse.warehouse_id,
           warehouse.code AS warehouse_code, count_type.stock_count_type_id,
           count_type.requires_freeze, reason.reason_code_id,
           dt.document_type_id, initial_status.status_id
    FROM organization owner
    JOIN warehouse_owner scope ON scope.owner_id = owner.organization_id AND scope.is_active
    JOIN warehouse ON warehouse.warehouse_id = scope.warehouse_id
                   AND warehouse.code = $2 AND warehouse.is_active
    JOIN stock_count_type count_type ON count_type.code = $4 AND count_type.is_active
    JOIN reason_code reason ON reason.module_code = 'INVENTORY'
                           AND reason.code = $5 AND reason.is_active
    JOIN document_type dt ON dt.code = 'STOCK_COUNT' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE owner.organization_id = $1
), numbered AS (
    SELECT context.*,
           generate_document_id('STOCK_COUNT', NULL, warehouse_code, $3) AS generated_id
    FROM context
)
INSERT INTO stock_count (
    stock_count_id, document_type_id, status_id, stock_count_type_id,
    owner_id, warehouse_id, business_date, reason_code_id,
    freeze_inventory, notes, created_by, updated_by
)
SELECT generated_id, document_type_id, status_id, stock_count_type_id,
       owner_id, warehouse_id, $3, reason_code_id,
       requires_freeze, $6, $7, $7
FROM numbered
RETURNING *;

-- 2.3 Generate a one-time balance snapshot for the selected scope.
-- NULL filters mean all values in the warehouse. Run before COUNTING.
-- $1 stock-count ID, $2 location ID or NULL, $3 item ID or NULL,
-- $4 inventory-status ID or NULL, $5 actor account ID
WITH count_context AS (
    SELECT count.*
    FROM stock_count count
    JOIN document_status status ON status.status_id = count.status_id
    WHERE count.stock_count_id = $1
      AND status.code = 'DRAFT'
      AND NOT EXISTS (
          SELECT 1 FROM stock_count_line existing
          WHERE existing.stock_count_id = count.stock_count_id
      )
), snapshot AS (
    SELECT balance.*,
           row_number() OVER (
               ORDER BY location.code, item.code,
                        balance.lot_id NULLS FIRST,
                        balance.handling_unit_id NULLS FIRST,
                        balance.inventory_status_id
           ) AS generated_line_no
    FROM count_context count
    JOIN inventory_balance balance
      ON balance.owner_id = count.owner_id
     AND balance.warehouse_id = count.warehouse_id
    JOIN warehouse_location location ON location.location_id = balance.location_id
    JOIN item ON item.item_id = balance.item_id
    WHERE balance.on_hand_qty > 0
      AND ($2::uuid IS NULL OR balance.location_id = $2)
      AND ($3::uuid IS NULL OR balance.item_id = $3)
      AND ($4::uuid IS NULL OR balance.inventory_status_id = $4)
)
INSERT INTO stock_count_line (
    stock_count_line_id, stock_count_id, line_no, balance_id,
    location_id, item_id, lot_id, handling_unit_id, inventory_status_id,
    system_qty, system_version_no, uom_id, created_by
)
SELECT $1 || '-L-' || lpad(snapshot.generated_line_no::text, 6, '0'),
       $1, snapshot.generated_line_no, snapshot.balance_id,
       snapshot.location_id, snapshot.item_id, snapshot.lot_id,
       snapshot.handling_unit_id, snapshot.inventory_status_id,
       snapshot.on_hand_qty, snapshot.version_no, snapshot.uom_id, $5
FROM snapshot
RETURNING *;

-- 2.4 Add found stock that was not in the system snapshot.
-- This line starts with system quantity zero and has no balance/version.
-- $1 count ID, $2 line number, $3 location ID, $4 item code,
-- $5 lot ID or NULL, $6 handling-unit ID or NULL,
-- $7 inventory-status code, $8 counted base quantity, $9 actor ID
INSERT INTO stock_count_line (
    stock_count_line_id, stock_count_id, line_no, balance_id,
    location_id, item_id, lot_id, handling_unit_id, inventory_status_id,
    system_qty, system_version_no, counted_qty, uom_id,
    counted_at, counted_by, created_by
)
SELECT count.stock_count_id || '-L-' || lpad($2::text, 6, '0'),
       count.stock_count_id, $2, NULL, location.location_id, item.item_id,
       lot.lot_id, hu.handling_unit_id, inventory_status.inventory_status_id,
       0, NULL, $8, item.base_uom_id, clock_timestamp(), $9, $9
FROM stock_count count
JOIN document_status status ON status.status_id = count.status_id
JOIN warehouse_location location
  ON location.location_id = $3
 AND location.warehouse_id = count.warehouse_id AND location.is_active
JOIN item ON item.owner_id = count.owner_id AND item.code = $4 AND item.is_active
JOIN inventory_status ON inventory_status.code = $7 AND inventory_status.is_active
LEFT JOIN inventory_lot lot
  ON lot.lot_id = $5 AND lot.owner_id = count.owner_id AND lot.item_id = item.item_id
LEFT JOIN handling_unit hu
  ON hu.handling_unit_id = $6
 AND hu.owner_id = count.owner_id AND hu.warehouse_id = count.warehouse_id
WHERE count.stock_count_id = $1
  AND status.code IN ('DRAFT', 'COUNTING')
  AND ($5::varchar IS NULL OR lot.lot_id IS NOT NULL)
  AND ($6::varchar IS NULL OR hu.handling_unit_id IS NOT NULL)
  AND (hu.handling_unit_id IS NULL OR hu.current_location_id IS NULL
       OR hu.current_location_id = location.location_id)
  AND $8 > 0
  AND NOT EXISTS (
      SELECT 1
      FROM inventory_balance existing
      WHERE existing.owner_id = count.owner_id
        AND existing.warehouse_id = count.warehouse_id
        AND existing.location_id = location.location_id
        AND existing.item_id = item.item_id
        AND existing.lot_id IS NOT DISTINCT FROM lot.lot_id
        AND existing.handling_unit_id IS NOT DISTINCT FROM hu.handling_unit_id
        AND existing.inventory_status_id = inventory_status.inventory_status_id
  )
RETURNING *;

-- 2.5 Start counting. $1 count ID, $2 actor ID
UPDATE stock_count count
SET status_id = counting.status_id,
    started_at = clock_timestamp(), updated_at = clock_timestamp(),
    updated_by = $2, version_no = count.version_no + 1
FROM document_status current_status
JOIN document_status counting
  ON counting.document_type_id = current_status.document_type_id
 AND counting.code = 'COUNTING' AND counting.is_active
WHERE count.stock_count_id = $1
  AND current_status.status_id = count.status_id
  AND current_status.code = 'DRAFT'
  AND EXISTS (
      SELECT 1 FROM stock_count_line line
      WHERE line.stock_count_id = count.stock_count_id
  )
RETURNING count.*;

-- 2.6 Record/re-record a physical quantity while COUNTING.
-- $1 count-line ID, $2 counted base quantity, $3 actor ID
UPDATE stock_count_line line
SET counted_qty = $2,
    counted_at = clock_timestamp(),
    counted_by = $3
FROM stock_count count, document_status status
WHERE line.stock_count_line_id = $1
  AND count.stock_count_id = line.stock_count_id
  AND status.status_id = count.status_id
  AND status.code = 'COUNTING'
  AND $2 >= 0
RETURNING line.*;

-- 2.7 Submit all completed counts for review. $1 count ID, $2 actor ID
UPDATE stock_count count
SET status_id = review.status_id,
    completed_at = clock_timestamp(), reviewed_at = clock_timestamp(),
    reviewed_by = $2, updated_at = clock_timestamp(), updated_by = $2,
    version_no = count.version_no + 1
FROM document_status current_status
JOIN document_status review
  ON review.document_type_id = current_status.document_type_id
 AND review.code = 'REVIEW' AND review.is_active
WHERE count.stock_count_id = $1
  AND current_status.status_id = count.status_id
  AND current_status.code = 'COUNTING'
  AND NOT EXISTS (
      SELECT 1 FROM stock_count_line line
      WHERE line.stock_count_id = count.stock_count_id
        AND line.counted_qty IS NULL
  )
RETURNING count.*;

-- 2.8 Variance review. $1 count ID, $2 only variances
SELECT line.stock_count_line_id, line.line_no,
       location.code AS location_code, item.code AS item_code,
       lot.lot_number, line.handling_unit_id,
       inventory_status.code AS inventory_status_code,
       line.system_qty, line.counted_qty,
       line.counted_qty - line.system_qty AS variance_qty,
       unit.code AS uom_code,
       line.system_version_no,
       balance.version_no AS current_version_no,
       (line.balance_id IS NULL OR balance.version_no = line.system_version_no)
           AS snapshot_is_current,
       line.posted_at, line.variance_movement_id
FROM stock_count_line line
JOIN warehouse_location location ON location.location_id = line.location_id
JOIN item ON item.item_id = line.item_id
JOIN inventory_status ON inventory_status.inventory_status_id = line.inventory_status_id
JOIN uom unit ON unit.uom_id = line.uom_id
LEFT JOIN inventory_lot lot ON lot.lot_id = line.lot_id
LEFT JOIN inventory_balance balance ON balance.balance_id = line.balance_id
WHERE line.stock_count_id = $1
  AND (NOT $2 OR line.counted_qty IS DISTINCT FROM line.system_qty)
ORDER BY line.line_no;

-- 2.9 Post one reviewed count line atomically.
-- Existing balances only post if their version still matches the snapshot.
-- Zero-variance lines are marked posted without creating a movement.
-- $1 count-line ID, $2 proposed balance ID for found stock,
-- $3 variance movement ID, $4 actor account ID
BEGIN;

SELECT balance.balance_id
FROM stock_count_line line
JOIN inventory_balance balance ON balance.balance_id = line.balance_id
WHERE line.stock_count_line_id = $1
FOR UPDATE OF balance;

WITH context AS (
    SELECT line.*, count.document_type_id, count.owner_id,
           count.warehouse_id, count.business_date, count.reason_code_id,
           line.counted_qty - line.system_qty AS variance_qty,
           balance.on_hand_qty AS current_on_hand_qty,
           balance.reserved_qty AS current_reserved_qty,
           balance.version_no AS current_version_no
    FROM stock_count_line line
    JOIN stock_count count ON count.stock_count_id = line.stock_count_id
    JOIN document_status status ON status.status_id = count.status_id
    LEFT JOIN inventory_balance balance ON balance.balance_id = line.balance_id
    WHERE line.stock_count_line_id = $1
      AND status.code = 'REVIEW'
      AND line.counted_qty IS NOT NULL
      AND line.posted_at IS NULL
      AND (
          (line.balance_id IS NULL AND line.system_qty = 0 AND line.counted_qty > 0)
          OR
          (line.balance_id IS NOT NULL
           AND balance.version_no = line.system_version_no
           AND balance.on_hand_qty = line.system_qty
           AND line.counted_qty >= balance.reserved_qty)
      )
), negative_update AS (
    UPDATE inventory_balance balance
    SET on_hand_qty = context.counted_qty,
        version_no = balance.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE context.variance_qty < 0
      AND balance.balance_id = context.balance_id
      AND balance.version_no = context.system_version_no
      AND context.counted_qty >= balance.reserved_qty
    RETURNING balance.*
), positive_existing AS (
    UPDATE inventory_balance balance
    SET on_hand_qty = context.counted_qty,
        version_no = balance.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE context.variance_qty > 0
      AND context.balance_id IS NOT NULL
      AND balance.balance_id = context.balance_id
      AND balance.version_no = context.system_version_no
    RETURNING balance.*
), positive_found AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT $2, context.owner_id,
           context.warehouse_id, context.location_id, context.item_id,
           context.lot_id, context.handling_unit_id,
           context.inventory_status_id, context.counted_qty, 0, context.uom_id
    FROM context
    WHERE context.variance_qty > 0
      AND context.balance_id IS NULL
    ON CONFLICT (
        owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id
    ) DO NOTHING
    RETURNING *
), unchanged AS (
    SELECT balance.*
    FROM context
    JOIN inventory_balance balance ON balance.balance_id = context.balance_id
    WHERE context.variance_qty = 0
), resulting_balance AS (
    SELECT * FROM negative_update
    UNION ALL SELECT * FROM positive_existing
    UNION ALL SELECT * FROM positive_found
    UNION ALL SELECT * FROM unchanged
), movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, handling_unit_id,
        from_location_id, to_location_id, from_status_id, to_status_id,
        quantity, uom_id, source_document_id, source_line_id,
        reason_code_id, created_by
    )
    SELECT $3, movement_type.movement_type_id, context.owner_id,
           context.warehouse_id, context.business_date, context.item_id,
           context.lot_id, context.handling_unit_id,
           CASE WHEN context.variance_qty < 0 THEN context.location_id END,
           CASE WHEN context.variance_qty > 0 THEN context.location_id END,
           CASE WHEN context.variance_qty < 0 THEN context.inventory_status_id END,
           CASE WHEN context.variance_qty > 0 THEN context.inventory_status_id END,
           abs(context.variance_qty), context.uom_id,
           context.stock_count_id, context.stock_count_line_id,
           context.reason_code_id, $4
    FROM context CROSS JOIN resulting_balance
    JOIN movement_type ON movement_type.code = 'COUNT_CORRECTION' AND movement_type.is_active
    WHERE context.variance_qty <> 0
    RETURNING movement_id
), posted_line AS (
    UPDATE stock_count_line line
    SET balance_id = resulting_balance.balance_id,
        variance_movement_id = movement.movement_id,
        posted_at = clock_timestamp(), posted_by = $4
    FROM resulting_balance
    LEFT JOIN movement ON true
    WHERE line.stock_count_line_id = $1
    RETURNING line.*
), closed_count AS (
    UPDATE stock_count count
    SET status_id = posted_status.status_id,
        posted_at = clock_timestamp(), posted_by = $4,
        updated_at = clock_timestamp(), updated_by = $4,
        version_no = count.version_no + 1
    FROM posted_line current_line, document_status posted_status
    WHERE count.stock_count_id = current_line.stock_count_id
      AND posted_status.document_type_id = count.document_type_id
      AND posted_status.code = 'POSTED'
      AND NOT EXISTS (
          SELECT 1 FROM stock_count_line other
          WHERE other.stock_count_id = current_line.stock_count_id
            AND other.stock_count_line_id <> current_line.stock_count_line_id
            AND other.posted_at IS NULL
      )
    RETURNING count.stock_count_id
), updated_hu AS (
    UPDATE handling_unit hu
    SET current_location_id = context.location_id
    FROM context, posted_line
    WHERE hu.handling_unit_id = context.handling_unit_id
      AND (hu.current_location_id IS NULL OR hu.current_location_id = context.location_id)
    RETURNING hu.handling_unit_id
)
SELECT posted_line.*,
       EXISTS (SELECT 1 FROM closed_count) AS count_completed
FROM posted_line;

COMMIT;

-- 2.10 Count list and progress.
-- $1 owner ID, $2 warehouse ID, $3 status or NULL,
-- $4 date from, $5 date until, $6 limit, $7 offset
SELECT count.stock_count_id, count.business_date,
       count_type.code AS count_type_code, status.code AS status_code,
       count.freeze_inventory, count.started_at, count.completed_at,
       count.posted_at,
       count(line.stock_count_line_id) AS total_lines,
       count(line.counted_at) AS counted_lines,
       count(line.posted_at) AS posted_lines,
       count(*) FILTER (
           WHERE line.counted_qty IS DISTINCT FROM line.system_qty
       ) AS variance_lines,
       count(*) OVER () AS total_rows
FROM stock_count count
JOIN stock_count_type count_type
  ON count_type.stock_count_type_id = count.stock_count_type_id
JOIN document_status status ON status.status_id = count.status_id
LEFT JOIN stock_count_line line ON line.stock_count_id = count.stock_count_id
WHERE count.owner_id = $1
  AND count.warehouse_id = $2
  AND ($3::varchar IS NULL OR status.code = $3)
  AND count.business_date BETWEEN $4 AND $5
GROUP BY count.stock_count_id, count.business_date, count_type.code,
         status.code, count.freeze_inventory, count.started_at,
         count.completed_at, count.posted_at
ORDER BY count.business_date DESC, count.stock_count_id DESC
LIMIT $6 OFFSET $7;

-- 2.11 Return a reviewed count to COUNTING for recount.
-- $1 count ID, $2 actor ID
UPDATE stock_count count
SET status_id = counting.status_id,
    reviewed_at = NULL, reviewed_by = NULL,
    completed_at = NULL,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = count.version_no + 1
FROM document_status current_status
JOIN document_status counting
  ON counting.document_type_id = current_status.document_type_id
 AND counting.code = 'COUNTING' AND counting.is_active
WHERE count.stock_count_id = $1
  AND current_status.status_id = count.status_id
  AND current_status.code = 'REVIEW'
  AND NOT EXISTS (
      SELECT 1 FROM stock_count_line line
      WHERE line.stock_count_id = count.stock_count_id
        AND line.posted_at IS NOT NULL
  )
RETURNING count.*;

-- 2.12 Cancel a count before variance posting starts.
-- $1 count ID, $2 actor ID
UPDATE stock_count count
SET status_id = cancelled.status_id,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = count.version_no + 1
FROM document_status cancelled
WHERE count.stock_count_id = $1
  AND cancelled.document_type_id = count.document_type_id
  AND cancelled.code = 'CANCELLED' AND cancelled.is_active
  AND NOT EXISTS (
      SELECT 1 FROM stock_count_line line
      WHERE line.stock_count_id = count.stock_count_id
        AND line.posted_at IS NOT NULL
  )
  AND EXISTS (
      SELECT 1 FROM document_status current_status
      WHERE current_status.status_id = count.status_id
        AND NOT current_status.is_final
  )
RETURNING count.*;
