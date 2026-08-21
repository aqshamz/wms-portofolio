-- Client delivery order, validation, allocation, and outbound-wave operations.
-- Execute one numbered operation at a time using prepared parameters.

SET search_path TO wms, public;

-- =============================================================================
-- 1. CLIENT DELIVERY ORDER CRUD
-- =============================================================================

-- 1.1 Create a client DO. Ship-to address is snapshotted from the partner master.
-- $1 owner ID, $2 customer/client partner code, $3 ship-to store partner code,
-- $4 warehouse code, $5 business date, $6 client DO number,
-- $7 customer order number or NULL, $8 external reference or NULL,
-- $9 requested ship timestamp or NULL, $10 notes, $11 actor account ID
WITH context AS (
    SELECT owner.organization_id AS owner_id,
           customer.partner_id AS customer_id,
           store.partner_id AS ship_to_partner_id,
           store.name AS ship_to_name,
           store.address_line_1, store.address_line_2, store.city,
           store.province, store.postal_code, store.country_code,
           warehouse.warehouse_id, warehouse.code AS warehouse_code,
           dt.document_type_id, initial_status.status_id
    FROM organization owner
    JOIN business_partner customer
      ON customer.owner_id = owner.organization_id
     AND customer.code = $2 AND customer.is_active
    JOIN business_partner store
      ON store.owner_id = owner.organization_id
     AND store.code = $3 AND store.is_active
    JOIN warehouse_owner scope
      ON scope.owner_id = owner.organization_id AND scope.is_active
    JOIN warehouse ON warehouse.warehouse_id = scope.warehouse_id
                  AND warehouse.code = $4 AND warehouse.is_active
    JOIN document_type dt ON dt.code = 'OUTBOUND' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE owner.organization_id = $1
), numbered AS (
    SELECT context.*,
           generate_document_id('OUTBOUND', $2, warehouse_code, $5) AS generated_id
    FROM context
)
INSERT INTO outbound_order (
    outbound_id, document_type_id, status_id, owner_id, customer_id,
    ship_to_partner_id, warehouse_id, business_date,
    client_delivery_order_no, customer_order_no, external_reference,
    requested_ship_at, ship_to_name, ship_to_address_1, ship_to_address_2,
    ship_to_city, ship_to_province, ship_to_postal_code,
    ship_to_country_code, notes, created_by, updated_by
)
SELECT generated_id, document_type_id, status_id, owner_id, customer_id,
       ship_to_partner_id, warehouse_id, $5, $6, $7, $8, $9,
       ship_to_name, address_line_1, address_line_2, city, province,
       postal_code, country_code, $10, $11, $11
FROM numbered
RETURNING *;

-- 1.2 Add a draft DO line in the item base UOM.
-- $1 outbound ID, $2 line number, $3 item code, $4 ordered base quantity,
-- $5 requested lot number or NULL, $6 client line reference or NULL,
-- $7 notes, $8 actor ID
INSERT INTO outbound_order_line (
    outbound_line_id, outbound_id, line_no, item_id, ordered_qty, uom_id,
    requested_lot_no, customer_line_reference, notes, created_by
)
SELECT outbound.outbound_id || '-L-' || lpad($2::text, 4, '0'),
       outbound.outbound_id, $2::integer, item.item_id, $4::numeric, item.base_uom_id,
       $5, $6, $7, $8
FROM outbound_order outbound
JOIN document_status status ON status.status_id = outbound.status_id
JOIN item ON item.owner_id = outbound.owner_id AND item.code = $3 AND item.is_active
WHERE outbound.outbound_id = $1
  AND status.code = 'DRAFT'
  AND $4::numeric > 0
RETURNING *;

-- 1.3 Update a draft DO header and refresh its destination snapshot.
-- $1 outbound ID, $2 ship-to partner code, $3 requested ship time,
-- $4 notes, $5 expected version, $6 actor ID
UPDATE outbound_order outbound
SET ship_to_partner_id = store.partner_id,
    requested_ship_at = $3,
    ship_to_name = store.name,
    ship_to_address_1 = store.address_line_1,
    ship_to_address_2 = store.address_line_2,
    ship_to_city = store.city,
    ship_to_province = store.province,
    ship_to_postal_code = store.postal_code,
    ship_to_country_code = store.country_code,
    notes = $4,
    updated_at = clock_timestamp(), updated_by = $6,
    version_no = outbound.version_no + 1
FROM business_partner store, document_status status
WHERE outbound.outbound_id = $1
  AND store.owner_id = outbound.owner_id
  AND store.code = $2 AND store.is_active
  AND status.status_id = outbound.status_id AND status.code = 'DRAFT'
  AND outbound.version_no = $5
RETURNING outbound.*;

-- 1.4 Delete a draft line. $1 outbound-line ID
DELETE FROM outbound_order_line line
USING outbound_order outbound, document_status status
WHERE line.outbound_line_id = $1
  AND outbound.outbound_id = line.outbound_id
  AND status.status_id = outbound.status_id AND status.code = 'DRAFT'
RETURNING line.*;

-- 1.5 DO list.
-- $1 owner ID, $2 warehouse ID, $3 status or NULL,
-- $4 DO number search or NULL, $5 date from, $6 date until,
-- $7 limit, $8 offset
SELECT outbound.outbound_id, outbound.client_delivery_order_no,
       outbound.business_date, status.code AS status_code,
       customer.code AS customer_code, store.code AS store_code,
       store.name AS store_name, outbound.requested_ship_at,
       sum(line.ordered_qty) AS ordered_qty,
       sum(line.allocated_qty) AS allocated_qty,
       sum(line.picked_qty) AS picked_qty,
       sum(line.packed_qty) AS packed_qty,
       sum(line.shipped_qty) AS shipped_qty,
       sum(line.delivered_qty) AS delivered_qty,
       count(*) OVER () AS total_rows
FROM outbound_order outbound
JOIN document_status status ON status.status_id = outbound.status_id
JOIN business_partner customer ON customer.partner_id = outbound.customer_id
LEFT JOIN business_partner store ON store.partner_id = outbound.ship_to_partner_id
JOIN outbound_order_line line ON line.outbound_id = outbound.outbound_id
WHERE outbound.owner_id = $1 AND outbound.warehouse_id = $2
  AND ($3::varchar IS NULL OR status.code = $3)
  AND ($4::varchar IS NULL OR outbound.client_delivery_order_no ILIKE '%' || $4 || '%')
  AND outbound.business_date BETWEEN $5 AND $6
GROUP BY outbound.outbound_id, outbound.client_delivery_order_no,
         outbound.business_date, status.code, customer.code,
         store.code, store.name, outbound.requested_ship_at
ORDER BY outbound.business_date DESC, outbound.outbound_id DESC
LIMIT $7 OFFSET $8;

-- =============================================================================
-- 2. VALIDATION AND RELEASE
-- =============================================================================

-- 2.1 Run configured validation rules and preserve every result.
-- $1 outbound ID, $2 validation-run ID, $3 notes, $4 actor account ID
BEGIN;

SELECT outbound_id FROM outbound_order WHERE outbound_id = $1 FOR UPDATE;

WITH outbound_context AS (
    SELECT outbound.*, dt.document_type_id AS validation_document_type_id,
           pending.status_id AS pending_status_id
    FROM outbound_order outbound
    JOIN document_status outbound_status ON outbound_status.status_id = outbound.status_id
    JOIN document_type dt ON dt.code = 'OUTBOUND_VALIDATION' AND dt.is_active
    JOIN document_status pending
      ON pending.document_type_id = dt.document_type_id
     AND pending.code = 'PENDING' AND pending.is_active
    WHERE outbound.outbound_id = $1 AND outbound_status.code = 'DRAFT'
), created_run AS (
    INSERT INTO outbound_validation_run (
        validation_run_id, outbound_id, document_type_id, status_id,
        notes, created_by
    )
    SELECT $2, outbound_id, validation_document_type_id,
           pending_status_id, $3, $4
    FROM outbound_context
    RETURNING *
), evaluated AS (
    SELECT rule.outbound_validation_rule_id, rule.code,
           CASE rule.handler_code
             WHEN 'REQUIRE_DO_NUMBER' THEN nullif(trim(context.client_delivery_order_no), '') IS NOT NULL
             WHEN 'REQUIRE_SHIP_TO' THEN context.ship_to_partner_id IS NOT NULL
                  AND nullif(trim(context.ship_to_name), '') IS NOT NULL
                  AND nullif(trim(context.ship_to_address_1), '') IS NOT NULL
             WHEN 'REQUIRE_ORDER_LINES' THEN EXISTS (
                  SELECT 1 FROM outbound_order_line line
                  WHERE line.outbound_id = context.outbound_id
             )
             WHEN 'REQUIRE_BASE_UOM' THEN NOT EXISTS (
                  SELECT 1 FROM outbound_order_line line
                  JOIN item ON item.item_id = line.item_id
                  WHERE line.outbound_id = context.outbound_id
                    AND line.uom_id <> item.base_uom_id
             )
             WHEN 'REQUIRE_REQUESTED_SHIP' THEN context.requested_ship_at IS NOT NULL
             ELSE false
           END AS passed
    FROM outbound_context context
    CROSS JOIN outbound_validation_rule rule
    WHERE rule.is_active
), inserted_results AS (
    INSERT INTO outbound_validation_result_detail (
        validation_result_id, validation_run_id,
        outbound_validation_rule_id, passed, result_message
    )
    SELECT $2 || '-R-' || lpad(row_number() OVER (ORDER BY rule.display_order)::text, 3, '0'),
           $2, evaluated.outbound_validation_rule_id, evaluated.passed,
           CASE WHEN evaluated.passed THEN 'Passed'
                ELSE rule.description END
    FROM evaluated
    JOIN outbound_validation_rule rule
      ON rule.outbound_validation_rule_id = evaluated.outbound_validation_rule_id
    RETURNING *
), outcome AS (
    SELECT NOT EXISTS (
        SELECT 1
        FROM inserted_results result
        JOIN outbound_validation_rule rule
          ON rule.outbound_validation_rule_id = result.outbound_validation_rule_id
        JOIN validation_severity severity
          ON severity.validation_severity_id = rule.validation_severity_id
        WHERE NOT result.passed AND severity.blocks_processing
    ) AS passed
), completed_run AS (
    UPDATE outbound_validation_run run
    SET status_id = final_status.status_id,
        validated_at = clock_timestamp(), validated_by = $4
    FROM outcome, document_status final_status
    WHERE run.validation_run_id = $2
      AND final_status.document_type_id = run.document_type_id
      AND final_status.code = CASE WHEN outcome.passed THEN 'PASSED' ELSE 'FAILED' END
    RETURNING run.*, outcome.passed
), validated_outbound AS (
    UPDATE outbound_order outbound
    SET status_id = validated.status_id,
        validated_at = clock_timestamp(), validated_by = $4,
        updated_at = clock_timestamp(), updated_by = $4,
        version_no = outbound.version_no + 1
    FROM completed_run run, document_status validated
    WHERE outbound.outbound_id = run.outbound_id
      AND run.passed
      AND validated.document_type_id = outbound.document_type_id
      AND validated.code = 'VALIDATED'
    RETURNING outbound.outbound_id
)
SELECT completed_run.*,
       EXISTS (SELECT 1 FROM validated_outbound) AS outbound_validated
FROM completed_run;

COMMIT;

-- 2.2 Validation result detail. $1 validation-run ID
SELECT rule.code AS rule_code, rule.name, severity.code AS severity_code,
       result.passed, result.result_message
FROM outbound_validation_result_detail result
JOIN outbound_validation_rule rule
  ON rule.outbound_validation_rule_id = result.outbound_validation_rule_id
JOIN validation_severity severity
  ON severity.validation_severity_id = rule.validation_severity_id
WHERE result.validation_run_id = $1
ORDER BY rule.display_order;

-- 2.3 Release a successfully validated DO. $1 outbound ID, $2 actor ID
UPDATE outbound_order outbound
SET status_id = released.status_id,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = outbound.version_no + 1
FROM document_status current_status, document_status released
WHERE outbound.outbound_id = $1
  AND current_status.status_id = outbound.status_id
  AND current_status.code = 'VALIDATED'
  AND released.document_type_id = outbound.document_type_id
  AND released.code = 'RELEASED' AND released.is_active
RETURNING outbound.*;

-- =============================================================================
-- 3. INVENTORY ALLOCATION
-- =============================================================================

-- 3.1 Candidate balances for one outbound line.
-- $1 outbound-line ID, $2 picking-strategy ID or NULL
SELECT balance.balance_id, location.code AS location_code,
       status.code AS inventory_status_code, lot.lot_number,
       lot.expiry_date, balance.handling_unit_id,
       balance.on_hand_qty - balance.reserved_qty AS available_qty,
       unit.code AS uom_code
FROM outbound_order_line line
JOIN outbound_order outbound ON outbound.outbound_id = line.outbound_id
JOIN inventory_balance balance
  ON balance.owner_id = outbound.owner_id
 AND balance.warehouse_id = outbound.warehouse_id
 AND balance.item_id = line.item_id
 AND balance.uom_id = line.uom_id
JOIN inventory_status status
  ON status.inventory_status_id = balance.inventory_status_id
 AND status.is_active AND status.is_allocatable
JOIN warehouse_location location ON location.location_id = balance.location_id
JOIN uom unit ON unit.uom_id = balance.uom_id
LEFT JOIN inventory_lot lot ON lot.lot_id = balance.lot_id
LEFT JOIN LATERAL (
    SELECT strategy_rule.sequence_no, sort_method.code AS sort_code
    FROM picking_strategy_rule strategy_rule
    JOIN picking_sort_method sort_method
      ON sort_method.picking_sort_method_id = strategy_rule.picking_sort_method_id
     AND sort_method.is_active
    WHERE strategy_rule.picking_strategy_id = $2
      AND strategy_rule.is_active
      AND (strategy_rule.inventory_status_id IS NULL
           OR strategy_rule.inventory_status_id = balance.inventory_status_id)
      AND (strategy_rule.zone_id IS NULL OR strategy_rule.zone_id = location.zone_id)
    ORDER BY strategy_rule.sequence_no
    LIMIT 1
) strategy ON true
WHERE line.outbound_line_id = $1
  AND balance.on_hand_qty > balance.reserved_qty
  AND (line.requested_lot_no IS NULL OR lot.lot_number = line.requested_lot_no)
ORDER BY strategy.sequence_no NULLS LAST,
         CASE WHEN strategy.sort_code = 'FEFO' THEN lot.expiry_date END NULLS LAST,
         CASE WHEN strategy.sort_code = 'FIFO' THEN lot.created_at END NULLS LAST,
         CASE WHEN strategy.sort_code = 'LOCATION' THEN location.pick_sequence END NULLS LAST,
         location.code, balance.balance_id;

-- 3.1a Allocation shortage view for released/partially allocated DOs.
-- $1 owner ID, $2 warehouse ID
SELECT outbound.outbound_id, outbound.client_delivery_order_no,
       line.outbound_line_id, line.line_no, item.code AS item_code,
       line.ordered_qty, line.allocated_qty,
       line.ordered_qty - line.allocated_qty AS remaining_to_allocate,
       COALESCE(available.available_qty, 0) AS currently_available_qty,
       greatest(line.ordered_qty - line.allocated_qty
                - COALESCE(available.available_qty, 0), 0) AS shortage_qty,
       unit.code AS uom_code
FROM outbound_order outbound
JOIN document_status outbound_status ON outbound_status.status_id = outbound.status_id
JOIN outbound_order_line line ON line.outbound_id = outbound.outbound_id
JOIN item ON item.item_id = line.item_id
JOIN uom unit ON unit.uom_id = line.uom_id
LEFT JOIN LATERAL (
    SELECT sum(balance.on_hand_qty - balance.reserved_qty) AS available_qty
    FROM inventory_balance balance
    JOIN inventory_status status
      ON status.inventory_status_id = balance.inventory_status_id
     AND status.is_active AND status.is_allocatable
    LEFT JOIN inventory_lot lot ON lot.lot_id = balance.lot_id
    WHERE balance.owner_id = outbound.owner_id
      AND balance.warehouse_id = outbound.warehouse_id
      AND balance.item_id = line.item_id AND balance.uom_id = line.uom_id
      AND balance.on_hand_qty > balance.reserved_qty
      AND (line.requested_lot_no IS NULL OR lot.lot_number = line.requested_lot_no)
) available ON true
WHERE outbound.owner_id = $1 AND outbound.warehouse_id = $2
  AND outbound_status.code IN ('RELEASED', 'PARTIALLY_ALLOCATED')
  AND line.allocated_qty < line.ordered_qty
ORDER BY outbound.requested_ship_at NULLS LAST,
         outbound.client_delivery_order_no, line.line_no;

-- 3.2 Allocate one balance atomically.
-- $1 outbound-line ID, $2 balance ID, $3 base quantity,
-- $4 reservation ID, $5 picking-strategy ID or NULL, $6 actor account ID
BEGIN;

SELECT balance_id FROM inventory_balance WHERE balance_id = $2 FOR UPDATE;

WITH context AS (
    SELECT line.*, outbound.owner_id, outbound.warehouse_id,
           outbound.document_type_id AS outbound_document_type_id,
           balance.on_hand_qty, balance.reserved_qty AS balance_reserved_qty,
           dt.document_type_id AS reservation_document_type_id,
           active_status.status_id AS reservation_status_id,
           strategy.picking_strategy_id
    FROM outbound_order_line line
    JOIN outbound_order outbound ON outbound.outbound_id = line.outbound_id
    JOIN document_status outbound_status ON outbound_status.status_id = outbound.status_id
    JOIN inventory_balance balance
      ON balance.balance_id = $2
     AND balance.owner_id = outbound.owner_id
     AND balance.warehouse_id = outbound.warehouse_id
     AND balance.item_id = line.item_id
     AND balance.uom_id = line.uom_id
    JOIN inventory_status inventory_status
      ON inventory_status.inventory_status_id = balance.inventory_status_id
     AND inventory_status.is_active AND inventory_status.is_allocatable
    JOIN document_type dt ON dt.code = 'RESERVATION' AND dt.is_active
    JOIN document_status active_status
      ON active_status.document_type_id = dt.document_type_id
     AND active_status.code = 'ACTIVE' AND active_status.is_active
    LEFT JOIN picking_strategy strategy
      ON strategy.picking_strategy_id = $5 AND strategy.is_active
     AND (strategy.owner_id IS NULL OR strategy.owner_id = outbound.owner_id)
     AND (strategy.warehouse_id IS NULL OR strategy.warehouse_id = outbound.warehouse_id)
    LEFT JOIN inventory_lot lot ON lot.lot_id = balance.lot_id
    WHERE line.outbound_line_id = $1
      AND outbound_status.code IN ('RELEASED', 'PARTIALLY_ALLOCATED')
      AND line.allocated_qty + $3 <= line.ordered_qty
      AND balance.on_hand_qty - balance.reserved_qty >= $3
      AND (line.requested_lot_no IS NULL OR lot.lot_number = line.requested_lot_no)
      AND ($5::uuid IS NULL OR strategy.picking_strategy_id IS NOT NULL)
      AND $3 > 0
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
        outbound_line_id,
        balance_id, reserved_qty, picked_qty, uom_id, created_by
    )
    SELECT $4, reservation_document_type_id, reservation_status_id,
           picking_strategy_id, outbound_line_id, $2, $3, 0, uom_id, $6
    FROM context CROSS JOIN reserved_balance
    RETURNING *
), updated_line AS (
    UPDATE outbound_order_line line
    SET allocated_qty = line.allocated_qty + $3
    FROM reservation
    WHERE line.outbound_line_id = reservation.outbound_line_id
    RETURNING line.*
), updated_outbound AS (
    UPDATE outbound_order outbound
    SET status_id = next_status.status_id,
        allocated_at = CASE WHEN next_status.code = 'ALLOCATED'
                            THEN clock_timestamp() ELSE outbound.allocated_at END,
        updated_at = clock_timestamp(), updated_by = $6,
        version_no = outbound.version_no + 1
    FROM updated_line current_line, document_status next_status
    WHERE outbound.outbound_id = current_line.outbound_id
      AND next_status.document_type_id = outbound.document_type_id
      AND next_status.code = CASE
          WHEN current_line.allocated_qty = current_line.ordered_qty
           AND NOT EXISTS (
               SELECT 1 FROM outbound_order_line other
               WHERE other.outbound_id = current_line.outbound_id
                 AND other.outbound_line_id <> current_line.outbound_line_id
                 AND other.allocated_qty < other.ordered_qty
           ) THEN 'ALLOCATED' ELSE 'PARTIALLY_ALLOCATED' END
    RETURNING outbound.*
)
SELECT reservation.*, updated_outbound.status_id AS outbound_status_id
FROM reservation CROSS JOIN updated_outbound;

COMMIT;

-- 3.3 Release the unpicked remainder of a reservation.
-- $1 reservation ID, $2 actor ID
BEGIN;

SELECT balance.balance_id
FROM inventory_reservation reservation
JOIN inventory_balance balance ON balance.balance_id = reservation.balance_id
WHERE reservation.reservation_id = $1
FOR UPDATE OF balance;

WITH context AS (
    SELECT reservation.*, line.outbound_id,
           reservation.reserved_qty - reservation.picked_qty AS release_qty
    FROM inventory_reservation reservation
    JOIN document_status status ON status.status_id = reservation.status_id
    JOIN outbound_order_line line ON line.outbound_line_id = reservation.outbound_line_id
    WHERE reservation.reservation_id = $1
      AND status.code IN ('ACTIVE', 'PARTIALLY_PICKED')
      AND reservation.reserved_qty > reservation.picked_qty
), released_balance AS (
    UPDATE inventory_balance balance
    SET reserved_qty = balance.reserved_qty - context.release_qty,
        version_no = balance.version_no + 1,
        updated_at = clock_timestamp()
    FROM context
    WHERE balance.balance_id = context.balance_id
      AND balance.reserved_qty >= context.release_qty
    RETURNING balance.balance_id
), released_reservation AS (
    UPDATE inventory_reservation reservation
    SET status_id = released.status_id, released_at = clock_timestamp()
    FROM context, released_balance, document_status released
    WHERE reservation.reservation_id = context.reservation_id
      AND released.document_type_id = reservation.document_type_id
      AND released.code = 'RELEASED'
    RETURNING reservation.*, context.release_qty, context.outbound_id
), updated_line AS (
    UPDATE outbound_order_line line
    SET allocated_qty = line.allocated_qty - released.release_qty
    FROM released_reservation released
    WHERE line.outbound_line_id = released.outbound_line_id
    RETURNING line.*
), updated_outbound AS (
    UPDATE outbound_order outbound
    SET status_id = next_status.status_id,
        updated_at = clock_timestamp(), updated_by = $2,
        version_no = outbound.version_no + 1
    FROM updated_line current_line, document_status next_status
    WHERE outbound.outbound_id = current_line.outbound_id
      AND next_status.document_type_id = outbound.document_type_id
      AND next_status.code = CASE WHEN current_line.allocated_qty = 0
          AND NOT EXISTS (
              SELECT 1 FROM outbound_order_line other
              WHERE other.outbound_id = current_line.outbound_id
                AND other.outbound_line_id <> current_line.outbound_line_id
                AND other.allocated_qty > 0
          ) THEN 'RELEASED' ELSE 'PARTIALLY_ALLOCATED' END
    RETURNING outbound.outbound_id
)
SELECT released_reservation.*
FROM released_reservation CROSS JOIN updated_line CROSS JOIN updated_outbound;

COMMIT;

-- =============================================================================
-- 4. OUTBOUND WAVE AND PICK-TASK CREATION
-- =============================================================================

-- 4.1 Create a draft wave.
-- $1 owner ID, $2 warehouse code, $3 business date, $4 wave-type code,
-- $5 picking-strategy ID or NULL, $6 planned release time or NULL,
-- $7 notes, $8 actor ID
WITH context AS (
    SELECT owner.organization_id AS owner_id, warehouse.warehouse_id,
           warehouse.code AS warehouse_code, wave_type.outbound_wave_type_id,
           strategy.picking_strategy_id, dt.document_type_id,
           initial_status.status_id
    FROM organization owner
    JOIN warehouse_owner scope ON scope.owner_id = owner.organization_id AND scope.is_active
    JOIN warehouse ON warehouse.warehouse_id = scope.warehouse_id
                   AND warehouse.code = $2 AND warehouse.is_active
    JOIN outbound_wave_type wave_type ON wave_type.code = $4 AND wave_type.is_active
    LEFT JOIN picking_strategy strategy
      ON strategy.picking_strategy_id = $5 AND strategy.is_active
     AND (strategy.owner_id IS NULL OR strategy.owner_id = owner.organization_id)
     AND (strategy.warehouse_id IS NULL OR strategy.warehouse_id = warehouse.warehouse_id)
    JOIN document_type dt ON dt.code = 'OUTBOUND_WAVE' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE owner.organization_id = $1
      AND ($5::uuid IS NULL OR strategy.picking_strategy_id IS NOT NULL)
), numbered AS (
    SELECT context.*,
           generate_document_id('OUTBOUND_WAVE', NULL, warehouse_code, $3) AS generated_id
    FROM context
)
INSERT INTO outbound_wave (
    wave_id, document_type_id, status_id, outbound_wave_type_id,
    owner_id, warehouse_id, picking_strategy_id, business_date,
    planned_release_at, notes, created_by, updated_by
)
SELECT generated_id, document_type_id, status_id, outbound_wave_type_id,
       owner_id, warehouse_id, picking_strategy_id, $3, $6, $7, $8, $8
FROM numbered
RETURNING *;

-- 4.2 Add a fully allocated DO to a draft wave.
-- $1 wave ID, $2 outbound ID, $3 actor ID
INSERT INTO outbound_wave_order (wave_id, outbound_id, added_by)
SELECT wave.wave_id, outbound.outbound_id, $3
FROM outbound_wave wave
JOIN document_status wave_status ON wave_status.status_id = wave.status_id
JOIN outbound_order outbound
  ON outbound.outbound_id = $2
 AND outbound.owner_id = wave.owner_id
 AND outbound.warehouse_id = wave.warehouse_id
JOIN document_status outbound_status ON outbound_status.status_id = outbound.status_id
WHERE wave.wave_id = $1
  AND wave_status.code = 'DRAFT'
  AND outbound_status.code = 'ALLOCATED'
  AND NOT EXISTS (
      SELECT 1
      FROM outbound_wave_order existing
      JOIN outbound_wave existing_wave ON existing_wave.wave_id = existing.wave_id
      JOIN document_status existing_status ON existing_status.status_id = existing_wave.status_id
      WHERE existing.outbound_id = outbound.outbound_id
        AND NOT existing_status.is_final
  )
RETURNING *;

-- 4.3 Release a wave and generate one pick task per active reservation.
-- $1 wave ID, $2 staging location ID, $3 task-priority code, $4 actor ID
WITH wave_context AS (
    SELECT wave.*, staging.location_id AS staging_location_id,
           task_type.task_type_id, open_status.task_status_id,
           priority.task_priority_id
    FROM outbound_wave wave
    JOIN document_status status ON status.status_id = wave.status_id
    JOIN warehouse_location staging
      ON staging.location_id = $2
     AND staging.warehouse_id = wave.warehouse_id AND staging.is_active
    JOIN location_type staging_type
      ON staging_type.location_type_id = staging.location_type_id
     AND staging_type.code = 'STAGING' AND staging_type.is_active
    JOIN task_type ON task_type.code = 'PICK' AND task_type.is_active
    JOIN task_status open_status ON open_status.code = 'OPEN' AND open_status.is_active
    JOIN task_priority priority ON priority.code = $3 AND priority.is_active
    WHERE wave.wave_id = $1 AND status.code = 'DRAFT'
), task_source AS (
    SELECT context.*, reservation.reservation_id,
           reservation.outbound_line_id, reservation.balance_id,
           reservation.reserved_qty - reservation.picked_qty AS planned_qty,
           reservation.uom_id, balance.location_id AS source_location_id,
           row_number() OVER (ORDER BY order_link.outbound_id,
                                      line.line_no, reservation.reservation_id) AS task_no
    FROM wave_context context
    JOIN outbound_wave_order order_link ON order_link.wave_id = context.wave_id
    JOIN outbound_order_line line ON line.outbound_id = order_link.outbound_id
    JOIN inventory_reservation reservation ON reservation.outbound_line_id = line.outbound_line_id
    JOIN document_status reservation_status ON reservation_status.status_id = reservation.status_id
    JOIN inventory_balance balance ON balance.balance_id = reservation.balance_id
    WHERE reservation_status.code IN ('ACTIVE', 'PARTIALLY_PICKED')
      AND reservation.reserved_qty > reservation.picked_qty
      AND balance.location_id <> context.staging_location_id
      AND (context.picking_strategy_id IS NULL
           OR reservation.picking_strategy_id = context.picking_strategy_id)
), created_tasks AS (
    INSERT INTO pick_task (
        pick_task_id, task_type_id, task_status_id, task_priority_id,
        wave_id, reservation_id, outbound_line_id,
        source_location_id, target_location_id,
        planned_qty, picked_qty, uom_id, created_by
    )
    SELECT wave_id || '-P-' || lpad(task_no::text, 6, '0'),
           task_type_id, task_status_id, task_priority_id, wave_id,
           reservation_id, outbound_line_id, source_location_id,
           staging_location_id, planned_qty, 0, uom_id, $4
    FROM task_source
    RETURNING *
), released_wave AS (
    UPDATE outbound_wave wave
    SET status_id = released.status_id, released_at = clock_timestamp(),
        updated_at = clock_timestamp(), updated_by = $4,
        version_no = wave.version_no + 1
    FROM wave_context context, document_status released
    WHERE wave.wave_id = context.wave_id
      AND released.document_type_id = wave.document_type_id
      AND released.code = 'RELEASED'
      AND EXISTS (SELECT 1 FROM created_tasks)
    RETURNING wave.*
), waved_orders AS (
    UPDATE outbound_order outbound
    SET status_id = waved.status_id,
        updated_at = clock_timestamp(), updated_by = $4,
        version_no = outbound.version_no + 1
    FROM outbound_wave_order link, released_wave wave, document_status waved
    WHERE link.wave_id = wave.wave_id
      AND outbound.outbound_id = link.outbound_id
      AND waved.document_type_id = outbound.document_type_id
      AND waved.code = 'WAVED'
    RETURNING outbound.outbound_id
)
SELECT released_wave.wave_id,
       (SELECT count(*) FROM created_tasks) AS created_pick_tasks,
       (SELECT count(*) FROM waved_orders) AS waved_orders
FROM released_wave;

-- 4.4 Wave/task detail. $1 wave ID
SELECT wave.wave_id, wave_status.code AS wave_status_code,
       outbound.client_delivery_order_no, line.line_no,
       item.code AS item_code, task.pick_task_id,
       task_status.code AS task_status_code,
       source.code AS source_location_code,
       target.code AS staging_location_code,
       task.planned_qty, task.picked_qty, task.short_qty,
       unit.code AS uom_code, task.assigned_to
FROM outbound_wave wave
JOIN document_status wave_status ON wave_status.status_id = wave.status_id
JOIN outbound_wave_order wave_order ON wave_order.wave_id = wave.wave_id
JOIN outbound_order outbound ON outbound.outbound_id = wave_order.outbound_id
JOIN outbound_order_line line ON line.outbound_id = outbound.outbound_id
JOIN item ON item.item_id = line.item_id
JOIN pick_task task ON task.outbound_line_id = line.outbound_line_id AND task.wave_id = wave.wave_id
JOIN task_status ON task_status.task_status_id = task.task_status_id
JOIN warehouse_location source ON source.location_id = task.source_location_id
LEFT JOIN warehouse_location target ON target.location_id = task.target_location_id
JOIN uom unit ON unit.uom_id = task.uom_id
WHERE wave.wave_id = $1
ORDER BY outbound.client_delivery_order_no, line.line_no, task.pick_task_id;

-- 4.5 Cancel an empty draft wave. $1 wave ID, $2 actor ID
UPDATE outbound_wave wave
SET status_id = cancelled.status_id,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = wave.version_no + 1
FROM document_status current_status, document_status cancelled
WHERE wave.wave_id = $1
  AND current_status.status_id = wave.status_id AND current_status.code = 'DRAFT'
  AND cancelled.document_type_id = wave.document_type_id
  AND cancelled.code = 'CANCELLED'
  AND NOT EXISTS (SELECT 1 FROM pick_task task WHERE task.wave_id = wave.wave_id)
RETURNING wave.*;

-- 4.6 Cancel a DO only when it has no active reservation or wave.
-- Release reservations first using operation 3.3.
-- $1 outbound ID, $2 actor ID
UPDATE outbound_order outbound
SET status_id = cancelled.status_id,
    completed_at = clock_timestamp(),
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = outbound.version_no + 1
FROM document_status current_status, document_status cancelled
WHERE outbound.outbound_id = $1
  AND current_status.status_id = outbound.status_id
  AND current_status.code IN ('DRAFT', 'VALIDATED', 'RELEASED', 'PARTIALLY_ALLOCATED', 'ALLOCATED')
  AND cancelled.document_type_id = outbound.document_type_id
  AND cancelled.code = 'CANCELLED'
  AND NOT EXISTS (
      SELECT 1 FROM outbound_order_line line
      JOIN inventory_reservation reservation ON reservation.outbound_line_id = line.outbound_line_id
      JOIN document_status reservation_status ON reservation_status.status_id = reservation.status_id
      WHERE line.outbound_id = outbound.outbound_id
        AND NOT reservation_status.is_final
  )
  AND NOT EXISTS (
      SELECT 1 FROM outbound_wave_order wave_order
      JOIN outbound_wave wave ON wave.wave_id = wave_order.wave_id
      JOIN document_status wave_status ON wave_status.status_id = wave.status_id
      WHERE wave_order.outbound_id = outbound.outbound_id
        AND NOT wave_status.is_final
  )
RETURNING outbound.*;
