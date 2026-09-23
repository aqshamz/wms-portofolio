# Inbound part 3: lifecycle, exceptions, reversal, and rework

This phase completes the operational inbound lifecycle around the normal flow
documented in parts 1 and 2. All routes require `Authorization: Bearer <token>`.
JSON routes also require `Content-Type: application/json`.

## Added routes

```text
PUT  /api/v1/inbound/purchase-orders/:id
POST /api/v1/inbound/purchase-orders/:id/lines
PUT  /api/v1/inbound/purchase-orders/:id/lines/:line_id
DELETE /api/v1/inbound/purchase-orders/:id/lines/:line_id
PUT  /api/v1/inbound/orders/:id
POST /api/v1/inbound/orders/:id/lines
PUT  /api/v1/inbound/orders/:id/lines/:line_id
DELETE /api/v1/inbound/orders/:id/lines/:line_id
PUT  /api/v1/inbound/receipts/:id
POST /api/v1/inbound/purchase-orders/:id/close
POST /api/v1/inbound/purchase-orders/:id/cancel
POST /api/v1/inbound/orders/:id/close
POST /api/v1/inbound/orders/:id/cancel
POST /api/v1/inbound/receipts/:id/cancel
POST /api/v1/inbound/receipts/:id/reverse
POST /api/v1/inbound/quality-inspections/:id/cancel
POST /api/v1/inbound/putaway-tasks/:id/assign
POST /api/v1/inbound/putaway-tasks/:id/retarget
POST /api/v1/inbound/putaway-tasks/:id/cancel
POST /api/v1/inbound/putaway-tasks/:id/reverse
GET  /api/v1/inbound/exceptions
GET  /api/v1/inbound/exceptions/:id
GET  /api/v1/inbound/quarantine-cases/:id/targets
GET  /api/v1/inbound/rework-tasks
GET  /api/v1/inbound/rework-tasks/:id
POST /api/v1/inbound/rework-tasks/:id/start
POST /api/v1/inbound/rework-tasks/:id/complete
```

There are 53 inbound routes in total; operational resource routes are
authenticated and owner/warehouse-scoped.
The quarantine disposition-type lookup is global reference data and only
requires authentication and `INBOUND.READ`, but no selected owner/warehouse.

The quarantine target lookup requires `INBOUND.QUARANTINE_DISPOSE`. It accepts
only `search`, `page`, and `page_size`, deriving owner/warehouse from the case.
It returns active, unlocked storage-capable locations matching the item's
highest-specificity active putaway strategy. Eligibility is filtered before
counting and pagination, with the same response shape as putaway target lookup.
It does not require a general master-data or inventory read permission.

## Receiving tolerances

Tolerance is captured on each PO line so later master-data changes cannot alter
an already approved order:

```json
{
  "item_id": "<item_uuid>",
  "ordered_qty": "10",
  "uom_id": "<uom_uuid>",
  "over_receipt_tolerance_pct": "10",
  "under_receipt_tolerance_pct": "5"
}
```

Both values default to zero and must be between 0 and 100. In this example the
API permits a maximum cumulative receipt of 11 units. A receipt that exceeds
10 but does not exceed 11 must include line-level `exception_notes`:

```json
{
  "inbound_line_id": "<inbound_line_id>",
  "received_qty": "7",
  "rejected_qty": "0",
  "exception_notes": "Supplier shipped one approved extra unit",
  "batches": [
    {
      "source_qty": "7",
      "received_location_id": "<qc_location_uuid>",
      "lot": {
        "lot_number": "OVER-LOT-001",
        "expiry_date": "2027-09-08"
      }
    }
  ]
}
```

Multiple receipts against one released inbound order are supported. The order
remains `PARTIALLY_RECEIVED` until its cumulative received quantity reaches its
expectation, including an allowed overage.

Rejected-at-dock quantities also require `exception_notes`. Optionally set
`exception_type_code` to `DAMAGED`, `WRONG_ITEM`, or `REJECTED_AT_DOCK`; the
last value is the default. Completing the receipt creates the classified
rejection and/or `OVER_RECEIPT` audit rows.

## Closing short and cancellation

The close and cancel endpoints use the same request:

```json
{
  "expected_version": 3,
  "reason": "Supplier confirmed the remaining quantity will not ship"
}
```

- `orders/:id/close` changes a released or partially received inbound order to
  `CLOSED`. Open receipts must be completed or cancelled first.
- `purchase-orders/:id/close` changes an approved or partially received PO to
  `CLOSED`. Every related inbound order must already be final.
- A shortage outside the PO line's under-receipt tolerance creates an
  `UNDER_RECEIPT` exception. A shortage within tolerance is accepted without a
  variance exception.
- PO cancellation is allowed only before receiving and when no active inbound
  order exists.
- Inbound cancellation can abandon its unreceived remainder, but no receipt may
  still be `OPEN`.
- Receipt cancellation is allowed only while the receipt is `OPEN`, before any
  inventory movement is posted.

Every cancellation creates a `CANCELLATION` exception record.

Cancelled records cannot be edited or reopened. A corrected document is
created normally with `supersedes_purchase_order_id`, `supersedes_inbound_id`,
or `supersedes_receipt_id`. The API validates matching business scope and
source relationships, limits a cancelled document to one direct successor,
and returns predecessor and successor IDs for audit navigation.

## Receipt reversal

Receipt reversal is intentionally narrow. It is allowed only while every
accepted batch is still untouched in `QC_PENDING` and no quality inspection has
ever been opened for the batch. This prevents reversal from bypassing QC,
quarantine, reservations, or putaway history.

Read every distinct affected balance first, then send its current version:

```http
POST /api/v1/inbound/receipts/<receipt_id>/reverse

{
  "expected_version": 2,
  "business_date": "2026-09-08",
  "reason": "Duplicate supplier delivery",
  "balances": [
    {
      "balance_id": "<initial_balance_id>",
      "expected_version": 2
    }
  ]
}
```

The request contains one entry per distinct balance, not one entry per batch.
A fully rejected receipt has no inventory balance and may send an empty array.
The API posts `RECEIPT_REVERSAL`, changes the receipt to `REVERSED`, recalculates
its inbound order and PO progress, and records a `REVERSAL` exception.

## QC cancellation

```http
POST /api/v1/inbound/quality-inspections/<inspection_id>/cancel

{
  "expected_version": 1,
  "reason": "Wrong sampling plan"
}
```

Only a pending inspection with intact `QC_PENDING` stock can be cancelled. The
old inspection becomes `WAIVED`; the response contains
`replacement_inspection_id` for a new child inspection over the same stock.

## Putaway assignment and retargeting

The task-scoped, paginated pickers are:

```text
GET /api/v1/inbound/putaway-tasks/:id/assignees?search=<name>&page=1&page_size=20
GET /api/v1/inbound/putaway-tasks/:id/targets?search=<location-or-zone>&page=1&page_size=20
```

Both require `INBOUND.ASSIGN` and the task's owner/warehouse access scope.
They accept only `search`, `page`, and `page_size`; clients cannot override
owner/warehouse scope. Assignees expose only account ID, username, and display
name, limited to login-enabled, unlocked accounts with both access grants and
effective `INBOUND.PUTAWAY` permission (an active direct grant or an active role
grant for an active permission). Filtering occurs before counting/pagination.
Assignment rechecks the same eligibility; ineligible accounts are rejected even
when submitted directly to the API. Active, unlocked Superadmins with effective
`*` permission are also eligible without individual access-scope grants.
An unassigned OPEN task can be claimed by a
scoped worker with `INBOUND.PUTAWAY`; assigned tasks can only be started and
completed by their assignee.

Targets are active, unlocked storage-capable locations matching category,
location-type, and zone rules at the most specific active putaway-strategy
scope. Filtering happens before counting/pagination. The posting/retarget
validator uses the same strategy selection, and always rechecks at execution.
These pickers do not require unrelated security/master-read permissions.

Assign an account with both owner/warehouse access and `INBOUND.PUTAWAY`:

```json
{
  "expected_version": 1,
  "account_id": "<account_uuid>"
}
```

```http
POST /api/v1/inbound/putaway-tasks/<task_id>/assign
```

An assigned task may be started only by its assignee. Retarget an `OPEN` or
`ASSIGNED` task to another location accepted by the active putaway strategy:

```http
POST /api/v1/inbound/putaway-tasks/<task_id>/retarget

{
  "expected_version": 2,
  "target_location_id": "<storage_location_uuid>"
}
```

### Cancel before completion

```http
POST /api/v1/inbound/putaway-tasks/<task_id>/cancel

{
  "expected_version": 1,
  "expected_balance_version": 2,
  "business_date": "2026-09-08",
  "reason": "Putaway plan must be rebuilt"
}
```

Cancellation moves the exact task quantity from `PUTAWAY_PENDING` back to
`QC_PENDING`, marks the task `CANCELLED`, and creates a replacement child
inspection. An in-progress task may be cancelled only by its assignee.
Cancellation also persists and returns `replacement_inspection_id`, so the
frontend can open that inspection even after a task refresh.

### Reverse after completion

Use the same request shape with the completed task's current version and the
current version of its `resulting_balance_id`:

```http
POST /api/v1/inbound/putaway-tasks/<task_id>/reverse
```

The result quantity must still be fully unreserved and `AVAILABLE` at the
recorded target. The API posts `PUTAWAY_REVERSAL`, returns it to `QC_PENDING`,
marks the task `REVERSED`, and returns `replacement_inspection_id`. Handling
units must still be an intact leaf HU containing one unreserved balance.

## Rework and reinspection

Rework task responses include nullable `assigned_username` for display in the
list and detail views. `assigned_to` remains the account ID used for assignment
and authorization. An unassigned task has no username; no extra account lookup
permission is required to display the assigned username.

The frontend execution view is `/inbound/rework`. `INBOUND.READ` allows scoped
lists and details; `INBOUND.REWORK` allows execution. Filters use `owner`,
`warehouse`, `status`, `search`, and `page`; the `task` parameter opens a detail
dialog. Quarantine history links directly to the corresponding task.

The view supports the existing start and complete endpoints only. OPEN tasks
are claimed when started; ASSIGNED tasks can be started only by their assignee,
and IN_PROGRESS tasks can be completed only by that worker. Completion requires
confirmation of the full planned base-unit quantity, allows optional result
notes up to 4000 characters, and links to the generated child inspection.
There is no partial completion, manual create, assignment, cancellation, or
delete UI because the rework API does not expose those operations. Rework does
not itself make stock available. Both actions send the current task version;
on a concurrency error, refresh the task before retrying.

Use `REWORK` on an open quarantine case:

```http
POST /api/v1/inbound/quarantine-cases/<case_id>/dispositions

{
  "expected_case_version": 1,
  "expected_balance_version": 2,
  "disposition_type_code": "REWORK",
  "disposition_qty": "2",
  "business_date": "2026-09-08",
  "decided_at": "2026-09-08T15:00:00+07:00",
  "client_decision_reference": "OWNER-REWORK-001",
  "decision_notes": "Owner approved rework",
  "work_instructions": "Replace seals and clean the outer packaging"
}
```

This moves the quantity from `QUARANTINE` to `QC_PENDING`, processes the
disposition, and creates an embedded `rework_task`. Execute it with:

```http
POST /api/v1/inbound/rework-tasks/<task_id>/start
{"expected_version": 1}
```

```http
POST /api/v1/inbound/rework-tasks/<task_id>/complete
{"expected_version": 2, "result_notes": "Seal replaced"}
```

Completion creates `reinspection_id`. Complete that child quality inspection
using the normal phase-2 QC endpoint. A failed reinspection opens a child
quarantine case linked through `parent_quarantine_case_id`; it may be accepted,
returned, disposed, or reworked again without losing lineage.

## Exception inquiry

Exception list and detail responses include nullable `created_by_display_name`,
`owner_name`, and `warehouse_name` for readable UI labels, without requiring
additional account or master-data lookups. These are current reference names;
the original `created_by`, `owner_id`, and `warehouse_id` remain unchanged as
audit identifiers. Owner/warehouse access checks and the immutable records are
unchanged. If a reference name is unavailable, the UI shows an unavailable label
instead of exposing an internal UUID.

```text
GET /api/v1/inbound/exceptions?owner_id=<uuid>&warehouse_id=<uuid>&status_code=UNDER_RECEIPT&page=1&page_size=100
GET /api/v1/inbound/exceptions/<exception_id>
```

For this inquiry, `status_code` filters `exception_type_code`. Supported values
are `OVER_RECEIPT`, `UNDER_RECEIPT`, `REJECTED_AT_DOCK`, `DAMAGED`,
`WRONG_ITEM`, `CANCELLATION`, and `REVERSAL`.

The frontend inquiry is available at `/inbound/exceptions` and requires
`INBOUND.READ`. It includes account-scoped warehouse/served-owner selection,
exception-type filtering, source-document/line/notes search, pagination, and
responsive desktop tables/mobile cards. The `exception` query parameter opens
the detail dialog; `owner`, `warehouse`, `type`, `search`, and `page` preserve
list filters in the URL. Switching scope clears the selected exception.

Exception records are immutable audit evidence, not tasks with a resolution
status. No create, edit, delete, or resolve action is exposed. Details preserve
the recorded expected, actual, and signed variance quantities without assuming
base units; document-level cancellation/reversal quantities may be absent.
Reasons/notes, source references, timestamp, and recording account ID are shown.
Follow-up actions belong to the original document's workflow.

All mutation endpoints use optimistic document and/or balance versions. On a
`409`, reload the document and affected balances before deciding whether the
business action should be retried.
