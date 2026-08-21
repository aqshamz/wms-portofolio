-- Inbound reports. Execute one numbered query at a time.

SET search_path TO wms, public;

-- 2.1 Purchase-order line fulfillment.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from date, $5 until date, $6 PO status code or NULL,
-- $7 vendor ID or NULL, $8 limit, $9 offset
WITH received AS (
    SELECT inbound_line.purchase_order_line_id,
           sum(receipt_line.received_qty) AS received_qty,
           sum(receipt_line.rejected_qty) AS rejected_qty,
           max(receipt.received_at) AS last_received_at
    FROM inbound_order_line inbound_line
    JOIN receipt_line ON receipt_line.inbound_line_id = inbound_line.inbound_line_id
    JOIN receipt ON receipt.receipt_id = receipt_line.receipt_id
    WHERE inbound_line.purchase_order_line_id IS NOT NULL
    GROUP BY inbound_line.purchase_order_line_id
)
SELECT purchase_order.purchase_order_id, purchase_order.purchase_order_no,
       purchase_order.business_date, owner.code AS owner_code,
       warehouse.code AS warehouse_code, vendor.code AS vendor_code,
       vendor.name AS vendor_name, status.code AS status_code,
       line.line_no, item.code AS item_code, item.name AS item_name,
       line.ordered_qty, COALESCE(received.received_qty, 0) AS received_qty,
       COALESCE(received.rejected_qty, 0) AS rejected_qty,
       line.ordered_qty - COALESCE(received.received_qty, 0) AS open_qty,
       CASE WHEN line.ordered_qty = 0 THEN 0
            ELSE round(COALESCE(received.received_qty, 0) * 100 / line.ordered_qty, 2)
       END AS receipt_percent,
       uom.code AS uom_code, purchase_order.expected_arrival_at,
       received.last_received_at
FROM purchase_order
JOIN purchase_order_line line ON line.purchase_order_id = purchase_order.purchase_order_id
JOIN document_status status ON status.status_id = purchase_order.status_id
JOIN organization owner ON owner.organization_id = purchase_order.owner_id
JOIN warehouse ON warehouse.warehouse_id = purchase_order.warehouse_id
JOIN business_partner vendor ON vendor.partner_id = purchase_order.vendor_id
JOIN item ON item.item_id = line.item_id
JOIN uom ON uom.uom_id = line.uom_id
LEFT JOIN received ON received.purchase_order_line_id = line.purchase_order_line_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1
      AND scope.owner_id = purchase_order.owner_id
      AND scope.warehouse_id = purchase_order.warehouse_id
  )
  AND ($2::uuid IS NULL OR purchase_order.owner_id = $2)
  AND ($3::uuid IS NULL OR purchase_order.warehouse_id = $3)
  AND purchase_order.business_date BETWEEN $4::date AND $5::date
  AND ($6::varchar IS NULL OR status.code = $6)
  AND ($7::uuid IS NULL OR purchase_order.vendor_id = $7)
ORDER BY purchase_order.business_date DESC, purchase_order.purchase_order_id, line.line_no
LIMIT $8 OFFSET $9;

-- 2.2 Receipt and QC daily summary.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from date, $5 until date
WITH receipt_quantity AS (
    SELECT receipt.receipt_id, receipt.owner_id, receipt.warehouse_id,
           receipt.business_date, count(*) AS receipt_line_count,
           sum(line.received_qty) AS received_qty,
           sum(line.rejected_qty) AS line_rejected_qty
    FROM receipt
    JOIN receipt_line line ON line.receipt_id = receipt.receipt_id
    WHERE receipt.business_date BETWEEN $4::date AND $5::date
    GROUP BY receipt.receipt_id
), inspection_quantity AS (
    SELECT line.receipt_id, count(inspection.inspection_id) AS inspection_count,
           sum(inspection.inspected_qty) AS inspected_qty,
           sum(inspection.passed_qty) AS passed_qty,
           sum(inspection.failed_qty) AS failed_qty
    FROM receipt_line line
    JOIN receipt_inventory inventory ON inventory.receipt_line_id = line.receipt_line_id
    JOIN quality_inspection inspection
      ON inspection.receipt_inventory_id = inventory.receipt_inventory_id
    GROUP BY line.receipt_id
)
SELECT quantity.business_date, owner.code AS owner_code,
       warehouse.code AS warehouse_code,
       count(*) AS receipt_count,
       sum(quantity.receipt_line_count) AS receipt_line_count,
       sum(quantity.received_qty) AS received_qty,
       sum(quantity.line_rejected_qty) AS line_rejected_qty,
       sum(COALESCE(inspection_summary.inspection_count, 0)) AS inspection_count,
       sum(COALESCE(inspection_summary.inspected_qty, 0)) AS inspected_qty,
       sum(COALESCE(inspection_summary.passed_qty, 0)) AS passed_qty,
       sum(COALESCE(inspection_summary.failed_qty, 0)) AS failed_qty,
       CASE WHEN sum(COALESCE(inspection_summary.inspected_qty, 0)) = 0 THEN NULL
            ELSE round(sum(COALESCE(inspection_summary.passed_qty, 0)) * 100
                       / sum(COALESCE(inspection_summary.inspected_qty, 0)), 2)
       END AS qc_pass_percent
FROM receipt_quantity quantity
JOIN organization owner ON owner.organization_id = quantity.owner_id
JOIN warehouse ON warehouse.warehouse_id = quantity.warehouse_id
LEFT JOIN inspection_quantity inspection_summary
  ON inspection_summary.receipt_id = quantity.receipt_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = quantity.owner_id
      AND scope.warehouse_id = quantity.warehouse_id
  )
  AND ($2::uuid IS NULL OR quantity.owner_id = $2)
  AND ($3::uuid IS NULL OR quantity.warehouse_id = $3)
GROUP BY quantity.business_date, owner.code, warehouse.code
ORDER BY quantity.business_date, owner.code, warehouse.code;

-- 2.3 QC inspection detail and exceptions.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from timestamp, $5 until timestamp, $6 result code or NULL,
-- $7 failures only, $8 limit, $9 offset
SELECT inspection.inspection_id, receipt.receipt_id, receipt.business_date,
       owner.code AS owner_code, warehouse.code AS warehouse_code,
       item.code AS item_code, lot.lot_number, inventory.handling_unit_id,
       quality_status.code AS quality_status_code,
       result.code AS result_code, result.name AS result_name,
       inspection.inspected_qty, inspection.passed_qty, inspection.failed_qty,
       uom.code AS uom_code, inspection.inspected_at,
       inspector.username AS inspected_by, inspection.notes
FROM quality_inspection inspection
JOIN receipt_inventory inventory
  ON inventory.receipt_inventory_id = inspection.receipt_inventory_id
JOIN receipt_line line ON line.receipt_line_id = inventory.receipt_line_id
JOIN receipt ON receipt.receipt_id = line.receipt_id
JOIN organization owner ON owner.organization_id = receipt.owner_id
JOIN warehouse ON warehouse.warehouse_id = receipt.warehouse_id
JOIN item ON item.item_id = inventory.item_id
JOIN uom ON uom.uom_id = inventory.base_uom_id
JOIN quality_status ON quality_status.quality_status_id = inspection.quality_status_id
LEFT JOIN inspection_result result
  ON result.inspection_result_id = inspection.inspection_result_id
LEFT JOIN inventory_lot lot ON lot.lot_id = inventory.lot_id
LEFT JOIN app_account inspector ON inspector.account_id = inspection.inspected_by
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = receipt.owner_id
      AND scope.warehouse_id = receipt.warehouse_id
  )
  AND ($2::uuid IS NULL OR receipt.owner_id = $2)
  AND ($3::uuid IS NULL OR receipt.warehouse_id = $3)
  AND inspection.inspected_at >= $4::timestamptz
  AND inspection.inspected_at < $5::timestamptz
  AND ($6::varchar IS NULL OR result.code = $6)
  AND (NOT $7 OR inspection.failed_qty > 0)
ORDER BY inspection.inspected_at DESC, inspection.inspection_id
LIMIT $8 OFFSET $9;

-- 2.4 Quarantine aging and client disposition summary.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 opened from date, $5 opened until date, $6 case status or NULL,
-- $7 open cases only, $8 limit, $9 offset
WITH disposition AS (
    SELECT decision.quarantine_case_id,
           sum(decision.disposition_qty) AS decided_qty,
           sum(decision.disposition_qty)
             FILTER (WHERE type.code = 'ACCEPT') AS accepted_qty,
           sum(decision.disposition_qty)
             FILTER (WHERE type.code = 'REWORK') AS rework_qty,
           sum(decision.disposition_qty)
             FILTER (WHERE type.code = 'RETURN') AS returned_qty,
           sum(decision.disposition_qty)
             FILTER (WHERE type.code = 'DISPOSE') AS disposed_qty,
           max(decision.decided_at) AS last_decided_at
    FROM quarantine_disposition decision
    JOIN quarantine_disposition_type type
      ON type.quarantine_disposition_type_id = decision.quarantine_disposition_type_id
    GROUP BY decision.quarantine_case_id
)
SELECT quarantine.quarantine_case_id, owner.code AS owner_code,
       warehouse.code AS warehouse_code, status.code AS status_code,
       item.code AS item_code, lot.lot_number,
       quarantine.quarantine_qty, COALESCE(disposition.decided_qty, 0) AS decided_qty,
       quarantine.quarantine_qty - COALESCE(disposition.decided_qty, 0) AS pending_qty,
       COALESCE(disposition.accepted_qty, 0) AS accepted_qty,
       COALESCE(disposition.rework_qty, 0) AS rework_qty,
       COALESCE(disposition.returned_qty, 0) AS returned_qty,
       COALESCE(disposition.disposed_qty, 0) AS disposed_qty,
       uom.code AS uom_code, quarantine.opened_at, quarantine.closed_at,
       CASE WHEN quarantine.closed_at IS NULL
            THEN current_date - quarantine.opened_at::date
            ELSE quarantine.closed_at::date - quarantine.opened_at::date END AS age_days,
       disposition.last_decided_at
FROM quarantine_case quarantine
JOIN document_status status ON status.status_id = quarantine.status_id
JOIN organization owner ON owner.organization_id = quarantine.owner_id
JOIN warehouse ON warehouse.warehouse_id = quarantine.warehouse_id
JOIN receipt_inventory inventory
  ON inventory.receipt_inventory_id = quarantine.receipt_inventory_id
JOIN item ON item.item_id = inventory.item_id
JOIN uom ON uom.uom_id = quarantine.uom_id
LEFT JOIN inventory_lot lot ON lot.lot_id = inventory.lot_id
LEFT JOIN disposition ON disposition.quarantine_case_id = quarantine.quarantine_case_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = quarantine.owner_id
      AND scope.warehouse_id = quarantine.warehouse_id
  )
  AND ($2::uuid IS NULL OR quarantine.owner_id = $2)
  AND ($3::uuid IS NULL OR quarantine.warehouse_id = $3)
  AND quarantine.opened_at::date BETWEEN $4::date AND $5::date
  AND ($6::varchar IS NULL OR status.code = $6)
  AND (NOT $7 OR quarantine.closed_at IS NULL)
ORDER BY age_days DESC, quarantine.opened_at
LIMIT $8 OFFSET $9;

-- 2.5 Putaway task productivity and backlog.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 created from, $5 created until, $6 task status or NULL,
-- $7 assignee ID or NULL, $8 limit, $9 offset
SELECT task.putaway_task_id, owner.code AS owner_code,
       warehouse.code AS warehouse_code, item.code AS item_code,
       source_location.code AS source_location_code,
       target_location.code AS target_location_code,
       task_status.code AS task_status_code, priority.code AS priority_code,
       task.planned_qty, task.completed_qty,
       task.planned_qty - task.completed_qty AS open_qty,
       uom.code AS uom_code, assignee.username AS assigned_to,
       task.created_at, task.started_at, task.completed_at,
       CASE WHEN task.started_at IS NULL THEN NULL
            ELSE extract(epoch FROM (COALESCE(task.completed_at, clock_timestamp())
                                     - task.started_at)) / 60 END AS elapsed_minutes
FROM putaway_task task
JOIN organization owner ON owner.organization_id = task.owner_id
JOIN warehouse ON warehouse.warehouse_id = task.warehouse_id
JOIN item ON item.item_id = task.item_id
JOIN warehouse_location source_location
  ON source_location.location_id = task.source_location_id
JOIN warehouse_location target_location
  ON target_location.location_id = task.target_location_id
JOIN task_status ON task_status.task_status_id = task.task_status_id
JOIN task_priority priority ON priority.task_priority_id = task.task_priority_id
JOIN uom ON uom.uom_id = task.uom_id
LEFT JOIN app_account assignee ON assignee.account_id = task.assigned_to
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = task.owner_id
      AND scope.warehouse_id = task.warehouse_id
  )
  AND ($2::uuid IS NULL OR task.owner_id = $2)
  AND ($3::uuid IS NULL OR task.warehouse_id = $3)
  AND task.created_at >= $4::timestamptz AND task.created_at < $5::timestamptz
  AND ($6::varchar IS NULL OR task_status.code = $6)
  AND ($7::uuid IS NULL OR task.assigned_to = $7)
ORDER BY task.created_at DESC, task.putaway_task_id
LIMIT $8 OFFSET $9;

-- 2.6 Vendor inbound performance.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 PO from date, $5 PO until date, $6 vendor ID or NULL
WITH po_quantity AS (
    SELECT purchase_order.purchase_order_id, purchase_order.owner_id,
           purchase_order.vendor_id, purchase_order.warehouse_id,
           purchase_order.expected_arrival_at,
           sum(line.ordered_qty) AS ordered_qty
    FROM purchase_order
    JOIN purchase_order_line line ON line.purchase_order_id = purchase_order.purchase_order_id
    WHERE purchase_order.business_date BETWEEN $4::date AND $5::date
    GROUP BY purchase_order.purchase_order_id
), receipt_quantity AS (
    SELECT inbound_line.purchase_order_line_id,
           sum(receipt_line.received_qty) AS received_qty,
           max(receipt.received_at) AS last_received_at
    FROM inbound_order_line inbound_line
    JOIN receipt_line ON receipt_line.inbound_line_id = inbound_line.inbound_line_id
    JOIN receipt ON receipt.receipt_id = receipt_line.receipt_id
    GROUP BY inbound_line.purchase_order_line_id
), po_received AS (
    SELECT po_quantity.*,
           COALESCE(sum(receipt_quantity.received_qty), 0) AS received_qty,
           max(receipt_quantity.last_received_at) AS last_received_at
    FROM po_quantity
    LEFT JOIN purchase_order_line line
      ON line.purchase_order_id = po_quantity.purchase_order_id
    LEFT JOIN receipt_quantity
      ON receipt_quantity.purchase_order_line_id = line.purchase_order_line_id
    GROUP BY po_quantity.purchase_order_id, po_quantity.owner_id,
             po_quantity.vendor_id, po_quantity.warehouse_id,
             po_quantity.expected_arrival_at, po_quantity.ordered_qty
)
SELECT owner.code AS owner_code, warehouse.code AS warehouse_code,
       vendor.code AS vendor_code, vendor.name AS vendor_name,
       count(*) AS purchase_order_count, sum(po_received.ordered_qty) AS ordered_qty,
       sum(po_received.received_qty) AS received_qty,
       round(sum(po_received.received_qty) * 100
             / NULLIF(sum(po_received.ordered_qty), 0), 2) AS fill_percent,
       count(*) FILTER (
           WHERE po_received.last_received_at IS NOT NULL
             AND (po_received.expected_arrival_at IS NULL
                  OR po_received.last_received_at <= po_received.expected_arrival_at)
       ) AS on_time_po_count,
       round(avg(extract(epoch FROM (
           po_received.last_received_at - po_received.expected_arrival_at
       )) / 3600) FILTER (
           WHERE po_received.last_received_at IS NOT NULL
             AND po_received.expected_arrival_at IS NOT NULL
       ), 2) AS average_arrival_variance_hours
FROM po_received
JOIN organization owner ON owner.organization_id = po_received.owner_id
JOIN warehouse ON warehouse.warehouse_id = po_received.warehouse_id
JOIN business_partner vendor ON vendor.partner_id = po_received.vendor_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INBOUND'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = po_received.owner_id
      AND scope.warehouse_id = po_received.warehouse_id
  )
  AND ($2::uuid IS NULL OR po_received.owner_id = $2)
  AND ($3::uuid IS NULL OR po_received.warehouse_id = $3)
  AND ($6::uuid IS NULL OR po_received.vendor_id = $6)
GROUP BY owner.code, warehouse.code, vendor.code, vendor.name
ORDER BY owner.code, warehouse.code, vendor.code;
