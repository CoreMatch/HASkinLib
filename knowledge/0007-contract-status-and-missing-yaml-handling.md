# 0007 — Contract status and missing-YAML handling

## Status

Accepted

## Date

2026-11-04

## Context

HA-Contract is the source of truth for externally visible behavior
(`docs/api/README.md`: "OpenAPI is the source for paths and response
structures"). For most of its life, HASkinLib had **no** OpenAPI file in
HA-Contract: `docs/api/openapi/haskinlib-business.yaml` did not exist, and
the README contract link went through break/fix cycles (commits `2828f25`,
`8fc0a88`, `02ad589`, `cb98ff01` in this repo document that churn).

## Decision

1. **Current state (verified 2026-11-04):** the gap is closed.
   `haskinlib-business.yaml` was added to HA-Contract in commit
   `4baf9a729ae448d1f953f2309b6898e27c896d16` ("docs: align contracts with
   the current service implementations", 2026-11-04), i.e. it was *rebuilt
   from the current implementation*, and this repo's README links to it.
2. **Handling while a contract file is missing** (policy for the future):
   - The implementation reference is the code plus the in-repo doc
     `docs/skinlib-api.md`, which is maintained alongside the code.
     HA-Contract archives hold the legacy docs.
   - Missing-contract gaps are fixed **in HA-Contract** (add/rebuild the
     OpenAPI file there). This knowledge base never invents external
     behavior to fill the gap, and the README/doc links are pointed at the
     contract only once the file exists.
   - When code and contract disagree, the contract wins for external
     behavior; the code is the thing that changes (or the contract is
     updated through HA-Contract's process if the code is intentional).

## Consequences

- The 2026-11-04 contract was generated from the implementation, so it
  describes current behavior by construction; future drift is detected by
  HA-Contract's check scripts, not by anything in this repo.
- `docs/skinlib-api.md` remains a *reference* (kept in sync with code), not
  the external contract; treat mismatches with HA-Contract as bugs in this
  doc.
- Known residual gap: `GET /profile/textures` is implemented but shelved
  (see [0001](./0001-service-scope-storage-and-preview.md)); its contract
  status should be re-checked whenever the feature is unshelved —
  **pending author confirmation** on whether it belongs in the contract.

## Evidence

- Contract file history: repo HA-Contract (@ `21c7147d`),
  `git log --diff-filter=A -- docs/api/openapi/haskinlib-business.yaml` →
  two additions: `24ba96d1b619f9f561256fec25b34ece7b74f72b` (2026-08-13)
  and `4baf9a729ae448d1f953f2309b6898e27c896d16` (2026-11-04); the first
  version was deleted by `badff2a3ce65b89fdb2fd937c25f33ff7bbf5fed`
  (2026-10-07), so the 2026-11-04 file is a rebuild of a once-deleted
  contract, not the first-ever addition.
- README contract link + "rebuilt 2026-11-04" note: `README.md:8-10`.
- Link break/fix history in this repo: commits
  `2828f25` (merge 'docs/drift-fix'), `8fc0a88` (fix broken OpenAPI link),
  `02ad589` (point README to rebuilt contract), `cb98ff01` (merge
  'docs/contract-align', current HEAD).
- "OpenAPI is the source for paths and response structures": repo
  HA-Contract, `docs/api/README.md` (Unified Constraints section).
- In-repo synced reference doc: `docs/skinlib-api.md` (endpoint list at
  `:3-10`); legacy docs archived under HA-Contract
  `archive/docs/HASkinLib` (referenced from `README.md:14`).
