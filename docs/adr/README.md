# Architecture Decision Records (ADRs)

This directory records the **architectural decisions** made on the Release
Catalog project. An ADR captures *what* was decided, *why*, and *what
trade-offs* it implies — so future contributors (and future us) understand the
reasoning instead of reverse-engineering it from code.

We use a lightweight [MADR](https://adr.github.io/madr/)-style format. The goal
is that any ADR takes **5–15 minutes** to write. If it feels heavier than that,
the decision is probably too small to record.

## When to write an ADR

Write one when a decision is **hard to reverse**, **affects multiple parts of
the system**, or **someone will reasonably ask "why is it done this way?"**
later. Examples:

- Choosing or replacing a datastore, framework, or protocol.
- Introducing an architectural pattern (layering, ports-and-adapters, etc.).
- Defining a contract boundary (API shape, source of truth).
- Deliberately deferring or stubbing something (e.g. auth, a real integration).

You do **not** need an ADR for routine code: renames, bug fixes, formatting,
or obvious local choices.

## PoC status note

This project is currently an **initial PoC**. Many decisions are provisional and
made to optimize iteration speed, not production hardening. Those ADRs are still
recorded as `Accepted`, but include a **PoC note** describing what would trigger
a revisit. This turns the ADR log into a ready-made "revisit before production"
checklist.

## Conventions

- **One decision per file.** Filename: `NNNN-kebab-case-title.md` (e.g.
  `0001-use-sqlite-for-the-poc.md`).
- **Numbering** is a zero-padded, monotonically increasing integer starting at
  `0001`. `0000-template.md` is the template and is not a real decision.
- **Never rewrite history.** Once an ADR is `Accepted`, don't edit the decision.
  If it changes, write a **new** ADR and set the old one's status to
  `Superseded by ADR-NNNN`, adding a link. Corrections of typos are fine.
- **Status values:** `Proposed`, `Accepted`, `Deprecated`, `Superseded by ADR-NNNN`.

## How to add a new ADR

1. Copy `0000-template.md` to the next number: `NNNN-your-title.md`.
2. Fill in Context, Options, Decision, and Consequences.
3. Set the status (`Proposed` while under discussion, `Accepted` once agreed).
4. Add a row to the **Index** below.

## Index

| ADR | Title | Status |
|-----|-------|--------|
| [0001](0001-use-sqlite-for-the-poc.md) | Use SQLite as the default store, with external PostgreSQL as an option | Accepted |
| [0002](0002-ports-and-adapters-with-package-by-feature.md) | Clean Architecture with package-by-feature | Accepted |
| [0003](0003-dependency-track-client-behind-a-port-with-a-stub.md) | Dependency-Track client behind a port, with an in-memory stub | Accepted |
| [0004](0004-openapi-as-the-api-contract-source-of-truth.md) | Code is the source of truth; OpenAPI generated at build time | Accepted |
| [0005](0005-recursive-bom-link-graph-resolution.md) | Recursive BOM-Link graph resolution with bounds | Accepted |
| [0006](0006-defer-authentication-with-a-placeholder.md) | Defer authentication with a placeholder middleware | Accepted |
| [0007](0007-git-like-version-model.md) | Git-like version model replacing snapshots and current-state | Accepted |
| [0008](0008-role-tagged-dt-projects-per-version.md) | Multiple role-tagged DT projects per version (ROOT/PROFILE/SUB) | Accepted |
| [0009](0009-version-lineage-graph-separate-from-sbom-graph.md) | Version-lineage graph (RC-DB-only) separate from the BOM-Link SBOM graph | Accepted |
| [0010](0010-release-branch-auto-provisions-first-version.md) | Release branch auto-provisions its first version, ROOT inherited from the forked main version | Accepted |
| [0011](0011-lineage-graph-axis-layout.md) | Version-lineage graph layout per axis (version blocks vs same-date columns) | Accepted |
