# Backend hardening

Hardening adds safety controls around the existing WMS business workflows. It
does not change inbound, inventory, outbound, or billing calculations.

## Included controls

- Bearer sessions remain server-side in `app_session`; revocation is immediate.
- Module RBAC checks `<MODULE>.READ` for GET/HEAD and `<MODULE>.WRITE` for
  mutations. Authentication self-service routes (`me`, logout, logout-all)
  require a valid session but no business-module grant.
- Exact-origin CORS, per-IP rate limiting, maximum request sizes, security
  headers, trusted-proxy configuration, HTTP timeouts, and graceful shutdown.
- Every response carries `X-Request-ID`; JSON responses also contain
  `request_id` so frontend error reports can be traced.
- Every POST, PUT, PATCH, and DELETE attempt is written to `api_audit_log` with
  its actor when authentication succeeded. Request bodies and bearer tokens are
  deliberately not stored.
- Versioned, serialized startup migrations are recorded in `schema_migration`.
- Reusable roles contribute permissions to an account while direct grants remain
  available for exceptions. Account and role changes protect the final security
  administrator from accidental lockout.
- A route-level OpenAPI document is served at `/openapi.json`, with Swagger UI
  at `/docs`. The module guides in `docs/` remain the source for exact request
  bodies and workflow examples.

## Bootstrap RBAC safely

The development default is `AUTH_RBAC_ENFORCED=false`, so an existing local
study database is not locked immediately. Before changing it to `true`, ensure
at least one administrator owns all grants:

1. Temporarily set `AUTH_BOOTSTRAP_ADMIN_ENABLED=true` and provide a unique
   password of at least 12 characters.
2. Start the API once. A new bootstrap administrator is created, or the named
   existing administrator is granted every active permission.
3. Log in as that administrator and set
   `AUTH_BOOTSTRAP_ADMIN_ENABLED=false`. Leaving it enabled would intentionally
   restore all permissions to that account at each startup.
4. Set `AUTH_RBAC_ENFORCED=true` and restart.

The login and `/api/v1/auth/me` response includes the account's permission
codes, including permissions inherited from roles. Use these security endpoints
with the administrator bearer token:

```text
GET    /api/v1/security/permissions
GET    /api/v1/security/accounts/:account_id/permissions
POST   /api/v1/security/accounts/:account_id/permissions
DELETE /api/v1/security/accounts/:account_id/permissions/:permission_id
```

Grant request:

```json
{
  "permission_id": "UUID-FROM-GET-SECURITY-PERMISSIONS"
}
```

Grants are idempotent. The API refuses to revoke the final active account's
`SECURITY.WRITE` permission, preventing an accidental total administrator
lockout.

## Environment settings

```dotenv
APP_ENV=development
APP_TRUSTED_PROXIES=
APP_READ_HEADER_TIMEOUT_SECONDS=5
APP_READ_TIMEOUT_SECONDS=30
APP_WRITE_TIMEOUT_SECONDS=30
APP_IDLE_TIMEOUT_SECONDS=60
APP_SHUTDOWN_TIMEOUT_SECONDS=10

CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
SECURITY_RATE_LIMIT_PER_MINUTE=600
SECURITY_MAX_REQUEST_BODY_BYTES=1048576
AUTH_RBAC_ENFORCED=false
```

`APP_TRUSTED_PROXIES` must contain only the proxy IP addresses or CIDR ranges
that are allowed to supply forwarded client addresses. Keep it empty when the
API is reached directly.

`APP_ENV=production` fails to start unless RBAC is enabled, database auto-create
is disabled, PostgreSQL TLS is enabled, and explicit non-wildcard CORS origins
are supplied.

## Verification

Run unit tests:

```powershell
go test ./...
go vet ./...
```

Run PostgreSQL integration tests against the configured test-capable database:

```powershell
$env:WMS_INTEGRATION_TEST = "1"
go test -p 1 ./...
```

`-p 1` is intentional: some legacy integration cases verify migrations against
the same development schema and would otherwise compete for PostgreSQL DDL
locks. Each workflow still runs both existing-schema and isolated fresh-schema
coverage where applicable.

After starting the API, verify:

```text
GET http://localhost:8080/health
GET http://localhost:8080/docs
GET http://localhost:8080/api/v1/auth/me
```

For `/api/v1/auth/me`, send `Authorization: Bearer <token>`. Its response should
have a request ID and permission list. A disallowed browser origin should return
403, an oversized request should return 413, and a missing module grant should
return 403 when RBAC is enabled.

## Deployment responsibilities

The application baseline is hardened, but production infrastructure still must
terminate HTTPS, store secrets outside source control, back up PostgreSQL,
monitor logs and health, and define retention/archival for `api_audit_log`. The
built-in rate limiter is per process; a multi-instance deployment should enforce
a shared limit at the gateway or use a distributed limiter.
