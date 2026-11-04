# 0001 — Service scope: texture storage and preview

## Status

Accepted

## Date

2026-08-02 (route surface settled with the pull endpoint, `d78272b6`)

## Context

The HA ecosystem needs a place where registered users can upload Minecraft
skins/capes, browse a public gallery, and where other services can fetch the
raw texture bytes by content hash. The Minecraft-facing protocol surface
(Yggdrasil endpoints) and the CustomSkinLoader translation live in other
services (see [0005](./0005-division-of-labor-with-skinproxy-and-yggdrasil.md)).

## Decision

HASkinLib is a standalone HTTP service (Gin, default port `:2701`) whose
responsibility boundary is **texture storage, metadata management, and
preview rendering**:

- `POST /texture/upload` — authenticated upload + validation + preview
  generation.
- `GET /texture/listpreview` — public paginated metadata gallery
  (16 items/page).
- `GET /texture/preview/:preview_file` — public WebP preview file serving.
- `GET /texture/pull/:hash` — public raw PNG serving by SHA-256 hash.
- `POST /texture/delete` — authenticated deletion of the caller's own
  texture (see [0006](./0006-delete-semantics-force-file-removal.md)).
- `GET /profile/textures` — authenticated per-user listing. Implemented,
  but marked "temporarily shelved" in the in-repo API doc; the route is
  still registered.

It owns its own MySQL tables (`hasl_texture_list_skin`,
`hasl_texture_list_cape`) and its own on-disk storage directories. It does
**not** implement any Yggdrasil endpoint, does not issue or validate game
sessions, and does not talk to Minecraft clients directly.

## Consequences

- Costs: gallery listing uses `FIND_IN_SET` tag filtering on a varchar
  column (no tag index), fine at current scale but a known limit.
- Costs: `GET /profile/textures` stays compiled and routed while shelved;
  operators should treat it as not part of the product surface.
- Constraint: anything Minecraft-protocol-facing belongs to
  HRPAuth-Yggdrasil-API / HASkinProxy, not here; adding Yggdrasil routes
  here would violate this boundary.
- The external path/response definitions for these routes live in
  HA-Contract (`docs/api/openapi/haskinlib-business.yaml`); this entry only
  records the internal scope decision.

## Evidence

- Routes: `main.go:60-66` (commit `cb98ff01`, HEAD of `main`).
- Default port `:2701`: `config/config.go:61`; also `build/config.yaml:9`.
- Page size 16: `controllers/list_preview_controller.go:13`.
- Tag filtering via `FIND_IN_SET`: `controllers/list_preview_controller.go:91,130`.
- Pull endpoint introduced: commit `d78272b6cd0f82d13f3f767b84990553bf2821e8`
  ("feat(api): add texture pull endpoint by sha256 hash", 2026-08-02).
- `/profile/textures` shelved status: `docs/skinlib-api.md:10` and
  `docs/skinlib-api.md:443-444` ("状态：暂时搁置"); route still registered at
  `main.go:66`. Shelving is a documented plan statement; the reason for
  shelving is not recorded in the repo — **pending author confirmation**.
- Positioning as HA-Contract skin-library component: `README.md:3`.
