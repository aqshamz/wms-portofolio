-- Stock-control reports. Execute one numbered query at a time.

SET search_path TO wms, public;

-- 3.1 Current stock-on-hand detail.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 item ID or NULL, $5 inventory-status code or NULL,
-- $6 location ID or NULL, $7 positive stock only, $8 limit, $9 offset
SELECT balance.balance_id, owner.code AS owner_code,
       warehouse.code AS warehouse_code, zone.code AS zone_code,
       location.code AS location_code, item.code AS item_code,
       item.name AS item_name, lot.lot_number, lot.expiry_date,
       CASE WHEN lot.expiry_date IS NULL THEN NULL
            ELSE lot.expiry_date - current_date END AS days_to_expiry,
       balance.handling_unit_id, status.code AS inventory_status_code,
       balance.on_hand_qty, balance.reserved_qty,
       balance.on_hand_qty - balance.reserved_qty AS available_qty,
       uom.code AS uom_code, balance.updated_at, balance.version_no
FROM inventory_balance balance
JOIN organization owner ON owner.organization_id = balance.owner_id
JOIN warehouse ON warehouse.warehouse_id = balance.warehouse_id
JOIN warehouse_location location ON location.location_id = balance.location_id
JOIN warehouse_zone zone ON zone.zone_id = location.zone_id
JOIN item ON item.item_id = balance.item_id
JOIN inventory_status status ON status.inventory_status_id = balance.inventory_status_id
JOIN uom ON uom.uom_id = balance.uom_id
LEFT JOIN inventory_lot lot ON lot.lot_id = balance.lot_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INVENTORY'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = balance.owner_id
      AND scope.warehouse_id = balance.warehouse_id
  )
  AND ($2::uuid IS NULL OR balance.owner_id = $2)
  AND ($3::uuid IS NULL OR balance.warehouse_id = $3)
  AND ($4::uuid IS NULL OR balance.item_id = $4)
  AND ($5::varchar IS NULL OR status.code = $5)
  AND ($6::uuid IS NULL OR balance.location_id = $6)
  AND (NOT $7 OR balance.on_hand_qty > 0)
ORDER BY owner.code, warehouse.code, item.code, lot.expiry_date NULLS LAST, location.code
LIMIT $8 OFFSET $9;

-- 3.2 Stock summary by item and inventory status.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 item category ID or NULL, $5 positive stock only
SELECT owner.code AS owner_code, warehouse.code AS warehouse_code,
       category.code AS category_code, item.code AS item_code,
       item.name AS item_name, status.code AS inventory_status_code,
       uom.code AS uom_code, count(*) AS balance_count,
       count(DISTINCT balance.location_id) AS location_count,
       count(DISTINCT balance.lot_id) FILTER (WHERE balance.lot_id IS NOT NULL) AS lot_count,
       count(DISTINCT balance.handling_unit_id)
         FILTER (WHERE balance.handling_unit_id IS NOT NULL) AS handling_unit_count,
       sum(balance.on_hand_qty) AS on_hand_qty,
       sum(balance.reserved_qty) AS reserved_qty,
       sum(balance.on_hand_qty - balance.reserved_qty) AS available_qty
FROM inventory_balance balance
JOIN organization owner ON owner.organization_id = balance.owner_id
JOIN warehouse ON warehouse.warehouse_id = balance.warehouse_id
JOIN item ON item.item_id = balance.item_id
LEFT JOIN item_category category ON category.category_id = item.category_id
JOIN inventory_status status ON status.inventory_status_id = balance.inventory_status_id
JOIN uom ON uom.uom_id = balance.uom_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INVENTORY'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = balance.owner_id
      AND scope.warehouse_id = balance.warehouse_id
  )
  AND ($2::uuid IS NULL OR balance.owner_id = $2)
  AND ($3::uuid IS NULL OR balance.warehouse_id = $3)
  AND ($4::uuid IS NULL OR item.category_id = $4)
  AND (NOT $5 OR balance.on_hand_qty > 0)
GROUP BY owner.code, warehouse.code, category.code,
         item.code, item.name, status.code, uom.code
ORDER BY owner.code, warehouse.code, item.code, status.code;

-- 3.3 Immutable inventory movement ledger.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from timestamp, $5 until timestamp, $6 movement-type code or NULL,
-- $7 item ID or NULL, $8 source document ID or NULL, $9 limit, $10 offset
SELECT movement.movement_id, movement.business_date, movement.occurred_at,
       owner.code AS owner_code, warehouse.code AS warehouse_code,
       type.code AS movement_type_code, item.code AS item_code,
       lot.lot_number, serial.serial_no, movement.handling_unit_id,
       from_location.code AS from_location_code,
       to_location.code AS to_location_code,
       from_status.code AS from_status_code, to_status.code AS to_status_code,
       movement.quantity, uom.code AS uom_code,
       movement.source_document_id, movement.source_line_id,
       reason.code AS reason_code, creator.username AS created_by,
       movement.notes
FROM inventory_movement movement
JOIN organization owner ON owner.organization_id = movement.owner_id
JOIN warehouse ON warehouse.warehouse_id = movement.warehouse_id
JOIN movement_type type ON type.movement_type_id = movement.movement_type_id
JOIN item ON item.item_id = movement.item_id
JOIN uom ON uom.uom_id = movement.uom_id
LEFT JOIN inventory_lot lot ON lot.lot_id = movement.lot_id
LEFT JOIN serial_number serial ON serial.serial_id = movement.serial_id
LEFT JOIN warehouse_location from_location
  ON from_location.location_id = movement.from_location_id
LEFT JOIN warehouse_location to_location
  ON to_location.location_id = movement.to_location_id
LEFT JOIN inventory_status from_status
  ON from_status.inventory_status_id = movement.from_status_id
LEFT JOIN inventory_status to_status
  ON to_status.inventory_status_id = movement.to_status_id
LEFT JOIN reason_code reason ON reason.reason_code_id = movement.reason_code_id
JOIN app_account creator ON creator.account_id = movement.created_by
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INVENTORY'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = movement.owner_id
      AND scope.warehouse_id = movement.warehouse_id
  )
  AND ($2::uuid IS NULL OR movement.owner_id = $2)
  AND ($3::uuid IS NULL OR movement.warehouse_id = $3)
  AND movement.occurred_at >= $4::timestamptz
  AND movement.occurred_at < $5::timestamptz
  AND ($6::varchar IS NULL OR type.code = $6)
  AND ($7::uuid IS NULL OR movement.item_id = $7)
  AND ($8::varchar IS NULL OR movement.source_document_id = $8)
ORDER BY movement.occurred_at DESC, movement.movement_id
LIMIT $9 OFFSET $10;

-- 3.4 Internal-movement completion report.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from date, $5 until date, $6 status code or NULL
WITH quantities AS (
    SELECT line.internal_move_id, count(*) AS line_count,
           sum(line.planned_qty) AS planned_qty,
           sum(line.completed_qty) AS completed_qty
    FROM internal_move_order_line line
    GROUP BY line.internal_move_id
)
SELECT move.internal_move_id, move.business_date,
       owner.code AS owner_code, warehouse.code AS warehouse_code,
       move_type.code AS move_type_code, status.code AS status_code,
       reason.code AS reason_code, quantities.line_count,
       quantities.planned_qty, quantities.completed_qty,
       quantities.planned_qty - quantities.completed_qty AS open_qty,
       move.requested_at, move.approved_at, move.completed_at,
       CASE WHEN move.completed_at IS NULL THEN NULL
            ELSE extract(epoch FROM (move.completed_at - move.requested_at)) / 60
       END AS request_to_complete_minutes
FROM internal_move_order move
JOIN quantities ON quantities.internal_move_id = move.internal_move_id
JOIN organization owner ON owner.organization_id = move.owner_id
JOIN warehouse ON warehouse.warehouse_id = move.warehouse_id
JOIN internal_move_type move_type
  ON move_type.internal_move_type_id = move.internal_move_type_id
JOIN document_status status ON status.status_id = move.status_id
LEFT JOIN reason_code reason ON reason.reason_code_id = move.reason_code_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INVENTORY'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = move.owner_id
      AND scope.warehouse_id = move.warehouse_id
  )
  AND ($2::uuid IS NULL OR move.owner_id = $2)
  AND ($3::uuid IS NULL OR move.warehouse_id = $3)
  AND move.business_date BETWEEN $4::date AND $5::date
  AND ($6::varchar IS NULL OR status.code = $6)
ORDER BY move.business_date DESC, move.internal_move_id;

-- 3.5 Inter-warehouse transfer progress and variance.
-- $1 account ID, $2 owner ID or NULL, $3 source warehouse ID or NULL,
-- $4 target warehouse ID or NULL, $5 from date, $6 until date,
-- $7 status code or NULL, $8 limit, $9 offset
SELECT transfer.transfer_id, transfer.business_date,
       owner.code AS owner_code, source.code AS source_warehouse_code,
       target.code AS target_warehouse_code, status.code AS status_code,
       line.line_no, item.code AS item_code, lot.lot_number,
       line.requested_qty, line.dispatched_qty, line.received_qty,
       line.requested_qty - line.dispatched_qty AS pending_dispatch_qty,
       line.dispatched_qty - line.received_qty AS in_transit_qty,
       uom.code AS uom_code, transfer.requested_transfer_at,
       transfer.approved_at, transfer.completed_at
FROM transfer_order transfer
JOIN transfer_order_line line ON line.transfer_id = transfer.transfer_id
JOIN organization owner ON owner.organization_id = transfer.owner_id
JOIN warehouse source ON source.warehouse_id = transfer.source_warehouse_id
JOIN warehouse target ON target.warehouse_id = transfer.target_warehouse_id
JOIN document_status status ON status.status_id = transfer.status_id
JOIN item ON item.item_id = line.item_id
JOIN uom ON uom.uom_id = line.uom_id
LEFT JOIN inventory_lot lot ON lot.lot_id = line.requested_lot_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INVENTORY'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = transfer.owner_id
      AND scope.warehouse_id = transfer.source_warehouse_id
  )
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = transfer.owner_id
      AND scope.warehouse_id = transfer.target_warehouse_id
  )
  AND ($2::uuid IS NULL OR transfer.owner_id = $2)
  AND ($3::uuid IS NULL OR transfer.source_warehouse_id = $3)
  AND ($4::uuid IS NULL OR transfer.target_warehouse_id = $4)
  AND transfer.business_date BETWEEN $5::date AND $6::date
  AND ($7::varchar IS NULL OR status.code = $7)
ORDER BY transfer.business_date DESC, transfer.transfer_id, line.line_no
LIMIT $8 OFFSET $9;

-- 3.6 Inventory status-change and adjustment ledger.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from date, $5 until date, $6 activity type ('STATUS_CHANGE',
-- 'ADJUSTMENT', or NULL), $7 limit, $8 offset
WITH activity AS (
    SELECT change.owner_id, change.warehouse_id, change.business_date,
           change.inventory_status_change_id AS document_id,
           'STATUS_CHANGE'::varchar AS activity_type,
           change_line.line_no, balance.item_id,
           change_line.quantity AS quantity,
           change_line.uom_id, reason.code AS reason_code,
           status.code AS document_status_code,
           change_line.movement_id, change_line.posted_at
    FROM inventory_status_change change
    JOIN inventory_status_change_line change_line
      ON change_line.inventory_status_change_id = change.inventory_status_change_id
    JOIN inventory_balance balance ON balance.balance_id = change_line.source_balance_id
    JOIN reason_code reason ON reason.reason_code_id = change.reason_code_id
    JOIN document_status status ON status.status_id = change.status_id
    WHERE change.business_date BETWEEN $4::date AND $5::date
    UNION ALL
    SELECT adjustment.owner_id, adjustment.warehouse_id, adjustment.business_date,
           adjustment.inventory_adjustment_id, 'ADJUSTMENT'::varchar,
           line.line_no, line.item_id,
           type.quantity_effect * line.adjustment_qty,
           line.uom_id, reason.code, status.code,
           line.movement_id, line.posted_at
    FROM inventory_adjustment adjustment
    JOIN inventory_adjustment_line line
      ON line.inventory_adjustment_id = adjustment.inventory_adjustment_id
    JOIN inventory_adjustment_type type
      ON type.inventory_adjustment_type_id = line.inventory_adjustment_type_id
    JOIN reason_code reason ON reason.reason_code_id = adjustment.reason_code_id
    JOIN document_status status ON status.status_id = adjustment.status_id
    WHERE adjustment.business_date BETWEEN $4::date AND $5::date
)
SELECT activity.business_date, activity.activity_type, activity.document_id,
       activity.line_no, owner.code AS owner_code,
       warehouse.code AS warehouse_code, item.code AS item_code,
       activity.quantity, uom.code AS uom_code, activity.reason_code,
       activity.document_status_code, activity.movement_id, activity.posted_at
FROM activity
JOIN organization owner ON owner.organization_id = activity.owner_id
JOIN warehouse ON warehouse.warehouse_id = activity.warehouse_id
JOIN item ON item.item_id = activity.item_id
JOIN uom ON uom.uom_id = activity.uom_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INVENTORY'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = activity.owner_id
      AND scope.warehouse_id = activity.warehouse_id
  )
  AND ($2::uuid IS NULL OR activity.owner_id = $2)
  AND ($3::uuid IS NULL OR activity.warehouse_id = $3)
  AND ($6::varchar IS NULL OR activity.activity_type = $6)
ORDER BY activity.business_date DESC, activity.document_id, activity.line_no
LIMIT $7 OFFSET $8;

-- 3.7 Stock-count variance report.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from date, $5 until date, $6 variances only,
-- $7 limit, $8 offset
SELECT stocktake.stock_count_id, stocktake.business_date,
       owner.code AS owner_code, warehouse.code AS warehouse_code,
       count_type.code AS count_type_code, status.code AS status_code,
       line.line_no, location.code AS location_code, item.code AS item_code,
       lot.lot_number, line.handling_unit_id,
       inventory_status.code AS inventory_status_code,
       line.system_qty, line.counted_qty,
       CASE WHEN line.counted_qty IS NULL THEN NULL
            ELSE line.counted_qty - line.system_qty END AS variance_qty,
       uom.code AS uom_code, line.counted_at, counter.username AS counted_by,
       line.variance_movement_id, line.posted_at
FROM stock_count stocktake
JOIN stock_count_line line ON line.stock_count_id = stocktake.stock_count_id
JOIN organization owner ON owner.organization_id = stocktake.owner_id
JOIN warehouse ON warehouse.warehouse_id = stocktake.warehouse_id
JOIN stock_count_type count_type
  ON count_type.stock_count_type_id = stocktake.stock_count_type_id
JOIN document_status status ON status.status_id = stocktake.status_id
JOIN warehouse_location location ON location.location_id = line.location_id
JOIN item ON item.item_id = line.item_id
JOIN inventory_status
  ON inventory_status.inventory_status_id = line.inventory_status_id
JOIN uom ON uom.uom_id = line.uom_id
LEFT JOIN inventory_lot lot ON lot.lot_id = line.lot_id
LEFT JOIN app_account counter ON counter.account_id = line.counted_by
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.INVENTORY'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = stocktake.owner_id
      AND scope.warehouse_id = stocktake.warehouse_id
  )
  AND ($2::uuid IS NULL OR stocktake.owner_id = $2)
  AND ($3::uuid IS NULL OR stocktake.warehouse_id = $3)
  AND stocktake.business_date BETWEEN $4::date AND $5::date
  AND (NOT $6 OR COALESCE(line.counted_qty, line.system_qty) <> line.system_qty)
ORDER BY stocktake.business_date DESC, stocktake.stock_count_id, line.line_no
LIMIT $7 OFFSET $8;
