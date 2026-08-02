-- 回滚为单表 texture_list，并恢复 type 列。

CREATE TABLE IF NOT EXISTS `texture_list` (
  `id` int NOT NULL AUTO_INCREMENT,
  `hash` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '纹理文件 sha256',
  `type` enum('skin','cape') COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '纹理类型',
  `uid` int unsigned NOT NULL COMMENT '所属用户, 对应 users.uid',
  `model` enum('default','slim') COLLATE utf8mb4_unicode_ci DEFAULT 'default' COMMENT '皮肤模型',
  `width` int NOT NULL DEFAULT 0 COMMENT '像素宽',
  `height` int NOT NULL DEFAULT 0 COMMENT '像素高',
  `file_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '原始文件名',
  `previewfile` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '预览图文件名',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '展示名',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '描述',
  `tags` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '标签, 逗号分隔',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_texture_list_uid_hash_type` (`uid`, `hash`, `type`),
  KEY `idx_texture_list_uid` (`uid`),
  CONSTRAINT `texture_list_ibfk_1` FOREIGN KEY (`uid`) REFERENCES `users` (`uid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET @has_texture_list_skin := (
  SELECT COUNT(*)
  FROM information_schema.TABLES
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'texture_list_skin'
);

SET @has_texture_list_cape := (
  SELECT COUNT(*)
  FROM information_schema.TABLES
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'texture_list_cape'
);

SET @copy_skin_back_sql := IF(
  @has_texture_list_skin > 0,
  'INSERT IGNORE INTO `texture_list` (`id`, `hash`, `type`, `uid`, `model`, `width`, `height`, `file_name`, `previewfile`, `name`, `description`, `tags`, `created_at`, `updated_at`) SELECT `id`, `hash`, ''skin'', `uid`, `model`, `width`, `height`, `file_name`, `previewfile`, `name`, `description`, `tags`, `created_at`, `updated_at` FROM `texture_list_skin`',
  'SELECT 1'
);
PREPARE stmt FROM @copy_skin_back_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @copy_cape_back_sql := IF(
  @has_texture_list_cape > 0,
  'INSERT IGNORE INTO `texture_list` (`id`, `hash`, `type`, `uid`, `model`, `width`, `height`, `file_name`, `previewfile`, `name`, `description`, `tags`, `created_at`, `updated_at`) SELECT `id`, `hash`, ''cape'', `uid`, ''default'', `width`, `height`, `file_name`, `previewfile`, `name`, `description`, `tags`, `created_at`, `updated_at` FROM `texture_list_cape`',
  'SELECT 1'
);
PREPARE stmt FROM @copy_cape_back_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

DROP TABLE IF EXISTS `texture_list_skin`;
DROP TABLE IF EXISTS `texture_list_cape`;
