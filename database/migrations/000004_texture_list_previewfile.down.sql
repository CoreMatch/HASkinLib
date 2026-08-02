SET @has_previewfile_column := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'texture_list'
    AND COLUMN_NAME = 'previewfile'
);
SET @drop_previewfile_column_sql := IF(
  @has_previewfile_column > 0,
  'ALTER TABLE `texture_list` DROP COLUMN `previewfile`',
  'SELECT 1'
);
PREPARE stmt FROM @drop_previewfile_column_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
