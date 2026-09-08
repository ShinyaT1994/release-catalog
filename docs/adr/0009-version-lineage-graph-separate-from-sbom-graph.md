# ADR-0009: Version-lineage graph (RC-DB-only) separate from the BOM-Link SBOM graph

- **Status:** Accepted
- **Date:** 2026-09-08
- **Deciders:** Project team
- **Tags:** backend, graph, frontend, visualization

## Context and Problem Statement

Users want to understand releases at two different levels: (1) the **release
timeline** — how versions relate across the main line and forked release lines;
and (2) the **SBOM tree** — the DT-project composition for a given version. The
existing graph feature resolves BOM-Link references from Dependency-Track. We
need to decide whether these two concerns share one graph or are separate.

## Decision Drivers

- The two views answer different questions (release lineage vs SBOM structure).
- The lineage view should not depend on Dependency-Track availability.
- Each view has a natural, different home in the UI.

## Considered Options

- **One combined graph** annotating SBOM nodes with lineage.
- **Two distinct graphs**, each with its own data source, endpoint, and page.

## Decision

We keep **two distinct graphs**:

1. **Version-lineage graph** — **nodes = versions, edges = parent/fork lineage**.
   Built **entirely from the Release Catalog DB** (branches + versions + fork
   edges); it makes **no Dependency-Track calls** and therefore **cannot return
   502**. Main is one line; each release branch (one per release profile) is its
   own line forked from a main node. Ordered by **release datetime**, with a
   toggle to order by version. **Location:** the **product main page**
   (release-timeline overview).
2. **BOM-Link SBOM graph** (existing resolver, ADR-0005) — DT-project
   composition. Resolved **per role-tagged project separately** (ROOT / PROFILE
   / SUB each get their own graph). Depends on DT and may `502`. **Location:**
   the **branch/version detail page**, per version.

## Consequences

- **Positive:** Each view is simple and purpose-built. The lineage graph is
  always available (no external dependency) and cheap. The SBOM graph keeps its
  existing semantics, now keyed per role-tagged project.
- **Negative / trade-offs:** Two graph builders and two sets of endpoints to
  maintain. Cross-referencing (e.g. jumping from a lineage node to that
  version's SBOM graphs) must be wired in the UI.
- **PoC note:** Product **risk** overlays (vulnerabilities per version, risk
  roll-up across ROOT/PROFILE/SUB) are **out of scope** now; both graphs are
  designed so risk annotations can be added later without structural change.

## Follow-ups

- [ ] Add risk/vulnerability annotations to the graphs in a later phase.
- [ ] Wire UI navigation from a lineage node to its per-project SBOM graphs.
