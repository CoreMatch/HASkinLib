UPDATE `texture_list_cape`
SET `name` = LEFT(`name`, 20)
WHERE CHAR_LENGTH(`name`) > 20;

ALTER TABLE `texture_list_cape`
  MODIFY COLUMN `name` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '展示名';