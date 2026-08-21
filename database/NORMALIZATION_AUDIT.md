# WMS normalization audit (1NF through 5NF)

## Decision

The operational schema is already predominantly in 3NF/BCNF, and its genuine
many-valued relationships are already decomposed through 4NF and 5NF. A blanket
"convert every table to 5NF" rewrite is not appropriate for this OLTP system.
It would remove useful integrity guards and immutable snapshots while adding
joins to almost every transaction.

The normalization target is therefore:

- 3NF/BCNF for mutable master and configuration data;
- 4NF/5NF for independent many-to-many relationships;
- explicit, documented snapshots for immutable warehouse and financial facts;
- propagated tenant keys where they provide declarative scope enforcement.

## Changes made by this audit

`app_module` is now a master relation. `app_permission`, `reason_code`,
`document_type`, and `billing_service` reference its stable code. This removes
an unchecked repeated determinant and prevents misspelled module values.

The bootstrap and CRUD catalogs were updated to validate active modules before
creating dependent records.

## Normal-form assessment

### First normal form (1NF)

Pass. Repeating business sets are not stored in comma-separated columns or SQL
arrays. Lines, serials, barcodes, roles, permissions, events, tiers, and
allocations have their own rows.

`billable_event.attributes` is the intentional exception in spirit, though
PostgreSQL treats `jsonb` as one typed value. It stores an immutable source
payload that is not used as a relational price, quantity, status, or key. Any
attribute that becomes searchable or controls billing must be promoted to a
typed column/master relation instead of remaining JSON.

### Second normal form (2NF)

Pass. Relations with composite keys describe the whole relationship, including:

- `account_role`, `role_permission`, and `menu_permission`;
- `warehouse_owner`, account owner/warehouse access, and partner types;
- `item_uom`, strategy rules, workflow transitions, and shipment packing;
- billing service/movement mappings and payment allocations.

Relationship metadata such as validity periods or assignment audit fields
depends on the complete relationship key.

### Third normal form and BCNF

Mutable masters are separated from transactions: organizations, partners,
items, UOMs, statuses, reasons, services, currencies, terms, tax rules, and
rate cards are referenced by keys rather than copied as editable descriptions.

The following repeated values are deliberate and must not be removed merely to
claim BCNF:

- `owner_id` and `warehouse_id` propagated into stock/transaction facts enforce
  tenant scope with composite foreign keys and support partition/index access;
- `document_type_id` beside `status_id` enables a composite foreign key that
  prevents assigning a status from another workflow;
- bill-to, ship-to, charge, rate, and tax values copied into issued documents
  are historical snapshots, not mutable master attributes;
- `inventory_balance` is the current-state projection while
  `inventory_movement` is the immutable ledger; they have different purposes;
- invoice/run totals are controlled summaries used for locking, reconciliation,
  and financial constraints, not alternative sources of master truth.

Removing those columns would weaken database-enforced integrity or change past
business documents when masters are edited.

### Fourth normal form (4NF)

Pass for the current requirements. Independent multivalued facts are separate:
an account can have many roles/scopes, a partner many types, an item many UOMs
and barcodes, a document many lines/events, and a payment many allocations.

### Fifth normal form (5NF)

Pass for the declared business rules. Remaining three-or-more-key relations are
facts or ordered rules whose meaning depends on the complete combination. For
example, a picking strategy rule and a scoped rate-card line cannot be projected
into independent pair tables and joined back without generating invalid
combinations. Decomposing them would create spurious rows.

## Why queries do not become shorter

Normalization narrows stored facts but normally lengthens reads because the
facts must be joined. The long WMS commands are mainly caused by workflow
validation, concurrency locks, stock invariants, idempotency, and audit posting,
not by wide tables.

For a short backend interface, keep the normalized schema and expose each
atomic workflow as a PostgreSQL function/procedure or a backend repository
method. For example, a multi-CTE invoice operation can become one call such as
`select * from create_billing_invoice(...)` without weakening the underlying
model.

## Future decompositions only when requirements expand

Add these relations only if the business requires the stated cardinality:

- address and entity-address-role tables for multiple effective-dated addresses;
- partner contact and contact-role tables for multiple contacts;
- invoice tax-component tables for multiple simultaneous tax jurisdictions;
- exchange-rate and currency-conversion facts for cross-currency settlement;
- typed billable-event attribute tables for attributes used in filtering/rating.

Creating them before those requirements exist would add empty joins, not
normalization value.
