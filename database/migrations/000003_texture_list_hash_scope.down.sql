SET @has_uid_hash_type_unique := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'texture_list'
    AND INDEX_NAME = 'uk_texture_list_uid_hash_type'
);
SET @drop_uid_hash_type_unique_sql := IF(
  @has_uid_hash_type_unique > 0,
  'ALTER TABLE `texture_list` DROP INDEX `uk_texture_list_uid_hash_type`',
  'SELECT 1'
);
PREPARE stmt FROM @drop_uid_hash_type_unique_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_old_hash_unique := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'texture_list'
    AND INDEX_NAME = 'uk_texture_list_hash'
);
SET @add_old_hash_unique_sql := IF(
  @has_old_hash_unique = 0,
  'ALTER TABLE `texture_list` ADD UNIQUE KEY `uk_texture_list_hash` (`hash`)',
  'SELECT 1'
);
PREPARE stmt FROM @add_old_hash_unique_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
