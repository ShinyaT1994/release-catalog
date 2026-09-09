# ADR-0010: Release branch auto-provisions its first version, ROOT inherited from the forked main version

- **Status:** Accepted
- **Date:** 2026-09-08
- **Deciders:** Project team
- **Tags:** backend, branch, version, workflow, ports-and-adapters, poc

## Context and Problem Statement

A release line is created from a specific main version (`forkedFromVersionId`).
Previously, creating a release line only recorded the fork point on the
`branch_line`; it did **not** materialize any version on the new release branch.
The version usecase already had logic to default a version's `ROOT` binding from
its forked main version, but that logic only fires when a version is created
with `forkedFromVersionId` set — and the only creation path (manual
`POST /branches/{id}/versions` from the UI) never passed it. As a result:

- A newly forked release branch had **zero versions**, so the lineage `fork`
  edge had no target node to connect to.
- Users had to hand-create the first version and then manually re-select the
  same `ROOT` project the main version already used.

We needed the release branch's first version to exist automatically, with its
`ROOT` inherited from the forked main version and status `incomplete`, so users
only fill in `PROFILE`/`SUB` bindings afterwards. Because branch and version are
separate feature packages (ADR-0002), this crosses a feature boundary.

## Decision Drivers

- The initial `ROOT` of a release branch equals the forked main version's
  `ROOT` — this should be automatic, not a manual re-entry.
- Preserve package-by-feature + ports-and-adapters (ADR-0002); no direct import
  of the version package from the branch package.
- Keep the version lifecycle rules (`incomplete` → `finalized`, ADR-0007) and
  the role-tagged project model (ADR-0008) intact.

## Considered Options

- **Option A — Auto-create on release-line creation via a `VersionCreator`
  port.** The branch usecase, after persisting the branch, calls a new
  nil-tolerant `VersionCreator` port; an adapter in `main.go` delegates to the
  version usecase's `Create`, which already defaults `ROOT` and sets
  `incomplete`.
- **Option B — Keep manual creation, fix the frontend.** Leave first-version
  creation manual but make the UI pass `forkedFromVersionId` on the first
  version of a release branch.
- **Option C — Merge branch and version into one feature/service** so the flow
  is a single internal call.

## Decision

We chose **Option A**.

The branch usecase gains an optional `VersionCreator` port (`CreateFirstVersion`)
and calls it when `forkedFromVersionId` is set. The port is **nil-tolerant** so
tests and non-forked flows are unaffected. In `cmd/server/main.go` the version
usecase is constructed before the branch usecase, and a `versionCreatorAdapter`
delegates to `version.UseCase.Create`, reusing the existing ROOT-defaulting and
`incomplete`-status logic. The first version's `versionString` **inherits the
forked main version's string** (fallback `0.1.0` if empty). Option B was rejected
because it puts a data-integrity rule in the client; Option C was rejected
because it collapses a boundary we deliberately keep separate.

The related lineage change — rendering nodes connected with real SVG lines
(parent = solid, fork = dashed) instead of text annotations — is a UI
implementation detail and is **not** itself an architectural decision.

## Consequences

- **Positive:** Forking a release line yields a ready first version whose `ROOT`
  matches the source main version; users can never forget to set `ROOT`. The
  lineage `fork` edge always connects to a real release-branch node. The
  cross-feature call stays behind a port, honoring ADR-0002.
- **Negative / trade-offs:** The branch feature now transitively triggers
  version creation, introducing a construction-ordering constraint in the
  wiring (version usecase before branch usecase) and a coupling that future
  contributors must be aware of. The `versionString` inheritance + `0.1.0`
  fallback is a convention encoded in the branch usecase, not the contract.
  Branch creation and first-version creation are **not** in a single
  transaction, so a failure after the branch is persisted leaves a branch with
  no first version (the call surfaces the error to the client).
- **PoC note:** Revisit before production — (1) whether the initial
  `versionString` should be inherited, blank, or derived from the release-line
  name; (2) whether branch + first-version creation should be atomic
  (transactional) or reconciled; (3) whether auto-provisioning should also apply
  to release lines created without a fork point.

## Follow-ups

- [ ] Decide the production behavior for the initial `versionString` default.
- [ ] Consider wrapping branch + first-version creation in a transaction.
