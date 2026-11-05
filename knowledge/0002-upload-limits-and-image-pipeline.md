# 0002 — Upload limits and image pipeline

## Status

Accepted

## Date

2026-08-02

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

- Defaults (2 MiB, +256 KiB, 5/min, 60 s): `config/config.go (field: max_upload_bytes)`,
  `config/config.go (field: max_request_bytes)`,
  `config/config.go (field: rate_limit_per_minute)`,
  `config/config.go (field: rate_limit_window_seconds)`;
  controller constants `controllers/upload_controller.go::defaultMaxTextureUploadBytes`,
  `controllers/upload_controller.go::defaultMaxTextureRequestBytes`,
  `controllers/upload_controller.go::defaultUploadRateLimitPerMinute`,
  `controllers/upload_controller.go::defaultUploadRateLimitWindowSeconds`.
- `MaxBytesReader` + 413 handling: `controllers/upload_controller.go::UploadTexture`
  (MaxBytesReader wrap + fileHeader.Size check),
  `controllers/upload_controller.go::isBodyTooLargeError`.
- Token+IP rate limiting: `controllers/upload_controller.go::UploadTexture`
  (token/ip `Allow` gate); limiter implementation
  `services/upload_rate_limiter.go::UploadRateLimiter`,
  `services/upload_rate_limiter.go::Allow`.
- PNG-only + size whitelist: `services/upload_service.go::readTextureFile`
  (PNG decode + size gate), `services/upload_service.go::isValidSkinSize`,
  `services/upload_service.go::isValidCapeSize`;
  error mapping `controllers/upload_controller.go::UploadTexture`
  (service-error → status/code switch).
- Cape 22x17 → 64x32 normalization: `services/upload_service.go::readTextureFile`
  (22x17 branch), `services/upload_service.go::normalizeCapeImage`.
- SHA-256 content addressing: `services/upload_service.go::calculateHash`;
  flat storage path `services/upload_service.go::saveTextureFile`.
- Upsert on (uid, hash): `services/upload_service.go::upsertSkinRecord`
  (skin), `services/upload_service.go::upsertCapeRecord` (cape).
- WebP preview rendering: `services/upload_service.go::encodeWebP`
  (encoder), `services/upload_service.go::generatePreviewImage` (pipeline),
  `services/upload_service.go::renderSkinPreview` /
  `services/upload_service.go::renderCapePreview` (skin/cape composition),
  `services/upload_service.go::previewScale` (`previewScale = 8`),
  `services/upload_service.go::buildPreviewFileName` (preview file naming).
