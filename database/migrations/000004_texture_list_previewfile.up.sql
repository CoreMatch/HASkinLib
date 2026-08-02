-- 为 texture_list 增加预览图文件名列。
SET @has_previewfile_column := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'texture_list'
    AND COLUMN_NAME = 'previewfile'
);
SET @add_previewfile_column_sql := IF(
  @has_previewfile_column = 0,
  'ALTER TABLE `texture_list` ADD COLUMN `previewfile` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT ''预览图文件名'' AFTER `file_name`',
  'SELECT 1'
);
PREPARE stmt FROM @add_previewfile_column_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
