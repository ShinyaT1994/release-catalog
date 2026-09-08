# ADR-0003: Dependency-Track client behind a port, with an in-memory stub

- **Status:** Accepted
- **Date:** 2026-09-08
- **Deciders:** Project team
- **Tags:** backend, integration, testing, poc

## Context and Problem Statement

The service depends on a Dependency-Track (DT) v5 instance for project and BOM
data. Requiring a running DT for every developer and every test would be slow
and fragile, especially while the app's own behavior is still changing.

## Decision Drivers

- Develop and test the app without a live DT instance.
- Isolate DT's HTTP/JSON details from graph-resolution logic.
- Allow a real DT to be dropped in without code changes.

## Considered Options

- **Direct HTTP calls from usecases** — simplest, but couples logic to DT and
  requires a live DT everywhere.
- **`Client` interface with HTTP + stub implementations** — swap at wiring time.
- **Record/replay HTTP fixtures** — realistic, but heavier to maintain.

## Decision

This is a **confirmed, intentional** mechanism for development and testing. We
use a **`dtclient.Client` interface (driven port)** with two implementations:
`NewHTTPClient` for real DT, and `NewStubClient`, an in-memory implementation.
Selection is controlled by the `RC_DT_STUB_MODE` env var (default `true`), and
the choice is made once in `cmd/server/main.go`. The graph resolver depends only
on the `Client` interface. **For production deployments the default is expected
to be `RC_DT_STUB_MODE=false`** (real DT).

## Consequences

- **Positive:** The app runs and tests pass with no external DT
  (`RC_DT_STUB_MODE=true`). Graph logic is tested against the stub. Swapping to
  real DT is a config change, not a code change.
- **Negative / trade-offs:** The stub can drift from real DT behavior. A
  standalone DT reference stack (`docker-compose.dt.yml`) is provided so
  integration testing is *possible*, but whether/when to run such tests is
  currently **undecided**.
- **PoC note:** Default is stub mode for local/PoC use; production should run
  with the stub disabled. A decision on stub-vs-real integration testing is
  deferred.

## Follow-ups

- [ ] Decide whether to add integration tests against the real DT reference
      stack (currently undecided).
- [ ] If integration testing is adopted, keep the stub's response shapes aligned
      with DT v5.
