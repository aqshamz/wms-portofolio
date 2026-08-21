-- Read-only helper views shared by the report query catalog.
-- Apply after wms_schema.sql. These views store no report data.

SET search_path TO wms, public;

CREATE OR REPLACE VIEW report_account_permission AS
SELECT DISTINCT account_role.account_id, permission.code AS permission_code
FROM account_role
JOIN app_role role ON role.role_id = account_role.role_id AND role.is_active
JOIN role_permission ON role_permission.role_id = role.role_id
JOIN app_permission permission
  ON permission.permission_id = role_permission.permission_id
 AND permission.is_active
WHERE account_role.valid_from <= clock_timestamp()
  AND (account_role.valid_until IS NULL
       OR account_role.valid_until > clock_timestamp());

CREATE OR REPLACE VIEW report_owner_scope AS
SELECT access.account_id, access.owner_id
FROM account_owner_access access
JOIN organization owner
  ON owner.organization_id = access.owner_id AND owner.is_active;

CREATE OR REPLACE VIEW report_warehouse_scope AS
SELECT owner_access.account_id, warehouse_owner.owner_id,
       warehouse_access.warehouse_id
FROM account_owner_access owner_access
JOIN warehouse_owner
  ON warehouse_owner.owner_id = owner_access.owner_id
 AND warehouse_owner.is_active
JOIN account_warehouse_access warehouse_access
  ON warehouse_access.account_id = owner_access.account_id
 AND warehouse_access.warehouse_id = warehouse_owner.warehouse_id
JOIN warehouse
  ON warehouse.warehouse_id = warehouse_access.warehouse_id
 AND warehouse.is_active;

COMMENT ON VIEW report_account_permission IS
    'Effective active permissions used by read-only report queries.';
COMMENT ON VIEW report_owner_scope IS
    'Owner/client rows visible to each application account.';
COMMENT ON VIEW report_warehouse_scope IS
    'Owner and warehouse combinations visible to each application account.';
