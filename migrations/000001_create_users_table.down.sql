DROP TABLE IF EXISTS `user`;
-- 统一使用反引号`xxx`，最稳, 有时候包含特殊字符（如 -、空格）, 可能和函数名 / 关键字冲突
-- Migration（结构变更）
-- 用途：建表, 加字段, 改索引, 改约束
-- 特点：只做结构, 必须有 up / down, 必须可回滚, 必须可重复执行（IF EXISTS / IF NOT EXISTS）