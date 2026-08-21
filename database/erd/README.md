# WMS ERD study package

This package is generated from the PostgreSQL schema and covers all 150
tables, their column types, keys, defaults, nullability, and foreign-key
relationships.

## Files

- `full_erd.dbml` - complete model for dbdiagram.io or another DBML viewer.
- `full_erd.mmd` - complete Mermaid ER diagram source.
- `domains/*.mmd` - smaller Mermaid diagrams for focused study.
- `DATA_DICTIONARY.md` - every table and column with PostgreSQL data type.
- `RELATIONSHIPS.md` - every foreign key and delete behavior.
- `erd_manifest.json` - structured source used by the interactive explorer.
- `generate-erd.js` - repeatable generator; run after schema changes.

## Suggested study order

1. Security and access (12 tables)
2. Master data and workflow configuration (36 tables)
3. Inventory identity (3 tables)
4. Inbound (13 tables)
5. Stock control (21 tables)
6. Outbound (41 tables)
7. Billing (24 tables)

The full diagram is intentionally large. Use the domain diagrams first, then
consult the relationship catalog for cross-domain foreign keys.

## Regenerate

From the workspace root:

`node database/erd/generate-erd.js`
