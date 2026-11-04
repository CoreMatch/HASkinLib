# HASkinLib Knowledge Base

This directory is the project's own memo book: Architecture Decision
Records (ADRs) that explain **why the code looks the way it does**, kept for
self-hosting and maintenance.

## Positioning

- These notes serve the people who operate and maintain HASkinLib. They
  record decisions, their costs, and the evidence behind them.
- **They do not define externally visible behavior.** External behavior
  (API paths, request/response shapes, error codes) is defined by
  [HA-Contract](https://github.com/CoreMatch/HA-Contract) — in particular
  `docs/api/openapi/haskinlib-business.yaml`, `docs/api/README.md` and
  `docs/api/error-codes.md`.
- If an entry here conflicts with HA-Contract, **fix the entry, not the
  contract**. If the code conflicts with the contract, that is a drift
  bug: record it in an entry and resolve it through HA-Contract's change
  process.

## Rules

- Entries are numbered `NNNN-<slug>.md`, written in English, and follow:
  Status / Date / Context / Decision / Consequences / Evidence.
- Evidence must be checked against the repository before writing: stable
  anchors resolved from the repo root — symbol anchors `file.ext::Symbol`,
  field anchors `file.ext (field: key)`, time-point anchors
  `file.ext:Lx-Ly @ <commit>` for historical snapshots, or bare line
  anchors `file.ext:Lx-Ly` only as a last resort; cross-repo citations are
  marked `[xref]` (absolute paths where possible). Anything that could not
  be verified is marked `not verified`. Gaps are stated honestly, never
  papered over.
- New entries get the next free number; superseded entries change Status
  to `Superseded by NNNN` instead of being deleted.

## Entries

| # | Title | Status | Date | One-line assertion |
|---|-------|--------|------|--------------------|
| [0001](./0001-service-scope-storage-and-preview.md) | Service scope: texture storage and preview | Accepted | 2026-08-02 | HASkinLib is the texture storage/management service (upload, list, preview, pull, delete); it is not a Yggdrasil endpoint provider. |
| [0002](./0002-upload-limits-and-image-pipeline.md) | Upload limits and image pipeline | Accepted | 2026-08-02 | Uploads are capped at 2 MiB, must be PNG in whitelisted canvas sizes, are content-addressed by SHA-256, and get a server-rendered WebP preview. |
| [0003](./0003-authentication-via-hrpauth-http-api.md) | Authentication via HRPAuth HTTP API | Accepted | 2026-10-14 | User authentication is delegated to HRPAuth over HTTP (`/internal/verify-token`, `/services/users/:uuid`) with OAuth2 client credentials; no shared code/library is used. |
| [0004](./0004-dedicated-database-and-migrations.md) | Dedicated database and standalone migrations | Accepted | 2026-10-14 | HASkinLib runs golang-migrate on its own migration chain and refuses to start against the legacy shared `schema_migrations` table (option 1: dedicated database). |
| [0005](./0005-division-of-labor-with-skinproxy-and-yggdrasil.md) | Division of labor with HASkinProxy and Yggdrasil-API | Accepted | 2026-11-04 | HASkinLib stores and manages textures; HRPAuth-Yggdrasil-API serves the Minecraft-facing surface; HASkinProxy translates to CSL. The texture-byte handoff is filesystem-level, not HTTP. |
| [0006](./0006-delete-semantics-force-file-removal.md) | Delete semantics: forced file removal | Accepted | 2026-09-06 | `POST /texture/delete` removes the caller's DB row and force-deletes the on-disk PNG and WebP even when other users reference the same hash. |
| [0007](./0007-contract-status-and-missing-yaml-handling.md) | Contract status and missing-YAML handling | Accepted | 2026-11-04 | HASkinLib had no OpenAPI contract until 2026-11-04; `haskinlib-business.yaml` was then rebuilt from the implementation. While a contract is missing, code + `docs/skinlib-api.md` are the implementation reference, and gaps are fixed in HA-Contract, not here. |
