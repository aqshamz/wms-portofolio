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

## Routes

```text
POST /api/v1/inbound/quality-inspections
GET  /api/v1/inbound/quality-inspections?owner_id=<uuid>&warehouse_id=<uuid>&status_code=PENDING&page=1&page_size=20
GET  /api/v1/inbound/quality-inspections/:id
POST /api/v1/inbound/quality-inspections/:id/complete

GET  /api/v1/inbound/putaway-tasks?owner_id=<uuid>&warehouse_id=<uuid>&status_code=OPEN&page=1&page_size=20
GET  /api/v1/inbound/putaway-tasks/:id
POST /api/v1/inbound/putaway-tasks/:id/start
POST /api/v1/inbound/putaway-tasks/:id/complete

GET  /api/v1/inbound/quarantine-disposition-types?active=true
GET  /api/v1/inbound/quarantine-cases?owner_id=<uuid>&warehouse_id=<uuid>&status_code=OPEN&page=1&page_size=20
GET  /api/v1/inbound/quarantine-cases/:id
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

```text
GET /api/v1/inventory/balances/<balance_id>
```

Copy its current `version_no` into `expected_balance_version`.

## 1. Open a quality inspection

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

`REWORK` now creates a task and child reinspection through the lifecycle API.
See [Inbound part 3](inbound-part-3-api.md) for the complete execution flow.

## Verification

Use these inquiries after each action:

```text
GET /api/v1/inbound/quality-inspections/<inspection_id>
GET /api/v1/inbound/putaway-tasks/<putaway_task_id>
GET /api/v1/inbound/quarantine-cases/<quarantine_case_id>
GET /api/v1/inventory/movements?owner_id=<owner_uuid>&search=<document_id>&page=1&page_size=100
GET /api/v1/inventory/balances?owner_id=<owner_uuid>&warehouse_id=<warehouse_uuid>&item_id=<item_uuid>&include_zero=true&page=1&page_size=100
```

The automated PostgreSQL workflow test covers partial QC, putaway, partial
return, accepted quarantine stock, final case closure, retry safety, and total
inventory conservation against both existing and fresh schemas.
