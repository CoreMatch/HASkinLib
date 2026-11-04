# 0002 — Upload limits and image pipeline

## Status

Accepted

## Date

2026-08-02 (commit `ba188415` limits/validation; commit `c828b2e3` preview pipeline)

## Context

User-uploaded images are stored on disk, served publicly, and rendered into
gallery previews. The service needed explicit size/format guardrails and a
cheap, uniform preview format for the web gallery.

## Decision

1. **Size limits**: single file ≤ 2 MiB (`MaxUploadBytes = 2 << 20`),
   whole request ≤ 2 MiB + 256 KiB (`MaxRequestBytes`), enforced with
   `http.MaxBytesReader` plus an explicit `fileHeader.Size` check returning
   413. Both values are configurable; defaults apply when unset.
2. **Rate limiting**: uploads are limited per access token *and* per client
   IP (default 5 per 60 s window), in-memory sliding window.
3. **Format validation**: file must decode as PNG. Skins must be 64x32 or
   64x64; capes must be 64x32 or 22x17. A 22x17 cape is normalized
   (padded) to 64x32 before storage. Skin model is `default` or `slim`;
   capes have no model.
4. **Content addressing**: stored bytes are the re-encoded PNG; the storage
   key and public identifier is the SHA-256 hex of those bytes. Raw files
   are stored flat as `<storage_dir>/<hash>`; re-uploading an existing
   (uid, hash) pair upserts metadata instead of creating a new row.
5. **Preview pipeline**: at upload time the server renders a flat 2D preview
   (skin: head/torso/arms/legs composite honoring slim arm width; cape: the
   10x16 outer face), nearest-neighbor scaled 8x, encoded as **WebP**
   (best compression, via `nativewebp`) and stored as
   `<preview_dir>/<hash>_<type>.webp`. Previews are served publicly with
   `Cache-Control: public, max-age=86400`.

## Consequences

- Costs: the whole file is read into memory and re-encoded; the 2 MiB cap
  keeps this acceptable. Raising the cap has memory implications.
- Costs: preview generation is synchronous in the upload path (CPU work per
  upload); the rate limiter bounds this.
- Costs: rate-limiter state is in-memory and per-process; it does not
  survive restarts and is not shared across replicas.
- Constraint: stored hash is over the *re-encoded* PNG, not the original
  upload bytes — the same source image can hash differently across Go PNG
  encoder versions. Anything verifying hashes externally must pull the
  stored bytes, not re-hash the user's file.
- Constraint: deleting a texture force-removes the shared files; see
  [0006](./0006-delete-semantics-force-file-removal.md).

## Evidence

- Defaults (2 MiB, +256 KiB, 5/min, 60 s): `config/config.go:76-83`;
  controller constants `controllers/upload_controller.go:21-26`.
- `MaxBytesReader` + 413 handling: `controllers/upload_controller.go:57,76-89`.
- Token+IP rate limiting: `controllers/upload_controller.go:67-70`;
  limiter implementation `services/upload_rate_limiter.go:23-47`.
- PNG-only + size whitelist: `services/upload_service.go:32,326-332`;
  error mapping `controllers/upload_controller.go:111-124`.
- Cape 22x17 → 64x32 normalization: `services/upload_service.go:309-313,334-339`.
- SHA-256 content addressing: `services/upload_service.go:378-381`;
  flat storage path `services/upload_service.go:393`.
- Upsert on (uid, hash): `services/upload_service.go:129-160` (skin),
  `:162-192` (cape).
- WebP preview rendering: `services/upload_service.go:349-357` (encoder),
  `:359-376` (pipeline), `:429-476` (skin/cape composition),
  `:40` (`previewScale = 8`), `:425-427` (preview file naming).
- Introduced by commits `ba188415138a0d4e87756ff690fc4dac4b12957b`
  ("feat: add upload rate limiting and texture validation", 2026-08-02) and
  `c828b2e321ee5331d094128890ee858148b91d16`
  ("feat(upload): add texture preview image support", 2026-08-02).
