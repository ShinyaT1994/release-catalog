# ADR-0002: Clean Architecture with package-by-feature

- **Status:** Accepted
- **Date:** 2026-09-08
- **Deciders:** Project team
- **Tags:** backend, architecture

## Context and Problem Statement

The backend has several cohesive concerns (products, branch lines, releases,
graph resolution) plus an external dependency (Dependency-Track). We need a
structure that keeps business logic testable and independent of frameworks and
I/O, while staying easy to navigate for a small team, and that leaves room to
grow.

## Decision Drivers

- Keep domain/usecase logic free of Echo and SQL specifics.
- Organize around **domains**, not around technical system layers.
- Keep implementation cost low while preserving reasonable extensibility.

## Considered Options

- **Package-by-layer** — top-level `handlers/`, `usecases/`, `repositories/`.
  Considered, but rejected: it frames the system from a *technical layer* view
  rather than a *domain* view, which we did not want.
- **Package-by-feature + Clean Architecture** — each feature owns its domain,
  usecase, repository, handler, and ports.
- **Flat package** — considered, but rejected: it loses the spatial/structural
  relationships between concerns, making the codebase harder to reason about.

## Decision

This is a **deliberate** design decision: **Clean Architecture combined with
package-by-feature**. Each feature package (`product`, `branch`, `release`,
`graph`) contains its own `domain.go`, `usecase.go`, `repository.go`,
`handler.go`, and `port.go`. Usecases depend on **ports** (interfaces), not
concrete adapters. Cross-feature dependencies are satisfied by small adapter
structs assembled at the composition root (`cmd/server/main.go`) using
**manual constructor injection** (no DI framework), e.g. `productFinderAdapter`,
`snapshotFinderAdapter`, `graphCSFinderAdapter` — so no feature imports another
feature's internals directly. The layering is organized by **domain**, and the
technical-layer axis is intentionally omitted.

## Consequences

- **Positive:** Features are self-contained and unit-testable with fake ports
  (see the `*_test.go` files). The composition root is the single place that
  knows how features connect. The structure balances low implementation cost
  with room to extend.
- **Negative / trade-offs:** More boilerplate — ports and adapter structs must
  be defined and maintained. Manual wiring in `main.go` grows as features are
  added.
- **PoC note:** Manual constructor injection is chosen deliberately after
  weighing DI frameworks. Two families were considered:
  - **Compile-time DI (code generation), e.g. `google/wire`** — generates the
    wiring code at build time; zero runtime overhead; missing dependencies fail
    at compile time; the output is ordinary Go, so it can be removed easily.
  - **Runtime DI (reflection), e.g. `uber-go/fx` / `dig`** — resolves the
    dependency graph at runtime and adds lifecycle management (start/stop
    hooks, module composition), at the cost of runtime overhead, harder
    debugging, and wiring errors surfacing only at runtime.

  At the current scale (~10 dependencies, single process, PoC with a
  frequently-changing structure) a DI framework would not yet solve a real
  pain: the decoupling and testability we want come from the ports/adapters
  design itself, not from a DI container. Manual wiring keeps the composition
  root as plain, IDE-navigable Go that fails fast at compile time. We therefore
  continue with manual constructor injection for now.

## Follow-ups

- [ ] Re-evaluate if `main.go` wiring becomes hard to read as adapters grow —
      first candidate is **`google/wire`** (compile-time generation, lowest
      risk, removable at any time).
- [ ] Consider **`uber-go/fx`** only if application lifecycle management
      (graceful shutdown, background workers, multi-server startup ordering)
      becomes complex enough to justify a runtime DI container.
