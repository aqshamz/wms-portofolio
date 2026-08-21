# WMS API

Initial backend foundation using Gin, GORM, and PostgreSQL.

## Run locally

1. Copy `.env.example` to `.env` and review the credentials.
2. Make sure PostgreSQL is listening on port `7654`.
3. Run `go run .`.
4. Check `GET http://localhost:8080/health`.

When `DB_AUTO_CREATE=true`, startup connects to the PostgreSQL `postgres`
administrative database and creates the configured `wms` database if it does
not exist. It also creates the `wms` schema. GORM table migration is deliberately
not enabled until the entity models are implemented and reviewed against
`../database/wms_schema.sql`.

## Package rules

- `routes`: endpoint registration only.
- `controller`: HTTP request validation and response handling, grouped by domain.
- `services`: business logic, grouped by domain.
- `repository`: the only application layer allowed to query GORM; one file per table.
- `models`: database entity structs; one file per table.
- `dto`: JSON request and response structs, grouped by domain.
- `middleware`: authentication, authorization, logging, and other guards.
- `utils`: small application-wide helpers without domain business logic.
- `config`: environment and infrastructure configuration.

Dependencies must flow in this direction:

`routes -> controller -> services -> repository -> models`

DTOs may be used by controllers and services. Controllers and services must not
query the database directly.

## Authentication

Startup migrates only these authentication tables: `account_status`,
`authentication_policy`, `session_revocation_reason`, `app_account`, and
`app_session`. With the local `.env`, an initial development administrator is
created once:

- username: `admin`
- password: `Admin@123456`

Change this password before using the API outside local development. The
bootstrap process never overwrites an existing account.

Authentication endpoints:

- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/logout-all`

Login request:

```json
{
  "identifier": "admin",
  "password": "Admin@123456"
}
```

For protected endpoints, copy `data.token` from the login response into the
request header:

```text
Authorization: Bearer <token>
```

Only the SHA-256 hash of this opaque token is stored in PostgreSQL.
