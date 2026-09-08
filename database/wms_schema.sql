-- WMS baseline schema for PostgreSQL 15+
-- Transaction identifiers are varchar business identifiers, never serial/identity keys.
-- Master/security identifiers use UUIDs because their business codes can change.

BEGIN;

CREATE SCHEMA IF NOT EXISTS wms;
SET search_path TO wms, public;

-- -----------------------------------------------------------------------------
-- Security: accounts, roles, permissions, and menu visibility
-- -----------------------------------------------------------------------------

CREATE TABLE account_status (
    account_status_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code              varchar(30) NOT NULL UNIQUE,
    name              varchar(100) NOT NULL,
    description       text,
    allows_login      boolean NOT NULL DEFAULT false,
    is_active         boolean NOT NULL DEFAULT true,
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE authentication_policy (
    authentication_policy_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                    varchar(40) NOT NULL UNIQUE,
    name                    varchar(100) NOT NULL,
    max_failed_attempts     integer NOT NULL,
    lockout_seconds         integer NOT NULL,
    session_ttl_seconds     integer NOT NULL,
    is_default              boolean NOT NULL DEFAULT false,
    is_active               boolean NOT NULL DEFAULT true,
    created_at              timestamptz NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT ck_auth_policy_values CHECK (
        max_failed_attempts > 0 AND
        lockout_seconds > 0 AND
        session_ttl_seconds > 0
    )
);

CREATE UNIQUE INDEX uq_authentication_policy_default
    ON authentication_policy(is_default)
    WHERE is_default AND is_active;

CREATE TABLE session_revocation_reason (
    session_revocation_reason_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                         varchar(40) NOT NULL UNIQUE,
    name                         varchar(100) NOT NULL,
    description                  text,
    is_active                    boolean NOT NULL DEFAULT true
);

CREATE TABLE app_module (
    module_id     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code          varchar(50) NOT NULL UNIQUE,
    name          varchar(100) NOT NULL,
    display_order integer NOT NULL DEFAULT 0,
    is_active     boolean NOT NULL DEFAULT true
);

CREATE TABLE app_account (
    account_id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username           varchar(100) NOT NULL UNIQUE,
    email              varchar(254) UNIQUE,
    display_name       varchar(150) NOT NULL,
    password_hash      text,
    external_subject   varchar(255) UNIQUE,
    account_status_id  uuid NOT NULL REFERENCES account_status(account_status_id),
    authentication_policy_id uuid REFERENCES authentication_policy(authentication_policy_id),
    preferred_timezone varchar(50),
    failed_login_count integer NOT NULL DEFAULT 0,
    locked_until       timestamptz,
    last_login_at      timestamptz,
    created_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by         uuid REFERENCES app_account(account_id),
    updated_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by         uuid REFERENCES app_account(account_id),
    version_no         integer NOT NULL DEFAULT 1,
    CONSTRAINT ck_account_identity CHECK (
        password_hash IS NOT NULL OR external_subject IS NOT NULL
    ),
    CONSTRAINT ck_account_failed_login_count CHECK (failed_login_count >= 0),
    CONSTRAINT ck_account_version CHECK (version_no > 0)
);

-- Store only a cryptographic hash of the opaque session token. The raw token
-- is returned to the client by the application and must never be stored here.
CREATE TABLE app_session (
    session_id       varchar(120) PRIMARY KEY,
    account_id       uuid NOT NULL REFERENCES app_account(account_id) ON DELETE CASCADE,
    token_hash       varchar(128) NOT NULL UNIQUE,
    issued_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    expires_at       timestamptz NOT NULL,
    last_seen_at     timestamptz NOT NULL DEFAULT clock_timestamp(),
    revoked_at       timestamptz,
    session_revocation_reason_id uuid REFERENCES session_revocation_reason(session_revocation_reason_id),
    ip_address       inet,
    user_agent       text,
    CONSTRAINT ck_session_period CHECK (expires_at > issued_at),
    CONSTRAINT ck_session_revocation CHECK (
        revoked_at IS NULL OR revoked_at >= issued_at
    )
);

CREATE TABLE app_role (
    role_id      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code         varchar(50) NOT NULL UNIQUE,
    name         varchar(100) NOT NULL,
    description  text,
    is_active    boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by   uuid REFERENCES app_account(account_id)
);

CREATE TABLE app_permission (
    permission_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code          varchar(100) NOT NULL UNIQUE,
    name          varchar(150) NOT NULL,
    module_code   varchar(50) NOT NULL REFERENCES app_module(code),
    description   text,
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE app_menu (
    menu_id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_menu_id uuid REFERENCES app_menu(menu_id),
    code           varchar(50) NOT NULL UNIQUE,
    label          varchar(100) NOT NULL,
    route          varchar(255),
    icon_name      varchar(100),
    display_order  integer NOT NULL DEFAULT 0,
    require_all_permissions boolean NOT NULL DEFAULT false,
    is_active      boolean NOT NULL DEFAULT true,
    created_at     timestamptz NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT ck_menu_not_own_parent CHECK (parent_menu_id IS NULL OR parent_menu_id <> menu_id)
);

CREATE TABLE account_role (
    account_id uuid NOT NULL REFERENCES app_account(account_id) ON DELETE CASCADE,
    role_id    uuid NOT NULL REFERENCES app_role(role_id) ON DELETE CASCADE,
    valid_from timestamptz NOT NULL DEFAULT clock_timestamp(),
    valid_until timestamptz,
    assigned_by uuid REFERENCES app_account(account_id),
    PRIMARY KEY (account_id, role_id),
    CONSTRAINT ck_account_role_period CHECK (valid_until IS NULL OR valid_until > valid_from)
);

CREATE TABLE role_permission (
    role_id      uuid NOT NULL REFERENCES app_role(role_id) ON DELETE CASCADE,
    permission_id uuid NOT NULL REFERENCES app_permission(permission_id) ON DELETE CASCADE,
    granted_by   uuid REFERENCES app_account(account_id),
    granted_at   timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (role_id, permission_id)
);

-- A menu can require several permissions. app_menu.require_all_permissions
-- determines whether all or any of the mapped permissions are required.
CREATE TABLE menu_permission (
    menu_id       uuid NOT NULL REFERENCES app_menu(menu_id) ON DELETE CASCADE,
    permission_id uuid NOT NULL REFERENCES app_permission(permission_id) ON DELETE CASCADE,
    PRIMARY KEY (menu_id, permission_id)
);

-- -----------------------------------------------------------------------------
-- Organization and warehouse masters
-- -----------------------------------------------------------------------------

CREATE TABLE organization (
    organization_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code            varchar(40) NOT NULL UNIQUE,
    name            varchar(150) NOT NULL,
    legal_name      varchar(200),
    tax_number      varchar(100),
    timezone_name   varchar(50) NOT NULL,
    address_line_1  varchar(255),
    address_line_2  varchar(255),
    city            varchar(100),
    province        varchar(100),
    postal_code     varchar(20),
    country_code    varchar(2),
    is_active       boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by      uuid REFERENCES app_account(account_id),
    updated_at      timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by      uuid REFERENCES app_account(account_id)
);

CREATE TABLE warehouse (
    warehouse_id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id         uuid NOT NULL REFERENCES organization(organization_id),
    code                varchar(40) NOT NULL,
    name                varchar(150) NOT NULL,
    timezone_name       varchar(50) NOT NULL,
    address_line_1      varchar(255),
    address_line_2      varchar(255),
    city                varchar(100),
    province            varchar(100),
    postal_code         varchar(20),
    country_code        varchar(2),
    is_active           boolean NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by          uuid REFERENCES app_account(account_id),
    updated_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by          uuid REFERENCES app_account(account_id),
    UNIQUE (operator_id, code),
    UNIQUE (warehouse_id, operator_id)
);

-- Owners/clients allowed to keep inventory in a warehouse.
CREATE TABLE warehouse_owner (
    warehouse_id uuid NOT NULL REFERENCES warehouse(warehouse_id),
    owner_id     uuid NOT NULL REFERENCES organization(organization_id),
    is_active    boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by   uuid REFERENCES app_account(account_id),
    PRIMARY KEY (warehouse_id, owner_id),
    UNIQUE (owner_id, warehouse_id)
);

CREATE TABLE location_type (
    location_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code             varchar(40) NOT NULL UNIQUE,
    name             varchar(100) NOT NULL,
    description      text,
    allows_receiving boolean NOT NULL DEFAULT false,
    allows_storage   boolean NOT NULL DEFAULT false,
    allows_picking   boolean NOT NULL DEFAULT false,
    allows_shipping  boolean NOT NULL DEFAULT false,
    is_active        boolean NOT NULL DEFAULT true
);

CREATE TABLE warehouse_zone (
    zone_id       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id  uuid NOT NULL REFERENCES warehouse(warehouse_id),
    code          varchar(40) NOT NULL,
    name          varchar(100) NOT NULL,
    description   text,
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by    uuid REFERENCES app_account(account_id),
    UNIQUE (warehouse_id, code),
    UNIQUE (zone_id, warehouse_id)
);

CREATE TABLE warehouse_location (
    location_id       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id      uuid NOT NULL REFERENCES warehouse(warehouse_id),
    zone_id           uuid NOT NULL,
    location_type_id  uuid NOT NULL REFERENCES location_type(location_type_id),
    code              varchar(60) NOT NULL,
    barcode           varchar(100),
    aisle             varchar(20),
    bay               varchar(20),
    level_no          varchar(20),
    position_no       varchar(20),
    pick_sequence     integer NOT NULL DEFAULT 0,
    max_weight        numeric(20,6),
    max_volume        numeric(20,6),
    is_pick_face      boolean NOT NULL DEFAULT false,
    is_locked         boolean NOT NULL DEFAULT false,
    is_active         boolean NOT NULL DEFAULT true,
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid REFERENCES app_account(account_id),
    UNIQUE (warehouse_id, code),
    UNIQUE (warehouse_id, barcode),
    UNIQUE (location_id, warehouse_id),
    FOREIGN KEY (zone_id, warehouse_id)
        REFERENCES warehouse_zone(zone_id, warehouse_id),
    CONSTRAINT ck_location_capacity CHECK (
        (max_weight IS NULL OR max_weight >= 0) AND
        (max_volume IS NULL OR max_volume >= 0) AND
        pick_sequence >= 0
    )
);

CREATE TABLE account_owner_access (
    account_id uuid NOT NULL REFERENCES app_account(account_id) ON DELETE CASCADE,
    owner_id   uuid NOT NULL REFERENCES organization(organization_id) ON DELETE CASCADE,
    granted_by uuid REFERENCES app_account(account_id),
    granted_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (account_id, owner_id)
);

CREATE TABLE account_warehouse_access (
    account_id   uuid NOT NULL REFERENCES app_account(account_id) ON DELETE CASCADE,
    warehouse_id uuid NOT NULL REFERENCES warehouse(warehouse_id) ON DELETE CASCADE,
    granted_by   uuid REFERENCES app_account(account_id),
    granted_at   timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (account_id, warehouse_id)
);

-- -----------------------------------------------------------------------------
-- Partner masters
-- -----------------------------------------------------------------------------

CREATE TABLE partner_type (
    partner_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code            varchar(40) NOT NULL UNIQUE,
    name            varchar(100) NOT NULL,
    description     text,
    is_active       boolean NOT NULL DEFAULT true
);

CREATE TABLE business_partner (
    partner_id       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id         uuid NOT NULL REFERENCES organization(organization_id),
    code             varchar(40) NOT NULL,
    name             varchar(150) NOT NULL,
    legal_name       varchar(200),
    tax_number       varchar(100),
    email            varchar(254),
    phone            varchar(50),
    address_line_1   varchar(255),
    address_line_2   varchar(255),
    city             varchar(100),
    province         varchar(100),
    postal_code      varchar(20),
    country_code     varchar(2),
    is_active        boolean NOT NULL DEFAULT true,
    created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by       uuid REFERENCES app_account(account_id),
    updated_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by       uuid REFERENCES app_account(account_id),
    UNIQUE (owner_id, code),
    UNIQUE (owner_id, partner_id)
);

CREATE TABLE business_partner_type (
    partner_id      uuid NOT NULL REFERENCES business_partner(partner_id) ON DELETE CASCADE,
    partner_type_id uuid NOT NULL REFERENCES partner_type(partner_type_id),
    PRIMARY KEY (partner_id, partner_type_id)
);

-- -----------------------------------------------------------------------------
-- Product and inventory masters
-- -----------------------------------------------------------------------------

CREATE TABLE uom (
    uom_id       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code         varchar(20) NOT NULL UNIQUE,
    name         varchar(100) NOT NULL,
    decimal_scale smallint NOT NULL DEFAULT 0,
    is_active    boolean NOT NULL DEFAULT true,
    CONSTRAINT ck_uom_scale CHECK (decimal_scale BETWEEN 0 AND 6)
);

CREATE TABLE item_category (
    category_id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id           uuid NOT NULL REFERENCES organization(organization_id),
    parent_category_id uuid,
    code               varchar(40) NOT NULL,
    name               varchar(100) NOT NULL,
    is_active          boolean NOT NULL DEFAULT true,
    UNIQUE (owner_id, code),
    UNIQUE (owner_id, category_id),
    FOREIGN KEY (owner_id, parent_category_id)
        REFERENCES item_category(owner_id, category_id),
    CONSTRAINT ck_category_not_own_parent CHECK (
        parent_category_id IS NULL OR parent_category_id <> category_id
    )
);

CREATE TABLE item (
    item_id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id             uuid NOT NULL REFERENCES organization(organization_id),
    category_id          uuid,
    code                 varchar(60) NOT NULL,
    name                 varchar(200) NOT NULL,
    description          text,
    base_uom_id          uuid NOT NULL REFERENCES uom(uom_id),
    weight               numeric(20,6),
    volume               numeric(20,6),
    lot_controlled       boolean NOT NULL DEFAULT false,
    serial_controlled    boolean NOT NULL DEFAULT false,
    shelf_life_days      integer,
    minimum_receive_days integer,
    is_active            boolean NOT NULL DEFAULT true,
    created_at           timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by           uuid REFERENCES app_account(account_id),
    updated_at           timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by           uuid REFERENCES app_account(account_id),
    UNIQUE (owner_id, code),
    UNIQUE (owner_id, item_id),
    FOREIGN KEY (owner_id, category_id)
        REFERENCES item_category(owner_id, category_id),
    CONSTRAINT ck_item_measurements CHECK (
        (weight IS NULL OR weight >= 0) AND
        (volume IS NULL OR volume >= 0) AND
        (shelf_life_days IS NULL OR shelf_life_days >= 0) AND
        (minimum_receive_days IS NULL OR minimum_receive_days >= 0)
    )
);

CREATE TABLE item_uom (
    item_uom_id      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id          uuid NOT NULL REFERENCES item(item_id) ON DELETE CASCADE,
    uom_id           uuid NOT NULL REFERENCES uom(uom_id),
    conversion_to_base numeric(20,6) NOT NULL,
    length           numeric(20,6),
    width            numeric(20,6),
    height           numeric(20,6),
    weight           numeric(20,6),
    is_receiving_uom boolean NOT NULL DEFAULT true,
    is_picking_uom   boolean NOT NULL DEFAULT true,
    is_active        boolean NOT NULL DEFAULT true,
    UNIQUE (item_id, uom_id),
    CONSTRAINT ck_item_uom_conversion CHECK (conversion_to_base > 0)
);

CREATE TABLE item_barcode (
    item_barcode_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id         uuid NOT NULL REFERENCES item(item_id) ON DELETE CASCADE,
    uom_id          uuid REFERENCES uom(uom_id),
    barcode         varchar(100) NOT NULL UNIQUE,
    is_primary      boolean NOT NULL DEFAULT false,
    is_active       boolean NOT NULL DEFAULT true
);

CREATE UNIQUE INDEX uq_item_primary_barcode
    ON item_barcode(item_id)
    WHERE is_primary AND is_active;

CREATE TABLE inventory_status (
    inventory_status_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                varchar(40) NOT NULL UNIQUE,
    name                varchar(100) NOT NULL,
    description         text,
    is_allocatable      boolean NOT NULL DEFAULT false,
    is_pickable         boolean NOT NULL DEFAULT false,
    is_active           boolean NOT NULL DEFAULT true
);

CREATE TABLE quality_status (
    quality_status_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code              varchar(40) NOT NULL UNIQUE,
    name              varchar(100) NOT NULL,
    description       text,
    is_active         boolean NOT NULL DEFAULT true
);

CREATE TABLE inspection_result (
    inspection_result_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                 varchar(40) NOT NULL UNIQUE,
    name                 varchar(100) NOT NULL,
    description          text,
    is_accepted          boolean NOT NULL DEFAULT false,
    is_active            boolean NOT NULL DEFAULT true
);

CREATE TABLE quarantine_disposition_type (
    quarantine_disposition_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                           varchar(40) NOT NULL UNIQUE,
    name                           varchar(100) NOT NULL,
    description                    text,
    releases_to_available          boolean NOT NULL DEFAULT false,
    requires_reinspection          boolean NOT NULL DEFAULT false,
    removes_inventory              boolean NOT NULL DEFAULT false,
    removal_movement_type_id       uuid,
    is_active                      boolean NOT NULL DEFAULT true,
    CONSTRAINT ck_quarantine_disposition_action CHECK (
        (CASE WHEN releases_to_available THEN 1 ELSE 0 END) +
        (CASE WHEN requires_reinspection THEN 1 ELSE 0 END) +
        (CASE WHEN removes_inventory THEN 1 ELSE 0 END) = 1
    )
);

CREATE TABLE reason_code (
    reason_code_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    module_code    varchar(50) NOT NULL REFERENCES app_module(code),
    code           varchar(40) NOT NULL,
    name           varchar(100) NOT NULL,
    description    text,
    requires_note  boolean NOT NULL DEFAULT false,
    is_active      boolean NOT NULL DEFAULT true,
    UNIQUE (module_code, code)
);

CREATE TABLE handling_unit_type (
    handling_unit_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                  varchar(40) NOT NULL UNIQUE,
    name                  varchar(100) NOT NULL,
    max_weight            numeric(20,6),
    max_volume            numeric(20,6),
    is_active             boolean NOT NULL DEFAULT true,
    CONSTRAINT ck_hu_type_capacity CHECK (
        (max_weight IS NULL OR max_weight >= 0) AND
        (max_volume IS NULL OR max_volume >= 0)
    )
);

-- -----------------------------------------------------------------------------
-- Workflow, tasks, strategies, and document numbering masters
-- -----------------------------------------------------------------------------

CREATE TABLE document_type (
    document_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code             varchar(40) NOT NULL UNIQUE,
    name             varchar(100) NOT NULL,
    module_code      varchar(50) NOT NULL REFERENCES app_module(code),
    description      text,
    is_active        boolean NOT NULL DEFAULT true
);

CREATE TABLE document_status (
    status_id       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    document_type_id uuid NOT NULL REFERENCES document_type(document_type_id),
    code            varchar(40) NOT NULL,
    name            varchar(100) NOT NULL,
    description     text,
    is_initial      boolean NOT NULL DEFAULT false,
    is_final        boolean NOT NULL DEFAULT false,
    is_cancelled    boolean NOT NULL DEFAULT false,
    display_order   integer NOT NULL DEFAULT 0,
    is_active       boolean NOT NULL DEFAULT true,
    UNIQUE (document_type_id, code),
    UNIQUE (document_type_id, status_id)
);

CREATE UNIQUE INDEX uq_document_status_initial
    ON document_status(document_type_id)
    WHERE is_initial;

CREATE TABLE document_status_transition (
    transition_id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    document_type_id     uuid NOT NULL REFERENCES document_type(document_type_id),
    from_status_id       uuid NOT NULL,
    to_status_id         uuid NOT NULL,
    required_permission_id uuid REFERENCES app_permission(permission_id),
    is_active            boolean NOT NULL DEFAULT true,
    UNIQUE (document_type_id, from_status_id, to_status_id),
    FOREIGN KEY (document_type_id, from_status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (document_type_id, to_status_id)
        REFERENCES document_status(document_type_id, status_id),
    CONSTRAINT ck_status_transition_different CHECK (from_status_id <> to_status_id)
);

CREATE TABLE task_type (
    task_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code         varchar(40) NOT NULL UNIQUE,
    name         varchar(100) NOT NULL,
    description  text,
    is_active    boolean NOT NULL DEFAULT true
);

CREATE TABLE task_status (
    task_status_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code           varchar(40) NOT NULL UNIQUE,
    name           varchar(100) NOT NULL,
    is_initial     boolean NOT NULL DEFAULT false,
    is_final       boolean NOT NULL DEFAULT false,
    is_cancelled   boolean NOT NULL DEFAULT false,
    is_active      boolean NOT NULL DEFAULT true
);

CREATE TABLE task_status_transition (
    task_status_transition_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    from_status_id            uuid NOT NULL REFERENCES task_status(task_status_id),
    to_status_id              uuid NOT NULL REFERENCES task_status(task_status_id),
    required_permission_id    uuid REFERENCES app_permission(permission_id),
    is_active                 boolean NOT NULL DEFAULT true,
    UNIQUE (from_status_id, to_status_id),
    CONSTRAINT ck_task_status_transition_different CHECK (from_status_id <> to_status_id)
);

CREATE TABLE task_priority (
    task_priority_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code             varchar(40) NOT NULL UNIQUE,
    name             varchar(100) NOT NULL,
    priority_value   integer NOT NULL UNIQUE,
    is_active        boolean NOT NULL DEFAULT true
);

CREATE TABLE putaway_strategy (
    putaway_strategy_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id            uuid REFERENCES organization(organization_id),
    warehouse_id        uuid REFERENCES warehouse(warehouse_id),
    code                varchar(40) NOT NULL,
    name                varchar(100) NOT NULL,
    description         text,
    is_active           boolean NOT NULL DEFAULT true,
    UNIQUE NULLS NOT DISTINCT (owner_id, warehouse_id, code)
);

CREATE TABLE putaway_strategy_rule (
    rule_id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    putaway_strategy_id  uuid NOT NULL REFERENCES putaway_strategy(putaway_strategy_id) ON DELETE CASCADE,
    sequence_no          integer NOT NULL,
    category_id          uuid REFERENCES item_category(category_id),
    location_type_id     uuid REFERENCES location_type(location_type_id),
    zone_id              uuid REFERENCES warehouse_zone(zone_id),
    minimum_empty_percent numeric(7,4),
    is_active            boolean NOT NULL DEFAULT true,
    UNIQUE (putaway_strategy_id, sequence_no),
    CONSTRAINT ck_putaway_empty_percent CHECK (
        minimum_empty_percent IS NULL OR minimum_empty_percent BETWEEN 0 AND 100
    )
);

CREATE TABLE picking_strategy (
    picking_strategy_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id            uuid REFERENCES organization(organization_id),
    warehouse_id        uuid REFERENCES warehouse(warehouse_id),
    code                varchar(40) NOT NULL,
    name                varchar(100) NOT NULL,
    description         text,
    is_active           boolean NOT NULL DEFAULT true,
    UNIQUE NULLS NOT DISTINCT (owner_id, warehouse_id, code)
);

CREATE TABLE picking_sort_method (
    picking_sort_method_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                   varchar(40) NOT NULL UNIQUE,
    name                   varchar(100) NOT NULL,
    description            text,
    is_active              boolean NOT NULL DEFAULT true
);

CREATE TABLE picking_strategy_rule (
    rule_id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    picking_strategy_id  uuid NOT NULL REFERENCES picking_strategy(picking_strategy_id) ON DELETE CASCADE,
    sequence_no          integer NOT NULL,
    inventory_status_id  uuid REFERENCES inventory_status(inventory_status_id),
    zone_id              uuid REFERENCES warehouse_zone(zone_id),
    picking_sort_method_id uuid NOT NULL REFERENCES picking_sort_method(picking_sort_method_id),
    is_active            boolean NOT NULL DEFAULT true,
    UNIQUE (picking_strategy_id, sequence_no)
);

CREATE TABLE document_number_rule (
    document_number_rule_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    document_type_id        uuid NOT NULL REFERENCES document_type(document_type_id),
    prefix                  varchar(20) NOT NULL,
    separator               varchar(3) NOT NULL DEFAULT '-',
    sequence_length         smallint NOT NULL DEFAULT 6,
    include_partner_code    boolean NOT NULL DEFAULT true,
    include_warehouse_code  boolean NOT NULL DEFAULT true,
    is_active               boolean NOT NULL DEFAULT true,
    effective_from          date NOT NULL DEFAULT current_date,
    effective_until         date,
    created_at              timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by              uuid REFERENCES app_account(account_id),
    CONSTRAINT ck_number_rule_length CHECK (sequence_length BETWEEN 3 AND 18),
    CONSTRAINT ck_number_rule_period CHECK (
        effective_until IS NULL OR effective_until >= effective_from
    )
);

CREATE UNIQUE INDEX uq_active_document_number_rule
    ON document_number_rule(document_type_id)
    WHERE is_active;

-- The counter is global per document type per business date. Vendor, owner, and
-- warehouse are deliberately not part of this key.
CREATE TABLE document_daily_counter (
    document_type_id uuid NOT NULL REFERENCES document_type(document_type_id),
    business_date    date NOT NULL,
    last_number      bigint NOT NULL,
    updated_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (document_type_id, business_date),
    CONSTRAINT ck_daily_counter_positive CHECK (last_number > 0)
);

CREATE OR REPLACE FUNCTION generate_document_id(
    p_document_type_code varchar,
    p_partner_code       varchar,
    p_warehouse_code     varchar,
    p_business_date      date
) RETURNS varchar
LANGUAGE plpgsql
AS $$
DECLARE
    v_document_type_id uuid;
    v_prefix           varchar(20);
    v_separator        varchar(3);
    v_sequence_length  smallint;
    v_include_partner  boolean;
    v_include_warehouse boolean;
    v_next_number      bigint;
    v_result           varchar;
BEGIN
    IF p_business_date IS NULL THEN
        RAISE EXCEPTION 'Business date is required';
    END IF;

    SELECT dt.document_type_id,
           nr.prefix,
           nr.separator,
           nr.sequence_length,
           nr.include_partner_code,
           nr.include_warehouse_code
      INTO STRICT v_document_type_id,
                  v_prefix,
                  v_separator,
                  v_sequence_length,
                  v_include_partner,
                  v_include_warehouse
      FROM document_type dt
      JOIN document_number_rule nr
        ON nr.document_type_id = dt.document_type_id
       AND nr.is_active
       AND nr.effective_from <= p_business_date
       AND (nr.effective_until IS NULL OR nr.effective_until >= p_business_date)
     WHERE dt.code = p_document_type_code
       AND dt.is_active;

    IF v_include_partner AND NULLIF(trim(p_partner_code), '') IS NULL THEN
        RAISE EXCEPTION 'Partner code is required for document type %', p_document_type_code;
    END IF;

    IF v_include_warehouse AND NULLIF(trim(p_warehouse_code), '') IS NULL THEN
        RAISE EXCEPTION 'Warehouse code is required for document type %', p_document_type_code;
    END IF;

    INSERT INTO document_daily_counter (
        document_type_id, business_date, last_number
    ) VALUES (
        v_document_type_id, p_business_date, 1
    )
    ON CONFLICT (document_type_id, business_date)
    DO UPDATE SET
        last_number = document_daily_counter.last_number + 1,
        updated_at = clock_timestamp()
    RETURNING last_number INTO v_next_number;

    IF length(v_next_number::text) > v_sequence_length THEN
        RAISE EXCEPTION 'Daily number % exceeds configured length % for document type %',
            v_next_number, v_sequence_length, p_document_type_code;
    END IF;

    v_result := v_prefix;

    IF v_include_partner THEN
        v_result := v_result || v_separator || upper(trim(p_partner_code));
    END IF;

    IF v_include_warehouse THEN
        v_result := v_result || v_separator || upper(trim(p_warehouse_code));
    END IF;

    v_result := v_result
        || v_separator || to_char(p_business_date, 'YYYYMMDD')
        || v_separator || lpad(v_next_number::text, v_sequence_length, '0');

    RETURN v_result;
EXCEPTION
    WHEN NO_DATA_FOUND THEN
        RAISE EXCEPTION 'No active number rule found for document type % on %',
            p_document_type_code, p_business_date;
    WHEN TOO_MANY_ROWS THEN
        RAISE EXCEPTION 'Multiple active number rules found for document type % on %',
            p_document_type_code, p_business_date;
END;
$$;

-- -----------------------------------------------------------------------------
-- Traceability entities
-- -----------------------------------------------------------------------------

CREATE TABLE inventory_lot (
    lot_id          varchar(120) PRIMARY KEY,
    owner_id        uuid NOT NULL REFERENCES organization(organization_id),
    item_id         uuid NOT NULL REFERENCES item(item_id),
    lot_number      varchar(100) NOT NULL,
    manufacture_date date,
    expiry_date     date,
    quality_status_id uuid REFERENCES quality_status(quality_status_id),
    created_at      timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by      uuid REFERENCES app_account(account_id),
    UNIQUE (owner_id, item_id, lot_number),
    UNIQUE (lot_id, owner_id, item_id),
    CONSTRAINT fk_identity_inventory_lot_item FOREIGN KEY (owner_id, item_id)
        REFERENCES item(owner_id, item_id),
    CONSTRAINT ck_identity_lot_number_nonempty CHECK (length(btrim(lot_number)) > 0),
    CONSTRAINT ck_lot_dates CHECK (
        expiry_date IS NULL OR manufacture_date IS NULL OR expiry_date >= manufacture_date
    )
);

CREATE TABLE serial_number (
    serial_id       varchar(160) PRIMARY KEY,
    owner_id        uuid NOT NULL REFERENCES organization(organization_id),
    item_id         uuid NOT NULL REFERENCES item(item_id),
    serial_no       varchar(120) NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by      uuid REFERENCES app_account(account_id),
    UNIQUE (owner_id, item_id, serial_no),
    UNIQUE (serial_id, owner_id, item_id),
    CONSTRAINT fk_identity_serial_number_item FOREIGN KEY (owner_id, item_id)
        REFERENCES item(owner_id, item_id),
    CONSTRAINT ck_identity_serial_no_nonempty CHECK (length(btrim(serial_no)) > 0)
);

CREATE TABLE handling_unit (
    handling_unit_id      varchar(120) PRIMARY KEY,
    warehouse_id          uuid NOT NULL REFERENCES warehouse(warehouse_id),
    owner_id              uuid NOT NULL REFERENCES organization(organization_id),
    handling_unit_type_id uuid NOT NULL REFERENCES handling_unit_type(handling_unit_type_id),
    parent_handling_unit_id varchar(120) REFERENCES handling_unit(handling_unit_id),
    current_location_id   uuid REFERENCES warehouse_location(location_id),
    barcode               varchar(120) NOT NULL UNIQUE,
    is_closed             boolean NOT NULL DEFAULT false,
    created_at            timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by            uuid REFERENCES app_account(account_id),
    CONSTRAINT ck_hu_not_own_parent CHECK (
        parent_handling_unit_id IS NULL OR parent_handling_unit_id <> handling_unit_id
    ),
    UNIQUE (handling_unit_id, owner_id, warehouse_id),
    CONSTRAINT fk_identity_hu_parent FOREIGN KEY (parent_handling_unit_id, owner_id, warehouse_id)
        REFERENCES handling_unit(handling_unit_id, owner_id, warehouse_id),
    CONSTRAINT fk_identity_hu_location FOREIGN KEY (current_location_id, warehouse_id)
        REFERENCES warehouse_location(location_id, warehouse_id),
    CONSTRAINT ck_identity_barcode_nonempty CHECK (length(btrim(barcode)) > 0),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id)
);

-- -----------------------------------------------------------------------------
-- Inbound and receiving transactions
-- -----------------------------------------------------------------------------

CREATE TABLE purchase_order (
    purchase_order_id  varchar(120) PRIMARY KEY,
    document_type_id   uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id          uuid NOT NULL,
    owner_id           uuid NOT NULL REFERENCES organization(organization_id),
    vendor_id          uuid NOT NULL REFERENCES business_partner(partner_id),
    warehouse_id       uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date      date NOT NULL,
    purchase_order_no  varchar(120) NOT NULL,
    ordered_at         timestamptz NOT NULL,
    expected_arrival_at timestamptz,
    notes              text,
    created_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by         uuid NOT NULL REFERENCES app_account(account_id),
    updated_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by         uuid REFERENCES app_account(account_id),
    version_no         integer NOT NULL DEFAULT 1,
    UNIQUE (owner_id, purchase_order_no),
    UNIQUE (purchase_order_id, owner_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, vendor_id)
        REFERENCES business_partner(owner_id, partner_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    CONSTRAINT ck_purchase_order_version CHECK (version_no > 0)
);

CREATE TABLE purchase_order_line (
    purchase_order_line_id varchar(150) PRIMARY KEY,
    purchase_order_id      varchar(120) NOT NULL,
    owner_id               uuid NOT NULL,
    line_no                integer NOT NULL,
    item_id                uuid NOT NULL,
    ordered_qty            numeric(20,6) NOT NULL,
    over_receipt_tolerance_pct numeric(7,4) NOT NULL DEFAULT 0,
    under_receipt_tolerance_pct numeric(7,4) NOT NULL DEFAULT 0,
    uom_id                 uuid NOT NULL REFERENCES uom(uom_id),
    vendor_item_code       varchar(100),
    expected_lot_no        varchar(100),
    expected_expiry_date   date,
    notes                  text,
    created_at             timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by             uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (purchase_order_id, line_no),
    UNIQUE (purchase_order_line_id, purchase_order_id),
    FOREIGN KEY (purchase_order_id, owner_id)
        REFERENCES purchase_order(purchase_order_id, owner_id) ON DELETE CASCADE,
    FOREIGN KEY (owner_id, item_id)
        REFERENCES item(owner_id, item_id),
    CONSTRAINT ck_purchase_order_line_no CHECK (line_no > 0),
    CONSTRAINT ck_purchase_order_line_qty CHECK (
        ordered_qty > 0
        AND over_receipt_tolerance_pct BETWEEN 0 AND 100
        AND under_receipt_tolerance_pct BETWEEN 0 AND 100
    )
);

CREATE TABLE inbound_order (
    inbound_id          varchar(120) PRIMARY KEY,
    document_type_id    uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id           uuid NOT NULL,
    owner_id            uuid NOT NULL REFERENCES organization(organization_id),
    vendor_id           uuid NOT NULL REFERENCES business_partner(partner_id),
    warehouse_id        uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date       date NOT NULL,
    expected_arrival_at timestamptz,
    external_reference  varchar(120),
    supplier_reference  varchar(120),
    notes                text,
    created_at           timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by           uuid NOT NULL REFERENCES app_account(account_id),
    updated_at           timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by           uuid REFERENCES app_account(account_id),
    version_no           integer NOT NULL DEFAULT 1,
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, vendor_id)
        REFERENCES business_partner(owner_id, partner_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    CONSTRAINT ck_inbound_version CHECK (version_no > 0)
);

CREATE TABLE inbound_order_line (
    inbound_line_id  varchar(150) PRIMARY KEY,
    inbound_id       varchar(120) NOT NULL REFERENCES inbound_order(inbound_id) ON DELETE CASCADE,
    purchase_order_line_id varchar(150) REFERENCES purchase_order_line(purchase_order_line_id),
    line_no          integer NOT NULL,
    item_id          uuid NOT NULL REFERENCES item(item_id),
    expected_qty     numeric(20,6) NOT NULL,
    uom_id           uuid NOT NULL REFERENCES uom(uom_id),
    expected_lot_no  varchar(100),
    expected_expiry_date date,
    customer_line_reference varchar(100),
    notes            text,
    created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by       uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (inbound_id, line_no),
    CONSTRAINT ck_inbound_line_no CHECK (line_no > 0),
    CONSTRAINT ck_inbound_expected_qty CHECK (expected_qty > 0)
);

CREATE TABLE receipt (
    receipt_id       varchar(120) PRIMARY KEY,
    document_type_id uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id        uuid NOT NULL,
    inbound_id       varchar(120) REFERENCES inbound_order(inbound_id),
    owner_id         uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id     uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date    date NOT NULL,
    received_at      timestamptz NOT NULL,
    dock_location_id uuid REFERENCES warehouse_location(location_id),
    vehicle_number   varchar(60),
    seal_number      varchar(60),
    delivery_note_no varchar(100),
    notes            text,
    created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by       uuid NOT NULL REFERENCES app_account(account_id),
    updated_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by       uuid REFERENCES app_account(account_id),
    version_no       integer NOT NULL DEFAULT 1,
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    CONSTRAINT ck_receipt_version CHECK (version_no > 0)
);

CREATE TABLE receipt_line (
    receipt_line_id  varchar(150) PRIMARY KEY,
    receipt_id       varchar(120) NOT NULL REFERENCES receipt(receipt_id) ON DELETE CASCADE,
    inbound_line_id  varchar(150) REFERENCES inbound_order_line(inbound_line_id),
    line_no          integer NOT NULL,
    item_id          uuid NOT NULL REFERENCES item(item_id),
    received_qty     numeric(20,6) NOT NULL,
    rejected_qty     numeric(20,6) NOT NULL DEFAULT 0,
    exception_notes  text,
    exception_type_code varchar(40),
    uom_id           uuid NOT NULL REFERENCES uom(uom_id),
    created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by       uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (receipt_id, line_no),
    CONSTRAINT ck_receipt_line_no CHECK (line_no > 0),
    CONSTRAINT ck_receipt_quantities CHECK (
        received_qty > 0 AND rejected_qty >= 0 AND rejected_qty <= received_qty
    )
);

-- One receipt line may be split across lots, handling units, locations, and QC
-- outcomes. Each row is a homogeneous received stock batch.
CREATE TABLE receipt_inventory (
    receipt_inventory_id       varchar(160) PRIMARY KEY,
    receipt_line_id            varchar(150) NOT NULL REFERENCES receipt_line(receipt_line_id) ON DELETE CASCADE,
    item_id                    uuid NOT NULL REFERENCES item(item_id),
    source_qty                 numeric(20,6) NOT NULL,
    source_uom_id              uuid NOT NULL REFERENCES uom(uom_id),
    base_qty                   numeric(20,6) NOT NULL,
    base_uom_id                uuid NOT NULL REFERENCES uom(uom_id),
    lot_id                     varchar(120) REFERENCES inventory_lot(lot_id),
    handling_unit_id           varchar(120) REFERENCES handling_unit(handling_unit_id),
    received_location_id       uuid NOT NULL REFERENCES warehouse_location(location_id),
    initial_inventory_status_id uuid NOT NULL REFERENCES inventory_status(inventory_status_id),
    initial_balance_id          varchar(160),
    created_at                 timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by                 uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_receipt_inventory_qty CHECK (source_qty > 0 AND base_qty > 0)
);

CREATE TABLE receipt_line_serial (
    receipt_inventory_id varchar(160) NOT NULL REFERENCES receipt_inventory(receipt_inventory_id) ON DELETE CASCADE,
    serial_id            varchar(160) NOT NULL REFERENCES serial_number(serial_id),
    PRIMARY KEY (receipt_inventory_id, serial_id),
    UNIQUE (serial_id)
);

CREATE TABLE quality_inspection (
    inspection_id       varchar(120) PRIMARY KEY,
    receipt_inventory_id varchar(160) NOT NULL REFERENCES receipt_inventory(receipt_inventory_id),
    parent_inspection_id varchar(120) REFERENCES quality_inspection(inspection_id),
    source_balance_id   varchar(160),
    quality_status_id   uuid NOT NULL REFERENCES quality_status(quality_status_id),
    inspection_result_id uuid REFERENCES inspection_result(inspection_result_id),
    inspected_qty       numeric(20,6) NOT NULL,
    passed_qty          numeric(20,6) NOT NULL DEFAULT 0,
    failed_qty          numeric(20,6) NOT NULL DEFAULT 0,
    inspected_at        timestamptz,
    inspected_by        uuid REFERENCES app_account(account_id),
    cancelled_at        timestamptz,
    cancelled_by        uuid REFERENCES app_account(account_id),
    cancellation_reason text,
    notes               text,
    created_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by          uuid NOT NULL REFERENCES app_account(account_id),
    updated_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by          uuid REFERENCES app_account(account_id),
    version_no          integer NOT NULL DEFAULT 1,
    CONSTRAINT ck_inspection_quantities CHECK (
        inspected_qty > 0 AND passed_qty >= 0 AND failed_qty >= 0
        AND passed_qty + failed_qty <= inspected_qty
    ),
    CONSTRAINT ck_quality_inspection_version CHECK (version_no > 0)
);

CREATE TABLE putaway_task (
    putaway_task_id   varchar(120) PRIMARY KEY,
    task_type_id      uuid NOT NULL REFERENCES task_type(task_type_id),
    task_status_id    uuid NOT NULL REFERENCES task_status(task_status_id),
    task_priority_id  uuid NOT NULL REFERENCES task_priority(task_priority_id),
    receipt_inventory_id varchar(160) NOT NULL REFERENCES receipt_inventory(receipt_inventory_id),
    inspection_id     varchar(120) NOT NULL UNIQUE REFERENCES quality_inspection(inspection_id),
    source_balance_id varchar(160) NOT NULL,
    owner_id          uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id      uuid NOT NULL REFERENCES warehouse(warehouse_id),
    item_id           uuid NOT NULL REFERENCES item(item_id),
    lot_id            varchar(120) REFERENCES inventory_lot(lot_id),
    handling_unit_id  varchar(120) REFERENCES handling_unit(handling_unit_id),
    source_location_id uuid NOT NULL REFERENCES warehouse_location(location_id),
    target_location_id uuid NOT NULL REFERENCES warehouse_location(location_id),
    planned_qty       numeric(20,6) NOT NULL,
    completed_qty     numeric(20,6) NOT NULL DEFAULT 0,
    uom_id            uuid NOT NULL REFERENCES uom(uom_id),
    assigned_to       uuid REFERENCES app_account(account_id),
    started_at        timestamptz,
    completed_at      timestamptz,
    inventory_movement_id varchar(140),
    resulting_balance_id varchar(160),
    reversal_movement_id varchar(140),
    reversed_at         timestamptz,
    reversed_by         uuid REFERENCES app_account(account_id),
    reversal_reason     text,
    replacement_inspection_id varchar(120) REFERENCES quality_inspection(inspection_id),
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    updated_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by        uuid REFERENCES app_account(account_id),
    version_no        integer NOT NULL DEFAULT 1,
    CONSTRAINT ck_putaway_qty CHECK (
        planned_qty > 0 AND completed_qty >= 0 AND completed_qty <= planned_qty
    ),
    CONSTRAINT ck_putaway_task_version CHECK (version_no > 0)
);

-- -----------------------------------------------------------------------------
-- Inventory balance, reservation, and immutable movement ledger
-- -----------------------------------------------------------------------------

CREATE TABLE movement_type (
    movement_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code             varchar(40) NOT NULL UNIQUE,
    name             varchar(100) NOT NULL,
    description      text,
    is_active        boolean NOT NULL DEFAULT true
);

CREATE TABLE internal_move_type (
    internal_move_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                  varchar(40) NOT NULL UNIQUE,
    name                  varchar(100) NOT NULL,
    description           text,
    requires_approval     boolean NOT NULL DEFAULT false,
    is_active             boolean NOT NULL DEFAULT true
);

CREATE TABLE inventory_adjustment_type (
    inventory_adjustment_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                         varchar(40) NOT NULL UNIQUE,
    name                         varchar(100) NOT NULL,
    description                  text,
    quantity_effect              smallint NOT NULL,
    requires_approval            boolean NOT NULL DEFAULT true,
    is_active                    boolean NOT NULL DEFAULT true,
    CONSTRAINT ck_adjustment_type_effect CHECK (quantity_effect IN (-1, 1))
);

CREATE TABLE stock_count_type (
    stock_count_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                varchar(40) NOT NULL UNIQUE,
    name                varchar(100) NOT NULL,
    description         text,
    requires_freeze     boolean NOT NULL DEFAULT false,
    is_active           boolean NOT NULL DEFAULT true
);

ALTER TABLE quarantine_disposition_type
    ADD CONSTRAINT fk_quarantine_disposition_removal_movement
    FOREIGN KEY (removal_movement_type_id) REFERENCES movement_type(movement_type_id);

CREATE TABLE inventory_balance (
    balance_id          varchar(160) PRIMARY KEY,
    owner_id            uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id        uuid NOT NULL REFERENCES warehouse(warehouse_id),
    location_id         uuid NOT NULL REFERENCES warehouse_location(location_id),
    item_id             uuid NOT NULL REFERENCES item(item_id),
    lot_id              varchar(120) REFERENCES inventory_lot(lot_id),
    handling_unit_id    varchar(120) REFERENCES handling_unit(handling_unit_id),
    inventory_status_id uuid NOT NULL REFERENCES inventory_status(inventory_status_id),
    on_hand_qty         numeric(20,6) NOT NULL DEFAULT 0,
    reserved_qty        numeric(20,6) NOT NULL DEFAULT 0,
    uom_id              uuid NOT NULL REFERENCES uom(uom_id),
    version_no          bigint NOT NULL DEFAULT 1,
    updated_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE NULLS NOT DISTINCT (
        owner_id, warehouse_id, location_id, item_id,
        lot_id, handling_unit_id, inventory_status_id
    ),
    UNIQUE (balance_id, owner_id, item_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    FOREIGN KEY (owner_id, item_id)
        REFERENCES item(owner_id, item_id),
    FOREIGN KEY (location_id, warehouse_id)
        REFERENCES warehouse_location(location_id, warehouse_id),
    FOREIGN KEY (lot_id, owner_id, item_id)
        REFERENCES inventory_lot(lot_id, owner_id, item_id),
    FOREIGN KEY (handling_unit_id, owner_id, warehouse_id)
        REFERENCES handling_unit(handling_unit_id, owner_id, warehouse_id),
    CONSTRAINT ck_inventory_balance_qty CHECK (
        on_hand_qty >= 0 AND reserved_qty >= 0 AND reserved_qty <= on_hand_qty
    ),
    CONSTRAINT ck_inventory_balance_version CHECK (version_no > 0)
);

CREATE TABLE inventory_movement (
    movement_id        varchar(140) PRIMARY KEY,
    movement_type_id   uuid NOT NULL REFERENCES movement_type(movement_type_id),
    owner_id           uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id       uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date      date NOT NULL,
    occurred_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    item_id            uuid NOT NULL REFERENCES item(item_id),
    lot_id             varchar(120) REFERENCES inventory_lot(lot_id),
    serial_id          varchar(160) REFERENCES serial_number(serial_id),
    handling_unit_id   varchar(120) REFERENCES handling_unit(handling_unit_id),
    from_location_id   uuid REFERENCES warehouse_location(location_id),
    to_location_id     uuid REFERENCES warehouse_location(location_id),
    from_status_id     uuid REFERENCES inventory_status(inventory_status_id),
    to_status_id       uuid REFERENCES inventory_status(inventory_status_id),
    quantity           numeric(20,6) NOT NULL,
    uom_id             uuid NOT NULL REFERENCES uom(uom_id),
    source_document_id varchar(140) NOT NULL,
    source_line_id     varchar(160),
    operation_key       varchar(160),
    operation_fingerprint varchar(64),
    reason_code_id     uuid REFERENCES reason_code(reason_code_id),
    notes               text,
    created_by          uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    FOREIGN KEY (owner_id, item_id)
        REFERENCES item(owner_id, item_id),
    FOREIGN KEY (lot_id, owner_id, item_id)
        REFERENCES inventory_lot(lot_id, owner_id, item_id),
    FOREIGN KEY (serial_id, owner_id, item_id)
        REFERENCES serial_number(serial_id, owner_id, item_id),
    FOREIGN KEY (handling_unit_id, owner_id, warehouse_id)
        REFERENCES handling_unit(handling_unit_id, owner_id, warehouse_id),
    FOREIGN KEY (from_location_id, warehouse_id)
        REFERENCES warehouse_location(location_id, warehouse_id),
    FOREIGN KEY (to_location_id, warehouse_id)
        REFERENCES warehouse_location(location_id, warehouse_id),
    CONSTRAINT ck_movement_qty CHECK (quantity > 0),
    CONSTRAINT ck_movement_has_effect CHECK (
        from_location_id IS DISTINCT FROM to_location_id OR
        from_status_id IS DISTINCT FROM to_status_id
    )
);

-- One authoritative pointer per serialized unit. Current location, lot,
-- handling unit, status and UOM are derived by joining its inventory balance.
CREATE TABLE serial_inventory (
    serial_id          varchar(160) PRIMARY KEY,
    balance_id         varchar(160) NOT NULL,
    owner_id           uuid NOT NULL,
    item_id            uuid NOT NULL,
    version_no         bigint NOT NULL DEFAULT 1,
    updated_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    FOREIGN KEY (serial_id, owner_id, item_id)
        REFERENCES serial_number(serial_id, owner_id, item_id),
    FOREIGN KEY (balance_id, owner_id, item_id)
        REFERENCES inventory_balance(balance_id, owner_id, item_id),
    CONSTRAINT ck_serial_inventory_version CHECK (version_no > 0)
);

ALTER TABLE putaway_task
    ADD CONSTRAINT fk_putaway_source_balance
    FOREIGN KEY (source_balance_id) REFERENCES inventory_balance(balance_id);

ALTER TABLE putaway_task
    ADD CONSTRAINT fk_putaway_inventory_movement
    FOREIGN KEY (inventory_movement_id) REFERENCES inventory_movement(movement_id);

ALTER TABLE putaway_task
    ADD CONSTRAINT fk_putaway_resulting_balance
    FOREIGN KEY (resulting_balance_id) REFERENCES inventory_balance(balance_id);

ALTER TABLE putaway_task
    ADD CONSTRAINT fk_putaway_reversal_movement
    FOREIGN KEY (reversal_movement_id) REFERENCES inventory_movement(movement_id);

ALTER TABLE receipt_inventory
    ADD CONSTRAINT fk_receipt_inventory_initial_balance
    FOREIGN KEY (initial_balance_id) REFERENCES inventory_balance(balance_id);

ALTER TABLE quality_inspection
    ADD CONSTRAINT fk_quality_inspection_source_balance
    FOREIGN KEY (source_balance_id) REFERENCES inventory_balance(balance_id);

CREATE TABLE quarantine_case (
    quarantine_case_id varchar(140) PRIMARY KEY,
    parent_quarantine_case_id varchar(140) REFERENCES quarantine_case(quarantine_case_id),
    document_type_id   uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id          uuid NOT NULL,
    receipt_inventory_id varchar(160) NOT NULL REFERENCES receipt_inventory(receipt_inventory_id),
    inspection_id      varchar(120) NOT NULL REFERENCES quality_inspection(inspection_id),
    quarantine_balance_id varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    owner_id           uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id       uuid NOT NULL REFERENCES warehouse(warehouse_id),
    quarantine_qty     numeric(20,6) NOT NULL,
    uom_id             uuid NOT NULL REFERENCES uom(uom_id),
    opened_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    closed_at          timestamptz,
    notes              text,
    created_by         uuid NOT NULL REFERENCES app_account(account_id),
    updated_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by         uuid REFERENCES app_account(account_id),
    version_no         integer NOT NULL DEFAULT 1,
    UNIQUE (inspection_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    CONSTRAINT ck_quarantine_case_qty CHECK (quarantine_qty > 0),
    CONSTRAINT ck_quarantine_case_period CHECK (closed_at IS NULL OR closed_at >= opened_at),
    CONSTRAINT ck_quarantine_case_version CHECK (version_no > 0)
);

CREATE TABLE quarantine_disposition (
    quarantine_disposition_id varchar(150) PRIMARY KEY,
    quarantine_case_id        varchar(140) NOT NULL REFERENCES quarantine_case(quarantine_case_id),
    document_type_id          uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id                 uuid NOT NULL,
    quarantine_disposition_type_id uuid NOT NULL REFERENCES quarantine_disposition_type(quarantine_disposition_type_id),
    disposition_qty           numeric(20,6) NOT NULL,
    uom_id                    uuid NOT NULL REFERENCES uom(uom_id),
    client_decision_reference varchar(120),
    decision_notes            text,
    decided_at                timestamptz NOT NULL,
    decided_by                uuid NOT NULL REFERENCES app_account(account_id),
    processed_at              timestamptz,
    inventory_movement_id     varchar(140) REFERENCES inventory_movement(movement_id),
    resulting_balance_id      varchar(160) REFERENCES inventory_balance(balance_id),
    target_location_id        uuid REFERENCES warehouse_location(location_id),
    created_at                timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by                uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (inventory_movement_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    CONSTRAINT ck_quarantine_disposition_qty CHECK (disposition_qty > 0),
    CONSTRAINT ck_quarantine_disposition_period CHECK (
        processed_at IS NULL OR processed_at >= decided_at
    )
);

CREATE TABLE rework_task (
    rework_task_id      varchar(140) PRIMARY KEY,
    quarantine_disposition_id varchar(150) NOT NULL UNIQUE REFERENCES quarantine_disposition(quarantine_disposition_id),
    task_type_id        uuid NOT NULL REFERENCES task_type(task_type_id),
    task_status_id      uuid NOT NULL REFERENCES task_status(task_status_id),
    task_priority_id    uuid NOT NULL REFERENCES task_priority(task_priority_id),
    source_balance_id   varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    planned_qty         numeric(20,6) NOT NULL,
    completed_qty       numeric(20,6) NOT NULL DEFAULT 0,
    uom_id              uuid NOT NULL REFERENCES uom(uom_id),
    assigned_to         uuid REFERENCES app_account(account_id),
    work_instructions   text,
    result_notes        text,
    started_at          timestamptz,
    completed_at        timestamptz,
    reinspection_id     varchar(120) UNIQUE REFERENCES quality_inspection(inspection_id),
    created_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by          uuid NOT NULL REFERENCES app_account(account_id),
	updated_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
	updated_by         uuid REFERENCES app_account(account_id),
	version_no         integer NOT NULL DEFAULT 1,
    CONSTRAINT ck_rework_qty CHECK (
        planned_qty > 0 AND completed_qty >= 0 AND completed_qty <= planned_qty
    ),
    CONSTRAINT ck_rework_period CHECK (
        completed_at IS NULL OR started_at IS NULL OR completed_at >= started_at
    ),
    CONSTRAINT ck_rework_version CHECK (version_no > 0)
);

CREATE TABLE inbound_exception (
    inbound_exception_id varchar(140) PRIMARY KEY,
    owner_id             uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id         uuid NOT NULL REFERENCES warehouse(warehouse_id),
    source_document_id   varchar(140) NOT NULL,
    source_line_id       varchar(160),
    exception_type_code  varchar(40) NOT NULL,
    expected_qty         numeric(20,6),
    actual_qty           numeric(20,6),
    variance_qty         numeric(20,6),
    notes                text,
    created_at           timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by           uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_inbound_exception_type CHECK (
        exception_type_code IN (
            'OVER_RECEIPT', 'UNDER_RECEIPT', 'REJECTED_AT_DOCK',
            'DAMAGED', 'WRONG_ITEM',
            'CANCELLATION', 'REVERSAL'
        )
    )
);

-- -----------------------------------------------------------------------------
-- Outbound, allocation, picking, packing, and shipping transactions
-- -----------------------------------------------------------------------------

CREATE TABLE carrier (
    carrier_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    business_partner_id uuid UNIQUE REFERENCES business_partner(partner_id),
    code       varchar(40) NOT NULL UNIQUE,
    name       varchar(150) NOT NULL,
    is_active  boolean NOT NULL DEFAULT true
);

CREATE TABLE carrier_service (
    carrier_service_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    carrier_id         uuid NOT NULL REFERENCES carrier(carrier_id),
    code               varchar(40) NOT NULL,
    name               varchar(100) NOT NULL,
    is_active          boolean NOT NULL DEFAULT true,
    UNIQUE (carrier_id, code)
);

CREATE TABLE carrier_driver (
    driver_id      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    carrier_id     uuid NOT NULL REFERENCES carrier(carrier_id),
    account_id     uuid UNIQUE REFERENCES app_account(account_id),
    code           varchar(40) NOT NULL,
    name           varchar(150) NOT NULL,
    phone_number   varchar(50),
    license_number varchar(80),
    is_active      boolean NOT NULL DEFAULT true,
    created_at     timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by     uuid REFERENCES app_account(account_id),
    UNIQUE (carrier_id, code)
);

CREATE TABLE validation_severity (
    validation_severity_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                   varchar(20) NOT NULL UNIQUE,
    name                   varchar(80) NOT NULL,
    blocks_processing      boolean NOT NULL DEFAULT true,
    is_active              boolean NOT NULL DEFAULT true
);

CREATE TABLE outbound_validation_rule (
    outbound_validation_rule_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                        varchar(50) NOT NULL UNIQUE,
    name                        varchar(120) NOT NULL,
    description                 text,
    handler_code                varchar(60) NOT NULL,
    validation_severity_id      uuid NOT NULL REFERENCES validation_severity(validation_severity_id),
    display_order               integer NOT NULL DEFAULT 0,
    is_active                   boolean NOT NULL DEFAULT true
);

CREATE TABLE outbound_check_exception_status (
    outbound_check_exception_status_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                               varchar(30) NOT NULL UNIQUE,
    name                               varchar(100) NOT NULL,
    is_final                           boolean NOT NULL DEFAULT false,
    is_active                          boolean NOT NULL DEFAULT true
);

CREATE TABLE outbound_check_resolution_type (
    outbound_check_resolution_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                              varchar(40) NOT NULL UNIQUE,
    name                              varchar(120) NOT NULL,
    description                       text,
    counts_as_stock_correction        boolean NOT NULL DEFAULT false,
    counts_as_replacement             boolean NOT NULL DEFAULT false,
    counts_as_short_acceptance        boolean NOT NULL DEFAULT false,
    requires_approval                 boolean NOT NULL DEFAULT false,
    is_active                         boolean NOT NULL DEFAULT true,
    CONSTRAINT ck_outbound_resolution_effect CHECK (
        (CASE WHEN counts_as_stock_correction THEN 1 ELSE 0 END) +
        (CASE WHEN counts_as_replacement THEN 1 ELSE 0 END) +
        (CASE WHEN counts_as_short_acceptance THEN 1 ELSE 0 END) = 1
    )
);

CREATE TABLE outbound_wave_type (
    outbound_wave_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                  varchar(40) NOT NULL UNIQUE,
    name                  varchar(100) NOT NULL,
    description           text,
    is_active             boolean NOT NULL DEFAULT true
);

CREATE TABLE outbound_check_result (
    outbound_check_result_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                     varchar(40) NOT NULL UNIQUE,
    name                     varchar(100) NOT NULL,
    description              text,
    is_pass                  boolean NOT NULL DEFAULT false,
    requires_note            boolean NOT NULL DEFAULT false,
    is_active                boolean NOT NULL DEFAULT true
);

CREATE TABLE delivery_event_type (
    delivery_event_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                   varchar(40) NOT NULL UNIQUE,
    name                   varchar(100) NOT NULL,
    description            text,
    marks_delivered        boolean NOT NULL DEFAULT false,
    marks_failed           boolean NOT NULL DEFAULT false,
    marks_returned         boolean NOT NULL DEFAULT false,
    is_active              boolean NOT NULL DEFAULT true,
    CONSTRAINT ck_delivery_event_outcome CHECK (
        (CASE WHEN marks_delivered THEN 1 ELSE 0 END) +
        (CASE WHEN marks_failed THEN 1 ELSE 0 END) +
        (CASE WHEN marks_returned THEN 1 ELSE 0 END) <= 1
    )
);

CREATE TABLE delivery_failure_reason (
    delivery_failure_reason_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                       varchar(40) NOT NULL UNIQUE,
    name                       varchar(100) NOT NULL,
    description                text,
    is_active                  boolean NOT NULL DEFAULT true
);

CREATE TABLE outbound_order (
    outbound_id        varchar(120) PRIMARY KEY,
    document_type_id   uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id          uuid NOT NULL,
    owner_id           uuid NOT NULL REFERENCES organization(organization_id),
    customer_id        uuid NOT NULL REFERENCES business_partner(partner_id),
    warehouse_id       uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date      date NOT NULL,
    requested_ship_at  timestamptz,
    external_reference varchar(120),
    customer_order_no  varchar(120),
    client_delivery_order_no varchar(120) NOT NULL,
    ship_to_partner_id uuid REFERENCES business_partner(partner_id),
    ship_to_name       varchar(150) NOT NULL,
    ship_to_address_1  varchar(255) NOT NULL,
    ship_to_address_2  varchar(255),
    ship_to_city       varchar(100),
    ship_to_province   varchar(100),
    ship_to_postal_code varchar(20),
    ship_to_country_code varchar(2),
    notes              text,
    validated_at       timestamptz,
    validated_by       uuid REFERENCES app_account(account_id),
    allocated_at       timestamptz,
    completed_at       timestamptz,
    created_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by         uuid NOT NULL REFERENCES app_account(account_id),
    updated_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by         uuid REFERENCES app_account(account_id),
    version_no         integer NOT NULL DEFAULT 1,
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, customer_id)
        REFERENCES business_partner(owner_id, partner_id),
    FOREIGN KEY (owner_id, ship_to_partner_id)
        REFERENCES business_partner(owner_id, partner_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    CONSTRAINT ck_outbound_version CHECK (version_no > 0),
    UNIQUE (owner_id, client_delivery_order_no)
);

CREATE TABLE outbound_order_line (
    outbound_line_id varchar(150) PRIMARY KEY,
    outbound_id      varchar(120) NOT NULL REFERENCES outbound_order(outbound_id) ON DELETE CASCADE,
    line_no          integer NOT NULL,
    item_id          uuid NOT NULL REFERENCES item(item_id),
    ordered_qty      numeric(20,6) NOT NULL,
    allocated_qty    numeric(20,6) NOT NULL DEFAULT 0,
    picked_qty       numeric(20,6) NOT NULL DEFAULT 0,
    checked_qty      numeric(20,6) NOT NULL DEFAULT 0,
    packed_qty       numeric(20,6) NOT NULL DEFAULT 0,
    shipped_qty      numeric(20,6) NOT NULL DEFAULT 0,
    delivered_qty    numeric(20,6) NOT NULL DEFAULT 0,
    rejected_qty     numeric(20,6) NOT NULL DEFAULT 0,
    short_accepted_qty numeric(20,6) NOT NULL DEFAULT 0,
    uom_id           uuid NOT NULL REFERENCES uom(uom_id),
    requested_lot_no varchar(100),
    customer_line_reference varchar(100),
    notes            text,
    created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by       uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (outbound_id, line_no),
    CONSTRAINT ck_outbound_line_no CHECK (line_no > 0),
    CONSTRAINT ck_outbound_ordered_qty CHECK (
        ordered_qty > 0 AND allocated_qty >= 0 AND picked_qty >= 0
        AND checked_qty >= 0 AND packed_qty >= 0
        AND shipped_qty >= 0 AND delivered_qty >= 0
        AND rejected_qty >= 0 AND rejected_qty <= picked_qty
        AND short_accepted_qty >= 0 AND short_accepted_qty <= rejected_qty
        AND delivered_qty <= shipped_qty
        AND shipped_qty <= packed_qty
        AND packed_qty <= checked_qty
        AND checked_qty <= picked_qty - rejected_qty
        AND checked_qty <= ordered_qty - short_accepted_qty
        AND picked_qty <= allocated_qty
        AND allocated_qty <= ordered_qty + rejected_qty
    )
);

CREATE TABLE outbound_validation_run (
    validation_run_id varchar(140) PRIMARY KEY,
    outbound_id       varchar(120) NOT NULL REFERENCES outbound_order(outbound_id),
    document_type_id  uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id         uuid NOT NULL,
    validated_at      timestamptz,
    validated_by      uuid REFERENCES app_account(account_id),
    notes             text,
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id)
);

CREATE TABLE outbound_validation_result_detail (
    validation_result_id varchar(170) PRIMARY KEY,
    validation_run_id    varchar(140) NOT NULL REFERENCES outbound_validation_run(validation_run_id) ON DELETE CASCADE,
    outbound_validation_rule_id uuid NOT NULL REFERENCES outbound_validation_rule(outbound_validation_rule_id),
    passed               boolean NOT NULL,
    result_message       text,
    created_at           timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (validation_run_id, outbound_validation_rule_id)
);

CREATE TABLE outbound_wave (
    wave_id               varchar(120) PRIMARY KEY,
    document_type_id      uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id             uuid NOT NULL,
    outbound_wave_type_id uuid NOT NULL REFERENCES outbound_wave_type(outbound_wave_type_id),
    owner_id              uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id          uuid NOT NULL REFERENCES warehouse(warehouse_id),
    picking_strategy_id   uuid REFERENCES picking_strategy(picking_strategy_id),
    business_date         date NOT NULL,
    planned_release_at    timestamptz,
    released_at           timestamptz,
    completed_at          timestamptz,
    notes                 text,
    version_no            bigint NOT NULL DEFAULT 1,
    created_at            timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by            uuid NOT NULL REFERENCES app_account(account_id),
    updated_at            timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by            uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    CONSTRAINT ck_outbound_wave_version CHECK (version_no > 0)
);

CREATE TABLE outbound_wave_order (
    wave_id     varchar(120) NOT NULL REFERENCES outbound_wave(wave_id) ON DELETE CASCADE,
    outbound_id varchar(120) NOT NULL REFERENCES outbound_order(outbound_id),
    added_at    timestamptz NOT NULL DEFAULT clock_timestamp(),
    added_by    uuid NOT NULL REFERENCES app_account(account_id),
    PRIMARY KEY (wave_id, outbound_id)
);

CREATE TABLE inventory_reservation (
    reservation_id    varchar(140) PRIMARY KEY,
    document_type_id  uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id         uuid NOT NULL,
    picking_strategy_id uuid REFERENCES picking_strategy(picking_strategy_id),
    outbound_line_id  varchar(150) NOT NULL REFERENCES outbound_order_line(outbound_line_id),
    balance_id        varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    reserved_qty      numeric(20,6) NOT NULL,
    picked_qty        numeric(20,6) NOT NULL DEFAULT 0,
    uom_id            uuid NOT NULL REFERENCES uom(uom_id),
    reserved_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    released_at       timestamptz,
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    CONSTRAINT ck_reservation_qty CHECK (
        reserved_qty > 0 AND picked_qty >= 0 AND picked_qty <= reserved_qty
    ),
    CONSTRAINT ck_reservation_period CHECK (released_at IS NULL OR released_at >= reserved_at)
);

CREATE TABLE pick_task (
    pick_task_id       varchar(140) PRIMARY KEY,
    task_type_id       uuid NOT NULL REFERENCES task_type(task_type_id),
    task_status_id     uuid NOT NULL REFERENCES task_status(task_status_id),
    task_priority_id   uuid NOT NULL REFERENCES task_priority(task_priority_id),
    wave_id            varchar(120) NOT NULL REFERENCES outbound_wave(wave_id),
    reservation_id     varchar(140) NOT NULL REFERENCES inventory_reservation(reservation_id),
    outbound_line_id   varchar(150) NOT NULL REFERENCES outbound_order_line(outbound_line_id),
    source_location_id uuid NOT NULL REFERENCES warehouse_location(location_id),
    target_location_id uuid REFERENCES warehouse_location(location_id),
    planned_qty        numeric(20,6) NOT NULL,
    picked_qty         numeric(20,6) NOT NULL DEFAULT 0,
    uom_id             uuid NOT NULL REFERENCES uom(uom_id),
    assigned_to        uuid REFERENCES app_account(account_id),
    started_at         timestamptz,
    completed_at       timestamptz,
    short_qty          numeric(20,6) NOT NULL DEFAULT 0,
    short_reason_code_id uuid REFERENCES reason_code(reason_code_id),
    result_notes       text,
    created_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by         uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_pick_qty CHECK (
        planned_qty > 0 AND picked_qty >= 0 AND short_qty >= 0
        AND picked_qty + short_qty <= planned_qty
    ),
    UNIQUE (reservation_id)
);

CREATE TABLE pick_execution (
    pick_execution_id     varchar(170) PRIMARY KEY,
    pick_task_id          varchar(140) NOT NULL REFERENCES pick_task(pick_task_id),
    source_balance_id     varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    staging_balance_id    varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    picked_qty            numeric(20,6) NOT NULL,
    uom_id                uuid NOT NULL REFERENCES uom(uom_id),
    movement_id           varchar(140) NOT NULL UNIQUE REFERENCES inventory_movement(movement_id),
    picked_at             timestamptz NOT NULL DEFAULT clock_timestamp(),
    picked_by             uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_pick_execution_qty CHECK (picked_qty > 0)
);

CREATE TABLE outbound_staging (
    staging_id       varchar(120) PRIMARY KEY,
    document_type_id uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id        uuid NOT NULL,
    outbound_id      varchar(120) NOT NULL REFERENCES outbound_order(outbound_id),
    wave_id          varchar(120) NOT NULL REFERENCES outbound_wave(wave_id),
    warehouse_id     uuid NOT NULL REFERENCES warehouse(warehouse_id),
    staging_location_id uuid NOT NULL REFERENCES warehouse_location(location_id),
    staged_at        timestamptz,
    staged_by        uuid REFERENCES app_account(account_id),
    notes            text,
    created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by       uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id)
);

CREATE TABLE outbound_staging_line (
    staging_line_id   varchar(160) PRIMARY KEY,
    staging_id        varchar(120) NOT NULL REFERENCES outbound_staging(staging_id) ON DELETE CASCADE,
    pick_execution_id varchar(170) NOT NULL UNIQUE REFERENCES pick_execution(pick_execution_id),
    staging_balance_id varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    staged_qty        numeric(20,6) NOT NULL,
    removed_qty       numeric(20,6) NOT NULL DEFAULT 0,
    uom_id            uuid NOT NULL REFERENCES uom(uom_id),
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_staging_line_qty CHECK (
        staged_qty > 0 AND removed_qty >= 0 AND removed_qty <= staged_qty
    )
);

CREATE TABLE outbound_check (
    outbound_check_id varchar(120) PRIMARY KEY,
    parent_check_id   varchar(120) REFERENCES outbound_check(outbound_check_id),
    document_type_id  uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id         uuid NOT NULL,
    staging_id        varchar(120) NOT NULL REFERENCES outbound_staging(staging_id),
    checked_at        timestamptz,
    checked_by        uuid REFERENCES app_account(account_id),
    notes             text,
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id)
);

CREATE TABLE outbound_check_line (
    outbound_check_line_id varchar(160) PRIMARY KEY,
    outbound_check_id      varchar(120) NOT NULL REFERENCES outbound_check(outbound_check_id) ON DELETE CASCADE,
    staging_line_id        varchar(160) NOT NULL REFERENCES outbound_staging_line(staging_line_id),
    line_no                integer NOT NULL,
    expected_qty           numeric(20,6) NOT NULL,
    checked_qty            numeric(20,6),
    exception_qty          numeric(20,6) NOT NULL DEFAULT 0,
    uom_id                 uuid NOT NULL REFERENCES uom(uom_id),
    outbound_check_result_id uuid REFERENCES outbound_check_result(outbound_check_result_id),
    notes                  text,
    checked_at             timestamptz,
    checked_by             uuid REFERENCES app_account(account_id),
    created_at             timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by             uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (outbound_check_id, line_no),
    CONSTRAINT ck_outbound_check_line_no CHECK (line_no > 0),
    CONSTRAINT ck_outbound_check_qty CHECK (
        expected_qty > 0 AND (checked_qty IS NULL OR checked_qty >= 0)
        AND exception_qty >= 0
    )
);

CREATE TABLE outbound_check_exception (
    outbound_check_exception_id varchar(170) PRIMARY KEY,
    outbound_check_line_id      varchar(160) NOT NULL UNIQUE REFERENCES outbound_check_line(outbound_check_line_id),
    status_id                   uuid NOT NULL REFERENCES outbound_check_exception_status(outbound_check_exception_status_id),
    exception_qty               numeric(20,6) NOT NULL,
    opened_at                   timestamptz NOT NULL DEFAULT clock_timestamp(),
    resolved_at                 timestamptz,
    resolved_by                 uuid REFERENCES app_account(account_id),
    notes                       text,
    created_by                  uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_outbound_exception_qty CHECK (exception_qty > 0),
    CONSTRAINT ck_outbound_exception_period CHECK (
        resolved_at IS NULL OR resolved_at >= opened_at
    )
);

CREATE TABLE outbound_check_resolution (
    outbound_check_resolution_id varchar(190) PRIMARY KEY,
    outbound_check_exception_id  varchar(170) NOT NULL REFERENCES outbound_check_exception(outbound_check_exception_id),
    outbound_check_resolution_type_id uuid NOT NULL REFERENCES outbound_check_resolution_type(outbound_check_resolution_type_id),
    resolved_qty                 numeric(20,6) NOT NULL,
    notes                        text,
    approved_at                  timestamptz,
    approved_by                  uuid REFERENCES app_account(account_id),
    created_at                   timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by                   uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_outbound_resolution_qty CHECK (resolved_qty > 0),
    CONSTRAINT ck_outbound_resolution_approval CHECK (
        (approved_at IS NULL) = (approved_by IS NULL)
    )
);

CREATE TABLE outbound_check_resolution_movement (
    outbound_check_resolution_id varchar(190) NOT NULL REFERENCES outbound_check_resolution(outbound_check_resolution_id) ON DELETE CASCADE,
    movement_id                  varchar(140) NOT NULL UNIQUE REFERENCES inventory_movement(movement_id),
    PRIMARY KEY (outbound_check_resolution_id, movement_id)
);

CREATE TABLE outbound_check_resolution_reservation (
    outbound_check_resolution_id varchar(190) NOT NULL REFERENCES outbound_check_resolution(outbound_check_resolution_id) ON DELETE CASCADE,
    reservation_id               varchar(140) NOT NULL UNIQUE REFERENCES inventory_reservation(reservation_id),
    PRIMARY KEY (outbound_check_resolution_id, reservation_id)
);

CREATE TABLE packing (
    packing_id       varchar(120) PRIMARY KEY,
    document_type_id uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id        uuid NOT NULL,
    outbound_id      varchar(120) NOT NULL REFERENCES outbound_order(outbound_id),
    warehouse_id     uuid NOT NULL REFERENCES warehouse(warehouse_id),
    packing_location_id uuid REFERENCES warehouse_location(location_id),
    packed_at        timestamptz,
    packed_by        uuid REFERENCES app_account(account_id),
    created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by       uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    UNIQUE (outbound_id)
);

CREATE TABLE packing_line (
    packing_line_id varchar(150) PRIMARY KEY,
    packing_id      varchar(120) NOT NULL REFERENCES packing(packing_id) ON DELETE CASCADE,
    pick_task_id    varchar(140) NOT NULL REFERENCES pick_task(pick_task_id),
    outbound_check_line_id varchar(160) NOT NULL REFERENCES outbound_check_line(outbound_check_line_id),
    source_balance_id varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    packing_balance_id varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    handling_unit_id varchar(120) REFERENCES handling_unit(handling_unit_id),
    packed_qty      numeric(20,6) NOT NULL,
    uom_id          uuid NOT NULL REFERENCES uom(uom_id),
    movement_id     varchar(140) NOT NULL UNIQUE REFERENCES inventory_movement(movement_id),
    created_at      timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by      uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (outbound_check_line_id),
    CONSTRAINT ck_packing_qty CHECK (packed_qty > 0)
);

CREATE TABLE shipment (
    shipment_id       varchar(120) PRIMARY KEY,
    document_type_id  uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id         uuid NOT NULL,
    owner_id           uuid NOT NULL REFERENCES organization(organization_id),
    carrier_service_id uuid REFERENCES carrier_service(carrier_service_id),
    warehouse_id      uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date     date NOT NULL,
    route_reference   varchar(100),
    tracking_number   varchar(150),
    vehicle_number    varchar(60),
    seal_number       varchar(60),
    shipped_at        timestamptz,
    dispatched_by     uuid REFERENCES app_account(account_id),
    notes             text,
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id)
);

CREATE TABLE shipment_order (
    shipment_id varchar(120) NOT NULL REFERENCES shipment(shipment_id) ON DELETE CASCADE,
    outbound_id varchar(120) NOT NULL REFERENCES outbound_order(outbound_id),
    added_at    timestamptz NOT NULL DEFAULT clock_timestamp(),
    added_by    uuid NOT NULL REFERENCES app_account(account_id),
    removed_at  timestamptz,
    removed_by  uuid REFERENCES app_account(account_id),
    PRIMARY KEY (shipment_id, outbound_id),
    CONSTRAINT ck_shipment_order_period CHECK (
        removed_at IS NULL OR removed_at >= added_at
    )
);

CREATE UNIQUE INDEX uq_active_shipment_order
    ON shipment_order(outbound_id)
    WHERE removed_at IS NULL;

CREATE TABLE shipment_driver (
    shipment_id varchar(120) NOT NULL REFERENCES shipment(shipment_id) ON DELETE CASCADE,
    driver_id   uuid NOT NULL REFERENCES carrier_driver(driver_id),
    is_primary  boolean NOT NULL DEFAULT false,
    assigned_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    assigned_by uuid NOT NULL REFERENCES app_account(account_id),
    PRIMARY KEY (shipment_id, driver_id)
);

CREATE UNIQUE INDEX uq_shipment_primary_driver
    ON shipment_driver(shipment_id)
    WHERE is_primary;

CREATE TABLE shipment_line (
    shipment_line_id varchar(160) PRIMARY KEY,
    shipment_id      varchar(120) NOT NULL REFERENCES shipment(shipment_id) ON DELETE CASCADE,
    packing_line_id  varchar(150) NOT NULL UNIQUE REFERENCES packing_line(packing_line_id),
    source_balance_id varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    shipped_qty      numeric(20,6) NOT NULL,
    uom_id           uuid NOT NULL REFERENCES uom(uom_id),
    movement_id      varchar(140) NOT NULL UNIQUE REFERENCES inventory_movement(movement_id),
    created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by       uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_shipment_line_qty CHECK (shipped_qty > 0)
);

CREATE TABLE shipment_packing (
    shipment_id varchar(120) NOT NULL REFERENCES shipment(shipment_id) ON DELETE CASCADE,
    packing_id  varchar(120) NOT NULL REFERENCES packing(packing_id),
    PRIMARY KEY (shipment_id, packing_id)
);

CREATE TABLE delivery (
    delivery_id       varchar(120) PRIMARY KEY,
    document_type_id  uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id         uuid NOT NULL,
    shipment_id       varchar(120) NOT NULL REFERENCES shipment(shipment_id),
    outbound_id       varchar(120) NOT NULL REFERENCES outbound_order(outbound_id),
    business_date     date NOT NULL,
    planned_delivery_at timestamptz,
    arrived_at        timestamptz,
    delivered_at      timestamptz,
    recipient_name    varchar(150),
    recipient_reference varchar(100),
    proof_reference   varchar(200),
    proof_uri         text,
    latitude          numeric(10,7),
    longitude         numeric(10,7),
    notes             text,
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    updated_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by        uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    UNIQUE (shipment_id, outbound_id),
    CONSTRAINT ck_delivery_coordinates CHECK (
        (latitude IS NULL AND longitude IS NULL)
        OR (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
    )
);

CREATE TABLE delivery_line (
    delivery_line_id varchar(160) PRIMARY KEY,
    delivery_id      varchar(120) NOT NULL REFERENCES delivery(delivery_id) ON DELETE CASCADE,
    shipment_line_id varchar(160) NOT NULL UNIQUE REFERENCES shipment_line(shipment_line_id),
    planned_qty      numeric(20,6) NOT NULL,
    delivered_qty    numeric(20,6) NOT NULL DEFAULT 0,
    returned_qty     numeric(20,6) NOT NULL DEFAULT 0,
    uom_id           uuid NOT NULL REFERENCES uom(uom_id),
    created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by       uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_delivery_line_qty CHECK (
        planned_qty > 0 AND delivered_qty >= 0 AND returned_qty >= 0
        AND delivered_qty + returned_qty <= planned_qty
    )
);

CREATE TABLE delivery_event (
    delivery_event_id varchar(160) PRIMARY KEY,
    delivery_id       varchar(120) NOT NULL REFERENCES delivery(delivery_id) ON DELETE CASCADE,
    delivery_event_type_id uuid NOT NULL REFERENCES delivery_event_type(delivery_event_type_id),
    event_at          timestamptz NOT NULL,
    delivery_failure_reason_id uuid REFERENCES delivery_failure_reason(delivery_failure_reason_id),
    recipient_name    varchar(150),
    recipient_reference varchar(100),
    proof_reference   varchar(200),
    proof_uri         text,
    notes             text,
    latitude          numeric(10,7),
    longitude         numeric(10,7),
    recorded_by       uuid NOT NULL REFERENCES app_account(account_id),
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT ck_delivery_event_coordinates CHECK (
        (latitude IS NULL AND longitude IS NULL)
        OR (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
    )
);

CREATE TABLE delivery_event_line (
    delivery_event_id varchar(160) NOT NULL REFERENCES delivery_event(delivery_event_id) ON DELETE CASCADE,
    delivery_line_id  varchar(160) NOT NULL REFERENCES delivery_line(delivery_line_id),
    delivered_qty     numeric(20,6) NOT NULL,
    PRIMARY KEY (delivery_event_id, delivery_line_id),
    CONSTRAINT ck_delivery_event_line_qty CHECK (delivered_qty > 0)
);

CREATE TABLE outbound_return_policy (
    owner_id                   uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id               uuid NOT NULL REFERENCES warehouse(warehouse_id),
    return_location_id         uuid NOT NULL,
    return_inventory_status_id uuid NOT NULL REFERENCES inventory_status(inventory_status_id),
    is_active                  boolean NOT NULL DEFAULT true,
    created_at                 timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by                 uuid NOT NULL REFERENCES app_account(account_id),
    updated_at                 timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by                 uuid REFERENCES app_account(account_id),
    PRIMARY KEY (owner_id, warehouse_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    FOREIGN KEY (return_location_id, warehouse_id)
        REFERENCES warehouse_location(location_id, warehouse_id)
);

CREATE TABLE delivery_return_line (
    delivery_return_line_id varchar(170) PRIMARY KEY,
    delivery_id             varchar(120) NOT NULL REFERENCES delivery(delivery_id),
    delivery_line_id        varchar(160) NOT NULL REFERENCES delivery_line(delivery_line_id),
    return_location_id      uuid NOT NULL REFERENCES warehouse_location(location_id),
    returned_balance_id     varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    returned_qty            numeric(20,6) NOT NULL,
    uom_id                  uuid NOT NULL REFERENCES uom(uom_id),
    movement_id             varchar(140) NOT NULL UNIQUE REFERENCES inventory_movement(movement_id),
    returned_at             timestamptz NOT NULL,
    returned_by             uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_delivery_return_qty CHECK (returned_qty > 0)
);

-- -----------------------------------------------------------------------------
-- Internal movement, warehouse transfer, and inventory-control transactions
-- -----------------------------------------------------------------------------

CREATE TABLE internal_move_order (
    internal_move_id      varchar(120) PRIMARY KEY,
    document_type_id      uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id             uuid NOT NULL,
    internal_move_type_id uuid NOT NULL REFERENCES internal_move_type(internal_move_type_id),
    owner_id              uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id          uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date         date NOT NULL,
    reason_code_id        uuid REFERENCES reason_code(reason_code_id),
    requested_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    approved_at           timestamptz,
    approved_by           uuid REFERENCES app_account(account_id),
    completed_at          timestamptz,
    notes                 text,
    version_no            bigint NOT NULL DEFAULT 1,
    created_at            timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by            uuid NOT NULL REFERENCES app_account(account_id),
    updated_at            timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by            uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    CONSTRAINT ck_internal_move_version CHECK (version_no > 0)
);

CREATE TABLE internal_move_order_line (
    internal_move_line_id varchar(150) PRIMARY KEY,
    internal_move_id      varchar(120) NOT NULL REFERENCES internal_move_order(internal_move_id) ON DELETE CASCADE,
    line_no               integer NOT NULL,
    source_balance_id     varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    target_location_id    uuid NOT NULL REFERENCES warehouse_location(location_id),
    planned_qty           numeric(20,6) NOT NULL,
    completed_qty         numeric(20,6) NOT NULL DEFAULT 0,
    uom_id                uuid NOT NULL REFERENCES uom(uom_id),
    assigned_to           uuid REFERENCES app_account(account_id),
    created_at            timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by            uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (internal_move_id, line_no),
    CONSTRAINT ck_internal_move_line_no CHECK (line_no > 0),
    CONSTRAINT ck_internal_move_line_qty CHECK (
        planned_qty > 0 AND completed_qty >= 0 AND completed_qty <= planned_qty
    )
);

CREATE TABLE internal_move_execution (
    internal_move_execution_id varchar(160) PRIMARY KEY,
    internal_move_line_id      varchar(150) NOT NULL REFERENCES internal_move_order_line(internal_move_line_id),
    quantity                   numeric(20,6) NOT NULL,
    source_balance_id          varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    destination_balance_id     varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    movement_id                varchar(140) NOT NULL UNIQUE REFERENCES inventory_movement(movement_id),
    executed_at                timestamptz NOT NULL DEFAULT clock_timestamp(),
    executed_by                uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_internal_move_execution_qty CHECK (quantity > 0),
    CONSTRAINT ck_internal_move_execution_balances CHECK (source_balance_id <> destination_balance_id)
);

CREATE TABLE transfer_order (
    transfer_id          varchar(120) PRIMARY KEY,
    document_type_id     uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id            uuid NOT NULL,
    owner_id             uuid NOT NULL REFERENCES organization(organization_id),
    source_warehouse_id  uuid NOT NULL REFERENCES warehouse(warehouse_id),
    target_warehouse_id  uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date        date NOT NULL,
    requested_transfer_at timestamptz,
    approved_at           timestamptz,
    approved_by           uuid REFERENCES app_account(account_id),
    completed_at          timestamptz,
    notes                text,
    version_no            bigint NOT NULL DEFAULT 1,
    created_at           timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by           uuid NOT NULL REFERENCES app_account(account_id),
    updated_at           timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by           uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, source_warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    FOREIGN KEY (owner_id, target_warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    CONSTRAINT ck_transfer_warehouses CHECK (source_warehouse_id <> target_warehouse_id),
    CONSTRAINT ck_transfer_version CHECK (version_no > 0)
);

CREATE TABLE transfer_order_line (
    transfer_line_id varchar(150) PRIMARY KEY,
    transfer_id      varchar(120) NOT NULL REFERENCES transfer_order(transfer_id) ON DELETE CASCADE,
    line_no          integer NOT NULL,
    item_id          uuid NOT NULL REFERENCES item(item_id),
    requested_qty    numeric(20,6) NOT NULL,
    dispatched_qty   numeric(20,6) NOT NULL DEFAULT 0,
    received_qty     numeric(20,6) NOT NULL DEFAULT 0,
    uom_id           uuid NOT NULL REFERENCES uom(uom_id),
    requested_lot_id varchar(120) REFERENCES inventory_lot(lot_id),
    created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by       uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (transfer_id, line_no),
    CONSTRAINT ck_transfer_line_no CHECK (line_no > 0),
    CONSTRAINT ck_transfer_qty CHECK (
        requested_qty > 0 AND dispatched_qty >= 0 AND received_qty >= 0
        AND received_qty <= dispatched_qty AND dispatched_qty <= requested_qty
    )
);

CREATE TABLE transfer_dispatch (
    transfer_dispatch_id varchar(120) PRIMARY KEY,
    transfer_id          varchar(120) NOT NULL REFERENCES transfer_order(transfer_id),
    document_type_id     uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id            uuid NOT NULL,
    business_date        date NOT NULL,
    carrier_partner_id   uuid REFERENCES business_partner(partner_id),
    tracking_number      varchar(150),
    vehicle_number       varchar(60),
    seal_number          varchar(60),
    dispatched_at        timestamptz,
    notes                text,
    created_at           timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by           uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id)
);

CREATE TABLE transfer_dispatch_line (
    transfer_dispatch_line_id varchar(150) PRIMARY KEY,
    transfer_dispatch_id      varchar(120) NOT NULL REFERENCES transfer_dispatch(transfer_dispatch_id) ON DELETE CASCADE,
    transfer_line_id          varchar(150) NOT NULL REFERENCES transfer_order_line(transfer_line_id),
    line_no                   integer NOT NULL,
    source_balance_id         varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    dispatched_qty            numeric(20,6) NOT NULL,
    received_qty              numeric(20,6) NOT NULL DEFAULT 0,
    uom_id                    uuid NOT NULL REFERENCES uom(uom_id),
    movement_id               varchar(140) UNIQUE REFERENCES inventory_movement(movement_id),
    confirmed_at              timestamptz,
    confirmed_by              uuid REFERENCES app_account(account_id),
    created_at                timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by                uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (transfer_dispatch_id, line_no),
    CONSTRAINT ck_transfer_dispatch_line_no CHECK (line_no > 0),
    CONSTRAINT ck_transfer_dispatch_qty CHECK (
        dispatched_qty > 0 AND received_qty >= 0 AND received_qty <= dispatched_qty
    )
);

CREATE TABLE transfer_receipt (
    transfer_receipt_id varchar(120) PRIMARY KEY,
    transfer_id         varchar(120) NOT NULL REFERENCES transfer_order(transfer_id),
    document_type_id    uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id           uuid NOT NULL,
    warehouse_id        uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date       date NOT NULL,
    received_at         timestamptz,
    notes               text,
    created_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by          uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id)
);

CREATE TABLE transfer_receipt_line (
    transfer_receipt_line_id varchar(150) PRIMARY KEY,
    transfer_receipt_id      varchar(120) NOT NULL REFERENCES transfer_receipt(transfer_receipt_id) ON DELETE CASCADE,
    transfer_dispatch_line_id varchar(150) NOT NULL REFERENCES transfer_dispatch_line(transfer_dispatch_line_id),
    line_no                  integer NOT NULL,
    target_location_id       uuid NOT NULL REFERENCES warehouse_location(location_id),
    received_qty             numeric(20,6) NOT NULL,
    uom_id                   uuid NOT NULL REFERENCES uom(uom_id),
    destination_balance_id   varchar(160) REFERENCES inventory_balance(balance_id),
    movement_id              varchar(140) UNIQUE REFERENCES inventory_movement(movement_id),
    posted_at                timestamptz,
    posted_by                uuid REFERENCES app_account(account_id),
    created_at               timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by               uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (transfer_receipt_id, line_no),
    CONSTRAINT ck_transfer_receipt_line_no CHECK (line_no > 0),
    CONSTRAINT ck_transfer_receipt_qty CHECK (received_qty > 0)
);

CREATE TABLE inventory_status_change (
    inventory_status_change_id varchar(120) PRIMARY KEY,
    document_type_id           uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id                  uuid NOT NULL,
    owner_id                   uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id               uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date              date NOT NULL,
    reason_code_id             uuid NOT NULL REFERENCES reason_code(reason_code_id),
    approved_at                timestamptz,
    approved_by                uuid REFERENCES app_account(account_id),
    posted_at                  timestamptz,
    notes                      text,
    version_no                 bigint NOT NULL DEFAULT 1,
    created_at                 timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by                 uuid NOT NULL REFERENCES app_account(account_id),
    updated_at                 timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by                 uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    CONSTRAINT ck_status_change_version CHECK (version_no > 0)
);

CREATE TABLE inventory_status_change_line (
    inventory_status_change_line_id varchar(150) PRIMARY KEY,
    inventory_status_change_id      varchar(120) NOT NULL REFERENCES inventory_status_change(inventory_status_change_id) ON DELETE CASCADE,
    line_no                         integer NOT NULL,
    source_balance_id               varchar(160) NOT NULL REFERENCES inventory_balance(balance_id),
    target_inventory_status_id      uuid NOT NULL REFERENCES inventory_status(inventory_status_id),
    quantity                        numeric(20,6) NOT NULL,
    uom_id                          uuid NOT NULL REFERENCES uom(uom_id),
    destination_balance_id          varchar(160) REFERENCES inventory_balance(balance_id),
    movement_id                     varchar(140) UNIQUE REFERENCES inventory_movement(movement_id),
    posted_at                       timestamptz,
    posted_by                       uuid REFERENCES app_account(account_id),
    created_at                      timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by                      uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (inventory_status_change_id, line_no),
    CONSTRAINT ck_status_change_line_no CHECK (line_no > 0),
    CONSTRAINT ck_status_change_qty CHECK (quantity > 0)
);

CREATE TABLE inventory_adjustment (
    inventory_adjustment_id varchar(120) PRIMARY KEY,
    document_type_id        uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id               uuid NOT NULL,
    owner_id                uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id            uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date           date NOT NULL,
    reason_code_id          uuid NOT NULL REFERENCES reason_code(reason_code_id),
    approved_at             timestamptz,
    approved_by             uuid REFERENCES app_account(account_id),
    posted_at               timestamptz,
    notes                   text,
    version_no              bigint NOT NULL DEFAULT 1,
    created_at              timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by              uuid NOT NULL REFERENCES app_account(account_id),
    updated_at              timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by              uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    CONSTRAINT ck_inventory_adjustment_version CHECK (version_no > 0)
);

CREATE TABLE inventory_adjustment_line (
    inventory_adjustment_line_id varchar(150) PRIMARY KEY,
    inventory_adjustment_id      varchar(120) NOT NULL REFERENCES inventory_adjustment(inventory_adjustment_id) ON DELETE CASCADE,
    line_no                      integer NOT NULL,
    inventory_adjustment_type_id uuid NOT NULL REFERENCES inventory_adjustment_type(inventory_adjustment_type_id),
    source_balance_id            varchar(160) REFERENCES inventory_balance(balance_id),
    location_id                  uuid NOT NULL REFERENCES warehouse_location(location_id),
    item_id                      uuid NOT NULL REFERENCES item(item_id),
    lot_id                       varchar(120) REFERENCES inventory_lot(lot_id),
    handling_unit_id             varchar(120) REFERENCES handling_unit(handling_unit_id),
    inventory_status_id          uuid NOT NULL REFERENCES inventory_status(inventory_status_id),
    adjustment_qty               numeric(20,6) NOT NULL,
    uom_id                       uuid NOT NULL REFERENCES uom(uom_id),
    resulting_balance_id         varchar(160) REFERENCES inventory_balance(balance_id),
    movement_id                  varchar(140) UNIQUE REFERENCES inventory_movement(movement_id),
    posted_at                    timestamptz,
    posted_by                    uuid REFERENCES app_account(account_id),
    notes                        text,
    created_at                   timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by                   uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (inventory_adjustment_id, line_no),
    CONSTRAINT ck_inventory_adjustment_line_no CHECK (line_no > 0),
    CONSTRAINT ck_inventory_adjustment_qty CHECK (adjustment_qty > 0)
);

CREATE TABLE stock_count (
    stock_count_id    varchar(120) PRIMARY KEY,
    document_type_id  uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id         uuid NOT NULL,
    stock_count_type_id uuid NOT NULL REFERENCES stock_count_type(stock_count_type_id),
    owner_id          uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id      uuid NOT NULL REFERENCES warehouse(warehouse_id),
    business_date     date NOT NULL,
    reason_code_id    uuid REFERENCES reason_code(reason_code_id),
    freeze_inventory  boolean NOT NULL DEFAULT false,
    started_at        timestamptz,
    completed_at      timestamptz,
    reviewed_at       timestamptz,
    reviewed_by       uuid REFERENCES app_account(account_id),
    posted_at         timestamptz,
    posted_by         uuid REFERENCES app_account(account_id),
    notes             text,
    version_no        bigint NOT NULL DEFAULT 1,
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    updated_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by        uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    CONSTRAINT ck_stock_count_period CHECK (
        completed_at IS NULL OR started_at IS NULL OR completed_at >= started_at
    ),
    CONSTRAINT ck_stock_count_version CHECK (version_no > 0)
);

CREATE TABLE stock_count_line (
    stock_count_line_id varchar(150) PRIMARY KEY,
    stock_count_id      varchar(120) NOT NULL REFERENCES stock_count(stock_count_id) ON DELETE CASCADE,
    line_no             integer NOT NULL,
    balance_id          varchar(160) REFERENCES inventory_balance(balance_id),
    location_id         uuid NOT NULL REFERENCES warehouse_location(location_id),
    item_id             uuid NOT NULL REFERENCES item(item_id),
    lot_id              varchar(120) REFERENCES inventory_lot(lot_id),
    handling_unit_id    varchar(120) REFERENCES handling_unit(handling_unit_id),
    inventory_status_id uuid NOT NULL REFERENCES inventory_status(inventory_status_id),
    system_qty          numeric(20,6) NOT NULL,
    system_version_no   bigint,
    counted_qty         numeric(20,6),
    uom_id              uuid NOT NULL REFERENCES uom(uom_id),
    counted_at          timestamptz,
    counted_by          uuid REFERENCES app_account(account_id),
    variance_movement_id varchar(140) UNIQUE REFERENCES inventory_movement(movement_id),
    posted_at           timestamptz,
    posted_by           uuid REFERENCES app_account(account_id),
    created_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by          uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (stock_count_id, line_no),
    CONSTRAINT ck_stock_count_line_no CHECK (line_no > 0),
    CONSTRAINT ck_stock_count_quantities CHECK (
        system_qty >= 0 AND (counted_qty IS NULL OR counted_qty >= 0)
    ),
    CONSTRAINT ck_stock_count_system_version CHECK (
        system_version_no IS NULL OR system_version_no > 0
    )
);

-- -----------------------------------------------------------------------------
-- Configurable 3PL billing, invoicing, credits, and payment allocation
-- -----------------------------------------------------------------------------

CREATE TABLE currency (
    currency_id   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code          varchar(3) NOT NULL UNIQUE,
    name          varchar(100) NOT NULL,
    decimal_scale smallint NOT NULL DEFAULT 2,
    is_active     boolean NOT NULL DEFAULT true,
    CONSTRAINT ck_currency_scale CHECK (decimal_scale BETWEEN 0 AND 6)
);

CREATE TABLE billing_cycle (
    billing_cycle_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code             varchar(40) NOT NULL UNIQUE,
    name             varchar(100) NOT NULL,
    period_days      integer,
    period_months    integer,
    is_on_demand     boolean NOT NULL DEFAULT false,
    is_active        boolean NOT NULL DEFAULT true,
    CONSTRAINT ck_billing_cycle_period CHECK (
        (is_on_demand AND period_days IS NULL AND period_months IS NULL)
        OR (NOT is_on_demand AND
            (CASE WHEN period_days IS NOT NULL THEN 1 ELSE 0 END) +
            (CASE WHEN period_months IS NOT NULL THEN 1 ELSE 0 END) = 1)
    ),
    CONSTRAINT ck_billing_cycle_positive CHECK (
        (period_days IS NULL OR period_days > 0)
        AND (period_months IS NULL OR period_months > 0)
    )
);

CREATE TABLE payment_term (
    payment_term_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code            varchar(40) NOT NULL UNIQUE,
    name            varchar(100) NOT NULL,
    due_days        integer NOT NULL,
    is_active       boolean NOT NULL DEFAULT true,
    CONSTRAINT ck_payment_term_days CHECK (due_days >= 0)
);

CREATE TABLE charge_basis (
    charge_basis_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code            varchar(40) NOT NULL UNIQUE,
    name            varchar(100) NOT NULL,
    description     text,
    requires_uom    boolean NOT NULL DEFAULT true,
    use_event_quantity boolean NOT NULL DEFAULT true,
    is_active       boolean NOT NULL DEFAULT true
);

CREATE TABLE tax_rule (
    tax_rule_id     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        uuid REFERENCES organization(organization_id),
    code            varchar(40) NOT NULL,
    name            varchar(120) NOT NULL,
    rate_percent    numeric(9,6) NOT NULL,
    is_inclusive    boolean NOT NULL DEFAULT false,
    effective_from  date NOT NULL,
    effective_until date,
    is_active       boolean NOT NULL DEFAULT true,
    UNIQUE NULLS NOT DISTINCT (owner_id, code, effective_from),
    CONSTRAINT ck_tax_rate CHECK (rate_percent BETWEEN 0 AND 100),
    CONSTRAINT ck_tax_period CHECK (
        effective_until IS NULL OR effective_until >= effective_from
    )
);

CREATE TABLE billing_service (
    billing_service_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code               varchar(50) NOT NULL UNIQUE,
    name               varchar(120) NOT NULL,
    module_code        varchar(50) NOT NULL REFERENCES app_module(code),
    description        text,
    default_charge_basis_id uuid NOT NULL REFERENCES charge_basis(charge_basis_id),
    default_uom_id     uuid REFERENCES uom(uom_id),
    is_active          boolean NOT NULL DEFAULT true
);

-- Configures which immutable inventory movements create each billable service.
-- This keeps operational-to-commercial mapping out of application code.
CREATE TABLE billing_service_movement_type (
    billing_service_id uuid NOT NULL REFERENCES billing_service(billing_service_id),
    movement_type_id   uuid NOT NULL REFERENCES movement_type(movement_type_id),
    is_active          boolean NOT NULL DEFAULT true,
    PRIMARY KEY (billing_service_id, movement_type_id),
    UNIQUE (movement_type_id)
);

CREATE TABLE billing_adjustment_type (
    billing_adjustment_type_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                       varchar(40) NOT NULL UNIQUE,
    name                       varchar(100) NOT NULL,
    amount_effect              smallint NOT NULL,
    is_active                  boolean NOT NULL DEFAULT true,
    CONSTRAINT ck_billing_adjustment_effect CHECK (amount_effect IN (-1, 1))
);

CREATE TABLE payment_method (
    payment_method_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code              varchar(40) NOT NULL UNIQUE,
    name              varchar(100) NOT NULL,
    requires_reference boolean NOT NULL DEFAULT true,
    is_active         boolean NOT NULL DEFAULT true
);

CREATE TABLE billing_account (
    billing_account_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id           uuid NOT NULL REFERENCES organization(organization_id),
    code               varchar(40) NOT NULL,
    name               varchar(150) NOT NULL,
    bill_to_partner_id uuid REFERENCES business_partner(partner_id),
    currency_id        uuid NOT NULL REFERENCES currency(currency_id),
    billing_cycle_id   uuid NOT NULL REFERENCES billing_cycle(billing_cycle_id),
    payment_term_id    uuid NOT NULL REFERENCES payment_term(payment_term_id),
    default_tax_rule_id uuid REFERENCES tax_rule(tax_rule_id),
    bill_to_name       varchar(200) NOT NULL,
    bill_to_tax_number varchar(100),
    bill_to_address_1  varchar(255) NOT NULL,
    bill_to_address_2  varchar(255),
    bill_to_city       varchar(100),
    bill_to_province   varchar(100),
    bill_to_postal_code varchar(20),
    bill_to_country_code varchar(2),
    is_active          boolean NOT NULL DEFAULT true,
    created_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by         uuid REFERENCES app_account(account_id),
    updated_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by         uuid REFERENCES app_account(account_id),
    UNIQUE (owner_id, code),
    UNIQUE (owner_id, billing_account_id),
    FOREIGN KEY (owner_id, bill_to_partner_id)
        REFERENCES business_partner(owner_id, partner_id)
);

CREATE TABLE billing_contract (
    billing_contract_id varchar(120) PRIMARY KEY,
    document_type_id    uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id           uuid NOT NULL,
    billing_account_id  uuid NOT NULL REFERENCES billing_account(billing_account_id),
    owner_id            uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id        uuid REFERENCES warehouse(warehouse_id),
    currency_id         uuid NOT NULL REFERENCES currency(currency_id),
    default_tax_rule_id uuid REFERENCES tax_rule(tax_rule_id),
    contract_number     varchar(100) NOT NULL,
    effective_from      date NOT NULL,
    effective_until     date,
    auto_renew          boolean NOT NULL DEFAULT false,
    notes               text,
    version_no          bigint NOT NULL DEFAULT 1,
    created_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by          uuid NOT NULL REFERENCES app_account(account_id),
    updated_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by          uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    FOREIGN KEY (owner_id, billing_account_id)
        REFERENCES billing_account(owner_id, billing_account_id),
    FOREIGN KEY (owner_id, warehouse_id)
        REFERENCES warehouse_owner(owner_id, warehouse_id),
    UNIQUE (billing_account_id, contract_number),
    CONSTRAINT ck_billing_contract_period CHECK (
        effective_until IS NULL OR effective_until >= effective_from
    ),
    CONSTRAINT ck_billing_contract_version CHECK (version_no > 0)
);

CREATE TABLE rate_card (
    rate_card_id       varchar(120) PRIMARY KEY,
    document_type_id   uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id          uuid NOT NULL,
    billing_contract_id varchar(120) NOT NULL REFERENCES billing_contract(billing_contract_id),
    version_label      varchar(40) NOT NULL,
    effective_from     date NOT NULL,
    effective_until    date,
    notes              text,
    version_no         bigint NOT NULL DEFAULT 1,
    created_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by         uuid NOT NULL REFERENCES app_account(account_id),
    updated_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by         uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    UNIQUE (billing_contract_id, version_label),
    CONSTRAINT ck_rate_card_period CHECK (
        effective_until IS NULL OR effective_until >= effective_from
    ),
    CONSTRAINT ck_rate_card_version CHECK (version_no > 0)
);

CREATE TABLE rate_card_line (
    rate_card_line_id varchar(150) PRIMARY KEY,
    rate_card_id      varchar(120) NOT NULL REFERENCES rate_card(rate_card_id) ON DELETE CASCADE,
    line_no           integer NOT NULL,
    billing_service_id uuid NOT NULL REFERENCES billing_service(billing_service_id),
    charge_basis_id   uuid NOT NULL REFERENCES charge_basis(charge_basis_id),
    uom_id            uuid REFERENCES uom(uom_id),
    warehouse_id      uuid REFERENCES warehouse(warehouse_id),
    item_id           uuid REFERENCES item(item_id),
    category_id       uuid REFERENCES item_category(category_id),
    unit_rate         numeric(20,6) NOT NULL,
    included_quantity numeric(20,6) NOT NULL DEFAULT 0,
    minimum_charge    numeric(20,4) NOT NULL DEFAULT 0,
    maximum_charge    numeric(20,4),
    rounding_increment numeric(20,6),
    priority_no       integer NOT NULL DEFAULT 100,
    tax_rule_id       uuid REFERENCES tax_rule(tax_rule_id),
    is_active         boolean NOT NULL DEFAULT true,
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (rate_card_id, line_no),
    CONSTRAINT ck_rate_card_line_no CHECK (line_no > 0),
    CONSTRAINT ck_rate_card_amounts CHECK (
        unit_rate >= 0 AND included_quantity >= 0 AND minimum_charge >= 0
        AND (maximum_charge IS NULL OR maximum_charge >= minimum_charge)
        AND (rounding_increment IS NULL OR rounding_increment > 0)
    )
);

CREATE TABLE rate_card_tier (
    rate_card_tier_id varchar(170) PRIMARY KEY,
    rate_card_line_id varchar(150) NOT NULL REFERENCES rate_card_line(rate_card_line_id) ON DELETE CASCADE,
    tier_no           integer NOT NULL,
    from_quantity     numeric(20,6) NOT NULL,
    until_quantity   numeric(20,6),
    unit_rate         numeric(20,6) NOT NULL,
    UNIQUE (rate_card_line_id, tier_no),
    CONSTRAINT ck_rate_tier_no CHECK (tier_no > 0),
    CONSTRAINT ck_rate_tier_range CHECK (
        from_quantity >= 0
        AND (until_quantity IS NULL OR until_quantity > from_quantity)
        AND unit_rate >= 0
    )
);

CREATE TABLE billable_event (
    billable_event_id varchar(140) PRIMARY KEY,
    document_type_id  uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id         uuid NOT NULL,
    billing_account_id uuid NOT NULL REFERENCES billing_account(billing_account_id),
    billing_contract_id varchar(120) NOT NULL REFERENCES billing_contract(billing_contract_id),
    billing_service_id uuid NOT NULL REFERENCES billing_service(billing_service_id),
    owner_id          uuid NOT NULL REFERENCES organization(organization_id),
    warehouse_id      uuid REFERENCES warehouse(warehouse_id),
    item_id           uuid REFERENCES item(item_id),
    category_id       uuid REFERENCES item_category(category_id),
    handling_unit_id  varchar(120) REFERENCES handling_unit(handling_unit_id),
    business_date     date NOT NULL,
    occurred_at       timestamptz NOT NULL,
    source_document_type_id uuid REFERENCES document_type(document_type_id),
    source_document_id varchar(140) NOT NULL,
    source_line_id    varchar(170),
    event_key         varchar(80) NOT NULL DEFAULT 'DEFAULT',
    quantity          numeric(20,6) NOT NULL,
    uom_id            uuid REFERENCES uom(uom_id),
    attributes        jsonb NOT NULL DEFAULT '{}'::jsonb,
    exclusion_reason_code_id uuid REFERENCES reason_code(reason_code_id),
    rated_at          timestamptz,
    rated_by          uuid REFERENCES app_account(account_id),
    excluded_at       timestamptz,
    excluded_by       uuid REFERENCES app_account(account_id),
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    UNIQUE NULLS NOT DISTINCT (
        billing_account_id, billing_service_id,
        source_document_id, source_line_id, event_key
    ),
    CONSTRAINT ck_billable_event_qty CHECK (quantity > 0)
);

CREATE TABLE billing_run (
    billing_run_id     varchar(120) PRIMARY KEY,
    document_type_id   uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id          uuid NOT NULL,
    billing_account_id uuid NOT NULL REFERENCES billing_account(billing_account_id),
    billing_contract_id varchar(120) NOT NULL REFERENCES billing_contract(billing_contract_id),
    currency_id        uuid NOT NULL REFERENCES currency(currency_id),
    period_from        date NOT NULL,
    period_until       date NOT NULL,
    calculated_at     timestamptz,
    reviewed_at       timestamptz,
    reviewed_by       uuid REFERENCES app_account(account_id),
    net_amount        numeric(20,4) NOT NULL DEFAULT 0,
    tax_amount        numeric(20,4) NOT NULL DEFAULT 0,
    gross_amount      numeric(20,4) NOT NULL DEFAULT 0,
    notes             text,
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    CONSTRAINT ck_billing_run_period CHECK (period_until >= period_from),
    CONSTRAINT ck_billing_run_amounts CHECK (
        net_amount >= 0 AND tax_amount >= 0 AND gross_amount >= 0
        AND gross_amount = net_amount + tax_amount
    )
);

CREATE TABLE billing_charge (
    billing_charge_id varchar(150) PRIMARY KEY,
    billing_run_id    varchar(120) NOT NULL REFERENCES billing_run(billing_run_id) ON DELETE CASCADE,
    billable_event_id varchar(140) NOT NULL UNIQUE REFERENCES billable_event(billable_event_id),
    rate_card_line_id varchar(150) NOT NULL REFERENCES rate_card_line(rate_card_line_id),
    billed_quantity   numeric(20,6) NOT NULL,
    uom_id            uuid REFERENCES uom(uom_id),
    unit_rate         numeric(20,6) NOT NULL,
    net_amount        numeric(20,4) NOT NULL,
    tax_rule_id       uuid REFERENCES tax_rule(tax_rule_id),
    tax_rate_percent  numeric(9,6) NOT NULL DEFAULT 0,
    tax_amount        numeric(20,4) NOT NULL DEFAULT 0,
    gross_amount      numeric(20,4) NOT NULL,
    currency_id       uuid NOT NULL REFERENCES currency(currency_id),
    calculated_at     timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    CONSTRAINT ck_billing_charge_values CHECK (
        billed_quantity > 0 AND unit_rate >= 0 AND net_amount >= 0
        AND tax_rate_percent BETWEEN 0 AND 100 AND tax_amount >= 0
        AND gross_amount = net_amount + tax_amount
    )
);

CREATE TABLE billing_invoice (
    billing_invoice_id varchar(120) PRIMARY KEY,
    document_type_id   uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id          uuid NOT NULL,
    billing_run_id     varchar(120) NOT NULL UNIQUE REFERENCES billing_run(billing_run_id),
    billing_account_id uuid NOT NULL REFERENCES billing_account(billing_account_id),
    billing_contract_id varchar(120) NOT NULL REFERENCES billing_contract(billing_contract_id),
    currency_id        uuid NOT NULL REFERENCES currency(currency_id),
    invoice_date       date NOT NULL,
    due_date           date NOT NULL,
    bill_to_name       varchar(200) NOT NULL,
    bill_to_tax_number varchar(100),
    bill_to_address    text NOT NULL,
    net_amount         numeric(20,4) NOT NULL DEFAULT 0,
    adjustment_amount  numeric(20,4) NOT NULL DEFAULT 0,
    tax_amount         numeric(20,4) NOT NULL DEFAULT 0,
    gross_amount       numeric(20,4) NOT NULL DEFAULT 0,
    credited_amount    numeric(20,4) NOT NULL DEFAULT 0,
    paid_amount        numeric(20,4) NOT NULL DEFAULT 0,
    balance_due        numeric(20,4) NOT NULL DEFAULT 0,
    issued_at          timestamptz,
    issued_by          uuid REFERENCES app_account(account_id),
    notes              text,
    created_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by         uuid NOT NULL REFERENCES app_account(account_id),
    updated_at         timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_by         uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    CONSTRAINT ck_invoice_dates CHECK (due_date >= invoice_date),
    CONSTRAINT ck_invoice_amounts CHECK (
        net_amount >= 0 AND tax_amount >= 0 AND gross_amount >= 0
        AND credited_amount >= 0 AND paid_amount >= 0 AND balance_due >= 0
        AND gross_amount = net_amount + adjustment_amount + tax_amount
        AND balance_due = gross_amount - credited_amount - paid_amount
    )
);

CREATE TABLE billing_invoice_line (
    billing_invoice_line_id varchar(150) PRIMARY KEY,
    billing_invoice_id      varchar(120) NOT NULL REFERENCES billing_invoice(billing_invoice_id) ON DELETE CASCADE,
    line_no                 integer NOT NULL,
    billing_charge_id       varchar(150) NOT NULL UNIQUE REFERENCES billing_charge(billing_charge_id),
    billing_service_id      uuid NOT NULL REFERENCES billing_service(billing_service_id),
    description             varchar(255) NOT NULL,
    quantity                numeric(20,6) NOT NULL,
    uom_id                  uuid REFERENCES uom(uom_id),
    unit_rate               numeric(20,6) NOT NULL,
    net_amount              numeric(20,4) NOT NULL,
    tax_rule_id             uuid REFERENCES tax_rule(tax_rule_id),
    tax_rate_percent        numeric(9,6) NOT NULL DEFAULT 0,
    tax_amount              numeric(20,4) NOT NULL DEFAULT 0,
    gross_amount            numeric(20,4) NOT NULL,
    UNIQUE (billing_invoice_id, line_no),
    CONSTRAINT ck_invoice_line_no CHECK (line_no > 0),
    CONSTRAINT ck_invoice_line_amounts CHECK (
        quantity > 0 AND unit_rate >= 0 AND net_amount >= 0
        AND tax_rate_percent BETWEEN 0 AND 100 AND tax_amount >= 0
        AND gross_amount = net_amount + tax_amount
    )
);

CREATE TABLE billing_invoice_adjustment (
    invoice_adjustment_id varchar(150) PRIMARY KEY,
    billing_invoice_id    varchar(120) NOT NULL REFERENCES billing_invoice(billing_invoice_id) ON DELETE CASCADE,
    line_no               integer NOT NULL,
    billing_adjustment_type_id uuid NOT NULL REFERENCES billing_adjustment_type(billing_adjustment_type_id),
    reason_code_id        uuid NOT NULL REFERENCES reason_code(reason_code_id),
    description           varchar(255) NOT NULL,
    amount                numeric(20,4) NOT NULL,
    tax_rule_id           uuid REFERENCES tax_rule(tax_rule_id),
    tax_rate_percent      numeric(9,6) NOT NULL DEFAULT 0,
    tax_amount            numeric(20,4) NOT NULL DEFAULT 0,
    created_at            timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by            uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (billing_invoice_id, line_no),
    CONSTRAINT ck_invoice_adjustment_line_no CHECK (line_no > 0),
    CONSTRAINT ck_invoice_adjustment_amount CHECK (
        amount > 0 AND tax_rate_percent BETWEEN 0 AND 100 AND tax_amount >= 0
    )
);

CREATE TABLE billing_credit_note (
    credit_note_id      varchar(120) PRIMARY KEY,
    document_type_id    uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id           uuid NOT NULL,
    billing_invoice_id  varchar(120) NOT NULL REFERENCES billing_invoice(billing_invoice_id),
    currency_id         uuid NOT NULL REFERENCES currency(currency_id),
    credit_date         date NOT NULL,
    reason_code_id      uuid NOT NULL REFERENCES reason_code(reason_code_id),
    net_amount          numeric(20,4) NOT NULL DEFAULT 0,
    tax_amount          numeric(20,4) NOT NULL DEFAULT 0,
    gross_amount        numeric(20,4) NOT NULL DEFAULT 0,
    issued_at           timestamptz,
    issued_by           uuid REFERENCES app_account(account_id),
    notes               text,
    created_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by          uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    CONSTRAINT ck_credit_note_amounts CHECK (
        net_amount >= 0 AND tax_amount >= 0 AND gross_amount = net_amount + tax_amount
    )
);

CREATE TABLE billing_credit_note_line (
    credit_note_line_id varchar(150) PRIMARY KEY,
    credit_note_id      varchar(120) NOT NULL REFERENCES billing_credit_note(credit_note_id) ON DELETE CASCADE,
    billing_invoice_line_id varchar(150) REFERENCES billing_invoice_line(billing_invoice_line_id),
    line_no             integer NOT NULL,
    description         varchar(255) NOT NULL,
    net_amount          numeric(20,4) NOT NULL,
    tax_rate_percent    numeric(9,6) NOT NULL DEFAULT 0,
    tax_amount          numeric(20,4) NOT NULL DEFAULT 0,
    gross_amount        numeric(20,4) NOT NULL,
    UNIQUE (credit_note_id, line_no),
    CONSTRAINT ck_credit_line_no CHECK (line_no > 0),
    CONSTRAINT ck_credit_line_amounts CHECK (
        net_amount > 0 AND tax_rate_percent BETWEEN 0 AND 100
        AND tax_amount >= 0 AND gross_amount = net_amount + tax_amount
    )
);

CREATE TABLE billing_payment (
    payment_id        varchar(120) PRIMARY KEY,
    document_type_id  uuid NOT NULL REFERENCES document_type(document_type_id),
    status_id         uuid NOT NULL,
    billing_account_id uuid NOT NULL REFERENCES billing_account(billing_account_id),
    currency_id       uuid NOT NULL REFERENCES currency(currency_id),
    payment_method_id uuid NOT NULL REFERENCES payment_method(payment_method_id),
    payment_date      date NOT NULL,
    amount            numeric(20,4) NOT NULL,
    unapplied_amount  numeric(20,4) NOT NULL,
    external_reference varchar(150),
    notes             text,
    created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
    created_by        uuid NOT NULL REFERENCES app_account(account_id),
    FOREIGN KEY (document_type_id, status_id)
        REFERENCES document_status(document_type_id, status_id),
    CONSTRAINT ck_payment_amounts CHECK (
        amount > 0 AND unapplied_amount >= 0 AND unapplied_amount <= amount
    )
);

CREATE TABLE billing_payment_allocation (
    payment_allocation_id varchar(150) PRIMARY KEY,
    payment_id            varchar(120) NOT NULL REFERENCES billing_payment(payment_id) ON DELETE CASCADE,
    allocation_no        integer NOT NULL,
    billing_invoice_id    varchar(120) NOT NULL REFERENCES billing_invoice(billing_invoice_id),
    allocated_amount      numeric(20,4) NOT NULL,
    allocated_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    allocated_by          uuid NOT NULL REFERENCES app_account(account_id),
    UNIQUE (payment_id, billing_invoice_id),
    UNIQUE (payment_id, allocation_no),
    CONSTRAINT ck_payment_allocation_no CHECK (allocation_no > 0),
    CONSTRAINT ck_payment_allocation_amount CHECK (allocated_amount > 0)
);

-- -----------------------------------------------------------------------------
-- Supporting indexes for common menu and operational queries
-- -----------------------------------------------------------------------------

CREATE INDEX ix_account_role_role ON account_role(role_id, account_id);
CREATE UNIQUE INDEX uq_account_username_ci ON app_account(lower(username));
CREATE UNIQUE INDEX uq_account_email_ci
    ON app_account(lower(email))
    WHERE email IS NOT NULL;
CREATE INDEX ix_session_account_active
    ON app_session(account_id, expires_at)
    WHERE revoked_at IS NULL;
CREATE INDEX ix_session_expiry
    ON app_session(expires_at)
    WHERE revoked_at IS NULL;
CREATE INDEX ix_role_permission_permission ON role_permission(permission_id, role_id);
CREATE INDEX ix_menu_parent_order ON app_menu(parent_menu_id, display_order);
CREATE INDEX ix_partner_owner_name ON business_partner(owner_id, name);
CREATE INDEX ix_location_warehouse_zone ON warehouse_location(warehouse_id, zone_id);
CREATE INDEX ix_item_owner_name ON item(owner_id, name);
CREATE INDEX ix_item_category_parent
    ON item_category(owner_id, parent_category_id);
CREATE INDEX ix_item_barcode_item
    ON item_barcode(item_id, is_active);

CREATE INDEX ix_billing_service_movement_active
    ON billing_service_movement_type(movement_type_id)
    WHERE is_active;

CREATE INDEX ix_inbound_owner_warehouse_date
    ON inbound_order(owner_id, warehouse_id, business_date DESC);
CREATE INDEX ix_purchase_order_owner_vendor_date
    ON purchase_order(owner_id, vendor_id, business_date DESC);
CREATE INDEX ix_purchase_order_status_date
    ON purchase_order(status_id, business_date DESC);
CREATE INDEX ix_purchase_order_line_item
    ON purchase_order_line(item_id, purchase_order_id);
CREATE INDEX ix_inbound_status_date
    ON inbound_order(status_id, business_date DESC);
CREATE INDEX ix_receipt_inbound ON receipt(inbound_id);
CREATE INDEX ix_receipt_owner_warehouse_date
    ON receipt(owner_id, warehouse_id, business_date DESC);
CREATE INDEX ix_receipt_line_item ON receipt_line(item_id, receipt_id);
CREATE INDEX ix_receipt_inventory_line
    ON receipt_inventory(receipt_line_id, item_id);
CREATE INDEX ix_quality_inspection_receipt_inventory
    ON quality_inspection(receipt_inventory_id, created_at DESC);
CREATE UNIQUE INDEX uq_quality_inspection_root_receipt_inventory
    ON quality_inspection(receipt_inventory_id) WHERE parent_inspection_id IS NULL;
CREATE UNIQUE INDEX uq_putaway_task_inventory_movement
    ON putaway_task(inventory_movement_id) WHERE inventory_movement_id IS NOT NULL;
CREATE INDEX ix_putaway_assignee_status ON putaway_task(assigned_to, task_status_id);

CREATE INDEX ix_balance_lookup
    ON inventory_balance(owner_id, warehouse_id, item_id, location_id, inventory_status_id);
CREATE INDEX ix_balance_lot ON inventory_balance(lot_id) WHERE lot_id IS NOT NULL;
CREATE INDEX ix_movement_item_time ON inventory_movement(owner_id, item_id, occurred_at DESC);
CREATE INDEX ix_movement_warehouse_date_type
    ON inventory_movement(owner_id, warehouse_id, business_date DESC, movement_type_id);
CREATE INDEX ix_movement_source ON inventory_movement(source_document_id, source_line_id);
CREATE UNIQUE INDEX uq_inventory_movement_operation_key
    ON inventory_movement(operation_key) WHERE operation_key IS NOT NULL;
CREATE INDEX ix_serial_inventory_balance ON serial_inventory(balance_id);
CREATE INDEX ix_quarantine_owner_status
    ON quarantine_case(owner_id, warehouse_id, status_id, opened_at DESC);
CREATE INDEX ix_quarantine_disposition_case
    ON quarantine_disposition(quarantine_case_id, decided_at);
CREATE INDEX ix_rework_assignee_status
    ON rework_task(assigned_to, task_status_id);

CREATE INDEX ix_outbound_owner_warehouse_date
    ON outbound_order(owner_id, warehouse_id, business_date DESC);
CREATE INDEX ix_outbound_status_date
    ON outbound_order(status_id, business_date DESC);
CREATE INDEX ix_outbound_validation_run_order
    ON outbound_validation_run(outbound_id, created_at DESC);
CREATE INDEX ix_reservation_outbound_line ON inventory_reservation(outbound_line_id);
CREATE INDEX ix_reservation_balance_active
    ON inventory_reservation(balance_id, released_at) WHERE released_at IS NULL;
CREATE INDEX ix_wave_owner_status_date
    ON outbound_wave(owner_id, warehouse_id, status_id, business_date DESC);
CREATE INDEX ix_wave_order_outbound ON outbound_wave_order(outbound_id, wave_id);
CREATE INDEX ix_pick_assignee_status ON pick_task(assigned_to, task_status_id);
CREATE INDEX ix_pick_wave_status ON pick_task(wave_id, task_status_id);
CREATE INDEX ix_pick_execution_task ON pick_execution(pick_task_id, picked_at);
CREATE INDEX ix_staging_wave ON outbound_staging(wave_id, status_id);
CREATE INDEX ix_staging_line_staging ON outbound_staging_line(staging_id);
CREATE INDEX ix_check_staging ON outbound_check(staging_id, status_id);
CREATE INDEX ix_check_exception_status
    ON outbound_check_exception(status_id, opened_at DESC);
CREATE INDEX ix_check_resolution_exception
    ON outbound_check_resolution(outbound_check_exception_id, created_at);
CREATE INDEX ix_packing_outbound ON packing(outbound_id, status_id);
CREATE INDEX ix_packing_line_packing ON packing_line(packing_id);
CREATE INDEX ix_shipment_owner_warehouse_date
    ON shipment(owner_id, warehouse_id, business_date DESC);
CREATE INDEX ix_shipment_order_outbound ON shipment_order(outbound_id, shipment_id);
CREATE INDEX ix_shipment_line_shipment ON shipment_line(shipment_id);
CREATE INDEX ix_delivery_outbound ON delivery(outbound_id, status_id, business_date DESC);
CREATE INDEX ix_delivery_shipment ON delivery(shipment_id, status_id);
CREATE INDEX ix_delivery_line_delivery ON delivery_line(delivery_id);
CREATE INDEX ix_delivery_event_delivery ON delivery_event(delivery_id, event_at DESC);
CREATE INDEX ix_delivery_return_delivery ON delivery_return_line(delivery_id);

CREATE INDEX ix_transfer_owner_date ON transfer_order(owner_id, business_date DESC);
CREATE INDEX ix_internal_move_owner_status_date
    ON internal_move_order(owner_id, warehouse_id, status_id, business_date DESC);
CREATE INDEX ix_internal_move_line_source
    ON internal_move_order_line(source_balance_id, internal_move_id);
CREATE INDEX ix_internal_move_execution_line
    ON internal_move_execution(internal_move_line_id, executed_at);
CREATE INDEX ix_transfer_dispatch_transfer
    ON transfer_dispatch(transfer_id, business_date DESC);
CREATE INDEX ix_transfer_dispatch_line_transfer_line
    ON transfer_dispatch_line(transfer_line_id, transfer_dispatch_id);
CREATE INDEX ix_transfer_receipt_transfer
    ON transfer_receipt(transfer_id, business_date DESC);
CREATE INDEX ix_transfer_receipt_line_dispatch
    ON transfer_receipt_line(transfer_dispatch_line_id, posted_at);
CREATE INDEX ix_status_change_owner_status_date
    ON inventory_status_change(owner_id, warehouse_id, status_id, business_date DESC);
CREATE INDEX ix_adjustment_owner_status_date
    ON inventory_adjustment(owner_id, warehouse_id, status_id, business_date DESC);
CREATE INDEX ix_stock_count_owner_warehouse_date
    ON stock_count(owner_id, warehouse_id, business_date DESC);
CREATE UNIQUE INDEX uq_stock_count_balance
    ON stock_count_line(stock_count_id, balance_id)
    WHERE balance_id IS NOT NULL;

CREATE INDEX ix_billing_contract_account_period
    ON billing_contract(billing_account_id, effective_from, effective_until);
CREATE INDEX ix_rate_card_contract_period
    ON rate_card(billing_contract_id, effective_from, effective_until);
CREATE INDEX ix_rate_card_line_service_scope
    ON rate_card_line(rate_card_id, billing_service_id, warehouse_id, item_id, category_id, priority_no);
CREATE INDEX ix_billable_event_account_date_status
    ON billable_event(billing_account_id, business_date, status_id);
CREATE INDEX ix_billable_event_source
    ON billable_event(source_document_id, source_line_id, event_key);
CREATE INDEX ix_billing_run_account_period
    ON billing_run(billing_account_id, period_from, period_until, status_id);
CREATE INDEX ix_billing_charge_run ON billing_charge(billing_run_id);
CREATE INDEX ix_invoice_account_date_status
    ON billing_invoice(billing_account_id, invoice_date DESC, status_id);
CREATE INDEX ix_credit_note_invoice ON billing_credit_note(billing_invoice_id, credit_date DESC);
CREATE INDEX ix_payment_account_date ON billing_payment(billing_account_id, payment_date DESC);
CREATE INDEX ix_payment_allocation_invoice ON billing_payment_allocation(billing_invoice_id);

COMMIT;
