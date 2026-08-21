# Billing SQL catalog

The billing flow is:

`billing account -> contract -> rate card -> billable event -> billing run -> charge -> invoice -> credit/payment`

Files:

1. `01_account_contract_rate_card.sql` - billing-account, tax-rule, contract,
   rate-card, service-rate, and tier maintenance.
2. `02_events_rating_run.sql` - idempotent operational event capture, storage
   snapshots, exclusions, billing runs, rating, tax snapshots, and review.
3. `03_invoice_credit_payment.sql` - invoice generation, adjustments, review,
   issue, credit notes, payments, allocation, and outstanding-balance inquiry.

No tax percentage, price, billing cycle, currency, payment term, charge basis,
or rounding rule is embedded in transactional queries. Those values come from
masters and are copied into charge/invoice snapshots so later master changes do
not rewrite financial history. Operational movement-to-service relationships
are configured in `billing_service_movement_type`.

All monetary values use `numeric`; never use PostgreSQL floating-point types for
money. Transaction IDs are varchar business IDs generated with daily counters.
Child IDs derive from their parent and an explicit line/allocation sequence.

Every operation containing `BEGIN` must use one database connection and roll
back if the expected final row is not returned. Issued invoices and credit notes
are not edited or deleted; corrections use credit notes and new billing events.

The starter data intentionally contains no tax rate and no service price. A tax
rule and rate card must be configured and approved for the applicable client,
warehouse, service, UOM, and effective period before rating succeeds.

Unit-storage and pallet-storage capture queries are end-of-day snapshots. Run
them once after operational posting for that business date; their unique event
keys make retries idempotent.
