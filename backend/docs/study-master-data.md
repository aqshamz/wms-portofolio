# Study master data

Created and verified in the local development database on 2026-09-04, through
the existing API. These are persisted sample records, not frontend dummy data.
Existing records were left untouched. A second run created zero new records.

## The example business

`STUDY_OP` operates `STUDY_WH` in Jakarta and `STUDY_WH_2` in Surabaya.
`STUDY_OWNER` is its client and owns the coffee inventory and serialized study
scanners. The warehouse-owner associations allow that client's goods to be
stored in both warehouses. Both operator and owner are organizations;
their roles come from these relationships, not a separate organization type.

The client buys from `STUDY_SUPPLIER` and sells to `STUDY_CUSTOMER`. Both business
partners, both coffee items and the scanner item belong to **STUDY_OWNER**, not
the operator.

Set up your own data in this order:

1. Organizations (operator and stock owner), then warehouse using operator UUID.
2. Warehouse-owner association using stock-owner UUID.
3. Zones and locations, referencing existing location-type UUIDs.
4. Owner's business partners and their partner-type assignments.
5. Owner's categories and items, referencing global UOM UUIDs.
6. Item-specific UOM conversions and barcodes.
7. Owner/warehouse strategies and rules; document/task configuration.

`OP_150533`, `OWN_150533`, and `WH_150533` were inactive when inspected. An
inactive referenced record causes validation failures even when its UUID is
correct. They have NOT been reactivated by this setup.

## Actual UUIDs in this database

These IDs are specific to this database. On a fresh database, discover IDs with
GET requests or use the returned seed-script object; never copy IDs across DBs.

| Code | Meaning | UUID |
| --- | --- | --- |
| STUDY_OP | Warehouse operator organization | 671a873f-2a78-40bb-b229-349493605c00 |
| STUDY_OWNER | Stock-owner organization | 5e6a104a-51c2-4e94-8643-e339bcd5242f |
| STUDY_WH | Warehouse | 59113433-3082-4208-87c1-330f4f56fcba |
| STUDY_SUPPLIER | Supplier business partner | 14072332-850c-4d74-bbce-c44d8817dd93 |
| STUDY_CUSTOMER | Customer business partner | c12b4413-7eb2-4b03-a3b1-b0583318690c |
| STUDY_COFFEE | Item category | 6b1348f8-628a-43ca-95dd-69851c4800ed |
| STUDY_COFFEE_250G | 250 g coffee pack | b9d199ad-343e-480e-a5a5-c8af97d3ddee |
| STUDY_COFFEE_1KG | 1 kg coffee pack | 89e7b8d1-6ea0-43ee-8e5f-0c78bc5ba9ad |
| EA | Global each UOM | 6df21b70-0d00-4406-b8d1-71aeca1b62fa |
| BOX | Global box UOM | 5f42c32b-95ed-4df0-b492-5c4e02309405 |
| STUDY_FEFO | Picking strategy | 39b6709f-6614-43b1-afbc-7638f0fc47e5 |
| STUDY_STORAGE | Putaway strategy | 0b85e4f7-427f-440c-b3d9-857c220556d3 |
| STUDY_RECEIPT | Document type | a224992f-a1fe-442d-b015-e919ee7d34c4 |
| STUDY_PICK | Task type | f930e9da-3057-488b-a079-0fd402a21e20 |

## Browse through your API client

Start your normal backend with `go run .` from `C:\WmsProject\backend`.
Use `http://localhost:8080` (or your configured port), not `0.0.0.0`.
The temporary port 8094 used during setup is not needed to access the saved data.

First, `POST /api/v1/auth/login` with JSON:

```json
{
  "identifier": "YOUR_USERNAME",
  "password": "YOUR_PASSWORD"
}
```

Copy `data.token`. For every master request set `Authorization: Bearer <token>`.
POST requests also use `Content-Type: application/json`. GET requests need no body.

Try these GET paths in order:

```text
/api/v1/master/organizations?search=STUDY&active=true
/api/v1/master/warehouses?search=STUDY&active=true
/api/v1/master/warehouses/59113433-3082-4208-87c1-330f4f56fcba/owners
/api/v1/master/warehouses/59113433-3082-4208-87c1-330f4f56fcba/zones
/api/v1/master/warehouses/59113433-3082-4208-87c1-330f4f56fcba/locations
/api/v1/master/business-partners?owner_id=5e6a104a-51c2-4e94-8643-e339bcd5242f
/api/v1/master/business-partners/14072332-850c-4d74-bbce-c44d8817dd93/types
/api/v1/master/items?owner_id=5e6a104a-51c2-4e94-8643-e339bcd5242f
/api/v1/master/items/b9d199ad-343e-480e-a5a5-c8af97d3ddee
```

Example complete URL:
`http://localhost:8080/api/v1/master/items/b9d199ad-343e-480e-a5a5-c8af97d3ddee`
The item detail includes both `uoms` and `barcodes`.

### Understand the warehouse-owner request

The following association is **already created**, so you do not need to send it
again. This is the correct request shape for your earlier problem:

```http
POST /api/v1/master/warehouses/59113433-3082-4208-87c1-330f4f56fcba/owners
Authorization: Bearer <token>
Content-Type: application/json

{
  "owner_id": "5e6a104a-51c2-4e94-8643-e339bcd5242f"
}
```

`owner_id` takes the organization's UUID, not `STUDY_OWNER` or `OP_150533`.
The warehouse path also takes a UUID. Both referenced records must be active.
Warehouse-owner assignment is distinct from account-owner/account-warehouse
access. This seed does not create accounts or grant access scopes.

## Locations and units

| Zone | Location | Location type / intended use |
| --- | --- | --- |
| STUDY_INBOUND | STUDY_RCV_01 | RECEIVING: incoming goods |
| STUDY_INBOUND | STUDY_QC_01 | QUARANTINE: goods awaiting disposition |
| STUDY_STORAGE | STUDY_BULK_01 | STORAGE: bulk storage |
| STUDY_STORAGE | STUDY_PICK_01 | PICK_FACE: picking location |
| STUDY_OUTBOUND | STUDY_STAGE_01 | STAGING: staged outbound goods |
| STUDY_OUTBOUND | STUDY_SHIP_01 | SHIPPING: dispatch area |

Both items use **EA as their base unit**: one EA is one sealed pack, not one gram.
For the 250 g item, `1 BOX = 24 EA`; for the 1 kg item, `1 BOX = 6 EA`.
Thus, two boxes of the 250 g item represent 48 EA. This is an explanation of the
conversion, not stock that was inserted into the database.

Each item has EA and BOX barcodes, such as `STUDY-COFFEE-250G-EA` and
`STUDY-COFFEE-250G-BOX`. These are study identifiers, not GS1/EAN codes.
Both items are lot-controlled, not serial-controlled, with shelf life 365 days
and minimum remaining life at receipt 90 days configured in the master.

Inspect shared references with GET:

```text
/api/v1/master/location-types
/api/v1/master/uoms
/api/v1/master/partner-types
/api/v1/master/inventory-statuses
/api/v1/master/quality-statuses
/api/v1/master/inspection-results
```

Inventory status describes stock availability (e.g. AVAILABLE/HOLD), quality
status describes quality state (e.g. PENDING/PASSED), and inspection result
records an inspection outcome (e.g. ACCEPTED/REJECTED). Existing reference rows
are reused; no duplicate study classifications were created.

## Workflow and strategies

GET these configuration paths:

```text
/api/v1/master/picking-strategies/39b6709f-6614-43b1-afbc-7638f0fc47e5/rules
/api/v1/master/putaway-strategies/0b85e4f7-427f-440c-b3d9-857c220556d3/rules
/api/v1/master/document-types/a224992f-a1fe-442d-b015-e919ee7d34c4/statuses
/api/v1/master/document-types/a224992f-a1fe-442d-b015-e919ee7d34c4/transitions
/api/v1/master/document-types/a224992f-a1fe-442d-b015-e919ee7d34c4/number-rules
/api/v1/master/task-types?search=STUDY
/api/v1/master/task-statuses
/api/v1/master/task-transitions
/api/v1/master/task-priorities
```

- Picking: AVAILABLE stock in STUDY_STORAGE, sorted FEFO (first expiry first out).
- Putaway: STUDY_COFFEE category, STUDY_STORAGE zone, STORAGE location type,
  minimum-empty percentage configured as 25.
- Receipt workflow: DRAFT (initial) to READY to COMPLETED (final), with cancellation
  allowed from DRAFT or READY into CANCELLED (final).
- Number rule: prefix STD_RCPT, separator `-`, six-digit sequence, partner and
  warehouse codes included, effective 2026-09-04 in this database.
- Task configuration: isolated STUDY_PICK type, existing shared statuses,
  transitions and priorities unchanged.

These are configuration records, not an implemented receipt/picking/putaway
execution flow. No inventory identities, lots, balances, receipts, tasks, or
movements were inserted. No document ID was allocated and no daily counter was
consumed by this setup. ID generation is a consuming POST, not a preview; see
the operational API guide before testing it.

Opening inventory is intentionally a separate opt-in step. See
[Study inventory data](study-inventory-data.md) and run
`scripts/seed-study-inventory.ps1` after this master script.

The current master APIs validate bearer sessions but do not yet enforce role
permissions or account owner/warehouse scope authorization.

## Repeatable setup script

With the local development backend running, execute in PowerShell:

```powershell
cd C:\WmsProject\backend
$study = & .\scripts\seed-study-data.ps1 -BaseUrl http://localhost:8080
$study | Select-Object created, reused
$study.items | Select-Object code, item_id
```

The first run here created 41 records/associations via API calls, plus two base
item-UOM rows automatically created by the item service. The verified rerun
reported `created = 0`, `reused = 41`.

The script logs in using bootstrap credentials from `.env` and logs that session
out in `finally`. If your login differs, supply an existing token using the
`WMS_STUDY_TOKEN` environment variable; that caller-owned session is not revoked.
Do not commit credentials or tokens. The script only accepts loopback URLs;
use it only against a development database.

It creates missing study records, reuses existing active records, and does not
reset your edited values. Inactive or incompatible scoped records cause a clear
error for manual review. It never replaces an existing number-rule history.
It is not one transaction across all requests: after partial failure, fix the
cause and rerun. There is intentionally no destructive reset/cleanup command.

More request bodies and validation rules: [Catalog API](catalog-api.md) and
[Operational configuration API](operational-configuration-api.md).
