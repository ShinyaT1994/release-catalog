---
name: adr-recording
description: Detect architectural decisions during coding and prompt the user to record them as an Architecture Decision Record (ADR) in docs/adr/. Use whenever the work involves choosing or replacing a datastore/framework/protocol, introducing an architectural pattern or layering, defining an API/contract boundary, or deliberately deferring/stubbing something (auth, an integration). Conduct the ADR intake questions in English.
---

# ADR Recording Skill

Your job with this skill is to notice when an **architecturally significant
decision** is being made during coding ("vibe coding") and to help the user
capture it as an ADR under `docs/adr/`. You **prompt and interview**; you do not
silently write ADRs on your own initiative without the user's go-ahead.

Conduct the interview in **English**, and accept the user's answers in English.

## 1. Detect a decision worth recording

Watch for these signals in the conversation or the diff. Any one is enough to
raise the question:

- **Datastore / infra choice:** adding or swapping a database, cache, queue,
  ORM, or driver (e.g. SQLite → PostgreSQL, adding Redis).
- **Framework / library with lock-in:** adopting a web framework, DI approach,
  auth library, or anything that will be hard to remove later.
- **Protocol / contract boundary:** defining or changing the API shape, choosing
  REST vs. gRPC, picking a serialization format, declaring a "source of truth".
- **Architectural pattern:** introducing layering, ports-and-adapters,
  package-by-feature, event-driven flow, caching strategy, etc.
- **Deliberate deferral / stub / placeholder:** stubbing an integration, a
  "TODO: implement auth" placeholder, an in-memory fake standing in for a real
  service, hard-coded values meant to be temporary.
- **Cross-cutting trade-off:** a choice that touches multiple packages or that
  someone will later ask "why is it done this way?" about.

Do **not** raise an ADR for routine work: renames, formatting, bug fixes,
obvious local implementation details, or reversible one-liners.

If unsure whether something qualifies, ask one short question:
> "This looks like an architectural decision (choosing X over Y). Want to record
> it as an ADR so the reasoning is captured?"

## 2. Prompt to record

When you detect a qualifying decision, pause and say something like:

> "Heads up — we just made an architectural decision: **<one-line summary>**.
> I recommend recording it as an ADR in `docs/adr/`. Shall I capture it? I'll
> ask a few quick questions (about 2 minutes)."

If the user declines, respect that and continue. Optionally offer to add a
`TODO(adr)` note near the code so it can be recorded later.

## 3. Interview — required fields (ask in English)

Ask these one at a time (or in a short batch), keeping it lightweight. Infer
sensible defaults from the code and the conversation, and offer them so the user
can just confirm:

1. **Title** — a short decision title. (Suggest one.)
2. **Context / problem** — what situation forces this decision?
3. **Options considered** — what alternatives were on the table? (List the ones
   already mentioned; ask if any are missing.)
4. **Decision** — which option, and the one-line reason.
5. **Consequences** — the main positive and the main trade-off / risk.
6. **Status** — `Proposed` or `Accepted`? (Default `Accepted` if the code is
   already written.)
7. **PoC note** *(this project is a PoC)* — is this provisional? If so, what
   would trigger a revisit before production?

Keep momentum: propose answers derived from context and let the user correct
them rather than asking open-ended questions cold.

## 4. Write the ADR

1. Determine the next number: read `docs/adr/`, find the highest existing
   `NNNN`, and use `NNNN+1` (zero-padded to 4 digits). `0000-template.md` is the
   template, not a decision.
2. Copy the structure of `docs/adr/0000-template.md` and fill in the answers.
3. Save as `docs/adr/NNNN-kebab-case-title.md`.
4. Add a row to the **Index** table in `docs/adr/README.md`.
5. Show the user the created file and the index update.

## 5. Conventions to honor

- One decision per file; filename `NNNN-kebab-case-title.md`.
- **Never rewrite an accepted decision.** If a past decision changes, create a
  **new** ADR and set the old one's status to `Superseded by ADR-NNNN` with a
  link, instead of editing the original.
- Status values: `Proposed`, `Accepted`, `Deprecated`, `Superseded by ADR-NNNN`.
- Because this is a PoC, always fill in the **PoC note** when the decision is
  provisional — it doubles as a "revisit before production" checklist.

## Example trigger → prompt

- User: "let's just use an in-memory map for the DT client for now"
  → "That's a deliberate stub decision (in-memory DT client instead of real
  Dependency-Track). Want me to record it as an ADR? A few quick questions and
  I'll write it to `docs/adr/`."
