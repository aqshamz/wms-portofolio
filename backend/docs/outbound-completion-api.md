# Outbound completion API

This final outbound slice closes the operational gaps around draft maintenance,
quality-exception recovery, transport setup, safe cancellation, inventory
identity consistency, and account scope authorization. All endpoints require the
same `Authorization: Bearer {token}` header used by the other API groups.

## Access scope

Every outbound order or transaction route now checks both active grants:

- `account_owner_access` for the authenticated account and owner;
- `account_warehouse_access` for the authenticated account and warehouse;
- an active `warehouse_owner` relationship for that owner/warehouse pair.

Outbound collection queries therefore require both `owner_id` and
`warehouse_id`. Carrier master routes are authenticated but global because a
carrier can serve multiple owners and warehouses.

## Draft delivery-order maintenance

Only an order in `DRAFT` can be edited. Every mutation uses the current
`expected_version` and returns the complete order with its new version.

```text
PUT    /api/v1/outbound/orders/{outbound_id}
POST   /api/v1/outbound/orders/{outbound_id}/lines
PUT    /api/v1/outbound/orders/{outbound_id}/lines/{outbound_line_id}
DELETE /api/v1/outbound/orders/{outbound_id}/lines/{outbound_line_id}
```

Example header update:

```json
{
  "expected_version": 1,
  "customer_id": "{customer_uuid}",
  "ship_to_partner_id": "{ship_to_uuid}",
  "requested_ship_at": "2026-09-09T15:00:00+07:00",
  "customer_reference": "STORE-REF-002",
  "notes": "Updated before release"
}
```

The final remaining line cannot be deleted; an outbound order must retain at
least one line.

## Failed-check replacement pick

A `DAMAGED`, `WRONG_ITEM`, or `SHORT` exception first needs a matching
`STOCK_CORRECTION` resolution. The correction movement must reference the same
owner, warehouse, item, staging location, and exact quantity.

Then create real replacement work:

```text
POST /api/v1/outbound/check-exceptions/{exception_id}/replacements
```

```json
{
  "balance_id": "{eligible_source_balance_id}",
  "replacement_qty": "1",
  "expected_balance_version": 4,
  "picking_strategy_id": "{strategy_uuid}",
  "priority_code": "HIGH",
  "notes": "Replacement after damaged carton"
}
```

The response contains a supplemental reservation, recovery wave, and pick task.
Start and confirm the returned task through the normal pick endpoints. A
successful confirmation consumes its reservation, creates a `PICK` movement,
adds the replacement execution to the original staging document, and resolves
the exception once all corrected quantity is fulfilled. Create a child check
with `parent_check_id` to verify the repaired staging document.

## Carrier and driver setup

```text
POST /api/v1/outbound/carriers
GET  /api/v1/outbound/carriers?active=true&search=express
GET  /api/v1/outbound/carriers/{carrier_id}
PUT  /api/v1/outbound/carriers/{carrier_id}

POST /api/v1/outbound/carriers/{carrier_id}/services
GET  /api/v1/outbound/carriers/{carrier_id}/services?active=true
GET  /api/v1/outbound/carriers/{carrier_id}/services/{service_id}
PUT  /api/v1/outbound/carriers/{carrier_id}/services/{service_id}

POST /api/v1/outbound/carriers/{carrier_id}/drivers
GET  /api/v1/outbound/carriers/{carrier_id}/drivers?active=true
GET  /api/v1/outbound/carriers/{carrier_id}/drivers/{driver_id}
PUT  /api/v1/outbound/carriers/{carrier_id}/drivers/{driver_id}
```

Create a carrier, service, and driver:

```json
{"code":"STUDY_CARRIER","name":"Study Express"}
```

```json
{"code":"REGULAR","name":"Regular delivery"}
```

```json
{
  "code":"DRIVER_001",
  "name":"Study Driver",
  "phone_number":"081234567890",
  "license_number":"SIM-B-STUDY"
}
```

Put `carrier_service_id` in `POST /shipments`, then assign a driver while the
shipment is `PLANNED`:

```text
POST   /api/v1/outbound/shipments/{shipment_id}/drivers
GET    /api/v1/outbound/shipments/{shipment_id}/drivers
DELETE /api/v1/outbound/shipments/{shipment_id}/drivers/{driver_id}
```

```json
{"driver_id":"{driver_uuid}","is_primary":true}
```

The driver must be active and belong to the carrier selected by the shipment
service. A shipment that specifies a carrier service cannot complete without one
primary driver.

## Safe cancellation

```text
POST /api/v1/outbound/packings/{packing_id}/cancel
POST /api/v1/outbound/shipments/{shipment_id}/cancel
POST /api/v1/outbound/deliveries/{delivery_id}/cancel
```

Cancellation is deliberately pre-execution only: packing must have no posted
lines, shipment must have no dispatched lines, and delivery must have no events.
After physical stock movement, use the auditable return/reversal workflow instead
of deleting operational history.

## Serial and handling-unit guarantees

Pick, pack, ship, and return now keep `serial_inventory` synchronized with the
movement balance. Serialized movement is one serial per movement. Handling-unit
stock must move as a complete single positive balance: its location advances to
staging and packing, clears when shipped, and is restored to the configured
return location on return-to-depot. Attempts to split or silently re-container
tracked handling-unit stock are rejected.

## Automated verification

```powershell
$env:GOCACHE='C:\WmsProject\.gocache'
go test ./...
go vet ./...

$env:WMS_INTEGRATION_TEST='1'
go test -run TestOutboundPostgreSQL -count=1 ./services/outbound
```

The PostgreSQL integration test runs once against the current schema and once
against a freshly migrated temporary schema. It covers the full happy path,
failed QC correction and replacement, recheck, carrier/primary-driver shipment,
partial delivery failure, and return-to-depot quarantine.
