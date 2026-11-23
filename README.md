<div align="center">
  <h1>HASkinLib</h1>
  <p><strong>A home for every skin and cape worth keeping.</strong></p>
  <p><a href="./README.zh-CN.md">简体中文</a></p>
</div>

HASkinLib is the HA ecosystem's skin and cape library for players and services that need upload, storage, preview, and public retrieval.

It pairs MySQL storage with OAuth2-protected user operations and has no Redis dependency.

## Why HASkinLib

- **Upload and preview.** Inspect skins and capes before publishing them.
- **Fit every model.** Support both default and slim player models.
- **Keep legacy capes.** Preserve access to cape styles from earlier Minecraft generations.
- **Browse large libraries.** Find textures with tags and pagination, backed by WebP previews.
- **Serve by public hash.** Retrieve shared texture data without a user session.
- **Use a simple footprint.** Rely on MySQL without Redis.

## Learn more

- [Skin library API](docs/skinlib-api.md) — local interface reference
- [HA Wiki](https://wiki.mcnb.dev/) — ecosystem documentation
- [HA-Contract](https://github.com/CoreMatch/HA-Contract) — ecosystem API contracts

## Limits

Public retrieval is available; texture detail and edit APIs are not currently provided.

## License & provenance

[AGPL-3.0](LICENSE). Source: [CoreMatch/HASkinLib](https://github.com/CoreMatch/HASkinLib).
