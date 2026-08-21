-- Multi-order shipment, driver assignment, partial delivery/POD, retry,
-- and safe return-to-depot operations. Execute one operation at a time.
SET search_path TO wms, public;

-- =============================================================================
-- 1. SHIPMENT MANIFEST
-- =============================================================================

-- 1.1 Create planned manifest.
-- $1 owner, $2 warehouse code, $3 business date, $4 carrier service/null,
-- $5 route/null, $6 tracking, $7 vehicle, $8 seal, $9 notes, $10 actor
WITH context AS (
    SELECT owner.organization_id AS owner_id, warehouse.warehouse_id,
           warehouse.code AS warehouse_code, dt.document_type_id,
           initial.status_id, service.carrier_service_id
    FROM organization owner
    JOIN warehouse_owner scope ON scope.owner_id = owner.organization_id AND scope.is_active
    JOIN warehouse ON warehouse.warehouse_id = scope.warehouse_id
                  AND warehouse.code = $2 AND warehouse.is_active
    JOIN document_type dt ON dt.code = 'SHIPMENT' AND dt.is_active
    JOIN document_status initial ON initial.document_type_id = dt.document_type_id
                                AND initial.is_initial AND initial.is_active
    LEFT JOIN carrier_service service ON service.carrier_service_id = $4 AND service.is_active
    LEFT JOIN carrier ON carrier.carrier_id = service.carrier_id AND carrier.is_active
    WHERE owner.organization_id = $1
      AND ($4::uuid IS NULL OR carrier.carrier_id IS NOT NULL)
), numbered AS (
    SELECT context.*,
           generate_document_id('SHIPMENT', NULL, warehouse_code, $3) AS generated_id
    FROM context
)
INSERT INTO shipment (
    shipment_id, document_type_id, status_id, owner_id, carrier_service_id,
    warehouse_id, business_date, route_reference, tracking_number,
    vehicle_number, seal_number, notes, created_by
)
SELECT generated_id, document_type_id, status_id, owner_id, carrier_service_id,
       warehouse_id, $3, $5, $6, $7, $8, $9, $10
FROM numbered RETURNING *;

-- 1.2 Add a packed DO. $1 shipment, $2 outbound, $3 actor
INSERT INTO shipment_order (shipment_id, outbound_id, added_by)
SELECT shipment.shipment_id, outbound.outbound_id, $3
FROM shipment
JOIN document_status shipment_status ON shipment_status.status_id = shipment.status_id
JOIN outbound_order outbound ON outbound.outbound_id = $2
 AND outbound.owner_id = shipment.owner_id AND outbound.warehouse_id = shipment.warehouse_id
JOIN document_status outbound_status ON outbound_status.status_id = outbound.status_id
JOIN packing ON packing.outbound_id = outbound.outbound_id
JOIN document_status packing_status ON packing_status.status_id = packing.status_id
WHERE shipment.shipment_id = $1 AND shipment_status.code = 'PLANNED'
  AND outbound_status.code = 'PACKED' AND packing_status.code = 'COMPLETED'
  AND NOT EXISTS (
      SELECT 1 FROM shipment_order existing
      JOIN shipment old_shipment ON old_shipment.shipment_id = existing.shipment_id
      JOIN document_status old_status ON old_status.status_id = old_shipment.status_id
      WHERE existing.outbound_id = outbound.outbound_id AND NOT old_status.is_cancelled
  )
RETURNING *;

-- 1.3 Assign driver; a new primary replaces the prior primary.
-- $1 shipment, $2 driver, $3 primary boolean, $4 actor
BEGIN;
WITH valid AS (
    SELECT shipment.shipment_id
    FROM shipment
    JOIN document_status status ON status.status_id = shipment.status_id AND status.code = 'PLANNED'
    JOIN carrier_service service ON service.carrier_service_id = shipment.carrier_service_id
    JOIN carrier_driver driver ON driver.driver_id = $2
      AND driver.carrier_id = service.carrier_id AND driver.is_active
    WHERE shipment.shipment_id = $1
)
UPDATE shipment_driver assignment SET is_primary = false
FROM valid WHERE assignment.shipment_id = valid.shipment_id
  AND assignment.is_primary AND $3;

INSERT INTO shipment_driver (shipment_id, driver_id, is_primary, assigned_by)
SELECT shipment.shipment_id, driver.driver_id, $3, $4
FROM shipment
JOIN document_status status ON status.status_id = shipment.status_id AND status.code = 'PLANNED'
JOIN carrier_service service ON service.carrier_service_id = shipment.carrier_service_id
JOIN carrier_driver driver ON driver.driver_id = $2
  AND driver.carrier_id = service.carrier_id AND driver.is_active
WHERE shipment.shipment_id = $1
ON CONFLICT (shipment_id, driver_id) DO UPDATE SET
    is_primary = EXCLUDED.is_primary, assigned_at = clock_timestamp(),
    assigned_by = EXCLUDED.assigned_by
RETURNING *;
COMMIT;

-- 1.4 Attach completed packing for a manifest DO. $1 shipment, $2 packing
INSERT INTO shipment_packing (shipment_id, packing_id)
SELECT shipment.shipment_id, packing.packing_id
FROM shipment
JOIN document_status shipment_status ON shipment_status.status_id = shipment.status_id
JOIN packing ON packing.packing_id = $2
JOIN shipment_order manifest ON manifest.shipment_id = shipment.shipment_id
                            AND manifest.outbound_id = packing.outbound_id
                            AND manifest.removed_at IS NULL
JOIN document_status packing_status ON packing_status.status_id = packing.status_id
WHERE shipment.shipment_id = $1 AND shipment_status.code = 'PLANNED'
  AND packing_status.code = 'COMPLETED'
  AND NOT EXISTS (SELECT 1 FROM shipment_packing x WHERE x.packing_id = packing.packing_id)
RETURNING *;

-- 1.5 Dispatch one packed line. $1 shipment, $2 packing line,
-- $3 movement ID, $4 actor
BEGIN;
SELECT balance.balance_id
FROM packing_line line JOIN inventory_balance balance ON balance.balance_id = line.packing_balance_id
WHERE line.packing_line_id = $2 FOR UPDATE OF balance;

WITH context AS (
    SELECT shipment.shipment_id, shipment.owner_id, shipment.warehouse_id,
           shipment.business_date, packing.outbound_id, line.packing_line_id,
           line.packing_balance_id, line.pick_task_id, line.packed_qty, line.uom_id,
           balance.location_id, balance.item_id, balance.lot_id,
           balance.handling_unit_id, balance.inventory_status_id
    FROM shipment
    JOIN document_status status ON status.status_id = shipment.status_id AND status.code = 'PLANNED'
    JOIN shipment_packing link ON link.shipment_id = shipment.shipment_id
    JOIN packing ON packing.packing_id = link.packing_id
    JOIN shipment_order manifest ON manifest.shipment_id = shipment.shipment_id
                                AND manifest.outbound_id = packing.outbound_id
                                AND manifest.removed_at IS NULL
    JOIN packing_line line ON line.packing_id = packing.packing_id AND line.packing_line_id = $2
    JOIN inventory_balance balance ON balance.balance_id = line.packing_balance_id
    WHERE shipment.shipment_id = $1 AND balance.on_hand_qty >= line.packed_qty
      AND NOT EXISTS (SELECT 1 FROM shipment_line x WHERE x.packing_line_id = line.packing_line_id)
), source_update AS (
    UPDATE inventory_balance balance
    SET on_hand_qty = balance.on_hand_qty - context.packed_qty,
        version_no = balance.version_no + 1, updated_at = clock_timestamp()
    FROM context WHERE balance.balance_id = context.packing_balance_id
      AND balance.on_hand_qty >= context.packed_qty RETURNING balance.balance_id
), movement AS (
    INSERT INTO inventory_movement (
        movement_id, movement_type_id, owner_id, warehouse_id, business_date,
        item_id, lot_id, handling_unit_id, from_location_id, to_location_id,
        from_status_id, to_status_id, quantity, uom_id,
        source_document_id, source_line_id, created_by
    )
    SELECT $3, type.movement_type_id, owner_id, warehouse_id, business_date,
           item_id, lot_id, handling_unit_id, location_id, NULL,
           inventory_status_id, NULL, packed_qty, uom_id,
           shipment_id, packing_line_id, $4
    FROM context CROSS JOIN source_update
    JOIN movement_type type ON type.code = 'SHIP' AND type.is_active
    RETURNING movement_id
), shipped AS (
    INSERT INTO shipment_line (
        shipment_line_id, shipment_id, packing_line_id, source_balance_id,
        shipped_qty, uom_id, movement_id, created_by
    )
    SELECT shipment_id || '-L-' || right(packing_line_id, 6), shipment_id,
           packing_line_id, packing_balance_id, packed_qty, uom_id,
           movement_id, $4 FROM context CROSS JOIN movement RETURNING *
), order_line_update AS (
    UPDATE outbound_order_line order_line
    SET shipped_qty = order_line.shipped_qty + shipped.shipped_qty
    FROM shipped JOIN packing_line packed ON packed.packing_line_id = shipped.packing_line_id
    JOIN pick_task task ON task.pick_task_id = packed.pick_task_id
    WHERE order_line.outbound_line_id = task.outbound_line_id RETURNING order_line.*
), hu_update AS (
    UPDATE handling_unit hu SET current_location_id = NULL
    FROM context, shipped WHERE hu.handling_unit_id = context.handling_unit_id RETURNING hu.handling_unit_id
)
SELECT shipped.* FROM shipped CROSS JOIN order_line_update;
COMMIT;

-- 1.6 Confirm manifest after all attached packing lines shipped.
-- $1 shipment, $2 actor
WITH shipped AS (
    UPDATE shipment
    SET status_id = final.status_id, shipped_at = clock_timestamp(), dispatched_by = $2
    FROM document_status current, document_status final
    WHERE shipment.shipment_id = $1 AND current.status_id = shipment.status_id
      AND current.code = 'PLANNED' AND final.document_type_id = shipment.document_type_id
      AND final.code = 'SHIPPED'
      AND EXISTS (SELECT 1 FROM shipment_order x WHERE x.shipment_id = shipment.shipment_id
                  AND x.removed_at IS NULL)
      AND EXISTS (SELECT 1 FROM shipment_packing x WHERE x.shipment_id = shipment.shipment_id)
      AND NOT EXISTS (
          SELECT 1 FROM shipment_packing link
          JOIN packing_line packed ON packed.packing_id = link.packing_id
          WHERE link.shipment_id = shipment.shipment_id
            AND NOT EXISTS (SELECT 1 FROM shipment_line line
                            WHERE line.shipment_id = shipment.shipment_id
                              AND line.packing_line_id = packed.packing_line_id)
      ) RETURNING shipment.*
), orders AS (
    UPDATE outbound_order outbound
    SET status_id = final.status_id, updated_at = clock_timestamp(),
        updated_by = $2, version_no = outbound.version_no + 1
    FROM shipped
    JOIN shipment_order manifest ON manifest.shipment_id = shipped.shipment_id
                                AND manifest.removed_at IS NULL,
         document_status final
    WHERE outbound.outbound_id = manifest.outbound_id
      AND final.document_type_id = outbound.document_type_id
      AND final.code = 'SHIPPED' RETURNING outbound.outbound_id
)
SELECT shipped.*, (SELECT count(*) FROM orders) AS outbound_count FROM shipped;

-- 1.7 Manifest trace. $1 shipment
SELECT shipment.shipment_id, status.code AS shipment_status_code,
       outbound.client_delivery_order_no, item.code AS item_code, lot.lot_number,
       line.shipped_qty, unit.code AS uom_code, line.movement_id,
       shipment.tracking_number, shipment.vehicle_number, shipment.route_reference,
       driver.name AS primary_driver_name, shipment.shipped_at
FROM shipment
JOIN document_status status ON status.status_id = shipment.status_id
JOIN shipment_line line ON line.shipment_id = shipment.shipment_id
JOIN packing_line packed ON packed.packing_line_id = line.packing_line_id
JOIN packing ON packing.packing_id = packed.packing_id
JOIN outbound_order outbound ON outbound.outbound_id = packing.outbound_id
JOIN inventory_balance source ON source.balance_id = line.source_balance_id
JOIN item ON item.item_id = source.item_id
LEFT JOIN inventory_lot lot ON lot.lot_id = source.lot_id
JOIN uom unit ON unit.uom_id = line.uom_id
LEFT JOIN shipment_driver assignment ON assignment.shipment_id = shipment.shipment_id AND assignment.is_primary
LEFT JOIN carrier_driver driver ON driver.driver_id = assignment.driver_id
WHERE shipment.shipment_id = $1
ORDER BY outbound.client_delivery_order_no, line.shipment_line_id;

-- =============================================================================
-- 2. DELIVERY AND PARTIAL POD
-- =============================================================================

-- 2.1 Create delivery for one manifest DO and copy its shipment lines.
-- $1 shipment, $2 outbound, $3 business date, $4 plan time, $5 notes, $6 actor
WITH context AS (
    SELECT shipment.shipment_id, outbound.outbound_id, warehouse.code AS warehouse_code,
           dt.document_type_id, initial.status_id
    FROM shipment
    JOIN document_status shipment_status ON shipment_status.status_id = shipment.status_id
    JOIN shipment_order manifest ON manifest.shipment_id = shipment.shipment_id
                                AND manifest.outbound_id = $2
                                AND manifest.removed_at IS NULL
    JOIN outbound_order outbound ON outbound.outbound_id = manifest.outbound_id
    JOIN warehouse ON warehouse.warehouse_id = shipment.warehouse_id
    JOIN document_type dt ON dt.code = 'DELIVERY' AND dt.is_active
    JOIN document_status initial ON initial.document_type_id = dt.document_type_id
                                AND initial.is_initial AND initial.is_active
    WHERE shipment.shipment_id = $1 AND shipment_status.code = 'SHIPPED'
      AND NOT EXISTS (SELECT 1 FROM delivery old JOIN document_status s ON s.status_id = old.status_id
                      WHERE old.shipment_id = shipment.shipment_id
                        AND old.outbound_id = outbound.outbound_id AND NOT s.is_cancelled)
), numbered AS (
    SELECT context.*, generate_document_id('DELIVERY', NULL, warehouse_code, $3) AS generated_id
    FROM context
), created AS (
    INSERT INTO delivery (
        delivery_id, document_type_id, status_id, shipment_id, outbound_id,
        business_date, planned_delivery_at, notes, created_by, updated_by
    )
    SELECT generated_id, document_type_id, status_id, shipment_id, outbound_id,
           $3, $4, $5, $6, $6 FROM numbered RETURNING *
), lines AS (
    INSERT INTO delivery_line (
        delivery_line_id, delivery_id, shipment_line_id, planned_qty,
        delivered_qty, returned_qty, uom_id, created_by
    )
    SELECT created.delivery_id || '-L-' || lpad(row_number() OVER (
               ORDER BY shipped.shipment_line_id)::text, 6, '0'),
           created.delivery_id, shipped.shipment_line_id, shipped.shipped_qty,
           0, 0, shipped.uom_id, $6
    FROM created JOIN shipment_line shipped ON shipped.shipment_id = created.shipment_id
    JOIN packing_line packed ON packed.packing_line_id = shipped.packing_line_id
    JOIN packing ON packing.packing_id = packed.packing_id AND packing.outbound_id = created.outbound_id
    RETURNING *
)
SELECT created.*, (SELECT count(*) FROM lines) AS line_count FROM created;

-- 2.2 Departure/arrival. $1 delivery, $2 event, $3 event code,
-- $4 time, $5 lat/null, $6 long/null, $7 notes, $8 actor
WITH context AS (
    SELECT delivery.*, type.delivery_event_type_id, type.code AS event_code
    FROM delivery JOIN document_status status ON status.status_id = delivery.status_id
    JOIN delivery_event_type type ON type.code = $3 AND type.is_active
    WHERE delivery.delivery_id = $1 AND type.code IN ('DEPARTED','ARRIVED_STORE')
      AND ((type.code = 'DEPARTED' AND status.code IN ('PLANNED','FAILED'))
        OR (type.code = 'ARRIVED_STORE' AND status.code = 'IN_TRANSIT'))
), event AS (
    INSERT INTO delivery_event (delivery_event_id, delivery_id, delivery_event_type_id,
        event_at, notes, latitude, longitude, recorded_by)
    SELECT $2, delivery_id, delivery_event_type_id, $4, $7, $5, $6, $8
    FROM context RETURNING *
), changed AS (
    UPDATE delivery SET status_id = next.status_id,
        arrived_at = CASE WHEN context.event_code = 'ARRIVED_STORE' THEN $4 ELSE delivery.arrived_at END,
        updated_at = clock_timestamp(), updated_by = $8
    FROM context, event, document_status next
    WHERE delivery.delivery_id = context.delivery_id
      AND next.document_type_id = delivery.document_type_id
      AND next.code = CASE WHEN context.event_code = 'DEPARTED' THEN 'IN_TRANSIT' ELSE 'ARRIVED' END
    RETURNING delivery.*
)
SELECT event.*, changed.status_id FROM event CROSS JOIN changed;

-- 2.3 Accept quantity with POD. $1 delivery, $2 delivery line, $3 qty,
-- $4 event, $5 time, $6 recipient, $7 recipient ref/null, $8 proof ref,
-- $9 proof URI/null, $10 lat/null, $11 long/null, $12 notes, $13 actor
WITH context AS (
    SELECT delivery.*, line.delivery_line_id, type.delivery_event_type_id
    FROM delivery JOIN document_status status ON status.status_id = delivery.status_id
    JOIN delivery_line line ON line.delivery_line_id = $2 AND line.delivery_id = delivery.delivery_id
    JOIN delivery_event_type type ON type.code = 'DELIVERED' AND type.is_active
    WHERE delivery.delivery_id = $1 AND status.code = 'ARRIVED' AND $3 > 0
      AND line.delivered_qty + line.returned_qty + $3 <= line.planned_qty
      AND nullif(trim($6),'') IS NOT NULL AND nullif(trim($8),'') IS NOT NULL
), event AS (
    INSERT INTO delivery_event (
        delivery_event_id, delivery_id, delivery_event_type_id, event_at,
        recipient_name, recipient_reference, proof_reference, proof_uri,
        notes, latitude, longitude, recorded_by
    ) SELECT $4, delivery_id, delivery_event_type_id, $5, $6, $7, $8, $9,
             $12, $10, $11, $13 FROM context RETURNING *
), event_line AS (
    INSERT INTO delivery_event_line (delivery_event_id, delivery_line_id, delivered_qty)
    SELECT event.delivery_event_id, context.delivery_line_id, $3
    FROM context CROSS JOIN event RETURNING *
), line_update AS (
    UPDATE delivery_line line SET delivered_qty = line.delivered_qty + event_line.delivered_qty
    FROM event_line WHERE line.delivery_line_id = event_line.delivery_line_id RETURNING line.*
), order_line_update AS (
    UPDATE outbound_order_line order_line
    SET delivered_qty = order_line.delivered_qty + event_line.delivered_qty
    FROM event_line JOIN delivery_line dl ON dl.delivery_line_id = event_line.delivery_line_id
    JOIN shipment_line sl ON sl.shipment_line_id = dl.shipment_line_id
    JOIN packing_line pl ON pl.packing_line_id = sl.packing_line_id
    JOIN pick_task task ON task.pick_task_id = pl.pick_task_id
    WHERE order_line.outbound_line_id = task.outbound_line_id RETURNING order_line.outbound_line_id
), delivery_update AS (
    UPDATE delivery SET recipient_name = $6, recipient_reference = $7,
        proof_reference = $8, proof_uri = $9, latitude = $10, longitude = $11,
        updated_at = clock_timestamp(), updated_by = $13
    FROM context, line_update WHERE delivery.delivery_id = context.delivery_id
    RETURNING delivery.delivery_id
)
SELECT event.*, event_line.delivery_line_id, event_line.delivered_qty
FROM event CROSS JOIN event_line CROSS JOIN order_line_update CROSS JOIN delivery_update;

-- 2.4 Finalize only when every line is fully delivered. $1 delivery, $2 time, $3 actor
WITH completed AS (
    UPDATE delivery SET status_id = final.status_id, delivered_at = $2,
        updated_at = clock_timestamp(), updated_by = $3
    FROM document_status current, document_status final
    WHERE delivery.delivery_id = $1 AND current.status_id = delivery.status_id
      AND current.code IN ('ARRIVED','FAILED')
      AND final.document_type_id = delivery.document_type_id AND final.code = 'DELIVERED'
      AND nullif(trim(delivery.recipient_name),'') IS NOT NULL
      AND nullif(trim(delivery.proof_reference),'') IS NOT NULL
      AND NOT EXISTS (SELECT 1 FROM delivery_line line WHERE line.delivery_id = delivery.delivery_id
                      AND line.delivered_qty < line.planned_qty)
    RETURNING delivery.*
), outbound_update AS (
    UPDATE outbound_order outbound SET status_id = final.status_id, completed_at = $2,
        updated_at = clock_timestamp(), updated_by = $3, version_no = outbound.version_no + 1
    FROM completed, document_status final
    WHERE outbound.outbound_id = completed.outbound_id
      AND final.document_type_id = outbound.document_type_id AND final.code = 'DELIVERED'
    RETURNING outbound.outbound_id
)
SELECT completed.* FROM completed CROSS JOIN outbound_update;

-- 2.5 Fail remaining delivery. $1 delivery, $2 event, $3 reason,
-- $4 time, $5 lat/null, $6 long/null, $7 notes, $8 actor
WITH context AS (
    SELECT delivery.*, type.delivery_event_type_id, reason.delivery_failure_reason_id
    FROM delivery JOIN document_status status ON status.status_id = delivery.status_id
    JOIN delivery_event_type type ON type.code = 'DELIVERY_FAILED' AND type.is_active AND type.marks_failed
    JOIN delivery_failure_reason reason ON reason.code = $3 AND reason.is_active
    WHERE delivery.delivery_id = $1 AND status.code IN ('IN_TRANSIT','ARRIVED')
      AND EXISTS (SELECT 1 FROM delivery_line line WHERE line.delivery_id = delivery.delivery_id
                  AND line.delivered_qty + line.returned_qty < line.planned_qty)
), event AS (
    INSERT INTO delivery_event (delivery_event_id, delivery_id, delivery_event_type_id,
        event_at, delivery_failure_reason_id, notes, latitude, longitude, recorded_by)
    SELECT $2, delivery_id, delivery_event_type_id, $4,
           delivery_failure_reason_id, $7, $5, $6, $8 FROM context RETURNING *
), failed AS (
    UPDATE delivery SET status_id = next.status_id, updated_at = clock_timestamp(), updated_by = $8
    FROM context, event, document_status next
    WHERE delivery.delivery_id = context.delivery_id
      AND next.document_type_id = delivery.document_type_id AND next.code = 'FAILED'
    RETURNING delivery.*
), outbound_update AS (
    UPDATE outbound_order outbound SET status_id = next.status_id,
        updated_at = clock_timestamp(), updated_by = $8, version_no = outbound.version_no + 1
    FROM failed, document_status next WHERE outbound.outbound_id = failed.outbound_id
      AND next.document_type_id = outbound.document_type_id AND next.code = 'DELIVERY_FAILED'
    RETURNING outbound.outbound_id
)
SELECT event.*, failed.status_id FROM event CROSS JOIN failed CROSS JOIN outbound_update;

-- 2.6 Delivery queue. $1 owner, $2 warehouse, $3 status/null,
-- $4 from date, $5 until date, $6 limit, $7 offset
SELECT delivery.delivery_id, delivery.business_date, status.code AS status_code,
       outbound.client_delivery_order_no, outbound.ship_to_name,
       shipment.shipment_id, shipment.tracking_number, shipment.vehicle_number,
       delivery.planned_delivery_at, delivery.arrived_at, delivery.delivered_at,
       qty.planned_qty, qty.delivered_qty, qty.returned_qty,
       count(*) OVER () AS total_rows
FROM delivery JOIN document_status status ON status.status_id = delivery.status_id
JOIN outbound_order outbound ON outbound.outbound_id = delivery.outbound_id
JOIN shipment ON shipment.shipment_id = delivery.shipment_id
JOIN LATERAL (SELECT sum(planned_qty) planned_qty, sum(delivered_qty) delivered_qty,
                    sum(returned_qty) returned_qty
              FROM delivery_line WHERE delivery_id = delivery.delivery_id) qty ON true
WHERE outbound.owner_id = $1 AND outbound.warehouse_id = $2
  AND ($3::varchar IS NULL OR status.code = $3)
  AND delivery.business_date BETWEEN $4 AND $5
ORDER BY delivery.business_date DESC, delivery.delivery_id DESC LIMIT $6 OFFSET $7;

-- =============================================================================
-- 3. RETURN TO DEPOT
-- =============================================================================

-- 3.1 Partial/full return using owner/warehouse safe-return policy.
-- $1 delivery, $2 delivery line, $3 qty, $4 balance ID, $5 movement ID,
-- $6 return-line ID, $7 time, $8 notes, $9 actor
BEGIN;
SELECT delivery_line_id FROM delivery_line WHERE delivery_line_id = $2 FOR UPDATE;
WITH context AS (
    SELECT delivery.delivery_id, shipment.owner_id, shipment.warehouse_id,
           delivery.business_date, line.delivery_line_id, line.uom_id,
           source.item_id, source.lot_id, source.handling_unit_id,
           policy.return_location_id, policy.return_inventory_status_id AS target_status_id,
           reason.reason_code_id
    FROM delivery JOIN document_status status ON status.status_id = delivery.status_id
    JOIN shipment ON shipment.shipment_id = delivery.shipment_id
    JOIN delivery_line line ON line.delivery_line_id = $2 AND line.delivery_id = delivery.delivery_id
    JOIN shipment_line shipped ON shipped.shipment_line_id = line.shipment_line_id
    JOIN inventory_balance source ON source.balance_id = shipped.source_balance_id
    JOIN outbound_return_policy policy ON policy.owner_id = shipment.owner_id
      AND policy.warehouse_id = shipment.warehouse_id AND policy.is_active
    JOIN warehouse_location location ON location.location_id = policy.return_location_id AND location.is_active
    JOIN inventory_status target ON target.inventory_status_id = policy.return_inventory_status_id
      AND target.is_active AND NOT target.is_allocatable
    JOIN reason_code reason ON reason.module_code = 'OUTBOUND'
      AND reason.code = 'DELIVERY_RETURN' AND reason.is_active
    WHERE delivery.delivery_id = $1 AND status.code = 'FAILED' AND $3 > 0
      AND line.delivered_qty + line.returned_qty + $3 <= line.planned_qty
), balance AS (
    INSERT INTO inventory_balance (balance_id, owner_id, warehouse_id, location_id,
        item_id, lot_id, handling_unit_id, inventory_status_id, on_hand_qty, reserved_qty, uom_id)
    SELECT $4, owner_id, warehouse_id, return_location_id, item_id, lot_id,
           handling_unit_id, target_status_id, $3, 0, uom_id FROM context
    ON CONFLICT (owner_id, warehouse_id, location_id, item_id, lot_id, handling_unit_id, inventory_status_id)
    DO UPDATE SET on_hand_qty = inventory_balance.on_hand_qty + EXCLUDED.on_hand_qty,
                  version_no = inventory_balance.version_no + 1, updated_at = clock_timestamp()
    RETURNING *
), movement AS (
    INSERT INTO inventory_movement (movement_id, movement_type_id, owner_id, warehouse_id,
        business_date, item_id, lot_id, handling_unit_id, from_location_id, to_location_id,
        from_status_id, to_status_id, quantity, uom_id, source_document_id,
        source_line_id, reason_code_id, notes, created_by)
    SELECT $5, type.movement_type_id, context.owner_id,
           context.warehouse_id, context.business_date,
           context.item_id, context.lot_id, context.handling_unit_id,
           NULL, context.return_location_id, NULL,
           context.target_status_id, $3, context.uom_id,
           context.delivery_id, context.delivery_line_id,
           context.reason_code_id, $8, $9 FROM context CROSS JOIN balance
    JOIN movement_type type ON type.code = 'DELIVERY_RETURN' AND type.is_active RETURNING movement_id
), returned AS (
    INSERT INTO delivery_return_line (delivery_return_line_id, delivery_id, delivery_line_id,
        return_location_id, returned_balance_id, returned_qty, uom_id,
        movement_id, returned_at, returned_by)
    SELECT $6, context.delivery_id, context.delivery_line_id,
           context.return_location_id, balance.balance_id,
           $3, context.uom_id, movement.movement_id, $7, $9
    FROM context CROSS JOIN balance CROSS JOIN movement
    RETURNING *
), line_update AS (
    UPDATE delivery_line line SET returned_qty = line.returned_qty + returned.returned_qty
    FROM returned WHERE line.delivery_line_id = returned.delivery_line_id
      AND line.delivered_qty + line.returned_qty + returned.returned_qty <= line.planned_qty
    RETURNING line.*
), hu_update AS (
    UPDATE handling_unit hu SET current_location_id = context.return_location_id
    FROM context, line_update WHERE hu.handling_unit_id = context.handling_unit_id RETURNING hu.handling_unit_id
)
SELECT returned.* FROM returned CROSS JOIN line_update;
COMMIT;

-- 3.2 Close fully accounted failed delivery. $1 delivery, $2 event,
-- $3 time, $4 notes, $5 actor
WITH context AS (
    SELECT delivery.*, type.delivery_event_type_id, sum(line.delivered_qty) AS delivered_total,
           CASE WHEN sum(line.delivered_qty) > 0 THEN 'PARTIALLY_DELIVERED' ELSE 'RETURNED' END final_code
    FROM delivery JOIN document_status status ON status.status_id = delivery.status_id
    JOIN delivery_event_type type ON type.code = 'RETURNED_TO_DEPOT' AND type.is_active AND type.marks_returned
    JOIN delivery_line line ON line.delivery_id = delivery.delivery_id
    WHERE delivery.delivery_id = $1 AND status.code = 'FAILED'
    GROUP BY delivery.delivery_id, type.delivery_event_type_id
    HAVING bool_and(line.delivered_qty + line.returned_qty = line.planned_qty)
), event AS (
    INSERT INTO delivery_event (delivery_event_id, delivery_id, delivery_event_type_id,
        event_at, notes, recorded_by)
    SELECT $2, delivery_id, delivery_event_type_id, $3, $4, $5 FROM context RETURNING *
), closed AS (
    UPDATE delivery SET status_id = final.status_id,
        delivered_at = CASE WHEN context.delivered_total > 0 THEN COALESCE(delivery.delivered_at,$3)
                            ELSE delivery.delivered_at END,
        updated_at = clock_timestamp(), updated_by = $5, notes = COALESCE($4,delivery.notes)
    FROM context, event, document_status final
    WHERE delivery.delivery_id = context.delivery_id
      AND final.document_type_id = delivery.document_type_id AND final.code = context.final_code
    RETURNING delivery.*
), outbound_update AS (
    UPDATE outbound_order outbound SET status_id = final.status_id, completed_at = $3,
        updated_at = clock_timestamp(), updated_by = $5, version_no = outbound.version_no + 1
    FROM context, closed, document_status final
    WHERE outbound.outbound_id = context.outbound_id
      AND final.document_type_id = outbound.document_type_id AND final.code = context.final_code
    RETURNING outbound.outbound_id
)
SELECT closed.*, event.delivery_event_id FROM closed CROSS JOIN event CROSS JOIN outbound_update;

-- 3.3 Event/POD timeline. $1 delivery
SELECT delivery.delivery_id, status.code status_code, event.event_at,
       type.code event_type_code, reason.code failure_reason_code, event.notes,
       event.latitude, event.longitude, event.recorded_by, event.recipient_name,
       event.recipient_reference, event.proof_reference, event.proof_uri,
       event_line.delivery_line_id, event_line.delivered_qty
FROM delivery JOIN document_status status ON status.status_id = delivery.status_id
LEFT JOIN delivery_event event ON event.delivery_id = delivery.delivery_id
LEFT JOIN delivery_event_type type ON type.delivery_event_type_id = event.delivery_event_type_id
LEFT JOIN delivery_failure_reason reason ON reason.delivery_failure_reason_id = event.delivery_failure_reason_id
LEFT JOIN delivery_event_line event_line ON event_line.delivery_event_id = event.delivery_event_id
WHERE delivery.delivery_id = $1
ORDER BY event.event_at, event.delivery_event_id, event_line.delivery_line_id;

-- 3.4 Cancel empty planned shipment and release its DO memberships.
-- $1 shipment, $2 actor
WITH eligible AS (
    SELECT shipment.shipment_id
    FROM shipment JOIN document_status current ON current.status_id = shipment.status_id
    WHERE shipment.shipment_id = $1 AND current.code = 'PLANNED'
      AND NOT EXISTS (SELECT 1 FROM shipment_line line WHERE line.shipment_id = shipment.shipment_id)
), released_orders AS (
    UPDATE shipment_order manifest
    SET removed_at = clock_timestamp(), removed_by = $2
    FROM eligible
    WHERE manifest.shipment_id = eligible.shipment_id AND manifest.removed_at IS NULL
    RETURNING manifest.outbound_id
)
UPDATE shipment SET status_id = cancelled.status_id
FROM eligible, document_status cancelled
WHERE shipment.shipment_id = eligible.shipment_id
  AND cancelled.document_type_id = shipment.document_type_id
  AND cancelled.code = 'CANCELLED'
RETURNING shipment.*, (SELECT count(*) FROM released_orders) AS released_order_count;

-- 3.5 Cancel delivery before departure. $1 delivery, $2 actor
UPDATE delivery SET status_id = cancelled.status_id,
    updated_at = clock_timestamp(), updated_by = $2
FROM document_status current, document_status cancelled
WHERE delivery.delivery_id = $1 AND current.status_id = delivery.status_id
  AND current.code = 'PLANNED' AND cancelled.document_type_id = delivery.document_type_id
  AND cancelled.code = 'CANCELLED'
  AND NOT EXISTS (SELECT 1 FROM delivery_event event WHERE event.delivery_id = delivery.delivery_id)
RETURNING delivery.*;
