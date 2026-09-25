# Inbound part 2: QC, putaway, and quarantine

Phase 2 starts from a completed receipt batch whose `initial_balance_id` points
to stock in `QC_PENDING`.

```text
QC_PENDING
   |
   +-- pass ----> PUTAWAY_PENDING --> putaway task --> AVAILABLE in storage
   |
   +-- fail ----> QUARANTINE ------> ACCEPT --> AVAILABLE in storage
                                      RETURN --> inventory removed (RETURN_TO_VENDOR)
                                      DISPOSE -> inventory removed (DISPOSE)
```

Every inventory action is written to the immutable movement ledger. Inspection
completion, its inventory splits, and creation of the resulting task/case occur
in one transaction. Putaway completion is atomic and safely retryable after the
task is final. Quarantine dispositions are atomic and protected by case and
balance versions, so a stale retry is rejected instead of being applied twice.

All quantities after receipt completion are canonical base-unit quantities.
The receipt batch's `base_qty` and `base_uom_id` become the inspection quantity;
passed stock becomes the putaway task quantity, failed stock becomes the
quarantine quantity, and rework/reinspection retain that same unit. Detail and
list responses expose `base_uom_code` so clients do not need to infer the unit
from the current Item-UOM configuration.

Quarantine disposition responses include nullable `target_location_code` for
readable history labels. `target_location_id` remains the internal identifier
used by requests. Decisions without a target location return a null code.

## Routes

```text
POST /api/v1/inbound/quality-inspections
GET  /api/v1/inbound/quality-inspections?owner_id=<uuid>&warehouse_id=<uuid>&status_code=PENDING&page=1&page_size=20
GET  /api/v1/inbound/quality-inspections/:id
POST /api/v1/inbound/quality-inspections/:id/complete
POST /api/v1/inbound/quality-inspections/:id/cancel

GET  /api/v1/inbound/putaway-tasks?owner_id=<uuid>&warehouse_id=<uuid>&status_code=OPEN&page=1&page_size=20
GET  /api/v1/inbound/putaway-tasks/:id
POST /api/v1/inbound/putaway-tasks/:id/start
POST /api/v1/inbound/putaway-tasks/:id/complete

GET  /api/v1/inbound/quarantine-disposition-types?active=true
GET  /api/v1/inbound/quarantine-cases?owner_id=<uuid>&warehouse_id=<uuid>&status_code=OPEN&page=1&page_size=20
GET  /api/v1/inbound/quarantine-cases/:id
GET  /api/v1/inbound/quarantine-cases/:id/targets?search=<code>&page=1&page_size=20
POST /api/v1/inbound/quarantine-cases/:id/dispositions
```

All routes require `Authorization: Bearer <token>`. JSON requests require
`Content-Type: application/json`. List endpoints require `owner_id`.

## Before testing

Complete the normal phase-1 example first. From its receipt response, keep:

- `receipt_inventory_id` from the accepted batch;
- `initial_balance_id` from that batch;
- the owner and warehouse UUIDs.

The target location must be active, unlocked, in the same warehouse, use a
location type with `allows_storage=true`, and match an active putaway-strategy
rule at the most specific configured owner/warehouse scope. The study master
already configures `STUDY_STORAGE`; use `STUDY_BULK_01` as the target.

Read the current source balance before every inventory-changing request:

For quality inspections, `GET /api/v1/inbound/quality-inspections/:id` (and
create/complete/cancel responses) includes `source_balance_version_no`,
`base_uom_code`, `is_indivisible`, and optional `serial_no` / `handling_unit_barcode`.
Use that source version as `expected_balance_version`, together with the
inspection's `version_no` as `expected_version`. This remains protected by
inbound owner/warehouse scope and does not require separate `INVENTORY.READ`
permission. Refresh after a concurrency conflict; never silently overwrite it.

For other inventory workflows, read the balance directly:

```text
GET /api/v1/inventory/balances/<balance_id>
```

Copy its current `version_no` into `expected_balance_version`.

## 1. Open a quality inspection

The start-inspection receipt picker uses
`GET /api/v1/inbound/receipts?owner_id=<uuid>&warehouse_id=<uuid>&inspection_eligible=true&page=1&page_size=10`.
This optional receipt-only boolean filter runs before counting and pagination.
It returns completed receipts with at least one posted batch whose full quantity
is still available in `QC_PENDING` and which has no existing inspection of any
status. A partly inspected receipt remains selectable until its last eligible
batch has an inspection. Pending, completed, and replacement inspections remain
in the quality-inspection list; normal receipt history is unchanged when the
filter is omitted or false. Normal owner/warehouse access scopes still apply.

```http
POST /api/v1/inbound/quality-inspections

{
  "receipt_inventory_id": "<receipt_inventory_id>",
  "notes": "Packaging and product quality inspection"
}
```

One root inspection is allowed per receipt batch. It inspects the full accepted
base quantity. The response starts with quality status `PENDING`, version `1`,
and no inventory movement.

## 2. Complete the inspection

All passed and failed quantities use the item's base UOM and must add up exactly
to `inspected_qty`.

### Fully accepted

```http
POST /api/v1/inbound/quality-inspections/<inspection_id>/complete

{
  "expected_version": 1,
  "expected_balance_version": <current_balance_version>,
  "passed_qty": "24",
  "failed_qty": "0",
  "putaway_target_location_id": "<STUDY_BULK_01_uuid>",
  "notes": "All samples passed"
}
```

This creates one `STATUS_CHANGE` movement from `QC_PENDING` to
`PUTAWAY_PENDING` and returns an embedded `putaway_task`.

### Partially accepted

```json
{
  "expected_version": 1,
  "expected_balance_version": 2,
  "passed_qty": "20",
  "failed_qty": "4",
  "putaway_target_location_id": "<STUDY_BULK_01_uuid>",
  "notes": "Four packs have damaged seals"
}
```

This creates both a putaway task for 20 and a quarantine case for 4. The
inspection result is `PARTIAL`.

### Fully rejected

```json
{
  "expected_version": 1,
  "expected_balance_version": 2,
  "passed_qty": "0",
  "failed_qty": "24",
  "notes": "Lot failed inspection"
}
```

No target location is supplied when nothing passed. The result is `REJECTED`,
and all stock moves to `QUARANTINE`.

A serialized batch or a handling-unit batch cannot be split between pass and
fail because one serial/HU cannot physically exist in two balances.

## 3. Execute the putaway task

The frontend is available at `/inbound/putaway` (Inbound → Putaway tasks).
Tasks are created by QC, not manually. An open task can be claimed by starting
it, or assigned to another scoped account with effective `INBOUND.PUTAWAY`
permission before it starts. An assigned task can only be started/completed by
its assignee. Completion is a
full-quantity physical move; partial completion is not supported.

`GET /api/v1/inbound/putaway-tasks/:id` and transition responses include
`source_balance_version_no` for completion/cancellation and, once completed,
`result_balance_version_no` for reversal. Use these together with the task's
`version_no`; no separate `INVENTORY.READ` permission is needed. Unit codes and
assigned display names are included in task list/detail responses. Refresh
after a concurrency conflict instead of silently adopting a newer version.

Start the returned task using its current version:

```http
POST /api/v1/inbound/putaway-tasks/<putaway_task_id>/start

{
  "expected_version": 1
}
```

The authenticated account becomes the assignee. Only that account can complete
the task. Reload the task and its `source_balance_id`, then complete it:

```http
POST /api/v1/inbound/putaway-tasks/<putaway_task_id>/complete

{
  "expected_version": 2,
  "expected_balance_version": <current_source_balance_version>,
  "business_date": "2026-09-08"
}
```

Completion creates one `PUTAWAY` movement that changes both location and status:
the stock leaves `PUTAWAY_PENDING` at QC and becomes `AVAILABLE` at the configured
storage target. The task returns `inventory_movement_id` and
`resulting_balance_id`.

For handling units, the operation also updates `handling_unit.current_location_id`.
Safety rules require the entire unreserved leaf HU balance to move; mixed,
partially reserved, or parent HUs are rejected.

## 4. Resolve quarantined stock

The frontend is available at `/inbound/quarantine`, with owner/warehouse-scoped
filters, responsive case lists, a detail dialog, and immutable disposition history.
`INBOUND.READ` allows viewing; `INBOUND.QUARANTINE_DISPOSE` allows decisions.
Cases are generated by failed inspections, not manually created or deleted.
QC details link directly to the resulting case using the `owner`, `warehouse`,
and `case` query parameters.

Case detail and disposition responses include `quarantine_balance_version_no`,
`available_qty`, `base_uom_code`, and `is_indivisible`, plus serial/handling-unit
labels when applicable. Decisions use this scoped snapshot rather than requiring
an unrelated `INVENTORY.READ` balance inquiry. Refresh after a concurrency error.
The frontend validates quantity against both case remaining quantity and
unreserved stock; serial/handling-unit stock cannot be split.

Discover available actions:

```text
GET /api/v1/inbound/quarantine-disposition-types?active=true
```

Supported phase-2 actions are `ACCEPT`, `RETURN`, and `DISPOSE`.

### Accept into storage

```http
POST /api/v1/inbound/quarantine-cases/<quarantine_case_id>/dispositions

{
  "expected_case_version": 1,
  "expected_balance_version": <current_quarantine_balance_version>,
  "disposition_type_code": "ACCEPT",
  "disposition_qty": "4",
  "business_date": "2026-09-08",
  "decided_at": "2026-09-08T10:00:00+07:00",
  "target_location_id": "<STUDY_BULK_01_uuid>",
  "client_decision_reference": "OWNER-DECISION-001",
  "decision_notes": "Owner accepts the cosmetic damage"
}
```

`ACCEPT` moves stock directly from `QUARANTINE` into `AVAILABLE` storage and
must include a valid putaway target.
It does not create a separate putaway task. The frontend requires confirmation
of the immediate stock move. Its target picker uses the case-specific lookup,
which applies the same active putaway strategy as the mutation validator.

### Return to vendor

```json
{
  "expected_case_version": 1,
  "expected_balance_version": 2,
  "disposition_type_code": "RETURN",
  "disposition_qty": "2",
  "business_date": "2026-09-08",
  "decided_at": "2026-09-08T11:00:00+07:00",
  "client_decision_reference": "OWNER-DECISION-002",
  "decision_notes": "Supplier approved return"
}
```

`RETURN` removes the quantity with a `RETURN_TO_VENDOR` movement.

### Dispose

Use the same request shape with `"disposition_type_code": "DISPOSE"`. It removes
the quantity with a `DISPOSE` movement. Neither removal action accepts a target
location.

A disposition may process part of a case. The case then becomes
`PARTIALLY_DECIDED`, increments its version, and retains the remaining quantity.
Reload the case and quarantine balance before the next disposition. Once total
processed quantity equals `quarantine_qty`, the case becomes `CLOSED`.
The frontend refreshes the detail snapshot and history after each decision.

`REWORK` now creates a task and child reinspection through the lifecycle API.
See [Inbound part 3](inbound-part-3-api.md) for the complete execution flow.
The Quarantine frontend requires work instructions for this action and links to
the resulting task at `/inbound/rework`. The dedicated Rework Tasks view supports
start, completion, result notes, and a link to reinspection when available.
A closed quarantine case may still have unfinished rework or reinspection.

## Verification

Use these inquiries after each action:

```text
GET /api/v1/inbound/quality-inspections/<inspection_id>
GET /api/v1/inbound/putaway-tasks/<putaway_task_id>
GET /api/v1/inbound/quarantine-cases/<quarantine_case_id>
GET /api/v1/inventory/movements?owner_id=<owner_uuid>&search=<document_id>&page=1&page_size=100
GET /api/v1/inventory/balances?owner_id=<owner_uuid>&warehouse_id=<warehouse_uuid>&item_id=<item_uuid>&include_zero=true&page=1&page_size=100
```

The automated PostgreSQL workflow test starts with a non-base receipt UOM and
verifies its converted base quantity through partial QC, putaway, quarantine,
rework, reinspection, and final available inventory. It also covers partial
return, accepted quarantine stock, final case closure, retry safety, and total
inventory conservation against both existing and fresh schemas.
