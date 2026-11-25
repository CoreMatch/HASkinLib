<div align="center">
  <h1>Armorial Skin Library</h1>
  <p><strong>为每一款值得珍藏的皮肤与披风安家。</strong></p>
  <p><a href="./README.md">English</a></p>
</div>

Armorial Skin Library 是 Armorial 生态的皮肤与披风库，为需要上传、存储、预览与公开获取的玩家和服务而建。

它采用 MySQL 存储，用户操作由 OAuth2 保护，并且不依赖 Redis。

## 为什么选择 Armorial Skin Library

- **上传并预览。** 在发布前检查皮肤与披风效果。
- **适配各种模型。** 同时支持默认与纤细玩家模型。
- **留住经典披风。** 保留对早期 Minecraft 世代披风样式的访问。
- **轻松浏览大型资源库。** 通过标签与分页查找纹理，并配有 WebP 预览。
- **按公开哈希获取。** 无需用户会话即可获取共享纹理数据。
- **保持轻量依赖。** 只依赖 MySQL，无需 Redis。

## 进一步了解

- [皮肤库 API](docs/skinlib-api.md) — 本地接口参考
- [Armorial Wiki](https://wiki.mcnb.dev/) — 生态文档
- [Armorial API Contracts](https://github.com/CoreMatch/HA-Contract) — 生态 API 契约

## 限制

目前支持公开获取，但尚未提供纹理详情与编辑 API。

## 许可证与来源

[AGPL-3.0](LICENSE)。来源：[CoreMatch/HASkinLib](https://github.com/CoreMatch/HASkinLib)。
