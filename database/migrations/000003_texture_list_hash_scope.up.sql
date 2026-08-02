-- 允许不同用户复用同一纹理文件；同一用户下相同 type+hash 只保留一条记录。
SET @has_old_hash_unique := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'texture_list'
    AND INDEX_NAME = 'uk_texture_list_hash'
);
SET @drop_old_hash_unique_sql := IF(
  @has_old_hash_unique > 0,
  'ALTER TABLE `texture_list` DROP INDEX `uk_texture_list_hash`',
  'SELECT 1'
);
PREPARE stmt FROM @drop_old_hash_unique_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_uid_hash_type_unique := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'texture_list'
    AND INDEX_NAME = 'uk_texture_list_uid_hash_type'
);
SET @add_uid_hash_type_unique_sql := IF(
  @has_uid_hash_type_unique = 0,
  'ALTER TABLE `texture_list` ADD UNIQUE KEY `uk_texture_list_uid_hash_type` (`uid`, `hash`, `type`)',
  'SELECT 1'
);
PREPARE stmt FROM @add_uid_hash_type_unique_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
