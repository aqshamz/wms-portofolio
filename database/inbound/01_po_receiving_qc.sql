-- Focused inbound queries: purchase order through quality-control result.
-- Execute one numbered operation at a time with prepared parameters.

SET search_path TO wms, public;

-- =============================================================================
-- 1. PURCHASE ORDER
-- =============================================================================

-- 1.1 Create PO in its configured initial status.
-- $1 owner ID, $2 vendor ID, $3 warehouse ID, $4 business date,
-- $5 client PO number, $6 ordered_at, $7 expected arrival, $8 notes,
-- $9 actor account ID
WITH context AS (
    SELECT
        owner.organization_id AS owner_id,
        vendor.partner_id AS vendor_id,
        vendor.code AS vendor_code,
        wh.warehouse_id,
        wh.code AS warehouse_code,
        dt.document_type_id,
        initial_status.status_id
    FROM organization owner
    JOIN business_partner vendor
      ON vendor.owner_id = owner.organization_id
     AND vendor.partner_id = $2
     AND vendor.is_active
    JOIN warehouse_owner wo
      ON wo.owner_id = owner.organization_id
     AND wo.warehouse_id = $3
     AND wo.is_active
    JOIN warehouse wh ON wh.warehouse_id = wo.warehouse_id AND wh.is_active
    JOIN document_type dt ON dt.code = 'PURCHASE_ORDER' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial
     AND initial_status.is_active
    WHERE owner.organization_id = $1
      AND owner.is_active
      AND EXISTS (
          SELECT 1
          FROM business_partner_type vendor_type
          JOIN partner_type type
            ON type.partner_type_id = vendor_type.partner_type_id
          WHERE vendor_type.partner_id = vendor.partner_id
            AND type.code IN ('SUPPLIER', 'FACTORY')
            AND type.is_active
      )
),
generated AS (
    SELECT
        context.*,
        generate_document_id(
            'PURCHASE_ORDER', vendor_code, warehouse_code, $4
        ) AS purchase_order_id
    FROM context
)
INSERT INTO purchase_order (
    purchase_order_id,
    document_type_id,
    status_id,
    owner_id,
    vendor_id,
    warehouse_id,
    business_date,
    purchase_order_no,
    ordered_at,
    expected_arrival_at,
    notes,
    created_by,
    updated_by
)
SELECT
    purchase_order_id,
    document_type_id,
    status_id,
    owner_id,
    vendor_id,
    warehouse_id,
    $4, $5, $6, $7, $8, $9, $9
FROM generated
RETURNING *;

-- 1.2 Add PO line. The PO header lock serializes line-number allocation.
-- $1 PO ID, $2 item ID, $3 ordered quantity, $4 UOM ID,
-- $5 vendor item code, $6 expected lot, $7 expected expiry,
-- $8 notes, $9 actor ID
WITH locked_po AS (
    SELECT po.purchase_order_id, po.owner_id
    FROM purchase_order po
    JOIN document_status status ON status.status_id = po.status_id
    WHERE po.purchase_order_id = $1
      AND status.code = 'DRAFT'
    FOR UPDATE OF po
),
next_line AS (
    SELECT
        locked_po.purchase_order_id,
        locked_po.owner_id,
        COALESCE(max(existing.line_no), 0) + 1 AS line_no
    FROM locked_po
    LEFT JOIN purchase_order_line existing
      ON existing.purchase_order_id = locked_po.purchase_order_id
    GROUP BY locked_po.purchase_order_id, locked_po.owner_id
)
INSERT INTO purchase_order_line (
    purchase_order_line_id,
    purchase_order_id,
    owner_id,
    line_no,
    item_id,
    ordered_qty,
    uom_id,
    vendor_item_code,
    expected_lot_no,
    expected_expiry_date,
    notes,
    created_by
)
SELECT
    next_line.purchase_order_id || '-L' || lpad(next_line.line_no::text, 4, '0'),
    next_line.purchase_order_id,
    next_line.owner_id,
    next_line.line_no,
    item.item_id,
    $3, $4, $5, $6, $7, $8, $9
FROM next_line
JOIN item
  ON item.owner_id = next_line.owner_id
 AND item.item_id = $2
 AND item.is_active
JOIN item_uom
  ON item_uom.item_id = item.item_id
 AND item_uom.uom_id = $4
 AND item_uom.is_active
RETURNING *;

-- 1.3 Approve PO through configured workflow.
-- Authorization must already confirm INBOUND.PO.APPROVE.
-- $1 PO ID, $2 actor account ID, $3 expected version
UPDATE purchase_order po
SET status_id  = target.status_id,
    updated_at = clock_timestamp(),
    updated_by = $2,
    version_no = po.version_no + 1
FROM document_status target
WHERE po.purchase_order_id = $1
  AND po.version_no = $3
  AND target.document_type_id = po.document_type_id
  AND target.code = 'APPROVED'
  AND target.is_active
  AND EXISTS (
      SELECT 1
      FROM document_status_transition transition
      WHERE transition.document_type_id = po.document_type_id
        AND transition.from_status_id = po.status_id
        AND transition.to_status_id = target.status_id
        AND transition.is_active
  )
  AND EXISTS (
      SELECT 1
      FROM purchase_order_line line
      WHERE line.purchase_order_id = po.purchase_order_id
  )
RETURNING po.*;

-- 1.4 Get PO with ordered/received quantities. $1 PO ID
SELECT
    po.purchase_order_id,
    po.purchase_order_no,
    po.business_date,
    po.ordered_at,
    po.expected_arrival_at,
    status.code AS status_code,
    owner.code AS owner_code,
    vendor.code AS vendor_code,
    vendor.name AS vendor_name,
    warehouse.code AS warehouse_code,
    po.version_no,
    lines.lines
FROM purchase_order po
JOIN document_status status ON status.status_id = po.status_id
JOIN organization owner ON owner.organization_id = po.owner_id
JOIN business_partner vendor ON vendor.partner_id = po.vendor_id
JOIN warehouse ON warehouse.warehouse_id = po.warehouse_id
LEFT JOIN LATERAL (
    SELECT jsonb_agg(
        jsonb_build_object(
            'lineId', pol.purchase_order_line_id,
            'lineNo', pol.line_no,
            'itemId', pol.item_id,
            'itemCode', item.code,
            'itemName', item.name,
            'orderedQty', pol.ordered_qty,
            'receivedQty', COALESCE(received.received_qty, 0),
            'uomCode', unit.code
        ) ORDER BY pol.line_no
    ) AS lines
    FROM purchase_order_line pol
    JOIN item ON item.item_id = pol.item_id
    JOIN uom unit ON unit.uom_id = pol.uom_id
    LEFT JOIN LATERAL (
        SELECT sum(rl.received_qty) AS received_qty
        FROM inbound_order_line iol
        JOIN receipt_line rl ON rl.inbound_line_id = iol.inbound_line_id
        WHERE iol.purchase_order_line_id = pol.purchase_order_line_id
    ) received ON true
    WHERE pol.purchase_order_id = po.purchase_order_id
) lines ON true
WHERE po.purchase_order_id = $1;

-- 1.5 Search PO menu.
-- $1 owner ID, $2 warehouse ID or null, $3 status code or null,
-- $4 search or null, $5 limit, $6 offset
SELECT
    po.purchase_order_id,
    po.purchase_order_no,
    po.business_date,
    vendor.code AS vendor_code,
    vendor.name AS vendor_name,
    warehouse.code AS warehouse_code,
    status.code AS status_code,
    po.expected_arrival_at,
    po.version_no,
    count(*) OVER () AS total_rows
FROM purchase_order po
JOIN business_partner vendor ON vendor.partner_id = po.vendor_id
JOIN warehouse ON warehouse.warehouse_id = po.warehouse_id
JOIN document_status status ON status.status_id = po.status_id
WHERE po.owner_id = $1
  AND ($2 IS NULL OR po.warehouse_id = $2)
  AND ($3 IS NULL OR status.code = $3)
  AND (
      $4 IS NULL OR
      po.purchase_order_id ILIKE '%' || $4 || '%' OR
      po.purchase_order_no ILIKE '%' || $4 || '%' OR
      vendor.name ILIKE '%' || $4 || '%'
  )
ORDER BY po.business_date DESC, po.purchase_order_id DESC
LIMIT $5 OFFSET $6;

-- =============================================================================
-- 2. INBOUND DELIVERY / ASN
-- =============================================================================

-- 2.1 Create an inbound delivery from an approved PO.
-- $1 PO ID, $2 business date, $3 expected arrival,
-- $4 external delivery reference, $5 supplier reference, $6 notes, $7 actor ID
WITH po_context AS (
    SELECT
        po.owner_id,
        po.vendor_id,
        vendor.code AS vendor_code,
        po.warehouse_id,
        warehouse.code AS warehouse_code,
        dt.document_type_id,
        initial_status.status_id
    FROM purchase_order po
    JOIN document_status po_status ON po_status.status_id = po.status_id
    JOIN business_partner vendor ON vendor.partner_id = po.vendor_id
    JOIN warehouse ON warehouse.warehouse_id = po.warehouse_id
    JOIN document_type dt ON dt.code = 'INBOUND' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial
     AND initial_status.is_active
    WHERE po.purchase_order_id = $1
      AND po_status.code IN ('APPROVED', 'PARTIALLY_RECEIVED')
),
generated AS (
    SELECT
        po_context.*,
        generate_document_id(
            'INBOUND', vendor_code, warehouse_code, $2
        ) AS inbound_id
    FROM po_context
)
INSERT INTO inbound_order (
    inbound_id,
    document_type_id,
    status_id,
    owner_id,
    vendor_id,
    warehouse_id,
    business_date,
    expected_arrival_at,
    external_reference,
    supplier_reference,
    notes,
    created_by,
    updated_by
)
SELECT
    inbound_id,
    document_type_id,
    status_id,
    owner_id,
    vendor_id,
    warehouse_id,
    $2, $3, $4, $5, $6, $7, $7
FROM generated
RETURNING *;

-- 2.2 Add delivery line from PO line without exceeding still-unplanned quantity.
-- $1 inbound ID, $2 PO line ID, $3 expected delivery quantity, $4 actor ID
WITH locked_inbound AS (
    SELECT inbound.*
    FROM inbound_order inbound
    JOIN document_status status ON status.status_id = inbound.status_id
    WHERE inbound.inbound_id = $1
      AND status.code = 'DRAFT'
    FOR UPDATE OF inbound
),
po_line AS (
    SELECT pol.*
    FROM purchase_order_line pol
    JOIN purchase_order po ON po.purchase_order_id = pol.purchase_order_id
    JOIN locked_inbound inbound
      ON inbound.owner_id = po.owner_id
     AND inbound.vendor_id = po.vendor_id
     AND inbound.warehouse_id = po.warehouse_id
    WHERE pol.purchase_order_line_id = $2
),
next_line AS (
    SELECT
        inbound.inbound_id,
        COALESCE(max(existing.line_no), 0) + 1 AS line_no
    FROM locked_inbound inbound
    LEFT JOIN inbound_order_line existing ON existing.inbound_id = inbound.inbound_id
    GROUP BY inbound.inbound_id
),
remaining AS (
    SELECT
        po_line.*,
        po_line.ordered_qty - COALESCE((
            SELECT sum(existing.expected_qty)
            FROM inbound_order_line existing
            WHERE existing.purchase_order_line_id = po_line.purchase_order_line_id
        ), 0) AS remaining_qty
    FROM po_line
)
INSERT INTO inbound_order_line (
    inbound_line_id,
    inbound_id,
    purchase_order_line_id,
    line_no,
    item_id,
    expected_qty,
    uom_id,
    expected_lot_no,
    expected_expiry_date,
    customer_line_reference,
    created_by
)
SELECT
    next_line.inbound_id || '-L' || lpad(next_line.line_no::text, 4, '0'),
    next_line.inbound_id,
    remaining.purchase_order_line_id,
    next_line.line_no,
    remaining.item_id,
    $3,
    remaining.uom_id,
    remaining.expected_lot_no,
    remaining.expected_expiry_date,
    remaining.purchase_order_line_id,
    $4
FROM next_line
CROSS JOIN remaining
WHERE $3 > 0
  AND $3 <= remaining.remaining_qty
RETURNING *;

-- 2.3 Release inbound delivery. $1 inbound ID, $2 actor ID, $3 expected version
UPDATE inbound_order inbound
SET status_id  = target.status_id,
    updated_at = clock_timestamp(),
    updated_by = $2,
    version_no = inbound.version_no + 1
FROM document_status target
WHERE inbound.inbound_id = $1
  AND inbound.version_no = $3
  AND target.document_type_id = inbound.document_type_id
  AND target.code = 'RELEASED'
  AND EXISTS (
      SELECT 1
      FROM document_status_transition transition
      WHERE transition.document_type_id = inbound.document_type_id
        AND transition.from_status_id = inbound.status_id
        AND transition.to_status_id = target.status_id
        AND transition.is_active
  )
  AND EXISTS (
      SELECT 1 FROM inbound_order_line line WHERE line.inbound_id = inbound.inbound_id
  )
RETURNING inbound.*;

-- 2.4 Search inbound-delivery menu.
-- $1 owner ID, $2 warehouse ID or null, $3 status code or null,
-- $4 date from, $5 date to, $6 limit, $7 offset
SELECT
    inbound.inbound_id,
    inbound.business_date,
    vendor.code AS vendor_code,
    vendor.name AS vendor_name,
    warehouse.code AS warehouse_code,
    status.code AS status_code,
    inbound.expected_arrival_at,
    inbound.supplier_reference,
    count(*) OVER () AS total_rows
FROM inbound_order inbound
JOIN business_partner vendor ON vendor.partner_id = inbound.vendor_id
JOIN warehouse ON warehouse.warehouse_id = inbound.warehouse_id
JOIN document_status status ON status.status_id = inbound.status_id
WHERE inbound.owner_id = $1
  AND ($2 IS NULL OR inbound.warehouse_id = $2)
  AND ($3 IS NULL OR status.code = $3)
  AND inbound.business_date BETWEEN $4 AND $5
ORDER BY inbound.business_date DESC, inbound.inbound_id DESC
LIMIT $6 OFFSET $7;

-- =============================================================================
-- 3. PHYSICAL RECEIPT
-- =============================================================================

-- 3.1 Create receipt for released delivery.
-- $1 inbound ID, $2 business date, $3 received_at, $4 dock location ID,
-- $5 vehicle number, $6 seal number, $7 delivery note, $8 notes, $9 actor ID
WITH inbound_context AS (
    SELECT
        inbound.*,
        vendor.code AS vendor_code,
        warehouse.code AS warehouse_code,
        dt.document_type_id AS receipt_document_type_id,
        initial_status.status_id AS receipt_status_id
    FROM inbound_order inbound
    JOIN document_status inbound_status ON inbound_status.status_id = inbound.status_id
    JOIN business_partner vendor ON vendor.partner_id = inbound.vendor_id
    JOIN warehouse ON warehouse.warehouse_id = inbound.warehouse_id
    JOIN warehouse_location dock
      ON dock.location_id = $4
     AND dock.warehouse_id = inbound.warehouse_id
     AND dock.is_active
    JOIN location_type dock_type
      ON dock_type.location_type_id = dock.location_type_id
     AND dock_type.allows_receiving
    JOIN document_type dt ON dt.code = 'RECEIPT' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial
     AND initial_status.is_active
    WHERE inbound.inbound_id = $1
      AND inbound_status.code IN ('RELEASED', 'PARTIALLY_RECEIVED')
),
generated AS (
    SELECT
        inbound_context.*,
        generate_document_id(
            'RECEIPT', vendor_code, warehouse_code, $2
        ) AS receipt_id
    FROM inbound_context
)
INSERT INTO receipt (
    receipt_id,
    document_type_id,
    status_id,
    inbound_id,
    owner_id,
    warehouse_id,
    business_date,
    received_at,
    dock_location_id,
    vehicle_number,
    seal_number,
    delivery_note_no,
    notes,
    created_by
)
SELECT
    receipt_id,
    receipt_document_type_id,
    receipt_status_id,
    inbound_id,
    owner_id,
    warehouse_id,
    $2, $3, $4, $5, $6, $7, $8, $9
FROM generated
RETURNING *;

-- 3.2 Add physical receipt line without exceeding delivery expectation.
-- $1 receipt ID, $2 inbound line ID, $3 received quantity,
-- $4 rejected-at-dock quantity, $5 actor ID
WITH locked_receipt AS (
    SELECT receipt.*
    FROM receipt
    JOIN document_status status ON status.status_id = receipt.status_id
    WHERE receipt.receipt_id = $1
      AND status.code = 'OPEN'
    FOR UPDATE OF receipt
),
delivery_line AS (
    SELECT line.*
    FROM inbound_order_line line
    JOIN locked_receipt receipt ON receipt.inbound_id = line.inbound_id
    WHERE line.inbound_line_id = $2
),
remaining AS (
    SELECT
        delivery_line.*,
        delivery_line.expected_qty - COALESCE((
            SELECT sum(existing.received_qty)
            FROM receipt_line existing
            WHERE existing.inbound_line_id = delivery_line.inbound_line_id
        ), 0) AS remaining_qty
    FROM delivery_line
),
next_line AS (
    SELECT
        receipt.receipt_id,
        COALESCE(max(existing.line_no), 0) + 1 AS line_no
    FROM locked_receipt receipt
    LEFT JOIN receipt_line existing ON existing.receipt_id = receipt.receipt_id
    GROUP BY receipt.receipt_id
)
INSERT INTO receipt_line (
    receipt_line_id,
    receipt_id,
    inbound_line_id,
    line_no,
    item_id,
    received_qty,
    rejected_qty,
    uom_id,
    created_by
)
SELECT
    next_line.receipt_id || '-L' || lpad(next_line.line_no::text, 4, '0'),
    next_line.receipt_id,
    remaining.inbound_line_id,
    next_line.line_no,
    remaining.item_id,
    $3,
    $4,
    remaining.uom_id,
    $5
FROM next_line
CROSS JOIN remaining
WHERE $3 > 0
  AND $4 >= 0
  AND $4 <= $3
  AND $3 <= remaining.remaining_qty
RETURNING *;

-- 3.3 Create or resolve an item lot.
-- $1 owner ID, $2 item ID, $3 lot number, $4 manufacture date,
-- $5 expiry date, $6 quality-status code, $7 business date, $8 actor ID
WITH context AS (
    SELECT
        owner.organization_id AS owner_id,
        item.item_id,
        quality_status.quality_status_id,
        generate_document_id('LOT', NULL, NULL, $7) AS lot_id
    FROM organization owner
    JOIN item
      ON item.owner_id = owner.organization_id
     AND item.item_id = $2
     AND item.lot_controlled
    JOIN quality_status ON quality_status.code = $6 AND quality_status.is_active
    WHERE owner.organization_id = $1
)
INSERT INTO inventory_lot (
    lot_id, owner_id, item_id, lot_number,
    manufacture_date, expiry_date, quality_status_id, created_by
)
SELECT
    lot_id, owner_id, item_id, $3, $4, $5, quality_status_id, $8
FROM context
ON CONFLICT (owner_id, item_id, lot_number)
DO UPDATE SET lot_number = EXCLUDED.lot_number
RETURNING *;

-- 3.4 Create handling unit.
-- $1 warehouse ID, $2 owner ID, $3 handling-unit type code,
-- $4 barcode, $5 parent HU ID or null, $6 current location ID or null,
-- $7 business date, $8 actor ID
WITH context AS (
    SELECT
        warehouse.warehouse_id,
        warehouse.code AS warehouse_code,
        warehouse_owner.owner_id,
        handling_unit_type.handling_unit_type_id,
        generate_document_id(
            'HANDLING_UNIT', NULL, warehouse.code, $7
        ) AS handling_unit_id
    FROM warehouse
    JOIN warehouse_owner
      ON warehouse_owner.warehouse_id = warehouse.warehouse_id
     AND warehouse_owner.owner_id = $2
     AND warehouse_owner.is_active
    JOIN handling_unit_type
      ON handling_unit_type.code = $3
     AND handling_unit_type.is_active
    WHERE warehouse.warehouse_id = $1
      AND ($5 IS NULL OR EXISTS (
          SELECT 1
          FROM handling_unit parent
          WHERE parent.handling_unit_id = $5
            AND parent.warehouse_id = warehouse.warehouse_id
            AND parent.owner_id = warehouse_owner.owner_id
      ))
      AND ($6 IS NULL OR EXISTS (
          SELECT 1
          FROM warehouse_location location
          WHERE location.location_id = $6
            AND location.warehouse_id = warehouse.warehouse_id
      ))
)
INSERT INTO handling_unit (
    handling_unit_id,
    warehouse_id,
    owner_id,
    handling_unit_type_id,
    parent_handling_unit_id,
    current_location_id,
    barcode,
    created_by
)
SELECT
    handling_unit_id,
    warehouse_id,
    owner_id,
    handling_unit_type_id,
    $5,
    $6,
    $4,
    $8
FROM context
RETURNING *;

-- 3.5 Register one homogeneous received batch, increase QC_PENDING balance,
-- and append RECEIVE movement atomically.
-- $1 receipt-inventory ID, $2 proposed balance ID, $3 movement ID,
-- $4 receipt line ID, $5 source quantity in receipt-line UOM,
-- $6 lot ID or null, $7 handling-unit ID or null, $8 staging location ID,
-- $9 actor ID
BEGIN;

SELECT receipt_line_id
FROM receipt_line
WHERE receipt_line_id = $4
FOR UPDATE;

WITH locked_line AS (
    SELECT
        rl.receipt_line_id,
        rl.receipt_id,
        rl.item_id,
        rl.uom_id AS source_uom_id,
        rl.received_qty - rl.rejected_qty AS receivable_qty,
        receipt.owner_id,
        receipt.warehouse_id,
        item.base_uom_id,
        item.lot_controlled,
        item.serial_controlled,
        item_uom.conversion_to_base,
        status.inventory_status_id,
        location.location_id,
        $5::numeric * item_uom.conversion_to_base AS base_qty
    FROM receipt_line rl
    JOIN receipt ON receipt.receipt_id = rl.receipt_id
    JOIN document_status receipt_status ON receipt_status.status_id = receipt.status_id
    JOIN item ON item.item_id = rl.item_id AND item.owner_id = receipt.owner_id
    JOIN item_uom
      ON item_uom.item_id = rl.item_id
     AND item_uom.uom_id = rl.uom_id
     AND item_uom.is_active
    JOIN inventory_status status ON status.code = 'QC_PENDING' AND status.is_active
    JOIN warehouse_location location
      ON location.location_id = $8
     AND location.warehouse_id = receipt.warehouse_id
     AND location.is_active
    WHERE rl.receipt_line_id = $4
      AND receipt_status.code = 'OPEN'
    FOR UPDATE OF rl
),
available_batch AS (
    SELECT locked_line.*
    FROM locked_line
    WHERE $5 > 0
      AND (NOT locked_line.lot_controlled OR $6 IS NOT NULL)
      AND $5 <= locked_line.receivable_qty - COALESCE((
          SELECT sum(existing.source_qty)
          FROM receipt_inventory existing
          WHERE existing.receipt_line_id = locked_line.receipt_line_id
      ), 0)
      AND (
          $6 IS NULL OR EXISTS (
              SELECT 1
              FROM inventory_lot lot
              WHERE lot.lot_id = $6
                AND lot.owner_id = locked_line.owner_id
                AND lot.item_id = locked_line.item_id
          )
      )
      AND (
          $7 IS NULL OR EXISTS (
              SELECT 1
              FROM handling_unit hu
              WHERE hu.handling_unit_id = $7
                AND hu.owner_id = locked_line.owner_id
                AND hu.warehouse_id = locked_line.warehouse_id
          )
      )
),
upserted_balance AS (
    INSERT INTO inventory_balance (
        balance_id,
        owner_id,
        warehouse_id,
        location_id,
        item_id,
        lot_id,
        handling_unit_id,
        inventory_status_id,
        on_hand_qty,
        reserved_qty,
        uom_id
    )
    SELECT
        $2,
        owner_id,
        warehouse_id,
        location_id,
        item_id,
        $6,
        $7,
        inventory_status_id,
        base_qty,
        0,
        base_uom_id
    FROM available_batch
    ON CONFLICT (
        owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id
    ) DO UPDATE SET
        on_hand_qty = inventory_balance.on_hand_qty + EXCLUDED.on_hand_qty,
        version_no = inventory_balance.version_no + 1,
        updated_at = clock_timestamp()
    RETURNING balance_id
),
created_batch AS (
    INSERT INTO receipt_inventory (
        receipt_inventory_id,
        receipt_line_id,
        item_id,
        source_qty,
        source_uom_id,
        base_qty,
        base_uom_id,
        lot_id,
        handling_unit_id,
        received_location_id,
        initial_inventory_status_id,
        initial_balance_id,
        created_by
    )
    SELECT
        $1,
        available_batch.receipt_line_id,
        available_batch.item_id,
        $5,
        available_batch.source_uom_id,
        available_batch.base_qty,
        available_batch.base_uom_id,
        $6,
        $7,
        available_batch.location_id,
        available_batch.inventory_status_id,
        upserted_balance.balance_id,
        $9
    FROM available_batch
    CROSS JOIN upserted_balance
    RETURNING *
),
created_movement AS (
    INSERT INTO inventory_movement (
        movement_id,
        movement_type_id,
        owner_id,
        warehouse_id,
        business_date,
        item_id,
        lot_id,
        handling_unit_id,
        to_location_id,
        to_status_id,
        quantity,
        uom_id,
        source_document_id,
        source_line_id,
        created_by
    )
    SELECT
        $3,
        movement_type.movement_type_id,
        available_batch.owner_id,
        available_batch.warehouse_id,
        receipt.business_date,
        available_batch.item_id,
        $6,
        $7,
        available_batch.location_id,
        available_batch.inventory_status_id,
        available_batch.base_qty,
        available_batch.base_uom_id,
        available_batch.receipt_id,
        created_batch.receipt_inventory_id,
        $9
    FROM available_batch
    CROSS JOIN created_batch
    JOIN receipt ON receipt.receipt_id = available_batch.receipt_id
    JOIN movement_type ON movement_type.code = 'RECEIVE' AND movement_type.is_active
    RETURNING movement_id
)
SELECT
    created_batch.*,
    upserted_balance.balance_id,
    created_movement.movement_id
FROM created_batch
CROSS JOIN upserted_balance
CROSS JOIN created_movement;

COMMIT;

-- 3.6 Register one serial number against a serial-controlled received batch.
-- $1 generated serial ID, $2 receipt-inventory ID, $3 serial number, $4 actor ID
WITH batch AS (
    SELECT
        receipt_inventory.receipt_inventory_id,
        receipt_inventory.item_id,
        receipt.owner_id
    FROM receipt_inventory
    JOIN receipt_line ON receipt_line.receipt_line_id = receipt_inventory.receipt_line_id
    JOIN receipt ON receipt.receipt_id = receipt_line.receipt_id
    JOIN item
      ON item.item_id = receipt_inventory.item_id
     AND item.serial_controlled
    WHERE receipt_inventory.receipt_inventory_id = $2
),
created_serial AS (
    INSERT INTO serial_number (
        serial_id, owner_id, item_id, serial_no, created_by
    )
    SELECT $1, owner_id, item_id, $3, $4
    FROM batch
    RETURNING serial_id
)
INSERT INTO receipt_line_serial (receipt_inventory_id, serial_id)
SELECT batch.receipt_inventory_id, created_serial.serial_id
FROM batch
CROSS JOIN created_serial
RETURNING *;

-- 3.7 Complete physical receipt and recalculate inbound/PO receipt progress.
-- This does not wait for QC; QC controls inventory availability separately.
-- $1 receipt ID, $2 actor account ID
WITH eligible_receipt AS (
    SELECT receipt.*
    FROM receipt
    JOIN document_status status ON status.status_id = receipt.status_id
    WHERE receipt.receipt_id = $1
      AND status.code = 'OPEN'
      AND EXISTS (
          SELECT 1 FROM receipt_line line WHERE line.receipt_id = receipt.receipt_id
      )
      AND NOT EXISTS (
          SELECT 1
          FROM receipt_line line
          WHERE line.receipt_id = receipt.receipt_id
            AND COALESCE((
                SELECT sum(batch.source_qty)
                FROM receipt_inventory batch
                WHERE batch.receipt_line_id = line.receipt_line_id
            ), 0) <> line.received_qty - line.rejected_qty
      )
),
completed_receipt AS (
    UPDATE receipt
    SET status_id = completed_status.status_id
    FROM eligible_receipt eligible
    JOIN document_status completed_status
      ON completed_status.document_type_id = eligible.document_type_id
     AND completed_status.code = 'COMPLETED'
    WHERE receipt.receipt_id = eligible.receipt_id
      AND EXISTS (
          SELECT 1
          FROM document_status_transition transition
          WHERE transition.document_type_id = receipt.document_type_id
            AND transition.from_status_id = receipt.status_id
            AND transition.to_status_id = completed_status.status_id
            AND transition.is_active
      )
    RETURNING receipt.*
),
inbound_progress AS (
    SELECT
        inbound.inbound_id,
        inbound.document_type_id,
        inbound.status_id,
        sum(delivery_line.expected_qty) AS expected_qty,
        COALESCE(sum(received.received_qty), 0) AS received_qty
    FROM inbound_order inbound
    JOIN completed_receipt completed ON completed.inbound_id = inbound.inbound_id
    JOIN inbound_order_line delivery_line ON delivery_line.inbound_id = inbound.inbound_id
    LEFT JOIN LATERAL (
        SELECT sum(line.received_qty) AS received_qty
        FROM receipt_line line
        JOIN receipt receipt_header ON receipt_header.receipt_id = line.receipt_id
        JOIN document_status receipt_status ON receipt_status.status_id = receipt_header.status_id
        WHERE line.inbound_line_id = delivery_line.inbound_line_id
          AND receipt_status.code = 'COMPLETED'
    ) received ON true
    GROUP BY inbound.inbound_id
),
updated_inbound AS (
    UPDATE inbound_order inbound
    SET status_id = target_status.status_id,
        updated_at = clock_timestamp(),
        updated_by = $2,
        version_no = inbound.version_no + 1
    FROM inbound_progress progress
    JOIN document_status target_status
      ON target_status.document_type_id = progress.document_type_id
     AND target_status.code = CASE
         WHEN progress.received_qty >= progress.expected_qty THEN 'RECEIVED'
         ELSE 'PARTIALLY_RECEIVED'
     END
    WHERE inbound.inbound_id = progress.inbound_id
      AND inbound.status_id <> target_status.status_id
      AND EXISTS (
          SELECT 1
          FROM document_status_transition transition
          WHERE transition.document_type_id = inbound.document_type_id
            AND transition.from_status_id = inbound.status_id
            AND transition.to_status_id = target_status.status_id
            AND transition.is_active
      )
    RETURNING inbound.*
),
affected_purchase_orders AS (
    SELECT DISTINCT po.purchase_order_id
    FROM completed_receipt completed
    JOIN receipt_line line ON line.receipt_id = completed.receipt_id
    JOIN inbound_order_line delivery_line ON delivery_line.inbound_line_id = line.inbound_line_id
    JOIN purchase_order_line po_line
      ON po_line.purchase_order_line_id = delivery_line.purchase_order_line_id
    JOIN purchase_order po ON po.purchase_order_id = po_line.purchase_order_id
),
purchase_order_progress AS (
    SELECT
        po.purchase_order_id,
        po.document_type_id,
        sum(po_line.ordered_qty) AS ordered_qty,
        COALESCE(sum(received.received_qty), 0) AS received_qty
    FROM affected_purchase_orders affected
    JOIN purchase_order po ON po.purchase_order_id = affected.purchase_order_id
    JOIN purchase_order_line po_line ON po_line.purchase_order_id = po.purchase_order_id
    LEFT JOIN LATERAL (
        SELECT sum(receipt_line.received_qty) AS received_qty
        FROM inbound_order_line delivery_line
        JOIN receipt_line ON receipt_line.inbound_line_id = delivery_line.inbound_line_id
        JOIN receipt receipt_header ON receipt_header.receipt_id = receipt_line.receipt_id
        JOIN document_status receipt_status ON receipt_status.status_id = receipt_header.status_id
        WHERE delivery_line.purchase_order_line_id = po_line.purchase_order_line_id
          AND receipt_status.code = 'COMPLETED'
    ) received ON true
    GROUP BY po.purchase_order_id
),
updated_purchase_orders AS (
    UPDATE purchase_order po
    SET status_id = target_status.status_id,
        updated_at = clock_timestamp(),
        updated_by = $2,
        version_no = po.version_no + 1
    FROM purchase_order_progress progress
    JOIN document_status target_status
      ON target_status.document_type_id = progress.document_type_id
     AND target_status.code = CASE
         WHEN progress.received_qty >= progress.ordered_qty THEN 'RECEIVED'
         ELSE 'PARTIALLY_RECEIVED'
     END
    WHERE po.purchase_order_id = progress.purchase_order_id
      AND po.status_id <> target_status.status_id
      AND EXISTS (
          SELECT 1
          FROM document_status_transition transition
          WHERE transition.document_type_id = po.document_type_id
            AND transition.from_status_id = po.status_id
            AND transition.to_status_id = target_status.status_id
            AND transition.is_active
      )
    RETURNING po.purchase_order_id, po.status_id
)
SELECT
    completed_receipt.receipt_id,
    completed_receipt.status_id AS receipt_status_id,
    updated_inbound.inbound_id,
    updated_inbound.status_id AS inbound_status_id,
    COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object(
                'purchaseOrderId', updated_purchase_orders.purchase_order_id,
                'statusId', updated_purchase_orders.status_id
            )
        )
        FROM updated_purchase_orders
    ), '[]'::jsonb) AS purchase_orders
FROM completed_receipt
LEFT JOIN updated_inbound ON updated_inbound.inbound_id = completed_receipt.inbound_id;

-- 3.8 Search receipt menu.
-- $1 owner ID, $2 warehouse ID or null, $3 status code or null,
-- $4 date from, $5 date to, $6 limit, $7 offset
SELECT
    receipt.receipt_id,
    receipt.inbound_id,
    receipt.business_date,
    receipt.received_at,
    status.code AS status_code,
    warehouse.code AS warehouse_code,
    receipt.delivery_note_no,
    receipt.vehicle_number,
    count(*) OVER () AS total_rows
FROM receipt
JOIN document_status status ON status.status_id = receipt.status_id
JOIN warehouse ON warehouse.warehouse_id = receipt.warehouse_id
WHERE receipt.owner_id = $1
  AND ($2 IS NULL OR receipt.warehouse_id = $2)
  AND ($3 IS NULL OR status.code = $3)
  AND receipt.business_date BETWEEN $4 AND $5
ORDER BY receipt.received_at DESC, receipt.receipt_id DESC
LIMIT $6 OFFSET $7;

-- =============================================================================
-- 4. QUALITY CONTROL
-- =============================================================================

-- 4.0 QC queue. $1 owner ID, $2 warehouse ID or null, $3 limit, $4 offset
SELECT
    batch.receipt_inventory_id,
    receipt.receipt_id,
    item.code AS item_code,
    item.name AS item_name,
    lot.lot_number,
    hu.barcode AS handling_unit_barcode,
    batch.base_qty,
    base_uom.code AS base_uom_code,
    location.code AS current_location_code,
    latest.inspection_id,
    latest.quality_status_code,
    count(*) OVER () AS total_rows
FROM receipt_inventory batch
JOIN receipt_line line ON line.receipt_line_id = batch.receipt_line_id
JOIN receipt ON receipt.receipt_id = line.receipt_id
JOIN item ON item.item_id = batch.item_id
LEFT JOIN inventory_lot lot ON lot.lot_id = batch.lot_id
LEFT JOIN handling_unit hu ON hu.handling_unit_id = batch.handling_unit_id
JOIN uom base_uom ON base_uom.uom_id = batch.base_uom_id
JOIN warehouse_location location ON location.location_id = batch.received_location_id
LEFT JOIN LATERAL (
    SELECT
        inspection.inspection_id,
        quality_status.code AS quality_status_code,
        inspection.inspection_result_id
    FROM quality_inspection inspection
    JOIN quality_status ON quality_status.quality_status_id = inspection.quality_status_id
    WHERE inspection.receipt_inventory_id = batch.receipt_inventory_id
    ORDER BY inspection.created_at DESC
    LIMIT 1
) latest ON true
WHERE receipt.owner_id = $1
  AND ($2 IS NULL OR receipt.warehouse_id = $2)
  AND (latest.inspection_id IS NULL OR latest.inspection_result_id IS NULL)
ORDER BY receipt.received_at, batch.receipt_inventory_id
LIMIT $3 OFFSET $4;

-- 4.1 Start QC for a received batch.
-- $1 inspection ID, $2 receipt-inventory ID, $3 actor ID,
-- $4 parent inspection ID or null (used after rework)
INSERT INTO quality_inspection (
    inspection_id,
    receipt_inventory_id,
    parent_inspection_id,
    quality_status_id,
    inspected_qty,
    created_by
)
SELECT
    $1,
    batch.receipt_inventory_id,
    $4,
    quality_status.quality_status_id,
    batch.base_qty,
    $3
FROM receipt_inventory batch
JOIN item ON item.item_id = batch.item_id
JOIN quality_status ON quality_status.code = 'PENDING' AND quality_status.is_active
WHERE batch.receipt_inventory_id = $2
  AND (
      NOT item.serial_controlled OR
      (
          SELECT count(*)::numeric
          FROM receipt_line_serial serial
          WHERE serial.receipt_inventory_id = batch.receipt_inventory_id
      ) = batch.base_qty
  )
  AND (
      $4 IS NOT NULL OR NOT EXISTS (
          SELECT 1
          FROM quality_inspection existing
          WHERE existing.receipt_inventory_id = batch.receipt_inventory_id
      )
  )
RETURNING *;

-- 4.2 QC PASS: move all inspected stock from QC_PENDING to AVAILABLE and create
-- putaway task. This single statement is atomic.
-- $1 inspection ID, $2 QC-pending source balance ID,
-- $3 proposed available balance ID, $4 movement ID, $5 putaway-task ID,
-- $6 target storage location ID, $7 task-priority code, $8 actor ID
BEGIN;

SELECT inspection_id
FROM quality_inspection
WHERE inspection_id = $1
  AND inspection_result_id IS NULL
FOR UPDATE;

WITH inspection_context AS (
    SELECT
        inspection.*,
        batch.receipt_inventory_id,
        batch.lot_id,
        batch.handling_unit_id,
        batch.base_uom_id,
        receipt.owner_id,
        receipt.warehouse_id,
        receipt.business_date,
        batch.item_id
    FROM quality_inspection inspection
    JOIN receipt_inventory batch
      ON batch.receipt_inventory_id = inspection.receipt_inventory_id
    JOIN receipt_line line ON line.receipt_line_id = batch.receipt_line_id
    JOIN receipt ON receipt.receipt_id = line.receipt_id
    WHERE inspection.inspection_id = $1
      AND inspection.inspection_result_id IS NULL
),
decremented_source AS (
    UPDATE inventory_balance source
    SET on_hand_qty = source.on_hand_qty - inspection.inspected_qty,
        version_no  = source.version_no + 1,
        updated_at  = clock_timestamp()
    FROM inspection_context inspection
    JOIN inventory_status pending_status ON pending_status.code = 'QC_PENDING'
    WHERE source.balance_id = $2
      AND source.inventory_status_id = pending_status.inventory_status_id
      AND source.item_id = inspection.item_id
      AND source.on_hand_qty - source.reserved_qty >= inspection.inspected_qty
    RETURNING source.*
),
available_balance AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT
        $3,
        source.owner_id,
        source.warehouse_id,
        source.location_id,
        source.item_id,
        source.lot_id,
        source.handling_unit_id,
        available_status.inventory_status_id,
        inspection.inspected_qty,
        0,
        source.uom_id
    FROM decremented_source source
    CROSS JOIN inspection_context inspection
    JOIN inventory_status available_status
      ON available_status.code = 'AVAILABLE'
     AND available_status.is_active
    ON CONFLICT (
        owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id
    ) DO UPDATE SET
        on_hand_qty = inventory_balance.on_hand_qty + EXCLUDED.on_hand_qty,
        version_no  = inventory_balance.version_no + 1,
        updated_at  = clock_timestamp()
    RETURNING *
),
completed_inspection AS (
    UPDATE quality_inspection inspection
    SET quality_status_id = quality_status.quality_status_id,
        inspection_result_id = result.inspection_result_id,
        passed_qty = inspection.inspected_qty,
        failed_qty = 0,
        inspected_at = clock_timestamp(),
        inspected_by = $8
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
        $4,
        movement_type.movement_type_id,
        source.owner_id,
        source.warehouse_id,
        inspection.business_date,
        source.item_id,
        source.lot_id,
        source.handling_unit_id,
        source.location_id,
        source.location_id,
        source.inventory_status_id,
        destination.inventory_status_id,
        inspection.inspected_qty,
        source.uom_id,
        inspection.inspection_id,
        inspection.receipt_inventory_id,
        $8
    FROM decremented_source source
    CROSS JOIN available_balance destination
    CROSS JOIN completed_inspection completed
    JOIN inspection_context inspection ON inspection.inspection_id = completed.inspection_id
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
        $5,
        task_type.task_type_id,
        task_status.task_status_id,
        priority.task_priority_id,
        inspection.receipt_inventory_id,
        destination.balance_id,
        inspection.owner_id,
        inspection.warehouse_id,
        inspection.item_id,
        inspection.lot_id,
        inspection.handling_unit_id,
        destination.location_id,
        target.location_id,
        inspection.inspected_qty,
        0,
        inspection.base_uom_id,
        $8
    FROM inspection_context inspection
    CROSS JOIN available_balance destination
    CROSS JOIN created_movement
    JOIN task_type ON task_type.code = 'PUTAWAY' AND task_type.is_active
    JOIN task_status ON task_status.code = 'OPEN' AND task_status.is_active
    JOIN task_priority priority ON priority.code = $7 AND priority.is_active
    JOIN warehouse_location target
      ON target.location_id = $6
     AND target.warehouse_id = inspection.warehouse_id
     AND target.is_active
    JOIN location_type target_type
      ON target_type.location_type_id = target.location_type_id
     AND target_type.allows_storage
    RETURNING *
)
SELECT * FROM created_task;

COMMIT;

-- 4.3 QC FAIL: move all inspected stock to QUARANTINE and open client case.
-- $1 inspection ID, $2 QC-pending source balance ID,
-- $3 proposed quarantine balance ID, $4 movement ID, $5 quarantine-case ID,
-- $6 quarantine location ID, $7 actor ID
BEGIN;

SELECT inspection_id
FROM quality_inspection
WHERE inspection_id = $1
  AND inspection_result_id IS NULL
FOR UPDATE;

WITH inspection_context AS (
    SELECT
        inspection.*,
        batch.receipt_inventory_id,
        batch.lot_id,
        batch.handling_unit_id,
        batch.base_uom_id,
        receipt.owner_id,
        receipt.warehouse_id,
        receipt.business_date,
        batch.item_id
    FROM quality_inspection inspection
    JOIN receipt_inventory batch
      ON batch.receipt_inventory_id = inspection.receipt_inventory_id
    JOIN receipt_line line ON line.receipt_line_id = batch.receipt_line_id
    JOIN receipt ON receipt.receipt_id = line.receipt_id
    WHERE inspection.inspection_id = $1
      AND inspection.inspection_result_id IS NULL
),
decremented_source AS (
    UPDATE inventory_balance source
    SET on_hand_qty = source.on_hand_qty - inspection.inspected_qty,
        version_no = source.version_no + 1,
        updated_at = clock_timestamp()
    FROM inspection_context inspection
    JOIN inventory_status pending_status ON pending_status.code = 'QC_PENDING'
    WHERE source.balance_id = $2
      AND source.inventory_status_id = pending_status.inventory_status_id
      AND source.item_id = inspection.item_id
      AND source.on_hand_qty - source.reserved_qty >= inspection.inspected_qty
    RETURNING source.*
),
quarantine_balance AS (
    INSERT INTO inventory_balance (
        balance_id, owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id,
        on_hand_qty, reserved_qty, uom_id
    )
    SELECT
        $3,
        source.owner_id,
        source.warehouse_id,
        quarantine_location.location_id,
        source.item_id,
        source.lot_id,
        source.handling_unit_id,
        quarantine_status.inventory_status_id,
        inspection.inspected_qty,
        0,
        source.uom_id
    FROM decremented_source source
    CROSS JOIN inspection_context inspection
    JOIN warehouse_location quarantine_location
      ON quarantine_location.location_id = $6
     AND quarantine_location.warehouse_id = source.warehouse_id
     AND quarantine_location.is_active
    JOIN location_type location_type
      ON location_type.location_type_id = quarantine_location.location_type_id
     AND location_type.code = 'QUARANTINE'
    JOIN inventory_status quarantine_status
      ON quarantine_status.code = 'QUARANTINE'
     AND quarantine_status.is_active
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
        passed_qty = 0,
        failed_qty = inspection.inspected_qty,
        inspected_at = clock_timestamp(),
        inspected_by = $7
    FROM quality_status
    CROSS JOIN inspection_result result
    WHERE inspection.inspection_id = $1
      AND quality_status.code = 'FAILED'
      AND result.code = 'REJECTED'
      AND EXISTS (SELECT 1 FROM quarantine_balance)
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
        $4,
        movement_type.movement_type_id,
        source.owner_id,
        source.warehouse_id,
        inspection.business_date,
        source.item_id,
        source.lot_id,
        source.handling_unit_id,
        source.location_id,
        destination.location_id,
        source.inventory_status_id,
        destination.inventory_status_id,
        inspection.inspected_qty,
        source.uom_id,
        inspection.inspection_id,
        inspection.receipt_inventory_id,
        $7
    FROM decremented_source source
    CROSS JOIN quarantine_balance destination
    CROSS JOIN completed_inspection completed
    JOIN inspection_context inspection ON inspection.inspection_id = completed.inspection_id
    JOIN movement_type ON movement_type.code = 'STATUS_CHANGE' AND movement_type.is_active
    RETURNING movement_id
),
created_case AS (
    INSERT INTO quarantine_case (
        quarantine_case_id,
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
        $5,
        dt.document_type_id,
        initial_status.status_id,
        inspection.receipt_inventory_id,
        inspection.inspection_id,
        destination.balance_id,
        inspection.owner_id,
        inspection.warehouse_id,
        inspection.inspected_qty,
        inspection.base_uom_id,
        $7,
        $7
    FROM inspection_context inspection
    CROSS JOIN quarantine_balance destination
    CROSS JOIN created_movement
    JOIN document_type dt ON dt.code = 'QUARANTINE_CASE' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial
     AND initial_status.is_active
    RETURNING *
)
SELECT * FROM created_case;

COMMIT;
