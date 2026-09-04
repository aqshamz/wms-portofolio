# Operational master configuration

This completes the requested **configuration** slice: document workflows,
document numbering/daily counters, picking strategies, putaway strategies, and
task configuration. It does not execute warehouse tasks, move stock, select
inventory, enforce workflow transitions on transaction documents, or implement
role/permission authorization.

## Start and authenticate

Restart the API from `C:\WmsProject\backend`:

```powershell
go run .
```

Use `http://localhost:8080`, or your configured port. Startup migrates these
tables and inserts operational reference data. No transaction documents, daily
counters, or example warehouse tasks are created by seeding.

Log in using `POST /api/v1/auth/login` and your existing
`{"identifier":"<username>","password":"<password>"}`. Copy `data.token`.

Every endpoint below requires:

```http
Authorization: Bearer <token>
Content-Type: application/json
```

All paths are under `/api/v1/master`. As in the existing master API, access is
currently restricted by **session authentication only**. Owner/warehouse
consistency checks validate data relationships, not the caller's authorization.

## Resources and routes

Each normal collection supports `POST` create, `GET` paginated list,
`GET /:id` detail, `PUT /:id` update, and `PATCH /:id/deactivate`.

| Collection | Purpose |
| --- | --- |
| `/modules` | Modules referenced by document types and permission metadata |
| `/document-types` | Document types |
| `/task-types` | Task types |
| `/task-statuses` | Shared task statuses |
| `/task-transitions` | Allowed task-status transitions |
| `/task-priorities` | Priorities; larger values have higher priority |
| `/picking-sort-methods` | Picking sort-method configuration |
| `/picking-strategies` | Scoped picking strategy headers |
| `/putaway-strategies` | Scoped putaway strategy headers |

The same five operations apply to nested collections:

- `/document-types/:id/statuses`
- `/document-types/:id/transitions`
- `/picking-strategies/:id/rules`
- `/putaway-strategies/:id/rules`

For nested detail/update/deactivation, append the child record's ID.
A child ID belonging to a different parent returns `404`.

`GET /workflow-permissions` and `GET /workflow-permissions/:id` provide
read-only permission lookup. Permission administration and role grants remain
part of the authentication/authorization work. New workflows may leave
`required_permission_id` null.

Numbering uses dedicated endpoints:

| Method | Path | Behavior |
| --- | --- | --- |
| GET | `/document-types/:id/number-rules` | Paginated rule history |
| POST | `/document-types/:id/number-rules` | Replace the active rule atomically |
| PATCH | `/document-types/:id/number-rules/:rule_id/deactivate` | Disable allocation for this rule |
| POST | `/document-types/:id/document-ids` | Allocate a new document ID |
| GET | `/document-types/:id/daily-counters` | Read counter usage within a date range |

There are **no counter create/update/delete/reset endpoints**.

Lists accept `page` (default 1, maximum 1000000), `page_size` (default 20,
maximum 100), and `active`. Code/name resources also accept `search`.
Document-type and permission lists accept `module_code`. Strategy lists accept
`owner_id` and `warehouse_id`; these are exact filters, not runtime default
selection. Omit the filters to see global and scoped strategies together.

Responses retain `success/message/data`. Paginated data includes `items`,
`page`, `page_size`, `total_items`, and `total_pages`.

## Document workflow configuration

Read `GET /modules`, then create a type or use one of the seeded types:

```http
POST /api/v1/master/document-types

{
  "code": "CUSTOM_INBOUND",
  "name": "Custom inbound",
  "module_code": "INBOUND",
  "description": "Inbound workflow configuration"
}
```

Create statuses under the returned `document_type_id`:

```http
POST /api/v1/master/document-types/<document_type_id>/statuses

{
  "code": "DRAFT",
  "name": "Draft",
  "description": null,
  "is_initial": true,
  "is_final": false,
  "is_cancelled": false,
  "display_order": 10
}
```

Create another status such as `RELEASED`, with `is_initial: false`, then link them:

```http
POST /api/v1/master/document-types/<document_type_id>/transitions

{
  "from_status_id": "<DRAFT status_id>",
  "to_status_id": "<RELEASED status_id>",
  "required_permission_id": null
}
```

Rules:

- Both statuses must belong to this document type.
- Active transitions require active statuses and a non-final source.
- Self-transitions are rejected. Workflow loops such as rework are allowed.
- A supplied permission must exist and be active. Saving its ID does not itself
  implement permission enforcement.
- At most one initial status exists per document type. Selecting a new initial
  status clears the old one atomically; a failed save restores the old selection.
- An initial status must be active and non-terminal. Cancelled means final.
- Deactivating a status clears its initial flag. It does not select a replacement
  or cascade into transition records.
- Deactivate outgoing transitions before marking their source status final.
- Statuses are listed by display order. Types and status codes remain stable.

Update a transition using its ID; endpoints are immutable:

```http
PUT /api/v1/master/document-types/<document_type_id>/transitions/<transition_id>

{"required_permission_id":null,"is_active":true}
```

## Document IDs and daily counters

Create/replace the active rule:

```http
POST /api/v1/master/document-types/<document_type_id>/number-rules

{
  "prefix": "CIN",
  "separator": "-",
  "sequence_length": 6,
  "include_partner_code": true,
  "include_warehouse_code": true,
  "effective_from": "<YYYY-MM-DD>"
}
```

`effective_from` must not be in the future or precede the latest rule's start
date. Replacement activates immediately: future scheduling is not supported
by this one-active-rule design. The date comparison uses `DB_TIMEZONE`.

Rules are versioned, not overwritten. Old rules are marked inactive and their
end date is closed. Same-day replacement is allowed; the retired version keeps
that day as its end date. Only the active rule is used for allocation, so
historical versions cannot be selected to back-allocate an ID. Replacement
never resets an existing counter or changes an existing document ID.

Allocate an ID:

```http
POST /api/v1/master/document-types/<document_type_id>/document-ids

{
  "business_date": "<YYYY-MM-DD>",
  "partner_id": "<partner_id>",
  "warehouse_id": "<warehouse_id>"
}
```

The IDs reference actual master data. A required partner/warehouse must be
active; when both are supplied, the partner's owner must have an active
assignment to that warehouse. Codes are resolved by the API, not accepted as
arbitrary caller text.

For prefix `CIN`, partner `SUP001`, warehouse `WH01`, and business date
2026-09-03, the first ID is:

```text
CIN-SUP001-WH01-20260903-000001
```

The response includes `document_id`, `document_type_id`,
`document_number_rule_id`, `business_date`, and `sequence_number`.

Important behavior:

- The counter key is **(document type, business date)**. Different owners,
  partners, and warehouses share the same sequence for that type/date.
- A new business date starts at 1. The caller supplies the date explicitly.
- Sequence width is 3–18 digits. Exhaustion returns `409` without incrementing.
- `sequence_number` and counter `last_number` are JSON **strings**, preserving
  integers beyond JavaScript's safe-integer limit.
- Prefix is 1–20 normalized code characters. Separator may be empty or up to
  three characters from `-`, `_`, `.`, and `/`.
- The generated ID must fit the existing 120-character transaction ID columns.
- **Every successful POST allocates another number.** It is not a preview,
  does not create a transaction document, and is not idempotent. Avoid blind
  retries after an uncertain network response; unused numbers can create gaps.
- Transaction services should use transaction-bound repositories and
  `GenerateDocumentID` inside the document-creation transaction. Failure of the
  outer transaction then rolls back both the number and the document.
- Rule edits and allocation are serialized by the document-type row lock;
  the counter increment itself is a single atomic PostgreSQL upsert.
- This API allocates through the Go service/repository. It does not replace
  the SQL `generate_document_id` function in the reference schema.

Read usage without consuming a number:

```http
GET /api/v1/master/document-types/<document_type_id>/daily-counters?date_from=2026-09-01&date_to=2026-09-30&page=1&page_size=20
```

Both dates are required, in `YYYY-MM-DD` format. No history is rewritten when
rules change.

## Picking strategies

Create a header. Omit or set `owner_id/warehouse_id` to null for broader scopes:

```http
POST /api/v1/master/picking-strategies

{
  "owner_id": "<organization_id>",
  "warehouse_id": "<warehouse_id>",
  "code": "CLIENT_FEFO",
  "name": "Client FEFO",
  "description": "Configured picking rules"
}
```

When both scope IDs are set, the owner must be assigned to the warehouse.
Codes are unique within the owner/warehouse scope, including null/global scopes.
Owner/warehouse scope is immutable after creation.

Read `GET /picking-sort-methods`, then add ordered rules:

```http
POST /api/v1/master/picking-strategies/<picking_strategy_id>/rules

{
  "sequence_no": 10,
  "inventory_status_id": "<AVAILABLE inventory_status_id>",
  "zone_id": "<zone_id>",
  "picking_sort_method_id": "<FEFO picking_sort_method_id>"
}
```

The inventory status is optional; if supplied it must be active, allocatable,
and pickable. The sort method is required and must be active. A zone filter
requires a warehouse-scoped strategy and a zone in that warehouse.

Rules are returned in ascending sequence. Sequence numbers are positive and
unique per strategy, including inactive rules. Null filters mean no restriction
for that field; they do not automatically choose stock.

## Putaway strategies

Create a header with the same scope/code/name fields at `POST /putaway-strategies`.

```http
POST /api/v1/master/putaway-strategies/<putaway_strategy_id>/rules

{
  "sequence_no": 10,
  "category_id": "<category_id>",
  "location_type_id": "<STORAGE location_type_id>",
  "zone_id": "<zone_id>",
  "minimum_empty_percent": "25.1234"
}
```

All filters other than sequence are optional. A category filter requires an
owner-scoped strategy and an active category in that owner's hierarchy. A zone
must belong to the strategy warehouse. The location type must be active.

`minimum_empty_percent` is a JSON decimal string from 0 through 100, with at
most four decimal places. It may be null to leave this criterion unset.
This config does not calculate capacity or choose a destination location yet.

## Task configuration

Create a type:

```http
POST /api/v1/master/task-types

{"code":"CUSTOM_TASK","name":"Custom task","description":"Configured work type"}
```

Task statuses use `code/name/is_initial/is_final/is_cancelled`.
They share one global initial-status selection, switched atomically.
Task transition requests use `from_status_id/to_status_id/required_permission_id`,
with the same self-transition, terminal-state, and active-reference checks as
document transitions.

Create a priority:

```http
POST /api/v1/master/task-priorities

{"code":"CRITICAL","name":"Critical","priority_value":500}
```

Priority values are non-negative, unique integers; larger values sort first.
Task-type/status/priority configuration does not create, assign, or execute tasks.

## Updates, deactivation, and errors

PUT replaces mutable fields; include `is_active` and all fields you want to
keep. Nullable fields omitted from PUT are cleared. Codes, parent IDs, strategy
scope, and transition endpoints are immutable and rejected if included in PUT.
Use a new transition if endpoints must change.

Deactivate endpoints take no body. Deactivation preserves history and does not
cascade. Parent deactivation blocks new child writes; child deactivation remains
available. Unlike items/partners, these configuration tables have no
`updated_at` version column and do not use `expected_updated_at`.

Expected status codes: `201` create/allocation/rule replacement; `200`
read/update/deactivate; `400` validation/business-rule failure; `401` missing or
invalid session; `404` missing record/wrong parent; `409` uniqueness conflict or
exhausted counter.

## Startup reference data

Seeds mirror the relevant parts of
`database/master/00_bootstrap_reference_data.sql`:

- 9 modules; 37 document types and numbering formats.
- 135 document statuses and 152 document transitions.
- Task types `PUTAWAY, PICK, REPLENISHMENT, STOCK_COUNT, REWORK`.
- Task statuses `OPEN, ASSIGNED, IN_PROGRESS, COMPLETED, CANCELLED`, with 7 transitions.
- Priorities `LOW=100, NORMAL=200, HIGH=300, URGENT=400`.
- Picking sort methods `FEFO, FIFO, LOCATION, LOT`.
- Global `DEFAULT_FEFO` picking strategy with an `AVAILABLE`/FEFO rule.

Seed counts describe a fresh database. Existing edits and deactivations are
preserved. Restarting does not replace a user-selected initial state, revive
a disabled numbering rule, or reset counters. No putaway strategy is invented,
since safe putaway depends on your actual warehouse configuration.

## Tests

```powershell
cd C:\WmsProject\backend
go test ./...
go vet ./...
$env:WMS_INTEGRATION_TEST = '1'
go test ./... -count=1
Remove-Item Env:WMS_INTEGRATION_TEST
```

Integration tests read `backend/.env` and require PostgreSQL plus permission
to create test schemas. Existing-schema and fresh-schema suites roll back all
fixtures. The 24-client concurrency suite uses a uniquely named committed test
schema so independent connections can share rows, then drops only that test
schema. Use a development/test database, not production.

Tests cover all 72 new route session guards, request validation, workflow
relationships/initial states, strategy scope checks, repeatable seeds, numbering
history, same-day rule replacement, exact large counters, rollback, new-day
numbering, and concurrent uniqueness.

The requested layering remains: routes → master controller → master services →
per-table repositories → database. GORM entities remain one per table in
`models/master`; request/response structs remain in `dto/master`.
