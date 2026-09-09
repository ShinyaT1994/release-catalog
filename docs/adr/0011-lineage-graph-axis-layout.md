# ADR-0011: Version-lineage graph layout per axis (version blocks vs same-date columns)

- **Status:** Accepted
- **Date:** 2026-09-08
- **Deciders:** Project team
- **Tags:** frontend, visualization, lineage, poc

## Context and Problem Statement

ADR-0009 defined a product-page **version-lineage graph** with a toggle between
`axis=version` and `axis=releaseDate`. The API returns the same nodes and
parent/fork edges for both; it only changes sort order. The first UI placed
**every node in its own column**, so:

- On the **version** axis, a forked release's first version (often the same
  string as the MAIN it forked from) did not sit under that MAIN, and the next
  MAIN version was not delayed until that MAIN's release revisions were drawn.
- On the **releaseDate** axis, two versions with the **same calendar date**
  occupied adjacent columns instead of lining up vertically, and there was no
  date scale.

We needed an explicit layout contract so the two axes answer different
questions without changing the graph payload.

## Decision Drivers

- **Version** view should show *which releases belong to which MAIN version*,
  including hyphenated revisions (`0.0.1` → `0.0.1-1`), and the next MAIN
  version must wait until those revisions are finished.
- **Release date** view should show *when* versions shipped: same day = same
  vertical column, with a readable date scale.
- Rows stay **branch swimlanes** (MAIN first) so each release line remains
  identifiable. Parent edges stay solid; fork edges stay dashed (visual
  encoding only).
- Layout is a **frontend** concern. The lineage API (ADR-0009) stays RC-DB-only
  nodes + parent/fork edges; it does not emit coordinates.

## Considered Options

- **Option A — One unique column per node.** Simple, matches API order. Same
  version family / same date never align; fork lines criss-cross.
- **Option B — Axis-specific column assignment on swimlanes.** Version axis:
  MAIN-owned blocks with forked releases shifted right. Date axis: one column
  per calendar date plus a tick scale.
- **Option C — Proportional time scale on the date axis** (pixel distance ∝
  elapsed time). Same-day nodes would still share an x, but sparse dates waste
  space and need extra tick policy (weeks/months).

## Decision

We chose **Option B**. Implementation lives in `LineageGraphView`
(`frontend/src/components/LineageGraphView.tsx`).

**Shared:** one row per branch line. MAIN lanes above RELEASE lanes.

**Version axis**

1. Walk the MAIN parent chain. Each MAIN version owns a **block**.
2. A release line belongs to the MAIN version reached by walking fork/parent
   edges from its first version.
3. That line's versions start **one column to the right** of the MAIN node;
   later revisions continue right (`col + 1 + i`).
4. Block width is `1 + longest related release chain`. The **next MAIN** starts
   after that width — so MAIN `0.0.2` is not drawn until every `0.0.1` /
   `0.0.1-*` revision in the previous block has a column.
5. A version-axis **fork** edge drops from the **bottom** of the source (MAIN)
   and turns right into the **left** of the first release node. Several forks
   from the same MAIN share that vertical spine.

**Release date axis**

1. Columns are **unique calendar dates** (`YYYY-MM-DD`); undated versions share
   a trailing `(no date)` column.
2. Nodes with the same date share an x-position (vertical alignment across
   lanes). Two versions on the same lane and date stack vertically in that
   cell.
3. A **date scale** (axis line, ticks, labels) is drawn under the plot. Faint
   vertical guides mark each date column.
4. Same-column forks (typical when MAIN and a release share a date) stay
   vertical; they detour around any node sitting between source and target.

Option A was rejected because it hid both lineage grouping and same-day
coincidence. Option C was rejected for the PoC: discrete date columns plus
ticks are enough to read "same day" without a continuous time mapping.

This ADR **does not supersede** ADR-0009 (two graphs, axis query param) or
ADR-0010 (auto-provisioned first release version). It specifies how the
lineage graph is **drawn**.

## Consequences

- **Positive:** Version view matches the intended MAIN-then-revision grouping;
  fork geometry (bottom → left) is unambiguous. Date view makes same-day
  releases obvious and gives a labeled horizontal scale. Backend lineage JSON
  is unchanged.
- **Negative / trade-offs:** Column logic is entirely in the frontend, so a
  future non-React client must reimplement it. Version-axis columns are not
  a parsed semver order (ADR-0007: version strings stay free-form); grouping
  is by fork-to-MAIN topology, not by parsing `0.0.1-1`.
- **PoC note:** Revisit if we need a proportional calendar scale, sticky lane
  labels on wide graphs, or server-side layout for other clients.

## Follow-ups

- [ ] Decide whether a production date axis should be proportional to elapsed
      time (Option C) rather than one column per distinct date.
- [ ] If another client consumes the lineage API, share this layout contract
      or move coordinates behind an endpoint.
