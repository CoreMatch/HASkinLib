-- 将 texture_list 拆分为 texture_list_skin / texture_list_cape，并移除 type 列。

CREATE TABLE IF NOT EXISTS `texture_list_skin` (
  `id` int NOT NULL AUTO_INCREMENT,
  `hash` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '纹理文件 sha256',
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
  UNIQUE KEY `uk_texture_list_skin_uid_hash` (`uid`, `hash`),
  KEY `idx_texture_list_skin_uid` (`uid`),
  CONSTRAINT `texture_list_skin_ibfk_1` FOREIGN KEY (`uid`) REFERENCES `users` (`uid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `texture_list_cape` (
  `id` int NOT NULL AUTO_INCREMENT,
  `hash` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '纹理文件 sha256',
  `uid` int unsigned NOT NULL COMMENT '所属用户, 对应 users.uid',
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
  UNIQUE KEY `uk_texture_list_cape_uid_hash` (`uid`, `hash`),
  KEY `idx_texture_list_cape_uid` (`uid`),
  CONSTRAINT `texture_list_cape_ibfk_1` FOREIGN KEY (`uid`) REFERENCES `users` (`uid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET @has_old_texture_list := (
  SELECT COUNT(*)
  FROM information_schema.TABLES
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'texture_list'
);

SET @copy_skin_sql := IF(
  @has_old_texture_list > 0,
  'INSERT IGNORE INTO `texture_list_skin` (`id`, `hash`, `uid`, `model`, `width`, `height`, `file_name`, `previewfile`, `name`, `description`, `tags`, `created_at`, `updated_at`) SELECT `id`, `hash`, `uid`, `model`, `width`, `height`, `file_name`, `previewfile`, `name`, `description`, `tags`, `created_at`, `updated_at` FROM `texture_list` WHERE `type` = ''skin''',
  'SELECT 1'
);
PREPARE stmt FROM @copy_skin_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @copy_cape_sql := IF(
  @has_old_texture_list > 0,
  'INSERT IGNORE INTO `texture_list_cape` (`id`, `hash`, `uid`, `width`, `height`, `file_name`, `previewfile`, `name`, `description`, `tags`, `created_at`, `updated_at`) SELECT `id`, `hash`, `uid`, `width`, `height`, `file_name`, `previewfile`, `name`, `description`, `tags`, `created_at`, `updated_at` FROM `texture_list` WHERE `type` = ''cape''',
  'SELECT 1'
);
PREPARE stmt FROM @copy_cape_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @drop_old_texture_list_sql := IF(
  @has_old_texture_list > 0,
  'DROP TABLE `texture_list`',
  'SELECT 1'
);
PREPARE stmt FROM @drop_old_texture_list_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
