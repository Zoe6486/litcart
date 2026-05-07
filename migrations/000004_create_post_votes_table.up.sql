-- CREATE TABLE `post_votes` (
--     `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
--     `post_id`     BIGINT UNSIGNED NOT NULL COMMENT '帖子业务ID(对应 post.post_id)',
--     `user_id`     BIGINT UNSIGNED NOT NULL COMMENT '投票用户ID',
--     `vote_type`   TINYINT NOT NULL COMMENT '投票类型: 1 = 点赞(like),-1 = 点踩(dislike), 0 = 取消投票（可选）',
--     `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '投票时间',
--     `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间（用于取消/改投时）',

--     -- 主键和唯一约束
--     PRIMARY KEY (`id`),
    
--     -- 核心防重：同一个用户对同一个帖子只能有一种投票状态
--     UNIQUE KEY `uk_post_user` (`post_id`, `user_id`),
    
--     -- 常见查询加速
--     KEY `idx_post_id` (`post_id`),              -- 按帖子统计点赞/点踩总数
--     KEY `idx_user_id` (`user_id`),              -- 查用户投过哪些帖子（个人点赞列表）
--     KEY `idx_post_type` (`post_id`, `vote_type`) -- 组合索引，加速按类型统计（如只看点赞）

-- ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=' 帖子投票表（点赞/点踩）';

-- 中文comment railway mysql报错改成英文版了
CREATE TABLE `post_votes` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Primary key',
    `post_id`     BIGINT UNSIGNED NOT NULL COMMENT 'Post business ID (post.post_id)',
    `user_id`     BIGINT UNSIGNED NOT NULL COMMENT 'Voting user ID',
    `vote_type`   TINYINT NOT NULL COMMENT '1=like, -1=dislike',
    `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Vote time',
    `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last updated',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_post_user` (`post_id`, `user_id`),
    KEY `idx_post_id` (`post_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_post_type` (`post_id`, `vote_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Post votes (like/dislike)';