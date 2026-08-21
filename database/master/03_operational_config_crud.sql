-- Operational master/configuration CRUD catalog.
-- Execute one numbered query at a time using prepared parameters.

SET search_path TO wms, public;

-- =============================================================================
-- 0. APPLICATION MODULE MASTER
-- =============================================================================

-- 0.1 Create a module. $1 code, $2 name, $3 display order
INSERT INTO app_module (code, name, display_order)
VALUES (upper($1), $2, $3)
RETURNING *;

-- 0.2 List modules. $1 active or null
SELECT *
FROM app_module
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY display_order, code;

-- 0.3 Update a module. Code remains stable after it is referenced.
-- $1 module ID, $2 name, $3 display order, $4 active
UPDATE app_module
SET name = $2, display_order = $3, is_active = $4
WHERE module_id = $1
RETURNING *;

-- =============================================================================
-- 1. INVENTORY AND QUALITY CONFIGURATION
-- =============================================================================

-- 1.1 Create inventory status.
-- $1 code, $2 name, $3 description, $4 allocatable, $5 pickable
INSERT INTO inventory_status (
    code, name, description, is_allocatable, is_pickable
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- 1.2 List inventory statuses. $1 active or null
SELECT *
FROM inventory_status
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY name;

-- 1.3 Update inventory status.
-- $1 ID, $2 name, $3 description, $4 allocatable, $5 pickable, $6 active
UPDATE inventory_status
SET name           = $2,
    description    = $3,
    is_allocatable = $4,
    is_pickable    = $5,
    is_active      = $6
WHERE inventory_status_id = $1
RETURNING *;

-- 1.4 Create quality status. $1 code, $2 name, $3 description
INSERT INTO quality_status (code, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- 1.5 List quality statuses. $1 active or null
SELECT *
FROM quality_status
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY name;

-- 1.6 Update quality status. $1 ID, $2 name, $3 description, $4 active
UPDATE quality_status
SET name = $2, description = $3, is_active = $4
WHERE quality_status_id = $1
RETURNING *;

-- 1.7 Create inspection result.
-- $1 code, $2 name, $3 description, $4 accepted
INSERT INTO inspection_result (code, name, description, is_accepted)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- 1.8 List inspection results. $1 active or null
SELECT *
FROM inspection_result
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY name;

-- 1.9 Update inspection result.
-- $1 ID, $2 name, $3 description, $4 accepted, $5 active
UPDATE inspection_result
SET name = $2, description = $3, is_accepted = $4, is_active = $5
WHERE inspection_result_id = $1
RETURNING *;

-- 1.10 Create handling-unit type.
-- $1 code, $2 name, $3 maximum weight, $4 maximum volume
INSERT INTO handling_unit_type (code, name, max_weight, max_volume)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- 1.11 List handling-unit types. $1 active or null
SELECT *
FROM handling_unit_type
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY name;

-- 1.12 Update handling-unit type.
-- $1 ID, $2 name, $3 max weight, $4 max volume, $5 active
UPDATE handling_unit_type
SET name = $2, max_weight = $3, max_volume = $4, is_active = $5
WHERE handling_unit_type_id = $1
RETURNING *;

-- =============================================================================
-- 2. REASONS AND MOVEMENT TYPES
-- =============================================================================

-- 2.1 Create reason code.
-- $1 module code, $2 code, $3 name, $4 description, $5 requires note
INSERT INTO reason_code (
    module_code, code, name, description, requires_note
)
SELECT module.code, $2, $3, $4, $5
FROM app_module module
WHERE module.code = $1 AND module.is_active
RETURNING *;

-- 2.2 Search reason codes.
-- $1 module code or null, $2 search or null, $3 active or null
SELECT *
FROM reason_code
WHERE ($1 IS NULL OR module_code = $1)
  AND ($2 IS NULL OR code ILIKE '%' || $2 || '%' OR name ILIKE '%' || $2 || '%')
  AND ($3 IS NULL OR is_active = $3)
ORDER BY module_code, name;

-- 2.3 Update reason code.
-- $1 ID, $2 name, $3 description, $4 requires note, $5 active
UPDATE reason_code
SET name = $2, description = $3, requires_note = $4, is_active = $5
WHERE reason_code_id = $1
RETURNING *;

-- 2.4 Create movement type. $1 code, $2 name, $3 description
INSERT INTO movement_type (code, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- 2.5 List movement types. $1 active or null
SELECT *
FROM movement_type
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY name;

-- 2.6 Update movement type. $1 ID, $2 name, $3 description, $4 active
UPDATE movement_type
SET name = $2, description = $3, is_active = $4
WHERE movement_type_id = $1
RETURNING *;

-- =============================================================================
-- 3. DOCUMENT TYPE, STATUS, AND WORKFLOW
-- =============================================================================

-- 3.1 Create document type.
-- $1 code, $2 name, $3 module code, $4 description
INSERT INTO document_type (code, name, module_code, description)
SELECT $1, $2, module.code, $4
FROM app_module module
WHERE module.code = $3 AND module.is_active
RETURNING *;

-- 3.2 List document types.
-- $1 module or null, $2 active or null
SELECT
    dt.*,
    (
        SELECT count(*)
        FROM document_status ds
        WHERE ds.document_type_id = dt.document_type_id
          AND ds.is_active
    ) AS active_status_count
FROM document_type dt
WHERE ($1 IS NULL OR dt.module_code = $1)
  AND ($2 IS NULL OR dt.is_active = $2)
ORDER BY dt.module_code, dt.name;

-- 3.3 Update document type.
-- Code remains stable after transaction usage.
-- $1 ID, $2 name, $3 module, $4 description, $5 active
UPDATE document_type target
SET name = $2, module_code = module.code,
    description = $4, is_active = $5
FROM app_module module
WHERE target.document_type_id = $1
  AND module.code = $3 AND module.is_active
RETURNING target.*;

-- 3.4 Create document status using document-type code.
-- $1 document type code, $2 status code, $3 name, $4 description,
-- $5 initial, $6 final, $7 cancelled, $8 display order
INSERT INTO document_status (
    document_type_id, code, name, description,
    is_initial, is_final, is_cancelled, display_order
)
SELECT
    document_type_id, $2, $3, $4, $5, $6, $7, $8
FROM document_type
WHERE code = $1
  AND is_active
RETURNING *;

-- 3.5 List statuses for document type. $1 document type code, $2 active or null
SELECT
    ds.status_id,
    ds.code,
    ds.name,
    ds.description,
    ds.is_initial,
    ds.is_final,
    ds.is_cancelled,
    ds.display_order,
    ds.is_active
FROM document_status ds
JOIN document_type dt ON dt.document_type_id = ds.document_type_id
WHERE dt.code = $1
  AND ($2 IS NULL OR ds.is_active = $2)
ORDER BY ds.display_order, ds.name;

-- 3.6 Update document status. If making it initial, first clear the old one.
-- Run in one transaction.
-- $1 status ID, $2 name, $3 description, $4 initial, $5 final,
-- $6 cancelled, $7 display order, $8 active
BEGIN;

UPDATE document_status other_status
SET is_initial = false
WHERE other_status.document_type_id = (
        SELECT selected.document_type_id
        FROM document_status selected
        WHERE selected.status_id = $1
      )
  AND other_status.status_id <> $1
  AND $4;

UPDATE document_status
SET name          = $2,
    description   = $3,
    is_initial    = $4,
    is_final      = $5,
    is_cancelled  = $6,
    display_order = $7,
    is_active     = $8
WHERE status_id = $1
RETURNING *;

COMMIT;

-- 3.7 Create allowed document transition.
-- $1 document type code, $2 from-status code, $3 to-status code,
-- $4 required permission code or null
INSERT INTO document_status_transition (
    document_type_id,
    from_status_id,
    to_status_id,
    required_permission_id
)
SELECT
    dt.document_type_id,
    from_status.status_id,
    to_status.status_id,
    permission.permission_id
FROM document_type dt
JOIN document_status from_status
  ON from_status.document_type_id = dt.document_type_id
 AND from_status.code = $2
JOIN document_status to_status
  ON to_status.document_type_id = dt.document_type_id
 AND to_status.code = $3
LEFT JOIN app_permission permission
  ON permission.code = $4
 AND permission.is_active
WHERE dt.code = $1
  AND ($4 IS NULL OR permission.permission_id IS NOT NULL)
ON CONFLICT (document_type_id, from_status_id, to_status_id)
DO UPDATE SET
    required_permission_id = EXCLUDED.required_permission_id,
    is_active = true
RETURNING *;

-- 3.8 List workflow transitions. $1 document type code
SELECT
    transition.transition_id,
    from_status.code AS from_status_code,
    from_status.name AS from_status_name,
    to_status.code AS to_status_code,
    to_status.name AS to_status_name,
    permission.code AS required_permission_code,
    transition.is_active
FROM document_status_transition transition
JOIN document_type dt ON dt.document_type_id = transition.document_type_id
JOIN document_status from_status ON from_status.status_id = transition.from_status_id
JOIN document_status to_status ON to_status.status_id = transition.to_status_id
LEFT JOIN app_permission permission
       ON permission.permission_id = transition.required_permission_id
WHERE dt.code = $1
ORDER BY from_status.display_order, to_status.display_order;

-- 3.9 Deactivate workflow transition. $1 transition ID
UPDATE document_status_transition
SET is_active = false
WHERE transition_id = $1
RETURNING *;

-- =============================================================================
-- 4. DOCUMENT NUMBERING
-- =============================================================================

-- 4.1 Get current number rule. $1 document type code
SELECT
    nr.*,
    dt.code AS document_type_code,
    dt.name AS document_type_name
FROM document_number_rule nr
JOIN document_type dt ON dt.document_type_id = nr.document_type_id
WHERE dt.code = $1
ORDER BY nr.effective_from DESC;

-- 4.2 Replace the active numbering rule atomically.
-- Existing transaction IDs never change.
-- $1 document type code, $2 prefix, $3 separator, $4 sequence length,
-- $5 include partner, $6 include warehouse, $7 effective from, $8 actor ID
BEGIN;

UPDATE document_number_rule current_rule
SET is_active       = false,
    effective_until = $7 - 1
FROM document_type dt
WHERE current_rule.document_type_id = dt.document_type_id
  AND dt.code = $1
  AND current_rule.is_active;

INSERT INTO document_number_rule (
    document_type_id,
    prefix,
    separator,
    sequence_length,
    include_partner_code,
    include_warehouse_code,
    effective_from,
    created_by
)
SELECT
    document_type_id,
    $2, $3, $4, $5, $6, $7, $8
FROM document_type
WHERE code = $1
  AND is_active
RETURNING *;

COMMIT;

-- 4.3 View daily counter usage. $1 document type code, $2 date from, $3 date to
SELECT
    dt.code AS document_type_code,
    counter.business_date,
    counter.last_number,
    counter.updated_at
FROM document_daily_counter counter
JOIN document_type dt ON dt.document_type_id = counter.document_type_id
WHERE dt.code = $1
  AND counter.business_date BETWEEN $2 AND $3
ORDER BY counter.business_date DESC;

-- =============================================================================
-- 5. TASK CONFIGURATION
-- =============================================================================

-- 5.1 Create task type. $1 code, $2 name, $3 description
INSERT INTO task_type (code, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- 5.2 List task types. $1 active or null
SELECT * FROM task_type
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY name;

-- 5.3 Update task type. $1 ID, $2 name, $3 description, $4 active
UPDATE task_type
SET name = $2, description = $3, is_active = $4
WHERE task_type_id = $1
RETURNING *;

-- 5.4 Create task status.
-- $1 code, $2 name, $3 initial, $4 final, $5 cancelled
INSERT INTO task_status (code, name, is_initial, is_final, is_cancelled)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- 5.5 List task statuses. $1 active or null
SELECT * FROM task_status
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY is_initial DESC, is_final, name;

-- 5.6 Update task status.
-- $1 ID, $2 name, $3 initial, $4 final, $5 cancelled, $6 active
UPDATE task_status
SET name = $2,
    is_initial = $3,
    is_final = $4,
    is_cancelled = $5,
    is_active = $6
WHERE task_status_id = $1
RETURNING *;

-- 5.7 Create/update task transition.
-- $1 from status code, $2 to status code, $3 permission code or null
INSERT INTO task_status_transition (
    from_status_id, to_status_id, required_permission_id
)
SELECT
    from_status.task_status_id,
    to_status.task_status_id,
    permission.permission_id
FROM task_status from_status
CROSS JOIN task_status to_status
LEFT JOIN app_permission permission
  ON permission.code = $3
 AND permission.is_active
WHERE from_status.code = $1
  AND to_status.code = $2
  AND ($3 IS NULL OR permission.permission_id IS NOT NULL)
ON CONFLICT (from_status_id, to_status_id)
DO UPDATE SET
    required_permission_id = EXCLUDED.required_permission_id,
    is_active = true
RETURNING *;

-- 5.8 List task transitions.
SELECT
    transition.task_status_transition_id,
    from_status.code AS from_status_code,
    to_status.code AS to_status_code,
    permission.code AS required_permission_code,
    transition.is_active
FROM task_status_transition transition
JOIN task_status from_status ON from_status.task_status_id = transition.from_status_id
JOIN task_status to_status ON to_status.task_status_id = transition.to_status_id
LEFT JOIN app_permission permission
       ON permission.permission_id = transition.required_permission_id
ORDER BY from_status.name, to_status.name;

-- 5.9 Create task priority. $1 code, $2 name, $3 priority value
INSERT INTO task_priority (code, name, priority_value)
VALUES ($1, $2, $3)
RETURNING *;

-- 5.10 List priorities. $1 active or null
SELECT * FROM task_priority
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY priority_value DESC;

-- 5.11 Update priority. $1 ID, $2 name, $3 value, $4 active
UPDATE task_priority
SET name = $2, priority_value = $3, is_active = $4
WHERE task_priority_id = $1
RETURNING *;

-- =============================================================================
-- 6. PUTAWAY AND PICKING STRATEGIES
-- =============================================================================

-- A null owner means a warehouse/global default; a null warehouse means the
-- strategy can apply across warehouses. More-specific matching belongs in the
-- backend service that selects a strategy.

-- 6.1 Create putaway strategy.
-- $1 owner ID or null, $2 warehouse ID or null, $3 code, $4 name, $5 description
INSERT INTO putaway_strategy (owner_id, warehouse_id, code, name, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- 6.2 List putaway strategies.
-- $1 owner ID or null, $2 warehouse ID or null, $3 active or null
SELECT
    strategy.*,
    owner.code AS owner_code,
    warehouse.code AS warehouse_code,
    (
        SELECT count(*)
        FROM putaway_strategy_rule rule
        WHERE rule.putaway_strategy_id = strategy.putaway_strategy_id
          AND rule.is_active
    ) AS active_rule_count
FROM putaway_strategy strategy
LEFT JOIN organization owner ON owner.organization_id = strategy.owner_id
LEFT JOIN warehouse ON warehouse.warehouse_id = strategy.warehouse_id
WHERE ($1 IS NULL OR strategy.owner_id = $1)
  AND ($2 IS NULL OR strategy.warehouse_id = $2)
  AND ($3 IS NULL OR strategy.is_active = $3)
ORDER BY strategy.name;

-- 6.3 Add putaway rule.
-- $1 strategy ID, $2 sequence, $3 category ID or null,
-- $4 location type ID or null, $5 zone ID or null, $6 minimum empty percent
INSERT INTO putaway_strategy_rule (
    putaway_strategy_id, sequence_no, category_id,
    location_type_id, zone_id, minimum_empty_percent
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- 6.4 Update putaway rule.
-- $1 rule ID, $2 sequence, $3 category, $4 location type,
-- $5 zone, $6 minimum empty percent, $7 active
UPDATE putaway_strategy_rule
SET sequence_no = $2,
    category_id = $3,
    location_type_id = $4,
    zone_id = $5,
    minimum_empty_percent = $6,
    is_active = $7
WHERE rule_id = $1
RETURNING *;

-- 6.5 Create picking sort method. $1 code, $2 name, $3 description
INSERT INTO picking_sort_method (code, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- 6.6 List picking sort methods. $1 active or null
SELECT * FROM picking_sort_method
WHERE ($1 IS NULL OR is_active = $1)
ORDER BY name;

-- 6.7 Create picking strategy.
-- $1 owner ID or null, $2 warehouse ID or null, $3 code, $4 name, $5 description
INSERT INTO picking_strategy (owner_id, warehouse_id, code, name, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- 6.8 List picking strategies.
-- $1 owner ID or null, $2 warehouse ID or null, $3 active or null
SELECT
    strategy.*,
    owner.code AS owner_code,
    warehouse.code AS warehouse_code,
    (
        SELECT count(*)
        FROM picking_strategy_rule rule
        WHERE rule.picking_strategy_id = strategy.picking_strategy_id
          AND rule.is_active
    ) AS active_rule_count
FROM picking_strategy strategy
LEFT JOIN organization owner ON owner.organization_id = strategy.owner_id
LEFT JOIN warehouse ON warehouse.warehouse_id = strategy.warehouse_id
WHERE ($1 IS NULL OR strategy.owner_id = $1)
  AND ($2 IS NULL OR strategy.warehouse_id = $2)
  AND ($3 IS NULL OR strategy.is_active = $3)
ORDER BY strategy.name;

-- 6.9 Add picking rule.
-- $1 strategy ID, $2 sequence, $3 inventory status ID or null,
-- $4 zone ID or null, $5 picking sort method code
INSERT INTO picking_strategy_rule (
    picking_strategy_id,
    sequence_no,
    inventory_status_id,
    zone_id,
    picking_sort_method_id
)
SELECT $1, $2, $3, $4, method.picking_sort_method_id
FROM picking_sort_method method
WHERE method.code = $5
  AND method.is_active
RETURNING *;

-- 6.10 Update picking rule.
-- $1 rule ID, $2 sequence, $3 inventory status, $4 zone,
-- $5 picking method ID, $6 active
UPDATE picking_strategy_rule
SET sequence_no = $2,
    inventory_status_id = $3,
    zone_id = $4,
    picking_sort_method_id = $5,
    is_active = $6
WHERE rule_id = $1
RETURNING *;

-- =============================================================================
-- 7. CARRIER AND SERVICE
-- =============================================================================

-- 7.1 Create carrier. $1 code, $2 name, $3 carrier business-partner ID or null
INSERT INTO carrier (code, name, business_partner_id)
SELECT $1, $2, partner.partner_id
FROM (SELECT 1) seed
LEFT JOIN business_partner partner
  ON partner.partner_id = $3 AND partner.is_active
WHERE $3::uuid IS NULL OR partner.partner_id IS NOT NULL
RETURNING *;

-- 7.2 List/search carriers. $1 search or null, $2 active or null
SELECT
    carrier.*,
    (
        SELECT count(*)
        FROM carrier_service service
        WHERE service.carrier_id = carrier.carrier_id
          AND service.is_active
    ) AS active_service_count
FROM carrier
WHERE ($1 IS NULL OR code ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%')
  AND ($2 IS NULL OR is_active = $2)
ORDER BY name;

-- 7.3 Update carrier. $1 ID, $2 name, $3 business-partner ID or null, $4 active
UPDATE carrier
SET name = $2, business_partner_id = $3, is_active = $4
WHERE carrier_id = $1
RETURNING *;

-- 7.4 Create carrier service. $1 carrier ID, $2 code, $3 name
INSERT INTO carrier_service (carrier_id, code, name)
VALUES ($1, $2, $3)
RETURNING *;

-- 7.5 List carrier services. $1 carrier ID, $2 active or null
SELECT *
FROM carrier_service
WHERE carrier_id = $1
  AND ($2 IS NULL OR is_active = $2)
ORDER BY name;

-- 7.6 Update carrier service. $1 ID, $2 name, $3 active
UPDATE carrier_service
SET name = $2, is_active = $3
WHERE carrier_service_id = $1
RETURNING *;

-- 7.7 Create driver. $1 carrier ID, $2 code, $3 name, $4 phone or null,
-- $5 license number or null, $6 linked account ID or null, $7 actor ID
INSERT INTO carrier_driver (
    carrier_id, code, name, phone_number, license_number, account_id, created_by
)
SELECT carrier.carrier_id, $2, $3, $4, $5, account.account_id, $7
FROM carrier
LEFT JOIN app_account account ON account.account_id = $6
WHERE carrier.carrier_id = $1 AND carrier.is_active
  AND ($6::uuid IS NULL OR account.account_id IS NOT NULL)
RETURNING *;

-- 7.8 List drivers. $1 carrier ID or null, $2 active or null
SELECT driver.*, carrier.code AS carrier_code, account.username
FROM carrier_driver driver
JOIN carrier ON carrier.carrier_id = driver.carrier_id
LEFT JOIN app_account account ON account.account_id = driver.account_id
WHERE ($1::uuid IS NULL OR driver.carrier_id = $1)
  AND ($2::boolean IS NULL OR driver.is_active = $2)
ORDER BY carrier.code, driver.name;

-- 7.9 Update driver. $1 driver ID, $2 name, $3 phone, $4 license,
-- $5 account ID or null, $6 active
UPDATE carrier_driver
SET name = $2, phone_number = $3, license_number = $4,
    account_id = $5, is_active = $6
WHERE driver_id = $1
RETURNING *;

-- 7.10 Configure the safe delivery-return destination for one owner/warehouse.
-- The selected status must be active and non-allocatable.
-- $1 owner ID, $2 warehouse ID, $3 return location ID,
-- $4 inventory-status code, $5 actor ID
INSERT INTO outbound_return_policy (
    owner_id, warehouse_id, return_location_id,
    return_inventory_status_id, created_by, updated_by
)
SELECT scope.owner_id, scope.warehouse_id, location.location_id,
       status.inventory_status_id, $5, $5
FROM warehouse_owner scope
JOIN warehouse_location location
  ON location.location_id = $3
 AND location.warehouse_id = scope.warehouse_id
 AND location.is_active
JOIN inventory_status status
  ON status.code = $4 AND status.is_active AND NOT status.is_allocatable
WHERE scope.owner_id = $1 AND scope.warehouse_id = $2 AND scope.is_active
ON CONFLICT (owner_id, warehouse_id) DO UPDATE SET
    return_location_id = EXCLUDED.return_location_id,
    return_inventory_status_id = EXCLUDED.return_inventory_status_id,
    is_active = true,
    updated_at = clock_timestamp(),
    updated_by = EXCLUDED.updated_by
RETURNING *;

-- 7.11 Read return policy. $1 owner ID, $2 warehouse ID
SELECT policy.*, location.code AS return_location_code,
       status.code AS return_inventory_status_code
FROM outbound_return_policy policy
JOIN warehouse_location location ON location.location_id = policy.return_location_id
JOIN inventory_status status
  ON status.inventory_status_id = policy.return_inventory_status_id
WHERE policy.owner_id = $1 AND policy.warehouse_id = $2;
