-- Picking, staging, checking, and packing operations.
-- Execute one numbered operation at a time using prepared parameters.

SET search_path TO wms, public;

-- =============================================================================
-- 1. PICKING
-- =============================================================================

-- 1.1 Open pick-task queue.
-- $1 warehouse ID, $2 assignee ID or NULL, $3 limit, $4 offset
SELECT task.pick_task_id, task.wave_id,
       outbound.client_delivery_order_no, line.line_no,
       item.code AS item_code, lot.lot_number,
       source.code AS source_location_code,
       target.code AS staging_location_code,
       task.planned_qty, task.picked_qty, task.short_qty,
       unit.code AS uom_code, priority.code AS priority_code,
       status.code AS task_status_code, task.assigned_to,
       count(*) OVER () AS total_rows
FROM pick_task task
JOIN task_status status ON status.task_status_id = task.task_status_id
JOIN task_priority priority ON priority.task_priority_id = task.task_priority_id
JOIN inventory_reservation reservation ON reservation.reservation_id = task.reservation_id
JOIN inventory_balance balance ON balance.balance_id = reservation.balance_id
JOIN outbound_order_line line ON line.outbound_line_id = task.outbound_line_id
JOIN outbound_order outbound ON outbound.outbound_id = line.outbound_id
JOIN item ON item.item_id = line.item_id
LEFT JOIN inventory_lot lot ON lot.lot_id = balance.lot_id
JOIN warehouse_location source ON source.location_id = task.source_location_id
LEFT JOIN warehouse_location target ON target.location_id = task.target_location_id
JOIN uom unit ON unit.uom_id = task.uom_id
WHERE outbound.warehouse_id = $1
  AND ($2::uuid IS NULL OR task.assigned_to = $2 OR task.assigned_to IS NULL)
  AND NOT status.is_final
ORDER BY priority.priority_value DESC, outbound.requested_ship_at NULLS LAST,
         task.created_at, task.pick_task_id
LIMIT $3 OFFSET $4;

-- 1.2 Assign/start a pick task. $1 task ID, $2 actor ID
UPDATE pick_task task
SET task_status_id = in_progress.task_status_id,
    assigned_to = COALESCE(task.assigned_to, $2),
    started_at = COALESCE(task.started_at, clock_timestamp())
FROM task_status current_status, task_status in_progress
WHERE task.pick_task_id = $1
  AND current_status.task_status_id = task.task_status_id
  AND current_status.code IN ('OPEN', 'ASSIGNED')
  AND in_progress.code = 'IN_PROGRESS' AND in_progress.is_active
  AND (task.assigned_to IS NULL OR task.assigned_to = $2)
RETURNING task.*;

-- 1.3 Confirm a partial or complete pick into staging atomically.
-- $1 pick-task ID, $2 picked base quantity,
-- $3 proposed staging balance ID, $4 movement ID,
-- $5 pick-execution ID, $6 actor ID
BEGIN;

SELECT balance.balance_id
FROM pick_task task
JOIN inventory_reservation reservation ON reservation.reservation_id = task.reservation_id
JOIN inventory_balance balance ON balance.balance_id = reservation.balance_id
WHERE task.pick_task_id = $1
FOR UPDATE OF balance;

WITH context AS (
    SELECT task.*, reservation.document_type_id AS reservation_document_type_id,
           reservation.status_id AS reservation_status_id,
           reservation.balance_id, reservation.reserved_qty,
           reservation.picked_qty AS reservation_picked_qty,
           line.outbound_id, line.item_id, outbound.owner_id,
           outbound.warehouse_id, outbound.business_date,
           balance.location_id AS actual_source_location_id,
           balance.lot_id, balance.handling_unit_id,
           balance.inventory_status_id, balance.on_hand_qty,
           balance.reserved_qty AS balance_reserved_qty
    FROM pick_task task
    JOIN task_status task_status ON task_status.task_status_id = task.task_status_id
    JOIN inventory_reservation reservation ON reservation.reservation_id = task.reservation_id
    JOIN document_status reservation_status ON reservation_status.status_id = reservation.status_id
    JOIN inventory_balance balance ON balance.balance_id = reservation.balance_id
    JOIN outbound_order_line line ON line.outbound_line_id = task.outbound_line_id
    JOIN outbound_order outbound ON outbound.outbound_id = line.outbound_id
    JOIN warehouse_location staging
      ON staging.location_id = task.target_location_id
     AND staging.warehouse_id = outbound.warehouse_id AND staging.is_active
    WHERE task.pick_task_id = $1
      AND task_status.code = 'IN_PROGRESS'
      AND reservation_status.code IN ('ACTIVE', 'PARTIALLY_PICKED')
      AND task.picked_qty + $2 <= task.planned_qty
      AND reservation.picked_qty + $2 <= reservation.reserved_qty
      AND balance.on_hand_qty >= $2 AND balance.reserved_qty >= $2
      AND task.source_location_id = balance.location_id
      AND (
          balance.handling_unit_id IS NULL
          OR ($2 = balance.on_hand_qty AND balance.reserved_qty = $2)
      )
      AND $2 > 0
), source_update AS (
    UPDATE inventory_balance balance
    SET on_hand_qty = balance.on_hand_qty - $2,
        reserved_qty = balance.reserved_qty - $2,
        version_no = balance.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE balance.balance_id = context.balance_id
      AND balance.on_hand_qty >= $2 AND balance.reserved_qty >= $2
    RETURNING balance.*
), staging_balance AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT $3, context.owner_id, context.warehouse_id,
           context.target_location_id, context.item_id, context.lot_id,
           context.handling_unit_id, context.inventory_status_id,
           $2, 0, context.uom_id
    FROM context CROSS JOIN source_update
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
    SELECT $4, movement_type.movement_type_id, context.owner_id,
           context.warehouse_id, context.business_date, context.item_id,
           context.lot_id, context.handling_unit_id,
           context.actual_source_location_id, context.target_location_id,
           context.inventory_status_id, context.inventory_status_id,
           $2, context.uom_id, context.pick_task_id,
           context.outbound_line_id, $6
    FROM context CROSS JOIN staging_balance
    JOIN movement_type ON movement_type.code = 'PICK' AND movement_type.is_active
    RETURNING movement_id
), execution AS (
    INSERT INTO pick_execution (
        pick_execution_id, pick_task_id, source_balance_id,
        staging_balance_id, picked_qty, uom_id, movement_id, picked_by
    )
    SELECT $5, context.pick_task_id, context.balance_id,
           staging_balance.balance_id, $2, context.uom_id,
           movement.movement_id, $6
    FROM context CROSS JOIN staging_balance CROSS JOIN movement
    RETURNING *
), updated_task AS (
    UPDATE pick_task task
    SET picked_qty = task.picked_qty + $2,
        task_status_id = next_status.task_status_id,
        completed_at = CASE WHEN task.picked_qty + $2 = task.planned_qty
                            THEN clock_timestamp() ELSE task.completed_at END
    FROM execution, task_status next_status
    WHERE task.pick_task_id = execution.pick_task_id
      AND next_status.code = CASE WHEN task.picked_qty + $2 = task.planned_qty
                                  THEN 'COMPLETED' ELSE 'IN_PROGRESS' END
    RETURNING task.*
), updated_reservation AS (
    UPDATE inventory_reservation reservation
    SET picked_qty = reservation.picked_qty + $2,
        status_id = next_status.status_id
    FROM context, document_status next_status
    WHERE reservation.reservation_id = context.reservation_id
      AND next_status.document_type_id = reservation.document_type_id
      AND next_status.code = CASE
          WHEN reservation.picked_qty + $2 = reservation.reserved_qty
          THEN 'CONSUMED' ELSE 'PARTIALLY_PICKED' END
    RETURNING reservation.*
), updated_line AS (
    UPDATE outbound_order_line line
    SET picked_qty = line.picked_qty + $2
    FROM context
    WHERE line.outbound_line_id = context.outbound_line_id
    RETURNING line.*
), updated_outbound AS (
    UPDATE outbound_order outbound
    SET status_id = CASE WHEN current_status.code = 'CHECK_FAILED'
                         THEN outbound.status_id ELSE picking.status_id END,
        updated_at = clock_timestamp(), updated_by = $6,
        version_no = outbound.version_no + 1
    FROM context, document_status current_status, document_status picking
    WHERE outbound.outbound_id = context.outbound_id
      AND current_status.status_id = outbound.status_id
      AND picking.document_type_id = outbound.document_type_id
      AND picking.code = 'PICKING'
    RETURNING outbound.outbound_id
), updated_wave AS (
    UPDATE outbound_wave wave
    SET status_id = next_status.status_id,
        completed_at = CASE WHEN next_status.code = 'COMPLETED'
                            THEN clock_timestamp() ELSE wave.completed_at END,
        updated_at = clock_timestamp(), updated_by = $6,
        version_no = wave.version_no + 1
    FROM updated_task current_task, document_status next_status
    WHERE wave.wave_id = current_task.wave_id
      AND next_status.document_type_id = wave.document_type_id
      AND next_status.code = CASE
          WHEN current_task.picked_qty + current_task.short_qty = current_task.planned_qty
           AND EXISTS (
               SELECT 1 FROM task_status current_task_status
               WHERE current_task_status.task_status_id = current_task.task_status_id
                 AND current_task_status.is_final
           )
           AND NOT EXISTS (
               SELECT 1 FROM pick_task other
               JOIN task_status other_status
                 ON other_status.task_status_id = other.task_status_id
               WHERE other.wave_id = current_task.wave_id
                 AND other.pick_task_id <> current_task.pick_task_id
                 AND NOT other_status.is_final
           ) THEN 'COMPLETED' ELSE 'IN_PROGRESS' END
    RETURNING wave.wave_id
), updated_hu AS (
    UPDATE handling_unit hu
    SET current_location_id = context.target_location_id
    FROM context, execution
    WHERE hu.handling_unit_id = context.handling_unit_id
    RETURNING hu.handling_unit_id
)
SELECT execution.*, updated_task.picked_qty AS task_picked_qty,
       updated_task.task_status_id
FROM execution CROSS JOIN updated_task CROSS JOIN updated_reservation
     CROSS JOIN updated_line CROSS JOIN updated_outbound CROSS JOIN updated_wave;

COMMIT;

-- 1.4 Close a short pick and release its unpicked reservation.
-- $1 pick-task ID, $2 OUTBOUND reason code, $3 result notes, $4 actor ID
BEGIN;

SELECT balance.balance_id
FROM pick_task task
JOIN inventory_reservation reservation ON reservation.reservation_id = task.reservation_id
JOIN inventory_balance balance ON balance.balance_id = reservation.balance_id
WHERE task.pick_task_id = $1
FOR UPDATE OF balance;

WITH context AS (
    SELECT task.*, reservation.balance_id, reservation.document_type_id,
           reservation.reserved_qty, reservation.picked_qty AS reservation_picked_qty,
           task.planned_qty - task.picked_qty AS short_qty_to_close,
           line.outbound_id, reason.reason_code_id
    FROM pick_task task
    JOIN task_status status ON status.task_status_id = task.task_status_id
    JOIN inventory_reservation reservation ON reservation.reservation_id = task.reservation_id
    JOIN outbound_order_line line ON line.outbound_line_id = task.outbound_line_id
    JOIN reason_code reason
      ON reason.module_code = 'OUTBOUND' AND reason.code = $2 AND reason.is_active
    WHERE task.pick_task_id = $1
      AND status.code = 'IN_PROGRESS'
      AND task.picked_qty < task.planned_qty
), released_balance AS (
    UPDATE inventory_balance balance
    SET reserved_qty = balance.reserved_qty - context.short_qty_to_close,
        version_no = balance.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE balance.balance_id = context.balance_id
      AND balance.reserved_qty >= context.short_qty_to_close
    RETURNING balance.balance_id
), closed_task AS (
    UPDATE pick_task task
    SET short_qty = context.short_qty_to_close,
        short_reason_code_id = context.reason_code_id,
        result_notes = $3, completed_at = clock_timestamp(),
        task_status_id = completed.task_status_id
    FROM context, released_balance, task_status completed
    WHERE task.pick_task_id = context.pick_task_id
      AND completed.code = 'COMPLETED'
    RETURNING task.*
), released_reservation AS (
    UPDATE inventory_reservation reservation
    SET status_id = released.status_id, released_at = clock_timestamp()
    FROM context, closed_task, document_status released
    WHERE reservation.reservation_id = context.reservation_id
      AND released.document_type_id = reservation.document_type_id
      AND released.code = 'RELEASED'
    RETURNING reservation.*
), reduced_allocation AS (
    UPDATE outbound_order_line line
    SET allocated_qty = line.allocated_qty - context.short_qty_to_close
    FROM context, released_reservation
    WHERE line.outbound_line_id = context.outbound_line_id
    RETURNING line.*
), updated_wave AS (
    UPDATE outbound_wave wave
    SET status_id = next_status.status_id,
        completed_at = CASE WHEN next_status.code = 'COMPLETED'
                            THEN clock_timestamp() ELSE wave.completed_at END,
        updated_at = clock_timestamp(), updated_by = $4,
        version_no = wave.version_no + 1
    FROM closed_task current_task, document_status next_status
    WHERE wave.wave_id = current_task.wave_id
      AND next_status.document_type_id = wave.document_type_id
      AND next_status.code = CASE WHEN NOT EXISTS (
          SELECT 1 FROM pick_task other
          JOIN task_status other_status
            ON other_status.task_status_id = other.task_status_id
          WHERE other.wave_id = current_task.wave_id
            AND other.pick_task_id <> current_task.pick_task_id
            AND NOT other_status.is_final
      ) THEN 'COMPLETED' ELSE 'IN_PROGRESS' END
    RETURNING wave.wave_id
)
SELECT closed_task.*, context.short_qty_to_close
FROM closed_task CROSS JOIN context CROSS JOIN reduced_allocation CROSS JOIN updated_wave;

COMMIT;

-- =============================================================================
-- 2. STAGING
-- =============================================================================

-- 2.1 Create an order staging document after all its pick tasks finish.
-- $1 outbound ID, $2 wave ID, $3 notes, $4 actor ID
WITH context AS (
    SELECT outbound.outbound_id, outbound.warehouse_id,
           (array_agg(DISTINCT target.location_id))[1] AS staging_location_id,
           warehouse.code AS warehouse_code, wave.wave_id,
           dt.document_type_id, initial_status.status_id,
           outbound.business_date
    FROM outbound_order outbound
    JOIN outbound_wave_order link ON link.outbound_id = outbound.outbound_id
    JOIN outbound_wave wave ON wave.wave_id = link.wave_id AND wave.wave_id = $2
    JOIN warehouse ON warehouse.warehouse_id = outbound.warehouse_id
    JOIN pick_task task ON task.wave_id = wave.wave_id
    JOIN outbound_order_line line
      ON line.outbound_line_id = task.outbound_line_id AND line.outbound_id = outbound.outbound_id
    JOIN warehouse_location target ON target.location_id = task.target_location_id
    JOIN document_type dt ON dt.code = 'OUTBOUND_STAGING' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE outbound.outbound_id = $1
      AND NOT EXISTS (
          SELECT 1 FROM pick_task remaining
          JOIN outbound_order_line remaining_line
            ON remaining_line.outbound_line_id = remaining.outbound_line_id
          JOIN task_status remaining_status
            ON remaining_status.task_status_id = remaining.task_status_id
          WHERE remaining.wave_id = wave.wave_id
            AND remaining_line.outbound_id = outbound.outbound_id
            AND NOT remaining_status.is_final
      )
    GROUP BY outbound.outbound_id, outbound.warehouse_id,
             warehouse.code, wave.wave_id, dt.document_type_id,
             initial_status.status_id, outbound.business_date
    HAVING count(DISTINCT target.location_id) = 1
), numbered AS (
    SELECT context.*,
           generate_document_id('OUTBOUND_STAGING', NULL, warehouse_code, business_date) AS generated_id
    FROM context
), created_staging AS (
    INSERT INTO outbound_staging (
        staging_id, document_type_id, status_id, outbound_id, wave_id,
        warehouse_id, staging_location_id, notes, created_by
    )
    SELECT generated_id, document_type_id, status_id, outbound_id, wave_id,
           warehouse_id, staging_location_id, $3, $4
    FROM numbered
    RETURNING *
), created_lines AS (
    INSERT INTO outbound_staging_line (
        staging_line_id, staging_id, pick_execution_id,
        staging_balance_id, staged_qty, uom_id, created_by
    )
    SELECT staging.staging_id || '-L-' || lpad(row_number() OVER (
               ORDER BY execution.pick_execution_id)::text, 6, '0'),
           staging.staging_id, execution.pick_execution_id,
           execution.staging_balance_id, execution.picked_qty,
           execution.uom_id, $4
    FROM created_staging staging
    JOIN outbound_order_line line ON line.outbound_id = staging.outbound_id
    JOIN pick_task task ON task.outbound_line_id = line.outbound_line_id
                       AND task.wave_id = staging.wave_id
    JOIN pick_execution execution ON execution.pick_task_id = task.pick_task_id
    RETURNING *
)
SELECT created_staging.*, (SELECT count(*) FROM created_lines) AS staging_lines
FROM created_staging;

-- 2.2 Confirm staging and advance the DO. $1 staging ID, $2 actor ID
WITH completed_staging AS (
    UPDATE outbound_staging staging
    SET status_id = completed.status_id,
        staged_at = clock_timestamp(), staged_by = $2
    FROM document_status current_status, document_status completed
    WHERE staging.staging_id = $1
      AND current_status.status_id = staging.status_id
      AND current_status.code = 'OPEN'
      AND completed.document_type_id = staging.document_type_id
      AND completed.code = 'COMPLETED'
      AND EXISTS (SELECT 1 FROM outbound_staging_line line WHERE line.staging_id = staging.staging_id)
    RETURNING staging.*
), updated_outbound AS (
    UPDATE outbound_order outbound
    SET status_id = staged.status_id,
        updated_at = clock_timestamp(), updated_by = $2,
        version_no = outbound.version_no + 1
    FROM completed_staging staging, document_status staged
    WHERE outbound.outbound_id = staging.outbound_id
      AND staged.document_type_id = outbound.document_type_id
      AND staged.code = 'STAGED'
    RETURNING outbound.outbound_id
)
SELECT completed_staging.* FROM completed_staging CROSS JOIN updated_outbound;

-- =============================================================================
-- 3. CHECKING
-- =============================================================================

-- 3.1 Create a check/recheck and copy staging lines.
-- $1 staging ID, $2 parent failed-check ID or NULL, $3 notes, $4 actor ID
WITH context AS (
    SELECT staging.*, warehouse.code AS warehouse_code,
           outbound.business_date,
           dt.document_type_id AS check_document_type_id,
           initial_status.status_id AS check_status_id
    FROM outbound_staging staging
    JOIN document_status staging_status ON staging_status.status_id = staging.status_id
    JOIN outbound_order outbound ON outbound.outbound_id = staging.outbound_id
    JOIN warehouse ON warehouse.warehouse_id = staging.warehouse_id
    JOIN document_type dt ON dt.code = 'OUTBOUND_CHECK' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    LEFT JOIN outbound_check parent
      ON parent.outbound_check_id = $2 AND parent.staging_id = staging.staging_id
    LEFT JOIN document_status parent_status ON parent_status.status_id = parent.status_id
    WHERE staging.staging_id = $1 AND staging_status.code = 'COMPLETED'
      AND ($2::varchar IS NULL OR parent_status.code = 'FAILED')
      AND ($2::varchar IS NULL OR NOT EXISTS (
          SELECT 1
          FROM outbound_check_line parent_line
          JOIN outbound_check_exception exception
            ON exception.outbound_check_line_id = parent_line.outbound_check_line_id
          JOIN outbound_check_exception_status exception_status
            ON exception_status.outbound_check_exception_status_id = exception.status_id
          WHERE parent_line.outbound_check_id = $2
            AND NOT exception_status.is_final
      ))
), numbered AS (
    SELECT context.*,
           generate_document_id('OUTBOUND_CHECK', NULL, warehouse_code, business_date) AS generated_id
    FROM context
), created_check AS (
    INSERT INTO outbound_check (
        outbound_check_id, parent_check_id, document_type_id, status_id,
        staging_id, notes, created_by
    )
    SELECT generated_id, $2, check_document_type_id, check_status_id,
           staging_id, $3, $4
    FROM numbered
    RETURNING *
), created_lines AS (
    INSERT INTO outbound_check_line (
        outbound_check_line_id, outbound_check_id, staging_line_id,
        line_no, expected_qty, uom_id, created_by
    )
    SELECT check_header.outbound_check_id || '-L-' || lpad(row_number() OVER (
               ORDER BY staging_line.staging_line_id)::text, 6, '0'),
           check_header.outbound_check_id, staging_line.staging_line_id,
           row_number() OVER (ORDER BY staging_line.staging_line_id),
           staging_line.staged_qty - staging_line.removed_qty,
           staging_line.uom_id, $4
    FROM created_check check_header
    JOIN outbound_staging_line staging_line ON staging_line.staging_id = check_header.staging_id
    WHERE staging_line.staged_qty > staging_line.removed_qty
    RETURNING *
), checking_outbound AS (
    UPDATE outbound_order outbound
    SET status_id = checking.status_id,
        updated_at = clock_timestamp(), updated_by = $4,
        version_no = outbound.version_no + 1
    FROM context, document_status checking
    WHERE outbound.outbound_id = context.outbound_id
      AND checking.document_type_id = outbound.document_type_id
      AND checking.code = 'CHECKING'
    RETURNING outbound.outbound_id
)
SELECT created_check.*, (SELECT count(*) FROM created_lines) AS check_lines
FROM created_check CROSS JOIN checking_outbound;

-- 3.2 Record one check result.
-- $1 check-line ID, $2 checked quantity, $3 result code,
-- $4 notes or NULL, $5 affected exception quantity or NULL, $6 actor ID
UPDATE outbound_check_line line
SET checked_qty = $2,
    exception_qty = CASE result.code
        WHEN 'PASS' THEN 0
        WHEN 'SHORT' THEN line.expected_qty - $2
        WHEN 'OVER' THEN $2 - line.expected_qty
        ELSE $5
    END,
    outbound_check_result_id = result.outbound_check_result_id,
    notes = $4, checked_at = clock_timestamp(), checked_by = $6
FROM outbound_check check_header, document_status check_status,
     outbound_check_result result
WHERE line.outbound_check_line_id = $1
  AND check_header.outbound_check_id = line.outbound_check_id
  AND check_status.status_id = check_header.status_id AND check_status.code = 'OPEN'
  AND result.code = $3 AND result.is_active
  AND (NOT result.requires_note OR nullif(trim($4), '') IS NOT NULL)
  AND $2 >= 0
  AND (
      (result.code = 'PASS' AND $2 = line.expected_qty AND COALESCE($5, 0) = 0)
      OR (result.code = 'SHORT' AND $2 < line.expected_qty)
      OR (result.code = 'OVER' AND $2 > line.expected_qty)
      OR (result.code IN ('WRONG_ITEM', 'DAMAGED')
          AND $5 > 0 AND $5 <= line.expected_qty)
  )
RETURNING line.*;

-- 3.3 Finalize check as PASSED or FAILED and update DO checked quantities.
-- $1 check ID, $2 actor ID
WITH outcome AS (
    SELECT check_header.outbound_check_id, staging.outbound_id,
           bool_and(result.is_pass AND line.checked_qty = line.expected_qty
                    AND line.exception_qty = 0) AS passed
    FROM outbound_check check_header
    JOIN document_status status ON status.status_id = check_header.status_id
    JOIN outbound_staging staging ON staging.staging_id = check_header.staging_id
    JOIN outbound_check_line line ON line.outbound_check_id = check_header.outbound_check_id
    LEFT JOIN outbound_check_result result
      ON result.outbound_check_result_id = line.outbound_check_result_id
    WHERE check_header.outbound_check_id = $1 AND status.code = 'OPEN'
    GROUP BY check_header.outbound_check_id, staging.outbound_id
    HAVING count(*) = count(line.outbound_check_result_id)
       AND count(*) = count(line.checked_qty)
), completed_check AS (
    UPDATE outbound_check check_header
    SET status_id = final_status.status_id,
        checked_at = clock_timestamp(), checked_by = $2
    FROM outcome, document_status final_status
    WHERE check_header.outbound_check_id = outcome.outbound_check_id
      AND final_status.document_type_id = check_header.document_type_id
      AND final_status.code = CASE WHEN outcome.passed THEN 'PASSED' ELSE 'FAILED' END
    RETURNING check_header.*, outcome.passed, outcome.outbound_id
), created_exceptions AS (
    INSERT INTO outbound_check_exception (
        outbound_check_exception_id, outbound_check_line_id,
        status_id, exception_qty, notes, created_by
    )
    SELECT line.outbound_check_line_id || '-EX', line.outbound_check_line_id,
           open_status.outbound_check_exception_status_id,
           line.exception_qty, line.notes, $2
    FROM completed_check completed
    JOIN outbound_check_line line
      ON line.outbound_check_id = completed.outbound_check_id
    JOIN outbound_check_result result
      ON result.outbound_check_result_id = line.outbound_check_result_id
    JOIN outbound_check_exception_status open_status
      ON open_status.code = 'OPEN' AND open_status.is_active
    WHERE NOT completed.passed AND NOT result.is_pass
      AND line.exception_qty > 0
    RETURNING outbound_check_exception_id
), checked_lines AS (
    UPDATE outbound_order_line outbound_line
    SET checked_qty = outbound_line.checked_qty + aggregated.checked_qty
    FROM (
        SELECT task.outbound_line_id, sum(check_line.checked_qty) AS checked_qty
        FROM completed_check completed
        JOIN outbound_check_line check_line
          ON check_line.outbound_check_id = completed.outbound_check_id
        JOIN outbound_staging_line staging_line
          ON staging_line.staging_line_id = check_line.staging_line_id
        JOIN pick_execution execution
          ON execution.pick_execution_id = staging_line.pick_execution_id
        JOIN pick_task task ON task.pick_task_id = execution.pick_task_id
        WHERE completed.passed
        GROUP BY task.outbound_line_id
    ) aggregated
    WHERE outbound_line.outbound_line_id = aggregated.outbound_line_id
    RETURNING outbound_line.outbound_line_id
), updated_outbound AS (
    UPDATE outbound_order outbound
    SET status_id = next_status.status_id,
        updated_at = clock_timestamp(), updated_by = $2,
        version_no = outbound.version_no + 1
    FROM completed_check completed, document_status next_status
    WHERE outbound.outbound_id = completed.outbound_id
      AND next_status.document_type_id = outbound.document_type_id
      AND next_status.code = CASE WHEN completed.passed THEN 'CHECKED' ELSE 'CHECK_FAILED' END
    RETURNING outbound.outbound_id
)
SELECT completed_check.* FROM completed_check CROSS JOIN updated_outbound;

-- 3.4 Open failed-check exception queue.
-- $1 owner ID, $2 warehouse ID, $3 exception-status code or NULL
SELECT exception.outbound_check_exception_id, exception.opened_at,
       exception.exception_qty, exception_status.code AS exception_status_code,
       result.code AS check_result_code, outbound.client_delivery_order_no,
       outbound.outbound_id, line.line_no AS check_line_no,
       item.code AS item_code, location.code AS staging_location_code,
       unit.code AS uom_code, exception.notes
FROM outbound_check_exception exception
JOIN outbound_check_exception_status exception_status
  ON exception_status.outbound_check_exception_status_id = exception.status_id
JOIN outbound_check_line line
  ON line.outbound_check_line_id = exception.outbound_check_line_id
JOIN outbound_check_result result
  ON result.outbound_check_result_id = line.outbound_check_result_id
JOIN outbound_staging_line staging_line
  ON staging_line.staging_line_id = line.staging_line_id
JOIN outbound_staging staging ON staging.staging_id = staging_line.staging_id
JOIN outbound_order outbound ON outbound.outbound_id = staging.outbound_id
JOIN pick_execution execution ON execution.pick_execution_id = staging_line.pick_execution_id
JOIN pick_task task ON task.pick_task_id = execution.pick_task_id
JOIN outbound_order_line order_line ON order_line.outbound_line_id = task.outbound_line_id
JOIN item ON item.item_id = order_line.item_id
JOIN warehouse_location location ON location.location_id = staging.staging_location_id
JOIN uom unit ON unit.uom_id = line.uom_id
WHERE outbound.owner_id = $1 AND outbound.warehouse_id = $2
  AND ($3::varchar IS NULL OR exception_status.code = $3)
ORDER BY exception.opened_at, exception.outbound_check_exception_id;

-- 3.5 Correct a physical SHORT by removing phantom staging stock.
-- This is an adjustment because the missing quantity cannot be physically moved.
-- $1 exception ID, $2 quantity, $3 movement ID, $4 resolution ID,
-- $5 notes, $6 actor ID
BEGIN;

SELECT balance.balance_id
FROM outbound_check_exception exception
JOIN outbound_check_line check_line
  ON check_line.outbound_check_line_id = exception.outbound_check_line_id
JOIN outbound_staging_line staging_line ON staging_line.staging_line_id = check_line.staging_line_id
JOIN inventory_balance balance ON balance.balance_id = staging_line.staging_balance_id
WHERE exception.outbound_check_exception_id = $1
FOR UPDATE OF balance;

WITH context AS (
    SELECT exception.*, check_line.staging_line_id,
           staging_line.staging_balance_id, staging_line.staged_qty,
           staging_line.removed_qty, balance.location_id,
           balance.inventory_status_id, balance.item_id, balance.lot_id,
           balance.handling_unit_id, balance.on_hand_qty,
           task.outbound_line_id, outbound.owner_id, outbound.warehouse_id,
           outbound.business_date, check_line.uom_id,
           resolution_type.outbound_check_resolution_type_id,
           reason.reason_code_id
    FROM outbound_check_exception exception
    JOIN outbound_check_exception_status exception_status
      ON exception_status.outbound_check_exception_status_id = exception.status_id
     AND NOT exception_status.is_final
    JOIN outbound_check_line check_line
      ON check_line.outbound_check_line_id = exception.outbound_check_line_id
    JOIN outbound_check_result result
      ON result.outbound_check_result_id = check_line.outbound_check_result_id
     AND result.code = 'SHORT'
    JOIN outbound_staging_line staging_line ON staging_line.staging_line_id = check_line.staging_line_id
    JOIN inventory_balance balance ON balance.balance_id = staging_line.staging_balance_id
    JOIN pick_execution execution ON execution.pick_execution_id = staging_line.pick_execution_id
    JOIN pick_task task ON task.pick_task_id = execution.pick_task_id
    JOIN outbound_order_line order_line ON order_line.outbound_line_id = task.outbound_line_id
    JOIN outbound_order outbound ON outbound.outbound_id = order_line.outbound_id
    JOIN outbound_check_resolution_type resolution_type
      ON resolution_type.code = 'STOCK_CORRECTION' AND resolution_type.is_active
    JOIN reason_code reason
      ON reason.module_code = 'OUTBOUND' AND reason.code = 'CHECK_FAILED' AND reason.is_active
    WHERE exception.outbound_check_exception_id = $1 AND $2 > 0
      AND balance.on_hand_qty >= $2
      AND staging_line.staged_qty - staging_line.removed_qty >= $2
      AND COALESCE((
          SELECT sum(resolution.resolved_qty)
          FROM outbound_check_resolution resolution
          JOIN outbound_check_resolution_type type
            ON type.outbound_check_resolution_type_id = resolution.outbound_check_resolution_type_id
          WHERE resolution.outbound_check_exception_id = exception.outbound_check_exception_id
            AND type.counts_as_stock_correction
      ), 0) + $2 <= exception.exception_qty
), source_update AS (
    UPDATE inventory_balance balance
    SET on_hand_qty = balance.on_hand_qty - $2,
        version_no = balance.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE balance.balance_id = context.staging_balance_id
      AND balance.on_hand_qty >= $2
    RETURNING balance.balance_id
), staging_update AS (
    UPDATE outbound_staging_line staging_line
    SET removed_qty = staging_line.removed_qty + $2
    FROM context, source_update
    WHERE staging_line.staging_line_id = context.staging_line_id
      AND staging_line.staged_qty - staging_line.removed_qty >= $2
    RETURNING staging_line.staging_line_id
), line_update AS (
    UPDATE outbound_order_line line
    SET rejected_qty = line.rejected_qty + $2
    FROM context, staging_update
    WHERE line.outbound_line_id = context.outbound_line_id
    RETURNING line.outbound_line_id
), movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, handling_unit_id, from_location_id, to_location_id,
        from_status_id, to_status_id, quantity, uom_id,
        source_document_id, source_line_id, reason_code_id, notes, created_by
    )
    SELECT $3, movement_type.movement_type_id, owner_id, warehouse_id,
           business_date, item_id, lot_id, handling_unit_id, location_id, NULL,
           inventory_status_id, NULL, $2, uom_id,
           outbound_check_exception_id, outbound_check_line_id,
           reason_code_id, $5, $6
    FROM context CROSS JOIN line_update
    JOIN movement_type ON movement_type.code = 'ADJUSTMENT' AND movement_type.is_active
    RETURNING movement_id
), resolution AS (
    INSERT INTO outbound_check_resolution (
        outbound_check_resolution_id, outbound_check_exception_id,
        outbound_check_resolution_type_id, resolved_qty, notes, created_by
    )
    SELECT $4, outbound_check_exception_id,
           outbound_check_resolution_type_id, $2, $5, $6
    FROM context CROSS JOIN movement
    RETURNING *
), linked AS (
    INSERT INTO outbound_check_resolution_movement (
        outbound_check_resolution_id, movement_id
    )
    SELECT resolution.outbound_check_resolution_id, movement.movement_id
    FROM resolution CROSS JOIN movement
    RETURNING *
)
UPDATE outbound_check_exception exception
SET status_id = progress.outbound_check_exception_status_id
FROM resolution, outbound_check_exception_status progress
WHERE exception.outbound_check_exception_id = resolution.outbound_check_exception_id
  AND progress.code = 'IN_PROGRESS'
RETURNING resolution.*;

COMMIT;

-- 3.6 Move DAMAGED or WRONG_ITEM stock out of active staging.
-- Target status must be non-allocatable.
-- $1 exception ID, $2 quantity, $3 target location ID,
-- $4 target inventory-status code, $5 target balance ID, $6 movement ID,
-- $7 resolution ID, $8 notes, $9 actor ID
BEGIN;

SELECT balance.balance_id
FROM outbound_check_exception exception
JOIN outbound_check_line check_line ON check_line.outbound_check_line_id = exception.outbound_check_line_id
JOIN outbound_staging_line staging_line ON staging_line.staging_line_id = check_line.staging_line_id
JOIN inventory_balance balance ON balance.balance_id = staging_line.staging_balance_id
WHERE exception.outbound_check_exception_id = $1
FOR UPDATE OF balance;

WITH context AS (
    SELECT exception.*, check_line.staging_line_id,
           staging_line.staging_balance_id, balance.location_id AS source_location_id,
           balance.item_id, balance.lot_id, balance.handling_unit_id,
           balance.inventory_status_id AS source_status_id,
           balance.on_hand_qty, check_line.uom_id, task.outbound_line_id,
           outbound.owner_id, outbound.warehouse_id, outbound.business_date,
           target.location_id AS target_location_id,
           target_status.inventory_status_id AS target_status_id,
           resolution_type.outbound_check_resolution_type_id,
           reason.reason_code_id
    FROM outbound_check_exception exception
    JOIN outbound_check_exception_status exception_status
      ON exception_status.outbound_check_exception_status_id = exception.status_id
     AND NOT exception_status.is_final
    JOIN outbound_check_line check_line ON check_line.outbound_check_line_id = exception.outbound_check_line_id
    JOIN outbound_check_result result
      ON result.outbound_check_result_id = check_line.outbound_check_result_id
     AND result.code IN ('DAMAGED', 'WRONG_ITEM')
    JOIN outbound_staging_line staging_line ON staging_line.staging_line_id = check_line.staging_line_id
    JOIN inventory_balance balance ON balance.balance_id = staging_line.staging_balance_id
    JOIN pick_execution execution ON execution.pick_execution_id = staging_line.pick_execution_id
    JOIN pick_task task ON task.pick_task_id = execution.pick_task_id
    JOIN outbound_order_line order_line ON order_line.outbound_line_id = task.outbound_line_id
    JOIN outbound_order outbound ON outbound.outbound_id = order_line.outbound_id
    JOIN warehouse_location target
      ON target.location_id = $3 AND target.warehouse_id = outbound.warehouse_id AND target.is_active
    JOIN inventory_status target_status
      ON target_status.code = $4 AND target_status.is_active AND NOT target_status.is_allocatable
    JOIN outbound_check_resolution_type resolution_type
      ON resolution_type.code = 'STOCK_CORRECTION' AND resolution_type.is_active
    JOIN reason_code reason
      ON reason.module_code = 'OUTBOUND' AND reason.code = 'CHECK_FAILED' AND reason.is_active
    WHERE exception.outbound_check_exception_id = $1 AND $2 > 0
      AND balance.on_hand_qty >= $2
      AND staging_line.staged_qty - staging_line.removed_qty >= $2
      AND (balance.handling_unit_id IS NULL OR balance.on_hand_qty = $2)
      AND COALESCE((
          SELECT sum(resolution.resolved_qty)
          FROM outbound_check_resolution resolution
          JOIN outbound_check_resolution_type type
            ON type.outbound_check_resolution_type_id = resolution.outbound_check_resolution_type_id
          WHERE resolution.outbound_check_exception_id = exception.outbound_check_exception_id
            AND type.counts_as_stock_correction
      ), 0) + $2 <= exception.exception_qty
), source_update AS (
    UPDATE inventory_balance balance
    SET on_hand_qty = balance.on_hand_qty - $2,
        version_no = balance.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE balance.balance_id = context.staging_balance_id AND balance.on_hand_qty >= $2
    RETURNING balance.balance_id
), target_balance AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id, lot_id,
        handling_unit_id, inventory_status_id, on_hand_qty, reserved_qty, uom_id
    )
    SELECT $5, owner_id, warehouse_id, target_location_id, item_id, lot_id,
           handling_unit_id, target_status_id, $2, 0, uom_id
    FROM context CROSS JOIN source_update
    ON CONFLICT (owner_id, warehouse_id, location_id, item_id,
                 lot_id, handling_unit_id, inventory_status_id)
    DO UPDATE SET on_hand_qty = inventory_balance.on_hand_qty + EXCLUDED.on_hand_qty,
                  version_no = inventory_balance.version_no + 1,
                  updated_at = clock_timestamp()
    RETURNING balance_id
), staging_update AS (
    UPDATE outbound_staging_line staging_line
    SET removed_qty = staging_line.removed_qty + $2
    FROM context, target_balance
    WHERE staging_line.staging_line_id = context.staging_line_id
      AND staging_line.staged_qty - staging_line.removed_qty >= $2
    RETURNING staging_line.staging_line_id
), line_update AS (
    UPDATE outbound_order_line line SET rejected_qty = line.rejected_qty + $2
    FROM context, staging_update
    WHERE line.outbound_line_id = context.outbound_line_id
    RETURNING line.outbound_line_id
), movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, handling_unit_id, from_location_id, to_location_id,
        from_status_id, to_status_id, quantity, uom_id,
        source_document_id, source_line_id, reason_code_id, notes, created_by
    )
    SELECT $6, movement_type.movement_type_id, owner_id, warehouse_id,
           business_date, item_id, lot_id, handling_unit_id,
           source_location_id, target_location_id, source_status_id, target_status_id,
           $2, uom_id, outbound_check_exception_id, outbound_check_line_id,
           reason_code_id, $8, $9
    FROM context CROSS JOIN line_update
    JOIN movement_type
      ON movement_type.code = 'OUTBOUND_CHECK_CORRECTION' AND movement_type.is_active
    RETURNING movement_id
), resolution AS (
    INSERT INTO outbound_check_resolution (
        outbound_check_resolution_id, outbound_check_exception_id,
        outbound_check_resolution_type_id, resolved_qty, notes, created_by
    )
    SELECT $7, outbound_check_exception_id,
           outbound_check_resolution_type_id, $2, $8, $9
    FROM context CROSS JOIN movement RETURNING *
), linked AS (
    INSERT INTO outbound_check_resolution_movement
        (outbound_check_resolution_id, movement_id)
    SELECT resolution.outbound_check_resolution_id, movement.movement_id
    FROM resolution CROSS JOIN movement RETURNING *
), updated_hu AS (
    UPDATE handling_unit hu SET current_location_id = context.target_location_id
    FROM context, linked
    WHERE hu.handling_unit_id = context.handling_unit_id
    RETURNING hu.handling_unit_id
)
UPDATE outbound_check_exception exception
SET status_id = progress.outbound_check_exception_status_id
FROM resolution, outbound_check_exception_status progress
WHERE exception.outbound_check_exception_id = resolution.outbound_check_exception_id
  AND progress.code = 'IN_PROGRESS'
RETURNING resolution.*;

COMMIT;

-- 3.7 Register and isolate OVER stock discovered during checking.
-- The extra stock is created only in a non-allocatable balance for investigation.
-- Parameters are the same as 3.6.
WITH context AS (
    SELECT exception.*, balance.item_id, balance.lot_id,
           balance.handling_unit_id, check_line.uom_id,
           outbound.owner_id, outbound.warehouse_id, outbound.business_date,
           target.location_id AS target_location_id,
           target_status.inventory_status_id AS target_status_id,
           resolution_type.outbound_check_resolution_type_id,
           reason.reason_code_id
    FROM outbound_check_exception exception
    JOIN outbound_check_exception_status exception_status
      ON exception_status.outbound_check_exception_status_id = exception.status_id
     AND NOT exception_status.is_final
    JOIN outbound_check_line check_line ON check_line.outbound_check_line_id = exception.outbound_check_line_id
    JOIN outbound_check_result result
      ON result.outbound_check_result_id = check_line.outbound_check_result_id AND result.code = 'OVER'
    JOIN outbound_staging_line staging_line ON staging_line.staging_line_id = check_line.staging_line_id
    JOIN inventory_balance balance ON balance.balance_id = staging_line.staging_balance_id
    JOIN outbound_staging staging ON staging.staging_id = staging_line.staging_id
    JOIN outbound_order outbound ON outbound.outbound_id = staging.outbound_id
    JOIN warehouse_location target
      ON target.location_id = $3 AND target.warehouse_id = outbound.warehouse_id AND target.is_active
    JOIN inventory_status target_status
      ON target_status.code = $4 AND target_status.is_active AND NOT target_status.is_allocatable
    JOIN outbound_check_resolution_type resolution_type
      ON resolution_type.code = 'STOCK_CORRECTION' AND resolution_type.is_active
    JOIN reason_code reason
      ON reason.module_code = 'OUTBOUND' AND reason.code = 'CHECK_FAILED' AND reason.is_active
    WHERE exception.outbound_check_exception_id = $1 AND $2 > 0
      AND COALESCE((
          SELECT sum(resolution.resolved_qty)
          FROM outbound_check_resolution resolution
          JOIN outbound_check_resolution_type type
            ON type.outbound_check_resolution_type_id = resolution.outbound_check_resolution_type_id
          WHERE resolution.outbound_check_exception_id = exception.outbound_check_exception_id
            AND type.counts_as_stock_correction
      ), 0) + $2 <= exception.exception_qty
), target_balance AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id, lot_id,
        handling_unit_id, inventory_status_id, on_hand_qty, reserved_qty, uom_id
    )
    SELECT $5, owner_id, warehouse_id, target_location_id, item_id, lot_id,
           NULL, target_status_id, $2, 0, uom_id
    FROM context
    ON CONFLICT (owner_id, warehouse_id, location_id, item_id,
                 lot_id, handling_unit_id, inventory_status_id)
    DO UPDATE SET on_hand_qty = inventory_balance.on_hand_qty + EXCLUDED.on_hand_qty,
                  version_no = inventory_balance.version_no + 1,
                  updated_at = clock_timestamp()
    RETURNING balance_id
), movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, from_location_id, to_location_id,
        from_status_id, to_status_id, quantity, uom_id,
        source_document_id, source_line_id, reason_code_id, notes, created_by
    )
    SELECT $6, movement_type.movement_type_id, owner_id, warehouse_id,
           business_date, item_id, lot_id, NULL, target_location_id,
           NULL, target_status_id, $2, uom_id,
           outbound_check_exception_id, outbound_check_line_id,
           reason_code_id, $8, $9
    FROM context CROSS JOIN target_balance
    JOIN movement_type ON movement_type.code = 'ADJUSTMENT' AND movement_type.is_active
    RETURNING movement_id
), resolution AS (
    INSERT INTO outbound_check_resolution (
        outbound_check_resolution_id, outbound_check_exception_id,
        outbound_check_resolution_type_id, resolved_qty, notes, created_by
    )
    SELECT $7, outbound_check_exception_id,
           outbound_check_resolution_type_id, $2, $8, $9
    FROM context CROSS JOIN movement RETURNING *
), linked AS (
    INSERT INTO outbound_check_resolution_movement
        (outbound_check_resolution_id, movement_id)
    SELECT resolution.outbound_check_resolution_id, movement.movement_id
    FROM resolution CROSS JOIN movement RETURNING *
)
UPDATE outbound_check_exception exception
SET status_id = progress.outbound_check_exception_status_id
FROM resolution, outbound_check_exception_status progress
WHERE exception.outbound_check_exception_id = resolution.outbound_check_exception_id
  AND progress.code = 'IN_PROGRESS'
RETURNING resolution.*;

-- 3.8 Accept a corrected SHORT without replacement.
-- $1 exception ID, $2 accepted-short quantity, $3 resolution ID,
-- $4 approval note, $5 approving actor ID
WITH context AS (
    SELECT exception.*, task.outbound_line_id,
           resolution_type.outbound_check_resolution_type_id
    FROM outbound_check_exception exception
    JOIN outbound_check_exception_status exception_status
      ON exception_status.outbound_check_exception_status_id = exception.status_id
     AND NOT exception_status.is_final
    JOIN outbound_check_line check_line ON check_line.outbound_check_line_id = exception.outbound_check_line_id
    JOIN outbound_check_result result
      ON result.outbound_check_result_id = check_line.outbound_check_result_id AND result.code = 'SHORT'
    JOIN outbound_staging_line staging_line ON staging_line.staging_line_id = check_line.staging_line_id
    JOIN pick_execution execution ON execution.pick_execution_id = staging_line.pick_execution_id
    JOIN pick_task task ON task.pick_task_id = execution.pick_task_id
    JOIN outbound_check_resolution_type resolution_type
      ON resolution_type.code = 'ACCEPT_SHORT' AND resolution_type.is_active
    WHERE exception.outbound_check_exception_id = $1 AND $2 > 0
      AND COALESCE((
          SELECT sum(resolution.resolved_qty)
          FROM outbound_check_resolution resolution
          JOIN outbound_check_resolution_type type
            ON type.outbound_check_resolution_type_id = resolution.outbound_check_resolution_type_id
          WHERE resolution.outbound_check_exception_id = exception.outbound_check_exception_id
            AND (type.counts_as_replacement OR type.counts_as_short_acceptance)
      ), 0) + $2 <= exception.exception_qty
      AND COALESCE((
          SELECT sum(resolution.resolved_qty)
          FROM outbound_check_resolution resolution
          JOIN outbound_check_resolution_type type
            ON type.outbound_check_resolution_type_id = resolution.outbound_check_resolution_type_id
          WHERE resolution.outbound_check_exception_id = exception.outbound_check_exception_id
            AND type.counts_as_stock_correction
      ), 0) >= $2
), resolution AS (
    INSERT INTO outbound_check_resolution (
        outbound_check_resolution_id, outbound_check_exception_id,
        outbound_check_resolution_type_id, resolved_qty, notes,
        approved_at, approved_by, created_by
    )
    SELECT $3, outbound_check_exception_id,
           outbound_check_resolution_type_id, $2, $4,
           clock_timestamp(), $5, $5
    FROM context RETURNING *
), updated_line AS (
    UPDATE outbound_order_line line
    SET short_accepted_qty = line.short_accepted_qty + resolution.resolved_qty
    FROM context, resolution
    WHERE line.outbound_line_id = context.outbound_line_id
    RETURNING line.*
)
SELECT resolution.* FROM resolution CROSS JOIN updated_line;

-- 3.9 Allocate replacement stock and create a supplemental pick task.
-- $1 exception ID, $2 source balance ID, $3 quantity, $4 reservation ID,
-- $5 picking-strategy ID or NULL, $6 pick-task ID, $7 resolution ID,
-- $8 task-priority code, $9 actor ID
BEGIN;

SELECT balance_id FROM inventory_balance WHERE balance_id = $2 FOR UPDATE;

WITH context AS (
    SELECT exception.outbound_check_exception_id, exception.exception_qty,
           check_line.outbound_check_line_id, task.outbound_line_id,
           staging.staging_id, staging.wave_id, staging.staging_location_id,
           outbound.owner_id, outbound.warehouse_id, outbound.business_date,
           line.item_id, line.uom_id, line.requested_lot_no,
           balance.location_id AS source_location_id,
           reservation_type.document_type_id AS reservation_document_type_id,
           active_status.status_id AS reservation_status_id,
           task_type.task_type_id, task_status.task_status_id,
           priority.task_priority_id, strategy.picking_strategy_id,
           resolution_type.outbound_check_resolution_type_id
    FROM outbound_check_exception exception
    JOIN outbound_check_exception_status exception_status
      ON exception_status.outbound_check_exception_status_id = exception.status_id
     AND NOT exception_status.is_final
    JOIN outbound_check_line check_line ON check_line.outbound_check_line_id = exception.outbound_check_line_id
    JOIN outbound_check_result result
      ON result.outbound_check_result_id = check_line.outbound_check_result_id
     AND result.code IN ('SHORT', 'DAMAGED', 'WRONG_ITEM')
    JOIN outbound_staging_line old_staging_line ON old_staging_line.staging_line_id = check_line.staging_line_id
    JOIN outbound_staging staging ON staging.staging_id = old_staging_line.staging_id
    JOIN pick_execution old_execution ON old_execution.pick_execution_id = old_staging_line.pick_execution_id
    JOIN pick_task task ON task.pick_task_id = old_execution.pick_task_id
    JOIN outbound_order_line line ON line.outbound_line_id = task.outbound_line_id
    JOIN outbound_order outbound ON outbound.outbound_id = line.outbound_id
    JOIN document_status outbound_status
      ON outbound_status.status_id = outbound.status_id AND outbound_status.code = 'CHECK_FAILED'
    JOIN inventory_balance balance
      ON balance.balance_id = $2 AND balance.owner_id = outbound.owner_id
     AND balance.warehouse_id = outbound.warehouse_id
     AND balance.item_id = line.item_id AND balance.uom_id = line.uom_id
    JOIN inventory_status inventory_status
      ON inventory_status.inventory_status_id = balance.inventory_status_id
     AND inventory_status.is_active AND inventory_status.is_allocatable
    LEFT JOIN inventory_lot lot ON lot.lot_id = balance.lot_id
    JOIN document_type reservation_type
      ON reservation_type.code = 'RESERVATION' AND reservation_type.is_active
    JOIN document_status active_status
      ON active_status.document_type_id = reservation_type.document_type_id
     AND active_status.code = 'ACTIVE' AND active_status.is_active
    JOIN task_type ON task_type.code = 'PICK' AND task_type.is_active
    JOIN task_status ON task_status.code = 'OPEN' AND task_status.is_active
    JOIN task_priority priority ON priority.code = $8 AND priority.is_active
    LEFT JOIN picking_strategy strategy
      ON strategy.picking_strategy_id = $5 AND strategy.is_active
     AND (strategy.owner_id IS NULL OR strategy.owner_id = outbound.owner_id)
     AND (strategy.warehouse_id IS NULL OR strategy.warehouse_id = outbound.warehouse_id)
    JOIN outbound_check_resolution_type resolution_type
      ON resolution_type.code = 'REPLACEMENT' AND resolution_type.is_active
    WHERE exception.outbound_check_exception_id = $1 AND $3 > 0
      AND balance.on_hand_qty - balance.reserved_qty >= $3
      AND (line.requested_lot_no IS NULL OR lot.lot_number = line.requested_lot_no)
      AND ($5::uuid IS NULL OR strategy.picking_strategy_id IS NOT NULL)
      AND COALESCE((
          SELECT sum(resolution.resolved_qty)
          FROM outbound_check_resolution resolution
          JOIN outbound_check_resolution_type type
            ON type.outbound_check_resolution_type_id = resolution.outbound_check_resolution_type_id
          WHERE resolution.outbound_check_exception_id = exception.outbound_check_exception_id
            AND type.counts_as_stock_correction
      ), 0) >= $3
      AND COALESCE((
          SELECT sum(resolution.resolved_qty)
          FROM outbound_check_resolution resolution
          JOIN outbound_check_resolution_type type
            ON type.outbound_check_resolution_type_id = resolution.outbound_check_resolution_type_id
          WHERE resolution.outbound_check_exception_id = exception.outbound_check_exception_id
            AND (type.counts_as_replacement OR type.counts_as_short_acceptance)
      ), 0) + $3 <= exception.exception_qty
), reserved_balance AS (
    UPDATE inventory_balance balance
    SET reserved_qty = balance.reserved_qty + $3,
        version_no = balance.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE balance.balance_id = $2
      AND balance.on_hand_qty - balance.reserved_qty >= $3
    RETURNING balance.balance_id
), reservation AS (
    INSERT INTO inventory_reservation (
        reservation_id, document_type_id, status_id, picking_strategy_id,
        outbound_line_id, balance_id, reserved_qty, picked_qty, uom_id, created_by
    )
    SELECT $4, reservation_document_type_id, reservation_status_id,
           picking_strategy_id, outbound_line_id, $2, $3, 0, uom_id, $9
    FROM context CROSS JOIN reserved_balance RETURNING *
), updated_line AS (
    UPDATE outbound_order_line line SET allocated_qty = line.allocated_qty + $3
    FROM reservation
    WHERE line.outbound_line_id = reservation.outbound_line_id RETURNING line.*
), resolution AS (
    INSERT INTO outbound_check_resolution (
        outbound_check_resolution_id, outbound_check_exception_id,
        outbound_check_resolution_type_id, resolved_qty, created_by
    )
    SELECT $7, outbound_check_exception_id,
           outbound_check_resolution_type_id, $3, $9
    FROM context CROSS JOIN updated_line RETURNING *
), linked AS (
    INSERT INTO outbound_check_resolution_reservation
        (outbound_check_resolution_id, reservation_id)
    SELECT resolution.outbound_check_resolution_id, reservation.reservation_id
    FROM resolution CROSS JOIN reservation RETURNING *
), created_task AS (
    INSERT INTO pick_task (
        pick_task_id, task_type_id, task_status_id, task_priority_id,
        wave_id, reservation_id, outbound_line_id,
        source_location_id, target_location_id,
        planned_qty, picked_qty, uom_id, created_by
    )
    SELECT $6, task_type_id, task_status_id, task_priority_id,
           wave_id, reservation.reservation_id, context.outbound_line_id,
           source_location_id, staging_location_id,
           $3, 0, context.uom_id, $9
    FROM context CROSS JOIN reservation CROSS JOIN linked RETURNING *
), updated_wave AS (
    UPDATE outbound_wave wave
    SET status_id = progress.status_id, completed_at = NULL,
        updated_at = clock_timestamp(), updated_by = $9,
        version_no = wave.version_no + 1
    FROM context, created_task, document_status progress
    WHERE wave.wave_id = context.wave_id
      AND progress.document_type_id = wave.document_type_id
      AND progress.code = 'IN_PROGRESS'
    RETURNING wave.wave_id
)
UPDATE outbound_check_exception exception
SET status_id = progress.outbound_check_exception_status_id
FROM resolution, outbound_check_exception_status progress
WHERE exception.outbound_check_exception_id = resolution.outbound_check_exception_id
  AND progress.code = 'IN_PROGRESS'
RETURNING resolution.*, (SELECT pick_task_id FROM created_task);

COMMIT;

-- 3.10 Attach a completed replacement pick execution to existing staging.
-- $1 staging ID, $2 pick-execution ID, $3 staging-line ID, $4 actor ID
INSERT INTO outbound_staging_line (
    staging_line_id, staging_id, pick_execution_id,
    staging_balance_id, staged_qty, uom_id, created_by
)
SELECT $3, staging.staging_id, execution.pick_execution_id,
       execution.staging_balance_id, execution.picked_qty,
       execution.uom_id, $4
FROM outbound_staging staging
JOIN pick_execution execution ON execution.pick_execution_id = $2
JOIN pick_task task
  ON task.pick_task_id = execution.pick_task_id AND task.wave_id = staging.wave_id
JOIN task_status task_status
  ON task_status.task_status_id = task.task_status_id AND task_status.code = 'COMPLETED'
JOIN outbound_order_line line
  ON line.outbound_line_id = task.outbound_line_id AND line.outbound_id = staging.outbound_id
JOIN outbound_check_resolution_reservation replacement
  ON replacement.reservation_id = task.reservation_id
WHERE staging.staging_id = $1
  AND NOT EXISTS (
      SELECT 1 FROM outbound_staging_line existing
      WHERE existing.pick_execution_id = execution.pick_execution_id
  )
RETURNING *;

-- 3.11 Close an exception only when its required evidence is complete.
-- $1 exception ID, $2 actor ID
WITH totals AS (
    SELECT exception.outbound_check_exception_id, exception.exception_qty,
           result.code AS result_code,
           COALESCE((
               SELECT sum(resolution.resolved_qty)
               FROM outbound_check_resolution resolution
               JOIN outbound_check_resolution_type type
                 ON type.outbound_check_resolution_type_id = resolution.outbound_check_resolution_type_id
               WHERE resolution.outbound_check_exception_id = exception.outbound_check_exception_id
                 AND type.counts_as_stock_correction
           ), 0) AS corrected_qty,
           COALESCE((
               SELECT sum(resolution.resolved_qty)
               FROM outbound_check_resolution resolution
               JOIN outbound_check_resolution_type type
                 ON type.outbound_check_resolution_type_id = resolution.outbound_check_resolution_type_id
               WHERE resolution.outbound_check_exception_id = exception.outbound_check_exception_id
                 AND type.counts_as_short_acceptance
                 AND resolution.approved_by IS NOT NULL
           ), 0) AS accepted_short_qty,
           COALESCE((
               SELECT sum(resolution.resolved_qty)
               FROM outbound_check_resolution resolution
               JOIN outbound_check_resolution_type type
                 ON type.outbound_check_resolution_type_id = resolution.outbound_check_resolution_type_id
               JOIN outbound_check_resolution_reservation replacement
                 ON replacement.outbound_check_resolution_id = resolution.outbound_check_resolution_id
               JOIN inventory_reservation reservation ON reservation.reservation_id = replacement.reservation_id
               JOIN document_status reservation_status
                 ON reservation_status.status_id = reservation.status_id
                AND reservation_status.code = 'CONSUMED'
               WHERE resolution.outbound_check_exception_id = exception.outbound_check_exception_id
                 AND type.counts_as_replacement
                 AND EXISTS (
                     SELECT 1
                     FROM pick_task replacement_task
                     JOIN pick_execution replacement_execution
                       ON replacement_execution.pick_task_id = replacement_task.pick_task_id
                     JOIN outbound_staging_line replacement_staging
                       ON replacement_staging.pick_execution_id = replacement_execution.pick_execution_id
                     WHERE replacement_task.reservation_id = reservation.reservation_id
                 )
           ), 0) AS replaced_qty
    FROM outbound_check_exception exception
    JOIN outbound_check_exception_status exception_status
      ON exception_status.outbound_check_exception_status_id = exception.status_id
     AND NOT exception_status.is_final
    JOIN outbound_check_line check_line ON check_line.outbound_check_line_id = exception.outbound_check_line_id
    JOIN outbound_check_result result
      ON result.outbound_check_result_id = check_line.outbound_check_result_id
    WHERE exception.outbound_check_exception_id = $1
), eligible AS (
    SELECT * FROM totals
    WHERE corrected_qty >= exception_qty
      AND CASE result_code
          WHEN 'SHORT' THEN replaced_qty + accepted_short_qty >= exception_qty
          WHEN 'DAMAGED' THEN replaced_qty >= exception_qty
          WHEN 'WRONG_ITEM' THEN replaced_qty >= exception_qty
          WHEN 'OVER' THEN true
          ELSE false
      END
)
UPDATE outbound_check_exception exception
SET status_id = resolved.outbound_check_exception_status_id,
    resolved_at = clock_timestamp(), resolved_by = $2
FROM eligible, outbound_check_exception_status resolved
WHERE exception.outbound_check_exception_id = eligible.outbound_check_exception_id
  AND resolved.code = 'RESOLVED' AND resolved.is_active
RETURNING exception.*;

-- =============================================================================
-- 4. PACKING
-- =============================================================================

-- 4.1 Create an open packing document after a passed check.
-- $1 outbound-check ID, $2 packing location ID, $3 actor ID
WITH context AS (
    SELECT check_header.outbound_check_id, staging.outbound_id,
           staging.warehouse_id, outbound.business_date,
           warehouse.code AS warehouse_code, location.location_id,
           dt.document_type_id, initial_status.status_id
    FROM outbound_check check_header
    JOIN document_status check_status ON check_status.status_id = check_header.status_id
    JOIN outbound_staging staging ON staging.staging_id = check_header.staging_id
    JOIN outbound_order outbound ON outbound.outbound_id = staging.outbound_id
    JOIN warehouse ON warehouse.warehouse_id = staging.warehouse_id
    JOIN warehouse_location location
      ON location.location_id = $2
     AND location.warehouse_id = staging.warehouse_id AND location.is_active
    JOIN location_type packing_type
      ON packing_type.location_type_id = location.location_type_id
     AND packing_type.code = 'PACKING' AND packing_type.is_active
    JOIN document_type dt ON dt.code = 'PACKING' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE check_header.outbound_check_id = $1 AND check_status.code = 'PASSED'
      AND NOT EXISTS (
          SELECT 1 FROM packing existing
          JOIN document_status existing_status ON existing_status.status_id = existing.status_id
          WHERE existing.outbound_id = staging.outbound_id
            AND NOT existing_status.is_cancelled
      )
), numbered AS (
    SELECT context.*,
           generate_document_id('PACKING', NULL, warehouse_code, business_date) AS generated_id
    FROM context
), created_packing AS (
    INSERT INTO packing (
        packing_id, document_type_id, status_id, outbound_id,
        warehouse_id, packing_location_id, created_by
    )
    SELECT generated_id, document_type_id, status_id, outbound_id,
           warehouse_id, location_id, $3
    FROM numbered
    RETURNING *
), updated_outbound AS (
    UPDATE outbound_order outbound
    SET status_id = packing_status.status_id,
        updated_at = clock_timestamp(), updated_by = $3,
        version_no = outbound.version_no + 1
    FROM created_packing packing, document_status packing_status
    WHERE outbound.outbound_id = packing.outbound_id
      AND packing_status.document_type_id = outbound.document_type_id
      AND packing_status.code = 'PACKING'
    RETURNING outbound.outbound_id
)
SELECT created_packing.* FROM created_packing CROSS JOIN updated_outbound;

-- 4.2 Pack one passed check line atomically.
-- $1 packing ID, $2 check-line ID, $3 handling-unit ID or NULL,
-- $4 proposed packing balance ID, $5 movement ID, $6 actor ID
BEGIN;

SELECT balance.balance_id
FROM outbound_check_line check_line
JOIN outbound_staging_line staging_line ON staging_line.staging_line_id = check_line.staging_line_id
JOIN inventory_balance balance ON balance.balance_id = staging_line.staging_balance_id
WHERE check_line.outbound_check_line_id = $2
FOR UPDATE OF balance;

WITH context AS (
    SELECT packing.packing_id, packing.outbound_id, packing.warehouse_id,
           packing.packing_location_id, check_line.outbound_check_line_id,
           check_line.checked_qty, check_line.uom_id,
           staging_line.staging_balance_id, execution.pick_task_id,
           task.outbound_line_id, outbound.owner_id, outbound.business_date,
           source.location_id AS source_location_id, source.item_id,
           source.lot_id, source.handling_unit_id AS source_handling_unit_id,
           source.inventory_status_id, hu.handling_unit_id AS target_handling_unit_id
    FROM packing
    JOIN document_status packing_status ON packing_status.status_id = packing.status_id
    JOIN outbound_order outbound ON outbound.outbound_id = packing.outbound_id
    JOIN outbound_check check_header ON check_header.staging_id = (
        SELECT staging_id FROM outbound_staging WHERE outbound_id = packing.outbound_id
    )
    JOIN document_status check_status ON check_status.status_id = check_header.status_id
    JOIN outbound_check_line check_line
      ON check_line.outbound_check_id = check_header.outbound_check_id
     AND check_line.outbound_check_line_id = $2
    JOIN outbound_check_result result
      ON result.outbound_check_result_id = check_line.outbound_check_result_id
     AND result.is_pass
    JOIN outbound_staging_line staging_line
      ON staging_line.staging_line_id = check_line.staging_line_id
    JOIN pick_execution execution ON execution.pick_execution_id = staging_line.pick_execution_id
    JOIN pick_task task ON task.pick_task_id = execution.pick_task_id
    JOIN inventory_balance source ON source.balance_id = staging_line.staging_balance_id
    LEFT JOIN handling_unit hu
      ON hu.handling_unit_id = $3
     AND hu.owner_id = outbound.owner_id AND hu.warehouse_id = packing.warehouse_id
    WHERE packing.packing_id = $1 AND packing_status.code = 'OPEN'
      AND check_status.code = 'PASSED'
      AND check_line.checked_qty = check_line.expected_qty
      AND source.on_hand_qty >= check_line.checked_qty
      AND source.location_id <> packing.packing_location_id
      AND ($3::varchar IS NULL OR hu.handling_unit_id IS NOT NULL)
), source_update AS (
    UPDATE inventory_balance balance
    SET on_hand_qty = balance.on_hand_qty - context.checked_qty,
        version_no = balance.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE balance.balance_id = context.staging_balance_id
      AND balance.on_hand_qty >= context.checked_qty
    RETURNING balance.balance_id
), packing_balance AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT $4, context.owner_id, context.warehouse_id,
           context.packing_location_id, context.item_id, context.lot_id,
           COALESCE(context.target_handling_unit_id, context.source_handling_unit_id),
           context.inventory_status_id, context.checked_qty, 0, context.uom_id
    FROM context CROSS JOIN source_update
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
    SELECT $5, movement_type.movement_type_id, context.owner_id,
           context.warehouse_id, context.business_date, context.item_id,
           context.lot_id, packing_balance.handling_unit_id,
           context.source_location_id, context.packing_location_id,
           context.inventory_status_id, context.inventory_status_id,
           context.checked_qty, context.uom_id, context.packing_id,
           context.outbound_check_line_id, $6
    FROM context CROSS JOIN packing_balance
    JOIN movement_type ON movement_type.code = 'PACK' AND movement_type.is_active
    RETURNING movement_id
), packed_line AS (
    INSERT INTO packing_line (
        packing_line_id, packing_id, pick_task_id, outbound_check_line_id,
        source_balance_id, packing_balance_id, handling_unit_id,
        packed_qty, uom_id, movement_id, created_by
    )
    SELECT context.packing_id || '-L-' || lpad(check_line.line_no::text, 6, '0'),
           context.packing_id, context.pick_task_id,
           context.outbound_check_line_id, context.staging_balance_id,
           packing_balance.balance_id, packing_balance.handling_unit_id,
           context.checked_qty, context.uom_id, movement.movement_id, $6
    FROM context
    JOIN outbound_check_line check_line
      ON check_line.outbound_check_line_id = context.outbound_check_line_id
    CROSS JOIN packing_balance CROSS JOIN movement
    RETURNING *
), updated_line AS (
    UPDATE outbound_order_line line
    SET packed_qty = line.packed_qty + packed.packed_qty
    FROM packed_line packed
    JOIN pick_task task ON task.pick_task_id = packed.pick_task_id
    WHERE line.outbound_line_id = task.outbound_line_id
    RETURNING line.*
), updated_hu AS (
    UPDATE handling_unit hu
    SET current_location_id = context.packing_location_id
    FROM context, packed_line
    WHERE hu.handling_unit_id = COALESCE(context.target_handling_unit_id,
                                         context.source_handling_unit_id)
    RETURNING hu.handling_unit_id
)
SELECT packed_line.* FROM packed_line CROSS JOIN updated_line;

COMMIT;

-- 4.3 Complete packing when every passed check line is packed.
-- $1 packing ID, $2 actor ID
WITH completed_packing AS (
    UPDATE packing packing
    SET status_id = completed.status_id,
        packed_at = clock_timestamp(), packed_by = $2
    FROM document_status current_status, document_status completed
    WHERE packing.packing_id = $1
      AND current_status.status_id = packing.status_id AND current_status.code = 'OPEN'
      AND completed.document_type_id = packing.document_type_id
      AND completed.code = 'COMPLETED'
      AND NOT EXISTS (
          SELECT 1
          FROM outbound_check check_header
          JOIN document_status check_status ON check_status.status_id = check_header.status_id
          JOIN outbound_staging staging ON staging.staging_id = check_header.staging_id
          JOIN outbound_check_line check_line ON check_line.outbound_check_id = check_header.outbound_check_id
          WHERE staging.outbound_id = packing.outbound_id
            AND check_status.code = 'PASSED'
            AND NOT EXISTS (
                SELECT 1 FROM packing_line packed
                WHERE packed.packing_id = packing.packing_id
                  AND packed.outbound_check_line_id = check_line.outbound_check_line_id
            )
      )
    RETURNING packing.*
), updated_outbound AS (
    UPDATE outbound_order outbound
    SET status_id = packed.status_id,
        updated_at = clock_timestamp(), updated_by = $2,
        version_no = outbound.version_no + 1
    FROM completed_packing packing, document_status packed
    WHERE outbound.outbound_id = packing.outbound_id
      AND packed.document_type_id = outbound.document_type_id
      AND packed.code = 'PACKED'
    RETURNING outbound.outbound_id
)
SELECT completed_packing.* FROM completed_packing CROSS JOIN updated_outbound;

-- 4.4 Cancel an empty packing document and return the DO to CHECKED.
-- $1 packing ID, $2 actor ID
WITH cancelled_packing AS (
    UPDATE packing packing
    SET status_id = cancelled.status_id
    FROM document_status current_status, document_status cancelled
    WHERE packing.packing_id = $1
      AND current_status.status_id = packing.status_id AND current_status.code = 'OPEN'
      AND cancelled.document_type_id = packing.document_type_id
      AND cancelled.code = 'CANCELLED'
      AND NOT EXISTS (SELECT 1 FROM packing_line line WHERE line.packing_id = packing.packing_id)
    RETURNING packing.*
), restored_outbound AS (
    UPDATE outbound_order outbound
    SET status_id = checked.status_id,
        updated_at = clock_timestamp(), updated_by = $2,
        version_no = outbound.version_no + 1
    FROM cancelled_packing packing, document_status checked
    WHERE outbound.outbound_id = packing.outbound_id
      AND checked.document_type_id = outbound.document_type_id
      AND checked.code = 'CHECKED'
    RETURNING outbound.outbound_id
)
SELECT cancelled_packing.* FROM cancelled_packing CROSS JOIN restored_outbound;
