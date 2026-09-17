# Operational action permissions

Operational routes require an exact action permission in addition to an active
session and the applicable SIS owner and Bandung warehouse scopes. `*` remains
the administrator override. Legacy module permissions such as `INBOUND.WRITE`,
`OUTBOUND.WRITE`, and `BILLING.WRITE` do not grant these protected actions.

## Superadmin

`SUPERADMIN` is the unrestricted system role. Its `*` permission authorizes all
current and future menus/actions and removes account owner/warehouse scope
restrictions for master data and inbound, outbound, and billing APIs. Individual
scope grants are not required. Ordinary roles, including those with every named
permission but without `*`, remain scoped. Login/account status, document-state
rules, optimistic concurrency, stock validation, and assigned-task workflow
rules still apply.

Bootstrap administrators receive this role when bootstrap seeding is enabled.
For an existing local development administrator, run from `backend`:

```text
go run ./cmd/seed-superadmin -account admin
```

This additive, repeatable command assigns only the named account, preserves its
existing roles/grants, and refuses production or non-loopback databases. It does
not reactivate a deliberately disabled role or permission. The wildcard grant
is visible in Permissions as **Unrestricted system access**; only trusted system
administrators should receive it. Deactivating/revoking its role removes the
wildcard on the next authenticated request, and last-security-administrator
protection includes wildcard grants. Bootstrap seeding restores missing grants
for the configured bootstrap administrator; disable bootstrap recovery when
managing those grants manually.

## Permission families

| Area | Permission codes |
| --- | --- |
| Inbound | `READ`, `PLAN`, `APPROVE`, `RECEIVE`, `QC`, `PUTAWAY`, `ASSIGN`, `QUARANTINE_DISPOSE`, `REWORK`, `CANCEL` |
| Inventory | `READ`, `IDENTITY`, `MOVE`, `STATUS_CHANGE`, `ADJUST`, `COUNT`, `TRANSFER` |
| Outbound | `READ`, `PLAN`, `PICK`, `STAGE`, `CHECK`, `PACK`, `TRANSPORT`, `SHIP`, `DELIVER`, `CANCEL`, `CONFIG` |
| Billing | `READ`, `CONFIGURE`, `PREPARE`, `APPROVE`, `ISSUE`, `PAYMENT` |

Prefix each code with its area, for example `INBOUND.QC`, `OUTBOUND.SHIP`, or
`BILLING.PAYMENT`.

## SIS and Bandung role matrix

| Role | Account | Responsibilities |
| --- | --- | --- |
| `WHADMIN` | `whm` | Warehouse approval, assignment, quarantine disposal, cancellation, inventory control, outbound configuration, shipment and delivery; no document planning or billing |
| `WAREHOUSE_PLANNER` | `planner.warehouse` | Create inbound and outbound plans; cannot approve or execute them |
| `RECEIVER` | `receiver.warehouse` | Receive goods, create inventory identities, and execute putaway |
| `PICKER` | `picker.warehouse` | Pick and stage outbound goods |
| `IC` | `ic.warehouse` | Inventory identity, movement, status, adjustment, count, and transfer control |
| `QC` | `qc.warehouse` | Perform inbound quality inspections; cannot decide quarantine disposal |
| `REWORK_OPERATOR` | `rework.warehouse` | Execute assigned quarantine rework; cannot inspect or decide disposition |
| `PACKER` | `packer.warehouse` | Check and pack outbound goods |
| `DISPATCHER` | `dispatcher.warehouse` | Ship and complete delivery operations |
| `BILLING_OPERATOR` | `billing.operator` | Configure contracts/rates and prepare billing runs and draft invoices |
| `BILLING_APPROVER` | `billing.approver` | Approve/review billing records and issue or void invoices/credit notes |
| `CASHIER` | `cashier.sis` | Read billing records and record payments |

Every operational account is scoped to owner `SIS` and warehouse `BDG_WH`.
Keep `BILLING_OPERATOR` and `BILLING_APPROVER` on separate people to preserve
maker-checker separation. Do not combine these roles through direct permission
exceptions.

Keep `WAREHOUSE_PLANNER` and `WHADMIN` on separate people as well. The planner
creates inbound/outbound plans and the manager approves or cancels them.

The warehouse manager is deliberately allowed to approve inbound plans and
quarantine disposal because those are warehouse-control decisions. Billing
approval, invoice issue, and payment remain SIS finance responsibilities.

## Repeatable development setup

Run `scripts/seed-sis-master-data.ps1` against a loopback API. The script is
additive and idempotent: it creates missing master data, permissions, roles, and
accounts; normalizes the twelve role permission sets; applies SIS/BDG scopes; and
activates the accounts. Newly generated passwords are printed once and are not
stored in the repository.
