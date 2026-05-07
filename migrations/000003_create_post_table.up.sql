CREATE TABLE `post` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `post_id` BIGINT NOT NULL COMMENT '业务帖子id',
    `title` VARCHAR(128) NOT NULL COMMENT '标题',
    `content` TEXT NOT NULL COMMENT '内容',
    `author_id` BIGINT NOT NULL COMMENT '作者id',
    `community_id` BIGINT NOT NULL COMMENT '所属社区',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 1正常 2删除 3审核中',
    `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_post_id` (`post_id`),
    KEY `idx_author_id` (`author_id`),
    KEY `idx_community_id` (`community_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- COLLATE=utf8mb4_unicode_ci排序规则