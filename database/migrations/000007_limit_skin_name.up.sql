UPDATE `texture_list_skin`
SET `name` = LEFT(`name`, 20)
WHERE CHAR_LENGTH(`name`) > 20;

ALTER TABLE `texture_list_skin`
  MODIFY COLUMN `name` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '展示名';