# ADR-0006: Defer authentication with a placeholder middleware

- **Status:** Accepted
- **Date:** 2026-09-08
- **Deciders:** Project team
- **Tags:** backend, security, poc

## Context and Problem Statement

The API will eventually need authentication and authorization. At the PoC stage,
the priority is exercising the domain (branches, snapshots, graph resolution),
not access control. We still want a clear, single seam where auth will live so
it can be added without restructuring the routing.

## Decision Drivers

- Don't block PoC iteration on an auth implementation.
- Reserve an explicit insertion point for auth in the middleware chain.
- Make the current "no auth" state obvious, not accidental.

## Considered Options

- **Placeholder middleware that passes all requests through.**
- **Implement authentication now** (e.g. JWT / API key / OIDC).
- **No auth wiring at all** (add it later wherever convenient).

## Decision

Deferring authentication is a **deliberate** decision. We use a **placeholder
middleware** (`middleware.AuthPlaceholder`) registered in the Echo chain in
`cmd/server/main.go`. It currently calls `next` for every request and carries a
`TODO: Implement JWT / API Key authentication`, reserving the exact location
where real auth will be enforced. The **authentication method is undecided**
(JWT, API key, OIDC, etc. are all still open).

## Consequences

- **Positive:** The middleware chain already has the auth seam; adding real auth
  is a localized change. The passthrough is explicit and discoverable in code.
- **Negative / trade-offs:** The API is currently unauthenticated — every
  endpoint is open.
- **PoC note:** This is a provisional PoC decision. The auth mechanism will be
  chosen and implemented in this middleware later; when that happens, update
  this ADR or write a superseding one.

## Follow-ups

- [ ] Choose an authentication method (JWT / API key / OIDC / other).
- [ ] Implement authentication (and authorization) in the placeholder middleware.
