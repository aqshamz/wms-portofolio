-- Master-data reports. Execute one numbered query at a time.

SET search_path TO wms, public;

-- 1.1 Organization/warehouse coverage.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 include inactive, $5 limit, $6 offset
SELECT owner.code AS owner_code, owner.name AS owner_name,
       warehouse.code AS warehouse_code, warehouse.name AS warehouse_name,
       operator.code AS operator_code, warehouse.timezone_name,
       warehouse.city, warehouse.province, warehouse.country_code,
       warehouse_owner.is_active AS owner_warehouse_active,
       count(DISTINCT zone.zone_id) AS zone_count,
       count(DISTINCT location.location_id) AS location_count,
       count(DISTINCT location.location_id)
           FILTER (WHERE location.is_active AND NOT location.is_locked) AS usable_location_count
FROM warehouse_owner
JOIN organization owner ON owner.organization_id = warehouse_owner.owner_id
JOIN warehouse ON warehouse.warehouse_id = warehouse_owner.warehouse_id
JOIN organization operator ON operator.organization_id = warehouse.operator_id
LEFT JOIN warehouse_zone zone ON zone.warehouse_id = warehouse.warehouse_id
LEFT JOIN warehouse_location location ON location.warehouse_id = warehouse.warehouse_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.MASTER'
)
  AND EXISTS (
    SELECT 1 FROM report_warehouse_scope scope
    WHERE scope.account_id = $1
      AND scope.owner_id = warehouse_owner.owner_id
      AND scope.warehouse_id = warehouse_owner.warehouse_id
  )
  AND ($2::uuid IS NULL OR owner.organization_id = $2)
  AND ($3::uuid IS NULL OR warehouse.warehouse_id = $3)
  AND ($4 OR (owner.is_active AND warehouse.is_active AND warehouse_owner.is_active))
GROUP BY owner.code, owner.name, warehouse.code, warehouse.name,
         operator.code, warehouse.timezone_name, warehouse.city,
         warehouse.province, warehouse.country_code, warehouse_owner.is_active
ORDER BY owner.code, warehouse.code
LIMIT $5 OFFSET $6;

-- 1.2 Warehouse-location utilization and current inventory occupancy.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 zone ID or NULL, $5 include empty locations, $6 limit, $7 offset
WITH stock AS (
    SELECT balance.owner_id, balance.location_id,
           count(*) AS balance_count,
           count(DISTINCT balance.item_id) AS item_count,
           count(DISTINCT balance.handling_unit_id)
             FILTER (WHERE balance.handling_unit_id IS NOT NULL) AS handling_unit_count,
           sum(balance.on_hand_qty) AS on_hand_qty,
           sum(balance.reserved_qty) AS reserved_qty
    FROM inventory_balance balance
    WHERE ($2::uuid IS NULL OR balance.owner_id = $2)
    GROUP BY balance.owner_id, balance.location_id
)
SELECT owner.code AS owner_code, warehouse.code AS warehouse_code,
       zone.code AS zone_code, location.code AS location_code,
       location_type.code AS location_type_code,
       location.is_pick_face, location.is_locked, location.is_active,
       location.max_weight, location.max_volume,
       COALESCE(stock.balance_count, 0) AS balance_count,
       COALESCE(stock.item_count, 0) AS item_count,
       COALESCE(stock.handling_unit_count, 0) AS handling_unit_count,
       COALESCE(stock.on_hand_qty, 0) AS on_hand_qty,
       COALESCE(stock.reserved_qty, 0) AS reserved_qty
FROM report_warehouse_scope scope
JOIN organization owner ON owner.organization_id = scope.owner_id
JOIN warehouse ON warehouse.warehouse_id = scope.warehouse_id
JOIN warehouse_location location ON location.warehouse_id = warehouse.warehouse_id
JOIN warehouse_zone zone ON zone.zone_id = location.zone_id
JOIN location_type ON location_type.location_type_id = location.location_type_id
LEFT JOIN stock ON stock.owner_id = owner.organization_id
               AND stock.location_id = location.location_id
WHERE scope.account_id = $1
  AND EXISTS (
      SELECT 1 FROM report_account_permission permission
      WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.MASTER'
  )
  AND ($2::uuid IS NULL OR scope.owner_id = $2)
  AND ($3::uuid IS NULL OR scope.warehouse_id = $3)
  AND ($4::uuid IS NULL OR zone.zone_id = $4)
  AND ($5 OR stock.location_id IS NOT NULL)
ORDER BY owner.code, warehouse.code, zone.code, location.code
LIMIT $6 OFFSET $7;

-- 1.3 Business-partner directory.
-- $1 account ID, $2 owner ID or NULL, $3 partner-type code or NULL,
-- $4 active or NULL, $5 search text or NULL, $6 limit, $7 offset
SELECT owner.code AS owner_code, partner.partner_id, partner.code,
       partner.name, partner.legal_name, partner.tax_number,
       partner.email, partner.phone, partner.city, partner.province,
       partner.country_code, partner.is_active,
       string_agg(DISTINCT type.code, ', ' ORDER BY type.code) AS partner_types,
       partner.created_at, partner.updated_at
FROM business_partner partner
JOIN organization owner ON owner.organization_id = partner.owner_id
LEFT JOIN business_partner_type assignment ON assignment.partner_id = partner.partner_id
LEFT JOIN partner_type type ON type.partner_type_id = assignment.partner_type_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.MASTER'
)
  AND EXISTS (
    SELECT 1 FROM report_owner_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = partner.owner_id
  )
  AND ($2::uuid IS NULL OR partner.owner_id = $2)
  AND ($3::varchar IS NULL OR EXISTS (
      SELECT 1 FROM business_partner_type requested_assignment
      JOIN partner_type requested_type
        ON requested_type.partner_type_id = requested_assignment.partner_type_id
      WHERE requested_assignment.partner_id = partner.partner_id
        AND requested_type.code = $3
  ))
  AND ($4::boolean IS NULL OR partner.is_active = $4)
  AND ($5::varchar IS NULL OR partner.code ILIKE '%' || $5 || '%'
       OR partner.name ILIKE '%' || $5 || '%'
       OR partner.tax_number ILIKE '%' || $5 || '%')
GROUP BY owner.code, partner.partner_id
ORDER BY owner.code, partner.code
LIMIT $6 OFFSET $7;

-- 1.4 Item catalog with category, base UOM, and control settings.
-- $1 account ID, $2 owner ID or NULL, $3 category ID or NULL,
-- $4 active or NULL, $5 search or NULL, $6 limit, $7 offset
SELECT owner.code AS owner_code, item.item_id, item.code, item.name,
       category.code AS category_code, category.name AS category_name,
       base_uom.code AS base_uom_code, item.weight, item.volume,
       item.lot_controlled, item.serial_controlled,
       item.shelf_life_days, item.minimum_receive_days, item.is_active,
       count(DISTINCT item_uom.item_uom_id) AS configured_uom_count,
       count(DISTINCT barcode.item_barcode_id)
         FILTER (WHERE barcode.is_active) AS active_barcode_count,
       max(barcode.barcode) FILTER (WHERE barcode.is_primary AND barcode.is_active)
         AS primary_barcode
FROM item
JOIN organization owner ON owner.organization_id = item.owner_id
JOIN uom base_uom ON base_uom.uom_id = item.base_uom_id
LEFT JOIN item_category category ON category.category_id = item.category_id
LEFT JOIN item_uom ON item_uom.item_id = item.item_id
LEFT JOIN item_barcode barcode ON barcode.item_id = item.item_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.MASTER'
)
  AND EXISTS (
    SELECT 1 FROM report_owner_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = item.owner_id
  )
  AND ($2::uuid IS NULL OR item.owner_id = $2)
  AND ($3::uuid IS NULL OR item.category_id = $3)
  AND ($4::boolean IS NULL OR item.is_active = $4)
  AND ($5::varchar IS NULL OR item.code ILIKE '%' || $5 || '%'
       OR item.name ILIKE '%' || $5 || '%'
       OR barcode.barcode ILIKE '%' || $5 || '%')
GROUP BY owner.code, item.item_id, category.code, category.name, base_uom.code
ORDER BY owner.code, item.code
LIMIT $6 OFFSET $7;

-- 1.5 Workflow/configuration dictionary.
-- $1 account ID, $2 module code or NULL, $3 active document types only
SELECT module.code AS module_code, document_type.code AS document_type_code,
       document_type.name AS document_type_name, document_type.is_active,
       count(DISTINCT status.status_id) AS status_count,
       count(DISTINCT status.status_id) FILTER (WHERE status.is_initial) AS initial_status_count,
       count(DISTINCT status.status_id) FILTER (WHERE status.is_final) AS final_status_count,
       count(DISTINCT transition.transition_id) AS transition_count,
       count(DISTINCT number_rule.document_number_rule_id)
         FILTER (WHERE number_rule.is_active) AS active_number_rule_count
FROM document_type
JOIN app_module module ON module.code = document_type.module_code
LEFT JOIN document_status status ON status.document_type_id = document_type.document_type_id
LEFT JOIN document_status_transition transition
  ON transition.document_type_id = document_type.document_type_id
LEFT JOIN document_number_rule number_rule
  ON number_rule.document_type_id = document_type.document_type_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.MASTER'
)
  AND ($2::varchar IS NULL OR module.code = $2)
  AND (NOT $3 OR document_type.is_active)
GROUP BY module.code, module.display_order, document_type.document_type_id
ORDER BY module.display_order, document_type.code;
