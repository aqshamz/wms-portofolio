# Billing API

Billing converts immutable warehouse movements into controlled receivables:

`billing contract -> active rate card -> billable events -> billing run -> invoice -> payment / credit note`

All routes start with `/api/v1/billing`, require `Authorization: Bearer <token>`, and enforce both owner and warehouse access. Quantities and money are JSON strings so decimal precision is not lost. Dates use `YYYY-MM-DD`. Every editable or workflow request uses the latest `version_no`; a stale version returns `409 Conflict`.

## Supported pricing

- `source_kind=MOVEMENT` collects matching immutable `inventory_movement` rows. Use movement types such as `RECEIVE`, `PUTAWAY`, `PICK`, `PACK`, and `SHIP`.
- `source_kind=MANUAL` records an authorized service that has no stock movement.
- `source_kind=STORAGE` rates an idempotent end-of-day on-hand snapshot.
- `billing_basis=QUANTITY` multiplies movement quantity by unit rate.
- `billing_basis=EVENT` charges one unit rate per event.

`minimum_charge` is applied before `tax_percent`. Collection is idempotent by movement and rate-line combination.

## Study flow

Replace values in braces with IDs returned by the master and inventory endpoints.

### 1. Contract

`POST /api/v1/billing/contracts`

```json
{
  "owner_id": "{owner_id}",
  "warehouse_id": "{warehouse_id}",
  "business_date": "2026-09-09",
  "currency_code": "IDR",
  "billing_cycle": "MONTHLY",
  "payment_term_days": 30,
  "effective_from": "2026-09-01",
  "notes": "September study contract"
}
```

Activate it with `POST /contracts/{billing_contract_id}/activate`:

```json
{"expected_version": 1}
```

Only one active contract may overlap for the same owner and warehouse.

### 2. Rate card

First call `GET /api/v1/inventory/movement-types` and copy the IDs for `RECEIVE`, `PUTAWAY`, `PICK`, `PACK`, and `SHIP`.

`POST /api/v1/billing/rate-cards`

```json
{
  "billing_contract_id": "{billing_contract_id}",
  "business_date": "2026-09-09",
  "name": "September handling rates",
  "effective_from": "2026-09-01",
  "lines": [
    {
      "service_code": "RECEIVING",
      "description": "Receiving per base unit",
      "source_kind": "MOVEMENT",
      "movement_type_id": "{receive_movement_type_id}",
      "billing_basis": "QUANTITY",
      "unit_rate": "750",
      "minimum_charge": "10000",
      "tax_percent": "11"
    },
    {
      "service_code": "PUTAWAY",
      "description": "Putaway per movement",
      "source_kind": "MOVEMENT",
      "movement_type_id": "{putaway_movement_type_id}",
      "billing_basis": "EVENT",
      "unit_rate": "5000",
      "minimum_charge": "0",
      "tax_percent": "11"
    },
    {
      "service_code": "PICKING",
      "description": "Picking per base unit",
      "source_kind": "MOVEMENT",
      "movement_type_id": "{pick_movement_type_id}",
      "billing_basis": "QUANTITY",
      "unit_rate": "600",
      "minimum_charge": "0",
      "tax_percent": "11"
    },
    {
      "service_code": "PACKING",
      "description": "Packing per event",
      "source_kind": "MOVEMENT",
      "movement_type_id": "{pack_movement_type_id}",
      "billing_basis": "EVENT",
      "unit_rate": "4000",
      "minimum_charge": "0",
      "tax_percent": "11"
    },
    {
      "service_code": "SHIPPING",
      "description": "Shipping per event",
      "source_kind": "MOVEMENT",
      "movement_type_id": "{ship_movement_type_id}",
      "billing_basis": "EVENT",
      "unit_rate": "15000",
      "minimum_charge": "0",
      "tax_percent": "11"
    },
    {
      "service_code": "ADMIN",
      "description": "Manual administration fee",
      "source_kind": "MANUAL",
      "billing_basis": "EVENT",
      "unit_rate": "25000",
      "minimum_charge": "0",
      "tax_percent": "11"
    }
  ]
}
```

Call `/rate-cards/{id}/approve` with the returned version, then `/rate-cards/{id}/activate` with the next version. Draft cards support header PUT plus line POST/PUT/DELETE. Active rates are immutable.

### 3. Events

`POST /api/v1/billing/events/collect`

```json
{
  "owner_id": "{owner_id}",
  "warehouse_id": "{warehouse_id}",
  "date_from": "2026-09-01",
  "date_until": "2026-09-30"
}
```

Calling it again is safe: existing events are counted, not duplicated. Review them with:

`GET /api/v1/billing/events?owner_id={owner_id}&warehouse_id={warehouse_id}&date_from=2026-09-01&date_until=2026-09-30`

Exclude a wrong pending event through `POST /events/{id}/exclude`:

```json
{"reason":"Not chargeable under the client agreement"}
```

Create a manual event with `POST /contracts/{contract_id}/events/manual`:

```json
{
  "rate_card_line_id": "{manual_rate_card_line_id}",
  "business_date": "2026-09-09",
  "source_document_id": "CLIENT-SERVICE-001",
  "quantity": "1",
  "notes": "Approved special handling"
}
```

For daily storage billing, add a `STORAGE` rate line and call
`POST /api/v1/billing/events/storage-snapshot` at end of day:

```json
{
  "owner_id": "{owner_id}",
  "warehouse_id": "{warehouse_id}",
  "business_date": "2026-09-09"
}
```

The date must be today's configured WMS business date. Repeating it is safe;
one event is kept per date, balance identity, and storage rate.

### 4. Billing run

`POST /api/v1/billing/runs`

```json
{
  "billing_contract_id": "{billing_contract_id}",
  "business_date": "2026-09-30",
  "period_from": "2026-09-01",
  "period_until": "2026-09-30",
  "notes": "September billing"
}
```

Call `/runs/{id}/calculate` with `{"expected_version":1}`, inspect `GET /runs/{id}`, then call `/runs/{id}/review` with the latest version. `/runs/{id}/reopen` removes calculated charges and makes its events pending again. Cancelling a draft or calculated run also releases its events.

### 5. Invoice

`POST /api/v1/billing/runs/{run_id}/invoice`

```json
{
  "expected_version": 3,
  "issue_date": "2026-09-30",
  "notes": "September warehouse services"
}
```

If omitted, `due_date` comes from contract payment terms. The invoice starts in `DRAFT`; use its current version to review, reopen, issue, or void it.

### 6. Payment and credit

`POST /api/v1/billing/invoices/{invoice_id}/payments`

```json
{
  "expected_version": 3,
  "business_date": "2026-10-10",
  "amount": "100000",
  "reference": "BANK-TRX-20261010-001",
  "notes": "Client transfer"
}
```

Multiple partial payments are supported and may not exceed outstanding value. At zero outstanding, the invoice becomes `PAID`.

`POST /api/v1/billing/invoices/{invoice_id}/credit-notes`

```json
{
  "expected_version": 4,
  "business_date": "2026-10-10",
  "amount": "25000",
  "reason": "Service-level credit"
}
```

A credit cannot exceed outstanding value. If it closes the balance, the invoice becomes `SETTLED`.

## Read endpoints

- Lists (`/contracts`, `/rate-cards`, `/events`, `/runs`, `/invoices`) require `owner_id` and `warehouse_id` and accept `page` and `page_size`.
- Detail routes return the full resource; rate cards include rates, runs include charges, and invoices include immutable invoice lines.
- `GET /invoices/{id}/payments` and `/invoices/{id}/credit-notes` expose the receivable audit trail.
