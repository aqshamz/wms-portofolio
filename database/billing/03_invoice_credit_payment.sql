-- Invoice, adjustment, credit-note, payment, and receivable queries.
-- Multi-statement operations must run on one connection and roll back on any error.

SET search_path TO wms, public;

-- =============================================================================
-- 1. INVOICE
-- =============================================================================

-- 1.1 Create a draft invoice from a reviewed run and snapshot every charge.
-- $1 billing-run ID, $2 invoice date, $3 notes, $4 actor ID
BEGIN;

WITH context AS (
    SELECT run.*, account.bill_to_name, account.bill_to_tax_number,
           account.bill_to_address_1, account.bill_to_address_2,
           account.bill_to_city, account.bill_to_province,
           account.bill_to_postal_code, account.bill_to_country_code,
           term.due_days,
           invoice_type.document_type_id AS invoice_document_type_id,
           initial_status.status_id AS invoice_status_id
    FROM billing_run run
    JOIN document_status run_status ON run_status.status_id = run.status_id
    JOIN billing_account account ON account.billing_account_id = run.billing_account_id
    JOIN payment_term term ON term.payment_term_id = account.payment_term_id
    JOIN document_type invoice_type
      ON invoice_type.code = 'INVOICE' AND invoice_type.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = invoice_type.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE run.billing_run_id = $1 AND run_status.code = 'REVIEWED'
      AND NOT EXISTS (
          SELECT 1 FROM billing_invoice existing
          WHERE existing.billing_run_id = run.billing_run_id
      )
    FOR UPDATE OF run
), numbered AS (
    SELECT context.*,
           generate_document_id('INVOICE', NULL, NULL, $2::date) AS generated_id
    FROM context
)
INSERT INTO billing_invoice (
    billing_invoice_id, document_type_id, status_id, billing_run_id,
    billing_account_id, billing_contract_id, currency_id,
    invoice_date, due_date, bill_to_name, bill_to_tax_number,
    bill_to_address, notes, created_by, updated_by
)
SELECT generated_id, invoice_document_type_id, invoice_status_id, billing_run_id,
       billing_account_id, billing_contract_id, currency_id,
       $2, $2::date + due_days, bill_to_name, bill_to_tax_number,
       concat_ws(E'\n', bill_to_address_1, bill_to_address_2,
                 concat_ws(', ', bill_to_city, bill_to_province, bill_to_postal_code),
                 bill_to_country_code),
       $3, $4, $4
FROM numbered;

WITH source_lines AS (
    SELECT invoice.billing_invoice_id, charge.*, event.billing_service_id,
           event.business_date, event.source_document_id,
           service.name AS service_name,
           row_number() OVER (
               ORDER BY event.business_date, service.code, charge.billing_charge_id
           )::integer AS line_no
    FROM billing_invoice invoice
    JOIN billing_charge charge ON charge.billing_run_id = invoice.billing_run_id
    JOIN billable_event event ON event.billable_event_id = charge.billable_event_id
    JOIN billing_service service ON service.billing_service_id = event.billing_service_id
    WHERE invoice.billing_run_id = $1
)
INSERT INTO billing_invoice_line (
    billing_invoice_line_id, billing_invoice_id, line_no,
    billing_charge_id, billing_service_id, description,
    quantity, uom_id, unit_rate, net_amount,
    tax_rule_id, tax_rate_percent, tax_amount, gross_amount
)
SELECT billing_invoice_id || '-L-' || lpad(line_no::text, 4, '0'),
       billing_invoice_id, line_no, billing_charge_id, billing_service_id,
       left(service_name || ' | ' || business_date::text || ' | ' || source_document_id, 255),
       billed_quantity, uom_id, unit_rate, net_amount,
       tax_rule_id, tax_rate_percent, tax_amount, gross_amount
FROM source_lines;

WITH totals AS (
    SELECT line.billing_invoice_id,
           sum(line.net_amount) AS net_amount,
           sum(line.tax_amount) AS tax_amount,
           sum(line.gross_amount) AS gross_amount
    FROM billing_invoice_line line
    JOIN billing_invoice invoice
      ON invoice.billing_invoice_id = line.billing_invoice_id
    WHERE invoice.billing_run_id = $1
    GROUP BY line.billing_invoice_id
)
UPDATE billing_invoice invoice
SET net_amount = totals.net_amount, adjustment_amount = 0,
    tax_amount = totals.tax_amount, gross_amount = totals.gross_amount,
    balance_due = totals.gross_amount,
    updated_at = clock_timestamp(), updated_by = $4
FROM totals
WHERE invoice.billing_invoice_id = totals.billing_invoice_id;

UPDATE billing_run run
SET status_id = invoiced.status_id
FROM document_status current_status, document_status invoiced
WHERE run.billing_run_id = $1
  AND current_status.status_id = run.status_id AND current_status.code = 'REVIEWED'
  AND invoiced.document_type_id = run.document_type_id AND invoiced.code = 'INVOICED';

SELECT invoice.*
FROM billing_invoice invoice
WHERE invoice.billing_run_id = $1;

COMMIT;

-- 1.2 Add a surcharge/discount to a draft invoice.
-- $1 invoice ID, $2 adjustment line number, $3 adjustment-type code,
-- $4 billing reason code, $5 description, $6 commercial amount (> 0),
-- $7 tax-rule ID or NULL, $8 actor ID
BEGIN;

WITH locked AS (
    SELECT invoice.*, currency.decimal_scale,
           adjustment_type.billing_adjustment_type_id,
           adjustment_type.amount_effect, reason.reason_code_id,
           tax.tax_rule_id, COALESCE(tax.rate_percent, 0) AS tax_rate_percent,
           COALESCE(tax.is_inclusive, false) AS is_inclusive
    FROM billing_invoice invoice
    JOIN document_status status ON status.status_id = invoice.status_id
    JOIN currency ON currency.currency_id = invoice.currency_id
    JOIN billing_adjustment_type adjustment_type
      ON adjustment_type.code = $3 AND adjustment_type.is_active
    JOIN reason_code reason
      ON reason.module_code = 'BILLING' AND reason.code = $4 AND reason.is_active
    LEFT JOIN tax_rule tax
      ON tax.tax_rule_id = $7 AND tax.is_active
     AND invoice.invoice_date BETWEEN tax.effective_from
         AND COALESCE(tax.effective_until, 'infinity'::date)
    WHERE invoice.billing_invoice_id = $1 AND status.code = 'DRAFT'
      AND ($7::uuid IS NULL OR tax.tax_rule_id IS NOT NULL)
    FOR UPDATE OF invoice
), calculated AS (
    SELECT locked.*,
           CASE WHEN is_inclusive AND tax_rate_percent > 0
                THEN round($6::numeric / (1 + tax_rate_percent / 100), decimal_scale)
                ELSE round($6::numeric, decimal_scale) END AS net_adjustment
    FROM locked
)
INSERT INTO billing_invoice_adjustment (
    invoice_adjustment_id, billing_invoice_id, line_no,
    billing_adjustment_type_id, reason_code_id, description,
    amount, tax_rule_id, tax_rate_percent, tax_amount, created_by
)
SELECT billing_invoice_id || '-ADJ-' || lpad($2::text, 3, '0'),
       billing_invoice_id, $2, billing_adjustment_type_id,
       reason_code_id, $5, net_adjustment, tax_rule_id,
       tax_rate_percent,
       CASE WHEN is_inclusive THEN round($6::numeric, decimal_scale) - net_adjustment
            ELSE round(net_adjustment * tax_rate_percent / 100, decimal_scale) END,
       $8
FROM calculated
WHERE $6::numeric > 0
RETURNING *;

WITH line_totals AS (
    SELECT line.billing_invoice_id,
           sum(line.net_amount) AS line_net, sum(line.tax_amount) AS line_tax
    FROM billing_invoice_line line
    WHERE line.billing_invoice_id = $1
    GROUP BY line.billing_invoice_id
), adjustment_totals AS (
    SELECT adjustment.billing_invoice_id,
           sum(adjustment_type.amount_effect * adjustment.amount) AS adjustment_net,
           sum(adjustment_type.amount_effect * adjustment.tax_amount) AS adjustment_tax
    FROM billing_invoice_adjustment adjustment
    JOIN billing_adjustment_type adjustment_type
      ON adjustment_type.billing_adjustment_type_id = adjustment.billing_adjustment_type_id
    WHERE adjustment.billing_invoice_id = $1
    GROUP BY adjustment.billing_invoice_id
)
UPDATE billing_invoice invoice
SET net_amount = line_totals.line_net,
    adjustment_amount = adjustment_totals.adjustment_net,
    tax_amount = line_totals.line_tax + adjustment_totals.adjustment_tax,
    gross_amount = line_totals.line_net + adjustment_totals.adjustment_net
                 + line_totals.line_tax + adjustment_totals.adjustment_tax,
    balance_due = line_totals.line_net + adjustment_totals.adjustment_net
                + line_totals.line_tax + adjustment_totals.adjustment_tax,
    updated_at = clock_timestamp(), updated_by = $8
FROM line_totals, adjustment_totals
WHERE invoice.billing_invoice_id = line_totals.billing_invoice_id
  AND invoice.billing_invoice_id = adjustment_totals.billing_invoice_id
RETURNING invoice.*;

COMMIT;

-- 1.3 Review a draft invoice after recomputing all snapshot totals.
-- $1 invoice ID, $2 reviewer ID
WITH line_totals AS (
    SELECT billing_invoice_id, sum(net_amount) AS line_net,
           sum(tax_amount) AS line_tax
    FROM billing_invoice_line WHERE billing_invoice_id = $1
    GROUP BY billing_invoice_id
), adjustment_totals AS (
    SELECT adjustment.billing_invoice_id,
           COALESCE(sum(kind.amount_effect * adjustment.amount), 0) AS adjustment_net,
           COALESCE(sum(kind.amount_effect * adjustment.tax_amount), 0) AS adjustment_tax
    FROM billing_invoice_adjustment adjustment
    JOIN billing_adjustment_type kind
      ON kind.billing_adjustment_type_id = adjustment.billing_adjustment_type_id
    WHERE adjustment.billing_invoice_id = $1
    GROUP BY adjustment.billing_invoice_id
)
UPDATE billing_invoice invoice
SET status_id = reviewed.status_id,
    net_amount = line_totals.line_net,
    adjustment_amount = COALESCE(adjustment_totals.adjustment_net, 0),
    tax_amount = line_totals.line_tax + COALESCE(adjustment_totals.adjustment_tax, 0),
    gross_amount = line_totals.line_net
                 + COALESCE(adjustment_totals.adjustment_net, 0)
                 + line_totals.line_tax
                 + COALESCE(adjustment_totals.adjustment_tax, 0),
    balance_due = line_totals.line_net
                + COALESCE(adjustment_totals.adjustment_net, 0)
                + line_totals.line_tax
                + COALESCE(adjustment_totals.adjustment_tax, 0),
    updated_at = clock_timestamp(), updated_by = $2
FROM line_totals
LEFT JOIN adjustment_totals
  ON adjustment_totals.billing_invoice_id = line_totals.billing_invoice_id,
     document_status current_status,
     document_status reviewed
WHERE invoice.billing_invoice_id = line_totals.billing_invoice_id
  AND current_status.status_id = invoice.status_id
  AND current_status.code = 'DRAFT'
  AND reviewed.document_type_id = invoice.document_type_id
  AND reviewed.code = 'REVIEWED'
  AND line_totals.line_net + COALESCE(adjustment_totals.adjustment_net, 0) >= 0
  AND line_totals.line_tax + COALESCE(adjustment_totals.adjustment_tax, 0) >= 0
RETURNING invoice.*;

-- 1.4 Issue a reviewed invoice. Snapshot values become immutable here.
-- $1 invoice ID, $2 actor ID
UPDATE billing_invoice invoice
SET status_id = issued.status_id,
    issued_at = clock_timestamp(), issued_by = $2,
    updated_at = clock_timestamp(), updated_by = $2
FROM document_status current_status, document_status issued
WHERE invoice.billing_invoice_id = $1
  AND current_status.status_id = invoice.status_id AND current_status.code = 'REVIEWED'
  AND issued.document_type_id = invoice.document_type_id AND issued.code = 'ISSUED'
  AND invoice.gross_amount > 0 AND invoice.balance_due = invoice.gross_amount
RETURNING invoice.*;

-- 1.5 Invoice detail. $1 invoice ID
SELECT invoice.*, status.code AS status_code, currency.code AS currency_code,
       account.code AS billing_account_code,
       COALESCE((
           SELECT jsonb_agg(to_jsonb(line) ORDER BY line.line_no)
           FROM billing_invoice_line line
           WHERE line.billing_invoice_id = invoice.billing_invoice_id
       ), '[]'::jsonb) AS lines,
       COALESCE((
           SELECT jsonb_agg(
               to_jsonb(adjustment) || jsonb_build_object(
                   'typeCode', kind.code, 'amountEffect', kind.amount_effect
               ) ORDER BY adjustment.line_no
           )
           FROM billing_invoice_adjustment adjustment
           JOIN billing_adjustment_type kind
             ON kind.billing_adjustment_type_id = adjustment.billing_adjustment_type_id
           WHERE adjustment.billing_invoice_id = invoice.billing_invoice_id
       ), '[]'::jsonb) AS adjustments
FROM billing_invoice invoice
JOIN document_status status ON status.status_id = invoice.status_id
JOIN currency ON currency.currency_id = invoice.currency_id
JOIN billing_account account ON account.billing_account_id = invoice.billing_account_id
WHERE invoice.billing_invoice_id = $1;

-- =============================================================================
-- 2. CREDIT NOTE
-- =============================================================================

-- 2.1 Create a draft credit note against an open issued invoice.
-- $1 invoice ID, $2 credit date, $3 billing reason code,
-- $4 required notes, $5 actor ID
WITH context AS (
    SELECT invoice.billing_invoice_id, invoice.currency_id,
           reason.reason_code_id, credit_type.document_type_id,
           initial_status.status_id
    FROM billing_invoice invoice
    JOIN document_status invoice_status ON invoice_status.status_id = invoice.status_id
    JOIN reason_code reason
      ON reason.module_code = 'BILLING' AND reason.code = $3 AND reason.is_active
    JOIN document_type credit_type
      ON credit_type.code = 'CREDIT_NOTE' AND credit_type.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = credit_type.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE invoice.billing_invoice_id = $1
      AND invoice_status.code IN ('ISSUED', 'PARTIALLY_PAID')
      AND invoice.balance_due > 0 AND nullif(btrim($4), '') IS NOT NULL
), numbered AS (
    SELECT context.*,
           generate_document_id('CREDIT_NOTE', NULL, NULL, $2::date) AS generated_id
    FROM context
)
INSERT INTO billing_credit_note (
    credit_note_id, document_type_id, status_id,
    billing_invoice_id, currency_id, credit_date,
    reason_code_id, notes, created_by
)
SELECT generated_id, document_type_id, status_id,
       billing_invoice_id, currency_id, $2, reason_code_id, $4, $5
FROM numbered
RETURNING *;

-- 2.2 Add a partial or full invoice-line credit.
-- $1 credit-note ID, $2 credit line number, $3 invoice-line ID,
-- $4 net amount to credit, $5 description
BEGIN;

WITH locked AS (
    SELECT credit.credit_note_id, credit.billing_invoice_id,
           invoice_line.billing_invoice_line_id,
           invoice_line.net_amount AS original_net,
           invoice_line.gross_amount AS original_gross,
           invoice_line.tax_rate_percent, currency.decimal_scale
    FROM billing_credit_note credit
    JOIN document_status credit_status ON credit_status.status_id = credit.status_id
    JOIN billing_invoice_line invoice_line
      ON invoice_line.billing_invoice_line_id = $3
     AND invoice_line.billing_invoice_id = credit.billing_invoice_id
    JOIN currency ON currency.currency_id = credit.currency_id
    WHERE credit.credit_note_id = $1 AND credit_status.code = 'DRAFT'
    FOR UPDATE OF credit, invoice_line
), calculated AS (
    SELECT locked.*, round($4::numeric, decimal_scale) AS credit_net,
           round($4::numeric * tax_rate_percent / 100, decimal_scale) AS credit_tax,
           COALESCE((
               SELECT sum(existing.gross_amount)
               FROM billing_credit_note_line existing
               JOIN billing_credit_note other
                 ON other.credit_note_id = existing.credit_note_id
               JOIN document_status other_status ON other_status.status_id = other.status_id
               WHERE existing.billing_invoice_line_id = locked.billing_invoice_line_id
                 AND other_status.code <> 'VOID'
           ), 0) AS already_credited
    FROM locked
)
INSERT INTO billing_credit_note_line (
    credit_note_line_id, credit_note_id, billing_invoice_line_id,
    line_no, description, net_amount, tax_rate_percent,
    tax_amount, gross_amount
)
SELECT credit_note_id || '-L-' || lpad($2::text, 3, '0'),
       credit_note_id, billing_invoice_line_id, $2, $5,
       credit_net, tax_rate_percent, credit_tax, credit_net + credit_tax
FROM calculated
WHERE credit_net > 0
  AND already_credited + credit_net + credit_tax <= original_gross;

WITH totals AS (
    SELECT line.credit_note_id, sum(line.net_amount) AS net_amount,
           sum(line.tax_amount) AS tax_amount,
           sum(line.gross_amount) AS gross_amount
    FROM billing_credit_note_line line
    WHERE line.credit_note_id = $1
    GROUP BY line.credit_note_id
)
UPDATE billing_credit_note credit
SET net_amount = totals.net_amount, tax_amount = totals.tax_amount,
    gross_amount = totals.gross_amount
FROM totals
WHERE credit.credit_note_id = totals.credit_note_id
RETURNING credit.*;

COMMIT;

-- 2.3 Issue a credit note and reduce the invoice open balance atomically.
-- $1 credit-note ID, $2 actor ID
BEGIN;

WITH locked AS (
    SELECT credit.credit_note_id, credit.document_type_id,
           credit.billing_invoice_id, credit.gross_amount,
           invoice.document_type_id AS invoice_document_type_id,
           invoice.balance_due
    FROM billing_credit_note credit
    JOIN document_status credit_status ON credit_status.status_id = credit.status_id
    JOIN billing_invoice invoice ON invoice.billing_invoice_id = credit.billing_invoice_id
    JOIN document_status invoice_status ON invoice_status.status_id = invoice.status_id
    WHERE credit.credit_note_id = $1 AND credit_status.code = 'DRAFT'
      AND invoice_status.code IN ('ISSUED', 'PARTIALLY_PAID')
      AND credit.gross_amount > 0 AND credit.gross_amount <= invoice.balance_due
    FOR UPDATE OF credit, invoice
), issued_credit AS (
    UPDATE billing_credit_note credit
    SET status_id = issued.status_id,
        issued_at = clock_timestamp(), issued_by = $2
    FROM locked, document_status issued
    WHERE credit.credit_note_id = locked.credit_note_id
      AND issued.document_type_id = credit.document_type_id AND issued.code = 'ISSUED'
    RETURNING credit.billing_invoice_id, credit.gross_amount
)
UPDATE billing_invoice invoice
SET credited_amount = invoice.credited_amount + issued_credit.gross_amount,
    balance_due = invoice.balance_due - issued_credit.gross_amount,
    status_id = CASE WHEN invoice.balance_due - issued_credit.gross_amount = 0
                     THEN settled.status_id ELSE invoice.status_id END,
    updated_at = clock_timestamp(), updated_by = $2
FROM issued_credit, document_status settled
WHERE invoice.billing_invoice_id = issued_credit.billing_invoice_id
  AND settled.document_type_id = invoice.document_type_id AND settled.code = 'SETTLED'
RETURNING invoice.*;

COMMIT;

-- =============================================================================
-- 3. PAYMENT AND RECEIVABLES
-- =============================================================================

-- 3.1 Record an unapplied payment.
-- $1 billing-account ID, $2 currency code, $3 payment-method code,
-- $4 payment date, $5 amount, $6 external reference or NULL,
-- $7 notes, $8 actor ID
WITH context AS (
    SELECT account.billing_account_id, currency.currency_id,
           method.payment_method_id, method.requires_reference,
           payment_type.document_type_id, initial_status.status_id
    FROM billing_account account
    JOIN currency ON currency.code = $2 AND currency.is_active
    JOIN payment_method method ON method.code = $3 AND method.is_active
    JOIN document_type payment_type
      ON payment_type.code = 'PAYMENT' AND payment_type.is_active
    JOIN document_status initial_status
      ON initial_status.document_type_id = payment_type.document_type_id
     AND initial_status.is_initial AND initial_status.is_active
    WHERE account.billing_account_id = $1 AND account.is_active
      AND account.currency_id = currency.currency_id
      AND (NOT method.requires_reference OR nullif(btrim($6), '') IS NOT NULL)
), numbered AS (
    SELECT context.*,
           generate_document_id('PAYMENT', NULL, NULL, $4::date) AS generated_id
    FROM context
)
INSERT INTO billing_payment (
    payment_id, document_type_id, status_id, billing_account_id,
    currency_id, payment_method_id, payment_date,
    amount, unapplied_amount, external_reference, notes, created_by
)
SELECT generated_id, document_type_id, status_id, billing_account_id,
       currency_id, payment_method_id, $4, $5, $5, $6, $7, $8
FROM numbered
RETURNING *;

-- 3.2 Allocate payment to one invoice. Repeat for additional invoices.
-- $1 payment ID, $2 invoice ID, $3 allocation sequence,
-- $4 allocation amount, $5 actor ID
BEGIN;

WITH locked AS (
    SELECT payment.payment_id, payment.document_type_id AS payment_document_type_id,
           payment.unapplied_amount, payment.billing_account_id,
           payment.currency_id, invoice.billing_invoice_id,
           invoice.document_type_id AS invoice_document_type_id,
           invoice.balance_due
    FROM billing_payment payment
    JOIN document_status payment_status ON payment_status.status_id = payment.status_id
    JOIN billing_invoice invoice ON invoice.billing_invoice_id = $2
    JOIN document_status invoice_status ON invoice_status.status_id = invoice.status_id
    WHERE payment.payment_id = $1
      AND payment_status.code IN ('RECEIVED', 'PARTIALLY_ALLOCATED')
      AND invoice_status.code IN ('ISSUED', 'PARTIALLY_PAID')
      AND payment.billing_account_id = invoice.billing_account_id
      AND payment.currency_id = invoice.currency_id
      AND $4::numeric > 0 AND $4::numeric <= payment.unapplied_amount
      AND $4::numeric <= invoice.balance_due
    FOR UPDATE OF payment, invoice
), allocated AS (
    INSERT INTO billing_payment_allocation (
        payment_allocation_id, payment_id, allocation_no,
        billing_invoice_id, allocated_amount, allocated_by
    )
    SELECT payment_id || '-A-' || lpad($3::text, 3, '0'),
           payment_id, $3, billing_invoice_id, $4, $5
    FROM locked
    ON CONFLICT (payment_id, billing_invoice_id) DO UPDATE
    SET allocated_amount = billing_payment_allocation.allocated_amount
                         + EXCLUDED.allocated_amount,
        allocated_at = clock_timestamp(), allocated_by = EXCLUDED.allocated_by
    RETURNING payment_id, billing_invoice_id
), payment_updated AS (
    UPDATE billing_payment payment
    SET unapplied_amount = payment.unapplied_amount - $4::numeric,
        status_id = CASE WHEN payment.unapplied_amount - $4::numeric = 0
                         THEN fully_allocated.status_id ELSE partly_allocated.status_id END
    FROM locked, allocated,
         document_status partly_allocated, document_status fully_allocated
    WHERE payment.payment_id = locked.payment_id
      AND allocated.payment_id = payment.payment_id
      AND partly_allocated.document_type_id = payment.document_type_id
      AND partly_allocated.code = 'PARTIALLY_ALLOCATED'
      AND fully_allocated.document_type_id = payment.document_type_id
      AND fully_allocated.code = 'ALLOCATED'
    RETURNING payment.payment_id
)
UPDATE billing_invoice invoice
SET paid_amount = invoice.paid_amount + $4::numeric,
    balance_due = invoice.balance_due - $4::numeric,
    status_id = CASE WHEN invoice.balance_due - $4::numeric = 0
                     THEN paid.status_id ELSE partly_paid.status_id END,
    updated_at = clock_timestamp(), updated_by = $5
FROM locked, payment_updated,
     document_status partly_paid, document_status paid
WHERE invoice.billing_invoice_id = locked.billing_invoice_id
  AND payment_updated.payment_id = locked.payment_id
  AND partly_paid.document_type_id = invoice.document_type_id
  AND partly_paid.code = 'PARTIALLY_PAID'
  AND paid.document_type_id = invoice.document_type_id AND paid.code = 'PAID'
RETURNING invoice.*;

COMMIT;

-- 3.3 Void only a completely unapplied payment. $1 payment ID
UPDATE billing_payment payment
SET status_id = voided.status_id
FROM document_status current_status, document_status voided
WHERE payment.payment_id = $1
  AND current_status.status_id = payment.status_id AND current_status.code = 'RECEIVED'
  AND payment.unapplied_amount = payment.amount
  AND NOT EXISTS (
      SELECT 1 FROM billing_payment_allocation allocation
      WHERE allocation.payment_id = payment.payment_id
  )
  AND voided.document_type_id = payment.document_type_id AND voided.code = 'VOID'
RETURNING payment.*;

-- 3.4 Open receivables / aging source query.
-- $1 owner ID, $2 as-of date, $3 billing-account ID or NULL
SELECT invoice.billing_invoice_id, account.code AS billing_account_code,
       invoice.invoice_date, invoice.due_date, currency.code AS currency_code,
       invoice.gross_amount, invoice.credited_amount, invoice.paid_amount,
       invoice.balance_due,
       GREATEST($2::date - invoice.due_date, 0) AS days_overdue,
       status.code AS status_code
FROM billing_invoice invoice
JOIN billing_account account ON account.billing_account_id = invoice.billing_account_id
JOIN currency ON currency.currency_id = invoice.currency_id
JOIN document_status status ON status.status_id = invoice.status_id
WHERE account.owner_id = $1
  AND ($3::uuid IS NULL OR account.billing_account_id = $3)
  AND invoice.invoice_date <= $2::date
  AND invoice.balance_due > 0
  AND status.code IN ('ISSUED', 'PARTIALLY_PAID')
ORDER BY invoice.due_date, invoice.billing_invoice_id;

-- 3.5 Payment and allocation detail. $1 payment ID
SELECT payment.payment_id, payment.payment_date,
       payment.amount, payment.unapplied_amount,
       payment_status.code AS payment_status, method.code AS payment_method,
       currency.code AS currency_code, payment.external_reference,
       allocation.allocation_no, allocation.allocated_amount,
       allocation.allocated_at, invoice.billing_invoice_id,
       invoice.balance_due AS invoice_balance_due
FROM billing_payment payment
JOIN document_status payment_status ON payment_status.status_id = payment.status_id
JOIN payment_method method ON method.payment_method_id = payment.payment_method_id
JOIN currency ON currency.currency_id = payment.currency_id
LEFT JOIN billing_payment_allocation allocation ON allocation.payment_id = payment.payment_id
LEFT JOIN billing_invoice invoice
  ON invoice.billing_invoice_id = allocation.billing_invoice_id
WHERE payment.payment_id = $1
ORDER BY allocation.allocation_no;
