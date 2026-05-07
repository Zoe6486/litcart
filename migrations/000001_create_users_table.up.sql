-- USE `db_litcart`;
-- CREATE TABLE `user` (
--     `id` bigint(20) NOT NULL AUTO_INCREMENT,
--     `user_id` bigint(20) NOT NULL,
--     `username` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
--     `password` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
--     `email` varchar(64) COLLATE utf8mb4_general_ci,
--     `gender` tinyint(4) NOT NULL DEFAULT '0',
--     `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
--     `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
--     PRIMARY KEY (`id`),
--     UNIQUE KEY `idx_username` (`username`) USING BTREE,
--     UNIQUE KEY `idx_user_id` (`user_id`) USING BTREE
-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
CREATE TABLE `user` (
    `id`          BIGINT       NOT NULL AUTO_INCREMENT,
    `user_id`     BIGINT       NOT NULL COMMENT 'Application-level unique identifier (snowflake)',
    `username`    VARCHAR(32)  NOT NULL COMMENT 'Unique display name',
    `email`       VARCHAR(255) NOT NULL COMMENT 'Unique email address used for login and recovery',
    `password`    VARCHAR(255) NOT NULL COMMENT 'bcrypt hashed password',
    `status`      TINYINT      NOT NULL DEFAULT 1 COMMENT '1:active 2:suspended 3:deleted',
    `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uidx_user_id`  (`user_id`),
    UNIQUE KEY `uidx_username` (`username`),
    UNIQUE KEY `uidx_email`    (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User accounts';