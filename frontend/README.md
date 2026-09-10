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

## Frontend rules

- Keep backend data in TanStack Query, not global client state.
- Keep filters and pagination in the URL when they should be shareable.
- Preserve decimal quantities as strings or `Decimal` values; do not convert
  inventory quantities to JavaScript `number`.
- Treat permission-based navigation as presentation only. The API remains the
  authorization authority.
- Show API request IDs in support-friendly error states.
- Design every workflow for keyboard, touch, mobile, and wide desktop layouts.
