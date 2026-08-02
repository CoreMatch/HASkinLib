SET @has_texture_list_cape := (
  SELECT COUNT(*)
  FROM information_schema.TABLES
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'texture_list_cape'
);

SET @has_model_column := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'texture_list_cape'
    AND COLUMN_NAME = 'model'
);

SET @add_model_column_sql := IF(
  @has_texture_list_cape > 0 AND @has_model_column = 0,
  'ALTER TABLE `texture_list_cape` ADD COLUMN `model` enum(''default'',''slim'') COLLATE utf8mb4_unicode_ci DEFAULT ''default'' COMMENT ''披风模型占位, 当前固定 default'' AFTER `uid`',
  'SELECT 1'
);
PREPARE stmt FROM @add_model_column_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
