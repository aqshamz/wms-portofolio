-- Billing reports. Execute one numbered query at a time.

SET search_path TO wms, public;

-- 5.1 Contract and rate-card coverage.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 effective on date, $5 contract status or NULL
SELECT account.code AS billing_account_code, owner.code AS owner_code,
       contract.billing_contract_id, contract.contract_number,
       warehouse.code AS warehouse_code, contract_status.code AS contract_status_code,
       contract.effective_from AS contract_from,
       contract.effective_until AS contract_until,
       currency.code AS currency_code, card.rate_card_id,
       card.version_label, card_status.code AS rate_card_status_code,
       card.effective_from AS rate_card_from, card.effective_until AS rate_card_until,
       count(DISTINCT line.rate_card_line_id) AS rate_line_count,
       count(DISTINCT line.billing_service_id) AS covered_service_count,
       count(DISTINCT tier.rate_card_tier_id) AS tier_count
FROM billing_contract contract
JOIN billing_account account ON account.billing_account_id = contract.billing_account_id
JOIN organization owner ON owner.organization_id = contract.owner_id
JOIN currency ON currency.currency_id = contract.currency_id
JOIN document_status contract_status ON contract_status.status_id = contract.status_id
LEFT JOIN warehouse ON warehouse.warehouse_id = contract.warehouse_id
LEFT JOIN rate_card card ON card.billing_contract_id = contract.billing_contract_id
LEFT JOIN document_status card_status ON card_status.status_id = card.status_id
LEFT JOIN rate_card_line line ON line.rate_card_id = card.rate_card_id AND line.is_active
LEFT JOIN rate_card_tier tier ON tier.rate_card_line_id = line.rate_card_line_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.BILLING'
)
  AND EXISTS (
    SELECT 1 FROM report_owner_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = contract.owner_id
  )
  AND (contract.warehouse_id IS NULL OR EXISTS (
      SELECT 1 FROM report_warehouse_scope scope
      WHERE scope.account_id = $1 AND scope.owner_id = contract.owner_id
        AND scope.warehouse_id = contract.warehouse_id
  ))
  AND ($2::uuid IS NULL OR contract.owner_id = $2)
  AND ($3::uuid IS NULL OR contract.warehouse_id = $3)
  AND contract.effective_from <= $4::date
  AND (contract.effective_until IS NULL OR contract.effective_until >= $4::date)
  AND ($5::varchar IS NULL OR contract_status.code = $5)
GROUP BY account.code, owner.code, contract.billing_contract_id,
         warehouse.code, contract_status.code, currency.code,
         card.rate_card_id, card_status.code
ORDER BY owner.code, account.code, contract.contract_number, card.effective_from DESC;

-- 5.2 Billable-event status summary by service and business date.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 from date, $5 until date, $6 service code or NULL
SELECT event.business_date, owner.code AS owner_code,
       warehouse.code AS warehouse_code, account.code AS billing_account_code,
       service.code AS service_code, status.code AS event_status_code,
       count(*) AS event_count, sum(event.quantity) AS event_quantity,
       uom.code AS uom_code,
       count(*) FILTER (WHERE event.exclusion_reason_code_id IS NOT NULL)
         AS excluded_event_count
FROM billable_event event
JOIN billing_account account ON account.billing_account_id = event.billing_account_id
JOIN organization owner ON owner.organization_id = event.owner_id
LEFT JOIN warehouse ON warehouse.warehouse_id = event.warehouse_id
JOIN billing_service service ON service.billing_service_id = event.billing_service_id
JOIN document_status status ON status.status_id = event.status_id
LEFT JOIN uom ON uom.uom_id = event.uom_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.BILLING'
)
  AND EXISTS (
    SELECT 1 FROM report_owner_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = event.owner_id
  )
  AND (event.warehouse_id IS NULL OR EXISTS (
      SELECT 1 FROM report_warehouse_scope scope
      WHERE scope.account_id = $1 AND scope.owner_id = event.owner_id
        AND scope.warehouse_id = event.warehouse_id
  ))
  AND ($2::uuid IS NULL OR event.owner_id = $2)
  AND ($3::uuid IS NULL OR event.warehouse_id = $3)
  AND event.business_date BETWEEN $4::date AND $5::date
  AND ($6::varchar IS NULL OR service.code = $6)
GROUP BY event.business_date, owner.code, warehouse.code,
         account.code, service.code, status.code, uom.code
ORDER BY event.business_date, owner.code, account.code, service.code, status.code;

-- 5.3 Rated revenue by service.
-- $1 account ID, $2 owner ID or NULL, $3 warehouse ID or NULL,
-- $4 event from date, $5 event until date, $6 billing-account ID or NULL
SELECT event.business_date, owner.code AS owner_code,
       warehouse.code AS warehouse_code, account.code AS billing_account_code,
       service.code AS service_code, currency.code AS currency_code,
       count(charge.billing_charge_id) AS charge_count,
       sum(charge.billed_quantity) AS billed_quantity,
       sum(charge.net_amount) AS net_amount,
       sum(charge.tax_amount) AS tax_amount,
       sum(charge.gross_amount) AS gross_amount,
       count(DISTINCT charge.tax_rule_id)
         FILTER (WHERE charge.tax_rule_id IS NOT NULL) AS applied_tax_rule_count
FROM billing_charge charge
JOIN billable_event event ON event.billable_event_id = charge.billable_event_id
JOIN billing_account account ON account.billing_account_id = event.billing_account_id
JOIN organization owner ON owner.organization_id = event.owner_id
LEFT JOIN warehouse ON warehouse.warehouse_id = event.warehouse_id
JOIN billing_service service ON service.billing_service_id = event.billing_service_id
JOIN currency ON currency.currency_id = charge.currency_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.BILLING'
)
  AND EXISTS (
    SELECT 1 FROM report_owner_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = event.owner_id
  )
  AND (event.warehouse_id IS NULL OR EXISTS (
      SELECT 1 FROM report_warehouse_scope scope
      WHERE scope.account_id = $1 AND scope.owner_id = event.owner_id
        AND scope.warehouse_id = event.warehouse_id
  ))
  AND ($2::uuid IS NULL OR event.owner_id = $2)
  AND ($3::uuid IS NULL OR event.warehouse_id = $3)
  AND event.business_date BETWEEN $4::date AND $5::date
  AND ($6::uuid IS NULL OR event.billing_account_id = $6)
GROUP BY event.business_date, owner.code, warehouse.code,
         account.code, service.code, currency.code
ORDER BY event.business_date, owner.code, account.code, service.code;

-- 5.4 Billing-run reconciliation.
-- $1 account ID, $2 owner ID or NULL, $3 billing-account ID or NULL,
-- $4 period from, $5 period until, $6 status code or NULL
WITH charge_totals AS (
    SELECT billing_run_id, count(*) AS charge_count,
           sum(net_amount) AS charge_net,
           sum(tax_amount) AS charge_tax,
           sum(gross_amount) AS charge_gross
    FROM billing_charge GROUP BY billing_run_id
)
SELECT run.billing_run_id, owner.code AS owner_code,
       account.code AS billing_account_code, contract.contract_number,
       status.code AS run_status_code, run.period_from, run.period_until,
       currency.code AS currency_code,
       COALESCE(charge_totals.charge_count, 0) AS charge_count,
       run.net_amount, COALESCE(charge_totals.charge_net, 0) AS calculated_charge_net,
       run.tax_amount, COALESCE(charge_totals.charge_tax, 0) AS calculated_charge_tax,
       run.gross_amount, COALESCE(charge_totals.charge_gross, 0) AS calculated_charge_gross,
       run.gross_amount - COALESCE(charge_totals.charge_gross, 0) AS variance_amount,
       run.calculated_at, run.reviewed_at, reviewer.username AS reviewed_by,
       invoice.billing_invoice_id
FROM billing_run run
JOIN billing_account account ON account.billing_account_id = run.billing_account_id
JOIN organization owner ON owner.organization_id = account.owner_id
JOIN billing_contract contract ON contract.billing_contract_id = run.billing_contract_id
JOIN document_status status ON status.status_id = run.status_id
JOIN currency ON currency.currency_id = run.currency_id
LEFT JOIN charge_totals ON charge_totals.billing_run_id = run.billing_run_id
LEFT JOIN app_account reviewer ON reviewer.account_id = run.reviewed_by
LEFT JOIN billing_invoice invoice ON invoice.billing_run_id = run.billing_run_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.BILLING'
)
  AND EXISTS (
    SELECT 1 FROM report_owner_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = account.owner_id
  )
  AND ($2::uuid IS NULL OR account.owner_id = $2)
  AND ($3::uuid IS NULL OR run.billing_account_id = $3)
  AND run.period_from <= $5::date AND run.period_until >= $4::date
  AND ($6::varchar IS NULL OR status.code = $6)
ORDER BY run.period_until DESC, run.billing_run_id;

-- 5.5 Invoice aging and open receivables.
-- $1 account ID, $2 owner ID or NULL, $3 billing-account ID or NULL,
-- $4 as-of date, $5 open invoices only, $6 limit, $7 offset
SELECT invoice.billing_invoice_id, owner.code AS owner_code,
       account.code AS billing_account_code, status.code AS invoice_status_code,
       invoice.invoice_date, invoice.due_date, currency.code AS currency_code,
       invoice.net_amount, invoice.adjustment_amount, invoice.tax_amount,
       invoice.gross_amount, invoice.credited_amount, invoice.paid_amount,
       invoice.balance_due,
       GREATEST($4::date - invoice.due_date, 0) AS days_overdue,
       CASE
         WHEN invoice.balance_due = 0 THEN 'CLOSED'
         WHEN $4::date <= invoice.due_date THEN 'CURRENT'
         WHEN $4::date - invoice.due_date <= 30 THEN '1-30'
         WHEN $4::date - invoice.due_date <= 60 THEN '31-60'
         WHEN $4::date - invoice.due_date <= 90 THEN '61-90'
         ELSE '90+'
       END AS aging_bucket
FROM billing_invoice invoice
JOIN billing_account account ON account.billing_account_id = invoice.billing_account_id
JOIN organization owner ON owner.organization_id = account.owner_id
JOIN document_status status ON status.status_id = invoice.status_id
JOIN currency ON currency.currency_id = invoice.currency_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.BILLING'
)
  AND EXISTS (
    SELECT 1 FROM report_owner_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = account.owner_id
  )
  AND ($2::uuid IS NULL OR account.owner_id = $2)
  AND ($3::uuid IS NULL OR invoice.billing_account_id = $3)
  AND invoice.invoice_date <= $4::date
  AND (NOT $5 OR invoice.balance_due > 0)
ORDER BY invoice.balance_due > 0 DESC, invoice.due_date, invoice.billing_invoice_id
LIMIT $6 OFFSET $7;

-- 5.6 Payment receipt and allocation report.
-- $1 account ID, $2 owner ID or NULL, $3 billing-account ID or NULL,
-- $4 from date, $5 until date, $6 payment status or NULL,
-- $7 limit, $8 offset
SELECT payment.payment_id, payment.payment_date, owner.code AS owner_code,
       account.code AS billing_account_code, status.code AS payment_status_code,
       method.code AS payment_method_code, currency.code AS currency_code,
       payment.amount, payment.unapplied_amount,
       payment.amount - payment.unapplied_amount AS allocated_amount,
       payment.external_reference, allocation.allocation_no,
       allocation.allocated_amount AS invoice_allocation_amount,
       invoice.billing_invoice_id, invoice.invoice_date, invoice.balance_due,
       allocation.allocated_at, allocator.username AS allocated_by
FROM billing_payment payment
JOIN billing_account account ON account.billing_account_id = payment.billing_account_id
JOIN organization owner ON owner.organization_id = account.owner_id
JOIN document_status status ON status.status_id = payment.status_id
JOIN payment_method method ON method.payment_method_id = payment.payment_method_id
JOIN currency ON currency.currency_id = payment.currency_id
LEFT JOIN billing_payment_allocation allocation ON allocation.payment_id = payment.payment_id
LEFT JOIN billing_invoice invoice
  ON invoice.billing_invoice_id = allocation.billing_invoice_id
LEFT JOIN app_account allocator ON allocator.account_id = allocation.allocated_by
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.BILLING'
)
  AND EXISTS (
    SELECT 1 FROM report_owner_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = account.owner_id
  )
  AND ($2::uuid IS NULL OR account.owner_id = $2)
  AND ($3::uuid IS NULL OR payment.billing_account_id = $3)
  AND payment.payment_date BETWEEN $4::date AND $5::date
  AND ($6::varchar IS NULL OR status.code = $6)
ORDER BY payment.payment_date DESC, payment.payment_id, allocation.allocation_no
LIMIT $7 OFFSET $8;

-- 5.7 Credits and invoice adjustments.
-- $1 account ID, $2 owner ID or NULL, $3 billing-account ID or NULL,
-- $4 from date, $5 until date, $6 entry type ('CREDIT_NOTE',
-- 'INVOICE_ADJUSTMENT', or NULL), $7 limit, $8 offset
WITH entry AS (
    SELECT invoice.billing_account_id, credit.credit_date AS entry_date,
           'CREDIT_NOTE'::varchar AS entry_type, credit.credit_note_id AS entry_id,
           invoice.billing_invoice_id, reason.code AS reason_code,
           credit.net_amount AS net_amount, credit.tax_amount,
           credit.gross_amount, status.code AS status_code,
           credit.notes AS description
    FROM billing_credit_note credit
    JOIN billing_invoice invoice ON invoice.billing_invoice_id = credit.billing_invoice_id
    JOIN reason_code reason ON reason.reason_code_id = credit.reason_code_id
    JOIN document_status status ON status.status_id = credit.status_id
    WHERE credit.credit_date BETWEEN $4::date AND $5::date
    UNION ALL
    SELECT invoice.billing_account_id, invoice.invoice_date,
           'INVOICE_ADJUSTMENT'::varchar, adjustment.invoice_adjustment_id,
           invoice.billing_invoice_id, reason.code,
           kind.amount_effect * adjustment.amount,
           kind.amount_effect * adjustment.tax_amount,
           kind.amount_effect * (adjustment.amount + adjustment.tax_amount),
           invoice_status.code, adjustment.description
    FROM billing_invoice_adjustment adjustment
    JOIN billing_invoice invoice
      ON invoice.billing_invoice_id = adjustment.billing_invoice_id
    JOIN billing_adjustment_type kind
      ON kind.billing_adjustment_type_id = adjustment.billing_adjustment_type_id
    JOIN reason_code reason ON reason.reason_code_id = adjustment.reason_code_id
    JOIN document_status invoice_status ON invoice_status.status_id = invoice.status_id
    WHERE invoice.invoice_date BETWEEN $4::date AND $5::date
)
SELECT entry.entry_date, entry.entry_type, entry.entry_id,
       entry.billing_invoice_id, owner.code AS owner_code,
       account.code AS billing_account_code, entry.reason_code,
       entry.net_amount, entry.tax_amount, entry.gross_amount,
       currency.code AS currency_code, entry.status_code, entry.description
FROM entry
JOIN billing_account account ON account.billing_account_id = entry.billing_account_id
JOIN organization owner ON owner.organization_id = account.owner_id
JOIN billing_invoice invoice ON invoice.billing_invoice_id = entry.billing_invoice_id
JOIN currency ON currency.currency_id = invoice.currency_id
WHERE EXISTS (
    SELECT 1 FROM report_account_permission permission
    WHERE permission.account_id = $1 AND permission.permission_code = 'REPORT.BILLING'
)
  AND EXISTS (
    SELECT 1 FROM report_owner_scope scope
    WHERE scope.account_id = $1 AND scope.owner_id = account.owner_id
  )
  AND ($2::uuid IS NULL OR account.owner_id = $2)
  AND ($3::uuid IS NULL OR entry.billing_account_id = $3)
  AND ($6::varchar IS NULL OR entry.entry_type = $6)
ORDER BY entry.entry_date DESC, entry.entry_id
LIMIT $7 OFFSET $8;
