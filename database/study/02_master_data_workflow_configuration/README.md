# Study Guide 02: Master Data and Workflow Configuration

This guide explains the WMS organization, warehouse, partner, item, inventory
classification, workflow, task, strategy, and document-number configuration.
Inventory identity (`inventory_lot`, `serial_number`, and `handling_unit`) is
intentionally reserved for Study Guide 03.

## 1. The four independent authorization questions

Before studying the master tables, keep these questions separate:

| Question | Tables that answer it |
|---|---|
| Who is the user? | `app_account`, `app_session` |
| What may the user do? | `account_role`, `app_role`, `role_permission`, `app_permission` |
| For which inventory owner/client? | `account_owner_access` |
| At which warehouse? | `account_warehouse_access` |

A role does not automatically provide owner or warehouse scope. An inbound
operator needs all three forms of authorization:

```text
functional permission: INBOUND.RECEIVE
owner scope:           CLIENT_FRESH
warehouse scope:       WH_JKT_01
```

For an operation with both `owner_id` and `warehouse_id`, the combination
should also exist in `warehouse_owner`.

## 2. Organization, owner, and warehouse concepts

### 2.1 `organization`

`organization` is a general company/legal-entity master. It is not always the
inventory owner.

An organization may play one of these contextual roles:

- 3PL warehouse operator
- client/inventory owner
- both operator and inventory owner

The role is determined by where the organization is referenced:

| Reference | Meaning of the organization |
|---|---|
| `warehouse.operator_id` | Company operating the warehouse |
| `warehouse_owner.owner_id` | Client allowed to keep inventory in that warehouse |
| `item.owner_id` | Client that owns the item master |
| Transaction `owner_id` | Client that owns the affected inventory/document |

Example organizations:

| Code | Name | Contextual role |
|---|---|---|
| `3PL_ID` | Nusantara Logistics | Warehouse operator |
| `CLIENT_FRESH` | Fresh Foods Indonesia | Inventory owner/client |
| `CLIENT_RETAIL` | Retail Makmur | Inventory owner/client |

### 2.2 What makes an account an "owner account"?

There is no `is_owner_account` parameter or account type in the current
schema. An account receives owner scope when this row exists:

```text
account_owner_access(account_id, owner_id)
```

Example:

```text
account_id = account for client.fresh.user
owner_id   = organization CLIENT_FRESH
```

That account can be limited to `CLIENT_FRESH` records. A 3PL supervisor can
receive several `account_owner_access` rows and work across several clients.

This table describes data scope; it does not describe employment or legal
membership. If the application later needs to record which organization
employs an account, add a separate organization-membership concept rather than
overloading `account_owner_access`.

### 2.3 `warehouse`

One row represents a warehouse operated by an organization.

Example:

```text
code        = WH_JKT_01
name        = Jakarta Distribution Centre
operator_id = organization 3PL_ID
timezone    = Asia/Jakarta
```

`operator_id` identifies who operates the facility. It is not the client whose
stock is stored there.

### 2.4 What gives an account warehouse access?

Warehouse scope exists when this row exists:

```text
account_warehouse_access(account_id, warehouse_id)
```

For example, an operator may have:

```text
role/permission:  INBOUND_OPERATOR -> INBOUND.RECEIVE
owner scope:      CLIENT_FRESH
warehouse scope:  WH_JKT_01
```

When authorizing a protected request, the backend supplies/checks:

```text
session token
required permission code
requested owner_id
requested warehouse_id
```

The account is allowed only if the session is valid, the role provides the
permission, and both scope rows exist. `granted_by` records who granted the
scope; it does not grant authority by itself.

### 2.5 `warehouse_owner`

Despite its name, `warehouse_owner` does not mean the legal owner of the
warehouse building. It means:

```text
the inventory owner/client is allowed to keep stock at this warehouse
```

Example:

| Warehouse | Owner/client | Meaning |
|---|---|---|
| `WH_JKT_01` | `CLIENT_FRESH` | Fresh Foods may store inventory in Jakarta DC. |
| `WH_JKT_01` | `CLIENT_RETAIL` | Retail Makmur may also store inventory there. |

The physical/operator relationship remains:

```text
warehouse.operator_id -> organization
```

Many inventory and transaction tables reference the composite
`(owner_id, warehouse_id)` relationship, preventing stock from being created
for a client that is not configured at that warehouse.

## 3. Warehouse layout

### 3.1 Difference between zone, location type, and location

| Concept | Meaning | Example |
|---|---|---|
| `warehouse_zone` | A named physical/operational area inside one warehouse | `DRY_STORAGE_A`, `FROZEN`, `OUTBOUND` |
| `location_type` | Reusable behavior/capability of a location | `STORAGE`, `PICK_FACE`, `SHIPPING` |
| `warehouse_location` | The actual addressable bin, dock, lane, workstation, or floor position | `A-01-02-03`, `RCV-DOCK-01` |

A zone belongs to one warehouse. A location belongs to one zone and has one
location type.

### 3.2 Starter `location_type` data

| Code | Receive | Store | Pick | Ship | Typical purpose |
|---|:---:|:---:|:---:|:---:|---|
| `RECEIVING` | Yes | No | No | No | Receiving dock/inbound staging |
| `STORAGE` | No | Yes | No | No | Reserve storage |
| `PICK_FACE` | No | Yes | Yes | No | Forward picking bin |
| `STAGING` | No | Yes | No | No | Outbound stock staged by order |
| `CHECKING` | No | Yes | No | No | Outbound verification area |
| `PACKING` | No | No | No | No | Packing workstation |
| `SHIPPING` | No | No | No | Yes | Dispatch lane/loading dock |
| `QUARANTINE` | No | Yes | No | No | Stock awaiting client decision |

The capability flags allow SQL to validate a destination without hardcoding a
particular location ID.

### 3.3 Example zones

Zones are warehouse-specific and do not have behavior flags.

| Code | Name | Purpose |
|---|---|---|
| `INBOUND` | Inbound Dock Area | Receiving and initial staging |
| `DRY_RESERVE` | Dry Reserve Storage | General reserve inventory |
| `CHILLER` | Chiller Storage | Temperature-controlled products |
| `FROZEN` | Frozen Storage | Frozen products |
| `PICKING_A` | Forward Picking A | Fast-moving pick faces |
| `QUARANTINE` | Quarantine Area | Rejected/held stock |
| `OUTBOUND` | Outbound Processing | Staging, checking, packing, shipping |

It is valid for a zone and a location type to have similar names. The zone says
where an area is; the type says what a particular location is allowed to do.

### 3.4 Example locations

| Location code | Zone | Location type | Example details |
|---|---|---|---|
| `RCV-DOCK-01` | `INBOUND` | `RECEIVING` | Receiving dock 1 |
| `DRY-A-01-01-01` | `DRY_RESERVE` | `STORAGE` | Aisle A, bay 01, level 01, position 01 |
| `PICK-A-01-01` | `PICKING_A` | `PICK_FACE` | Forward pick bin |
| `QTN-01` | `QUARANTINE` | `QUARANTINE` | Quarantine holding location |
| `STG-DO-01` | `OUTBOUND` | `STAGING` | Delivery-order staging lane |
| `CHK-01` | `OUTBOUND` | `CHECKING` | Checking station |
| `PCK-01` | `OUTBOUND` | `PACKING` | Packing workstation |
| `SHP-LANE-01` | `OUTBOUND` | `SHIPPING` | Dispatch lane |

`max_weight` and `max_volume` describe capacity. `is_pick_face` is a faster
location-level marker for forward pick bins, while the location type provides
the general picking capability.

## 4. Business partners

### 4.1 `partner_type`

Your interpretation is correct. Starter types include:

| Code | Meaning |
|---|---|
| `SUPPLIER` | Vendor supplying goods |
| `FACTORY` | Production house/manufacturer |
| `CUSTOMER` | Outbound customer |
| `STORE` | Retail/branch destination |
| `CARRIER` | Transportation provider |
| `OTHER` | Other relationship |

### 4.2 `business_partner`

This stores the partner's actual code, name, legal/tax details, contacts, and
address. It is scoped by `owner_id` because different clients may use the same
partner code for different companies.

Example:

```text
owner       = CLIENT_FRESH
partner code= VEND-001
name        = Bandung Vegetable Supplier
```

### 4.3 `business_partner_type`

This bridge assigns one or more types to a partner. A factory can also be a
supplier, and a customer can also be a store.

```text
Bandung Vegetable Supplier
  -> SUPPLIER
  -> FACTORY
```

This many-to-many model avoids a hardcoded single partner classification.

## 5. Item and UOM masters

### 5.1 `item_category`

Your examples are reasonable. Categories are owned by a client and can form a
hierarchy.

Example:

```text
FOOD
  -> FRESH
       -> VEGETABLE
       -> FRUIT
  -> FROZEN_FOOD
  -> FAST_FOOD
```

Categories should describe the client's useful storage, reporting, billing, or
handling classification. Avoid categories that have no business use.

### 5.2 `item`

An item belongs to an owner and may belong to one category.

Important controls include:

| Column | Purpose |
|---|---|
| `base_uom_id` | Canonical quantity unit used for normalized stock quantities |
| `weight`, `volume` | Measurement for one base unit under the deployment's measurement convention |
| `lot_controlled` | Requires lot identity during receiving/inventory handling |
| `serial_controlled` | Requires individual serial identities |
| `shelf_life_days` | Expected product shelf life |
| `minimum_receive_days` | Minimum remaining shelf life accepted at receiving |

### 5.3 `uom` and `item_uom`

`uom` contains globally reusable quantity units such as:

```text
EA, BOX, CTN, PLT, KG, G, L, M
```

`item_uom` defines which UOMs are valid for a specific item and how they convert
to its base UOM.

Example item: bottled juice, base UOM `EA`.

| UOM | Conversion to base | Receiving | Picking | Meaning |
|---|---:|:---:|:---:|---|
| `EA` | 1 | Yes | Yes | One bottle |
| `BOX` | 12 | Yes | Yes | One box = 12 bottles |
| `CTN` | 48 | Yes | No | One carton = 48 bottles |
| `PLT` | 1,920 | Yes | No | One pallet = 1,920 bottles |

The conversion formula is:

```text
base quantity = source quantity * conversion_to_base
```

Therefore receiving and picking can be configured with different allowed UOMs
using `is_receiving_uom` and `is_picking_uom`.

However, the current operational implementation is only partially complete:

- Receiving converts source UOM quantities into base quantities.
- The receiving query currently checks that the item UOM is active but does
  not yet require `is_receiving_uom = true`.
- The starter outbound validation currently requires the base UOM, so alternate
  case/box picking is not yet fully implemented even though
  `is_picking_uom` exists.

Before backend development, decide whether outbound operates only in base UOM
or supports alternate picking UOMs. If alternate UOM is required, allocation,
pick, check, pack, and ship queries must consistently convert and validate it.

### 5.4 `item_barcode`

A barcode identifies an item and may identify a particular item UOM.

Example:

| Barcode | UOM | Meaning |
|---|---|---|
| `8990000000011` | `EA` | Single bottle barcode |
| `18990000000018` | `BOX` | Box barcode |
| `28990000000015` | `CTN` | Carton barcode |

One active barcode may be marked primary for the item. UOM-specific barcodes
help receiving and picking infer both the item and scanned packaging level.

### 5.5 Measurement-unit decision

The current columns `item.weight`, `item.volume`, `item_uom.length`, `width`,
`height`, `weight`, location capacity, and handling-unit capacity do not store
their measurement UOMs.

This requires one of two explicit decisions:

1. Establish database-wide units such as kilograms, cubic metres, and
   centimetres in configuration/documentation; or
2. Add explicit weight, dimension, and volume UOM references.

Because measurements affect capacity and billing, relying on an undocumented
unit convention would be risky. This should be finalized before backend and
billing implementation.

## 6. Inventory and quality classifications

### 6.1 `inventory_status`

Inventory status describes the operational usability of a stock balance.

| Code | Allocatable | Pickable | Purpose |
|---|:---:|:---:|---|
| `QC_PENDING` | No | No | Received stock waiting for QC |
| `AVAILABLE` | Yes | Yes | Normal stock available for outbound use |
| `HOLD` | No | No | Temporarily held stock |
| `QUARANTINE` | No | No | Stock awaiting a client decision |
| `DAMAGED` | No | No | Damaged stock |
| `EXPIRED` | No | No | Expired stock |

This status is stored on `inventory_balance` and controls whether allocation
and picking may use the balance. Expiry belongs here as an operational stock
state, not in `quality_status`.

### 6.2 `quality_status`

Quality status describes the lifecycle/state of an inspection or the recorded
quality state of a lot.

| Code | Meaning |
|---|---|
| `PENDING` | Inspection has not finished |
| `PASSED` | Requirements were satisfied |
| `FAILED` | Requirements were not satisfied |
| `WAIVED` | Authorized person waived inspection |

It is used by `quality_inspection.quality_status_id` and can also be recorded on
`inventory_lot.quality_status_id`.

`EXPIRED` should not normally be a quality workflow status. Lot expiry is
determined from `expiry_date`, after which stock can be changed to the
`EXPIRED` inventory status.

### 6.3 `inspection_result`

Inspection result records the quantitative outcome after inspection finishes:

| Code | `is_accepted` | Meaning |
|---|:---:|---|
| `ACCEPTED` | Yes | All inspected quantity accepted |
| `PARTIAL` | Yes | Some quantity accepted and some rejected |
| `REJECTED` | No | All inspected quantity rejected |

The distinction is:

```text
quality_status    = state of the QC process
inspection_result = outcome of the completed QC
```

For example, a new inspection can be `PENDING` with no result. A completed
inspection can be `PASSED/ACCEPTED`, `FAILED/REJECTED`, or a partial result
with separate `passed_qty` and `failed_qty`.

### 6.4 `quarantine_disposition_type`

This table configures the client's decision for quarantined stock.

| Code | Effect |
|---|---|
| `ACCEPT` | Release the quantity to `AVAILABLE` and create putaway work |
| `REWORK` | Require rework and later reinspection |
| `RETURN` | Remove stock through return-to-vendor movement |
| `DISPOSE` | Remove stock through disposal movement |

Exactly one behavior flag must be true:

```text
releases_to_available
requires_reinspection
removes_inventory
```

For removal decisions, `removal_movement_type_id` identifies the inventory
movement type used for the audit ledger.

### 6.5 `reason_code`

A reason code records why an exceptional or controlled action happened. It is
grouped by `module_code` and can require a user note.

Examples:

| Module | Code | Use |
|---|---|---|
| `INVENTORY` | `DAMAGE` | Damaged stock/status change |
| `INVENTORY` | `COUNT_VARIANCE` | Physical count differs from system stock |
| `INVENTORY` | `STATUS_HOLD` | Stock placed on hold |
| `INBOUND` | `OVER_RECEIPT` | Received more than expected |
| `OUTBOUND` | `SHORT_PICK` | Picker could not fulfill planned quantity |
| `GENERAL` | `CANCELLATION` | Document/task cancellation |
| `BILLING` | `BILLING_ERROR` | Billing correction |

Reason codes are referenced from movements, short picks, internal moves,
status changes, adjustments, stock counts, billable-event exclusions, invoice
adjustments, and credit notes.

## 7. Document workflow configuration

### 7.1 `document_type`

`document_type` identifies the business kind of a document/transaction. It is
not limited to purchase orders and delivery orders.

Examples:

| Document type | Module code |
|---|---|
| `PURCHASE_ORDER` | `INBOUND` |
| `RECEIPT` | `INBOUND` |
| `QUALITY_INSPECTION` | `INBOUND` |
| `OUTBOUND` | `OUTBOUND` |
| `SHIPMENT` | `OUTBOUND` |
| `TRANSFER` | `INVENTORY` |
| `STOCK_COUNT` | `INVENTORY` |
| `INVOICE` | `BILLING` |

`module_code` must be the owning application module such as `INBOUND`,
`OUTBOUND`, `INVENTORY`, or `BILLING`.

### 7.2 `document_status`

Statuses belong to a specific document type. `DRAFT` for a purchase order and
`DRAFT` for an outbound order are separate rows.

Examples:

```text
PURCHASE_ORDER: DRAFT -> APPROVED -> PARTIALLY_RECEIVED -> RECEIVED
OUTBOUND:       DRAFT -> VALIDATED -> RELEASED -> ALLOCATED -> ... -> DELIVERED
INVOICE:        DRAFT -> REVIEWED -> ISSUED -> PARTIALLY_PAID -> PAID
```

Useful flags are:

- `is_initial`: starting status for a new document
- `is_final`: normal or abnormal terminal state
- `is_cancelled`: terminal cancellation/void state
- `display_order`: presentation order

`CREATED`, `APPROVED`, and `REJECTED` can be valid statuses, but they should be
configured only for document types whose workflow actually uses them. The
starter model commonly uses `DRAFT`, `OPEN`, `PENDING`, or `PLANNED` as the
initial state instead of a universal `CREATED` status.

### 7.3 `document_status_transition`

This table defines allowed workflow edges and prevents arbitrary jumps.

Example:

```text
DRAFT -> APPROVED       allowed
DRAFT -> CANCELLED      allowed
DRAFT -> RECEIVED       not configured, therefore not allowed
```

`required_permission_id` can state that a transition needs a specific
permission, such as approval. The starter transition rows currently leave this
column null, so permission requirements must either be populated here or
enforced separately by backend services.

The transition table is configuration. It does not automatically update a
transaction row; the update query/service must verify the transition while
changing the document status.

## 8. Document IDs and daily counters

### 8.1 `document_number_rule`

This table configures how varchar transaction IDs are formatted for every
configured document type, including purchase orders, inbound orders, receipts,
outbound orders, movements, transfers, tasks, shipments, invoices, and others.

Configuration includes:

```text
prefix
separator
sequence length
include partner code?
include warehouse code?
effective dates
```

Example inbound rule:

```text
document type     = INBOUND
prefix            = INB
separator         = -
sequence length   = 6
include partner   = true
include warehouse = true
```

Generated result:

```text
INB-VEND01-WHJKT-20260819-000001
```

### 8.2 `document_daily_counter`

This table stores the last number used for each document type and business
date:

```text
PRIMARY KEY (document_type_id, business_date)
```

Therefore each document type restarts at 1 each day:

```text
INBOUND on 2026-08-19: 000001, 000002, 000003
INBOUND on 2026-08-20: 000001, 000002, 000003
RECEIPT on 2026-08-20: 000001, 000002, ...
```

The sequence is global per document type/day. Although partner and warehouse
codes can appear in the formatted ID, they are deliberately not part of the
counter key. Consequently:

```text
INB-VEND-A-WH1-20260819-000001
INB-VEND-B-WH2-20260819-000002
```

This is not an auto-increment primary key. `last_number` is only an atomic
counter used to build the varchar business ID. The resulting varchar is then
stored as the transaction primary key. PostgreSQL's upsert locks the one
counter row safely when concurrent requests request the next number.

## 9. Picking strategies

### 9.1 Strategy versus sort method

`picking_strategy` is a named policy applicable globally, to an owner, to a
warehouse, or to an owner/warehouse combination.

`picking_sort_method` is the algorithm used by one rule. Starter methods are:

| Code | Meaning |
|---|---|
| `FEFO` | Earliest expiry date first |
| `FIFO` | Oldest received stock first |
| `LOCATION` | Warehouse location sequence |
| `LOT` | Lot-number sequence |

`LIFO` is not part of the starter data, but it can be added if a real business
case requires newest stock first. It is generally unsuitable for expiring food
because it leaves old inventory behind.

### 9.2 `picking_strategy_rule`

Rules can filter by inventory status and zone, then specify the sort method.
`sequence_no` determines rule order.

Current default example:

```text
strategy:         DEFAULT_FEFO
owner/warehouse:  null/null (global fallback)
inventory status: AVAILABLE
zone:             any
sort method:      FEFO
```

A more specific example could be:

```text
CLIENT_FRESH + WH_JKT_01
  rule 10: AVAILABLE stock in CHILLER, sorted FEFO
  rule 20: AVAILABLE stock in DRY_RESERVE, sorted FIFO
```

Selection precedence between global and owner/warehouse strategies belongs in
the backend strategy-selection service and must be implemented consistently.

## 10. Putaway strategies

`putaway_strategy` is a named policy for selecting eligible destination
locations. Like picking strategy, null owner/warehouse values represent less
specific defaults.

`putaway_strategy_rule` can use:

- `sequence_no`: evaluation order
- `category_id`: item-category condition
- `location_type_id`: required destination type
- `zone_id`: preferred/required warehouse zone
- `minimum_empty_percent`: required remaining capacity percentage

Example strategy:

```text
Strategy: CLIENT_FRESH_JKT_PUTAWAY
Owner:    CLIENT_FRESH
Warehouse: WH_JKT_01

Rule 10: FROZEN_FOOD -> FROZEN zone -> STORAGE type
Rule 20: VEGETABLE   -> CHILLER zone -> STORAGE type
Rule 30: any category -> DRY_RESERVE zone -> STORAGE type
```

The tables describe eligible destinations and priority. The actual algorithm
that calculates current free capacity and chooses the best bin belongs in the
putaway service/query. No default putaway strategy is currently inserted by
the bootstrap script, so one must be configured for the real warehouse layout.

## 11. Task configuration

### 11.1 `task_type`

This identifies the operational work being performed.

Starter values:

```text
PUTAWAY
PICK
REPLENISHMENT
STOCK_COUNT
REWORK
```

### 11.2 `task_status`

Task status is shared by all task types in the current model:

```text
OPEN
ASSIGNED
IN_PROGRESS
COMPLETED
CANCELLED
```

Flags identify the single initial state, final states, and cancellation state.

### 11.3 `task_status_transition`

This table defines allowed task-status changes:

```text
OPEN -> ASSIGNED
OPEN -> IN_PROGRESS
OPEN -> CANCELLED
ASSIGNED -> IN_PROGRESS
ASSIGNED -> CANCELLED
IN_PROGRESS -> COMPLETED
IN_PROGRESS -> CANCELLED
```

`required_permission_id` can protect a transition. Starter rows do not yet
assign required permissions.

Because `task_status_transition` has no `task_type_id`, every task type shares
the same transition graph. If different task types later require different
workflows, the model will need task-type-specific transitions.

### 11.4 `task_priority`

Priority is configurable data with a sortable numeric value:

| Code | Value |
|---|---:|
| `LOW` | 100 |
| `NORMAL` | 200 |
| `HIGH` | 300 |
| `URGENT` | 400 |

Operational queues sort `priority_value DESC`, so larger values appear first.
Using spaced numeric values makes it possible to insert a future level between
existing priorities without renumbering everything.

## 12. Corrected mental model

```text
organization
  = a company/legal entity

warehouse.operator_id
  = company operating the facility

warehouse_owner
  = clients allowed to keep inventory in the facility

account_role / role_permission
  = what an account may do

account_owner_access
  = whose inventory/data an account may access

account_warehouse_access
  = where an account may operate

business_partner
  = supplier, factory, customer, store, or carrier belonging to a client scope

inventory_status
  = whether stock is operationally usable

quality_status
  = state of the quality process

inspection_result
  = completed inspection outcome

document_type + document_status + transition
  = configurable workflow definition

document_number_rule + daily_counter
  = concurrency-safe varchar business ID generation
```

## 13. Decisions to confirm before backend development

1. Does the application need explicit account-to-employer organization
   membership, separate from owner data scope?
2. Should the generic authorization check also verify the requested
   `(owner_id, warehouse_id)` exists in `warehouse_owner`?
3. What are the standard units for weight, dimensions, and volume?
4. Will outbound support alternate picking UOMs or require base UOM only?
5. Should `is_receiving_uom` and `is_picking_uom` be enforced in every relevant
   operational query?
6. Which document transitions require explicit permissions?
7. Do all task types share one transition graph?
8. What strategy precedence applies: owner+warehouse, owner-only,
   warehouse-only, then global?
9. Which putaway strategy and rules apply to the actual warehouse zones?

## 14. Source files

- [Master and workflow table DDL](../../wms_schema.sql)
- [Organization and warehouse CRUD](../../master/01_organization_warehouse_crud.sql)
- [Partner and item CRUD](../../master/02_partner_item_crud.sql)
- [Operational configuration CRUD](../../master/03_operational_config_crud.sql)
- [Starter reference data](../../master/00_bootstrap_reference_data.sql)
- [Master ER diagram](../../erd/domains/master.mmd)
- [Configuration ER diagram](../../erd/domains/configuration.mmd)

## 15. Suggested study checklist

- [ ] Explain operator organization versus inventory-owner organization.
- [ ] Trace permission, owner scope, and warehouse scope independently.
- [ ] Explain `warehouse_owner` without calling it the building owner.
- [ ] Design zones, location types, and actual locations for one warehouse.
- [ ] Create one partner with two partner types.
- [ ] Create one item with base, receiving, and picking UOM examples.
- [ ] Explain inventory status versus quality status versus inspection result.
- [ ] Trace one document through configured status transitions.
- [ ] Generate sample document IDs across two business dates.
- [ ] Design one putaway and one picking strategy.
- [ ] Resolve the nine backend design decisions above.

