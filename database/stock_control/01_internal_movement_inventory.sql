-- Stock inquiry, internal movement, and inventory status-change operations.
-- Execute one numbered operation at a time with prepared parameters.

SET search_path TO wms, public;

-- =============================================================================
-- 1. STOCK INQUIRY AND AUDIT
-- =============================================================================

-- 1.1 Search current balances.
-- $1 owner ID, $2 warehouse ID, $3 item code or NULL, $4 location code or NULL,
-- $5 inventory-status code or NULL, $6 lot number or NULL,
-- $7 include zero balances, $8 limit, $9 offset
SELECT
    balance.balance_id,
    item.code AS item_code,
    item.name AS item_name,
    location.code AS location_code,
    status.code AS inventory_status_code,
    lot.lot_number,
    balance.handling_unit_id,
    balance.on_hand_qty,
    balance.reserved_qty,
    balance.on_hand_qty - balance.reserved_qty AS available_qty,
    unit.code AS base_uom_code,
    balance.version_no,
    balance.updated_at,
    count(*) OVER () AS total_rows
FROM inventory_balance balance
JOIN item ON item.item_id = balance.item_id
JOIN warehouse_location location ON location.location_id = balance.location_id
JOIN inventory_status status ON status.inventory_status_id = balance.inventory_status_id
JOIN uom unit ON unit.uom_id = balance.uom_id
LEFT JOIN inventory_lot lot ON lot.lot_id = balance.lot_id
WHERE balance.owner_id = $1
  AND balance.warehouse_id = $2
  AND ($3::varchar IS NULL OR item.code = $3)
  AND ($4::varchar IS NULL OR location.code = $4)
  AND ($5::varchar IS NULL OR status.code = $5)
  AND ($6::varchar IS NULL OR lot.lot_number = $6)
  AND ($7 OR balance.on_hand_qty > 0)
ORDER BY item.code, location.pick_sequence NULLS LAST, location.code,
         lot.lot_number NULLS FIRST, status.code
LIMIT $8 OFFSET $9;

-- 1.2 Aggregate stock by item and status.
-- $1 owner ID, $2 warehouse ID, $3 item-code search or NULL
SELECT
    item.code AS item_code,
    item.name AS item_name,
    status.code AS inventory_status_code,
    sum(balance.on_hand_qty) AS on_hand_qty,
    sum(balance.reserved_qty) AS reserved_qty,
    sum(balance.on_hand_qty - balance.reserved_qty) AS available_qty,
    unit.code AS base_uom_code
FROM inventory_balance balance
JOIN item ON item.item_id = balance.item_id
JOIN inventory_status status ON status.inventory_status_id = balance.inventory_status_id
JOIN uom unit ON unit.uom_id = balance.uom_id
WHERE balance.owner_id = $1
  AND balance.warehouse_id = $2
  AND ($3::varchar IS NULL OR item.code ILIKE '%' || $3 || '%')
GROUP BY item.item_id, item.code, item.name, status.code, unit.code
ORDER BY item.code, status.code;

-- 1.3 Immutable movement ledger.
-- $1 owner ID, $2 warehouse ID, $3 item code or NULL,
-- $4 source document ID or NULL, $5 occurred-from or NULL,
-- $6 occurred-until or NULL, $7 limit, $8 offset
SELECT
    movement.movement_id,
    movement.occurred_at,
    movement.business_date,
    movement_type.code AS movement_type_code,
    item.code AS item_code,
    lot.lot_number,
    movement.handling_unit_id,
    from_location.code AS from_location_code,
    to_location.code AS to_location_code,
    from_status.code AS from_status_code,
    to_status.code AS to_status_code,
    movement.quantity,
    unit.code AS uom_code,
    movement.source_document_id,
    movement.source_line_id,
    reason.code AS reason_code,
    movement.notes,
    movement.created_by,
    count(*) OVER () AS total_rows
FROM inventory_movement movement
JOIN movement_type ON movement_type.movement_type_id = movement.movement_type_id
JOIN item ON item.item_id = movement.item_id
JOIN uom unit ON unit.uom_id = movement.uom_id
LEFT JOIN inventory_lot lot ON lot.lot_id = movement.lot_id
LEFT JOIN warehouse_location from_location ON from_location.location_id = movement.from_location_id
LEFT JOIN warehouse_location to_location ON to_location.location_id = movement.to_location_id
LEFT JOIN inventory_status from_status ON from_status.inventory_status_id = movement.from_status_id
LEFT JOIN inventory_status to_status ON to_status.inventory_status_id = movement.to_status_id
LEFT JOIN reason_code reason ON reason.reason_code_id = movement.reason_code_id
WHERE movement.owner_id = $1
  AND movement.warehouse_id = $2
  AND ($3::varchar IS NULL OR item.code = $3)
  AND ($4::varchar IS NULL OR movement.source_document_id = $4)
  AND ($5::timestamptz IS NULL OR movement.occurred_at >= $5)
  AND ($6::timestamptz IS NULL OR movement.occurred_at < $6)
ORDER BY movement.occurred_at DESC, movement.movement_id DESC
LIMIT $7 OFFSET $8;

-- 1.4 Active internal-movement setup values.
SELECT code, name, description, requires_approval
FROM internal_move_type
WHERE is_active
ORDER BY name;

-- =============================================================================
-- 2. INTERNAL MOVEMENT / REPLENISHMENT / CONSOLIDATION
-- =============================================================================

-- 2.1 Create a draft movement and generate its varchar ID.
-- $1 owner ID, $2 warehouse code, $3 business date, $4 movement-type code,
-- $5 reason-code or NULL, $6 notes, $7 actor account ID
WITH context AS (
    SELECT
        owner.organization_id AS owner_id,
        warehouse.warehouse_id,
        warehouse.code AS warehouse_code,
        move_type.internal_move_type_id,
        reason.reason_code_id,
        dt.document_type_id,
        initial_status.status_id
    FROM organization owner
    JOIN warehouse_owner scope ON scope.owner_id = owner.organization_id AND scope.is_active
    JOIN warehouse ON warehouse.warehouse_id = scope.warehouse_id
                   AND warehouse.code = $2 AND warehouse.is_active
    JOIN internal_move_type move_type ON move_type.code = $4 AND move_type.is_active
    JOIN document_type dt ON dt.code = 'INTERNAL_MOVE' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    LEFT JOIN reason_code reason
      ON reason.module_code = 'INVENTORY' AND reason.code = $5 AND reason.is_active
    WHERE owner.organization_id = $1
      AND ($5::varchar IS NULL OR reason.reason_code_id IS NOT NULL)
), numbered AS (
    SELECT context.*,
           generate_document_id('INTERNAL_MOVE', NULL, context.warehouse_code, $3) AS generated_id
    FROM context
)
INSERT INTO internal_move_order (
    internal_move_id, document_type_id, status_id, internal_move_type_id,
    owner_id, warehouse_id, business_date, reason_code_id, notes,
    created_by, updated_by
)
SELECT generated_id, document_type_id, status_id, internal_move_type_id,
       owner_id, warehouse_id, $3, reason_code_id, $6, $7, $7
FROM numbered
RETURNING *;

-- 2.2 Add a draft movement line. Quantity must use the item base UOM.
-- $1 movement ID, $2 line number, $3 source balance ID,
-- $4 target location ID, $5 planned base quantity, $6 assignee or NULL,
-- $7 actor account ID
INSERT INTO internal_move_order_line (
    internal_move_line_id, internal_move_id, line_no, source_balance_id,
    target_location_id, planned_qty, uom_id, assigned_to, created_by
)
SELECT
    move.internal_move_id || '-L-' || lpad($2::text, 4, '0'),
    move.internal_move_id,
    $2,
    source.balance_id,
    target.location_id,
    $5,
    source.uom_id,
    $6,
    $7
FROM internal_move_order move
JOIN document_status move_status ON move_status.status_id = move.status_id
JOIN inventory_balance source
  ON source.balance_id = $3
 AND source.owner_id = move.owner_id
 AND source.warehouse_id = move.warehouse_id
JOIN warehouse_location target
  ON target.location_id = $4
 AND target.warehouse_id = move.warehouse_id
 AND target.is_active
WHERE move.internal_move_id = $1
  AND move_status.code = 'DRAFT'
  AND source.location_id <> target.location_id
  AND source.uom_id = (SELECT base_uom_id FROM item WHERE item_id = source.item_id)
  AND source.on_hand_qty - source.reserved_qty >= $5
  AND $5 > 0
RETURNING *;

-- 2.3 Update a draft line. $1 line ID, $2 target location ID,
-- $3 planned quantity, $4 assignee or NULL, $5 actor account ID
UPDATE internal_move_order_line line
SET target_location_id = target.location_id,
    planned_qty = $3,
    assigned_to = $4
FROM internal_move_order move, document_status status,
     inventory_balance source, warehouse_location target
WHERE line.internal_move_line_id = $1
  AND move.internal_move_id = line.internal_move_id
  AND status.status_id = move.status_id
  AND source.balance_id = line.source_balance_id
  AND target.location_id = $2
  AND target.warehouse_id = move.warehouse_id
  AND target.is_active
  AND status.code = 'DRAFT'
  AND source.location_id <> target.location_id
  AND source.on_hand_qty - source.reserved_qty >= $3
  AND $3 > 0
RETURNING line.*;

-- 2.4 Delete a draft line. $1 line ID
DELETE FROM internal_move_order_line line
USING internal_move_order move, document_status status
WHERE line.internal_move_line_id = $1
  AND move.internal_move_id = line.internal_move_id
  AND status.status_id = move.status_id
  AND status.code = 'DRAFT'
RETURNING line.*;

-- 2.5 Approve/release a movement with at least one line.
-- $1 movement ID, $2 actor account ID
UPDATE internal_move_order move
SET status_id = approved.status_id,
    approved_at = clock_timestamp(),
    approved_by = $2,
    updated_at = clock_timestamp(),
    updated_by = $2,
    version_no = move.version_no + 1
FROM document_status current_status
JOIN document_status approved
  ON approved.document_type_id = current_status.document_type_id
 AND approved.code = 'APPROVED' AND approved.is_active
WHERE move.internal_move_id = $1
  AND current_status.status_id = move.status_id
  AND current_status.code = 'DRAFT'
  AND EXISTS (
      SELECT 1 FROM internal_move_order_line line
      WHERE line.internal_move_id = move.internal_move_id
  )
  AND EXISTS (
      SELECT 1 FROM document_status_transition transition
      WHERE transition.document_type_id = move.document_type_id
        AND transition.from_status_id = move.status_id
        AND transition.to_status_id = approved.status_id
        AND transition.is_active
  )
RETURNING move.*;

-- 2.6 Execute all or part of one movement line atomically.
-- A handling unit may only be moved as a whole, preventing one HU from appearing
-- in two physical locations.
-- $1 movement-line ID, $2 executed base quantity,
-- $3 proposed destination balance ID, $4 movement ID,
-- $5 execution ID, $6 actor account ID
BEGIN;

SELECT source.balance_id
FROM internal_move_order_line line
JOIN inventory_balance source ON source.balance_id = line.source_balance_id
WHERE line.internal_move_line_id = $1
FOR UPDATE OF source;

WITH context AS (
    SELECT
        line.*,
        move.document_type_id,
        move.status_id AS move_status_id,
        move.owner_id,
        move.warehouse_id,
        move.business_date,
        move.reason_code_id,
        source.location_id AS source_location_id,
        source.item_id,
        source.lot_id,
        source.handling_unit_id,
        source.inventory_status_id,
        source.on_hand_qty,
        source.reserved_qty,
        source.uom_id AS source_uom_id
    FROM internal_move_order_line line
    JOIN internal_move_order move ON move.internal_move_id = line.internal_move_id
    JOIN document_status status ON status.status_id = move.status_id
    JOIN inventory_balance source ON source.balance_id = line.source_balance_id
    JOIN warehouse_location target
      ON target.location_id = line.target_location_id
     AND target.warehouse_id = move.warehouse_id
     AND target.is_active
    WHERE line.internal_move_line_id = $1
      AND status.code IN ('APPROVED', 'IN_PROGRESS')
      AND line.completed_qty + $2 <= line.planned_qty
      AND source.on_hand_qty - source.reserved_qty >= $2
      AND source.location_id <> line.target_location_id
      AND ($2 > 0)
      AND (
          source.handling_unit_id IS NULL
          OR (source.reserved_qty = 0 AND $2 = source.on_hand_qty)
      )
), decremented AS (
    UPDATE inventory_balance source
    SET on_hand_qty = source.on_hand_qty - $2,
        version_no = source.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE source.balance_id = context.source_balance_id
      AND source.on_hand_qty - source.reserved_qty >= $2
    RETURNING source.*
), destination AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT
        $3, context.owner_id, context.warehouse_id, context.target_location_id,
        context.item_id, context.lot_id, context.handling_unit_id,
        context.inventory_status_id, $2, 0, context.source_uom_id
    FROM context JOIN decremented ON decremented.balance_id = context.source_balance_id
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
        quantity, uom_id, source_document_id, source_line_id,
        reason_code_id, created_by
    )
    SELECT
        $4, movement_type.movement_type_id, context.owner_id,
        context.warehouse_id, context.business_date, context.item_id,
        context.lot_id, context.handling_unit_id, context.source_location_id,
        context.target_location_id, context.inventory_status_id,
        context.inventory_status_id, $2, context.source_uom_id,
        context.internal_move_id, context.internal_move_line_id,
        context.reason_code_id, $6
    FROM context
    CROSS JOIN destination
    JOIN movement_type ON movement_type.code = 'INTERNAL_MOVE' AND movement_type.is_active
    RETURNING movement_id
), execution AS (
    INSERT INTO internal_move_execution (
        internal_move_execution_id, internal_move_line_id, quantity,
        source_balance_id, destination_balance_id, movement_id,
        executed_by
    )
    SELECT $5, context.internal_move_line_id, $2, context.source_balance_id,
           destination.balance_id, movement.movement_id, $6
    FROM context CROSS JOIN destination CROSS JOIN movement
    RETURNING *
), updated_line AS (
    UPDATE internal_move_order_line line
    SET completed_qty = line.completed_qty + $2
    FROM execution
    WHERE line.internal_move_line_id = execution.internal_move_line_id
    RETURNING line.*
), updated_move AS (
    UPDATE internal_move_order move
    SET status_id = next_status.status_id,
        completed_at = CASE WHEN next_status.code = 'COMPLETED'
                            THEN clock_timestamp() ELSE move.completed_at END,
        updated_at = clock_timestamp(),
        updated_by = $6,
        version_no = move.version_no + 1
    FROM updated_line current_line, document_status next_status
    WHERE move.internal_move_id = current_line.internal_move_id
      AND next_status.document_type_id = move.document_type_id
      AND next_status.code = CASE
         WHEN current_line.completed_qty = current_line.planned_qty
          AND NOT EXISTS (
              SELECT 1 FROM internal_move_order_line other
              WHERE other.internal_move_id = current_line.internal_move_id
                AND other.internal_move_line_id <> current_line.internal_move_line_id
                AND other.completed_qty < other.planned_qty
          ) THEN 'COMPLETED' ELSE 'IN_PROGRESS' END
    RETURNING move.*
), updated_hu AS (
    UPDATE handling_unit hu
    SET current_location_id = destination.location_id
    FROM context CROSS JOIN destination
    WHERE hu.handling_unit_id = context.handling_unit_id
    RETURNING hu.handling_unit_id
)
SELECT updated_move.internal_move_id,
       updated_line.internal_move_line_id,
       updated_line.completed_qty,
       execution.internal_move_execution_id,
       execution.destination_balance_id,
       execution.movement_id
FROM updated_move CROSS JOIN updated_line CROSS JOIN execution;

COMMIT;

-- 2.7 List movements.
-- $1 owner ID, $2 warehouse ID, $3 status code or NULL,
-- $4 business-date from, $5 business-date until, $6 limit, $7 offset
SELECT
    move.internal_move_id,
    move.business_date,
    move_type.code AS movement_type_code,
    status.code AS status_code,
    reason.code AS reason_code,
    move.requested_at,
    move.completed_at,
    count(line.internal_move_line_id) AS line_count,
    sum(line.planned_qty) AS planned_qty,
    sum(line.completed_qty) AS completed_qty,
    count(*) OVER () AS total_rows
FROM internal_move_order move
JOIN internal_move_type move_type ON move_type.internal_move_type_id = move.internal_move_type_id
JOIN document_status status ON status.status_id = move.status_id
LEFT JOIN reason_code reason ON reason.reason_code_id = move.reason_code_id
LEFT JOIN internal_move_order_line line ON line.internal_move_id = move.internal_move_id
WHERE move.owner_id = $1
  AND move.warehouse_id = $2
  AND ($3::varchar IS NULL OR status.code = $3)
  AND move.business_date >= $4
  AND move.business_date <= $5
GROUP BY move.internal_move_id, move.business_date, move_type.code, status.code,
         reason.code, move.requested_at, move.completed_at
ORDER BY move.business_date DESC, move.internal_move_id DESC
LIMIT $6 OFFSET $7;

-- 2.8 Cancel an unexecuted movement. $1 movement ID, $2 actor ID
UPDATE internal_move_order move
SET status_id = cancelled.status_id,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = move.version_no + 1
FROM document_status cancelled
WHERE move.internal_move_id = $1
  AND cancelled.document_type_id = move.document_type_id
  AND cancelled.code = 'CANCELLED' AND cancelled.is_active
  AND NOT EXISTS (
      SELECT 1
      FROM internal_move_order_line line
      JOIN internal_move_execution execution
        ON execution.internal_move_line_id = line.internal_move_line_id
      WHERE line.internal_move_id = move.internal_move_id
  )
  AND EXISTS (
      SELECT 1 FROM document_status current_status
      WHERE current_status.status_id = move.status_id
        AND NOT current_status.is_final
  )
RETURNING move.*;

-- =============================================================================
-- 3. INVENTORY STATUS HOLD / RELEASE / RECLASSIFICATION
-- =============================================================================

-- 3.1 Create a draft status-change document.
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
    JOIN document_type dt ON dt.code = 'INVENTORY_STATUS_CHANGE' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE owner.organization_id = $1
), numbered AS (
    SELECT context.*,
           generate_document_id('INVENTORY_STATUS_CHANGE', NULL, warehouse_code, $3) AS generated_id
    FROM context
)
INSERT INTO inventory_status_change (
    inventory_status_change_id, document_type_id, status_id,
    owner_id, warehouse_id, business_date, reason_code_id, notes,
    created_by, updated_by
)
SELECT generated_id, document_type_id, status_id, owner_id, warehouse_id,
       $3, reason_code_id, $5, $6, $6
FROM numbered
RETURNING *;

-- 3.2 Add a draft status-change line.
-- $1 document ID, $2 line number, $3 source balance ID,
-- $4 target inventory-status code, $5 base quantity, $6 actor account ID
INSERT INTO inventory_status_change_line (
    inventory_status_change_line_id, inventory_status_change_id, line_no,
    source_balance_id, target_inventory_status_id, quantity, uom_id, created_by
)
SELECT change.inventory_status_change_id || '-L-' || lpad($2::text, 4, '0'),
       change.inventory_status_change_id, $2, source.balance_id,
       target_status.inventory_status_id, $5, source.uom_id, $6
FROM inventory_status_change change
JOIN document_status document_status ON document_status.status_id = change.status_id
JOIN inventory_balance source
  ON source.balance_id = $3
 AND source.owner_id = change.owner_id
 AND source.warehouse_id = change.warehouse_id
JOIN inventory_status target_status ON target_status.code = $4 AND target_status.is_active
WHERE change.inventory_status_change_id = $1
  AND document_status.code = 'DRAFT'
  AND source.inventory_status_id <> target_status.inventory_status_id
  AND source.on_hand_qty - source.reserved_qty >= $5
  AND (
      source.handling_unit_id IS NULL
      OR (source.reserved_qty = 0 AND $5 = source.on_hand_qty)
  )
  AND $5 > 0
RETURNING *;

-- 3.3 Approve a status-change document. $1 document ID, $2 actor ID
UPDATE inventory_status_change change
SET status_id = approved.status_id,
    approved_at = clock_timestamp(), approved_by = $2,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = change.version_no + 1
FROM document_status current_status
JOIN document_status approved
  ON approved.document_type_id = current_status.document_type_id
 AND approved.code = 'APPROVED' AND approved.is_active
WHERE change.inventory_status_change_id = $1
  AND current_status.status_id = change.status_id
  AND current_status.code = 'DRAFT'
  AND EXISTS (
      SELECT 1 FROM inventory_status_change_line line
      WHERE line.inventory_status_change_id = change.inventory_status_change_id
  )
RETURNING change.*;

-- 3.4 Post one approved status-change line atomically.
-- Repeat until every line is posted; the final line closes the document.
-- $1 status-change line ID, $2 proposed destination balance ID,
-- $3 movement ID, $4 actor account ID
BEGIN;

SELECT source.balance_id
FROM inventory_status_change_line line
JOIN inventory_balance source ON source.balance_id = line.source_balance_id
WHERE line.inventory_status_change_line_id = $1
FOR UPDATE OF source;

WITH context AS (
    SELECT line.*, change.document_type_id, change.status_id AS document_status_id,
           change.owner_id, change.warehouse_id, change.business_date,
           change.reason_code_id, source.location_id, source.item_id,
           source.lot_id, source.handling_unit_id,
           source.inventory_status_id AS source_status_id,
           source.on_hand_qty, source.reserved_qty
    FROM inventory_status_change_line line
    JOIN inventory_status_change change
      ON change.inventory_status_change_id = line.inventory_status_change_id
    JOIN document_status status ON status.status_id = change.status_id
    JOIN inventory_balance source ON source.balance_id = line.source_balance_id
    WHERE line.inventory_status_change_line_id = $1
      AND status.code = 'APPROVED'
      AND line.posted_at IS NULL
      AND source.inventory_status_id <> line.target_inventory_status_id
      AND source.on_hand_qty - source.reserved_qty >= line.quantity
      AND (
          source.handling_unit_id IS NULL
          OR (source.reserved_qty = 0 AND line.quantity = source.on_hand_qty)
      )
), decremented AS (
    UPDATE inventory_balance source
    SET on_hand_qty = source.on_hand_qty - context.quantity,
        version_no = source.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE source.balance_id = context.source_balance_id
      AND source.on_hand_qty - source.reserved_qty >= context.quantity
    RETURNING source.*
), destination AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT $2, context.owner_id, context.warehouse_id, context.location_id,
           context.item_id, context.lot_id, context.handling_unit_id,
           context.target_inventory_status_id, context.quantity, 0, context.uom_id
    FROM context JOIN decremented ON decremented.balance_id = context.source_balance_id
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
        item_id, lot_id, handling_unit_id, from_location_id, to_location_id,
        from_status_id, to_status_id, quantity, uom_id,
        source_document_id, source_line_id, reason_code_id, created_by
    )
    SELECT $3, movement_type.movement_type_id, context.owner_id,
           context.warehouse_id, context.business_date, context.item_id,
           context.lot_id, context.handling_unit_id, context.location_id,
           context.location_id, context.source_status_id,
           context.target_inventory_status_id, context.quantity, context.uom_id,
           context.inventory_status_change_id,
           context.inventory_status_change_line_id, context.reason_code_id, $4
    FROM context CROSS JOIN destination
    JOIN movement_type ON movement_type.code = 'STATUS_CHANGE' AND movement_type.is_active
    RETURNING movement_id
), posted_line AS (
    UPDATE inventory_status_change_line line
    SET destination_balance_id = destination.balance_id,
        movement_id = movement.movement_id,
        posted_at = clock_timestamp(), posted_by = $4
    FROM destination CROSS JOIN movement
    WHERE line.inventory_status_change_line_id = $1
    RETURNING line.*
), closed_document AS (
    UPDATE inventory_status_change change
    SET status_id = posted.status_id,
        posted_at = clock_timestamp(), updated_at = clock_timestamp(),
        updated_by = $4, version_no = change.version_no + 1
    FROM posted_line current_line, document_status posted
    WHERE change.inventory_status_change_id = current_line.inventory_status_change_id
      AND posted.document_type_id = change.document_type_id
      AND posted.code = 'POSTED'
      AND NOT EXISTS (
          SELECT 1 FROM inventory_status_change_line other
          WHERE other.inventory_status_change_id = current_line.inventory_status_change_id
            AND other.inventory_status_change_line_id <> current_line.inventory_status_change_line_id
            AND other.posted_at IS NULL
      )
    RETURNING change.inventory_status_change_id
)
SELECT posted_line.*,
       EXISTS (SELECT 1 FROM closed_document) AS document_completed
FROM posted_line;

COMMIT;

-- 3.5 Status-change detail and posting progress. $1 document ID
SELECT
    change.inventory_status_change_id,
    document_status.code AS document_status_code,
    reason.code AS reason_code,
    line.line_no,
    line.inventory_status_change_line_id,
    item.code AS item_code,
    location.code AS location_code,
    lot.lot_number,
    source_status.code AS source_status_code,
    target_status.code AS target_status_code,
    line.quantity,
    unit.code AS uom_code,
    line.posted_at,
    line.movement_id
FROM inventory_status_change change
JOIN document_status document_status ON document_status.status_id = change.status_id
JOIN reason_code reason ON reason.reason_code_id = change.reason_code_id
JOIN inventory_status_change_line line
  ON line.inventory_status_change_id = change.inventory_status_change_id
JOIN inventory_balance source ON source.balance_id = line.source_balance_id
JOIN item ON item.item_id = source.item_id
JOIN warehouse_location location ON location.location_id = source.location_id
JOIN inventory_status source_status ON source_status.inventory_status_id = source.inventory_status_id
JOIN inventory_status target_status
  ON target_status.inventory_status_id = line.target_inventory_status_id
JOIN uom unit ON unit.uom_id = line.uom_id
LEFT JOIN inventory_lot lot ON lot.lot_id = source.lot_id
WHERE change.inventory_status_change_id = $1
ORDER BY line.line_no;

-- 3.6 Cancel a status-change document only before any line posts.
-- $1 document ID, $2 actor ID
UPDATE inventory_status_change change
SET status_id = cancelled.status_id,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = change.version_no + 1
FROM document_status cancelled
WHERE change.inventory_status_change_id = $1
  AND cancelled.document_type_id = change.document_type_id
  AND cancelled.code = 'CANCELLED' AND cancelled.is_active
  AND NOT EXISTS (
      SELECT 1 FROM inventory_status_change_line line
      WHERE line.inventory_status_change_id = change.inventory_status_change_id
        AND line.posted_at IS NOT NULL
  )
  AND EXISTS (
      SELECT 1 FROM document_status current_status
      WHERE current_status.status_id = change.status_id
        AND NOT current_status.is_final
  )
RETURNING change.*;
