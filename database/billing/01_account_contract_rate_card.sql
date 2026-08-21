-- Billing-account, contract, tax, rate-card, and tier maintenance.
-- Execute one numbered operation at a time using prepared parameters.

SET search_path TO wms, public;

-- =============================================================================
-- 0. BILLING REFERENCE MASTERS
-- =============================================================================

-- 0.1 Upsert currency. $1 code, $2 name, $3 decimal scale, $4 active
INSERT INTO currency (code, name, decimal_scale, is_active)
VALUES (upper($1), $2, $3, $4)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name, decimal_scale = EXCLUDED.decimal_scale,
    is_active = EXCLUDED.is_active
RETURNING *;

-- 0.2 Upsert billing cycle. Exactly one period unit is used unless on-demand.
-- $1 code, $2 name, $3 period days or NULL, $4 period months or NULL,
-- $5 on demand, $6 active
INSERT INTO billing_cycle (
    code, name, period_days, period_months, is_on_demand, is_active
)
VALUES (upper($1), $2, $3, $4, $5, $6)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name, period_days = EXCLUDED.period_days,
    period_months = EXCLUDED.period_months,
    is_on_demand = EXCLUDED.is_on_demand, is_active = EXCLUDED.is_active
RETURNING *;

-- 0.3 Upsert payment term. $1 code, $2 name, $3 due days, $4 active
INSERT INTO payment_term (code, name, due_days, is_active)
VALUES (upper($1), $2, $3, $4)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name, due_days = EXCLUDED.due_days,
    is_active = EXCLUDED.is_active
RETURNING *;

-- 0.4 Upsert charge basis.
-- $1 code, $2 name, $3 description, $4 requires UOM,
-- $5 use event quantity (false means one charge unit per event), $6 active
INSERT INTO charge_basis (
    code, name, description, requires_uom, use_event_quantity, is_active
)
VALUES (upper($1), $2, $3, $4, $5, $6)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name, description = EXCLUDED.description,
    requires_uom = EXCLUDED.requires_uom,
    use_event_quantity = EXCLUDED.use_event_quantity,
    is_active = EXCLUDED.is_active
RETURNING *;

-- 0.5 Upsert billable service.
-- $1 code, $2 name, $3 module code, $4 description,
-- $5 default charge-basis code, $6 default UOM code or NULL, $7 active
INSERT INTO billing_service (
    code, name, module_code, description,
    default_charge_basis_id, default_uom_id, is_active
)
SELECT upper($1), $2, module.code, $4, basis.charge_basis_id, unit.uom_id, $7
FROM charge_basis basis
JOIN app_module module ON module.code = upper($3) AND module.is_active
LEFT JOIN uom unit ON unit.code = $6 AND unit.is_active
WHERE basis.code = $5 AND basis.is_active
  AND ($6::varchar IS NULL OR unit.uom_id IS NOT NULL)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name, module_code = EXCLUDED.module_code,
    description = EXCLUDED.description,
    default_charge_basis_id = EXCLUDED.default_charge_basis_id,
    default_uom_id = EXCLUDED.default_uom_id,
    is_active = EXCLUDED.is_active
RETURNING *;

-- 0.6 Upsert payment method.
-- $1 code, $2 name, $3 reference required, $4 active
INSERT INTO payment_method (code, name, requires_reference, is_active)
VALUES (upper($1), $2, $3, $4)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name, requires_reference = EXCLUDED.requires_reference,
    is_active = EXCLUDED.is_active
RETURNING *;

-- 0.7 Configure one movement type to produce one billing service.
-- $1 movement-type code, $2 billing-service code, $3 active
INSERT INTO billing_service_movement_type (
    billing_service_id, movement_type_id, is_active
)
SELECT service.billing_service_id, movement.movement_type_id, $3
FROM billing_service service, movement_type movement
WHERE service.code = $2 AND movement.code = $1
ON CONFLICT (movement_type_id) DO UPDATE
SET billing_service_id = EXCLUDED.billing_service_id,
    is_active = EXCLUDED.is_active
RETURNING *;

-- =============================================================================
-- 1. BILLING ACCOUNT AND TAX CONFIGURATION
-- =============================================================================

-- 1.1 Create a billing account and snapshot bill-to details.
-- $1 owner ID, $2 account code, $3 account name,
-- $4 bill-to partner ID or NULL, $5 currency code,
-- $6 billing-cycle code, $7 payment-term code,
-- $8 default tax-rule ID or NULL, $9 actor ID
INSERT INTO billing_account (
    owner_id, code, name, bill_to_partner_id, currency_id,
    billing_cycle_id, payment_term_id, default_tax_rule_id,
    bill_to_name, bill_to_tax_number, bill_to_address_1,
    bill_to_address_2, bill_to_city, bill_to_province,
    bill_to_postal_code, bill_to_country_code, created_by, updated_by
)
SELECT owner.organization_id, $2, $3, partner.partner_id,
       currency.currency_id, cycle.billing_cycle_id, term.payment_term_id,
       tax.tax_rule_id,
       COALESCE(partner.name, owner.legal_name, owner.name),
       COALESCE(partner.tax_number, owner.tax_number),
       COALESCE(partner.address_line_1, owner.address_line_1),
       COALESCE(partner.address_line_2, owner.address_line_2),
       COALESCE(partner.city, owner.city),
       COALESCE(partner.province, owner.province),
       COALESCE(partner.postal_code, owner.postal_code),
       COALESCE(partner.country_code, owner.country_code), $9, $9
FROM organization owner
JOIN currency ON currency.code = $5 AND currency.is_active
JOIN billing_cycle cycle ON cycle.code = $6 AND cycle.is_active
JOIN payment_term term ON term.code = $7 AND term.is_active
LEFT JOIN business_partner partner
  ON partner.partner_id = $4 AND partner.owner_id = owner.organization_id AND partner.is_active
LEFT JOIN tax_rule tax
  ON tax.tax_rule_id = $8 AND tax.is_active
 AND (tax.owner_id IS NULL OR tax.owner_id = owner.organization_id)
WHERE owner.organization_id = $1
  AND ($4::uuid IS NULL OR partner.partner_id IS NOT NULL)
  AND ($8::uuid IS NULL OR tax.tax_rule_id IS NOT NULL)
  AND COALESCE(partner.address_line_1, owner.address_line_1) IS NOT NULL
RETURNING *;

-- 1.2 Update billing settings using updated_at as optimistic concurrency.
-- $1 billing-account ID, $2 name, $3 currency code,
-- $4 billing-cycle code, $5 payment-term code,
-- $6 default tax-rule ID or NULL, $7 expected updated_at, $8 actor ID
UPDATE billing_account account
SET name = $2, currency_id = currency.currency_id,
    billing_cycle_id = cycle.billing_cycle_id,
    payment_term_id = term.payment_term_id,
    default_tax_rule_id = tax.tax_rule_id,
    updated_at = clock_timestamp(), updated_by = $8
FROM currency, billing_cycle cycle, payment_term term
LEFT JOIN tax_rule tax ON tax.tax_rule_id = $6 AND tax.is_active
WHERE account.billing_account_id = $1
  AND currency.code = $3 AND currency.is_active
  AND cycle.code = $4 AND cycle.is_active
  AND term.code = $5 AND term.is_active
  AND ($6::uuid IS NULL OR (tax.tax_rule_id IS NOT NULL
       AND (tax.owner_id IS NULL OR tax.owner_id = account.owner_id)))
  AND account.updated_at = $7
RETURNING account.*;

-- 1.3 Soft-deactivate a billing account with no active contract. $1 account ID, $2 actor ID
UPDATE billing_account account
SET is_active = false, updated_at = clock_timestamp(), updated_by = $2
WHERE account.billing_account_id = $1
  AND NOT EXISTS (
      SELECT 1 FROM billing_contract contract
      JOIN document_status status ON status.status_id = contract.status_id
      WHERE contract.billing_account_id = account.billing_account_id
        AND status.code IN ('ACTIVE', 'SUSPENDED')
  )
RETURNING account.*;

-- 1.4 Create a versioned tax rule. No tax rate is supplied by starter data.
-- $1 owner ID or NULL for global, $2 code, $3 name,
-- $4 rate percent, $5 inclusive, $6 effective from,
-- $7 effective until or NULL
INSERT INTO tax_rule (
    owner_id, code, name, rate_percent, is_inclusive,
    effective_from, effective_until
)
SELECT $1, $2, $3, $4, $5, $6, $7
WHERE NOT EXISTS (
    SELECT 1 FROM tax_rule existing
    WHERE existing.owner_id IS NOT DISTINCT FROM $1
      AND existing.code = $2
      AND daterange(existing.effective_from,
                    COALESCE(existing.effective_until + 1, 'infinity'::date), '[)')
          && daterange($6::date, COALESCE($7::date + 1, 'infinity'::date), '[)')
      AND existing.is_active
)
RETURNING *;

-- 1.5 Billing account list. $1 owner ID, $2 include inactive
SELECT account.*, currency.code AS currency_code,
       cycle.code AS billing_cycle_code, term.code AS payment_term_code,
       tax.code AS default_tax_code
FROM billing_account account
JOIN currency ON currency.currency_id = account.currency_id
JOIN billing_cycle cycle ON cycle.billing_cycle_id = account.billing_cycle_id
JOIN payment_term term ON term.payment_term_id = account.payment_term_id
LEFT JOIN tax_rule tax ON tax.tax_rule_id = account.default_tax_rule_id
WHERE account.owner_id = $1 AND ($2 OR account.is_active)
ORDER BY account.code;

-- =============================================================================
-- 2. BILLING CONTRACT
-- =============================================================================

-- 2.1 Create a draft contract and generate its varchar ID.
-- $1 billing-account ID, $2 warehouse ID or NULL,
-- $3 contract number, $4 effective from, $5 effective until or NULL,
-- $6 auto renew, $7 default tax-rule ID or NULL,
-- $8 notes, $9 actor ID
WITH context AS (
    SELECT account.*, warehouse.code AS warehouse_code,
           COALESCE($7, account.default_tax_rule_id) AS selected_tax_rule_id,
           dt.document_type_id, initial_status.status_id
    FROM billing_account account
    LEFT JOIN warehouse_owner scope
      ON scope.owner_id = account.owner_id AND scope.warehouse_id = $2 AND scope.is_active
    LEFT JOIN warehouse ON warehouse.warehouse_id = scope.warehouse_id AND warehouse.is_active
    JOIN document_type dt ON dt.code = 'BILLING_CONTRACT' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    LEFT JOIN tax_rule tax
      ON tax.tax_rule_id = COALESCE($7, account.default_tax_rule_id)
     AND tax.is_active AND (tax.owner_id IS NULL OR tax.owner_id = account.owner_id)
    WHERE account.billing_account_id = $1 AND account.is_active
      AND ($2::uuid IS NULL OR warehouse.warehouse_id IS NOT NULL)
      AND (COALESCE($7, account.default_tax_rule_id) IS NULL OR tax.tax_rule_id IS NOT NULL)
), numbered AS (
    SELECT context.*,
           generate_document_id('BILLING_CONTRACT', NULL, NULL, $4) AS generated_id
    FROM context
)
INSERT INTO billing_contract (
    billing_contract_id, document_type_id, status_id,
    billing_account_id, owner_id, warehouse_id, currency_id,
    default_tax_rule_id, contract_number, effective_from,
    effective_until, auto_renew, notes, created_by, updated_by
)
SELECT generated_id, document_type_id, status_id, billing_account_id,
       owner_id, $2, currency_id, selected_tax_rule_id,
       $3, $4, $5, $6, $8, $9, $9
FROM numbered
RETURNING *;

-- 2.2 Activate a contract only when it has an active approved rate card.
-- $1 contract ID, $2 actor ID
UPDATE billing_contract contract
SET status_id = active.status_id,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = contract.version_no + 1
FROM document_status current_status, document_status active
WHERE contract.billing_contract_id = $1
  AND current_status.status_id = contract.status_id
  AND current_status.code IN ('DRAFT', 'SUSPENDED')
  AND active.document_type_id = contract.document_type_id
  AND active.code = 'ACTIVE'
  AND EXISTS (
      SELECT 1 FROM rate_card card
      JOIN document_status card_status ON card_status.status_id = card.status_id
      WHERE card.billing_contract_id = contract.billing_contract_id
        AND card_status.code IN ('APPROVED', 'ACTIVE')
  )
RETURNING contract.*;

-- 2.3 Suspend an active contract. $1 contract ID, $2 actor ID
UPDATE billing_contract contract
SET status_id = suspended.status_id,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = contract.version_no + 1
FROM document_status current_status, document_status suspended
WHERE contract.billing_contract_id = $1
  AND current_status.status_id = contract.status_id AND current_status.code = 'ACTIVE'
  AND suspended.document_type_id = contract.document_type_id
  AND suspended.code = 'SUSPENDED'
RETURNING contract.*;

-- =============================================================================
-- 3. RATE CARD AND TIERS
-- =============================================================================

-- 3.1 Create a draft rate card.
-- $1 contract ID, $2 version label, $3 effective from,
-- $4 effective until or NULL, $5 notes, $6 actor ID
WITH context AS (
    SELECT contract.billing_contract_id, dt.document_type_id,
           initial_status.status_id
    FROM billing_contract contract
    JOIN document_status contract_status ON contract_status.status_id = contract.status_id
    JOIN document_type dt ON dt.code = 'RATE_CARD' AND dt.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = dt.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE contract.billing_contract_id = $1
      AND contract_status.code IN ('DRAFT', 'ACTIVE', 'SUSPENDED')
      AND $3::date >= contract.effective_from
      AND (contract.effective_until IS NULL OR $4::date IS NOT NULL
           AND $4::date <= contract.effective_until)
), numbered AS (
    SELECT context.*,
           generate_document_id('RATE_CARD', NULL, NULL, $3) AS generated_id
    FROM context
)
INSERT INTO rate_card (
    rate_card_id, document_type_id, status_id, billing_contract_id,
    version_label, effective_from, effective_until,
    notes, created_by, updated_by
)
SELECT generated_id, document_type_id, status_id, billing_contract_id,
       $2, $3, $4, $5, $6, $6
FROM numbered
RETURNING *;

-- 3.2 Add a draft rate line.
-- $1 rate-card ID, $2 line number, $3 service code, $4 charge-basis code,
-- $5 UOM code or NULL, $6 warehouse ID or NULL, $7 item ID or NULL,
-- $8 category ID or NULL, $9 unit rate, $10 included quantity,
-- $11 minimum charge, $12 maximum charge or NULL,
-- $13 rounding increment or NULL, $14 priority, $15 tax-rule ID or NULL,
-- $16 actor ID
INSERT INTO rate_card_line (
    rate_card_line_id, rate_card_id, line_no, billing_service_id,
    charge_basis_id, uom_id, warehouse_id, item_id, category_id,
    unit_rate, included_quantity, minimum_charge, maximum_charge,
    rounding_increment, priority_no, tax_rule_id, created_by
)
SELECT card.rate_card_id || '-L-' || lpad($2::text, 4, '0'),
       card.rate_card_id, $2, service.billing_service_id,
       basis.charge_basis_id, unit.uom_id, warehouse.warehouse_id,
       item.item_id, category.category_id, $9, $10, $11, $12,
       $13, $14, tax.tax_rule_id, $16
FROM rate_card card
JOIN document_status status ON status.status_id = card.status_id
JOIN billing_contract contract ON contract.billing_contract_id = card.billing_contract_id
JOIN billing_service service ON service.code = $3 AND service.is_active
JOIN charge_basis basis ON basis.code = $4 AND basis.is_active
LEFT JOIN uom unit ON unit.code = $5 AND unit.is_active
LEFT JOIN warehouse_owner scope
  ON scope.owner_id = contract.owner_id AND scope.warehouse_id = $6 AND scope.is_active
LEFT JOIN warehouse ON warehouse.warehouse_id = scope.warehouse_id AND warehouse.is_active
LEFT JOIN item ON item.item_id = $7 AND item.owner_id = contract.owner_id AND item.is_active
LEFT JOIN item_category category
  ON category.category_id = $8 AND category.owner_id = contract.owner_id AND category.is_active
LEFT JOIN tax_rule tax
  ON tax.tax_rule_id = $15 AND tax.is_active
 AND (tax.owner_id IS NULL OR tax.owner_id = contract.owner_id)
WHERE card.rate_card_id = $1 AND status.code = 'DRAFT'
  AND ($5::varchar IS NULL OR unit.uom_id IS NOT NULL)
  AND (NOT basis.requires_uom OR unit.uom_id IS NOT NULL)
  AND ($6::uuid IS NULL OR warehouse.warehouse_id IS NOT NULL)
  AND ($7::uuid IS NULL OR item.item_id IS NOT NULL)
  AND ($8::uuid IS NULL OR category.category_id IS NOT NULL)
  AND ($15::uuid IS NULL OR tax.tax_rule_id IS NOT NULL)
RETURNING *;

-- 3.3 Add a non-overlapping quantity tier to a draft line.
-- $1 rate-line ID, $2 tier number, $3 from quantity,
-- $4 until quantity or NULL, $5 unit rate
INSERT INTO rate_card_tier (
    rate_card_tier_id, rate_card_line_id, tier_no,
    from_quantity, until_quantity, unit_rate
)
SELECT line.rate_card_line_id || '-T-' || lpad($2::text, 3, '0'),
       line.rate_card_line_id, $2, $3, $4, $5
FROM rate_card_line line
JOIN rate_card card ON card.rate_card_id = line.rate_card_id
JOIN document_status status ON status.status_id = card.status_id
WHERE line.rate_card_line_id = $1 AND status.code = 'DRAFT'
  AND NOT EXISTS (
      SELECT 1 FROM rate_card_tier existing
      WHERE existing.rate_card_line_id = line.rate_card_line_id
        AND numrange(existing.from_quantity, existing.until_quantity, '[)')
            && numrange($3::numeric, $4::numeric, '[)')
  )
RETURNING *;

-- 3.4 Approve a complete draft rate card. $1 rate-card ID, $2 actor ID
UPDATE rate_card card
SET status_id = approved.status_id,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = card.version_no + 1
FROM document_status current_status, document_status approved
WHERE card.rate_card_id = $1
  AND current_status.status_id = card.status_id AND current_status.code = 'DRAFT'
  AND approved.document_type_id = card.document_type_id
  AND approved.code = 'APPROVED'
  AND EXISTS (SELECT 1 FROM rate_card_line line WHERE line.rate_card_id = card.rate_card_id AND line.is_active)
RETURNING card.*;

-- 3.5 Activate an approved rate card without overlapping another active card.
-- $1 rate-card ID, $2 actor ID
UPDATE rate_card card
SET status_id = active.status_id,
    updated_at = clock_timestamp(), updated_by = $2,
    version_no = card.version_no + 1
FROM document_status current_status, document_status active
WHERE card.rate_card_id = $1
  AND current_status.status_id = card.status_id AND current_status.code = 'APPROVED'
  AND active.document_type_id = card.document_type_id AND active.code = 'ACTIVE'
  AND NOT EXISTS (
      SELECT 1 FROM rate_card other
      JOIN document_status other_status ON other_status.status_id = other.status_id
      WHERE other.billing_contract_id = card.billing_contract_id
        AND other.rate_card_id <> card.rate_card_id
        AND other_status.code = 'ACTIVE'
        AND daterange(other.effective_from,
                      COALESCE(other.effective_until + 1, 'infinity'::date), '[)')
            && daterange(card.effective_from,
                         COALESCE(card.effective_until + 1, 'infinity'::date), '[)')
  )
RETURNING card.*;

-- 3.6 Effective rate-card detail. $1 contract ID, $2 business date
SELECT card.rate_card_id, card.version_label,
       service.code AS service_code, basis.code AS charge_basis_code,
       line.line_no, unit.code AS uom_code, warehouse.code AS warehouse_code,
       item.code AS item_code, category.code AS category_code,
       line.unit_rate, line.included_quantity, line.minimum_charge,
       line.maximum_charge, line.rounding_increment, line.priority_no,
       tax.code AS tax_code, tax.rate_percent, tax.is_inclusive,
       COALESCE((
           SELECT jsonb_agg(jsonb_build_object(
               'tierNo', tier.tier_no, 'from', tier.from_quantity,
               'until', tier.until_quantity, 'unitRate', tier.unit_rate
           ) ORDER BY tier.tier_no)
           FROM rate_card_tier tier
           WHERE tier.rate_card_line_id = line.rate_card_line_id
       ), '[]'::jsonb) AS tiers
FROM rate_card card
JOIN document_status status ON status.status_id = card.status_id
JOIN rate_card_line line ON line.rate_card_id = card.rate_card_id AND line.is_active
JOIN billing_service service ON service.billing_service_id = line.billing_service_id
JOIN charge_basis basis ON basis.charge_basis_id = line.charge_basis_id
LEFT JOIN uom unit ON unit.uom_id = line.uom_id
LEFT JOIN warehouse ON warehouse.warehouse_id = line.warehouse_id
LEFT JOIN item ON item.item_id = line.item_id
LEFT JOIN item_category category ON category.category_id = line.category_id
LEFT JOIN tax_rule tax ON tax.tax_rule_id = line.tax_rule_id
WHERE card.billing_contract_id = $1 AND status.code = 'ACTIVE'
  AND card.effective_from <= $2
  AND (card.effective_until IS NULL OR card.effective_until >= $2)
ORDER BY service.code, line.priority_no, line.line_no;
