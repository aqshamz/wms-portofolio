# WMS report query catalog

Apply `00_reporting_setup.sql` after the main schema. Apply the bootstrap file
to create the Reporting module, five report permissions, and the Reports menu.

Files:

1. `01_master_reports.sql` - organizations, warehouses, locations, partners,
   items, packaging, and workflow configuration.
2. `02_inbound_reports.sql` - PO fulfillment, receipts, QC, quarantine, and
   putaway performance.
3. `03_stock_control_reports.sql` - on-hand stock, movements, internal moves,
   transfers, adjustments, and count variance.
4. `04_outbound_reports.sql` - DO fulfillment, allocation/picking, waves,
   packing/shipping, and store delivery.
5. `05_billing_reports.sql` - contract/rate coverage, events, charges, billing
   runs, invoices, credits, payments, and aging.

Each numbered query is independent. `$1` is always the requesting application
account ID. Every query checks its report permission and owner/warehouse scope
inside SQL. Optional filters use typed nullable parameters; the backend should
still cap export ranges and page large detail reports.

These are operational reports over authoritative OLTP tables. No duplicate
report tables are created. For high-volume historical analytics, replicate the
immutable facts into a separate reporting database rather than adding heavy
materialized aggregates to the transaction database.
