# ADR-0001: Use SQLite as the default store, with external PostgreSQL as an option

- **Status:** Accepted
- **Date:** 2026-09-08
- **Deciders:** Project team
- **Tags:** backend, database, poc

## Context and Problem Statement

The Release Catalog needs persistent storage for products, branch lines,
current state, snapshots, and the BOM-Link index. At the PoC stage we want the
fewest moving parts possible so contributors can clone, build, and run without
provisioning external infrastructure — while keeping the option to run against
an external database open.

## Decision Drivers

- Zero-setup local development (no separate DB server to run).
- Fast iteration; the schema is still changing.
- Allow an external database (e.g. PostgreSQL) to be used without redesign.

## Considered Options

- **SQLite (embedded file) as the default.**
- **PostgreSQL as the default** — production-grade, but requires a running
  instance for every developer.
- **In-memory only** — simplest, but loses data across restarts.

## Decision

We chose **SQLite** (`mattn/go-sqlite3`, CGo) as the **default** store, opened
with `_journal_mode=WAL` and `_foreign_keys=on`, with the schema applied on
startup by `database.Migrate`. The design intent is that an **external database
can be plugged in** later; **PostgreSQL** is the presumed target simply because
it is the standard OSS choice — no other engine (e.g. MySQL) has been evaluated,
and none is ruled out. SQLite is not strictly "PoC-only": it may remain viable
for some deployments, with PostgreSQL available as an opt-in for those that need
it.

## Consequences

- **Positive:** No external dependency to run the app by default; tests and
  local dev are trivial. Foreign keys and WAL give reasonable
  integrity/concurrency.
- **Negative / trade-offs:** CGo requires a C toolchain to build. SQLite's
  concurrency and type handling differ from PostgreSQL, so behaviors will need
  re-validation when running against an external DB. Migrations are currently a
  single inlined SQL string, not versioned migration files.
- **PoC note:** The trigger to introduce PostgreSQL support is **"PoC complete
  and the specification has stabilized."** Repository access is already funneled
  through per-feature interfaces (see ADR-0002), so an external-DB adapter can be
  added without touching usecases. Introduce a versioned migration tool at that
  point.

## Follow-ups

- [ ] After the PoC stabilizes, add a PostgreSQL adapter and a versioned
      migration tool.
- [ ] Confirm SHA-256 / integer column semantics port cleanly to PostgreSQL.
