-- Organization, warehouse, owner assignment, zone, and location CRUD catalog.
-- Execute one numbered query at a time using prepared parameters.

SET search_path TO wms, public;

-- =============================================================================
-- 1. ORGANIZATION / INVENTORY OWNER
-- =============================================================================

-- 1.1 Create organization.
-- $1 code, $2 name, $3 legal name, $4 tax number, $5 timezone,
-- $6 address 1, $7 address 2, $8 city, $9 province, $10 postal code,
-- $11 country code, $12 actor account ID
INSERT INTO organization (
    code, name, legal_name, tax_number, timezone_name,
    address_line_1, address_line_2, city, province, postal_code, country_code,
    created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10, $11,
    $12, $12
)
RETURNING *;

-- 1.2 Get organization. $1 organization ID
SELECT *
FROM organization
WHERE organization_id = $1;

-- 1.3 Search organizations.
-- $1 search or null, $2 active or null, $3 limit, $4 offset
SELECT
    organization_id,
    code,
    name,
    legal_name,
    timezone_name,
    city,
    country_code,
    is_active,
    updated_at,
    count(*) OVER () AS total_rows
FROM organization
WHERE ($1 IS NULL OR code ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%')
  AND ($2 IS NULL OR is_active = $2)
ORDER BY name, organization_id
LIMIT $3 OFFSET $4;

-- 1.4 Update organization with updated_at concurrency protection.
-- $1 ID, $2 name, $3 legal name, $4 tax, $5 timezone,
-- $6 address 1, $7 address 2, $8 city, $9 province, $10 postal,
-- $11 country, $12 active, $13 actor ID, $14 expected updated_at
UPDATE organization
SET name           = $2,
    legal_name     = $3,
    tax_number     = $4,
    timezone_name  = $5,
    address_line_1 = $6,
    address_line_2 = $7,
    city           = $8,
    province       = $9,
    postal_code    = $10,
    country_code   = $11,
    is_active      = $12,
    updated_at     = clock_timestamp(),
    updated_by     = $13
WHERE organization_id = $1
  AND updated_at = $14
RETURNING *;

-- 1.5 Deactivate organization rather than deleting referenced history.
-- $1 organization ID, $2 actor ID, $3 expected updated_at
UPDATE organization
SET is_active  = false,
    updated_at = clock_timestamp(),
    updated_by = $2
WHERE organization_id = $1
  AND updated_at = $3
RETURNING organization_id, code, is_active, updated_at;

-- =============================================================================
-- 2. WAREHOUSE
-- =============================================================================

-- 2.1 Create warehouse.
-- $1 operator organization ID, $2 code, $3 name, $4 timezone,
-- $5 address 1, $6 address 2, $7 city, $8 province, $9 postal,
-- $10 country, $11 actor account ID
INSERT INTO warehouse (
    operator_id, code, name, timezone_name,
    address_line_1, address_line_2, city, province, postal_code, country_code,
    created_by, updated_by
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8, $9, $10,
    $11, $11
)
RETURNING *;

-- 2.2 Get warehouse with operator and owner count. $1 warehouse ID
SELECT
    w.*,
    operator.code AS operator_code,
    operator.name AS operator_name,
    (
        SELECT count(*)
        FROM warehouse_owner wo
        WHERE wo.warehouse_id = w.warehouse_id
          AND wo.is_active
    ) AS active_owner_count
FROM warehouse w
JOIN organization operator ON operator.organization_id = w.operator_id
WHERE w.warehouse_id = $1;

-- 2.3 Search warehouses.
-- $1 operator ID or null, $2 search or null, $3 active or null,
-- $4 limit, $5 offset
SELECT
    w.warehouse_id,
    w.code,
    w.name,
    operator.code AS operator_code,
    operator.name AS operator_name,
    w.timezone_name,
    w.city,
    w.is_active,
    w.updated_at,
    count(*) OVER () AS total_rows
FROM warehouse w
JOIN organization operator ON operator.organization_id = w.operator_id
WHERE ($1 IS NULL OR w.operator_id = $1)
  AND ($2 IS NULL OR w.code ILIKE '%' || $2 || '%' OR w.name ILIKE '%' || $2 || '%')
  AND ($3 IS NULL OR w.is_active = $3)
ORDER BY w.name, w.warehouse_id
LIMIT $4 OFFSET $5;

-- 2.4 Update warehouse.
-- Code and operator remain stable after operational use.
-- $1 ID, $2 name, $3 timezone, $4 address 1, $5 address 2,
-- $6 city, $7 province, $8 postal, $9 country, $10 active,
-- $11 actor ID, $12 expected updated_at
UPDATE warehouse
SET name           = $2,
    timezone_name  = $3,
    address_line_1 = $4,
    address_line_2 = $5,
    city           = $6,
    province       = $7,
    postal_code    = $8,
    country_code   = $9,
    is_active      = $10,
    updated_at     = clock_timestamp(),
    updated_by     = $11
WHERE warehouse_id = $1
  AND updated_at = $12
RETURNING *;

-- 2.5 Deactivate warehouse. $1 ID, $2 actor ID, $3 expected updated_at
UPDATE warehouse
SET is_active  = false,
    updated_at = clock_timestamp(),
    updated_by = $2
WHERE warehouse_id = $1
  AND updated_at = $3
RETURNING warehouse_id, code, is_active, updated_at;

-- =============================================================================
-- 3. WAREHOUSE OWNER ASSIGNMENT
-- =============================================================================

-- 3.1 Allow an owner to use a warehouse.
-- $1 warehouse ID, $2 owner organization ID, $3 actor ID
INSERT INTO warehouse_owner (warehouse_id, owner_id, created_by)
VALUES ($1, $2, $3)
ON CONFLICT (warehouse_id, owner_id)
DO UPDATE SET is_active = true
RETURNING *;

-- 3.2 List owners assigned to warehouse. $1 warehouse ID
SELECT
    wo.owner_id,
    owner.code AS owner_code,
    owner.name AS owner_name,
    wo.is_active,
    wo.created_at
FROM warehouse_owner wo
JOIN organization owner ON owner.organization_id = wo.owner_id
WHERE wo.warehouse_id = $1
ORDER BY owner.name;

-- 3.3 Disable owner assignment rather than deleting it.
-- $1 warehouse ID, $2 owner ID
UPDATE warehouse_owner
SET is_active = false
WHERE warehouse_id = $1
  AND owner_id = $2
RETURNING *;

-- =============================================================================
-- 4. WAREHOUSE ZONE
-- =============================================================================

-- 4.1 Create zone.
-- $1 warehouse ID, $2 code, $3 name, $4 description, $5 actor ID
INSERT INTO warehouse_zone (
    warehouse_id, code, name, description, created_by
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- 4.2 List/search zones.
-- $1 warehouse ID, $2 search or null, $3 active or null
SELECT
    z.zone_id,
    z.warehouse_id,
    z.code,
    z.name,
    z.description,
    z.is_active,
    count(l.location_id) AS location_count
FROM warehouse_zone z
LEFT JOIN warehouse_location l ON l.zone_id = z.zone_id
WHERE z.warehouse_id = $1
  AND ($2 IS NULL OR z.code ILIKE '%' || $2 || '%' OR z.name ILIKE '%' || $2 || '%')
  AND ($3 IS NULL OR z.is_active = $3)
GROUP BY z.zone_id
ORDER BY z.code;

-- 4.3 Update zone. $1 zone ID, $2 name, $3 description, $4 active
UPDATE warehouse_zone
SET name        = $2,
    description = $3,
    is_active   = $4
WHERE zone_id = $1
RETURNING *;

-- 4.4 Deactivate zone. $1 zone ID
UPDATE warehouse_zone
SET is_active = false
WHERE zone_id = $1
RETURNING zone_id, code, is_active;

-- =============================================================================
-- 5. LOCATION TYPE
-- =============================================================================

-- 5.1 Create location type.
-- $1 code, $2 name, $3 description, $4 receiving, $5 storage,
-- $6 picking, $7 shipping
INSERT INTO location_type (
    code, name, description,
    allows_receiving, allows_storage, allows_picking, allows_shipping
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- 5.2 List location types. $1 active or null
SELECT *
FROM location_type
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY name;

-- 5.3 Update location type.
-- $1 ID, $2 name, $3 description, $4 receiving, $5 storage,
-- $6 picking, $7 shipping, $8 active
UPDATE location_type
SET name             = $2,
    description      = $3,
    allows_receiving = $4,
    allows_storage   = $5,
    allows_picking   = $6,
    allows_shipping  = $7,
    is_active        = $8
WHERE location_type_id = $1
RETURNING *;

-- =============================================================================
-- 6. WAREHOUSE LOCATION
-- =============================================================================

-- 6.1 Create location using location type code.
-- $1 warehouse ID, $2 zone ID, $3 location type code,
-- $4 code, $5 barcode, $6 aisle, $7 bay, $8 level, $9 position,
-- $10 max weight, $11 max volume, $12 pick face, $13 pick sequence,
-- $14 actor ID
INSERT INTO warehouse_location (
    warehouse_id,
    zone_id,
    location_type_id,
    code,
    barcode,
    aisle,
    bay,
    level_no,
    position_no,
    max_weight,
    max_volume,
    is_pick_face,
    pick_sequence,
    created_by
)
SELECT
    $1, $2, lt.location_type_id,
    $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14
FROM location_type lt
WHERE lt.code = $3
  AND lt.is_active
RETURNING *;

-- 6.2 Get location. $1 location ID
SELECT
    l.*,
    w.code AS warehouse_code,
    w.name AS warehouse_name,
    z.code AS zone_code,
    z.name AS zone_name,
    lt.code AS location_type_code,
    lt.name AS location_type_name
FROM warehouse_location l
JOIN warehouse w ON w.warehouse_id = l.warehouse_id
JOIN warehouse_zone z ON z.zone_id = l.zone_id
JOIN location_type lt ON lt.location_type_id = l.location_type_id
WHERE l.location_id = $1;

-- 6.3 Search locations.
-- $1 warehouse ID, $2 zone ID or null, $3 location type code or null,
-- $4 search or null, $5 active or null, $6 limit, $7 offset
SELECT
    l.location_id,
    l.code,
    l.barcode,
    z.code AS zone_code,
    lt.code AS location_type_code,
    l.aisle,
    l.bay,
    l.level_no,
    l.position_no,
    l.pick_sequence,
    l.is_pick_face,
    l.is_locked,
    l.is_active,
    count(*) OVER () AS total_rows
FROM warehouse_location l
JOIN warehouse_zone z ON z.zone_id = l.zone_id
JOIN location_type lt ON lt.location_type_id = l.location_type_id
WHERE l.warehouse_id = $1
  AND ($2 IS NULL OR l.zone_id = $2)
  AND ($3 IS NULL OR lt.code = $3)
  AND ($4 IS NULL OR l.code ILIKE '%' || $4 || '%' OR l.barcode ILIKE '%' || $4 || '%')
  AND ($5 IS NULL OR l.is_active = $5)
ORDER BY l.code, l.location_id
LIMIT $6 OFFSET $7;

-- 6.4 Update location. Warehouse and code remain stable after use.
-- $1 location ID, $2 zone ID, $3 location type ID, $4 barcode,
-- $5 aisle, $6 bay, $7 level, $8 position, $9 max weight, $10 max volume,
-- $11 pick face, $12 pick sequence, $13 locked, $14 active
UPDATE warehouse_location
SET zone_id          = $2,
    location_type_id = $3,
    barcode          = $4,
    aisle            = $5,
    bay              = $6,
    level_no         = $7,
    position_no      = $8,
    max_weight       = $9,
    max_volume       = $10,
    is_pick_face     = $11,
    pick_sequence    = $12,
    is_locked        = $13,
    is_active        = $14
WHERE location_id = $1
RETURNING *;

-- 6.5 Deactivate location. $1 location ID
UPDATE warehouse_location
SET is_active = false
WHERE location_id = $1
RETURNING location_id, code, is_active;
