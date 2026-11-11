# HASkinLib

本项目是 HA-Contract 的皮肤库服务组件。

## 文档

所有接口文档和开发规范已统一收口至根目录：
- [API 标准与规范](https://github.com/CoreMatch/HA-Contract/blob/main/docs/api/README.md)
- [OpenAPI 定义](https://github.com/CoreMatch/HA-Contract/blob/main/docs/api/openapi/haskinlib-business.yaml)（2026-11-04 按当前实现重建）
- [错误码表](https://github.com/CoreMatch/HA-Contract/blob/main/docs/api/error-codes.md)

本仓库内的 [docs/skinlib-api.md](./docs/skinlib-api.md) 继续作为与代码同步维护的接口参考。

旧文档已归档至 [archive/docs/HASkinLib](https://github.com/CoreMatch/HA-Contract/tree/main/archive/docs/HASkinLib)。

## 部署事实

部署事实以仓库根目录的 [`deploy-manifest.yaml`](deploy-manifest.yaml) 为准
（供 HA-Cupbearer 编排器读取；`manager: aapanel` 表示进程由面板管理，
编排器只读不碰）：

- 监听 `:2701`（对外）；健康端点 `GET /health`。
- 依赖 MySQL；运行时经 `core_api` 调用 HRPAuth（非启动硬依赖）。无 Redis 依赖。
- 配置文件 `config.yaml`（模板版本 7）；数据库迁移启动时自动执行。
- 数据目录：`./data/textures`、`./data/previews`。
