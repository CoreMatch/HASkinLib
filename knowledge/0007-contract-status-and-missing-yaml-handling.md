# 0007 — Contract status and missing-YAML handling

## Status

Accepted

## Context

HA-Contract is the source of truth for externally visible behavior
(`/home/lnb/Desktop/HA/HA-Contract/docs/api/README.md`: "OpenAPI is the
source for paths and response structures"). HASkinLib's external contract
is
`/home/lnb/Desktop/HA/HA-Contract/docs/api/openapi/haskinlib-business.yaml`
in HA-Contract.

## Decision

1. **Contract status:** `haskinlib-business.yaml` exists in HA-Contract and
   was *rebuilt from the current implementation*; this repo's README links
   to it.
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

- The contract was generated from the implementation, so it describes
  current behavior by construction; future drift is detected by
  HA-Contract's check scripts, not by anything in this repo.
- `docs/skinlib-api.md` remains a *reference* (kept in sync with code), not
  the external contract; treat mismatches with HA-Contract as bugs in this
  doc.
- Known residual gap: `GET /profile/textures` is implemented but shelved
  (see [0001](./0001-service-scope-storage-and-preview.md)); its contract
  status should be re-checked whenever the feature is unshelved —
  **pending author confirmation** on whether it belongs in the contract.

## Evidence

- README contract link + "rebuilt from the current implementation" note:
  `README.md:L8-L10`.
- "OpenAPI is the source for paths and response structures": repo
  HA-Contract, `/home/lnb/Desktop/HA/HA-Contract/docs/api/README.md`
  (Unified Constraints section).
- In-repo synced reference doc: `docs/skinlib-api.md` (endpoint list at
  `docs/skinlib-api.md:L3-L10`); legacy docs archived under HA-Contract
  `/home/lnb/Desktop/HA/HA-Contract/archive/docs/HASkinLib` (referenced
  from `README.md:L14`).
