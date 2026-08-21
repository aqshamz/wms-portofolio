-- Outbound reports. Execute one numbered query at a time.

SET search_path TO wms, public;

-- 4.1 Client delivery-order fulfillment by line.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from date, $5 until date, $6 status code or NULL,
-- $7 customer ID or NULL, $8 limit, $9 offset
SELECT outbound.outbound_id, outbound.client_delivery_order_no,
       outbound.business_date, owner.code AS owner_code,
       warehouse.code AS warehouse_code, customer.code AS customer_code,
       customer.name AS customer_name, status.code AS status_code,
       outbound.ship_to_name, outbound.requested_ship_at,
       line.line_no, item.code AS item_code, item.name AS item_name,
       line.ordered_qty, line.allocated_qty, line.picked_qty,
       line.rejected_qty, line.short_accepted_qty,
       line.checked_qty, line.packed_qty, line.shipped_qty,
       line.delivered_qty,
       line.ordered_qty - line.short_accepted_qty AS fulfillment_target_qty,
       line.ordered_qty - line.short_accepted_qty - line.delivered_qty AS open_qty,
       round(line.delivered_qty * 100 /
             NULLIF(line.ordered_qty - line.short_accepted_qty, 0), 2) AS delivered_percent,
       uom.code AS uom_code
FROM outbound_order outbound
JOIN outbound_order_line line ON line.outbound_id = outbound.outbound_id
JOIN organization owner ON owner.organization_id = outbound.owner_id
JOIN warehouse ON warehouse.warehouse_id = outbound.warehouse_id
JOIN business_partner customer ON customer.partner_id = outbound.customer_id
JOIN document_status status ON status.status_id = outbound.status_id
JOIN item ON item.item_id = line.item_id
JOIN uom ON uom.uom_id = line.uom_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.OUTBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = outbound.owner_id
      AND scope.warehouse_id = outbound.warehouse_id
  )
  AND ($2::uuid IS NULL OR outbound.owner_id = $2)
  AND ($3::uuid IS NULL OR outbound.warehouse_id = $3)
  AND outbound.business_date BETWEEN $4::date AND $5::date
  AND ($6::varchar IS NULL OR status.code = $6)
  AND ($7::uuid IS NULL OR outbound.customer_id = $7)
ORDER BY outbound.business_date DESC, outbound.outbound_id, line.line_no
LIMIT $8 OFFSET $9;

-- 4.2 Allocation and picking exceptions.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from date, $5 until date, $6 exceptions only,
-- $7 task status or NULL, $8 limit, $9 offset
SELECT outbound.outbound_id, outbound.client_delivery_order_no,
       outbound.business_date, owner.code AS owner_code,
       warehouse.code AS warehouse_code, line.line_no,
       item.code AS item_code, line.ordered_qty, line.allocated_qty,
       line.rejected_qty, line.short_accepted_qty, line.picked_qty,
       greatest(line.ordered_qty - line.short_accepted_qty
                - (line.allocated_qty - line.rejected_qty), 0) AS allocation_short_qty,
       line.allocated_qty - line.picked_qty AS unpicked_qty,
       reservation.reservation_id, reservation.reserved_qty,
       reservation.picked_qty AS reservation_picked_qty,
       reservation_status.code AS reservation_status_code,
       task.pick_task_id, task.planned_qty, task.picked_qty AS task_picked_qty,
       task.short_qty, task_status.code AS task_status_code,
       short_reason.code AS short_reason_code, assignee.username AS assigned_to,
       task.started_at, task.completed_at, uom.code AS uom_code
FROM outbound_order outbound
JOIN outbound_order_line line ON line.outbound_id = outbound.outbound_id
JOIN organization owner ON owner.organization_id = outbound.owner_id
JOIN warehouse ON warehouse.warehouse_id = outbound.warehouse_id
JOIN item ON item.item_id = line.item_id
JOIN uom ON uom.uom_id = line.uom_id
LEFT JOIN inventory_reservation reservation
  ON reservation.outbound_line_id = line.outbound_line_id
LEFT JOIN document_status reservation_status
  ON reservation_status.status_id = reservation.status_id
LEFT JOIN pick_task task ON task.reservation_id = reservation.reservation_id
LEFT JOIN task_status ON task_status.task_status_id = task.task_status_id
LEFT JOIN reason_code short_reason ON short_reason.reason_code_id = task.short_reason_code_id
LEFT JOIN app_account assignee ON assignee.account_id = task.assigned_to
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.OUTBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = outbound.owner_id
      AND scope.warehouse_id = outbound.warehouse_id
  )
  AND ($2::uuid IS NULL OR outbound.owner_id = $2)
  AND ($3::uuid IS NULL OR outbound.warehouse_id = $3)
  AND outbound.business_date BETWEEN $4::date AND $5::date
  AND (NOT $6
       OR line.allocated_qty - line.rejected_qty < line.ordered_qty - line.short_accepted_qty
       OR COALESCE(task.short_qty, 0) > 0)
  AND ($7::varchar IS NULL OR task_status.code = $7)
ORDER BY outbound.business_date DESC, outbound.outbound_id, line.line_no
LIMIT $8 OFFSET $9;

-- 4.3 Outbound-wave productivity.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from date, $5 until date, $6 wave status or NULL
WITH order_count AS (
    SELECT wave_id, count(*) AS outbound_order_count
    FROM outbound_wave_order GROUP BY wave_id
), task_quantity AS (
    SELECT task.wave_id, count(*) AS pick_task_count,
           count(*) FILTER (WHERE task_state.code = 'COMPLETED') AS completed_task_count,
           sum(task.planned_qty) AS planned_qty,
           sum(task.picked_qty) AS picked_qty,
           sum(task.short_qty) AS short_qty
    FROM pick_task task
    JOIN task_status task_state ON task_state.task_status_id = task.task_status_id
    GROUP BY task.wave_id
)
SELECT wave.wave_id, wave.business_date, owner.code AS owner_code,
       warehouse.code AS warehouse_code, wave_type.code AS wave_type_code,
       status.code AS status_code, strategy.code AS picking_strategy_code,
       COALESCE(order_count.outbound_order_count, 0) AS outbound_order_count,
       COALESCE(task_quantity.pick_task_count, 0) AS pick_task_count,
       COALESCE(task_quantity.completed_task_count, 0) AS completed_task_count,
       COALESCE(task_quantity.planned_qty, 0) AS planned_qty,
       COALESCE(task_quantity.picked_qty, 0) AS picked_qty,
       COALESCE(task_quantity.short_qty, 0) AS short_qty,
       wave.planned_release_at, wave.released_at, wave.completed_at,
       CASE WHEN wave.released_at IS NULL THEN NULL
            ELSE extract(epoch FROM (COALESCE(wave.completed_at, clock_timestamp())
                                     - wave.released_at)) / 60 END AS elapsed_minutes
FROM outbound_wave wave
JOIN organization owner ON owner.organization_id = wave.owner_id
JOIN warehouse ON warehouse.warehouse_id = wave.warehouse_id
JOIN outbound_wave_type wave_type
  ON wave_type.outbound_wave_type_id = wave.outbound_wave_type_id
JOIN document_status status ON status.status_id = wave.status_id
LEFT JOIN picking_strategy strategy ON strategy.picking_strategy_id = wave.picking_strategy_id
LEFT JOIN order_count ON order_count.wave_id = wave.wave_id
LEFT JOIN task_quantity ON task_quantity.wave_id = wave.wave_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.OUTBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = wave.owner_id
      AND scope.warehouse_id = wave.warehouse_id
  )
  AND ($2::uuid IS NULL OR wave.owner_id = $2)
  AND ($3::uuid IS NULL OR wave.warehouse_id = $3)
  AND wave.business_date BETWEEN $4::date AND $5::date
  AND ($6::varchar IS NULL OR status.code = $6)
ORDER BY wave.business_date DESC, wave.wave_id;

-- 4.4 Checking, packing, and shipment reconciliation by DO.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from date, $5 until date, $6 discrepancies only,
-- $7 limit, $8 offset
WITH ordered AS (
    SELECT outbound_id, sum(ordered_qty) AS ordered_qty,
           sum(short_accepted_qty) AS short_accepted_qty,
           sum(picked_qty) AS picked_qty,
           sum(rejected_qty) AS rejected_qty,
           sum(checked_qty) AS checked_qty,
           sum(packed_qty) AS packed_qty,
           sum(shipped_qty) AS shipped_qty
    FROM outbound_order_line GROUP BY outbound_id
), check_summary AS (
    SELECT staging.outbound_id,
           count(DISTINCT check_header.outbound_check_id) AS check_count,
           sum(check_line.expected_qty) AS expected_check_qty,
           sum(COALESCE(check_line.checked_qty, 0)) AS actual_check_qty,
           count(*) FILTER (WHERE result.outbound_check_result_id IS NOT NULL
                              AND NOT result.is_pass) AS failed_check_line_count
    FROM outbound_staging staging
    JOIN outbound_check check_header ON check_header.staging_id = staging.staging_id
    JOIN outbound_check_line check_line
      ON check_line.outbound_check_id = check_header.outbound_check_id
    LEFT JOIN outbound_check_result result
      ON result.outbound_check_result_id = check_line.outbound_check_result_id
    GROUP BY staging.outbound_id
), packing_summary AS (
    SELECT packing.outbound_id, count(DISTINCT packing.packing_id) AS packing_count,
           sum(line.packed_qty) AS packed_line_qty
    FROM packing JOIN packing_line line ON line.packing_id = packing.packing_id
    GROUP BY packing.outbound_id
), shipment_summary AS (
    SELECT packing.outbound_id, count(DISTINCT shipment.shipment_id) AS shipment_count,
           sum(line.shipped_qty) AS shipment_line_qty,
           max(shipment.shipped_at) AS last_shipped_at
    FROM shipment JOIN shipment_line line ON line.shipment_id = shipment.shipment_id
    JOIN packing_line packed ON packed.packing_line_id = line.packing_line_id
    JOIN packing ON packing.packing_id = packed.packing_id
    GROUP BY packing.outbound_id
)
SELECT outbound.outbound_id, outbound.client_delivery_order_no,
       outbound.business_date, owner.code AS owner_code,
       warehouse.code AS warehouse_code, status.code AS status_code,
       ordered.ordered_qty, ordered.short_accepted_qty,
       ordered.picked_qty, ordered.rejected_qty, ordered.checked_qty,
       ordered.packed_qty, ordered.shipped_qty,
       COALESCE(check_summary.check_count, 0) AS check_count,
       COALESCE(check_summary.failed_check_line_count, 0) AS failed_check_line_count,
       COALESCE(packing_summary.packing_count, 0) AS packing_count,
       COALESCE(packing_summary.packed_line_qty, 0) AS packed_line_qty,
       COALESCE(shipment_summary.shipment_count, 0) AS shipment_count,
       COALESCE(shipment_summary.shipment_line_qty, 0) AS shipment_line_qty,
       shipment_summary.last_shipped_at
FROM outbound_order outbound
JOIN ordered ON ordered.outbound_id = outbound.outbound_id
JOIN organization owner ON owner.organization_id = outbound.owner_id
JOIN warehouse ON warehouse.warehouse_id = outbound.warehouse_id
JOIN document_status status ON status.status_id = outbound.status_id
LEFT JOIN check_summary ON check_summary.outbound_id = outbound.outbound_id
LEFT JOIN packing_summary ON packing_summary.outbound_id = outbound.outbound_id
LEFT JOIN shipment_summary ON shipment_summary.outbound_id = outbound.outbound_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.OUTBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = outbound.owner_id
      AND scope.warehouse_id = outbound.warehouse_id
  )
  AND ($2::uuid IS NULL OR outbound.owner_id = $2)
  AND ($3::uuid IS NULL OR outbound.warehouse_id = $3)
  AND outbound.business_date BETWEEN $4::date AND $5::date
  AND (NOT $6 OR ordered.checked_qty <> ordered.picked_qty - ordered.rejected_qty
       OR ordered.checked_qty <> ordered.ordered_qty - ordered.short_accepted_qty
       OR ordered.packed_qty <> ordered.checked_qty
       OR ordered.shipped_qty <> ordered.packed_qty
       OR COALESCE(check_summary.failed_check_line_count, 0) > 0)
ORDER BY outbound.business_date DESC, outbound.outbound_id
LIMIT $7 OFFSET $8;

-- 4.5 Store-delivery performance.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from date, $5 until date, $6 delivery status or NULL,
-- $7 late/failed only, $8 limit, $9 offset
SELECT delivery.delivery_id, delivery.business_date,
       outbound.client_delivery_order_no, owner.code AS owner_code,
       warehouse.code AS warehouse_code, outbound.ship_to_name,
       status.code AS delivery_status_code, shipment.shipment_id,
       shipment.tracking_number, delivery.planned_delivery_at,
       delivery.arrived_at, delivery.delivered_at,
       CASE WHEN delivery.delivered_at IS NULL OR delivery.planned_delivery_at IS NULL
            THEN NULL
            ELSE extract(epoch FROM (delivery.delivered_at - delivery.planned_delivery_at)) / 3600
       END AS delivery_variance_hours,
       delivery.recipient_name, delivery.recipient_reference,
       delivery.proof_reference, delivery.proof_uri
FROM delivery
JOIN outbound_order outbound ON outbound.outbound_id = delivery.outbound_id
JOIN shipment ON shipment.shipment_id = delivery.shipment_id
JOIN organization owner ON owner.organization_id = outbound.owner_id
JOIN warehouse ON warehouse.warehouse_id = outbound.warehouse_id
JOIN document_status status ON status.status_id = delivery.status_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.OUTBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = outbound.owner_id
      AND scope.warehouse_id = outbound.warehouse_id
  )
  AND ($2::uuid IS NULL OR outbound.owner_id = $2)
  AND ($3::uuid IS NULL OR outbound.warehouse_id = $3)
  AND delivery.business_date BETWEEN $4::date AND $5::date
  AND ($6::varchar IS NULL OR status.code = $6)
  AND (NOT $7 OR status.code IN ('FAILED', 'RETURNED', 'PARTIALLY_DELIVERED')
       OR (delivery.delivered_at IS NOT NULL
           AND delivery.planned_delivery_at IS NOT NULL
           AND delivery.delivered_at > delivery.planned_delivery_at))
ORDER BY delivery.business_date DESC, delivery.delivery_id
LIMIT $8 OFFSET $9;

-- 4.6 Delivery failure events and returned inventory.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from timestamp, $5 until timestamp, $6 failure reason code or NULL,
-- $7 limit, $8 offset
WITH returned AS (
    SELECT delivery_id, sum(returned_qty) AS returned_qty,
           count(*) AS return_line_count, max(returned_at) AS last_returned_at
    FROM delivery_return_line GROUP BY delivery_id
)
SELECT event.delivery_event_id, event.event_at, delivery.delivery_id,
       outbound.client_delivery_order_no, owner.code AS owner_code,
       warehouse.code AS warehouse_code, event_type.code AS event_type_code,
       failure.code AS failure_reason_code, failure.name AS failure_reason_name,
       event.notes, event.latitude, event.longitude,
       COALESCE(returned.return_line_count, 0) AS return_line_count,
       COALESCE(returned.returned_qty, 0) AS returned_qty,
       returned.last_returned_at, recorder.username AS recorded_by
FROM delivery_event event
JOIN delivery ON delivery.delivery_id = event.delivery_id
JOIN outbound_order outbound ON outbound.outbound_id = delivery.outbound_id
JOIN organization owner ON owner.organization_id = outbound.owner_id
JOIN warehouse ON warehouse.warehouse_id = outbound.warehouse_id
JOIN delivery_event_type event_type
  ON event_type.delivery_event_type_id = event.delivery_event_type_id
LEFT JOIN delivery_failure_reason failure
  ON failure.delivery_failure_reason_id = event.delivery_failure_reason_id
LEFT JOIN returned ON returned.delivery_id = delivery.delivery_id
JOIN app_account recorder ON recorder.account_id = event.recorded_by
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.OUTBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = outbound.owner_id
      AND scope.warehouse_id = outbound.warehouse_id
  )
  AND ($2::uuid IS NULL OR outbound.owner_id = $2)
  AND ($3::uuid IS NULL OR outbound.warehouse_id = $3)
  AND event.event_at >= $4::timestamptz AND event.event_at < $5::timestamptz
  AND ($6::varchar IS NULL OR failure.code = $6)
  AND (event_type.marks_failed OR event_type.marks_returned
       OR event.delivery_failure_reason_id IS NOT NULL)
ORDER BY event.event_at DESC, event.delivery_event_id
LIMIT $7 OFFSET $8;
