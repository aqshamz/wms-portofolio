# WMS Frontend

Responsive warehouse-management frontend built with Next.js, TypeScript,
Tailwind CSS, and TanStack.

## Local development

1. Copy `.env.example` to `.env.local`.
2. Start the Go API on `http://localhost:8080`.
3. Install dependencies with `npm install`.
4. Run `npm run dev` and open `http://localhost:3000`.

## Project structure

- `src/app`: routes, layouts, and route-level composition.
- `src/components/ui`: reusable visual primitives.
- `src/components/layout`: application-wide layout components.
- `src/features`: domain-oriented screens and components.
- `src/lib/api`: typed API contracts and transport.
- `src/lib`: framework-independent helpers and configuration.
- `src/test`: shared test setup.

## Authentication flow

The Go API remains responsible for validating credentials, creating sessions,
and enforcing permissions. Next.js acts as a backend-for-frontend (BFF):

1. The browser posts the typed login form to `POST /api/auth/login`.
2. The route handler validates the payload with Zod and calls the Go API.
3. The opaque bearer token is stored in a `Secure`, `HttpOnly`, `SameSite=Lax`
   cookie. The token is never returned to browser JavaScript or local storage.
4. The workspace server layout validates the session through `/auth/me` before
   rendering protected content.
5. Browser API calls use `/api/backend/api/v1/...`; the proxy attaches the
   bearer token on the server.
6. A confirmed `401` clears the cookie. A temporary API or database outage
   returns `503` and preserves the session.

Authorization helpers filter navigation and report categories using the
permission codes returned by the backend. This improves the UI, but every data
endpoint must still authorize the request on the Go API.

The implementation demonstrates useful current TypeScript patterns:

- Zod schemas provide runtime validation and infer the matching TypeScript type.
- `SessionState` is a discriminated union, forcing callers to handle signed-in,
  signed-out, and service-unavailable states separately.
- Server Components validate identity and pass only the safe user profile into
  a small Client Component context for interactive permission-aware UI.

## Frontend rules

- Keep backend data in TanStack Query, not global client state.
- Keep filters and pagination in the URL when they should be shareable.
- Preserve decimal quantities as strings or `Decimal` values; do not convert
  inventory quantities to JavaScript `number`.
- Treat permission-based navigation as presentation only. The API remains the
  authorization authority.
- Show API request IDs in support-friendly error states.
- Design every workflow for keyboard, touch, mobile, and wide desktop layouts.
