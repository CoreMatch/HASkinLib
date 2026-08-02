-- 兼容已执行旧版 000005 的数据库：删除 texture_list_cape.model 列。
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

SET @drop_model_column_sql := IF(
  @has_texture_list_cape > 0 AND @has_model_column > 0,
  'ALTER TABLE `texture_list_cape` DROP COLUMN `model`',
  'SELECT 1'
);
PREPARE stmt FROM @drop_model_column_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
