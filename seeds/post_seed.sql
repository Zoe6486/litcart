INSERT INTO post (post_id, title, content, author_id, community_id, status)
VALUES (
    FLOOR(RAND() * 9000000000000000) + 1000000000000000,  -- snowflake 大数字, 
    'test title', 
    'test content', 
    (SELECT user_id FROM user LIMIT 1), -- 选择user数据库第一个user
    (SELECT community_id FROM community LIMIT 1), -- 选择community数据库第一个community
    1
);