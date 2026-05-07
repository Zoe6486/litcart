CREATE TABLE `community` (
    `id`             BIGINT       NOT NULL AUTO_INCREMENT,
    `community_id`   BIGINT       NOT NULL COMMENT 'Application-level unique identifier (snowflake)',
    `community_name` VARCHAR(128) NOT NULL COMMENT 'Unique community name',
    `introduction`   VARCHAR(256)          COMMENT 'Brief description of the community',
    `status`         TINYINT      NOT NULL DEFAULT 1 COMMENT '1:active 2:archived 3:deleted',
    `created_at`     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uidx_community_id`   (`community_id`),
    UNIQUE KEY `uidx_community_name` (`community_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Communities';