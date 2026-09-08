# ADR-0005: Recursive BOM-Link graph resolution with bounds

- **Status:** Accepted
- **Date:** 2026-09-08
- **Deciders:** Project team
- **Tags:** backend, graph, algorithm

## Context and Problem Statement

A release's SBOM composition is expressed through CycloneDX **BOM-Link**
external references (`urn:cdx:<serial>/<version>#<bom-ref>`). To present the full
release dependency graph we must follow these links from a root DT project into
child projects, which can form arbitrary — and possibly cyclic — graphs.

## Decision Drivers

- Traverse an unbounded, possibly cyclic reference graph safely.
- Surface links that cannot be resolved rather than failing silently.
- Bound resource usage so a pathological graph cannot exhaust the server.

## Considered Options

- **Recursive DFS with a visited set and depth/node bounds.**
- **Iterative BFS with a queue.**
- **Precompute/persist the whole graph offline.**

## Decision

We implemented **recursive depth-first resolution** in `graph.Resolver`. This
was the straightforward approach taken; alternatives (BFS, offline precompute)
were **not formally compared** at decision time. It maintains a `visited` set and
a `projectUUID -> nodeID` map. Cycles are detected via the visited set (counted
in `cyclesDetected`) and the existing node is reused. Unresolvable references
produce nodes/edges with a `resolutionStatus` (`missing_project`, `missing_bom`,
`missing_bom_ref`, `invalid`) rather than aborting. Traversal is bounded by
`maxDepth` (default 10) and `maxNodes` (default 1000), both overridable via query
parameters. When DT is unreachable, graph endpoints return HTTP `502`.

## Consequences

- **Positive:** Cycles are handled deterministically; partial/broken graphs are
  reported with explicit statuses and summarized in `metadata`. Bounds cap
  worst-case cost.
- **Negative / trade-offs:** Recursion depth is tied to `maxDepth`. The result
  is computed on demand per request; there is no caching or persistence yet.
- **Known limitation (accepted):** `metadata.maxDepthReached` is currently always
  reported as `false`. This is **intentionally left as-is for now** — accurate
  depth-limit reporting is deemed unnecessary at the PoC stage.
- **PoC note:** DFS was chosen for simplicity without benchmarking alternatives;
  revisit if traversal characteristics become a problem.

## Follow-ups

- [ ] The `bom_link_index` table (present in the schema, currently unused) is
      **intended to be populated** to persist resolved edges — resolution is
      slow, and persistence is a planned performance improvement.
- [ ] Reconsider `maxDepthReached` reporting if/when depth limits become
      user-visible concerns.
