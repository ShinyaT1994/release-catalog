# ADR-0008: Multiple role-tagged DT projects per version (ROOT / PROFILE / SUB)

- **Status:** Accepted
- **Date:** 2026-09-08
- **Deciders:** Project team
- **Tags:** backend, domain, data-model, dependency-track

## Context and Problem Statement

Originally a branch bound a **single** root DT project. Real cases need more than
one DT project per version: a root system plus **environment-difference**
settings and additional **sub** components — each being its own DT project with
its own UUID and BOM. We also want to know each project's **role**.

## Decision Drivers

- Bind several DT projects to a single version.
- Make each project's purpose explicit via a role.
- Support environment-specific differences (profiles) that are themselves SBOMs.

## Considered Options

- **Single root only** (status quo) — insufficient.
- **A set of DT projects with a free-form label** — flexible but unstructured.
- **A set of DT projects tagged with a fixed role enum** — structured, queryable.

## Decision

Each **version** binds a **set of role-tagged DT projects**. Role is a **fixed
enum**: **`ROOT`**, **`PROFILE`** (environmental difference), **`SUB`**.

- Cardinality per version: **exactly one `ROOT`**, **zero-or-more `PROFILE`**,
  **zero-or-more `SUB`** (a version may have neither PROFILE nor SUB).
- **Main versions have no `PROFILE`** (main has no environment).
- Every role-tagged binding is a **real DT project** (own UUID + BOM), not
  attached metadata.
- Bindings are held **per version** (not once per branch), captured in a
  `version_dt_project` table.
- For a **release** version, `ROOT` **defaults from the forked main version but
  is independently editable/overridable** (each release version owns its own
  set). A release line's first version may start with only a `PROFILE` and no
  `ROOT` until filled in.
- The **one-`ROOT`** rule is enforced in the usecase layer (optionally backed by
  a filtered unique index for defense-in-depth).

## Consequences

- **Positive:** Versions can model root + environment-diff + sub components with
  clear roles. Each project resolves independently for SBOM graphing (see
  ADR-0009 / plan §1.4).
- **Negative / trade-offs:** More complex than a single UUID column; usecase
  logic must enforce cardinality that the schema cannot fully express. Editable
  ROOT means a release version can drift from its forked main root — intended,
  but must be surfaced clearly in the UI.
- **PoC note:** The role enum is intentionally small and closed. Adding roles
  later is a schema/CHECK change; keep it enum, not free-form.

## Follow-ups

- [ ] Add the filtered unique index for one-ROOT if usecase enforcement proves
      insufficient.
