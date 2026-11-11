# 0005 — Division of labor with HASkinProxy and Yggdrasil-API

## Status

Accepted

## Context

Three services touch "skins" in the HA stack and their boundaries are easy
to confuse:

- **HASkinLib** (this repo) — user-facing texture library.
- **HRPAuth-Yggdrasil-API** — the Minecraft-facing Yggdrasil protocol
  service (auth, sessions, texture download URLs inside profiles).
- **HASkinProxy** — CustomSkinLoader (CSL) protocol translation for clients
  that speak CSL instead of Yggdrasil.

## Decision

Responsibility split:

1. **HASkinLib = storage & management.** Upload validation, metadata,
   gallery, previews, deletion, and raw-byte serving by hash
   (`GET /texture/pull/:hash`). It never speaks Yggdrasil.
2. **HRPAuth-Yggdrasil-API = protocol surface for Minecraft clients.** It
   serves `GET /textures/:hash`, `GET /previews/:fileName`, profile texture
   properties, etc., and builds texture URLs from its configured callback
   URL. It reads texture bytes **from its own local filesystem**
   (`textures_storage` + `/textures/<hash>`), not from HASkinLib over HTTP.
3. **HASkinProxy = stateless CSL translation + cache.** It resolves
   username → UUID → Yggdrasil profile against the upstream Yggdrasil
   surface and re-shapes the result into CSL JSON; it has no database.

The texture-byte handoff between HASkinLib and Yggdrasil-API is therefore
**filesystem-level**: HASkinLib writes `<storage_dir>/<hash>`; the
Yggdrasil-API service reads `<textures_storage>/textures/<hash>`. The two
services deliberately do **not** share one configured mount point directly;
the deployment layer points both at the same content (e.g.
`storage_dir` set to `<textures_storage>/textures`).

## Consequences

- Benefit: no HTTP coupling between the two storage-adjacent services;
  texture download latency is local file I/O on the Yggdrasil-API side.
- Design note (confirmed by the author): the differing mount
  points are **intentional separation, not a gap** — storage and management
  belong to HASkinLib, while the protocol surface reads from Yggdrasil-API's
  own local directory, and the deployment layer wires the two to the same
  content. The Cupbearer templates intentionally use different relative paths
  (`./data/textures` for HASkinLib vs `./storage` textures_storage for
  Yggdrasil-API); that separation is by design.
- Constraint: preview file naming (`<hash>_<type>.webp`) and the flat
  `<hash>` texture naming are de facto shared conventions; changing either
  breaks the Yggdrasil-API side silently (404s).
- HASkinProxy does not call HASkinLib at all; its texture fetches go to the
  Yggdrasil surface (`/textures/:hash`). See HASkinProxy's knowledge base.

## Evidence

- HASkinLib storage layout: `services/upload_service.go::saveTextureFile`
  (flat `<storage_dir>/<hash>`), `services/upload_service.go::savePreviewFile`
  + `services/upload_service.go::buildPreviewFileName` (preview naming).
- HASkinLib pull endpoint (public raw bytes): `controllers/pull_texture_controller.go::Pull`.
- Yggdrasil-API serves textures from local disk (repo HRPAuth-Yggdrasil-API
  @ `124b2fa1`):
  [xref] `/home/lnb/Desktop/HA/hrpauth/HRPAuth-Yggdrasil-API/services/texture_service.go:L232-L243 @ 124b2fa1`
  (`GetTexturePath` = `<textures_storage>/textures/<hash>`),
  [xref] `/home/lnb/Desktop/HA/hrpauth/HRPAuth-Yggdrasil-API/services/texture_service.go:L41-L42 @ 124b2fa1`
  (subdir constants),
  [xref] `/home/lnb/Desktop/HA/hrpauth/HRPAuth-Yggdrasil-API/controllers/yggdrasil_controller.go:L575-L592 @ 124b2fa1`
  (`DownloadTexture`), routes in
  [xref] `/home/lnb/Desktop/HA/hrpauth/HRPAuth-Yggdrasil-API/main.go:L187-L189 @ 124b2fa1`.
- Yggdrasil-API builds profile texture URLs from callback URL:
  [xref] `/home/lnb/Desktop/HA/hrpauth/HRPAuth-Yggdrasil-API/services/texture_service.go:L316 @ 124b2fa1`
  (`callbackURL + "/textures/" + hash`).
- Cupbearer templates (repo HA-Cupbearer @ `657bcb88`):
  [xref] `/home/lnb/Desktop/HA/maintanance/HA-Cupbearer/deploy/config/haskinlib.yaml.tmpl:L14-L20 @ 657bcb88`
  (`storage_dir: ./data/textures`, `preview_storage_dir: ./data/previews`),
  [xref] `/home/lnb/Desktop/HA/maintanance/HA-Cupbearer/deploy/config/hrpauth-yggdrasil-api.yaml.tmpl:L72 @ 657bcb88`
  (`textures_storage: ./storage`). The differing paths are the intentional
  separation described above; the deployment layer points both at the same
  content.
- HASkinProxy's upstream is the Yggdrasil surface, not HASkinLib: repo
  HASkinProxy,
  [xref] `/home/lnb/Desktop/HA/hrpauth/HASkinProxy/internal/hrpauth/client.go::FetchTexture`
  (`FetchTexture` → `{base}/textures/:hash`); see HASkinProxy knowledge
  base entry 0002.
