-- Business partner, UOM, category, item, conversion, and barcode CRUD catalog.
-- Execute one numbered query at a time using prepared parameters.

SET search_path TO wms, public;

-- =============================================================================
-- 1. PARTNER TYPE AND BUSINESS PARTNER
-- =============================================================================

-- 1.1 Create partner type. $1 code, $2 name, $3 description
INSERT INTO partner_type (code, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- 1.2 List partner types. $1 active or null
SELECT *
FROM partner_type
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY name;

-- 1.3 Update partner type. $1 ID, $2 name, $3 description, $4 active
UPDATE partner_type
SET name        = $2,
    description = $3,
    is_active   = $4
WHERE partner_type_id = $1
RETURNING *;

-- 1.4 Create business partner.
-- $1 owner ID, $2 code, $3 name, $4 legal name, $5 tax number,
-- $6 email, $7 phone, $8 address 1, $9 address 2, $10 city,
-- $11 province, $12 postal, $13 country, $14 actor ID
INSERT INTO business_partner (
    owner_id, code, name, legal_name, tax_number, email, phone,
    address_line_1, address_line_2, city, province, postal_code, country_code,
    created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13,
    $14, $14
)
RETURNING *;

-- 1.5 Get partner with all assigned types. $1 partner ID
SELECT
    bp.*,
    owner.code AS owner_code,
    owner.name AS owner_name,
    COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object('id', pt.partner_type_id, 'code', pt.code, 'name', pt.name)
            ORDER BY pt.name
        )
        FROM business_partner_type bpt
        JOIN partner_type pt ON pt.partner_type_id = bpt.partner_type_id
        WHERE bpt.partner_id = bp.partner_id
    ), '[]'::jsonb) AS partner_types
FROM business_partner bp
JOIN organization owner ON owner.organization_id = bp.owner_id
WHERE bp.partner_id = $1;

-- 1.6 Search partners.
-- $1 owner ID, $2 partner type code or null, $3 search or null,
-- $4 active or null, $5 limit, $6 offset
SELECT
    bp.partner_id,
    bp.code,
    bp.name,
    bp.email,
    bp.phone,
    bp.city,
    bp.is_active,
    bp.updated_at,
    count(*) OVER () AS total_rows
FROM business_partner bp
WHERE bp.owner_id = $1
  AND (
      $2 IS NULL OR EXISTS (
          SELECT 1
          FROM business_partner_type bpt
          JOIN partner_type pt ON pt.partner_type_id = bpt.partner_type_id
          WHERE bpt.partner_id = bp.partner_id
            AND pt.code = $2
            AND pt.is_active
      )
  )
  AND ($3 IS NULL OR bp.code ILIKE '%' || $3 || '%' OR bp.name ILIKE '%' || $3 || '%')
  AND ($4 IS NULL OR bp.is_active = $4)
ORDER BY bp.name, bp.partner_id
LIMIT $5 OFFSET $6;

-- 1.7 Update partner. Owner and code remain stable after operational use.
-- $1 ID, $2 name, $3 legal name, $4 tax number, $5 email, $6 phone,
-- $7 address 1, $8 address 2, $9 city, $10 province, $11 postal,
-- $12 country, $13 active, $14 actor ID, $15 expected updated_at
UPDATE business_partner
SET name           = $2,
    legal_name     = $3,
    tax_number     = $4,
    email          = $5,
    phone          = $6,
    address_line_1 = $7,
    address_line_2 = $8,
    city           = $9,
    province       = $10,
    postal_code    = $11,
    country_code   = $12,
    is_active      = $13,
    updated_at     = clock_timestamp(),
    updated_by     = $14
WHERE partner_id = $1
  AND updated_at = $15
RETURNING *;

-- 1.8 Deactivate partner. $1 ID, $2 actor ID, $3 expected updated_at
UPDATE business_partner
SET is_active  = false,
    updated_at = clock_timestamp(),
    updated_by = $2
WHERE partner_id = $1
  AND updated_at = $3
RETURNING partner_id, code, is_active, updated_at;

-- 1.9 Assign type to partner. $1 partner ID, $2 partner type code
INSERT INTO business_partner_type (partner_id, partner_type_id)
SELECT $1, partner_type_id
FROM partner_type
WHERE code = $2
  AND is_active
ON CONFLICT (partner_id, partner_type_id) DO NOTHING
RETURNING *;

-- 1.10 Remove type from partner. $1 partner ID, $2 partner type code
DELETE FROM business_partner_type bpt
USING partner_type pt
WHERE bpt.partner_id = $1
  AND bpt.partner_type_id = pt.partner_type_id
  AND pt.code = $2
RETURNING bpt.*;

-- =============================================================================
-- 2. UNIT OF MEASURE
-- =============================================================================

-- 2.1 Create UOM. $1 code, $2 name, $3 decimal scale
INSERT INTO uom (code, name, decimal_scale)
VALUES ($1, $2, $3)
RETURNING *;

-- 2.2 List/search UOM. $1 search or null, $2 active or null
SELECT *
FROM uom
WHERE ($1 IS NULL OR code ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%')
  AND ($2 IS NULL OR is_active = $2)
ORDER BY name;

-- 2.3 Update UOM. Code remains stable after item usage.
-- $1 ID, $2 name, $3 decimal scale, $4 active
UPDATE uom
SET name          = $2,
    decimal_scale = $3,
    is_active     = $4
WHERE uom_id = $1
RETURNING *;

-- =============================================================================
-- 3. ITEM CATEGORY
-- =============================================================================

-- 3.1 Create category. Parent must belong to the same owner.
-- $1 owner ID, $2 parent category ID or null, $3 code, $4 name
INSERT INTO item_category (owner_id, parent_category_id, code, name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- 3.2 List category hierarchy for owner. $1 owner ID, $2 active or null
WITH RECURSIVE category_tree AS (
    SELECT
        c.category_id,
        c.parent_category_id,
        c.code,
        c.name,
        c.is_active,
        0 AS depth,
        ARRAY[c.name]::varchar[] AS sort_path
    FROM item_category c
    WHERE c.owner_id = $1
      AND c.parent_category_id IS NULL
      AND ($2 IS NULL OR c.is_active = $2)

    UNION ALL

    SELECT
        child.category_id,
        child.parent_category_id,
        child.code,
        child.name,
        child.is_active,
        parent.depth + 1,
        parent.sort_path || child.name
    FROM item_category child
    JOIN category_tree parent ON parent.category_id = child.parent_category_id
    WHERE child.owner_id = $1
      AND ($2 IS NULL OR child.is_active = $2)
)
SELECT *
FROM category_tree
ORDER BY sort_path;

-- 3.3 Update category. $1 ID, $2 parent ID or null, $3 name, $4 active
UPDATE item_category
SET parent_category_id = $2,
    name               = $3,
    is_active          = $4
WHERE category_id = $1
RETURNING *;

-- 3.4 Deactivate category. $1 category ID
UPDATE item_category
SET is_active = false
WHERE category_id = $1
RETURNING category_id, code, is_active;

-- =============================================================================
-- 4. ITEM
-- =============================================================================

-- 4.1 Create item and its required base-UOM conversion atomically.
-- $1 owner ID, $2 category ID or null, $3 code, $4 name, $5 description,
-- $6 base UOM ID, $7 weight, $8 volume, $9 lot controlled,
-- $10 serial controlled, $11 shelf-life days, $12 minimum receive days,
-- $13 actor account ID
WITH created_item AS (
    INSERT INTO item (
        owner_id,
        category_id,
        code,
        name,
        description,
        base_uom_id,
        weight,
        volume,
        lot_controlled,
        serial_controlled,
        shelf_life_days,
        minimum_receive_days,
        created_by,
        updated_by
    ) VALUES (
        $1, $2, $3, $4, $5, $6,
        $7, $8, $9, $10, $11, $12,
        $13, $13
    )
    RETURNING *
),
created_base_uom AS (
    INSERT INTO item_uom (
        item_id, uom_id, conversion_to_base,
        is_receiving_uom, is_picking_uom
    )
    SELECT item_id, base_uom_id, 1, true, true
    FROM created_item
    RETURNING item_id
)
SELECT created_item.*
FROM created_item
JOIN created_base_uom USING (item_id);

-- 4.2 Get item with UOMs and barcodes. $1 item ID
SELECT
    i.*,
    owner.code AS owner_code,
    owner.name AS owner_name,
    category.code AS category_code,
    category.name AS category_name,
    base_uom.code AS base_uom_code,
    COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object(
                'itemUomId', iu.item_uom_id,
                'uomId', unit.uom_id,
                'code', unit.code,
                'conversionToBase', iu.conversion_to_base,
                'receiving', iu.is_receiving_uom,
                'picking', iu.is_picking_uom,
                'active', iu.is_active
            ) ORDER BY iu.conversion_to_base
        )
        FROM item_uom iu
        JOIN uom unit ON unit.uom_id = iu.uom_id
        WHERE iu.item_id = i.item_id
    ), '[]'::jsonb) AS uoms,
    COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object(
                'itemBarcodeId', ib.item_barcode_id,
                'barcode', ib.barcode,
                'uomId', ib.uom_id,
                'primary', ib.is_primary,
                'active', ib.is_active
            ) ORDER BY ib.is_primary DESC, ib.barcode
        )
        FROM item_barcode ib
        WHERE ib.item_id = i.item_id
    ), '[]'::jsonb) AS barcodes
FROM item i
JOIN organization owner ON owner.organization_id = i.owner_id
LEFT JOIN item_category category ON category.category_id = i.category_id
JOIN uom base_uom ON base_uom.uom_id = i.base_uom_id
WHERE i.item_id = $1;

-- 4.3 Search items.
-- $1 owner ID, $2 category ID or null, $3 search or null,
-- $4 active or null, $5 limit, $6 offset
SELECT
    i.item_id,
    i.code,
    i.name,
    category.code AS category_code,
    base_uom.code AS base_uom_code,
    i.lot_controlled,
    i.serial_controlled,
    i.is_active,
    i.updated_at,
    count(*) OVER () AS total_rows
FROM item i
LEFT JOIN item_category category ON category.category_id = i.category_id
JOIN uom base_uom ON base_uom.uom_id = i.base_uom_id
WHERE i.owner_id = $1
  AND ($2 IS NULL OR i.category_id = $2)
  AND ($3 IS NULL OR i.code ILIKE '%' || $3 || '%' OR i.name ILIKE '%' || $3 || '%')
  AND ($4 IS NULL OR i.is_active = $4)
ORDER BY i.name, i.item_id
LIMIT $5 OFFSET $6;

-- 4.4 Update item. Owner, code, and base UOM remain stable after stock exists.
-- $1 ID, $2 category ID or null, $3 name, $4 description,
-- $5 weight, $6 volume, $7 lot controlled, $8 serial controlled,
-- $9 shelf-life days, $10 minimum receive days, $11 active,
-- $12 actor ID, $13 expected updated_at
UPDATE item
SET category_id          = $2,
    name                 = $3,
    description          = $4,
    weight               = $5,
    volume               = $6,
    lot_controlled       = $7,
    serial_controlled    = $8,
    shelf_life_days      = $9,
    minimum_receive_days = $10,
    is_active            = $11,
    updated_at           = clock_timestamp(),
    updated_by           = $12
WHERE item_id = $1
  AND updated_at = $13
RETURNING *;

-- 4.5 Deactivate item. $1 ID, $2 actor ID, $3 expected updated_at
UPDATE item
SET is_active  = false,
    updated_at = clock_timestamp(),
    updated_by = $2
WHERE item_id = $1
  AND updated_at = $3
RETURNING item_id, code, is_active, updated_at;

-- =============================================================================
-- 5. ITEM UOM CONVERSION
-- =============================================================================

-- 5.1 Add item UOM.
-- $1 item ID, $2 UOM ID, $3 conversion to base,
-- $4 length, $5 width, $6 height, $7 weight,
-- $8 receiving UOM, $9 picking UOM
INSERT INTO item_uom (
    item_id, uom_id, conversion_to_base,
    length, width, height, weight,
    is_receiving_uom, is_picking_uom
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- 5.2 Update item UOM.
-- $1 item UOM ID, $2 conversion, $3 length, $4 width, $5 height,
-- $6 weight, $7 receiving, $8 picking, $9 active
UPDATE item_uom
SET conversion_to_base = $2,
    length             = $3,
    width              = $4,
    height             = $5,
    weight             = $6,
    is_receiving_uom   = $7,
    is_picking_uom     = $8,
    is_active          = $9
WHERE item_uom_id = $1
RETURNING *;

-- 5.3 Deactivate item UOM. $1 item UOM ID
UPDATE item_uom
SET is_active = false
WHERE item_uom_id = $1
RETURNING item_uom_id, item_id, uom_id, is_active;

-- =============================================================================
-- 6. ITEM BARCODE
-- =============================================================================

-- 6.1 Create barcode. $1 item ID, $2 UOM ID or null, $3 barcode, $4 primary
INSERT INTO item_barcode (item_id, uom_id, barcode, is_primary)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- 6.2 Make one barcode primary atomically.
-- $1 item ID, $2 barcode row ID
BEGIN;

UPDATE item_barcode
SET is_primary = false
WHERE item_id = $1
  AND is_primary;

UPDATE item_barcode
SET is_primary = true,
    is_active  = true
WHERE item_id = $1
  AND item_barcode_id = $2
RETURNING *;

COMMIT;

-- 6.3 Update barcode. $1 barcode row ID, $2 UOM ID or null, $3 barcode, $4 active
UPDATE item_barcode
SET uom_id    = $2,
    barcode   = $3,
    is_active = $4
WHERE item_barcode_id = $1
RETURNING *;

-- 6.4 Deactivate barcode. $1 barcode row ID
UPDATE item_barcode
SET is_active  = false,
    is_primary = false
WHERE item_barcode_id = $1
RETURNING *;

