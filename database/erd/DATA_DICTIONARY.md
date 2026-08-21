# WMS data dictionary

Generated from `database/wms_schema.sql`: **150 tables**, **1536 columns**, and **622 foreign-key relationships**.

Legend: PK = primary key, FK = foreign key, UK = single-column unique key, NN = not null.

## Security & access

### `account_status`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `account_status_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(30)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `allows_login` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |

### `authentication_policy`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `authentication_policy_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `max_failed_attempts` | `integer` | N |  |  |
| `lockout_seconds` | `integer` | N |  |  |
| `session_ttl_seconds` | `integer` | N |  |  |
| `is_default` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |

### `session_revocation_reason`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `session_revocation_reason_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

### `app_module`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `module_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(50)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `display_order` | `integer` | N |  | `0` |
| `is_active` | `boolean` | N |  | `true` |

### `app_account`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `account_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `username` | `varchar(100)` | N | UK |  |
| `email` | `varchar(254)` | Y | UK |  |
| `display_name` | `varchar(150)` | N |  |  |
| `password_hash` | `text` | Y |  |  |
| `external_subject` | `varchar(255)` | Y | UK |  |
| `account_status_id` | `uuid` | N | FK |  |
| `authentication_policy_id` | `uuid` | Y | FK |  |
| `preferred_timezone` | `varchar(50)` | Y |  |  |
| `failed_login_count` | `integer` | N |  | `0` |
| `locked_until` | `timestamptz` | Y |  |  |
| `last_login_at` | `timestamptz` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | Y | FK |  |
| `version_no` | `integer` | N |  | `1` |

Relationships:
- `app_account(account_status_id)` → `account_status(account_status_id)`
- `app_account(authentication_policy_id)` → `authentication_policy(authentication_policy_id)` (optional)
- `app_account(created_by)` → `app_account(account_id)` (optional)
- `app_account(updated_by)` → `app_account(account_id)` (optional)

### `app_session`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `session_id` | `varchar(120)` | N | PK |  |
| `account_id` | `uuid` | N | FK |  |
| `token_hash` | `varchar(128)` | N | UK |  |
| `issued_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `expires_at` | `timestamptz` | N |  |  |
| `last_seen_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `revoked_at` | `timestamptz` | Y |  |  |
| `session_revocation_reason_id` | `uuid` | Y | FK |  |
| `ip_address` | `inet` | Y |  |  |
| `user_agent` | `text` | Y |  |  |

Relationships:
- `app_session(account_id)` → `app_account(account_id)`; ON DELETE CASCADE
- `app_session(session_revocation_reason_id)` → `session_revocation_reason(session_revocation_reason_id)` (optional)

### `app_role`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `role_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(50)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |

Relationships:
- `app_role(created_by)` → `app_account(account_id)` (optional)

### `app_permission`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `permission_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(100)` | N | UK |  |
| `name` | `varchar(150)` | N |  |  |
| `module_code` | `varchar(50)` | N | FK |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |

Relationships:
- `app_permission(module_code)` → `app_module(code)`

### `app_menu`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `menu_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `parent_menu_id` | `uuid` | Y | FK |  |
| `code` | `varchar(50)` | N | UK |  |
| `label` | `varchar(100)` | N |  |  |
| `route` | `varchar(255)` | Y |  |  |
| `icon_name` | `varchar(100)` | Y |  |  |
| `display_order` | `integer` | N |  | `0` |
| `require_all_permissions` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |

Relationships:
- `app_menu(parent_menu_id)` → `app_menu(menu_id)` (optional)

### `account_role`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `account_id` | `uuid` | N | PK, FK |  |
| `role_id` | `uuid` | N | PK, FK |  |
| `valid_from` | `timestamptz` | N |  | `clock_timestamp()` |
| `valid_until` | `timestamptz` | Y |  |  |
| `assigned_by` | `uuid` | Y | FK |  |

Relationships:
- `account_role(account_id)` → `app_account(account_id)`; ON DELETE CASCADE
- `account_role(role_id)` → `app_role(role_id)`; ON DELETE CASCADE
- `account_role(assigned_by)` → `app_account(account_id)` (optional)

### `role_permission`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `role_id` | `uuid` | N | PK, FK |  |
| `permission_id` | `uuid` | N | PK, FK |  |
| `granted_by` | `uuid` | Y | FK |  |
| `granted_at` | `timestamptz` | N |  | `clock_timestamp()` |

Relationships:
- `role_permission(role_id)` → `app_role(role_id)`; ON DELETE CASCADE
- `role_permission(permission_id)` → `app_permission(permission_id)`; ON DELETE CASCADE
- `role_permission(granted_by)` → `app_account(account_id)` (optional)

### `menu_permission`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `menu_id` | `uuid` | N | PK, FK |  |
| `permission_id` | `uuid` | N | PK, FK |  |

Relationships:
- `menu_permission(menu_id)` → `app_menu(menu_id)`; ON DELETE CASCADE
- `menu_permission(permission_id)` → `app_permission(permission_id)`; ON DELETE CASCADE

## Master data

### `organization`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `organization_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(150)` | N |  |  |
| `legal_name` | `varchar(200)` | Y |  |  |
| `tax_number` | `varchar(100)` | Y |  |  |
| `timezone_name` | `varchar(50)` | N |  |  |
| `address_line_1` | `varchar(255)` | Y |  |  |
| `address_line_2` | `varchar(255)` | Y |  |  |
| `city` | `varchar(100)` | Y |  |  |
| `province` | `varchar(100)` | Y |  |  |
| `postal_code` | `varchar(20)` | Y |  |  |
| `country_code` | `varchar(2)` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | Y | FK |  |

Relationships:
- `organization(created_by)` → `app_account(account_id)` (optional)
- `organization(updated_by)` → `app_account(account_id)` (optional)

### `warehouse`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `warehouse_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `operator_id` | `uuid` | N | FK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(150)` | N |  |  |
| `timezone_name` | `varchar(50)` | N |  |  |
| `address_line_1` | `varchar(255)` | Y |  |  |
| `address_line_2` | `varchar(255)` | Y |  |  |
| `city` | `varchar(100)` | Y |  |  |
| `province` | `varchar(100)` | Y |  |  |
| `postal_code` | `varchar(20)` | Y |  |  |
| `country_code` | `varchar(2)` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | Y | FK |  |

Relationships:
- `warehouse(operator_id)` → `organization(organization_id)`
- `warehouse(created_by)` → `app_account(account_id)` (optional)
- `warehouse(updated_by)` → `app_account(account_id)` (optional)

### `warehouse_owner`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `warehouse_id` | `uuid` | N | PK, FK |  |
| `owner_id` | `uuid` | N | PK, FK |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |

Relationships:
- `warehouse_owner(warehouse_id)` → `warehouse(warehouse_id)`
- `warehouse_owner(owner_id)` → `organization(organization_id)`
- `warehouse_owner(created_by)` → `app_account(account_id)` (optional)

### `location_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `location_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `allows_receiving` | `boolean` | N |  | `false` |
| `allows_storage` | `boolean` | N |  | `false` |
| `allows_picking` | `boolean` | N |  | `false` |
| `allows_shipping` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

### `warehouse_zone`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `zone_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `warehouse_id` | `uuid` | N | FK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |

Relationships:
- `warehouse_zone(warehouse_id)` → `warehouse(warehouse_id)`
- `warehouse_zone(created_by)` → `app_account(account_id)` (optional)

### `warehouse_location`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `location_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `warehouse_id` | `uuid` | N | FK |  |
| `zone_id` | `uuid` | N | FK |  |
| `location_type_id` | `uuid` | N | FK |  |
| `code` | `varchar(60)` | N |  |  |
| `barcode` | `varchar(100)` | Y |  |  |
| `aisle` | `varchar(20)` | Y |  |  |
| `bay` | `varchar(20)` | Y |  |  |
| `level_no` | `varchar(20)` | Y |  |  |
| `position_no` | `varchar(20)` | Y |  |  |
| `pick_sequence` | `integer` | N |  | `0` |
| `max_weight` | `numeric(20,6)` | Y |  |  |
| `max_volume` | `numeric(20,6)` | Y |  |  |
| `is_pick_face` | `boolean` | N |  | `false` |
| `is_locked` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |

Relationships:
- `warehouse_location(warehouse_id)` → `warehouse(warehouse_id)`
- `warehouse_location(location_type_id)` → `location_type(location_type_id)`
- `warehouse_location(created_by)` → `app_account(account_id)` (optional)
- `warehouse_location(zone_id, warehouse_id)` → `warehouse_zone(zone_id, warehouse_id)`

### `account_owner_access`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `account_id` | `uuid` | N | PK, FK |  |
| `owner_id` | `uuid` | N | PK, FK |  |
| `granted_by` | `uuid` | Y | FK |  |
| `granted_at` | `timestamptz` | N |  | `clock_timestamp()` |

Relationships:
- `account_owner_access(account_id)` → `app_account(account_id)`; ON DELETE CASCADE
- `account_owner_access(owner_id)` → `organization(organization_id)`; ON DELETE CASCADE
- `account_owner_access(granted_by)` → `app_account(account_id)` (optional)

### `account_warehouse_access`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `account_id` | `uuid` | N | PK, FK |  |
| `warehouse_id` | `uuid` | N | PK, FK |  |
| `granted_by` | `uuid` | Y | FK |  |
| `granted_at` | `timestamptz` | N |  | `clock_timestamp()` |

Relationships:
- `account_warehouse_access(account_id)` → `app_account(account_id)`; ON DELETE CASCADE
- `account_warehouse_access(warehouse_id)` → `warehouse(warehouse_id)`; ON DELETE CASCADE
- `account_warehouse_access(granted_by)` → `app_account(account_id)` (optional)

### `partner_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `partner_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

### `business_partner`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `partner_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `owner_id` | `uuid` | N | FK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(150)` | N |  |  |
| `legal_name` | `varchar(200)` | Y |  |  |
| `tax_number` | `varchar(100)` | Y |  |  |
| `email` | `varchar(254)` | Y |  |  |
| `phone` | `varchar(50)` | Y |  |  |
| `address_line_1` | `varchar(255)` | Y |  |  |
| `address_line_2` | `varchar(255)` | Y |  |  |
| `city` | `varchar(100)` | Y |  |  |
| `province` | `varchar(100)` | Y |  |  |
| `postal_code` | `varchar(20)` | Y |  |  |
| `country_code` | `varchar(2)` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | Y | FK |  |

Relationships:
- `business_partner(owner_id)` → `organization(organization_id)`
- `business_partner(created_by)` → `app_account(account_id)` (optional)
- `business_partner(updated_by)` → `app_account(account_id)` (optional)

### `business_partner_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `partner_id` | `uuid` | N | PK, FK |  |
| `partner_type_id` | `uuid` | N | PK, FK |  |

Relationships:
- `business_partner_type(partner_id)` → `business_partner(partner_id)`; ON DELETE CASCADE
- `business_partner_type(partner_type_id)` → `partner_type(partner_type_id)`

### `uom`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `uom_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(20)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `decimal_scale` | `smallint` | N |  | `0` |
| `is_active` | `boolean` | N |  | `true` |

### `item_category`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `category_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `owner_id` | `uuid` | N | FK |  |
| `parent_category_id` | `uuid` | Y | FK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(100)` | N |  |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `item_category(owner_id)` → `organization(organization_id)`
- `item_category(owner_id, parent_category_id)` → `item_category(owner_id, category_id)` (optional)

### `item`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `item_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `owner_id` | `uuid` | N | FK |  |
| `category_id` | `uuid` | Y | FK |  |
| `code` | `varchar(60)` | N |  |  |
| `name` | `varchar(200)` | N |  |  |
| `description` | `text` | Y |  |  |
| `base_uom_id` | `uuid` | N | FK |  |
| `weight` | `numeric(20,6)` | Y |  |  |
| `volume` | `numeric(20,6)` | Y |  |  |
| `lot_controlled` | `boolean` | N |  | `false` |
| `serial_controlled` | `boolean` | N |  | `false` |
| `shelf_life_days` | `integer` | Y |  |  |
| `minimum_receive_days` | `integer` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | Y | FK |  |

Relationships:
- `item(owner_id)` → `organization(organization_id)`
- `item(base_uom_id)` → `uom(uom_id)`
- `item(created_by)` → `app_account(account_id)` (optional)
- `item(updated_by)` → `app_account(account_id)` (optional)
- `item(owner_id, category_id)` → `item_category(owner_id, category_id)` (optional)

### `item_uom`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `item_uom_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `item_id` | `uuid` | N | FK |  |
| `uom_id` | `uuid` | N | FK |  |
| `conversion_to_base` | `numeric(20,6)` | N |  |  |
| `length` | `numeric(20,6)` | Y |  |  |
| `width` | `numeric(20,6)` | Y |  |  |
| `height` | `numeric(20,6)` | Y |  |  |
| `weight` | `numeric(20,6)` | Y |  |  |
| `is_receiving_uom` | `boolean` | N |  | `true` |
| `is_picking_uom` | `boolean` | N |  | `true` |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `item_uom(item_id)` → `item(item_id)`; ON DELETE CASCADE
- `item_uom(uom_id)` → `uom(uom_id)`

### `item_barcode`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `item_barcode_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `item_id` | `uuid` | N | FK |  |
| `uom_id` | `uuid` | Y | FK |  |
| `barcode` | `varchar(100)` | N | UK |  |
| `is_primary` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `item_barcode(item_id)` → `item(item_id)`; ON DELETE CASCADE
- `item_barcode(uom_id)` → `uom(uom_id)` (optional)

### `inventory_status`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `inventory_status_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_allocatable` | `boolean` | N |  | `false` |
| `is_pickable` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

### `quality_status`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `quality_status_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

### `inspection_result`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `inspection_result_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_accepted` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

### `quarantine_disposition_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `quarantine_disposition_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `releases_to_available` | `boolean` | N |  | `false` |
| `requires_reinspection` | `boolean` | N |  | `false` |
| `removes_inventory` | `boolean` | N |  | `false` |
| `removal_movement_type_id` | `uuid` | Y | FK |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `quarantine_disposition_type(removal_movement_type_id)` → `movement_type(movement_type_id)` (optional)

### `reason_code`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `reason_code_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `module_code` | `varchar(50)` | N | FK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `requires_note` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `reason_code(module_code)` → `app_module(code)`

### `handling_unit_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `handling_unit_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `max_weight` | `numeric(20,6)` | Y |  |  |
| `max_volume` | `numeric(20,6)` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

## Workflow & configuration

### `document_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `document_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `module_code` | `varchar(50)` | N | FK |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `document_type(module_code)` → `app_module(code)`

### `document_status`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `status_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `document_type_id` | `uuid` | N | FK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_initial` | `boolean` | N |  | `false` |
| `is_final` | `boolean` | N |  | `false` |
| `is_cancelled` | `boolean` | N |  | `false` |
| `display_order` | `integer` | N |  | `0` |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `document_status(document_type_id)` → `document_type(document_type_id)`

### `document_status_transition`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `transition_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `document_type_id` | `uuid` | N | FK |  |
| `from_status_id` | `uuid` | N | FK |  |
| `to_status_id` | `uuid` | N | FK |  |
| `required_permission_id` | `uuid` | Y | FK |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `document_status_transition(document_type_id)` → `document_type(document_type_id)`
- `document_status_transition(required_permission_id)` → `app_permission(permission_id)` (optional)
- `document_status_transition(document_type_id, from_status_id)` → `document_status(document_type_id, status_id)`
- `document_status_transition(document_type_id, to_status_id)` → `document_status(document_type_id, status_id)`

### `task_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `task_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

### `task_status`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `task_status_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `is_initial` | `boolean` | N |  | `false` |
| `is_final` | `boolean` | N |  | `false` |
| `is_cancelled` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

### `task_status_transition`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `task_status_transition_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `from_status_id` | `uuid` | N | FK |  |
| `to_status_id` | `uuid` | N | FK |  |
| `required_permission_id` | `uuid` | Y | FK |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `task_status_transition(from_status_id)` → `task_status(task_status_id)`
- `task_status_transition(to_status_id)` → `task_status(task_status_id)`
- `task_status_transition(required_permission_id)` → `app_permission(permission_id)` (optional)

### `task_priority`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `task_priority_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `priority_value` | `integer` | N | UK |  |
| `is_active` | `boolean` | N |  | `true` |

### `putaway_strategy`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `putaway_strategy_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `owner_id` | `uuid` | Y | FK |  |
| `warehouse_id` | `uuid` | Y | FK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `putaway_strategy(owner_id)` → `organization(organization_id)` (optional)
- `putaway_strategy(warehouse_id)` → `warehouse(warehouse_id)` (optional)

### `putaway_strategy_rule`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `rule_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `putaway_strategy_id` | `uuid` | N | FK |  |
| `sequence_no` | `integer` | N |  |  |
| `category_id` | `uuid` | Y | FK |  |
| `location_type_id` | `uuid` | Y | FK |  |
| `zone_id` | `uuid` | Y | FK |  |
| `minimum_empty_percent` | `numeric(7,4)` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `putaway_strategy_rule(putaway_strategy_id)` → `putaway_strategy(putaway_strategy_id)`; ON DELETE CASCADE
- `putaway_strategy_rule(category_id)` → `item_category(category_id)` (optional)
- `putaway_strategy_rule(location_type_id)` → `location_type(location_type_id)` (optional)
- `putaway_strategy_rule(zone_id)` → `warehouse_zone(zone_id)` (optional)

### `picking_strategy`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `picking_strategy_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `owner_id` | `uuid` | Y | FK |  |
| `warehouse_id` | `uuid` | Y | FK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `picking_strategy(owner_id)` → `organization(organization_id)` (optional)
- `picking_strategy(warehouse_id)` → `warehouse(warehouse_id)` (optional)

### `picking_sort_method`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `picking_sort_method_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

### `picking_strategy_rule`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `rule_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `picking_strategy_id` | `uuid` | N | FK |  |
| `sequence_no` | `integer` | N |  |  |
| `inventory_status_id` | `uuid` | Y | FK |  |
| `zone_id` | `uuid` | Y | FK |  |
| `picking_sort_method_id` | `uuid` | N | FK |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `picking_strategy_rule(picking_strategy_id)` → `picking_strategy(picking_strategy_id)`; ON DELETE CASCADE
- `picking_strategy_rule(inventory_status_id)` → `inventory_status(inventory_status_id)` (optional)
- `picking_strategy_rule(zone_id)` → `warehouse_zone(zone_id)` (optional)
- `picking_strategy_rule(picking_sort_method_id)` → `picking_sort_method(picking_sort_method_id)`

### `document_number_rule`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `document_number_rule_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `document_type_id` | `uuid` | N | FK |  |
| `prefix` | `varchar(20)` | N |  |  |
| `separator` | `varchar(3)` | N |  | `'-'` |
| `sequence_length` | `smallint` | N |  | `6` |
| `include_partner_code` | `boolean` | N |  | `true` |
| `include_warehouse_code` | `boolean` | N |  | `true` |
| `is_active` | `boolean` | N |  | `true` |
| `effective_from` | `date` | N |  | `current_date` |
| `effective_until` | `date` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |

Relationships:
- `document_number_rule(document_type_id)` → `document_type(document_type_id)`
- `document_number_rule(created_by)` → `app_account(account_id)` (optional)

### `document_daily_counter`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `document_type_id` | `uuid` | N | PK, FK |  |
| `business_date` | `date` | N | PK |  |
| `last_number` | `bigint` | N |  |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |

Relationships:
- `document_daily_counter(document_type_id)` → `document_type(document_type_id)`

## Inventory identity

### `inventory_lot`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `lot_id` | `varchar(120)` | N | PK |  |
| `owner_id` | `uuid` | N | FK |  |
| `item_id` | `uuid` | N | FK |  |
| `lot_number` | `varchar(100)` | N |  |  |
| `manufacture_date` | `date` | Y |  |  |
| `expiry_date` | `date` | Y |  |  |
| `quality_status_id` | `uuid` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |

Relationships:
- `inventory_lot(owner_id)` → `organization(organization_id)`
- `inventory_lot(item_id)` → `item(item_id)`
- `inventory_lot(quality_status_id)` → `quality_status(quality_status_id)` (optional)
- `inventory_lot(created_by)` → `app_account(account_id)` (optional)

### `serial_number`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `serial_id` | `varchar(160)` | N | PK |  |
| `owner_id` | `uuid` | N | FK |  |
| `item_id` | `uuid` | N | FK |  |
| `serial_no` | `varchar(120)` | N |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |

Relationships:
- `serial_number(owner_id)` → `organization(organization_id)`
- `serial_number(item_id)` → `item(item_id)`
- `serial_number(created_by)` → `app_account(account_id)` (optional)

### `handling_unit`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `handling_unit_id` | `varchar(120)` | N | PK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `handling_unit_type_id` | `uuid` | N | FK |  |
| `parent_handling_unit_id` | `varchar(120)` | Y | FK |  |
| `current_location_id` | `uuid` | Y | FK |  |
| `barcode` | `varchar(120)` | N | UK |  |
| `is_closed` | `boolean` | N |  | `false` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |

Relationships:
- `handling_unit(warehouse_id)` → `warehouse(warehouse_id)`
- `handling_unit(owner_id)` → `organization(organization_id)`
- `handling_unit(handling_unit_type_id)` → `handling_unit_type(handling_unit_type_id)`
- `handling_unit(parent_handling_unit_id)` → `handling_unit(handling_unit_id)` (optional)
- `handling_unit(current_location_id)` → `warehouse_location(location_id)` (optional)
- `handling_unit(created_by)` → `app_account(account_id)` (optional)
- `handling_unit(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

## Inbound

### `purchase_order`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `purchase_order_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `vendor_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `purchase_order_no` | `varchar(120)` | N |  |  |
| `ordered_at` | `timestamptz` | N |  |  |
| `expected_arrival_at` | `timestamptz` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | Y | FK |  |
| `version_no` | `integer` | N |  | `1` |

Relationships:
- `purchase_order(document_type_id)` → `document_type(document_type_id)`
- `purchase_order(owner_id)` → `organization(organization_id)`
- `purchase_order(vendor_id)` → `business_partner(partner_id)`
- `purchase_order(warehouse_id)` → `warehouse(warehouse_id)`
- `purchase_order(created_by)` → `app_account(account_id)`
- `purchase_order(updated_by)` → `app_account(account_id)` (optional)
- `purchase_order(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `purchase_order(owner_id, vendor_id)` → `business_partner(owner_id, partner_id)`
- `purchase_order(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `purchase_order_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `purchase_order_line_id` | `varchar(150)` | N | PK |  |
| `purchase_order_id` | `varchar(120)` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `item_id` | `uuid` | N | FK |  |
| `ordered_qty` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `vendor_item_code` | `varchar(100)` | Y |  |  |
| `expected_lot_no` | `varchar(100)` | Y |  |  |
| `expected_expiry_date` | `date` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `purchase_order_line(uom_id)` → `uom(uom_id)`
- `purchase_order_line(created_by)` → `app_account(account_id)`
- `purchase_order_line(purchase_order_id, owner_id)` → `purchase_order(purchase_order_id, owner_id)`; ON DELETE CASCADE
- `purchase_order_line(owner_id, item_id)` → `item(owner_id, item_id)`

### `inbound_order`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `inbound_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `vendor_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `expected_arrival_at` | `timestamptz` | Y |  |  |
| `external_reference` | `varchar(120)` | Y |  |  |
| `supplier_reference` | `varchar(120)` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | Y | FK |  |
| `version_no` | `integer` | N |  | `1` |

Relationships:
- `inbound_order(document_type_id)` → `document_type(document_type_id)`
- `inbound_order(owner_id)` → `organization(organization_id)`
- `inbound_order(vendor_id)` → `business_partner(partner_id)`
- `inbound_order(warehouse_id)` → `warehouse(warehouse_id)`
- `inbound_order(created_by)` → `app_account(account_id)`
- `inbound_order(updated_by)` → `app_account(account_id)` (optional)
- `inbound_order(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `inbound_order(owner_id, vendor_id)` → `business_partner(owner_id, partner_id)`
- `inbound_order(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `inbound_order_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `inbound_line_id` | `varchar(150)` | N | PK |  |
| `inbound_id` | `varchar(120)` | N | FK |  |
| `purchase_order_line_id` | `varchar(150)` | Y | FK |  |
| `line_no` | `integer` | N |  |  |
| `item_id` | `uuid` | N | FK |  |
| `expected_qty` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `expected_lot_no` | `varchar(100)` | Y |  |  |
| `expected_expiry_date` | `date` | Y |  |  |
| `customer_line_reference` | `varchar(100)` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `inbound_order_line(inbound_id)` → `inbound_order(inbound_id)`; ON DELETE CASCADE
- `inbound_order_line(purchase_order_line_id)` → `purchase_order_line(purchase_order_line_id)` (optional)
- `inbound_order_line(item_id)` → `item(item_id)`
- `inbound_order_line(uom_id)` → `uom(uom_id)`
- `inbound_order_line(created_by)` → `app_account(account_id)`

### `receipt`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `receipt_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `inbound_id` | `varchar(120)` | Y | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `received_at` | `timestamptz` | N |  |  |
| `dock_location_id` | `uuid` | Y | FK |  |
| `vehicle_number` | `varchar(60)` | Y |  |  |
| `seal_number` | `varchar(60)` | Y |  |  |
| `delivery_note_no` | `varchar(100)` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `receipt(document_type_id)` → `document_type(document_type_id)`
- `receipt(inbound_id)` → `inbound_order(inbound_id)` (optional)
- `receipt(owner_id)` → `organization(organization_id)`
- `receipt(warehouse_id)` → `warehouse(warehouse_id)`
- `receipt(dock_location_id)` → `warehouse_location(location_id)` (optional)
- `receipt(created_by)` → `app_account(account_id)`
- `receipt(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `receipt(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `receipt_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `receipt_line_id` | `varchar(150)` | N | PK |  |
| `receipt_id` | `varchar(120)` | N | FK |  |
| `inbound_line_id` | `varchar(150)` | Y | FK |  |
| `line_no` | `integer` | N |  |  |
| `item_id` | `uuid` | N | FK |  |
| `received_qty` | `numeric(20,6)` | N |  |  |
| `rejected_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `receipt_line(receipt_id)` → `receipt(receipt_id)`; ON DELETE CASCADE
- `receipt_line(inbound_line_id)` → `inbound_order_line(inbound_line_id)` (optional)
- `receipt_line(item_id)` → `item(item_id)`
- `receipt_line(uom_id)` → `uom(uom_id)`
- `receipt_line(created_by)` → `app_account(account_id)`

### `receipt_inventory`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `receipt_inventory_id` | `varchar(160)` | N | PK |  |
| `receipt_line_id` | `varchar(150)` | N | FK |  |
| `item_id` | `uuid` | N | FK |  |
| `source_qty` | `numeric(20,6)` | N |  |  |
| `source_uom_id` | `uuid` | N | FK |  |
| `base_qty` | `numeric(20,6)` | N |  |  |
| `base_uom_id` | `uuid` | N | FK |  |
| `lot_id` | `varchar(120)` | Y | FK |  |
| `handling_unit_id` | `varchar(120)` | Y | FK |  |
| `received_location_id` | `uuid` | N | FK |  |
| `initial_inventory_status_id` | `uuid` | N | FK |  |
| `initial_balance_id` | `varchar(160)` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `receipt_inventory(receipt_line_id)` → `receipt_line(receipt_line_id)`; ON DELETE CASCADE
- `receipt_inventory(item_id)` → `item(item_id)`
- `receipt_inventory(source_uom_id)` → `uom(uom_id)`
- `receipt_inventory(base_uom_id)` → `uom(uom_id)`
- `receipt_inventory(lot_id)` → `inventory_lot(lot_id)` (optional)
- `receipt_inventory(handling_unit_id)` → `handling_unit(handling_unit_id)` (optional)
- `receipt_inventory(received_location_id)` → `warehouse_location(location_id)`
- `receipt_inventory(initial_inventory_status_id)` → `inventory_status(inventory_status_id)`
- `receipt_inventory(created_by)` → `app_account(account_id)`
- `receipt_inventory(initial_balance_id)` → `inventory_balance(balance_id)` (optional)

### `receipt_line_serial`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `receipt_inventory_id` | `varchar(160)` | N | PK, FK |  |
| `serial_id` | `varchar(160)` | N | PK, FK, UK |  |

Relationships:
- `receipt_line_serial(receipt_inventory_id)` → `receipt_inventory(receipt_inventory_id)`; ON DELETE CASCADE
- `receipt_line_serial(serial_id)` → `serial_number(serial_id)`

### `quality_inspection`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `inspection_id` | `varchar(120)` | N | PK |  |
| `receipt_inventory_id` | `varchar(160)` | N | FK |  |
| `parent_inspection_id` | `varchar(120)` | Y | FK |  |
| `quality_status_id` | `uuid` | N | FK |  |
| `inspection_result_id` | `uuid` | Y | FK |  |
| `inspected_qty` | `numeric(20,6)` | N |  |  |
| `passed_qty` | `numeric(20,6)` | N |  | `0` |
| `failed_qty` | `numeric(20,6)` | N |  | `0` |
| `inspected_at` | `timestamptz` | Y |  |  |
| `inspected_by` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `quality_inspection(receipt_inventory_id)` → `receipt_inventory(receipt_inventory_id)`
- `quality_inspection(parent_inspection_id)` → `quality_inspection(inspection_id)` (optional)
- `quality_inspection(quality_status_id)` → `quality_status(quality_status_id)`
- `quality_inspection(inspection_result_id)` → `inspection_result(inspection_result_id)` (optional)
- `quality_inspection(inspected_by)` → `app_account(account_id)` (optional)
- `quality_inspection(created_by)` → `app_account(account_id)`

### `putaway_task`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `putaway_task_id` | `varchar(120)` | N | PK |  |
| `task_type_id` | `uuid` | N | FK |  |
| `task_status_id` | `uuid` | N | FK |  |
| `task_priority_id` | `uuid` | N | FK |  |
| `receipt_inventory_id` | `varchar(160)` | N | FK |  |
| `source_balance_id` | `varchar(160)` | Y | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `item_id` | `uuid` | N | FK |  |
| `lot_id` | `varchar(120)` | Y | FK |  |
| `handling_unit_id` | `varchar(120)` | Y | FK |  |
| `source_location_id` | `uuid` | N | FK |  |
| `target_location_id` | `uuid` | N | FK |  |
| `planned_qty` | `numeric(20,6)` | N |  |  |
| `completed_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `assigned_to` | `uuid` | Y | FK |  |
| `started_at` | `timestamptz` | Y |  |  |
| `completed_at` | `timestamptz` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `putaway_task(task_type_id)` → `task_type(task_type_id)`
- `putaway_task(task_status_id)` → `task_status(task_status_id)`
- `putaway_task(task_priority_id)` → `task_priority(task_priority_id)`
- `putaway_task(receipt_inventory_id)` → `receipt_inventory(receipt_inventory_id)`
- `putaway_task(owner_id)` → `organization(organization_id)`
- `putaway_task(warehouse_id)` → `warehouse(warehouse_id)`
- `putaway_task(item_id)` → `item(item_id)`
- `putaway_task(lot_id)` → `inventory_lot(lot_id)` (optional)
- `putaway_task(handling_unit_id)` → `handling_unit(handling_unit_id)` (optional)
- `putaway_task(source_location_id)` → `warehouse_location(location_id)`
- `putaway_task(target_location_id)` → `warehouse_location(location_id)`
- `putaway_task(uom_id)` → `uom(uom_id)`
- `putaway_task(assigned_to)` → `app_account(account_id)` (optional)
- `putaway_task(created_by)` → `app_account(account_id)`
- `putaway_task(source_balance_id)` → `inventory_balance(balance_id)` (optional)

### `quarantine_case`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `quarantine_case_id` | `varchar(140)` | N | PK |  |
| `parent_quarantine_case_id` | `varchar(140)` | Y | FK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `receipt_inventory_id` | `varchar(160)` | N | FK |  |
| `inspection_id` | `varchar(120)` | N | FK, UK |  |
| `quarantine_balance_id` | `varchar(160)` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `quarantine_qty` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `opened_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `closed_at` | `timestamptz` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | Y | FK |  |
| `version_no` | `integer` | N |  | `1` |

Relationships:
- `quarantine_case(parent_quarantine_case_id)` → `quarantine_case(quarantine_case_id)` (optional)
- `quarantine_case(document_type_id)` → `document_type(document_type_id)`
- `quarantine_case(receipt_inventory_id)` → `receipt_inventory(receipt_inventory_id)`
- `quarantine_case(inspection_id)` → `quality_inspection(inspection_id)`
- `quarantine_case(quarantine_balance_id)` → `inventory_balance(balance_id)`
- `quarantine_case(owner_id)` → `organization(organization_id)`
- `quarantine_case(warehouse_id)` → `warehouse(warehouse_id)`
- `quarantine_case(uom_id)` → `uom(uom_id)`
- `quarantine_case(created_by)` → `app_account(account_id)`
- `quarantine_case(updated_by)` → `app_account(account_id)` (optional)
- `quarantine_case(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `quarantine_case(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `quarantine_disposition`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `quarantine_disposition_id` | `varchar(150)` | N | PK |  |
| `quarantine_case_id` | `varchar(140)` | N | FK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `quarantine_disposition_type_id` | `uuid` | N | FK |  |
| `disposition_qty` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `client_decision_reference` | `varchar(120)` | Y |  |  |
| `decision_notes` | `text` | Y |  |  |
| `decided_at` | `timestamptz` | N |  |  |
| `decided_by` | `uuid` | N | FK |  |
| `processed_at` | `timestamptz` | Y |  |  |
| `inventory_movement_id` | `varchar(140)` | Y | FK, UK |  |
| `resulting_balance_id` | `varchar(160)` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `quarantine_disposition(quarantine_case_id)` → `quarantine_case(quarantine_case_id)`
- `quarantine_disposition(document_type_id)` → `document_type(document_type_id)`
- `quarantine_disposition(quarantine_disposition_type_id)` → `quarantine_disposition_type(quarantine_disposition_type_id)`
- `quarantine_disposition(uom_id)` → `uom(uom_id)`
- `quarantine_disposition(decided_by)` → `app_account(account_id)`
- `quarantine_disposition(inventory_movement_id)` → `inventory_movement(movement_id)` (optional)
- `quarantine_disposition(resulting_balance_id)` → `inventory_balance(balance_id)` (optional)
- `quarantine_disposition(created_by)` → `app_account(account_id)`
- `quarantine_disposition(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `rework_task`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `rework_task_id` | `varchar(140)` | N | PK |  |
| `quarantine_disposition_id` | `varchar(150)` | N | FK, UK |  |
| `task_type_id` | `uuid` | N | FK |  |
| `task_status_id` | `uuid` | N | FK |  |
| `task_priority_id` | `uuid` | N | FK |  |
| `planned_qty` | `numeric(20,6)` | N |  |  |
| `completed_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `assigned_to` | `uuid` | Y | FK |  |
| `work_instructions` | `text` | Y |  |  |
| `result_notes` | `text` | Y |  |  |
| `started_at` | `timestamptz` | Y |  |  |
| `completed_at` | `timestamptz` | Y |  |  |
| `reinspection_id` | `varchar(120)` | Y | FK, UK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `rework_task(quarantine_disposition_id)` → `quarantine_disposition(quarantine_disposition_id)`
- `rework_task(task_type_id)` → `task_type(task_type_id)`
- `rework_task(task_status_id)` → `task_status(task_status_id)`
- `rework_task(task_priority_id)` → `task_priority(task_priority_id)`
- `rework_task(uom_id)` → `uom(uom_id)`
- `rework_task(assigned_to)` → `app_account(account_id)` (optional)
- `rework_task(reinspection_id)` → `quality_inspection(inspection_id)` (optional)
- `rework_task(created_by)` → `app_account(account_id)`

## Stock control

### `movement_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `movement_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

### `internal_move_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `internal_move_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `requires_approval` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

### `inventory_adjustment_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `inventory_adjustment_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `quantity_effect` | `smallint` | N |  |  |
| `requires_approval` | `boolean` | N |  | `true` |
| `is_active` | `boolean` | N |  | `true` |

### `stock_count_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `stock_count_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `requires_freeze` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

### `inventory_balance`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `balance_id` | `varchar(160)` | N | PK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `location_id` | `uuid` | N | FK |  |
| `item_id` | `uuid` | N | FK |  |
| `lot_id` | `varchar(120)` | Y | FK |  |
| `handling_unit_id` | `varchar(120)` | Y | FK |  |
| `inventory_status_id` | `uuid` | N | FK |  |
| `on_hand_qty` | `numeric(20,6)` | N |  | `0` |
| `reserved_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `version_no` | `bigint` | N |  | `1` |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |

Relationships:
- `inventory_balance(owner_id)` → `organization(organization_id)`
- `inventory_balance(warehouse_id)` → `warehouse(warehouse_id)`
- `inventory_balance(location_id)` → `warehouse_location(location_id)`
- `inventory_balance(item_id)` → `item(item_id)`
- `inventory_balance(lot_id)` → `inventory_lot(lot_id)` (optional)
- `inventory_balance(handling_unit_id)` → `handling_unit(handling_unit_id)` (optional)
- `inventory_balance(inventory_status_id)` → `inventory_status(inventory_status_id)`
- `inventory_balance(uom_id)` → `uom(uom_id)`
- `inventory_balance(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `inventory_movement`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `movement_id` | `varchar(140)` | N | PK |  |
| `movement_type_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `occurred_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `item_id` | `uuid` | N | FK |  |
| `lot_id` | `varchar(120)` | Y | FK |  |
| `serial_id` | `varchar(160)` | Y | FK |  |
| `handling_unit_id` | `varchar(120)` | Y | FK |  |
| `from_location_id` | `uuid` | Y | FK |  |
| `to_location_id` | `uuid` | Y | FK |  |
| `from_status_id` | `uuid` | Y | FK |  |
| `to_status_id` | `uuid` | Y | FK |  |
| `quantity` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `source_document_id` | `varchar(140)` | N |  |  |
| `source_line_id` | `varchar(160)` | Y |  |  |
| `reason_code_id` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `inventory_movement(movement_type_id)` → `movement_type(movement_type_id)`
- `inventory_movement(owner_id)` → `organization(organization_id)`
- `inventory_movement(warehouse_id)` → `warehouse(warehouse_id)`
- `inventory_movement(item_id)` → `item(item_id)`
- `inventory_movement(lot_id)` → `inventory_lot(lot_id)` (optional)
- `inventory_movement(serial_id)` → `serial_number(serial_id)` (optional)
- `inventory_movement(handling_unit_id)` → `handling_unit(handling_unit_id)` (optional)
- `inventory_movement(from_location_id)` → `warehouse_location(location_id)` (optional)
- `inventory_movement(to_location_id)` → `warehouse_location(location_id)` (optional)
- `inventory_movement(from_status_id)` → `inventory_status(inventory_status_id)` (optional)
- `inventory_movement(to_status_id)` → `inventory_status(inventory_status_id)` (optional)
- `inventory_movement(uom_id)` → `uom(uom_id)`
- `inventory_movement(reason_code_id)` → `reason_code(reason_code_id)` (optional)
- `inventory_movement(created_by)` → `app_account(account_id)`
- `inventory_movement(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `internal_move_order`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `internal_move_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `internal_move_type_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `reason_code_id` | `uuid` | Y | FK |  |
| `requested_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `approved_at` | `timestamptz` | Y |  |  |
| `approved_by` | `uuid` | Y | FK |  |
| `completed_at` | `timestamptz` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `version_no` | `bigint` | N |  | `1` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | N | FK |  |

Relationships:
- `internal_move_order(document_type_id)` → `document_type(document_type_id)`
- `internal_move_order(internal_move_type_id)` → `internal_move_type(internal_move_type_id)`
- `internal_move_order(owner_id)` → `organization(organization_id)`
- `internal_move_order(warehouse_id)` → `warehouse(warehouse_id)`
- `internal_move_order(reason_code_id)` → `reason_code(reason_code_id)` (optional)
- `internal_move_order(approved_by)` → `app_account(account_id)` (optional)
- `internal_move_order(created_by)` → `app_account(account_id)`
- `internal_move_order(updated_by)` → `app_account(account_id)`
- `internal_move_order(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `internal_move_order(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `internal_move_order_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `internal_move_line_id` | `varchar(150)` | N | PK |  |
| `internal_move_id` | `varchar(120)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `source_balance_id` | `varchar(160)` | N | FK |  |
| `target_location_id` | `uuid` | N | FK |  |
| `planned_qty` | `numeric(20,6)` | N |  |  |
| `completed_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `assigned_to` | `uuid` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `internal_move_order_line(internal_move_id)` → `internal_move_order(internal_move_id)`; ON DELETE CASCADE
- `internal_move_order_line(source_balance_id)` → `inventory_balance(balance_id)`
- `internal_move_order_line(target_location_id)` → `warehouse_location(location_id)`
- `internal_move_order_line(uom_id)` → `uom(uom_id)`
- `internal_move_order_line(assigned_to)` → `app_account(account_id)` (optional)
- `internal_move_order_line(created_by)` → `app_account(account_id)`

### `internal_move_execution`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `internal_move_execution_id` | `varchar(160)` | N | PK |  |
| `internal_move_line_id` | `varchar(150)` | N | FK |  |
| `quantity` | `numeric(20,6)` | N |  |  |
| `source_balance_id` | `varchar(160)` | N | FK |  |
| `destination_balance_id` | `varchar(160)` | N | FK |  |
| `movement_id` | `varchar(140)` | N | FK, UK |  |
| `executed_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `executed_by` | `uuid` | N | FK |  |

Relationships:
- `internal_move_execution(internal_move_line_id)` → `internal_move_order_line(internal_move_line_id)`
- `internal_move_execution(source_balance_id)` → `inventory_balance(balance_id)`
- `internal_move_execution(destination_balance_id)` → `inventory_balance(balance_id)`
- `internal_move_execution(movement_id)` → `inventory_movement(movement_id)`
- `internal_move_execution(executed_by)` → `app_account(account_id)`

### `transfer_order`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `transfer_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `source_warehouse_id` | `uuid` | N | FK |  |
| `target_warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `requested_transfer_at` | `timestamptz` | Y |  |  |
| `approved_at` | `timestamptz` | Y |  |  |
| `approved_by` | `uuid` | Y | FK |  |
| `completed_at` | `timestamptz` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `version_no` | `bigint` | N |  | `1` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | N | FK |  |

Relationships:
- `transfer_order(document_type_id)` → `document_type(document_type_id)`
- `transfer_order(owner_id)` → `organization(organization_id)`
- `transfer_order(source_warehouse_id)` → `warehouse(warehouse_id)`
- `transfer_order(target_warehouse_id)` → `warehouse(warehouse_id)`
- `transfer_order(approved_by)` → `app_account(account_id)` (optional)
- `transfer_order(created_by)` → `app_account(account_id)`
- `transfer_order(updated_by)` → `app_account(account_id)`
- `transfer_order(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `transfer_order(owner_id, source_warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`
- `transfer_order(owner_id, target_warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `transfer_order_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `transfer_line_id` | `varchar(150)` | N | PK |  |
| `transfer_id` | `varchar(120)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `item_id` | `uuid` | N | FK |  |
| `requested_qty` | `numeric(20,6)` | N |  |  |
| `dispatched_qty` | `numeric(20,6)` | N |  | `0` |
| `received_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `requested_lot_id` | `varchar(120)` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `transfer_order_line(transfer_id)` → `transfer_order(transfer_id)`; ON DELETE CASCADE
- `transfer_order_line(item_id)` → `item(item_id)`
- `transfer_order_line(uom_id)` → `uom(uom_id)`
- `transfer_order_line(requested_lot_id)` → `inventory_lot(lot_id)` (optional)
- `transfer_order_line(created_by)` → `app_account(account_id)`

### `transfer_dispatch`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `transfer_dispatch_id` | `varchar(120)` | N | PK |  |
| `transfer_id` | `varchar(120)` | N | FK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `carrier_partner_id` | `uuid` | Y | FK |  |
| `tracking_number` | `varchar(150)` | Y |  |  |
| `vehicle_number` | `varchar(60)` | Y |  |  |
| `seal_number` | `varchar(60)` | Y |  |  |
| `dispatched_at` | `timestamptz` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `transfer_dispatch(transfer_id)` → `transfer_order(transfer_id)`
- `transfer_dispatch(document_type_id)` → `document_type(document_type_id)`
- `transfer_dispatch(carrier_partner_id)` → `business_partner(partner_id)` (optional)
- `transfer_dispatch(created_by)` → `app_account(account_id)`
- `transfer_dispatch(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `transfer_dispatch_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `transfer_dispatch_line_id` | `varchar(150)` | N | PK |  |
| `transfer_dispatch_id` | `varchar(120)` | N | FK |  |
| `transfer_line_id` | `varchar(150)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `source_balance_id` | `varchar(160)` | N | FK |  |
| `dispatched_qty` | `numeric(20,6)` | N |  |  |
| `received_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `movement_id` | `varchar(140)` | Y | FK, UK |  |
| `confirmed_at` | `timestamptz` | Y |  |  |
| `confirmed_by` | `uuid` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `transfer_dispatch_line(transfer_dispatch_id)` → `transfer_dispatch(transfer_dispatch_id)`; ON DELETE CASCADE
- `transfer_dispatch_line(transfer_line_id)` → `transfer_order_line(transfer_line_id)`
- `transfer_dispatch_line(source_balance_id)` → `inventory_balance(balance_id)`
- `transfer_dispatch_line(uom_id)` → `uom(uom_id)`
- `transfer_dispatch_line(movement_id)` → `inventory_movement(movement_id)` (optional)
- `transfer_dispatch_line(confirmed_by)` → `app_account(account_id)` (optional)
- `transfer_dispatch_line(created_by)` → `app_account(account_id)`

### `transfer_receipt`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `transfer_receipt_id` | `varchar(120)` | N | PK |  |
| `transfer_id` | `varchar(120)` | N | FK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `received_at` | `timestamptz` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `transfer_receipt(transfer_id)` → `transfer_order(transfer_id)`
- `transfer_receipt(document_type_id)` → `document_type(document_type_id)`
- `transfer_receipt(warehouse_id)` → `warehouse(warehouse_id)`
- `transfer_receipt(created_by)` → `app_account(account_id)`
- `transfer_receipt(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `transfer_receipt_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `transfer_receipt_line_id` | `varchar(150)` | N | PK |  |
| `transfer_receipt_id` | `varchar(120)` | N | FK |  |
| `transfer_dispatch_line_id` | `varchar(150)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `target_location_id` | `uuid` | N | FK |  |
| `received_qty` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `destination_balance_id` | `varchar(160)` | Y | FK |  |
| `movement_id` | `varchar(140)` | Y | FK, UK |  |
| `posted_at` | `timestamptz` | Y |  |  |
| `posted_by` | `uuid` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `transfer_receipt_line(transfer_receipt_id)` → `transfer_receipt(transfer_receipt_id)`; ON DELETE CASCADE
- `transfer_receipt_line(transfer_dispatch_line_id)` → `transfer_dispatch_line(transfer_dispatch_line_id)`
- `transfer_receipt_line(target_location_id)` → `warehouse_location(location_id)`
- `transfer_receipt_line(uom_id)` → `uom(uom_id)`
- `transfer_receipt_line(destination_balance_id)` → `inventory_balance(balance_id)` (optional)
- `transfer_receipt_line(movement_id)` → `inventory_movement(movement_id)` (optional)
- `transfer_receipt_line(posted_by)` → `app_account(account_id)` (optional)
- `transfer_receipt_line(created_by)` → `app_account(account_id)`

### `inventory_status_change`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `inventory_status_change_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `reason_code_id` | `uuid` | N | FK |  |
| `approved_at` | `timestamptz` | Y |  |  |
| `approved_by` | `uuid` | Y | FK |  |
| `posted_at` | `timestamptz` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `version_no` | `bigint` | N |  | `1` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | N | FK |  |

Relationships:
- `inventory_status_change(document_type_id)` → `document_type(document_type_id)`
- `inventory_status_change(owner_id)` → `organization(organization_id)`
- `inventory_status_change(warehouse_id)` → `warehouse(warehouse_id)`
- `inventory_status_change(reason_code_id)` → `reason_code(reason_code_id)`
- `inventory_status_change(approved_by)` → `app_account(account_id)` (optional)
- `inventory_status_change(created_by)` → `app_account(account_id)`
- `inventory_status_change(updated_by)` → `app_account(account_id)`
- `inventory_status_change(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `inventory_status_change(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `inventory_status_change_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `inventory_status_change_line_id` | `varchar(150)` | N | PK |  |
| `inventory_status_change_id` | `varchar(120)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `source_balance_id` | `varchar(160)` | N | FK |  |
| `target_inventory_status_id` | `uuid` | N | FK |  |
| `quantity` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `destination_balance_id` | `varchar(160)` | Y | FK |  |
| `movement_id` | `varchar(140)` | Y | FK, UK |  |
| `posted_at` | `timestamptz` | Y |  |  |
| `posted_by` | `uuid` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `inventory_status_change_line(inventory_status_change_id)` → `inventory_status_change(inventory_status_change_id)`; ON DELETE CASCADE
- `inventory_status_change_line(source_balance_id)` → `inventory_balance(balance_id)`
- `inventory_status_change_line(target_inventory_status_id)` → `inventory_status(inventory_status_id)`
- `inventory_status_change_line(uom_id)` → `uom(uom_id)`
- `inventory_status_change_line(destination_balance_id)` → `inventory_balance(balance_id)` (optional)
- `inventory_status_change_line(movement_id)` → `inventory_movement(movement_id)` (optional)
- `inventory_status_change_line(posted_by)` → `app_account(account_id)` (optional)
- `inventory_status_change_line(created_by)` → `app_account(account_id)`

### `inventory_adjustment`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `inventory_adjustment_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `reason_code_id` | `uuid` | N | FK |  |
| `approved_at` | `timestamptz` | Y |  |  |
| `approved_by` | `uuid` | Y | FK |  |
| `posted_at` | `timestamptz` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `version_no` | `bigint` | N |  | `1` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | N | FK |  |

Relationships:
- `inventory_adjustment(document_type_id)` → `document_type(document_type_id)`
- `inventory_adjustment(owner_id)` → `organization(organization_id)`
- `inventory_adjustment(warehouse_id)` → `warehouse(warehouse_id)`
- `inventory_adjustment(reason_code_id)` → `reason_code(reason_code_id)`
- `inventory_adjustment(approved_by)` → `app_account(account_id)` (optional)
- `inventory_adjustment(created_by)` → `app_account(account_id)`
- `inventory_adjustment(updated_by)` → `app_account(account_id)`
- `inventory_adjustment(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `inventory_adjustment(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `inventory_adjustment_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `inventory_adjustment_line_id` | `varchar(150)` | N | PK |  |
| `inventory_adjustment_id` | `varchar(120)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `inventory_adjustment_type_id` | `uuid` | N | FK |  |
| `source_balance_id` | `varchar(160)` | Y | FK |  |
| `location_id` | `uuid` | N | FK |  |
| `item_id` | `uuid` | N | FK |  |
| `lot_id` | `varchar(120)` | Y | FK |  |
| `handling_unit_id` | `varchar(120)` | Y | FK |  |
| `inventory_status_id` | `uuid` | N | FK |  |
| `adjustment_qty` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `resulting_balance_id` | `varchar(160)` | Y | FK |  |
| `movement_id` | `varchar(140)` | Y | FK, UK |  |
| `posted_at` | `timestamptz` | Y |  |  |
| `posted_by` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `inventory_adjustment_line(inventory_adjustment_id)` → `inventory_adjustment(inventory_adjustment_id)`; ON DELETE CASCADE
- `inventory_adjustment_line(inventory_adjustment_type_id)` → `inventory_adjustment_type(inventory_adjustment_type_id)`
- `inventory_adjustment_line(source_balance_id)` → `inventory_balance(balance_id)` (optional)
- `inventory_adjustment_line(location_id)` → `warehouse_location(location_id)`
- `inventory_adjustment_line(item_id)` → `item(item_id)`
- `inventory_adjustment_line(lot_id)` → `inventory_lot(lot_id)` (optional)
- `inventory_adjustment_line(handling_unit_id)` → `handling_unit(handling_unit_id)` (optional)
- `inventory_adjustment_line(inventory_status_id)` → `inventory_status(inventory_status_id)`
- `inventory_adjustment_line(uom_id)` → `uom(uom_id)`
- `inventory_adjustment_line(resulting_balance_id)` → `inventory_balance(balance_id)` (optional)
- `inventory_adjustment_line(movement_id)` → `inventory_movement(movement_id)` (optional)
- `inventory_adjustment_line(posted_by)` → `app_account(account_id)` (optional)
- `inventory_adjustment_line(created_by)` → `app_account(account_id)`

### `stock_count`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `stock_count_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `stock_count_type_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `reason_code_id` | `uuid` | Y | FK |  |
| `freeze_inventory` | `boolean` | N |  | `false` |
| `started_at` | `timestamptz` | Y |  |  |
| `completed_at` | `timestamptz` | Y |  |  |
| `reviewed_at` | `timestamptz` | Y |  |  |
| `reviewed_by` | `uuid` | Y | FK |  |
| `posted_at` | `timestamptz` | Y |  |  |
| `posted_by` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `version_no` | `bigint` | N |  | `1` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | N | FK |  |

Relationships:
- `stock_count(document_type_id)` → `document_type(document_type_id)`
- `stock_count(stock_count_type_id)` → `stock_count_type(stock_count_type_id)`
- `stock_count(owner_id)` → `organization(organization_id)`
- `stock_count(warehouse_id)` → `warehouse(warehouse_id)`
- `stock_count(reason_code_id)` → `reason_code(reason_code_id)` (optional)
- `stock_count(reviewed_by)` → `app_account(account_id)` (optional)
- `stock_count(posted_by)` → `app_account(account_id)` (optional)
- `stock_count(created_by)` → `app_account(account_id)`
- `stock_count(updated_by)` → `app_account(account_id)`
- `stock_count(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `stock_count(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `stock_count_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `stock_count_line_id` | `varchar(150)` | N | PK |  |
| `stock_count_id` | `varchar(120)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `balance_id` | `varchar(160)` | Y | FK |  |
| `location_id` | `uuid` | N | FK |  |
| `item_id` | `uuid` | N | FK |  |
| `lot_id` | `varchar(120)` | Y | FK |  |
| `handling_unit_id` | `varchar(120)` | Y | FK |  |
| `inventory_status_id` | `uuid` | N | FK |  |
| `system_qty` | `numeric(20,6)` | N |  |  |
| `system_version_no` | `bigint` | Y |  |  |
| `counted_qty` | `numeric(20,6)` | Y |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `counted_at` | `timestamptz` | Y |  |  |
| `counted_by` | `uuid` | Y | FK |  |
| `variance_movement_id` | `varchar(140)` | Y | FK, UK |  |
| `posted_at` | `timestamptz` | Y |  |  |
| `posted_by` | `uuid` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `stock_count_line(stock_count_id)` → `stock_count(stock_count_id)`; ON DELETE CASCADE
- `stock_count_line(balance_id)` → `inventory_balance(balance_id)` (optional)
- `stock_count_line(location_id)` → `warehouse_location(location_id)`
- `stock_count_line(item_id)` → `item(item_id)`
- `stock_count_line(lot_id)` → `inventory_lot(lot_id)` (optional)
- `stock_count_line(handling_unit_id)` → `handling_unit(handling_unit_id)` (optional)
- `stock_count_line(inventory_status_id)` → `inventory_status(inventory_status_id)`
- `stock_count_line(uom_id)` → `uom(uom_id)`
- `stock_count_line(counted_by)` → `app_account(account_id)` (optional)
- `stock_count_line(variance_movement_id)` → `inventory_movement(movement_id)` (optional)
- `stock_count_line(posted_by)` → `app_account(account_id)` (optional)
- `stock_count_line(created_by)` → `app_account(account_id)`

## Outbound

### `carrier`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `carrier_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `business_partner_id` | `uuid` | Y | FK, UK |  |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(150)` | N |  |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `carrier(business_partner_id)` → `business_partner(partner_id)` (optional)

### `carrier_service`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `carrier_service_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `carrier_id` | `uuid` | N | FK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(100)` | N |  |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `carrier_service(carrier_id)` → `carrier(carrier_id)`

### `carrier_driver`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `driver_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `carrier_id` | `uuid` | N | FK |  |
| `account_id` | `uuid` | Y | FK, UK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(150)` | N |  |  |
| `phone_number` | `varchar(50)` | Y |  |  |
| `license_number` | `varchar(80)` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |

Relationships:
- `carrier_driver(carrier_id)` → `carrier(carrier_id)`
- `carrier_driver(account_id)` → `app_account(account_id)` (optional)
- `carrier_driver(created_by)` → `app_account(account_id)` (optional)

### `validation_severity`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `validation_severity_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(20)` | N | UK |  |
| `name` | `varchar(80)` | N |  |  |
| `blocks_processing` | `boolean` | N |  | `true` |
| `is_active` | `boolean` | N |  | `true` |

### `outbound_validation_rule`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_validation_rule_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(50)` | N | UK |  |
| `name` | `varchar(120)` | N |  |  |
| `description` | `text` | Y |  |  |
| `handler_code` | `varchar(60)` | N |  |  |
| `validation_severity_id` | `uuid` | N | FK |  |
| `display_order` | `integer` | N |  | `0` |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `outbound_validation_rule(validation_severity_id)` → `validation_severity(validation_severity_id)`

### `outbound_check_exception_status`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_check_exception_status_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(30)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `is_final` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

### `outbound_check_resolution_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_check_resolution_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(120)` | N |  |  |
| `description` | `text` | Y |  |  |
| `counts_as_stock_correction` | `boolean` | N |  | `false` |
| `counts_as_replacement` | `boolean` | N |  | `false` |
| `counts_as_short_acceptance` | `boolean` | N |  | `false` |
| `requires_approval` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

### `outbound_wave_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_wave_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

### `outbound_check_result`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_check_result_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_pass` | `boolean` | N |  | `false` |
| `requires_note` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

### `delivery_event_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `delivery_event_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `marks_delivered` | `boolean` | N |  | `false` |
| `marks_failed` | `boolean` | N |  | `false` |
| `marks_returned` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

### `delivery_failure_reason`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `delivery_failure_reason_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

### `outbound_order`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `customer_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `requested_ship_at` | `timestamptz` | Y |  |  |
| `external_reference` | `varchar(120)` | Y |  |  |
| `customer_order_no` | `varchar(120)` | Y |  |  |
| `client_delivery_order_no` | `varchar(120)` | N |  |  |
| `ship_to_partner_id` | `uuid` | Y | FK |  |
| `ship_to_name` | `varchar(150)` | N |  |  |
| `ship_to_address_1` | `varchar(255)` | N |  |  |
| `ship_to_address_2` | `varchar(255)` | Y |  |  |
| `ship_to_city` | `varchar(100)` | Y |  |  |
| `ship_to_province` | `varchar(100)` | Y |  |  |
| `ship_to_postal_code` | `varchar(20)` | Y |  |  |
| `ship_to_country_code` | `varchar(2)` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `validated_at` | `timestamptz` | Y |  |  |
| `validated_by` | `uuid` | Y | FK |  |
| `allocated_at` | `timestamptz` | Y |  |  |
| `completed_at` | `timestamptz` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | Y | FK |  |
| `version_no` | `integer` | N |  | `1` |

Relationships:
- `outbound_order(document_type_id)` → `document_type(document_type_id)`
- `outbound_order(owner_id)` → `organization(organization_id)`
- `outbound_order(customer_id)` → `business_partner(partner_id)`
- `outbound_order(warehouse_id)` → `warehouse(warehouse_id)`
- `outbound_order(ship_to_partner_id)` → `business_partner(partner_id)` (optional)
- `outbound_order(validated_by)` → `app_account(account_id)` (optional)
- `outbound_order(created_by)` → `app_account(account_id)`
- `outbound_order(updated_by)` → `app_account(account_id)` (optional)
- `outbound_order(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `outbound_order(owner_id, customer_id)` → `business_partner(owner_id, partner_id)`
- `outbound_order(owner_id, ship_to_partner_id)` → `business_partner(owner_id, partner_id)` (optional)
- `outbound_order(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `outbound_order_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_line_id` | `varchar(150)` | N | PK |  |
| `outbound_id` | `varchar(120)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `item_id` | `uuid` | N | FK |  |
| `ordered_qty` | `numeric(20,6)` | N |  |  |
| `allocated_qty` | `numeric(20,6)` | N |  | `0` |
| `picked_qty` | `numeric(20,6)` | N |  | `0` |
| `checked_qty` | `numeric(20,6)` | N |  | `0` |
| `packed_qty` | `numeric(20,6)` | N |  | `0` |
| `shipped_qty` | `numeric(20,6)` | N |  | `0` |
| `delivered_qty` | `numeric(20,6)` | N |  | `0` |
| `rejected_qty` | `numeric(20,6)` | N |  | `0` |
| `short_accepted_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `requested_lot_no` | `varchar(100)` | Y |  |  |
| `customer_line_reference` | `varchar(100)` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `outbound_order_line(outbound_id)` → `outbound_order(outbound_id)`; ON DELETE CASCADE
- `outbound_order_line(item_id)` → `item(item_id)`
- `outbound_order_line(uom_id)` → `uom(uom_id)`
- `outbound_order_line(created_by)` → `app_account(account_id)`

### `outbound_validation_run`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `validation_run_id` | `varchar(140)` | N | PK |  |
| `outbound_id` | `varchar(120)` | N | FK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `validated_at` | `timestamptz` | Y |  |  |
| `validated_by` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `outbound_validation_run(outbound_id)` → `outbound_order(outbound_id)`
- `outbound_validation_run(document_type_id)` → `document_type(document_type_id)`
- `outbound_validation_run(validated_by)` → `app_account(account_id)` (optional)
- `outbound_validation_run(created_by)` → `app_account(account_id)`
- `outbound_validation_run(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `outbound_validation_result_detail`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `validation_result_id` | `varchar(170)` | N | PK |  |
| `validation_run_id` | `varchar(140)` | N | FK |  |
| `outbound_validation_rule_id` | `uuid` | N | FK |  |
| `passed` | `boolean` | N |  |  |
| `result_message` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |

Relationships:
- `outbound_validation_result_detail(validation_run_id)` → `outbound_validation_run(validation_run_id)`; ON DELETE CASCADE
- `outbound_validation_result_detail(outbound_validation_rule_id)` → `outbound_validation_rule(outbound_validation_rule_id)`

### `outbound_wave`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `wave_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `outbound_wave_type_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `picking_strategy_id` | `uuid` | Y | FK |  |
| `business_date` | `date` | N |  |  |
| `planned_release_at` | `timestamptz` | Y |  |  |
| `released_at` | `timestamptz` | Y |  |  |
| `completed_at` | `timestamptz` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `version_no` | `bigint` | N |  | `1` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | N | FK |  |

Relationships:
- `outbound_wave(document_type_id)` → `document_type(document_type_id)`
- `outbound_wave(outbound_wave_type_id)` → `outbound_wave_type(outbound_wave_type_id)`
- `outbound_wave(owner_id)` → `organization(organization_id)`
- `outbound_wave(warehouse_id)` → `warehouse(warehouse_id)`
- `outbound_wave(picking_strategy_id)` → `picking_strategy(picking_strategy_id)` (optional)
- `outbound_wave(created_by)` → `app_account(account_id)`
- `outbound_wave(updated_by)` → `app_account(account_id)`
- `outbound_wave(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `outbound_wave(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `outbound_wave_order`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `wave_id` | `varchar(120)` | N | PK, FK |  |
| `outbound_id` | `varchar(120)` | N | PK, FK |  |
| `added_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `added_by` | `uuid` | N | FK |  |

Relationships:
- `outbound_wave_order(wave_id)` → `outbound_wave(wave_id)`; ON DELETE CASCADE
- `outbound_wave_order(outbound_id)` → `outbound_order(outbound_id)`
- `outbound_wave_order(added_by)` → `app_account(account_id)`

### `inventory_reservation`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `reservation_id` | `varchar(140)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `picking_strategy_id` | `uuid` | Y | FK |  |
| `outbound_line_id` | `varchar(150)` | N | FK |  |
| `balance_id` | `varchar(160)` | N | FK |  |
| `reserved_qty` | `numeric(20,6)` | N |  |  |
| `picked_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `reserved_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `released_at` | `timestamptz` | Y |  |  |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `inventory_reservation(document_type_id)` → `document_type(document_type_id)`
- `inventory_reservation(picking_strategy_id)` → `picking_strategy(picking_strategy_id)` (optional)
- `inventory_reservation(outbound_line_id)` → `outbound_order_line(outbound_line_id)`
- `inventory_reservation(balance_id)` → `inventory_balance(balance_id)`
- `inventory_reservation(uom_id)` → `uom(uom_id)`
- `inventory_reservation(created_by)` → `app_account(account_id)`
- `inventory_reservation(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `pick_task`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `pick_task_id` | `varchar(140)` | N | PK |  |
| `task_type_id` | `uuid` | N | FK |  |
| `task_status_id` | `uuid` | N | FK |  |
| `task_priority_id` | `uuid` | N | FK |  |
| `wave_id` | `varchar(120)` | N | FK |  |
| `reservation_id` | `varchar(140)` | N | FK, UK |  |
| `outbound_line_id` | `varchar(150)` | N | FK |  |
| `source_location_id` | `uuid` | N | FK |  |
| `target_location_id` | `uuid` | Y | FK |  |
| `planned_qty` | `numeric(20,6)` | N |  |  |
| `picked_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `assigned_to` | `uuid` | Y | FK |  |
| `started_at` | `timestamptz` | Y |  |  |
| `completed_at` | `timestamptz` | Y |  |  |
| `short_qty` | `numeric(20,6)` | N |  | `0` |
| `short_reason_code_id` | `uuid` | Y | FK |  |
| `result_notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `pick_task(task_type_id)` → `task_type(task_type_id)`
- `pick_task(task_status_id)` → `task_status(task_status_id)`
- `pick_task(task_priority_id)` → `task_priority(task_priority_id)`
- `pick_task(wave_id)` → `outbound_wave(wave_id)`
- `pick_task(reservation_id)` → `inventory_reservation(reservation_id)`
- `pick_task(outbound_line_id)` → `outbound_order_line(outbound_line_id)`
- `pick_task(source_location_id)` → `warehouse_location(location_id)`
- `pick_task(target_location_id)` → `warehouse_location(location_id)` (optional)
- `pick_task(uom_id)` → `uom(uom_id)`
- `pick_task(assigned_to)` → `app_account(account_id)` (optional)
- `pick_task(short_reason_code_id)` → `reason_code(reason_code_id)` (optional)
- `pick_task(created_by)` → `app_account(account_id)`

### `pick_execution`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `pick_execution_id` | `varchar(170)` | N | PK |  |
| `pick_task_id` | `varchar(140)` | N | FK |  |
| `source_balance_id` | `varchar(160)` | N | FK |  |
| `staging_balance_id` | `varchar(160)` | N | FK |  |
| `picked_qty` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `movement_id` | `varchar(140)` | N | FK, UK |  |
| `picked_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `picked_by` | `uuid` | N | FK |  |

Relationships:
- `pick_execution(pick_task_id)` → `pick_task(pick_task_id)`
- `pick_execution(source_balance_id)` → `inventory_balance(balance_id)`
- `pick_execution(staging_balance_id)` → `inventory_balance(balance_id)`
- `pick_execution(uom_id)` → `uom(uom_id)`
- `pick_execution(movement_id)` → `inventory_movement(movement_id)`
- `pick_execution(picked_by)` → `app_account(account_id)`

### `outbound_staging`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `staging_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `outbound_id` | `varchar(120)` | N | FK |  |
| `wave_id` | `varchar(120)` | N | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `staging_location_id` | `uuid` | N | FK |  |
| `staged_at` | `timestamptz` | Y |  |  |
| `staged_by` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `outbound_staging(document_type_id)` → `document_type(document_type_id)`
- `outbound_staging(outbound_id)` → `outbound_order(outbound_id)`
- `outbound_staging(wave_id)` → `outbound_wave(wave_id)`
- `outbound_staging(warehouse_id)` → `warehouse(warehouse_id)`
- `outbound_staging(staging_location_id)` → `warehouse_location(location_id)`
- `outbound_staging(staged_by)` → `app_account(account_id)` (optional)
- `outbound_staging(created_by)` → `app_account(account_id)`
- `outbound_staging(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `outbound_staging_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `staging_line_id` | `varchar(160)` | N | PK |  |
| `staging_id` | `varchar(120)` | N | FK |  |
| `pick_execution_id` | `varchar(170)` | N | FK, UK |  |
| `staging_balance_id` | `varchar(160)` | N | FK |  |
| `staged_qty` | `numeric(20,6)` | N |  |  |
| `removed_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `outbound_staging_line(staging_id)` → `outbound_staging(staging_id)`; ON DELETE CASCADE
- `outbound_staging_line(pick_execution_id)` → `pick_execution(pick_execution_id)`
- `outbound_staging_line(staging_balance_id)` → `inventory_balance(balance_id)`
- `outbound_staging_line(uom_id)` → `uom(uom_id)`
- `outbound_staging_line(created_by)` → `app_account(account_id)`

### `outbound_check`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_check_id` | `varchar(120)` | N | PK |  |
| `parent_check_id` | `varchar(120)` | Y | FK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `staging_id` | `varchar(120)` | N | FK |  |
| `checked_at` | `timestamptz` | Y |  |  |
| `checked_by` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `outbound_check(parent_check_id)` → `outbound_check(outbound_check_id)` (optional)
- `outbound_check(document_type_id)` → `document_type(document_type_id)`
- `outbound_check(staging_id)` → `outbound_staging(staging_id)`
- `outbound_check(checked_by)` → `app_account(account_id)` (optional)
- `outbound_check(created_by)` → `app_account(account_id)`
- `outbound_check(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `outbound_check_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_check_line_id` | `varchar(160)` | N | PK |  |
| `outbound_check_id` | `varchar(120)` | N | FK |  |
| `staging_line_id` | `varchar(160)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `expected_qty` | `numeric(20,6)` | N |  |  |
| `checked_qty` | `numeric(20,6)` | Y |  |  |
| `exception_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `outbound_check_result_id` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `checked_at` | `timestamptz` | Y |  |  |
| `checked_by` | `uuid` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `outbound_check_line(outbound_check_id)` → `outbound_check(outbound_check_id)`; ON DELETE CASCADE
- `outbound_check_line(staging_line_id)` → `outbound_staging_line(staging_line_id)`
- `outbound_check_line(uom_id)` → `uom(uom_id)`
- `outbound_check_line(outbound_check_result_id)` → `outbound_check_result(outbound_check_result_id)` (optional)
- `outbound_check_line(checked_by)` → `app_account(account_id)` (optional)
- `outbound_check_line(created_by)` → `app_account(account_id)`

### `outbound_check_exception`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_check_exception_id` | `varchar(170)` | N | PK |  |
| `outbound_check_line_id` | `varchar(160)` | N | FK, UK |  |
| `status_id` | `uuid` | N | FK |  |
| `exception_qty` | `numeric(20,6)` | N |  |  |
| `opened_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `resolved_at` | `timestamptz` | Y |  |  |
| `resolved_by` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `outbound_check_exception(outbound_check_line_id)` → `outbound_check_line(outbound_check_line_id)`
- `outbound_check_exception(status_id)` → `outbound_check_exception_status(outbound_check_exception_status_id)`
- `outbound_check_exception(resolved_by)` → `app_account(account_id)` (optional)
- `outbound_check_exception(created_by)` → `app_account(account_id)`

### `outbound_check_resolution`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_check_resolution_id` | `varchar(190)` | N | PK |  |
| `outbound_check_exception_id` | `varchar(170)` | N | FK |  |
| `outbound_check_resolution_type_id` | `uuid` | N | FK |  |
| `resolved_qty` | `numeric(20,6)` | N |  |  |
| `notes` | `text` | Y |  |  |
| `approved_at` | `timestamptz` | Y |  |  |
| `approved_by` | `uuid` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `outbound_check_resolution(outbound_check_exception_id)` → `outbound_check_exception(outbound_check_exception_id)`
- `outbound_check_resolution(outbound_check_resolution_type_id)` → `outbound_check_resolution_type(outbound_check_resolution_type_id)`
- `outbound_check_resolution(approved_by)` → `app_account(account_id)` (optional)
- `outbound_check_resolution(created_by)` → `app_account(account_id)`

### `outbound_check_resolution_movement`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_check_resolution_id` | `varchar(190)` | N | PK, FK |  |
| `movement_id` | `varchar(140)` | N | PK, FK, UK |  |

Relationships:
- `outbound_check_resolution_movement(outbound_check_resolution_id)` → `outbound_check_resolution(outbound_check_resolution_id)`; ON DELETE CASCADE
- `outbound_check_resolution_movement(movement_id)` → `inventory_movement(movement_id)`

### `outbound_check_resolution_reservation`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `outbound_check_resolution_id` | `varchar(190)` | N | PK, FK |  |
| `reservation_id` | `varchar(140)` | N | PK, FK, UK |  |

Relationships:
- `outbound_check_resolution_reservation(outbound_check_resolution_id)` → `outbound_check_resolution(outbound_check_resolution_id)`; ON DELETE CASCADE
- `outbound_check_resolution_reservation(reservation_id)` → `inventory_reservation(reservation_id)`

### `packing`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `packing_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `outbound_id` | `varchar(120)` | N | FK, UK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `packing_location_id` | `uuid` | Y | FK |  |
| `packed_at` | `timestamptz` | Y |  |  |
| `packed_by` | `uuid` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `packing(document_type_id)` → `document_type(document_type_id)`
- `packing(outbound_id)` → `outbound_order(outbound_id)`
- `packing(warehouse_id)` → `warehouse(warehouse_id)`
- `packing(packing_location_id)` → `warehouse_location(location_id)` (optional)
- `packing(packed_by)` → `app_account(account_id)` (optional)
- `packing(created_by)` → `app_account(account_id)`
- `packing(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `packing_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `packing_line_id` | `varchar(150)` | N | PK |  |
| `packing_id` | `varchar(120)` | N | FK |  |
| `pick_task_id` | `varchar(140)` | N | FK |  |
| `outbound_check_line_id` | `varchar(160)` | N | FK, UK |  |
| `source_balance_id` | `varchar(160)` | N | FK |  |
| `packing_balance_id` | `varchar(160)` | N | FK |  |
| `handling_unit_id` | `varchar(120)` | Y | FK |  |
| `packed_qty` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `movement_id` | `varchar(140)` | N | FK, UK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `packing_line(packing_id)` → `packing(packing_id)`; ON DELETE CASCADE
- `packing_line(pick_task_id)` → `pick_task(pick_task_id)`
- `packing_line(outbound_check_line_id)` → `outbound_check_line(outbound_check_line_id)`
- `packing_line(source_balance_id)` → `inventory_balance(balance_id)`
- `packing_line(packing_balance_id)` → `inventory_balance(balance_id)`
- `packing_line(handling_unit_id)` → `handling_unit(handling_unit_id)` (optional)
- `packing_line(uom_id)` → `uom(uom_id)`
- `packing_line(movement_id)` → `inventory_movement(movement_id)`
- `packing_line(created_by)` → `app_account(account_id)`

### `shipment`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `shipment_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `carrier_service_id` | `uuid` | Y | FK |  |
| `warehouse_id` | `uuid` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `route_reference` | `varchar(100)` | Y |  |  |
| `tracking_number` | `varchar(150)` | Y |  |  |
| `vehicle_number` | `varchar(60)` | Y |  |  |
| `seal_number` | `varchar(60)` | Y |  |  |
| `shipped_at` | `timestamptz` | Y |  |  |
| `dispatched_by` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `shipment(document_type_id)` → `document_type(document_type_id)`
- `shipment(owner_id)` → `organization(organization_id)`
- `shipment(carrier_service_id)` → `carrier_service(carrier_service_id)` (optional)
- `shipment(warehouse_id)` → `warehouse(warehouse_id)`
- `shipment(dispatched_by)` → `app_account(account_id)` (optional)
- `shipment(created_by)` → `app_account(account_id)`
- `shipment(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `shipment(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`

### `shipment_order`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `shipment_id` | `varchar(120)` | N | PK, FK |  |
| `outbound_id` | `varchar(120)` | N | PK, FK |  |
| `added_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `added_by` | `uuid` | N | FK |  |
| `removed_at` | `timestamptz` | Y |  |  |
| `removed_by` | `uuid` | Y | FK |  |

Relationships:
- `shipment_order(shipment_id)` → `shipment(shipment_id)`; ON DELETE CASCADE
- `shipment_order(outbound_id)` → `outbound_order(outbound_id)`
- `shipment_order(added_by)` → `app_account(account_id)`
- `shipment_order(removed_by)` → `app_account(account_id)` (optional)

### `shipment_driver`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `shipment_id` | `varchar(120)` | N | PK, FK |  |
| `driver_id` | `uuid` | N | PK, FK |  |
| `is_primary` | `boolean` | N |  | `false` |
| `assigned_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `assigned_by` | `uuid` | N | FK |  |

Relationships:
- `shipment_driver(shipment_id)` → `shipment(shipment_id)`; ON DELETE CASCADE
- `shipment_driver(driver_id)` → `carrier_driver(driver_id)`
- `shipment_driver(assigned_by)` → `app_account(account_id)`

### `shipment_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `shipment_line_id` | `varchar(160)` | N | PK |  |
| `shipment_id` | `varchar(120)` | N | FK |  |
| `packing_line_id` | `varchar(150)` | N | FK, UK |  |
| `source_balance_id` | `varchar(160)` | N | FK |  |
| `shipped_qty` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `movement_id` | `varchar(140)` | N | FK, UK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `shipment_line(shipment_id)` → `shipment(shipment_id)`; ON DELETE CASCADE
- `shipment_line(packing_line_id)` → `packing_line(packing_line_id)`
- `shipment_line(source_balance_id)` → `inventory_balance(balance_id)`
- `shipment_line(uom_id)` → `uom(uom_id)`
- `shipment_line(movement_id)` → `inventory_movement(movement_id)`
- `shipment_line(created_by)` → `app_account(account_id)`

### `shipment_packing`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `shipment_id` | `varchar(120)` | N | PK, FK |  |
| `packing_id` | `varchar(120)` | N | PK, FK |  |

Relationships:
- `shipment_packing(shipment_id)` → `shipment(shipment_id)`; ON DELETE CASCADE
- `shipment_packing(packing_id)` → `packing(packing_id)`

### `delivery`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `delivery_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `shipment_id` | `varchar(120)` | N | FK |  |
| `outbound_id` | `varchar(120)` | N | FK |  |
| `business_date` | `date` | N |  |  |
| `planned_delivery_at` | `timestamptz` | Y |  |  |
| `arrived_at` | `timestamptz` | Y |  |  |
| `delivered_at` | `timestamptz` | Y |  |  |
| `recipient_name` | `varchar(150)` | Y |  |  |
| `recipient_reference` | `varchar(100)` | Y |  |  |
| `proof_reference` | `varchar(200)` | Y |  |  |
| `proof_uri` | `text` | Y |  |  |
| `latitude` | `numeric(10,7)` | Y |  |  |
| `longitude` | `numeric(10,7)` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | N | FK |  |

Relationships:
- `delivery(document_type_id)` → `document_type(document_type_id)`
- `delivery(shipment_id)` → `shipment(shipment_id)`
- `delivery(outbound_id)` → `outbound_order(outbound_id)`
- `delivery(created_by)` → `app_account(account_id)`
- `delivery(updated_by)` → `app_account(account_id)`
- `delivery(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `delivery_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `delivery_line_id` | `varchar(160)` | N | PK |  |
| `delivery_id` | `varchar(120)` | N | FK |  |
| `shipment_line_id` | `varchar(160)` | N | FK, UK |  |
| `planned_qty` | `numeric(20,6)` | N |  |  |
| `delivered_qty` | `numeric(20,6)` | N |  | `0` |
| `returned_qty` | `numeric(20,6)` | N |  | `0` |
| `uom_id` | `uuid` | N | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `delivery_line(delivery_id)` → `delivery(delivery_id)`; ON DELETE CASCADE
- `delivery_line(shipment_line_id)` → `shipment_line(shipment_line_id)`
- `delivery_line(uom_id)` → `uom(uom_id)`
- `delivery_line(created_by)` → `app_account(account_id)`

### `delivery_event`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `delivery_event_id` | `varchar(160)` | N | PK |  |
| `delivery_id` | `varchar(120)` | N | FK |  |
| `delivery_event_type_id` | `uuid` | N | FK |  |
| `event_at` | `timestamptz` | N |  |  |
| `delivery_failure_reason_id` | `uuid` | Y | FK |  |
| `recipient_name` | `varchar(150)` | Y |  |  |
| `recipient_reference` | `varchar(100)` | Y |  |  |
| `proof_reference` | `varchar(200)` | Y |  |  |
| `proof_uri` | `text` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `latitude` | `numeric(10,7)` | Y |  |  |
| `longitude` | `numeric(10,7)` | Y |  |  |
| `recorded_by` | `uuid` | N | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |

Relationships:
- `delivery_event(delivery_id)` → `delivery(delivery_id)`; ON DELETE CASCADE
- `delivery_event(delivery_event_type_id)` → `delivery_event_type(delivery_event_type_id)`
- `delivery_event(delivery_failure_reason_id)` → `delivery_failure_reason(delivery_failure_reason_id)` (optional)
- `delivery_event(recorded_by)` → `app_account(account_id)`

### `delivery_event_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `delivery_event_id` | `varchar(160)` | N | PK, FK |  |
| `delivery_line_id` | `varchar(160)` | N | PK, FK |  |
| `delivered_qty` | `numeric(20,6)` | N |  |  |

Relationships:
- `delivery_event_line(delivery_event_id)` → `delivery_event(delivery_event_id)`; ON DELETE CASCADE
- `delivery_event_line(delivery_line_id)` → `delivery_line(delivery_line_id)`

### `outbound_return_policy`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `owner_id` | `uuid` | N | PK, FK |  |
| `warehouse_id` | `uuid` | N | PK, FK |  |
| `return_location_id` | `uuid` | N | FK |  |
| `return_inventory_status_id` | `uuid` | N | FK |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | Y | FK |  |

Relationships:
- `outbound_return_policy(owner_id)` → `organization(organization_id)`
- `outbound_return_policy(warehouse_id)` → `warehouse(warehouse_id)`
- `outbound_return_policy(return_inventory_status_id)` → `inventory_status(inventory_status_id)`
- `outbound_return_policy(created_by)` → `app_account(account_id)`
- `outbound_return_policy(updated_by)` → `app_account(account_id)` (optional)
- `outbound_return_policy(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)`
- `outbound_return_policy(return_location_id, warehouse_id)` → `warehouse_location(location_id, warehouse_id)`

### `delivery_return_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `delivery_return_line_id` | `varchar(170)` | N | PK |  |
| `delivery_id` | `varchar(120)` | N | FK |  |
| `delivery_line_id` | `varchar(160)` | N | FK |  |
| `return_location_id` | `uuid` | N | FK |  |
| `returned_balance_id` | `varchar(160)` | N | FK |  |
| `returned_qty` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | N | FK |  |
| `movement_id` | `varchar(140)` | N | FK, UK |  |
| `returned_at` | `timestamptz` | N |  |  |
| `returned_by` | `uuid` | N | FK |  |

Relationships:
- `delivery_return_line(delivery_id)` → `delivery(delivery_id)`
- `delivery_return_line(delivery_line_id)` → `delivery_line(delivery_line_id)`
- `delivery_return_line(return_location_id)` → `warehouse_location(location_id)`
- `delivery_return_line(returned_balance_id)` → `inventory_balance(balance_id)`
- `delivery_return_line(uom_id)` → `uom(uom_id)`
- `delivery_return_line(movement_id)` → `inventory_movement(movement_id)`
- `delivery_return_line(returned_by)` → `app_account(account_id)`

## Billing

### `currency`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `currency_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(3)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `decimal_scale` | `smallint` | N |  | `2` |
| `is_active` | `boolean` | N |  | `true` |

### `billing_cycle`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `billing_cycle_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `period_days` | `integer` | Y |  |  |
| `period_months` | `integer` | Y |  |  |
| `is_on_demand` | `boolean` | N |  | `false` |
| `is_active` | `boolean` | N |  | `true` |

### `payment_term`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `payment_term_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `due_days` | `integer` | N |  |  |
| `is_active` | `boolean` | N |  | `true` |

### `charge_basis`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `charge_basis_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `description` | `text` | Y |  |  |
| `requires_uom` | `boolean` | N |  | `true` |
| `use_event_quantity` | `boolean` | N |  | `true` |
| `is_active` | `boolean` | N |  | `true` |

### `tax_rule`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `tax_rule_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `owner_id` | `uuid` | Y | FK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(120)` | N |  |  |
| `rate_percent` | `numeric(9,6)` | N |  |  |
| `is_inclusive` | `boolean` | N |  | `false` |
| `effective_from` | `date` | N |  |  |
| `effective_until` | `date` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `tax_rule(owner_id)` → `organization(organization_id)` (optional)

### `billing_service`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `billing_service_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(50)` | N | UK |  |
| `name` | `varchar(120)` | N |  |  |
| `module_code` | `varchar(50)` | N | FK |  |
| `description` | `text` | Y |  |  |
| `default_charge_basis_id` | `uuid` | N | FK |  |
| `default_uom_id` | `uuid` | Y | FK |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `billing_service(module_code)` → `app_module(code)`
- `billing_service(default_charge_basis_id)` → `charge_basis(charge_basis_id)`
- `billing_service(default_uom_id)` → `uom(uom_id)` (optional)

### `billing_service_movement_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `billing_service_id` | `uuid` | N | PK, FK |  |
| `movement_type_id` | `uuid` | N | PK, FK, UK |  |
| `is_active` | `boolean` | N |  | `true` |

Relationships:
- `billing_service_movement_type(billing_service_id)` → `billing_service(billing_service_id)`
- `billing_service_movement_type(movement_type_id)` → `movement_type(movement_type_id)`

### `billing_adjustment_type`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `billing_adjustment_type_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `amount_effect` | `smallint` | N |  |  |
| `is_active` | `boolean` | N |  | `true` |

### `payment_method`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `payment_method_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `code` | `varchar(40)` | N | UK |  |
| `name` | `varchar(100)` | N |  |  |
| `requires_reference` | `boolean` | N |  | `true` |
| `is_active` | `boolean` | N |  | `true` |

### `billing_account`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `billing_account_id` | `uuid` | N | PK | `gen_random_uuid()` |
| `owner_id` | `uuid` | N | FK |  |
| `code` | `varchar(40)` | N |  |  |
| `name` | `varchar(150)` | N |  |  |
| `bill_to_partner_id` | `uuid` | Y | FK |  |
| `currency_id` | `uuid` | N | FK |  |
| `billing_cycle_id` | `uuid` | N | FK |  |
| `payment_term_id` | `uuid` | N | FK |  |
| `default_tax_rule_id` | `uuid` | Y | FK |  |
| `bill_to_name` | `varchar(200)` | N |  |  |
| `bill_to_tax_number` | `varchar(100)` | Y |  |  |
| `bill_to_address_1` | `varchar(255)` | N |  |  |
| `bill_to_address_2` | `varchar(255)` | Y |  |  |
| `bill_to_city` | `varchar(100)` | Y |  |  |
| `bill_to_province` | `varchar(100)` | Y |  |  |
| `bill_to_postal_code` | `varchar(20)` | Y |  |  |
| `bill_to_country_code` | `varchar(2)` | Y |  |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | Y | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | Y | FK |  |

Relationships:
- `billing_account(owner_id)` → `organization(organization_id)`
- `billing_account(bill_to_partner_id)` → `business_partner(partner_id)` (optional)
- `billing_account(currency_id)` → `currency(currency_id)`
- `billing_account(billing_cycle_id)` → `billing_cycle(billing_cycle_id)`
- `billing_account(payment_term_id)` → `payment_term(payment_term_id)`
- `billing_account(default_tax_rule_id)` → `tax_rule(tax_rule_id)` (optional)
- `billing_account(created_by)` → `app_account(account_id)` (optional)
- `billing_account(updated_by)` → `app_account(account_id)` (optional)
- `billing_account(owner_id, bill_to_partner_id)` → `business_partner(owner_id, partner_id)` (optional)

### `billing_contract`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `billing_contract_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `billing_account_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | Y | FK |  |
| `currency_id` | `uuid` | N | FK |  |
| `default_tax_rule_id` | `uuid` | Y | FK |  |
| `contract_number` | `varchar(100)` | N |  |  |
| `effective_from` | `date` | N |  |  |
| `effective_until` | `date` | Y |  |  |
| `auto_renew` | `boolean` | N |  | `false` |
| `notes` | `text` | Y |  |  |
| `version_no` | `bigint` | N |  | `1` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | N | FK |  |

Relationships:
- `billing_contract(document_type_id)` → `document_type(document_type_id)`
- `billing_contract(billing_account_id)` → `billing_account(billing_account_id)`
- `billing_contract(owner_id)` → `organization(organization_id)`
- `billing_contract(warehouse_id)` → `warehouse(warehouse_id)` (optional)
- `billing_contract(currency_id)` → `currency(currency_id)`
- `billing_contract(default_tax_rule_id)` → `tax_rule(tax_rule_id)` (optional)
- `billing_contract(created_by)` → `app_account(account_id)`
- `billing_contract(updated_by)` → `app_account(account_id)`
- `billing_contract(document_type_id, status_id)` → `document_status(document_type_id, status_id)`
- `billing_contract(owner_id, billing_account_id)` → `billing_account(owner_id, billing_account_id)`
- `billing_contract(owner_id, warehouse_id)` → `warehouse_owner(owner_id, warehouse_id)` (optional)

### `rate_card`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `rate_card_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `billing_contract_id` | `varchar(120)` | N | FK |  |
| `version_label` | `varchar(40)` | N |  |  |
| `effective_from` | `date` | N |  |  |
| `effective_until` | `date` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `version_no` | `bigint` | N |  | `1` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | N | FK |  |

Relationships:
- `rate_card(document_type_id)` → `document_type(document_type_id)`
- `rate_card(billing_contract_id)` → `billing_contract(billing_contract_id)`
- `rate_card(created_by)` → `app_account(account_id)`
- `rate_card(updated_by)` → `app_account(account_id)`
- `rate_card(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `rate_card_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `rate_card_line_id` | `varchar(150)` | N | PK |  |
| `rate_card_id` | `varchar(120)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `billing_service_id` | `uuid` | N | FK |  |
| `charge_basis_id` | `uuid` | N | FK |  |
| `uom_id` | `uuid` | Y | FK |  |
| `warehouse_id` | `uuid` | Y | FK |  |
| `item_id` | `uuid` | Y | FK |  |
| `category_id` | `uuid` | Y | FK |  |
| `unit_rate` | `numeric(20,6)` | N |  |  |
| `included_quantity` | `numeric(20,6)` | N |  | `0` |
| `minimum_charge` | `numeric(20,4)` | N |  | `0` |
| `maximum_charge` | `numeric(20,4)` | Y |  |  |
| `rounding_increment` | `numeric(20,6)` | Y |  |  |
| `priority_no` | `integer` | N |  | `100` |
| `tax_rule_id` | `uuid` | Y | FK |  |
| `is_active` | `boolean` | N |  | `true` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `rate_card_line(rate_card_id)` → `rate_card(rate_card_id)`; ON DELETE CASCADE
- `rate_card_line(billing_service_id)` → `billing_service(billing_service_id)`
- `rate_card_line(charge_basis_id)` → `charge_basis(charge_basis_id)`
- `rate_card_line(uom_id)` → `uom(uom_id)` (optional)
- `rate_card_line(warehouse_id)` → `warehouse(warehouse_id)` (optional)
- `rate_card_line(item_id)` → `item(item_id)` (optional)
- `rate_card_line(category_id)` → `item_category(category_id)` (optional)
- `rate_card_line(tax_rule_id)` → `tax_rule(tax_rule_id)` (optional)
- `rate_card_line(created_by)` → `app_account(account_id)`

### `rate_card_tier`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `rate_card_tier_id` | `varchar(170)` | N | PK |  |
| `rate_card_line_id` | `varchar(150)` | N | FK |  |
| `tier_no` | `integer` | N |  |  |
| `from_quantity` | `numeric(20,6)` | N |  |  |
| `until_quantity` | `numeric(20,6)` | Y |  |  |
| `unit_rate` | `numeric(20,6)` | N |  |  |

Relationships:
- `rate_card_tier(rate_card_line_id)` → `rate_card_line(rate_card_line_id)`; ON DELETE CASCADE

### `billable_event`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `billable_event_id` | `varchar(140)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `billing_account_id` | `uuid` | N | FK |  |
| `billing_contract_id` | `varchar(120)` | N | FK |  |
| `billing_service_id` | `uuid` | N | FK |  |
| `owner_id` | `uuid` | N | FK |  |
| `warehouse_id` | `uuid` | Y | FK |  |
| `item_id` | `uuid` | Y | FK |  |
| `category_id` | `uuid` | Y | FK |  |
| `handling_unit_id` | `varchar(120)` | Y | FK |  |
| `business_date` | `date` | N |  |  |
| `occurred_at` | `timestamptz` | N |  |  |
| `source_document_type_id` | `uuid` | Y | FK |  |
| `source_document_id` | `varchar(140)` | N |  |  |
| `source_line_id` | `varchar(170)` | Y |  |  |
| `event_key` | `varchar(80)` | N |  | `'DEFAULT'` |
| `quantity` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | Y | FK |  |
| `attributes` | `jsonb` | N |  | `'{}'::jsonb` |
| `exclusion_reason_code_id` | `uuid` | Y | FK |  |
| `rated_at` | `timestamptz` | Y |  |  |
| `rated_by` | `uuid` | Y | FK |  |
| `excluded_at` | `timestamptz` | Y |  |  |
| `excluded_by` | `uuid` | Y | FK |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `billable_event(document_type_id)` → `document_type(document_type_id)`
- `billable_event(billing_account_id)` → `billing_account(billing_account_id)`
- `billable_event(billing_contract_id)` → `billing_contract(billing_contract_id)`
- `billable_event(billing_service_id)` → `billing_service(billing_service_id)`
- `billable_event(owner_id)` → `organization(organization_id)`
- `billable_event(warehouse_id)` → `warehouse(warehouse_id)` (optional)
- `billable_event(item_id)` → `item(item_id)` (optional)
- `billable_event(category_id)` → `item_category(category_id)` (optional)
- `billable_event(handling_unit_id)` → `handling_unit(handling_unit_id)` (optional)
- `billable_event(source_document_type_id)` → `document_type(document_type_id)` (optional)
- `billable_event(uom_id)` → `uom(uom_id)` (optional)
- `billable_event(exclusion_reason_code_id)` → `reason_code(reason_code_id)` (optional)
- `billable_event(rated_by)` → `app_account(account_id)` (optional)
- `billable_event(excluded_by)` → `app_account(account_id)` (optional)
- `billable_event(created_by)` → `app_account(account_id)`
- `billable_event(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `billing_run`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `billing_run_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `billing_account_id` | `uuid` | N | FK |  |
| `billing_contract_id` | `varchar(120)` | N | FK |  |
| `currency_id` | `uuid` | N | FK |  |
| `period_from` | `date` | N |  |  |
| `period_until` | `date` | N |  |  |
| `calculated_at` | `timestamptz` | Y |  |  |
| `reviewed_at` | `timestamptz` | Y |  |  |
| `reviewed_by` | `uuid` | Y | FK |  |
| `net_amount` | `numeric(20,4)` | N |  | `0` |
| `tax_amount` | `numeric(20,4)` | N |  | `0` |
| `gross_amount` | `numeric(20,4)` | N |  | `0` |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `billing_run(document_type_id)` → `document_type(document_type_id)`
- `billing_run(billing_account_id)` → `billing_account(billing_account_id)`
- `billing_run(billing_contract_id)` → `billing_contract(billing_contract_id)`
- `billing_run(currency_id)` → `currency(currency_id)`
- `billing_run(reviewed_by)` → `app_account(account_id)` (optional)
- `billing_run(created_by)` → `app_account(account_id)`
- `billing_run(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `billing_charge`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `billing_charge_id` | `varchar(150)` | N | PK |  |
| `billing_run_id` | `varchar(120)` | N | FK |  |
| `billable_event_id` | `varchar(140)` | N | FK, UK |  |
| `rate_card_line_id` | `varchar(150)` | N | FK |  |
| `billed_quantity` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | Y | FK |  |
| `unit_rate` | `numeric(20,6)` | N |  |  |
| `net_amount` | `numeric(20,4)` | N |  |  |
| `tax_rule_id` | `uuid` | Y | FK |  |
| `tax_rate_percent` | `numeric(9,6)` | N |  | `0` |
| `tax_amount` | `numeric(20,4)` | N |  | `0` |
| `gross_amount` | `numeric(20,4)` | N |  |  |
| `currency_id` | `uuid` | N | FK |  |
| `calculated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `billing_charge(billing_run_id)` → `billing_run(billing_run_id)`; ON DELETE CASCADE
- `billing_charge(billable_event_id)` → `billable_event(billable_event_id)`
- `billing_charge(rate_card_line_id)` → `rate_card_line(rate_card_line_id)`
- `billing_charge(uom_id)` → `uom(uom_id)` (optional)
- `billing_charge(tax_rule_id)` → `tax_rule(tax_rule_id)` (optional)
- `billing_charge(currency_id)` → `currency(currency_id)`
- `billing_charge(created_by)` → `app_account(account_id)`

### `billing_invoice`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `billing_invoice_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `billing_run_id` | `varchar(120)` | N | FK, UK |  |
| `billing_account_id` | `uuid` | N | FK |  |
| `billing_contract_id` | `varchar(120)` | N | FK |  |
| `currency_id` | `uuid` | N | FK |  |
| `invoice_date` | `date` | N |  |  |
| `due_date` | `date` | N |  |  |
| `bill_to_name` | `varchar(200)` | N |  |  |
| `bill_to_tax_number` | `varchar(100)` | Y |  |  |
| `bill_to_address` | `text` | N |  |  |
| `net_amount` | `numeric(20,4)` | N |  | `0` |
| `adjustment_amount` | `numeric(20,4)` | N |  | `0` |
| `tax_amount` | `numeric(20,4)` | N |  | `0` |
| `gross_amount` | `numeric(20,4)` | N |  | `0` |
| `credited_amount` | `numeric(20,4)` | N |  | `0` |
| `paid_amount` | `numeric(20,4)` | N |  | `0` |
| `balance_due` | `numeric(20,4)` | N |  | `0` |
| `issued_at` | `timestamptz` | Y |  |  |
| `issued_by` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |
| `updated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `updated_by` | `uuid` | N | FK |  |

Relationships:
- `billing_invoice(document_type_id)` → `document_type(document_type_id)`
- `billing_invoice(billing_run_id)` → `billing_run(billing_run_id)`
- `billing_invoice(billing_account_id)` → `billing_account(billing_account_id)`
- `billing_invoice(billing_contract_id)` → `billing_contract(billing_contract_id)`
- `billing_invoice(currency_id)` → `currency(currency_id)`
- `billing_invoice(issued_by)` → `app_account(account_id)` (optional)
- `billing_invoice(created_by)` → `app_account(account_id)`
- `billing_invoice(updated_by)` → `app_account(account_id)`
- `billing_invoice(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `billing_invoice_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `billing_invoice_line_id` | `varchar(150)` | N | PK |  |
| `billing_invoice_id` | `varchar(120)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `billing_charge_id` | `varchar(150)` | N | FK, UK |  |
| `billing_service_id` | `uuid` | N | FK |  |
| `description` | `varchar(255)` | N |  |  |
| `quantity` | `numeric(20,6)` | N |  |  |
| `uom_id` | `uuid` | Y | FK |  |
| `unit_rate` | `numeric(20,6)` | N |  |  |
| `net_amount` | `numeric(20,4)` | N |  |  |
| `tax_rule_id` | `uuid` | Y | FK |  |
| `tax_rate_percent` | `numeric(9,6)` | N |  | `0` |
| `tax_amount` | `numeric(20,4)` | N |  | `0` |
| `gross_amount` | `numeric(20,4)` | N |  |  |

Relationships:
- `billing_invoice_line(billing_invoice_id)` → `billing_invoice(billing_invoice_id)`; ON DELETE CASCADE
- `billing_invoice_line(billing_charge_id)` → `billing_charge(billing_charge_id)`
- `billing_invoice_line(billing_service_id)` → `billing_service(billing_service_id)`
- `billing_invoice_line(uom_id)` → `uom(uom_id)` (optional)
- `billing_invoice_line(tax_rule_id)` → `tax_rule(tax_rule_id)` (optional)

### `billing_invoice_adjustment`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `invoice_adjustment_id` | `varchar(150)` | N | PK |  |
| `billing_invoice_id` | `varchar(120)` | N | FK |  |
| `line_no` | `integer` | N |  |  |
| `billing_adjustment_type_id` | `uuid` | N | FK |  |
| `reason_code_id` | `uuid` | N | FK |  |
| `description` | `varchar(255)` | N |  |  |
| `amount` | `numeric(20,4)` | N |  |  |
| `tax_rule_id` | `uuid` | Y | FK |  |
| `tax_rate_percent` | `numeric(9,6)` | N |  | `0` |
| `tax_amount` | `numeric(20,4)` | N |  | `0` |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `billing_invoice_adjustment(billing_invoice_id)` → `billing_invoice(billing_invoice_id)`; ON DELETE CASCADE
- `billing_invoice_adjustment(billing_adjustment_type_id)` → `billing_adjustment_type(billing_adjustment_type_id)`
- `billing_invoice_adjustment(reason_code_id)` → `reason_code(reason_code_id)`
- `billing_invoice_adjustment(tax_rule_id)` → `tax_rule(tax_rule_id)` (optional)
- `billing_invoice_adjustment(created_by)` → `app_account(account_id)`

### `billing_credit_note`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `credit_note_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `billing_invoice_id` | `varchar(120)` | N | FK |  |
| `currency_id` | `uuid` | N | FK |  |
| `credit_date` | `date` | N |  |  |
| `reason_code_id` | `uuid` | N | FK |  |
| `net_amount` | `numeric(20,4)` | N |  | `0` |
| `tax_amount` | `numeric(20,4)` | N |  | `0` |
| `gross_amount` | `numeric(20,4)` | N |  | `0` |
| `issued_at` | `timestamptz` | Y |  |  |
| `issued_by` | `uuid` | Y | FK |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `billing_credit_note(document_type_id)` → `document_type(document_type_id)`
- `billing_credit_note(billing_invoice_id)` → `billing_invoice(billing_invoice_id)`
- `billing_credit_note(currency_id)` → `currency(currency_id)`
- `billing_credit_note(reason_code_id)` → `reason_code(reason_code_id)`
- `billing_credit_note(issued_by)` → `app_account(account_id)` (optional)
- `billing_credit_note(created_by)` → `app_account(account_id)`
- `billing_credit_note(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `billing_credit_note_line`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `credit_note_line_id` | `varchar(150)` | N | PK |  |
| `credit_note_id` | `varchar(120)` | N | FK |  |
| `billing_invoice_line_id` | `varchar(150)` | Y | FK |  |
| `line_no` | `integer` | N |  |  |
| `description` | `varchar(255)` | N |  |  |
| `net_amount` | `numeric(20,4)` | N |  |  |
| `tax_rate_percent` | `numeric(9,6)` | N |  | `0` |
| `tax_amount` | `numeric(20,4)` | N |  | `0` |
| `gross_amount` | `numeric(20,4)` | N |  |  |

Relationships:
- `billing_credit_note_line(credit_note_id)` → `billing_credit_note(credit_note_id)`; ON DELETE CASCADE
- `billing_credit_note_line(billing_invoice_line_id)` → `billing_invoice_line(billing_invoice_line_id)` (optional)

### `billing_payment`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `payment_id` | `varchar(120)` | N | PK |  |
| `document_type_id` | `uuid` | N | FK |  |
| `status_id` | `uuid` | N | FK |  |
| `billing_account_id` | `uuid` | N | FK |  |
| `currency_id` | `uuid` | N | FK |  |
| `payment_method_id` | `uuid` | N | FK |  |
| `payment_date` | `date` | N |  |  |
| `amount` | `numeric(20,4)` | N |  |  |
| `unapplied_amount` | `numeric(20,4)` | N |  |  |
| `external_reference` | `varchar(150)` | Y |  |  |
| `notes` | `text` | Y |  |  |
| `created_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `created_by` | `uuid` | N | FK |  |

Relationships:
- `billing_payment(document_type_id)` → `document_type(document_type_id)`
- `billing_payment(billing_account_id)` → `billing_account(billing_account_id)`
- `billing_payment(currency_id)` → `currency(currency_id)`
- `billing_payment(payment_method_id)` → `payment_method(payment_method_id)`
- `billing_payment(created_by)` → `app_account(account_id)`
- `billing_payment(document_type_id, status_id)` → `document_status(document_type_id, status_id)`

### `billing_payment_allocation`

| Column | PostgreSQL type | Null | Key | Default |
|---|---|:---:|---|---|
| `payment_allocation_id` | `varchar(150)` | N | PK |  |
| `payment_id` | `varchar(120)` | N | FK |  |
| `allocation_no` | `integer` | N |  |  |
| `billing_invoice_id` | `varchar(120)` | N | FK |  |
| `allocated_amount` | `numeric(20,4)` | N |  |  |
| `allocated_at` | `timestamptz` | N |  | `clock_timestamp()` |
| `allocated_by` | `uuid` | N | FK |  |

Relationships:
- `billing_payment_allocation(payment_id)` → `billing_payment(payment_id)`; ON DELETE CASCADE
- `billing_payment_allocation(billing_invoice_id)` → `billing_invoice(billing_invoice_id)`
- `billing_payment_allocation(allocated_by)` → `app_account(account_id)`

## Reporting helper views

- `report_account_permission`
- `report_owner_scope`
- `report_warehouse_scope`
