# ADR-0004: Code is the source of truth; OpenAPI is generated at build time

- **Status:** Accepted
- **Date:** 2026-09-08
- **Deciders:** Project team
- **Tags:** api, contract, frontend

## Context and Problem Statement

The backend (Go) and frontend (React/TypeScript) are developed together and must
agree on the HTTP contract. We need a single authoritative definition of the API
— but in practice the implementation leads, and the spec follows the code.

## Decision Drivers

- One authoritative description of endpoints and schemas.
- Enable typed frontend clients.
- Avoid the spec and implementation diverging over time.

## Considered Options

- **Code-first: the Go implementation is authoritative and the OpenAPI document
  is generated from it** (at build time).
- **Spec-first: hand-maintain OpenAPI as the contract, code follows it.**
- **No formal contract** — rely on docs and convention.

## Decision

We chose **code-first**: the **Go implementation is the source of truth**, and
`api/openapi.yaml` is treated as a **generated artifact produced at build
time** ("製本" generated during build). The frontend consumes that generated
spec to produce its types (`openapi-typescript` → `frontend/src/api/schema.ts`,
used via `openapi-typescript-fetch`).

> Note: at the time of writing, `api/openapi.yaml` is still maintained by hand
> and the build-time generation is a target state, not yet wired up. The
> *positioning* — code is truth, spec is generated — is the decision; the
> generation step is a follow-up.

## Consequences

- **Positive:** No manual spec/implementation reconciliation once generation is
  in place — the spec cannot drift from the code because it is derived from it.
  The frontend still gets end-to-end typing from the generated spec.
- **Negative / trade-offs:** Until build-time generation is implemented, the
  hand-maintained `openapi.yaml` **can** fall out of sync with the handlers, so
  this remains a manual discipline in the interim.
- **PoC note:** The code-as-truth positioning is a **standing decision**, not a
  PoC-only choice. Only the generation mechanism is outstanding.

## Follow-ups

- [ ] Wire up build-time OpenAPI generation from the Go code.
- [ ] Regenerate `frontend/src/api/schema.ts` from the generated spec as part of
      the build.
