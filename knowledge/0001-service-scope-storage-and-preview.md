# 0001 — Service scope: texture storage and preview

## Status

Accepted

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
- `GET /profile/textures` — authenticated per-user listing. Part of the
  contract surface (author decision 2026-11-12); registered in
  HA-Contract `haskinlib-business.yaml` 1.1.0.

It owns its own MySQL tables (`hasl_texture_list_skin`,
`hasl_texture_list_cape`) and its own on-disk storage directories. It does
**not** implement any Yggdrasil endpoint, does not issue or validate game
sessions, and does not talk to Minecraft clients directly.

## Consequences

- Costs: gallery listing uses `FIND_IN_SET` tag filtering on a varchar
  column (no tag index), fine at current scale but a known limit.
- `GET /profile/textures` is contract surface (author decision 2026-11-12);
  it is covered by HA-Contract's route-drift and field-level checks like the
  other routes, so external changes to it are drift-gated.
- Constraint: anything Minecraft-protocol-facing belongs to
  HRPAuth-Yggdrasil-API / HASkinProxy, not here; adding Yggdrasil routes
  here would violate this boundary.
- The external path/response definitions for these routes live in
  HA-Contract (`/home/lnb/Desktop/HA/HA-Contract/docs/api/openapi/haskinlib-business.yaml`);
  this entry only records the internal scope decision.

## Evidence

- Routes: `main.go::main`.
- Default port `:2701`: `config/config.go (field: port)`; also
  `/home/lnb/Desktop/HA/hrpauth/HASkinLib/build/config.yaml (field: port)`
  (gitignored local runtime config).
- Page size 16: `controllers/list_preview_controller.go::defaultListPreviewPageSize`.
- Tag filtering via `FIND_IN_SET`: `controllers/list_preview_controller.go::querySkinPreviewItems`,
  `controllers/list_preview_controller.go::queryCapePreviewItems`.
- `/profile/textures` contract status: `docs/skinlib-api.md:L10-L10 fp=0885fe64` (listed
  without a shelved marker) and `docs/skinlib-api.md:L443-L446 fp=958fe983` (status note
  recording the 2026-11-12 decision); route registered at
  `main.go::GetMyTextures`. The earlier "temporarily shelved" marker and the
  pending-author-confirmation gap were removed when the author confirmed the
  endpoint belongs in the contract (HA-Contract `haskinlib-business.yaml`
  1.1.0).
- Positioning as HA-Contract skin-library component: `README.md:L3-L3 fp=729b6ac1 ~ 本项目是 HA-Contract 的皮肤库服务组件。`.
