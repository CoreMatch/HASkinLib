# 0006 — Delete semantics: forced file removal

## Status

Accepted

## Context

Textures are content-addressed: multiple users can hold rows pointing at the
same `<hash>` file (e.g. re-uploads of the same PNG). Deletion had to decide
whether disk cleanup is reference-counted or forced.

## Decision

`POST /texture/delete` (authenticated, caller's own rows only):

1. Delete the `(uid, hash)` row from `texture_list_skin` or
   `texture_list_cape` for the token's user. The `uid` comes from the
   verified token; there is no way to target another user's row.
2. **Force-remove** the on-disk raw PNG (`<storage_dir>/<hash>`) and the
   WebP preview (`<preview_dir>/<hash>_<type>.webp`) regardless of whether
   other users' rows still reference the same hash. File-removal failures
   are logged as warnings and do not fail the request.

## Consequences

- Known sharp edge: after user A deletes a texture that user B also
  uploaded, B's database row survives but `GET /texture/pull/:hash` and
  `GET /texture/preview/...` for that hash return 404 until B (or anyone)
  re-uploads the same bytes. This is intentional per the code comments and
  documented in the in-repo API doc.
- Cost: no reference counting / garbage collection; storage reclamation is
  aggressive at the price of the edge case above.
- Constraint: any future reference-counted cleanup must also consider that
  HRPAuth-Yggdrasil-API serves the same files from its own storage view
  (see [0005](./0005-division-of-labor-with-skinproxy-and-yggdrasil.md)).
- The external error contract for this route lives in HA-Contract
  (`haskinlib-business.yaml` / error-codes); this entry records the
  internal semantics decision only.

## Evidence

- Force-delete comment and implementation:
  `services/delete_texture_service.go::DeleteTexture` ("强制清理磁盘上的原始 PNG 文件和
  WebP 预览图,即使其他用户仍引用相同 hash") and
  `services/delete_texture_service.go::removeTextureFiles`
  (comment "强制删除策略").
- Row deletion scoped to caller uid: `services/delete_texture_service.go::deleteSkinTexture`,
  `services/delete_texture_service.go::deleteCapeTexture`.
- Controller (auth + param validation, no uid input accepted):
  `controllers/delete_texture_controller.go::DeleteTexture`.
- Documented side effect incl. cross-user impact: `docs/skinlib-api.md:L528-L532 fp=3e2f4748 ~ **副作用**`.
