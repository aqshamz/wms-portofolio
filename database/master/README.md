# WMS master-data setup and CRUD

The cross-module normalization rationale is documented in
`../NORMALIZATION_AUDIT.md`.

These files are intentionally separated by application menu so the master-data
layer can be reviewed without reading the transaction schema.

## Setup order

1. Apply `../wms_schema.sql` to an empty PostgreSQL 15+ database.
2. Review and apply `00_bootstrap_reference_data.sql`.
3. Create the first account with `../auth_queries.sql` and assign `SUPER_ADMIN`.
4. Create organizations and warehouses with `01_organization_warehouse_crud.sql`.
5. Create partners and items with `02_partner_item_crud.sql`.
6. Review operational configuration with `03_operational_config_crud.sql`.
7. Review the focused inbound flow in `../inbound/README.md` before backend work.

The CRUD files are query catalogs. Execute one numbered query at a time through
prepared statements; do not execute a whole CRUD catalog as one script.

## Conventions to keep stable in the backend

- Master primary keys are UUIDs. Business `code` values are separate unique keys.
- Transaction primary keys are generated varchar identifiers.
- Codes used in historical document IDs are treated as immutable after use.
- Referenced master rows are deactivated with `is_active = false`, not deleted.
- `organization`, `warehouse`, `business_partner`, and `item` updates use
  `updated_at` for optimistic concurrency.
- Statuses, allowed transitions, reasons, task configuration, numbering rules,
  menus, and permissions are database master data rather than application enums.
- Owner and warehouse scope are checked separately from functional permission.
- Password hashing and raw session-token generation remain application duties.

## Bootstrap behavior

The bootstrap script uses conflict-safe inserts. Re-running it does not create
duplicate business codes and does not overwrite existing customized rows. The
starter names, workflows, permissions, menus, and number formats should still be
reviewed before production deployment.

## Performance verification

Query text length is not a performance measurement. Validate representative list
and authorization queries after loading realistic data with:

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT ...;
```

Do this in a development database because `EXPLAIN ANALYZE` executes the query.
