-- HASkinLib 自有表：纹理列表。
-- 注意：本迁移只管理自有表，users 表由 HRPAuth 负责创建与维护，
-- 此处仅以外键引用 users.uid，不创建、不修改 users 表。
CREATE TABLE IF NOT EXISTS `texture_list` (
  `id` int NOT NULL AUTO_INCREMENT,
  `hash` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '纹理文件 sha256',
  `type` enum('skin','cape') COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '纹理类型',
  `uid` int unsigned NOT NULL COMMENT '所属用户, 对应 users.uid',
  `model` enum('default','slim') COLLATE utf8mb4_unicode_ci DEFAULT 'default' COMMENT '皮肤模型',
  `width` int NOT NULL DEFAULT 0 COMMENT '像素宽',
  `height` int NOT NULL DEFAULT 0 COMMENT '像素高',
  `file_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '原始文件名',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '展示名',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '描述',
  `tags` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '标签, 逗号分隔',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_texture_list_hash` (`hash`),
  KEY `idx_texture_list_uid` (`uid`),
  CONSTRAINT `texture_list_ibfk_1` FOREIGN KEY (`uid`) REFERENCES `users` (`uid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
