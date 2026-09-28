# Outbound disposal transactions

Disposal is a controlled outbound lifecycle for failed quality stock. A
quarantine `DISPOSE` decision creates a `PLANNED` transaction; it does not
remove inventory. This separates an owner's decision from the warehouse's
physical disposal and preserves a cancellable audit trail.

## Lifecycle

```text
quarantine DISPOSE decision
        |
        v
DSP-... PLANNED (stock remains QUARANTINE, quantity is committed)
        |
        +-- complete --> COMPLETED + DISPOSE movement + case processed
        |
        +-- cancel ----> CANCELLED + commitment released, no movement
```

All list and detail routes require `OUTBOUND.READ`. Completion requires
`OUTBOUND.DISPOSE`; cancellation requires `OUTBOUND.CANCEL`. Account access
scope is enforced for the owner and warehouse on every route.

## Routes

```text
GET  /api/v1/outbound/disposals
GET  /api/v1/outbound/disposals/:id
POST /api/v1/outbound/disposals/:id/complete
POST /api/v1/outbound/disposals/:id/cancel
```

List requests require `owner_id` and `warehouse_id`. Optional filters are
`status_code`, `search`, `page`, and `page_size`.

### Complete

```json
{
  "expected_version": 1,
  "expected_balance_version": 7,
  "completed_at": "2026-09-25T10:00:00+07:00"
}
```

The balance version is refreshed with disposal detail. Completion fails safely
if the source balance changed, lacks sufficient quarantine quantity, or its
serial/handling-unit identity is no longer eligible. On success, the transaction
stores the resulting `inventory_movement_id` and completion audit fields.

### Cancel

```json
{
  "expected_version": 1,
  "reason": "Owner withdrew disposal authorization"
}
```

Cancellation is allowed only from `PLANNED`. It records the actor, timestamp,
and reason, cancels the linked quarantine disposition, and recalculates the
case's committed and remaining quantities.

## Frontend

The scoped transaction list is available at `/outbound/disposals`. Quarantine
history links to its generated transaction, and disposal detail links back to
the source case. The UI explicitly distinguishes processed, pending, and
undecided quarantine quantities.
