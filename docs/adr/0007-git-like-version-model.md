# ADR-0007: Git-like version model replacing snapshots and current-state

- **Status:** Accepted
- **Date:** 2026-09-08
- **Deciders:** Project team
- **Tags:** backend, domain, data-model

## Context and Problem Statement

The original model gave each branch a single mutable `branch_current_state` and
a flat list of immutable `snapshot`s. Requirements have grown: we need an
explicit, ordered **version history** per branch, release branches that **fork
from a specific main version**, and versions that can be created incrementally
and corrected. The team framed the desired mental model as **git history**.

## Decision Drivers

- Represent version history as a lineage (lines + fork points), like git.
- Allow a version to be created early ("preset") and completed later.
- Allow correction of human mistakes without a heavy supersede mechanism.
- No backward compatibility required (PoC, no production data).

## Considered Options

- **Keep snapshot + current-state**, add ordering on top.
- **Git-like `version` entity** that owns history, replacing both snapshot and
  current-state.
- **Event-sourced history** (append-only events, projected state).

## Decision

We adopt a **git-like `version` entity** and **remove both `snapshot` and
`branch_current_state`**. A branch owns an **ordered list of versions**; there
is no separate mutable working area — "current" is the latest version.

- **Lines are linear within a branch**; branching happens only at **release
  fork points**. A release branch's first version records
  `forkedFromVersionId` pointing at the main version it forked from.
- **Ordering is by release datetime**, not by parsing the free-form version
  string (e.g. `1.0`, `1.0-1`); the system does not interpret version strings.
- **Partial creation:** the only required field to create a version is the
  **version string**; DT projects and release metadata are filled in later.
- **Status** is explicit: `incomplete` → `finalized`, so the timeline can
  distinguish not-yet-finalized versions.
- **Soft immutability:** every field remains editable after creation; edits bump
  `updatedAt`. We deliberately do **not** freeze fields or use
  correction-by-supersede.

## Consequences

- **Positive:** History is first-class and matches the team's git mental model.
  Incremental authoring is natural; corrections are simple.
- **Negative / trade-offs:** "Immutable" is only a convention (soft), so there
  is no cryptographic/audit guarantee that a finalized version never changed
  beyond the `updatedAt` timestamp. Removing snapshots/current-state is a
  breaking schema change.
- **PoC note:** Soft immutability is acceptable for the PoC. If audited
  immutability becomes a requirement, revisit (e.g. append-only revisions or a
  change log) in a superseding ADR.

## Follow-ups

- [ ] If provenance/audit hardening is needed, design an immutability guarantee.
