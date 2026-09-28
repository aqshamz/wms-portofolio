# Return to vendor transactions

A quarantine `RETURN` decision creates a controlled outbound document for the
vendor on the original inbound order. The vendor is inherited rather than
entered by the user. Planning commits the case quantity but does not change
inventory.

```text
quarantine RETURN decision
        |
        v
RTV-... PLANNED (stock remains QUARANTINE)
        |
        +-- complete --> COMPLETED + RETURN_TO_VENDOR movement
        |
        +-- cancel ----> CANCELLED + commitment released, no movement
```

Viewing requires `OUTBOUND.READ`, completion requires
`OUTBOUND.RETURN_TO_VENDOR`, and cancellation requires `OUTBOUND.CANCEL`.
Owner and warehouse access scope is enforced on every route.

## Routes

```text
GET  /api/v1/outbound/vendor-returns
GET  /api/v1/outbound/vendor-returns/:id
POST /api/v1/outbound/vendor-returns/:id/complete
POST /api/v1/outbound/vendor-returns/:id/cancel
```

List requests require `owner_id` and `warehouse_id`; `status_code`, `search`,
`page`, and `page_size` are optional. Search covers the transaction, case, item,
lot, and vendor identity.

### Complete

```json
{
  "expected_version": 1,
  "expected_balance_version": 7,
  "completed_at": "2026-09-25T15:00:00+07:00"
}
```

Completion revalidates the planned quarantine balance, serial/handling-unit
identity, current document version, and balance version. It posts the inventory
movement, processes the linked disposition, and updates or closes the quarantine
case in the same database transaction.

### Cancel

```json
{
  "expected_version": 1,
  "reason": "Vendor declined the return authorization"
}
```

Cancellation is available only while planned. It records the reason and actor,
cancels the linked disposition, and makes that quantity undecided again.

The frontend is available at `/outbound/vendor-returns`. It shows the inherited
vendor, source identity, current stock snapshot, audit fields, and links back to
the quarantine case.
