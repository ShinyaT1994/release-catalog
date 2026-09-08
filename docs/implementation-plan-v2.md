# Release Catalog — Implementation Plan v2 (Git-like Version Model)

Status: **Planning** · Date: 2026-09-08 · Supersedes the single-root /
snapshot model from `implementation-plan.md`.

> No backward compatibility required — the current single-root schema/API may be
> broken cleanly (PoC, no production data).

This plan covers three requested features:

1. **Multiple role-tagged DT projects per version** (`ROOT` / `PROFILE` / `SUB`).
2. **Git-like version history per branch** (versions replace snapshots +
   current-state).
3. **Release-history visualization** — a version-lineage timeline/graph
   (product page) distinct from the existing BOM-Link SBOM graph (branch/version
   detail).

Product **risk data** (vulnerabilities / known issues) is **out of scope** for
this plan; the model is left extensible so it can be layered on later.

---

## 1. Domain model

### 1.1 Hierarchy

```
Product
 ├── Main branch            (exactly one, type = MAIN)
 │     └── Version[]         (linear line, ordered by release datetime)
 └── Release branch[]        (N; type = RELEASE; each = one release profile)
       └── Version[]         (line per release branch, ordered by release datetime)
```

- **Main `1:N` Release branch.** A release branch is **forked from a specific
  main version**. That main version therefore relates to N release branches.
- **Release branch `1:N` Release version.**
- A branch **owns an ordered list of versions**. There is **no mutable
  "current state"** — "current" = the latest version on the line.

### 1.2 Version

A **Version** is the git-like commit on a branch line. It replaces the old
`snapshot` and the old `branch_current_state`.

| Field | Notes |
|-------|-------|
| `id` | UUID |
| `branchLineId` | owning branch |
| `versionString` | free-form (`1.0`, `1.0-1`); system does **not** parse/auto-increment |
| `status` | `incomplete` → `finalized` (explicit; drives timeline visuals) |
| `parentVersionId` | previous version on the same line (linear within a line) |
| `forkedFromVersionId` | for a release branch's **first** version: the main version it forked from (nullable otherwise) |
| release metadata | `location`, `releaseDate`, `customer`, … (release branches) |
| `createdAt` / `updatedAt` | soft-immutability audit |

- **Creation is partial ("preset"):** the minimum required field is
  **`versionString`** only. Everything else (DT projects, release metadata) is
  filled in later.
- **Soft immutability:** every field is editable after creation (including the
  version string and DT-project bindings); edits bump `updatedAt`. No
  supersede/freeze mechanic.
- **Ordering:** version lines are ordered by **release datetime** (not by
  parsing the version string).

### 1.3 Role-tagged DT projects (per version)

Each version binds a **set** of DT projects, each tagged with a **role**:

- Roles (fixed enum): **`ROOT`**, **`PROFILE`** (environmental difference),
  **`SUB`**.
- Cardinality per version: **exactly one `ROOT`**, **zero-or-more `PROFILE`**,
  **zero-or-more `SUB`** (sometimes no PROFILE and no SUB at all).
- **Main versions have no `PROFILE`** (main has no environment).
- Every role-tagged project is a **real DT project** (own UUID + BOM).
- **ROOT on a release version defaults from the forked main version but is
  independently editable/overridable** (Q-A = Option 2). The first release
  version may start with only a `PROFILE` and no `ROOT` until filled in.

Each binding row: `{ versionId, role, dtProjectUuid, bomSerial, bomVersion,
bomSha256, sourceRevision }`.

### 1.4 Two graphs (different purposes, different pages)

| | Version-lineage graph | BOM-Link SBOM graph (existing) |
|---|---|---|
| Purpose | Overview of the **release timeline** | Overview of the **SBOM tree** |
| Nodes | **Versions** | DT projects (BOM-Link resolution) |
| Edges | parent / fork lineage | `urn:cdx:` BOM-Link references |
| Data source | **RC DB only** (no DT calls) | Dependency-Track |
| Failure mode | cannot 502 | `502` when DT unreachable |
| Scope | Main = 1 line; each release branch = its own line forked from a main node | **per role-tagged project** (separate graph each) |
| UI location | **Product main page** | **Branch/version detail page**, per version |

---

## 2. Data model / schema (SQLite, breaking change)

Replace the `snapshot`, `branch_current_state`, and `bom_link_index`-centric
model. `product` is unchanged. `branch_line` keeps its fork pointer but retargets
it to a **version**.

```sql
-- product: unchanged

-- branch_line: fork now points to a main VERSION, not a snapshot
CREATE TABLE branch_line (
    id                    TEXT PRIMARY KEY,
    product_id            TEXT NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    type                  TEXT NOT NULL CHECK(type IN ('MAIN','RELEASE')),
    name                  TEXT NOT NULL,
    display_name          TEXT NOT NULL DEFAULT '',
    source_branch_line_id TEXT REFERENCES branch_line(id),
    forked_from_version_id TEXT REFERENCES version(id),  -- main version the release forked from
    status                TEXT NOT NULL DEFAULT 'active'
        CHECK(status IN ('active','maintenance','security_only','end_of_support','closed')),
    created_at            TEXT NOT NULL,
    updated_at            TEXT NOT NULL,
    closed_at             TEXT,
    UNIQUE(product_id, name)
);

-- version: replaces snapshot + branch_current_state
CREATE TABLE version (
    id                  TEXT PRIMARY KEY,
    branch_line_id      TEXT NOT NULL REFERENCES branch_line(id) ON DELETE CASCADE,
    version_string      TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'incomplete'
        CHECK(status IN ('incomplete','finalized')),
    parent_version_id   TEXT REFERENCES version(id),        -- previous version on this line
    forked_from_version_id TEXT REFERENCES version(id),     -- set on a release line's first version
    -- release metadata (release branches; nullable on main)
    location            TEXT,
    customer            TEXT,
    release_date        TEXT,
    created_at          TEXT NOT NULL,
    updated_at          TEXT NOT NULL,
    UNIQUE(branch_line_id, version_string)
);
CREATE INDEX idx_version_branch ON version(branch_line_id);
CREATE INDEX idx_version_parent ON version(parent_version_id);
CREATE INDEX idx_version_fork   ON version(forked_from_version_id);

-- role-tagged DT project bindings (many per version)
CREATE TABLE version_dt_project (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    version_id          TEXT NOT NULL REFERENCES version(id) ON DELETE CASCADE,
    role                TEXT NOT NULL CHECK(role IN ('ROOT','PROFILE','SUB')),
    dt_project_uuid     TEXT,
    bom_serial_number   TEXT,
    bom_version         INTEGER,
    bom_sha256          TEXT,
    source_revision     TEXT,
    label               TEXT,               -- optional human label (e.g. which profile)
    created_at          TEXT NOT NULL,
    updated_at          TEXT NOT NULL
);
CREATE INDEX idx_vdp_version ON version_dt_project(version_id);
-- exactly-one-ROOT is enforced in the usecase layer (SQLite lacks partial-unique on filtered rows easily);
-- optionally: CREATE UNIQUE INDEX idx_vdp_one_root ON version_dt_project(version_id) WHERE role='ROOT';
```

Notes:
- **Exactly-one-`ROOT`** per version is enforced in the usecase (a filtered
  unique index is possible in SQLite and can be added as defense-in-depth).
- The existing `graph` BOM-Link resolution stays; it now takes a
  `dtProjectUuid` from a **role-tagged binding** rather than a branch's single
  root.

---

## 3. API surface (draft, `/api/v1`)

**Removed:** `PUT/GET /branches/{id}/current`, `POST /branches/{id}/snapshots`,
`POST/GET /branches/{id}/releases`, `GET /releases/{id}`,
`GET /branches/{id}/current/graph`, `GET /releases/{id}/graph`.

**Branches / fork**
| Method & Path | Description |
|---|---|
| `POST /products/{productId}/release-lines` | Create a release branch, forked from a main **version** (`forkedFromVersionId`) |

**Versions (git-like)**
| Method & Path | Description |
|---|---|
| `POST /branches/{branchId}/versions` | Create (preset) a version — body: `{ versionString }` only required |
| `GET /branches/{branchId}/versions` | List versions on a branch (ordered by release datetime) |
| `GET /versions/{versionId}` | Get a version (incl. role-tagged projects + metadata) |
| `PATCH /versions/{versionId}` | Edit any field (soft immutability); set `status`, metadata, etc. |
| `DELETE /versions/{versionId}` | (optional) remove a preset version |

**Role-tagged DT projects**
| Method & Path | Description |
|---|---|
| `PUT /versions/{versionId}/projects/{role}` | Set/replace the `ROOT` (single) |
| `POST /versions/{versionId}/projects` | Add a `PROFILE` / `SUB` binding |
| `DELETE /versions/{versionId}/projects/{bindingId}` | Remove a binding |

**Graphs**
| Method & Path | Description |
|---|---|
| `GET /products/{productId}/lineage-graph` | **Version-lineage graph** (RC DB only). Query: `axis=version|releaseDate` |
| `GET /versions/{versionId}/projects/{role}/graph` | **BOM-Link SBOM graph** for one role-tagged project (`maxDepth`, `maxNodes`; may `502`) |

**Timeline**
| Method & Path | Description |
|---|---|
| `GET /products/{productId}/timeline` | Product-scope timeline (all branches) |
| `GET /branches/{branchId}/timeline` | Branch-scope timeline |

`axis` toggles the horizontal ordering between **version** and **release
datetime**.

---

## 4. Backend package changes (Clean Architecture, package-by-feature — ADR-0002)

- **`version/`** (new) — replaces `release/`. Domain (`Version`, `Role`,
  `VersionDTProject`, `Status`), usecase (create/preset, edit, list,
  set/add/remove role project, enforce one-ROOT, default ROOT from fork),
  repository (`version`, `version_dt_project`), handler, ports.
- **`branch/`** — drop `CurrentState*` entirely; retarget fork to
  `forkedFromVersionId`. `ListByProductID` unchanged.
- **`graph/`** — split into two usecases:
  - `sbom` (existing resolver) now keyed by a role-tagged project UUID.
  - `lineage` (new) — builds nodes=versions / edges=parent+fork purely from the
    `version` table; no DT client dependency.
- **`dtclient/`** — unchanged (still behind the port + stub, ADR-0003).
- **`cmd/server/main.go`** — rewire adapters for `version` and the two graph
  usecases; remove current-state adapters.

---

## 5. Frontend changes (React, ADR-0004 code-first types)

- **Product main page:** version-lineage graph (main line + forked release
  lines, nodes = versions) with a **version / release-datetime axis toggle**;
  product-scope timeline.
- **Branch detail page:** per-version panel; each role-tagged project
  (`ROOT`/`PROFILE`/`SUB`) links to its **own** BOM-Link SBOM graph.
- **Version editor:** preset (version string) → fill DT projects + release
  metadata → mark `finalized`. Status badge (`incomplete`/`finalized`).

---

## 6. Phased build plan

1. **Schema + domain** — new `version` / `version_dt_project`, drop
   snapshot/current-state; `version` feature package (domain + repository).
2. **Version usecases/API** — preset create, partial edit (soft immutability),
   list ordered by release datetime, one-ROOT enforcement, ROOT default-from-fork.
3. **Role-tagged projects API** — set ROOT / add PROFILE|SUB / remove.
4. **SBOM graph retarget** — resolve per role-tagged project; per-version detail
   endpoint.
5. **Version-lineage graph + timeline** — RC-DB-only builder; product & branch
   scope; axis toggle.
6. **Frontend** — product page (lineage + timeline), branch/version detail
   (per-project SBOM graphs), version editor.
7. **Tests** — usecase tests for one-ROOT, fork/default-ROOT, ordering, partial
   create/edit; lineage builder tests; SBOM resolver tests retargeted.

---

## 7. Explicitly out of scope (this plan)

- Product risk: vulnerabilities / known issues per version, and risk
  aggregation (ROOT+PROFILE+SUB roll-up). Model kept extensible for a later
  phase.
- Authentication (still deferred — ADR-0006).

---

## 8. Related ADRs

- **ADR-0007** — Git-like version model replacing snapshots + current-state.
- **ADR-0008** — Multiple role-tagged DT projects per version (ROOT/PROFILE/SUB).
- **ADR-0009** — Version-lineage graph (RC-DB-only) separate from the BOM-Link
  SBOM graph.
