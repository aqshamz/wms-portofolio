# Partial fulfillment and closing demand short

The delivery order (DO) retains its original order-UOM and base-quantity snapshots.
All allocation and fulfillment progress, including shortages, is in frozen base units.

## Normal flow

1. A planner chooses **Allocate available stock** in DO details. The API uses the
   configured picking strategy and `allow_partial=true`. No stock moves yet.
2. `ALLOCATED` and `PARTIALLY_ALLOCATED` DOs can enter a wave. Only active
   reservations are picked, checked, packed and dispatched in that cycle.
3. Completing the shipment sets the DO to `PARTIALLY_FULFILLED` if any demand is
   outstanding; otherwise it sets `SHIPPED`. Available demand can be allocated
   again after the previous warehouse cycle's shipment is complete.
4. The next wave, staging, check, packing and shipment use the **same DO** with
   cumulative line counters. Completed document membership remains audit history.
5. Deliveries are recorded per shipment. An earlier shipment's delivery cannot
   overwrite a backorder or an active warehouse cycle. Full delivery completion
   requires final outcomes for every dispatched shipment, not just the latest one.

One warehouse fulfillment cycle is active per DO at a time. This does not prevent
an earlier shipment from being delivered while the next cycle is processed.
There is no automatic backorder cloning, stock creation, or substitution.
Whole-HU and serialized-stock constraints continue to apply.

An allocation with no additional stock leaves the DO status and version unchanged.
The API's existing strict allocation option (`allow_partial=false`) still rolls back
the entire reservation attempt when insufficient stock is available.

## Quantities and audit

DO line responses include:

- `open_qty`: ordered base quantity minus shipped, accepted checking shortage and
  closed-short quantity. This includes demand currently being processed.
- `unallocated_qty`: ordered base quantity minus allocated and closed-short quantity.
- `closed_short_qty`: uncommitted demand explicitly cancelled by an authorized user.
- `short_accepted_qty`: the existing approved checking-exception shortage. It is
  distinct from DO demand closure and must not be counted twice.

For example, order 100, ship 70: open demand remains 30 on the same DO. Ship another
30 later, or explicitly close the remaining 30 short. Ordered quantity remains 100.

## Authorized closure

`POST /api/v1/outbound/orders/:id/close-short` accepts:

```json
{"expected_version": 12, "reason": "Customer agreed to cancel the remaining demand"}
```

The endpoint requires `OUTBOUND.CLOSE_SHORT` and the existing owner/warehouse
access scope. `OUTBOUND.PLAN`, `OUTBOUND.WRITE` and `OUTBOUND.CANCEL` do not confer
this permission. Grant the new permission only to designated approvers; superadmin
retains its existing universal access.

The UI requires confirmation and a nonblank reason. Closure:

- is allowed in `RELEASED` or `PARTIALLY_FULFILLED` only;
- rejects stale versions, active draft waves and reserved/picked/in-process demand;
- closes **all** remaining demand across the DO's lines atomically;
- records the account, time, reason and each line's closed-short base quantity;
- sets final demand status `CLOSED_SHORT`, preventing later allocation;
- does not change inventory, ordered snapshots or shipment/POD history.

Delivery and return tracking of already dispatched stock continues on the linked
delivery documents; those events never replace the DO's closed-short status.
Returned stock is not automatically converted into a new fulfillment demand.

## Deployment and verification

Restart the API to apply migration **28** and seed the new statuses, transitions and
permission. The migration backfills packing source checks and completed shipment
links without marking history as removed. It refuses an ambiguous legacy packing
that spans several checks instead of silently dropping its lines.

PostgreSQL integration tests use a fresh transaction-scoped schema and roll back:

```powershell
$env:WMS_INTEGRATION_TEST = '1'
go test ./services/outbound -run '^TestOutboundPostgreSQL/fresh_true$' -count=1
```

They cover partial waves, two shipments for one DO, no-stock retries, previous-cycle
delivery isolation, packing history, multi-line close-short, required reasons,
stale versions, reserved-stock rejection and delivery after demand closure.
