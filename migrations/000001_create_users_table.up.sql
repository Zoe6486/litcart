-- CREATE TABLE `user` (
--     `id`          BIGINT       NOT NULL AUTO_INCREMENT,
--     `user_id`     BIGINT       NOT NULL COMMENT 'Application-level unique identifier (snowflake)',
--     `username`    VARCHAR(32)  NOT NULL COMMENT 'Unique display name',
--     `email`       VARCHAR(255) NOT NULL COMMENT 'Unique email address used for login and recovery',
--     `password`    VARCHAR(255) NOT NULL COMMENT 'bcrypt hashed password',
--     `status`      TINYINT      NOT NULL DEFAULT 1 COMMENT '1:active 2:suspended 3:deleted',
--     `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
--     PRIMARY KEY (`id`),
--     UNIQUE KEY `uidx_user_id`  (`user_id`),
--     UNIQUE KEY `uidx_username` (`username`),
--     UNIQUE KEY `uidx_email`    (`email`)
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User accounts';


-- User 表
--
-- 关键设计:
--   1. collation = utf8mb4_0900_ai_ci  → email/username 大小写不敏感(_ai_ci = accent/case insensitive)
--   2. created_at / updated_at 由 DB 自动维护,Go 端不主动赋值
--   3. uidx_username / uidx_email 唯一索引,与 mapInsertError 里的字符串匹配
--
-- 注意:utf8mb4_0900_ai_ci 是 MySQL 8 的默认值,MySQL 5.7 用 utf8mb4_unicode_ci 替代。

CREATE TABLE IF NOT EXISTS `user` (
    `id`             BIGINT          NOT NULL AUTO_INCREMENT COMMENT '自增主键(InnoDB 顺序写入用)',
    `user_id`        BIGINT          NOT NULL COMMENT '业务 ID(Snowflake)',
    `username`       VARCHAR(32)     NOT NULL,
    `email`          VARCHAR(255)    NOT NULL,
    `password`       VARCHAR(72)     NOT NULL COMMENT 'bcrypt hash',
    `status`         TINYINT         NOT NULL DEFAULT 1 COMMENT '1=active 2=suspended 3=deleted',
    `email_verified` TINYINT(1)      NOT NULL DEFAULT 0,
    `created_at`     TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uidx_user_id`  (`user_id`),
    UNIQUE KEY `uidx_username` (`username`),
    UNIQUE KEY `uidx_email`    (`email`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_0900_ai_ci;