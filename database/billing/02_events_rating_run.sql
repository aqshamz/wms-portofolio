-- Billable-event capture, rating, and billing-run workflow.
-- Execute one numbered operation at a time using prepared parameters.

SET search_path TO wms, public;

-- =============================================================================
-- 1. CAPTURE IMMUTABLE BILLABLE EVENTS
-- =============================================================================

-- 1.1 Capture one generic operational event idempotently.
-- $1 contract ID, $2 service code, $3 warehouse ID,
-- $4 item ID or NULL, $5 handling-unit ID or NULL,
-- $6 business date, $7 occurred_at, $8 source document-type code or NULL,
-- $9 source document ID, $10 source line ID or NULL, $11 event key,
-- $12 quantity, $13 UOM code or NULL, $14 JSON attributes, $15 actor ID
WITH context AS (
    SELECT contract.billing_contract_id, contract.billing_account_id,
           contract.owner_id, contract.warehouse_id AS contract_warehouse_id,
           warehouse.warehouse_id, warehouse.code AS warehouse_code,
           service.billing_service_id,
           item.item_id, item.category_id,
           hu.handling_unit_id,
           COALESCE(unit.uom_id, service.default_uom_id) AS uom_id,
           source_type.document_type_id AS source_document_type_id,
           event_type.document_type_id, initial_status.status_id
    FROM billing_contract contract
    JOIN billing_account account
      ON account.billing_account_id = contract.billing_account_id AND account.is_active
    JOIN document_status contract_status ON contract_status.status_id = contract.status_id
    JOIN warehouse_owner scope
      ON scope.owner_id = contract.owner_id AND scope.warehouse_id = $3 AND scope.is_active
    JOIN warehouse ON warehouse.warehouse_id = scope.warehouse_id AND warehouse.is_active
    JOIN billing_service service ON service.code = $2 AND service.is_active
    LEFT JOIN item
      ON item.item_id = $4 AND item.owner_id = contract.owner_id AND item.is_active
    LEFT JOIN handling_unit hu
      ON hu.handling_unit_id = $5 AND hu.owner_id = contract.owner_id
     AND hu.warehouse_id = warehouse.warehouse_id
    LEFT JOIN uom unit ON unit.code = $13 AND unit.is_active
    LEFT JOIN document_type source_type
      ON source_type.code = $8 AND source_type.is_active
    JOIN document_type event_type
      ON event_type.code = 'BILLABLE_EVENT' AND event_type.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = event_type.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE contract.billing_contract_id = $1
      AND contract_status.code = 'ACTIVE'
      AND contract.effective_from <= $6::date
      AND (contract.effective_until IS NULL OR contract.effective_until >= $6::date)
      AND (contract.warehouse_id IS NULL OR contract.warehouse_id = warehouse.warehouse_id)
      AND ($4::uuid IS NULL OR item.item_id IS NOT NULL)
      AND ($5::varchar IS NULL OR hu.handling_unit_id IS NOT NULL)
      AND ($8::varchar IS NULL OR source_type.document_type_id IS NOT NULL)
      AND ($13::varchar IS NULL OR unit.uom_id IS NOT NULL)
      AND (NOT EXISTS (
              SELECT 1 FROM charge_basis basis
              WHERE basis.charge_basis_id = service.default_charge_basis_id
                AND basis.requires_uom
           ) OR COALESCE(unit.uom_id, service.default_uom_id) IS NOT NULL)
), numbered AS (
    SELECT context.*,
           generate_document_id(
               'BILLABLE_EVENT', NULL, warehouse_code, $6::date
           ) AS generated_id
    FROM context
)
INSERT INTO billable_event (
    billable_event_id, document_type_id, status_id,
    billing_account_id, billing_contract_id, billing_service_id,
    owner_id, warehouse_id, item_id, category_id, handling_unit_id,
    business_date, occurred_at, source_document_type_id,
    source_document_id, source_line_id, event_key,
    quantity, uom_id, attributes, created_by
)
SELECT generated_id, document_type_id, status_id,
       billing_account_id, billing_contract_id, billing_service_id,
       owner_id, warehouse_id, item_id, category_id, handling_unit_id,
       $6, $7, source_document_type_id, $9, $10, $11,
       $12, uom_id, COALESCE($14::jsonb, '{}'::jsonb), $15
FROM numbered
ON CONFLICT DO NOTHING
RETURNING *;

-- 1.2 Capture all configured inventory movements for a contract and period.
-- The billing_service_movement_type master controls the mapping.
-- $1 contract ID, $2 from date, $3 until date, $4 actor ID
WITH eligible AS (
    SELECT movement.*, warehouse.code AS warehouse_code,
           item.category_id, service.billing_service_id,
           contract.billing_account_id, contract.billing_contract_id,
           event_type.document_type_id, initial_status.status_id
    FROM billing_contract contract
    JOIN document_status contract_status ON contract_status.status_id = contract.status_id
    JOIN inventory_movement movement ON movement.owner_id = contract.owner_id
    JOIN warehouse ON warehouse.warehouse_id = movement.warehouse_id
    JOIN item ON item.item_id = movement.item_id
    JOIN billing_service_movement_type mapping
      ON mapping.movement_type_id = movement.movement_type_id AND mapping.is_active
    JOIN billing_service service
      ON service.billing_service_id = mapping.billing_service_id AND service.is_active
    JOIN document_type event_type
      ON event_type.code = 'BILLABLE_EVENT' AND event_type.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = event_type.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE contract.billing_contract_id = $1
      AND contract_status.code IN ('ACTIVE', 'SUSPENDED')
      AND movement.business_date BETWEEN $2::date AND $3::date
      AND movement.business_date >= contract.effective_from
      AND (contract.effective_until IS NULL
           OR movement.business_date <= contract.effective_until)
      AND (contract.warehouse_id IS NULL
           OR contract.warehouse_id = movement.warehouse_id)
), numbered AS (
    SELECT eligible.*,
           generate_document_id(
               'BILLABLE_EVENT', NULL, warehouse_code, business_date
           ) AS generated_id
    FROM eligible
)
INSERT INTO billable_event (
    billable_event_id, document_type_id, status_id,
    billing_account_id, billing_contract_id, billing_service_id,
    owner_id, warehouse_id, item_id, category_id, handling_unit_id,
    business_date, occurred_at, source_document_id, source_line_id,
    event_key, quantity, uom_id, attributes, created_by
)
SELECT generated_id, document_type_id, status_id,
       billing_account_id, billing_contract_id, billing_service_id,
       owner_id, warehouse_id, item_id, category_id, handling_unit_id,
       business_date, occurred_at, movement_id, source_line_id,
       'INVENTORY_MOVEMENT', quantity, uom_id,
       jsonb_build_object(
           'movementId', movement_id,
           'operationalSourceDocumentId', source_document_id
       ), $4
FROM numbered
ON CONFLICT DO NOTHING
RETURNING *;

-- 1.3 End-of-day unit-storage snapshot. Run once after operational posting.
-- $1 contract ID, $2 snapshot/business date, $3 actor ID
WITH eligible AS (
    SELECT balance.*, warehouse.code AS warehouse_code, item.category_id,
           contract.billing_account_id, contract.billing_contract_id,
           service.billing_service_id, event_type.document_type_id,
           initial_status.status_id
    FROM billing_contract contract
    JOIN document_status contract_status ON contract_status.status_id = contract.status_id
    JOIN inventory_balance balance ON balance.owner_id = contract.owner_id
    JOIN warehouse ON warehouse.warehouse_id = balance.warehouse_id
    JOIN item ON item.item_id = balance.item_id
    JOIN billing_service service
      ON service.code = 'STORAGE_UNIT_DAY' AND service.is_active
    JOIN document_type event_type
      ON event_type.code = 'BILLABLE_EVENT' AND event_type.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = event_type.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE contract.billing_contract_id = $1
      AND contract_status.code = 'ACTIVE'
      AND contract.effective_from <= $2::date
      AND (contract.effective_until IS NULL OR contract.effective_until >= $2::date)
      AND (contract.warehouse_id IS NULL OR contract.warehouse_id = balance.warehouse_id)
      AND balance.on_hand_qty > 0
), numbered AS (
    SELECT eligible.*,
           generate_document_id(
               'BILLABLE_EVENT', NULL, warehouse_code, $2::date
           ) AS generated_id
    FROM eligible
)
INSERT INTO billable_event (
    billable_event_id, document_type_id, status_id,
    billing_account_id, billing_contract_id, billing_service_id,
    owner_id, warehouse_id, item_id, category_id, handling_unit_id,
    business_date, occurred_at, source_document_id, event_key,
    quantity, uom_id, attributes, created_by
)
SELECT generated_id, document_type_id, status_id,
       billing_account_id, billing_contract_id, billing_service_id,
       owner_id, warehouse_id, item_id, category_id, handling_unit_id,
       $2, clock_timestamp(), balance_id, $2::text,
       on_hand_qty, uom_id,
       jsonb_build_object('snapshotVersion', version_no), $3
FROM numbered
ON CONFLICT DO NOTHING
RETURNING *;

-- 1.4 End-of-day occupied-pallet/HU snapshot (one event per non-empty HU).
-- $1 contract ID, $2 snapshot/business date, $3 actor ID
WITH eligible AS (
    SELECT DISTINCT balance.owner_id, balance.warehouse_id,
           balance.handling_unit_id, warehouse.code AS warehouse_code,
           contract.billing_account_id, contract.billing_contract_id,
           service.billing_service_id, service.default_uom_id AS uom_id,
           event_type.document_type_id, initial_status.status_id
    FROM billing_contract contract
    JOIN document_status contract_status ON contract_status.status_id = contract.status_id
    JOIN inventory_balance balance ON balance.owner_id = contract.owner_id
    JOIN warehouse ON warehouse.warehouse_id = balance.warehouse_id
    JOIN billing_service service
      ON service.code = 'STORAGE_PALLET_DAY' AND service.is_active
    JOIN document_type event_type
      ON event_type.code = 'BILLABLE_EVENT' AND event_type.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = event_type.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE contract.billing_contract_id = $1
      AND contract_status.code = 'ACTIVE'
      AND contract.effective_from <= $2::date
      AND (contract.effective_until IS NULL OR contract.effective_until >= $2::date)
      AND (contract.warehouse_id IS NULL OR contract.warehouse_id = balance.warehouse_id)
      AND balance.on_hand_qty > 0 AND balance.handling_unit_id IS NOT NULL
), numbered AS (
    SELECT eligible.*,
           generate_document_id(
               'BILLABLE_EVENT', NULL, warehouse_code, $2::date
           ) AS generated_id
    FROM eligible
)
INSERT INTO billable_event (
    billable_event_id, document_type_id, status_id,
    billing_account_id, billing_contract_id, billing_service_id,
    owner_id, warehouse_id, handling_unit_id,
    business_date, occurred_at, source_document_id, event_key,
    quantity, uom_id, attributes, created_by
)
SELECT generated_id, document_type_id, status_id,
       billing_account_id, billing_contract_id, billing_service_id,
       owner_id, warehouse_id, handling_unit_id,
       $2, clock_timestamp(), handling_unit_id, $2::text,
       1, uom_id, jsonb_build_object('occupied', true), $3
FROM numbered
ON CONFLICT DO NOTHING
RETURNING *;

-- 1.5 Exclude a pending event, retaining the reason in the immutable record.
-- $1 event ID, $2 billing reason code, $3 required note, $4 actor ID
UPDATE billable_event event
SET status_id = excluded.status_id,
    exclusion_reason_code_id = reason.reason_code_id,
    attributes = event.attributes || jsonb_build_object('exclusionNote', $3),
    excluded_at = clock_timestamp(), excluded_by = $4
FROM document_status current_status, document_status excluded, reason_code reason
WHERE event.billable_event_id = $1
  AND current_status.status_id = event.status_id AND current_status.code = 'PENDING'
  AND excluded.document_type_id = event.document_type_id AND excluded.code = 'EXCLUDED'
  AND reason.module_code = 'BILLING' AND reason.code = $2 AND reason.is_active
  AND nullif(btrim($3), '') IS NOT NULL
RETURNING event.*;

-- =============================================================================
-- 2. BILLING RUN AND RATING
-- =============================================================================

-- 2.1 Create a non-overlapping draft run.
-- $1 contract ID, $2 period from, $3 period until, $4 notes, $5 actor ID
WITH context AS (
    SELECT contract.billing_contract_id, contract.billing_account_id,
           contract.currency_id, run_type.document_type_id, initial_status.status_id
    FROM billing_contract contract
    JOIN document_status contract_status ON contract_status.status_id = contract.status_id
    JOIN document_type run_type ON run_type.code = 'BILLING_RUN' AND run_type.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = run_type.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE contract.billing_contract_id = $1
      AND contract_status.code IN ('ACTIVE', 'SUSPENDED')
      AND $2::date >= contract.effective_from
      AND (contract.effective_until IS NULL OR $3::date <= contract.effective_until)
      AND NOT EXISTS (
          SELECT 1 FROM billing_run existing
          JOIN document_status status ON status.status_id = existing.status_id
          WHERE existing.billing_contract_id = contract.billing_contract_id
            AND status.code <> 'CANCELLED'
            AND daterange(existing.period_from, existing.period_until + 1, '[)')
                && daterange($2::date, $3::date + 1, '[)')
      )
), numbered AS (
    SELECT context.*,
           generate_document_id('BILLING_RUN', NULL, NULL, $3::date) AS generated_id
    FROM context
)
INSERT INTO billing_run (
    billing_run_id, document_type_id, status_id,
    billing_account_id, billing_contract_id, currency_id,
    period_from, period_until, notes, created_by
)
SELECT generated_id, document_type_id, status_id,
       billing_account_id, billing_contract_id, currency_id,
       $2, $3, $4, $5
FROM numbered
RETURNING *;

-- 2.2 Rate one pending event into a draft run.
-- The selected line is the most specific match, then lowest priority/line number.
-- $1 billing-run ID, $2 event ID, $3 actor ID
WITH locked AS (
    SELECT run.billing_run_id, run.billing_contract_id,
           run.billing_account_id, run.currency_id, currency.decimal_scale,
           event.billable_event_id, event.billing_service_id,
           event.warehouse_id, event.item_id, event.category_id,
           event.business_date, event.quantity, event.uom_id,
           contract.default_tax_rule_id AS contract_tax_rule_id,
           account.default_tax_rule_id AS account_tax_rule_id
    FROM billing_run run
    JOIN document_status run_status ON run_status.status_id = run.status_id
    JOIN billable_event event
      ON event.billable_event_id = $2
     AND event.billing_contract_id = run.billing_contract_id
     AND event.billing_account_id = run.billing_account_id
     AND event.business_date BETWEEN run.period_from AND run.period_until
    JOIN document_status event_status ON event_status.status_id = event.status_id
    JOIN billing_contract contract
      ON contract.billing_contract_id = run.billing_contract_id
    JOIN billing_account account
      ON account.billing_account_id = run.billing_account_id
    JOIN currency ON currency.currency_id = run.currency_id
    WHERE run.billing_run_id = $1
      AND run_status.code = 'DRAFT' AND event_status.code = 'PENDING'
    FOR UPDATE OF run, event
), selected AS (
    SELECT locked.*, matched.rate_card_line_id, matched.charge_basis_id,
           matched.use_event_quantity, matched.line_uom_id,
           matched.unit_rate AS base_unit_rate,
           matched.included_quantity, matched.minimum_charge,
           matched.maximum_charge, matched.rounding_increment,
           matched.line_tax_rule_id
    FROM locked
    JOIN LATERAL (
        SELECT line.rate_card_line_id, line.charge_basis_id,
               basis.use_event_quantity, line.uom_id AS line_uom_id,
               line.unit_rate, line.included_quantity,
               line.minimum_charge, line.maximum_charge,
               line.rounding_increment, line.tax_rule_id AS line_tax_rule_id
        FROM rate_card card
        JOIN document_status card_status ON card_status.status_id = card.status_id
        JOIN rate_card_line line ON line.rate_card_id = card.rate_card_id AND line.is_active
        JOIN charge_basis basis ON basis.charge_basis_id = line.charge_basis_id AND basis.is_active
        WHERE card.billing_contract_id = locked.billing_contract_id
          AND card_status.code = 'ACTIVE'
          AND locked.business_date BETWEEN card.effective_from
              AND COALESCE(card.effective_until, 'infinity'::date)
          AND line.billing_service_id = locked.billing_service_id
          AND (line.warehouse_id IS NULL OR line.warehouse_id = locked.warehouse_id)
          AND (line.item_id IS NULL OR line.item_id = locked.item_id)
          AND (line.category_id IS NULL OR line.category_id = locked.category_id)
          AND (line.uom_id IS NULL OR line.uom_id = locked.uom_id)
        ORDER BY
          ((line.warehouse_id IS NOT NULL)::int
           + (line.item_id IS NOT NULL)::int
           + (line.category_id IS NOT NULL)::int
           + (line.uom_id IS NOT NULL)::int) DESC,
          line.priority_no, line.line_no
        LIMIT 1
    ) matched ON true
), quantities AS (
    SELECT selected.*,
           CASE WHEN use_event_quantity THEN quantity ELSE 1::numeric END AS billed_qty
    FROM selected
), priced AS (
    SELECT quantities.*,
           COALESCE(tier.unit_rate, base_unit_rate) AS selected_unit_rate,
           GREATEST(billed_qty - included_quantity, 0) AS chargeable_qty,
           tax.tax_rule_id AS applied_tax_rule_id,
           COALESCE(tax.rate_percent, 0) AS applied_tax_rate,
           COALESCE(tax.is_inclusive, false) AS tax_inclusive
    FROM quantities
    LEFT JOIN LATERAL (
        SELECT tier.unit_rate
        FROM rate_card_tier tier
        WHERE tier.rate_card_line_id = quantities.rate_card_line_id
          AND quantities.billed_qty >= tier.from_quantity
          AND (tier.until_quantity IS NULL OR quantities.billed_qty < tier.until_quantity)
        ORDER BY tier.from_quantity DESC LIMIT 1
    ) tier ON true
    LEFT JOIN LATERAL (
        SELECT rule.tax_rule_id, rule.rate_percent, rule.is_inclusive
        FROM tax_rule rule
        WHERE rule.tax_rule_id = COALESCE(
                  quantities.line_tax_rule_id,
                  quantities.contract_tax_rule_id,
                  quantities.account_tax_rule_id
              )
          AND rule.is_active
          AND quantities.business_date BETWEEN rule.effective_from
              AND COALESCE(rule.effective_until, 'infinity'::date)
    ) tax ON true
    WHERE COALESCE(
              quantities.line_tax_rule_id,
              quantities.contract_tax_rule_id,
              quantities.account_tax_rule_id
          ) IS NULL
       OR tax.tax_rule_id IS NOT NULL
), extended AS (
    SELECT priced.*,
           CASE
             WHEN rounding_increment IS NULL THEN chargeable_qty
             ELSE ceil(chargeable_qty / rounding_increment) * rounding_increment
           END AS rounded_chargeable_qty
    FROM priced
), base_amount AS (
    SELECT extended.*,
           round(LEAST(
               COALESCE(maximum_charge, 9999999999999999.9999::numeric),
               GREATEST(minimum_charge, rounded_chargeable_qty * selected_unit_rate)
           ), decimal_scale) AS commercial_amount
    FROM extended
), amounts AS (
    SELECT base_amount.*,
           CASE WHEN tax_inclusive AND applied_tax_rate > 0
                THEN round(commercial_amount / (1 + applied_tax_rate / 100), decimal_scale)
                ELSE commercial_amount END AS calculated_net
    FROM base_amount
), final_amounts AS (
    SELECT amounts.*,
           CASE WHEN tax_inclusive
                THEN commercial_amount - calculated_net
                ELSE round(calculated_net * applied_tax_rate / 100, decimal_scale)
           END AS calculated_tax
    FROM amounts
), inserted AS (
    INSERT INTO billing_charge (
        billing_charge_id, billing_run_id, billable_event_id,
        rate_card_line_id, billed_quantity, uom_id, unit_rate,
        net_amount, tax_rule_id, tax_rate_percent, tax_amount,
        gross_amount, currency_id, created_by
    )
    SELECT generate_document_id('BILLING_CHARGE', NULL, NULL, business_date),
           billing_run_id, billable_event_id, rate_card_line_id,
           billed_qty, COALESCE(line_uom_id, uom_id), selected_unit_rate,
           calculated_net, applied_tax_rule_id, applied_tax_rate,
           calculated_tax, calculated_net + calculated_tax,
           currency_id, $3
    FROM final_amounts
    RETURNING billable_event_id
)
UPDATE billable_event event
SET status_id = rated.status_id,
    rated_at = clock_timestamp(), rated_by = $3
FROM inserted, document_status rated
WHERE event.billable_event_id = inserted.billable_event_id
  AND rated.document_type_id = event.document_type_id AND rated.code = 'RATED'
RETURNING event.*;

-- 2.3 Finish calculation only when every in-period event is rated or excluded.
-- $1 billing-run ID
WITH totals AS (
    SELECT run.billing_run_id,
           COALESCE(sum(charge.net_amount), 0) AS net_amount,
           COALESCE(sum(charge.tax_amount), 0) AS tax_amount,
           COALESCE(sum(charge.gross_amount), 0) AS gross_amount
    FROM billing_run run
    LEFT JOIN billing_charge charge ON charge.billing_run_id = run.billing_run_id
    WHERE run.billing_run_id = $1
    GROUP BY run.billing_run_id
)
UPDATE billing_run run
SET status_id = calculated.status_id,
    net_amount = totals.net_amount,
    tax_amount = totals.tax_amount,
    gross_amount = totals.gross_amount,
    calculated_at = clock_timestamp()
FROM totals, document_status current_status, document_status calculated
WHERE run.billing_run_id = totals.billing_run_id
  AND current_status.status_id = run.status_id AND current_status.code = 'DRAFT'
  AND calculated.document_type_id = run.document_type_id
  AND calculated.code = 'CALCULATED'
  AND EXISTS (SELECT 1 FROM billing_charge charge WHERE charge.billing_run_id = run.billing_run_id)
  AND NOT EXISTS (
      SELECT 1 FROM billable_event event
      JOIN document_status status ON status.status_id = event.status_id
      WHERE event.billing_contract_id = run.billing_contract_id
        AND event.business_date BETWEEN run.period_from AND run.period_until
        AND status.code = 'PENDING'
  )
RETURNING run.*;

-- 2.4 Review/approve a calculated run. $1 billing-run ID, $2 reviewer ID
UPDATE billing_run run
SET status_id = reviewed.status_id,
    reviewed_at = clock_timestamp(), reviewed_by = $2
FROM document_status current_status, document_status reviewed
WHERE run.billing_run_id = $1
  AND current_status.status_id = run.status_id AND current_status.code = 'CALCULATED'
  AND reviewed.document_type_id = run.document_type_id AND reviewed.code = 'REVIEWED'
RETURNING run.*;

-- 2.5 Run review detail. $1 billing-run ID
SELECT run.billing_run_id, run.period_from, run.period_until,
       run_status.code AS run_status, account.code AS billing_account_code,
       currency.code AS currency_code,
       event.billable_event_id, service.code AS service_code,
       event.business_date, event.source_document_id, event.source_line_id,
       event.quantity AS event_quantity, event_uom.code AS event_uom,
       charge.billed_quantity, charge.unit_rate, charge.net_amount,
       charge.tax_rate_percent, charge.tax_amount, charge.gross_amount,
       rate_line.rate_card_line_id
FROM billing_run run
JOIN document_status run_status ON run_status.status_id = run.status_id
JOIN billing_account account ON account.billing_account_id = run.billing_account_id
JOIN currency ON currency.currency_id = run.currency_id
JOIN billing_charge charge ON charge.billing_run_id = run.billing_run_id
JOIN billable_event event ON event.billable_event_id = charge.billable_event_id
JOIN billing_service service ON service.billing_service_id = event.billing_service_id
JOIN rate_card_line rate_line ON rate_line.rate_card_line_id = charge.rate_card_line_id
LEFT JOIN uom event_uom ON event_uom.uom_id = event.uom_id
WHERE run.billing_run_id = $1
ORDER BY event.business_date, service.code, event.billable_event_id;
