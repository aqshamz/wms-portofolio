# Inbound part 1: PO through receiving

This slice implements the inbound flow up to stock waiting for quality control:

```text
PURCHASE_ORDER: DRAFT -> APPROVED -> PARTIALLY_RECEIVED | RECEIVED
INBOUND:        DRAFT -> RELEASED -> PARTIALLY_RECEIVED | RECEIVED
RECEIPT:        OPEN -> COMPLETED
                                  |
                                  +-> RECEIVE movement -> QC_PENDING balance
```

Continue from the resulting `QC_PENDING` balance with
[Inbound part 2](inbound-part-2-api.md), which covers inspection, disposition,
putaway, and negative outcomes.

## Important behavior

- A purchase order belongs to one owner, supplier, and warehouse.
- An inbound order is the ASN/expected-arrival document created from selected
  purchase-order lines. Multiple inbound orders may split a PO quantity.
- A receipt can split an accepted line across lots, handling units, serials,
  and receiving/QC locations.
- `received_qty` is the physical quantity counted at the dock.
  `rejected_qty` is the portion rejected immediately at the dock.
- Accepted quantity is `received_qty - rejected_qty`. Batch source quantities
  must equal accepted quantity exactly.
- A fully rejected line uses `batches: []` and creates no inventory.
- Creating an `OPEN` receipt does not change inventory. Completing it performs
  all RECEIVE postings and status transitions in one database transaction.
- Accepted inventory is posted in the item's base UOM with inventory status
  `QC_PENDING`. A newly created lot has quality status `PENDING`.
- A completion retry on an already completed receipt is safe and does not
  create duplicate movements.
- State changes require `expected_version`; stale writes return HTTP `409`.
- All routes require a valid, non-revoked Bearer session. Owner-scoped routes
  also require both account-owner and account-warehouse access plus an active
  warehouse-owner relationship. List requests therefore require both
  `owner_id` and `warehouse_id`.

## Routes

```text
POST /api/v1/inbound/purchase-orders
GET  /api/v1/inbound/purchase-orders?owner_id=<uuid>&warehouse_id=<uuid>&status_code=APPROVED&search=<text>&page=1&page_size=20
GET  /api/v1/inbound/purchase-orders/:id
PUT  /api/v1/inbound/purchase-orders/:id
POST /api/v1/inbound/purchase-orders/:id/lines
PUT  /api/v1/inbound/purchase-orders/:id/lines/:line_id
DELETE /api/v1/inbound/purchase-orders/:id/lines/:line_id
POST /api/v1/inbound/purchase-orders/:id/approve

POST /api/v1/inbound/orders
GET  /api/v1/inbound/orders?owner_id=<uuid>&warehouse_id=<uuid>&status_code=RELEASED&page=1&page_size=20
GET  /api/v1/inbound/orders/:id
PUT  /api/v1/inbound/orders/:id
POST /api/v1/inbound/orders/:id/lines
PUT  /api/v1/inbound/orders/:id/lines/:line_id
DELETE /api/v1/inbound/orders/:id/lines/:line_id
POST /api/v1/inbound/orders/:id/release

POST /api/v1/inbound/receipts
GET  /api/v1/inbound/receipts?owner_id=<uuid>&warehouse_id=<uuid>&status_code=OPEN&page=1&page_size=20
GET  /api/v1/inbound/receipts/:id
PUT  /api/v1/inbound/receipts/:id
POST /api/v1/inbound/receipts/:id/complete
```

List endpoints require `owner_id` and `warehouse_id`. Every request rejects unknown JSON fields,
extra JSON values, oversized bodies, unknown query parameters, and repeated
query parameters.

## Draft editing and cancelled-document replacement

Purchase-order and inbound-order headers and lines can be added, updated, or
deleted only while the document is `DRAFT`. At least one line must remain.
Every mutation requires the latest `expected_version` and increments the
document version.

`PUT /receipts/:id` is available only while the receipt is `OPEN` and no batch
has been posted. It replaces the editable receipt header and its complete
line/batch draft atomically. `business_date` and `inbound_id` remain immutable.
If any corrected line, quantity, lot, serial, handling unit, or location is
invalid, the whole edit is rolled back.

Cancelled, completed, closed, and reversed documents are immutable. Create a
new document instead and supply the appropriate optional predecessor field:

```json
{
  "supersedes_purchase_order_id": "<cancelled_purchase_order_id>"
}
```

Inbound orders use `supersedes_inbound_id`; receipts use
`supersedes_receipt_id`. The predecessor must be cancelled and must have the
same business scope and source relationship. Only one direct successor is
allowed. Responses expose both the `supersedes_*` and `successor_*` IDs.

## Test with the study master data

Start the API and seed the study master if needed:

```powershell
cd C:\WmsProject\backend
go run .
```

In another PowerShell window:

```powershell
cd C:\WmsProject\backend
.\scripts\seed-study-data.ps1 -BaseUrl http://localhost:8080 | Out-Null
```

Log in using `POST /api/v1/auth/login`, copy `data.token`, and add these headers
to every request:

```text
Authorization: Bearer <token>
Content-Type: application/json
```

Discover the current UUIDs instead of copying IDs from another database:

```text
GET /api/v1/master/organizations?search=STUDY_OWNER&active=true
GET /api/v1/master/warehouses?search=STUDY_WH&active=true
GET /api/v1/master/business-partners?owner_id=<owner_uuid>&search=STUDY_SUPPLIER&active=true
GET /api/v1/master/items?owner_id=<owner_uuid>&search=STUDY_COFFEE_250G&active=true
GET /api/v1/master/warehouses/<warehouse_uuid>/locations
```

Use the UUID of `EA` from the chosen item's `uoms`. Use `STUDY_RCV_01` as the
dock location and `STUDY_QC_01` as the accepted-stock location.

### 1. Create the purchase order

```http
POST /api/v1/inbound/purchase-orders

{
  "owner_id": "<owner_uuid>",
  "vendor_id": "<supplier_uuid>",
  "warehouse_id": "<warehouse_uuid>",
  "business_date": "2026-09-07",
  "purchase_order_no": "CLIENT-PO-STUDY-001",
  "ordered_at": "2026-09-07T08:00:00+07:00",
  "expected_arrival_at": "2026-09-07T09:00:00+07:00",
  "notes": "Inbound part 1 study",
  "lines": [
    {
      "item_id": "<item_uuid>",
      "ordered_qty": "2",
      "uom_id": "<box_uom_uuid>",
      "expected_lot_no": "STUDY-INB-LOT-001",
      "expected_expiry_date": "2027-09-07"
    }
  ]
}
```

The study 250 g item converts `1 BOX = 24 EA`; two boxes will eventually post
48 EA. Save `data.purchase_order_id`, `data.version_no`, and
`data.lines[0].purchase_order_line_id`.

### 2. Approve the PO

```http
POST /api/v1/inbound/purchase-orders/<purchase_order_id>/approve

{
  "expected_version": 1
}
```

Use the actual version returned by the previous response.

### 3. Create and release the inbound order/ASN

```http
POST /api/v1/inbound/orders

{
  "purchase_order_id": "<purchase_order_id>",
  "business_date": "2026-09-07",
  "expected_arrival_at": "2026-09-07T09:00:00+07:00",
  "external_reference": "ASN-STUDY-001",
  "supplier_reference": "DELIVERY-STUDY-001",
  "lines": [
    {
      "purchase_order_line_id": "<purchase_order_line_id>",
      "expected_qty": "2"
    }
  ]
}
```

Save `data.inbound_id`, `data.version_no`, and
`data.lines[0].inbound_line_id`, then release it:

```http
POST /api/v1/inbound/orders/<inbound_id>/release

{
  "expected_version": 1
}
```

### 4. Open the receipt

This example counts two boxes, rejects one box at the dock, and accepts one box.
It will post 24 EA when completed.

```http
POST /api/v1/inbound/receipts

{
  "inbound_id": "<inbound_id>",
  "business_date": "2026-09-07",
  "received_at": "2026-09-07T09:15:00+07:00",
  "dock_location_id": "<STUDY_RCV_01_uuid>",
  "vehicle_number": "B 1234 STUDY",
  "delivery_note_no": "DN-STUDY-001",
  "lines": [
    {
      "inbound_line_id": "<inbound_line_id>",
      "received_qty": "2",
      "rejected_qty": "1",
      "batches": [
        {
          "source_qty": "1",
          "received_location_id": "<STUDY_QC_01_uuid>",
          "lot": {
            "lot_number": "STUDY-INB-LOT-001",
            "manufacture_date": "2026-08-01",
            "expiry_date": "2027-08-01"
          }
        }
      ]
    }
  ]
}
```

Before completing, verify the response has `status_code: "OPEN"`,
`base_qty: "24.000000"`, and `initial_balance_id: null`. Querying inventory at
this point must show no movement with this receipt ID.

For a full dock rejection, set `rejected_qty` equal to `received_qty` and send
`"batches": []`.

### 5. Complete and verify

```http
POST /api/v1/inbound/receipts/<receipt_id>/complete

{
  "expected_version": 1
}
```

The result should contain:

- receipt status `COMPLETED`;
- a non-null batch `initial_balance_id`;
- one `RECEIVE` movement for accepted base quantity;
- a balance at `STUDY_QC_01` with status `QC_PENDING`;
- inbound and PO progress changed to `RECEIVED` when their full physical
  expected quantity has been accounted for.

Verify through inquiry:

```text
GET /api/v1/inventory/movements?owner_id=<owner_uuid>&source_document_id=<receipt_id>&page=1&page_size=20
GET /api/v1/inventory/balances?owner_id=<owner_uuid>&warehouse_id=<warehouse_uuid>&item_id=<item_uuid>&page=1&page_size=20
GET /api/v1/inbound/receipts/<receipt_id>
GET /api/v1/inbound/orders/<inbound_id>
GET /api/v1/inbound/purchase-orders/<purchase_order_id>
```

Use a new `purchase_order_no`, ASN reference, delivery note, and lot number when
repeating the example. The external PO number is unique per owner.

## Automated verification

Fast tests (PostgreSQL integration test skips by default):

```powershell
cd C:\WmsProject\backend
go test ./...
```

Full PostgreSQL test, including both the existing schema and a freshly migrated
temporary schema:

```powershell
cd C:\WmsProject\backend
$env:WMS_INTEGRATION_TEST = '1'
go test -p 1 ./...
Remove-Item Env:WMS_INTEGRATION_TEST
```

The inbound integration test rolls back all fixtures. It covers workflow
transitions, full and partial dock rejection, minimum shelf life, lot creation
rollback, UOM conversion, delayed inventory posting, QC_PENDING stock,
document progress, and idempotent receipt completion.
