-- Authentication and authorization query examples for PostgreSQL
-- Schema prerequisite: database/wms_schema.sql
-- This is a query catalog: execute one numbered query at a time, not this whole
-- file as a single script.
--
-- Parameter style uses PostgreSQL positional parameters: $1, $2, ...
-- The application must:
--   1. Hash passwords with Argon2id (preferred) or bcrypt before sending them.
--   2. Generate a cryptographically random raw session token.
--   3. Send only SHA-256(raw_session_token) as token_hash to PostgreSQL.
--   4. Return the raw token to the client in a Secure, HttpOnly cookie.

SET search_path TO wms, public;

-- =============================================================================
-- 1. AUTHENTICATION POLICY MASTER
-- =============================================================================

-- 1.1 Create an authentication policy.
-- $1 code, $2 name, $3 max attempts, $4 lockout seconds,
-- $5 session TTL seconds, $6 is default
INSERT INTO authentication_policy (
    code,
    name,
    max_failed_attempts,
    lockout_seconds,
    session_ttl_seconds,
    is_default
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- 1.2 List authentication policies.
SELECT
    authentication_policy_id,
    code,
    name,
    max_failed_attempts,
    lockout_seconds,
    session_ttl_seconds,
    is_default,
    is_active
FROM authentication_policy
ORDER BY is_default DESC, name;

-- 1.3 Update a policy.
-- $1 policy ID, $2 name, $3 max attempts, $4 lockout seconds,
-- $5 session TTL seconds, $6 active
UPDATE authentication_policy
SET name                = $2,
    max_failed_attempts = $3,
    lockout_seconds     = $4,
    session_ttl_seconds = $5,
    is_active           = $6
WHERE authentication_policy_id = $1
RETURNING *;

-- 1.4 Create a session revocation reason.
-- $1 code, $2 name, $3 description
INSERT INTO session_revocation_reason (code, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- 1.5 List session revocation reasons.
SELECT
    session_revocation_reason_id,
    code,
    name,
    description,
    is_active
FROM session_revocation_reason
ORDER BY name;

-- =============================================================================
-- 2. ACCOUNT STATUS MASTER
-- =============================================================================

-- 2.1 Create an account status. Status codes are data, not database enums.
-- $1 code, $2 name, $3 description, $4 allows login, $5 active
INSERT INTO account_status (code, name, description, allows_login, is_active)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- 2.2 List account statuses.
SELECT account_status_id, code, name, description, allows_login, is_active
FROM account_status
ORDER BY name;

-- =============================================================================
-- 3. ACCOUNT CRUD
-- =============================================================================

-- 3.1 Create an account.
-- password_hash must already be produced by the application.
-- For SSO accounts, password_hash can be null and external_subject is required.
-- $1 username, $2 email, $3 display name, $4 password hash,
-- $5 external subject, $6 account status code, $7 auth policy code or null,
-- $8 timezone, $9 actor account ID or null during initial bootstrap
INSERT INTO app_account (
    username,
    email,
    display_name,
    password_hash,
    external_subject,
    account_status_id,
    authentication_policy_id,
    preferred_timezone,
    created_by,
    updated_by
)
SELECT
    $1,
    $2,
    $3,
    $4,
    $5,
    account_status.account_status_id,
    authentication_policy.authentication_policy_id,
    $8,
    $9,
    $9
FROM account_status
LEFT JOIN authentication_policy
       ON authentication_policy.code = $7
      AND authentication_policy.is_active
WHERE account_status.code = $6
  AND account_status.is_active
  AND ($7 IS NULL OR authentication_policy.authentication_policy_id IS NOT NULL)
RETURNING
    account_id,
    username,
    email,
    display_name,
    account_status_id,
    authentication_policy_id,
    preferred_timezone,
    created_at,
    version_no;

-- 3.2 Get an account and its assigned roles/data scopes.
-- $1 account ID
SELECT
    a.account_id,
    a.username,
    a.email,
    a.display_name,
    ast.code AS status_code,
    ast.name AS status_name,
    ap.code AS authentication_policy_code,
    a.preferred_timezone,
    a.failed_login_count,
    a.locked_until,
    a.last_login_at,
    a.created_at,
    a.updated_at,
    a.version_no,
    COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object('id', r.role_id, 'code', r.code, 'name', r.name)
            ORDER BY r.name
        )
        FROM account_role ar
        JOIN app_role r ON r.role_id = ar.role_id
        WHERE ar.account_id = a.account_id
          AND r.is_active
          AND ar.valid_from <= clock_timestamp()
          AND (ar.valid_until IS NULL OR ar.valid_until > clock_timestamp())
    ), '[]'::jsonb) AS roles,
    COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object('id', o.organization_id, 'code', o.code, 'name', o.name)
            ORDER BY o.name
        )
        FROM account_owner_access aoa
        JOIN organization o ON o.organization_id = aoa.owner_id
        WHERE aoa.account_id = a.account_id
    ), '[]'::jsonb) AS owner_access,
    COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object('id', w.warehouse_id, 'code', w.code, 'name', w.name)
            ORDER BY w.name
        )
        FROM account_warehouse_access awa
        JOIN warehouse w ON w.warehouse_id = awa.warehouse_id
        WHERE awa.account_id = a.account_id
    ), '[]'::jsonb) AS warehouse_access
FROM app_account a
JOIN account_status ast ON ast.account_status_id = a.account_status_id
LEFT JOIN authentication_policy ap
       ON ap.authentication_policy_id = a.authentication_policy_id
WHERE a.account_id = $1;

-- 3.3 Search/list accounts.
-- $1 search text or null, $2 status code or null, $3 page size, $4 offset
SELECT
    a.account_id,
    a.username,
    a.email,
    a.display_name,
    ast.code AS status_code,
    a.locked_until,
    a.last_login_at,
    a.version_no,
    count(*) OVER () AS total_rows
FROM app_account a
JOIN account_status ast ON ast.account_status_id = a.account_status_id
WHERE (
        $1 IS NULL OR
        a.username ILIKE '%' || $1 || '%' OR
        a.email ILIKE '%' || $1 || '%' OR
        a.display_name ILIKE '%' || $1 || '%'
      )
  AND ($2 IS NULL OR ast.code = $2)
ORDER BY a.display_name, a.account_id
LIMIT $3 OFFSET $4;

-- 3.4 Update an account using optimistic concurrency.
-- $1 account ID, $2 username, $3 email, $4 display name,
-- $5 status code, $6 auth policy code or null, $7 timezone,
-- $8 expected version, $9 actor account ID
UPDATE app_account a
SET username                 = $2,
    email                    = $3,
    display_name             = $4,
    account_status_id        = ast.account_status_id,
    authentication_policy_id = ap.authentication_policy_id,
    preferred_timezone       = $7,
    updated_at               = clock_timestamp(),
    updated_by               = $9,
    version_no               = a.version_no + 1
FROM account_status ast
LEFT JOIN authentication_policy ap
       ON ap.code = $6
      AND ap.is_active
WHERE a.account_id = $1
  AND a.version_no = $8
  AND ast.code = $5
  AND ast.is_active
  AND ($6 IS NULL OR ap.authentication_policy_id IS NOT NULL)
RETURNING
    a.account_id,
    a.username,
    a.email,
    a.display_name,
    a.account_status_id,
    a.authentication_policy_id,
    a.updated_at,
    a.version_no;

-- If this returns zero rows, the record was missing, the supplied master code
-- was invalid, or another request already changed the account version.

-- 3.5 Change/reset password and revoke every current session.
-- Run these statements in one database transaction.
-- Roll back and do not execute the session update if the account update returns
-- zero rows.
-- $1 account ID, $2 new application-generated password hash,
-- $3 actor account ID, $4 session revocation reason code
BEGIN;

UPDATE app_account
SET password_hash      = $2,
    failed_login_count = 0,
    locked_until       = NULL,
    updated_at         = clock_timestamp(),
    updated_by         = $3,
    version_no         = version_no + 1
WHERE account_id = $1
RETURNING account_id, updated_at, version_no;

UPDATE app_session s
SET revoked_at = clock_timestamp(),
    session_revocation_reason_id = rr.session_revocation_reason_id
FROM session_revocation_reason rr
WHERE s.account_id = $1
  AND s.revoked_at IS NULL
  AND rr.code = $4
  AND rr.is_active;

COMMIT;

-- 3.6 "Delete" account safely by changing its configured status and revoking
-- all sessions. $2 must be the desired disabled/deleted status code.
-- $1 account ID, $2 target status code, $3 actor ID, $4 expected version,
-- $5 session revocation reason code
-- Roll back and do not execute the session update if the account update returns
-- zero rows.
BEGIN;

UPDATE app_account a
SET account_status_id = ast.account_status_id,
    updated_at        = clock_timestamp(),
    updated_by        = $3,
    version_no        = a.version_no + 1
FROM account_status ast
WHERE a.account_id = $1
  AND a.version_no = $4
  AND ast.code = $2
  AND ast.is_active
RETURNING a.account_id, ast.code AS status_code, a.version_no;

UPDATE app_session s
SET revoked_at = clock_timestamp(),
    session_revocation_reason_id = rr.session_revocation_reason_id
FROM session_revocation_reason rr
WHERE s.account_id = $1
  AND s.revoked_at IS NULL
  AND rr.code = $5
  AND rr.is_active;

COMMIT;

-- =============================================================================
-- 4. ROLE, PERMISSION, MENU, AND DATA-SCOPE MANAGEMENT
-- =============================================================================

-- 4.1 Create role. $1 code, $2 name, $3 description, $4 actor ID
INSERT INTO app_role (code, name, description, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- 4.2 Create permission.
-- Example code shape: ACCOUNT.READ or INBOUND.APPROVE, but values are master data.
-- $1 code, $2 name, $3 module code, $4 description
INSERT INTO app_permission (code, name, module_code, description)
SELECT $1, $2, module.code, $4
FROM app_module module
WHERE module.code = $3 AND module.is_active
RETURNING *;

-- 4.3 Assign a role to an account.
-- $1 account ID, $2 role code, $3 valid from, $4 valid until, $5 actor ID
INSERT INTO account_role (
    account_id, role_id, valid_from, valid_until, assigned_by
)
SELECT $1, role_id, $3, $4, $5
FROM app_role
WHERE code = $2
  AND is_active
ON CONFLICT (account_id, role_id)
DO UPDATE SET
    valid_from  = EXCLUDED.valid_from,
    valid_until = EXCLUDED.valid_until,
    assigned_by = EXCLUDED.assigned_by
RETURNING *;

-- 4.4 Revoke a role from an account.
-- $1 account ID, $2 role code
DELETE FROM account_role ar
USING app_role r
WHERE ar.account_id = $1
  AND ar.role_id = r.role_id
  AND r.code = $2
RETURNING ar.*;

-- 4.5 Grant a permission to a role.
-- $1 role code, $2 permission code, $3 actor ID
INSERT INTO role_permission (role_id, permission_id, granted_by)
SELECT r.role_id, p.permission_id, $3
FROM app_role r
CROSS JOIN app_permission p
WHERE r.code = $1
  AND r.is_active
  AND p.code = $2
  AND p.is_active
ON CONFLICT (role_id, permission_id)
DO UPDATE SET
    granted_by = EXCLUDED.granted_by,
    granted_at = clock_timestamp()
RETURNING *;

-- 4.6 Revoke permission from role.
-- $1 role code, $2 permission code
DELETE FROM role_permission rp
USING app_role r, app_permission p
WHERE rp.role_id = r.role_id
  AND rp.permission_id = p.permission_id
  AND r.code = $1
  AND p.code = $2
RETURNING rp.*;

-- 4.7 Create a menu.
-- $1 parent menu ID or null, $2 code, $3 label, $4 route, $5 icon,
-- $6 display order, $7 require all mapped permissions
INSERT INTO app_menu (
    parent_menu_id,
    code,
    label,
    route,
    icon_name,
    display_order,
    require_all_permissions
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- 4.8 Map a required permission to a menu.
-- $1 menu code, $2 permission code
INSERT INTO menu_permission (menu_id, permission_id)
SELECT m.menu_id, p.permission_id
FROM app_menu m
CROSS JOIN app_permission p
WHERE m.code = $1
  AND p.code = $2
  AND m.is_active
  AND p.is_active
ON CONFLICT (menu_id, permission_id) DO NOTHING
RETURNING *;

-- 4.9 Remove required permission from menu.
-- $1 menu code, $2 permission code
DELETE FROM menu_permission mp
USING app_menu m, app_permission p
WHERE mp.menu_id = m.menu_id
  AND mp.permission_id = p.permission_id
  AND m.code = $1
  AND p.code = $2
RETURNING mp.*;

-- 4.10 Grant owner scope to account.
-- $1 account ID, $2 owner code, $3 actor ID
INSERT INTO account_owner_access (account_id, owner_id, granted_by)
SELECT $1, organization_id, $3
FROM organization
WHERE code = $2
  AND is_active
ON CONFLICT (account_id, owner_id)
DO UPDATE SET
    granted_by = EXCLUDED.granted_by,
    granted_at = clock_timestamp()
RETURNING *;

-- 4.11 Grant warehouse scope to account.
-- $1 account ID, $2 warehouse ID, $3 actor ID
INSERT INTO account_warehouse_access (account_id, warehouse_id, granted_by)
SELECT $1, warehouse_id, $3
FROM warehouse
WHERE warehouse_id = $2
  AND is_active
ON CONFLICT (account_id, warehouse_id)
DO UPDATE SET
    granted_by = EXCLUDED.granted_by,
    granted_at = clock_timestamp()
RETURNING *;

-- =============================================================================
-- 5. LOGIN
-- =============================================================================

-- Login is deliberately a two-step application flow. SQL must never receive a
-- plain password.

-- 5.1 Find the account and hash for password verification in the application.
-- Use the same generic "invalid credentials" response for every failure.
-- $1 username or email
SELECT
    a.account_id,
    a.username,
    a.email,
    a.display_name,
    a.password_hash,
    a.external_subject,
    ast.code AS status_code,
    ast.is_active AS status_is_active,
    ast.allows_login,
    a.failed_login_count,
    a.locked_until,
    (a.locked_until IS NOT NULL AND a.locked_until > clock_timestamp()) AS is_locked,
    p.authentication_policy_id,
    p.max_failed_attempts,
    p.lockout_seconds,
    p.session_ttl_seconds
FROM app_account a
JOIN account_status ast ON ast.account_status_id = a.account_status_id
JOIN LATERAL (
    SELECT
        candidate.authentication_policy_id,
        candidate.max_failed_attempts,
        candidate.lockout_seconds,
        candidate.session_ttl_seconds
    FROM authentication_policy candidate
    WHERE candidate.is_active
      AND (
          candidate.authentication_policy_id = a.authentication_policy_id OR
          candidate.is_default
      )
    ORDER BY
        (candidate.authentication_policy_id = a.authentication_policy_id) DESC,
        candidate.is_default DESC
    LIMIT 1
) p ON true
WHERE lower(a.username) = lower($1)
   OR lower(a.email) = lower($1)
LIMIT 1;

-- The application verifies the supplied password against password_hash using
-- Argon2id/bcrypt. Do not compare password text inside SQL.

-- 5.2 Record a failed password verification and apply configured lockout.
-- $1 account ID
WITH selected_policy AS (
    SELECT
        a.account_id,
        p.max_failed_attempts,
        p.lockout_seconds
    FROM app_account a
    JOIN LATERAL (
        SELECT candidate.max_failed_attempts, candidate.lockout_seconds
        FROM authentication_policy candidate
        WHERE candidate.is_active
          AND (
              candidate.authentication_policy_id = a.authentication_policy_id OR
              candidate.is_default
          )
        ORDER BY
            (candidate.authentication_policy_id = a.authentication_policy_id) DESC,
            candidate.is_default DESC
        LIMIT 1
    ) p ON true
    WHERE a.account_id = $1
)
UPDATE app_account a
SET failed_login_count = a.failed_login_count + 1,
    locked_until = CASE
        WHEN a.failed_login_count + 1 >= p.max_failed_attempts
        THEN clock_timestamp() + make_interval(secs => p.lockout_seconds)
        ELSE a.locked_until
    END,
    updated_at = clock_timestamp(),
    version_no = a.version_no + 1
FROM selected_policy p
WHERE a.account_id = p.account_id
RETURNING a.account_id, a.failed_login_count, a.locked_until;

-- 5.3 Complete successful login and create a session.
-- Run only after the application verifies the password.
-- $1 account ID, $2 generated session ID, $3 SHA-256 token hash,
-- $4 IP address or null, $5 user agent or null
WITH eligible_account AS (
    SELECT a.account_id, p.session_ttl_seconds
    FROM app_account a
    JOIN account_status ast ON ast.account_status_id = a.account_status_id
    JOIN LATERAL (
        SELECT candidate.session_ttl_seconds
        FROM authentication_policy candidate
        WHERE candidate.is_active
          AND (
              candidate.authentication_policy_id = a.authentication_policy_id OR
              candidate.is_default
          )
        ORDER BY
            (candidate.authentication_policy_id = a.authentication_policy_id) DESC,
            candidate.is_default DESC
        LIMIT 1
    ) p ON true
    WHERE a.account_id = $1
      AND ast.is_active
      AND ast.allows_login
      AND (a.locked_until IS NULL OR a.locked_until <= clock_timestamp())
),
updated_account AS (
    UPDATE app_account a
    SET failed_login_count = 0,
        locked_until       = NULL,
        last_login_at      = clock_timestamp(),
        updated_at         = clock_timestamp(),
        version_no         = a.version_no + 1
    FROM eligible_account ea
    WHERE a.account_id = ea.account_id
    RETURNING a.account_id
)
INSERT INTO app_session (
    session_id,
    account_id,
    token_hash,
    expires_at,
    ip_address,
    user_agent
)
SELECT
    $2,
    ua.account_id,
    $3,
    clock_timestamp() + make_interval(secs => ea.session_ttl_seconds),
    $4,
    $5
FROM updated_account ua
JOIN eligible_account ea ON ea.account_id = ua.account_id
RETURNING session_id, account_id, issued_at, expires_at;

-- If successful login returns zero rows, do not create a cookie: the account
-- became disabled/locked or has no active authentication policy.

-- =============================================================================
-- 6. SESSION VALIDATION AND LOGOUT
-- =============================================================================

-- 6.1 Validate session. $1 SHA-256 token hash
SELECT
    s.session_id,
    s.account_id,
    a.username,
    a.display_name,
    a.preferred_timezone,
    s.issued_at,
    s.expires_at,
    s.last_seen_at
FROM app_session s
JOIN app_account a ON a.account_id = s.account_id
JOIN account_status ast ON ast.account_status_id = a.account_status_id
WHERE s.token_hash = $1
  AND s.revoked_at IS NULL
  AND s.expires_at > clock_timestamp()
  AND ast.is_active
  AND ast.allows_login
  AND (a.locked_until IS NULL OR a.locked_until <= clock_timestamp());

-- 6.2 Touch last activity after successful validation. $1 token hash
UPDATE app_session
SET last_seen_at = clock_timestamp()
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND expires_at > clock_timestamp()
RETURNING session_id, last_seen_at;

-- 6.3 Logout current session.
-- $1 token hash, $2 session revocation reason code
UPDATE app_session s
SET revoked_at = clock_timestamp(),
    session_revocation_reason_id = rr.session_revocation_reason_id
FROM session_revocation_reason rr
WHERE s.token_hash = $1
  AND s.revoked_at IS NULL
  AND rr.code = $2
  AND rr.is_active
RETURNING s.session_id, s.revoked_at;

-- 6.4 Logout from all devices.
-- $1 current token hash, $2 session revocation reason code
WITH current_account AS (
    SELECT account_id
    FROM app_session
    WHERE token_hash = $1
      AND revoked_at IS NULL
      AND expires_at > clock_timestamp()
)
UPDATE app_session s
SET revoked_at = clock_timestamp(),
    session_revocation_reason_id = rr.session_revocation_reason_id
FROM current_account ca
JOIN session_revocation_reason rr
  ON rr.code = $2
 AND rr.is_active
WHERE s.account_id = ca.account_id
  AND s.revoked_at IS NULL
RETURNING s.session_id, s.revoked_at;

-- 6.5 Periodic cleanup. Retention boundary is supplied by configuration/job.
-- $1 timestamptz: delete sessions expired or revoked before this time
DELETE FROM app_session
WHERE (expires_at < $1)
   OR (revoked_at IS NOT NULL AND revoked_at < $1)
RETURNING session_id;

-- =============================================================================
-- 7. CURRENT SESSION PERMISSIONS AND MENU
-- =============================================================================

-- 7.1 List all effective permissions for a valid session.
-- $1 SHA-256 token hash
SELECT DISTINCT
    p.permission_id,
    p.code,
    p.name,
    p.module_code
FROM app_session s
JOIN app_account a ON a.account_id = s.account_id
JOIN account_status ast ON ast.account_status_id = a.account_status_id
JOIN account_role ar ON ar.account_id = a.account_id
JOIN app_role r ON r.role_id = ar.role_id
JOIN role_permission rp ON rp.role_id = r.role_id
JOIN app_permission p ON p.permission_id = rp.permission_id
WHERE s.token_hash = $1
  AND s.revoked_at IS NULL
  AND s.expires_at > clock_timestamp()
  AND ast.is_active
  AND ast.allows_login
  AND (a.locked_until IS NULL OR a.locked_until <= clock_timestamp())
  AND r.is_active
  AND p.is_active
  AND ar.valid_from <= clock_timestamp()
  AND (ar.valid_until IS NULL OR ar.valid_until > clock_timestamp())
ORDER BY p.module_code, p.code;

-- 7.2 Check one permission for one valid session.
-- $1 token hash, $2 permission code
SELECT EXISTS (
    SELECT 1
    FROM app_session s
    JOIN app_account a ON a.account_id = s.account_id
    JOIN account_status ast ON ast.account_status_id = a.account_status_id
    JOIN account_role ar ON ar.account_id = a.account_id
    JOIN app_role r ON r.role_id = ar.role_id
    JOIN role_permission rp ON rp.role_id = r.role_id
    JOIN app_permission p ON p.permission_id = rp.permission_id
    WHERE s.token_hash = $1
      AND p.code = $2
      AND s.revoked_at IS NULL
      AND s.expires_at > clock_timestamp()
      AND ast.is_active
      AND ast.allows_login
      AND (a.locked_until IS NULL OR a.locked_until <= clock_timestamp())
      AND r.is_active
      AND p.is_active
      AND ar.valid_from <= clock_timestamp()
      AND (ar.valid_until IS NULL OR ar.valid_until > clock_timestamp())
) AS is_allowed;

-- 7.3 Return menus visible to the session.
-- Menus without permission mappings are visible to every valid account.
-- For mapped menus, require_all_permissions selects AND versus OR behavior.
-- $1 SHA-256 token hash
WITH valid_account AS (
    SELECT a.account_id
    FROM app_session s
    JOIN app_account a ON a.account_id = s.account_id
    JOIN account_status ast ON ast.account_status_id = a.account_status_id
    WHERE s.token_hash = $1
      AND s.revoked_at IS NULL
      AND s.expires_at > clock_timestamp()
      AND ast.is_active
      AND ast.allows_login
      AND (a.locked_until IS NULL OR a.locked_until <= clock_timestamp())
),
granted_permissions AS (
    SELECT DISTINCT rp.permission_id
    FROM valid_account va
    JOIN account_role ar ON ar.account_id = va.account_id
    JOIN app_role r ON r.role_id = ar.role_id
    JOIN role_permission rp ON rp.role_id = r.role_id
    JOIN app_permission p ON p.permission_id = rp.permission_id
    WHERE r.is_active
      AND p.is_active
      AND ar.valid_from <= clock_timestamp()
      AND (ar.valid_until IS NULL OR ar.valid_until > clock_timestamp())
)
SELECT
    m.menu_id,
    m.parent_menu_id,
    m.code,
    m.label,
    m.route,
    m.icon_name,
    m.display_order
FROM app_menu m
WHERE m.is_active
  AND EXISTS (SELECT 1 FROM valid_account)
  AND (
      NOT EXISTS (
          SELECT 1
          FROM menu_permission mp
          WHERE mp.menu_id = m.menu_id
      )
      OR (
          m.require_all_permissions
          AND NOT EXISTS (
              SELECT 1
              FROM menu_permission required_mp
              WHERE required_mp.menu_id = m.menu_id
                AND NOT EXISTS (
                    SELECT 1
                    FROM granted_permissions gp
                    WHERE gp.permission_id = required_mp.permission_id
                )
          )
      )
      OR (
          NOT m.require_all_permissions
          AND EXISTS (
              SELECT 1
              FROM menu_permission allowed_mp
              JOIN granted_permissions gp
                ON gp.permission_id = allowed_mp.permission_id
              WHERE allowed_mp.menu_id = m.menu_id
          )
      )
  )
ORDER BY m.parent_menu_id NULLS FIRST, m.display_order, m.label;

-- 7.4 Return the current session context in one result.
-- $1 SHA-256 token hash
WITH valid_account AS (
    SELECT a.account_id, a.username, a.display_name, s.session_id, s.expires_at
    FROM app_session s
    JOIN app_account a ON a.account_id = s.account_id
    JOIN account_status ast ON ast.account_status_id = a.account_status_id
    WHERE s.token_hash = $1
      AND s.revoked_at IS NULL
      AND s.expires_at > clock_timestamp()
      AND ast.is_active
      AND ast.allows_login
      AND (a.locked_until IS NULL OR a.locked_until <= clock_timestamp())
)
SELECT
    va.session_id,
    va.expires_at,
    va.account_id,
    va.username,
    va.display_name,
    COALESCE((
        SELECT jsonb_agg(DISTINCT r.code)
        FROM account_role ar
        JOIN app_role r ON r.role_id = ar.role_id
        WHERE ar.account_id = va.account_id
          AND r.is_active
          AND ar.valid_from <= clock_timestamp()
          AND (ar.valid_until IS NULL OR ar.valid_until > clock_timestamp())
    ), '[]'::jsonb) AS roles,
    COALESCE((
        SELECT jsonb_agg(DISTINCT p.code)
        FROM account_role ar
        JOIN app_role r ON r.role_id = ar.role_id
        JOIN role_permission rp ON rp.role_id = r.role_id
        JOIN app_permission p ON p.permission_id = rp.permission_id
        WHERE ar.account_id = va.account_id
          AND r.is_active
          AND p.is_active
          AND ar.valid_from <= clock_timestamp()
          AND (ar.valid_until IS NULL OR ar.valid_until > clock_timestamp())
    ), '[]'::jsonb) AS permissions,
    COALESCE((
        SELECT jsonb_agg(aoa.owner_id)
        FROM account_owner_access aoa
        WHERE aoa.account_id = va.account_id
    ), '[]'::jsonb) AS owner_ids,
    COALESCE((
        SELECT jsonb_agg(awa.warehouse_id)
        FROM account_warehouse_access awa
        WHERE awa.account_id = va.account_id
    ), '[]'::jsonb) AS warehouse_ids
FROM valid_account va;

-- 7.5 Permission plus owner/warehouse data-scope check.
-- Use this before loading protected WMS records.
-- $1 token hash, $2 permission code, $3 owner ID, $4 warehouse ID
SELECT EXISTS (
    SELECT 1
    FROM app_session s
    JOIN app_account a ON a.account_id = s.account_id
    JOIN account_status ast ON ast.account_status_id = a.account_status_id
    JOIN account_role ar ON ar.account_id = a.account_id
    JOIN app_role r ON r.role_id = ar.role_id
    JOIN role_permission rp ON rp.role_id = r.role_id
    JOIN app_permission p ON p.permission_id = rp.permission_id
    JOIN account_owner_access aoa
      ON aoa.account_id = a.account_id
     AND aoa.owner_id = $3
    JOIN account_warehouse_access awa
      ON awa.account_id = a.account_id
     AND awa.warehouse_id = $4
    WHERE s.token_hash = $1
      AND p.code = $2
      AND s.revoked_at IS NULL
      AND s.expires_at > clock_timestamp()
      AND ast.is_active
      AND ast.allows_login
      AND (a.locked_until IS NULL OR a.locked_until <= clock_timestamp())
      AND r.is_active
      AND p.is_active
      AND ar.valid_from <= clock_timestamp()
      AND (ar.valid_until IS NULL OR ar.valid_until > clock_timestamp())
) AS is_allowed;
