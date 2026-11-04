# 0005 — Division of labor with HASkinProxy and Yggdrasil-API

## Status

Accepted

## Date

2026-11-04 (recorded when the knowledge base was created; the division
itself emerged with the pull endpoint and preview API, 2026-08-02)

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
Yggdrasil-API service reads `<textures_storage>/textures/<hash>`. For the
end-to-end flow to work, deployment must make these point at the same
content (same volume, or `storage_dir` set to `<textures_storage>/textures`).

## Consequences

- Benefit: no HTTP coupling between the two storage-adjacent services;
  texture download latency is local file I/O on the Yggdrasil-API side.
- Cost / gap: the wiring is a deployment-time invariant with **no runtime
  check**. The deployer templates use different relative paths
  (`./data/textures` for HASkinLib vs `./storage` textures_storage for
  Yggdrasil-API), so the invariant holds only via the actual deployment
  layout (volumes/mounts). This is **not verified** from the repositories
  alone; confirm against the running deployment before debugging
  "uploaded skin not visible in game" issues.
- Constraint: preview file naming (`<hash>_<type>.webp`) and the flat
  `<hash>` texture naming are de facto shared conventions; changing either
  breaks the Yggdrasil-API side silently (404s).
- HASkinProxy does not call HASkinLib at all; its texture fetches go to the
  Yggdrasil surface (`/textures/:hash`). See HASkinProxy's knowledge base.

## Evidence

- HASkinLib storage layout: `services/upload_service.go:383-405` (flat
  `<storage_dir>/<hash>`), `:407-427` (preview naming).
- HASkinLib pull endpoint (public raw bytes): `controllers/pull_texture_controller.go:26-59`.
- Yggdrasil-API serves textures from local disk (repo HRPAuth-Yggdrasil-API
  @ `124b2fa1`): `services/texture_service.go:232-243`
  (`GetTexturePath` = `<textures_storage>/textures/<hash>`), `:41-42`
  (subdir constants), `controllers/yggdrasil_controller.go:575-592`
  (`DownloadTexture`), routes in `main.go:187-189`.
- Yggdrasil-API builds profile texture URLs from callback URL:
  `services/texture_service.go:316` (`callbackURL + "/textures/" + hash`).
- Deployer templates (repo HA-Services-Deployer @ `657bcb88`):
  `deploy/config/haskinlib.yaml.tmpl:14-20`
  (`storage_dir: ./data/textures`, `preview_storage_dir: ./data/previews`),
  `deploy/config/hrpauth-yggdrasil-api.yaml.tmpl:72`
  (`textures_storage: ./storage`). Path mismatch ⇒ deployment-level wiring
  required; **not verified** against a live deployment.
- HASkinProxy's upstream is the Yggdrasil surface, not HASkinLib: repo
  HASkinProxy, `internal/hrpauth/client.go` (`FetchTexture` →
  `{base}/textures/:hash`); see HASkinProxy knowledge base entry 0002.
