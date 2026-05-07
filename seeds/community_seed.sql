-- Initial communities for a Reddit-like forum
-- Safe to re-run (idempotent)
-- 可以插入时间，也可以不插入自动生成，这里就不写了
INSERT INTO community (community_id, community_name, introduction)
VALUES
  (1, 'golang', 'Discussion about the Go programming language'),
  (2, 'frontend', 'Frontend development: React, CSS, performance'),
  (3, 'backend', 'Backend engineering, APIs, databases'),
  (4, 'devops', 'DevOps, infrastructure, CI/CD, cloud'),
  (5, 'career', 'Career advice, interviews, and growth'),
  (6, 'random', 'Anything interesting, off-topic discussions')
ON DUPLICATE KEY UPDATE
  community_name = VALUES(community_name),
  introduction = VALUES(introduction);

-- 注意：1. --后面必须有一个空格（必须是-- ，而不是--中文紧贴着，--紧贴着就报错了）
-- 2. ON DUPLICATE KEY UPDATE的用途：
-- 因为000002_create_community_table.up 的表里有两个唯一约束：
-- UNIQUE KEY idx_community_id (community_id)
-- UNIQUE KEY idx_community_name (community_name)
-- 所以只要发生下面任何一种情况：community_id 已存在或 community_name 已存在
-- 👉 就会触发 ON DUPLICATE KEY UPDATE