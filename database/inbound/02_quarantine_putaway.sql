-- Focused inbound queries: quarantine disposition, rework, and putaway.
-- Execute one numbered operation at a time with prepared parameters.

SET search_path TO wms, public;

-- =============================================================================
-- 1. QUARANTINE CASE AND CLIENT DECISION
-- =============================================================================

-- 1.1 List open quarantine cases.
-- $1 owner ID, $2 warehouse ID, $3 limit, $4 offset
SELECT
    qc.quarantine_case_id,
    qc.parent_quarantine_case_id,
    qc.opened_at,
    status.code AS status_code,
    item.code AS item_code,
    item.name AS item_name,
    lot.lot_number,
    qc.quarantine_qty,
    unit.code AS uom_code,
    balance.location_id,
    location.code AS quarantine_location_code,
    qc.quarantine_qty - COALESCE(decided.decided_qty, 0) AS undecided_qty,
    count(*) OVER () AS total_rows
FROM quarantine_case qc
JOIN document_status status ON status.status_id = qc.status_id
JOIN receipt_inventory batch ON batch.receipt_inventory_id = qc.receipt_inventory_id
JOIN item ON item.item_id = batch.item_id
LEFT JOIN inventory_lot lot ON lot.lot_id = batch.lot_id
JOIN uom unit ON unit.uom_id = qc.uom_id
JOIN inventory_balance balance ON balance.balance_id = qc.quarantine_balance_id
JOIN warehouse_location location ON location.location_id = balance.location_id
LEFT JOIN LATERAL (
    SELECT sum(disposition.disposition_qty) AS decided_qty
    FROM quarantine_disposition disposition
    JOIN document_status disposition_status
      ON disposition_status.status_id = disposition.status_id
    WHERE disposition.quarantine_case_id = qc.quarantine_case_id
      AND NOT disposition_status.is_cancelled
) decided ON true
WHERE qc.owner_id = $1
  AND qc.warehouse_id = $2
  AND NOT status.is_final
ORDER BY qc.opened_at, qc.quarantine_case_id
LIMIT $3 OFFSET $4;

-- 1.2 Get quarantine case, inspection history, and decisions. $1 case ID
SELECT
    qc.*,
    case_status.code AS case_status_code,
    item.code AS item_code,
    item.name AS item_name,
    lot.lot_number,
    balance.on_hand_qty AS current_quarantine_balance_qty,
    COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object(
                'inspectionId', inspection.inspection_id,
                'parentInspectionId', inspection.parent_inspection_id,
                'status', quality_status.code,
                'result', result.code,
                'inspectedQty', inspection.inspected_qty,
                'passedQty', inspection.passed_qty,
                'failedQty', inspection.failed_qty,
                'inspectedAt', inspection.inspected_at
            ) ORDER BY inspection.created_at
        )
        FROM quality_inspection inspection
        JOIN quality_status ON quality_status.quality_status_id = inspection.quality_status_id
        LEFT JOIN inspection_result result
          ON result.inspection_result_id = inspection.inspection_result_id
        WHERE inspection.receipt_inventory_id = qc.receipt_inventory_id
    ), '[]'::jsonb) AS inspections,
    COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object(
                'dispositionId', disposition.quarantine_disposition_id,
                'type', disposition_type.code,
                'quantity', disposition.disposition_qty,
                'status', disposition_status.code,
                'reference', disposition.client_decision_reference,
                'decidedAt', disposition.decided_at,
                'processedAt', disposition.processed_at
            ) ORDER BY disposition.decided_at
        )
        FROM quarantine_disposition disposition
        JOIN quarantine_disposition_type disposition_type
          ON disposition_type.quarantine_disposition_type_id = disposition.quarantine_disposition_type_id
        JOIN document_status disposition_status
          ON disposition_status.status_id = disposition.status_id
        WHERE disposition.quarantine_case_id = qc.quarantine_case_id
    ), '[]'::jsonb) AS dispositions
FROM quarantine_case qc
JOIN document_status case_status ON case_status.status_id = qc.status_id
JOIN receipt_inventory batch ON batch.receipt_inventory_id = qc.receipt_inventory_id
JOIN item ON item.item_id = batch.item_id
LEFT JOIN inventory_lot lot ON lot.lot_id = batch.lot_id
JOIN inventory_balance balance ON balance.balance_id = qc.quarantine_balance_id
WHERE qc.quarantine_case_id = $1;

-- 1.3 Record a client decision without allocating more than case quantity.
-- The case row is locked so two concurrent decisions cannot over-allocate.
-- $1 disposition ID, $2 quarantine-case ID, $3 type code,
-- $4 base quantity, $5 client decision reference, $6 notes,
-- $7 decided_at, $8 actor account ID
BEGIN;

SELECT quarantine_case_id
FROM quarantine_case
WHERE quarantine_case_id = $2
FOR UPDATE;

WITH locked_case AS (
    SELECT qc.*
    FROM quarantine_case qc
    JOIN document_status status ON status.status_id = qc.status_id
    WHERE qc.quarantine_case_id = $2
      AND NOT status.is_final
    FOR UPDATE OF qc
),
remaining AS (
    SELECT
        locked_case.*,
        locked_case.quarantine_qty - COALESCE((
            SELECT sum(existing.disposition_qty)
            FROM quarantine_disposition existing
            JOIN document_status existing_status ON existing_status.status_id = existing.status_id
            WHERE existing.quarantine_case_id = locked_case.quarantine_case_id
              AND NOT existing_status.is_cancelled
        ), 0) AS remaining_qty
    FROM locked_case
),
created_disposition AS (
    INSERT INTO quarantine_disposition (
        quarantine_disposition_id,
        quarantine_case_id,
        document_type_id,
        status_id,
        quarantine_disposition_type_id,
        disposition_qty,
        uom_id,
        client_decision_reference,
        decision_notes,
        decided_at,
        decided_by,
        created_by
    )
    SELECT
        $1,
        remaining.quarantine_case_id,
        dt.document_type_id,
        initial_status.status_id,
        disposition_type.quarantine_disposition_type_id,
        $4,
        remaining.uom_id,
        $5,
        $6,
        $7,
        $8,
        $8
    FROM remaining
    JOIN quarantine_disposition_type disposition_type
      ON disposition_type.code = $3
     AND disposition_type.is_active
    JOIN document_type dt ON dt.code = 'QUARANTINE_DISPOSITION' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial
     AND initial_status.is_active
    WHERE $4 > 0
      AND $4 <= remaining.remaining_qty
    RETURNING *
),
updated_case AS (
    UPDATE quarantine_case qc
    SET status_id = partial_status.status_id,
        updated_at = clock_timestamp(),
        updated_by = $8,
        version_no = qc.version_no + 1
    FROM document_status partial_status
    WHERE qc.quarantine_case_id = $2
      AND partial_status.document_type_id = qc.document_type_id
      AND partial_status.code = 'PARTIALLY_DECIDED'
      AND EXISTS (SELECT 1 FROM created_disposition)
    RETURNING qc.quarantine_case_id
)
SELECT created_disposition.*
FROM created_disposition
CROSS JOIN updated_case;

COMMIT;

-- =============================================================================
-- 2. ACCEPT QUARANTINED STOCK
-- =============================================================================

-- 2.1 Process ACCEPT: QUARANTINE -> AVAILABLE and create putaway task.
-- $1 disposition ID, $2 proposed available balance ID, $3 movement ID,
-- $4 putaway-task ID, $5 target storage location ID,
-- $6 task-priority code, $7 actor account ID
BEGIN;

SELECT qc.quarantine_case_id
FROM quarantine_disposition disposition
JOIN quarantine_case qc ON qc.quarantine_case_id = disposition.quarantine_case_id
WHERE disposition.quarantine_disposition_id = $1
FOR UPDATE OF qc;

WITH disposition_context AS (
    SELECT
        disposition.*,
        qc.document_type_id AS case_document_type_id,
        qc.quarantine_qty,
        qc.quarantine_balance_id,
        qc.owner_id,
        qc.warehouse_id,
        qc.receipt_inventory_id,
        batch.item_id,
        batch.lot_id,
        batch.handling_unit_id,
        batch.base_uom_id,
        receipt.business_date
    FROM quarantine_disposition disposition
    JOIN document_status disposition_status ON disposition_status.status_id = disposition.status_id
    JOIN quarantine_disposition_type disposition_type
      ON disposition_type.quarantine_disposition_type_id = disposition.quarantine_disposition_type_id
     AND disposition_type.releases_to_available
    JOIN quarantine_case qc ON qc.quarantine_case_id = disposition.quarantine_case_id
    JOIN receipt_inventory batch ON batch.receipt_inventory_id = qc.receipt_inventory_id
    JOIN receipt_line line ON line.receipt_line_id = batch.receipt_line_id
    JOIN receipt ON receipt.receipt_id = line.receipt_id
    WHERE disposition.quarantine_disposition_id = $1
      AND disposition_status.code = 'DECIDED'
    FOR UPDATE OF qc
),
decremented_source AS (
    UPDATE inventory_balance source
    SET on_hand_qty = source.on_hand_qty - context.disposition_qty,
        version_no = source.version_no + 1,
        updated_at = clock_timestamp()
    FROM disposition_context context
    WHERE source.balance_id = context.quarantine_balance_id
      AND source.on_hand_qty - source.reserved_qty >= context.disposition_qty
    RETURNING source.*
),
available_balance AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT
        $2,
        source.owner_id,
        source.warehouse_id,
        source.location_id,
        source.item_id,
        source.lot_id,
        source.handling_unit_id,
        available_status.inventory_status_id,
        context.disposition_qty,
        0,
        source.uom_id
    FROM decremented_source source
    CROSS JOIN disposition_context context
    JOIN inventory_status available_status
      ON available_status.code = 'AVAILABLE'
     AND available_status.is_active
    ON CONFLICT (
        owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id
    ) DO UPDATE SET
        on_hand_qty = inventory_balance.on_hand_qty + EXCLUDED.on_hand_qty,
        version_no = inventory_balance.version_no + 1,
        updated_at = clock_timestamp()
    RETURNING *
),
created_movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, handling_unit_id,
        from_location_id, to_location_id, from_status_id, to_status_id,
        quantity, uom_id, source_document_id, source_line_id, created_by
    )
    SELECT
        $3,
        movement_type.movement_type_id,
        source.owner_id,
        source.warehouse_id,
        context.business_date,
        source.item_id,
        source.lot_id,
        source.handling_unit_id,
        source.location_id,
        source.location_id,
        source.inventory_status_id,
        destination.inventory_status_id,
        context.disposition_qty,
        source.uom_id,
        context.quarantine_case_id,
        context.quarantine_disposition_id,
        $7
    FROM decremented_source source
    CROSS JOIN available_balance destination
    CROSS JOIN disposition_context context
    JOIN movement_type ON movement_type.code = 'STATUS_CHANGE' AND movement_type.is_active
    RETURNING movement_id
),
created_task AS (
    INSERT INTO putaway_task (
        putaway_task_id, task_type_id, task_status_id, task_priority_id,
        receipt_inventory_id, source_balance_id,
        owner_id, warehouse_id, item_id, lot_id, handling_unit_id,
        source_location_id, target_location_id,
        planned_qty, completed_qty, uom_id, created_by
    )
    SELECT
        $4,
        task_type.task_type_id,
        task_status.task_status_id,
        priority.task_priority_id,
        context.receipt_inventory_id,
        destination.balance_id,
        context.owner_id,
        context.warehouse_id,
        context.item_id,
        context.lot_id,
        context.handling_unit_id,
        destination.location_id,
        target.location_id,
        context.disposition_qty,
        0,
        context.base_uom_id,
        $7
    FROM disposition_context context
    CROSS JOIN available_balance destination
    CROSS JOIN created_movement
    JOIN task_type ON task_type.code = 'PUTAWAY' AND task_type.is_active
    JOIN task_status ON task_status.code = 'OPEN' AND task_status.is_active
    JOIN task_priority priority ON priority.code = $6 AND priority.is_active
    JOIN warehouse_location target
      ON target.location_id = $5
     AND target.warehouse_id = context.warehouse_id
     AND target.is_active
    JOIN location_type target_type
      ON target_type.location_type_id = target.location_type_id
     AND target_type.allows_storage
    RETURNING *
),
processed_disposition AS (
    UPDATE quarantine_disposition disposition
    SET status_id = processed_status.status_id,
        processed_at = clock_timestamp(),
        inventory_movement_id = movement.movement_id,
        resulting_balance_id = task.source_balance_id
    FROM document_status processed_status
    CROSS JOIN created_movement movement
    CROSS JOIN created_task task
    WHERE disposition.quarantine_disposition_id = $1
      AND processed_status.document_type_id = disposition.document_type_id
      AND processed_status.code = 'PROCESSED'
    RETURNING disposition.*
),
updated_case AS (
    UPDATE quarantine_case qc
    SET status_id = target_status.status_id,
        closed_at = CASE WHEN target_status.code = 'CLOSED' THEN clock_timestamp() ELSE NULL END,
        updated_at = clock_timestamp(),
        updated_by = $7,
        version_no = qc.version_no + 1
    FROM disposition_context context
    JOIN document_status target_status
      ON target_status.document_type_id = context.case_document_type_id
     AND target_status.code = CASE
         WHEN COALESCE((
             SELECT sum(existing.disposition_qty)
             FROM quarantine_disposition existing
             WHERE existing.quarantine_case_id = context.quarantine_case_id
               AND existing.processed_at IS NOT NULL
               AND existing.quarantine_disposition_id <> context.quarantine_disposition_id
         ), 0) + context.disposition_qty >= context.quarantine_qty
         THEN 'CLOSED'
         ELSE 'PARTIALLY_DECIDED'
     END
    WHERE qc.quarantine_case_id = context.quarantine_case_id
      AND EXISTS (SELECT 1 FROM processed_disposition)
    RETURNING qc.*
)
SELECT
    processed_disposition.*,
    created_task.putaway_task_id,
    updated_case.status_id AS quarantine_case_status_id
FROM processed_disposition
CROSS JOIN created_task
CROSS JOIN updated_case;

COMMIT;

-- =============================================================================
-- 3. RETURN OR DISPOSE
-- =============================================================================

-- 3.1 Process a removal disposition. The configured disposition type selects
-- RETURN_TO_VENDOR versus DISPOSE movement; the query contains no action switch.
-- $1 disposition ID, $2 movement ID, $3 actor account ID
BEGIN;

SELECT qc.quarantine_case_id
FROM quarantine_disposition disposition
JOIN quarantine_case qc ON qc.quarantine_case_id = disposition.quarantine_case_id
WHERE disposition.quarantine_disposition_id = $1
FOR UPDATE OF qc;

WITH disposition_context AS (
    SELECT
        disposition.*,
        disposition_type.removal_movement_type_id,
        qc.document_type_id AS case_document_type_id,
        qc.quarantine_qty,
        qc.quarantine_balance_id,
        qc.owner_id,
        qc.warehouse_id,
        batch.item_id,
        batch.lot_id,
        batch.handling_unit_id,
        receipt.business_date
    FROM quarantine_disposition disposition
    JOIN document_status disposition_status ON disposition_status.status_id = disposition.status_id
    JOIN quarantine_disposition_type disposition_type
      ON disposition_type.quarantine_disposition_type_id = disposition.quarantine_disposition_type_id
     AND disposition_type.removes_inventory
     AND disposition_type.removal_movement_type_id IS NOT NULL
    JOIN quarantine_case qc ON qc.quarantine_case_id = disposition.quarantine_case_id
    JOIN receipt_inventory batch ON batch.receipt_inventory_id = qc.receipt_inventory_id
    JOIN receipt_line line ON line.receipt_line_id = batch.receipt_line_id
    JOIN receipt ON receipt.receipt_id = line.receipt_id
    WHERE disposition.quarantine_disposition_id = $1
      AND disposition_status.code = 'DECIDED'
    FOR UPDATE OF qc
),
decremented_source AS (
    UPDATE inventory_balance source
    SET on_hand_qty = source.on_hand_qty - context.disposition_qty,
        version_no = source.version_no + 1,
        updated_at = clock_timestamp()
    FROM disposition_context context
    WHERE source.balance_id = context.quarantine_balance_id
      AND source.on_hand_qty - source.reserved_qty >= context.disposition_qty
    RETURNING source.*
),
created_movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, handling_unit_id,
        from_location_id, from_status_id,
        quantity, uom_id, source_document_id, source_line_id, created_by
    )
    SELECT
        $2,
        context.removal_movement_type_id,
        source.owner_id,
        source.warehouse_id,
        context.business_date,
        source.item_id,
        source.lot_id,
        source.handling_unit_id,
        source.location_id,
        source.inventory_status_id,
        context.disposition_qty,
        source.uom_id,
        context.quarantine_case_id,
        context.quarantine_disposition_id,
        $3
    FROM decremented_source source
    CROSS JOIN disposition_context context
    RETURNING movement_id
),
processed_disposition AS (
    UPDATE quarantine_disposition disposition
    SET status_id = processed_status.status_id,
        processed_at = clock_timestamp(),
        inventory_movement_id = movement.movement_id
    FROM document_status processed_status
    CROSS JOIN created_movement movement
    WHERE disposition.quarantine_disposition_id = $1
      AND processed_status.document_type_id = disposition.document_type_id
      AND processed_status.code = 'PROCESSED'
    RETURNING disposition.*
),
updated_case AS (
    UPDATE quarantine_case qc
    SET status_id = target_status.status_id,
        closed_at = CASE WHEN target_status.code = 'CLOSED' THEN clock_timestamp() ELSE NULL END,
        updated_at = clock_timestamp(),
        updated_by = $3,
        version_no = qc.version_no + 1
    FROM disposition_context context
    JOIN document_status target_status
      ON target_status.document_type_id = context.case_document_type_id
     AND target_status.code = CASE
         WHEN COALESCE((
             SELECT sum(existing.disposition_qty)
             FROM quarantine_disposition existing
             WHERE existing.quarantine_case_id = context.quarantine_case_id
               AND existing.processed_at IS NOT NULL
               AND existing.quarantine_disposition_id <> context.quarantine_disposition_id
         ), 0) + context.disposition_qty >= context.quarantine_qty
         THEN 'CLOSED'
         ELSE 'PARTIALLY_DECIDED'
     END
    WHERE qc.quarantine_case_id = context.quarantine_case_id
      AND EXISTS (SELECT 1 FROM processed_disposition)
    RETURNING qc.*
)
SELECT processed_disposition.*, updated_case.status_id AS quarantine_case_status_id
FROM processed_disposition
CROSS JOIN updated_case;

COMMIT;

-- =============================================================================
-- 4. REWORK AND REINSPECTION
-- =============================================================================

-- 4.1 Create rework task for a REWORK client disposition.
-- $1 rework-task ID, $2 disposition ID, $3 priority code,
-- $4 assigned account ID or null, $5 instructions, $6 actor account ID
INSERT INTO rework_task (
    rework_task_id,
    quarantine_disposition_id,
    task_type_id,
    task_status_id,
    task_priority_id,
    planned_qty,
    completed_qty,
    uom_id,
    assigned_to,
    work_instructions,
    created_by
)
SELECT
    $1,
    disposition.quarantine_disposition_id,
    task_type.task_type_id,
    task_status.task_status_id,
    priority.task_priority_id,
    disposition.disposition_qty,
    0,
    disposition.uom_id,
    $4,
    $5,
    $6
FROM quarantine_disposition disposition
JOIN document_status disposition_status ON disposition_status.status_id = disposition.status_id
JOIN quarantine_disposition_type disposition_type
  ON disposition_type.quarantine_disposition_type_id = disposition.quarantine_disposition_type_id
 AND disposition_type.requires_reinspection
JOIN task_type ON task_type.code = 'REWORK' AND task_type.is_active
JOIN task_status ON task_status.code = 'OPEN' AND task_status.is_active
JOIN task_priority priority ON priority.code = $3 AND priority.is_active
WHERE disposition.quarantine_disposition_id = $2
  AND disposition_status.code = 'DECIDED'
RETURNING *;

-- 4.2 Assign rework. $1 task ID, $2 account ID
UPDATE rework_task task
SET task_status_id = assigned_status.task_status_id,
    assigned_to = $2
FROM task_status assigned_status
WHERE task.rework_task_id = $1
  AND assigned_status.code = 'ASSIGNED'
  AND EXISTS (
      SELECT 1
      FROM task_status current_status
      JOIN task_status_transition transition
        ON transition.from_status_id = current_status.task_status_id
       AND transition.to_status_id = assigned_status.task_status_id
       AND transition.is_active
      WHERE current_status.task_status_id = task.task_status_id
  )
RETURNING task.*;

-- 4.3 Start rework. $1 task ID, $2 assigned account ID
UPDATE rework_task task
SET task_status_id = in_progress_status.task_status_id,
    started_at = clock_timestamp()
FROM task_status in_progress_status
WHERE task.rework_task_id = $1
  AND task.assigned_to = $2
  AND in_progress_status.code = 'IN_PROGRESS'
  AND EXISTS (
      SELECT 1
      FROM task_status_transition transition
      WHERE transition.from_status_id = task.task_status_id
        AND transition.to_status_id = in_progress_status.task_status_id
        AND transition.is_active
  )
RETURNING task.*;

-- 4.4 Complete rework and create pending reinspection atomically.
-- $1 rework-task ID, $2 reinspection ID, $3 result notes, $4 actor ID
WITH completed_task AS (
    UPDATE rework_task task
    SET task_status_id = completed_status.task_status_id,
        completed_qty = task.planned_qty,
        completed_at = clock_timestamp(),
        result_notes = $3
    FROM task_status completed_status
    WHERE task.rework_task_id = $1
      AND completed_status.code = 'COMPLETED'
      AND EXISTS (
          SELECT 1
          FROM task_status_transition transition
          WHERE transition.from_status_id = task.task_status_id
            AND transition.to_status_id = completed_status.task_status_id
            AND transition.is_active
      )
    RETURNING task.*
),
created_inspection AS (
    INSERT INTO quality_inspection (
        inspection_id,
        receipt_inventory_id,
        parent_inspection_id,
        quality_status_id,
        inspected_qty,
        created_by
    )
    SELECT
        $2,
        qc.receipt_inventory_id,
        qc.inspection_id,
        pending_status.quality_status_id,
        completed_task.completed_qty,
        $4
    FROM completed_task
    JOIN quarantine_disposition disposition
      ON disposition.quarantine_disposition_id = completed_task.quarantine_disposition_id
    JOIN quarantine_case qc ON qc.quarantine_case_id = disposition.quarantine_case_id
    JOIN quality_status pending_status ON pending_status.code = 'PENDING' AND pending_status.is_active
    RETURNING *
),
linked_task AS (
    UPDATE rework_task task
    SET reinspection_id = inspection.inspection_id
    FROM created_inspection inspection
    WHERE task.rework_task_id = $1
    RETURNING task.*
)
SELECT * FROM created_inspection;

-- 4.5 Reinspection PASS. This performs the same inventory transition as ACCEPT,
-- creates putaway, and completes the REWORK disposition.
-- $1 reinspection ID, $2 proposed available balance ID, $3 movement ID,
-- $4 putaway-task ID, $5 target storage location ID,
-- $6 priority code, $7 actor ID
BEGIN;

SELECT qc.quarantine_case_id
FROM quality_inspection inspection
JOIN rework_task task ON task.reinspection_id = inspection.inspection_id
JOIN quarantine_disposition disposition
  ON disposition.quarantine_disposition_id = task.quarantine_disposition_id
JOIN quarantine_case qc ON qc.quarantine_case_id = disposition.quarantine_case_id
WHERE inspection.inspection_id = $1
FOR UPDATE OF qc;

WITH context AS (
    SELECT
        inspection.*,
        task.rework_task_id,
        disposition.quarantine_disposition_id,
        disposition.quarantine_case_id,
        disposition.disposition_qty,
        disposition.document_type_id AS disposition_document_type_id,
        qc.document_type_id AS case_document_type_id,
        qc.quarantine_qty,
        qc.quarantine_balance_id,
        qc.owner_id,
        qc.warehouse_id,
        qc.receipt_inventory_id,
        batch.item_id,
        batch.lot_id,
        batch.handling_unit_id,
        batch.base_uom_id,
        receipt.business_date
    FROM quality_inspection inspection
    JOIN rework_task task ON task.reinspection_id = inspection.inspection_id
    JOIN quarantine_disposition disposition
      ON disposition.quarantine_disposition_id = task.quarantine_disposition_id
    JOIN quarantine_case qc ON qc.quarantine_case_id = disposition.quarantine_case_id
    JOIN receipt_inventory batch ON batch.receipt_inventory_id = qc.receipt_inventory_id
    JOIN receipt_line line ON line.receipt_line_id = batch.receipt_line_id
    JOIN receipt ON receipt.receipt_id = line.receipt_id
    WHERE inspection.inspection_id = $1
      AND inspection.inspection_result_id IS NULL
    FOR UPDATE OF qc
),
decremented_source AS (
    UPDATE inventory_balance source
    SET on_hand_qty = source.on_hand_qty - context.inspected_qty,
        version_no = source.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE source.balance_id = context.quarantine_balance_id
      AND source.on_hand_qty - source.reserved_qty >= context.inspected_qty
    RETURNING source.*
),
available_balance AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT
        $2, source.owner_id, source.warehouse_id, source.location_id,
        source.item_id, source.lot_id, source.handling_unit_id,
        available_status.inventory_status_id,
        context.inspected_qty, 0, source.uom_id
    FROM decremented_source source
    CROSS JOIN context
    JOIN inventory_status available_status
      ON available_status.code = 'AVAILABLE'
     AND available_status.is_active
    ON CONFLICT (
        owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id
    ) DO UPDATE SET
        on_hand_qty = inventory_balance.on_hand_qty + EXCLUDED.on_hand_qty,
        version_no = inventory_balance.version_no + 1,
        updated_at = clock_timestamp()
    RETURNING *
),
completed_inspection AS (
    UPDATE quality_inspection inspection
    SET quality_status_id = quality_status.quality_status_id,
        inspection_result_id = result.inspection_result_id,
        passed_qty = inspection.inspected_qty,
        failed_qty = 0,
        inspected_at = clock_timestamp(),
        inspected_by = $7
    FROM quality_status
    CROSS JOIN inspection_result result
    WHERE inspection.inspection_id = $1
      AND quality_status.code = 'PASSED'
      AND result.code = 'ACCEPTED'
      AND EXISTS (SELECT 1 FROM available_balance)
    RETURNING inspection.*
),
created_movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, handling_unit_id,
        from_location_id, to_location_id, from_status_id, to_status_id,
        quantity, uom_id, source_document_id, source_line_id, created_by
    )
    SELECT
        $3, movement_type.movement_type_id,
        source.owner_id, source.warehouse_id, context.business_date,
        source.item_id, source.lot_id, source.handling_unit_id,
        source.location_id, source.location_id,
        source.inventory_status_id, destination.inventory_status_id,
        context.inspected_qty, source.uom_id,
        context.quarantine_case_id, context.quarantine_disposition_id, $7
    FROM decremented_source source
    CROSS JOIN available_balance destination
    CROSS JOIN context
    CROSS JOIN completed_inspection
    JOIN movement_type ON movement_type.code = 'STATUS_CHANGE' AND movement_type.is_active
    RETURNING movement_id
),
created_task AS (
    INSERT INTO putaway_task (
        putaway_task_id, task_type_id, task_status_id, task_priority_id,
        receipt_inventory_id, source_balance_id,
        owner_id, warehouse_id, item_id, lot_id, handling_unit_id,
        source_location_id, target_location_id,
        planned_qty, completed_qty, uom_id, created_by
    )
    SELECT
        $4, task_type.task_type_id, task_status.task_status_id, priority.task_priority_id,
        context.receipt_inventory_id, destination.balance_id,
        context.owner_id, context.warehouse_id, context.item_id,
        context.lot_id, context.handling_unit_id,
        destination.location_id, target.location_id,
        context.inspected_qty, 0, context.base_uom_id, $7
    FROM context
    CROSS JOIN available_balance destination
    CROSS JOIN created_movement
    JOIN task_type ON task_type.code = 'PUTAWAY' AND task_type.is_active
    JOIN task_status ON task_status.code = 'OPEN' AND task_status.is_active
    JOIN task_priority priority ON priority.code = $6 AND priority.is_active
    JOIN warehouse_location target
      ON target.location_id = $5
     AND target.warehouse_id = context.warehouse_id
     AND target.is_active
    JOIN location_type target_type
      ON target_type.location_type_id = target.location_type_id
     AND target_type.allows_storage
    RETURNING *
),
processed_disposition AS (
    UPDATE quarantine_disposition disposition
    SET status_id = processed_status.status_id,
        processed_at = clock_timestamp(),
        inventory_movement_id = movement.movement_id,
        resulting_balance_id = task.source_balance_id
    FROM document_status processed_status
    CROSS JOIN created_movement movement
    CROSS JOIN created_task task
    CROSS JOIN context
    WHERE disposition.quarantine_disposition_id = context.quarantine_disposition_id
      AND processed_status.document_type_id = disposition.document_type_id
      AND processed_status.code = 'PROCESSED'
    RETURNING disposition.*
),
updated_case AS (
    UPDATE quarantine_case qc
    SET status_id = target_status.status_id,
        closed_at = CASE WHEN target_status.code = 'CLOSED' THEN clock_timestamp() ELSE NULL END,
        updated_at = clock_timestamp(),
        updated_by = $7,
        version_no = qc.version_no + 1
    FROM context
    JOIN document_status target_status
      ON target_status.document_type_id = context.case_document_type_id
     AND target_status.code = CASE
         WHEN COALESCE((
             SELECT sum(existing.disposition_qty)
             FROM quarantine_disposition existing
             WHERE existing.quarantine_case_id = context.quarantine_case_id
               AND existing.processed_at IS NOT NULL
               AND existing.quarantine_disposition_id <> context.quarantine_disposition_id
         ), 0) + context.disposition_qty >= context.quarantine_qty
         THEN 'CLOSED' ELSE 'PARTIALLY_DECIDED' END
    WHERE qc.quarantine_case_id = context.quarantine_case_id
      AND EXISTS (SELECT 1 FROM processed_disposition)
    RETURNING qc.*
)
SELECT created_task.*, updated_case.status_id AS quarantine_case_status_id
FROM created_task
CROSS JOIN updated_case;

COMMIT;

-- 4.6 Reinspection FAIL: stock remains quarantined, the REWORK decision is
-- completed, and a child quarantine case is opened for a new client decision.
-- $1 reinspection ID, $2 child quarantine-case ID, $3 actor account ID
BEGIN;

SELECT qc.quarantine_case_id
FROM quality_inspection inspection
JOIN rework_task task ON task.reinspection_id = inspection.inspection_id
JOIN quarantine_disposition disposition
  ON disposition.quarantine_disposition_id = task.quarantine_disposition_id
JOIN quarantine_case qc ON qc.quarantine_case_id = disposition.quarantine_case_id
WHERE inspection.inspection_id = $1
FOR UPDATE OF qc;

WITH context AS (
    SELECT
        inspection.*,
        disposition.quarantine_disposition_id,
        disposition.quarantine_case_id,
        disposition.disposition_qty,
        qc.document_type_id AS case_document_type_id,
        qc.quarantine_qty,
        qc.quarantine_balance_id,
        qc.owner_id,
        qc.warehouse_id,
        qc.receipt_inventory_id,
        qc.uom_id
    FROM quality_inspection inspection
    JOIN rework_task task ON task.reinspection_id = inspection.inspection_id
    JOIN quarantine_disposition disposition
      ON disposition.quarantine_disposition_id = task.quarantine_disposition_id
    JOIN quarantine_case qc ON qc.quarantine_case_id = disposition.quarantine_case_id
    WHERE inspection.inspection_id = $1
      AND inspection.inspection_result_id IS NULL
    FOR UPDATE OF qc
),
completed_inspection AS (
    UPDATE quality_inspection inspection
    SET quality_status_id = quality_status.quality_status_id,
        inspection_result_id = result.inspection_result_id,
        passed_qty = 0,
        failed_qty = inspection.inspected_qty,
        inspected_at = clock_timestamp(),
        inspected_by = $3
    FROM quality_status
    CROSS JOIN inspection_result result
    WHERE inspection.inspection_id = $1
      AND quality_status.code = 'FAILED'
      AND result.code = 'REJECTED'
    RETURNING inspection.*
),
processed_disposition AS (
    UPDATE quarantine_disposition disposition
    SET status_id = processed_status.status_id,
        processed_at = clock_timestamp()
    FROM document_status processed_status
    CROSS JOIN context
    CROSS JOIN completed_inspection
    WHERE disposition.quarantine_disposition_id = context.quarantine_disposition_id
      AND processed_status.document_type_id = disposition.document_type_id
      AND processed_status.code = 'PROCESSED'
    RETURNING disposition.*
),
child_case AS (
    INSERT INTO quarantine_case (
        quarantine_case_id,
        parent_quarantine_case_id,
        document_type_id,
        status_id,
        receipt_inventory_id,
        inspection_id,
        quarantine_balance_id,
        owner_id,
        warehouse_id,
        quarantine_qty,
        uom_id,
        created_by,
        updated_by
    )
    SELECT
        $2,
        context.quarantine_case_id,
        context.case_document_type_id,
        initial_status.status_id,
        context.receipt_inventory_id,
        context.inspection_id,
        context.quarantine_balance_id,
        context.owner_id,
        context.warehouse_id,
        context.inspected_qty,
        context.uom_id,
        $3,
        $3
    FROM context
    CROSS JOIN processed_disposition
    JOIN document_status initial_status
      ON initial_status.document_type_id = context.case_document_type_id
     AND initial_status.is_initial
     AND initial_status.is_active
    RETURNING *
),
updated_parent_case AS (
    UPDATE quarantine_case parent
    SET status_id = target_status.status_id,
        closed_at = CASE WHEN target_status.code = 'CLOSED' THEN clock_timestamp() ELSE NULL END,
        updated_at = clock_timestamp(),
        updated_by = $3,
        version_no = parent.version_no + 1
    FROM context
    JOIN document_status target_status
      ON target_status.document_type_id = context.case_document_type_id
     AND target_status.code = CASE
         WHEN COALESCE((
             SELECT sum(existing.disposition_qty)
             FROM quarantine_disposition existing
             WHERE existing.quarantine_case_id = context.quarantine_case_id
               AND existing.processed_at IS NOT NULL
               AND existing.quarantine_disposition_id <> context.quarantine_disposition_id
         ), 0) + context.disposition_qty >= context.quarantine_qty
         THEN 'CLOSED' ELSE 'PARTIALLY_DECIDED' END
    WHERE parent.quarantine_case_id = context.quarantine_case_id
      AND EXISTS (SELECT 1 FROM child_case)
    RETURNING parent.*
)
SELECT * FROM child_case;

COMMIT;

-- =============================================================================
-- 5. PUTAWAY COMPLETION: ACCEPTED STOCK BECOMES STORED
-- =============================================================================

-- 5.0 Putaway queue.
-- $1 owner ID, $2 warehouse ID or null, $3 assignee ID or null,
-- $4 status code or null, $5 limit, $6 offset
SELECT
    task.putaway_task_id,
    task.receipt_inventory_id,
    item.code AS item_code,
    item.name AS item_name,
    lot.lot_number,
    hu.barcode AS handling_unit_barcode,
    source.code AS source_location_code,
    target.code AS target_location_code,
    task.planned_qty,
    unit.code AS uom_code,
    priority.code AS priority_code,
    status.code AS status_code,
    task.assigned_to,
    count(*) OVER () AS total_rows
FROM putaway_task task
JOIN item ON item.item_id = task.item_id
LEFT JOIN inventory_lot lot ON lot.lot_id = task.lot_id
LEFT JOIN handling_unit hu ON hu.handling_unit_id = task.handling_unit_id
JOIN warehouse_location source ON source.location_id = task.source_location_id
JOIN warehouse_location target ON target.location_id = task.target_location_id
JOIN uom unit ON unit.uom_id = task.uom_id
JOIN task_priority priority ON priority.task_priority_id = task.task_priority_id
JOIN task_status status ON status.task_status_id = task.task_status_id
WHERE task.owner_id = $1
  AND ($2 IS NULL OR task.warehouse_id = $2)
  AND ($3 IS NULL OR task.assigned_to = $3)
  AND ($4 IS NULL OR status.code = $4)
ORDER BY priority.priority_value DESC, task.created_at, task.putaway_task_id
LIMIT $5 OFFSET $6;

-- 5.1 Assign putaway task. $1 task ID, $2 assignee account ID
UPDATE putaway_task task
SET task_status_id = assigned_status.task_status_id,
    assigned_to = $2
FROM task_status assigned_status
WHERE task.putaway_task_id = $1
  AND assigned_status.code = 'ASSIGNED'
  AND EXISTS (
      SELECT 1
      FROM task_status_transition transition
      WHERE transition.from_status_id = task.task_status_id
        AND transition.to_status_id = assigned_status.task_status_id
        AND transition.is_active
  )
RETURNING task.*;

-- 5.2 Start putaway task. $1 task ID, $2 assignee account ID
UPDATE putaway_task task
SET task_status_id = in_progress_status.task_status_id,
    assigned_to = $2,
    started_at = clock_timestamp()
FROM task_status in_progress_status
WHERE task.putaway_task_id = $1
  AND task.assigned_to = $2
  AND in_progress_status.code = 'IN_PROGRESS'
  AND EXISTS (
      SELECT 1
      FROM task_status_transition transition
      WHERE transition.from_status_id = task.task_status_id
        AND transition.to_status_id = in_progress_status.task_status_id
        AND transition.is_active
  )
RETURNING task.*;

-- 5.3 Complete a whole putaway task atomically.
-- $1 putaway-task ID, $2 proposed destination balance ID,
-- $3 movement ID, $4 actor account ID
WITH task_context AS (
    SELECT task.*, receipt.business_date
    FROM putaway_task task
    JOIN task_status status ON status.task_status_id = task.task_status_id
    JOIN receipt_inventory batch ON batch.receipt_inventory_id = task.receipt_inventory_id
    JOIN receipt_line line ON line.receipt_line_id = batch.receipt_line_id
    JOIN receipt ON receipt.receipt_id = line.receipt_id
    WHERE task.putaway_task_id = $1
      AND status.code = 'IN_PROGRESS'
      AND task.source_location_id <> task.target_location_id
    FOR UPDATE OF task
),
decremented_source AS (
    UPDATE inventory_balance source
    SET on_hand_qty = source.on_hand_qty - task.planned_qty,
        version_no = source.version_no + 1,
        updated_at = clock_timestamp()
    FROM task_context task
    WHERE source.balance_id = task.source_balance_id
      AND source.location_id = task.source_location_id
      AND source.on_hand_qty - source.reserved_qty >= task.planned_qty
    RETURNING source.*
),
destination_balance AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT
        $2,
        source.owner_id,
        source.warehouse_id,
        task.target_location_id,
        source.item_id,
        source.lot_id,
        source.handling_unit_id,
        source.inventory_status_id,
        task.planned_qty,
        0,
        source.uom_id
    FROM decremented_source source
    CROSS JOIN task_context task
    ON CONFLICT (
        owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id
    ) DO UPDATE SET
        on_hand_qty = inventory_balance.on_hand_qty + EXCLUDED.on_hand_qty,
        version_no = inventory_balance.version_no + 1,
        updated_at = clock_timestamp()
    RETURNING *
),
created_movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, handling_unit_id,
        from_location_id, to_location_id, from_status_id, to_status_id,
        quantity, uom_id, source_document_id, source_line_id, created_by
    )
    SELECT
        $3,
        movement_type.movement_type_id,
        source.owner_id,
        source.warehouse_id,
        task.business_date,
        source.item_id,
        source.lot_id,
        source.handling_unit_id,
        source.location_id,
        destination.location_id,
        source.inventory_status_id,
        destination.inventory_status_id,
        task.planned_qty,
        source.uom_id,
        task.putaway_task_id,
        task.receipt_inventory_id,
        $4
    FROM decremented_source source
    CROSS JOIN destination_balance destination
    CROSS JOIN task_context task
    JOIN movement_type ON movement_type.code = 'PUTAWAY' AND movement_type.is_active
    RETURNING movement_id
),
completed_task AS (
    UPDATE putaway_task task
    SET task_status_id = completed_status.task_status_id,
        completed_qty = task.planned_qty,
        completed_at = clock_timestamp(),
        assigned_to = COALESCE(task.assigned_to, $4)
    FROM task_status completed_status
    WHERE task.putaway_task_id = $1
      AND completed_status.code = 'COMPLETED'
      AND EXISTS (SELECT 1 FROM created_movement)
      AND EXISTS (
          SELECT 1
          FROM task_status_transition transition
          WHERE transition.from_status_id = task.task_status_id
            AND transition.to_status_id = completed_status.task_status_id
            AND transition.is_active
      )
    RETURNING task.*
),
updated_handling_unit AS (
    UPDATE handling_unit hu
    SET current_location_id = destination.location_id
    FROM destination_balance destination
    CROSS JOIN completed_task task
    WHERE hu.handling_unit_id = task.handling_unit_id
    RETURNING hu.handling_unit_id
)
SELECT
    completed_task.*,
    destination_balance.balance_id AS stored_balance_id,
    created_movement.movement_id
FROM completed_task
CROSS JOIN destination_balance
CROSS JOIN created_movement;

-- 5.4 Stored inventory trace for one receipt batch. $1 receipt-inventory ID
SELECT
    batch.receipt_inventory_id,
    receipt.receipt_id,
    inbound.inbound_id,
    po.purchase_order_id,
    po.purchase_order_no,
    item.code AS item_code,
    item.name AS item_name,
    lot.lot_number,
    hu.barcode AS handling_unit_barcode,
    balance.balance_id,
    warehouse.code AS warehouse_code,
    location.code AS storage_location_code,
    inventory_status.code AS inventory_status_code,
    balance.on_hand_qty,
    unit.code AS base_uom_code
FROM receipt_inventory batch
JOIN receipt_line line ON line.receipt_line_id = batch.receipt_line_id
JOIN receipt ON receipt.receipt_id = line.receipt_id
JOIN inbound_order inbound ON inbound.inbound_id = receipt.inbound_id
LEFT JOIN inbound_order_line inbound_line ON inbound_line.inbound_line_id = line.inbound_line_id
LEFT JOIN purchase_order_line po_line
  ON po_line.purchase_order_line_id = inbound_line.purchase_order_line_id
LEFT JOIN purchase_order po ON po.purchase_order_id = po_line.purchase_order_id
JOIN item ON item.item_id = batch.item_id
LEFT JOIN inventory_lot lot ON lot.lot_id = batch.lot_id
LEFT JOIN handling_unit hu ON hu.handling_unit_id = batch.handling_unit_id
JOIN inventory_balance balance
  ON balance.owner_id = receipt.owner_id
 AND balance.warehouse_id = receipt.warehouse_id
 AND balance.item_id = batch.item_id
 AND balance.lot_id IS NOT DISTINCT FROM batch.lot_id
 AND balance.handling_unit_id IS NOT DISTINCT FROM batch.handling_unit_id
 AND balance.on_hand_qty > 0
JOIN warehouse ON warehouse.warehouse_id = balance.warehouse_id
JOIN warehouse_location location ON location.location_id = balance.location_id
JOIN location_type location_type
  ON location_type.location_type_id = location.location_type_id
 AND location_type.allows_storage
JOIN inventory_status ON inventory_status.inventory_status_id = balance.inventory_status_id
JOIN uom unit ON unit.uom_id = balance.uom_id
WHERE batch.receipt_inventory_id = $1
ORDER BY location.code;
